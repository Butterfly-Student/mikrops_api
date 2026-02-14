package outbound_port

import "go-template/internal/model"

type NotificationDatabasePort interface {
	Create(notification *model.Notification) error
	FindByID(id string) (*model.Notification, error)
	FindAll() ([]model.Notification, error)
	Find(filter model.NotificationFilter) ([]model.Notification, error)
	Update(notification *model.Notification) error
	Delete(id string) error
	FindByStatus(statuses []string) ([]model.Notification, error)
	FindByRecipient(recipient string) ([]model.Notification, error)
}

type NotificationTemplateDatabasePort interface {
	Create(template *model.NotificationTemplate) error
	FindByID(id string) (*model.NotificationTemplate, error)
	FindAll() ([]model.NotificationTemplate, error)
	FindByName(name string) (*model.NotificationTemplate, error)
	Update(template *model.NotificationTemplate) error
	Delete(id string) error
}
