package main

import (
	"fmt"
)

type WalletResponse struct {
	ID          string `json:"id"`
	Network     string `json:"network"`
	Status      string `json:"status"`
	Name        string `json:"name"`
	Address     string `json:"address"`
	DateCreated string `json:"dateCreated"`
	SigningKey  struct {
		Curve     string `json:"curve"`
		Scheme    string `json:"scheme"`
		PublicKey string `json:"publicKey"`
	} `json:"signingKey"`
}

func main() {

	// Create wallet
	walletData := map[string]interface{}{
		"network":         "EthereumGoerli",
		"name":            "my-wallet",
		"delayDelegation": true,
	}

	fmt.Printf("Wallet Created: %+v\n", walletData)
}
