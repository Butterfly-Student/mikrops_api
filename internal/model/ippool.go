package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ─── IP Pools (DB entity) ─────────────────────────────────────────────────────

// IPPool represents an IP address pool managed by the application.
// Maps to the ip_pools table (differs from IpPool which is a RouterOS API DTO).
type IPPool struct {
	ID           uuid.UUID       `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	PoolName     string          `gorm:"size:100;not null" json:"pool_name" validate:"required,max=100"` // must match RouterOS name
	RouterID     uuid.UUID       `gorm:"type:uuid;not null" json:"router_id" validate:"required"`
	Router       *MikrotikRouter `gorm:"foreignKey:RouterID;constraint:OnDelete:RESTRICT" json:"router,omitempty"`
	Network      string          `gorm:"size:18;not null" json:"network" validate:"required"` // CIDR
	Gateway      string          `gorm:"size:15;not null" json:"gateway" validate:"required"`
	DNSPrimary   *string         `gorm:"size:15;default:'8.8.8.8'" json:"dns_primary,omitempty"`
	DNSSecondary *string         `gorm:"size:15;default:'8.8.4.4'" json:"dns_secondary,omitempty"`
	ServiceType  string     `gorm:"size:20;not null" json:"service_type" validate:"required,oneof=pppoe hotspot static vpn"`
	TotalIP      int             `gorm:"not null" json:"total_ip"`
	UsedIP       int             `gorm:"default:0" json:"used_ip"`
	IsActive     *bool           `gorm:"default:true" json:"is_active"`
}

func (IPPool) TableName() string { return "ip_pools" }

// IPPoolInput for creating/updating an IP pool
type IPPoolInput struct {
	PoolName     string    `json:"pool_name" validate:"required,max=100"`
	RouterID     uuid.UUID `json:"router_id" validate:"required"`
	Network      string    `json:"network" validate:"required"`
	Gateway      string    `json:"gateway" validate:"required"`
	DNSPrimary   *string   `json:"dns_primary"`
	DNSSecondary *string   `json:"dns_secondary"`
	ServiceType  string    `json:"service_type" validate:"required,oneof=pppoe hotspot static vpn"`
	TotalIP      int       `json:"total_ip" validate:"required,min=1"`
	IsActive     *bool     `json:"is_active"`
}

// IPAssignment records which IP is assigned to which subscription
type IPAssignment struct {
	ID             uuid.UUID     `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	SubscriptionID uuid.UUID     `gorm:"type:uuid;not null" json:"subscription_id" validate:"required"`
	Subscription   *Subscription `gorm:"foreignKey:SubscriptionID;constraint:OnDelete:RESTRICT" json:"subscription,omitempty"`
	IPPoolID       uuid.UUID     `gorm:"type:uuid;not null" json:"ip_pool_id" validate:"required"`
	IPPool         *IPPool       `gorm:"foreignKey:IPPoolID;constraint:OnDelete:RESTRICT" json:"ip_pool,omitempty"`
	IPAddress      string        `gorm:"size:18;not null" json:"ip_address" validate:"required"`
	MacAddress     *string       `gorm:"size:17" json:"mac_address,omitempty"`
	AssignedAt     time.Time     `gorm:"default:CURRENT_TIMESTAMP;not null" json:"assigned_at"`
	ReleasedAt     *time.Time    `json:"released_at,omitempty"`
	IsActive       *bool         `gorm:"default:true" json:"is_active"`
}

func (IPAssignment) TableName() string { return "ip_assignments" }

// ─── RouterOS API DTO (kept from old ippool.go) ────────────────────────────

// IpPool (RouterOS API DTO) represents a MikroTik IP Pool as returned by the API.
// Use IPPool (capital letters) for DB entities.
type IpPool struct {
	ID       string `json:"id,omitempty"`
	Name     string `json:"name" validate:"required"`
	Ranges   string `json:"ranges" validate:"required"`
	NextPool string `json:"next_pool"`
	Comment  string `json:"comment"`
}

// IPPoolFilter for querying IP pools
type IPPoolFilter struct {
	RouterID    *uuid.UUID   `json:"router_id"`
	ServiceType *ServiceType `json:"service_type"`
	IsActive    *bool        `json:"is_active"`
}

// IPAssignmentFilter for querying IP assignments
type IPAssignmentFilter struct {
	SubscriptionID *uuid.UUID `json:"subscription_id"`
	IPPoolID       *uuid.UUID `json:"ip_pool_id"`
	IsActive       *bool      `json:"is_active"`
}

// keep gorm import used by soft-delete if ever needed
var _ = gorm.DeletedAt{}
