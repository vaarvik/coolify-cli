package cmd

import (
	"coolify-cli/client"
	"coolify-cli/config"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

var instancesCmd = &cobra.Command{
	Use:   "instances",
	Short: "Manage Coolify instances",
	Long:  `Manage your Coolify instances including adding new ones, setting tokens, and changing defaults.`,
}

var instancesAddCmd = &cobra.Command{
	Use:   "add [name] [fqdn] [token]",
	Short: "Add a new Coolify instance",
	Long: `Add a new Coolify instance to your configuration.

Examples:
  coolify-cli instances add myserver https://coolify.mycompany.com my-token-123
  coolify-cli instances add -d myserver https://coolify.mycompany.com my-token-123`,
	Args: cobra.ExactArgs(3),
	RunE: runInstancesAddCommand,
}

var instancesSetCmd = &cobra.Command{
	Use:   "set",
	Short: "Set instance properties",
	Long:  `Set various properties for Coolify instances.`,
}

var instancesSetTokenCmd = &cobra.Command{
	Use:   "token [instance-name] [token]",
	Short: "Set token for an existing instance",
	Long: `Set or update the API token for an existing Coolify instance.

Examples:
  coolify-cli instances set token cloud my-cloud-token-123
  coolify-cli instances set token myserver my-server-token-456`,
	Args: cobra.ExactArgs(2),
	RunE: runInstancesSetTokenCommand,
}

var instancesSetDefaultCmd = &cobra.Command{
	Use:   "default [instance-name]",
	Short: "Set default instance",
	Long: `Set which Coolify instance should be used by default.

Examples:
  coolify-cli instances set default cloud
  coolify-cli instances set default myserver`,
	Args: cobra.ExactArgs(1),
	RunE: runInstancesSetDefaultCommand,
}

var instancesListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all configured instances",
	Long:  `List all configured Coolify instances with their details.`,
	RunE:  runInstancesListCommand,
}

var instancesRemoveCmd = &cobra.Command{
	Use:   "remove [instance-name]",
	Short: "Remove an instance",
	Long: `Remove a Coolify instance from your configuration.

Examples:
  coolify-cli instances remove myserver`,
	Args: cobra.ExactArgs(1),
	RunE: runInstancesRemoveCommand,
}

var (
	makeDefault bool
	skipTest    bool
)

func init() {
	rootCmd.AddCommand(instancesCmd)

	// Add subcommands
	instancesCmd.AddCommand(instancesAddCmd)
	instancesCmd.AddCommand(instancesSetCmd)
	instancesCmd.AddCommand(instancesListCmd)
	instancesCmd.AddCommand(instancesRemoveCmd)

	// Add set subcommands
	instancesSetCmd.AddCommand(instancesSetTokenCmd)
	instancesSetCmd.AddCommand(instancesSetDefaultCmd)

	// Add flags
	instancesAddCmd.Flags().BoolVarP(&makeDefault, "default", "d", false, "Make this instance the default")
	instancesAddCmd.Flags().BoolVar(&skipTest, "skip-test", false, "Skip connection test when adding instance")
}

func runInstancesAddCommand(cmd *cobra.Command, args []string) error {
	name := args[0]
	fqdn := args[1]
	token := args[2]

	cfg, err := config.LoadWithoutValidation()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Test connection before adding (unless skipped)
	if !skipTest {
		fmt.Printf("🧪 Testing connection to %s...\n", fqdn)

		// Create a temporary instance for testing
		tempInstance := &config.Instance{
			FQDN:  fqdn,
			Name:  name,
			Token: token,
		}

		// Create a temporary client
		tempClient := &client.Client{}
		tempClient.SetInstance(tempInstance)

		if err := tempClient.TestConnection(); err != nil {
			if strings.Contains(err.Error(), "failed to connect") {
				fmt.Printf("❌ Connection test failed: Cannot reach %s\n", fqdn)
				fmt.Printf("\n💡 Troubleshooting:\n")
				fmt.Printf("  • Check if the URL is correct and accessible\n")
				fmt.Printf("  • Verify the Coolify instance is running\n")
				fmt.Printf("  • Try accessing %s in your browser\n", fqdn)
				fmt.Printf("\nTo add anyway, use --skip-test flag\n")
				return fmt.Errorf("connection test failed")
			} else if strings.Contains(err.Error(), "401") || strings.Contains(err.Error(), "authentication failed") {
				fmt.Printf("❌ Authentication failed: Invalid token for %s\n", fqdn)
				fmt.Printf("\n💡 Fix this by:\n")
				fmt.Printf("  • Get a valid token from %s/security/api-tokens\n", fqdn)
				fmt.Printf("\nTo add anyway, use --skip-test flag\n")
				return fmt.Errorf("authentication failed")
			} else {
				fmt.Printf("⚠️  Connection test failed: %v\n", err)
				fmt.Printf("Adding instance anyway...\n")
			}
		} else {
			fmt.Printf("✅ Connection test successful!\n")
		}
	}

	if err := cfg.AddInstance(name, fqdn, token, makeDefault); err != nil {
		return fmt.Errorf("failed to add instance: %w", err)
	}

	if err := cfg.Save(); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	fmt.Printf("✅ Added instance '%s' (%s)\n", name, fqdn)
	if makeDefault {
		fmt.Printf("🎯 Set '%s' as the default instance\n", name)
	}

	return nil
}

func runInstancesSetTokenCommand(cmd *cobra.Command, args []string) error {
	name := args[0]
	token := args[1]

	cfg, err := config.LoadWithoutValidation()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if err := cfg.SetInstanceToken(name, token); err != nil {
		return fmt.Errorf("failed to set token: %w", err)
	}

	if err := cfg.Save(); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	fmt.Printf("✅ Updated token for instance '%s'\n", name)

	return nil
}

func runInstancesSetDefaultCommand(cmd *cobra.Command, args []string) error {
	name := args[0]

	cfg, err := config.LoadWithoutValidation()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if err := cfg.SetDefaultInstance(name); err != nil {
		return fmt.Errorf("failed to set default: %w", err)
	}

	if err := cfg.Save(); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	fmt.Printf("🎯 Set '%s' as the default instance\n", name)

	return nil
}

func runInstancesListCommand(cmd *cobra.Command, args []string) error {
	cfg, err := config.LoadWithoutValidation()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	fmt.Println("Configured Coolify Instances:")
	fmt.Println()

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

func runInstancesRemoveCommand(cmd *cobra.Command, args []string) error {
	name := args[0]

	cfg, err := config.LoadWithoutValidation()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if err := cfg.RemoveInstance(name); err != nil {
		return fmt.Errorf("failed to remove instance: %w", err)
	}

	if err := cfg.Save(); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	fmt.Printf("✅ Removed instance '%s'\n", name)

	return nil
}
