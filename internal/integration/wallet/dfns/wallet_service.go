package dfns

import (
	"errors"
	"fmt"
	"fortis/entity/constants"
	"fortis/entity/models"
	"fortis/infrastructure/config"
	dbmodels "fortis/internal/integration/db/models"
	"fortis/pkg/utils"
	"log"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// CreateWallet ensures the user exists and creates wallets for each configured network.
func (p *DFNSWalletProvider) CreateWallet(request models.WalletRequest) (*models.WalletResponse, error) {
	log.Printf("[INFO] CreateWallet: Initiating wallet creation for UserID: %s\n", request.UserID)

	// Retrieve user details from the database
	dbUser, err := p.dbClient.FindUserByID(request.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve user %s from DB: %w", request.UserID, err)
	}

	// If user is not active, return an error
	if !dbUser.IsActive {
		return nil, errors.New(constants.InactiveUser + dbUser.ID)
	}

	// Iterate over configured networks and create/fetch wallets
	var response models.WalletResponse
	response.Addresses = make(map[string]string)
	for _, network := range config.GetNetworks() {
		log.Printf("[INFO] CreateWallet: Processing wallet for User: %s, Network: %s\n", dbUser.Name, network)

		wallet, err := p.createOrFetchWallet(dbUser, network)
		if err != nil {
			return nil, fmt.Errorf("failed to process wallet for network %s: %w", network, err)
		}

		log.Printf("[INFO] CreateWallet: Wallet successfully processed for User: %s, Network: %s, Address: %s\n",
			dbUser.Name, network, wallet.Address)

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
		log.Printf("[INFO] createOrFetchWallet: Wallet already exists for User: %s, Network: %s\n", dbUser.Name, network)
		return &wallet, nil
	} else if err != gorm.ErrRecordNotFound {
		// Return if any error other than "wallet not found" occurs
		return nil, fmt.Errorf("error retrieving wallet from DB: %w", err)
	}

	log.Printf("[INFO] createOrFetchWallet: No existing wallet found. Creating new wallet for User: %s, Network: %s\n",
		dbUser.Name, network)

	// Extract DFNS user details from stored metadata
	var userResponse models.DFNSUserRegistrationResponse
	utils.UnmarshalFromJSON(dbUser.Metadata, &userResponse)

	// Construct wallet creation request for DFNS
	walletRequest := &models.DFNSWalletRequest{
		Network:    network,
		Name:       fmt.Sprintf("%s-%s-wallet", dbUser.Name, network),
		DelegateTo: userResponse.User.ID, // Assigning wallet ownership to DFNS user
	}

	// Send API request to DFNS to create a new wallet
	walletResponse, err := APIClient[models.DFNSWalletResponse](walletRequest, "POST", constants.CreateWalletsURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create wallet on DFNS for network %s: %w", network, err)
	}

	log.Printf("[INFO] createOrFetchWallet: Successfully created wallet on DFNS for User: %s, Network: %s, Address: %s\n",
		dbUser.Name, network, walletResponse.Address)

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
