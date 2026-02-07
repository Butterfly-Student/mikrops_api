package model

import "time"

type Tenant struct {
	ID string `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	TenantInput
}

func (Tenant) TableName() string {
	return "tenants"
}

type TenantInput struct {
	Name                  string     `json:"name" gorm:"column:name;type:varchar(255);not null"`
	Slug                  string     `json:"slug" gorm:"column:slug;type:varchar(100);uniqueIndex;not null"`
	Email                 string     `json:"email" gorm:"column:email;type:varchar(255)"`
	Phone                 string     `json:"phone" gorm:"column:phone;type:varchar(50)"`
	Address               string     `json:"address" gorm:"column:address;type:text"`
	LogoURL               string     `json:"logo_url" gorm:"column:logo_url;type:varchar(500)"`
	MaxNas                int        `json:"max_nas" gorm:"column:max_nas;default:3"`
	SubscriptionPlan      string     `json:"subscription_plan" gorm:"column:subscription_plan;type:varchar(50);default:'basic'"`
	SubscriptionExpiresAt *time.Time `json:"subscription_expires_at" gorm:"column:subscription_expires_at"`
	IsActive              bool       `json:"is_active" gorm:"column:is_active;default:true"`
	CreatedAt             time.Time  `json:"created_at" gorm:"column:created_at;autoCreateTime"`
	UpdatedAt             time.Time  `json:"updated_at" gorm:"column:updated_at;autoUpdateTime"`
}

type TenantFilter struct {
	IDs    []string `json:"ids"`
	Slugs  []string `json:"slugs"`
	Emails []string `json:"emails"`
}

func TenantPrepare(v *TenantInput) {
	v.CreatedAt = time.Now()
	v.UpdatedAt = time.Now()
	if v.MaxNas == 0 {
		v.MaxNas = 3
	}
	if v.SubscriptionPlan == "" {
		v.SubscriptionPlan = "basic"
	}
}

func (f TenantFilter) IsEmpty() bool {
	return len(f.IDs) == 0 && len(f.Slugs) == 0 && len(f.Emails) == 0
}
