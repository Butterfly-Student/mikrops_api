package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type MikrotikRouter struct {
	ID                uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Name              string         `gorm:"not null" json:"name" validate:"required"`
	Address           string         `gorm:"not null" json:"address" validate:"required"` // IP:Port
	ApiPort           *int           `gorm:"default:8728" json:"api_port"`
	RestPort          *int           `gorm:"default:80" json:"rest_port"`
	Username          string         `gorm:"not null" json:"username" validate:"required"`
	Password          string         `gorm:"not null" json:"password" validate:"required"`
	PasswordEncrypted *string        `gorm:"type:text" json:"password_encrypted"`
	UseSSL            *bool          `gorm:"default:false" json:"use_ssl"`
	RouterOSVersion   *string        `gorm:"size:50" json:"router_os_version"`
	Identity          *string        `gorm:"size:255" json:"identity"`
	IsActive          *bool          `gorm:"default:true" json:"is_active"`
	LastSeenAt        *time.Time     `json:"last_seen_at"`
	CreatedAt         time.Time      `json:"created_at"`
	UpdatedAt         time.Time      `json:"updated_at"`
	DeletedAt         gorm.DeletedAt `gorm:"index" json:"-"`
}

// MikrotikRouterInput for creating/updating a MikroTik router
type MikrotikRouterInput struct {
	Name     string `json:"name" validate:"required"`
	Address  string `json:"address" validate:"required"` // IP address of the router
	ApiPort  *int   `json:"api_port"`                    // RouterOS API port, default 8728
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
	UseSSL   *bool  `json:"use_ssl"`
	IsActive *bool  `json:"is_active"`
}

// MikrotikTestResult holds the result of a connection test
type MikrotikTestResult struct {
	Success         bool   `json:"success"`
	Message         string `json:"message"`
	RouterOSVersion string `json:"router_os_version,omitempty"`
	Identity        string `json:"identity,omitempty"`
}

// PPPoE Models for JSON communication (from RouterOS)

type PppoeSecret struct {
	ID            string `json:"id,omitempty"` // .id from mikrotik
	Name          string `json:"name" validate:"required"`
	Password      string `json:"password" validate:"required"`
	Service       string `json:"service"` // pppoe
	CallerID      string `json:"caller_id"`
	Profile       string `json:"profile"`
	LocalAddress  string `json:"local_address"`
	RemoteAddress string `json:"remote_address"`
	Routes        string `json:"routes"`
	Comment       string `json:"comment"`
	Disabled      bool   `json:"disabled"`
}

type PppoeProfile struct {
	ID                string `json:"id,omitempty"`
	Name              string `json:"name" validate:"required"`
	LocalAddress      string `json:"local_address"`
	RemoteAddress     string `json:"remote_address"`
	Bridge            string `json:"bridge"`
	ChangeTCPMSS      string `json:"change_tcp_mss"` // default, yes, no
	RateLimit         string `json:"rate_limit"`
	ParentQueue       string `json:"parent_queue"`
	QueueType         string `json:"queue_type"`
	InsertQueueBefore string `json:"insert_queue_before"` // bottom, first, or queue name
	OnlyOne           string `json:"only_one"`            // default, yes, no
	UseMPLS           string `json:"use_mpls"`
	UseCompression    string `json:"use_compression"`
	UseEncryption     string `json:"use_encryption"`
	UseIPv6           string `json:"use_ipv6"`
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
	ProfileName string `json:"profile_name"` // PPP profile name, e.g., "isolir"
	AddressList string `json:"address_list"` // Firewall address list, e.g., "isolated-users"
	PortalIP    string `json:"portal_ip"`    // Captive portal IP address
	PortalPort  string `json:"portal_port"`  // Captive portal port (default "80")
	DNSServer   string `json:"dns_server"`   // DNS server for isolated users
	RateLimit   string `json:"rate_limit"`   // Minimal rate limit, e.g., "256k/256k"
}

// DefaultIsolationConfig returns a config with sensible defaults (portal IP must be set)
func DefaultIsolationConfig() IsolationConfig {
	return IsolationConfig{
		ProfileName: "isolir",
		AddressList: "isolated-users",
		PortalPort:  "80",
		RateLimit:   "256k/256k",
	}
}

// Reuse callback and webhook structs from before
type PppoeCallbackData struct {
	User       string `json:"user"`
	IP         string `json:"ip-address"`
	CallerID   string `json:"caller-id"` // MAC Address usually
	SessionID  string `json:"session-id"`
	Interface  string `json:"interface"`
	Uptime     string `json:"uptime"`
	BytesIn    int64  `json:"bytes-in"`
	BytesOut   int64  `json:"bytes-out"`
	PacketsIn  int64  `json:"packets-in"`
	PacketsOut int64  `json:"packets-out"`
	RouterID   uint   `json:"router_id"` // Included in callback URL query
}

// WebSocketMessage represents the message sent to connected clients
type WebSocketMessage struct {
	Event string      `json:"event"` // "session_up", "session_down"
	Data  interface{} `json:"data"`
}
