package client

import (
	"coolify-cli/config"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
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

// GetApplicationLogs fetches logs for a specific application
func (c *Client) GetApplicationLogs(applicationID string) ([]ParsedLogLine, error) {
	endpoint := fmt.Sprintf("/applications/%s/logs", applicationID)

	resp, err := c.makeRequest("GET", endpoint)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Parse the JSON response
	var logsResponse LogsResponse
	if err := json.Unmarshal(body, &logsResponse); err != nil {
		return nil, fmt.Errorf("failed to parse logs response: %w", err)
	}

	// Parse the log content
	return c.parseLogContent(logsResponse.Logs), nil
}

// parseLogContent parses the raw log content and extracts structured information
func (c *Client) parseLogContent(logContent string) []ParsedLogLine {
	if logContent == "" {
		return []ParsedLogLine{}
	}

	// Remove ANSI escape sequences for cleaner parsing
	ansiRegex := regexp.MustCompile(`\x1b\[[0-9;]*m`)
	cleanContent := ansiRegex.ReplaceAllString(logContent, "")

	lines := strings.Split(cleanContent, "\n")
	var parsedLines []ParsedLogLine

	// Regex patterns for different log formats
	infoPattern := regexp.MustCompile(`INFO \((\d+)\): ([a-f0-9-]+) (GET|POST|PUT|DELETE|PATCH) (http://[^\s]+) - (\d+) query params, (\d+) body keys`)
	responsePattern := regexp.MustCompile(`INFO \((\d+)\): ([a-f0-9-]+) Response: (\d+)`)

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		parsedLine := ParsedLogLine{
			Raw:       line,
			Timestamp: time.Now().Format("2006-01-02 15:04:05"),
		}

		// Try to match INFO request pattern
		if matches := infoPattern.FindStringSubmatch(line); len(matches) > 0 {
			parsedLine.Level = "INFO"
			parsedLine.RequestID = matches[2]
			parsedLine.Method = matches[3]
			parsedLine.URL = matches[4]
			parsedLine.Message = fmt.Sprintf("%s %s", matches[3], matches[4])
		} else if matches := responsePattern.FindStringSubmatch(line); len(matches) > 0 {
			// Match response pattern
			parsedLine.Level = "INFO"
			parsedLine.RequestID = matches[2]
			parsedLine.Status = matches[3]
			parsedLine.Message = fmt.Sprintf("Response: %s", matches[3])
		} else {
			// Generic log line
			parsedLine.Level = "INFO"
			parsedLine.Message = line
		}

		parsedLines = append(parsedLines, parsedLine)
	}

	return parsedLines
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
