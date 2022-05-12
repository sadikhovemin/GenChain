package chaincode

import (
	"encoding/json"
	"fmt"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	Pailler "github.com/hyperledger/fabric-samples/asset-transfer-basic/chaincode-go/chaincode/paillerCrypto"
	"math/big"
	"strconv"
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

type PaillerKey struct {
	PatientFamilyID int                 `json:"patientFamilyID"`
	Key             *Pailler.PrivateKey `json:"key"`
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

	selectedFamilyID := patients[0].PatientFamilyID
	var paillerAssets []PaillerKey

	publicKey, privateKey, _ := Pailler.GenerateKeyPair(selectedFamilyID)
	N, g := publicKey.ToString()
	publicKey2, _ := Pailler.NewPublicKey(N, g)

	paillerAsset := PaillerKey{PatientFamilyID: patients[0].PatientFamilyID, Key: privateKey}
	paillerAssets = append(paillerAssets, paillerAsset)

	for _, asset := range patients {
		if selectedFamilyID != asset.PatientFamilyID {
			selectedFamilyID = asset.PatientFamilyID
			publicKey, privateKey, _ = Pailler.GenerateKeyPair(selectedFamilyID)
			N, g = publicKey.ToString()
			publicKey2, _ = Pailler.NewPublicKey(N, g)
			pailler := PaillerKey{PatientFamilyID: asset.PatientFamilyID, Key: privateKey}
			paillerAssets = append(paillerAssets, pailler)
		}

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

	for _, paillerAsset := range paillerAssets {
		paillerJSON, err := json.Marshal(paillerAsset)
		if err != nil {
			return err
		}
		keyFamilyID := strconv.Itoa(paillerAsset.PatientFamilyID)
		err = ctx.GetStub().PutState(keyFamilyID, paillerJSON)
		if err != nil {
			return err
		}
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

// ReadAsset returns the asset stored in the world state with given id.
func (s *SmartContract) ReadPailler(ctx contractapi.TransactionContextInterface, patientFamilyID string) error {
	assetJSON, err := ctx.GetStub().GetState(patientFamilyID)
	if err != nil {
		return err
	}
	if assetJSON == nil {
		return err
	}

	var asset PaillerKey
	err = json.Unmarshal(assetJSON, &asset)
	if err != nil {
		return err
	}

	fmt.Println(asset.Key.Pk.N)

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

	var paillerAsset PaillerKey
	patientFamilyID_Int, err := strconv.Atoi(patientFamilyID)
	patientNationalID_Int, err := strconv.Atoi(patientNationalID)

	asset, err := ctx.GetStub().GetState(patientFamilyID)
	if err != nil {
		return err
	}

	if len(asset) == 0 {
		fmt.Println("Family Tree Doesn't Exist")
		_, privateKey, _ := Pailler.GenerateKeyPair(patientFamilyID_Int)
		fmt.Println("Keys Generated...")
		paillerAsset = PaillerKey{
			PatientFamilyID: patientFamilyID_Int,
			Key:             privateKey,
		}
		fmt.Println("Pailler Props Generated...")
	} else {
		err = json.Unmarshal(asset, &paillerAsset)
	}

	firstDiseaseEncrypted, err := paillerAsset.Key.Pk.Encrypt(int64(firstDisease), int64(patientNationalID_Int)) // First Disease Value Encryption
	if err != nil {
		return fmt.Errorf("encryption error")
	}
	secondDiseaseEncrypted, err := paillerAsset.Key.Pk.Encrypt(int64(secondDisease), int64(patientNationalID_Int)) // Second Disease Value Encryption
	thirdDiseaseEncrypted, err := paillerAsset.Key.Pk.Encrypt(int64(thirdDisease), int64(patientNationalID_Int))   // Third Disease Value Encryption

	patient :=
		Patient{
			PatientName:         patientName,
			PatientNationalID:   patientNationalID_Int,
			PatientFamilyID:     patientFamilyID_Int,
			PatientDiseaseTable: [3]*big.Int{firstDiseaseEncrypted, secondDiseaseEncrypted, thirdDiseaseEncrypted},
		}

	fmt.Println(patient)

	patientJSON, err := json.Marshal(patient) // Patient Information Encoded as JSON
	if err != nil {
		return fmt.Errorf("asset cannot encoded right now")
	}

	err = ctx.GetStub().PutState(patientNationalID, patientJSON) // Patinet Information Saved To The Ledger
	if err != nil {
		return fmt.Errorf("failed in put state")
	}

	paillerJSON, err := json.Marshal(paillerAsset) // Pailler Key Information Encoded as JSON
	if err != nil {
		return fmt.Errorf("asset cannot encoded right now")
	}

	err = ctx.GetStub().PutState(patientFamilyID, paillerJSON) // Pailler Key Information Saved To The Ledger
	if err != nil {
		return fmt.Errorf("failed in put state")
	}

	return nil
}

// TransferAsset updates the owner field of asset with given id in world state.
func (s *SmartContract) TransferAsset(ctx contractapi.TransactionContextInterface, patientNationalID string, diseaseIndex int) error {

	ancestorIds := []string{"115", "116", "119", "120"}

	level := 0

	patient := new(Patient)
	patientProps := new(PaillerKey)
	disease := new(Diseases)

	patientAsset, err := ctx.GetStub().GetState(patientNationalID)
	if err != nil {
		return fmt.Errorf("patient doesn't exist")
	}
	err = json.Unmarshal(patientAsset, patient)
	if err != nil {
		return fmt.Errorf("patient can't be fetched")
	}

	patientFamilyID_String := strconv.Itoa(patient.PatientFamilyID)

	paillerAsset, err := ctx.GetStub().GetState(patientFamilyID_String)
	if err != nil {
		return fmt.Errorf("pailler Props can't be fetched")
	}
	err = json.Unmarshal(paillerAsset, patientProps)
	if err != nil {
		return fmt.Errorf("PaillerProps can't be fetched")
	}

	diseaseAsset, err := ctx.GetStub().GetState("DiseaseTable")
	if err != nil {
		return fmt.Errorf("pailler Props can't be fetched")
	}
	err = json.Unmarshal(diseaseAsset, disease)
	if err != nil {
		return fmt.Errorf("PaillerProps can't be fetched")
	}

	diseaseProbability := disease.Achondroplasia
	fmt.Println(diseaseProbability)
	result, _ := patientProps.Key.Pk.Encrypt(0, int64(patient.PatientNationalID))

	fmt.Println("******************************")
	fmt.Println(patient.PatientName)
	fmt.Println(patient.PatientNationalID)
	fmt.Println(patientProps.PatientFamilyID)
	fmt.Println("******************************")

	for i := 0; i < 4; i += 2 {

		level++

		fatherId := ancestorIds[i]
		fatherPatient := getPatient(ctx, fatherId)
		fmt.Println("Father Name Is : " + fatherPatient.PatientName)
		probability, _ := fatherPatient.calcualte(patientProps, diseaseProbability, level, diseaseIndex)
		fmt.Println("Father Calculation Done " + fatherPatient.PatientName)
		fmt.Print("Probability : ")
		fmt.Println(probability)
		fmt.Print("Decrypt Probability : ")
		fmt.Println(patientProps.Key.Decrypt(probability))
		result, _ = patientProps.Key.Pk.Add(result, probability)

		fmt.Print("Result : ")
		fmt.Println(result)
		fmt.Print("Decrypt Result : ")
		fmt.Println(patientProps.Key.Decrypt(result))

		motherId := ancestorIds[i+1]
		motherPatient := getPatient(ctx, motherId)
		fmt.Println("Mother Name Is : " + motherPatient.PatientName)
		probability, err = motherPatient.calcualte(patientProps, diseaseProbability, level, diseaseIndex)
		fmt.Println("Mother Calculation Done " + motherPatient.PatientName)
		fmt.Print("Probability : ")
		fmt.Println(probability)
		fmt.Print("Decrypt Probability : ")
		fmt.Println(patientProps.Key.Decrypt(probability))
		result, _ = patientProps.Key.Pk.Add(result, probability)

		fmt.Print("Result : ")
		fmt.Println(result)
		fmt.Print("Decrypt Result : ")
		fmt.Println(patientProps.Key.Decrypt(result))
	}

	fmt.Println("******************************")
	fmt.Println(result)
	fmt.Println("******************************")
	fmt.Println(patientProps.Key.Decrypt(result))
	fmt.Println("******************************")

	return nil

}

func getPatient(ctx contractapi.TransactionContextInterface, nationalID string) *Patient {
	patient := new(Patient)
	patientAsset, err := ctx.GetStub().GetState(nationalID)
	if err != nil {
		fmt.Println("GetState Error")
	}
	err = json.Unmarshal(patientAsset, patient)
	if err != nil {
		fmt.Println("Unmarshal Error")
	}
	return patient
}

func getPaillerKey(ctx contractapi.TransactionContextInterface, familyID string) *PaillerKey {
	paillerKey := new(PaillerKey)
	paillerAsset, err := ctx.GetStub().GetState(familyID)
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
	fmt.Println(t)
	diseaseValue := t.PatientDiseaseTable[diseaseIndex]
	fmt.Println(diseaseValue)
	diseaseContribution := diseaseProbability / level
	fmt.Println(diseaseContribution)
	return patientProps.Key.Pk.MultPlaintext(diseaseValue, int64(diseaseContribution))
}
