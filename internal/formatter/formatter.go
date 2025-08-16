package formatter

import (
	"coolify-cli/client"
	"fmt"
	"strings"
)

// Color codes for terminal output
const (
	Reset  = "\033[0m"
	Bold   = "\033[1m"

	// Text colors
	Red    = "\033[31m"
	Green  = "\033[32m"
	Yellow = "\033[33m"
	Blue   = "\033[34m"
	Purple = "\033[35m"
	Cyan   = "\033[36m"
	Gray   = "\033[37m"
	White  = "\033[97m"

	// Background colors
	BgRed    = "\033[41m"
	BgGreen  = "\033[42m"
	BgYellow = "\033[43m"
)

// LogFormatter handles pretty printing of logs
type LogFormatter struct {
	ShowTimestamps bool
	ShowRequestIDs bool
	ColorOutput    bool
	CompactMode    bool
}

// NewLogFormatter creates a new log formatter
func NewLogFormatter(colorOutput, showTimestamps, showRequestIDs, compactMode bool) *LogFormatter {
	return &LogFormatter{
		ShowTimestamps: showTimestamps,
		ShowRequestIDs: showRequestIDs,
		ColorOutput:    colorOutput,
		CompactMode:    compactMode,
	}
}

// FormatLogLine formats a single log line with colors and structure
func (f *LogFormatter) FormatLogLine(log client.ParsedLogLine) string {
	var parts []string

	// Add timestamp if enabled
	if f.ShowTimestamps && log.Timestamp != "" {
		timestamp := f.colorize(Gray, fmt.Sprintf("[%s]", log.Timestamp))
		parts = append(parts, timestamp)
	}

	// Add log level with color
	if log.Level != "" {
		levelColor := f.getLevelColor(log.Level)
		level := f.colorize(levelColor, fmt.Sprintf("[%s]", log.Level))
		parts = append(parts, level)
	}

	// Add request ID if available and enabled
	if f.ShowRequestIDs && log.RequestID != "" {
		requestID := f.colorize(Purple, fmt.Sprintf("[%s]", log.RequestID[:8])) // Show first 8 chars
		parts = append(parts, requestID)
	}

	// Format the main message based on log type
	message := f.formatMessage(log)
	parts = append(parts, message)

	return strings.Join(parts, " ")
}

// formatMessage formats the main log message with appropriate colors
func (f *LogFormatter) formatMessage(log client.ParsedLogLine) string {
	// HTTP request logs
	if log.Method != "" && log.URL != "" {
		method := f.colorize(f.getMethodColor(log.Method), log.Method)
		url := f.colorize(Cyan, log.URL)
		return fmt.Sprintf("%s %s", method, url)
	}

	// HTTP response logs
	if log.Status != "" {
		statusColor := f.getStatusColor(log.Status)
		status := f.colorize(statusColor, fmt.Sprintf("→ %s", log.Status))
		return status
	}

	// Generic message
	return log.Message
}

// colorize applies color to text if color output is enabled
func (f *LogFormatter) colorize(color, text string) string {
	if !f.ColorOutput {
		return text
	}
	return color + text + Reset
}

// getLevelColor returns the appropriate color for a log level
func (f *LogFormatter) getLevelColor(level string) string {
	switch strings.ToUpper(level) {
	case "ERROR":
		return Red
	case "WARN", "WARNING":
		return Yellow
	case "INFO":
		return Green
	case "DEBUG":
		return Blue
	default:
		return White
	}
}

// getMethodColor returns the appropriate color for HTTP methods
func (f *LogFormatter) getMethodColor(method string) string {
	switch strings.ToUpper(method) {
	case "GET":
		return Green
	case "POST":
		return Blue
	case "PUT":
		return Yellow
	case "DELETE":
		return Red
	case "PATCH":
		return Purple
	default:
		return White
	}
}

// getStatusColor returns the appropriate color for HTTP status codes
func (f *LogFormatter) getStatusColor(status string) string {
	if len(status) == 0 {
		return White
	}

	switch status[0] {
	case '2': // 2xx success
		return Green
	case '3': // 3xx redirect
		return Yellow
	case '4': // 4xx client error
		return Red
	case '5': // 5xx server error
		return BgRed + White
	default:
		return White
	}
}

// FormatHeader creates a formatted header for the log output
func (f *LogFormatter) FormatHeader(appID string) string {
	if !f.ColorOutput {
		return fmt.Sprintf("=== Logs for application: %s ===", appID)
	}

	header := f.colorize(Bold+Cyan, "=== Logs for application: ") +
			 f.colorize(Bold+White, appID) +
			 f.colorize(Bold+Cyan, " ===")
	return header
}

// FormatSeparator creates a separator line
func (f *LogFormatter) FormatSeparator() string {
	if f.CompactMode {
		return ""
	}
	return f.colorize(Gray, strings.Repeat("-", 80))
}
