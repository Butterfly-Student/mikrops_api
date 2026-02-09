package model

import (
	"encoding/json"
	"time"
)

type ActivityLog struct {
	ID string `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	ActivityLogInput
}

func (ActivityLog) TableName() string {
	return "activity_logs"
}

type ActivityLogInput struct {
	TenantID     string          `json:"tenant_id" gorm:"column:tenant_id;type:uuid;not null;index"`
	UserID       *string         `json:"user_id" gorm:"column:user_id;type:uuid"`
	UserType     string          `json:"user_type" gorm:"column:user_type;type:varchar(50)"`
	Action       string          `json:"action" gorm:"column:action;type:varchar(100);not null"`
	ResourceType string          `json:"resource_type" gorm:"column:resource_type;type:varchar(100)"`
	ResourceID   string          `json:"resource_id" gorm:"column:resource_id;type:varchar(100)"`
	Details      json.RawMessage `json:"details" gorm:"column:details;type:jsonb"`
	IPAddress    string          `json:"ip_address" gorm:"column:ip_address;type:varchar(45)"`
	CreatedAt    time.Time       `json:"created_at" gorm:"column:created_at;autoCreateTime"`
}

type ActivityLogFilter struct {
	IDs           []string `json:"ids"`
	TenantIDs     []string `json:"tenant_ids"`
	UserIDs       []string `json:"user_ids"`
	UserTypes     []string `json:"user_types"`
	Actions       []string `json:"actions"`
	ResourceTypes []string `json:"resource_types"`
	ResourceIDs   []string `json:"resource_ids"`
}

func (f ActivityLogFilter) IsEmpty() bool {
	return len(f.IDs) == 0 && len(f.TenantIDs) == 0 && len(f.Actions) == 0 &&
		len(f.ResourceTypes) == 0
}
