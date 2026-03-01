package model

import (
	"time"

	"github.com/google/uuid"
)

// SyncOperation represents the RouterOS operation to perform
type SyncOperation string

const (
	SyncOperationAdd     SyncOperation = "add"
	SyncOperationUpdate  SyncOperation = "update"
	SyncOperationDelete  SyncOperation = "delete"
	SyncOperationEnable  SyncOperation = "enable"
	SyncOperationDisable SyncOperation = "disable"
)

// SyncResourceType represents the RouterOS resource type
type SyncResourceType string

const (
	SyncResourcePPPSecret   SyncResourceType = "ppp_secret"
	SyncResourceHotspotUser SyncResourceType = "hotspot_user"
	SyncResourceIPBinding   SyncResourceType = "ip_binding"
	SyncResourceQueue       SyncResourceType = "queue"
	SyncResourceAddress     SyncResourceType = "address"
)

// SyncQueueStatus represents the processing status of a sync queue item
type SyncQueueStatus string

const (
	SyncQueueStatusPending    SyncQueueStatus = "pending"
	SyncQueueStatusProcessing SyncQueueStatus = "processing"
	SyncQueueStatusDone       SyncQueueStatus = "done"
	SyncQueueStatusFailed     SyncQueueStatus = "failed"
)

// MikrotikSyncQueue is an async operation queue for MikroTik synchronization.
// A background worker picks up pending items and executes them against the router.
type MikrotikSyncQueue struct {
	ID             uuid.UUID              `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	RouterID       uuid.UUID              `gorm:"type:uuid;not null" json:"router_id" validate:"required"`
	Router         *MikrotikRouter        `gorm:"foreignKey:RouterID;constraint:OnDelete:RESTRICT" json:"router,omitempty"`
	SubscriptionID *uuid.UUID             `gorm:"type:uuid" json:"subscription_id,omitempty"`
	Subscription   *Subscription          `gorm:"foreignKey:SubscriptionID;constraint:OnDelete:SET NULL" json:"subscription,omitempty"`
	Operation      SyncOperation          `gorm:"size:20;not null" json:"operation" validate:"required"`
	ResourceType   SyncResourceType       `gorm:"size:20;not null" json:"resource_type" validate:"required"`
	Payload        map[string]interface{} `gorm:"serializer:json;type:jsonb;not null" json:"payload"`
	Priority       int8                   `gorm:"default:5" json:"priority"` // 1=urgent, 10=low
	Status         SyncQueueStatus        `gorm:"size:20;not null;default:pending" json:"status"`
	Attempts       int16                  `gorm:"default:0" json:"attempts"`
	MaxAttempts    int16                  `gorm:"default:3" json:"max_attempts"`
	ErrorMessage   *string                `gorm:"type:text" json:"error_message,omitempty"`
	ScheduledAt    time.Time              `gorm:"not null;default:CURRENT_TIMESTAMP" json:"scheduled_at"`
	ProcessedAt    *time.Time             `json:"processed_at,omitempty"`
	CreatedAt      time.Time              `json:"created_at"`
}

func (MikrotikSyncQueue) TableName() string { return "mikrotik_sync_queue" }

// MikrotikSyncQueueInput for enqueuing a new sync operation
type MikrotikSyncQueueInput struct {
	RouterID       uuid.UUID              `json:"router_id" validate:"required"`
	SubscriptionID *uuid.UUID             `json:"subscription_id"`
	Operation      string                 `json:"operation" validate:"required,oneof=add update delete enable disable"`
	ResourceType   string                 `json:"resource_type" validate:"required,oneof=ppp_secret hotspot_user ip_binding queue address"`
	Payload        map[string]interface{} `json:"payload" validate:"required"`
	Priority       *int8                  `json:"priority" validate:"omitempty,min=1,max=10"`
	ScheduledAt    *time.Time             `json:"scheduled_at"`
}

// MikrotikSyncQueueFilter for querying the queue
type MikrotikSyncQueueFilter struct {
	RouterID       *uuid.UUID        `json:"router_id"`
	SubscriptionID *uuid.UUID        `json:"subscription_id"`
	Status         *SyncQueueStatus  `json:"status"`
	Operation      *SyncOperation    `json:"operation"`
	ResourceType   *SyncResourceType `json:"resource_type"`
}
