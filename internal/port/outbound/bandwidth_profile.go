package outbound_port

import "go-template/internal/model"

//go:generate mockgen -source=bandwidth_profile.go -destination=./../../../tests/mocks/port/mock_bandwidth_profile.go
type BandwidthProfileDatabasePort interface {
	Create(profile *model.BandwidthProfile) error
	FindByFilter(filter model.BandwidthProfileFilter) ([]model.BandwidthProfile, error)
	Update(profile *model.BandwidthProfile) error
	Delete(id string) error
}
