package model

import (
	"encoding/json"
	"strconv"
	"time"

	"github.com/google/uuid"
)

// SettingValueType represents the type of setting value
type SettingValueType string

const (
	SettingValueTypeString  SettingValueType = "string"
	SettingValueTypeNumber  SettingValueType = "number"
	SettingValueTypeBoolean SettingValueType = "boolean"
	SettingValueTypeJSON    SettingValueType = "json"
)

// Setting categories
const (
	SettingCategoryBilling      = "billing"
	SettingCategoryCompany      = "company"
	SettingCategoryWhatsApp     = "whatsapp"
	SettingCategoryXendit       = "xendit"
	SettingCategoryPayment      = "payment"
	SettingCategorySystem       = "system"
	SettingCategoryNotification = "notification"
)

type SystemSetting struct {
	ID          uuid.UUID        `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Key         string           `json:"key" gorm:"uniqueIndex;not null"`
	Value       string           `json:"value"`
	ValueType   SettingValueType `json:"value_type" gorm:"type:setting_value_type;not null;default:'string'"`
	Category    string           `json:"category" gorm:"not null;index"`
	Description string           `json:"description"`
	IsPublic    bool             `json:"is_public" gorm:"default:false"`
	UpdatedBy   *uint            `json:"updated_by" gorm:"index"`
	UpdatedByUser *User          `json:"updated_by_user,omitempty" gorm:"foreignKey:UpdatedBy;references:ID"`
	CreatedAt   time.Time        `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time        `json:"updated_at" gorm:"autoUpdateTime"`
}

type SystemSettingInput struct {
	Key         string           `json:"key" binding:"required"`
	Value       string           `json:"value"`
	ValueType   SettingValueType `json:"value_type" binding:"required,oneof=string number boolean json"`
	Category    string           `json:"category" binding:"required"`
	Description string           `json:"description"`
	IsPublic    bool             `json:"is_public"`
}

type SystemSettingFilter struct {
	Keys       []string `json:"keys"`
	Categories []string `json:"categories"`
	IsPublic   *bool    `json:"is_public"`
	Search     string   `json:"search"`
}

func (f SystemSettingFilter) IsEmpty() bool {
	return len(f.Keys) == 0 && len(f.Categories) == 0 && f.IsPublic == nil && f.Search == ""
}

// TableName specifies the table name for GORM
func (SystemSetting) TableName() string {
	return "system_settings"
}

// GetString returns the value as string
func (s *SystemSetting) GetString() string {
	return s.Value
}

// GetInt returns the value as int
func (s *SystemSetting) GetInt() (int, error) {
	return strconv.Atoi(s.Value)
}

// GetInt64 returns the value as int64
func (s *SystemSetting) GetInt64() (int64, error) {
	return strconv.ParseInt(s.Value, 10, 64)
}

// GetFloat returns the value as float64
func (s *SystemSetting) GetFloat() (float64, error) {
	return strconv.ParseFloat(s.Value, 64)
}

// GetBool returns the value as bool
func (s *SystemSetting) GetBool() (bool, error) {
	return strconv.ParseBool(s.Value)
}

// GetJSON unmarshals the value as JSON
func (s *SystemSetting) GetJSON(target interface{}) error {
	return json.Unmarshal([]byte(s.Value), target)
}

// GetIntDefault returns the value as int with default fallback
func (s *SystemSetting) GetIntDefault(defaultValue int) int {
	val, err := s.GetInt()
	if err != nil {
		return defaultValue
	}
	return val
}

// GetFloat64Default returns the value as float64 with default fallback
func (s *SystemSetting) GetFloat64Default(defaultValue float64) float64 {
	val, err := s.GetFloat()
	if err != nil {
		return defaultValue
	}
	return val
}

// GetBoolDefault returns the value as bool with default fallback
func (s *SystemSetting) GetBoolDefault(defaultValue bool) bool {
	val, err := s.GetBool()
	if err != nil {
		return defaultValue
	}
	return val
}
