package models

import (
	"time"

	"gorm.io/gorm"
)

type InflightTransaction struct {
	Challenge            string         `gorm:"column:challenge;type:varchar(50);primaryKey"` // Unique identifier for the transaction
	ChallengeIdentifier  string         `gorm:"column:challenge_identifier;type:text"`        // Unique challenge identifier
	URL                  string         `gorm:"column:url;type:varchar(255)"`                 // Target URL for transaction processing
	AuthToken            string         `gorm:"column:auth_token;type:text"`                  // User authentication token
	RequestPayload       []byte         `gorm:"column:request_payload;type:jsonb"`            // Transfer request in JSONB format
	TransferPayload      []byte         `gorm:"column:transfer_payload;type:jsonb"`           // Transaction request in JSONB format
	UserChallengePayload []byte         `gorm:"column:user_challenge_payload;type:jsonb"`     // User's challenge response in JSONB format
	SenderInfo           []byte         `gorm:"column:sender_info;type:jsonb"`                // Sender information in JSONB format
	CreatedAt            time.Time      `gorm:"column:created_at;autoCreateTime"`             // Record creation timestamp
	UpdatedAt            time.Time      `gorm:"column:updated_at;autoUpdateTime"`             // Record update timestamp
	DeletedAt            gorm.DeletedAt `gorm:"column:deleted_at;index"`                      // Soft delete timestamp
}

func (InflightTransaction) TableName() string {
	return "inflight_transactions"
}

type Transaction struct {
	TxHash          string         `gorm:"column:tx_hash;type:varchar(255);primaryKey"`
	SenderName      string         `gorm:"column:sender_name;type:varchar(50);not null"`
	SenderAddress   string         `gorm:"column:sender_address;type:varchar(255);not null"`
	ReceiverName    string         `gorm:"column:receiver_name;type:varchar(50);not null"`
	ReceiverAddress string         `gorm:"column:receiver_address;type:varchar(255);not null"`
	Amount          string         `gorm:"column:amount;type:varchar(50);not null"`
	Denom           string         `gorm:"column:denom;type:varchar(50);not null"`
	Provider        string         `gorm:"column:provider;type:varchar(50);not null"`
	Network         string         `gorm:"column:network;type:varchar(50);not null"`
	TypeFee         bool           `gorm:"column:type_fee;type:bool"`
	Status          string         `gorm:"column:status;type:varchar(50);not null"`
	UTR             string         `gorm:"column:utr;type:varchar(255);not null"`
	CreatedAt       time.Time      `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt       time.Time      `gorm:"column:updated_at;autoUpdateTime"`
	DeletedAt       gorm.DeletedAt `gorm:"column:deleted_at;index"`
}

func (Transaction) TableName() string {
	return "transactions"
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
	TypeFee         bool      `gorm:"column:type_fee;type:bool"`
	Status          string    `gorm:"column:status;type:varchar(50);not null"`
	FailureReason   string    `gorm:"column:failure_reason;type:varchar(1024)"` // TODO: how to check
	TxHash          string    `gorm:"column:tx_hash;type:varchar(255)"`
	UTR             string    `gorm:"column:utr;type:varchar(255);not null"`
	TxMeta          []byte    `gorm:"column:tx_meta;type:jsonb"`
	CreatedAt       time.Time `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt       time.Time `gorm:"column:updated_at;autoUpdateTime"`
}

func (TransactionLog) TableName() string {
	return "transaction_logs"
}
