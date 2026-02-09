package postgres_outbound_adapter

import (
	"mikrops/internal/model"
	outbound_port "mikrops/internal/port/outbound"

	"gorm.io/gorm"
)

type pppoeAccountAdapter struct {
	db *gorm.DB
}

func NewPppoeAccountAdapter(db *gorm.DB) outbound_port.PppoeAccountDatabasePort {
	return &pppoeAccountAdapter{db: db}
}

func (a *pppoeAccountAdapter) Create(data model.PppoeAccountInput) (model.PppoeAccount, error) {
	account := model.PppoeAccount{PppoeAccountInput: data}
	result := a.db.Create(&account)
	return account, result.Error
}

func (a *pppoeAccountAdapter) FindByFilter(filter model.PppoeAccountFilter) ([]model.PppoeAccount, error) {
	var accounts []model.PppoeAccount
	query := a.db.Model(&model.PppoeAccount{})
	query = applyPppoeAccountFilter(query, filter)

	if filter.WithCustomer {
		query = query.Preload("Customer")
	}
	if filter.WithNas {
		query = query.Preload("Nas")
	}
	if filter.WithPackage {
		query = query.Preload("InternetPackage")
	}

	result := query.Find(&accounts)
	return accounts, result.Error
}

func (a *pppoeAccountAdapter) FindByID(id string) (model.PppoeAccount, error) {
	var account model.PppoeAccount
	result := a.db.Preload("Customer").Preload("Nas").Preload("InternetPackage").Where("id = ?", id).First(&account)
	return account, result.Error
}

func (a *pppoeAccountAdapter) Update(id string, data model.PppoeAccountInput) error {
	return a.db.Model(&model.PppoeAccount{}).Where("id = ?", id).Updates(data).Error
}

func (a *pppoeAccountAdapter) Delete(id string) error {
	return a.db.Where("id = ?", id).Delete(&model.PppoeAccount{}).Error
}

func applyPppoeAccountFilter(query *gorm.DB, filter model.PppoeAccountFilter) *gorm.DB {
	if len(filter.IDs) > 0 {
		query = query.Where("id IN ?", filter.IDs)
	}
	if len(filter.TenantIDs) > 0 {
		query = query.Where("tenant_id IN ?", filter.TenantIDs)
	}
	if len(filter.CustomerIDs) > 0 {
		query = query.Where("customer_id IN ?", filter.CustomerIDs)
	}
	if len(filter.NasIDs) > 0 {
		query = query.Where("nas_id IN ?", filter.NasIDs)
	}
	if len(filter.PackageIDs) > 0 {
		query = query.Where("package_id IN ?", filter.PackageIDs)
	}
	if len(filter.Usernames) > 0 {
		query = query.Where("username IN ?", filter.Usernames)
	}
	if len(filter.Statuses) > 0 {
		query = query.Where("status IN ?", filter.Statuses)
	}
	if len(filter.SyncStatuses) > 0 {
		query = query.Where("sync_status IN ?", filter.SyncStatuses)
	}
	return query
}
