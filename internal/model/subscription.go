package model

import "time"

const (
	SubscriptionStatusActive    = "active"
	SubscriptionStatusSuspended = "suspended"
	SubscriptionStatusExpired   = "expired"
	SubscriptionStatusCancelled = "cancelled"
	SubscriptionStatusVacation  = "vacation"
)

type Subscription struct {
	ID string `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	SubscriptionInput
	Customer        *Customer        `json:"customer,omitempty" gorm:"foreignKey:CustomerID"`
	InternetPackage *InternetPackage `json:"internet_package,omitempty" gorm:"foreignKey:PackageID"`
}

func (Subscription) TableName() string {
	return "subscriptions"
}

type SubscriptionInput struct {
	TenantID           string     `json:"tenant_id" gorm:"column:tenant_id;type:uuid;not null;index"`
	CustomerID         string     `json:"customer_id" gorm:"column:customer_id;type:uuid;not null"`
	PackageID          string     `json:"package_id" gorm:"column:package_id;type:uuid;not null"`
	NasID              string     `json:"nas_id" gorm:"column:nas_id;type:uuid;not null"`
	Status             string     `json:"status" gorm:"column:status;type:varchar(50);default:'active'"`
	StartDate          time.Time  `json:"start_date" gorm:"column:start_date;not null"`
	EndDate            time.Time  `json:"end_date" gorm:"column:end_date;not null"`
	AutoRenew          bool       `json:"auto_renew" gorm:"column:auto_renew;default:true"`
	VacationStart      *time.Time `json:"vacation_start" gorm:"column:vacation_start"`
	VacationEnd        *time.Time `json:"vacation_end" gorm:"column:vacation_end"`
	MikrotikQueueName  string     `json:"mikrotik_queue_name" gorm:"column:mikrotik_queue_name;type:varchar(255)"`
	MikrotikSecretName string     `json:"mikrotik_secret_name" gorm:"column:mikrotik_secret_name;type:varchar(255)"`
	CreatedAt          time.Time  `json:"created_at" gorm:"column:created_at;autoCreateTime"`
	UpdatedAt          time.Time  `json:"updated_at" gorm:"column:updated_at;autoUpdateTime"`
}

type SubscriptionFilter struct {
	IDs          []string `json:"ids"`
	TenantIDs    []string `json:"tenant_ids"`
	CustomerIDs  []string `json:"customer_ids"`
	PackageIDs   []string `json:"package_ids"`
	NasIDs       []string `json:"nas_ids"`
	Statuses     []string `json:"statuses"`
	AutoRenew    *bool    `json:"auto_renew,omitempty"`
	WithTenant   bool     `json:"-"`
	WithCustomer bool     `json:"-"`
	WithPackage  bool     `json:"-"`
	WithNas      bool     `json:"-"`
}

func (f SubscriptionFilter) IsEmpty() bool {
	return len(f.IDs) == 0 && len(f.TenantIDs) == 0 && len(f.CustomerIDs) == 0 &&
		len(f.PackageIDs) == 0 && len(f.Statuses) == 0
}
