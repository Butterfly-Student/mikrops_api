package postgres_outbound_adapter

import (
	"mikrops/internal/model"
	outbound_port "mikrops/internal/port/outbound"
	"time"

	"gorm.io/gorm"
)

type subscriptionAdapter struct {
	db *gorm.DB
}

func NewSubscriptionAdapter(db *gorm.DB) outbound_port.SubscriptionDatabasePort {
	return &subscriptionAdapter{db: db}
}

func (a *subscriptionAdapter) Create(data model.SubscriptionInput) (model.Subscription, error) {
	subscription := model.Subscription{SubscriptionInput: data}
	result := a.db.Create(&subscription)
	return subscription, result.Error
}

func (a *subscriptionAdapter) FindByFilter(filter model.SubscriptionFilter) ([]model.Subscription, error) {
	var subscriptions []model.Subscription
	query := a.db.Model(&model.Subscription{})
	query = applySubscriptionFilter(query, filter)

	if filter.WithTenant {
		query = query.Preload("Tenant")
	}
	if filter.WithCustomer {
		query = query.Preload("Customer")
	}
	if filter.WithPackage {
		query = query.Preload("Package")
	}
	if filter.WithNas {
		query = query.Preload("Nas")
	}

	result := query.Find(&subscriptions)
	return subscriptions, result.Error
}

func (a *subscriptionAdapter) FindByID(id string) (model.Subscription, error) {
	var subscription model.Subscription
	result := a.db.
		Preload("Tenant").
		Preload("Customer").
		Preload("Package").
		Preload("Nas").
		Where("id = ?", id).
		First(&subscription)
	return subscription, result.Error
}

func (a *subscriptionAdapter) Update(id string, data model.SubscriptionInput) error {
	return a.db.Model(&model.Subscription{}).Where("id = ?", id).Updates(data).Error
}

func (a *subscriptionAdapter) FindExpiring(daysBeforeExpiry int, tenantID string) ([]model.Subscription, error) {
	var subscriptions []model.Subscription
	expiryDate := time.Now().AddDate(0, 0, daysBeforeExpiry)

	result := a.db.
		Preload("Customer").
		Preload("Package").
		Where("tenant_id = ?", tenantID).
		Where("status = ?", model.SubscriptionStatusActive).
		Where("end_date <= ?", expiryDate).
		Where("end_date >= ?", time.Now()).
		Find(&subscriptions)

	return subscriptions, result.Error
}

func applySubscriptionFilter(query *gorm.DB, filter model.SubscriptionFilter) *gorm.DB {
	if len(filter.IDs) > 0 {
		query = query.Where("id IN ?", filter.IDs)
	}
	if len(filter.TenantIDs) > 0 {
		query = query.Where("tenant_id IN ?", filter.TenantIDs)
	}
	if len(filter.CustomerIDs) > 0 {
		query = query.Where("customer_id IN ?", filter.CustomerIDs)
	}
	if len(filter.PackageIDs) > 0 {
		query = query.Where("package_id IN ?", filter.PackageIDs)
	}
	if len(filter.NasIDs) > 0 {
		query = query.Where("nas_id IN ?", filter.NasIDs)
	}
	if len(filter.Statuses) > 0 {
		query = query.Where("status IN ?", filter.Statuses)
	}
	if filter.AutoRenew != nil {
		query = query.Where("auto_renew = ?", *filter.AutoRenew)
	}
	return query
}
