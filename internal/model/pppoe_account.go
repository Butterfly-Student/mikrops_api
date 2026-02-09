package model

import "time"

const (
	PppoeStatusActive   = "active"
	PppoeStatusIsolated = "isolated"
	PppoeStatusDisabled = "disabled"

	PppoeSyncStatusSynced  = "synced"
	PppoeSyncStatusPending = "pending"
	PppoeSyncStatusError   = "error"
)

type PppoeAccount struct {
	ID string `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	PppoeAccountInput
	Customer        *Customer        `json:"customer,omitempty" gorm:"foreignKey:CustomerID"`
	Nas             *Nas             `json:"nas,omitempty" gorm:"foreignKey:NasID"`
	InternetPackage *InternetPackage `json:"internet_package,omitempty" gorm:"foreignKey:PackageID"`
}

func (PppoeAccount) TableName() string {
	return "pppoe_accounts"
}

type PppoeAccountInput struct {
	TenantID          string     `json:"tenant_id" gorm:"column:tenant_id;type:uuid;not null;index"`
	CustomerID        string     `json:"customer_id" gorm:"column:customer_id;type:uuid;not null"`
	SubscriptionID    *string    `json:"subscription_id" gorm:"column:subscription_id;type:uuid"`
	NasID             string     `json:"nas_id" gorm:"column:nas_id;type:uuid;not null"`
	PackageID         string     `json:"package_id" gorm:"column:package_id;type:uuid;not null"`
	Username          string     `json:"username" gorm:"column:username;type:varchar(100);not null"`
	PasswordEncrypted string     `json:"-" gorm:"column:password_encrypted;type:text;not null"`
	Password          string     `json:"password,omitempty" gorm:"-"`
	ProfileName       string     `json:"profile_name" gorm:"column:profile_name;type:varchar(100)"`
	OriginalProfile   string     `json:"original_profile" gorm:"column:original_profile;type:varchar(100)"`
	Status            string     `json:"status" gorm:"column:status;type:varchar(50);default:'active'"`
	SyncStatus        string     `json:"sync_status" gorm:"column:sync_status;type:varchar(50);default:'pending'"`
	LastSyncAt        *time.Time `json:"last_sync_at" gorm:"column:last_sync_at"`
	CreatedAt         time.Time  `json:"created_at" gorm:"column:created_at;autoCreateTime"`
	UpdatedAt         time.Time  `json:"updated_at" gorm:"column:updated_at;autoUpdateTime"`
}

type PppoeAccountFilter struct {
	IDs          []string `json:"ids"`
	TenantIDs    []string `json:"tenant_ids"`
	CustomerIDs  []string `json:"customer_ids"`
	NasIDs       []string `json:"nas_ids"`
	PackageIDs   []string `json:"package_ids"`
	Usernames    []string `json:"usernames"`
	Statuses     []string `json:"statuses"`
	SyncStatuses []string `json:"sync_statuses"`
	WithCustomer bool     `json:"-"`
	WithNas      bool     `json:"-"`
	WithPackage  bool     `json:"-"`
}

func (f PppoeAccountFilter) IsEmpty() bool {
	return len(f.IDs) == 0 && len(f.TenantIDs) == 0 && len(f.CustomerIDs) == 0 &&
		len(f.NasIDs) == 0 && len(f.Usernames) == 0 && len(f.Statuses) == 0
}
