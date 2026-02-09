package postgres_outbound_adapter

import (
	"mikrops/internal/model"
	outbound_port "mikrops/internal/port/outbound"

	"gorm.io/gorm"
)

type mikrotikSyncLogAdapter struct {
	db *gorm.DB
}

func NewMikrotikSyncLogAdapter(db *gorm.DB) outbound_port.MikrotikSyncLogDatabasePort {
	return &mikrotikSyncLogAdapter{db: db}
}

func (a *mikrotikSyncLogAdapter) Create(data model.MikrotikSyncLogInput) (model.MikrotikSyncLog, error) {
	syncLog := model.MikrotikSyncLog{MikrotikSyncLogInput: data}
	result := a.db.Create(&syncLog)
	return syncLog, result.Error
}

func (a *mikrotikSyncLogAdapter) FindByFilter(filter model.MikrotikSyncLogFilter) ([]model.MikrotikSyncLog, error) {
	var logs []model.MikrotikSyncLog
	query := a.db.Model(&model.MikrotikSyncLog{})
	query = applyMikrotikSyncLogFilter(query, filter)
	result := query.Order("created_at DESC").Find(&logs)
	return logs, result.Error
}

func (a *mikrotikSyncLogAdapter) FindByID(id string) (model.MikrotikSyncLog, error) {
	var syncLog model.MikrotikSyncLog
	result := a.db.Where("id = ?", id).First(&syncLog)
	return syncLog, result.Error
}

func applyMikrotikSyncLogFilter(query *gorm.DB, filter model.MikrotikSyncLogFilter) *gorm.DB {
	if len(filter.IDs) > 0 {
		query = query.Where("id IN ?", filter.IDs)
	}
	if len(filter.TenantIDs) > 0 {
		query = query.Where("tenant_id IN ?", filter.TenantIDs)
	}
	if len(filter.NasIDs) > 0 {
		query = query.Where("nas_id IN ?", filter.NasIDs)
	}
	if len(filter.Actions) > 0 {
		query = query.Where("action IN ?", filter.Actions)
	}
	if len(filter.ResourceTypes) > 0 {
		query = query.Where("resource_type IN ?", filter.ResourceTypes)
	}
	if len(filter.Statuses) > 0 {
		query = query.Where("status IN ?", filter.Statuses)
	}
	return query
}
