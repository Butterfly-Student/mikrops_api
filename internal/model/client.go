package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// AdminUserRole represents roles available for admin users
type AdminUserRole string

const (
	AdminRoleSuperAdmin AdminUserRole = "superadmin"
	AdminRoleAdmin      AdminUserRole = "admin"
	AdminRoleCS         AdminUserRole = "cs" // Customer Service
	AdminRoleBilling    AdminUserRole = "billing"
	AdminRoleTechnician AdminUserRole = "technician"
	AdminRoleReadOnly   AdminUserRole = "readonly"
)

// AdminUser represents an ISP internal staff/admin account.
// Maps to the admin_users table.
type AdminUser struct {
	ID           uuid.UUID     `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	FullName     string        `gorm:"size:100;not null" json:"full_name" validate:"required,max=100"`
	Email        string        `gorm:"size:100;uniqueIndex;not null" json:"email" validate:"required,email,max=100"`
	Phone        *string       `gorm:"size:20" json:"phone,omitempty"`
	PasswordHash string        `gorm:"size:255;not null" json:"-"`
	Role         AdminUserRole `gorm:"size:20;not null;default:cs" json:"role" validate:"required,oneof=superadmin admin cs billing technician readonly"`
	IsActive     *bool         `gorm:"default:true" json:"is_active"`
	LastLogin    *time.Time    `json:"last_login,omitempty"`
	LastIP       *string       `gorm:"size:45" json:"last_ip,omitempty"`
	BearerKey    *string       `gorm:"size:255;uniqueIndex" json:"-"` // API key programmatic access

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName specifies the table name for AdminUser
func (AdminUser) TableName() string {
	return "admin_users"
}

// AdminUserInput for creating/updating an admin user
type AdminUserInput struct {
	FullName string  `json:"full_name" validate:"required,max=100"`
	Email    string  `json:"email" validate:"required,email,max=100"`
	Phone    *string `json:"phone" validate:"omitempty,max=20"`
	Password string  `json:"password" validate:"required,min=6"`
	Role     string  `json:"role" validate:"required,oneof=superadmin admin cs billing technician readonly"`
	IsActive *bool   `json:"is_active"`
}

// AdminUserFilter for querying admin users
type AdminUserFilter struct {
	Emails     []string        `json:"emails"`
	BearerKeys []string        `json:"bearer_keys"`
	Roles      []AdminUserRole `json:"roles"`
	IsActive   *bool           `json:"is_active"`
	Search     *string         `json:"search"` // name, email
}

func (f AdminUserFilter) IsEmpty() bool {
	return len(f.Emails) == 0 && len(f.BearerKeys) == 0 && len(f.Roles) == 0 && f.IsActive == nil && f.Search == nil
}

// ------------------------------------------------------------------
// Backward-compatibility aliases (used by existing adapter / domain layers).
// These will be removed once those layers are migrated to AdminUser.
// ------------------------------------------------------------------

// Client is a backward-compatibility alias for AdminUser.
// Deprecated: use AdminUser instead.
type Client = AdminUser

// ClientInput is a backward-compatibility alias for AdminUserInput.
// Deprecated: use AdminUserInput instead.
type ClientInput = AdminUserInput

// ClientFilter is a backward-compatibility alias for AdminUserFilter.
// Deprecated: use AdminUserFilter instead.
type ClientFilter = AdminUserFilter
