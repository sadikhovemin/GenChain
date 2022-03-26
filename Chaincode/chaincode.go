/*
Copyright IBM Corp. All Rights Reserved.
SPDX-License-Identifier: Apache-2.0
*/

package simple

import (
	"encoding/json"
	"fmt"
	"github.com/stefanomozart/paillier"
	"math/big"
	"os"
	"strconv"

	"github.com/hyperledger/fabric-chaincode-go/shim"
	pb "github.com/hyperledger/fabric-protos-go/peer"
)

type Patient struct {
	PatientName         string      `json:"patientName"`
	PatientNationalID   string      `json:"patientNationalID"`
	PatientFamilyID     string      `json:"patientFamilyID"`
	PatientDiseaseTable [3]*big.Int `json:"patientDiseaseTable"`
}

type Diseases struct{}

type PaillerProps struct {
	PatientFamilyID string               `json:"patientFamilyID"`
	PublicKey       *paillier.PublicKey  `json:"publicKey"`
	PrivateKey      *paillier.PrivateKey `json:"privateKey"`
}

type QueryResult struct {
	Key    string `json:"Key"`
	Record *Patient
}

func (t *Patient) Init(stub shim.ChaincodeStubInterface) pb.Response {

	fmt.Println("Init invoked...")

	patientAssets := []Patient{
		{PatientName: "Esad", PatientNationalID: "1", PatientFamilyID: "2", PatientDiseaseTable: [3]*big.Int{}},
		{PatientName: "Emin", PatientNationalID: "3", PatientFamilyID: "2", PatientDiseaseTable: [3]*big.Int{}},
		{PatientName: "Abdullah", PatientNationalID: "5", PatientFamilyID: "6", PatientDiseaseTable: [3]*big.Int{}},
	}

	fmt.Println("Ledger Created...")

	for _, patient := range patientAssets {

		publicKey, privateKey, _ := paillier.GenerateKeyPair(3072)

		paillerAsset := PaillerProps{
			PublicKey:       publicKey,
			PrivateKey:      privateKey,
			PatientFamilyID: patient.PatientFamilyID,
		}

		for _, disease := range patient.PatientDiseaseTable {
			disease, _ = publicKey.Encrypt(disease.Int64())
		}

		patientJSON, err := json.Marshal(patient)
		if err != nil {
			return shim.Error("Json Mars")
		}

		paillerJSON, err := json.Marshal(paillerAsset)
		if err != nil {
			return shim.Error("Json Mars")
		}

		err = stub.PutState(patient.PatientNationalID, patientJSON)
		if err != nil {
			return shim.Error("Cannot put Patient to the ledger")
		}

		err = stub.PutState(paillerAsset.PatientFamilyID, paillerJSON)
		if err != nil {
			return shim.Error("Cannot put PaillerProps to the ledger")
		}

	}
	fmt.Println("Init returning with success")
	return shim.Success(nil)
}

func (t *Patient) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	fmt.Println("ex02 Invoke")
	if os.Getenv("DEVMODE_ENABLED") != "" {
		fmt.Println("invoking in devmode")
	}
	function, args := stub.GetFunctionAndParameters()
	switch function {
	case "invoke":
		// Make payment of X units from A to B
		return t.invoke(stub, args)
	case "readAllPatients":
		// Deletes an entity from its state
		return t.readAllPatients(stub)
	case "addPatient":
		// Deletes an entity from its state
		return t.addPatient(stub, args)
	case "deletePatient":
		// Deletes an entity from its state
		return t.deletePatient(stub, args)
	case "queryPatient":
		// the old "Query" is now implemented in invoke
		return t.queryPatient(stub, args)
	case "respond":
		// return with an error
		return t.respond(stub, args)
	case "mspid":
		// Checks the shim's GetMSPID() API
		return t.mspid(args)
	case "event":
		return t.event(stub, args)
	default:
		return shim.Error(`Invalid invoke function name. Expecting "invoke", "delete", "query", "respond", "mspid", or "event"`)
	}
}

// Add a patient asset to the ledger
func (t *Patient) addPatient(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 6 {
		return shim.Error("Incorrect number of arguments. Expecting 6")
	}

	val, err := stub.GetState("3")
	if err != nil {
		firstDisease, _ := strconv.Atoi(args[3])
		secondDisease, _ := strconv.Atoi(args[4])
		thirdDisease, _ := strconv.Atoi(args[5])
		asset :=
			Patient{
				PatientName:         args[0],
				PatientNationalID:   args[1],
				PatientFamilyID:     args[2],
				PatientDiseaseTable: [3]*big.Int{firstDisease, secondDisease, thirdDisease},
			}

		assetJSON, err := json.Marshal(asset)
		if err != nil {
			return shim.Error("Asset cannot encoded right now...")
		}

		err = stub.PutState(asset.PatientNationalID, assetJSON)
		if err != nil {
			return shim.Error("Failed in put state...")
		}

		return shim.Success(nil)
	}
	return shim.Error("Patient Already Exist : " + string(val))
}

