package hotspot

import (
	"context"

	routeros "github.com/go-routeros/routeros/v3"
)

// Client defines the interface for MikroTik Hotspot management operations.
// This interface follows the Mikhmon v3 API patterns for managing
// hotspot users, profiles, vouchers, sessions, and sales records.
type Client interface {
	// User operations
	UserService

	// Profile operations
	ProfileService

	// Session operations
	SessionService

	// Voucher operations
	VoucherService

	// Sales operations
	SalesService

	// Scheduler operations
	SchedulerService

	// Client configuration
	GetRouterID() uint
	GetConfig() *Config
	SetConfig(*Config)

	// Close closes the client connection
	Close() error
}

// UserService provides hotspot user management operations
type UserService interface {
	CreateUser(ctx context.Context, user *User) error
	GetUser(ctx context.Context, username string) (*User, error)
	GetAllUsers(ctx context.Context, filter *UserFilter) ([]User, error)
	GetUsersByProfile(ctx context.Context, profile string) ([]User, error)
	GetUsersByComment(ctx context.Context, comment string) ([]User, error)
	UpdateUser(ctx context.Context, username string, updates *UserUpdate) error
	DeleteUser(ctx context.Context, username string) error
	DisableUser(ctx context.Context, username string) error
	EnableUser(ctx context.Context, username string) error
	RemoveExpiredUsers(ctx context.Context, profile string) (int, error)
	RemoveUnusedVouchers(ctx context.Context, profile string) (int, error)
	BatchCreateUsers(ctx context.Context, users []User) (*VoucherResult, error)
	BatchRemoveUsers(ctx context.Context, usernames []string) (int, error)
}

// ProfileService provides hotspot profile management operations
type ProfileService interface {
	CreateProfile(ctx context.Context, profile *Profile) error
	GetProfile(ctx context.Context, name string) (*Profile, error)
	GetAllProfiles(ctx context.Context) ([]Profile, error)
	UpdateProfile(ctx context.Context, name string, updates *ProfileUpdate) error
	DeleteProfile(ctx context.Context, name string) error
	SyncProfileToRouter(ctx context.Context, profile *Profile) error
	GetProfileSettings(ctx context.Context, name string) (*Profile, error)
}

// SessionService provides active hotspot session management operations
type SessionService interface {
	GetActiveSessions(ctx context.Context) ([]Session, error)
	GetSessionsByServer(ctx context.Context, server string) ([]Session, error)
	GetSessionByUsername(ctx context.Context, username string) (*Session, error)
	DisconnectUser(ctx context.Context, username string) error
	GetSessionStats(ctx context.Context) (*SessionStats, error)
}

// VoucherService provides voucher generation operations
type VoucherService interface {
	GenerateVouchers(ctx context.Context, gen *VoucherGenerator) (*VoucherResult, error)
	GenerateUserPasswordMode(ctx context.Context, gen *VoucherGenerator) (*VoucherResult, error)
}

// SalesService provides sales recording operations
type SalesService interface {
	RecordSale(ctx context.Context, sale *Sale) error
	GetAllSales(ctx context.Context, filter *SaleFilter) ([]Sale, error)
	GetSalesByDateRange(ctx context.Context, startDate, endDate string) ([]Sale, error)
	GetSalesByPrefix(ctx context.Context, prefix string) ([]Sale, error)
	GetTotalRevenue(ctx context.Context, startDate, endDate string) (float64, error)
	DeleteSale(ctx context.Context, scriptID string) error
}

// SchedulerService provides scheduler management operations
type SchedulerService interface {
	CreateExpiryScheduler(ctx context.Context, profileName string) error
	RemoveExpiryScheduler(ctx context.Context, profileName string) error
	GetAllSchedulers(ctx context.Context) ([]Scheduler, error)
	GetSchedulerByName(ctx context.Context, name string) (*Scheduler, error)
}

// RouterOSClient defines the interface for RouterOS command execution
type RouterOSClient interface {
	Run(args ...string) (*routeros.Reply, error)
	Close() error
}
