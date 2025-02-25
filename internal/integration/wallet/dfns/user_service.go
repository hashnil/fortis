package dfns

import (
	"encoding/json"
	"fmt"
	"fortis/entity/constants"
	"fortis/entity/models"
	dbmodels "fortis/internal/integration/db/models"
	"fortis/pkg/utils"
)

func (p *DFNSWalletProvider) RegisterDelegatedUser(request *models.CreateUserRequest) (*models.CreateUserResponse, error) {
	// Ensure user is registered (create a new user if not found)
	userResponse, existingUser, err := p.registerOrFetchUser(*request)
	if err != nil {
		return nil, fmt.Errorf("failed to register or fetch user: %w", err)
	}

	return &models.CreateUserResponse{
		ExistingUser:            existingUser,
		Challenge:               userResponse.Challenge,
		TempAuthenticationToken: userResponse.TemporaryAuthenticationToken,
	}, nil
}

// registerOrFetchUser checks if the user exists in the database or registers a new user in DFNS.
func (p *DFNSWalletProvider) registerOrFetchUser(request models.CreateUserRequest) (*models.DFNSUserRegistrationResponse, bool, error) {
	// Check if user exists in DB
	user, err := p.dbClient.FindUserByID(request.UserID)
	if err != nil {
		return nil, false, fmt.Errorf("failed to check user %s in DB: %w", request.UserID, err)
	}

	if user.ID != "" {
		var userResponse models.DFNSUserRegistrationResponse
		if err := json.Unmarshal(user.Metadata, &userResponse); err != nil {
			return nil, false, fmt.Errorf("failed to unmarshal user metadata: %w", err)
		}
		return &userResponse, true, nil
	}

	// User not found, register a new delegated user
	userResponse, err := p.registerDelegatedUser(request.Username)
	if err != nil {
		return nil, false, fmt.Errorf("failed to register delegated user %s in DFNS: %w", request.Username, err)
	}

	// Save user details in DB
	newUser := dbmodels.User{
		ID:       request.UserID,
		Name:     request.Username,
		Provider: constants.DFNS,
		IsActive: false, // Activation pending
		Metadata: utils.MarshalToJSON(userResponse),
	}

	if err := p.dbClient.CreateUser(newUser); err != nil {
		return nil, false, fmt.Errorf("failed to store user in DB: %w", err)
	}

	return userResponse, false, nil
}

// registerDelegatedUser registers a new user in DFNS and returns the response.
func (p *DFNSWalletProvider) registerDelegatedUser(username string) (*models.DFNSUserRegistrationResponse, error) {
	userData := models.DFNSUserRegistrationRequest{
		Kind:  "EndUser",
		Email: username,
	}

	return APIClient[models.DFNSUserRegistrationResponse](userData, "POST", "/auth/registration/delegated", nil)
}
