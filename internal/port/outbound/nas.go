package outbound_port

import "mikrops/internal/model"

//go:generate mockgen -source=nas.go -destination=./../../../tests/mocks/port/mock_nas.go
type NasDatabasePort interface {
	Create(data model.NasInput) (model.Nas, error)
	FindByFilter(filter model.NasFilter) ([]model.Nas, error)
	FindByID(id string) (model.Nas, error)
	Update(id string, data model.NasInput) error
	Delete(id string) error
	CountByTenantID(tenantID string) (int, error)
}
