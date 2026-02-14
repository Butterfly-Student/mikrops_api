package model

import (
	"database/sql/driver"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// CustomerStatus represents the status of a customer
type CustomerStatus string

const (
	CustomerStatusPending    CustomerStatus = "pending"
	CustomerStatusActive     CustomerStatus = "active"
	CustomerStatusSuspended  CustomerStatus = "suspended"
	CustomerStatusIsolated   CustomerStatus = "isolated"
	CustomerStatusTerminated CustomerStatus = "terminated"
)

// PPPServiceType represents the PPP service type
type PPPServiceType string

const (
	PPPServiceTypePPPoE PPPServiceType = "pppoe"
	PPPServiceTypePPTP  PPPServiceType = "pptp"
	PPPServiceTypeL2TP  PPPServiceType = "l2tp"
	PPPServiceTypeOVPN  PPPServiceType = "ovpn"
	PPPServiceTypeAny   PPPServiceType = "any"
)

// BillingCycleType represents the billing cycle
type BillingCycleType string

const (
	BillingCycleMonthly      BillingCycleType = "monthly"
	BillingCycleQuarterly    BillingCycleType = "quarterly"
	BillingCycleSemiAnnually BillingCycleType = "semi-annually"
	BillingCycleYearly       BillingCycleType = "yearly"
)

// Coordinates represents geographic coordinates
type Coordinates struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

// Scan implements sql.Scanner for Coordinates
func (c *Coordinates) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(bytes, c)
}

// Value implements driver.Valuer for Coordinates
func (c Coordinates) Value() (driver.Value, error) {
	return json.Marshal(c)
}

