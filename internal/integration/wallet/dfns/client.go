package dfns

import (
	"bytes"
	"encoding/json"
	"fmt"
	"fortis/entity/constants"
	"fortis/infrastructure/config"
	"fortis/pkg/utils"
	"io"
	"net/http"
	"sync"

	"github.com/dfns/dfns-sdk-go/credentials"
	api "github.com/dfns/dfns-sdk-go/dfnsapiclient"
)

var (
	serviceAccountClient *http.Client
	once                 sync.Once
)

// Initialize and return the singleton DFNS client for service account calls.
func getServiceAccountClient() (*http.Client, error) {
	var err error
	once.Do(func() {
		// Configure credentials for authentication
		conf := &credentials.AsymmetricKeySignerConfig{
			PrivateKey: string(config.DFNS_PRIVATE_KEY), // DFNS Credential Private Key
			CredID:     config.GetCredentialID(),        // DFNS Credential ID
		}

		// Create a DFNS API client instance with authentication
		authToken := config.GetAuthToken()
		apiOptions, apiErr := api.NewDfnsAPIOptions(&api.DfnsAPIConfig{
			AppID:     config.GetAppID(),   // DFNS Application ID
			AuthToken: &authToken,          // Authentication Token
			BaseURL:   config.GetBaseURL(), // DFNS API Base URL
		}, credentials.NewAsymmetricKeySigner(conf))
		if apiErr != nil {
			err = fmt.Errorf("error creating DFNS API options for Service-Account: %w", apiErr)
			return
		}

		serviceAccountClient = api.CreateDfnsAPIClient(apiOptions)
	})

	return serviceAccountClient, err
}

// Creates a new DFNS client for user-specific authentication.
func newUserAuthClient(userAuthToken *string) (*http.Client, error) {
	conf := &credentials.AsymmetricKeySignerConfig{
		PrivateKey: string(config.DFNS_PRIVATE_KEY),
		CredID:     config.GetCredentialID(),
	}

	apiOptions, err := api.NewDfnsAPIOptions(&api.DfnsAPIConfig{
		AppID:     config.GetAppID(),
		AuthToken: userAuthToken, // Dynamic Auth Token
		BaseURL:   config.GetBaseURL(),
	}, credentials.NewAsymmetricKeySigner(conf))
	if err != nil {
		return nil, fmt.Errorf("error creating DFNS API options for End-User: %w", err)
	}

	return api.CreateDfnsAPIClient(apiOptions), nil
}

func APIClient[T any](request interface{}, httpMethod, URL string, userAuthToken *string) (*T, error) {
	var (
		dfnsClient *http.Client
		err        error
	)

	// Decide which client to use
	if userAuthToken != nil {
		dfnsClient, err = newUserAuthClient(userAuthToken)
	} else {
		dfnsClient, err = getServiceAccountClient()
	}
	if err != nil {
		return nil, err
	}

	// Prepare request body if applicable
	var requestBody io.Reader
	if httpMethod == http.MethodPost && request != nil {
		requestBody = bytes.NewBuffer(utils.MarshalToJSON(request))
	}

	// Create HTTP request
	req, err := http.NewRequest(httpMethod, config.GetBaseURL()+URL, requestBody)
	if err != nil {
		return nil, fmt.Errorf("error creating HTTP request: %w", err)
	}

	// Set additional headers for user-authenticated requests
	if isUserActionURL(URL) {
		req.Header.Set("x-dfns-useraction", "false")
	}

	// Execute the request
	resp, err := dfnsClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error executing request: %w", err)
	}
	defer resp.Body.Close()

	// Check for HTTP response errors
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("DFNS API error: %s", resp.Status)
	}

	// Decode response JSON
	var response T
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("error decoding JSON response: %v", err)
	}

	return &response, nil
}

func isUserActionURL(url string) bool {
	return url == constants.CompleteUserRegistrationURL ||
		url == constants.UserActionSignatureChallengeURL ||
		url == constants.UserActionSignatureURL
}
