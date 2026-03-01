package model

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ServiceType represents the type of ISP service
type ServiceType string

const (
	ServiceTypePPPoE    ServiceType = "pppoe"
	ServiceTypeHotspot  ServiceType = "hotspot"
	ServiceTypeStaticIP ServiceType = "static_ip"
	ServiceTypeVPN      ServiceType = "vpn"
)

// BillingCycle represents billing cycle duration
type BillingCycle string

const (
	BillingCycleDaily   BillingCycle = "daily"
	BillingCycleWeekly  BillingCycle = "weekly"
	BillingCycleMonthly BillingCycle = "monthly"
	BillingCycleYearly  BillingCycle = "yearly"
)

// ProfileCategory represents customer category for a bandwidth plan
type ProfileCategory string

const (
	ProfileCategoryResidential ProfileCategory = "residential"
	ProfileCategoryBusiness    ProfileCategory = "business"
	ProfileCategoryCorporate   ProfileCategory = "corporate"
	ProfileCategoryPromo       ProfileCategory = "promo"
)

// BandwidthProfile represents a bandwidth plan/package
type BandwidthProfile struct {
	ID          uuid.UUID       `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	ProfileCode string          `gorm:"size:50;uniqueIndex;not null" json:"profile_code" validate:"required,max=50"`
	Name        string          `gorm:"size:100;not null" json:"name" validate:"required,max=100"`
	Description *string         `gorm:"type:text" json:"description,omitempty"`
	ServiceType ServiceType     `gorm:"size:20;not null;default:pppoe" json:"service_type" validate:"required,oneof=pppoe hotspot static_ip vpn"`
	Category    ProfileCategory `gorm:"size:20;not null;default:residential" json:"category" validate:"required,oneof=residential business corporate promo"`

	// MikroTik profile name (PPP profile / hotspot profile)
	PppProfileName string `gorm:"size:100;not null" json:"ppp_profile_name" validate:"required,max=100"`

	// PPP Network Configuration
	LocalAddress  *string `gorm:"size:45" json:"local_address,omitempty"`
	RemoteAddress *string `gorm:"size:45" json:"remote_address,omitempty"`
	ParentQueue   *string `gorm:"size:100" json:"parent_queue,omitempty"`
	DNSServer     *string `gorm:"size:100" json:"dns_server,omitempty"`

	// Speed (kbps)
	DownloadSpeed int64 `gorm:"not null" json:"download_speed" validate:"required,min=1"`
	UploadSpeed   int64 `gorm:"not null" json:"upload_speed" validate:"required,min=1"`

	// Burst
	BurstDownload  *int64 `json:"burst_download,omitempty"`
	BurstUpload    *int64 `json:"burst_upload,omitempty"`
	BurstThreshold *int   `gorm:"default:80" json:"burst_threshold,omitempty" validate:"omitempty,min=1,max=100"`
	BurstTime      *int   `gorm:"default:8" json:"burst_time,omitempty" validate:"omitempty,min=1"`

	// Queue
	Priority    *int    `gorm:"default:8" json:"priority,omitempty" validate:"omitempty,min=1,max=8"`
	QueueType   *string `gorm:"size:20;default:default" json:"queue_type,omitempty"`
	SharedUsers *int    `gorm:"default:1" json:"shared_users,omitempty" validate:"omitempty,min=1"`
	QueueName   *string `gorm:"size:20" json:"queue_name,omitempty"`

	// Data cap
	QuotaGB *int `json:"quota_gb,omitempty"` // NULL = unlimited (in GB)

	// IP Pool / Address Pool (name in RouterOS)
	AddressPool *string `gorm:"size:100" json:"address_pool,omitempty"`

	// Billing
	BillingCycle BillingCycle `gorm:"size:20;not null;default:monthly" json:"billing_cycle"`

	// Pricing
	PriceMonthly      float64  `gorm:"type:decimal(12,2);not null" json:"price_monthly" validate:"required,min=0"`
	PriceInstallation float64  `gorm:"type:decimal(12,2);default:0" json:"price_installation" validate:"min=0"`
	TaxRate           *float64 `gorm:"type:decimal(5,4);default:0.11" json:"tax_rate,omitempty" validate:"omitempty,min=0,max=1"`

	// Visibility
	IsActive  *bool `gorm:"default:true" json:"is_active"`
	IsVisible *bool `gorm:"default:true" json:"is_visible"`
	SortOrder *int  `gorm:"default:0" json:"sort_order"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName specifies the table name
func (BandwidthProfile) TableName() string { return "bandwidth_profiles" }

// BandwidthProfileInput for creating/updating
type BandwidthProfileInput struct {
	ProfileCode       string   `json:"profile_code" validate:"required,max=50"`
	Name              string   `json:"name" validate:"required,max=100"`
	Description       *string  `json:"description"`
	ServiceType       string   `json:"service_type" validate:"required,oneof=pppoe hotspot static_ip vpn"`
	Category          string   `json:"category" validate:"required,oneof=residential business corporate promo"`
	PppProfileName    string   `json:"ppp_profile_name" validate:"required,max=100"`
	LocalAddress      *string  `json:"local_address" validate:"omitempty,max=45"`
	RemoteAddress     *string  `json:"remote_address" validate:"omitempty,max=45"`
	ParentQueue       *string  `json:"parent_queue" validate:"omitempty,max=100"`
	DNSServer         *string  `json:"dns_server" validate:"omitempty,max=100"`
	DownloadSpeed     int64    `json:"download_speed" validate:"required,min=1"`
	UploadSpeed       int64    `json:"upload_speed" validate:"required,min=1"`
	BurstDownload     *int64   `json:"burst_download" validate:"omitempty,min=1"`
	BurstUpload       *int64   `json:"burst_upload" validate:"omitempty,min=1"`
	BurstThreshold    *int     `json:"burst_threshold" validate:"omitempty,min=1,max=100"`
	BurstTime         *int     `json:"burst_time" validate:"omitempty,min=1"`
	Priority          *int     `json:"priority" validate:"omitempty,min=1,max=8"`
	QueueType         *string  `json:"queue_type" validate:"omitempty,max=20"`
	SharedUsers       *int     `json:"shared_users" validate:"omitempty,min=1"`
	QueueName         *string  `json:"queue_name" validate:"omitempty,max=20"`
	QuotaGB           *int     `json:"quota_gb" validate:"omitempty,min=0"`
	AddressPool       *string  `json:"address_pool" validate:"omitempty,max=100"`
	BillingCycle      string   `json:"billing_cycle" validate:"omitempty,oneof=daily weekly monthly yearly"`
	PriceMonthly      float64  `json:"price_monthly" validate:"required,min=0"`
	PriceInstallation float64  `json:"price_installation" validate:"min=0"`
	TaxRate           *float64 `json:"tax_rate" validate:"omitempty,min=0,max=1"`
	IsActive          *bool    `json:"is_active"`
	IsVisible         *bool    `json:"is_visible"`
	SortOrder         *int     `json:"sort_order"`
}

// BandwidthProfileFilter for querying
type BandwidthProfileFilter struct {
	ServiceType *ServiceType     `json:"service_type"`
	Category    *ProfileCategory `json:"category"`
	IsActive    *bool            `json:"is_active"`
	IsVisible   *bool            `json:"is_visible"`
	MinPrice    *float64         `json:"min_price"`
	MaxPrice    *float64         `json:"max_price"`
	Search      *string          `json:"search"`
}

// ToModel converts input to BandwidthProfile
func (i *BandwidthProfileInput) ToModel() *BandwidthProfile {
	bp := &BandwidthProfile{
		ProfileCode:       i.ProfileCode,
		Name:              i.Name,
		Description:       i.Description,
		PppProfileName:    i.PppProfileName,
		LocalAddress:      i.LocalAddress,
		RemoteAddress:     i.RemoteAddress,
		ParentQueue:       i.ParentQueue,
		DNSServer:         i.DNSServer,
		DownloadSpeed:     i.DownloadSpeed,
		UploadSpeed:       i.UploadSpeed,
		BurstDownload:     i.BurstDownload,
		BurstUpload:       i.BurstUpload,
		BurstThreshold:    i.BurstThreshold,
		BurstTime:         i.BurstTime,
		Priority:          i.Priority,
		QueueType:         i.QueueType,
		SharedUsers:       i.SharedUsers,
		QueueName:         i.QueueName,
		QuotaGB:           i.QuotaGB,
		AddressPool:       i.AddressPool,
		PriceMonthly:      i.PriceMonthly,
		PriceInstallation: i.PriceInstallation,
		TaxRate:           i.TaxRate,
		IsActive:          i.IsActive,
		IsVisible:         i.IsVisible,
		SortOrder:         i.SortOrder,
	}
	bp.ServiceType = ServiceType(i.ServiceType)
	bp.Category = ProfileCategory(i.Category)
	if i.BillingCycle != "" {
		bp.BillingCycle = BillingCycle(i.BillingCycle)
	} else {
		bp.BillingCycle = BillingCycleMonthly
	}
	return bp
}

// GetRateLimit returns rate limit string for MikroTik ("upload/download")
func (b *BandwidthProfile) GetRateLimit() string {
	return formatRateLimit(b.UploadSpeed, b.DownloadSpeed, b.BurstUpload, b.BurstDownload, b.BurstThreshold, b.BurstTime)
}

func formatRateLimit(upload, download int64, burstUpload, burstDownload *int64, burstThreshold, burstTime *int) string {
	r := formatSpeed(upload) + "/" + formatSpeed(download)
	if burstUpload != nil && burstDownload != nil {
		r += " " + formatSpeed(*burstUpload) + "/" + formatSpeed(*burstDownload)
		if burstThreshold != nil {
			t := 8
			if burstTime != nil {
				t = *burstTime
			}
			r += " " + formatInt(*burstThreshold) + "/" + formatInt(t)
		}
	}
	return r
}

func formatSpeed(kbps int64) string {
	if kbps >= 1000000 {
		return formatFloat(float64(kbps)/1000000) + "G"
	} else if kbps >= 1000 {
		return formatFloat(float64(kbps)/1000) + "M"
	}
	return fmt.Sprintf("%d", kbps) + "k"
}

func formatInt(n int) string     { return fmt.Sprintf("%d", n) }
func formatInt64(n int64) string { return fmt.Sprintf("%d", n) }
func formatFloat(f float64) string {
	s := fmt.Sprintf("%.1f", f)
	if s[len(s)-2:] == ".0" {
		return s[:len(s)-2]
	}
	return s
}

// keep formatInt64 used elsewhere
var _ = formatInt64
