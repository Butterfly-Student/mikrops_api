package bandwidth_profile

import (
	"context"

	"github.com/google/uuid"
	"github.com/palantir/stacktrace"

	"go-template/internal/model"
	outbound_port "go-template/internal/port/outbound"
)

type BandwidthProfileDomain interface {
	Create(ctx context.Context, input model.BandwidthProfileInput) (*model.BandwidthProfile, error)
	CreateWithRouter(ctx context.Context, routerID string, input model.BandwidthProfileInput) (*model.BandwidthProfile, error)
	GetByID(ctx context.Context, id string) (*model.BandwidthProfile, error)
	GetByCode(ctx context.Context, code string) (*model.BandwidthProfile, error)
	List(ctx context.Context, filter *model.BandwidthProfileFilter) ([]model.BandwidthProfile, error)
	Update(ctx context.Context, id string, input model.BandwidthProfileInput) (*model.BandwidthProfile, error)
	UpdateWithRouter(ctx context.Context, routerID string, id string, input model.BandwidthProfileInput) (*model.BandwidthProfile, error)
	Delete(ctx context.Context, id string) error
	DeleteWithRouter(ctx context.Context, routerID string, id string) error
	SyncToMikrotik(ctx context.Context, id string, routerID string) error
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

// Create saves a BandwidthProfile to DB only (no MikroTik push).
// Use CreateWithRouter if you need MikroTik-first behavior.
func (d *domain) Create(ctx context.Context, input model.BandwidthProfileInput) (*model.BandwidthProfile, error) {
	profile := input.ToModel()
	profile.ID = uuid.New()
	if err := d.dbPort.BandwidthProfile().Create(ctx, profile); err != nil {
		return nil, stacktrace.Propagate(err, "failed to create bandwidth profile")
	}
	return profile, nil
}

// CreateWithRouter creates a BandwidthProfile — MikroTik first, then DB.
// If MikroTik fails, DB is never touched.
func (d *domain) CreateWithRouter(ctx context.Context, routerID string, input model.BandwidthProfileInput) (*model.BandwidthProfile, error) {
	// Get router
	router, err := d.dbPort.Mikrotik().FindByID(routerID)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to find mikrotik router")
	}

	// Build PPPoE profile for MikroTik
	profileModel := input.ToModel()
	pppProfile := &model.PppoeProfile{
		Name:      profileModel.PppProfileName,
		RateLimit: profileModel.GetRateLimit(),
		Comment:   profileModel.Name,
	}
	if profileModel.Description != nil && *profileModel.Description != "" {
		pppProfile.Comment = profileModel.Name + " - " + *profileModel.Description
	}

	// Check if profile already exists on MikroTik
	existingProfile, _ := d.mikrotikPort.GetProfile(router, profileModel.PppProfileName)
	if existingProfile != nil {
		// Update existing profile on MikroTik
		pppProfile.ID = existingProfile.ID
		if err := d.mikrotikPort.UpdateProfile(router, pppProfile); err != nil {
			return nil, stacktrace.Propagate(err, "failed to update ppp profile on mikrotik")
		}
	} else {
		// Create new profile on MikroTik
		if err := d.mikrotikPort.CreateProfile(router, pppProfile); err != nil {
			return nil, stacktrace.Propagate(err, "failed to create ppp profile on mikrotik")
		}
	}

	// MikroTik OK — now save to DB
	profileModel.ID = uuid.New()
	if err := d.dbPort.BandwidthProfile().Create(ctx, profileModel); err != nil {
		// DB failed — attempt to undo MikroTik (best effort)
		if existingProfile == nil {
			_ = d.mikrotikPort.DeleteProfile(router, pppProfile.ID)
		}
		return nil, stacktrace.Propagate(err, "mikrotik ok but failed to save to database")
	}

	return profileModel, nil
}

func (d *domain) GetByID(ctx context.Context, id string) (*model.BandwidthProfile, error) {
	profile, err := d.dbPort.BandwidthProfile().FindByID(ctx, id)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to get bandwidth profile by id")
	}

	return profile, nil
}

