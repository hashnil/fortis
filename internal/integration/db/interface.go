package db

import (
	"fortis/internal/integration/db/models"
)

type Client interface {
	FindUserWallet(userID string) (models.Wallet, error)
	FindWalletByNetwork(userID, network string) (models.Wallet, error)
	CreateWallet(models.Wallet) error
}
