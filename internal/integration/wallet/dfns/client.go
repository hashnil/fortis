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
	transport   = &http.Transport{} // Reusable transport
	clientMutex sync.Mutex
)

func newDFNSClient(userAuthToken *string) (*http.Client, error) {
	clientMutex.Lock()
	defer clientMutex.Unlock()

	// Configure credentials for authentication
	conf := &credentials.AsymmetricKeySignerConfig{
		PrivateKey: string(config.DFNS_PRIVATE_KEY), // DFNS Credential Private Key
		CredID:     config.GetCredentialID(),        // DFNS Credential ID
	}

	// Create a DFNS API client instance with authentication
	authToken := config.GetAuthToken()
	if userAuthToken != nil {
		authToken = *userAuthToken
	}

	apiOptions, err := api.NewDfnsAPIOptions(&api.DfnsAPIConfig{
		AppID:     config.GetAppID(),   // DFNS Application ID
		AuthToken: &authToken,          // Authentication Token
		BaseURL:   config.GetBaseURL(), // DFNS API Base URL
	}, credentials.NewAsymmetricKeySigner(conf))
	if err != nil {
		return nil, fmt.Errorf("error creating DFNS API options: %w", err)
	}

	// Create a new DFNS client with the reusable transport
	client := api.CreateDfnsAPIClient(apiOptions)
	client.Transport = transport // Attach shared transport

	return client, nil
}

func APIClient[T any](request interface{}, httpMethod, URL string, userAuthToken *string) (*T, error) {
	// Create a new client for each request (but with shared transport)
	dfnsClient, err := newDFNSClient(userAuthToken)
	if err != nil {
		return nil, err
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

	if userAuthToken != nil {
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

	// Decode response JSON into the provided response structure
	var response T
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("error decoding JSON response: %v", err)
	}

	return &response, nil
}
