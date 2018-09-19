package main

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
)

// InsuranceChaincode simple Chaincode implementation
type InsuranceChaincode struct {
}

// ============================================================================================================================
// Concept Definitions - Concept struct types
// ============================================================================================================================

// AddressConcept - address type
type AddressConcept struct {
	Class        string `json:"$class"` //base.Address
	AddressLine1 string `json:"addressLine1"`
	AddressLine2 string `json:"addressLine2"`
	AddressLine3 string `json:"addressLine3,omitempty"`
}

// LocationConcept - location type
type LocationConcept struct {
	Class       string  `json:"$class"` // accident.Location
	Longitude   float64 `json:"longitude"`
	Latitude    float64 `json:"latitude"`
	Description string  `json:"description,omitempty"`
}

// GoodsConcept - goods type
type GoodsConcept struct {
	Class    string   `json:"$class"`   // accident.Goods
	Vehicles []string `json:"vehicles"` // class name + # +registrationid
}

// EstimateConcept - estimate type
type EstimateConcept struct {
	Class          string  `json:"$class"` // vehiclerepair.Estimate
	Type           string  `json:"type"`   // This can be either REPAIR or REPLACE
	Description    string  `json:"description"`
	CostOfParts    float32 `json:"costOfParts,omitempty"`
	CostOfLabor    float32 `json:"costOfLabor,omitempty"`
	CostOfRefinish float32 `json:"costOfRefinish,omitempty"`
	TotalCost      float32 `json:"totalCost"`
}

// ============================================================================================================================
// Abstract Definitions - Abstract struct types
// ============================================================================================================================

// CompanyAbstract - company type
type CompanyAbstract struct {
	TradeName string         `json:"tradeName"`
	Address   AddressConcept `json:"address"`
}

// ============================================================================================================================
// Participants Definitions - Users the ledger will store
// ============================================================================================================================

// Registrant - participanting policy holder, vehicle owner
type Registrant struct {
	Class                string         `json:"$class"` // base.Registrant
	IdentificationNumber string         `json:"identificationNumber"`
	LegalEntity          string         `json:"legalEntity"` // This can be INDIVIDUAL, CORPORATION or LEASER
	Name                 string         `json:"name"`
	Initials             string         `json:"initials,omitempty"`
	Address              AddressConcept `json:"address"`
}

// Insurer - participating insurer
type Insurer struct {
	Class string `json:"$class"` // base.Insurer
	CompanyAbstract
	Signature string `json:"signature"`
}

// EmergencyServices - participatin ERS
type EmergencyServices struct {
	Class string `json:"$class"` // base.EmergencyServices
	CompanyAbstract
	Location LocationConcept `json:"location"`
}

// RepairShop - participating repair shop
type RepairShop struct {
	Class string `json:"$class"` // base.RepairShop
	CompanyAbstract
	Phone string `json:"phone,omitempty"`
	Email string `json:"email,omitempty"`
}

// ============================================================================================================================
// Asset Definitions - Assets the ledger will store
// ============================================================================================================================

// Vehicle = asset type of vehicle
type Vehicle struct {
	Class              string    `json:"$class"` // base.Vehicle
	RegistrationNumber string    `json:"registrationNumber"`
	LicencePlate       string    `json:"licencePlate"`
	DateFirstAdmission time.Time `json:"dateFirstAdmission"`
	DateAscription     time.Time `json:"dateAscription"`
	Owner              string    `json:"owner"` // Registrant class name + # + registrationId
	Make               string    `json:"make"`
	Model              string    `json:"model"`
	Color              string    `json:"color,omitempty"`
	MaxMass            int       `json:"maxMass,omitempty"`
	MaxSeating         int       `json:"maxSeating"`
}

// AccidentReport - asset type of accident report
type AccidentReport struct {
	Class         string          `json:"$class"` // accident.AccidentReport
	AccidentID    string          `json:"accidentId"`
	OccuredAt     time.Time       `json:"occuredAt"`
	Status        string          `json:"status"` // This can be NEW, RESPONDING or RESOLVED
	Location      LocationConcept `json:"location"`
	Description   string          `json:"accidentDescription,omitempty"`
	InvolvedGoods GoodsConcept    `json:"involvedGoods,omitempty"`
	RespondingERS string          `json:"respondingERS,omitempty"` // Emergency Services class name + # + tradeName
}

// QuoteRequest - asset type of quote request
type QuoteRequest struct {
	Class             string `json:"$class"` // vehiclerepair.QuoteRequest
	RequestID         string `json:"requestId"`
	AccidentReport    string `json:"accidentReport"`   // Accident report class name + # + accidentId
	VehicleInsurance  string `json:"vehicleInsurance"` // Insurance policy class name + # + policyId
	DamageDescription string `json:"damageDescription"`
}

// RepairQuote - asset type of repair quote
type RepairQuote struct {
	Class         string            `json:"$class"` // vehiclerepair.RepairQuote
	QuoteID       string            `json:"quoteId"`
	QuoteRequest  string            `json:"quoteRequest"` // Quote request class name + # + requestId
	Estimator     string            `json:"estimator"`    // Repair shop class name + # + tradeName
	Estimates     []EstimateConcept `json:"estimates"`
	TotalParts    float32           `json:"totalParts"`
	TotalLabor    float32           `json:"totalLabor"`
	TotalRefinish float32           `json:"totalRefinish"`
	Tax           float32           `json:"tax"`
	Total         float32           `json:"total"`
}

// InsurancePolicy - asset type of insurance policy
type InsurancePolicy struct {
	Class             string    `json:"$class"` // insurance.InsurancePolicy
	PolicyID          string    `json:"policyId"`
	AutorisedBy       string    `json:"autorisedBy"`
	ValidFrom         time.Time `json:"validFrom"`
	ValidTo           time.Time `json:"validTo"`
	RegisteredVehicle string    `json:"registeredVehicle"` // Vehicle class name + # + registrationNumber
	CountryCode       string    `json:"countryCode"`
	InsurerCode       string    `json:"insurerCode"`
	PolicyNumber      int64     `json:"policyNumber"`
	VehicleCategory   string    `json:"vehicleCategory"`
	VehicleMake       string    `json:"vehicleMake"`
	Coverage          []string  `json:"coverage"`
	PolicyHolder      string    `json:"policyHolder"` // Registrant class name + # + identificationNumber
	IssuedBy          string    `json:"issuedBy"`     // Insurer class name + # + tradeName
}

