package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// SubscriptionStatus represents the service status for a subscription
type SubscriptionStatus string

const (
	SubscriptionStatusPending    SubscriptionStatus = "pending"
	SubscriptionStatusActive     SubscriptionStatus = "active"
	SubscriptionStatusSuspended  SubscriptionStatus = "suspended"
	SubscriptionStatusIsolated   SubscriptionStatus = "isolated"
	SubscriptionStatusExpired    SubscriptionStatus = "expired"
	SubscriptionStatusTerminated SubscriptionStatus = "terminated"
)

// VPNType represents VPN protocol type
type VPNType string

const (
	VPNTypePPTP VPNType = "pptp"
	VPNTypeL2TP VPNType = "l2tp"
	VPNTypeSSTP VPNType = "sstp"
	VPNTypeOVPN VPNType = "ovpn"
)

// Subscription represents an active service subscription linking a Customer to a Plan on a Router.
// Replaces the old PPPoE fields that were previously on Customer.
type Subscription struct {
	ID         uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	CustomerID uuid.UUID `gorm:"type:uuid;not null" json:"customer_id" validate:"required"`
	PlanID     uuid.UUID `gorm:"type:uuid;not null" json:"plan_id" validate:"required"`
	RouterID   uuid.UUID `gorm:"type:uuid;not null" json:"router_id" validate:"required"`

	// Service Type — must match plan's service_type
	ServiceType ServiceType `gorm:"size:20;not null;default:pppoe" json:"service_type"`

	// MikroTik credentials
	Username string `gorm:"size:100;uniqueIndex;not null" json:"username" validate:"required,max=100"`
	Password string `gorm:"size:255;not null" json:"-"` // encrypted

	// Network config
	StaticIP   *string  `gorm:"size:45" json:"static_ip,omitempty"`
	Gateway    *string  `gorm:"size:15" json:"gateway,omitempty"`
	MacAddress *string  `gorm:"size:17" json:"mac_address,omitempty"`
	VPNType    *VPNType `gorm:"size:20" json:"vpn_type,omitempty"`

	// Status & billing period
	Status       SubscriptionStatus `gorm:"size:20;not null;default:pending" json:"status"`
	ActivatedAt  *time.Time         `json:"activated_at,omitempty"`
	ExpiredAt    *time.Time         `json:"expired_at,omitempty"`
	ExpiryDate   *time.Time         `gorm:"type:date" json:"expiry_date,omitempty"`
	BillingCycle BillingCycle       `gorm:"size:20;default:monthly" json:"billing_cycle"`
	BillingDay   *int               `gorm:"default:1" json:"billing_day,omitempty" validate:"omitempty,min=1,max=31"`

	// Suspend / terminate
	SuspendReason *string    `gorm:"type:text" json:"suspend_reason,omitempty"`
	TerminatedAt  *time.Time `json:"terminated_at,omitempty"`

	// MikroTik sync tracking
	MtSynced   bool       `gorm:"default:false" json:"mt_synced"`
	MtLastSync *time.Time `json:"mt_last_sync,omitempty"`
	MtError    *string    `gorm:"type:text" json:"mt_error,omitempty"`

	// Previous plan (for restore after isolation)
	PreviousPlanID *uuid.UUID `gorm:"type:uuid" json:"previous_plan_id,omitempty"`

	Notes     *string        `gorm:"type:text" json:"notes,omitempty"`
	CreatedBy *uuid.UUID     `gorm:"type:uuid" json:"created_by,omitempty"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// Relations (loaded on demand)
	Customer *Customer         `gorm:"foreignKey:CustomerID;constraint:OnDelete:RESTRICT" json:"customer,omitempty"`
	Plan     *BandwidthProfile `gorm:"foreignKey:PlanID;constraint:OnDelete:RESTRICT" json:"plan,omitempty"`
	Router   *MikrotikRouter   `gorm:"foreignKey:RouterID;constraint:OnDelete:RESTRICT" json:"router,omitempty"`
}

// TableName specifies the table name
func (Subscription) TableName() string { return "subscriptions" }

// SubscriptionInput for creating/updating a subscription
type SubscriptionInput struct {
	CustomerID   uuid.UUID  `json:"customer_id" validate:"required"`
	PlanID       uuid.UUID  `json:"plan_id" validate:"required"`
	RouterID     uuid.UUID  `json:"router_id" validate:"required"`
	ServiceType  string     `json:"service_type" validate:"required,oneof=pppoe hotspot static_ip vpn"`
	Username     string     `json:"username" validate:"required,max=100"`
	Password     string     `json:"password" validate:"required,max=255"`
	StaticIP     *string    `json:"static_ip" validate:"omitempty,ip"`
	Gateway      *string    `json:"gateway" validate:"omitempty,ip"`
	MacAddress   *string    `json:"mac_address" validate:"omitempty"`
	VPNType      *string    `json:"vpn_type" validate:"omitempty,oneof=pptp l2tp sstp ovpn"`
	Status       *string    `json:"status" validate:"omitempty,oneof=pending active suspended isolated expired terminated"`
	BillingCycle *string    `json:"billing_cycle" validate:"omitempty,oneof=daily weekly monthly yearly"`
	BillingDay   *int       `json:"billing_day" validate:"omitempty,min=1,max=31"`
	ExpiryDate   *time.Time `json:"expiry_date"`
	Notes        *string    `json:"notes"`
}

// SubscriptionFilter for querying subscriptions
type SubscriptionFilter struct {
	CustomerID  *uuid.UUID          `json:"customer_id"`
	RouterID    *uuid.UUID          `json:"router_id"`
	PlanID      *uuid.UUID          `json:"plan_id"`
	Status      *SubscriptionStatus `json:"status"`
	ServiceType *ServiceType        `json:"service_type"`
	MtSynced    *bool               `json:"mt_synced"`
	Username    *string             `json:"username"`
	ExpiryFrom  *time.Time          `json:"expiry_from"`
	ExpiryTo    *time.Time          `json:"expiry_to"`
}

// ToModel converts input to Subscription
func (i *SubscriptionInput) ToModel() *Subscription {
	s := &Subscription{
		CustomerID:  i.CustomerID,
		PlanID:      i.PlanID,
		RouterID:    i.RouterID,
		ServiceType: ServiceType(i.ServiceType),
		Username:    i.Username,
		Password:    i.Password,
		StaticIP:    i.StaticIP,
		Gateway:     i.Gateway,
		MacAddress:  i.MacAddress,
		ExpiryDate:  i.ExpiryDate,
		Notes:       i.Notes,
	}
	if i.VPNType != nil {
		vt := VPNType(*i.VPNType)
		s.VPNType = &vt
	}
	if i.Status != nil {
		s.Status = SubscriptionStatus(*i.Status)
	} else {
		s.Status = SubscriptionStatusPending
	}
	if i.BillingCycle != nil {
		s.BillingCycle = BillingCycle(*i.BillingCycle)
	} else {
		s.BillingCycle = BillingCycleMonthly
	}
	s.BillingDay = i.BillingDay
	return s
}

// IsActive returns true when subscription is active
func (s *Subscription) IsActive() bool { return s.Status == SubscriptionStatusActive }

// NeedsMtSync returns true when the subscription is not yet synced to MikroTik
func (s *Subscription) NeedsMtSync() bool { return !s.MtSynced }
