package config

import (
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

// Config holds the CLI configuration
type Config struct {
	AppKey       string `mapstructure:"app_key"`
	SessionToken string `mapstructure:"session_token"`
	BaseURL      string `mapstructure:"base_url"`
}

// DefaultBaseURL is the default Anytype local API URL
const DefaultBaseURL = "http://localhost:31009"

// LoadConfig loads the configuration from config file and environment variables
func LoadConfig() (*Config, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	configDir := filepath.Join(home, ".anytype-cli")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return nil, err
	}

	configFilePath := filepath.Join(configDir, "config")

	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(configDir)

	// Environment variables
	viper.SetEnvPrefix("ANYTYPE")
	viper.AutomaticEnv()

	// Set defaults
	viper.SetDefault("base_url", DefaultBaseURL)

	// If config file doesn't exist, create it
	if _, err := os.Stat(configFilePath + ".yaml"); os.IsNotExist(err) {
		defaultConfig := Config{
			BaseURL: DefaultBaseURL,
		}
		viper.Set("base_url", defaultConfig.BaseURL)
		if err := viper.SafeWriteConfig(); err != nil {
			return nil, err
		}
	} else {
		// Read the config file
		if err := viper.ReadInConfig(); err != nil {
			return nil, err
		}
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, err
	}

	return &config, nil
}

// SaveConfig saves the configuration to disk
func SaveConfig(config *Config) error {
	viper.Set("app_key", config.AppKey)
	viper.Set("session_token", config.SessionToken)
	viper.Set("base_url", config.BaseURL)
	return viper.WriteConfig()
}
