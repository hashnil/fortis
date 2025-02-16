package models

import "time"

// DFNSUserRegistrationRequest represents the request payload for registering a user in DFNS.
type DFNSUserRegistrationRequest struct {
	Email string `json:"email"`
	Kind  string `json:"kind"`
}

// DFNSUserRegistrationResponse represents the response received after user registration in DFNS.
type DFNSUserRegistrationResponse struct {
	RP struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"rp"`
	User struct {
		ID          string `json:"id"`
		Name        string `json:"name"`
		DisplayName string `json:"displayName"`
	} `json:"user"`
	TemporaryAuthenticationToken string `json:"temporaryAuthenticationToken"`
	SupportedCredentialKinds     struct {
		FirstFactor  []string `json:"firstFactor"`
		SecondFactor []string `json:"secondFactor"`
	} `json:"supportedCredentialKinds"`
	Challenge       string `json:"challenge"`
	PubKeyCredParam []struct {
		Type string `json:"type"`
		Alg  int    `json:"alg"`
	} `json:"pubKeyCredParam"`
	Attestation        string `json:"attestation"`
	ExcludeCredentials []struct {
		Type       string `json:"type"`
		ID         string `json:"id"`
		Transports string `json:"transports"`
	} `json:"excludeCredentials"`
	AuthenticatorSelection struct {
		AuthenticatorAttachment string `json:"authenticatorAttachment"`
		ResidentKey             string `json:"residentKey"`
		RequireResidentKey      bool   `json:"requireResidentKey"`
		UserVerification        string `json:"userVerification"`
	} `json:"authenticatorSelection"`
}

// DFNSWalletRequest represents the request structure for creating a new wallet in DFNS.
type DFNSWalletRequest struct {
	Network    string `json:"network"`
	Name       string `json:"name"`
	DelegateTo string `json:"delegateTo"`
}

// DFNSWalletResponse represents the response received after creating a wallet in DFNS.
type DFNSWalletResponse struct {
	ID          string `json:"id"`
	Network     string `json:"network"`
	Status      string `json:"status"`
	Name        string `json:"name"`
	Address     string `json:"address"`
	DateCreated string `json:"dateCreated"`
	SigningKey  struct {
		Curve     string `json:"curve"`
		Scheme    string `json:"scheme"`
		PublicKey string `json:"publicKey"`
	} `json:"signingKey"`
}

// DFNSTransactionRequest represents the request payload for initiating a blockchain transaction.
type DFNSTransactionRequest struct {
	Kind       string `json:"kind"`
	Contract   string `json:"contract,omitempty"` // From account for EVM chains
	Mint       string `json:"mint,omitempty"`     // From account for Solana chain
	To         string `json:"to"`
	Amount     string `json:"amount"`
	ExternalID string `json:"externalID"`
}

// DFNSTransactionResponse represents the response received after processing a transaction.
type DFNSTransactionResponse struct {
	ID        string `json:"id"`
	WalletID  string `json:"walletId"`
	Network   string `json:"network"`
	Requester struct {
		UserID  string `json:"userId"`
		TokenID string `json:"tokenId"`
		AppID   string `json:"appId"`
	} `json:"requester"`
	RequestBody struct {
		Kind     string `json:"kind"`
		Contract string `json:"contract"`
		Mint     string `json:"mint"`
		Amount   string `json:"amount"`
		To       string `json:"to"`
	} `json:"requestBody"`
	Metadata struct {
		Asset struct {
			Symbol   string             `json:"symbol"`
			Decimals int                `json:"decimals"`
			Verified bool               `json:"verified"`
			Quotes   map[string]float64 `json:"quotes"`
		} `json:"asset"`
	} `json:"metadata"`
	Status          string    `json:"status"`
	Fee             string    `json:"fee"`
	TxHash          string    `json:"txHash"`
	DateRequested   time.Time `json:"dateRequested"`
	DateBroadcasted time.Time `json:"dateBroadcasted"`
	DateConfirmed   time.Time `json:"dateConfirmed"`
}
