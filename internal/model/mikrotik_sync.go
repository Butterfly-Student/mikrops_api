package model

import (
	"github.com/google/uuid"
)

// MikrotikSyncAction represents the action to perform on MikroTik
type MikrotikSyncAction string

const (
	MikrotikSyncActionCreate MikrotikSyncAction = "create"
	MikrotikSyncActionUpdate MikrotikSyncAction = "update"
	MikrotikSyncActionDelete MikrotikSyncAction = "delete"
	MikrotikSyncActionEnable MikrotikSyncAction = "enable"
	MikrotikSyncActionDisable MikrotikSyncAction = "disable"
)

// MikrotikServiceType represents the type of service to sync
type MikrotikServiceType string

const (
	MikrotikServiceTypePPPoE   MikrotikServiceType = "pppoe"
	MikrotikServiceTypeHotspot MikrotikServiceType = "hotspot"
	MikrotikServiceTypeStaticIP MikrotikServiceType = "static_ip"
	MikrotikServiceTypeVPN     MikrotikServiceType = "vpn"
)

// MikrotikSyncMessage is the message payload for MikroTik sync operations
type MikrotikSyncMessage struct {
	SubscriptionID uuid.UUID           `json:"subscription_id" validate:"required"`
	RouterID       uuid.UUID           `json:"router_id" validate:"required"`
	Action         MikrotikSyncAction  `json:"action" validate:"required"`
	ServiceType    MikrotikServiceType `json:"service_type" validate:"required"`
	
	// Service data - populated based on ServiceType
	PPPoEData   *PPPoESyncData   `json:"pppoe_data,omitempty"`
	HotspotData *HotspotSyncData `json:"hotspot_data,omitempty"`
}

// PPPoESyncData contains PPPoE secret configuration
type PPPoESyncData struct {
	Username       string `json:"username" validate:"required"`
	Password       string `json:"password" validate:"required"`
	Profile        string `json:"profile" validate:"required"`
	LocalAddress   string `json:"local_address,omitempty"`
	RemoteAddress  string `json:"remote_address,omitempty"`
	CallerID       string `json:"caller_id,omitempty"` // MAC address binding
	RateLimit      string `json:"rate_limit,omitempty"`
	Comment        string `json:"comment,omitempty"`
}

// HotspotSyncData contains Hotspot user configuration
type HotspotSyncData struct {
	Username      string `json:"username" validate:"required"`
	Password      string `json:"password" validate:"required"`
	Profile       string `json:"profile,omitempty"`
	SharedUsers   int    `json:"shared_users,omitempty"`
	RateLimit     string `json:"rate_limit,omitempty"`
	Comment       string `json:"comment,omitempty"`
}

// MikrotikSyncResult represents the result of a sync operation
type MikrotikSyncResult struct {
	Success      bool   `json:"success"`
	ErrorMessage string `json:"error_message,omitempty"`
	ResourceID   string `json:"resource_id,omitempty"` // Internal MikroTik ID
}
