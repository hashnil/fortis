package config

import (
	"fmt"
	"fortis/entity/constants"
	"os"
	"strings"

	"github.com/getpanda/commons/pkg/auth/secrets"
	"github.com/spf13/viper"
)

var (
	JWT_SECRET       []byte
	NONCE_SECRET     []byte
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
		configYAML, err := secrets.FetchSecretFromAWS(constants.APP_CONFIG_SECRET)
		if err != nil {
			return fmt.Errorf("failed to fetch secret: %v", err)
		}

		viper.SetConfigType("yaml")
		if err := viper.ReadConfig(strings.NewReader(configYAML)); err != nil {
			return fmt.Errorf("failed to parse secret `%s`: %v", constants.APP_CONFIG_SECRET, err)
		}

		// Fetch DFNS Private Key from AWS Secret Manager
		dfnsPvtKey, err := secrets.FetchSecretFromAWS(constants.DFNS_PVT_KEY_SECRET)
		if err != nil {
			return fmt.Errorf("failed to fetch secret `%s`: %v", constants.DFNS_PVT_KEY_SECRET, err)
		}
		DFNS_PRIVATE_KEY = []byte(dfnsPvtKey)
	}

	// Enable automatic environment variable binding
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Load secrets required for authentication of request
	if err := secrets.LoadNonceSecrets(); err != nil {
		return fmt.Errorf("failed to load authentication secrets: %w", err)
	}
	JWT_SECRET = secrets.JWT_SECRET
	NONCE_SECRET = secrets.NONCE_SECRET

	return nil
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

func GetNetworks() []string {
	return viper.GetStringSlice("wallet.dfns.networks")
}
