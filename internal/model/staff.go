package model

import "time"

type Staff struct {
	ID string `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	StaffInput
	Role *Role `json:"role,omitempty" gorm:"foreignKey:RoleID"`
}

func (Staff) TableName() string {
	return "staffs"
}

type StaffInput struct {
	TenantID     string     `json:"tenant_id" gorm:"column:tenant_id;type:uuid;not null;index"`
	RoleID       int        `json:"role_id" gorm:"column:role_id;not null"`
	Email        string     `json:"email" gorm:"column:email;type:varchar(255);not null"`
	PasswordHash string     `json:"-" gorm:"column:password_hash;type:varchar(255);not null"`
	Password     string     `json:"password,omitempty" gorm:"-"`
	FullName     string     `json:"full_name" gorm:"column:full_name;type:varchar(255);not null"`
	Phone        string     `json:"phone" gorm:"column:phone;type:varchar(50)"`
	IsActive     bool       `json:"is_active" gorm:"column:is_active;default:true"`
	LastLoginAt  *time.Time `json:"last_login_at" gorm:"column:last_login_at"`
	CreatedAt    time.Time  `json:"created_at" gorm:"column:created_at;autoCreateTime"`
	UpdatedAt    time.Time  `json:"updated_at" gorm:"column:updated_at;autoUpdateTime"`
}

type StaffFilter struct {
	IDs       []string `json:"ids"`
	TenantIDs []string `json:"tenant_ids"`
	Emails    []string `json:"emails"`
	RoleIDs   []int    `json:"role_ids"`
}

func (f StaffFilter) IsEmpty() bool {
	return len(f.IDs) == 0 && len(f.TenantIDs) == 0 && len(f.Emails) == 0 && len(f.RoleIDs) == 0
}
