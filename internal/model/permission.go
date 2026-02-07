package model

import "time"

const (
	ResourceTenant       = "tenant"
	ResourceStaff        = "staff"
	ResourceNas          = "nas"
	ResourcePackage      = "internet_package"
	ResourceCustomer     = "customer"
	ResourceSubscription = "subscription"
	ResourcePaymentMethod = "payment_method"
	ResourceInvoice      = "invoice"
	ResourcePayment      = "payment"
	ResourceReport       = "report"

	ActionCreate = "create"
	ActionRead   = "read"
	ActionUpdate = "update"
	ActionDelete = "delete"
	ActionManage = "manage"
)

type Permission struct {
	ID int `json:"id" gorm:"primaryKey;autoIncrement"`
	PermissionInput
}

func (Permission) TableName() string {
	return "permissions"
}

type PermissionInput struct {
	Name        string    `json:"name" gorm:"column:name;type:varchar(100);uniqueIndex;not null"`
	Resource    string    `json:"resource" gorm:"column:resource;type:varchar(100);not null"`
	Action      string    `json:"action" gorm:"column:action;type:varchar(50);not null"`
	Description string    `json:"description" gorm:"column:description;type:text"`
	CreatedAt   time.Time `json:"created_at" gorm:"column:created_at;autoCreateTime"`
	UpdatedAt   time.Time `json:"updated_at" gorm:"column:updated_at;autoUpdateTime"`
}

type PermissionFilter struct {
	IDs       []int    `json:"ids"`
	Resources []string `json:"resources"`
	Actions   []string `json:"actions"`
}

func (f PermissionFilter) IsEmpty() bool {
	return len(f.IDs) == 0 && len(f.Resources) == 0 && len(f.Actions) == 0
}

type RolePermission struct {
	RoleID       int `json:"role_id" gorm:"primaryKey;column:role_id"`
	PermissionID int `json:"permission_id" gorm:"primaryKey;column:permission_id"`
}

func (RolePermission) TableName() string {
	return "role_permissions"
}