// InsuranceClaim - asset type of insurance claim
type InsuranceClaim struct {
	Class          string    `json:"$class"` // insurance.InsuranceClaim
	ClaimID        string    `json:"claimId"`
	DateOfClaim    time.Time `json:"dateOfClaim"`
	Status         string    `json:"status"`         // This can be NEW, ACCEPTED, DECLINED or RESOLVED
	AccidentReport string    `json:"accidentReport"` // Accident report class name + # + accidentId
	Claimant       string    `json:"claimant"`       // Insurance policy class name + # + policyId
	Defendant      string    `json:"defendant"`      // Insurance policy class name + # + policyId
	CostOfRepair   string    `json:"costOfRepair"`   // Repair Quote class name + # + quoteId
}

// AssetEntry - entry of created asset, used in setup
type AssetEntry struct {
	Class   string `json:"$class"`
	AssetID string `json:"assetId"`
}

// ============================================================================================================================
// Event Definitions - Events the ledger will emit
// ============================================================================================================================

// NewAccidentEvent - new accident event type
type NewAccidentEvent struct {
	AccidentID string `json:"accidentId"`
	Location   string `json:"location"` //Longitude, Latitude
}

// ReportUpdateEvent - updated accident event type
type ReportUpdateEvent struct {
	AccidentID string `json:"accidentId"`
	Reason     string `json:"reason"`
}

// RequestForQuoteEvent - new quote request event type
type RequestForQuoteEvent struct {
	RequestID         string `json:"requestId"`
	VehicleMake       string `json:"vehicleMake"`
	VehicleModel      string `json:"vehicleModel"`
	DamageDescription string `json:"damageDescription"`
}

// NewQuoteOfferEvent - new repaire quote event type
type NewQuoteOfferEvent struct {
	RequestID     string  `json:"requestId"`
	QuoteID       string  `json:"quoteId"`
	TotalEstimate float32 `json:"totalEstimate"`
}

// NewClaimEvent - new insurance claim event type
type NewClaimEvent struct {
	ClaimID      string  `json:"claimId"`
	ClaimantID   string  `json:"claimantPolicyId"`
	DefendantID  string  `json:"defendantPolicyId"`
	CostOfRepair float32 `json:"costOfRepair"`
}

// ============================================================================================================================
// Chaincode functions
// ============================================================================================================================

// Main - Start the chaincode
func main() {
	err := shim.Start(new(InsuranceChaincode))
	if err != nil {
		fmt.Printf("Error starting InsuranceChain chaincode - %s", err)
	}
}

// Init -  Initialize the chaincode
func (t *InsuranceChaincode) Init(stub shim.ChaincodeStubInterface) pb.Response {
	return shim.Success(nil)
}

// Invoke - Our entry point for Invocations
func (t *InsuranceChaincode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	function, args := stub.GetFunctionAndParameters()
	fmt.Println("invoke is running " + function)

	// Handle different functions
	if function == "setupAssets" { // setup demo assets
		return t.setupAssets(stub, args)
	} else if function == "readAssetData" {
		return t.readAssetData(stub, args)
	} else if function == "reportAccident" { // report new accident
		return t.reportAccident(stub, args)
	} else if function == "updateReport" { // update accident report
		return t.updateReport(stub, args)
	} else if function == "requestQuote" { // request quote for repair
		return t.requestQuote(stub, args)
	} else if function == "offerQuote" { // offer repair quote
		return t.offerQuote(stub, args)
	} else if function == "issuePolicy" { // issue inusrance policy
		return t.issuePolicy(stub, args)
	} else if function == "sendClaim" { // offer repair quote
		return t.sendClaim(stub, args)
	}

	return shim.Error("Received unknown invoke function name - '" + function + "'")
}

// reportAccident - Create a new accident report, store into state
func (t *InsuranceChaincode) reportAccident(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var err error

	// simple data model arguments
	// 0=longitude  1=latitude  2=occuredAt               3=reporting vehicle
	// 52.0920511   5.06641270  2018-08-03T10:20:20.325Z  JN6ND01S3GX194659

	if len(args) < 2 {
		return shim.Error(fmt.Sprintf("Incorrect number of arguments. Expecting minimum of 2, got %d", len(args)))
	}

	// === Check input variables ===
	if len(args[0]) <= 0 {
		return shim.Error("1st argument must be a non-empty string")
	}
	if len(args[1]) <= 0 {
		return shim.Error("2nd argument must be a non-empty string")
	}

	longitude, err := strconv.ParseFloat(args[0], 64)
	if err != nil {
		return shim.Error("1st argument must be a floating point string")
	}

	latitude, err := strconv.ParseFloat(args[1], 64)
	if err != nil {
		return shim.Error("2nd argument must be a floating point string")
	}

	occuredAt := time.Now()

	// === Parse occuredAt dateTime format ===
	if len(args[2]) > 0 {
		occuredAt, err = time.Parse(time.RFC3339, args[2])
		if err != nil {
			return shim.Error("3rd argument must be a RFC3339 dateTime string")
		}
	}

	// === Check if optional vehicle exists ===
	var vehicleRef string
	if len(args[3]) > 0 {
		vehicleRef = fmt.Sprintf("%s#%s", "base.Vehicle", args[3])
		vehicleAsBytes, err := stub.GetState(vehicleRef)
		if err != nil {
			return shim.Error("Failed to get vehicle: " + err.Error())
		} else if vehicleAsBytes == nil {
			return shim.Error("This vehicle doesn't exists: " + vehicleRef)
		}
	}

	// === Create report object
	accidentObjClass := "accident.AccidentReport"
	//accidentID, err := strconv.ParseInt("1534180781", 10, 64) //static id for testing
	accidentID := time.Now().Unix()
	location := LocationConcept{"accident.Location", longitude, latitude, ""}
	accidentReport := &AccidentReport{Class: accidentObjClass, OccuredAt: occuredAt, Status: "NEW", Location: location}
	if vehicleRef != "" {
		var vehicles []string
		vehicles[0] = vehicleRef
		involvedGoods := GoodsConcept{"accident.Goods", vehicles}
		accidentReport.InvolvedGoods = involvedGoods
	}

	// === Marshal accident report
	accidentJSONasBytes, err := json.Marshal(accidentReport)
	if err != nil {
		return shim.Error(err.Error())
	}

	// === Save accident to state ===
	accidentRef := fmt.Sprintf("%s#%d", accidentObjClass, accidentID)
	err = stub.PutState(accidentRef, accidentJSONasBytes)
	if err != nil {
		return shim.Error(err.Error())
	}

	// === Emit NewAccident event ===
	locationStr := fmt.Sprintf("%f, %f", longitude, latitude)
	newAccident := &NewAccidentEvent{strconv.FormatInt(accidentID, 10), locationStr}
	eventJSONasBytes, err := json.Marshal(newAccident)
	if err != nil {
		return shim.Error(err.Error())
	}
	stub.SetEvent("NewAccidentEvent", eventJSONasBytes)

	fmt.Println("- Accident report successfully created")
	return shim.Success(eventJSONasBytes)

}

