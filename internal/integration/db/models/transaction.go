package models

import "time"

type TransactionStatus string

var (
	Pending     TransactionStatus = "Pending"
	Executing   TransactionStatus = "Executing"
	Broadcasted TransactionStatus = "Broadcasted"
	Confirmed   TransactionStatus = "Confirmed"
	Failure     TransactionStatus = "Failure"
	Rejected    TransactionStatus = "Rejected"
)

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
