package outbound_port

import "mikrops/internal/model"

//go:generate mockgen -source=mikrotik_sync_log.go -destination=./../../../tests/mocks/port/mock_mikrotik_sync_log.go
type MikrotikSyncLogDatabasePort interface {
	Create(data model.MikrotikSyncLogInput) (model.MikrotikSyncLog, error)
	FindByFilter(filter model.MikrotikSyncLogFilter) ([]model.MikrotikSyncLog, error)
	FindByID(id string) (model.MikrotikSyncLog, error)
}
