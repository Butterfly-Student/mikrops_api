package hotspot

import (
	"context"

	"go-template/internal/model"
	outbound_port "go-template/internal/port/outbound"
	pkghotspot "go-template/pkg/hotspot"
)

// HotspotDomain defines business operations for MikroTik Hotspot management.
// All data is stored directly in RouterOS — no external database required.
type HotspotDomain interface {
	// Profile Management
	CreateProfile(ctx context.Context, routerID string, profile *model.HotspotProfile) error
	GetProfile(ctx context.Context, routerID string, name string) (*model.HotspotProfile, error)
	GetAllProfiles(ctx context.Context, routerID string) ([]model.HotspotProfile, error)
	UpdateProfile(ctx context.Context, routerID string, name string, updates *model.HotspotProfileUpdate) error
	DeleteProfile(ctx context.Context, routerID string, name string) error

	// User Management
	CreateUser(ctx context.Context, routerID string, user *model.HotspotUser) error
	GetUser(ctx context.Context, routerID string, username string) (*model.HotspotUser, error)
	GetAllUsers(ctx context.Context, routerID string, filter *model.HotspotUserFilter) ([]model.HotspotUser, error)
	UpdateUser(ctx context.Context, routerID string, username string, updates *model.HotspotUserUpdate) error
	DeleteUser(ctx context.Context, routerID string, username string) error

	// Voucher Generation
	GenerateVouchers(ctx context.Context, routerID string, gen *model.HotspotVoucherGenerator) (*model.HotspotVoucherResult, error)
	GenerateUserPasswordMode(ctx context.Context, routerID string, gen *model.HotspotVoucherGenerator) (*model.HotspotVoucherResult, error)

	// Session Management
	GetActiveSessions(ctx context.Context, routerID string) ([]model.HotspotSession, error)
	GetSessionStats(ctx context.Context, routerID string) (*model.HotspotSessionStats, error)
	DisconnectUser(ctx context.Context, routerID string, username string) error

	// Sales
	RecordSale(ctx context.Context, routerID string, sale *model.HotspotSale) error
	GetAllSales(ctx context.Context, routerID string, filter *model.HotspotSaleFilter) ([]model.HotspotSale, error)
	GetTotalRevenue(ctx context.Context, routerID string, startDate, endDate string) (float64, error)

	// Expiry Schedulers
	CreateExpiryScheduler(ctx context.Context, routerID string, profileName string) error
	RemoveExpiryScheduler(ctx context.Context, routerID string, profileName string) error
}

type domain struct {
	dbPort      outbound_port.DatabasePort
	hotspotPort outbound_port.HotspotPort
}

func NewHotspotDomain(
	dbPort outbound_port.DatabasePort,
	hotspotPort outbound_port.HotspotPort,
) HotspotDomain {
	return &domain{
		dbPort:      dbPort,
		hotspotPort: hotspotPort,
	}
}

func (d *domain) getRouter(routerID string) (*model.MikrotikRouter, error) {
	return d.dbPort.Mikrotik().FindByID(routerID)
}

func (d *domain) getClient(ctx context.Context, routerID string) (pkghotspot.Client, error) {
	router, err := d.getRouter(routerID)
	if err != nil {
		return nil, err
	}
	return d.hotspotPort.GetHotspotClient(router)
}

// Profile Management

func (d *domain) CreateProfile(ctx context.Context, routerID string, profile *model.HotspotProfile) error {
	client, err := d.getClient(ctx, routerID)
	if err != nil {
		return err
	}
	return client.CreateProfile(ctx, profile)
}

func (d *domain) GetProfile(ctx context.Context, routerID string, name string) (*model.HotspotProfile, error) {
	client, err := d.getClient(ctx, routerID)
	if err != nil {
		return nil, err
	}
	return client.GetProfile(ctx, name)
}

func (d *domain) GetAllProfiles(ctx context.Context, routerID string) ([]model.HotspotProfile, error) {
	client, err := d.getClient(ctx, routerID)
	if err != nil {
		return nil, err
	}
	return client.GetAllProfiles(ctx)
}

func (d *domain) UpdateProfile(ctx context.Context, routerID string, name string, updates *model.HotspotProfileUpdate) error {
	client, err := d.getClient(ctx, routerID)
	if err != nil {
		return err
	}
	return client.UpdateProfile(ctx, name, updates)
}

