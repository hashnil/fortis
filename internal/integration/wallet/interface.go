package wallet

import "fortis/entity/models"

type Provider interface {
	RegisterDelegatedUser(request *models.CreateUserRequest) (*models.CreateUserResponse, error)
	CreateWallet(request *models.WalletRequest) (*models.WalletResponse, error)
	TransferAssets(request *models.TransactionRequest) (*models.TransactionResponse, error)
}
