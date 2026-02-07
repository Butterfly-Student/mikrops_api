package model

import "time"

const (
	RoleAdmin      = "admin"
	RoleTechnician = "technician"
	RoleSales      = "sales"
	RoleFinance    = "finance"
)

type Role struct {
	ID int `json:"id" gorm:"primaryKey;autoIncrement"`
	RoleInput
}

func (Role) TableName() string {
	return "roles"
}

type RoleInput struct {
	Name        string    `json:"name" gorm:"column:name;type:varchar(100);uniqueIndex;not null"`
	Description string    `json:"description" gorm:"column:description;type:text"`
	CreatedAt   time.Time `json:"created_at" gorm:"column:created_at;autoCreateTime"`
	UpdatedAt   time.Time `json:"updated_at" gorm:"column:updated_at;autoUpdateTime"`
}

type RoleFilter struct {
	IDs   []int    `json:"ids"`
	Names []string `json:"names"`
}

func (f RoleFilter) IsEmpty() bool {
	return len(f.IDs) == 0 && len(f.Names) == 0
}
