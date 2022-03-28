/*
Copyright IBM Corp. All Rights Reserved.
SPDX-License-Identifier: Apache-2.0
*/

package simple

import (
	"encoding/json"
	"fmt"
	"github.com/hyperledger/fabric-chaincode-go/shim"
	pb "github.com/hyperledger/fabric-protos-go/peer"
	"github.com/hyperledger/fabric/integration/chaincode/simple/paillerCrypto"
	"math/big"
	"os"
	"strconv"
)

type Patient struct {
	PatientName         string      `json:"patientName"`
	PatientNationalID   string      `json:"patientNationalID"`
	PatientFamilyID     string      `json:"patientFamilyID"`
	PatientDiseaseTable [3]*big.Int `json:"patientDiseaseTable"`
}

type PaillerKey struct {
	PatientFamilyID string              `json:"patientFamilyID"`
	Key             *Pailler.PrivateKey `json:"key"`
}

type PatientResult struct {
	Key    string `json:"Key"`
	Record *Patient
}

type PaillerResult struct {
	Key    string `json:"Key"`
	Record *PaillerKey
}

func (t *Patient) Init(stub shim.ChaincodeStubInterface) pb.Response {

	fmt.Println("Init invoked...")

	patient := Patient{PatientName: "Sevde", PatientNationalID: "30", PatientFamilyID: "20", PatientDiseaseTable: [3]*big.Int{}}

	fmt.Println("Ledger Created...")

	publicKey, privateKey, _ := Pailler.GenerateKeyPair(1024)
	N, g := publicKey.ToString()
	publicKey2, err := Pailler.NewPublicKey(N, g)

	fmt.Println("Keys Generated...")

	paillerAsset := PaillerKey{
		PatientFamilyID: patient.PatientFamilyID,
		Key:             privateKey,
	}

	fmt.Println("Pailler Props Generated...")

	for index := range patient.PatientDiseaseTable {
		patient.PatientDiseaseTable[index], _ = publicKey2.Encrypt(0)
	}

	fmt.Println("Encryption Done...")

	patientJSON, err := json.Marshal(patient)
	if err != nil {
		return shim.Error("Json Mars")
	}
	err = stub.PutState(patient.PatientNationalID, patientJSON)
	if err != nil {
		return shim.Error("Cannot put Patient to the ledger")
	}

	paillerJSON, err := json.Marshal(paillerAsset)
	if err != nil {
		return shim.Error("Json Mars")
	}
	err = stub.PutState(paillerAsset.PatientFamilyID, paillerJSON)
	if err != nil {
		return shim.Error("Cannot put PaillerProps to the ledger")
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
	case "changeDisease":
		return t.changeDisease(stub, args)
	case "readAllPatients":
		return t.readAllPatients(stub)
	case "readAllPailler":
		return t.readAllPailler(stub)
	case "addPatient":
		return t.addPatient(stub, args)
	case "deletePatient":
		return t.deletePatient(stub, args)
	case "queryPatient":
		return t.queryPatient(stub, args)
	default:
		return shim.Error(`Invalid invoke function name. Expecting "invoke", "delete", "query", "respond", "mspid", or "event"`)
	}
}

// Add a patient to the state
func (t *Patient) addPatient(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 6 {
		return shim.Error("Incorrect number of arguments. Expecting 6")
	}

	var paillerAsset PaillerKey
	patientNationalID := args[1]
	patientFamilyID := args[2]

	val, err := stub.GetState(patientNationalID)
	if err != nil {
		return shim.Error("Patient Already Exist : " + string(val))
	}

	asset, err := stub.GetState(patientFamilyID)
	if err != nil {
		return shim.Error("Error From Pailler Ledger")
	}

	if len(asset) == 0 {
		fmt.Println("Family Tree Doesn't Exist")
		_, privateKey, _ := Pailler.GenerateKeyPair(1024)
		fmt.Println("Keys Generated...")
		paillerAsset = PaillerKey{
			PatientFamilyID: patientFamilyID,
			Key:             privateKey,
		}
		fmt.Println("Pailler Props Generated...")
	} else {
		err = json.Unmarshal(asset, &paillerAsset)
	}

	fmt.Println(paillerAsset.PatientFamilyID)
	fmt.Println("Fetched Pailler Asset")

	firstDisease, err := strconv.Atoi(args[3])
	if err != nil {
		return shim.Error("Atoi error")
	}
	secondDisease, err := strconv.Atoi(args[4])
	thirdDisease, err := strconv.Atoi(args[5])

	firstDiseaseEncrypted, err := paillerAsset.Key.Pk.Encrypt(int64(firstDisease))
	if err != nil {
		return shim.Error("Encryption error")
	}
	secondDiseaseEncrypted, err := paillerAsset.Key.Pk.Encrypt(int64(secondDisease))
	thirdDiseaseEncrypted, err := paillerAsset.Key.Pk.Encrypt(int64(thirdDisease))

	fmt.Println("Encryption Done...")

	patient :=
		Patient{
			PatientName:         args[0],
			PatientNationalID:   patientNationalID,
			PatientFamilyID:     patientFamilyID,
			PatientDiseaseTable: [3]*big.Int{firstDiseaseEncrypted, secondDiseaseEncrypted, thirdDiseaseEncrypted},
		}

	patientJSON, err := json.Marshal(patient)
	if err != nil {
		return shim.Error("Asset cannot encoded right now...")
	}

	fmt.Println(patient)

	err = stub.PutState(patient.PatientNationalID, patientJSON)
	if err != nil {
		return shim.Error("Failed in put state...")
	}

	paillerJSON, err := json.Marshal(paillerAsset)
	if err != nil {
		return shim.Error("Asset cannot encoded right now...")
	}

	fmt.Println(paillerAsset)

	err = stub.PutState(paillerAsset.PatientFamilyID, paillerJSON)
	if err != nil {
		return shim.Error("Failed in put state...")
	}

	fmt.Println("Patient Successfully Saved...")

	return shim.Success(nil)
}

