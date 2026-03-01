package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// MikrotikRouterStatus represents the connectivity status of a router
type MikrotikRouterStatus string

const (
	MikrotikRouterStatusOnline  MikrotikRouterStatus = "online"
	MikrotikRouterStatusOffline MikrotikRouterStatus = "offline"
	MikrotikRouterStatusUnknown MikrotikRouterStatus = "unknown"
)

// MikrotikRouter represents a MikroTik device managed by the system
type MikrotikRouter struct {
	ID                uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Name              string    `gorm:"not null" json:"name" validate:"required"`
	Address           string    `gorm:"not null" json:"address" validate:"required"` // IP:Port
	ApiPort           *int      `gorm:"default:8728" json:"api_port"`
	RestPort          *int      `gorm:"default:80" json:"rest_port"`
	Username          string    `gorm:"not null" json:"username" validate:"required"`
	Password          string    `gorm:"not null" json:"password" validate:"required"`
	PasswordEncrypted *string   `gorm:"type:text" json:"password_encrypted,omitempty"`
	UseSSL            *bool     `gorm:"default:false" json:"use_ssl"`
	RouterOSVersion   *string   `gorm:"size:50" json:"router_os_version,omitempty"`
	Identity          *string   `gorm:"size:255" json:"identity,omitempty"`
	IsActive          *bool     `gorm:"default:true" json:"is_active"`

	// Area / location grouping
	Area *string `gorm:"size:100" json:"area,omitempty"`

	// Multi-router: marks primary router for a location
	IsMaster *bool `gorm:"default:false" json:"is_master"`

	// Connectivity status (updated by poller)
	Status   MikrotikRouterStatus `gorm:"size:20;not null;default:unknown" json:"status"`
	LastPing *time.Time           `json:"last_ping,omitempty"`

	Notes      *string    `gorm:"type:text" json:"notes,omitempty"`
	LastSeenAt *time.Time `json:"last_seen_at,omitempty"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName specifies the table name
func (MikrotikRouter) TableName() string { return "mikrotik_routers" }

// MikrotikRouterInput for creating/updating a router
type MikrotikRouterInput struct {
	Name     string  `json:"name" validate:"required"`
	Address  string  `json:"address" validate:"required"`
	ApiPort  *int    `json:"api_port"`
	Username string  `json:"username" validate:"required"`
	Password string  `json:"password" validate:"required"`
	UseSSL   *bool   `json:"use_ssl"`
	IsActive *bool   `json:"is_active"`
	Area     *string `json:"area" validate:"omitempty,max=100"`
	IsMaster *bool   `json:"is_master"`
}

// MikrotikTestResult holds the result of a connection test
type MikrotikTestResult struct {
	Success         bool   `json:"success"`
	Message         string `json:"message"`
	RouterOSVersion string `json:"router_os_version,omitempty"`
	Identity        string `json:"identity,omitempty"`
}

// PPPoE Models for RouterOS API communication

type PppoeSecret struct {
	ID            string `json:"id,omitempty"`
	Name          string `json:"name" validate:"required"`
	Password      string `json:"password" validate:"required"`
	Service       string `json:"service"`
	CallerID      string `json:"caller_id"`
	Profile       string `json:"profile"`
	LocalAddress  string `json:"local_address"`
	RemoteAddress string `json:"remote_address"`
	LimitBytesIn  int64  `json:"limit_bytes_in"`
	LimitBytesOut int64  `json:"limit_bytes_out"`
	Routes        string `json:"routes"`
	Comment       string `json:"comment"`
	Disabled      bool   `json:"disabled"`
}

type PppoeProfile struct {
	ID                string `json:"id,omitempty"`
	Name              string `json:"name" validate:"required"`
	LocalAddress      string `json:"local_address"`
	RemoteAddress     string `json:"remote_address"`
	RateLimit         string `json:"rate_limit"`
	ParentQueue       string `json:"parent_queue"`
	QueueType         string `json:"queue_type"`
	InsertQueueBefore string `json:"insert_queue_before"`
	OnlyOne           string `json:"only_one"`
	DNSServer         string `json:"dns_server"`
	AddressList       string `json:"address_list"`
	InterfaceList     string `json:"interface_list"`
	IncomingFilter    string `json:"incoming_filter"`
	OutgoingFilter    string `json:"outgoing_filter"`
	OnUp              string `json:"on_up"`
	OnDown            string `json:"on_down"`
	Comment           string `json:"comment"`
}

type PppoeActive struct {
	ID            string `json:"id,omitempty"`
	Name          string `json:"name"`
	Service       string `json:"service"`
	CallerID      string `json:"caller_id"`
	Address       string `json:"address"`
	Uptime        string `json:"uptime"`
	Encoding      string `json:"encoding"`
	SessionID     string `json:"session_id"`
	LimitBytesIn  int64  `json:"limit_bytes_in"`
	LimitBytesOut int64  `json:"limit_bytes_out"`
	Radius        bool   `json:"radius"`
}

// IsolationConfig holds settings for customer isolation on a MikroTik router
type IsolationConfig struct {
	ProfileName string `json:"profile_name"`
	AddressList string `json:"address_list"`
	PortalIP    string `json:"portal_ip"`
	PortalPort  string `json:"portal_port"`
	DNSServer   string `json:"dns_server"`
	RateLimit   string `json:"rate_limit"`
}

func DefaultIsolationConfig() IsolationConfig {
	return IsolationConfig{
		ProfileName: "isolir",
		AddressList: "isolated-users",
		PortalPort:  "80",
		RateLimit:   "256k/256k",
	}
}

type PppoeCallbackData struct {
	User       string `json:"user"`
	IP         string `json:"ip-address"`
	CallerID   string `json:"caller-id"`
	SessionID  string `json:"session-id"`
	Interface  string `json:"interface"`
	Uptime     string `json:"uptime"`
	BytesIn    int64  `json:"bytes-in"`
	BytesOut   int64  `json:"bytes-out"`
	PacketsIn  int64  `json:"packets-in"`
	PacketsOut int64  `json:"packets-out"`
	RouterID   uint   `json:"router_id"`
}

type WebSocketMessage struct {
	Event string      `json:"event"`
	Data  interface{} `json:"data"`
}
