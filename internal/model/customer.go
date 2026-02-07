package model

import "time"

type Customer struct {
	ID string `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	CustomerInput
}

func (Customer) TableName() string {
	return "customers"
}

type CustomerInput struct {
	TenantID       string    `json:"tenant_id" gorm:"column:tenant_id;type:uuid;not null;index"`
	FullName       string    `json:"full_name" gorm:"column:full_name;type:varchar(255);not null"`
	Email          string    `json:"email" gorm:"column:email;type:varchar(255)"`
	Phone          string    `json:"phone" gorm:"column:phone;type:varchar(50)"`
	Address        string    `json:"address" gorm:"column:address;type:text"`
	IdentityNumber string    `json:"identity_number" gorm:"column:identity_number;type:varchar(50)"`
	Username       string    `json:"username" gorm:"column:username;type:varchar(100)"`
	PasswordHash   string    `json:"-" gorm:"column:password_hash;type:varchar(255)"`
	Password       string    `json:"password,omitempty" gorm:"-"`
	PppoeUsername  string    `json:"pppoe_username" gorm:"column:pppoe_username;type:varchar(100)"`
	PppoePassword  string    `json:"pppoe_password" gorm:"column:pppoe_password;type:varchar(255)"`
	StaticIP       string    `json:"static_ip" gorm:"column:static_ip;type:varchar(45)"`
	NasID          *string   `json:"nas_id" gorm:"column:nas_id;type:uuid"`
	IsActive       bool      `json:"is_active" gorm:"column:is_active;default:true"`
	RegisteredAt   time.Time `json:"registered_at" gorm:"column:registered_at"`
	CreatedAt      time.Time `json:"created_at" gorm:"column:created_at;autoCreateTime"`
	UpdatedAt      time.Time `json:"updated_at" gorm:"column:updated_at;autoUpdateTime"`
}

type CustomerFilter struct {
	IDs            []string `json:"ids"`
	TenantIDs      []string `json:"tenant_ids"`
	Emails         []string `json:"emails"`
	Usernames      []string `json:"usernames"`
	PppoeUsernames []string `json:"pppoe_usernames"`
	NasIDs         []string `json:"nas_ids"`
}

func (f CustomerFilter) IsEmpty() bool {
	return len(f.IDs) == 0 && len(f.TenantIDs) == 0 && len(f.Emails) == 0 &&
		len(f.Usernames) == 0 && len(f.PppoeUsernames) == 0 && len(f.NasIDs) == 0
}