// Deletes an entity from state
func (t *Patient) deletePatient(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	nationalID := args[0]

	// Delete the key from the state in ledger
	err := stub.DelState(nationalID)
	if err != nil {
		return shim.Error("Failed to delete state")
	}

	return shim.Success(nil)
}

// Deletes an entity from state
func (t *Patient) readAllPatients(stub shim.ChaincodeStubInterface) pb.Response {

	startKey := ""
	endKey := ""

	resultsIterator, err := stub.GetStateByRange(startKey, endKey)

	if err != nil {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}
	defer func(resultsIterator shim.StateQueryIteratorInterface) {
		err := resultsIterator.Close()
		if err != nil {
			fmt.Println("Error in iterator...")
		}
	}(resultsIterator)

	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()

		if err != nil {
			return shim.Error("Incorrect number of arguments. Expecting 1")
		}

		patient := new(Patient)
		_ = json.Unmarshal(queryResponse.Value, patient)

		queryResult := QueryResult{Key: queryResponse.Key, Record: patient}
		fmt.Println("*************************************************")
		fmt.Println("Key : " + string(queryResult.Key) + " |  Name : " + (queryResult.Record.PatientName))
		fmt.Print("[ ")
		for _, value := range queryResult.Record.PatientDiseaseTable {
			fmt.Print(strconv.Itoa(value) + " ")
		}
		fmt.Print("]")
		fmt.Println("*************************************************")
	}
	return shim.Success(nil)
}

// Transaction makes payment of X units from A to B
func (t *Patient) invoke(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var A, B string    // Entities
	var Aval, Bval int // Asset holdings
	var X int          // Transaction value
	var err error

	if len(args) != 3 {
		return shim.Error("Incorrect number of arguments. Expecting 3")
	}

	A = args[0]
	B = args[1]

	// Get the state from the ledger
	// TODO: will be nice to have a GetAllState call to ledger
	Avalbytes, err := stub.GetState(A)
	if err != nil {
		return shim.Error("Failed to get state")
	}
	if Avalbytes == nil {
		return shim.Error("Entity not found")
	}
	Aval, _ = strconv.Atoi(string(Avalbytes))

	Bvalbytes, err := stub.GetState(B)
	if err != nil {
		return shim.Error("Failed to get state")
	}
	if Bvalbytes == nil {
		return shim.Error("Entity not found")
	}
	Bval, _ = strconv.Atoi(string(Bvalbytes))

	// Perform the execution
	X, err = strconv.Atoi(args[2])
	if err != nil {
		return shim.Error("Invalid transaction amount, expecting a integer value")
	}
	Aval = Aval - X
	Bval = Bval + X
	fmt.Printf("Aval = %d, Bval = %d\n", Aval, Bval)

	// Write the state back to the ledger
	err = stub.PutState(A, []byte(strconv.Itoa(Aval)))
	if err != nil {
		return shim.Error(err.Error())
	}

	err = stub.PutState(B, []byte(strconv.Itoa(Bval)))
	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success(nil)
}

// query callback representing the query of a chaincode
func (t *Patient) queryPatient(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	nationalID := args[0]

	asset, err := stub.GetState(nationalID)
	if err != nil {
		return shim.Error("Patient doesn't exist")
	}

	fmt.Println(string(asset))
	return shim.Success(nil)
}

// respond simply generates a response payload from the args
func (t *Patient) respond(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 3 {
		return shim.Error("expected three arguments")
	}

	status, err := strconv.ParseInt(args[0], 10, 32)
	if err != nil {
		return shim.Error(err.Error())
	}
	message := args[1]
	payload := []byte(args[2])

	return pb.Response{
		Status:  int32(status),
		Message: message,
		Payload: payload,
	}
}

// mspid simply calls shim.GetMSPID() to verify the mspid was properly passed from the peer
// via the CORE_PEER_LOCALMSPID env var
func (t *Patient) mspid(args []string) pb.Response {
	if len(args) != 0 {
		return shim.Error("expected no arguments")
	}

	// Get the mspid from the env var
	mspid, err := shim.GetMSPID()
	if err != nil {
		jsonResp := "{\"Error\":\"Failed to get mspid\"}"
		return shim.Error(jsonResp)
	}

	if mspid == "" {
		jsonResp := "{\"Error\":\"Empty mspid\"}"
		return shim.Error(jsonResp)
	}

	fmt.Printf("MSPID:%s\n", mspid)
	return shim.Success([]byte(mspid))
}

// event emits a chaincode event
func (t *Patient) event(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments. Expecting 2")
	}

	if err := stub.SetEvent(args[0], []byte(args[1])); err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success(nil)
}
