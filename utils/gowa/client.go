package gowa

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type Client struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
	enabled    bool
}

// NewClient creates a new Gowa HTTP client
func NewClient(config GowaConfig) *Client {
	timeout := 30 * time.Second
	if config.Timeout > 0 {
		timeout = time.Duration(config.Timeout) * time.Second
	}

	return &Client{
		baseURL:  config.BaseURL,
		apiKey:   config.APIKey,
		enabled:  config.Enabled,
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}
}

// IsEnabled returns whether Gowa is enabled
func (c *Client) IsEnabled() bool {
	return c.enabled
}

// doRequest performs an HTTP request to the Gowa API
func (c *Client) doRequest(endpoint string, payload any) (*GowaResponse, error) {
	if !c.enabled {
		return nil, fmt.Errorf("Gowa is disabled")
	}

	// Marshal payload to JSON
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload: %w", err)
	}

	// Build full URL
	url := fmt.Sprintf("%s%s", c.baseURL, endpoint)

	// Create HTTP request
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.apiKey))

	// Send request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Parse response
	var result GowaResponse
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Check HTTP status
	if resp.StatusCode != http.StatusOK {
		return &result, fmt.Errorf("Gowa API returned status %d: %s", resp.StatusCode, result.Message)
	}

	return &result, nil
}

// GetHealth checks if Gowa server is accessible
func (c *Client) GetHealth() error {
	if !c.enabled {
		return fmt.Errorf("Gowa is disabled")
	}

	url := fmt.Sprintf("%s/health", c.baseURL)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create health check request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to perform health check: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Gowa health check failed with status %d", resp.StatusCode)
	}

	return nil
}
