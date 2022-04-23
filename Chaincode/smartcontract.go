package chaincode

import (
	"encoding/json"
	"fmt"
	Pailler "github.com/hyperledger/fabric-samples/asset-transfer-basic/chaincode-go/chaincode/paillerCrypto"
	"math/big"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// SmartContract provides functions for managing an Asset
type SmartContract struct {
	contractapi.Contract
}

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

// InitLedger adds a base set of assets to the ledger
func (s *SmartContract) InitLedger(ctx contractapi.TransactionContextInterface) error {
	patients := []Patient{
		{PatientName: "Erhan", PatientNationalID: "112", PatientFamilyID: "20", PatientDiseaseTable: [3]*big.Int{}},
		{PatientName: "Aysegul", PatientNationalID: "113", PatientFamilyID: "20", PatientDiseaseTable: [3]*big.Int{}},
		{PatientName: "Ahmet", PatientNationalID: "111", PatientFamilyID: "20", PatientDiseaseTable: [3]*big.Int{}},
		{PatientName: "Recep", PatientNationalID: "114", PatientFamilyID: "21", PatientDiseaseTable: [3]*big.Int{}},
		{PatientName: "Nusret", PatientNationalID: "115", PatientFamilyID: "22", PatientDiseaseTable: [3]*big.Int{}},
		{PatientName: "Asli", PatientNationalID: "116", PatientFamilyID: "22", PatientDiseaseTable: [3]*big.Int{}},
		{PatientName: "Safiye", PatientNationalID: "117", PatientFamilyID: "22", PatientDiseaseTable: [3]*big.Int{}},
		{PatientName: "Mushab", PatientNationalID: "118", PatientFamilyID: "22", PatientDiseaseTable: [3]*big.Int{}},
		{PatientName: "Sakir", PatientNationalID: "119", PatientFamilyID: "22", PatientDiseaseTable: [3]*big.Int{}},
		{PatientName: "Yadigar", PatientNationalID: "120", PatientFamilyID: "22", PatientDiseaseTable: [3]*big.Int{}},
		{PatientName: "Hamza", PatientNationalID: "121", PatientFamilyID: "22", PatientDiseaseTable: [3]*big.Int{}},
	}

	previousFamilyID := patients[0].PatientFamilyID
	var paillerAssets []PaillerKey

	publicKey, privateKey, _ := Pailler.GenerateKeyPair(1024)
	N, g := publicKey.ToString()
	publicKey2, _ := Pailler.NewPublicKey(N, g)

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
			return err
		}
		err = ctx.GetStub().PutState(patient.PatientNationalID, patientJSON)
		if err != nil {
			return err
		}

		fmt.Println("Encryption Done for Patient : " + patient.PatientName)
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
		err = ctx.GetStub().PutState(paillerAsset.PatientFamilyID, paillerJSON)
		if err != nil {
			return err
		}
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

	asset, err := ctx.GetStub().GetState(patientFamilyID)
	if err != nil {
		return err
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

	firstDiseaseEncrypted, err := paillerAsset.Key.Pk.Encrypt(int64(firstDisease)) // First Disease Value Encryption
	if err != nil {
		return fmt.Errorf("encryption error")
	}
	secondDiseaseEncrypted, err := paillerAsset.Key.Pk.Encrypt(int64(secondDisease)) // Second Disease Value Encryption
	thirdDiseaseEncrypted, err := paillerAsset.Key.Pk.Encrypt(int64(thirdDisease))   // Third Disease Value Encryption

	patient :=
		Patient{
			PatientName:         patientName,
			PatientNationalID:   patientNationalID,
			PatientFamilyID:     patientFamilyID,
			PatientDiseaseTable: [3]*big.Int{firstDiseaseEncrypted, secondDiseaseEncrypted, thirdDiseaseEncrypted},
		}

	patientJSON, err := json.Marshal(patient) // Patient Information Encoded as JSON
	if err != nil {
		return fmt.Errorf("asset cannot encoded right now")
	}

	err = ctx.GetStub().PutState(patient.PatientNationalID, patientJSON) // Patinet Information Saved To The Ledger
	if err != nil {
		return fmt.Errorf("failed in put state")
	}

	paillerJSON, err := json.Marshal(paillerAsset) // Pailler Key Information Encoded as JSON
	if err != nil {
		return fmt.Errorf("asset cannot encoded right now")
	}

	err = ctx.GetStub().PutState(paillerAsset.PatientFamilyID, paillerJSON) // Pailler Key Information Saved To The Ledger
	if err != nil {
		return fmt.Errorf("failed in put state")
	}

	return nil
}

// ReadAsset returns the asset stored in the world state with given id.
func (s *SmartContract) ReadAsset(ctx contractapi.TransactionContextInterface, patientNationalID string) (*Patient, error) {
	assetJSON, err := ctx.GetStub().GetState(patientNationalID)
	if err != nil {
		return nil, fmt.Errorf("failed to read from world state: %v", err)
	}
	if assetJSON == nil {
		return nil, fmt.Errorf("the asset %s does not exist", patientNationalID)
	}

	var asset Patient
	err = json.Unmarshal(assetJSON, &asset)
	if err != nil {
		return nil, err
	}

	return &asset, nil
}

// UpdateAsset updates an existing asset in the world state with provided parameters.
func (s *SmartContract) UpdateAsset(ctx contractapi.TransactionContextInterface, patientNationalID string, diseaseIndex int) error {
	exists, err := s.AssetExists(ctx, patientNationalID)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("the asset %s does not exist", patientNationalID)
	}

	patient := getPatient(ctx, patientNationalID)
	paillerKey := getPaillerKey(ctx, patient.PatientFamilyID)

	patient.PatientDiseaseTable[diseaseIndex], err = paillerKey.Key.Pk.Encrypt(1)
	if err != nil {
		return fmt.Errorf("patient's disease value cannot assigned to encrypted 1")
	}

	err = ctx.GetStub().DelState(patientNationalID)
	if err != nil {
		return fmt.Errorf("delete State Error")
	}

	patientJSON, err := json.Marshal(patient)
	if err != nil {
		return fmt.Errorf("json Mars")
	}

	return ctx.GetStub().PutState(patient.PatientNationalID, patientJSON)
}

