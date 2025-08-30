package tss

import (
	"fmt"
	"os"
	"sync"

	"tsm/internal/keystore"

	ecdsaKeygen "github.com/bnb-chain/tss-lib/v2/ecdsa/keygen"
	eddsaKeygen "github.com/bnb-chain/tss-lib/v2/eddsa/keygen"
	"github.com/bnb-chain/tss-lib/v2/tss"
)

const (
	testThreshold = 2
	testPartyNum  = 3
)

func GenerateECDSAKeys(keyID string) {
	fmt.Println("--- Starting ECDSA Key Generation ---")
	partyIDs := tss.GenerateTestPartyIDs(testPartyNum)
	p2pCtx := tss.NewPeerContext(partyIDs)

	params := make([]*tss.Parameters, testPartyNum)
	for i := 0; i < testPartyNum; i++ {
		params[i] = tss.NewParameters(tss.S256(), p2pCtx, partyIDs[i], testPartyNum, testThreshold)
	}

	outChs := make(chan tss.Message, testPartyNum)
	endChs := make(chan *ecdsaKeygen.LocalPartySaveData, testPartyNum)
	errChs := make(chan *tss.Error, testPartyNum)

	parties := make([]tss.Party, testPartyNum)
	for i := 0; i < testPartyNum; i++ {
		parties[i] = ecdsaKeygen.NewLocalParty(params[i], outChs, endChs)
	}

	runKeygen(parties, errChs, outChs)

	var saveDatas []*ecdsaKeygen.LocalPartySaveData
	for i := 0; i < testPartyNum; i++ {
		saveData := <-endChs
		saveDatas = append(saveDatas, saveData)
	}

	fmt.Println("ECDSA Key generation completed successfully!")
	// Additional validation and logging can be added here if needed

	for i, saveData := range saveDatas {
		if err := keystore.SaveECDSAKey(keyID, i, saveData); err != nil {
			fmt.Printf("Error saving ECDSA key for party %d: %v\n", i, err)
			os.Exit(1)
		}
		fmt.Printf("ECDSA key for party %d saved.\n", i)
	}
	fmt.Println("--- Finished ECDSA Key Generation ---\n")
}

func GenerateEdDSAKeys(keyID string) {
	fmt.Println("--- Starting EdDSA Key Generation ---")
	partyIDs := tss.GenerateTestPartyIDs(testPartyNum)
	p2pCtx := tss.NewPeerContext(partyIDs)

	params := make([]*tss.Parameters, testPartyNum)
	for i := 0; i < testPartyNum; i++ {
		params[i] = tss.NewParameters(tss.Edwards(), p2pCtx, partyIDs[i], testPartyNum, testThreshold)
	}

	outChs := make(chan tss.Message, testPartyNum)
	endChs := make(chan *eddsaKeygen.LocalPartySaveData, testPartyNum)
	errChs := make(chan *tss.Error, testPartyNum)

	parties := make([]tss.Party, testPartyNum)
	for i := 0; i < testPartyNum; i++ {
		parties[i] = eddsaKeygen.NewLocalParty(params[i], outChs, endChs)
	}

	runKeygen(parties, errChs, outChs)

	var saveDatas []*eddsaKeygen.LocalPartySaveData
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
						go func(p tss.Party) {
							wireBytes, _, _ := msg.WireBytes()
							p.UpdateFromBytes(wireBytes, msg.GetFrom(), true)
						}(p)
					}
				} else {
					for _, p := range parties {
						if p.PartyID().Index == dest[0].Index {
							go func(p tss.Party) {
								wireBytes, _, _ := msg.WireBytes()
								p.UpdateFromBytes(wireBytes, msg.GetFrom(), false)
							}(p)
							break
						}
					}
				}
			case err := <-errChs:
				fmt.Printf("Error: %s\n", err.Error())
				return // Exit goroutine on error
			}
		}
	}()
	wg.Wait()
}