func (d *domain) GetByCode(ctx context.Context, code string) (*model.BandwidthProfile, error) {
	profile, err := d.dbPort.BandwidthProfile().FindByCode(ctx, code)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to get bandwidth profile by code")
	}

	return profile, nil
}

func (d *domain) List(ctx context.Context, filter *model.BandwidthProfileFilter) ([]model.BandwidthProfile, error) {
	profiles, err := d.dbPort.BandwidthProfile().FindAll(ctx, filter)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to list bandwidth profiles")
	}

	return profiles, nil
}

func (d *domain) Update(ctx context.Context, id string, input model.BandwidthProfileInput) (*model.BandwidthProfile, error) {
	// Get existing profile
	profile, err := d.dbPort.BandwidthProfile().FindByID(ctx, id)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to find bandwidth profile")
	}

	// Update fields
	profile.ProfileCode = input.ProfileCode
	profile.Name = input.Name
	profile.Description = input.Description
	profile.Category = model.ProfileCategory(input.Category)
	profile.PppProfileName = input.PppProfileName
	profile.DownloadSpeed = input.DownloadSpeed
	profile.UploadSpeed = input.UploadSpeed
	profile.BurstDownload = input.BurstDownload
	profile.BurstUpload = input.BurstUpload
	profile.BurstThreshold = input.BurstThreshold
	profile.BurstTime = input.BurstTime
	profile.Priority = input.Priority
	profile.QueueType = input.QueueType
	profile.SharedUsers = input.SharedUsers
	profile.QueueName = input.QueueName
	profile.PriceMonthly = input.PriceMonthly
	profile.PriceInstallation = input.PriceInstallation
	profile.TaxRate = input.TaxRate
	profile.IsActive = input.IsActive
	profile.IsVisible = input.IsVisible
	profile.SortOrder = input.SortOrder

	if err := d.dbPort.BandwidthProfile().Update(ctx, profile); err != nil {
		return nil, stacktrace.Propagate(err, "failed to update bandwidth profile")
	}

	return profile, nil
}

// UpdateWithRouter updates on MikroTik FIRST, then updates DB.
func (d *domain) UpdateWithRouter(ctx context.Context, routerID string, id string, input model.BandwidthProfileInput) (*model.BandwidthProfile, error) {
	// Get existing profile from DB
	profile, err := d.dbPort.BandwidthProfile().FindByID(ctx, id)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to find bandwidth profile")
	}

	// Get router
	router, err := d.dbPort.Mikrotik().FindByID(routerID)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to find mikrotik router")
	}

	// Keep old profile name for rollback and lookup
	oldPppProfileName := profile.PppProfileName

	// Apply new values to profile struct
	profile.ProfileCode = input.ProfileCode
	profile.Name = input.Name
	profile.Description = input.Description
	profile.Category = model.ProfileCategory(input.Category)
	profile.PppProfileName = input.PppProfileName
	profile.DownloadSpeed = input.DownloadSpeed
	profile.UploadSpeed = input.UploadSpeed
	profile.BurstDownload = input.BurstDownload
	profile.BurstUpload = input.BurstUpload
	profile.BurstThreshold = input.BurstThreshold
	profile.BurstTime = input.BurstTime
	profile.Priority = input.Priority
	profile.QueueType = input.QueueType
	profile.SharedUsers = input.SharedUsers
	profile.QueueName = input.QueueName
	profile.PriceMonthly = input.PriceMonthly
	profile.PriceInstallation = input.PriceInstallation
	profile.TaxRate = input.TaxRate
	profile.IsActive = input.IsActive
	profile.IsVisible = input.IsVisible
	profile.SortOrder = input.SortOrder

	newPppProfile := d.convertToMikrotikProfile(profile)

	// Find existing profile on MikroTik (by old name in case it was renamed)
	existingMikrotikProfile, _ := d.mikrotikPort.GetProfile(router, oldPppProfileName)
	if existingMikrotikProfile != nil {
		newPppProfile.ID = existingMikrotikProfile.ID
		if err := d.mikrotikPort.UpdateProfile(router, newPppProfile); err != nil {
			return nil, stacktrace.Propagate(err, "failed to update ppp profile on mikrotik")
		}
	} else {
		// Doesn't exist on MikroTik yet — create it
		if err := d.mikrotikPort.CreateProfile(router, newPppProfile); err != nil {
			return nil, stacktrace.Propagate(err, "failed to create ppp profile on mikrotik")
		}
	}

	// MikroTik OK — update DB
	if err := d.dbPort.BandwidthProfile().Update(ctx, profile); err != nil {
		// Best-effort rollback: restore old values on MikroTik
		if existingMikrotikProfile != nil {
			_ = d.mikrotikPort.UpdateProfile(router, existingMikrotikProfile)
		}
		return nil, stacktrace.Propagate(err, "mikrotik ok but failed to update database")
	}

	return profile, nil
}

