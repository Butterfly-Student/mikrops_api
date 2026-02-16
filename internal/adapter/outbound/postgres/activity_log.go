package postgres_outbound_adapter

import (
	"errors"

	"go-template/internal/model"
	outbound_port "go-template/internal/port/outbound"
	"gorm.io/gorm"
)

const tableActivityLog = "activity_logs"

type ActivityLogAdapter struct {
	db *gorm.DB
}

func NewActivityLogAdapter(db *gorm.DB) outbound_port.ActivityLogDatabasePort {
	return &ActivityLogAdapter{db: db}
}

func (a *ActivityLogAdapter) Create(log *model.ActivityLog) error {
	return a.db.Create(log).Error
}

func (a *ActivityLogAdapter) FindByID(id string) (*model.ActivityLog, error) {
	var log model.ActivityLog
	err := a.db.Where("id = ?", id).First(&log).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("activity log not found")
		}
		return nil, err
	}
	return &log, nil
}

func (a *ActivityLogAdapter) FindAll() ([]model.ActivityLog, error) {
	var logs []model.ActivityLog
	if err := a.db.Order("created_at DESC").Find(&logs).Error; err != nil {
		return nil, err
	}
	return logs, nil
}

func (a *ActivityLogAdapter) Find(filter model.ActivityLogFilter) ([]model.ActivityLog, error) {
	var logs []model.ActivityLog
	query := a.db

	if !filter.IsEmpty() {
		if len(filter.IDs) > 0 {
			query = query.Where("id IN ?", filter.IDs)
		}
		if len(filter.UserIDs) > 0 {
			query = query.Where("user_id IN ?", filter.UserIDs)
		}
		if len(filter.Actions) > 0 {
			query = query.Where("action IN ?", filter.Actions)
		}
		if len(filter.EntityTypes) > 0 {
			query = query.Where("entity_type IN ?", filter.EntityTypes)
		}
		if len(filter.EntityIDs) > 0 {
			query = query.Where("entity_id IN ?", filter.EntityIDs)
		}
		if filter.CreatedAtStart != nil {
			query = query.Where("created_at >= ?", *filter.CreatedAtStart)
		}
		if filter.CreatedAtEnd != nil {
			query = query.Where("created_at <= ?", *filter.CreatedAtEnd)
		}
		if filter.Search != nil {
			search := "%" + *filter.Search + "%"
			query = query.Where("description ILIKE ?", search)
		}
	}

	if err := query.Order("created_at DESC").Find(&logs).Error; err != nil {
		return nil, err
	}
	return logs, nil
}

func (a *ActivityLogAdapter) FindByEntity(entityType string, entityID string) ([]model.ActivityLog, error) {
	var logs []model.ActivityLog
	if err := a.db.Where("entity_type = ? AND entity_id = ?", entityType, entityID).Order("created_at DESC").Find(&logs).Error; err != nil {
		return nil, err
	}
	return logs, nil
}

func (a *ActivityLogAdapter) FindByUserID(userID string) ([]model.ActivityLog, error) {
	var logs []model.ActivityLog
	if err := a.db.Where("user_id = ?", userID).Order("created_at DESC").Find(&logs).Error; err != nil {
		return nil, err
	}
	return logs, nil
}

func (a *ActivityLogAdapter) FindByAction(action string) ([]model.ActivityLog, error) {
	var logs []model.ActivityLog
	if err := a.db.Where("action = ?", action).Order("created_at DESC").Find(&logs).Error; err != nil {
		return nil, err
	}
	return logs, nil
}
