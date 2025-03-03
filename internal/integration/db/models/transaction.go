package models

import (
	"time"

	"gorm.io/gorm"
)

type InflightTransaction struct {
	Challenge            string         `gorm:"column:challenge;type:varchar(50);primaryKey"` // Unique identifier for the transaction
	URL                  string         `gorm:"column:url;type:varchar(255)"`                 // Target URL for transaction processing
	AuthToken            string         `gorm:"column:auth_token;type:varchar(1024)"`         // User authentication token
	TransferPayload      []byte         `gorm:"column:transfer_payload;type:jsonb"`           // Transaction request in JSONB format
	UserChallengePayload []byte         `gorm:"column:user_challenge_payload;type:jsonb"`     // User's challenge response in JSONB format
	CreatedAt            time.Time      `gorm:"column:created_at;autoCreateTime"`             // Record creation timestamp
	UpdatedAt            time.Time      `gorm:"column:updated_at;autoUpdateTime"`             // Record update timestamp
	DeletedAt            gorm.DeletedAt `gorm:"column:deleted_at;index"`                      // Soft delete timestamp
}

func (InflightTransaction) TableName() string {
	return "inflight_transactions"
}

type TransactionLog struct {
	ID              string    `gorm:"column:id;type:varchar(50);primaryKey"`
	SenderName      string    `gorm:"column:sender_name;type:varchar(50);not null"`
	SenderAddress   string    `gorm:"column:sender_address;type:varchar(255);not null"`
	ReceiverName    string    `gorm:"column:receiver_name;type:varchar(50);not null"`
	ReceiverAddress string    `gorm:"column:receiver_address;type:varchar(255);not null"`
	Amount          string    `gorm:"column:amount;type:varchar(50);not null"`
	Denom           string    `gorm:"column:denom;type:varchar(50);not null"`
	Provider        string    `gorm:"column:provider;type:varchar(50);not null"`
	Network         string    `gorm:"column:network;type:varchar(50);not null"`
	Status          string    `gorm:"column:status;type:varchar(50);not null"`
	FeeType         bool      `gorm:"column:fee_type;type:bool;default:false"`
	Retries         int       `gorm:"column:retries;type:int;default:0"`
	TxHash          string    `gorm:"column:tx_hash;type:varchar(255)"`
	UTR             string    `gorm:"column:utr;type:varchar(255);not null"`
	TxMeta          []byte    `gorm:"column:tx_meta;type:jsonb"`
	CreatedAt       time.Time `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt       time.Time `gorm:"column:updated_at;autoUpdateTime"`
}

func (TransactionLog) TableName() string {
	return "transaction_logs"
}
