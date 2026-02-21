package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// CustomerStatus represents customer service status
type CustomerStatus string

const (
	CustomerStatusPending    CustomerStatus = "pending"    // menunggu instalasi
	CustomerStatusActive     CustomerStatus = "active"     // aktif berlangganan
	CustomerStatusSuspended  CustomerStatus = "suspended"  // ditangguhkan manual
	CustomerStatusIsolated   CustomerStatus = "isolated"   // diisolasi karena telat bayar
	CustomerStatusTerminated CustomerStatus = "terminated" // berakhir
)

// BillingCycle represents billing cycle type
type BillingCycle string

const (
	BillingCycleMonthly   BillingCycle = "monthly"
	BillingCycleQuarterly BillingCycle = "quarterly"
	BillingCycleYearly    BillingCycle = "yearly"
)

// PPPService represents PPP service type
type PPPService string

const (
	PPPServicePPPoE PPPService = "pppoe"
	PPPServicePPTP  PPPService = "pptp"
	PPPServiceL2TP  PPPService = "l2tp"
	PPPServiceOVPN  PPPService = "ovpn"
)

// Customer represents internet service customer
type Customer struct {
	ID           uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	CustomerCode string         `gorm:"size:50;uniqueIndex;not null" json:"customer_code" validate:"required"`

	// Personal Information
	FullName string  `gorm:"size:100;not null" json:"full_name" validate:"required"`
	Email    *string `gorm:"size:100;uniqueIndex" json:"email" validate:"omitempty,email"`
	Phone    string  `gorm:"size:20;not null" json:"phone" validate:"required"`

	// Address Information
	Address   *string  `gorm:"type:text" json:"address"`
	Latitude  *float64 `gorm:"type:decimal(10,8)" json:"latitude"`
	Longitude *float64 `gorm:"type:decimal(11,8)" json:"longitude"`

	// Service Information
	Status           CustomerStatus `gorm:"size:20;not null;default:pending" json:"status"`
	ActivationDate   *time.Time     `gorm:"type:date" json:"activation_date"`
	InstallationDate *time.Time     `gorm:"type:date" json:"installation_date"`
	TerminationDate  *time.Time     `gorm:"type:date" json:"termination_date"`
	ExpiryDate       *time.Time     `gorm:"type:date" json:"expiry_date"` // tanggal jatuh tempo

	// Mikrotik Configuration
	RouterID          *uuid.UUID `gorm:"type:uuid" json:"router_id"`
	Router            *MikrotikRouter `gorm:"foreignKey:RouterID;constraint:OnDelete:RESTRICT" json:"router,omitempty"`
	PppSecretName     *string    `gorm:"size:100;uniqueIndex" json:"ppp_secret_name"`
	PppSecretPassword *string    `gorm:"size:255" json:"ppp_secret_password"` // encrypted
	PppService        PPPService `gorm:"size:20;default:pppoe" json:"ppp_service"`
	StaticIP          *string    `gorm:"size:45" json:"static_ip"`
	MacAddress        *string    `gorm:"size:17" json:"mac_address"`

	// Package Information
	ProfileID         *uuid.UUID        `gorm:"type:uuid" json:"profile_id"`
	Profile           *BandwidthProfile `gorm:"foreignKey:ProfileID;constraint:OnDelete:RESTRICT" json:"profile,omitempty"`
	PreviousProfileID *uuid.UUID        `gorm:"type:uuid" json:"previous_profile_id"`

	// Billing Information
	BillingCycle            BillingCycle `gorm:"size:20;default:monthly" json:"billing_cycle"`
	BillingDay              *int         `gorm:"default:1" json:"billing_day" validate:"omitempty,min=1,max=31"`
	PaymentMethodPreference *string      `gorm:"size:20" json:"payment_method_preference"`
	AutoIsolate             *bool        `gorm:"default:true" json:"auto_isolate"`
	GracePeriodDays         *int         `gorm:"default:3" json:"grace_period_days" validate:"omitempty,min=0"`

	Notes *string                `gorm:"type:text" json:"notes"`
	Tags  map[string]interface{} `gorm:"type:jsonb" json:"tags"`

	// Customer Portal
	PortalPassword  *string    `gorm:"size:255" json:"-"`                         // bcrypt hashed, never exposed
	PortalLastLogin *time.Time `json:"portal_last_login,omitempty"`

	CreatedBy *uuid.UUID `gorm:"type:uuid" json:"created_by"`
	UpdatedBy *uuid.UUID `gorm:"type:uuid" json:"updated_by"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName specifies the table name for Customer
func (Customer) TableName() string {
	return "customers"
}

// CustomerInput for creating/updating customer
type CustomerInput struct {
	CustomerCode            string                 `json:"customer_code" validate:"required,max=50"`
	FullName                string                 `json:"full_name" validate:"required,max=100"`
	Email                   *string                `json:"email" validate:"omitempty,email,max=100"`
	Phone                   string                 `json:"phone" validate:"required,max=20"`
	Address                 *string                `json:"address"`
	Latitude                *float64               `json:"latitude" validate:"omitempty,min=-90,max=90"`
	Longitude               *float64               `json:"longitude" validate:"omitempty,min=-180,max=180"`
	Status                  *string                `json:"status" validate:"omitempty,oneof=pending active suspended isolated terminated"`
	ActivationDate          *time.Time             `json:"activation_date"`
	InstallationDate        *time.Time             `json:"installation_date"`
	TerminationDate         *time.Time             `json:"termination_date"`
	ExpiryDate              *time.Time             `json:"expiry_date"`
	RouterID                *uuid.UUID             `json:"router_id" validate:"omitempty,uuid"`
	PppSecretName           *string                `json:"ppp_secret_name" validate:"omitempty,max=100"`
	PppSecretPassword       *string                `json:"ppp_secret_password" validate:"omitempty,max=255"`
	PppService              *string                `json:"ppp_service" validate:"omitempty,oneof=pppoe pptp l2tp ovpn"`
	StaticIP                *string                `json:"static_ip" validate:"omitempty,ip"`
	MacAddress              *string                `json:"mac_address" validate:"omitempty,mac"`
	ProfileID               *uuid.UUID             `json:"profile_id" validate:"omitempty,uuid"`
	BillingCycle            *string                `json:"billing_cycle" validate:"omitempty,oneof=monthly quarterly yearly"`
	BillingDay              *int                   `json:"billing_day" validate:"omitempty,min=1,max=31"`
	PaymentMethodPreference *string                `json:"payment_method_preference" validate:"omitempty,oneof=cash transfer e-wallet auto-debit"`
	AutoIsolate             *bool                  `json:"auto_isolate"`
	GracePeriodDays         *int                   `json:"grace_period_days" validate:"omitempty,min=0"`
	Notes                   *string                `json:"notes"`
	Tags                    map[string]interface{} `json:"tags"`
}

// CustomerFilter for filtering customers
type CustomerFilter struct {
	Status           *CustomerStatus `json:"status"`
	RouterID         *uuid.UUID      `json:"router_id"`
	ProfileID        *uuid.UUID      `json:"profile_id"`
	BillingCycle     *BillingCycle   `json:"billing_cycle"`
	AutoIsolate      *bool           `json:"auto_isolate"`
	ExpiryDateFrom   *time.Time      `json:"expiry_date_from"`
	ExpiryDateTo     *time.Time      `json:"expiry_date_to"`
	Search           *string         `json:"search"` // search by name, code, phone, email
	IncludeDeleted   bool            `json:"include_deleted"`
}

// ToModel converts input to model
func (i *CustomerInput) ToModel() *Customer {
	customer := &Customer{
		CustomerCode:            i.CustomerCode,
		FullName:                i.FullName,
		Email:                   i.Email,
		Phone:                   i.Phone,
		Address:                 i.Address,
		Latitude:                i.Latitude,
		Longitude:               i.Longitude,
		ActivationDate:          i.ActivationDate,
		InstallationDate:        i.InstallationDate,
		TerminationDate:         i.TerminationDate,
		ExpiryDate:              i.ExpiryDate,
		RouterID:                i.RouterID,
		PppSecretName:           i.PppSecretName,
		PppSecretPassword:       i.PppSecretPassword,
		StaticIP:                i.StaticIP,
		MacAddress:              i.MacAddress,
		ProfileID:               i.ProfileID,
		BillingDay:              i.BillingDay,
		PaymentMethodPreference: i.PaymentMethodPreference,
		AutoIsolate:             i.AutoIsolate,
		GracePeriodDays:         i.GracePeriodDays,
		Notes:                   i.Notes,
		Tags:                    i.Tags,
	}

	// Set status
	if i.Status != nil {
		customer.Status = CustomerStatus(*i.Status)
	} else {
		customer.Status = CustomerStatusPending
	}

	// Set PPP service
	if i.PppService != nil {
		customer.PppService = PPPService(*i.PppService)
	} else {
		customer.PppService = PPPServicePPPoE
	}

	// Set billing cycle
	if i.BillingCycle != nil {
		customer.BillingCycle = BillingCycle(*i.BillingCycle)
	} else {
		customer.BillingCycle = BillingCycleMonthly
	}

	return customer
}

// IsActive checks if customer is in active status
func (c *Customer) IsActive() bool {
	return c.Status == CustomerStatusActive
}

// IsIsolated checks if customer is isolated
func (c *Customer) IsIsolated() bool {
	return c.Status == CustomerStatusIsolated
}

// CanBeIsolated checks if customer can be auto-isolated
func (c *Customer) CanBeIsolated() bool {
	if c.AutoIsolate == nil {
		return true // default is true
	}
	return *c.AutoIsolate && c.IsActive()
}

// GetGracePeriodDays returns grace period days
func (c *Customer) GetGracePeriodDays() int {
	if c.GracePeriodDays == nil {
		return 3 // default
	}
	return *c.GracePeriodDays
}

// GetBillingDay returns billing day
func (c *Customer) GetBillingDay() int {
	if c.BillingDay == nil {
		return 1 // default
	}
	return *c.BillingDay
}
