package model

import "time"

type PaymentMethod struct {
	ID string `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	PaymentMethodInput
}

func (PaymentMethod) TableName() string {
	return "payment_methods"
}

type PaymentMethodInput struct {
	TenantID      string    `json:"tenant_id" gorm:"column:tenant_id;type:uuid;not null;index"`
	Name          string    `json:"name" gorm:"column:name;type:varchar(255);not null"`
	Type          string    `json:"type" gorm:"column:type;type:varchar(50);not null"`
	AccountName   string    `json:"account_name" gorm:"column:account_name;type:varchar(255)"`
	AccountNumber string    `json:"account_number" gorm:"column:account_number;type:varchar(100)"`
	BankName      string    `json:"bank_name" gorm:"column:bank_name;type:varchar(255)"`
	Instructions  string    `json:"instructions" gorm:"column:instructions;type:text"`
	IsActive      bool      `json:"is_active" gorm:"column:is_active;default:true"`
	CreatedAt     time.Time `json:"created_at" gorm:"column:created_at;autoCreateTime"`
	UpdatedAt     time.Time `json:"updated_at" gorm:"column:updated_at;autoUpdateTime"`
}

type PaymentMethodFilter struct {
	IDs       []string `json:"ids"`
	TenantIDs []string `json:"tenant_ids"`
	Types     []string `json:"types"`
}

func (f PaymentMethodFilter) IsEmpty() bool {
	return len(f.IDs) == 0 && len(f.TenantIDs) == 0 && len(f.Types) == 0
}
