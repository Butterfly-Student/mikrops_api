package postgres_outbound_adapter

import (
	"mikrops/internal/model"
	outbound_port "mikrops/internal/port/outbound"

	"gorm.io/gorm"
)

type permissionAdapter struct {
	db *gorm.DB
}

func NewPermissionAdapter(db *gorm.DB) outbound_port.PermissionDatabasePort {
	return &permissionAdapter{db: db}
}

func (a *permissionAdapter) FindByRoleID(roleID int) ([]model.Permission, error) {
	var permissions []model.Permission
	result := a.db.
		Joins("JOIN role_permissions ON role_permissions.permission_id = permissions.id").
		Where("role_permissions.role_id = ?", roleID).
		Find(&permissions)
	return permissions, result.Error
}

func (a *permissionAdapter) FindAll() ([]model.Permission, error) {
	var permissions []model.Permission
	result := a.db.Find(&permissions)
	return permissions, result.Error
}
