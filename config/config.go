package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

// Config represents the CLI configuration structure
type Config struct {
	APIKey  string `mapstructure:"api_key" yaml:"api_key"`
	HostURL string `mapstructure:"host_url" yaml:"host_url"`
	APIPath string `mapstructure:"api_path" yaml:"api_path"`
}

// GetBaseURL returns the complete base URL for API calls
func (c *Config) GetBaseURL() string {
	return c.HostURL + c.APIPath
}

var globalConfig *Config

// Load reads the configuration from the config file
func Load() (*Config, error) {
	if globalConfig != nil {
		return globalConfig, nil
	}

	// Set config file name and paths
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")

	// Add config search paths
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get user home directory: %w", err)
	}

	configDir := filepath.Join(homeDir, ".coolify-cli")
	viper.AddConfigPath(configDir)
	viper.AddConfigPath(".")

	// Set defaults
	viper.SetDefault("host_url", "https://app.coolify.io")
	viper.SetDefault("api_path", "/api/v1")

	// Try to read config file
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Config file not found, create default config
			return createDefaultConfig(configDir)
		}
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	if config.APIKey == "" {
		return nil, fmt.Errorf("API key is required. Please set it in the config file at %s/config.yaml", configDir)
	}

	globalConfig = &config
	return globalConfig, nil
}

// createDefaultConfig creates a default configuration file
func createDefaultConfig(configDir string) (*Config, error) {
	// Create config directory if it doesn't exist
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create config directory: %w", err)
	}

	configPath := filepath.Join(configDir, "config.yaml")

	// Create default config content
	defaultConfig := `# Coolify CLI Configuration
# Get your API key from your Coolify instance
api_key: "your-api-key-here"

# Your Coolify instance URL (without /api/v1)
host_url: "https://app.coolify.io"

# API path (usually /api/v1)
api_path: "/api/v1"
`

	if err := os.WriteFile(configPath, []byte(defaultConfig), 0600); err != nil {
		return nil, fmt.Errorf("failed to create default config file: %w", err)
	}

	fmt.Printf("Created default config file at: %s\n", configPath)
	fmt.Println("Please edit the config file and set your API key.")

	return nil, fmt.Errorf("please configure your API key in %s", configPath)
}

// GetConfig returns the loaded configuration
func GetConfig() *Config {
	return globalConfig
}
