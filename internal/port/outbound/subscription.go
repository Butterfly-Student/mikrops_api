package outbound_port

import (
	"context"

	"go-template/internal/model"
)

type SubscriptionDatabasePort interface {
	Create(ctx context.Context, subscription *model.Subscription) error
	FindByID(ctx context.Context, id string) (*model.Subscription, error)
	FindByCustomerID(ctx context.Context, customerID string) ([]model.Subscription, error)
	FindActiveByCustomerID(ctx context.Context, customerID string) ([]model.Subscription, error)
	Update(ctx context.Context, subscription *model.Subscription) error
	// ListUsernames returns all subscription usernames for uniqueness checking
	ListUsernames(ctx context.Context) ([]string, error)
}
