package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"tsm/internal/tss"
)

func main() {
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
	var sigBytes []byte
	var errSign error

	switch algChoice {
	case 1:
		signature, err := tss.SignECDSA(keyID, message)
		if err != nil {
			errSign = err
		} else if signature != nil {
			sigBytes = signature.GetSignature()
		}
	case 2:
		signature, err := tss.SignEdDSA(keyID, message)
		if err != nil {
			errSign = err
		} else if signature != nil {
			sigBytes = signature.GetSignature()
		}
	}

	if errSign != nil {
		fmt.Printf("Error signing message: %v\n", errSign)
	} else if sigBytes != nil {
		fmt.Printf("Signature generated successfully!\n")
		fmt.Printf("Signature (hex): %x\n", sigBytes)
	} else {
		fmt.Println("Failed to generate signature, but no error was reported.")
	}
}
