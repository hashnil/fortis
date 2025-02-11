package config

import (
	"context"
	"fmt"
	"fortis/entity/constants"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/spf13/viper"
)

var (
	DFNS_PRIVATE_KEY []byte

	configPath = "./infrastructure/config/"
)

// LoadConfig loads the application configuration.
func LoadConfig() error {
	// Check the environment to determine config source
	if os.Getenv("ENV") == "develop" {
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
		// Fetch configuration from AWS Secret Manager
		configYAML, err := fetchSecretFromAWS(constants.APP_CONFIG_SECRET)
		if err != nil {
			return fmt.Errorf("failed to fetch secret: %v", err)
		}

		viper.SetConfigType("yaml")
		if err := viper.ReadConfig(strings.NewReader(configYAML)); err != nil {
			return fmt.Errorf("failed to parse secret `%s`: %v", constants.APP_CONFIG_SECRET, err)
		}

		// Fetch DFNS Private Key from AWS Secret Manager
		dfnsPvtKey, err := fetchSecretFromAWS(constants.DFNS_PVT_KEY_SECRET)
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

// fetchSecretFromAWS fetches the secret from AWS Secrets Manager
func fetchSecretFromAWS(secretName string) (string, error) {
	ctx := context.Background()

	// Load AWS configuration
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to load AWS config: %v", err)
	}

	client := secretsmanager.NewFromConfig(cfg)

	// Request to get the secret value
	req := &secretsmanager.GetSecretValueInput{
		SecretId: aws.String(secretName),
	}

	result, err := client.GetSecretValue(ctx, req)
	if err != nil {
		return "", fmt.Errorf("failed to retrieve secret `%s`: %v", secretName, err)
	}

	return aws.ToString(result.SecretString), nil
}

func GetAppID() string {
	return viper.GetString("wallet.dfns.app_id")
}

func GetAuthToken() string {
	return viper.GetString("wallet.dfns.auth_token")
}

func GetBaseURL() string {
	return viper.GetString("wallet.dfns.base_url")
}

func GetCredentialID() string {
	return viper.GetString("wallet.dfns.credential_id")
}
