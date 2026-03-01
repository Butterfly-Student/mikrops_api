package gowa

import (
	"context"
	"fmt"
	"net/http"
)

// Login initiates a QR code login for the default device.
// Returns a LoginResponse containing the QR code link and duration.
func (c *Client) Login(ctx context.Context) (*LoginResponse, error) {
	return c.LoginWithDevice(ctx, "")
}

// LoginWithDevice initiates a QR code login for the specified device.
// If deviceID is empty, the default device ID from config is used.
func (c *Client) LoginWithDevice(ctx context.Context, deviceID string) (*LoginResponse, error) {
	body, statusCode, err := c.doRequest(ctx, http.MethodGet, "/app/login", nil, deviceID)
	if err != nil {
		return nil, fmt.Errorf("login request failed: %w", err)
	}

	var result LoginResponse
	if err := parseResponse(body, statusCode, &result); err != nil {
		return nil, fmt.Errorf("login failed: %w", err)
	}

	return &result, nil
}

// LoginWithCode initiates a pairing code login for the default device.
// phone should be in format: 628912344551
func (c *Client) LoginWithCode(ctx context.Context, phone string) (*LoginWithCodeResponse, error) {
	return c.LoginWithCodeAndDevice(ctx, phone, "")
}

// LoginWithCodeAndDevice initiates a pairing code login for the specified device.
// phone should be in format: 628912344551
// If deviceID is empty, the default device ID from config is used.
func (c *Client) LoginWithCodeAndDevice(ctx context.Context, phone string, deviceID string) (*LoginWithCodeResponse, error) {
	params := map[string]string{
		"phone": phone,
	}

	body, statusCode, err := c.doRequestWithQuery(ctx, "/app/login-with-code", params, deviceID)
	if err != nil {
		return nil, fmt.Errorf("login with code request failed: %w", err)
	}

	var result LoginWithCodeResponse
	if err := parseResponse(body, statusCode, &result); err != nil {
		return nil, fmt.Errorf("login with code failed: %w", err)
	}

	return &result, nil
}

// Logout removes the database and logs out from WhatsApp for the default device.
func (c *Client) Logout(ctx context.Context) (*GenericResponse, error) {
	return c.LogoutWithDevice(ctx, "")
}

// LogoutWithDevice removes the database and logs out from WhatsApp for the specified device.
// If deviceID is empty, the default device ID from config is used.
func (c *Client) LogoutWithDevice(ctx context.Context, deviceID string) (*GenericResponse, error) {
	body, statusCode, err := c.doRequest(ctx, http.MethodGet, "/app/logout", nil, deviceID)
	if err != nil {
		return nil, fmt.Errorf("logout request failed: %w", err)
	}

	var result GenericResponse
	if err := parseResponse(body, statusCode, &result); err != nil {
		return nil, fmt.Errorf("logout failed: %w", err)
	}

	return &result, nil
}

// Reconnect reconnects to the WhatsApp server for the default device.
func (c *Client) Reconnect(ctx context.Context) (*GenericResponse, error) {
	return c.ReconnectWithDevice(ctx, "")
}

// ReconnectWithDevice reconnects to the WhatsApp server for the specified device.
// If deviceID is empty, the default device ID from config is used.
func (c *Client) ReconnectWithDevice(ctx context.Context, deviceID string) (*GenericResponse, error) {
	body, statusCode, err := c.doRequest(ctx, http.MethodGet, "/app/reconnect", nil, deviceID)
	if err != nil {
		return nil, fmt.Errorf("reconnect request failed: %w", err)
	}

	var result GenericResponse
	if err := parseResponse(body, statusCode, &result); err != nil {
		return nil, fmt.Errorf("reconnect failed: %w", err)
	}

	return &result, nil
}

// GetStatus returns the connection status for the default device.
func (c *Client) GetStatus(ctx context.Context) (*AppStatusResponse, error) {
	return c.GetStatusWithDevice(ctx, "")
}

// GetStatusWithDevice returns the connection status for the specified device.
// If deviceID is empty, the default device ID from config is used.
func (c *Client) GetStatusWithDevice(ctx context.Context, deviceID string) (*AppStatusResponse, error) {
	body, statusCode, err := c.doRequest(ctx, http.MethodGet, "/app/status", nil, deviceID)
	if err != nil {
		return nil, fmt.Errorf("get status request failed: %w", err)
	}

	var result AppStatusResponse
	if err := parseResponse(body, statusCode, &result); err != nil {
		return nil, fmt.Errorf("get status failed: %w", err)
	}

	return &result, nil
}

// GetConnectedDevices returns all connected devices.
func (c *Client) GetConnectedDevices(ctx context.Context) ([]map[string]interface{}, error) {
	body, statusCode, err := c.doRequest(ctx, http.MethodGet, "/app/devices", nil, "")
	if err != nil {
		return nil, fmt.Errorf("get connected devices request failed: %w", err)
	}

	var result struct {
		Code    string                   `json:"code"`
		Message string                   `json:"message"`
		Results []map[string]interface{} `json:"results"`
	}
	if err := parseResponse(body, statusCode, &result); err != nil {
		return nil, fmt.Errorf("get connected devices failed: %w", err)
	}

	return result.Results, nil
}
