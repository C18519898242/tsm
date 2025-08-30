package main

import (
	"fmt"
	"os"
)

func main() {
	fmt.Println("Select a key generation example to run:")
	fmt.Println("1. ECDSA Key Generation")
	fmt.Println("2. EdDSA Key Generation")
	fmt.Print("Enter your choice (1 or 2): ")

	var choice int
	_, err := fmt.Scanf("%d", &choice)
	if err != nil {
		fmt.Printf("Invalid input: %v\n", err)
		os.Exit(1)
	}

	switch choice {
	case 1:
		RunECDSAKeygenExample()
	case 2:
		RunEdDSAKeygenExample()
	default:
		fmt.Println("Invalid choice. Please enter 1 or 2.")
		os.Exit(1)
	}
}
