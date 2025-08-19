package client

import (
	"coolify-cli/config"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// Client represents the Coolify API client
type Client struct {
	httpClient *http.Client
	instance   *config.Instance
}

// LogEntry represents a single log entry from the Coolify API
type LogEntry struct {
	Timestamp string `json:"timestamp"`
	Message   string `json:"message"`
	Level     string `json:"level,omitempty"`
	RequestID string `json:"request_id,omitempty"`
	Method    string `json:"method,omitempty"`
	URL       string `json:"url,omitempty"`
	Status    int    `json:"status,omitempty"`
}

// LogsResponse represents the response from the logs endpoint
type LogsResponse struct {
	Logs string `json:"logs"`
}

// ParsedLogLine represents a parsed log line with extracted information
type ParsedLogLine struct {
	Timestamp string
	Level     string
	RequestID string
	Method    string
	URL       string
	Status    string
	Message   string
	Raw       string
}

// NewClient creates a new Coolify API client using the default instance
func NewClient() (*Client, error) {
	return NewClientForInstance("")
}

// NewClientForInstance creates a new Coolify API client for a specific instance
func NewClientForInstance(instanceName string) (*Client, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	var instance *config.Instance
	if instanceName == "" {
		instance = cfg.GetDefaultInstance()
		if instance == nil {
			return nil, fmt.Errorf("no default instance configured")
		}
	} else {
		instance = cfg.GetInstanceByName(instanceName)
		if instance == nil {
			return nil, fmt.Errorf("instance '%s' not found in config", instanceName)
		}
	}

	return &Client{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		instance: instance,
	}, nil
}

// SetInstance sets the instance for the client (used for testing)
func (c *Client) SetInstance(instance *config.Instance) {
	c.instance = instance
	if c.httpClient == nil {
		c.httpClient = &http.Client{
			Timeout: 30 * time.Second,
		}
	}
}

// makeRequest performs an HTTP request with Bearer token authentication
func (c *Client) makeRequest(method, endpoint string) (*http.Response, error) {
	url := fmt.Sprintf("%s%s", c.instance.GetBaseURL(), endpoint)

	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add Bearer token authentication
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.instance.Token))
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "coolify-cli/1.0")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Coolify instance at %s: %w", c.instance.FQDN, err)
	}

	return resp, nil
}

// Application represents a Coolify application
type Application struct {
	UUID     string                 `json:"uuid"`
	Name     string                 `json:"name"`
	Status   string                 `json:"status"`
	URL      string                 `json:"url,omitempty"`
	RawData  map[string]interface{} `json:"-"` // Store any additional fields from API
}

// GetApplications fetches all applications
func (c *Client) GetApplications() ([]Application, error) {
	resp, err := c.makeRequest("GET", "/applications")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	// First decode into raw JSON to capture all fields
	var rawData []map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&rawData); err != nil {
		return nil, fmt.Errorf("failed to parse applications response: %w", err)
	}

	// Convert to Application structs while preserving raw data
	var apps []Application
	for _, raw := range rawData {
		var app Application
		// Convert back to JSON
		rawJSON, err := json.Marshal(raw)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal raw data: %w", err)
		}
		// Decode into Application struct
		if err := json.Unmarshal(rawJSON, &app); err != nil {
			return nil, fmt.Errorf("failed to unmarshal application: %w", err)
		}
		// Store raw data
		app.RawData = raw
		apps = append(apps, app)
	}

	return apps, nil
}

// GetApplicationLogs fetches logs for a specific application and returns raw log content
func (c *Client) GetApplicationLogs(applicationID string) (string, error) {
	endpoint := fmt.Sprintf("/applications/%s/logs", applicationID)

	resp, err := c.makeRequest("GET", endpoint)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	// Parse the JSON response
	var logsResponse LogsResponse
	if err := json.Unmarshal(body, &logsResponse); err != nil {
		return "", fmt.Errorf("failed to parse logs response: %w", err)
	}

	// Return the raw log content exactly as received
	return logsResponse.Logs, nil
}

