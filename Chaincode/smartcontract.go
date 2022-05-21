package chaincode

import (
	"encoding/json"
	"fmt"
	"math/big"
	"strconv"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	Pailler "github.com/hyperledger/fabric-samples/asset-transfer-basic/chaincode-go/chaincode/paillerCrypto"
)

// SmartContract provides functions for managing an Asset
type SmartContract struct {
	contractapi.Contract
}

type Patient struct {
	PatientName         string      `json:"patientName"`
	PatientNationalID   int         `json:"patientNationalID"`
	PatientFamilyID     int         `json:"patientFamilyID"`
	PatientDiseaseTable [3]*big.Int `json:"patientDiseaseTable"`
}

type Diseases struct {
	SickleCellDisease int `json:"sickleCellDisease"`
	Type2Diabetes     int `json:"type2Diabetes"`
	Achondroplasia    int `json:"achondroplasia"`
}

// InitLedger adds a base set of assets to the ledger
func (s *SmartContract) InitLedger(ctx contractapi.TransactionContextInterface) error {

	var zero = new(big.Int).SetInt64(0)
	var one = new(big.Int).SetInt64(1)

	patients := []Patient{
		{PatientName: "Erhan", PatientNationalID: 112, PatientFamilyID: 20, PatientDiseaseTable: [3]*big.Int{zero, zero, zero}},
		{PatientName: "Aysegul", PatientNationalID: 113, PatientFamilyID: 20, PatientDiseaseTable: [3]*big.Int{zero, zero, zero}},
		{PatientName: "Ahmet", PatientNationalID: 111, PatientFamilyID: 20, PatientDiseaseTable: [3]*big.Int{zero, zero, zero}},
		{PatientName: "Recep", PatientNationalID: 114, PatientFamilyID: 21, PatientDiseaseTable: [3]*big.Int{zero, zero, zero}},
		{PatientName: "Nusret", PatientNationalID: 115, PatientFamilyID: 22, PatientDiseaseTable: [3]*big.Int{zero, zero, zero}},
		{PatientName: "Asli", PatientNationalID: 116, PatientFamilyID: 22, PatientDiseaseTable: [3]*big.Int{zero, zero, zero}},
		{PatientName: "Safiye", PatientNationalID: 117, PatientFamilyID: 22, PatientDiseaseTable: [3]*big.Int{zero, zero, zero}},
		{PatientName: "Mushab", PatientNationalID: 118, PatientFamilyID: 22, PatientDiseaseTable: [3]*big.Int{zero, zero, zero}},
		{PatientName: "Sakir", PatientNationalID: 119, PatientFamilyID: 22, PatientDiseaseTable: [3]*big.Int{zero, one, zero}},
		{PatientName: "Yadigar", PatientNationalID: 120, PatientFamilyID: 22, PatientDiseaseTable: [3]*big.Int{zero, one, zero}},
		{PatientName: "Hamza", PatientNationalID: 121, PatientFamilyID: 22, PatientDiseaseTable: [3]*big.Int{zero, zero, zero}},
	}

	for _, asset := range patients {
		publicKey, _, _ := Pailler.GenerateKeyPair(asset.PatientFamilyID)
		N, g := publicKey.ToString()
		publicKey2, _ := Pailler.NewPublicKey(N, g)

		for index := range asset.PatientDiseaseTable {
			asset.PatientDiseaseTable[index], _ = publicKey2.Encrypt(asset.PatientDiseaseTable[index].Int64(), int64(asset.PatientNationalID))
		}

		assetJSON, err := json.Marshal(asset)
		if err != nil {
			return err
		}

		keyNationalID := strconv.Itoa(asset.PatientNationalID)
		err = ctx.GetStub().PutState(keyNationalID, assetJSON)
		if err != nil {
			return fmt.Errorf("failed to put to world state. %v", err)
		}
	}

	disease := Diseases{SickleCellDisease: 100, Type2Diabetes: 70, Achondroplasia: 50}
	diseaseJSON, err := json.Marshal(disease)
	if err != nil {
		return err
	}
	err = ctx.GetStub().PutState("DiseaseTable", diseaseJSON)
	if err != nil {
		return err
	}

	return nil
}

