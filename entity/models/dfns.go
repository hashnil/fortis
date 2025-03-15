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

// DFNSCompleteUserRegistrationResponse is the response returned when the user completes the registration process.
type DFNSCompleteUserRegistrationResponse struct {
	Credential struct {
		UUID           string `json:"uuid"`
		CredentialKind string `json:"credentialKind"`
		Name           string `json:"name"`
	} `json:"credential"`
	User struct {
		ID       string `json:"id"`
		Username string `json:"username"`
		OrgID    string `json:"orgId"`
	} `json:"user"`
}

// DFNSWebhookRequest represents the request to register a webhook.
type DFNSWebhookRequest struct {
	URL         string   `json:"url"`
	Description string   `json:"description"`
	Events      []string `json:"events"`
}

// DFNSWebhookResponse represents the response from webhook registration.
type DFNSWebhookResponse struct {
	ID          string   `json:"id"`
	URL         string   `json:"url"`
	Description string   `json:"description"`
	Events      []string `json:"events"`
	Status      string   `json:"status"`
	Secret      string   `json:"secret"`
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
	Kind      string `json:"kind"`
	Contract  string `json:"contract,omitempty"` // From account for EVM chains
	Mint      string `json:"mint,omitempty"`     // From account for Solana chain
	To        string `json:"to"`
	Amount    string `json:"amount"`
	CreateATA bool   `json:"createDestinationAccount,omitempty"` // Is used to create AssociatedTokenAccount for Spl
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

// LoginToken represents the authentication token returned upon user login.
type LoginToken struct {
	Token string `json:"token"`
}

// UserActionSignatureChallengeRequest represents the payload to initiate a user action signature challenge.
type UserActionSignatureChallengeRequest struct {
	UserActionPayload    string `json:"userActionPayload"`    // JSON-encoded body of the request being signed
	UserActionHTTPMethod string `json:"userActionHttpMethod"` // HTTP method of the request being signed (e.g., POST, PUT, DELETE, GET)
	UserActionHTTPPath   string `json:"userActionHttpPath"`   // Path of the request being signed
}

// UserActionSignatureChallengeResponse represents the response for signing a user action.
type UserActionSignatureChallengeResponse struct {
	SupportedCredentialKinds []struct {
		Kind                 string `json:"kind"`                 // "Fido2" or "Key"
		Factor               string `json:"factor"`               // "first", "second", or "either"
		RequiresSecondFactor bool   `json:"requiresSecondFactor"` // Indicates if second factor is required
	} `json:"supportedCredentialKinds"`
	Challenge                 string `json:"challenge"`                 // Unique challenge value
	ChallengeIdentifier       string `json:"challengeIdentifier"`       // Temporary authentication token
	ExternalAuthenticationUrl string `json:"externalAuthenticationUrl"` // Optional URL for cross-device signing
	AllowCredentials          struct {
		Key []struct {
			Type string `json:"type"` // Always "public-key"
			ID   string `json:"id"`   // Unique identifier for the credential
		} `json:"key"`
		PasswordProtectedKey []struct {
			Type                string `json:"type"`                // Always "public-key"
			ID                  string `json:"id"`                  // Unique identifier
			EncryptedPrivateKey string `json:"encryptedPrivateKey"` // Encrypted private key
		} `json:"passwordProtectedKey"`
		WebAuthn []struct {
			Type       string   `json:"type"`       // Always "public-key"
			ID         string   `json:"id"`         // Unique identifier
			Transports []string `json:"transports"` // Optional list of transports
		} `json:"webauthn"`
	} `json:"allowCredentials"`
}

// UserActionSigningRequest represents the request payload for completing user action signing.
type UserActionSigningRequest struct {
	ChallengeIdentifier string     `json:"challengeIdentifier"`
	FirstFactor         AuthFactor `json:"firstFactor"`
	SecondFactor        AuthFactor `json:"secondFactor,omitempty"`
}

type CredentialAssertion struct {
	CredID            string `json:"credId"`            // Credential ID
	ClientData        string `json:"clientData"`        // Base64-encoded client data
	AuthenticatorData string `json:"authenticatorData"` // Base64-encoded authenticator data
	Signature         string `json:"signature"`         // Digital signature for authentication
	UserHandle        string `json:"userHandle"`        // User identifier
}

type AuthFactor struct {
	Kind                string              `json:"kind"`
	CredentialAssertion CredentialAssertion `json:"credentialAssertion"`
}

// UserActionSigningResponse gets the user action signature
type UserActionSigningResponse struct {
	UserAction string `json:"userAction"`
}
