package whatsapp

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type WhatsAppUtil struct {
	apiURL    string
	apiKey    string
	isEnabled bool
}

type WhatsAppConfig struct {
	APIURL  string
	APIKey  string
	Enabled bool
}

type WhatsAppMessage struct {
	APIKey  string `json:"api_key"`
	Target  string `json:"target"`
	Message string `json:"message"`
}

type WhatsAppResponse struct {
	Status  bool   `json:"status"`
	Message string `json:"message"`
}

func NewWhatsAppUtil(config WhatsAppConfig) *WhatsAppUtil {
	return &WhatsAppUtil{
		apiURL:    config.APIURL,
		apiKey:    config.APIKey,
		isEnabled: config.Enabled,
	}
}

func (w *WhatsAppUtil) SendMessage(phone string, message string) error {
	if !w.isEnabled {
		return fmt.Errorf("WhatsApp sending is disabled")
	}

	// Format phone number (remove + and ensure country code)
	formattedPhone := formatPhoneNumber(phone)

	// Create message payload
	msg := WhatsAppMessage{
		APIKey:  w.apiKey,
		Target:  formattedPhone,
		Message: message,
	}

	// Marshal to JSON
	jsonData, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	// Send request
	resp, err := http.Post(w.apiURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to send WhatsApp message: %w", err)
	}
	defer resp.Body.Close()

	// Parse response
	var result WhatsAppResponse
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return fmt.Errorf("failed to parse WhatsApp response: %w", err)
	}

	if !result.Status {
		return fmt.Errorf("WhatsApp API error: %s", result.Message)
	}

	return nil
}

func (w *WhatsAppUtil) SendMultipleMessages(phoneNumbers []string, message string) error {
	if !w.isEnabled {
		return fmt.Errorf("WhatsApp sending is disabled")
	}

	for _, phone := range phoneNumbers {
		if err := w.SendMessage(phone, message); err != nil {
			return err
		}
	}

	return nil
}

func (w *WhatsAppUtil) ValidatePhoneNumber(phone string) bool {
	if phone == "" {
		return false
	}

	// Basic validation - should start with country code
	// This is a simple validation, can be improved with regex
	formatted := formatPhoneNumber(phone)
	return len(formatted) >= 10
}

func (w *WhatsAppUtil) IsEnabled() bool {
	return w.isEnabled
}

func formatPhoneNumber(phone string) string {
	// Remove + and spaces
	formatted := phone
	formatted = fmt.Sprintf("%s", formatted)
	formatted = fmt.Sprintf("%s", formatted)

	// Ensure starts with country code (e.g., 62 for Indonesia)
	if len(formatted) > 0 && formatted[0] == '0' {
		formatted = "62" + formatted[1:]
	}

	return formatted
}
