package model

import "time"

type TenantSetting struct {
	ID string `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	TenantSettingInput
}

func (TenantSetting) TableName() string {
	return "tenant_settings"
}

type TenantSettingInput struct {
	TenantID                 string    `json:"tenant_id" gorm:"column:tenant_id;type:uuid;not null;uniqueIndex"`
	CutoffDay                int       `json:"cutoff_day" gorm:"column:cutoff_day;default:1"`
	GracePeriod              int       `json:"grace_period" gorm:"column:grace_period;default:3"`
	IsolirProfileName        string    `json:"isolir_profile_name" gorm:"column:isolir_profile_name;type:varchar(100);default:'ISOLIR_MIKROPS'"`
	WaGatewayAPI             string    `json:"wa_gateway_api" gorm:"column:wa_gateway_api;type:text"`
	AutoApproveRegistration  bool      `json:"auto_approve_registration" gorm:"column:auto_approve_registration;default:false"`
	DefaultPppPasswordType   string    `json:"default_ppp_password_type" gorm:"column:default_ppp_password_type;type:varchar(20);default:'random'"`
	DefaultPppPasswordLength int       `json:"default_ppp_password_length" gorm:"column:default_ppp_password_length;default:8"`
	CreatedAt                time.Time `json:"created_at" gorm:"column:created_at;autoCreateTime"`
	UpdatedAt                time.Time `json:"updated_at" gorm:"column:updated_at;autoUpdateTime"`
}

type TenantSettingFilter struct {
	IDs       []string `json:"ids"`
	TenantIDs []string `json:"tenant_ids"`
}

func (f TenantSettingFilter) IsEmpty() bool {
	return len(f.IDs) == 0 && len(f.TenantIDs) == 0
}
