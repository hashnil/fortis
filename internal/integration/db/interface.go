package db

import (
	"fortis/internal/integration/db/models"
)

type Client interface {
	// Fetch requests
	FindUserByID(userID string) (models.User, error)
	FindWalletByNameAndNetwork(username, provider, network string) (models.Wallet, error)
	GetInflightTransaction(challenge string) (models.InflightTransaction, error)

	// Create requests
	CreateUser(models.User) error
	CreateWallet(models.Wallet) error
	CreateInflightTransaction(models.InflightTransaction) error
	CreateTransactionLog(models.TransactionLog) error

	// Update requests
	UpdateUser(models.User) error
}
