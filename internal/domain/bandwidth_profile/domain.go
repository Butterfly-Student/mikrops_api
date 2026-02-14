package bandwidth_profile

import (
	"context"
	"errors"
	"fmt"

	"go-template/internal/model"
	outbound_port "go-template/internal/port/outbound"
	"go-template/utils/log"
)

type BandwidthProfileDomain interface {
	CreateProfile(ctx context.Context, input model.BandwidthProfileInput) (*model.BandwidthProfile, error)
	GetProfile(ctx context.Context, id string) (*model.BandwidthProfile, error)
	ListProfiles(ctx context.Context, filter model.BandwidthProfileFilter) ([]model.BandwidthProfile, error)
	UpdateProfile(ctx context.Context, id string, input model.BandwidthProfileInput) (*model.BandwidthProfile, error)
	DeleteProfile(ctx context.Context, id string) error
	SyncToMikrotik(ctx context.Context, profile *model.BandwidthProfile) error
	GetIsolatedProfile(ctx context.Context) (*model.BandwidthProfile, error)
	GetProfileByCode(ctx context.Context, code string) (*model.BandwidthProfile, error)
}

type domain struct {
	dbPort       outbound_port.DatabasePort
	mikrotikPort outbound_port.MikrotikPort
}

func NewBandwidthProfileDomain(
	dbPort outbound_port.DatabasePort,
	mikrotikPort outbound_port.MikrotikPort,
) BandwidthProfileDomain {
	return &domain{
		dbPort:       dbPort,
		mikrotikPort: mikrotikPort,
	}
}

func (d *domain) CreateProfile(ctx context.Context, input model.BandwidthProfileInput) (*model.BandwidthProfile, error) {
	if input.ProfileCode == nil || *input.ProfileCode == "" {
		return nil, errors.New("profile_code is required")
	}

	var result interface{}
	var err error
	result, err = d.dbPort.DoInTransaction(func(repo outbound_port.DatabasePort) (interface{}, error) {
		profile := &model.BandwidthProfile{
			ProfileCode:       *input.ProfileCode,
			Name:              input.Name,
			Description:       input.Description,
			Category:          input.Category,
			PppProfileName:    input.PppProfileName,
			DownloadSpeed:     input.DownloadSpeed,
			UploadSpeed:       input.UploadSpeed,
			BurstDownload:     input.BurstDownload,
			BurstUpload:       input.BurstUpload,
			BurstThreshold:    input.BurstThreshold,
			BurstTime:         input.BurstTime,
			Priority:          input.Priority,
			QueueType:         input.QueueType,
			SharedUsers:       input.SharedUsers,
			QueueName:         input.QueueName,
			PriceMonthly:      input.PriceMonthly,
			PriceInstallation: input.PriceInstallation,
			TaxRate:           input.TaxRate,
			IsActive:          input.IsActive,
			IsVisible:         input.IsVisible,
			SortOrder:         input.SortOrder,
		}

		model.BandwidthProfilePrepare(profile)

		err := repo.BandwidthProfile().Create(profile)
		if err != nil {
			return nil, err
		}

		if input.Category == "pppoe" || input.Category == "isolated" {
			mikrotikProfile := &model.PppoeProfile{
				Name:      profile.PppProfileName,
				RateLimit: fmt.Sprintf("%dK/%dK", profile.UploadSpeed, profile.DownloadSpeed),
			}

			routers, err := repo.Mikrotik().FindAll()
			if err == nil {
				for _, router := range routers {
					if router.IsActive != nil && *router.IsActive {
						err := d.mikrotikPort.CreateProfile(&router, mikrotikProfile)
						if err != nil {
							log.WithContext(ctx).Warn(fmt.Sprintf("Failed to sync profile to router - profile: %s, router: %s, error: %v",
								profile.Name, router.Name, err))
						}
					}
				}
			}
		}

		return profile, nil
	})

	if err != nil {
		return nil, err
	}

	return result.(*model.BandwidthProfile), nil
}

func (d *domain) GetProfile(ctx context.Context, id string) (*model.BandwidthProfile, error) {
	return d.dbPort.BandwidthProfile().FindByID(id)
}

func (d *domain) ListProfiles(ctx context.Context, filter model.BandwidthProfileFilter) ([]model.BandwidthProfile, error) {
	if filter.IsEmpty() {
		return d.dbPort.BandwidthProfile().FindAll()
	}
	return d.dbPort.BandwidthProfile().Find(filter)
}

