package dfns

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
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

	// Return authentication details for new user
	return &models.CreateUserResponse{
		Challenge:           userResponse.Challenge,
		AuthenticationToken: userResponse.TemporaryAuthenticationToken,
	}, nil
}

// registerOrFetchUser checks if the user exists in the database or registers a new user in DFNS.
func (p *DFNSWalletProvider) registerOrFetchUser(request models.CreateUserRequest) (*models.DFNSUserRegistrationResponse, error) {
	// Check if user exists in DB
	user, err := p.dbClient.FindUserByID(request.UserID)
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("failed to check user %s in DB: %w", request.UserID, err)
	}

	if err == nil && user.IsActive {
		return nil, errors.New(constants.DuplicateUser)
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

	return userResponse, nil
}

// registerDelegatedUser calls DFNS API to register a new user.
func (p *DFNSWalletProvider) registerDelegatedUser(username string) (*models.DFNSUserRegistrationResponse, error) {
	userData := models.DFNSUserRegistrationRequest{
		Kind:  "EndUser",
		Email: username,
	}
	return APIClient[models.DFNSUserRegistrationResponse](userData, "POST", "/auth/registration/delegated", nil)
}

// restartDelegatedRegistration calls DFNS API to restart user registration.
func (p *DFNSWalletProvider) restartDelegatedRegistration(username string) (*models.DFNSUserRegistrationResponse, error) {
	userData := models.DFNSUserRegistrationRequest{
		Kind:  "EndUser",
		Email: username,
	}
	return APIClient[models.DFNSUserRegistrationResponse](userData, "POST", "/auth/registration/delegated/restart", nil)
}

// ActivateDelegatedUser activates a user by completing their registration in DFNS.
func (p *DFNSWalletProvider) ActivateDelegatedUser(request models.ActivateUserRequest) (*models.ActivateUserResponse, error) {
	// Retrieve user details from the database
	user, err := p.dbClient.FindUserByID(request.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve user %s from DB: %w", request.UserID, err)
	}

	// If user is already active, return an error
	if user.IsActive {
		return nil, errors.New(constants.DuplicateUser)
	}

	// Parse user metadata to extract DFNS registration details
	var dfnsUser models.DFNSUserRegistrationResponse
	if err := json.Unmarshal(user.Metadata, &dfnsUser); err != nil {
		return nil, fmt.Errorf("failed to unmarshal DFNS user metadata: %w", err)
	}

	// Complete user registration using provided credentials and temporary authentication token
	completeUserRegResp, err := p.completeUserRegistration(request.CredentialInfo, dfnsUser.TemporaryAuthenticationToken)
	if err != nil {
		return nil, fmt.Errorf("failed to complete user registration: %w", err)
	}

	// Mark user as active in the database
	if err := p.dbClient.UpdateUser(dbmodels.User{
		ID:        user.ID,
		IsActive:  true,
		AuthToken: completeUserRegResp.Authentication.Token,
	}); err != nil {
		return nil, fmt.Errorf("failed to update user token in DB: %w", err)
	}

	return nil, nil
}

// completeUserRegistration finalizes the registration process by submitting credentials.
func (p *DFNSWalletProvider) completeUserRegistration(
	credentials map[string]models.CredentialInfo, tempAuthToken string,
) (*models.DFNSCompleteUserRegistrationResponse, error) {
	// Validate that required credentials are provided
	requiredKeys := []string{"Fido2", "Key"}
	for _, key := range requiredKeys {
		if _, exists := credentials[key]; !exists {
			return nil, fmt.Errorf("missing required credential: %s", key)
		}
	}

	// Construct request payload for completing user registration
	requestPayload := map[string]interface{}{
		"firstFactorCredential": map[string]interface{}{
			"credentialKind": "Fido2",
			"credentialInfo": map[string]string{
				"credId":          credentials["Fido2"].CredentialID,
				"clientData":      credentials["Fido2"].ClientData,
				"attestationData": credentials["Fido2"].AttestationData,
			},
		},
		"secondFactorCredential": map[string]interface{}{
			"credentialKind": "Key",
			"credentialInfo": map[string]string{
				"credId":          credentials["Key"].CredentialID,
				"clientData":      credentials["Key"].ClientData,
				"attestationData": credentials["Key"].AttestationData,
			},
		},
	}

	// Call the API to complete registration
	return APIClient[models.DFNSCompleteUserRegistrationResponse](requestPayload, "POST", "/auth/registration", &tempAuthToken)
}

// TODO: remove
func getCredentials(challenge string) map[string]interface{} {
	clientData, _ := json.Marshal(map[string]string{
		"challenge": challenge,
		"type":      "key.create",
	})

	hash := sha256Hex(clientData)

	// Generate a new RSA key pair
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		log.Fatalf("error generating key: %v", err)
	}

	publicKeyDER, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		log.Fatalf("error marshaling public key: %v", err)
	}

	publicKeyPEM := string(pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: publicKeyDER}))

	credInfo, _ := json.Marshal(map[string]string{
		"clientDataHash": string(hash),
		"publicKey":      publicKeyPEM,
	})

	signature, err := rsa.SignPKCS1v15(rand.Reader, privateKey, crypto.SHA256, sha256Hex(credInfo))
	if err != nil {
		log.Fatalf("error signing data: %v", err)
	}

	attestationData, _ := json.Marshal(map[string]string{
		"publicKey": publicKeyPEM,
		"signature": string(sha256Hex(signature)),
	})

	credID := toBase64URL(make([]byte, 16))
	rand.Read([]byte(credID))

	requestPayload := map[string]interface{}{
		"firstFactorCredential": map[string]interface{}{
			"credentialKind": "Key",
			"credentialInfo": map[string]string{
				"credId":          credID,
				"clientData":      toBase64URL(clientData),
				"attestationData": toBase64URL(attestationData),
			},
		},
	}

	return requestPayload
}

func toBase64URL(data []byte) string {
	return base64.RawURLEncoding.EncodeToString(data)
}

func sha256Hex(data []byte) []byte {
	hash := sha256.Sum256(data)
	return hash[:]
}
