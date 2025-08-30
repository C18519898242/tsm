package tss

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"

	"github.com/bnb-chain/tss-lib/v2/ecdsa/keygen"
)

const (
	preParamsFolder = "./keys"
	preParamsPrefix = "pre-params"
)

func preParamsFileName(partyIndex int) string {
	return fmt.Sprintf("%s_%d.json", preParamsPrefix, partyIndex)
}

// LoadPreParams loads pre-computed parameters for a party from a file.
func LoadPreParams(partyIndex int) (*keygen.LocalPreParams, error) {
	fileName := preParamsFileName(partyIndex)
	filePath := filepath.Join(preParamsFolder, fileName)

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil, err
	}

	bz, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read pre-params file %s: %w", filePath, err)
	}

	var params keygen.LocalPreParams
	if err := json.Unmarshal(bz, &params); err != nil {
		return nil, fmt.Errorf("failed to unmarshal pre-params from %s: %w", filePath, err)
	}
	return &params, nil
}

// GenerateAndSavePreParams generates and saves pre-computed parameters for all parties.
// This is a one-time, expensive operation.
func GenerateAndSavePreParams(partyCount int) ([]*keygen.LocalPreParams, error) {
	fmt.Printf("Generating new pre-parameters for %d parties... This may take a long time.\n", partyCount)
	preParams := make([]*keygen.LocalPreParams, partyCount)
	var wg sync.WaitGroup
	concurrency := runtime.NumCPU()
	if concurrency > partyCount {
		concurrency = partyCount
	}
	tasks := make(chan int, partyCount)

	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for partyIndex := range tasks {
				fmt.Printf("Generating for party %d...\n", partyIndex)
				// The library's default timeout is 10 minutes, but we'll give it 15 for safety.
				params, err := keygen.GeneratePreParams(15 * time.Minute)
				if err != nil {
					fmt.Printf("Error generating pre-params for party %d: %v\n", partyIndex, err)
					// In a real app, you'd handle this more gracefully.
					// For this CLI, we'll just let it fail.
					continue
				}
				preParams[partyIndex] = params
			}
		}()
	}

	for i := 0; i < partyCount; i++ {
		tasks <- i
	}
	close(tasks)
	wg.Wait()

	// Check if all were generated
	for i, params := range preParams {
		if params == nil {
			return nil, fmt.Errorf("failed to generate pre-params for all parties (party %d failed)", i)
		}
	}

	// Save them
	for i, params := range preParams {
		if err := savePreParams(i, params); err != nil {
			return nil, fmt.Errorf("failed to save pre-params for party %d: %w", i, err)
		}
		fmt.Printf("Saved pre-params for party %d.\n", i)
	}

	fmt.Println("Finished generating pre-parameters.")
	return preParams, nil
}

func savePreParams(partyIndex int, params *keygen.LocalPreParams) error {
	fileName := preParamsFileName(partyIndex)
	filePath := filepath.Join(preParamsFolder, fileName)

	if err := os.MkdirAll(preParamsFolder, 0700); err != nil {
		return fmt.Errorf("failed to create pre-params folder: %w", err)
	}

	bz, err := json.MarshalIndent(params, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal pre-params: %w", err)
	}

	return os.WriteFile(filePath, bz, 0600)
}
