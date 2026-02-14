package xendit

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Client represents Xendit API client
type Client struct {
	APIKey     string
	SecretKey  string
	BaseURL    string
	HTTPClient *http.Client
}

// NewClient creates a new Xendit client
func NewClient(apiKey, secretKey string) *Client {
	return &Client{
		APIKey:    apiKey,
		SecretKey: secretKey,
		BaseURL:   "https://api.xendit.co",
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// NewTestClient creates a test client for development
func NewTestClient(apiKey, secretKey string) *Client {
	client := NewClient(apiKey, secretKey)
	client.BaseURL = "https://api.xendit.co" // Xendit doesn't have separate test URL, use API key
	return client
}

// doRequest executes HTTP request
func (c *Client) doRequest(method, path string, body interface{}) ([]byte, error) {
	var reqBody io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		reqBody = bytes.NewBuffer(jsonBody)
	}

	req, err := http.NewRequest(method, c.BaseURL+path, reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.SetBasicAuth(c.APIKey, "")
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("xendit API error (status %d): %s", resp.StatusCode, string(respBody))
	}

	return respBody, nil
}

// VerifyWebhookSignature verifies the webhook callback signature
func (c *Client) VerifyWebhookSignature(webhookToken, signature, payload string) bool {
	mac := hmac.New(sha256.New, []byte(c.SecretKey))
	mac.Write([]byte(webhookToken + payload))
	expectedSignature := hex.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(signature), []byte(expectedSignature))
}
