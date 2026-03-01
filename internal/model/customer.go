package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// CustomerStatus represents customer account status
type CustomerStatus string

const (
	CustomerStatusPending    CustomerStatus = "pending"
	CustomerStatusActive     CustomerStatus = "active"
	CustomerStatusSuspended  CustomerStatus = "suspended"
	CustomerStatusIsolated   CustomerStatus = "isolated"
	CustomerStatusTerminated CustomerStatus = "terminated"
)

// Customer represents the identity of an ISP subscriber.
// Service-specific configs (PPPoE creds, IP, router) are in Subscription.
type Customer struct {
	ID           uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	CustomerCode string    `gorm:"size:50;uniqueIndex;not null" json:"customer_code" validate:"required,max=50"`

	// Personal Identity
	FullName     string  `gorm:"size:100;not null" json:"full_name" validate:"required,max=100"`
	Email        *string `gorm:"size:100;uniqueIndex" json:"email,omitempty" validate:"omitempty,email,max=100"`
	Phone        string  `gorm:"size:20;not null" json:"phone" validate:"required,max=20"`
	IDCardNumber *string `gorm:"size:30" json:"id_card_number,omitempty"` // NIK

	// Address
	Address   *string  `gorm:"type:text" json:"address,omitempty"`
	Latitude  *float64 `gorm:"type:decimal(10,8)" json:"latitude,omitempty" validate:"omitempty,min=-90,max=90"`
	Longitude *float64 `gorm:"type:decimal(11,8)" json:"longitude,omitempty" validate:"omitempty,min=-180,max=180"`

	// Service Status
	Status           CustomerStatus `gorm:"size:20;not null;default:pending" json:"status"`
	ActivationDate   *time.Time     `gorm:"type:date" json:"activation_date,omitempty"`
	InstallationDate *time.Time     `gorm:"type:date" json:"installation_date,omitempty"`
	TerminationDate  *time.Time     `gorm:"type:date" json:"termination_date,omitempty"`

	// Billing behaviour
	AutoIsolate     *bool `gorm:"default:true" json:"auto_isolate"`
	GracePeriodDays *int  `gorm:"default:3" json:"grace_period_days" validate:"omitempty,min=0"`

	// Customer portal login
	PortalPassword  *string    `gorm:"size:255" json:"-"` // bcrypt hashed, never exposed
	PortalLastLogin *time.Time `json:"portal_last_login,omitempty"`

	Notes *string `gorm:"type:text" json:"notes,omitempty"`
	Tags  *string  `gorm:"type:jsonb" json:"tags,omitempty"` // JSON tags

	// Audit
	CreatedBy *uuid.UUID     `gorm:"type:uuid" json:"created_by,omitempty"`
	UpdatedBy *uuid.UUID     `gorm:"type:uuid" json:"updated_by,omitempty"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// Relations (loaded on demand)
	Subscriptions []Subscription `gorm:"foreignKey:CustomerID" json:"subscriptions,omitempty"`
}

// TableName specifies the table name
func (Customer) TableName() string { return "customers" }

// CustomerInput for creating/updating a customer
type CustomerInput struct {
	CustomerCode     string     `json:"customer_code" validate:"required,max=50"`
	FullName         string     `json:"full_name" validate:"required,max=100"`
	Email            *string    `json:"email" validate:"omitempty,email,max=100"`
	Phone            string     `json:"phone" validate:"required,max=20"`
	IDCardNumber     *string    `json:"id_card_number" validate:"omitempty,max=30"`
	Address          *string    `json:"address"`
	Latitude         *float64   `json:"latitude" validate:"omitempty,min=-90,max=90"`
	Longitude        *float64   `json:"longitude" validate:"omitempty,min=-180,max=180"`
	Status           *string    `json:"status" validate:"omitempty,oneof=pending active suspended isolated terminated"`
	ActivationDate   *time.Time `json:"activation_date"`
	InstallationDate *time.Time `json:"installation_date"`
	TerminationDate  *time.Time `json:"termination_date"`
	AutoIsolate      *bool      `json:"auto_isolate"`
	GracePeriodDays  *int       `json:"grace_period_days" validate:"omitempty,min=0"`
	Notes            *string    `json:"notes"`
	Tags             *string    `json:"tags"` // JSON string
}

// CustomerFilter for querying customers
type CustomerFilter struct {
	Status         *CustomerStatus `json:"status"`
	AutoIsolate    *bool           `json:"auto_isolate"`
	Search         *string         `json:"search"` // name, code, phone, email
	IncludeDeleted bool            `json:"include_deleted"`
}

// ToModel converts input to Customer
func (i *CustomerInput) ToModel() *Customer {
	c := &Customer{
		CustomerCode:     i.CustomerCode,
		FullName:         i.FullName,
		Email:            i.Email,
		Phone:            i.Phone,
		IDCardNumber:     i.IDCardNumber,
		Address:          i.Address,
		Latitude:         i.Latitude,
		Longitude:        i.Longitude,
		ActivationDate:   i.ActivationDate,
		InstallationDate: i.InstallationDate,
		TerminationDate:  i.TerminationDate,
		AutoIsolate:      i.AutoIsolate,
		GracePeriodDays:  i.GracePeriodDays,
		Notes:            i.Notes,
		Tags:             i.Tags,
	}
	if i.Status != nil {
		c.Status = CustomerStatus(*i.Status)
	} else {
		c.Status = CustomerStatusPending
	}
	return c
}

// IsActive returns true when customer account is active
func (c *Customer) IsActive() bool { return c.Status == CustomerStatusActive }

// IsIsolated returns true when customer is isolated
func (c *Customer) IsIsolated() bool { return c.Status == CustomerStatusIsolated }

// CanBeIsolated returns true when auto-isolation is enabled and customer is active
func (c *Customer) CanBeIsolated() bool {
	if c.AutoIsolate == nil {
		return true
	}
	return *c.AutoIsolate && c.IsActive()
}

// GetGracePeriodDays returns grace period days (default 3)
func (c *Customer) GetGracePeriodDays() int {
	if c.GracePeriodDays == nil {
		return 3
	}
	return *c.GracePeriodDays
}
