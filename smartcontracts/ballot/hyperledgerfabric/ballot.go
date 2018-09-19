package main

// imports needed for chaincode
import (
  "encoding/json"
  "fmt"
  "github.com/hyperledger/fabric/core/chaincode/shim"
  pb "github.com/hyperledger/fabric/protos/peer"
)

// Ballot implements a voting chaincode to manage event proposals
type Ballot struct {}

// Voter declares a new complex type which will be used
// for variables later. It will represent a single voter.
type Voter struct {
  VID    string `json:"vid`     // registration id of a voter
  Weight uint   `json:"weight"` // weight is accumulated
  Voted  bool   `json:"voted"`  // if true, person voted
  Vote   string `json:"vote"`   // name of the voted proposal
}

// Proposal is a type for a single proposal.
type Proposal struct {
  Name      string `json:"proposalName"` //short name
  VoteCount uint   `json:"voteCount"`    //number of votes
}

// Main function to start chaincode
func main() {
  err := shim.Start(new(Ballot))
  if err != nil {
    fmt.Printf("Error starting Ballot chaincode: %s", err)
  }
}

// Init initializes chaincode
func (t *Ballot) Init(stub shim.ChaincodeStubInterface) pb.Response {
  // Get the args from the transaction proposal
  // In our case it is the registration id of the chairman
  args := stub.GetStringArgs()
  if len(args) != 1 {
    return shim.Error("Incorrect arguments.
           Expecting registration id of chairman")
  }

  // Set up any variables by calling stub.PutState()
  // We store the chairman value on the ledger as a byte array
  err := stub.PutState("chairman", []byte(args[0]))
  if err != nil {
    return shim.Error(fmt.Sprintf("Failed to assign chairman:
                      %s", args[0]))
  }

  return shim.Success(nil)
}

// Invoke - Our entry point for Invocations
func (t *Ballot) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
  function, args := stub.GetFunctionAndParameters()
  fmt.Println("invoke is running " + function)
  // Handle different functions
  if function == "initProposals" {
    return t.initProposals(stub, args)
  }
  else if function == "giveRightToVote" {
    return t.giveRightToVote(stub, args)
  } else if function == "vote" { //receive vote from individual
    return t.vote(stub, args)
  }

  // Return error when unknown function is received
  fmt.Println("invoke did not find func: " + function)
  return shim.Error("Received unknown function invocation")
}

// createProposals - create for each proposal a new entry in state
func (t *Ballot) initProposals(stub shim.ChaincodeStubInterface, args []string) peer.Response {
  chairmanAsBytes, err := stub.GetState("chairman")
  if chairmanAsBytes != nil {
    chairman := string(chairmanAsBytes[:])
    if chairman != args[0] {
      return shim.Error("Invoker is not the chairman")
    }
  } else {
    return shim.Error("No chairman found: " + err.Error())
  }

  for i := 1; i < len(args); i++ {
    // Create proposal object and marshal to JSON
    var proposalName = args[i]
    proposal := &Proposal{proposalName, 0}
    proposalJSONasBytes, err := json.Marshal(proposal)
    if err != nil { return shim.Error(err.Error()) }

    // Check if proposal doesn't exists
    proposalAsBytes, err := stub.GetState(proposalName)
    if proposalAsBytes == nil {
      // Save proposal to state
      err = stub.PutState(proposalName, proposalJSONasBytes)
      if err != nil { return shim.Error(err.Error()) }
    }
  }

  // Proposals saved
  fmt.Println("- end init proposals")
  return shim.Success(nil)
}

// giveRightToVote - give the right to vote to a specific id
func (t *Ballot) giveRightToVote(stub shim.ChaincodeStubInterface, args []string) peer.Response {
  // Same code as above to only allow chairman to invoke - skipped
  // Retrieve voter from state to check if not yet voted - skipped
  // Create voter based on arguments
  // Args only include registration id from individual
  vid := args[0]
  voter := &Voter{vid, 1, false, ""}
  voterJSONasBytes, err := json.Marshal(voter)
  if err != nil {
      return shim.Error(err.Error())
  }

  // Save voter to state
  err = stub.PutState(vid, voterJSONasBytes)
  if err != nil {
    return shim.Error(err.Error())
  }

  return shim.Success(nil)
}

// vote - receive and store vote from individual
func (t *Ballot) vote(stub shim.ChaincodeStubInterface, args []string) peer.Response {
  // Check if voter exists based on arguments
  // Args include reg id, and the name of proposal voted for
  vid := args[0]
  voteProposal := args[1]
  voterAsBytes, err := stub.GetState(vid)
  if err != nil {
    return shim.Error("Failed to get voter:" + err.Error())
  } else if voterAsBytes == nil {
    return shim.Error("Voter does not exist")
  }

  // Retrieve the proposal
  proposalAsBytes, err := stub.GetState(voteProposal)
  if err != nil {
    return shim.Error("Failed to get proposal:" + err.Error())
  } else if proposalAsBytes == nil {
    return shim.Error("Proposal does not exist")
  }

  // Placeholder for stored proposal object
  proposalObj := Proposal{}
  err = json.Unmarshal(proposalAsBytes, &proposalObj) //unmarshal
  if err != nil {
    return shim.Error(err.Error())
  }

  // Placeholder for stored proposal object
  voterObj := Voter{}
  err = json.Unmarshal(voterAsBytes, &voterObj) //unmarshal it
  if err != nil {
    return shim.Error(err.Error())
  }

  proposalObj.VoteCount += voterObj.Weight // Update vote count
  voterObj.Voted = true // Update voter voted status

  // Store updated proposal object
  proposalAsBytes, err = json.Marshal(proposalObj)
  err = stub.PutState(voteProposal, proposalAsBytes) //rewrite
  if err != nil {
    return shim.Error(err.Error())
  }

  // Store updated voter object -- skipped
  return shim.Success(nil)
}