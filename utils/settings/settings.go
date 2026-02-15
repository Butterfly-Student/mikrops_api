package settings

import (
	"fmt"
	"strconv"
	"strings"

	"go-template/internal/model"
	outbound_port "go-template/internal/port/outbound"
)

// Helper functions to get settings from database

// GetStringSetting retrieves a string setting value
func GetStringSetting(dbPort outbound_port.DatabasePort, key string) (string, error) {
	setting, err := dbPort.SystemSetting().FindByKey(key)
	if err != nil {
		return "", fmt.Errorf("setting not found: %s", key)
	}

	if setting.Value == nil {
		return "", fmt.Errorf("setting value is nil: %s", key)
	}

	return *setting.Value, nil
}

// GetStringSettingWithDefault retrieves a string setting value with default
func GetStringSettingWithDefault(dbPort outbound_port.DatabasePort, key string, defaultValue string) string {
	value, err := GetStringSetting(dbPort, key)
	if err != nil {
		return defaultValue
	}
	return value
}

// GetBoolSetting retrieves a boolean setting value
func GetBoolSetting(dbPort outbound_port.DatabasePort, key string) (bool, error) {
	value, err := GetStringSetting(dbPort, key)
	if err != nil {
		return false, err
	}

	return strings.ToLower(value) == "true" || value == "1", nil
}

// GetBoolSettingWithDefault retrieves a boolean setting value with default
func GetBoolSettingWithDefault(dbPort outbound_port.DatabasePort, key string, defaultValue bool) bool {
	value, err := GetBoolSetting(dbPort, key)
	if err != nil {
		return defaultValue
	}
	return value
}

// GetIntSetting retrieves an integer setting value
func GetIntSetting(dbPort outbound_port.DatabasePort, key string) (int, error) {
	value, err := GetStringSetting(dbPort, key)
	if err != nil {
		return 0, err
	}

	intValue, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("invalid integer value for key %s: %w", key, err)
	}

	return intValue, nil
}

// GetIntSettingWithDefault retrieves an integer setting value with default
func GetIntSettingWithDefault(dbPort outbound_port.DatabasePort, key string, defaultValue int) int {
	value, err := GetIntSetting(dbPort, key)
	if err != nil {
		return defaultValue
	}
	return value
}

// GetSettingsByCategory retrieves all settings in a category
func GetSettingsByCategory(dbPort outbound_port.DatabasePort, category string) ([]model.SystemSetting, error) {
	return dbPort.SystemSetting().FindByCategory(category)
}
func EmailConfigFromSettings(dbPort outbound_port.DatabasePort) map[string]string {
	settings := make(map[string]string)

	keys := []string{
		"email.smtp_host",
		"email.smtp_port",
		"email.smtp_user",
		"email.smtp_password",
		"email.from_email",
		"email.from_name",
	}

	for _, key := range keys {
		if value, err := GetStringSetting(dbPort, key); err == nil {
			settings[strings.TrimPrefix(key, "email.")]= value
		}
	}

	return settings
}

// GowaConfigFromSettings builds Gowa config from settings
func GowaConfigFromSettings(dbPort outbound_port.DatabasePort) map[string]string {
	settings := make(map[string]string)

	keys := []string{
		"gowa.api_url",
		"gowa.api_key",
	}

	for _, key := range keys {
		if value, err := GetStringSetting(dbPort, key); err == nil {
			settings[strings.TrimPrefix(key, "gowa.")] = value
		}
	}

	return settings
}

// XenditConfigFromSettings builds Xendit config from settings
func XenditConfigFromSettings(dbPort outbound_port.DatabasePort) map[string]string {
	settings := make(map[string]string)

	keys := []string{
		"xendit.secret_key",
		"xendit.api_key",
		"xendit.environment",
	}

	for _, key := range keys {
		if value, err := GetStringSetting(dbPort, key); err == nil {
			settings[strings.TrimPrefix(key, "xendit.")] = value
		}
	}

	return settings
}

// WhatsAppGroupsFromSettings builds WhatsApp group IDs from settings
func WhatsAppGroupsFromSettings(dbPort outbound_port.DatabasePort) map[string]string {
	settings := make(map[string]string)

	keys := []string{
		"whatsapp.group_billing",
		"whatsapp.group_support",
		"whatsapp.group_notifications",
	}

	for _, key := range keys {
		if value, err := GetStringSetting(dbPort, key); err == nil {
			groupName := strings.TrimPrefix(key, "whatsapp.group_")
			settings[groupName] = value
		}
	}

	return settings
}
