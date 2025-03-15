package constants

const (
	// ID prefixes
	UserPrefix   = "us-"
	WalletPrefix = "wa-"

	// TODO: Chain ID's
	Solana = "SolanaDevnet"
	Base   = "BaseSepolia"

	// DFNS API's
	DelegatedRegistrationURL        = "/auth/registration/delegated"
	DelegatedRegistrationRestartURL = "/auth/registration/delegated/restart"
	CompleteUserRegistrationURL     = "/auth/registration"
	DelegatedLoginURL               = "/auth/login/delegated"
	CreateWalletsURL                = "/wallets"
	TransferAssetsURL               = "/wallets/%s/transfers"
	UserActionSignatureChallengeURL = "/auth/action/init"
	UserActionSignatureURL          = "/auth/action"

	// Transfer challenge keys
	FundTransferChallenge = "fundTransfer"
	FeeTransferChallenge  = "feeTransfer"
)

// TokenDecimals stores the decimal places for each token
var TokenDecimals = map[string]int{
	"USDC": 6,
	"USDT": 6,
	"SOL":  9,
	"ETH":  18,
}

type TransactionStatus string

var (
	Pending     TransactionStatus = "Pending"
	Executing   TransactionStatus = "Executing"
	Broadcasted TransactionStatus = "Broadcasted"
	Confirmed   TransactionStatus = "Confirmed"
	Failure     TransactionStatus = "Failure"
	Rejected    TransactionStatus = "Rejected"
)
