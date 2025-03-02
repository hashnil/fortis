package models

// --- User Registration --- //
type CreateUserRequest struct {
	UserID   string `json:"user_id,omitempty"` // with prefix: us-
	Username string `json:"username" binding:"required"`
}

type CreateUserResponse struct {
	Result       string `json:"result"`
	ExistingUser bool   `json:"existing_user,omitempty"`
	Challenge    string `json:"challenge"`
}

// --- User Activation --- //
type ActivateUserRequest struct {
	UserID         string                    `json:"user_id,omitempty"` // with prefix: us-
	CredentialInfo map[string]CredentialInfo `json:"credential_info" binding:"required"`
}

type CredentialInfo struct {
	CredentialID    string `json:"credential_id" binding:"required"`
	ClientData      string `json:"client_data" binding:"required"`
	AttestationData string `json:"attestation_data" binding:"required"`
}

type ActivateUserResponse struct {
	Result string `json:"result"`
}

// --- Wallet Management --- //
type WalletRequest struct {
	UserID string `json:"user_id,omitempty"` // with prefix: us-
}

type WalletResponse struct {
	Result    string            `json:"result"`
	Addresses map[string]string `json:"addresses"`
}

// --- Transactions --- //
type InitTransferRequest struct {
	UserID    string `json:"user_id,omitempty"`             // Identifier for the sender with prefix (e.g., "us-")
	ToAccount string `json:"to_account" binding:"required"` // Address of the receiver
	Amount    string `json:"amount" binding:"required"`     // Transfer amount (in smallest unit)
	Fee       string `json:"fee" binding:"required"`        // Transaction fee (in native token)
	Denom     string `json:"denom" binding:"required"`      // Token/Currency type (e.g., "ETH", "BTC", "USDC")
	Memo      string `json:"memo,omitempty"`                // Optional transaction note
}

type InitTransferResponse struct {
	Result    string            `json:"result"`
	Challenge map[string]string `json:"challenge"`
}

type TransferRequest struct {
	UserID         string                    `json:"user_id,omitempty"` // Identifier for the sender with prefix (e.g., "us-")
	CredentialInfo map[string]CredentialInfo `json:"credential_info" binding:"required"`
}

type TransferResponse struct {
	Result string `json:"result"`
}
