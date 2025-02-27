package models

import "time"

type User struct {
	ID        string    `gorm:"column:id;type:varchar(50);primaryKey"`       // us-<uuid>: Unique User ID
	Name      string    `gorm:"column:name;type:varchar(50);not null;index"` // Username (Indexed)
	Provider  string    `gorm:"column:provider;type:varchar(50);not null"`   // Wallet provider (e.g., 'dfns', 'coinbase')
	IsActive  bool      `gorm:"column:is_active;type:bool;default:false"`    // User status (true = Active)
	AuthToken string    `gorm:"column:auth_token;type:varchar(255)"`         // User authentication token
	Metadata  []byte    `gorm:"column:metadata;type:jsonb;not null"`         // Additional user metadata (stored as JSONB)
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime"`            // Record creation timestamp
	UpdatedAt time.Time `gorm:"column:updated_at;autoUpdateTime"`            // Record update timestamp
}

func (User) TableName() string {
	return "users"
}
