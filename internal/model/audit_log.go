package model

import (
	"time"

	"github.com/google/uuid"
)

// AuditLog records important changes to entities in the system
type AuditLog struct {
	ID         uuid.UUID              `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	AdminID    *uuid.UUID             `gorm:"type:uuid" json:"admin_id,omitempty"` // NULL = system/scheduler action
	Admin      *AdminUser             `gorm:"foreignKey:AdminID;constraint:OnDelete:SET NULL" json:"admin,omitempty"`
	Action     string                 `gorm:"size:100;not null" json:"action" validate:"required"`     // create, update, delete, suspend, activate…
	EntityType string                 `gorm:"size:50;not null" json:"entity_type" validate:"required"` // customer, subscription, invoice…
	EntityID   uuid.UUID              `gorm:"type:uuid;not null" json:"entity_id" validate:"required"`
	OldValue   map[string]interface{} `gorm:"serializer:json;type:jsonb" json:"old_value,omitempty"`
	NewValue   map[string]interface{} `gorm:"serializer:json;type:jsonb" json:"new_value,omitempty"`
	IPAddress  *string                `gorm:"size:45" json:"ip_address,omitempty"`
	UserAgent  *string                `gorm:"size:255" json:"user_agent,omitempty"`
	Notes      *string                `gorm:"type:text" json:"notes,omitempty"`
	CreatedAt  time.Time              `gorm:"not null;default:CURRENT_TIMESTAMP" json:"created_at"`
}

func (AuditLog) TableName() string { return "audit_logs" }

// AuditLogFilter for querying audit logs
type AuditLogFilter struct {
	AdminID    *uuid.UUID `json:"admin_id"`
	EntityType *string    `json:"entity_type"`
	EntityID   *uuid.UUID `json:"entity_id"`
	Action     *string    `json:"action"`
	DateFrom   *time.Time `json:"date_from"`
	DateTo     *time.Time `json:"date_to"`
}
