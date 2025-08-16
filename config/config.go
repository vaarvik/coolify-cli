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
	return LoadWithValidation(true)
}

// LoadWithoutValidation loads config without validating tokens (useful for management commands)
func LoadWithoutValidation() (*Config, error) {
	return LoadWithValidation(false)
}

// LoadWithValidation reads the configuration with optional validation
func LoadWithValidation(validateTokens bool) (*Config, error) {
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

	// Only validate tokens if requested (skip for management commands)
	if validateTokens {
		// Check that default instance has a token
		defaultInstance := config.GetDefaultInstance()
		if defaultInstance == nil {
			return nil, fmt.Errorf("no default instance found in config")
		}

		if defaultInstance.Token == "" {
			return nil, fmt.Errorf("token is required for default instance '%s'. Please set it in %s", defaultInstance.Name, configPath)
		}
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

	// Create default config structure with only cloud instance
	defaultConfig := Config{
		Instances: []Instance{
			{
				FQDN:    "https://app.coolify.io",
				Name:    "cloud",
				Token:   "",
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

// Save writes the configuration to file
func (c *Config) Save() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get user home directory: %w", err)
	}

	configPath := filepath.Join(homeDir, ".coolify-cli", "config.json")

	// Update last update check time
	c.LastUpdateCheckTime = time.Now()

	// Marshal to pretty JSON
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0600); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// AddInstance adds a new instance to the configuration
func (c *Config) AddInstance(name, fqdn, token string, isDefault bool) error {
	// Check if instance name already exists
	if c.GetInstanceByName(name) != nil {
		return fmt.Errorf("instance '%s' already exists", name)
	}

	// If this is being set as default, unset other defaults
	if isDefault {
		for i := range c.Instances {
			c.Instances[i].Default = false
		}
	}

	// Add new instance
	newInstance := Instance{
		FQDN:    fqdn,
		Name:    name,
		Token:   token,
		Default: isDefault,
	}

	c.Instances = append(c.Instances, newInstance)
	return nil
}

// SetInstanceToken sets the token for an existing instance
func (c *Config) SetInstanceToken(name, token string) error {
	instance := c.GetInstanceByName(name)
	if instance == nil {
		return fmt.Errorf("instance '%s' not found", name)
	}

	instance.Token = token
	return nil
}

// SetDefaultInstance sets an instance as the default
func (c *Config) SetDefaultInstance(name string) error {
	targetInstance := c.GetInstanceByName(name)
	if targetInstance == nil {
		return fmt.Errorf("instance '%s' not found", name)
	}

	// Unset all defaults first
	for i := range c.Instances {
		c.Instances[i].Default = false
	}

	// Set the target as default
	targetInstance.Default = true
	return nil
}

// RemoveInstance removes an instance from the configuration
func (c *Config) RemoveInstance(name string) error {
	for i, instance := range c.Instances {
		if instance.Name == name {
			c.Instances = append(c.Instances[:i], c.Instances[i+1:]...)

			// If we removed the default instance and there are others, make the first one default
			if instance.Default && len(c.Instances) > 0 {
				c.Instances[0].Default = true
			}

			return nil
		}
	}

	return fmt.Errorf("instance '%s' not found", name)
}
