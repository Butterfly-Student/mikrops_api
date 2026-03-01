package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// RegistrationStatus represents the review status of a customer registration
type RegistrationStatus string

const (
	RegistrationStatusPending  RegistrationStatus = "pending"
	RegistrationStatusApproved RegistrationStatus = "approved"
	RegistrationStatusRejected RegistrationStatus = "rejected"
)

// CustomerRegistration represents a public registration request from a prospective customer.
// This is service-type agnostic - works for PPPoE, Hotspot, Static IP, and VPN.
type CustomerRegistration struct {
	ID uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`

	// Prospective customer information
	FullName  string   `gorm:"size:100;not null" json:"full_name"`
	Email     *string  `gorm:"size:100" json:"email"`
	Phone     string   `gorm:"size:20;not null" json:"phone"`
	Address   *string  `gorm:"type:text" json:"address"`
	Latitude  *float64 `gorm:"type:decimal(10,8)" json:"latitude"`
	Longitude *float64 `gorm:"type:decimal(11,8)" json:"longitude"`
	Notes     *string  `gorm:"type:text" json:"notes"` // message from customer

	// Selected package and optional preferred router
	BandwidthProfileID *uuid.UUID        `gorm:"type:uuid" json:"bandwidth_profile_id"`
	BandwidthProfile   *BandwidthProfile `gorm:"foreignKey:BandwidthProfileID;constraint:OnDelete:RESTRICT" json:"bandwidth_profile,omitempty"`
	PreferredRouterID  *uuid.UUID        `gorm:"type:uuid" json:"preferred_router_id"`

	// Review status
	Status          RegistrationStatus `gorm:"size:20;not null;default:pending" json:"status"`
	RejectionReason *string            `gorm:"type:text" json:"rejection_reason"`

	// Approval tracking
	ApprovedBy *uuid.UUID `gorm:"type:uuid" json:"approved_by"`
	ApprovedAt *time.Time `json:"approved_at"`

	// Link to created customer (set when approved)
	CustomerID *uuid.UUID `gorm:"type:uuid" json:"customer_id"`
	Customer   *Customer  `gorm:"foreignKey:CustomerID;constraint:OnDelete:SET NULL" json:"customer,omitempty"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName specifies the table name for CustomerRegistration
func (CustomerRegistration) TableName() string {
	return "customer_registrations"
}

// RegistrationInput is the public registration form payload
type RegistrationInput struct {
	FullName           string     `json:"full_name" validate:"required,max=100"`
	Email              *string    `json:"email" validate:"omitempty,email,max=100"`
	Phone              string     `json:"phone" validate:"required,max=20"`
	Address            *string    `json:"address"`
	Latitude           *float64   `json:"latitude" validate:"omitempty,min=-90,max=90"`
	Longitude          *float64   `json:"longitude" validate:"omitempty,min=-180,max=180"`
	Notes              *string    `json:"notes"`
	BandwidthProfileID *uuid.UUID `json:"bandwidth_profile_id" validate:"required,uuid"`
	PreferredRouterID  *uuid.UUID `json:"preferred_router_id" validate:"omitempty,uuid"`
}

// RegistrationApproveInput is the admin approval payload
type RegistrationApproveInput struct {
	RouterID uuid.UUID `json:"router_id" validate:"required"`
}

// RegistrationRejectInput is the admin rejection payload
type RegistrationRejectInput struct {
	Reason string `json:"reason" validate:"required,min=3,max=500"`
}

// RegistrationFilter for listing registrations
type RegistrationFilter struct {
	Status             *RegistrationStatus `json:"status"`
	BandwidthProfileID *uuid.UUID          `json:"bandwidth_profile_id"`
	Search             *string             `json:"search"` // name, email, phone
}

// RegistrationApprovalResult is the approval response — includes sensitive credentials shown only once.
// Credentials are generated based on the service type of the selected bandwidth profile.
type RegistrationApprovalResult struct {
	Registration          *CustomerRegistration `json:"registration"`
	Customer              *Customer             `json:"customer"`
	Subscription          *Subscription         `json:"subscription"`
	InitialPortalPassword string                `json:"initial_portal_password"` // plain-text, shown once
	// Service credentials - type depends on BandwidthProfile.ServiceType:
	// - PPPoE: PPP username/password
	// - Hotspot: Hotspot username/password  
	// - Static IP: Identifier/IP info
	// - VPN: VPN username/password
	ServiceUsername string `json:"service_username"`
	ServicePassword string `json:"service_password"` // plain-text, shown once
}
