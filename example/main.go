package main

import (
	"fmt"
	"os"
	"sync"

	"github.com/bnb-chain/tss-lib/v2/ecdsa/keygen"
	eddsaKeygen "github.com/bnb-chain/tss-lib/v2/eddsa/keygen" // Alias for eddsa keygen
	"github.com/bnb-chain/tss-lib/v2/tss"
)

const (
	// T and N values for the ECDSA key generation
	// T is the threshold, N is the total number of parties
	// T+1 parties are required to reconstruct the key
	ecdsaTestThreshold = 2
	ecdsaTestPartyNum  = 3

	// T and N values for the EdDSA key generation
	// T is the threshold, N is the total number of parties
	// T+1 parties are required to reconstruct the key
	eddsaTestThreshold = 2
	eddsaTestPartyNum  = 3
)

func runECDSAKeygenExample() {
	fmt.Println("--- Starting ECDSA Key Generation Example ---")
	// 1. Create Party IDs
	partyIDs := tss.GenerateTestPartyIDs(ecdsaTestPartyNum)
	p2pCtx := tss.NewPeerContext(partyIDs)

	// 2. Create Parameters for each party
	params := make([]*tss.Parameters, ecdsaTestPartyNum)
	for i := 0; i < ecdsaTestPartyNum; i++ {
		params[i] = tss.NewParameters(tss.S256(), p2pCtx, partyIDs[i], ecdsaTestPartyNum, ecdsaTestThreshold)
	}

	// 3. Create channels for inter-party communication
	// Each party needs an outbound channel for messages and an inbound channel for messages from other parties
	// Also, an end channel to signal completion and receive the generated key data
	outChs := make(chan tss.Message, ecdsaTestPartyNum)
	endChs := make(chan *keygen.LocalPartySaveData, ecdsaTestPartyNum)
	errChs := make(chan *tss.Error, ecdsaTestPartyNum)

	// 4. Create and start local parties
	parties := make([]tss.Party, ecdsaTestPartyNum)
	for i := 0; i < ecdsaTestPartyNum; i++ {
		parties[i] = keygen.NewLocalParty(params[i], outChs, endChs)
	}

	// 5. Start all parties in goroutines
	var wg sync.WaitGroup
	for i := 0; i < ecdsaTestPartyNum; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			if err := parties[idx].Start(); err != nil {
				errChs <- err
			}
		}(i)
	}

	// 6. Message routing
	// This goroutine listens for messages from any party and routes them to the correct recipient(s)
	go func() {
		for {
			select {
			case msg := <-outChs:
				dest := msg.GetTo()
				if dest == nil { // broadcast message
					for _, p := range parties {
						if p.PartyID().Index == msg.GetFrom().Index {
							continue
						}
						go func(p tss.Party) {
							wireBytes, _, err := msg.WireBytes()
							if err != nil {
								errChs <- tss.NewError(err, "failed to get wire bytes", 0, msg.GetFrom())
								return
							}
							if _, err := p.UpdateFromBytes(wireBytes, msg.GetFrom(), true); err != nil {
								errChs <- err
							}
						}(p)
					}
				} else { // unicast message
					for _, p := range parties {
						if p.PartyID().Index == dest[0].Index {
							go func(p tss.Party) {
								wireBytes, _, err := msg.WireBytes()
								if err != nil {
									errChs <- tss.NewError(err, "failed to get wire bytes", 0, msg.GetFrom())
									return
								}
								if _, err := p.UpdateFromBytes(wireBytes, msg.GetFrom(), false); err != nil {
									errChs <- err
								}
							}(p)
							break
						}
					}
				}
			case err := <-errChs:
				fmt.Printf("Error: %s\n", err.Error())
				os.Exit(1)
			}
		}
	}()

	// 7. Wait for all parties to finish and collect save data
	var saveDatas []*keygen.LocalPartySaveData
	for i := 0; i < ecdsaTestPartyNum; i++ {
		saveData := <-endChs
		saveDatas = append(saveDatas, saveData)
	}
	wg.Wait() // Ensure all party goroutines have finished

	fmt.Println("ECDSA Key generation completed successfully!")

	// 8. Verify the generated public key
	// All parties should have the same public key
	firstPubKey := saveDatas[0].ECDSAPub
	for i := 1; i < ecdsaTestPartyNum; i++ {
		if firstPubKey.X().Cmp(saveDatas[i].ECDSAPub.X()) != 0 || firstPubKey.Y().Cmp(saveDatas[i].ECDSAPub.Y()) != 0 {
			fmt.Printf("Error: Public keys do not match for party %d\n", i)
			os.Exit(1)
		}
	}
	fmt.Printf("Generated ECDSA Public Key: (%s, %s)\n", firstPubKey.X().String(), firstPubKey.Y().String())

	// Example of how to use the generated key shares (for signing, not part of keygen)
	// This part is just for demonstration and would typically be in a separate signing process
	fmt.Println("\nDemonstrating ECDSA key share usage (for signing):")
	// For a real signing process, you would use the `ecdsa/signing` package
	// Here, we'll just show that each party has a share of the private key
	for i, saveData := range saveDatas {
		fmt.Printf("Party %d private share: %s\n", i, saveData.Xi.String())
	}
	fmt.Println("--- Finished ECDSA Key Generation Example ---\n")
}

