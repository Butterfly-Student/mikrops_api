package postgres_outbound_adapter

import (
	"errors"

	"go-template/internal/model"
	outbound_port "go-template/internal/port/outbound"
	"gorm.io/gorm"
)

const tablePayment = "payments"

type PaymentAdapter struct {
	db *gorm.DB
}

func NewPaymentAdapter(
	db *gorm.DB,
) outbound_port.PaymentDatabasePort {
	return &PaymentAdapter{
		db: db,
	}
}

func (a *PaymentAdapter) Create(payment *model.Payment) error {
	if err := a.db.Table(tablePayment).Create(payment).Error; err != nil {
		return err
	}
	return nil
}

func (a *PaymentAdapter) FindByID(id string) (*model.Payment, error) {
	var payment model.Payment
	if err := a.db.Table(tablePayment).Preload("Customer").Preload("Invoice").Preload("CreatedByUser").Preload("ProcessedByUser").Where("id = ? AND deleted_at IS NULL", id).First(&payment).Error; err != nil {
		return nil, err
	}
	return &payment, nil
}

func (a *PaymentAdapter) Find(filter model.PaymentFilter) ([]model.Payment, error) {
	var payments []model.Payment
	query := a.db.Table(tablePayment).Preload("Customer").Preload("Invoice").Preload("CreatedByUser").Preload("ProcessedByUser").Where("deleted_at IS NULL")

	query = a.applyFilters(query, filter)

	if err := query.Find(&payments).Error; err != nil {
		return nil, err
	}
	return payments, nil
}

func (a *PaymentAdapter) Count(filter model.PaymentFilter) (int64, error) {
	var count int64
	query := a.db.Table(tablePayment).Where("deleted_at IS NULL")

	query = a.applyFilters(query, filter)

	if err := query.Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

func (a *PaymentAdapter) applyFilters(query *gorm.DB, filter model.PaymentFilter) *gorm.DB {
	if !filter.IsEmpty() {
		if len(filter.IDs) > 0 {
			query = query.Where("id IN ?", filter.IDs)
		}
		if len(filter.PaymentNumbers) > 0 {
			query = query.Where("payment_number IN ?", filter.PaymentNumbers)
		}
		if len(filter.CustomerIDs) > 0 {
			query = query.Where("customer_id IN ?", filter.CustomerIDs)
		}
		if filter.CustomerID != nil {
			query = query.Where("customer_id = ?", *filter.CustomerID)
		}
		if len(filter.InvoiceIDs) > 0 {
			query = query.Where("invoice_id IN ?", filter.InvoiceIDs)
		}
		if len(filter.Status) > 0 {
			query = query.Where("status IN ?", filter.Status)
		}
		if len(filter.PaymentMethod) > 0 {
			query = query.Where("payment_method IN ?", filter.PaymentMethod)
		}
		if filter.PaymentStart != nil {
			query = query.Where("payment_date >= ?", *filter.PaymentStart)
		}
		if filter.PaymentEnd != nil {
			query = query.Where("payment_date <= ?", *filter.PaymentEnd)
		}
		if filter.AmountMin != nil {
			query = query.Where("amount >= ?", *filter.AmountMin)
		}
		if filter.AmountMax != nil {
			query = query.Where("amount <= ?", *filter.AmountMax)
		}
		if len(filter.EwalletProvider) > 0 {
			query = query.Where("ewallet_provider IN ?", filter.EwalletProvider)
		}
		if filter.TransactionRef != nil {
			query = query.Where("transaction_reference = ?", *filter.TransactionRef)
		}
		if filter.Search != nil {
			search := "%" + *filter.Search + "%"
			query = query.Where("payment_number ILIKE ? OR receipt_number ILIKE ? OR notes ILIKE ?", search, search, search)
		}
		if filter.IsProcessed != nil {
			if *filter.IsProcessed {
				query = query.Where("status = ?", "confirmed")
			} else {
				query = query.Where("status IN ?", []string{"pending", "rejected"})
			}
		}
		if filter.IsRefunded != nil {
			if *filter.IsRefunded {
				query = query.Where("status = ?", "refunded")
			} else {
				query = query.Where("status != ?", "refunded")
			}
		}
	}

	return query
}

func (a *PaymentAdapter) Update(payment *model.Payment) error {
	result := a.db.Table(tablePayment).Save(payment)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("no rows affected")
	}
	return nil
}

func (a *PaymentAdapter) Delete(id string) error {
	result := a.db.Table(tablePayment).Where("id = ?", id).Update("deleted_at", "NOW()")
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("no rows affected")
	}
	return nil
}

func (a *PaymentAdapter) FindByXenditExternalID(externalID string) (*model.Payment, error) {
	var payment model.Payment
	if err := a.db.Table(tablePayment).Preload("Customer").Preload("Invoice").Preload("CreatedByUser").Preload("ProcessedByUser").Where("transaction_reference = ? AND deleted_at IS NULL", externalID).First(&payment).Error; err != nil {
		return nil, err
	}
	return &payment, nil
}

func (a *PaymentAdapter) FindByCustomerID(customerID string) ([]model.Payment, error) {
	var payments []model.Payment
	if err := a.db.Table(tablePayment).Preload("Invoice").Preload("CreatedByUser").Preload("ProcessedByUser").Where("customer_id = ? AND deleted_at IS NULL", customerID).Find(&payments).Error; err != nil {
		return nil, err
	}
	return payments, nil
}

func (a *PaymentAdapter) FindByInvoiceID(invoiceID string) ([]model.Payment, error) {
	var payments []model.Payment
	if err := a.db.Table(tablePayment).Preload("Customer").Preload("CreatedByUser").Preload("ProcessedByUser").Where("invoice_id = ? AND deleted_at IS NULL", invoiceID).Find(&payments).Error; err != nil {
		return nil, err
	}
	return payments, nil
}
