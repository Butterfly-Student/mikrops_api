package postgres_outbound_adapter

import (
	"errors"

	"go-template/internal/model"
	outbound_port "go-template/internal/port/outbound"
	"gorm.io/gorm"
)

const tableBandwidthProfile = "bandwidth_profiles"

type BandwidthProfileAdapter struct {
	db *gorm.DB
}

func NewBandwidthProfileAdapter(db *gorm.DB) outbound_port.BandwidthProfileDatabasePort {
	return &BandwidthProfileAdapter{db: db}
}

func (a *BandwidthProfileAdapter) Create(profile *model.BandwidthProfile) error {
	model.BandwidthProfilePrepare(profile)
	return a.db.Create(profile).Error
}

func (a *BandwidthProfileAdapter) FindByID(id string) (*model.BandwidthProfile, error) {
	var profile model.BandwidthProfile
	err := a.db.Where("id = ?", id).First(&profile).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("bandwidth profile not found")
		}
		return nil, err
	}
	return &profile, nil
}

func (a *BandwidthProfileAdapter) FindAll() ([]model.BandwidthProfile, error) {
	var profiles []model.BandwidthProfile
	err := a.db.Order("sort_order ASC, created_at ASC").Find(&profiles).Error
	return profiles, err
}

func (a *BandwidthProfileAdapter) Find(filter model.BandwidthProfileFilter) ([]model.BandwidthProfile, error) {
	var profiles []model.BandwidthProfile
	query := a.db.Model(&model.BandwidthProfile{})

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

	err := query.Order("sort_order ASC, created_at ASC").Find(&profiles).Error
	return profiles, err
}

func (a *BandwidthProfileAdapter) FindByProfileCode(code string) (*model.BandwidthProfile, error) {
	var profile model.BandwidthProfile
	err := a.db.Where("profile_code = ?", code).First(&profile).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("bandwidth profile not found")
		}
		return nil, err
	}
	return &profile, nil
}

func (a *BandwidthProfileAdapter) FindByCategory(category string) ([]model.BandwidthProfile, error) {
	var profiles []model.BandwidthProfile
	err := a.db.Where("category = ?", category).Order("sort_order ASC").Find(&profiles).Error
	return profiles, err
}

func (a *BandwidthProfileAdapter) Update(profile *model.BandwidthProfile) error {
	return a.db.Save(profile).Error
}

func (a *BandwidthProfileAdapter) Delete(id string) error {
	return a.db.Delete(&model.BandwidthProfile{}, "id = ?", id).Error
}