// ReadAsset returns the asset stored in the world state with given id.
func (s *SmartContract) ReadAsset(ctx contractapi.TransactionContextInterface, patientNationalID string) error {
	assetJSON, err := ctx.GetStub().GetState(patientNationalID)
	if err != nil {
		return err
	}
	if assetJSON == nil {
		return err
	}

	var asset Patient
	err = json.Unmarshal(assetJSON, &asset)
	if err != nil {
		return err
	}

	fmt.Println(asset)

	return err
}

// AssetExists returns true when asset with given ID exists in world state
func (s *SmartContract) AssetExists(ctx contractapi.TransactionContextInterface, id string) (bool, error) {
	assetJSON, err := ctx.GetStub().GetState(id)
	if err != nil {
		return false, fmt.Errorf("failed to read from world state: %v", err)
	}
	return assetJSON != nil, nil
}

// ChangeAsset returns true when asset with given ID exists in world state
func (s *SmartContract) ChangeAsset(ctx contractapi.TransactionContextInterface, patientNationalID int, diseaseIndex int) error {
	patient := getPatient(ctx, patientNationalID)

	publicKey, _, _ := Pailler.GenerateKeyPair(patient.PatientFamilyID)
	N, g := publicKey.ToString()
	publicKey2, _ := Pailler.NewPublicKey(N, g)

	patient.PatientDiseaseTable[diseaseIndex], _ = publicKey2.Encrypt(1, int64(patientNationalID))

	fmt.Println("Patient's Disease Value Is Changed")
	fmt.Println(patient)

	return nil
}

// GetAllAssets returns all assets found in world state
func (s *SmartContract) GetAllAssets(ctx contractapi.TransactionContextInterface) error {
	// range query with empty string for startKey and endKey does an
	// open-ended query of all assets in the chaincode namespace.
	resultsIterator, err := ctx.GetStub().GetStateByRange("", "")
	if err != nil {
		return err
	}
	defer resultsIterator.Close()

	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return err
		}

		var asset Patient
		err = json.Unmarshal(queryResponse.Value, &asset)
		if err != nil {
			return err
		}
		fmt.Println(asset)
	}

	return nil
}

// CreateAsset issues a new asset to the world state with given details.
func (s *SmartContract) CreateAsset(ctx contractapi.TransactionContextInterface, patientName string, patientNationalID string, patientFamilyID string, firstDisease int, secondDisease int, thirdDisease int) error {
	exists, err := s.AssetExists(ctx, patientNationalID)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("the asset %s already exists", patientNationalID)
	}

	patientfamilyidInt, err := strconv.Atoi(patientFamilyID)
	patientnationalidInt, err := strconv.Atoi(patientNationalID)

	publicKey, _, _ := Pailler.GenerateKeyPair(patientfamilyidInt)
	N, g := publicKey.ToString()
	publicKey2, _ := Pailler.NewPublicKey(N, g)

	firstDiseaseEncrypted, err := publicKey2.Encrypt(int64(firstDisease), int64(patientnationalidInt)) // First Disease Value Encryption
	if err != nil {
		return fmt.Errorf("encryption error")
	}
	secondDiseaseEncrypted, err := publicKey2.Encrypt(int64(secondDisease), int64(patientnationalidInt)) // Second Disease Value Encryption
	thirdDiseaseEncrypted, err := publicKey2.Encrypt(int64(thirdDisease), int64(patientnationalidInt))   // Third Disease Value Encryption

	patient :=
		Patient{
			PatientName:         patientName,
			PatientNationalID:   patientnationalidInt,
			PatientFamilyID:     patientfamilyidInt,
			PatientDiseaseTable: [3]*big.Int{firstDiseaseEncrypted, secondDiseaseEncrypted, thirdDiseaseEncrypted},
		}

	patientJSON, err := json.Marshal(patient) // Patient Information Encoded as JSON
	if err != nil {
		return fmt.Errorf("asset cannot encoded right now")
	}

	err = ctx.GetStub().PutState(patientNationalID, patientJSON) // Patinet Information Saved To The Ledger
	if err != nil {
		return fmt.Errorf("failed in put state")
	}

	return nil
}

func (s *SmartContract) DeleteAsset(ctx contractapi.TransactionContextInterface, patientNationalID string) error {
	// Delete the key from the state in ledger
	err := ctx.GetStub().DelState(patientNationalID)
	if err != nil {
		return err
	}
	return nil
}

