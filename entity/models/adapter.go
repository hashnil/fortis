package models

// --- User Registration --- //
type CreateUserRequest struct {
	UserID   string `json:"user_id,omitempty"` // with prefix: us-
	Username string `json:"username" binding:"required"`
}

type CreateUserResponse struct {
	Result              string `json:"result"`
	ExistingUser        bool   `json:"existing_user,omitempty"`
	Challenge           string `json:"challenge,omitempty"`
	AuthenticationToken string `json:"auth_token,omitempty"`
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
	Addresses map[string]string `json:"addresses,omitempty"`
}

// --- Transactions --- //
type TransactionRequest struct {
	UserID      string `json:"user_id,omitempty"` // for from account identification; with prefix: us-
	FromAccount string `json:"from_account,omitempty"`
	ToAccount   string `json:"to_account"`
	Amount      string `json:"amount"`
	Denom       string `json:"denom"`
	Fee         string `json:"fee"`
}

type TransactionResponse struct {
	Result string `json:"result"`
}
