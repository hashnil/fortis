package wallet

import "fortis/entity/models"

type Provider interface {
	RegisterDelegatedUser(models.CreateUserRequest) (*models.CreateUserResponse, error)
	ActivateDelegatedUser(models.ActivateUserRequest) error
	CreateWallet(models.WalletRequest) (*models.WalletResponse, error)
	TransferAssets(models.TransferRequest) (*models.TransferResponse, error)
}
