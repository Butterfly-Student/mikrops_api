package model

import (
	"time"

	"github.com/google/uuid"
)

type BandwidthProfile struct {
	ID                uuid.UUID  `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	ProfileCode       string     `json:"profile_code" gorm:"unique;not null" validate:"required"`
	Name              string     `json:"name" gorm:"not null" validate:"required"`
	Description       *string    `json:"description" gorm:"type:text"`
	Category          string     `json:"category" gorm:"not null" validate:"required,oneof=pppoe ip-static hotspot isolated"`
	PppProfileName    string     `json:"ppp_profile_name" gorm:"not null" validate:"required"`
	DownloadSpeed     int64      `json:"download_speed" gorm:"not null" validate:"required,min=1"`
	UploadSpeed       int64      `json:"upload_speed" gorm:"not null" validate:"required,min=1"`
	BurstDownload     *int64     `json:"burst_download" validate:"min=1"`
	BurstUpload       *int64     `json:"burst_upload" validate:"min=1"`
	BurstThreshold    *int       `json:"burst_threshold" validate:"min=0,max=100"`
	BurstTime         *int       `json:"burst_time" validate:"min=1"`
	Priority          *int       `json:"priority" gorm:"default:8" validate:"min=1,max=8"`
	QueueType         *string    `json:"queue_type" gorm:"default:'default'"`
	SharedUsers       *int       `json:"shared_users" gorm:"default:1" validate:"min=1"`
	QueueName         *string    `json:"queue_name" validate:"max=20"`
	PriceMonthly      float64    `json:"price_monthly" gorm:"not null" validate:"required,min=0"`
	PriceInstallation float64    `json:"price_installation" gorm:"default:0" validate:"min=0"`
	TaxRate           float64    `json:"tax_rate" gorm:"default:0.11" validate:"min=0,max=1"`
	IsActive          *bool      `json:"is_active" gorm:"default:true"`
	IsVisible         *bool      `json:"is_visible" gorm:"default:true"`
	SortOrder         *int       `json:"sort_order" gorm:"default:0"`
	CreatedAt         time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt         time.Time  `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt         *time.Time `json:"-" gorm:"index"`
}

type BandwidthProfileInput struct {
	ProfileCode       *string `json:"profile_code" validate:"required"`
	Name              string  `json:"name" validate:"required"`
	Description       *string `json:"description"`
	Category          string  `json:"category" validate:"required,oneof=pppoe ip-static hotspot isolated"`
	PppProfileName    string  `json:"ppp_profile_name" validate:"required"`
	DownloadSpeed     int64   `json:"download_speed" validate:"required,min=1"`
	UploadSpeed       int64   `json:"upload_speed" validate:"required,min=1"`
	BurstDownload     *int64  `json:"burst_download" validate:"min=1"`
	BurstUpload       *int64  `json:"burst_upload" validate:"min=1"`
	BurstThreshold    *int    `json:"burst_threshold" validate:"min=0,max=100"`
	BurstTime         *int    `json:"burst_time" validate:"min=1"`
	Priority          *int    `json:"priority" validate:"min=1,max=8"`
	QueueType         *string `json:"queue_type"`
	SharedUsers       *int    `json:"shared_users" validate:"min=1"`
	QueueName         *string `json:"queue_name" validate:"max=20"`
	PriceMonthly      float64 `json:"price_monthly" validate:"required,min=0"`
	PriceInstallation float64 `json:"price_installation" validate:"min=0"`
	TaxRate           float64 `json:"tax_rate" validate:"min=0,max=1"`
	IsActive          *bool   `json:"is_active"`
	IsVisible         *bool   `json:"is_visible"`
	SortOrder         *int    `json:"sort_order"`
}

type BandwidthProfileFilter struct {
	IDs          []uuid.UUID `json:"ids"`
	ProfileCodes []string    `json:"profile_codes"`
	Categories   []string    `json:"categories"`
	IsActive     *bool       `json:"is_active"`
	IsVisible    *bool       `json:"is_visible"`
}

func (f BandwidthProfileFilter) IsEmpty() bool {
	return len(f.IDs) == 0 && len(f.ProfileCodes) == 0 &&
		len(f.Categories) == 0 && f.IsActive == nil && f.IsVisible == nil
}

func BandwidthProfilePrepare(v *BandwidthProfile) {
	if v.Priority == nil {
		priority := 8
		v.Priority = &priority
	}
	if v.IsActive == nil {
		isActive := true
		v.IsActive = &isActive
	}
	if v.IsVisible == nil {
		isVisible := true
		v.IsVisible = &isVisible
	}
	if v.SortOrder == nil {
		sortOrder := 0
		v.SortOrder = &sortOrder
	}
	if v.SharedUsers == nil {
		sharedUsers := 1
		v.SharedUsers = &sharedUsers
	}
	if v.TaxRate == 0 {
		v.TaxRate = 0.11
	}
	if v.QueueType == nil {
		queueType := "default"
		v.QueueType = &queueType
	}
}
