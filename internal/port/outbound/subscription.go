package outbound_port

import (
	"mikrops/internal/model"
)

//go:generate mockgen -source=subscription.go -destination=./../../../tests/mocks/port/mock_subscription.go
type SubscriptionDatabasePort interface {
	Create(data model.SubscriptionInput) (model.Subscription, error)
	FindByFilter(filter model.SubscriptionFilter) ([]model.Subscription, error)
	FindByID(id string) (model.Subscription, error)
	Update(id string, data model.SubscriptionInput) error
	FindExpiring(daysBeforeExpiry int, tenantID string) ([]model.Subscription, error)
}
