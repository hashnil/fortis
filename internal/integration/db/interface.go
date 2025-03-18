package db

import (
	"fortis/internal/integration/db/models"
)

type Client interface {
	// Fetch requests
	FindUserByID(userID string) (models.User, error)
	FindWalletByNameAndNetwork(username, provider, network string) (models.Wallet, error)
	GetInflightTransaction(challenge string) (models.InflightTransaction, error)
	GetTransaction(txHash string) (models.Transaction, error)

	// Create requests
	CreateUser(models.User) error
	CreateWallet(models.Wallet) error
	CreateInflightTransaction(models.InflightTransaction) error
	CreateTransaction(models.Transaction) error
	CreateTransactionLog(models.TransactionLog) error

	// Update requests
	UpdateUser(models.User) error

	// Delete requests
	DeleteInflightTransaction(string) error
	DeleteTransaction(string) error
}
