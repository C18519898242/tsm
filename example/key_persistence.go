package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	ecdsaKeygen "github.com/bnb-chain/tss-lib/v2/ecdsa/keygen"
	eddsaKeygen "github.com/bnb-chain/tss-lib/v2/eddsa/keygen"
)

const (
	ecdsaKeyFilePrefix = "ecdsa_key_party_"
	eddsaKeyFilePrefix = "eddsa_key_party_"
	keyFileSuffix      = ".json"
	keysDir            = "keys"
)

// SaveECDSAKey saves the ECDSA LocalPartySaveData for a given party index to a JSON file.
func SaveECDSAKey(partyIndex int, data *ecdsaKeygen.LocalPartySaveData) error {
	if _, err := os.Stat(keysDir); os.IsNotExist(err) {
		os.Mkdir(keysDir, 0755)
	}
	fileName := filepath.Join(keysDir, fmt.Sprintf("%s%d%s", ecdsaKeyFilePrefix, partyIndex, keyFileSuffix))
	file, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal ECDSA key data: %w", err)
	}
	return ioutil.WriteFile(fileName, file, 0644)
}

// LoadECDSAKey loads the ECDSA LocalPartySaveData for a given party index from a JSON file.
func LoadECDSAKey(partyIndex int) (*ecdsaKeygen.LocalPartySaveData, error) {
	fileName := filepath.Join(keysDir, fmt.Sprintf("%s%d%s", ecdsaKeyFilePrefix, partyIndex, keyFileSuffix))
	file, err := ioutil.ReadFile(fileName)
	if err != nil {
		return nil, fmt.Errorf("failed to read ECDSA key file: %w", err)
	}
	var data ecdsaKeygen.LocalPartySaveData
	if err := json.Unmarshal(file, &data); err != nil {
		return nil, fmt.Errorf("failed to unmarshal ECDSA key data: %w", err)
	}
	return &data, nil
}

// SaveEdDSAKey saves the EdDSA LocalPartySaveData for a given party index to a JSON file.
func SaveEdDSAKey(partyIndex int, data *eddsaKeygen.LocalPartySaveData) error {
	if _, err := os.Stat(keysDir); os.IsNotExist(err) {
		os.Mkdir(keysDir, 0755)
	}
	fileName := filepath.Join(keysDir, fmt.Sprintf("%s%d%s", eddsaKeyFilePrefix, partyIndex, keyFileSuffix))
	file, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal EdDSA key data: %w", err)
	}
	return ioutil.WriteFile(fileName, file, 0644)
}

// LoadEdDSAKey loads the EdDSA LocalPartySaveData for a given party index from a JSON file.
func LoadEdDSAKey(partyIndex int) (*eddsaKeygen.LocalPartySaveData, error) {
	fileName := filepath.Join(keysDir, fmt.Sprintf("%s%d%s", eddsaKeyFilePrefix, partyIndex, keyFileSuffix))
	file, err := ioutil.ReadFile(fileName)
	if err != nil {
		return nil, fmt.Errorf("failed to read EdDSA key file: %w", err)
	}
	var data eddsaKeygen.LocalPartySaveData
	if err := json.Unmarshal(file, &data); err != nil {
		return nil, fmt.Errorf("failed to unmarshal EdDSA key data: %w", err)
	}
	return &data, nil
}
