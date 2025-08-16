package cmd

import (
	"coolify-cli/client"
	"fmt"

	"github.com/spf13/cobra"
)

var applicationsCmd = &cobra.Command{
	Use:     "applications",
	Aliases: []string{"apps", "app"},
	Short:   "Manage Coolify applications",
	Long:    `List and manage your Coolify applications.`,
}

var applicationsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all applications",
	Long:  `List all applications in your Coolify instance.`,
	RunE:  runApplicationsListCommand,
}

var showRaw bool

func init() {
	rootCmd.AddCommand(applicationsCmd)
	applicationsCmd.AddCommand(applicationsListCmd)

	// Add flags
	applicationsListCmd.Flags().BoolVar(&showRaw, "raw", false, "Show all raw data from API")
}

func runApplicationsListCommand(cmd *cobra.Command, args []string) error {
	c, err := client.NewClient()
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}

	apps, err := c.GetApplications()
	if err != nil {
		return fmt.Errorf("failed to fetch applications: %w", err)
	}

	if len(apps) == 0 {
		fmt.Println("No applications found.")
		return nil
	}

	fmt.Println("Applications:")
	for _, app := range apps {
		fmt.Printf("  • %s\n", app.Name)
		fmt.Printf("    UUID: %s\n", app.UUID)
		fmt.Printf("    Status: %s\n", app.Status)
		if app.URL != "" {
			fmt.Printf("    URL: %s\n", app.URL)
		}

		if showRaw {
			fmt.Println("    Raw Data:")
			for key, value := range app.RawData {
				// Skip fields we already showed
				if key == "uuid" || key == "name" || key == "status" || key == "url" {
					continue
				}
				fmt.Printf("      %s: %v\n", key, value)
			}
		}
		fmt.Println()
	}

	return nil
}