type Customer struct {
	ID                      uuid.UUID         `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	CustomerCode            string            `json:"customer_code" gorm:"uniqueIndex;not null"`
	FullName                string            `json:"full_name" gorm:"not null"`
	Email                   string            `json:"email" gorm:"index"`
	Phone                   string            `json:"phone" gorm:"index"`
	Address                 string            `json:"address"`
	Coordinates             *Coordinates      `json:"coordinates" gorm:"type:geography(POINT,4326)"`
	Status                  CustomerStatus    `json:"status" gorm:"type:customer_status;not null;default:'pending';index"`
	ActivationDate          *time.Time        `json:"activation_date"`
	InstallationDate        *time.Time        `json:"installation_date"`
	TerminationDate         *time.Time        `json:"termination_date"`
	ExpiryDate              *time.Time        `json:"expiry_date" gorm:"index"`
	RouterID                *uuid.UUID        `json:"router_id" gorm:"type:uuid;index"`
	Router                  *MikrotikRouter   `json:"router,omitempty" gorm:"foreignKey:RouterID;references:ID"`
	PppSecretName           string            `json:"ppp_secret_name" gorm:"uniqueIndex"`
	PppSecretPassword       string            `json:"ppp_secret_password"`
	PppService              PPPServiceType    `json:"ppp_service" gorm:"type:ppp_service_type;default:'pppoe'"`
	StaticIP                string            `json:"static_ip"`
	MacAddress              string            `json:"mac_address"`
	ProfileID               *uuid.UUID        `json:"profile_id" gorm:"type:uuid;index"`
	Profile                 *BandwidthProfile `json:"profile,omitempty" gorm:"foreignKey:ProfileID;references:ID"`
	BillingCycle            BillingCycleType  `json:"billing_cycle" gorm:"type:billing_cycle_type;default:'monthly'"`
	BillingDay              int               `json:"billing_day" gorm:"default:1"`
	PaymentMethodPreference string            `json:"payment_method_preference"`
	AutoIsolate             bool              `json:"auto_isolate" gorm:"default:true"`
	GracePeriodDays         int               `json:"grace_period_days" gorm:"default:3"`
	Notes                   string            `json:"notes"`
	Tags                    JSONBArray        `json:"tags" gorm:"type:jsonb;default:'[]'"`
	CreatedBy               *uint             `json:"created_by" gorm:"index"`
	CreatedByUser           *User             `json:"created_by_user,omitempty" gorm:"foreignKey:CreatedBy;references:ID"`
	UpdatedBy               *uint             `json:"updated_by" gorm:"index"`
	UpdatedByUser           *User             `json:"updated_by_user,omitempty" gorm:"foreignKey:UpdatedBy;references:ID"`
	CreatedAt               time.Time         `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt               time.Time         `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt               *time.Time        `json:"deleted_at,omitempty" gorm:"index"`
}

type CustomerInput struct {
	CustomerCode            string           `json:"customer_code" binding:"required"`
	FullName                string           `json:"full_name" binding:"required"`
	Email                   string           `json:"email" binding:"omitempty,email"`
	Phone                   string           `json:"phone"`
	Address                 string           `json:"address"`
	Coordinates             *Coordinates     `json:"coordinates"`
	Status                  CustomerStatus   `json:"status" binding:"omitempty,oneof=pending active suspended isolated terminated"`
	ActivationDate          *time.Time       `json:"activation_date"`
	InstallationDate        *time.Time       `json:"installation_date"`
	TerminationDate         *time.Time       `json:"termination_date"`
	ExpiryDate              *time.Time       `json:"expiry_date"`
	RouterID                *uuid.UUID       `json:"router_id"`
	PppSecretName           string           `json:"ppp_secret_name" binding:"required"`
	PppSecretPassword       string           `json:"ppp_secret_password" binding:"required"`
	PppService              PPPServiceType   `json:"ppp_service" binding:"omitempty,oneof=pppoe pptp l2tp ovpn any"`
	StaticIP                string           `json:"static_ip"`
	MacAddress              string           `json:"mac_address"`
	ProfileID               *uuid.UUID       `json:"profile_id" binding:"required"`
	BillingCycle            BillingCycleType `json:"billing_cycle" binding:"omitempty,oneof=monthly quarterly semi-annually yearly"`
	BillingDay              int              `json:"billing_day" binding:"omitempty,min=1,max=31"`
	PaymentMethodPreference string           `json:"payment_method_preference"`
	AutoIsolate             bool             `json:"auto_isolate"`
	GracePeriodDays         int              `json:"grace_period_days" binding:"omitempty,min=0"`
	Notes                   string           `json:"notes"`
	Tags                    []string         `json:"tags"`
}

type CustomerFilter struct {
	IDs           []uuid.UUID      `json:"ids"`
	CustomerCodes []string         `json:"customer_codes"`
	Statuses      []CustomerStatus `json:"statuses"`
	RouterIDs     []uuid.UUID      `json:"router_ids"`
	ProfileIDs    []uuid.UUID      `json:"profile_ids"`
	ExpiredBefore *time.Time       `json:"expired_before"`
	ExpiredAfter  *time.Time       `json:"expired_after"`
	Search        string           `json:"search"`
	Page          int              `json:"page"`
	Limit         int              `json:"limit"`
}

func (f CustomerFilter) IsEmpty() bool {
	return len(f.IDs) == 0 && len(f.CustomerCodes) == 0 && len(f.Statuses) == 0 &&
		len(f.RouterIDs) == 0 && len(f.ProfileIDs) == 0 && f.ExpiredBefore == nil &&
		f.ExpiredAfter == nil && f.Search == ""
}

// TableName specifies the table name for GORM
func (Customer) TableName() string {
	return "customers"
}

// IsExpired checks if customer's expiry date has passed
func (c *Customer) IsExpired() bool {
	if c.ExpiryDate == nil {
		return false
	}
	return c.ExpiryDate.Before(time.Now())
}

// IsInGracePeriod checks if customer is within grace period
func (c *Customer) IsInGracePeriod() bool {
	if c.ExpiryDate == nil {
		return false
	}
	gracePeriodEnd := c.ExpiryDate.AddDate(0, 0, c.GracePeriodDays)
	now := time.Now()
	return now.After(*c.ExpiryDate) && now.Before(gracePeriodEnd)
}

// ShouldBeIsolated checks if customer should be isolated
func (c *Customer) ShouldBeIsolated() bool {
	if !c.AutoIsolate {
		return false
	}
	if c.Status == CustomerStatusIsolated || c.Status == CustomerStatusTerminated {
		return false
	}
	return c.IsExpired() && !c.IsInGracePeriod()
}
