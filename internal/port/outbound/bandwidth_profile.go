package outbound_port

import "go-template/internal/model"

type BandwidthProfileDatabasePort interface {
	Create(profile *model.BandwidthProfile) error
	FindByID(id string) (*model.BandwidthProfile, error)
	FindAll() ([]model.BandwidthProfile, error)
	Find(filter model.BandwidthProfileFilter) ([]model.BandwidthProfile, error)
	FindByProfileCode(code string) (*model.BandwidthProfile, error)
	FindByCategory(category string) ([]model.BandwidthProfile, error)
	Update(profile *model.BandwidthProfile) error
	Delete(id string) error
}