// Delete a patient from state
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

// Read all patients that in the state
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

		queryResult := PatientResult{Key: queryResponse.Key, Record: patient}
		fmt.Println("*************************************************")
		fmt.Println("Key : " + string(queryResult.Key) + " |  Name : " + (queryResult.Record.PatientName))
		fmt.Print("[ ")
		for _, value := range queryResult.Record.PatientDiseaseTable {
			fmt.Println(value)
		}
		fmt.Print("]")
		fmt.Println("*************************************************")
	}
	return shim.Success(nil)
}

// Read all pailler props that in the state
func (t *Patient) readAllPailler(stub shim.ChaincodeStubInterface) pb.Response {

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

		pailler := new(PaillerKey)
		_ = json.Unmarshal(queryResponse.Value, pailler)

		queryResult := PaillerResult{Key: queryResponse.Key, Record: pailler}
		fmt.Println("*************************************************")
		fmt.Println(queryResult.Key)
		fmt.Println(queryResult.Record)
		fmt.Println("*************************************************")
	}
	return shim.Success(nil)
}

// Query callback representing the query of a chaincode
func (t *Patient) queryPatient(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	patient := new(Patient)
	patientProps := new(PaillerKey)
	nationalID := args[0]

	patientAsset, err := stub.GetState(nationalID)
	if err != nil {
		return shim.Error("Patient doesn't exist")
	}

	err = json.Unmarshal(patientAsset, patient)
	if err != nil {
		return shim.Error("Patient can't be fetched")
	}

	fmt.Println("Patient's Name : " + patient.PatientName)

	paillerAsset, err := stub.GetState(patient.PatientFamilyID)
	if err != nil {
		return shim.Error("Pailler Props can't be fetched")
	}

	err = json.Unmarshal(paillerAsset, patientProps)
	if err != nil {
		return shim.Error("Patient can't be fetched")
	}

	fmt.Println("Patient's FamilyID : " + patientProps.PatientFamilyID)
	fmt.Println("Patient's Diseses Values")

	for _, encryptedValue := range patient.PatientDiseaseTable {
		fmt.Println(encryptedValue)
		decryptedValue, err := patientProps.Key.Decrypt(encryptedValue)
		if err != nil {
			return shim.Error("Patient Disease Value cannot decrypted")
		}
		fmt.Println(decryptedValue)
	}

	return shim.Success(nil)
}

// Change the value of the disease for patient in the state
func (t *Patient) changeDisease(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments. Expecting 2")
	}

	patient := new(Patient)
	patientProps := new(PaillerKey)
	patientNationalID := args[0]

	diseaseIndex, err := strconv.Atoi(args[1])
	if err != nil {
		return shim.Error("Second argument is not an integer")
	}

	patientAsset, err := stub.GetState(patientNationalID)
	if err != nil {
		return shim.Error("Patient doesn't exist")
	}
	err = json.Unmarshal(patientAsset, patient)
	if err != nil {
		return shim.Error("Patient can't be fetched")
	}

	paillerAsset, err := stub.GetState(patient.PatientFamilyID)
	if err != nil {
		return shim.Error("Pailler Props can't be fetched")
	}
	err = json.Unmarshal(paillerAsset, patientProps)
	if err != nil {
		return shim.Error("Patient can't be fetched")
	}

	patient.PatientDiseaseTable[diseaseIndex], err = patientProps.Key.Pk.Encrypt(1)
	if err != nil {
		return shim.Error("Patient's disease value cannot assigned to encrypted 1")
	}

	return shim.Success(nil)
}

func (t *Patient) calculateDiseaseProbability(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments. Expecting 2")
	}

	patient := new(Patient)
	patientProps := new(PaillerKey)
	patientNationalID := args[0]

	diseaseIndex, err := strconv.Atoi(args[1])
	if err != nil {
		return shim.Error("Second argument is not an integer")
	}

	patientAsset, err := stub.GetState(patientNationalID)
	if err != nil {
		return shim.Error("Patient doesn't exist")
	}
	err = json.Unmarshal(patientAsset, patient)
	if err != nil {
		return shim.Error("Patient can't be fetched")
	}

	paillerAsset, err := stub.GetState(patient.PatientFamilyID)
	if err != nil {
		return shim.Error("Pailler Props can't be fetched")
	}
	err = json.Unmarshal(paillerAsset, patientProps)
	if err != nil {
		return shim.Error("Patient can't be fetched")
	}

	patient.PatientDiseaseTable[diseaseIndex], err = patientProps.Key.Pk.Encrypt(1)
	if err != nil {
		return shim.Error("Patient's disease value cannot assigned to encrypted 1")
	}

	return shim.Success(nil)
}
