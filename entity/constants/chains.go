package constants

const (
	// ID prefixes
	UserPrefix   = "us-"
	WalletPrefix = "wa-"

	// DFNS API's
	DelegatedRegistrationURL        = "/auth/registration/delegated"
	DelegatedRegistrationRestartURL = "/auth/registration/delegated/restart"
	CompleteUserRegistrationURL     = "/auth/registration"
	DelegatedLoginURL               = "/auth/login/delegated"
	CreateWalletsURL                = "/wallets"
	TransferAssetsURL               = "/wallets/%s/transfers"
)
