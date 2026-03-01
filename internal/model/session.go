package model

import (
	"time"

	"github.com/google/uuid"
)

// ActiveSession represents a currently connected customer session (from MikroTik polling)
type ActiveSession struct {
	ID             uuid.UUID       `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	SubscriptionID uuid.UUID       `gorm:"type:uuid;not null" json:"subscription_id"`
	Subscription   *Subscription   `gorm:"foreignKey:SubscriptionID;constraint:OnDelete:RESTRICT" json:"subscription,omitempty"`
	RouterID       uuid.UUID       `gorm:"type:uuid;not null" json:"router_id"`
	Router         *MikrotikRouter `gorm:"foreignKey:RouterID;constraint:OnDelete:RESTRICT" json:"router,omitempty"`
	SessionID      string          `gorm:"size:100;not null" json:"session_id"` // .id from RouterOS
	CallerID       *string         `gorm:"size:50" json:"caller_id,omitempty"`  // MAC / caller-id
	IPAddress      string          `gorm:"size:15;not null" json:"ip_address"`
	BytesIn        int64           `gorm:"default:0" json:"bytes_in"`
	BytesOut       int64           `gorm:"default:0" json:"bytes_out"`
	UptimeSeconds  int             `gorm:"default:0" json:"uptime_seconds"`
	ConnectedAt    time.Time       `gorm:"not null" json:"connected_at"`
	LastUpdated    time.Time       `gorm:"not null" json:"last_updated"`
}

func (ActiveSession) TableName() string { return "active_sessions" }

// SessionHistory records the completed session after a customer disconnects
type SessionHistory struct {
	ID              uuid.UUID       `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	SubscriptionID  uuid.UUID       `gorm:"type:uuid;not null" json:"subscription_id"`
	Subscription    *Subscription   `gorm:"foreignKey:SubscriptionID;constraint:OnDelete:RESTRICT" json:"subscription,omitempty"`
	RouterID        uuid.UUID       `gorm:"type:uuid;not null" json:"router_id"`
	Router          *MikrotikRouter `gorm:"foreignKey:RouterID;constraint:OnDelete:RESTRICT" json:"router,omitempty"`
	SessionID       *string         `gorm:"size:100" json:"session_id,omitempty"`
	IPAddress       *string         `gorm:"size:15" json:"ip_address,omitempty"`
	BytesIn         int64           `gorm:"default:0" json:"bytes_in"`
	BytesOut        int64           `gorm:"default:0" json:"bytes_out"`
	UptimeSeconds   int             `gorm:"default:0" json:"uptime_seconds"`
	ConnectedAt     time.Time       `gorm:"not null" json:"connected_at"`
	DisconnectedAt  time.Time       `gorm:"not null" json:"disconnected_at"`
	DisconnectCause *string         `gorm:"size:100" json:"disconnect_cause,omitempty"`
}

func (SessionHistory) TableName() string { return "session_history" }

// SessionFilter for querying sessions
type SessionFilter struct {
	SubscriptionID *uuid.UUID `json:"subscription_id"`
	RouterID       *uuid.UUID `json:"router_id"`
	DateFrom       *time.Time `json:"date_from"`
	DateTo         *time.Time `json:"date_to"`
}