func (d *domain) UpdateProfile(ctx context.Context, id string, input model.BandwidthProfileInput) (*model.BandwidthProfile, error) {
	var result interface{}
	var err error
	result, err = d.dbPort.DoInTransaction(func(repo outbound_port.DatabasePort) (interface{}, error) {
		profile, err := repo.BandwidthProfile().FindByID(id)
		if err != nil {
			return nil, errors.New("profile not found")
		}

		if input.Name != "" {
			profile.Name = input.Name
		}
		if input.Description != nil {
			profile.Description = input.Description
		}
		if input.DownloadSpeed != 0 {
			profile.DownloadSpeed = input.DownloadSpeed
		}
		if input.UploadSpeed != 0 {
			profile.UploadSpeed = input.UploadSpeed
		}
		if input.PriceMonthly != 0 {
			profile.PriceMonthly = input.PriceMonthly
		}
		if input.TaxRate != 0 {
			profile.TaxRate = input.TaxRate
		}
		if input.IsActive != nil {
			profile.IsActive = input.IsActive
		}
		if input.IsVisible != nil {
			profile.IsVisible = input.IsVisible
		}
		if input.Priority != nil {
			profile.Priority = input.Priority
		}
		if input.BurstDownload != nil {
			profile.BurstDownload = input.BurstDownload
		}
		if input.BurstUpload != nil {
			profile.BurstUpload = input.BurstUpload
		}

		err = repo.BandwidthProfile().Update(profile)
		if err != nil {
			return nil, err
		}

		if profile.Category == "pppoe" || profile.Category == "isolated" {
			mikrotikProfile := &model.PppoeProfile{
				ID:        "",
				Name:      profile.PppProfileName,
				RateLimit: fmt.Sprintf("%dK/%dK", profile.UploadSpeed, profile.DownloadSpeed),
			}

			routers, err := repo.Mikrotik().FindAll()
			if err == nil {
				for _, router := range routers {
					if router.IsActive != nil && *router.IsActive {
						err := d.mikrotikPort.UpdateProfile(&router, mikrotikProfile)
						if err != nil {
							log.WithContext(ctx).Warn(fmt.Sprintf("Failed to sync profile update to router - profile: %s, router: %s, error: %v",
								profile.Name, router.Name, err))
						}
					}
				}
			}
		}

		return profile, nil
	})

	if err != nil {
		return nil, err
	}

	return result.(*model.BandwidthProfile), nil
}

func (d *domain) DeleteProfile(ctx context.Context, id string) error {
	_, err := d.dbPort.BandwidthProfile().FindByID(id)
	if err != nil {
		return errors.New("profile not found")
	}

	customers, err := d.dbPort.Customer().FindByProfileID(id)
	if err == nil && len(customers) > 0 {
		return errors.New("cannot delete profile that is in use by customers")
	}

	return d.dbPort.BandwidthProfile().Delete(id)
}

func (d *domain) SyncToMikrotik(ctx context.Context, profile *model.BandwidthProfile) error {
	if profile.Category != "pppoe" && profile.Category != "isolated" {
		return nil
	}

	mikrotikProfile := &model.PppoeProfile{
		Name:      profile.PppProfileName,
		RateLimit: fmt.Sprintf("%dK/%dK", profile.UploadSpeed, profile.DownloadSpeed),
	}

	routers, err := d.dbPort.Mikrotik().FindAll()
	if err != nil {
		return err
	}

	for _, router := range routers {
		if router.IsActive != nil && *router.IsActive {
			existingProfile, err := d.mikrotikPort.GetProfile(&router, profile.PppProfileName)
			if err == nil && existingProfile != nil {
				err = d.mikrotikPort.UpdateProfile(&router, mikrotikProfile)
			} else {
				err = d.mikrotikPort.CreateProfile(&router, mikrotikProfile)
			}

			if err != nil {
				return fmt.Errorf("failed to sync profile %s to router %s: %w",
					profile.Name, router.Name, err)
			}
		}
	}

	return nil
}

func (d *domain) GetIsolatedProfile(ctx context.Context) (*model.BandwidthProfile, error) {
	profiles, err := d.dbPort.BandwidthProfile().FindByCategory("isolated")
	if err != nil {
		return nil, err
	}

	if len(profiles) == 0 {
		return nil, errors.New("isolated profile not found. Please create a profile with category='isolated'")
	}

	if len(profiles) == 0 {
		return nil, errors.New("isolated profile not found. Please create a profile with category='isolated'")
	}

	return &profiles[0], nil
}

func (d *domain) GetProfileByCode(ctx context.Context, code string) (*model.BandwidthProfile, error) {
	return d.dbPort.BandwidthProfile().FindByProfileCode(code)
}
