package model

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// BandwidthProfileCategory represents the category of bandwidth profile
type BandwidthProfileCategory string

const (
	BandwidthProfileCategoryPPPoE     BandwidthProfileCategory = "pppoe"
	BandwidthProfileCategoryIPStatic  BandwidthProfileCategory = "ip-static"
	BandwidthProfileCategoryHotspot   BandwidthProfileCategory = "hotspot"
	BandwidthProfileCategoryIsolated  BandwidthProfileCategory = "isolated"
)

type BandwidthProfile struct {
	ID              uuid.UUID                `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	ProfileCode     string                   `json:"profile_code" gorm:"uniqueIndex;not null"`
	Name            string                   `json:"name" gorm:"not null"`
	Description     string                   `json:"description"`
	Category        BandwidthProfileCategory `json:"category" gorm:"type:bandwidth_profile_category;not null;default:'pppoe'"`
	PppProfileName  string                   `json:"ppp_profile_name"`
	DownloadSpeed   int64                    `json:"download_speed" gorm:"not null;default:0"`
	UploadSpeed     int64                    `json:"upload_speed" gorm:"not null;default:0"`
	BurstDownload   int64                    `json:"burst_download" gorm:"default:0"`
	BurstUpload     int64                    `json:"burst_upload" gorm:"default:0"`
	BurstThreshold  int                      `json:"burst_threshold" gorm:"default:80"`
	BurstTime       int                      `json:"burst_time" gorm:"default:60"`
	Priority        int                      `json:"priority" gorm:"default:8"`
	QueueType       string                   `json:"queue_type" gorm:"default:'default'"`
	SharedUsers     int                      `json:"shared_users" gorm:"default:1"`
	QueueName       string                   `json:"queue_name"`
	PriceMonthly    decimal.Decimal          `json:"price_monthly" gorm:"type:decimal(12,2);default:0"`
	PriceInstallation decimal.Decimal        `json:"price_installation" gorm:"type:decimal(12,2);default:0"`
	TaxRate         decimal.Decimal          `json:"tax_rate" gorm:"type:decimal(5,4);default:0.1100"`
	IsActive        bool                     `json:"is_active" gorm:"default:true"`
	IsVisible       bool                     `json:"is_visible" gorm:"default:true"`
	SortOrder       int                      `json:"sort_order" gorm:"default:0"`
	CreatedAt       time.Time                `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt       time.Time                `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt       *time.Time               `json:"deleted_at,omitempty" gorm:"index"`
}

type BandwidthProfileInput struct {
	ProfileCode       string                   `json:"profile_code" binding:"required"`
	Name              string                   `json:"name" binding:"required"`
	Description       string                   `json:"description"`
	Category          BandwidthProfileCategory `json:"category" binding:"required,oneof=pppoe ip-static hotspot isolated"`
	PppProfileName    string                   `json:"ppp_profile_name"`
	DownloadSpeed     int64                    `json:"download_speed" binding:"required,min=0"`
	UploadSpeed       int64                    `json:"upload_speed" binding:"required,min=0"`
	BurstDownload     int64                    `json:"burst_download" binding:"min=0"`
	BurstUpload       int64                    `json:"burst_upload" binding:"min=0"`
	BurstThreshold    int                      `json:"burst_threshold" binding:"min=0,max=100"`
	BurstTime         int                      `json:"burst_time" binding:"min=0"`
	Priority          int                      `json:"priority" binding:"min=1,max=8"`
	QueueType         string                   `json:"queue_type"`
	SharedUsers       int                      `json:"shared_users" binding:"min=1"`
	QueueName         string                   `json:"queue_name"`
	PriceMonthly      decimal.Decimal          `json:"price_monthly"`
	PriceInstallation decimal.Decimal          `json:"price_installation"`
	TaxRate           decimal.Decimal          `json:"tax_rate"`
	IsActive          bool                     `json:"is_active"`
	IsVisible         bool                     `json:"is_visible"`
	SortOrder         int                      `json:"sort_order"`
}

type BandwidthProfileFilter struct {
	IDs          []uuid.UUID                `json:"ids"`
	ProfileCodes []string                   `json:"profile_codes"`
	Categories   []BandwidthProfileCategory `json:"categories"`
	IsActive     *bool                      `json:"is_active"`
	IsVisible    *bool                      `json:"is_visible"`
	Search       string                     `json:"search"`
	Page         int                        `json:"page"`
	Limit        int                        `json:"limit"`
}

func (f BandwidthProfileFilter) IsEmpty() bool {
	return len(f.IDs) == 0 && len(f.ProfileCodes) == 0 && len(f.Categories) == 0 && f.IsActive == nil && f.IsVisible == nil && f.Search == ""
}

// TableName specifies the table name for GORM
func (BandwidthProfile) TableName() string {
	return "bandwidth_profiles"
}

// FormatSpeed formats speed in kbps to human readable format
func (bp *BandwidthProfile) FormatDownloadSpeed() string {
	if bp.DownloadSpeed >= 1024*1024 {
		return decimal.NewFromInt(bp.DownloadSpeed).Div(decimal.NewFromInt(1024 * 1024)).StringFixed(2) + " Gbps"
	} else if bp.DownloadSpeed >= 1024 {
		return decimal.NewFromInt(bp.DownloadSpeed).Div(decimal.NewFromInt(1024)).StringFixed(2) + " Mbps"
	}
	return decimal.NewFromInt(bp.DownloadSpeed).String() + " Kbps"
}

func (bp *BandwidthProfile) FormatUploadSpeed() string {
	if bp.UploadSpeed >= 1024*1024 {
		return decimal.NewFromInt(bp.UploadSpeed).Div(decimal.NewFromInt(1024 * 1024)).StringFixed(2) + " Gbps"
	} else if bp.UploadSpeed >= 1024 {
		return decimal.NewFromInt(bp.UploadSpeed).Div(decimal.NewFromInt(1024)).StringFixed(2) + " Mbps"
	}
	return decimal.NewFromInt(bp.UploadSpeed).String() + " Kbps"
}
