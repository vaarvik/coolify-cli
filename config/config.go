package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Instance represents a single Coolify instance configuration
type Instance struct {
	FQDN    string `json:"fqdn"`
	Name    string `json:"name"`
	Token   string `json:"token"`
	Default bool   `json:"default,omitempty"`
}

// Config represents the CLI configuration structure
type Config struct {
	Instances             []Instance `json:"instances"`
	LastUpdateCheckTime   time.Time  `json:"lastupdatechecktime"`
}

// GetDefaultInstance returns the default instance or the first one if no default is set
func (c *Config) GetDefaultInstance() *Instance {
	// First, look for an instance marked as default
	for i := range c.Instances {
		if c.Instances[i].Default {
			return &c.Instances[i]
		}
	}
	
	// If no default is set, return the first instance
	if len(c.Instances) > 0 {
		return &c.Instances[0]
	}
	
	return nil
}

// GetInstanceByName returns an instance by name
func (c *Config) GetInstanceByName(name string) *Instance {
	for i := range c.Instances {
		if c.Instances[i].Name == name {
			return &c.Instances[i]
		}
	}
	return nil
}

// GetBaseURL returns the complete base URL for API calls for an instance
func (i *Instance) GetBaseURL() string {
	return i.FQDN + "/api/v1"
}

var globalConfig *Config

// Load reads the configuration from the config file
func Load() (*Config, error) {
	if globalConfig != nil {
		return globalConfig, nil
	}

	// Get config file path
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get user home directory: %w", err)
	}
	
	configDir := filepath.Join(homeDir, ".coolify-cli")
	configPath := filepath.Join(configDir, "config.json")

	// Try to read config file
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// Config file not found, create default config
		return createDefaultConfig(configDir)
	}

	// Read and parse JSON config
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config JSON: %w", err)
	}

	// Validate that we have at least one instance
	if len(config.Instances) == 0 {
		return nil, fmt.Errorf("no Coolify instances configured. Please add at least one instance to %s", configPath)
	}

	// Check that default instance has a token
	defaultInstance := config.GetDefaultInstance()
	if defaultInstance == nil {
		return nil, fmt.Errorf("no default instance found in config")
	}
	
	if defaultInstance.Token == "" {
		return nil, fmt.Errorf("token is required for default instance '%s'. Please set it in %s", defaultInstance.Name, configPath)
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

	configPath := filepath.Join(configDir, "config.json")

	// Create default config structure
	defaultConfig := Config{
		Instances: []Instance{
			{
				FQDN:    "https://app.coolify.io",
				Name:    "cloud",
				Token:   "",
			},
			{
				FQDN:    "http://localhost:8000",
				Name:    "localhost",
				Token:   "",
			},
			{
				FQDN:    "https://coolify.yourdomain.com",
				Name:    "yourdomain",
				Token:   "your-token-here",
				Default: true,
			},
		},
		LastUpdateCheckTime: time.Now(),
	}

	// Marshal to pretty JSON
	data, err := json.MarshalIndent(defaultConfig, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal default config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0600); err != nil {
		return nil, fmt.Errorf("failed to create default config file: %w", err)
	}

	fmt.Printf("Created default config file at: %s\n", configPath)
	fmt.Println("Please edit the config file and set your tokens for the instances you want to use.")

	return nil, fmt.Errorf("please configure your tokens in %s", configPath)
}

// GetConfig returns the loaded configuration
func GetConfig() *Config {
	return globalConfig
}
