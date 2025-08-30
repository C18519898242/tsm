package tss

import (
	"fmt"
	"os"
	"sync"

	"tsm/internal/keystore"

	"github.com/bnb-chain/tss-lib/v2/ecdsa/keygen"
	eddsa_keygen "github.com/bnb-chain/tss-lib/v2/eddsa/keygen"
	"github.com/bnb-chain/tss-lib/v2/tss"
)

const (
	testThreshold = 2
	testPartyNum  = 3
)

// GenerateECDSAKeys generates ECDSA keys using a known-good pattern from the tss-lib examples.
func GenerateECDSAKeys(keyID string) {
	fmt.Println("--- Starting ECDSA Key Generation ---")

	// 1. Load or generate pre-parameters
	preParams, err := loadOrGeneratePreParams(testPartyNum)
	if err != nil {
		fmt.Printf("Error with pre-parameters: %v\n", err)
		os.Exit(1)
	}

	// 2. Create communication channels
	outCh := make(chan tss.Message, testPartyNum*testPartyNum)
	endCh := make(chan *keygen.LocalPartySaveData, testPartyNum)
	errCh := make(chan *tss.Error, testPartyNum)

	// 3. Create and start parties
	parties := make([]tss.Party, 0, testPartyNum)
	pIDs := tss.GenerateTestPartyIDs(testPartyNum) // Use the library's tested party ID generator
	p2pCtx := tss.NewPeerContext(pIDs)

	for i := 0; i < testPartyNum; i++ {
		params := tss.NewParameters(tss.S256(), p2pCtx, pIDs[i], testPartyNum, testThreshold)
		P := keygen.NewLocalParty(params, outCh, endCh, *preParams[i]).(*keygen.LocalParty)
		parties = append(parties, P)
	}

	var wg sync.WaitGroup
	for _, p := range parties {
		wg.Add(1)
		go func(p tss.Party) {
			defer wg.Done()
			if err := p.Start(); err != nil {
				errCh <- err
			}
		}(p)
	}

	// 4. Message routing
	go func() {
		for {
			select {
			case err := <-errCh:
				fmt.Fprintf(os.Stderr, "Error from party: %s\n", err)
				return
			case msg := <-outCh:
				dest := msg.GetTo()
				if dest == nil { // broadcast
					for _, p := range parties {
						if p.PartyID().Index == msg.GetFrom().Index {
							continue
						}
						go func(p tss.Party) {
							wireBytes, _, _ := msg.WireBytes()
							if _, err := p.UpdateFromBytes(wireBytes, msg.GetFrom(), msg.IsBroadcast()); err != nil {
								errCh <- err
							}
						}(p)
					}
				} else { // point-to-point
					if dest[0].Index == msg.GetFrom().Index {
						fmt.Fprintf(os.Stderr, "party %d tried to send a message to itself\n", dest[0].Index)
						continue
					}
					go func(p tss.Party) {
						wireBytes, _, _ := msg.WireBytes()
						if _, err := p.UpdateFromBytes(wireBytes, msg.GetFrom(), msg.IsBroadcast()); err != nil {
							errCh <- err
						}
					}(parties[dest[0].Index])
				}
			}
		}
	}()

	wg.Wait()

	// 5. Collect save data
	saveDatas := make([]*keygen.LocalPartySaveData, 0, testPartyNum)
	for i := 0; i < testPartyNum; i++ {
		saveDatas = append(saveDatas, <-endCh)
	}

	fmt.Println("ECDSA Key generation completed successfully!")

	// 6. Save keys
	for i, saveData := range saveDatas {
		// The index from the save data might not be sequential, but for this test setup it will be.
		// A more robust implementation would map party IDs to save data.
		if err := keystore.SaveECDSAKey(keyID, i, saveData); err != nil {
			fmt.Printf("Error saving ECDSA key for party %d: %v\n", i, err)
			os.Exit(1)
		}
		fmt.Printf("ECDSA key for party %d saved.\n", i)
	}
	fmt.Println("--- Finished ECDSA Key Generation ---\n")
}

// loadOrGeneratePreParams attempts to load pre-computed parameters from disk,
// or generates and saves them if they cannot be found.
func loadOrGeneratePreParams(partyCount int) ([]*keygen.LocalPreParams, error) {
	preParams := make([]*keygen.LocalPreParams, partyCount)
	var err error
	for i := 0; i < partyCount; i++ {
		preParams[i], err = LoadPreParams(i)
		if err != nil {
			fmt.Println("Could not load pre-parameters, generating new ones...")
			return GenerateAndSavePreParams(partyCount)
		}
	}
	fmt.Println("Loaded existing pre-parameters.")
	return preParams, nil
}

// GenerateEdDSAKeys is left as-is for now, focusing on fixing ECDSA.
func GenerateEdDSAKeys(keyID string) {
	fmt.Println("--- Starting EdDSA Key Generation ---")
	partyIDs := tss.GenerateTestPartyIDs(testPartyNum)
	p2pCtx := tss.NewPeerContext(partyIDs)
	params := make([]*tss.Parameters, testPartyNum)
	for i := 0; i < testPartyNum; i++ {
		params[i] = tss.NewParameters(tss.Edwards(), p2pCtx, partyIDs[i], testPartyNum, testThreshold)
	}

	outChs := make(chan tss.Message, testPartyNum)
	endChs := make(chan *eddsa_keygen.LocalPartySaveData, testPartyNum)
	errChs := make(chan *tss.Error, testPartyNum)

	parties := make([]tss.Party, testPartyNum)
	for i := 0; i < testPartyNum; i++ {
		parties[i] = eddsa_keygen.NewLocalParty(params[i], outChs, endChs)
	}

	// This old runKeygen is problematic, but we leave it for now.
	runKeygen(parties, errChs, outChs)

	var saveDatas []*eddsa_keygen.LocalPartySaveData
	for i := 0; i < testPartyNum; i++ {
		saveData := <-endChs
		saveDatas = append(saveDatas, saveData)
	}

	fmt.Println("EdDSA Key generation completed successfully!")

	for i, saveData := range saveDatas {
		if err := keystore.SaveEdDSAKey(keyID, i, saveData); err != nil {
			fmt.Printf("Error saving EdDSA key for party %d: %v\n", i, err)
			os.Exit(1)
		}
		fmt.Printf("EdDSA key for party %d saved.\n", i)
	}
	fmt.Println("--- Finished EdDSA Key Generation ---")
}

// runKeygen is the old, problematic implementation. Kept for GenerateEdDSAKeys.
func runKeygen(parties []tss.Party, errChs chan *tss.Error, outChs chan tss.Message) {
	var wg sync.WaitGroup
	for i := 0; i < len(parties); i++ {
		wg.Add(1)
		go func(pIdx int) {
			defer wg.Done()
			if err := parties[pIdx].Start(); err != nil {
				errChs <- err
			}
		}(i)
	}

	go func() {
		for {
			select {
			case msg := <-outChs:
				dest := msg.GetTo()
				if dest == nil {
					for _, p := range parties {
						if p.PartyID().Index == msg.GetFrom().Index {
							continue
						}
						wireBytes, _, _ := msg.WireBytes()
						go p.UpdateFromBytes(wireBytes, msg.GetFrom(), true)
					}
				} else {
					for _, p := range parties {
						if p.PartyID().Index == dest[0].Index {
							wireBytes, _, _ := msg.WireBytes()
							go p.UpdateFromBytes(wireBytes, msg.GetFrom(), false)
							break
						}
					}
				}
			}
		}
	}()
	wg.Wait()
}