// updateReport - Update the report
func (t *InsuranceChaincode) updateReport(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var err error
	var reason string

	// simple data model arguments
	// 0=accidentId  1=respondingERS     2=description           3=other vehicle
	// 1534180781    NYPD 34th Precinct  Nose to tail collision  1HTZR0007JH586991

	if len(args) < 2 {
		return shim.Error("Incorrect number of arguments. Expecting minimum of 2")
	}

	// === Check input variables ===
	if len(args[0]) <= 0 {
		return shim.Error("1st argument must be a non-empty string")
	}
	if len(args[1]) <= 0 {
		return shim.Error("2nd argument must be a non-empty string")
	}

	accidentID := args[0]
	respondingERS := args[1]
	description := args[2]
	otherVehicle := args[3]

	// === Check if AccidentReport asset exists
	accidentRef := fmt.Sprintf("%s#%s", "accident.AccidentReport", accidentID)
	reportAsBytes, err := stub.GetState(accidentRef)
	if err != nil {
		return shim.Error("Failed to get accident report: " + err.Error())
	} else if reportAsBytes == nil {
		return shim.Error("This accident report doesn't exists: " + accidentRef)
	}

	// === Check if EmergencyServices asset exists
	ersRef := fmt.Sprintf("%s#%s", "base.EmergencyServices", respondingERS)
	ersAsBytes, err := stub.GetState(ersRef)
	if err != nil {
		return shim.Error("Failed to get accident report: " + err.Error())
	} else if ersAsBytes == nil {
		return shim.Error("This accident report doesn't exists: " + ersRef)
	}

	// === Unmarshal the report to an object
	accidentReport := AccidentReport{}
	if err = json.Unmarshal(reportAsBytes, accidentReport); err != nil {
		return shim.Error("Failed to unmarshal accident report: " + err.Error())
	}

	// === Update reponsing ERS if not yet assigned
	if accidentReport.RespondingERS == "" {
		accidentReport.RespondingERS = ersRef
		reason = fmt.Sprintf("Emergencency Services (%s) responding to accident", respondingERS)
	} else {
		return shim.Error("Emergency Services alsready responding: " + accidentReport.RespondingERS)
	}

	// === Check if description is given
	if len(description) > 0 {
		accidentReport.Description = description
		reason = "Description of accident updated"
	}

	// === Check if other vehicle exists
	if len(otherVehicle) > 0 {
		vehicleRef := fmt.Sprintf("%s#%s", "base.Vehicle", otherVehicle)
		vehicleAsBytes, err := stub.GetState(vehicleRef)
		if err != nil {
			return shim.Error("Failed to get vehicle: " + err.Error())
		} else if vehicleAsBytes == nil {
			return shim.Error("Added vehicle doesn't exists: " + vehicleRef)
		}

		var vehicles []string
		vehicles = accidentReport.InvolvedGoods.Vehicles
		vehicles = append(vehicles, otherVehicle)
		accidentReport.InvolvedGoods.Vehicles = vehicles
		reason = "Another vehicle added to the report"
	}

	// === Marshal the updated accident report
	accidentJSONasBytes, err := json.Marshal(accidentReport)
	if err != nil {
		return shim.Error(err.Error())
	}

	// === Save accident report to state ===
	err = stub.PutState(accidentRef, accidentJSONasBytes)
	if err != nil {
		return shim.Error(err.Error())
	}

	// === Emit ReportUpdate event ===
	reportUpdate := &ReportUpdateEvent{accidentID, reason}
	eventJSONasBytes, err := json.Marshal(reportUpdate)
	if err != nil {
		return shim.Error(err.Error())
	}
	stub.SetEvent("ReportUpdateEvent", eventJSONasBytes)

	fmt.Println("- Accident report successfully updated")
	return shim.Success(eventJSONasBytes)
}

// requestQuote - Request a new quote for repairs
func (t *InsuranceChaincode) requestQuote(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var err error

	// simple data model arguments
	// 0=accidentId  1=insurancePolicy  2=description
	// 1534180781    USA-AX203-3459802  Scratch on back bumper (2x0.1 inches)

	if len(args) != 3 {
		return shim.Error("Incorrect number of arguments. Expecting exactly 3")
	}

	// === Check input variables ===
	if len(args[0]) <= 0 {
		return shim.Error("1st argument must be a non-empty string")
	}
	if len(args[1]) <= 0 {
		return shim.Error("2nd argument must be a non-empty string")
	}
	if len(args[2]) <= 0 {
		return shim.Error("3rd argument must be a non-empty string")
	}

	accidentID := args[0]
	insurancePolicyID := args[1]
	description := args[2]

	// === Check if AccidentReport asset exists
	accidentRef := fmt.Sprintf("%s#%s", "accident.AccidentReport", accidentID)
	reportAsBytes, err := stub.GetState(accidentRef)
	if err != nil {
		return shim.Error("Failed to get accident report: " + err.Error())
	} else if reportAsBytes == nil {
		return shim.Error("This accident report doesn't exists: " + accidentRef)
	}

	// === Unmarshal the report to an object
	accidentReport := AccidentReport{}
	if err = json.Unmarshal(reportAsBytes, accidentReport); err != nil {
		return shim.Error("Failed to unmarshal accident report: " + err.Error())
	}

	// === Check if InsurancePolicy asset exists
	policyRef := fmt.Sprintf("%s#%s", "insurance.InsurancePolicy", insurancePolicyID)
	policyAsBytes, err := stub.GetState(policyRef)
	if err != nil {
		return shim.Error("Failed to get insurance policy: " + err.Error())
	} else if policyAsBytes == nil {
		return shim.Error("Given insurance policy doesn't exists: " + policyRef)
	}

	// === Unmarshal the policy to an object
	insurancePolicy := InsurancePolicy{}
	if err = json.Unmarshal(policyAsBytes, insurancePolicy); err != nil {
		return shim.Error("Failed to unmarshal insurance policy: " + err.Error())
	}

	// === Check if vehicle is involved in accident
	vehicles := accidentReport.InvolvedGoods.Vehicles

	// save the items in map
	vmap := make(map[string]bool)
	for i := 0; i < len(vehicles); i++ {
		vmap[vehicles[i]] = true
	}

	// Check if registered vehicle is involved in accident
	var vehicleReg string
	vehicleReg = insurancePolicy.RegisteredVehicle
	if _, ok := vmap[vehicleReg]; ok {
		return shim.Error("Insured vehicle is not involved accident: " + vehicleReg)
	}

	// === Retrieve insured vehicle
	vehicleRef := insurancePolicy.RegisteredVehicle
	vehicleAsBytes, err := stub.GetState(vehicleRef)
	if err != nil {
		return shim.Error("Failed to get vehicle: " + err.Error())
	} else if vehicleAsBytes == nil {
		return shim.Error("Insured vehicle doesn't exists: " + vehicleRef)
	}

	// === Unmarshal the vehicle to an object
	vehicle := Vehicle{}
	if err = json.Unmarshal(vehicleAsBytes, vehicle); err != nil {
		return shim.Error("Failed to unmarshal vehicle: " + err.Error())
	}

	// === Create new QuoteRequest object
	requestObjClass := "vehiclerepair.QuoteRequest"
	//requestID, err := strconv.ParseInt("1000000001", 10, 64) //static id for testing
	requestID := time.Now().Unix()
	quoteRequest := &QuoteRequest{requestObjClass, strconv.FormatInt(requestID, 10), accidentRef, policyRef, description}

	// === Marshal the quote request
	requestJSONasBytes, err := json.Marshal(quoteRequest)
	if err != nil {
		return shim.Error(err.Error())
	}

	// === Save request to state
	requestRef := fmt.Sprintf("%s#%d", requestObjClass, requestID)
	err = stub.PutState(requestRef, requestJSONasBytes)
	if err != nil {
		return shim.Error(err.Error())
	}

	// === Emit RequestForQuote event
	newQuoteRequest := &RequestForQuoteEvent{strconv.FormatInt(requestID, 10), vehicle.Make, vehicle.Model, description}
	eventJSONasBytes, err := json.Marshal(newQuoteRequest)
	if err != nil {
		return shim.Error(err.Error())
	}
	stub.SetEvent("RequestForQuoteEvent", eventJSONasBytes)

	fmt.Println("- Request of quote for repairs successfully created")
	return shim.Success(eventJSONasBytes)
}

