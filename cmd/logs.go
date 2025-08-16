package cmd

import (
	"coolify-cli/client"
	"coolify-cli/internal/formatter"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

var logsCmd = &cobra.Command{
	Use:   "logs [application-uuid-or-name]",
	Short: "Fetch logs for a Coolify application",
	Long: `Fetch and display logs for a specific Coolify application.
You can provide either the application UUID or name as an argument.
If using a name, it must be unique across all applications.

Examples:
  coolify-cli logs nk4kcskcsswg0wskk88skcsg
  coolify-cli logs my-app-name`,
	Args: cobra.ExactArgs(1),
	RunE: runLogsCommand,
}

var (
	follow     bool
	tail       int
	timestamps bool
	noColor    bool
	compact    bool
	requestIDs bool
	instance   string
)

func init() {
	rootCmd.AddCommand(logsCmd)

	// Add flags for the logs command
	logsCmd.Flags().BoolVarP(&follow, "follow", "f", false, "Follow log output (stream logs)")
	logsCmd.Flags().IntVarP(&tail, "tail", "n", 100, "Number of lines to show from the end of the logs")
	logsCmd.Flags().BoolVarP(&timestamps, "timestamps", "t", true, "Show timestamps")
	logsCmd.Flags().BoolVar(&noColor, "no-color", false, "Disable colored output")
	logsCmd.Flags().BoolVarP(&compact, "compact", "c", false, "Compact output (less spacing)")
	logsCmd.Flags().BoolVarP(&requestIDs, "request-ids", "r", false, "Show request IDs")
	logsCmd.Flags().StringVarP(&instance, "instance", "i", "", "Coolify instance to use (default: use default instance)")
}

func runLogsCommand(cmd *cobra.Command, args []string) error {
	applicationIdentifier := args[0]

	// Create client for the specified instance
	var c *client.Client
	var err error
	if instance != "" {
		c, err = client.NewClientForInstance(instance)
	} else {
		c, err = client.NewClient()
	}
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}

	// Resolve application identifier to UUID
	applicationUUID, err := resolveApplicationIdentifier(c, applicationIdentifier)
	if err != nil {
		return err
	}

	verbose, _ := cmd.Flags().GetBool("verbose")
	if verbose {
		fmt.Printf("Fetching logs for application: %s (UUID: %s)\n", applicationIdentifier, applicationUUID)
	}

	if follow {
		return followLogs(c, applicationUUID, verbose)
	}

	return fetchLogs(c, applicationUUID, verbose)
}

func fetchLogs(c *client.Client, applicationID string, verbose bool) error {
	logs, err := c.GetApplicationLogs(applicationID)
	if err != nil {
		// Check if it's a connection error
		if strings.Contains(err.Error(), "failed to connect") {
			return fmt.Errorf("❌ Connection failed: %w\n\n💡 Troubleshooting:\n  • Check if your Coolify instance is running and accessible\n  • Verify the instance URL is correct: run 'coolify-cli instances list'\n  • Ensure your token is valid: get a new one from /security/api-tokens", err)
		}
		return fmt.Errorf("failed to fetch logs: %w", err)
	}

	if len(logs) == 0 {
		fmt.Println("No logs found for this application.")
		return nil
	}

	// Create formatter
	colorOutput := !noColor && isTerminal()
	logFormatter := formatter.NewLogFormatter(colorOutput, timestamps, requestIDs, compact)

	// Display header
	if verbose {
		fmt.Println(logFormatter.FormatHeader(applicationID))
		if sep := logFormatter.FormatSeparator(); sep != "" {
			fmt.Println(sep)
		}
	}

	// Display logs
	displayLogs(logs, logFormatter)

	return nil
}

func followLogs(c *client.Client, applicationID string, verbose bool) error {
	// Create formatter
	colorOutput := !noColor && isTerminal()
	logFormatter := formatter.NewLogFormatter(colorOutput, timestamps, requestIDs, compact)

	if verbose {
		fmt.Println(logFormatter.FormatHeader(applicationID))
		if sep := logFormatter.FormatSeparator(); sep != "" {
			fmt.Println(sep)
		}
		fmt.Println("Following logs... (Press Ctrl+C to stop)")
		fmt.Println()
	}

	// Implement improved polling with shorter intervals for better responsiveness
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	var seenLines map[string]bool = make(map[string]bool)

	for {
		select {
		case <-ticker.C:
			logs, err := c.GetApplicationLogs(applicationID)
			if err != nil {
				if strings.Contains(err.Error(), "failed to connect") {
					fmt.Printf("❌ Connection lost to Coolify instance. Retrying...\n")
					if verbose {
						fmt.Printf("Details: %v\n", err)
					}
				} else if verbose {
					fmt.Printf("Error fetching logs: %v\n", err)
				}
				continue
			}

			// Show only new unique logs
			var newLogs []client.ParsedLogLine
			for _, log := range logs {
				// Use a combination of timestamp and message as unique identifier
				logKey := fmt.Sprintf("%s_%s", log.Timestamp, log.Message)
				if !seenLines[logKey] {
					seenLines[logKey] = true
					newLogs = append(newLogs, log)
				}
			}

			if len(newLogs) > 0 {
				displayLogs(newLogs, logFormatter)
			}
		}
	}
}

func displayLogs(logs []client.ParsedLogLine, logFormatter *formatter.LogFormatter) {
	for _, log := range logs {
		formattedLine := logFormatter.FormatLogLine(log)
		fmt.Println(formattedLine)

		// Add spacing between logs if not in compact mode
		if !compact {
			// Only add spacing for request/response pairs
			if log.Status != "" || (log.Method != "" && log.URL != "") {
				// Don't add extra spacing
			}
		}
	}
}

// resolveApplicationIdentifier resolves an application identifier (UUID or name) to a UUID
func resolveApplicationIdentifier(c *client.Client, identifier string) (string, error) {
	// If it looks like a UUID (long string), use it directly
	if len(identifier) >= 20 {
		return identifier, nil
	}

	// Otherwise, treat it as a name and look it up
	apps, err := c.GetApplications()
	if err != nil {
		return "", fmt.Errorf("failed to fetch applications: %w", err)
	}

	var matchingApps []string
	for _, app := range apps {
		if app.Name == identifier {
			matchingApps = append(matchingApps, app.UUID)
		}
	}

	if len(matchingApps) == 0 {
		return "", fmt.Errorf("no application found with name '%s'", identifier)
	}

	if len(matchingApps) > 1 {
		return "", fmt.Errorf("multiple applications found with name '%s'. Please use the UUID instead:\n%s",
			identifier, strings.Join(matchingApps, "\n"))
	}

	return matchingApps[0], nil
}

// isTerminal checks if output is going to a terminal (for color detection)
func isTerminal() bool {
	if os.Getenv("TERM") == "dumb" {
		return false
	}

	fileInfo, err := os.Stdout.Stat()
	if err != nil {
		return false
	}

	return (fileInfo.Mode() & os.ModeCharDevice) != 0
}
