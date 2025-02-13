package models

type WalletRequest struct {
	UserID   string `json:"user_id,omitempty"`
	Username string `json:"username"`
}

type WalletResponse struct {
	Result string `json:"result"`
}
