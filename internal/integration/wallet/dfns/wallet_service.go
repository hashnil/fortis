package dfns

import (
	"encoding/json"
	"fmt"
	"fortis/entity/constants"
	"fortis/entity/models"
	"fortis/internal/integration/db"
	dbmodels "fortis/internal/integration/db/models"
	"fortis/internal/integration/wallet"
	"fortis/pkg/utils"

	"github.com/google/uuid"
)

type DFNSWalletProvider struct {
	dbClient db.Client
}

func NewDFNSWalletProvider(dbClient db.Client) wallet.Provider {
	return &DFNSWalletProvider{dbClient: dbClient}
}

// CreateWallet ensures the user exists and creates wallets for multiple chains, storing them in the database.
func (p *DFNSWalletProvider) CreateWallet(request *models.WalletRequest) (*models.WalletResponse, error) {
	// Ensure user is registered (create a new user if not found)
	userResponse, err := p.registerOrFetchUser(*request)
	if err != nil {
		return nil, fmt.Errorf("failed to register or fetch user: %w", err)
	}

	// Create wallets for specified networks (if they don’t exist)
	networks := []string{constants.Solana, constants.Base}

	for _, network := range networks {
		_, err := p.createOrFetchWallet(*request, *userResponse, network)
		if err != nil {
			return nil, fmt.Errorf("failed to create or fetch wallet for chain %s: %w", network, err)
		}
	}

	return &models.WalletResponse{}, nil
}

// registerOrFetchUser checks if the user exists in the database or registers a new user in DFNS.
func (p *DFNSWalletProvider) registerOrFetchUser(request models.WalletRequest) (*models.DFNSUserRegistrationResponse, error) {
	// Check if user exists in DB
	user, err := p.dbClient.FindUserWallet(constants.UserPrefix + request.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to check user %s in DB: %w", request.UserID, err)
	}

	if user.ID != "" {
		var userResponse models.DFNSUserRegistrationResponse
		if err := json.Unmarshal(user.UserMeta, &userResponse); err != nil {
			return nil, fmt.Errorf("failed to unmarshal user metadata: %w", err)
		}
		return &userResponse, nil
	}

	// User not found, register a new delegated user
	return p.registerDelegatedUser(request.Username)
}

// registerDelegatedUser registers a new user in DFNS and returns the response.
func (p *DFNSWalletProvider) registerDelegatedUser(username string) (*models.DFNSUserRegistrationResponse, error) {
	userData := models.DFNSUserRegistrationRequest{
		Kind:  "EndUser",
		Email: username,
	}

	return APIClient[models.DFNSUserRegistrationResponse](userData, "POST", "/auth/registration/delegated")
}

// createOrFetchWallet checks if a wallet exists for a specific network or creates a new wallet in DFNS.
func (p *DFNSWalletProvider) createOrFetchWallet(
	internalUserData models.WalletRequest, dfnsUserData models.DFNSUserRegistrationResponse, network string,
) (*dbmodels.Wallet, error) {
	// Check if wallet exists in DB for the given network
	walletRecord, err := p.dbClient.FindWalletByNetwork(internalUserData.UserID, constants.DFNS, network)
	if err != nil {
		return nil, fmt.Errorf("error checking wallet in DB: %w", err)
	}

	if walletRecord.ID != "" {
		return &walletRecord, nil
	}

	// Wallet not found, create a new wallet in DFNS
	walletRequest := &models.DFNSWalletRequest{
		Network:    network,
		Name:       fmt.Sprintf("%s-%s-wallet", internalUserData.Username, network),
		DelegateTo: dfnsUserData.User.ID,
	}

	walletResponse, err := APIClient[models.DFNSWalletResponse](walletRequest, "POST", "/wallets")
	if err != nil {
		return nil, fmt.Errorf("failed to create wallet for network %s: %w", network, err)
	}

	// Prepare wallet data for database storage
	newWallet := dbmodels.Wallet{
		ID:         constants.WalletPrefix + uuid.NewString(),
		UserID:     constants.UserPrefix + internalUserData.UserID,
		Username:   internalUserData.Username,
		Provider:   constants.DFNS,
		Network:    network,
		Name:       walletResponse.Name,
		Address:    walletResponse.Address,
		IsActive:   walletResponse.Status == "Active",
		UserMeta:   utils.MarshalToJSON(dfnsUserData),
		WalletMeta: utils.MarshalToJSON(walletResponse),
	}

	// Store the wallet in the database
	if err := p.dbClient.CreateWallet(newWallet); err != nil {
		return nil, fmt.Errorf("failed to store wallet in DB: %w", err)
	}

	return &newWallet, nil
}