func runEdDSAKeygenExample() {
	fmt.Println("--- Starting EdDSA Key Generation Example ---")
	// 1. Create Party IDs
	partyIDs := tss.GenerateTestPartyIDs(eddsaTestPartyNum)
	p2pCtx := tss.NewPeerContext(partyIDs)

	// 2. Create Parameters for each party
	// For EdDSA, we use the Ed25519 curve
	params := make([]*tss.Parameters, eddsaTestPartyNum)
	for i := 0; i < eddsaTestPartyNum; i++ {
		params[i] = tss.NewParameters(tss.Edwards(), p2pCtx, partyIDs[i], eddsaTestPartyNum, eddsaTestThreshold)
	}

	// 3. Create channels for inter-party communication
	// Each party needs an outbound channel for messages and an inbound channel for messages from other parties
	// Also, an end channel to signal completion and receive the generated key data
	outChs := make(chan tss.Message, eddsaTestPartyNum)
	endChs := make(chan *eddsaKeygen.LocalPartySaveData, eddsaTestPartyNum)
	errChs := make(chan *tss.Error, eddsaTestPartyNum)

	// 4. Create and start local parties
	parties := make([]tss.Party, eddsaTestPartyNum)
	for i := 0; i < eddsaTestPartyNum; i++ {
		parties[i] = eddsaKeygen.NewLocalParty(params[i], outChs, endChs)
	}

	// 5. Start all parties in goroutines
	var wg sync.WaitGroup
	for i := 0; i < eddsaTestPartyNum; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			if err := parties[idx].Start(); err != nil {
				errChs <- err
			}
		}(i)
	}

	// 6. Message routing
	// This goroutine listens for messages from any party and routes them to the correct recipient(s)
	go func() {
		for {
			select {
			case msg := <-outChs:
				dest := msg.GetTo()
				if dest == nil { // broadcast message
					for _, p := range parties {
						if p.PartyID().Index == msg.GetFrom().Index {
							continue
						}
						go func(p tss.Party) {
							wireBytes, _, err := msg.WireBytes()
							if err != nil {
								errChs <- tss.NewError(err, "failed to get wire bytes", 0, msg.GetFrom())
								return
							}
							if _, err := p.UpdateFromBytes(wireBytes, msg.GetFrom(), true); err != nil {
								errChs <- err
							}
						}(p)
					}
				} else { // unicast message
					for _, p := range parties {
						if p.PartyID().Index == dest[0].Index {
							go func(p tss.Party) {
								wireBytes, _, err := msg.WireBytes()
								if err != nil {
									errChs <- tss.NewError(err, "failed to get wire bytes", 0, msg.GetFrom())
									return
								}
								if _, err := p.UpdateFromBytes(wireBytes, msg.GetFrom(), false); err != nil {
									errChs <- err
								}
							}(p)
							break
						}
					}
				}
			case err := <-errChs:
				fmt.Printf("Error: %s\n", err.Error())
				os.Exit(1)
			}
		}
	}()

	// 7. Wait for all parties to finish and collect save data
	var saveDatas []*eddsaKeygen.LocalPartySaveData
	for i := 0; i < eddsaTestPartyNum; i++ {
		saveData := <-endChs
		saveDatas = append(saveDatas, saveData)
	}
	wg.Wait() // Ensure all party goroutines have finished

	fmt.Println("EdDSA Key generation completed successfully!")

	// 8. Verify the generated public key
	// All parties should have the same public key
	firstPubKey := saveDatas[0].EDDSAPub
	for i := 1; i < eddsaTestPartyNum; i++ {
		if firstPubKey.X().Cmp(saveDatas[i].EDDSAPub.X()) != 0 || firstPubKey.Y().Cmp(saveDatas[i].EDDSAPub.Y()) != 0 {
			fmt.Printf("Error: Public keys do not match for party %d\n", i)
			os.Exit(1)
		}
	}
	fmt.Printf("Generated EdDSA Public Key: (%s, %s)\n", firstPubKey.X().String(), firstPubKey.Y().String())

	// Example of how to use the generated key shares (for signing, not part of keygen)
	// This part is just for demonstration and would typically be in a separate signing process
	fmt.Println("\nDemonstrating EdDSA key share usage (for signing):")
	// For a real signing process, you would use the `eddsa/signing` package
	// Here, we'll just show that each party has a share of the private key
	for i, saveData := range saveDatas {
		fmt.Printf("Party %d private share: %s\n", i, saveData.Xi.String())
	}
	fmt.Println("--- Finished EdDSA Key Generation Example ---")
}

func main() {
	runECDSAKeygenExample()
	runEdDSAKeygenExample()
}
