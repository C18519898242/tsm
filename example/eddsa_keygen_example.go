package main

import (
	"fmt"
	"os"
	"sync"

	eddsaKeygen "github.com/bnb-chain/tss-lib/v2/eddsa/keygen" // Alias for eddsa keygen
	"github.com/bnb-chain/tss-lib/v2/tss"
)

const (
	eddsaTestThreshold = 2
	eddsaTestPartyNum  = 3
)

func RunEdDSAKeygenExample() {
	fmt.Println("--- Starting EdDSA Key Generation Example ---")
	partyIDs := tss.GenerateTestPartyIDs(eddsaTestPartyNum)
	p2pCtx := tss.NewPeerContext(partyIDs)

	params := make([]*tss.Parameters, eddsaTestPartyNum)
	for i := 0; i < eddsaTestPartyNum; i++ {
		params[i] = tss.NewParameters(tss.Edwards(), p2pCtx, partyIDs[i], eddsaTestPartyNum, eddsaTestThreshold)
	}

	outChs := make(chan tss.Message, eddsaTestPartyNum)
	endChs := make(chan *eddsaKeygen.LocalPartySaveData, eddsaTestPartyNum)
	errChs := make(chan *tss.Error, eddsaTestPartyNum)

	parties := make([]tss.Party, eddsaTestPartyNum)
	for i := 0; i < eddsaTestPartyNum; i++ {
		parties[i] = eddsaKeygen.NewLocalParty(params[i], outChs, endChs)
	}

	var wg sync.WaitGroup
	for i := 0; i < eddsaTestPartyNum; i++ {
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
				} else {
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

	var saveDatas []*eddsaKeygen.LocalPartySaveData
	for i := 0; i < eddsaTestPartyNum; i++ {
		saveData := <-endChs
		saveDatas = append(saveDatas, saveData)
	}
	wg.Wait()

	fmt.Println("EdDSA Key generation completed successfully!")

	firstPubKey := saveDatas[0].EDDSAPub
	for i := 1; i < eddsaTestPartyNum; i++ {
		if firstPubKey.X().Cmp(saveDatas[i].EDDSAPub.X()) != 0 || firstPubKey.Y().Cmp(saveDatas[i].EDDSAPub.Y()) != 0 {
			fmt.Printf("Error: Public keys do not match for party %d\n", i)
			os.Exit(1)
		}
	}
	fmt.Printf("Generated EdDSA Public Key: (%s, %s)\n", firstPubKey.X().String(), firstPubKey.Y().String())

	fmt.Println("\nDemonstrating EdDSA key share usage (for signing):")
	for i, saveData := range saveDatas {
		fmt.Printf("Party %d private share: %s\n", i, saveData.Xi.String())
	}
	fmt.Println("--- Finished EdDSA Key Generation Example ---")

	// Save the generated keys
	for i, saveData := range saveDatas {
		if err := SaveEdDSAKey(i, saveData); err != nil {
			fmt.Printf("Error saving EdDSA key for party %d: %v\n", i, err)
			os.Exit(1)
		}
		fmt.Printf("EdDSA key for party %d saved to %s/%s%d%s\n", i, keysDir, eddsaKeyFilePrefix, i, keyFileSuffix)
	}
}
