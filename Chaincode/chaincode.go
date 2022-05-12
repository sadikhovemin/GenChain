/*
Copyright IBM Corp. All Rights Reserved.
SPDX-License-Identifier: Apache-2.0
*/

package simple

import (
	"encoding/json"
	"fmt"
	"math/big"
	"os"
	"strconv"

	"github.com/hyperledger/fabric-chaincode-go/shim"
	pb "github.com/hyperledger/fabric-protos-go/peer"
	Pailler "github.com/hyperledger/fabric-samples/asset-transfer-basic/chaincode-go/chaincode/paillerCrypto"
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

type Diseases struct {
	SickleCellDisease int `json:"sickleCellDisease"`
	Type2Diabetes     int `json:"type2Diabetes"`
	Achondroplasia    int `json:"achondroplasia"`
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

	var zero = new(big.Int).SetInt64(0)

	patients := []Patient{
		{PatientName: "Erhan", PatientNationalID: "112", PatientFamilyID: "20", PatientDiseaseTable: [3]*big.Int{zero, zero, zero}},
		{PatientName: "Aysegul", PatientNationalID: "113", PatientFamilyID: "20", PatientDiseaseTable: [3]*big.Int{zero, zero, zero}},
		{PatientName: "Ahmet", PatientNationalID: "111", PatientFamilyID: "20", PatientDiseaseTable: [3]*big.Int{zero, zero, zero}},
		{PatientName: "Recep", PatientNationalID: "114", PatientFamilyID: "21", PatientDiseaseTable: [3]*big.Int{zero, zero, zero}},
		{PatientName: "Nusret", PatientNationalID: "115", PatientFamilyID: "22", PatientDiseaseTable: [3]*big.Int{zero, zero, zero}},
		{PatientName: "Asli", PatientNationalID: "116", PatientFamilyID: "22", PatientDiseaseTable: [3]*big.Int{zero, zero, zero}},
		{PatientName: "Safiye", PatientNationalID: "117", PatientFamilyID: "22", PatientDiseaseTable: [3]*big.Int{zero, zero, zero}},
		{PatientName: "Mushab", PatientNationalID: "118", PatientFamilyID: "22", PatientDiseaseTable: [3]*big.Int{zero, zero, zero}},
		{PatientName: "Sakir", PatientNationalID: "119", PatientFamilyID: "22", PatientDiseaseTable: [3]*big.Int{zero, zero, zero}},
		{PatientName: "Yadigar", PatientNationalID: "120", PatientFamilyID: "22", PatientDiseaseTable: [3]*big.Int{zero, zero, zero}},
		{PatientName: "Hamza", PatientNationalID: "121", PatientFamilyID: "22", PatientDiseaseTable: [3]*big.Int{zero, zero, zero}},
	}

	previousFamilyID := patients[0].PatientFamilyID
	var paillerAssets []PaillerKey

	publicKey, privateKey, _ := Pailler.GenerateKeyPair(1024)
	N, g := publicKey.ToString()
	publicKey2, _ := Pailler.NewPublicKey(N, g)

	fmt.Println("Keys Generated...")

	for _, patient := range patients {
		if previousFamilyID != patient.PatientFamilyID {
			previousFamilyID = patient.PatientFamilyID
			publicKey, privateKey, _ = Pailler.GenerateKeyPair(1024)
			N, g = publicKey.ToString()
			publicKey2, _ = Pailler.NewPublicKey(N, g)
			pailler := PaillerKey{PatientFamilyID: patient.PatientFamilyID, Key: privateKey}
			paillerAssets = append(paillerAssets, pailler)
			fmt.Println("Keys Generated For FamilyID" + patient.PatientFamilyID)
		}
		for index, _ := range patient.PatientDiseaseTable {
			patient.PatientDiseaseTable[index], _ = publicKey2.Encrypt(0)
		}

		patientJSON, err := json.Marshal(patient)
		if err != nil {
			return shim.Error("Json Mars")
		}
		err = stub.PutState(patient.PatientNationalID, patientJSON)
		if err != nil {
			return shim.Error("Cannot put Patient to the ledger")
		}

		fmt.Println("Encryption Done for Patient : " + patient.PatientName)
	}

	disease := Diseases{SickleCellDisease: 100, Type2Diabetes: 70, Achondroplasia: 50}
	diseaseJSON, err := json.Marshal(disease)
	if err != nil {
		return shim.Error("Json Mars")
	}
	err = stub.PutState("DiseaseTable", diseaseJSON)
	if err != nil {
		return shim.Error("Cannot put Patient to the ledger")
	}

	for _, paillerAsset := range paillerAssets {
		paillerJSON, err := json.Marshal(paillerAsset)
		if err != nil {
			return shim.Error("Json Mars")
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
	case "calculateDiseaseProbabilityWithoutTree":
		return t.calculateDiseaseProbabilityWithoutTree(stub, args)
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
	paillerKey := new(PaillerKey)
	patientNationalID := args[0]

	patient = getPatient(stub, patientNationalID)
	paillerKey = getPaillerKey(stub, patient.PatientFamilyID)
	fmt.Println("Patient's Name : " + patient.PatientName)
	fmt.Println("Patient's FamilyID : " + paillerKey.PatientFamilyID)
	fmt.Println("Patient's Diseses Values")

	for _, encryptedValue := range patient.PatientDiseaseTable {
		fmt.Println(encryptedValue)
		decryptedValue, err := paillerKey.Key.Decrypt(encryptedValue)
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
	paillerKey := new(PaillerKey)
	patientNationalID := args[0]

	diseaseIndex, err := strconv.Atoi(args[1])
	if err != nil {
		return shim.Error("Second argument is not an integer")
	}

	patient = getPatient(stub, patientNationalID)
	paillerKey = getPaillerKey(stub, patient.PatientFamilyID)

	patient.PatientDiseaseTable[diseaseIndex], err = paillerKey.Key.Pk.Encrypt(1)
	if err != nil {
		return shim.Error("Patient's disease value cannot assigned to encrypted 1")
	}

	err = stub.DelState(patientNationalID)
	if err != nil {
		return shim.Error("Delete State Error")
	}

	patientJSON, err := json.Marshal(patient)
	if err != nil {
		return shim.Error("Json Mars")
	}
	err = stub.PutState(patient.PatientNationalID, patientJSON)
	if err != nil {
		return shim.Error("Cannot put Patient to the ledger")
	}

	return shim.Success(nil)
}

func (t *Patient) calculateDiseaseProbabilityWithoutTree(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	ancestorIds := []string{"115", "116", "119", "120"}

	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments. Expecting 2")
	}

	level := 0

	patient := new(Patient)
	patientProps := new(PaillerKey)
	disease := new(Diseases)
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
		return shim.Error("PaillerProps can't be fetched")
	}

	diseaseAsset, err := stub.GetState(patient.PatientFamilyID)
	if err != nil {
		return shim.Error("Pailler Props can't be fetched")
	}
	err = json.Unmarshal(diseaseAsset, disease)
	if err != nil {
		return shim.Error("PaillerProps can't be fetched")
	}

	diseaseProbability := disease.Achondroplasia
	result, _ := patientProps.Key.Pk.Encrypt(0)

	fmt.Println(patient.PatientName)
	fmt.Println(patient.PatientNationalID)
	fmt.Println(patientProps.PatientFamilyID)

	for i := 0; i < 4; i += 2 {

		level++

		fatherId := ancestorIds[i]
		fatherPatient := getPatient(stub, fatherId)
		fmt.Println("Father Name Is : " + fatherPatient.PatientName)
		probability, _ := fatherPatient.calcualte(patientProps, diseaseProbability, level, diseaseIndex)
		fmt.Println("Father Calculation Done " + fatherPatient.PatientName)
		fmt.Println(probability)
		result, err = patientProps.Key.Pk.Add(result, probability)
		if err != nil {
			fmt.Println(err)
			return shim.Error("Father Calculation Error")
		}
		fmt.Println(result)

		motherId := ancestorIds[i+1]
		motherPatient := getPatient(stub, motherId)
		fmt.Println("Mother Name Is : " + motherPatient.PatientName)
		probability, err = motherPatient.calcualte(patientProps, diseaseProbability, level, diseaseIndex)
		fmt.Println("Mother Calculation Done" + motherPatient.PatientName)
		if err != nil {
			return shim.Error("Mother Calculation Error")
		}
		result, err = patientProps.Key.Pk.Add(result, probability)
		fmt.Println(result)

	}

	fmt.Println("******************************")
	fmt.Println(result)
	fmt.Println("******************************")
	fmt.Println(patientProps.Key.Decrypt(result))
	fmt.Println("******************************")

	return shim.Success(nil)
}

func getPatient(stub shim.ChaincodeStubInterface, nationalID string) *Patient {

	patient := new(Patient)
	patientAsset, err := stub.GetState(nationalID)
	if err != nil {
		fmt.Println("GetState Error")
	}
	err = json.Unmarshal(patientAsset, patient)
	if err != nil {
		fmt.Println("Unmarshal Error")
	}
	return patient
}

func getPaillerKey(stub shim.ChaincodeStubInterface, familyID string) *PaillerKey {

	paillerKey := new(PaillerKey)
	paillerAsset, err := stub.GetState(familyID)
	if err != nil {
		fmt.Println("GetState Error")
	}
	err = json.Unmarshal(paillerAsset, paillerKey)
	if err != nil {
		fmt.Println("Unmarshal Error")
	}
	return paillerKey
}

func (t *Patient) calcualte(patientProps *PaillerKey, diseaseProbability int, level int, diseaseIndex int) (*big.Int, error) {
	diseaseValue := t.PatientDiseaseTable[diseaseIndex]
	diseaseContribution := diseaseProbability / level
	return patientProps.Key.Pk.MultPlaintext(diseaseValue, int64(diseaseContribution))
}
