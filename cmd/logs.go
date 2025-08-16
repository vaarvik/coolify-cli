package cmd

import (
	"coolify-cli/client"
	"coolify-cli/internal/formatter"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
)

var logsCmd = &cobra.Command{
	Use:   "logs [application-id]",
	Short: "Fetch logs for a Coolify application",
	Long: `Fetch and display logs for a specific Coolify application.
You need to provide the application ID as an argument.

Example:
  coolify-cli logs nk4kcskcsswg0wskk88skcsg`,
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
}

func runLogsCommand(cmd *cobra.Command, args []string) error {
	applicationID := args[0]

	// Validate application ID format (basic validation)
	if len(applicationID) < 10 {
		return fmt.Errorf("invalid application ID format: %s", applicationID)
	}

	// Create client
	c, err := client.NewClient()
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}

	// Skip connection test and try to fetch logs directly

	verbose, _ := cmd.Flags().GetBool("verbose")
	if verbose {
		fmt.Printf("Fetching logs for application: %s\n", applicationID)
	}

	if follow {
		return followLogs(c, applicationID, verbose)
	}

	return fetchLogs(c, applicationID, verbose)
}

func fetchLogs(c *client.Client, applicationID string, verbose bool) error {
	logs, err := c.GetApplicationLogs(applicationID)
	if err != nil {
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
				if verbose {
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
