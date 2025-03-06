package constants

const (
	InvalidRequestParser = "Failed to parse request body"
	InvalidRequest       = "Invalid Request Body"
	DuplicateUser        = "user is already registered "
	InactiveUser         = "user is not registered or activated "

	ErrInvalidProvider    = "Invalid wallet provider"
	ErrCreateUser         = "Failed to create user"
	ErrActivateUser       = "Failed to activate user"
	ErrCreateWallet       = "Failed to create wallet"
	ErrInitTransferAssets = "Failed to create transfer assets payload"
	ErrTransferAssets     = "Failed to transfer assets"
	ErrExistingUser       = "User is already registered and activated"
	ErrInactiveUser       = "User is not registered or activated"
)
