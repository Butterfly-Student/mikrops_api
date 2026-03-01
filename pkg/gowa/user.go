package gowa

import (
	"context"
	"fmt"
	"net/http"
)

// GetUserInfo returns information about a WhatsApp user.
// phone should be in format: 6289685028129@s.whatsapp.net
func (c *Client) GetUserInfo(ctx context.Context, phone string) (*UserInfoResponse, error) {
	return c.GetUserInfoWithDevice(ctx, phone, "")
}

// GetUserInfoWithDevice returns information about a WhatsApp user using the specified device.
// If deviceID is empty, the default device ID from config is used.
func (c *Client) GetUserInfoWithDevice(ctx context.Context, phone string, deviceID string) (*UserInfoResponse, error) {
	params := map[string]string{
		"phone": phone,
	}

	body, statusCode, err := c.doRequestWithQuery(ctx, "/user/info", params, deviceID)
	if err != nil {
		return nil, fmt.Errorf("get user info request failed: %w", err)
	}

	var result UserInfoResponse
	if err := parseResponse(body, statusCode, &result); err != nil {
		return nil, fmt.Errorf("get user info failed: %w", err)
	}

	return &result, nil
}

// CheckUser checks if a phone number is registered on WhatsApp.
// phone should be in format: 628912344551 (without @s.whatsapp.net)
func (c *Client) CheckUser(ctx context.Context, phone string) (bool, error) {
	return c.CheckUserWithDevice(ctx, phone, "")
}

// CheckUserWithDevice checks if a phone number is registered on WhatsApp using the specified device.
// If deviceID is empty, the default device ID from config is used.
func (c *Client) CheckUserWithDevice(ctx context.Context, phone string, deviceID string) (bool, error) {
	params := map[string]string{
		"phone": phone,
	}

	body, statusCode, err := c.doRequestWithQuery(ctx, "/user/check", params, deviceID)
	if err != nil {
		return false, fmt.Errorf("check user request failed: %w", err)
	}

	var result UserCheckResponse
	if err := parseResponse(body, statusCode, &result); err != nil {
		return false, fmt.Errorf("check user failed: %w", err)
	}

	return result.Results.IsOnWhatsApp, nil
}

// GetMyContacts returns the list of contacts for the default device.
func (c *Client) GetMyContacts(ctx context.Context) ([]map[string]interface{}, error) {
	return c.GetMyContactsWithDevice(ctx, "")
}

// GetMyContactsWithDevice returns the list of contacts for the specified device.
// If deviceID is empty, the default device ID from config is used.
func (c *Client) GetMyContactsWithDevice(ctx context.Context, deviceID string) ([]map[string]interface{}, error) {
	body, statusCode, err := c.doRequest(ctx, http.MethodGet, "/user/my/contacts", nil, deviceID)
	if err != nil {
		return nil, fmt.Errorf("get my contacts request failed: %w", err)
	}

	var result struct {
		Code    string `json:"code"`
		Message string `json:"message"`
		Results struct {
			Data []map[string]interface{} `json:"data"`
		} `json:"results"`
	}
	if err := parseResponse(body, statusCode, &result); err != nil {
		return nil, fmt.Errorf("get my contacts failed: %w", err)
	}

	return result.Results.Data, nil
}

// FormatPhoneNumber formats a phone number to WhatsApp JID format.
// Input: 6289685028129
// Output: 6289685028129@s.whatsapp.net
func FormatPhoneNumber(phone string) string {
	return phone + "@s.whatsapp.net"
}

// FormatGroupJID formats a group ID to WhatsApp group JID format.
// Input: 120363347168689807
// Output: 120363347168689807@g.us
func FormatGroupJID(groupID string) string {
	return groupID + "@g.us"
}
