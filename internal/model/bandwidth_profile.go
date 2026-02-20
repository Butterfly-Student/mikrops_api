package model

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// BandwidthProfile represents a bandwidth package/profile for customers
type BandwidthProfile struct {
	ID          uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	ProfileCode string         `gorm:"size:50;uniqueIndex;not null" json:"profile_code" validate:"required"`
	Name        string         `gorm:"size:100;not null" json:"name" validate:"required"`
	Description *string        `gorm:"type:text" json:"description"`
	Category    string         `gorm:"size:20;not null" json:"category" validate:"required,oneof=residential business corporate promo"`

	// Mikrotik PPP Profile Name
	PppProfileName string `gorm:"size:100;not null" json:"ppp_profile_name" validate:"required"`

	// Speed Configuration (in kbps)
	DownloadSpeed int64 `gorm:"not null" json:"download_speed" validate:"required,min=1"`
	UploadSpeed   int64 `gorm:"not null" json:"upload_speed" validate:"required,min=1"`

	// Burst Configuration
	BurstDownload   *int64 `json:"burst_download"`
	BurstUpload     *int64 `json:"burst_upload"`
	BurstThreshold  *int   `gorm:"default:80" json:"burst_threshold" validate:"omitempty,min=1,max=100"`
	BurstTime       *int   `gorm:"default:8" json:"burst_time" validate:"omitempty,min=1"`

	// Queue Configuration
	Priority    *int    `gorm:"default:8" json:"priority" validate:"omitempty,min=1,max=8"`
	QueueType   *string `gorm:"size:20;default:default" json:"queue_type"`
	SharedUsers *int    `gorm:"default:1" json:"shared_users" validate:"omitempty,min=1"`
	QueueName   *string `gorm:"size:20" json:"queue_name"` // for IP static

	// Pricing
	PriceMonthly     float64  `gorm:"type:decimal(12,2);not null" json:"price_monthly" validate:"required,min=0"`
	PriceInstallation float64 `gorm:"type:decimal(12,2);default:0" json:"price_installation" validate:"min=0"`
	TaxRate          *float64 `gorm:"type:decimal(5,4);default:0.11" json:"tax_rate" validate:"omitempty,min=0,max=1"` // 11% PPN

	// Status
	IsActive  *bool `gorm:"default:true" json:"is_active"`
	IsVisible *bool `gorm:"default:true" json:"is_visible"` // visible for new customers
	SortOrder *int  `gorm:"default:0" json:"sort_order"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName specifies the table name for BandwidthProfile
func (BandwidthProfile) TableName() string {
	return "bandwidth_profiles"
}

// BandwidthProfileInput for creating/updating bandwidth profile
type BandwidthProfileInput struct {
	ProfileCode       string   `json:"profile_code" validate:"required,max=50"`
	Name              string   `json:"name" validate:"required,max=100"`
	Description       *string  `json:"description"`
	Category          string   `json:"category" validate:"required,oneof=residential business corporate promo"`
	PppProfileName    string   `json:"ppp_profile_name" validate:"required,max=100"`
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
	PriceMonthly      float64  `json:"price_monthly" validate:"required,min=0"`
	PriceInstallation float64  `json:"price_installation" validate:"min=0"`
	TaxRate           *float64 `json:"tax_rate" validate:"omitempty,min=0,max=1"`
	IsActive          *bool    `json:"is_active"`
	IsVisible         *bool    `json:"is_visible"`
	SortOrder         *int     `json:"sort_order"`
}

// BandwidthProfileFilter for filtering bandwidth profiles
type BandwidthProfileFilter struct {
	Category  *string `json:"category"`
	IsActive  *bool   `json:"is_active"`
	IsVisible *bool   `json:"is_visible"`
	MinPrice  *float64 `json:"min_price"`
	MaxPrice  *float64 `json:"max_price"`
	Search    *string `json:"search"` // search by name, code, or description
}

// ToModel converts input to model
func (i *BandwidthProfileInput) ToModel() *BandwidthProfile {
	return &BandwidthProfile{
		ProfileCode:       i.ProfileCode,
		Name:              i.Name,
		Description:       i.Description,
		Category:          i.Category,
		PppProfileName:    i.PppProfileName,
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
		PriceMonthly:      i.PriceMonthly,
		PriceInstallation: i.PriceInstallation,
		TaxRate:           i.TaxRate,
		IsActive:          i.IsActive,
		IsVisible:         i.IsVisible,
		SortOrder:         i.SortOrder,
	}
}

// GetRateLimit returns rate limit string for Mikrotik (format: "upload/download")
func (b *BandwidthProfile) GetRateLimit() string {
	return formatRateLimit(b.UploadSpeed, b.DownloadSpeed, b.BurstUpload, b.BurstDownload, b.BurstThreshold, b.BurstTime)
}

// formatRateLimit formats rate limit for Mikrotik
// Format: "upload[/download] [burst-upload[/burst-download] [burst-threshold[/burst-time [priority]]]]"
func formatRateLimit(upload, download int64, burstUpload, burstDownload *int64, burstThreshold, burstTime *int) string {
	rateLimit := ""

	// Basic rate limit (required)
	rateLimit += formatSpeed(upload) + "/" + formatSpeed(download)

	// Add burst if configured
	if burstUpload != nil && burstDownload != nil {
		rateLimit += " " + formatSpeed(*burstUpload) + "/" + formatSpeed(*burstDownload)

		// Add burst threshold and time if configured
		if burstThreshold != nil {
			threshold := *burstThreshold
			time := 8 // default
			if burstTime != nil {
				time = *burstTime
			}
			rateLimit += " " + formatInt(threshold) + "/" + formatInt(time)
		}
	}

	return rateLimit
}

// formatSpeed formats speed from kbps to Mikrotik format (k, M, G)
func formatSpeed(kbps int64) string {
	if kbps >= 1000000 {
		return formatFloat(float64(kbps)/1000000) + "G"
	} else if kbps >= 1000 {
		return formatFloat(float64(kbps)/1000) + "M"
	}
	return formatInt64(kbps) + "k"
}

// Helper functions for formatting
func formatInt64(n int64) string {
	return fmt.Sprintf("%d", n)
}

func formatInt(n int) string {
	return fmt.Sprintf("%d", n)
}

func formatFloat(f float64) string {
	// Remove trailing zeros
	s := fmt.Sprintf("%.1f", f)
	if s[len(s)-2:] == ".0" {
		return s[:len(s)-2]
	}
	return s
}
