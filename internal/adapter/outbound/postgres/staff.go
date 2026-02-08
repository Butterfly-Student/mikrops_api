package postgres_outbound_adapter

import (
	"mikrops/internal/model"
	outbound_port "mikrops/internal/port/outbound"

	"gorm.io/gorm"
)

type staffAdapter struct {
	db *gorm.DB
}

func NewStaffAdapter(db *gorm.DB) outbound_port.StaffDatabasePort {
	return &staffAdapter{db: db}
}

func (a *staffAdapter) Create(data model.StaffInput) (model.Staff, error) {
	staff := model.Staff{StaffInput: data}
	result := a.db.Create(&staff)
	return staff, result.Error
}

func (a *staffAdapter) FindByFilter(filter model.StaffFilter) ([]model.Staff, error) {
	var staffs []model.Staff
	query := a.db.Model(&model.Staff{})
	query = applyStaffFilter(query, filter)

	if filter.WithRole {
		query = query.Preload("Role")
	}
	if filter.WithTenant {
		query = query.Preload("Tenant")
	}

	result := query.Find(&staffs)
	return staffs, result.Error
}

func (a *staffAdapter) FindByID(id string) (model.Staff, error) {
	var staff model.Staff
	result := a.db.Preload("Role").Preload("Tenant").Where("id = ?", id).First(&staff)
	return staff, result.Error
}

func (a *staffAdapter) FindByEmail(email string) (model.Staff, error) {
	var staff model.Staff
	result := a.db.Preload("Role").Preload("Tenant").Where("email = ?", email).First(&staff)
	return staff, result.Error
}

func (a *staffAdapter) Update(id string, data model.StaffInput) error {
	return a.db.Model(&model.Staff{}).Where("id = ?", id).Updates(data).Error
}

func (a *staffAdapter) Delete(id string) error {
	return a.db.Where("id = ?", id).Delete(&model.Staff{}).Error
}

func applyStaffFilter(query *gorm.DB, filter model.StaffFilter) *gorm.DB {
	if len(filter.IDs) > 0 {
		query = query.Where("id IN ?", filter.IDs)
	}
	if len(filter.TenantIDs) > 0 {
		query = query.Where("tenant_id IN ?", filter.TenantIDs)
	}
	if len(filter.RoleIDs) > 0 {
		query = query.Where("role_id IN ?", filter.RoleIDs)
	}
	if len(filter.Emails) > 0 {
		query = query.Where("email IN ?", filter.Emails)
	}
	if filter.IsActive != nil {
		query = query.Where("is_active = ?", *filter.IsActive)
	}
	return query
}
