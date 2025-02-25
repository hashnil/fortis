package postgresql

import (
	"fmt"
	"fortis/internal/integration/db/models"

	"gorm.io/gorm"
)

// FindUserByID retrieves the first user for a given user ID.
func (db *PostgresSQLClient) FindUserByID(userID string) (models.User, error) {
	var user models.User
	err := db.client.Where("id = ?", userID).First(&user).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return models.User{}, nil // Wallet not found
		}
		return models.User{}, fmt.Errorf("failed to find user: %w", err)
	}
	return user, nil
}

// CreateUser inserts a new user record into the database.
func (db *PostgresSQLClient) CreateUser(user models.User) error {
	return db.client.Create(&user).Error
}

// FindUserWallet retrieves the first wallet for a given user ID.
func (db *PostgresSQLClient) FindUserWallet(userID string) (models.Wallet, error) {
	var wallet models.Wallet
	err := db.client.Where("user_id = ?", userID).First(&wallet).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return models.Wallet{}, nil // Wallet not found
		}
		return models.Wallet{}, fmt.Errorf("failed to find user wallet: %w", err)
	}
	return wallet, nil
}

// FindWalletByNetwork retrieves a wallet for a given user and network.
func (db *PostgresSQLClient) FindWalletByNetwork(userID, provider, network string) (models.Wallet, error) {
	var wallet models.Wallet
	err := db.client.Where("user_id = ? AND provider = ? AND network = ?", userID, provider, network).First(&wallet).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return models.Wallet{}, nil // Wallet not found
		}
		return models.Wallet{}, fmt.Errorf("failed to find wallet for network %s: %w", network, err)
	}
	return wallet, nil
}

// CreateWallet inserts a new wallet record into the database.
func (db *PostgresSQLClient) CreateWallet(wallet models.Wallet) error {
	return db.client.Create(&wallet).Error
}

// GetWalletByUsername retrieves the first wallet for a given username, network and provider.
func (db *PostgresSQLClient) GetWalletByUsername(username, provider, network string) (models.Wallet, error) {
	var wallet models.Wallet
	err := db.client.Where("username = ? AND provider = ? AND network = ?", username, provider, network).First(&wallet).Error
	return wallet, err
}

// CreateTransactionLog stores a transaction log entry in the database.
func (db *PostgresSQLClient) CreateTransactionLog(transactionLog models.TransactionLog) error {
	return db.client.Create(&transactionLog).Error
}
