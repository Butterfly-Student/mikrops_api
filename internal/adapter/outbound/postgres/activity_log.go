package postgres_outbound_adapter

import (
	"mikrops/internal/model"
	outbound_port "mikrops/internal/port/outbound"

	"gorm.io/gorm"
)

type activityLogAdapter struct {
	db *gorm.DB
}

func NewActivityLogAdapter(db *gorm.DB) outbound_port.ActivityLogDatabasePort {
	return &activityLogAdapter{db: db}
}

func (a *activityLogAdapter) Create(data model.ActivityLogInput) (model.ActivityLog, error) {
	log := model.ActivityLog{ActivityLogInput: data}
	result := a.db.Create(&log)
	return log, result.Error
}

func (a *activityLogAdapter) FindByFilter(filter model.ActivityLogFilter) ([]model.ActivityLog, error) {
	var logs []model.ActivityLog
	query := a.db.Model(&model.ActivityLog{})
	query = applyActivityLogFilter(query, filter)
	result := query.Order("created_at DESC").Find(&logs)
	return logs, result.Error
}

func (a *activityLogAdapter) FindByID(id string) (model.ActivityLog, error) {
	var log model.ActivityLog
	result := a.db.Where("id = ?", id).First(&log)
	return log, result.Error
}

func applyActivityLogFilter(query *gorm.DB, filter model.ActivityLogFilter) *gorm.DB {
	if len(filter.IDs) > 0 {
		query = query.Where("id IN ?", filter.IDs)
	}
	if len(filter.TenantIDs) > 0 {
		query = query.Where("tenant_id IN ?", filter.TenantIDs)
	}
	if len(filter.UserIDs) > 0 {
		query = query.Where("user_id IN ?", filter.UserIDs)
	}
	if len(filter.UserTypes) > 0 {
		query = query.Where("user_type IN ?", filter.UserTypes)
	}
	if len(filter.Actions) > 0 {
		query = query.Where("action IN ?", filter.Actions)
	}
	if len(filter.ResourceTypes) > 0 {
		query = query.Where("resource_type IN ?", filter.ResourceTypes)
	}
	if len(filter.ResourceIDs) > 0 {
		query = query.Where("resource_id IN ?", filter.ResourceIDs)
	}
	return query
}
