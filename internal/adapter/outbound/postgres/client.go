package postgres_outbound_adapter

import (
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"go-template/internal/model"
	outbound_port "go-template/internal/port/outbound"
)

const tableClient = "admin_users"

type clientAdapter struct {
	db *gorm.DB
}

func NewClientAdapter(
	db *gorm.DB,
) outbound_port.ClientDatabasePort {
	return &clientAdapter{
		db: db,
	}
}

// Upsert inserts or updates client records
func (adapter *clientAdapter) Upsert(datas []model.AdminUserInput) error {
	// Build the data structures for GORM
	users := make([]map[string]interface{}, len(datas))
	for i, data := range datas {
		users[i] = map[string]interface{}{
			"full_name": data.FullName,
			"email":     data.Email,
			"phone":     data.Phone,
			"role":      data.Role,
			"is_active": data.IsActive,
		}
	}

	// Use GORM's Clauses for ON CONFLICT handling
	return adapter.db.Table(tableClient).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "email"}},
			DoUpdates: clause.AssignmentColumns([]string{"full_name", "phone", "role", "is_active", "updated_at"}),
		}).
		Create(users).Error
}

// FindByFilter retrieves clients based on filter criteria
func (adapter *clientAdapter) FindByFilter(filter model.AdminUserFilter, lock bool) ([]model.AdminUser, error) {
	var users []model.AdminUser

	query := adapter.db.Table(tableClient)

	// Apply filters
	if len(filter.Emails) > 0 {
		query = query.Where("email IN ?", filter.Emails)
	}

	if len(filter.Roles) > 0 {
		query = query.Where("role IN ?", filter.Roles)
	}

	if filter.IsActive != nil {
		query = query.Where("is_active = ?", *filter.IsActive)
	}

	if filter.Search != nil && *filter.Search != "" {
		searchPattern := "%" + *filter.Search + "%"
		query = query.Where("full_name ILIKE ? OR email ILIKE ?", searchPattern, searchPattern)
	}

	// Add row locking if requested
	if lock {
		query = query.Clauses(clause.Locking{Strength: "UPDATE"})
	}

	// Execute query
	err := query.Find(&users).Error
	if err != nil {
		return nil, err
	}

	return users, nil
}

// DeleteByFilter deletes clients based on filter criteria
func (adapter *clientAdapter) DeleteByFilter(filter model.AdminUserFilter) error {
	query := adapter.db.Table(tableClient)

	// Apply filters
	if len(filter.Emails) > 0 {
		query = query.Where("email IN ?", filter.Emails)
	}

	if len(filter.Roles) > 0 {
		query = query.Where("role IN ?", filter.Roles)
	}

	if filter.IsActive != nil {
		query = query.Where("is_active = ?", *filter.IsActive)
	}

	// Execute delete
	return query.Delete(&model.AdminUser{}).Error
}

// IsExists checks if a user exists by email
func (adapter *clientAdapter) IsExists(email string) (bool, error) {
	var count int64

	err := adapter.db.Table(tableClient).
		Where("email = ?", email).
		Count(&count).Error

	if err != nil {
		return false, err
	}

	return count > 0, nil
}
