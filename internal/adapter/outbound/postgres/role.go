package postgres_outbound_adapter

import (
	"mikrops/internal/model"
	outbound_port "mikrops/internal/port/outbound"

	"gorm.io/gorm"
)

type roleAdapter struct {
	db *gorm.DB
}

func NewRoleAdapter(db *gorm.DB) outbound_port.RoleDatabasePort {
	return &roleAdapter{db: db}
}

func (a *roleAdapter) FindAll() ([]model.Role, error) {
	var roles []model.Role
	result := a.db.Find(&roles)
	return roles, result.Error
}

func (a *roleAdapter) FindByID(id int) (model.Role, error) {
	var role model.Role
	result := a.db.Where("id = ?", id).First(&role)
	return role, result.Error
}

func (a *roleAdapter) FindByName(name string) (model.Role, error) {
	var role model.Role
	result := a.db.Where("name = ?", name).First(&role)
	return role, result.Error
}
