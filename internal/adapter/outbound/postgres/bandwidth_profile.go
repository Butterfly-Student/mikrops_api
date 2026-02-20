package postgres_outbound_adapter

import (
	"context"

	"github.com/google/uuid"
	"github.com/palantir/stacktrace"
	"gorm.io/gorm"

	"go-template/internal/model"
	outbound_port "go-template/internal/port/outbound"
)

const tableBandwidthProfile = "bandwidth_profiles"

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

func (a *bandwidthProfileAdapter) Create(ctx context.Context, profile *model.BandwidthProfile) error {
	if err := a.db.WithContext(ctx).Create(profile).Error; err != nil {
		return stacktrace.Propagate(err, "failed to create bandwidth profile in database")
	}
	return nil
}

func (a *bandwidthProfileAdapter) FindByID(ctx context.Context, id string) (*model.BandwidthProfile, error) {
	var profile model.BandwidthProfile

	profileID, err := uuid.Parse(id)
	if err != nil {
		return nil, stacktrace.Propagate(err, "invalid profile id format")
	}

	if err := a.db.WithContext(ctx).Where("id = ?", profileID).First(&profile).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, stacktrace.NewError("bandwidth profile not found")
		}
		return nil, stacktrace.Propagate(err, "failed to find bandwidth profile by id")
	}

	return &profile, nil
}

func (a *bandwidthProfileAdapter) FindByCode(ctx context.Context, code string) (*model.BandwidthProfile, error) {
	var profile model.BandwidthProfile

	if err := a.db.WithContext(ctx).Where("profile_code = ?", code).First(&profile).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, stacktrace.NewError("bandwidth profile not found")
		}
		return nil, stacktrace.Propagate(err, "failed to find bandwidth profile by code")
	}

	return &profile, nil
}

func (a *bandwidthProfileAdapter) FindAll(ctx context.Context, filter *model.BandwidthProfileFilter) ([]model.BandwidthProfile, error) {
	var profiles []model.BandwidthProfile

	query := a.db.WithContext(ctx).Model(&model.BandwidthProfile{})

	// Apply filters if provided
	if filter != nil {
		if filter.Category != nil {
			query = query.Where("category = ?", *filter.Category)
		}

		if filter.IsActive != nil {
			query = query.Where("is_active = ?", *filter.IsActive)
		}

		if filter.IsVisible != nil {
			query = query.Where("is_visible = ?", *filter.IsVisible)
		}

		if filter.MinPrice != nil {
			query = query.Where("price_monthly >= ?", *filter.MinPrice)
		}

		if filter.MaxPrice != nil {
			query = query.Where("price_monthly <= ?", *filter.MaxPrice)
		}

		if filter.Search != nil && *filter.Search != "" {
			searchPattern := "%" + *filter.Search + "%"
			query = query.Where(
				"name ILIKE ? OR profile_code ILIKE ? OR description ILIKE ?",
				searchPattern, searchPattern, searchPattern,
			)
		}
	}

	// Order by sort_order and name
	query = query.Order("sort_order ASC, name ASC")

	if err := query.Find(&profiles).Error; err != nil {
		return nil, stacktrace.Propagate(err, "failed to find bandwidth profiles")
	}

	return profiles, nil
}

func (a *bandwidthProfileAdapter) Update(ctx context.Context, profile *model.BandwidthProfile) error {
	if err := a.db.WithContext(ctx).Save(profile).Error; err != nil {
		return stacktrace.Propagate(err, "failed to update bandwidth profile in database")
	}
	return nil
}

func (a *bandwidthProfileAdapter) Delete(ctx context.Context, id string) error {
	profileID, err := uuid.Parse(id)
	if err != nil {
		return stacktrace.Propagate(err, "invalid profile id format")
	}

	// Soft delete using GORM
	if err := a.db.WithContext(ctx).Delete(&model.BandwidthProfile{}, profileID).Error; err != nil {
		return stacktrace.Propagate(err, "failed to delete bandwidth profile from database")
	}

	return nil
}
