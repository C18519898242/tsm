package tss

import (
	"fmt"
	"math/big"
	"sort"
	"sync"
	"tsm/internal/keystore"

	"github.com/bnb-chain/tss-lib/v2/common"
	"github.com/bnb-chain/tss-lib/v2/crypto"
	ecdsaKeygen "github.com/bnb-chain/tss-lib/v2/ecdsa/keygen"
	ecdsaSigning "github.com/bnb-chain/tss-lib/v2/ecdsa/signing"
	eddsaKeygen "github.com/bnb-chain/tss-lib/v2/eddsa/keygen"
	eddsaSigning "github.com/bnb-chain/tss-lib/v2/eddsa/signing"
	"github.com/bnb-chain/tss-lib/v2/tss"
)

// SignECDSA performs a threshold signature for ECDSA.
func SignECDSA(keyID string, message string) (*common.SignatureData, error) {
	fmt.Println("Loading ECDSA keys...")
	keys, err := keystore.LoadAllECDSAKeys(keyID)
	if err != nil {
		return nil, fmt.Errorf("error loading ECDSA keys: %w", err)
	}

	partyNum := len(keys)
	if partyNum < 2 {
		return nil, fmt.Errorf("not enough keys to perform signing")
	}
	threshold := testThreshold

	// 1. Create and prepare the signing parties
	partyIDs := tss.GenerateTestPartyIDs(partyNum)
	// The keys need to be sorted by ShareID to match the party IDs
	sort.Slice(keys, func(i, j int) bool {
		return keys[i].ShareID.Cmp(keys[j].ShareID) < 0
	})
	for i := 0; i < partyNum; i++ {
		partyIDs[i].Key = keys[i].ShareID.Bytes()
	}
	sort.Sort(partyIDs)

	// 2. Create the peer context
	p2pCtx := tss.NewPeerContext(partyIDs)

	// 3. Create a map from ShareID to key data for easy lookup
	keyMap := make(map[string]ecdsaKeygen.LocalPartySaveData)
	for _, key := range keys {
		keyMap[key.ShareID.String()] = key
	}

	// 4. Create and start the signing parties
	parties := make([]tss.Party, partyNum)
	errCh := make(chan *tss.Error, partyNum)
	outCh := make(chan tss.Message, partyNum)
	endCh := make(chan *common.SignatureData, partyNum)
	msg := new(big.Int).SetBytes([]byte(message))

	for i := 0; i < partyNum; i++ {
		pID := partyIDs[i]
		params := tss.NewParameters(tss.S256(), p2pCtx, pID, partyNum, threshold)
		keyData := keyMap[pID.KeyInt().String()]
		parties[i] = ecdsaSigning.NewLocalParty(msg, params, keyData, outCh, endCh)
	}

	runSigning(parties, errCh, outCh)

	select {
	case sig := <-endCh:
		fmt.Println("ECDSA Signature generated!")
		return sig, nil
	case err := <-errCh:
		return nil, fmt.Errorf("signing error: %w", err)
	}
}

// SignEdDSA performs a threshold signature for EdDSA.
func SignEdDSA(keyID string, message string) (*common.SignatureData, *crypto.ECPoint, error) {
	fmt.Println("Loading EdDSA keys...")
	keys, err := keystore.LoadAllEdDSAKeys(keyID)
	if err != nil {
		return nil, nil, fmt.Errorf("error loading EdDSA keys: %w", err)
	}

	partyNum := len(keys)
	if partyNum < 2 {
		return nil, nil, fmt.Errorf("not enough keys to perform signing")
	}
	pubKey := keys[0].EDDSAPub
	threshold := testThreshold

	// 1. Create and prepare the signing parties
	partyIDs := tss.GenerateTestPartyIDs(partyNum)
	// The keys need to be sorted by ShareID to match the party IDs
	sort.Slice(keys, func(i, j int) bool {
		return keys[i].ShareID.Cmp(keys[j].ShareID) < 0
	})
	for i := 0; i < partyNum; i++ {
		partyIDs[i].Key = keys[i].ShareID.Bytes()
	}
	sort.Sort(partyIDs)

	// 2. Create the peer context
	p2pCtx := tss.NewPeerContext(partyIDs)

	// 3. Create a map from ShareID to key data for easy lookup
	keyMap := make(map[string]eddsaKeygen.LocalPartySaveData)
	for _, key := range keys {
		keyMap[key.ShareID.String()] = key
	}

	// 4. Create and start the signing parties
	parties := make([]tss.Party, partyNum)
	errCh := make(chan *tss.Error, partyNum)
	outCh := make(chan tss.Message, partyNum)
	endCh := make(chan *common.SignatureData, partyNum)
	msg := new(big.Int).SetBytes([]byte(message))

	for i := 0; i < partyNum; i++ {
		pID := partyIDs[i]
		params := tss.NewParameters(tss.Edwards(), p2pCtx, pID, partyNum, threshold)
		keyData := keyMap[pID.KeyInt().String()]
		parties[i] = eddsaSigning.NewLocalParty(msg, params, keyData, outCh, endCh)
	}

	runSigning(parties, errCh, outCh)

	select {
	case sig := <-endCh:
		fmt.Println("EdDSA Signature generated!")
		return sig, pubKey, nil
	case err := <-errCh:
		return nil, nil, fmt.Errorf("signing error: %w", err)
	}
}

func runSigning(parties []tss.Party, errCh chan *tss.Error, outCh chan tss.Message) {
	var wg sync.WaitGroup
	for i := 0; i < len(parties); i++ {
		wg.Add(1)
		go func(p tss.Party) {
			defer wg.Done()
			if err := p.Start(); err != nil {
				errCh <- err
			}
		}(parties[i])
	}

	// Message routing
	go func() {
		for {
			select {
			case msg := <-outCh:
				dest := msg.GetTo()
				if dest == nil {
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
				} else {
					for _, p := range parties {
						if p.PartyID().Index == dest[0].Index {
							go func(p tss.Party) {
								wireBytes, _, _ := msg.WireBytes()
								if _, err := p.UpdateFromBytes(wireBytes, msg.GetFrom(), msg.IsBroadcast()); err != nil {
									errCh <- err
								}
							}(p)
							break
						}
					}
				}
			case <-errCh:
				return
			}
		}
	}()
	wg.Wait()
}
