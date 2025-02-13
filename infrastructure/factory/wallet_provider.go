package factory

import (
	"fmt"
	"fortis/entity/constants"
	"fortis/internal/integration/db"
	"fortis/internal/integration/wallet"
	"fortis/internal/integration/wallet/dfns"
)

// Wallet Provider factory
func InitWalletProvider(provider string, dbClient db.Client) (wallet.Provider, error) {
	switch provider {
	case constants.DFNS:
		// Initialize DFNS client for wallet operations.
		return dfns.NewDFNSWalletProvider(dbClient), nil
	default:
		return nil, fmt.Errorf("unsupported wallet-provider type: %s", provider)
	}
}
