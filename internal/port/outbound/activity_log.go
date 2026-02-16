package outbound_port

import "go-template/internal/model"

type ActivityLogDatabasePort interface {
	Create(log *model.ActivityLog) error
	FindByID(id string) (*model.ActivityLog, error)
	FindAll() ([]model.ActivityLog, error)
	Find(filter model.ActivityLogFilter) ([]model.ActivityLog, error)
	FindByEntity(entityType string, entityID string) ([]model.ActivityLog, error)
	FindByUserID(userID string) ([]model.ActivityLog, error)
	FindByAction(action string) ([]model.ActivityLog, error)
}