func (d *domain) DeleteProfile(ctx context.Context, routerID string, name string) error {
	client, err := d.getClient(ctx, routerID)
	if err != nil {
		return err
	}
	return client.DeleteProfile(ctx, name)
}

// User Management

func (d *domain) CreateUser(ctx context.Context, routerID string, user *model.HotspotUser) error {
	client, err := d.getClient(ctx, routerID)
	if err != nil {
		return err
	}
	return client.CreateUser(ctx, user)
}

func (d *domain) GetUser(ctx context.Context, routerID string, username string) (*model.HotspotUser, error) {
	client, err := d.getClient(ctx, routerID)
	if err != nil {
		return nil, err
	}
	return client.GetUser(ctx, username)
}

func (d *domain) GetAllUsers(ctx context.Context, routerID string, filter *model.HotspotUserFilter) ([]model.HotspotUser, error) {
	client, err := d.getClient(ctx, routerID)
	if err != nil {
		return nil, err
	}
	return client.GetAllUsers(ctx, filter)
}

func (d *domain) UpdateUser(ctx context.Context, routerID string, username string, updates *model.HotspotUserUpdate) error {
	client, err := d.getClient(ctx, routerID)
	if err != nil {
		return err
	}
	return client.UpdateUser(ctx, username, updates)
}

func (d *domain) DeleteUser(ctx context.Context, routerID string, username string) error {
	client, err := d.getClient(ctx, routerID)
	if err != nil {
		return err
	}
	return client.DeleteUser(ctx, username)
}

// Voucher Generation

func (d *domain) GenerateVouchers(ctx context.Context, routerID string, gen *model.HotspotVoucherGenerator) (*model.HotspotVoucherResult, error) {
	client, err := d.getClient(ctx, routerID)
	if err != nil {
		return nil, err
	}
	return client.GenerateVouchers(ctx, gen)
}

func (d *domain) GenerateUserPasswordMode(ctx context.Context, routerID string, gen *model.HotspotVoucherGenerator) (*model.HotspotVoucherResult, error) {
	client, err := d.getClient(ctx, routerID)
	if err != nil {
		return nil, err
	}
	return client.GenerateUserPasswordMode(ctx, gen)
}

// Session Management

func (d *domain) GetActiveSessions(ctx context.Context, routerID string) ([]model.HotspotSession, error) {
	client, err := d.getClient(ctx, routerID)
	if err != nil {
		return nil, err
	}
	return client.GetActiveSessions(ctx)
}

func (d *domain) GetSessionStats(ctx context.Context, routerID string) (*model.HotspotSessionStats, error) {
	client, err := d.getClient(ctx, routerID)
	if err != nil {
		return nil, err
	}
	return client.GetSessionStats(ctx)
}

func (d *domain) DisconnectUser(ctx context.Context, routerID string, username string) error {
	client, err := d.getClient(ctx, routerID)
	if err != nil {
		return err
	}
	return client.DisconnectUser(ctx, username)
}

// Sales

func (d *domain) RecordSale(ctx context.Context, routerID string, sale *model.HotspotSale) error {
	client, err := d.getClient(ctx, routerID)
	if err != nil {
		return err
	}
	return client.RecordSale(ctx, sale)
}

func (d *domain) GetAllSales(ctx context.Context, routerID string, filter *model.HotspotSaleFilter) ([]model.HotspotSale, error) {
	client, err := d.getClient(ctx, routerID)
	if err != nil {
		return nil, err
	}
	return client.GetAllSales(ctx, filter)
}

func (d *domain) GetTotalRevenue(ctx context.Context, routerID string, startDate, endDate string) (float64, error) {
	client, err := d.getClient(ctx, routerID)
	if err != nil {
		return 0, err
	}
	return client.GetTotalRevenue(ctx, startDate, endDate)
}

// Expiry Schedulers

func (d *domain) CreateExpiryScheduler(ctx context.Context, routerID string, profileName string) error {
	client, err := d.getClient(ctx, routerID)
	if err != nil {
		return err
	}
	return client.CreateExpiryScheduler(ctx, profileName)
}

func (d *domain) RemoveExpiryScheduler(ctx context.Context, routerID string, profileName string) error {
	client, err := d.getClient(ctx, routerID)
	if err != nil {
		return err
	}
	return client.RemoveExpiryScheduler(ctx, profileName)
}
