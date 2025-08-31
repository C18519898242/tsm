package main

import (
	"bufio"
	"fmt"
	"math/big"
	"os"
	"strings"
	"tsm/internal/tss"

	"github.com/bnb-chain/tss-lib/v2/common"
	"github.com/bnb-chain/tss-lib/v2/crypto"
	"github.com/decred/dcrd/dcrec/edwards/v2"
)

// ecPointToEncodedBytes converts an EC point to the 32-byte compressed format.
func ecPointToEncodedBytes(x *big.Int, y *big.Int) *[32]byte {
	pk := edwards.NewPublicKey(x, y)
	serialized := pk.Serialize()
	var s [32]byte
	copy(s[:], serialized)
	return &s
}

func main() {
	fmt.Println("Application starting...")
	for {
		fmt.Println("\nSelect an operation:")
		fmt.Println("1. Generate new keys")
		fmt.Println("2. Sign a message")
		fmt.Println("3. Exit")
		fmt.Print("Enter your choice (1-3): ")

		var mainChoice int
		_, err := fmt.Scanf("%d\n", &mainChoice) // Consume the newline
		if err != nil {
			fmt.Println("Invalid input. Please enter a number.")
			// Clear the buffer in case of non-integer input that Scanf leaves behind
			reader := bufio.NewReader(os.Stdin)
			reader.ReadString('\n')
			continue
		}

		switch mainChoice {
		case 1:
			handleGenerateKeys()
		case 2:
			handleSignMessage()
		case 3:
			fmt.Println("Exiting...")
			os.Exit(0)
		default:
			fmt.Println("Invalid choice. Please enter a number between 1 and 3.")
		}
	}
}

func handleGenerateKeys() {
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Enter a KeyID for the new keys: ")
	keyID, err := reader.ReadString('\n')
	if err != nil {
		fmt.Printf("Failed to read KeyID: %v\n", err)
		return
	}
	keyID = strings.TrimSpace(keyID)
	if keyID == "" {
		fmt.Println("KeyID cannot be empty.")
		return
	}

	fmt.Printf("\nGenerating ECDSA and EdDSA keys for KeyID: %s\n", keyID)
	tss.GenerateECDSAKeys(keyID)
	fmt.Println() // Add a newline for better separation
	tss.GenerateEdDSAKeys(keyID)
}

func handleSignMessage() {
	reader := bufio.NewReader(os.Stdin)

	// 1. Select Algorithm
	fmt.Println("\nSelect signature algorithm:")
	fmt.Println("1. ECDSA")
	fmt.Println("2. EdDSA")
	fmt.Print("Enter your choice (1-2): ")
	var algChoice int
	_, err := fmt.Scanf("%d\n", &algChoice)
	if err != nil || (algChoice != 1 && algChoice != 2) {
		fmt.Println("Invalid choice. Please enter 1 for ECDSA or 2 for EdDSA.")
		// Clear the buffer
		reader.ReadString('\n')
		return
	}

	// 2. Get KeyID
	fmt.Print("Enter the KeyID of the keys to use for signing: ")
	keyID, err := reader.ReadString('\n')
	if err != nil {
		fmt.Printf("Failed to read KeyID: %v\n", err)
		return
	}
	keyID = strings.TrimSpace(keyID)
	if keyID == "" {
		fmt.Println("KeyID cannot be empty.")
		return
	}

	// 3. Get Message
	fmt.Print("Enter the message to sign: ")
	message, err := reader.ReadString('\n')
	if err != nil {
		fmt.Printf("Failed to read message: %v\n", err)
		return
	}
	message = strings.TrimSpace(message)
	if message == "" {
		fmt.Println("Message cannot be empty.")
		return
	}

	// 4. Call signing function
	fmt.Printf("\nSigning message with KeyID '%s'...\n", keyID)

	var signature *common.SignatureData
	var pubKey *crypto.ECPoint

	if algChoice == 1 {
		signature, err = tss.SignECDSA(keyID, message)
	} else {
		signature, pubKey, err = tss.SignEdDSA(keyID, message)
	}

	if err != nil {
		fmt.Printf("Error signing message: %v\n", err)
		return
	}
	if signature == nil {
		fmt.Println("Failed to generate signature, but no error was reported.")
		return
	}

	fmt.Printf("Signature generated successfully!\n")
	fmt.Printf("Signature (hex): %x\n", signature.GetSignature())

	// Verification
	if algChoice == 1 {
		fmt.Println("ECDSA signing does not have verification implemented in this CLI.")
	} else if algChoice == 2 {
		pkBytes := ecPointToEncodedBytes(pubKey.X(), pubKey.Y())
		parsedPk, err := edwards.ParsePubKey((*pkBytes)[:])
		if err != nil {
			fmt.Println("Error parsing public key for verification:", err)
			return
		}

		r := new(big.Int).SetBytes(signature.GetR())
		s := new(big.Int).SetBytes(signature.GetS())

		ok := edwards.Verify(parsedPk, []byte(message), r, s)
		if ok {
			fmt.Println("✅ Signature Verified Successfully!")
		} else {
			fmt.Println("❌ Signature Verification Failed!")
		}
	}
}