// offerQuote - Offer a quote for repair
func (t *InsuranceChaincode) offerQuote(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var err error

	// simple data model arguments
	// 0=requestId  1=repairShop       2=json{estimates[]}                                                                                              3=tax (%)
	// 1534180781   USA Automotive NY  [{"type":"REPAIR","description":"Scratch removal","costOfPart":30.6,"costOfLabor":100,"totalCost":130.6},{...}]  11

	if len(args) != 4 {
		return shim.Error("Incorrect number of arguments. Expecting exactly 4")
	}

	// === Check input variables
	if len(args[0]) <= 0 {
		return shim.Error("1st argument must be a non-empty string")
	}
	if len(args[1]) <= 0 {
		return shim.Error("2nd argument must be a non-empty string")
	}
	if len(args[2]) <= 0 {
		return shim.Error("3rd argument must be a non-empty string")
	}
	if len(args[3]) <= 0 {
		return shim.Error("4th argument must be a non-empty string")
	}

	// === Initialize variables
	requestID := args[0]
	repairShopID := args[1]
	estimatesAsBytes := []byte(args[2])
	tax, err := strconv.ParseFloat(args[3], 32)
	if err != nil {
		return shim.Error("4th argument must be a floating point string")
	} else if 0 <= tax && tax >= 100 {
		return shim.Error("4th argument must be between 0 and 100")
	}

	// === Unmarshal estimates array and calculate totals
	var estimates []EstimateConcept
	if err = json.Unmarshal(estimatesAsBytes, &estimates); err != nil {
		return shim.Error("Failed to unmarshal estimate array: " + err.Error())
	}

	// === Check if QuoteRequest asset exists
	requestObjClass := "vehiclerepair.QuoteRequest"
	requestRef := fmt.Sprintf("%s#%s", requestObjClass, requestID)
	requestAsBytes, err := stub.GetState(requestRef)
	if err != nil {
		return shim.Error("Failed to get quote request: " + err.Error())
	} else if requestAsBytes == nil {
		return shim.Error("Given quote request doesn't exists: " + requestRef)
	}

	// === Check if RepairShop asset exists
	shopObjClass := "base.RepairShop"
	shopRef := fmt.Sprintf("%s#%s", shopObjClass, repairShopID)
	shopAsBytes, err := stub.GetState(shopRef)
	if err != nil {
		return shim.Error("Failed to get repair shop: " + err.Error())
	} else if shopAsBytes == nil {
		return shim.Error("Given repair shop doesn't exists: " + shopRef)
	}

	// === Calculate totals
	var totalParts float32
	var totalLabor float32
	var totalRefinish float32
	var totalEstimates float32
	var total float32
	for _, estimate := range estimates {
		totalParts = totalParts + estimate.CostOfParts
		totalLabor = totalLabor + estimate.CostOfLabor
		totalRefinish = totalRefinish + estimate.CostOfRefinish
		totalEstimates = total + estimate.TotalCost
	}
	total = totalEstimates * float32(tax/100+1)

	// === Create new RepairQuote object
	quoteObjClass := "vehiclerepair.RepairQuote"
	//quoteID, err := strconv.ParseInt("1000000001", 10, 64) //static id for testing
	quoteID := time.Now().Unix()
	repairQuote := &RepairQuote{quoteObjClass, strconv.FormatInt(quoteID, 10), requestRef, shopRef, estimates, totalParts, totalLabor, totalRefinish, float32(tax), total}

	// === Marshal the quote request
	quoteJSONasBytes, err := json.Marshal(repairQuote)
	if err != nil {
		return shim.Error(err.Error())
	}

	// === Save request to state
	quoteRef := fmt.Sprintf("%s#%d", quoteObjClass, quoteID)
	err = stub.PutState(quoteRef, quoteJSONasBytes)
	if err != nil {
		return shim.Error(err.Error())
	}

	// === Emit NewQuoteOffer event
	newQuoteOffer := &NewQuoteOfferEvent{requestID, strconv.FormatInt(quoteID, 10), totalEstimates}
	eventJSONasBytes, err := json.Marshal(newQuoteOffer)
	if err != nil {
		return shim.Error(err.Error())
	}
	stub.SetEvent("NewQuoteOfferEvent", eventJSONasBytes)

	fmt.Println("- Offer quote for repairs successfully created")
	return shim.Success(eventJSONasBytes)
}

