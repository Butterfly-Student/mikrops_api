package postgres_outbound_adapter

import (
	"time"

	"gorm.io/gorm"

	"go-template/internal/model"
	outbound_port "go-template/internal/port/outbound"
)

type customerAdapter struct {
	db *gorm.DB
}

func NewCustomerAdapter(
	db *gorm.DB,
) outbound_port.CustomerDatabasePort {
	return &customerAdapter{
		db: db,
	}
}

func (adapter *customerAdapter) Create(customer *model.Customer) error {
	return adapter.db.Create(customer).Error
}

func (adapter *customerAdapter) FindByFilter(filter model.CustomerFilter) ([]model.Customer, error) {
	var customers []model.Customer

	query := adapter.db.Model(&model.Customer{})

	// Preload relationships
	query = query.Preload("Router").Preload("Profile").Preload("CreatedByUser").Preload("UpdatedByUser")

	// Apply filters
	if len(filter.IDs) > 0 {
		query = query.Where("id IN ?", filter.IDs)
	}

	if len(filter.CustomerCodes) > 0 {
		query = query.Where("customer_code IN ?", filter.CustomerCodes)
	}

	if len(filter.Statuses) > 0 {
		query = query.Where("status IN ?", filter.Statuses)
	}

	if len(filter.RouterIDs) > 0 {
		query = query.Where("router_id IN ?", filter.RouterIDs)
	}

	if len(filter.ProfileIDs) > 0 {
		query = query.Where("profile_id IN ?", filter.ProfileIDs)
	}

	if filter.ExpiredBefore != nil {
		query = query.Where("expiry_date < ?", filter.ExpiredBefore)
	}

	if filter.ExpiredAfter != nil {
		query = query.Where("expiry_date > ?", filter.ExpiredAfter)
	}

	if filter.Search != "" {
		searchPattern := "%" + filter.Search + "%"
		query = query.Where("full_name ILIKE ? OR email ILIKE ? OR phone ILIKE ? OR customer_code ILIKE ? OR address ILIKE ?",
			searchPattern, searchPattern, searchPattern, searchPattern, searchPattern)
	}

	// Exclude soft-deleted records
	query = query.Where("deleted_at IS NULL")

	// Order by created_at desc
	query = query.Order("created_at DESC")

	// Pagination
	if filter.Limit > 0 {
		query = query.Limit(filter.Limit)
		if filter.Page > 0 {
			offset := (filter.Page - 1) * filter.Limit
			query = query.Offset(offset)
		}
	}

	// Execute query
	err := query.Find(&customers).Error
	if err != nil {
		return nil, err
	}

	return customers, nil
}

func (adapter *customerAdapter) Update(customer *model.Customer) error {
	return adapter.db.Save(customer).Error
}

func (adapter *customerAdapter) Delete(id string) error {
	// Soft delete
	return adapter.db.Model(&model.Customer{}).
		Where("id = ?", id).
		Update("deleted_at", gorm.Expr("CURRENT_TIMESTAMP")).Error
}

func (adapter *customerAdapter) FindExpiredCustomers() ([]model.Customer, error) {
	var customers []model.Customer

	now := time.Now()

	query := adapter.db.Model(&model.Customer{}).
		Preload("Router").
		Preload("Profile").
		Where("expiry_date < ?", now).
		Where("status IN ?", []string{string(model.CustomerStatusActive), string(model.CustomerStatusSuspended)}).
		Where("deleted_at IS NULL")

	err := query.Find(&customers).Error
	if err != nil {
		return nil, err
	}

	return customers, nil
}

func (adapter *customerAdapter) FindExpiringCustomers(days int) ([]model.Customer, error) {
	var customers []model.Customer

	now := time.Now()
	futureDate := now.AddDate(0, 0, days)

	query := adapter.db.Model(&model.Customer{}).
		Preload("Router").
		Preload("Profile").
		Where("expiry_date BETWEEN ? AND ?", now, futureDate).
		Where("status = ?", string(model.CustomerStatusActive)).
		Where("deleted_at IS NULL")

	err := query.Find(&customers).Error
	if err != nil {
		return nil, err
	}

	return customers, nil
}
