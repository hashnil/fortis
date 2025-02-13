package models

type DFNSUserRegistrationRequest struct {
	Email string `json:"email"`
	Kind  string `json:"kind"`
}

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

type DFNSWalletRequest struct {
	Network    string `json:"network"`
	Name       string `json:"name"`
	DelegateTo string `json:"delegateTo"`
}

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