// issuePolicy - Create a new insurance policy
func (t *InsuranceChaincode) issuePolicy(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var err error

	// simple data model arguments
	// authorisedBy|validFrom|validTo|registeredVehicle|
	// countryCode|insurerCode|policyNumber|vehicleCategory|vehicleMake|coverage|policyHolder|issuedBy
	//
	// State of New York|2018-08-01T00:00:00.000Z|2020-08-01T00:00:00.000Z|JN6ND01S3GX194659
	// USA|AX203|3459802|AF|BMW|US,CA,MX|908123764|AXA Insurance

	if len(args) != 12 {
		return shim.Error("Incorrect number of arguments. Expecting exactly 12")
	}

	// === Check input variables ===
	if len(args[0]) <= 0 {
		return shim.Error("1st argument must be a non-empty string")
	}
	if len(args[1]) <= 0 {
		return shim.Error("2nd argument must be a non-empty string")
	}
	if len(args[2]) <= 0 {
		return shim.Error("3rd argument must be a non-empty string")
	}
	if len(args[3]) <= 0 {
		return shim.Error("4th argument must be a non-empty string")
	}
	if len(args[4]) <= 0 {
		return shim.Error("5th argument must be a non-empty string")
	}
	if len(args[5]) <= 0 {
		return shim.Error("6th argument must be a non-empty string")
	}
	if len(args[6]) <= 0 {
		return shim.Error("7th argument must be a non-empty string")
	}
	if len(args[7]) <= 0 {
		return shim.Error("8th argument must be a non-empty string")
	}
	if len(args[8]) <= 0 {
		return shim.Error("9th argument must be a non-empty string")
	}
	if len(args[9]) <= 0 {
		return shim.Error("10th argument must be a non-empty string")
	}
	if len(args[10]) <= 0 {
		return shim.Error("11th argument must be a non-empty string")
	}
	if len(args[11]) <= 0 {
		return shim.Error("12th argument must be a non-empty string")
	}

	authorisedBy := args[0]
	validFrom, err := time.Parse(time.RFC3339, args[1])
	if err != nil {
		return shim.Error("2nd argument must be a RFC3339 dateTime string")
	}

	validTo, err := time.Parse(time.RFC3339, args[2])
	if err != nil {
		return shim.Error("3rd argument must be a RFC3339 dateTime string")
	}

	vehicleReg := args[3]
	countryCode := args[4]
	insurerCode := args[5]

	policyNumber, err := strconv.ParseInt(args[6], 10, 64)
	if err != nil {
		return shim.Error("7th argument must be a valid integer")
	}

	vehicleCat := args[7]
	vehicleMake := args[8]
	coverage := strings.Split(args[9], ",")
	policyHolder := args[10]
	issuedBy := args[11]

	// === Check if vehicle exists
	vehicleRef := fmt.Sprintf("%s#%s", "base.Vehicle", vehicleReg)
	vehicleAsBytes, err := stub.GetState(vehicleRef)
	if err != nil {
		return shim.Error("Failed to get vehicle: " + err.Error())
	} else if vehicleAsBytes == nil {
		return shim.Error("Given vehicle doesn't exists: " + vehicleRef)
	}

	// === Unmarshal vehicle asset
	var vehicle Vehicle
	if err = json.Unmarshal(vehicleAsBytes, &vehicle); err != nil {
		return shim.Error("Failed to unmarshal vehicle asset: " + err.Error())
	}
	// TODO check if vehicle is not already insured

	// === Check if policy holder exists
	holderRef := fmt.Sprintf("%s#%s", "base.Registrant", policyHolder)
	holderAsBytes, err := stub.GetState(holderRef)
	if err != nil {
		return shim.Error("Failed to get policy holder: " + err.Error())
	} else if holderAsBytes == nil {
		return shim.Error("Given policy holder doesn't exists: " + holderRef)
	}

	// === check if vehicle is owned by policy holder
	if vehicle.Owner != holderRef {
		return shim.Error("The vehicle is not owned by the newly assigned policy holder")
	}

	// === Check if insurer issueing policy exists
	insurerRef := fmt.Sprintf("%s#%s", "base.Insurer", issuedBy)
	insurerAsBytes, err := stub.GetState(insurerRef)
	if err != nil {
		return shim.Error("Failed to get insurer: " + err.Error())
	} else if insurerAsBytes == nil {
		return shim.Error("Given insurer doesn't exists: " + insurerRef)
	}

	// === Create policy object and marchal to JSON ===
	policyObjClass := "insurance.InsurancePolicy"
	policyID := fmt.Sprintf("%s-%s-%d", countryCode, insurerCode, policyNumber)
	insurancePolicy := &InsurancePolicy{policyObjClass, policyID, authorisedBy, validFrom, validTo, vehicleRef, countryCode, insurerCode, policyNumber, vehicleCat, vehicleMake, coverage, holderRef, insurerRef}
	policyJSONasBytes, err := json.Marshal(insurancePolicy)
	if err != nil {
		return shim.Error(err.Error())
	}

	// === Save insurance policy to state ===
	policyRef := fmt.Sprintf("%s#%s", policyObjClass, policyID)
	err = stub.PutState(policyRef, policyJSONasBytes)
	if err != nil {
		return shim.Error(err.Error())
	}

	fmt.Println("- Insurance policy successfully issued")
	return shim.Success(policyJSONasBytes)
}

