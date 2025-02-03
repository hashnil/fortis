package config

import (
	"context"
	"fmt"
	"fortis/entity/constants"
	"os"
	"strings"

	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	"cloud.google.com/go/secretmanager/apiv1/secretmanagerpb"
	"github.com/spf13/viper"
)

var (
	APP_ID        = viper.GetString("wallet.dfns.app_id")
	AUTH_TOKEN    = viper.GetString("wallet.dfns.auth_token")
	BASE_URL      = viper.GetString("wallet.dfns.base_url")
	CREDENTIAL_ID = viper.GetString("wallet.dfns.credential_id")

	DFNS_PRIVATE_KEY []byte

	configPath = "./infrastructure/config/"
)

// LoadConfig loads the application configuration.
func LoadConfig() error {
	// Check the environment to determine config source
	if os.Getenv("ENV") == "" {
		// Load configuration from local YAML file
		viper.SetConfigName("config")
		viper.SetConfigType("yaml")
		viper.AddConfigPath(configPath)

		err := viper.ReadInConfig()
		if err != nil {
			return fmt.Errorf("failed to read local config: %v", err)
		}

		// Load the key.pem secret in global variable
		DFNS_PRIVATE_KEY, err = os.ReadFile(configPath + "dfns_key.pem")
		if err != nil {
			return fmt.Errorf("failed to read key.pem: %v", err)
		}
	} else {
		// Fetch configuration from Google Secret Manager
		configYAML, err := fetchSecretFromGCP(constants.APP_CONFIG_SECRET)
		if err != nil {
			return fmt.Errorf("failed to fetch secret: %v", err)
		}

		viper.SetConfigType("yaml")
		if err := viper.ReadConfig(strings.NewReader(configYAML)); err != nil {
			return fmt.Errorf("failed to parse secret `%s`: %v", constants.APP_CONFIG_SECRET, err)
		}

		// Fetch DFNS Private Key from Google Secret Manager
		dfnsPvtKey, err := fetchSecretFromGCP(constants.DFNS_PVT_KEY_SECRET)
		if err != nil {
			return fmt.Errorf("failed to fetch secret `%s`: %v", constants.DFNS_PVT_KEY_SECRET, err)
		}
		DFNS_PRIVATE_KEY = []byte(dfnsPvtKey)
	}

	// Enable automatic environment variable binding
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	return nil
}

// fetchSecretFromGCP fetches the secret from Google Secret Manager
func fetchSecretFromGCP(secretName string) (string, error) {
	ctx := context.Background()
	client, err := secretmanager.NewClient(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to create secret manager client: %v", err)
	}
	defer client.Close()

	// Construct the secret version name (e.g., "projects/<project-id>/secrets/<secret-name>/versions/latest")
	secretVersion := fmt.Sprintf("projects/%s/secrets/%s/versions/latest", constants.PROJECT_ID, secretName)
	req := &secretmanagerpb.AccessSecretVersionRequest{
		Name: secretVersion,
	}

	// Retrieve the secret
	result, err := client.AccessSecretVersion(ctx, req)
	if err != nil {
		return "", fmt.Errorf("failed to access secret: %v", err)
	}

	return string(result.Payload.Data), nil
}
