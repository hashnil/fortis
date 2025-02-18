package models

type WalletRequest struct {
	UserID   string `json:"user_id,omitempty"` // with prefix: us-
	Username string `json:"username,omitempty"`
}

type WalletResponse struct {
	Result string `json:"result"`
}

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