// ParseLogContent parses the raw log content and extracts structured information
func (c *Client) ParseLogContent(logContent string) []ParsedLogLine {
	if logContent == "" {
		return []ParsedLogLine{}
	}

	// Remove ANSI escape sequences for cleaner parsing
	ansiRegex := regexp.MustCompile(`\x1b\[[0-9;]*m`)
	cleanContent := ansiRegex.ReplaceAllString(logContent, "")

	lines := strings.Split(cleanContent, "\n")
	var parsedLines []ParsedLogLine

	// Regex patterns for different log formats (updated to match actual Coolify log format)
	// HTTP request pattern: INFO (18): uuid GET http://... - N query params, N body keys
	httpRequestPattern := regexp.MustCompile(`INFO \((\d+)\): ([a-f0-9-]+) (GET|POST|PUT|DELETE|PATCH) (http://[^\s]+) - (\d+) query params, (\d+) body keys`)
	// HTTP response pattern: INFO (18): uuid Response: 200
	httpResponsePattern := regexp.MustCompile(`INFO \((\d+)\): ([a-f0-9-]+) Response: (\d+)`)
	// Auth pattern: INFO (18): uuid Auth via Bearer Token
	authPattern := regexp.MustCompile(`INFO \((\d+)\): ([a-f0-9-]+) Auth via Bearer Token`)
	// Generic INFO pattern: INFO (18): uuid message
	genericInfoPattern := regexp.MustCompile(`INFO \((\d+)\): ([a-f0-9-]+) (.+)`)
	// TraceId pattern: traceId: "uuid"
	traceIdPattern := regexp.MustCompile(`^\s*traceId: "([a-f0-9-]+)"`)
	// Generic INFO without request ID: INFO (18): message
	simpleInfoPattern := regexp.MustCompile(`INFO \((\d+)\): (.+)`)

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		parsedLine := ParsedLogLine{
			Raw:       line,
			Timestamp: c.extractTimestamp(line),
		}

		// Try to match different log patterns in order of specificity
		if matches := httpRequestPattern.FindStringSubmatch(line); len(matches) > 0 {
			// HTTP request: INFO (18): uuid GET http://... - N query params, N body keys
			parsedLine.Level = "INFO"
			parsedLine.RequestID = matches[2]
			parsedLine.Method = matches[3]
			parsedLine.URL = matches[4]
			parsedLine.Message = fmt.Sprintf("%s %s", matches[3], matches[4])
		} else if matches := httpResponsePattern.FindStringSubmatch(line); len(matches) > 0 {
			// HTTP response: INFO (18): uuid Response: 200
			parsedLine.Level = "INFO"
			parsedLine.RequestID = matches[2]
			parsedLine.Status = matches[3]
			parsedLine.Message = fmt.Sprintf("Response: %s", matches[3])
		} else if matches := authPattern.FindStringSubmatch(line); len(matches) > 0 {
			// Auth: INFO (18): uuid Auth via Bearer Token
			parsedLine.Level = "INFO"
			parsedLine.RequestID = matches[2]
			parsedLine.Message = "Auth via Bearer Token"
		} else if matches := traceIdPattern.FindStringSubmatch(line); len(matches) > 0 {
			// TraceId: traceId: "uuid"
			parsedLine.Level = "INFO"
			parsedLine.RequestID = matches[1]
			parsedLine.Message = fmt.Sprintf("traceId: \"%s\"", matches[1])
		} else if matches := genericInfoPattern.FindStringSubmatch(line); len(matches) > 0 {
			// Generic INFO with request ID: INFO (18): uuid message
			parsedLine.Level = "INFO"
			parsedLine.RequestID = matches[2]
			parsedLine.Message = matches[3]
		} else if matches := simpleInfoPattern.FindStringSubmatch(line); len(matches) > 0 {
			// Simple INFO without request ID: INFO (18): message
			parsedLine.Level = "INFO"
			parsedLine.Message = matches[2]
		} else {
			// Generic log line - extract timestamp and keep as-is
			parsedLine.Level = "INFO"
			// Remove timestamp from message if it was at the beginning
			message := line
			if timestampMatch := regexp.MustCompile(`^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}(\.\d+)?Z?\s*`).FindString(line); timestampMatch != "" {
				message = strings.TrimSpace(strings.TrimPrefix(line, timestampMatch))
			}
			parsedLine.Message = message
		}

		parsedLines = append(parsedLines, parsedLine)
	}

	return parsedLines
}

