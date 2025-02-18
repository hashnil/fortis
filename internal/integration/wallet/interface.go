package wallet

import "fortis/entity/models"

type Provider interface {
	CreateWallet(request *models.WalletRequest) (*models.WalletResponse, error)
	TransferAssets(request *models.TransactionRequest) (*models.TransactionResponse, error)
}
