package outbound_port

import "mikrops/internal/model"

//go:generate mockgen -source=staff.go -destination=./../../../tests/mocks/port/mock_staff.go
type StaffDatabasePort interface {
	Create(data model.StaffInput) (model.Staff, error)
	FindByFilter(filter model.StaffFilter) ([]model.Staff, error)
	FindByID(id string) (model.Staff, error)
	FindByEmail(email string) (model.Staff, error)
	Update(id string, data model.StaffInput) error
	Delete(id string) error
}

type StaffCachePort interface {
	Set(data model.Staff) error
	Get(staffID string) (model.Staff, error)
	Delete(staffID string) error
}
