package keystore

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	ecdsaKeygen "github.com/bnb-chain/tss-lib/v2/ecdsa/keygen"
	eddsaKeygen "github.com/bnb-chain/tss-lib/v2/eddsa/keygen"
)

const (
	keyFolder = "./keys"
)

func SaveECDSAKey(keyID string, partyIndex int, keyData *ecdsaKeygen.LocalPartySaveData) error {
	fileName := fmt.Sprintf("%s_ecdsa_key_party_%d.json", keyID, partyIndex)
	return saveKey(fileName, keyData)
}

func SaveEdDSAKey(keyID string, partyIndex int, keyData *eddsaKeygen.LocalPartySaveData) error {
	fileName := fmt.Sprintf("%s_eddsa_key_party_%d.json", keyID, partyIndex)
	return saveKey(fileName, keyData)
}

func LoadAllECDSAKeys(keyID string) ([]ecdsaKeygen.LocalPartySaveData, error) {
	var keys []ecdsaKeygen.LocalPartySaveData
	i := 0
	for {
		fileName := fmt.Sprintf("%s_ecdsa_key_party_%d.json", keyID, i)
		filePath := filepath.Join(keyFolder, fileName)
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			break // No more files for this keyID
		}

		file, err := os.ReadFile(filePath)
		if err != nil {
			return nil, fmt.Errorf("failed to read key file %s: %w", filePath, err)
		}

		var keyData ecdsaKeygen.LocalPartySaveData
		if err := json.Unmarshal(file, &keyData); err != nil {
			return nil, fmt.Errorf("failed to unmarshal key data from %s: %w", filePath, err)
		}
		keys = append(keys, keyData)
		i++
	}
	if len(keys) == 0 {
		return nil, fmt.Errorf("no ECDSA keys found for keyID: %s", keyID)
	}
	return keys, nil
}

func LoadAllEdDSAKeys(keyID string) ([]eddsaKeygen.LocalPartySaveData, error) {
	var keys []eddsaKeygen.LocalPartySaveData
	i := 0
	for {
		fileName := fmt.Sprintf("%s_eddsa_key_party_%d.json", keyID, i)
		filePath := filepath.Join(keyFolder, fileName)
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			break // No more files for this keyID
		}

		file, err := os.ReadFile(filePath)
		if err != nil {
			return nil, fmt.Errorf("failed to read key file %s: %w", filePath, err)
		}

		var keyData eddsaKeygen.LocalPartySaveData
		if err := json.Unmarshal(file, &keyData); err != nil {
			return nil, fmt.Errorf("failed to unmarshal key data from %s: %w", filePath, err)
		}
		keys = append(keys, keyData)
		i++
	}
	if len(keys) == 0 {
		return nil, fmt.Errorf("no EdDSA keys found for keyID: %s", keyID)
	}
	return keys, nil
}

func saveKey(fileName string, keyData interface{}) error {
	if err := os.MkdirAll(keyFolder, 0700); err != nil {
		return fmt.Errorf("failed to create key folder: %w", err)
	}

	filePath := filepath.Join(keyFolder, fileName)
	file, err := json.MarshalIndent(keyData, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal key data: %w", err)
	}

	return os.WriteFile(filePath, file, 0600)
}
