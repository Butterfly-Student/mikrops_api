package model

import "time"

type InternetPackage struct {
	ID string `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	InternetPackageInput
}

func (InternetPackage) TableName() string {
	return "internet_packages"
}

type InternetPackageInput struct {
	TenantID      string    `json:"tenant_id" gorm:"column:tenant_id;type:uuid;not null;index"`
	Name          string    `json:"name" gorm:"column:name;type:varchar(255);not null"`
	Description   string    `json:"description" gorm:"column:description;type:text"`
	Type          string    `json:"type" gorm:"column:type;type:varchar(50);not null"`
	UploadRate    string    `json:"upload_rate" gorm:"column:upload_rate;type:varchar(50)"`
	DownloadRate  string    `json:"download_rate" gorm:"column:download_rate;type:varchar(50)"`
	UploadBurst   string    `json:"upload_burst" gorm:"column:upload_burst;type:varchar(50)"`
	DownloadBurst string    `json:"download_burst" gorm:"column:download_burst;type:varchar(50)"`
	Price         int64     `json:"price" gorm:"column:price;not null"`
	BillingCycle  string    `json:"billing_cycle" gorm:"column:billing_cycle;type:varchar(50);default:'monthly'"`
	ValidityDays  int       `json:"validity_days" gorm:"column:validity_days;default:30"`
	ProfileName   string    `json:"profile_name" gorm:"column:profile_name;type:varchar(100)"`
	IsActive      bool      `json:"is_active" gorm:"column:is_active;default:true"`
	CreatedAt     time.Time `json:"created_at" gorm:"column:created_at;autoCreateTime"`
	UpdatedAt     time.Time `json:"updated_at" gorm:"column:updated_at;autoUpdateTime"`
}

type InternetPackageFilter struct {
	IDs           []string `json:"ids"`
	TenantIDs     []string `json:"tenant_ids"`
	Types         []string `json:"types"`
	Names         []string `json:"names"`
	BillingCycles []string `json:"billing_cycles"`
	IsActive      *bool    `json:"is_active,omitempty"`
	WithTenant    bool     `json:"-"`
}

func (f InternetPackageFilter) IsEmpty() bool {
	return len(f.IDs) == 0 && len(f.TenantIDs) == 0 && len(f.Types) == 0 && len(f.Names) == 0
}
