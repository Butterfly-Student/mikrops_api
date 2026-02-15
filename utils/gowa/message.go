package gowa

import (
	"fmt"
	"regexp"
	"strings"
)

// SendTextMessage sends a text message to an individual phone number
func (c *Client) SendTextMessage(phone string, message string) (*MessageReceipt, error) {
	// Format phone number
	formattedPhone, err := formatPhoneNumber(phone)
	if err != nil {
		return nil, fmt.Errorf("invalid phone number: %w", err)
	}

	// Create request payload
	payload := SendMessageRequest{
		ChatID:  formattedPhone,
		Message: message,
	}

	// Send request
	resp, err := c.doRequest("/api/message/send-text", payload)
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

// SendMultipleMessages sends the same message to multiple phone numbers
func (c *Client) SendMultipleMessages(phones []string, message string) ([]*MessageReceipt, []error) {
	var receipts []*MessageReceipt
	var errs []error

	for _, phone := range phones {
		receipt, err := c.SendTextMessage(phone, message)
		if err != nil {
			errs = append(errs, fmt.Errorf("failed to send to %s: %w", phone, err))
			continue
		}
		receipts = append(receipts, receipt)
	}

	return receipts, errs
}

// ValidatePhoneNumber validates if a phone number is properly formatted
func ValidatePhoneNumber(phone string) bool {
	if phone == "" {
		return false
	}

	// Remove common separators
	cleaned := strings.ReplaceAll(phone, " ", "")
	cleaned = strings.ReplaceAll(cleaned, "-", "")
	cleaned = strings.ReplaceAll(cleaned, "(", "")
	cleaned = strings.ReplaceAll(cleaned, ")", "")
	cleaned = strings.ReplaceAll(cleaned, "+", "")

	// Check if it's all digits
	if !regexp.MustCompile(`^\d+$`).MatchString(cleaned) {
		return false
	}

	// Check length (should be at least 10 digits)
	return len(cleaned) >= 10
}

// formatPhoneNumber formats a phone number for WhatsApp
func formatPhoneNumber(phone string) (string, error) {
	if phone == "" {
		return "", fmt.Errorf("phone number is empty")
	}

	// Remove common separators
	cleaned := strings.ReplaceAll(phone, " ", "")
	cleaned = strings.ReplaceAll(cleaned, "-", "")
	cleaned = strings.ReplaceAll(cleaned, "(", "")
	cleaned = strings.ReplaceAll(cleaned, ")", "")

	// Remove + if present
	if strings.HasPrefix(cleaned, "+") {
		cleaned = cleaned[1:]
	}

	// If starts with 0, replace with country code (62 for Indonesia)
	if strings.HasPrefix(cleaned, "0") {
		cleaned = "62" + cleaned[1:]
	}

	// Validate
	if !regexp.MustCompile(`^\d+$`).MatchString(cleaned) {
		return "", fmt.Errorf("phone number contains invalid characters")
	}

	if len(cleaned) < 10 {
		return "", fmt.Errorf("phone number is too short")
	}

	return cleaned, nil
}

// parseMessageReceipt parses message receipt from response data
func parseMessageReceipt(data any) (*MessageReceipt, error) {
	if data == nil {
		return nil, fmt.Errorf("no data in response")
	}

	// Try to convert to map
	dataMap, ok := data.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("invalid data format")
	}

	receipt := &MessageReceipt{}

	if messageID, ok := dataMap["message_id"].(string); ok {
		receipt.MessageID = messageID
	}

	if chatID, ok := dataMap["chat_id"].(string); ok {
		receipt.ChatID = chatID
	}

	if timestamp, ok := dataMap["timestamp"].(string); ok {
		receipt.Timestamp = timestamp
	}

	return receipt, nil
}