func (d *domain) Delete(ctx context.Context, id string) error {
	if err := d.dbPort.BandwidthProfile().Delete(ctx, id); err != nil {
		return stacktrace.Propagate(err, "failed to delete bandwidth profile")
	}

	return nil
}

// DeleteWithRouter deletes from MikroTik FIRST, then deletes from DB.
func (d *domain) DeleteWithRouter(ctx context.Context, routerID string, id string) error {
	// Get profile from DB
	profile, err := d.dbPort.BandwidthProfile().FindByID(ctx, id)
	if err != nil {
		return stacktrace.Propagate(err, "failed to find bandwidth profile")
	}

	// Get router
	router, err := d.dbPort.Mikrotik().FindByID(routerID)
	if err != nil {
		return stacktrace.Propagate(err, "failed to find mikrotik router")
	}

	// Find and delete profile from MikroTik
	existingMikrotikProfile, _ := d.mikrotikPort.GetProfile(router, profile.PppProfileName)
	if existingMikrotikProfile != nil {
		if err := d.mikrotikPort.DeleteProfile(router, existingMikrotikProfile.ID); err != nil {
			return stacktrace.Propagate(err, "failed to delete ppp profile from mikrotik")
		}
	}

	// MikroTik OK — delete from DB
	if err := d.dbPort.BandwidthProfile().Delete(ctx, id); err != nil {
		// Best-effort rollback: recreate profile on MikroTik
		if existingMikrotikProfile != nil {
			_ = d.mikrotikPort.CreateProfile(router, existingMikrotikProfile)
		}
		return stacktrace.Propagate(err, "mikrotik ok but failed to delete from database")
	}

	return nil
}

func (d *domain) SyncToMikrotik(ctx context.Context, id string, routerID string) error {
	// Get bandwidth profile
	profile, err := d.dbPort.BandwidthProfile().FindByID(ctx, id)
	if err != nil {
		return stacktrace.Propagate(err, "failed to get bandwidth profile")
	}

	// Get router
	router, err := d.dbPort.Mikrotik().FindByID(routerID)
	if err != nil {
		return stacktrace.Propagate(err, "failed to get mikrotik router")
	}

	// Convert BandwidthProfile to PppoeProfile
	pppProfile := d.convertToMikrotikProfile(profile)

	// Check if profile exists on MikroTik
	existingProfile, err := d.mikrotikPort.GetProfile(router, profile.PppProfileName)
	if err != nil {
		// Profile doesn't exist, create it
		if err := d.mikrotikPort.CreateProfile(router, pppProfile); err != nil {
			return stacktrace.Propagate(err, "failed to create profile on mikrotik")
		}
	} else {
		// Profile exists, update it
		pppProfile.ID = existingProfile.ID
		if err := d.mikrotikPort.UpdateProfile(router, pppProfile); err != nil {
			return stacktrace.Propagate(err, "failed to update profile on mikrotik")
		}
	}

	return nil
}

func (d *domain) convertToMikrotikProfile(profile *model.BandwidthProfile) *model.PppoeProfile {
	comment := profile.Name
	if profile.Description != nil && *profile.Description != "" {
		comment = comment + " - " + *profile.Description
	}

	return &model.PppoeProfile{
		Name:      profile.PppProfileName,
		RateLimit: profile.GetRateLimit(),
		Comment:   comment,
	}
}
