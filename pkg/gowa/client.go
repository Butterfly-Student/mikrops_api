package gowa

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

// Client is the Gowa WhatsApp gateway HTTP client.
type Client struct {
	config     *Config
	httpClient *http.Client
}

// New creates a new Gowa client with the given configuration.
func New(cfg *Config) *Client {
	return &Client{
		config: cfg,
		httpClient: &http.Client{
			Timeout: cfg.Timeout,
		},
	}
}

// NewFromEnv creates a new Gowa client using configuration from environment variables.
func NewFromEnv() (*Client, error) {
	cfg, err := ConfigFromEnv()
	if err != nil {
		return nil, fmt.Errorf("failed to load gowa config from env: %w", err)
	}
	return New(cfg), nil
}

// WithDeviceID returns a new client with the specified device ID.
// This is useful when you need to use a different device than the default.
func (c *Client) WithDeviceID(deviceID string) *Client {
	newCfg := *c.config
	newCfg.DeviceID = deviceID
	return &Client{
		config:     &newCfg,
		httpClient: c.httpClient,
	}
}

// doRequest performs an HTTP request to the Gowa API.
func (c *Client) doRequest(ctx context.Context, method, path string, body interface{}, deviceID string) ([]byte, int, error) {
	fullURL := c.config.BaseURL + path

	var reqBody io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to marshal request body: %w", err)
		}
		reqBody = bytes.NewBuffer(jsonData)
	}

	req, err := http.NewRequestWithContext(ctx, method, fullURL, reqBody)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to create request: %w", err)
	}

	// Set basic auth
	req.SetBasicAuth(c.config.Username, c.config.Password)

	// Set content type for requests with body
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	// Set device ID header if provided
	if deviceID != "" {
		req.Header.Set("X-Device-Id", deviceID)
	} else if c.config.DeviceID != "" {
		req.Header.Set("X-Device-Id", c.config.DeviceID)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, fmt.Errorf("failed to read response body: %w", err)
	}

	return respBody, resp.StatusCode, nil
}

// doRequestWithQuery performs an HTTP GET request with query parameters.
func (c *Client) doRequestWithQuery(ctx context.Context, path string, params map[string]string, deviceID string) ([]byte, int, error) {
	u, err := url.Parse(c.config.BaseURL + path)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to parse URL: %w", err)
	}

	if len(params) > 0 {
		q := u.Query()
		for k, v := range params {
			q.Set(k, v)
		}
		u.RawQuery = q.Encode()
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to create request: %w", err)
	}

	req.SetBasicAuth(c.config.Username, c.config.Password)

	if deviceID != "" {
		req.Header.Set("X-Device-Id", deviceID)
	} else if c.config.DeviceID != "" {
		req.Header.Set("X-Device-Id", c.config.DeviceID)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, fmt.Errorf("failed to read response body: %w", err)
	}

	return respBody, resp.StatusCode, nil
}

// parseResponse parses the JSON response body into the given target.
func parseResponse(body []byte, statusCode int, target interface{}) error {
	if statusCode >= 400 {
		var errResp ErrorResponse
		if err := json.Unmarshal(body, &errResp); err != nil {
			return fmt.Errorf("request failed with status %d: %s", statusCode, string(body))
		}
		return fmt.Errorf("request failed with status %d: %s - %s", statusCode, errResp.Code, errResp.Message)
	}

	if err := json.Unmarshal(body, target); err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}

	return nil
}
