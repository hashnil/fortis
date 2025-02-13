package dfns

import (
	"bytes"
	"encoding/json"
	"fmt"
	"fortis/infrastructure/config"
	"io"
	"net/http"
	"sync"

	"github.com/dfns/dfns-sdk-go/credentials"
	api "github.com/dfns/dfns-sdk-go/dfnsapiclient"
)

var (
	dfnsClient    *http.Client
	dfnsOnce      sync.Once
	clientInitErr error
)

func initDFNSClient() {
	dfnsOnce.Do(func() {
		// Configure credentials for authentication
		conf := &credentials.AsymmetricKeySignerConfig{
			PrivateKey: string(config.DFNS_PRIVATE_KEY), // DFNS Credential Private Key
			CredID:     config.GetCredentialID(),        // DFNS Credential ID
		}

		// Create a DFNS API client instance with authentication
		authToken := config.GetAuthToken()
		apiOptions, err := api.NewDfnsAPIOptions(&api.DfnsAPIConfig{
			AppID:     config.GetAppID(),   // DFNS Application ID
			AuthToken: &authToken,          // Authentication Token
			BaseURL:   config.GetBaseURL(), // DFNS API Base URL
		}, credentials.NewAsymmetricKeySigner(conf))
		if err != nil {
			clientInitErr = fmt.Errorf("error creating DFNS API options: %w", err)
		}

		// Create DFNS API client
		dfnsClient = api.CreateDfnsAPIClient(apiOptions)
	})
}

func APIClient[T any](request interface{}, httpMethod, URL string) (*T, error) {
	initDFNSClient()
	if clientInitErr != nil {
		return nil, clientInitErr
	}

	// Prepare request body if applicable
	var requestBody io.Reader
	if httpMethod == http.MethodPost && request != nil {
		jsonData, err := json.Marshal(request)
		if err != nil {
			return nil, fmt.Errorf("error marshaling JSON: %w", err)
		}
		requestBody = bytes.NewBuffer(jsonData)
	}

	// Create an HTTP request
	req, err := http.NewRequest(httpMethod, config.GetBaseURL()+URL, requestBody)
	if err != nil {
		return nil, fmt.Errorf("error creating HTTP request: %w", err)
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

	// Decode response JSON into the provided response structure
	var response T
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("error decoding JSON response: %v", err)
	}

	return &response, nil
}
