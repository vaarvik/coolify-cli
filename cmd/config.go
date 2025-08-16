package cmd

import (
	"coolify-cli/client"
	"coolify-cli/config"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage Coolify CLI configuration",
	Long:  `Manage your Coolify CLI configuration including API key and base URL.`,
}

var configShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show current configuration",
	Long:  `Display the current configuration settings.`,
	RunE:  runConfigShowCommand,
}

var configTestCmd = &cobra.Command{
	Use:   "test",
	Short: "Test connection to Coolify API",
	Long:  `Test the connection to your Coolify instance using the configured API key.`,
	RunE:  runConfigTestCommand,
}

var configInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize configuration file",
	Long:  `Create a new configuration file with default settings.`,
	RunE:  runConfigInitCommand,
}

func init() {
	rootCmd.AddCommand(configCmd)
	configCmd.AddCommand(configShowCmd)
	configCmd.AddCommand(configTestCmd)
	configCmd.AddCommand(configInitCmd)
}

func runConfigShowCommand(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	fmt.Println("Current Configuration:")
	fmt.Printf("  Host URL:   %s\n", cfg.HostURL)
	fmt.Printf("  API Path:   %s\n", cfg.APIPath)
	fmt.Printf("  Full URL:   %s\n", cfg.GetBaseURL())

	// Mask the API key for security
	apiKey := cfg.APIKey
	if len(apiKey) > 8 {
		apiKey = apiKey[:4] + "..." + apiKey[len(apiKey)-4:]
	}
	fmt.Printf("  API Key:  %s\n", apiKey)

	// Show config file location
	homeDir, _ := os.UserHomeDir()
	configPath := filepath.Join(homeDir, ".coolify-cli", "config.yaml")
	fmt.Printf("  Config:   %s\n", configPath)

	return nil
}

func runConfigTestCommand(cmd *cobra.Command, args []string) error {
	fmt.Println("Testing connection to Coolify API...")

	c, err := client.NewClient()
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}

	if err := c.TestConnection(); err != nil {
		fmt.Printf("❌ Connection failed: %v\n", err)
		return err
	}

	fmt.Println("✅ Connection successful!")
	return nil
}

func runConfigInitCommand(cmd *cobra.Command, args []string) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get user home directory: %w", err)
	}

	configDir := filepath.Join(homeDir, ".coolify-cli")
	configPath := filepath.Join(configDir, "config.yaml")

	// Check if config file already exists
	if _, err := os.Stat(configPath); err == nil {
		fmt.Printf("Configuration file already exists at: %s\n", configPath)
		fmt.Println("Use 'coolify-cli config show' to view current settings.")
		return nil
	}

	// Create config directory if it doesn't exist
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

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
		return fmt.Errorf("failed to create config file: %w", err)
	}

	fmt.Printf("✅ Configuration file created at: %s\n", configPath)
	fmt.Println("📝 Please edit the file and set your API key.")
	fmt.Println("🧪 Use 'coolify-cli config test' to verify your configuration.")

	return nil
}
