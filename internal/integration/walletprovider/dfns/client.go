package dfns

import (
	"bytes"
	"encoding/json"
	"fmt"
	"fortis/infrastructure/config"
	"net/http"

	"github.com/dfns/dfns-sdk-go/credentials"
	api "github.com/dfns/dfns-sdk-go/dfnsapiclient"
)

func APIClient(request, response interface{}, httpMethod, URL string) error {
	// Configure credentials for authentication
	conf := &credentials.AsymmetricKeySignerConfig{
		PrivateKey: config.PRIVATE_KEY, // DFNS Credential Private Key
		CredID:     config.CRED_ID,     // DFNS Credential ID
	}

	// Create a DFNS API client instance with authentication
	apiOptions, err := api.NewDfnsAPIOptions(&api.DfnsAPIConfig{
		AppID:     config.APP_ID,      // DFNS Application ID
		AuthToken: &config.AUTH_TOKEN, // Authentication Token
		BaseURL:   config.BASE_URL,    // DFNS API Base URL
	}, credentials.NewAsymmetricKeySigner(conf))
	if err != nil {
		return fmt.Errorf("error creating DFNS API options: %w", err)
	}

	// Convert request body to JSON format
	jsonData, err := json.Marshal(request)
	if err != nil {
		return fmt.Errorf("error marshaling JSON: %w", err)
	}

	// Create an HTTP request with the given HTTP method and URL
	req, err := http.NewRequest(httpMethod, apiOptions.BaseURL+URL, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("error creating HTTP request: %w", err)
	}

	// Create DFNS API client
	dfnsClient := api.CreateDfnsAPIClient(apiOptions)

	// Execute the request
	resp, err := dfnsClient.Do(req)
	if err != nil {
		return fmt.Errorf("error executing request: %w", err)
	}
	defer resp.Body.Close()

	// Check for HTTP response errors
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("DFNS API error: %s", resp.Status)
	}

	// Decode response JSON into the provided response structure
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return fmt.Errorf("error decoding JSON response: %v", err)
	}
	return nil
}