// extractTimestamp attempts to extract a timestamp from a log line
// Supports multiple common timestamp formats and falls back to current time
func (c *Client) extractTimestamp(line string) string {
	// Common timestamp patterns (ordered by specificity)
	timestampPatterns := []struct {
		regex  *regexp.Regexp
		layout string
	}{
		// Coolify format with nanoseconds: 2025-08-19T06:49:35.131504808Z
		{regexp.MustCompile(`^(\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}\.\d{9}Z)`), "2006-01-02T15:04:05.000000000Z"},
		// ISO 8601 with microseconds: 2024-01-15T14:30:45.123456Z
		{regexp.MustCompile(`^(\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}\.\d{6}Z?)`), "2006-01-02T15:04:05.000000Z"},
		// ISO 8601 with milliseconds: 2024-01-15T14:30:45.123Z
		{regexp.MustCompile(`^(\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}\.\d{3}Z?)`), "2006-01-02T15:04:05.000Z"},
		// ISO 8601 basic: 2024-01-15T14:30:45Z
		{regexp.MustCompile(`^(\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}Z?)`), "2006-01-02T15:04:05Z"},
		// ISO 8601 with timezone: 2024-01-15T14:30:45+00:00
		{regexp.MustCompile(`^(\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}[+-]\d{2}:\d{2})`), "2006-01-02T15:04:05-07:00"},
		// Docker/Container logs: 2024-01-15 14:30:45
		{regexp.MustCompile(`^(\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2})`), "2006-01-02 15:04:05"},
		// Syslog format: Jan 15 14:30:45
		{regexp.MustCompile(`^([A-Za-z]{3} \d{1,2} \d{2}:\d{2}:\d{2})`), "Jan 2 15:04:05"},
		// Unix timestamp with brackets: [1705329045]
		{regexp.MustCompile(`^\[(\d{10})\]`), "unix"},
		// Timestamp at beginning with brackets: [2024-01-15 14:30:45]
		{regexp.MustCompile(`^\[(\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2})\]`), "2006-01-02 15:04:05"},
		// Timestamp at beginning: 2024/01/15 14:30:45
		{regexp.MustCompile(`^(\d{4}/\d{2}/\d{2} \d{2}:\d{2}:\d{2})`), "2006/01/02 15:04:05"},
	}

	for _, pattern := range timestampPatterns {
		if matches := pattern.regex.FindStringSubmatch(line); len(matches) > 1 {
			timestampStr := matches[1]

			// Handle Unix timestamp specially
			if pattern.layout == "unix" {
				if unixSeconds, err := strconv.ParseInt(timestampStr, 10, 64); err == nil {
					unixTime := time.Unix(unixSeconds, 0)
					return unixTime.Format("2006-01-02 15:04:05")
				}
				continue
			}

			// Try to parse with the specified layout
			if parsedTime, err := time.Parse(pattern.layout, timestampStr); err == nil {
				return parsedTime.Format("2006-01-02 15:04:05")
			}

			// If parsing fails, try with current year for formats like "Jan 15 14:30:45"
			if pattern.layout == "Jan 2 15:04:05" {
				currentYear := time.Now().Year()
				fullTimestamp := fmt.Sprintf("%d %s", currentYear, timestampStr)
				if parsedTime, err := time.Parse("2006 Jan 2 15:04:05", fullTimestamp); err == nil {
					return parsedTime.Format("2006-01-02 15:04:05")
				}
			}
		}
	}

	// If no timestamp pattern matches, fall back to current time
	return time.Now().Format("2006-01-02 15:04:05")
}

// TestConnection tests the connection to the Coolify API
func (c *Client) TestConnection() error {
	// Try a simple request to test authentication
	// Use /applications endpoint which is more likely to exist
	resp, err := c.makeRequest("GET", "/applications")
	if err != nil {
		return fmt.Errorf("connection test failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return fmt.Errorf("authentication failed: invalid API key")
	}

	// Accept 200, 404, or other non-auth errors as successful connection
	// The important thing is that we can reach the API and authenticate
	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API returned error %d: %s", resp.StatusCode, string(body))
	}

	return nil
}
