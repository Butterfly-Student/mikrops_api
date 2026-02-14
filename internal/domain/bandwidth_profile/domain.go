package bandwidth_profile

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/palantir/stacktrace"

	"go-template/internal/model"
	outbound_port "go-template/internal/port/outbound"
)

type BandwidthProfileDomain interface {
	Create(ctx context.Context, input model.BandwidthProfileInput) (*model.BandwidthProfile, error)
	FindByID(ctx context.Context, id string) (*model.BandwidthProfile, error)
	FindByFilter(ctx context.Context, filter model.BandwidthProfileFilter) ([]model.BandwidthProfile, error)
	Update(ctx context.Context, id string, input model.BandwidthProfileInput) (*model.BandwidthProfile, error)
	Delete(ctx context.Context, id string) error
	GetIsolatedProfile(ctx context.Context) (*model.BandwidthProfile, error)
	SyncToMikrotik(ctx context.Context, profileID string, routerID string) error
}

type bandwidthProfileDomain struct {
	databasePort outbound_port.DatabasePort
	mikrotikPort outbound_port.MikrotikPort
}

func NewBandwidthProfileDomain(
	databasePort outbound_port.DatabasePort,
	mikrotikPort outbound_port.MikrotikPort,
) BandwidthProfileDomain {
	return &bandwidthProfileDomain{
		databasePort: databasePort,
		mikrotikPort: mikrotikPort,
	}
}

func (d *bandwidthProfileDomain) Create(ctx context.Context, input model.BandwidthProfileInput) (*model.BandwidthProfile, error) {
	// Validate input
	if input.ProfileCode == "" {
		return nil, stacktrace.NewError("profile code is required")
	}
	if input.Name == "" {
		return nil, stacktrace.NewError("profile name is required")
	}

	// Check if profile code already exists
	databaseBandwidthProfilePort := d.databasePort.BandwidthProfile()
	existing, err := databaseBandwidthProfilePort.FindByFilter(model.BandwidthProfileFilter{
		ProfileCodes: []string{input.ProfileCode},
	})
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to check existing profile")
	}
	if len(existing) > 0 {
		return nil, stacktrace.NewError("profile code already exists")
	}

	// Create profile
	profile := &model.BandwidthProfile{
		ID:                uuid.New(),
		ProfileCode:       input.ProfileCode,
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

	// Set default PPP profile name if not provided
	if profile.PppProfileName == "" {
		profile.PppProfileName = input.ProfileCode
	}

	err = databaseBandwidthProfilePort.Create(profile)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to create bandwidth profile")
	}

	return profile, nil
}

func (d *bandwidthProfileDomain) FindByID(ctx context.Context, id string) (*model.BandwidthProfile, error) {
	if id == "" {
		return nil, stacktrace.NewError("id is required")
	}

	profileID, err := uuid.Parse(id)
	if err != nil {
		return nil, stacktrace.Propagate(err, "invalid uuid format")
	}

	databaseBandwidthProfilePort := d.databasePort.BandwidthProfile()
	profiles, err := databaseBandwidthProfilePort.FindByFilter(model.BandwidthProfileFilter{
		IDs: []uuid.UUID{profileID},
	})
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to find bandwidth profile")
	}

	if len(profiles) == 0 {
		return nil, stacktrace.NewError("bandwidth profile not found")
	}

	return &profiles[0], nil
}

func (d *bandwidthProfileDomain) FindByFilter(ctx context.Context, filter model.BandwidthProfileFilter) ([]model.BandwidthProfile, error) {
	databaseBandwidthProfilePort := d.databasePort.BandwidthProfile()
	profiles, err := databaseBandwidthProfilePort.FindByFilter(filter)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to find bandwidth profiles")
	}

	return profiles, nil
}

func (d *bandwidthProfileDomain) Update(ctx context.Context, id string, input model.BandwidthProfileInput) (*model.BandwidthProfile, error) {
	// Find existing profile
	profile, err := d.FindByID(ctx, id)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to find bandwidth profile")
	}

	// Check if profile code is being changed and already exists
	if input.ProfileCode != profile.ProfileCode {
		databaseBandwidthProfilePort := d.databasePort.BandwidthProfile()
		existing, err := databaseBandwidthProfilePort.FindByFilter(model.BandwidthProfileFilter{
			ProfileCodes: []string{input.ProfileCode},
		})
		if err != nil {
			return nil, stacktrace.Propagate(err, "failed to check existing profile")
		}
		if len(existing) > 0 {
			return nil, stacktrace.NewError("profile code already exists")
		}
	}

	// Update profile fields
	profile.ProfileCode = input.ProfileCode
	profile.Name = input.Name
	profile.Description = input.Description
	profile.Category = input.Category
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

	if profile.PppProfileName == "" {
		profile.PppProfileName = input.ProfileCode
	}

	databaseBandwidthProfilePort := d.databasePort.BandwidthProfile()
	err = databaseBandwidthProfilePort.Update(profile)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to update bandwidth profile")
	}

	return profile, nil
}

