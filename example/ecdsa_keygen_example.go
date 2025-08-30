package main

import (
	"fmt"
	"os"
	"sync"

	"github.com/bnb-chain/tss-lib/v2/ecdsa/keygen"
	"github.com/bnb-chain/tss-lib/v2/tss"
)

const (
	ecdsaTestThreshold = 2
	ecdsaTestPartyNum  = 3
)

func RunECDSAKeygenExample() {
	fmt.Println("--- Starting ECDSA Key Generation Example ---")
	partyIDs := tss.GenerateTestPartyIDs(ecdsaTestPartyNum)
	p2pCtx := tss.NewPeerContext(partyIDs)

	params := make([]*tss.Parameters, ecdsaTestPartyNum)
	for i := 0; i < ecdsaTestPartyNum; i++ {
		params[i] = tss.NewParameters(tss.S256(), p2pCtx, partyIDs[i], ecdsaTestPartyNum, ecdsaTestThreshold)
	}

	outChs := make(chan tss.Message, ecdsaTestPartyNum)
	endChs := make(chan *keygen.LocalPartySaveData, ecdsaTestPartyNum)
	errChs := make(chan *tss.Error, ecdsaTestPartyNum)

	parties := make([]tss.Party, ecdsaTestPartyNum)
	for i := 0; i < ecdsaTestPartyNum; i++ {
		parties[i] = keygen.NewLocalParty(params[i], outChs, endChs)
	}

	var wg sync.WaitGroup
	for i := 0; i < ecdsaTestPartyNum; i++ {
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

	var saveDatas []*keygen.LocalPartySaveData
	for i := 0; i < ecdsaTestPartyNum; i++ {
		saveData := <-endChs
		saveDatas = append(saveDatas, saveData)
	}
	wg.Wait()

	fmt.Println("ECDSA Key generation completed successfully!")

	firstPubKey := saveDatas[0].ECDSAPub
	for i := 1; i < ecdsaTestPartyNum; i++ {
		if firstPubKey.X().Cmp(saveDatas[i].ECDSAPub.X()) != 0 || firstPubKey.Y().Cmp(saveDatas[i].ECDSAPub.Y()) != 0 {
			fmt.Printf("Error: Public keys do not match for party %d\n", i)
			os.Exit(1)
		}
	}
	fmt.Printf("Generated ECDSA Public Key: (%s, %s)\n", firstPubKey.X().String(), firstPubKey.Y().String())

	fmt.Println("\nDemonstrating ECDSA key share usage (for signing):")
	for i, saveData := range saveDatas {
		fmt.Printf("Party %d private share: %s\n", i, saveData.Xi.String())
	}
	fmt.Println("--- Finished ECDSA Key Generation Example ---\n")
}
