package dfns

import (
	"bytes"
	"encoding/json"
	"fmt"
	"fortis/infrastructure/config"
	"io"
	"net/http"

	"github.com/dfns/dfns-sdk-go/credentials"
	api "github.com/dfns/dfns-sdk-go/dfnsapiclient"
)

func APIClient[T any](request *T, httpMethod, URL string) (*T, error) {
	// Configure credentials for authentication
	conf := &credentials.AsymmetricKeySignerConfig{
		PrivateKey: string(config.DFNS_PRIVATE_KEY), // DFNS Credential Private Key
		CredID:     config.CREDENTIAL_ID,            // DFNS Credential ID
	}

	// Create a DFNS API client instance with authentication
	apiOptions, err := api.NewDfnsAPIOptions(&api.DfnsAPIConfig{
		AppID:     config.APP_ID,      // DFNS Application ID
		AuthToken: &config.AUTH_TOKEN, // Authentication Token
		BaseURL:   config.BASE_URL,    // DFNS API Base URL
	}, credentials.NewAsymmetricKeySigner(conf))
	if err != nil {
		return nil, fmt.Errorf("error creating DFNS API options: %w", err)
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
	req, err := http.NewRequest(httpMethod, apiOptions.BaseURL+URL, requestBody)
	if err != nil {
		return nil, fmt.Errorf("error creating HTTP request: %w", err)
	}

	// Create DFNS API client
	dfnsClient := api.CreateDfnsAPIClient(apiOptions)

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
