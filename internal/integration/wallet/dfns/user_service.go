package dfns

import (
	"errors"
	"fmt"
	"fortis/entity/constants"
	"fortis/entity/models"
	dbmodels "fortis/internal/integration/db/models"
	"fortis/pkg/utils"
	"log"
	"strings"

	"gorm.io/gorm"
)

// RegisterDelegatedUser ensures the user is registered, handling duplicates and restarts if necessary.
func (p *DFNSWalletProvider) RegisterDelegatedUser(request models.CreateUserRequest) (*models.CreateUserResponse, error) {
	// Ensure user is registered (create a new user if not found)
	userResponse, err := p.registerOrFetchUser(request)

	// Handle user already exists scenario
	if err != nil {
		if strings.Contains(err.Error(), constants.DuplicateUser) {
			return &models.CreateUserResponse{ExistingUser: true}, nil
		}
		return nil, fmt.Errorf("failed to register or fetch user: %w", err)
	}

	return &models.CreateUserResponse{Challenge: userResponse.Challenge}, nil
}

// registerOrFetchUser checks if the user exists in the database or registers a new user in DFNS.
func (p *DFNSWalletProvider) registerOrFetchUser(request models.CreateUserRequest) (*models.DFNSUserRegistrationResponse, error) {
	// Check if user exists in DB
	user, err := p.dbClient.FindUserByID(request.UserID)
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("failed to check user %s in DB: %w", request.UserID, err)
	}

	if err == nil && user.IsActive {
		return nil, errors.New(constants.DuplicateUser + user.ID)
	}

	// Constrcut user object
	dbUser := dbmodels.User{
		ID:       request.UserID,
		Name:     request.Username,
		Provider: constants.DFNS,
		IsActive: false, // Activation pending
	}

	// If user exists and not already registered, attempt to restart the registration process
	if err == nil && !user.IsActive {
		// Call delegatedRegistrationRestart to refresh the authentication token and challenge
		refreshedUserResponse, err := p.restartDelegatedRegistration(request.Username)
		if err != nil {
			return nil, fmt.Errorf("failed to restart delegated registration for user %s: %w", request.UserID, err)
		}

		// Update user details in the database
		dbUser.Metadata = utils.MarshalToJSON(refreshedUserResponse)
		if err := p.dbClient.UpdateUser(dbUser); err != nil {
			return nil, fmt.Errorf("failed to update user in DB: %w", err)
		}

		log.Println("[INFO] Successfully restarted registration for:", request.Username)
		return refreshedUserResponse, nil
	}

	// User not found, register a new delegated user
	userResponse, err := p.registerDelegatedUser(request.Username)
	if err != nil {
		return nil, fmt.Errorf("failed to register delegated user %s in DFNS: %w", request.Username, err)
	}

	// Save new user in database
	dbUser.Metadata = utils.MarshalToJSON(userResponse)
	if err := p.dbClient.CreateUser(dbUser); err != nil {
		return nil, fmt.Errorf("failed to store user in DB: %w", err)
	}

	log.Println("[INFO] Successfully registered new user:", request.Username)
	return userResponse, nil
}

// registerDelegatedUser calls DFNS API to register a new user.
func (p *DFNSWalletProvider) registerDelegatedUser(username string) (*models.DFNSUserRegistrationResponse, error) {
	userData := models.DFNSUserRegistrationRequest{
		Kind:  "EndUser",
		Email: username,
	}
	return APIClient[models.DFNSUserRegistrationResponse](userData, "POST", constants.DelegatedRegistrationURL, nil)
}

// restartDelegatedRegistration calls DFNS API to restart user registration.
func (p *DFNSWalletProvider) restartDelegatedRegistration(username string) (*models.DFNSUserRegistrationResponse, error) {
	userData := models.DFNSUserRegistrationRequest{
		Kind:  "EndUser",
		Email: username,
	}
	return APIClient[models.DFNSUserRegistrationResponse](userData, "POST", constants.DelegatedRegistrationRestartURL, nil)
}

// ActivateDelegatedUser activates a user by completing their registration in DFNS.
func (p *DFNSWalletProvider) ActivateDelegatedUser(request models.ActivateUserRequest) error {
	// Retrieve user details from the database
	user, err := p.dbClient.FindUserByID(request.UserID)
	if err != nil {
		return fmt.Errorf("failed to retrieve user %s from DB: %w", request.UserID, err)
	}

	// If user is already active, return an error
	if user.IsActive {
		return errors.New(constants.DuplicateUser + user.ID)
	}

	// Parse user metadata to extract DFNS registration details
	var dfnsUser models.DFNSUserRegistrationResponse
	utils.UnmarshalFromJSON(user.Metadata, &dfnsUser)

	// Complete user registration using provided credentials and temporary authentication token
	_, err = p.completeUserRegistration(request.CredentialInfo, dfnsUser.TemporaryAuthenticationToken, dfnsUser.Challenge)
	if err != nil {
		return fmt.Errorf("failed to complete user registration: %w", err)
	}

	// Mark user as active in the database
	if err := p.dbClient.UpdateUser(dbmodels.User{
		ID:       user.ID,
		IsActive: true,
	}); err != nil {
		return fmt.Errorf("failed to update user activation status in DB: %w", err)
	}

	return nil
}

// completeUserRegistration finalizes the registration process by submitting credentials.
func (p *DFNSWalletProvider) completeUserRegistration(
	credentials []models.CredentialInfo, tempAuthToken, challenge string,
) (*models.DFNSCompleteUserRegistrationResponse, error) {
	// Ensure at least one credential is provided
	if len(credentials) == 0 || len(credentials) > 2 {
		return nil, fmt.Errorf("invalid number of credentials provided, must be 1 or 2")
	}

	// Construct request payload for completing user registration
	requestPayload := map[string]interface{}{}

	if len(credentials) == 1 {
		requestPayload["firstFactorCredential"] = map[string]interface{}{
			"credentialKind": credentials[0].CredentialKind,
			"credentialInfo": map[string]string{
				"credId":          credentials[0].CredentialID,
				"clientData":      credentials[0].ClientData,
				"attestationData": credentials[0].AttestationData,
			},
		}
	} else if len(credentials) == 2 {
		// Ensure Fido2 is the first factor and Key is the second factor
		var firstFactor, secondFactor *models.CredentialInfo
		for _, cred := range credentials {
			if strings.ToLower(cred.CredentialKind) == "fido2" {
				firstFactor = &cred
			} else if strings.ToLower(cred.CredentialKind) == "key" {
				secondFactor = &cred
			}
		}

		if firstFactor == nil || secondFactor == nil {
			return nil, fmt.Errorf("must provide both Fido2 and Key credentials in order")
		}

		requestPayload["firstFactorCredential"] = map[string]interface{}{
			"credentialKind": firstFactor.CredentialKind,
			"credentialInfo": map[string]string{
				"credId":          firstFactor.CredentialID,
				"clientData":      firstFactor.ClientData,
				"attestationData": firstFactor.AttestationData,
			},
		}

		// TODO: remove if not required
		requestPayload["secondFactorCredential"] = map[string]interface{}{
			"credentialKind": secondFactor.CredentialKind,
			"credentialInfo": map[string]string{
				"credId":          secondFactor.CredentialID,
				"clientData":      secondFactor.ClientData,
				"attestationData": secondFactor.AttestationData,
			},
		}
	}

	// TODO: remove
	requestPayload = utils.GetAttestationData(challenge)

	// Call the API to complete registration
	return APIClient[models.DFNSCompleteUserRegistrationResponse](requestPayload, "POST", "/auth/registration", &tempAuthToken)
}
