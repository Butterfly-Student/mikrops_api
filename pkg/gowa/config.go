// Package gowa provides a reusable WhatsApp gateway client for the Gowa API.
// It supports multi-device management, sending messages, and group operations.
//
// Configuration is loaded from environment variables:
//   - GOWA_BASE_URL: Base URL of the Gowa API server (default: http://localhost:3000)
//   - GOWA_USERNAME: Basic auth username
//   - GOWA_PASSWORD: Basic auth password
//   - GOWA_DEVICE_ID: Default device ID to use for requests
//   - GOWA_TIMEOUT: HTTP request timeout in seconds (default: 30)
package gowa

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// Config holds the configuration for the Gowa WhatsApp gateway client.
type Config struct {
	// BaseURL is the base URL of the Gowa API server.
	BaseURL string

	// Username is the basic auth username.
	Username string

	// Password is the basic auth password.
	Password string

	// DeviceID is the default device ID to use for requests.
	DeviceID string

	// Timeout is the HTTP request timeout.
	Timeout time.Duration
}

// ConfigFromEnv loads configuration from environment variables.
// Required environment variables:
//   - GOWA_BASE_URL
//   - GOWA_USERNAME
//   - GOWA_PASSWORD
//
// Optional environment variables:
//   - GOWA_DEVICE_ID (default: "")
//   - GOWA_TIMEOUT (default: 30 seconds)
func ConfigFromEnv() (*Config, error) {
	baseURL := os.Getenv("GOWA_BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:3000"
	}

	username := os.Getenv("GOWA_USERNAME")
	if username == "" {
		return nil, fmt.Errorf("GOWA_USERNAME environment variable is required")
	}

	password := os.Getenv("GOWA_PASSWORD")
	if password == "" {
		return nil, fmt.Errorf("GOWA_PASSWORD environment variable is required")
	}

	deviceID := os.Getenv("GOWA_DEVICE_ID")

	timeout := 30 * time.Second
	if timeoutStr := os.Getenv("GOWA_TIMEOUT"); timeoutStr != "" {
		timeoutSecs, err := strconv.Atoi(timeoutStr)
		if err != nil {
			return nil, fmt.Errorf("invalid GOWA_TIMEOUT value: %w", err)
		}
		timeout = time.Duration(timeoutSecs) * time.Second
	}

	return &Config{
		BaseURL:  baseURL,
		Username: username,
		Password: password,
		DeviceID: deviceID,
		Timeout:  timeout,
	}, nil
}