// DeleteAsset deletes an given asset from the world state.
func (s *SmartContract) DeleteAsset(ctx contractapi.TransactionContextInterface, patientNationalID string) error {
	exists, err := s.AssetExists(ctx, patientNationalID)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("the asset %s does not exist", patientNationalID)
	}

	return ctx.GetStub().DelState(patientNationalID)
}

// AssetExists returns true when asset with given ID exists in world state
func (s *SmartContract) AssetExists(ctx contractapi.TransactionContextInterface, id string) (bool, error) {
	assetJSON, err := ctx.GetStub().GetState(id)
	if err != nil {
		return false, fmt.Errorf("failed to read from world state: %v", err)
	}

	return assetJSON != nil, nil
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

	paillerAsset, err := ctx.GetStub().GetState(patient.PatientFamilyID)
	if err != nil {
		return fmt.Errorf("pailler Props can't be fetched")
	}
	err = json.Unmarshal(paillerAsset, patientProps)
	if err != nil {
		return fmt.Errorf("PaillerProps can't be fetched")
	}

	diseaseAsset, err := ctx.GetStub().GetState(patient.PatientFamilyID)
	if err != nil {
		return fmt.Errorf("pailler Props can't be fetched")
	}
	err = json.Unmarshal(diseaseAsset, disease)
	if err != nil {
		return fmt.Errorf("PaillerProps can't be fetched")
	}

	diseaseProbability := disease.Achondroplasia
	result, _ := patientProps.Key.Pk.Encrypt(0)

	fmt.Println(patient.PatientName)
	fmt.Println(patient.PatientNationalID)
	fmt.Println(patientProps.PatientFamilyID)

	for i := 0; i < 4; i += 2 {

		level++

		fatherId := ancestorIds[i]
		fatherPatient := getPatient(ctx, fatherId)
		fmt.Println("Father Name Is : " + fatherPatient.PatientName)
		probability, _ := fatherPatient.calcualte(patientProps, diseaseProbability, level, diseaseIndex)
		fmt.Println("Father Calculation Done " + fatherPatient.PatientName)
		fmt.Println(probability)
		result, err = patientProps.Key.Pk.Add(result, probability)
		if err != nil {
			fmt.Println(err)
			return fmt.Errorf("father Calculation Error")
		}
		fmt.Println(result)

		motherId := ancestorIds[i+1]
		motherPatient := getPatient(ctx, motherId)
		fmt.Println("Mother Name Is : " + motherPatient.PatientName)
		probability, err = motherPatient.calcualte(patientProps, diseaseProbability, level, diseaseIndex)
		fmt.Println("Mother Calculation Done" + motherPatient.PatientName)
		if err != nil {
			return fmt.Errorf("mother Calculation Error")
		}
		result, err = patientProps.Key.Pk.Add(result, probability)
		fmt.Println(result)
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
	diseaseValue := t.PatientDiseaseTable[diseaseIndex]
	diseaseContribution := diseaseProbability / level
	return patientProps.Key.Pk.MultPlaintext(diseaseValue, int64(diseaseContribution))
}
