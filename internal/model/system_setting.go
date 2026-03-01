package model

import (
	"time"

	"github.com/google/uuid"
)

// SettingType represents the data type of a system setting value
type SettingType string

const (
	SettingTypeString   SettingType = "string"
	SettingTypeInteger  SettingType = "integer"
	SettingTypeBoolean  SettingType = "boolean"
	SettingTypeJSON     SettingType = "json"
	SettingTypePassword SettingType = "password"
)

// SystemSetting represents a global key-value configuration entry
type SystemSetting struct {
	ID          uuid.UUID   `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	GroupName   string      `gorm:"size:50;not null" json:"group_name" validate:"required,max=50"`
	KeyName     string      `gorm:"size:100;not null" json:"key_name" validate:"required,max=100"`
	Value       *string     `gorm:"type:text" json:"value,omitempty"`
	Type        SettingType `gorm:"size:20;default:string" json:"type"`
	Label       *string     `gorm:"size:150" json:"label,omitempty"`
	Description *string     `gorm:"type:text" json:"description,omitempty"`
	IsEncrypted bool        `gorm:"default:false" json:"is_encrypted"` // value is AES-256 encrypted
	IsPublic    bool        `gorm:"default:false" json:"is_public"`    // readable without auth
	UpdatedAt   *time.Time  `json:"updated_at,omitempty"`
	UpdatedBy   *uuid.UUID  `gorm:"type:uuid" json:"updated_by,omitempty"`
	Updater     *AdminUser  `gorm:"foreignKey:UpdatedBy;constraint:OnDelete:SET NULL" json:"updater,omitempty"`
}

func (SystemSetting) TableName() string { return "system_settings" }

// SettingHistory records every change to a SystemSetting
type SettingHistory struct {
	ID        uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	GroupName string     `gorm:"size:50;not null" json:"group_name"`
	KeyName   string     `gorm:"size:100;not null" json:"key_name"`
	OldValue  *string    `gorm:"type:text" json:"old_value,omitempty"`
	NewValue  *string    `gorm:"type:text" json:"new_value,omitempty"`
	ChangedBy uuid.UUID  `gorm:"type:uuid;not null" json:"changed_by"`
	Changer   *AdminUser `gorm:"foreignKey:ChangedBy;constraint:OnDelete:RESTRICT" json:"changer,omitempty"`
	ChangedAt time.Time  `gorm:"default:CURRENT_TIMESTAMP;not null" json:"changed_at"`
	IPAddress *string    `gorm:"size:45" json:"ip_address,omitempty"`
	UserAgent *string    `gorm:"size:255" json:"user_agent,omitempty"`
}

func (SettingHistory) TableName() string { return "settings_history" }

// SystemSettingInput for creating/updating a setting
type SystemSettingInput struct {
	GroupName   string  `json:"group_name" validate:"required,max=50"`
	KeyName     string  `json:"key_name" validate:"required,max=100"`
	Value       *string `json:"value"`
	Type        string  `json:"type" validate:"omitempty,oneof=string integer boolean json password"`
	Label       *string `json:"label" validate:"omitempty,max=150"`
	Description *string `json:"description"`
	IsEncrypted *bool   `json:"is_encrypted"`
	IsPublic    *bool   `json:"is_public"`
}

// SystemSettingFilter for querying settings
type SystemSettingFilter struct {
	GroupName *string `json:"group_name"`
	IsPublic  *bool   `json:"is_public"`
	Search    *string `json:"search"` // group_name, key_name, label
}
