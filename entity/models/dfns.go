package models

type WalletRequest struct {
	Network    string `json:"network"`
	Name       string `json:"name"`
	DelegateTo string `json:"delegateTo"`
}

type WalletResponse struct {
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
