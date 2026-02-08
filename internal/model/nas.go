package model

import "time"

type Nas struct {
	ID string `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	NasInput
}

func (Nas) TableName() string {
	return "nas"
}

type NasInput struct {
	TenantID          string     `json:"tenant_id" gorm:"column:tenant_id;type:uuid;not null;index"`
	Name              string     `json:"name" gorm:"column:name;type:varchar(255);not null"`
	Host              string     `json:"host" gorm:"column:host;type:varchar(255);not null"`
	ApiPort           int        `json:"api_port" gorm:"column:api_port;default:8728"`
	RestPort          int        `json:"rest_port" gorm:"column:rest_port;default:80"`
	Username          string     `json:"username" gorm:"column:username;type:varchar(255);not null"`
	PasswordEncrypted string     `json:"-" gorm:"column:password_encrypted;type:text"`
	Password          string     `json:"password,omitempty" gorm:"-"`
	UseSSL            bool       `json:"use_ssl" gorm:"column:use_ssl;default:false"`
	RouterOsVersion   string     `json:"router_os_version" gorm:"column:router_os_version;type:varchar(50)"`
	Identity          string     `json:"identity" gorm:"column:identity;type:varchar(255)"`
	IsActive          bool       `json:"is_active" gorm:"column:is_active;default:true"`
	LastSeenAt        *time.Time `json:"last_seen_at" gorm:"column:last_seen_at"`
	CreatedAt         time.Time  `json:"created_at" gorm:"column:created_at;autoCreateTime"`
	UpdatedAt         time.Time  `json:"updated_at" gorm:"column:updated_at;autoUpdateTime"`
}

type NasFilter struct {
	IDs        []string `json:"ids"`
	TenantIDs  []string `json:"tenant_ids"`
	Hosts      []string `json:"hosts"`
	IsActive   *bool    `json:"is_active,omitempty"`
	WithTenant bool     `json:"-"`
}

func (f NasFilter) IsEmpty() bool {
	return len(f.IDs) == 0 && len(f.TenantIDs) == 0 && len(f.Hosts) == 0
}
