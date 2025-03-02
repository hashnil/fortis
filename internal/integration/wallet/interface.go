package wallet

import "fortis/entity/models"

type Provider interface {
	RegisterDelegatedUser(models.CreateUserRequest) (*models.CreateUserResponse, error)
	ActivateDelegatedUser(models.ActivateUserRequest) error
	CreateWallet(models.WalletRequest) (*models.WalletResponse, error)
	InitTransferAssets(models.InitTransferRequest) (*models.InitTransferResponse, error)
	// TransferAssets(models.TransferRequest) (*models.TransferResponse, error)
}
