package gowa

// SendMessageRequest represents the request to send a message
type SendMessageRequest struct {
	ChatID  string `json:"chat_id"` // Phone number for individual, or group ID
	Message string `json:"message"`
}

// SendGroupMessageRequest represents the request to send a message to a group
type SendGroupMessageRequest struct {
	GroupID string `json:"group_id"`
	Message string `json:"message"`
}

// GowaResponse represents the standard response from Gowa API
type GowaResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

// DeliveryStatus represents the delivery status of a message
type DeliveryStatus struct {
	MessageID string `json:"message_id"`
	Status    string `json:"status"` // sent, delivered, read, failed
	Timestamp string `json:"timestamp"`
}

// GowaConfig represents the configuration for Gowa client
type GowaConfig struct {
	BaseURL   string
	APIKey    string
	Timeout   int // seconds
	Enabled   bool
}

// MessageReceipt represents the receipt/acknowledgment of sent message
type MessageReceipt struct {
	MessageID string `json:"message_id"`
	ChatID    string `json:"chat_id"`
	Timestamp  string `json:"timestamp"`
}
