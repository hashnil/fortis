package models

import "time"

type Wallet struct {
	ID         string    `gorm:"column:id;type:varchar(50);primaryKey"`                                  // wa-<uuid>: Unique wallet ID
	UserID     string    `gorm:"column:user_id;type:varchar(50);not null;uniqueIndex:user_network_idx"`  // us-<uuid>: User ID
	Username   string    `gorm:"column:username;type:varchar(50);not null"`                              // Username
	Provider   string    `gorm:"column:provider;type:varchar(50);not null;uniqueIndex:user_network_idx"` // Wallet provider (e.g., 'dfns', 'coinbase')
	Network    string    `gorm:"column:network;type:varchar(50);not null;uniqueIndex:user_network_idx"`  // Blockchain network (e.g., 'Solana', 'Base', 'Matic')
	Name       string    `gorm:"column:name;type:varchar(100)"`                                          // Wallet name (e.g., "My awesome wallet")
	Address    string    `gorm:"column:address;type:varchar(255);not null"`                              // Blockchain wallet address
	IsActive   bool      `gorm:"column:is_active;type:bool;default:true"`                                // Wallet status (true = Active)
	UserMeta   []byte    `gorm:"column:user_meta;type:jsonb;not null"`                                   // Additional user metadata (stored as JSONB)
	WalletMeta []byte    `gorm:"column:wallet_meta;type:jsonb;not null"`                                 // Additional wallet metadata (stored as JSONB)
	CreatedAt  time.Time `gorm:"column:created_at;autoCreateTime"`                                       // Record creation timestamp
	UpdatedAt  time.Time `gorm:"column:updated_at;autoUpdateTime"`                                       // Record update timestamp
}

func (Wallet) TableName() string {
	return "wallets"
}
