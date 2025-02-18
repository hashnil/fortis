package db

import (
	"fortis/internal/integration/db/models"
)

type Client interface {
	FindUserWallet(userID string) (models.Wallet, error)
	FindWalletByNetwork(userID, provider, network string) (models.Wallet, error)
	CreateWallet(models.Wallet) error
	GetWalletByUsername(username, provider, network string) (models.Wallet, error)
	CreateTransactionLog(models.TransactionLog) error
}
