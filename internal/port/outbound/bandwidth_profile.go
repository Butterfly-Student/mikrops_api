package outbound_port

import (
	"context"

	"go-template/internal/model"
)

//go:generate mockgen -source=bandwidth_profile.go -destination=./../../../tests/mocks/port/mock_bandwidth_profile_database_port.go

type BandwidthProfileDatabasePort interface {
	Create(ctx context.Context, profile *model.BandwidthProfile) error
	FindByID(ctx context.Context, id string) (*model.BandwidthProfile, error)
	FindByCode(ctx context.Context, code string) (*model.BandwidthProfile, error)
	FindAll(ctx context.Context, filter *model.BandwidthProfileFilter) ([]model.BandwidthProfile, error)
	Update(ctx context.Context, profile *model.BandwidthProfile) error
	Delete(ctx context.Context, id string) error
}
