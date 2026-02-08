package outbound_port

import "mikrops/internal/model"

//go:generate mockgen -source=internet_package.go -destination=./../../../tests/mocks/port/mock_internet_package.go
type InternetPackageDatabasePort interface {
	Create(data model.InternetPackageInput) (model.InternetPackage, error)
	FindByFilter(filter model.InternetPackageFilter) ([]model.InternetPackage, error)
	FindByID(id string) (model.InternetPackage, error)
	Update(id string, data model.InternetPackageInput) error
	Delete(id string) error
}
