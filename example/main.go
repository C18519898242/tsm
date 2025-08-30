package main

import (
	"fmt"
	"os"
)

func main() {
	fmt.Println("Select an operation:")
	fmt.Println("1. Generate new keys (ECDSA or EdDSA)")
	fmt.Println("2. Load existing keys (ECDSA or EdDSA)")
	fmt.Print("Enter your choice (1 or 2): ")

	var mainChoice int
	_, err := fmt.Scanf("%d\n", &mainChoice) // Consume the newline
	if err != nil {
		fmt.Printf("Invalid input: %v\n", err)
		os.Exit(1)
	}

	switch mainChoice {
	case 1:
		fmt.Println("\nSelect a key generation example to run:")
		fmt.Println("1. ECDSA Key Generation")
		fmt.Println("2. EdDSA Key Generation")
		fmt.Print("Enter your choice (1 or 2): ")

		var genChoice int
		_, err := fmt.Scanf("%d\n", &genChoice) // Consume the newline
		if err != nil {
			fmt.Printf("Invalid input: %v\n", err)
			os.Exit(1)
		}

		switch genChoice {
		case 1:
			RunECDSAKeygenExample()
		case 2:
			RunEdDSAKeygenExample()
		default:
			fmt.Println("Invalid choice. Please enter 1 or 2.")
			os.Exit(1)
		}
	case 2:
		fmt.Println("\nSelect key type to load:")
		fmt.Println("1. ECDSA Keys")
		fmt.Println("2. EdDSA Keys")
		fmt.Print("Enter your choice (1 or 2): ")

		var loadChoice int
		_, err := fmt.Scanf("%d\n", &loadChoice) // Consume the newline
		if err != nil {
			fmt.Printf("Invalid input: %v\n", err)
			os.Exit(1)
		}

		fmt.Print("Enter party index to load (e.g., 0, 1, 2): ")
		var partyIndex int
		_, err = fmt.Scanf("%d\n", &partyIndex) // Consume the newline
		if err != nil {
			fmt.Printf("Invalid input for party index: %v\n", err)
			os.Exit(1)
		}

		switch loadChoice {
		case 1:
			data, err := LoadECDSAKey(partyIndex)
			if err != nil {
				fmt.Printf("Error loading ECDSA key for party %d: %v\n", partyIndex, err)
				os.Exit(1)
			}
			fmt.Printf("Successfully loaded ECDSA key for party %d. Public Key: (%s, %s)\n", partyIndex, data.ECDSAPub.X().String(), data.ECDSAPub.Y().String())
			fmt.Printf("Private share: %s\n", data.Xi.String())
		case 2:
			data, err := LoadEdDSAKey(partyIndex)
			if err != nil {
				fmt.Printf("Error loading EdDSA key for party %d: %v\n", partyIndex, err)
				os.Exit(1)
			}
			fmt.Printf("Successfully loaded EdDSA key for party %d. Public Key: (%s, %s)\n", partyIndex, data.EDDSAPub.X().String(), data.EDDSAPub.Y().String())
			fmt.Printf("Private share: %s\n", data.Xi.String())
		default:
			fmt.Println("Invalid choice. Please enter 1 or 2.")
			os.Exit(1)
		}
	default:
		fmt.Println("Invalid choice. Please enter 1 or 2.")
		os.Exit(1)
	}
}
