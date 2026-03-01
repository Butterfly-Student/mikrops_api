package outbound_port

import "go-template/internal/model"

//go:generate mockgen -source=client.go -destination=./../../../tests/mocks/port/mock_client.go
type ClientDatabasePort interface {
	Upsert(datas []model.AdminUserInput) error
	FindByFilter(filter model.AdminUserFilter, lock bool) ([]model.AdminUser, error)
	DeleteByFilter(filter model.AdminUserFilter) error
	IsExists(bearerKey string) (bool, error)
}

type ClientMessagePort interface {
	PublishUpsert(datas []model.AdminUserInput) error
}

type ClientCachePort interface {
	Set(data model.AdminUser) error
	Get(bearerKey string) (model.AdminUser, error)
}

type ClientWorkflowPort interface {
	StartUpsert(data model.AdminUserInput) error
}