// TransferAsset updates the owner field of asset with given id in world state.
func (s *SmartContract) TransferAsset(ctx contractapi.TransactionContextInterface, patientNationalID string, diseaseIndex int) error {

	ancestorIds := []int{115, 116, 119, 120}

	level := 0

	patient := new(Patient)
	disease := new(Diseases)

	patientAsset, err := ctx.GetStub().GetState(patientNationalID)
	if err != nil {
		return fmt.Errorf("patient doesn't exist")
	}
	err = json.Unmarshal(patientAsset, patient)
	if err != nil {
		return fmt.Errorf("patient can't be fetched")
	}

	publicKey, privateKey, _ := Pailler.GenerateKeyPair(patient.PatientFamilyID)
	N, g := publicKey.ToString()
	publicKey2, _ := Pailler.NewPublicKey(N, g)

	diseaseAsset, err := ctx.GetStub().GetState("DiseaseTable")
	if err != nil {
		return fmt.Errorf("pailler Props can't be fetched")
	}
	err = json.Unmarshal(diseaseAsset, disease)
	if err != nil {
		return fmt.Errorf("PaillerProps can't be fetched")
	}

	diseaseProbability := disease.Achondroplasia
	result, _ := publicKey2.Encrypt(0, int64(patient.PatientNationalID))

	fmt.Println("******************************")
	fmt.Println(patient.PatientName)
	fmt.Println(patient.PatientNationalID)
	fmt.Println("******************************")

	for i := 0; i < 4; i += 2 {

		level++

		fatherId := ancestorIds[i]
		fatherPatient := getPatient(ctx, fatherId)
		probability, _ := fatherPatient.calculate(publicKey2, diseaseProbability, level, diseaseIndex)
		fmt.Println("Father's Calculation Done " + fatherPatient.PatientName)
		fmt.Print("Probability : ")
		fmt.Println(probability)
		dResult, _ := privateKey.Decrypt(probability)
		dResultString := strconv.Itoa(int(dResult))
		fmt.Println(" Total Result Decrypted Probability is  : " + dResultString + "%")
		result, _ = publicKey2.Add(result, probability)

		fmt.Print("Result : ")
		fmt.Println(result)
		fmt.Print("Decrypt Result : ")
		fmt.Println(privateKey.Decrypt(result))

		motherId := ancestorIds[i+1]
		motherPatient := getPatient(ctx, motherId)
		probability, err = motherPatient.calculate(publicKey2, diseaseProbability, level, diseaseIndex)
		fmt.Println("Mother's Calculation Done " + motherPatient.PatientName)
		fmt.Print("Probability : ")
		fmt.Println(probability)
		dResult, _ = privateKey.Decrypt(probability)
		dResultString = strconv.Itoa(int(dResult))
		fmt.Println(" Total Result Decrypted Probability is  : " + dResultString + "%")
		result, _ = publicKey2.Add(result, probability)

		fmt.Print("Final Result : ")
		fmt.Println(result)
		fmt.Print("Decrypted Final Result : ")
		fmt.Println(privateKey.Decrypt(result))
	}

	fmt.Println("******************************")
	fmt.Println("Encrypted Result : " + result.String())
	fmt.Println("******************************")
	dResult, err := privateKey.Decrypt(result)
	dResultString := strconv.Itoa(int(dResult))
	if err != nil {
		return fmt.Errorf("PaillerProps can't be fetched")
	}
	fmt.Println("The Probability of Patient's Having Achondroplasia Disease Is : " + dResultString + "%")
	fmt.Println("******************************")
	return nil

}

func getPatient(ctx contractapi.TransactionContextInterface, nationalID int) *Patient {
	patient := new(Patient)
	nationalIDString := strconv.Itoa(nationalID)
	patientAsset, err := ctx.GetStub().GetState(nationalIDString)
	if err != nil {
		fmt.Println("GetState Error")
	}
	err = json.Unmarshal(patientAsset, patient)
	if err != nil {
		fmt.Println("Unmarshal Error")
	}
	return patient
}

func (t *Patient) calculate(patientProps *Pailler.PublicKey, diseaseProbability int, level int, diseaseIndex int) (*big.Int, error) {
	diseaseValue := t.PatientDiseaseTable[diseaseIndex]
	diseaseContribution := diseaseProbability / level
	return patientProps.MultPlaintext(diseaseValue, int64(diseaseContribution))
}