// sendClaim - Send a new insurance claim to defendant
func (t *InsuranceChaincode) sendClaim(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var err error

	// simple data model arguments
	// 0=accidentId  1=claimantPolicyId  2=defendantPolicyId  3=repairQuoteId
	// 1534180781    USA-AX203-3459802   USA-AS204-1042919    1000000001

	if len(args) != 4 {
		return shim.Error("Incorrect number of arguments. Expecting exactly 4")
	}

	// === Check input variables ===
	if len(args[0]) <= 0 {
		return shim.Error("1st argument must be a non-empty string")
	}
	if len(args[1]) <= 0 {
		return shim.Error("2nd argument must be a non-empty string")
	}
	if len(args[2]) <= 0 {
		return shim.Error("3rd argument must be a non-empty string")
	}
	if len(args[3]) <= 0 {
		return shim.Error("4th argument must be a non-empty string")
	}

	accidentID := args[0]
	claimantPolicyID := args[1]
	defendantPolicyID := args[2]
	repairQuoteID := args[3]

	// === Check if AccidentReport asset exists
	accidentRef := fmt.Sprintf("%s#%s", "accident.AccidentReport", accidentID)
	reportAsBytes, err := stub.GetState(accidentRef)
	if err != nil {
		return shim.Error("Failed to get accident report: " + err.Error())
	} else if reportAsBytes == nil {
		return shim.Error("This accident report doesn't exists: " + accidentRef)
	}

	// === Unmarshal the report to an object
	accidentReport := AccidentReport{}
	if err = json.Unmarshal(reportAsBytes, accidentReport); err != nil {
		return shim.Error("Failed to unmarshal accident report: " + err.Error())
	}

	// === Check if InsurancePolicy asset of claimant exists
	claimantRef := fmt.Sprintf("%s#%s", "insurance.InsurancePolicy", claimantPolicyID)
	claimantAsBytes, err := stub.GetState(claimantRef)
	if err != nil {
		return shim.Error("Failed to get insurance policy: " + err.Error())
	} else if claimantAsBytes == nil {
		return shim.Error("Insurance policy of claimant doesn't exists: " + claimantRef)
	}

	// === Unmarshal the policy to an object
	claimantPolicy := InsurancePolicy{}
	if err = json.Unmarshal(claimantAsBytes, claimantPolicy); err != nil {
		return shim.Error("Failed to unmarshal insurance policy of claimaint: " + err.Error())
	}

	// === Check if InsurancePolicy asset of defantdant exists
	defendantRef := fmt.Sprintf("%s#%s", "insurance.InsurancePolicy", defendantPolicyID)
	defendantAsBytes, err := stub.GetState(defendantRef)
	if err != nil {
		return shim.Error("Failed to get insurance policy: " + err.Error())
	} else if defendantAsBytes == nil {
		return shim.Error("Insurance policy of defendant doesn't exists: " + defendantRef)
	}

	// === Unmarshal the policy to an object
	defendantPolicy := InsurancePolicy{}
	if err = json.Unmarshal(defendantAsBytes, defendantPolicy); err != nil {
		return shim.Error("Failed to unmarshal insurance policy of defendant: " + err.Error())
	}

	// === Check if claimant and defendant are involved in accident
	vehicles := accidentReport.InvolvedGoods.Vehicles

	// save the items in map
	vmap := make(map[string]bool)
	for i := 0; i < len(vehicles); i++ {
		vmap[vehicles[i]] = true
	}

	// Check if registered vehicle of claimant is involved in accident
	vehicleClaimantRef := claimantPolicy.RegisteredVehicle
	if _, ok := vmap[vehicleClaimantRef]; ok {
		return shim.Error("Insured vehicle of claimant is not involved accident: " + vehicleClaimantRef)
	}

	// Check if registered vehicle of defendant is involved in accident
	vehicleDefendantRef := defendantPolicy.RegisteredVehicle
	if _, ok := vmap[vehicleDefendantRef]; ok {
		return shim.Error("Insured vehicle of claimant is not involved accident: " + vehicleDefendantRef)
	}

	// === Check if RepairQuote asset exists
	quoteRef := fmt.Sprintf("%s#%s", "vehiclerepair.RepairQuote", repairQuoteID)
	quoteAsBytes, err := stub.GetState(quoteRef)
	if err != nil {
		return shim.Error("Failed to get repair quote: " + err.Error())
	} else if quoteAsBytes == nil {
		return shim.Error("Given repair quote doesn't exists: " + quoteRef)
	}

	// === Unmarshal the quote to an object
	repairQuote := RepairQuote{}
	if err = json.Unmarshal(quoteAsBytes, repairQuote); err != nil {
		return shim.Error("Failed to unmarshal repair quote: " + err.Error())
	}

	// === Create claim object and marchal to JSON ===
	claimObjClass := "insurance.InsuranceClaim"
	//claimID, err := strconv.ParseInt("1000000001", 10, 64) //static id for testing
	claimID := time.Now().Unix()
	dateOfClaim := time.Now()
	insuranceClaim := &InsuranceClaim{claimObjClass, strconv.FormatInt(claimID, 10), dateOfClaim, "NEW", accidentRef, claimantRef, defendantRef, quoteRef}
	claimJSONasBytes, err := json.Marshal(insuranceClaim)
	if err != nil {
		return shim.Error(err.Error())
	}

	// === Save insurance claim to state ===
	claimRef := fmt.Sprintf("%s#%d", claimObjClass, claimID)
	err = stub.PutState(claimRef, claimJSONasBytes)
	if err != nil {
		return shim.Error(err.Error())
	}

	// === Emit NewQuoteOffer event
	newClaim := &NewClaimEvent{strconv.FormatInt(claimID, 10), claimantPolicyID, defendantPolicyID, repairQuote.Total}
	eventJSONasBytes, err := json.Marshal(newClaim)
	if err != nil {
		return shim.Error(err.Error())
	}
	stub.SetEvent("NewClaimEvent", eventJSONasBytes)

	fmt.Println("- Insurance claim successfully send to defendant")
	return shim.Success(eventJSONasBytes)
}

// readAssetData - Get a accident report from chaincode state
func (t *InsuranceChaincode) readAssetData(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var assetClass, assetID, assetRef, assetType, jsonResp string
	var err error

	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments. Expecting Class and ID/NAme of asset to query")
	}

	assetClass = args[0]
	assetType = strings.SplitAfter(assetClass, ".")[1]
	assetID = args[1]

	assetRef = fmt.Sprintf("%s#%s", assetClass, assetID)
	valAsbytes, err := stub.GetState(assetRef) //get the asset from chaincode state
	if err != nil {
		jsonResp = "{\"Error\":\"Failed to get state for " + assetType + " with id" + assetID + "\"}"
		return shim.Error(jsonResp)
	} else if valAsbytes == nil {
		jsonResp = "{\"Error\":\"" + assetType + " does not exist: " + assetID + "\"}"
		return shim.Error(jsonResp)
	}

	return shim.Success(valAsbytes)
}

