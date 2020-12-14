## $5 Tech Unlocked 2021!
[Buy and download this Book for only $5 on PacktPub.com](https://www.packtpub.com/product/blockchain-across-oracle/9781788474290)
-----
*If you have read this book, please leave a review on [Amazon.com](https://www.amazon.com/gp/product/1788474295).     Potential readers can then use your unbiased opinion to help them make purchase decisions. Thank you. The $5 campaign         runs from __December 15th 2020__ to __January 13th 2021.__*

# Blockchain across Oracle
This is the code repository for [Blockchain across Oracle](https://www.packtpub.com/big-data-and-business-intelligence/blockchain-across-oracle?utm_source=github&utm_medium=repository&utm_campaign=9781788474290), published by [Packt](https://www.packtpub.com/?utm_source=github). It contains all the supporting project files necessary to work through the book from start to finish.

## About the Book
Blockchain across Oracle is a professional orientation for Oracle developers to get up to speed with the details and implications of the Blockchain across Oracle. Learn everything from Blockchain concepts, through to working with Oracle Blockchain Cloud Service, expert guidance on the effects of the Blockchain on key markets, and the impact to Oracle customers.

## Instructions and Navigation
All of the code is organized into folders.
* smartcontracts/ballot: source code of smart contract, "ballot", used in chapter 8 for both Ethereum as HL Fabric
* smartcontracts/insurancechain: source code of smart contract, "insurancechain", used in chapter 11 to 13
* postman/insurancechain: Postman collection of REST API calls to test the smart contract, "insurancechain".


The code will look like the following:
```
func main() {
  err := shim.Start(new(InsuranceChaincode))
  if err != nil {
    fmt.Printf("Error starting chaincode - %s", err)
  }
}
```

 

## Related Products
* [Blockchain By Example](https://www.packtpub.com/big-data-and-business-intelligence/blockchain-example?utm_source=github&utm_medium=repository&utm_campaign=9781788475686)

* [Mastering Blockchain - Second Edition](https://www.packtpub.com/big-data-and-business-intelligence/mastering-blockchain-second-edition?utm_source=github&utm_medium=repository&utm_campaign=9781788839044)

* [Tokenomics](https://www.packtpub.com/big-data-and-business-intelligence/tokenomics?utm_source=github&utm_medium=repository&utm_campaign=9781789136326)
