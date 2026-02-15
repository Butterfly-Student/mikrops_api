package gowa

import (
	"fmt"
)

// SendToGroup sends a message to a WhatsApp group
func (c *Client) SendToGroup(groupID string, message string) (*MessageReceipt, error) {
	if groupID == "" {
		return nil, fmt.Errorf("group ID is required")
	}

	if message == "" {
		return nil, fmt.Errorf("message is required")
	}

	// Create request payload
	payload := SendGroupMessageRequest{
		GroupID: groupID,
		Message: message,
	}

	// Send request
	resp, err := c.doRequest("/api/message/send-group", payload)
	if err != nil {
		return nil, err
	}

	if !resp.Success {
		return nil, fmt.Errorf("Gowa API error: %s", resp.Message)
	}

	// Parse receipt from Data
	receipt, err := parseMessageReceipt(resp.Data)
	if err != nil {
		return nil, fmt.Errorf("failed to parse message receipt: %w", err)
	}

	return receipt, nil
}

// SendToMultipleGroups sends the same message to multiple groups
func (c *Client) SendToMultipleGroups(groupIDs []string, message string) ([]*MessageReceipt, []error) {
	var receipts []*MessageReceipt
	var errs []error

	for _, groupID := range groupIDs {
		receipt, err := c.SendToGroup(groupID, message)
		if err != nil {
			errs = append(errs, fmt.Errorf("failed to send to group %s: %w", groupID, err))
			continue
		}
		receipts = append(receipts, receipt)
	}

	return receipts, errs
}
