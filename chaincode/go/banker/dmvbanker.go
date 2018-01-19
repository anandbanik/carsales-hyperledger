package main

import (
	"fmt"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
	"strings"
	"encoding/pem"
	"crypto/x509"
	"encoding/json"
	"strconv"
	//"bytes"
)

var logger = shim.NewLogger("DmvDealerChaincode")

type DmvDealerChaincode struct {
}

type Loan struct {
	LoanId string `json:"loanId"`
	VinNumber string `json:"vin"`
	Amount int `json:"amount"`
	SsnNumber string `json:"ssnNumber"`
	LoanPeriod int `json:"loanPeriod"`
	Apr int `json:"apr"`
	MonthlyPayment int `json:"monthlyPayment"`
	Status string `json:"status"`
	Org string `json:"org"`
}

func (t *DmvDealerChaincode) Init(stub shim.ChaincodeStubInterface) pb.Response {
	logger.Debug("Init")
	return shim.Success(nil)
}

func (t *DmvDealerChaincode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	logger.Debug("Invoke")
	function, args := stub.GetFunctionAndParameters()
	if function == "loan" {
		return t.loan(stub, args)
	}else if function == "query" {
		return t.query(stub, args)
	}

	return pb.Response{Status:403, Message:"unknown function name"}
}
func (t *DmvDealerChaincode) loan(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	if len(args) != 6 {
		return pb.Response{Status:403, Message:"incorrect number of arguments"}
	}

	creatorBytes, err := stub.GetCreator()
	if err != nil {
		return shim.Error("cannot get creator")
	}

	ssn, org := getCreator(creatorBytes)
	if org != "" {
		logger.Debug("Org is null")
	}

	amount, err := strconv.Atoi(args[2])
	if(err != nil){
		return shim.Error("Invalid amount, expecting a integer value")
	}

	loanPeriod, err := strconv.Atoi(args[3])
	if(err != nil){
		return shim.Error("Invalid Loan amount, expecting a integer value")
	}

	interest, err := strconv.Atoi(args[4])
	if(err != nil){
		return shim.Error("Invalid APR amount, expecting a integer value")
	}

	monthlyPayment, err := strconv.Atoi(args[5])
	if(err != nil){
		return shim.Error("Invalid Monthly Payment amount, expecting a integer value")
	}

	loanObj := &Loan{
				LoanId: args[0],
				VinNumber: args[1],
				Amount: amount,
				SsnNumber: ssn,
				LoanPeriod: loanPeriod,
				Apr: interest,
				MonthlyPayment: monthlyPayment,
				Org: org}

	jsonLoanObj, err := json.Marshal(loanObj)	
	if(err != nil){
		return shim.Error("Cannot create Json Object")
	}

	logger.Debug("Json Obj: "+string(jsonLoanObj))

	key := ssn+"@"+args[1]

	err = stub.PutState(key, jsonLoanObj)

	if err != nil {
		return shim.Error("cannot put state")
	}

	return shim.Success(nil)
}

func (t *DmvDealerChaincode) query(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 1 {
		return pb.Response{Status:403, Message:"incorrect number of arguments"}
	}

	if args[0] == "health" {
		logger.Info("Health status Ok")
		return shim.Success(nil)
	}

	creatorBytes, err := stub.GetCreator()
	if err != nil {
		return shim.Error("cannot get creator")
	}

	ssn, org := getCreator(creatorBytes)

	if org != ""{
		logger.Debug("Org is null")
	}
	
	key := ssn+"@"+args[0]

	bytes, err := stub.GetState(key)
	if err != nil {
		return shim.Error("cannot get state")
	}

	return shim.Success(bytes)
}


var getCreator = func (certificate []byte) (string, string) {
	data := certificate[strings.Index(string(certificate), "-----"): strings.LastIndex(string(certificate), "-----")+5]
	block, _ := pem.Decode([]byte(data))
	cert, _ := x509.ParseCertificate(block.Bytes)
	organization := cert.Issuer.Organization[0]
	commonName := cert.Subject.CommonName
	logger.Debug("commonName: " + commonName + ", organization: " + organization)


	organizationShort := strings.Split(organization, ".")[0]

	return commonName, organizationShort
}

func main() {
	err := shim.Start(new(DmvDealerChaincode))
	if err != nil {
		fmt.Printf("Error starting chaincode: %s", err)
	}
}