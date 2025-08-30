package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func main() {
	for {
		fmt.Println("\nSelect an operation:")
		fmt.Println("1. Generate new keys")
		fmt.Println("2. Sign (Not implemented)")
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
			fmt.Println("\nSigning functionality is not yet implemented.")
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
	RunECDSAKeygenExample(keyID)
	fmt.Println() // Add a newline for better separation
	RunEdDSAKeygenExample(keyID)
}
