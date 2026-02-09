package model

import (
	"encoding/json"
	"time"
)

const (
	SyncActionCreate  = "create"
	SyncActionUpdate  = "update"
	SyncActionDelete  = "delete"
	SyncActionIsolate = "isolate"
	SyncActionRestore = "restore"

	SyncStatusSuccess = "success"
	SyncStatusFailed  = "failed"
)

type MikrotikSyncLog struct {
	ID string `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	MikrotikSyncLogInput
}

func (MikrotikSyncLog) TableName() string {
	return "mikrotik_sync_logs"
}

type MikrotikSyncLogInput struct {
	TenantID           string          `json:"tenant_id" gorm:"column:tenant_id;type:uuid;not null;index"`
	NasID              string          `json:"nas_id" gorm:"column:nas_id;type:uuid;not null"`
	Action             string          `json:"action" gorm:"column:action;type:varchar(50);not null"`
	ResourceType       string          `json:"resource_type" gorm:"column:resource_type;type:varchar(50);not null"`
	ResourceIdentifier string          `json:"resource_identifier" gorm:"column:resource_identifier;type:varchar(255)"`
	Status             string          `json:"status" gorm:"column:status;type:varchar(50);not null"`
	ErrorMessage       string          `json:"error_message" gorm:"column:error_message;type:text"`
	RequestPayload     json.RawMessage `json:"request_payload" gorm:"column:request_payload;type:jsonb"`
	CreatedAt          time.Time       `json:"created_at" gorm:"column:created_at;autoCreateTime"`
}

type MikrotikSyncLogFilter struct {
	IDs           []string `json:"ids"`
	TenantIDs     []string `json:"tenant_ids"`
	NasIDs        []string `json:"nas_ids"`
	Actions       []string `json:"actions"`
	ResourceTypes []string `json:"resource_types"`
	Statuses      []string `json:"statuses"`
}

func (f MikrotikSyncLogFilter) IsEmpty() bool {
	return len(f.IDs) == 0 && len(f.TenantIDs) == 0 && len(f.NasIDs) == 0 &&
		len(f.Actions) == 0
}