func (d *bandwidthProfileDomain) Delete(ctx context.Context, id string) error {
	// Find existing profile
	profile, err := d.FindByID(ctx, id)
	if err != nil {
		return stacktrace.Propagate(err, "failed to find bandwidth profile")
	}

	// Check if profile is being used by customers
	databaseCustomerPort := d.databasePort.Customer()
	customers, err := databaseCustomerPort.FindByFilter(model.CustomerFilter{
		ProfileIDs: []uuid.UUID{profile.ID},
	})
	if err != nil {
		return stacktrace.Propagate(err, "failed to check customers using profile")
	}

	if len(customers) > 0 {
		return stacktrace.NewError("cannot delete profile: currently being used by %d customer(s)", len(customers))
	}

	databaseBandwidthProfilePort := d.databasePort.BandwidthProfile()
	err = databaseBandwidthProfilePort.Delete(profile.ID.String())
	if err != nil {
		return stacktrace.Propagate(err, "failed to delete bandwidth profile")
	}

	return nil
}

func (d *bandwidthProfileDomain) GetIsolatedProfile(ctx context.Context) (*model.BandwidthProfile, error) {
	databaseBandwidthProfilePort := d.databasePort.BandwidthProfile()
	profiles, err := databaseBandwidthProfilePort.FindByFilter(model.BandwidthProfileFilter{
		Categories: []model.BandwidthProfileCategory{model.BandwidthProfileCategoryIsolated},
	})
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to find isolated profile")
	}

	if len(profiles) == 0 {
		return nil, stacktrace.NewError("isolated profile not found")
	}

	return &profiles[0], nil
}

func (d *bandwidthProfileDomain) SyncToMikrotik(ctx context.Context, profileID string, routerID string) error {
	// Find profile
	profile, err := d.FindByID(ctx, profileID)
	if err != nil {
		return stacktrace.Propagate(err, "failed to find bandwidth profile")
	}

	// Find router
	routerUUID, err := uuid.Parse(routerID)
	if err != nil {
		return stacktrace.Propagate(err, "invalid router uuid format")
	}

	databaseMikrotikPort := d.databasePort.Mikrotik()
	router, err := databaseMikrotikPort.FindByID(routerUUID.String())
	if err != nil {
		return stacktrace.Propagate(err, "failed to find mikrotik router")
	}

	// Create PPP profile on MikroTik
	pppProfile := &model.PppoeProfile{
		Name:          profile.PppProfileName,
		LocalAddress:  "10.10.10.1",
		RemoteAddress: "pppoe-pool",
		RateLimit:     formatRateLimit(profile.UploadSpeed, profile.DownloadSpeed, profile.BurstUpload, profile.BurstDownload, profile.BurstTime, profile.Priority),
	}

	mikrotikPppoePort := d.mikrotikPort.Pppoe()
	err = mikrotikPppoePort.CreateProfile(router, pppProfile)
	if err != nil {
		return stacktrace.Propagate(err, "failed to create ppp profile on mikrotik")
	}

	return nil
}

// formatRateLimit formats the rate limit string for MikroTik
// Format: rx-rate[/tx-rate] [rx-burst-rate[/tx-burst-rate] [rx-burst-threshold[/tx-burst-threshold] [rx-burst-time[/tx-burst-time] [priority]]]]
func formatRateLimit(upload, download, burstUpload, burstDownload int64, burstTime, priority int) string {
	// Convert kbps to bps
	uploadBps := upload * 1024
	downloadBps := download * 1024
	burstUploadBps := burstUpload * 1024
	burstDownloadBps := burstDownload * 1024

	// Basic rate limit: upload/download
	rateLimit := ""
	rateLimit += intToString(uploadBps) + "/" + intToString(downloadBps)

	// Add burst rates if configured
	if burstUpload > 0 && burstDownload > 0 {
		rateLimit += " " + intToString(burstUploadBps) + "/" + intToString(burstDownloadBps)

		// Add burst threshold (80% of download speed by default)
		threshold := (downloadBps * 80) / 100
		rateLimit += " " + intToString(threshold) + "/" + intToString(threshold)

		// Add burst time
		rateLimit += " " + intToString(int64(burstTime)) + "/" + intToString(int64(burstTime))

		// Add priority
		rateLimit += " " + intToString(int64(priority))
	}

	return rateLimit
}

func intToString(i int64) string {
	return fmt.Sprintf("%d", i)
}
