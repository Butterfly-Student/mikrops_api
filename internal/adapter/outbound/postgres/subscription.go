package postgres_outbound_adapter

import (
	"context"

	"github.com/google/uuid"
	"github.com/palantir/stacktrace"
	"gorm.io/gorm"

	"go-template/internal/model"
	outbound_port "go-template/internal/port/outbound"
)

type subscriptionAdapter struct {
	db *gorm.DB
}

func NewSubscriptionAdapter(
	db *gorm.DB,
) outbound_port.SubscriptionDatabasePort {
	return &subscriptionAdapter{
		db: db,
	}
}

func (a *subscriptionAdapter) Create(ctx context.Context, subscription *model.Subscription) error {
	if err := a.db.WithContext(ctx).Create(subscription).Error; err != nil {
		return stacktrace.Propagate(err, "failed to create subscription")
	}
	return nil
}

func (a *subscriptionAdapter) FindByID(ctx context.Context, id string) (*model.Subscription, error) {
	var subscription model.Subscription

	subscriptionID, err := uuid.Parse(id)
	if err != nil {
		return nil, stacktrace.Propagate(err, "invalid subscription id format")
	}

	if err := a.db.WithContext(ctx).
		Preload("Customer").
		Preload("Plan").
		Preload("Router").
		Where("id = ?", subscriptionID).
		First(&subscription).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, stacktrace.NewError("subscription not found")
		}
		return nil, stacktrace.Propagate(err, "failed to find subscription by id")
	}

	return &subscription, nil
}

func (a *subscriptionAdapter) FindByCustomerID(ctx context.Context, customerID string) ([]model.Subscription, error) {
	var subscriptions []model.Subscription

	custID, err := uuid.Parse(customerID)
	if err != nil {
		return nil, stacktrace.Propagate(err, "invalid customer id format")
	}

	if err := a.db.WithContext(ctx).
		Preload("Plan").
		Preload("Router").
		Where("customer_id = ?", custID).
		Find(&subscriptions).Error; err != nil {
		return nil, stacktrace.Propagate(err, "failed to find subscriptions by customer id")
	}

	return subscriptions, nil
}

func (a *subscriptionAdapter) FindActiveByCustomerID(ctx context.Context, customerID string) ([]model.Subscription, error) {
	var subscriptions []model.Subscription

	custID, err := uuid.Parse(customerID)
	if err != nil {
		return nil, stacktrace.Propagate(err, "invalid customer id format")
	}

	if err := a.db.WithContext(ctx).
		Preload("Plan").
		Preload("Router").
		Where("customer_id = ? AND status = ?", custID, model.SubscriptionStatusActive).
		Find(&subscriptions).Error; err != nil {
		return nil, stacktrace.Propagate(err, "failed to find active subscriptions by customer id")
	}

	return subscriptions, nil
}

func (a *subscriptionAdapter) Update(ctx context.Context, subscription *model.Subscription) error {
	if err := a.db.WithContext(ctx).Save(subscription).Error; err != nil {
		return stacktrace.Propagate(err, "failed to update subscription")
	}
	return nil
}

func (a *subscriptionAdapter) ListUsernames(ctx context.Context) ([]string, error) {
	var usernames []string

	if err := a.db.WithContext(ctx).
		Model(&model.Subscription{}).
		Where("deleted_at IS NULL").
		Pluck("username", &usernames).Error; err != nil {
		return nil, stacktrace.Propagate(err, "failed to list subscription usernames")
	}

	return usernames, nil
}
