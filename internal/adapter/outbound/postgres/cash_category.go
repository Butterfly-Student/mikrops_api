package postgres_outbound_adapter

import (
	"errors"

	"go-template/internal/model"
	outbound_port "go-template/internal/port/outbound"
	"gorm.io/gorm"
)

const tableCashCategory = "cash_categories"

type CashCategoryAdapter struct {
	db *gorm.DB
}

func NewCashCategoryAdapter(db *gorm.DB) outbound_port.CashCategoryDatabasePort {
	return &CashCategoryAdapter{db: db}
}

func (a *CashCategoryAdapter) Create(category *model.CashCategory) error {
	return a.db.Create(category).Error
}

func (a *CashCategoryAdapter) FindByID(id string) (*model.CashCategory, error) {
	var category model.CashCategory
	err := a.db.Preload("ParentCategory").Where("id = ?", id).First(&category).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("category not found")
		}
		return nil, err
	}
	return &category, nil
}

func (a *CashCategoryAdapter) FindAll() ([]model.CashCategory, error) {
	var categories []model.CashCategory
	if err := a.db.Preload("ParentCategory").Where("is_active = true").Order("sort_order ASC, name ASC").Find(&categories).Error; err != nil {
		return nil, err
	}
	return categories, nil
}

func (a *CashCategoryAdapter) Find(filter model.CashCategoryFilter) ([]model.CashCategory, error) {
	var categories []model.CashCategory
	query := a.db.Preload("ParentCategory").Where("deleted_at IS NULL")

	if !filter.IsEmpty() {
		if len(filter.IDs) > 0 {
			query = query.Where("id IN ?", filter.IDs)
		}
		if len(filter.Codes) > 0 {
			query = query.Where("code IN ?", filter.Codes)
		}
		if filter.Type != nil {
			query = query.Where("type = ?", *filter.Type)
		}
		if filter.IsActive != nil {
			query = query.Where("is_active = ?", *filter.IsActive)
		}
		if filter.IsSystem != nil {
			query = query.Where("is_system = ?", *filter.IsSystem)
		}
		if filter.Search != nil {
			search := "%" + *filter.Search + "%"
			query = query.Where("name ILIKE ? OR code ILIKE ?", search, search)
		}
	}

	if err := query.Order("sort_order ASC, name ASC").Find(&categories).Error; err != nil {
		return nil, err
	}
	return categories, nil
}

func (a *CashCategoryAdapter) Update(category *model.CashCategory) error {
	result := a.db.Save(category)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("no rows affected")
	}
	return nil
}

func (a *CashCategoryAdapter) Delete(id string) error {
	result := a.db.Where("id = ?", id).Delete(&model.CashCategory{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("no rows affected")
	}
	return nil
}

func (a *CashCategoryAdapter) FindByCode(code string) (*model.CashCategory, error) {
	var category model.CashCategory
	err := a.db.Where("code = ?", code).First(&category).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("category not found")
		}
		return nil, err
	}
	return &category, nil
}

func (a *CashCategoryAdapter) FindByType(cashType string) ([]model.CashCategory, error) {
	var categories []model.CashCategory
	if err := a.db.Where("type = ?", cashType).Order("sort_order ASC, name ASC").Find(&categories).Error; err != nil {
		return nil, err
	}
	return categories, nil
}
