package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type MikrotikRouter struct {
	ID                uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Name              string         `gorm:"not null" json:"name" validate:"required"`
	Address           string         `gorm:"not null" json:"address" validate:"required"`
	ApiPort           *int           `gorm:"default:8728" json:"api_port"`
	RestPort          *int           `gorm:"default:80" json:"rest_port"`
	Username          string         `gorm:"not null" json:"username" validate:"required"`
	Password          string         `gorm:"not null" json:"password" validate:"required"`
	PasswordEncrypted *string        `gorm:"type:text" json:"password_encrypted"`
	UseSSL            *bool          `gorm:"default:false" json:"use_ssl"`
	RouterOSVersion   *string        `gorm:"size:50" json:"router_os_version"`
	Identity          *string        `gorm:"size:255" json:"identity"`
	IsActive          *bool          `gorm:"default:true" json:"is_active"`
	LastSeenAt        *time.Time     `json:"last_seen_at"`
	CreatedAt         time.Time      `json:"created_at"`
	UpdatedAt         time.Time      `json:"updated_at"`
	DeletedAt         gorm.DeletedAt `gorm:"index" json:"-"`
}

func (m *MikrotikRouter) IsRouterActive() bool {
	return m.IsActive != nil && *m.IsActive
}
