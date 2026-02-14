package postgres_outbound_adapter

import (
	"gorm.io/gorm"

	"go-template/internal/model"
	outbound_port "go-template/internal/port/outbound"
)

type bandwidthProfileAdapter struct {
	db *gorm.DB
}

func NewBandwidthProfileAdapter(
	db *gorm.DB,
) outbound_port.BandwidthProfileDatabasePort {
	return &bandwidthProfileAdapter{
		db: db,
	}
}

func (adapter *bandwidthProfileAdapter) Create(profile *model.BandwidthProfile) error {
	return adapter.db.Create(profile).Error
}

func (adapter *bandwidthProfileAdapter) FindByFilter(filter model.BandwidthProfileFilter) ([]model.BandwidthProfile, error) {
	var profiles []model.BandwidthProfile

	query := adapter.db.Model(&model.BandwidthProfile{})

	// Apply filters
	if len(filter.IDs) > 0 {
		query = query.Where("id IN ?", filter.IDs)
	}

	if len(filter.ProfileCodes) > 0 {
		query = query.Where("profile_code IN ?", filter.ProfileCodes)
	}

	if len(filter.Categories) > 0 {
		query = query.Where("category IN ?", filter.Categories)
	}

	if filter.IsActive != nil {
		query = query.Where("is_active = ?", *filter.IsActive)
	}

	if filter.IsVisible != nil {
		query = query.Where("is_visible = ?", *filter.IsVisible)
	}

	if filter.Search != "" {
		searchPattern := "%" + filter.Search + "%"
		query = query.Where("name ILIKE ? OR description ILIKE ? OR profile_code ILIKE ?",
			searchPattern, searchPattern, searchPattern)
	}

	// Exclude soft-deleted records
	query = query.Where("deleted_at IS NULL")

	// Order by sort_order and name
	query = query.Order("sort_order ASC, name ASC")

	// Pagination
	if filter.Limit > 0 {
		query = query.Limit(filter.Limit)
		if filter.Page > 0 {
			offset := (filter.Page - 1) * filter.Limit
			query = query.Offset(offset)
		}
	}

	// Execute query
	err := query.Find(&profiles).Error
	if err != nil {
		return nil, err
	}

	return profiles, nil
}

func (adapter *bandwidthProfileAdapter) Update(profile *model.BandwidthProfile) error {
	return adapter.db.Save(profile).Error
}

func (adapter *bandwidthProfileAdapter) Delete(id string) error {
	// Soft delete
	return adapter.db.Model(&model.BandwidthProfile{}).
		Where("id = ?", id).
		Update("deleted_at", gorm.Expr("CURRENT_TIMESTAMP")).Error
}
