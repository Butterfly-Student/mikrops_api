package postgres_outbound_adapter

import (
	"mikrops/internal/model"
	outbound_port "mikrops/internal/port/outbound"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type clientAdapter struct {
	db *gorm.DB
}

func NewClientAdapter(
	db *gorm.DB,
) outbound_port.ClientDatabasePort {
	return &clientAdapter{
		db: db,
	}
}

func (a *clientAdapter) Upsert(datas []model.ClientInput) error {
	clients := make([]model.Client, len(datas))
	for i, d := range datas {
		clients[i] = model.Client{ClientInput: d}
	}

	result := a.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "bearer_key"}},
		DoUpdates: clause.AssignmentColumns([]string{"name", "updated_at"}),
	}).Create(&clients)

	return result.Error
}

func (a *clientAdapter) FindByFilter(filter model.ClientFilter, lock bool) ([]model.Client, error) {
	var clients []model.Client
	query := a.db.Model(&model.Client{})
	query = applyClientFilter(query, filter)

	if lock {
		query = query.Clauses(clause.Locking{Strength: "UPDATE"})
	}

	result := query.Find(&clients)
	if result.Error != nil {
		return nil, result.Error
	}

	return clients, nil
}

func (a *clientAdapter) DeleteByFilter(filter model.ClientFilter) error {
	query := a.db.Model(&model.Client{})
	query = applyClientFilter(query, filter)
	return query.Delete(&model.Client{}).Error
}

func (a *clientAdapter) IsExists(bearerKey string) (bool, error) {
	var count int64
	result := a.db.Model(&model.Client{}).Where("bearer_key = ?", bearerKey).Count(&count)
	if result.Error != nil {
		return false, result.Error
	}
	return count > 0, nil
}

func applyClientFilter(query *gorm.DB, filter model.ClientFilter) *gorm.DB {
	if len(filter.IDs) > 0 {
		query = query.Where("id IN ?", filter.IDs)
	}
	if len(filter.Names) > 0 {
		query = query.Where("name IN ?", filter.Names)
	}
	if len(filter.BearerKeys) > 0 {
		query = query.Where("bearer_key IN ?", filter.BearerKeys)
	}
	return query
}