// setupAssets - Create all example assets
func (t *InsuranceChaincode) setupAssets(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	// Setup has no arguments, just run the transactions

	var err error
	var address AddressConcept
	var location LocationConcept
	addressObjClass := "base.Address"
	locationObjClass := "accident.Location"
	registrantObjClass := "base.Registrant"
	ersObjClass := "base.EmergencyServices"
	insurerObjClass := "base.Insurer"
	shopObjClass := "base.RepairShop"
	vehicleObjClass := "base.Vehicle"
	policyObjClass := "insurance.InsurancePolicy"

	// === Create registrant - AutoLease
	address = AddressConcept{addressObjClass, "4300 Broadway", "New York, NY 10033", "United States"}
	registrantAl := &Registrant{registrantObjClass, "908123764", "LEASER", "AutoLease", "", address}
	regAlJSONasBytes, err := json.Marshal(registrantAl)
	if err != nil {
		return shim.Error(err.Error())
	}

	// === Create registrant - John Smith
	address = AddressConcept{addressObjClass, "28 Clinton Ave", "Jersey City, NJ 07304", "United States"}
	registrantJS := &Registrant{registrantObjClass, "170632064", "INDIVIDUAL", "Smith", "J.", address}
	regJsJSONasBytes, err := json.Marshal(registrantJS)
	if err != nil {
		return shim.Error(err.Error())
	}

	// === Create emergency services - NYPD 34th Precinct
	address = AddressConcept{addressObjClass, "4295 Broadway", "New York, NY 10033", "United States"}
	location = LocationConcept{locationObjClass, 40.851498, -73.935389, "Police Station"}
	ersNYPD := &EmergencyServices{ersObjClass, CompanyAbstract{"NYPD 34th Precinct", address}, location}
	ersNYPDJSONasBytes, err := json.Marshal(ersNYPD)
	if err != nil {
		return shim.Error(err.Error())
	}

	// === Create insurer AllSecur Insurance
	address = AddressConcept{addressObjClass, "59 Maiden Ln", "New York, NY 10038", "United States"}
	signatureAS := "iVBORw0KGgoAAAANSUhEUgAAAFQAAAAtCAMAAAAHmLz6AAAAQlBMVEX29vYuLi41NTWFhYU5OTlQUFAXFxe8vLxra2uhoaGKiopfX19zc3NFRUUeHh7X19e4uLh7e3uZmZlcXFw9PT0AAABLGcruAAAAAXRSTlMAQObYZgAAAZ5JREFUSIntVmtvgzAMZOu2vlb8yv//rTvbEFIVtRWwb5x0LUqTi53YR7tux463wMDljA/aTFKVgi65kSwJBf2Rk6vBkuw2VGVLDtCBq0CWrKKUXIUycISfrawUNUoGiPn8pXK0p0teQiXpiBotYH9crFf8MEvS4TmTsVOer3wWI0ShERw0GQfhXHr7HEFJFwxRPGkQY8s0Ud89guySjpO47OLEXVO93ptWUjmIyrqih5JwDZThJazX8yrJuIiSgZ5U9ajuUF1NXe3dkFUCYxeikAzuebu4keaMemUa+pj7wgZp3JtYDJkjTLrB8KxdxTHJNKzwk/FVmqX0EH4zgDhxgqa/Vzfndl8292rDhTF/EES8vaj3ujN6fCc0A4OM3HU95anCrYYz8GRiS5I+zm7GYusd1ByksSeqtWSDQYV09LBGMjZn26Mo1x/Vi7/78bHJ6aAl03xER1l3OmtbuYyb/TBbWL/BqdU12x7TQrTEApks50FVVe9uz1+dPO7DKH+UXDzH8SmuiUuEgs6dTX4BuI1gu78CO3bs2PHf+AMMxRtHDqOPeQAAAABJRU5ErkJggg=="
	insurerAS := &Insurer{insurerObjClass, CompanyAbstract{"Allsecur Insurance", address}, signatureAS}
	insAsJSONasBytes, err := json.Marshal(insurerAS)
	if err != nil {
		return shim.Error(err.Error())
	}

	// === Create insurer AXA Insurance
	address = AddressConcept{addressObjClass, "888 Bergen Ave", "Jersey City, NJ 07306", "United States"}
	signatureAX := "iVBORw0KGgoAAAANSUhEUgAAAKUAAAAxBAMAAABJ8nS8AAAAKlBMVEX29vYTExMXFxdMTExfX1+hoaE9PT1cXFy4uLiZmZkeHh57e3vX19cAAAC4vRUVAAAAAXRSTlMAQObYZgAAAmFJREFUWIXtVE1v2kAQndN6bS+/oP+o1/4BTlWrgMQptBVIPtFGFMnKITQpSL60EBoknxoaBWlPSZM6kk9EKOx/6eyuP4hlq9Tk1PIkZN7O+u28mfEC7PC/gxeS0rC9IlIe87CIlIY5DgtIefQGjQJSHnvMzyPXW0iaHnXzyKstNC3HdHII9XN3b4ZzMDh7g3+66wTsLSTNFRhw8gGANWKiMjW30LRcIEc1i6NYTNR6UmNj9NeaTRzzJdYRLhOikJRzLnjhy++cvFW6umnfY5K+7ElE1HoiNBaPukXb4+R/W3h5khOx+OjJGbpKiUTUIsOH8GucjC3P6Qcv4wgFWsXVMTwGq39TU0OXtZRIHESZcEi/1YpMogF3ccSXblirmU0UpWzZDmUwJggcAlYF40HPvj7oDhPGTh5FkS6X38cVP89ImvgukZaEs0ZQD1XsEJorObIAHjA87xid/gDq6QjVVRrTWkYThweItCTOUhKVzgqNDkr3+AENgVSBDVDaA+LrCFGO6GndzWj28fdeVryzSAnoFvXDwS1mPPzESQN6LlyjD9uHIdeRitpIRLZFgG7gM8iWXDgJQfySRVzyt0CnArvC2ZSzGraaoKMogiMtjZ5mJRVmsDbhM/1Q0g1YARnMXdiHQECHo4rFng+drisjdgu3XAgnVxM7k9520Y18rB8rqLT2ObkPYBp8kW4rL86shasiTLg/26Pby1xNR01ISiQC/fju2hN93Mlr9Nlih88gjsBNfTHD2ozyM1XX3J/BJnsF7+ehtfnWTWFuc6cXAG/Pp9d8eskddthhh38WvwHtUtriRXlodwAAAABJRU5ErkJggg=="
	insurerAX := &Insurer{insurerObjClass, CompanyAbstract{"AXA Insurance", address}, signatureAX}
	insAxJSONasBytes, err := json.Marshal(insurerAX)
	if err != nil {
		return shim.Error(err.Error())
	}

	// === Create repairshop USA Automotice New York City
	address = AddressConcept{addressObjClass, "225 Delancey St", "New York, NY 10002", "United States"}
	shopNYC := &RepairShop{shopObjClass, CompanyAbstract{"USA Automotive NYC", address}, "", "nyc@usa-automotive.com"}
	shopNycJSONasBytes, err := json.Marshal(shopNYC)
	if err != nil {
		return shim.Error(err.Error())
	}

	// === Create repairshop USA Automotice Jersey City
	address = AddressConcept{addressObjClass, "5 West Side Ave", "Jersey City, NJ 07305", "United States"}
	shopJC := &RepairShop{shopObjClass, CompanyAbstract{"USA Automotive JC", address}, "", "jersey@usa-automotive.com"}
	shopJcJSONasBytes, err := json.Marshal(shopJC)
	if err != nil {
		return shim.Error(err.Error())
	}

	// === Create vehicle JN6ND01S3GX194659
	dateFirstAdmissionOne, err := time.Parse(time.RFC3339, "2018-01-12T00:00:00Z")
	dateAscriptionOne, err := time.Parse(time.RFC3339, "2018-01-13T00:00:00Z")
	vehicleOne := &Vehicle{vehicleObjClass, "JN6ND01S3GX194659", "WPD 9321", dateFirstAdmissionOne, dateAscriptionOne, "base.Registrant#9081237645", "BMW", "X5 Estate 3.0i", "Black", 2595, 5}
	vOneJSONasBytes, err := json.Marshal(vehicleOne)
	if err != nil {
		return shim.Error(err.Error())
	}

	// === Create vehicle 1HTZR0007JH586991
	dateFirstAdmissionTwo, err := time.Parse(time.RFC3339, "2014-09-28T00:00:00Z")
	dateAscriptionTwo, err := time.Parse(time.RFC3339, "2018-10-01T00:00:00Z")
	vehicleTwo := &Vehicle{vehicleObjClass, "1HTZR0007JH586991", "B63-AGM", dateFirstAdmissionTwo, dateAscriptionTwo, "base.Registrant#170632064", "Toyota", "Prius", "Red", 1526, 4}
	vTwoJSONasBytes, err := json.Marshal(vehicleTwo)
	if err != nil {
		return shim.Error(err.Error())
	}

	// === Create insurance policy USA-AS204-1042919
	dateValidFrom, err := time.Parse(time.RFC3339, "2018-05-01T00:00:00Z")
	dateValidTo, err := time.Parse(time.RFC3339, "2020-04-30T00:00:00Z")
	insurancePolicy := &InsurancePolicy{policyObjClass, "USA-AS204-1042919", "State of New Jersey", dateValidFrom, dateValidTo, "base.Vehicle#1HTZR0007JH586991", "USA", "AS204", 1042919, "AF", "Toyota", []string{"US", "CA"}, "base.Registrant#170632064", "base.Insurer#AXA Insurance"}
	pOneJSONasBytes, err := json.Marshal(insurancePolicy)
	if err != nil {
		return shim.Error(err.Error())
	}

	// ============================================================================================================================
	// Store test data in blockchain
	// ============================================================================================================================
	var stateRef string
	var assetList []AssetEntry
	var assetEntry AssetEntry

	// Registrant - AutoLease
	stateRef = fmt.Sprintf("%s#%s", registrantObjClass, registrantAl.IdentificationNumber)
	err = stub.PutState(stateRef, regAlJSONasBytes)
	if err != nil {
		return shim.Error(err.Error())
	}
	assetEntry = AssetEntry{registrantObjClass, registrantAl.IdentificationNumber}
	assetList = append(assetList, assetEntry)

	// Registrant - John Smith
	stateRef = fmt.Sprintf("%s#%s", registrantObjClass, registrantJS.IdentificationNumber)
	err = stub.PutState(stateRef, regJsJSONasBytes)
	if err != nil {
		return shim.Error(err.Error())
	}
	assetEntry = AssetEntry{registrantObjClass, registrantJS.IdentificationNumber}
	assetList = append(assetList, assetEntry)

	// Emergency Services - NYPD
	stateRef = fmt.Sprintf("%s#%s", ersObjClass, ersNYPD.TradeName)
	err = stub.PutState(stateRef, ersNYPDJSONasBytes)
	if err != nil {
		return shim.Error(err.Error())
	}
	assetEntry = AssetEntry{ersObjClass, ersNYPD.TradeName}
	assetList = append(assetList, assetEntry)

	// Insurer - AllSecur Insurance
	stateRef = fmt.Sprintf("%s#%s", insurerObjClass, insurerAS.TradeName)
	err = stub.PutState(stateRef, insAsJSONasBytes)
	if err != nil {
		return shim.Error(err.Error())
	}
	assetEntry = AssetEntry{insurerObjClass, insurerAS.TradeName}
	assetList = append(assetList, assetEntry)

	// Insurer - AXA Insurance
	stateRef = fmt.Sprintf("%s#%s", insurerObjClass, insurerAX.TradeName)
	err = stub.PutState(stateRef, insAxJSONasBytes)
	if err != nil {
		return shim.Error(err.Error())
	}
	assetEntry = AssetEntry{insurerObjClass, insurerAX.TradeName}
	assetList = append(assetList, assetEntry)

	// RepairShop - USA Automotive NYC
	stateRef = fmt.Sprintf("%s#%s", shopObjClass, shopNYC.TradeName)
	err = stub.PutState(stateRef, shopNycJSONasBytes)
	if err != nil {
		return shim.Error(err.Error())
	}
	assetEntry = AssetEntry{shopObjClass, shopNYC.TradeName}
	assetList = append(assetList, assetEntry)

	// RepairShop - USA Automotive NYC
	stateRef = fmt.Sprintf("%s#%s", shopObjClass, shopJC.TradeName)
	err = stub.PutState(stateRef, shopJcJSONasBytes)
	if err != nil {
		return shim.Error(err.Error())
	}
	assetEntry = AssetEntry{shopObjClass, shopJC.TradeName}
	assetList = append(assetList, assetEntry)

	// Vehicle - JN6ND01S3GX194659
	stateRef = fmt.Sprintf("%s#%s", vehicleObjClass, vehicleOne.RegistrationNumber)
	err = stub.PutState(stateRef, vOneJSONasBytes)
	if err != nil {
		return shim.Error(err.Error())
	}
	assetEntry = AssetEntry{vehicleObjClass, vehicleOne.RegistrationNumber}
	assetList = append(assetList, assetEntry)

	// RepairShop - 1HTZR0007JH586991
	stateRef = fmt.Sprintf("%s#%s", vehicleObjClass, vehicleTwo.RegistrationNumber)
	err = stub.PutState(stateRef, vTwoJSONasBytes)
	if err != nil {
		return shim.Error(err.Error())
	}
	assetEntry = AssetEntry{vehicleObjClass, vehicleTwo.RegistrationNumber}
	assetList = append(assetList, assetEntry)

	// Insurance Policy - USA-AS204-1042919
	stateRef = fmt.Sprintf("%s#%s", policyObjClass, insurancePolicy.PolicyID)
	err = stub.PutState(stateRef, pOneJSONasBytes)
	if err != nil {
		return shim.Error(err.Error())
	}
	assetEntry = AssetEntry{policyObjClass, insurancePolicy.PolicyID}
	assetList = append(assetList, assetEntry)

	// Marshal AssetList
	assetsJSONasBytes, err := json.Marshal(assetList)
	if err != nil {
		return shim.Error(err.Error())
	}

	fmt.Println("- Setup created example assets")
	return shim.Success(assetsJSONasBytes)
}
