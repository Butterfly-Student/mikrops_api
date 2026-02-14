package postgres_outbound_adapter

import (
	"errors"

	"go-template/internal/model"
	outbound_port "go-template/internal/port/outbound"
	"gorm.io/gorm"
)

const tablePaymentAllocation = "payment_allocations"

type PaymentAllocationAdapter struct {
	db *gorm.DB
}

func NewPaymentAllocationAdapter(
	db *gorm.DB,
) outbound_port.PaymentAllocationDatabasePort {
	return &PaymentAllocationAdapter{
		db: db,
	}
}

func (a *PaymentAllocationAdapter) Create(allocation *model.PaymentAllocation) error {
	if err := a.db.Table(tablePaymentAllocation).Create(allocation).Error; err != nil {
		return err
	}
	return nil
}

func (a *PaymentAllocationAdapter) FindByID(id string) (*model.PaymentAllocation, error) {
	var allocation model.PaymentAllocation
	if err := a.db.Table(tablePaymentAllocation).Preload("Payment").Preload("Invoice").Where("id = ?", id).First(&allocation).Error; err != nil {
		return nil, err
	}
	return &allocation, nil
}

func (a *PaymentAllocationAdapter) FindByPaymentID(paymentID string) ([]model.PaymentAllocation, error) {
	var allocations []model.PaymentAllocation
	if err := a.db.Table(tablePaymentAllocation).Preload("Invoice").Where("payment_id = ?", paymentID).Find(&allocations).Error; err != nil {
		return nil, err
	}
	return allocations, nil
}

func (a *PaymentAllocationAdapter) FindByInvoiceID(invoiceID string) ([]model.PaymentAllocation, error) {
	var allocations []model.PaymentAllocation
	if err := a.db.Table(tablePaymentAllocation).Preload("Payment").Preload("Payment.Customer").Where("invoice_id = ?", invoiceID).Find(&allocations).Error; err != nil {
		return nil, err
	}
	return allocations, nil
}

func (a *PaymentAllocationAdapter) Delete(id string) error {
	result := a.db.Table(tablePaymentAllocation).Where("id = ?", id).Delete(&model.PaymentAllocation{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("no rows affected")
	}
	return nil
}
