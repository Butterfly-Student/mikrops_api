package xendit

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/google/uuid"
)

const (
	ProductionAPIURL  = "https://api.xendit.co"
	DevelopmentAPIURL = "https://api.xendit.co"
)

type Client struct {
	apiKey       string
	apiURL       string
	webhookToken string
	httpClient   *http.Client
}

func NewClient(apiKey string, env Environment) *Client {
	apiURL := ProductionAPIURL
	if env == EnvironmentDevelopment {
		apiURL = DevelopmentAPIURL
	}

	return &Client{
		apiKey: apiKey,
		apiURL: apiURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (c *Client) SetWebhookToken(token string) {
	c.webhookToken = token
}

func (c *Client) makeRequest(method, endpoint string, body interface{}, result interface{}) error {
	var reqBody io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("failed to marshal request body: %w", err)
		}
		reqBody = bytes.NewBuffer(jsonData)
	}

	url := fmt.Sprintf("%s%s", c.apiURL, endpoint)
	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.SetBasicAuth(c.apiKey, "")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode >= 400 {
		var errResp ErrorResponse
		if err := json.Unmarshal(respBody, &errResp); err != nil {
			return fmt.Errorf("request failed with status %d: %s", resp.StatusCode, string(respBody))
		}
		return fmt.Errorf("xendit error [%s]: %s", errResp.ErrorCode, errResp.Message)
	}

	if err := json.Unmarshal(respBody, result); err != nil {
		return fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return nil
}

func (c *Client) GenerateID() string {
	return uuid.New().String()
}

func (c *Client) GetCurrentTimestamp() time.Time {
	return time.Now()
}
