package outbound_port

import "mikrops/internal/model"

//go:generate mockgen -source=activity_log.go -destination=./../../../tests/mocks/port/mock_activity_log.go
type ActivityLogDatabasePort interface {
	Create(data model.ActivityLogInput) (model.ActivityLog, error)
	FindByFilter(filter model.ActivityLogFilter) ([]model.ActivityLog, error)
	FindByID(id string) (model.ActivityLog, error)
}
