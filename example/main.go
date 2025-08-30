package main

import (
	"fmt"
	"os"
	"sync"

	"github.com/bnb-chain/tss-lib/v2/ecdsa/keygen"
	"github.com/bnb-chain/tss-lib/v2/tss"
)

const (
	// T and N values for the key generation
	// T is the threshold, N is the total number of parties
	// T+1 parties are required to reconstruct the key
	testThreshold = 2
	testPartyNum  = 3
)

func main() {
	// 1. Create Party IDs
	partyIDs := tss.GenerateTestPartyIDs(testPartyNum)
	p2pCtx := tss.NewPeerContext(partyIDs)

	// 2. Create Parameters for each party
	params := make([]*tss.Parameters, testPartyNum)
	for i := 0; i < testPartyNum; i++ {
		params[i] = tss.NewParameters(tss.S256(), p2pCtx, partyIDs[i], testPartyNum, testThreshold)
	}

	// 3. Create channels for inter-party communication
	// Each party needs an outbound channel for messages and an inbound channel for messages from other parties
	// Also, an end channel to signal completion and receive the generated key data
	outChs := make(chan tss.Message, testPartyNum)
	endChs := make(chan *keygen.LocalPartySaveData, testPartyNum)
	errChs := make(chan *tss.Error, testPartyNum)

	// 4. Create and start local parties
	parties := make([]tss.Party, testPartyNum)
	for i := 0; i < testPartyNum; i++ {
		parties[i] = keygen.NewLocalParty(params[i], outChs, endChs)
	}

	// 5. Start all parties in goroutines
	var wg sync.WaitGroup
	for i := 0; i < testPartyNum; i++ {
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
	for i := 0; i < testPartyNum; i++ {
		saveData := <-endChs
		saveDatas = append(saveDatas, saveData)
	}
	wg.Wait() // Ensure all party goroutines have finished

	fmt.Println("Key generation completed successfully!")

	// 8. Verify the generated public key
	// All parties should have the same public key
	firstPubKey := saveDatas[0].ECDSAPub
	for i := 1; i < testPartyNum; i++ {
		if firstPubKey.X().Cmp(saveDatas[i].ECDSAPub.X()) != 0 || firstPubKey.Y().Cmp(saveDatas[i].ECDSAPub.Y()) != 0 {
			fmt.Printf("Error: Public keys do not match for party %d\n", i)
			os.Exit(1)
		}
	}
	fmt.Printf("Generated Public Key: (%s, %s)\n", firstPubKey.X().String(), firstPubKey.Y().String())

	// Example of how to use the generated key shares (for signing, not part of keygen)
	// This part is just for demonstration and would typically be in a separate signing process
	fmt.Println("\nDemonstrating key share usage (for signing):")
	// For a real signing process, you would use the `ecdsa/signing` package
	// Here, we'll just show that each party has a share of the private key
	for i, saveData := range saveDatas {
		fmt.Printf("Party %d private share: %s\n", i, saveData.Xi.String())
	}
}
