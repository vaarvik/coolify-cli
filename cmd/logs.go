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

	if logs == "" {
		fmt.Println("No logs found for this application.")
		return nil
	}

	// Create formatter for beautiful output
	colorOutput := !noColor && isTerminal()
	logFormatter := formatter.NewLogFormatter(colorOutput, timestamps, requestIDs, compact)

	// Display header
	if verbose {
		fmt.Println(logFormatter.FormatHeader(applicationID))
		if sep := logFormatter.FormatSeparator(); sep != "" {
			fmt.Println(sep)
		}
	}

	// Format and display the raw logs beautifully
	displayFormattedLogs(logs, logFormatter)

	return nil
}

func followLogs(c *client.Client, applicationID string, verbose bool) error {
	// Create formatter for beautiful output
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

	// Poll and print only new log lines
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	var lastLine string
	var prevLen int
	var initialized bool

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

			lines := strings.Split(logs, "\n")
			// Drop trailing empty line (common with newline-terminated payloads)
			if len(lines) > 0 && strings.TrimSpace(lines[len(lines)-1]) == "" {
				lines = lines[:len(lines)-1]
			}
			if len(lines) == 0 {
				continue
			}

			startIdx := 0
			if !initialized {
				if tail > 0 && len(lines) > tail {
					startIdx = len(lines) - tail
				}
			} else {
				// Prefer anchor by content: search from end for the last line we printed
				if lastLine != "" {
					for i := len(lines) - 1; i >= 0; i-- {
						if strings.TrimSpace(lines[i]) == strings.TrimSpace(lastLine) {
							startIdx = i + 1
							break
						}
					}
				}
				if startIdx == 0 { // anchor not found
					if len(lines) > prevLen {
						// Assume pure append: print the delta by length
						startIdx = prevLen
					} else {
						// Likely rotation/reset: print a reasonable tail
						if tail > 0 && len(lines) > tail {
							startIdx = len(lines) - tail
						} else {
							startIdx = 0
						}
					}
				}
			}

			if startIdx < len(lines) {
				segment := strings.Join(lines[startIdx:], "\n")
				displayFormattedLogs(segment, logFormatter)
			}

			lastLine = lines[len(lines)-1]
			prevLen = len(lines)
			initialized = true
		}
	}
}

// displayFormattedLogs takes raw log content and applies beautiful formatting
func displayFormattedLogs(rawLogs string, logFormatter *formatter.LogFormatter) {
	if rawLogs == "" {
		return
	}

	// Split raw logs into individual lines
	lines := strings.Split(rawLogs, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Parse each line for formatting while keeping the original content
		parsedLine := parseLogLine(line)

		// Format and display the line beautifully
		formattedLine := logFormatter.FormatLogLine(parsedLine)
		fmt.Println(formattedLine)

		// Add spacing between logs if not in compact mode
		if !compact {
			// Only add spacing for request/response pairs
			if parsedLine.Status != "" || (parsedLine.Method != "" && parsedLine.URL != "") {
				// Don't add extra spacing for now
			}
		}
	}
}

// parseLogLine parses a single raw log line into structured data for formatting
func parseLogLine(line string) client.ParsedLogLine {
	// Create a temporary client to use the existing parsing logic
	c := &client.Client{}

	// Use the existing parsing logic but for a single line
	parsed := c.ParseLogContent(line)
	if len(parsed) > 0 {
		return parsed[0]
	}

	// Fallback - return raw line
	return client.ParsedLogLine{
		Raw:     line,
		Level:   "INFO",
		Message: line,
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
