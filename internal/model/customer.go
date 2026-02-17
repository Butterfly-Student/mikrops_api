package model

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

type Customer struct {
	ID                uuid.UUID         `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	CustomerCode      string            `json:"customer_code" gorm:"unique;not null" validate:"required"`
	FullName          string            `json:"full_name" gorm:"not null" validate:"required"`
	Email             *string           `json:"email" gorm:"unique" validate:"omitempty,email"`
	Phone             string            `json:"phone" gorm:"not null" validate:"required"`
	Address           *string           `json:"address" gorm:"type:text"`
	Latitude          *float64          `json:"latitude" gorm:"type:decimal(10,8)"`
	Longitude         *float64          `json:"longitude" gorm:"type:decimal(11,8)"`
	Status            string            `json:"status" gorm:"default:'pending';not null" validate:"required,oneof=pending active suspended isolated terminated"`
	ActivationDate    *time.Time        `json:"activation_date"`
	InstallationDate  *time.Time        `json:"installation_date"`
	TerminationDate   *time.Time        `json:"termination_date"`
	ExpiryDate        *time.Time        `json:"expiry_date" validate:"required"`
	RouterID          *uuid.UUID        `json:"router_id" gorm:"type:uuid" validate:"required"`
	Router            *MikrotikRouter   `json:"router,omitempty" gorm:"foreignKey:RouterID"`
	PppSecretName     *string           `json:"ppp_secret_name" gorm:"unique"`
	PppSecretPassword *string           `json:"-" gorm:"type:text"`
	PppService        *string           `json:"ppp_service" gorm:"default:'pppoe'" validate:"omitempty,oneof=pppoe pptp l2tp ovpn"`
	StaticIP          *string           `json:"static_ip" validate:"omitempty,ipv4"`
	MACAddress        *string           `json:"mac_address" validate:"omitempty,mac"`
	ProfileID         *uuid.UUID        `json:"profile_id" gorm:"type:uuid" validate:"required"`
	Profile           *BandwidthProfile `json:"profile,omitempty" gorm:"foreignKey:ProfileID"`
	BillingCycle      *string           `json:"billing_cycle" gorm:"default:'monthly'" validate:"omitempty,oneof=monthly quarterly yearly"`
	BillingDay        *int              `json:"billing_day" gorm:"default:1" validate:"omitempty,min=1,max=31"`
	PaymentMethodPref *string           `json:"payment_method_preference" gorm:"column:payment_method_preference" validate:"omitempty,oneof=cash transfer e-wallet auto-debit"`
	AutoIsolate       *bool             `json:"auto_isolate" gorm:"default:true"`
	GracePeriodDays   *int              `json:"grace_period_days" gorm:"default:3" validate:"omitempty,min=0"`
	Notes             *string           `json:"notes" gorm:"type:text"`
	Tags              pq.StringArray    `json:"tags" gorm:"type:text[] default:'{}'"`
	CreatedBy         *uint             `json:"created_by" gorm:"type:integer"`
	UpdatedBy         *uint             `json:"updated_by" gorm:"type:integer"`
	CreatedAt         time.Time         `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt         time.Time         `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt         *time.Time        `json:"-" gorm:"index"`
}

type CustomerInput struct {
	CustomerCode      *string        `json:"customer_code" validate:"required"`
	FullName          string         `json:"full_name" validate:"required"`
	Email             *string        `json:"email" validate:"omitempty,email"`
	Phone             string         `json:"phone" validate:"required"`
	Address           *string        `json:"address"`
	Latitude          *float64       `json:"latitude"`
	Longitude         *float64       `json:"longitude"`
	Status            *string        `json:"status" validate:"omitempty,oneof=pending active suspended isolated terminated"`
	ActivationDate    *time.Time     `json:"activation_date"`
	InstallationDate  *time.Time     `json:"installation_date"`
	TerminationDate   *time.Time     `json:"termination_date"`
	ExpiryDate        *time.Time     `json:"expiry_date" validate:"required"`
	RouterID          *uuid.UUID     `json:"router_id" validate:"required"`
	PppSecretName     *string        `json:"ppp_secret_name"`
	PppSecretPassword *string        `json:"-" validate:"required_with=PppSecretName"`
	PppService        *string        `json:"ppp_service" validate:"omitempty,oneof=pppoe pptp l2tp ovpn"`
	StaticIP          *string        `json:"static_ip" validate:"omitempty,ipv4"`
	MACAddress        *string        `json:"mac_address" validate:"omitempty,mac"`
	ProfileID         *uuid.UUID     `json:"profile_id" validate:"required"`
	BillingCycle      *string        `json:"billing_cycle" validate:"omitempty,oneof=monthly quarterly yearly"`
	BillingDay        *int           `json:"billing_day" validate:"omitempty,min=1,max=31"`
	PaymentMethodPref *string        `json:"payment_method_preference" validate:"omitempty,oneof=cash transfer e-wallet auto-debit"`
	AutoIsolate       *bool          `json:"auto_isolate"`
	GracePeriodDays   *int           `json:"grace_period_days" validate:"omitempty,min=0"`
	Notes             *string        `json:"notes"`
	Tags              pq.StringArray `json:"tags"`
}

type CustomerFilter struct {
	IDs            []uuid.UUID `json:"ids"`
	CustomerCodes  []string    `json:"customer_codes"`
	RouterIDs      []uuid.UUID `json:"router_ids"`
	ProfileIDs     []uuid.UUID `json:"profile_ids"`
	Status         []string    `json:"status"`
	Email          *string     `json:"email"`
	Phone          *string     `json:"phone"`
	IsActive       *bool       `json:"is_active"`
	IsExpired      *bool       `json:"is_expired"`
	WillExpireSoon *int        `json:"will_expire_soon"`
	Search         *string     `json:"search"`
}

func (f CustomerFilter) IsEmpty() bool {
	return len(f.IDs) == 0 && len(f.CustomerCodes) == 0 &&
		len(f.RouterIDs) == 0 && len(f.ProfileIDs) == 0 &&
		len(f.Status) == 0 && f.Email == nil &&
		f.Phone == nil && f.IsActive == nil &&
		f.IsExpired == nil && f.WillExpireSoon == nil && f.Search == nil
}

func CustomerPrepare(v *Customer) {
	if v.Status == "" {
		v.Status = "pending"
	}
	if v.AutoIsolate == nil {
		autoIsolate := true
		v.AutoIsolate = &autoIsolate
	}
	if v.GracePeriodDays == nil {
		gracePeriodDays := 3
		v.GracePeriodDays = &gracePeriodDays
	}
	if v.BillingCycle == nil {
		billingCycle := "monthly"
		v.BillingCycle = &billingCycle
	}
	if v.BillingDay == nil {
		billingDay := 1
		v.BillingDay = &billingDay
	}
	if v.PppService == nil {
		pppService := "pppoe"
		v.PppService = &pppService
	}
	if v.Tags == nil {
		v.Tags = pq.StringArray{}
	}
}

func ValidateCustomerForMikrotik(customer *Customer) error {
	if customer.PppSecretName == nil {
		return errors.New("ppp_secret_name is required for Mikrotik sync")
	}
	if customer.PppSecretPassword == nil {
		return errors.New("ppp_secret_password is required for Mikrotik sync")
	}
	if customer.RouterID == nil {
		return errors.New("router_id is required for Mikrotik sync")
	}
	if customer.ProfileID == nil {
		return errors.New("profile_id is required for Mikrotik sync")
	}
	return nil
}
