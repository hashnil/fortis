package utils

import (
	"fmt"
	"fortis/entity/constants"
	"math/big"
	"math/rand"
)

func GenerateUTR() string {
	return fmt.Sprintf("UTR-%d", rand.Int63())
}

// ConvertToSmallestUnit converts a human-readable amount to the smallest unit (string representation)
func ConvertToSmallestUnit(amount string, token string) (string, error) {
	decimals, exists := constants.TokenDecimals[token]
	if !exists {
		return "", fmt.Errorf("unsupported token: %s", token)
	}

	// Convert string amount to big.Float
	amtFloat, success := new(big.Float).SetString(amount)
	if !success {
		return "", fmt.Errorf("invalid amount format: %s", amount)
	}

	// Compute 10^decimals as big.Int
	multiplier := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(decimals)), nil)
	multiplierFloat := new(big.Float).SetInt(multiplier)

	// Multiply amount by 10^decimals
	result := new(big.Float).Mul(amtFloat, multiplierFloat)

	// Convert to big.Int (truncates decimals)
	smallestUnit := new(big.Int)
	result.Int(smallestUnit)

	// Return string representation
	return smallestUnit.String(), nil
}
