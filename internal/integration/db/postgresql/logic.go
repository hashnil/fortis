package postgresql

import (
	"fmt"
	"fortis/internal/integration/db/models"
)

// FindUserByID fetches a user by their unique ID.
func (db *PostgresSQLClient) FindUserByID(userID string) (models.User, error) {
	var user models.User
	err := db.client.First(&user, "id = ?", userID).Error
	return user, err
}

// CreateUser adds a new user record to the database.
func (db *PostgresSQLClient) CreateUser(user models.User) error {
	return db.client.Create(&user).Error
}

// UpdateUser updates an existing user record in the database.
func (db *PostgresSQLClient) UpdateUser(user models.User) error {
	return db.client.Model(&models.User{}).Where("id = ?", user.ID).Updates(user).Error
}

// FindWalletByNameAndNetwork retrieves a wallet based on username, provider, and network.
func (db *PostgresSQLClient) FindWalletByNameAndNetwork(username, provider, network string) (models.Wallet, error) {
	var wallet models.Wallet
	err := db.client.First(&wallet, "username = ? AND provider = ? AND network = ?", username, provider, network).Error
	return wallet, err
}

// CreateWallet inserts a new wallet record.
func (db *PostgresSQLClient) CreateWallet(wallet models.Wallet) error {
	return db.client.Create(&wallet).Error
}

// CreateInflightTransaction inserts a inflight transaction record.
func (db *PostgresSQLClient) CreateInflightTransaction(inflightTransaction models.InflightTransaction) error {
	return db.client.Create(&inflightTransaction).Error
}

// GetInflightTransaction fetches an inflight transaction by their unique challenge identifier.
func (db *PostgresSQLClient) GetInflightTransaction(challenge string) (models.InflightTransaction, error) {
	var inflightTransaction models.InflightTransaction
	err := db.client.First(&inflightTransaction, "challenge = ?", challenge).Error
	return inflightTransaction, err
}

// DeleteInflightTransaction deletes an inflight transaction record.
func (db *PostgresSQLClient) DeleteInflightTransaction(challenge string) error {
	result := db.client.Where("challenge = ?", challenge).Delete(&models.InflightTransaction{})
	if result.Error != nil {
		return fmt.Errorf("error deleting inflight transaction: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("no inflight transaction found for challenge: %s", challenge)
	}
	return nil
}

// CreateTransaction stores a transaction entry.
func (db *PostgresSQLClient) CreateTransaction(transaction models.Transaction) error {
	return db.client.Create(&transaction).Error
}

// TODO: deleted_at check
// GetTransaction fetches a transaction by their unique transaction hash.
func (db *PostgresSQLClient) GetTransaction(txHash string) (models.Transaction, error) {
	var transaction models.Transaction
	err := db.client.First(&transaction, "tx_hash = ?", txHash).Error
	return transaction, err
}

// DeleteTransaction deletes a transaction record.
func (db *PostgresSQLClient) DeleteTransaction(txHash string) error {
	result := db.client.Where("tx_hash = ?", txHash).Delete(&models.Transaction{})
	if result.Error != nil {
		return fmt.Errorf("error deleting transaction: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("no transaction found for txHash: %s", txHash)
	}
	return nil
}

// CreateTransactionLog stores a transaction log entry.
func (db *PostgresSQLClient) CreateTransactionLog(transactionLog models.TransactionLog) error {
	return db.client.Create(&transactionLog).Error
}
