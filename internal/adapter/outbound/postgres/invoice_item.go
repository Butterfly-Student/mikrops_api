package postgres_outbound_adapter

import (
	"errors"

	"go-template/internal/model"
	outbound_port "go-template/internal/port/outbound"
	"gorm.io/gorm"
)

const tableInvoiceItem = "invoice_items"

type InvoiceItemAdapter struct {
	db *gorm.DB
}

func NewInvoiceItemAdapter(db *gorm.DB) outbound_port.InvoiceItemDatabasePort {
	return &InvoiceItemAdapter{db: db}
}

func (a *InvoiceItemAdapter) Create(item *model.InvoiceItem) error {
	return a.db.Create(item).Error
}

func (a *InvoiceItemAdapter) FindByID(id string) (*model.InvoiceItem, error) {
	var item model.InvoiceItem
	err := a.db.Preload("Invoice").Preload("Profile").Where("id = ?", id).First(&item).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("invoice item not found")
		}
		return nil, err
	}
	return &item, nil
}

func (a *InvoiceItemAdapter) FindByInvoiceID(invoiceID string) ([]model.InvoiceItem, error) {
	var items []model.InvoiceItem
	if err := a.db.Preload("Profile").Where("invoice_id = ?", invoiceID).Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

func (a *InvoiceItemAdapter) Update(item *model.InvoiceItem) error {
	result := a.db.Save(item)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("no rows affected")
	}
	return nil
}

func (a *InvoiceItemAdapter) Delete(id string) error {
	result := a.db.Where("id = ?", id).Delete(&model.InvoiceItem{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("no rows affected")
	}
	return nil
}
