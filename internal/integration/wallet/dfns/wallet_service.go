package dfns

import (
	"encoding/json"
	"fmt"
	"fortis/entity/constants"
	"fortis/entity/models"
	"fortis/infrastructure/config"
	dbmodels "fortis/internal/integration/db/models"
	"fortis/pkg/utils"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// CreateWallet ensures the user exists and creates wallets for each configured network.
func (p *DFNSWalletProvider) CreateWallet(request models.WalletRequest) (*models.WalletResponse, error) {
	// Retrieve user details from the database
	dbUser, err := p.dbClient.FindUserByID(request.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve user %s from DB: %w", request.UserID, err)
	}

	// Iterate over configured networks and create/fetch wallets
	var response models.WalletResponse
	for _, network := range config.GetNetworks() {
		wallet, err := p.createOrFetchWallet(dbUser, network)
		if err != nil {
			return nil, fmt.Errorf("failed to process wallet for network %s: %w", network, err)
		}

		// Append wallet address to response
		response.Addresses[network] = wallet.Address
	}

	return &response, nil
}

// createOrFetchWallet checks if a wallet exists for a specific network or creates a new wallet in DFNS.
func (p *DFNSWalletProvider) createOrFetchWallet(dbUser dbmodels.User, network string) (*dbmodels.Wallet, error) {
	// Check if a wallet already exists in the database for this user and network
	wallet, err := p.dbClient.FindWalletByNameAndNetwork(dbUser.Name, constants.DFNS, network)
	if err == nil {
		// Wallet already exists, return it
		return &wallet, nil
	} else if err != gorm.ErrRecordNotFound {
		// Return if any error other than "wallet not found" occurs
		return nil, fmt.Errorf("error retrieving wallet from DB: %w", err)
	}

	// Extract DFNS user details from stored metadata
	var userResponse models.DFNSUserRegistrationResponse
	if err := json.Unmarshal(dbUser.Metadata, &userResponse); err != nil {
		return nil, fmt.Errorf("failed to parse user metadata: %w", err)
	}

	// Construct wallet creation request for DFNS
	walletRequest := &models.DFNSWalletRequest{
		Network:    network,
		Name:       fmt.Sprintf("%s-%s-wallet", dbUser.Name, network),
		DelegateTo: userResponse.User.ID, // Assigning wallet ownership to DFNS user
	}

	// Send API request to DFNS to create a new wallet
	walletResponse, err := APIClient[models.DFNSWalletResponse](walletRequest, "POST", "/wallets", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create wallet on DFNS for network %s: %w", network, err)
	}

	// Construct wallet model for database storage
	newWallet := dbmodels.Wallet{
		ID:       constants.WalletPrefix + uuid.NewString(),
		UserID:   dbUser.ID,
		Username: dbUser.Name,
		Provider: constants.DFNS,
		Network:  network,
		Name:     walletResponse.Name,
		Address:  walletResponse.Address,
		IsActive: walletResponse.Status == "Active",
		Metadata: utils.MarshalToJSON(walletResponse), // Store full response for reference
	}

	// Persist wallet details in the database
	if err := p.dbClient.CreateWallet(newWallet); err != nil {
		return nil, fmt.Errorf("failed to store wallet in DB: %w", err)
	}

	return &newWallet, nil
}
