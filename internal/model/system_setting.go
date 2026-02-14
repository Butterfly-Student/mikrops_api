package model

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

type SystemSetting struct {
	ID          uuid.UUID  `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Key         string     `json:"key" gorm:"unique;not null;size:100" validate:"required"`
	Value       *string    `json:"value" gorm:"type:text"`
	ValueType   string     `json:"value_type" gorm:"default:'string';not null" validate:"required,oneof=string number boolean json"`
	Category    *string    `json:"category" gorm:"size:50"`
	Description *string    `json:"description" gorm:"type:text"`
	IsPublic    *bool      `json:"is_public" gorm:"default:false"`
	UpdatedBy   *uuid.UUID `json:"updated_by" gorm:"type:uuid"`
	UpdatedAt   time.Time  `json:"updated_at" gorm:"autoUpdateTime"`
}

type SystemSettingInput struct {
	Key         *string `json:"key" validate:"required"`
	Value       *string `json:"value"`
	ValueType   *string `json:"value_type" validate:"required,oneof=string number boolean json"`
	Category    *string `json:"category"`
	Description *string `json:"description"`
	IsPublic    *bool   `json:"is_public"`
}

type SystemSettingFilter struct {
	Keys     []string `json:"keys"`
	Category *string  `json:"category"`
	IsPublic *bool    `json:"is_public"`
}

func (f SystemSettingFilter) IsEmpty() bool {
	return len(f.Keys) == 0 && f.Category == nil && f.IsPublic == nil
}

func SystemSettingPrepare(v *SystemSetting) {
	if v.ValueType == "" {
		v.ValueType = "string"
	}
	if v.IsPublic == nil {
		isPublic := false
		v.IsPublic = &isPublic
	}
}

func (s *SystemSetting) GetString(defaultValue string) string {
	if s.Value == nil {
		return defaultValue
	}
	return *s.Value
}

func (s *SystemSetting) GetInt(defaultValue int) int {
	if s.Value == nil {
		return defaultValue
	}
	var result int
	_, err := fmt.Sscanf(*s.Value, "%d", &result)
	if err != nil {
		return defaultValue
	}
	return result
}

func (s *SystemSetting) GetBool(defaultValue bool) bool {
	if s.Value == nil {
		return defaultValue
	}
	return *s.Value == "true"
}

func (s *SystemSetting) GetFloat(defaultValue float64) float64 {
	if s.Value == nil {
		return defaultValue
	}
	var result float64
	_, err := fmt.Sscanf(*s.Value, "%f", &result)
	if err != nil {
		return defaultValue
	}
	return result
}
