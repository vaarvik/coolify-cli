package cmd

import (
	"coolify-cli/client"
	"coolify-cli/config"
	"fmt"
	"os"
	"path/filepath"
	"strings"

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
	cfg, err := config.LoadWithoutValidation()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	homeDir, _ := os.UserHomeDir()
	fmt.Println("Current Configuration:")
	fmt.Printf("  Config file: %s\n", filepath.Join(homeDir, ".coolify-cli", "config.json"))
	fmt.Printf("  Last update check: %s\n", cfg.LastUpdateCheckTime.Format("2006-01-02 15:04:05"))
	fmt.Println("\nInstances:")

	for i, instance := range cfg.Instances {
		prefix := "  "
		if instance.Default {
			prefix = "* "
		}
		fmt.Printf("%s[%d] %s\n", prefix, i+1, instance.Name)
		fmt.Printf("    FQDN: %s\n", instance.FQDN)
		fmt.Printf("    Full URL: %s\n", instance.GetBaseURL())

		// Mask the token for security
		token := instance.Token
		if len(token) > 8 {
			token = token[:4] + "..." + token[len(token)-4:]
		} else if token == "" {
			token = "(not configured)"
		}
		fmt.Printf("    Token: %s\n", token)
		fmt.Println()
	}

	return nil
}

func runConfigTestCommand(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	defaultInstance := cfg.GetDefaultInstance()
	fmt.Printf("Testing connection to Coolify instance '%s' at %s...\n", defaultInstance.Name, defaultInstance.FQDN)

	c, err := client.NewClient()
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}

	if err := c.TestConnection(); err != nil {
		if strings.Contains(err.Error(), "failed to connect") {
			fmt.Printf("❌ Connection failed: Cannot reach Coolify instance\n")
			fmt.Printf("🔗 Instance: %s (%s)\n", defaultInstance.Name, defaultInstance.FQDN)
			fmt.Printf("\n💡 Troubleshooting:\n")
			fmt.Printf("  • Check if the instance URL is correct and accessible\n")
			fmt.Printf("  • Verify the instance is running and not behind a firewall\n")
			fmt.Printf("  • Try accessing %s in your browser\n", defaultInstance.FQDN)
			fmt.Printf("  • Check your internet connection\n")
		} else if strings.Contains(err.Error(), "401") || strings.Contains(err.Error(), "authentication failed") {
			fmt.Printf("❌ Authentication failed: Invalid or expired token\n")
			fmt.Printf("🔑 Instance: %s (%s)\n", defaultInstance.Name, defaultInstance.FQDN)
			fmt.Printf("\n💡 Fix this by:\n")
			fmt.Printf("  • Get a new token from %s/security/api-tokens\n", defaultInstance.FQDN)
			fmt.Printf("  • Update it with: coolify-cli instances set token %s <new-token>\n", defaultInstance.Name)
		} else {
			fmt.Printf("❌ Connection failed: %v\n", err)
		}
		return err
	}

	fmt.Printf("✅ Connection successful to %s!\n", defaultInstance.Name)
	return nil
}

func runConfigInitCommand(cmd *cobra.Command, args []string) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get user home directory: %w", err)
	}

	configDir := filepath.Join(homeDir, ".coolify-cli")
	configPath := filepath.Join(configDir, "config.json")

	// Check if config file already exists
	if _, err := os.Stat(configPath); err == nil {
		fmt.Printf("Configuration file already exists at: %s\n", configPath)
		fmt.Println("Use 'coolify-cli config show' to view current settings.")
		return nil
	}

	// Try to create the default config (this will handle directory creation)
	_, err = config.Load() // This will trigger createDefaultConfig if file doesn't exist
	if err != nil {
		// Expected error when config is created but not configured
		if strings.Contains(err.Error(), "please configure your tokens") {
			fmt.Printf("✅ Configuration file created at: %s\n", configPath)
			fmt.Println("📝 Please edit the file and set your tokens for the instances you want to use.")
			fmt.Println("🧪 Use 'coolify-cli config test' to verify your configuration.")
			return nil
		}
		return err
	}

	return nil
}
