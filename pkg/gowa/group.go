package gowa

import (
	"context"
	"fmt"
	"net/http"
)

// GetMyGroups returns all groups that the authenticated user has joined.
// Note: WhatsApp protocol limits this to a maximum of 500 groups.
func (c *Client) GetMyGroups(ctx context.Context) ([]Group, error) {
	return c.GetMyGroupsWithDevice(ctx, "")
}

// GetMyGroupsWithDevice returns all groups for the specified device.
// If deviceID is empty, the default device ID from config is used.
// Note: WhatsApp protocol limits this to a maximum of 500 groups.
func (c *Client) GetMyGroupsWithDevice(ctx context.Context, deviceID string) ([]Group, error) {
	body, statusCode, err := c.doRequest(ctx, http.MethodGet, "/user/my/groups", nil, deviceID)
	if err != nil {
		return nil, fmt.Errorf("get my groups request failed: %w", err)
	}

	var result UserGroupResponse
	if err := parseResponse(body, statusCode, &result); err != nil {
		return nil, fmt.Errorf("get my groups failed: %w", err)
	}

	return result.Results.Data, nil
}

// GetGroupInfo returns detailed information about a specific group.
// groupJID should be in format: 120363347168689807@g.us
func (c *Client) GetGroupInfo(ctx context.Context, groupJID string) (*GroupInfoResponse, error) {
	return c.GetGroupInfoWithDevice(ctx, groupJID, "")
}

// GetGroupInfoWithDevice returns detailed information about a specific group using the specified device.
// If deviceID is empty, the default device ID from config is used.
func (c *Client) GetGroupInfoWithDevice(ctx context.Context, groupJID string, deviceID string) (*GroupInfoResponse, error) {
	params := map[string]string{
		"group_id": groupJID,
	}

	body, statusCode, err := c.doRequestWithQuery(ctx, "/group/info", params, deviceID)
	if err != nil {
		return nil, fmt.Errorf("get group info request failed: %w", err)
	}

	var result GroupInfoResponse
	if err := parseResponse(body, statusCode, &result); err != nil {
		return nil, fmt.Errorf("get group info failed: %w", err)
	}

	return &result, nil
}

// GetGroupInviteLink returns the invite link for a specific group.
// groupJID should be in format: 120363347168689807@g.us
func (c *Client) GetGroupInviteLink(ctx context.Context, groupJID string) (*GetGroupInviteLinkResponse, error) {
	return c.GetGroupInviteLinkWithDevice(ctx, groupJID, "")
}

// GetGroupInviteLinkWithDevice returns the invite link for a specific group using the specified device.
// If deviceID is empty, the default device ID from config is used.
func (c *Client) GetGroupInviteLinkWithDevice(ctx context.Context, groupJID string, deviceID string) (*GetGroupInviteLinkResponse, error) {
	params := map[string]string{
		"group_id": groupJID,
	}

	body, statusCode, err := c.doRequestWithQuery(ctx, "/group/invite-link", params, deviceID)
	if err != nil {
		return nil, fmt.Errorf("get group invite link request failed: %w", err)
	}

	var result GetGroupInviteLinkResponse
	if err := parseResponse(body, statusCode, &result); err != nil {
		return nil, fmt.Errorf("get group invite link failed: %w", err)
	}

	return &result, nil
}

// FindGroupByName searches for a group by name from the user's joined groups.
// Returns the first group that matches the name (case-insensitive partial match).
func (c *Client) FindGroupByName(ctx context.Context, name string) (*Group, error) {
	return c.FindGroupByNameWithDevice(ctx, name, "")
}

// FindGroupByNameWithDevice searches for a group by name using the specified device.
// If deviceID is empty, the default device ID from config is used.
func (c *Client) FindGroupByNameWithDevice(ctx context.Context, name string, deviceID string) (*Group, error) {
	groups, err := c.GetMyGroupsWithDevice(ctx, deviceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get groups: %w", err)
	}

	nameLower := toLower(name)
	for _, group := range groups {
		if contains(toLower(group.Name), nameLower) {
			return &group, nil
		}
	}

	return nil, fmt.Errorf("group with name '%s' not found", name)
}

// GetGroupJIDByName returns the JID of a group by its name.
// This is a convenience method for getting the group ID to use in SendGroupMessage.
func (c *Client) GetGroupJIDByName(ctx context.Context, name string) (string, error) {
	group, err := c.FindGroupByName(ctx, name)
	if err != nil {
		return "", err
	}
	return group.JID, nil
}

// toLower converts a string to lowercase.
func toLower(s string) string {
	result := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			c += 'a' - 'A'
		}
		result[i] = c
	}
	return string(result)
}

// contains checks if s contains substr.
func contains(s, substr string) bool {
	if len(substr) == 0 {
		return true
	}
	if len(s) < len(substr) {
		return false
	}
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
