package postgres_outbound_adapter

import (
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"

	"go-template/internal/model"
	outbound_port "go-template/internal/port/outbound"
)

type invoiceAdapter struct {
	db *gorm.DB
}

func NewInvoiceAdapter(
	db *gorm.DB,
) outbound_port.InvoiceDatabasePort {
	return &invoiceAdapter{
		db: db,
	}
}

func (adapter *invoiceAdapter) Create(invoice *model.Invoice) error {
	return adapter.db.Create(invoice).Error
}

func (adapter *invoiceAdapter) CreateWithItems(invoice *model.Invoice, items []model.InvoiceItem) error {
	return adapter.db.Transaction(func(tx *gorm.DB) error {
		// Create invoice
		if err := tx.Create(invoice).Error; err != nil {
			return err
		}

		// Set invoice ID for all items
		for i := range items {
			items[i].InvoiceID = invoice.ID
		}

		// Create items
		if len(items) > 0 {
			if err := tx.Create(&items).Error; err != nil {
				return err
			}
		}

		// Load items back to invoice
		invoice.Items = items

		return nil
	})
}

func (adapter *invoiceAdapter) FindByFilter(filter model.InvoiceFilter) ([]model.Invoice, error) {
	var invoices []model.Invoice

	query := adapter.db.Model(&model.Invoice{})

	// Preload relationships
	query = query.Preload("Customer").
		Preload("Items").
		Preload("Items.Profile").
		Preload("CreatedByUser").
		Preload("UpdatedByUser")

	// Apply filters
	if len(filter.IDs) > 0 {
		query = query.Where("id IN ?", filter.IDs)
	}

	if len(filter.InvoiceNumbers) > 0 {
		query = query.Where("invoice_number IN ?", filter.InvoiceNumbers)
	}

	if len(filter.CustomerIDs) > 0 {
		query = query.Where("customer_id IN ?", filter.CustomerIDs)
	}

	if len(filter.Statuses) > 0 {
		query = query.Where("status IN ?", filter.Statuses)
	}

	if len(filter.PaymentStatuses) > 0 {
		query = query.Where("payment_status IN ?", filter.PaymentStatuses)
	}

	if len(filter.InvoiceTypes) > 0 {
		query = query.Where("invoice_type IN ?", filter.InvoiceTypes)
	}

	if filter.BillingMonth != nil {
		query = query.Where("billing_month = ?", *filter.BillingMonth)
	}

	if filter.BillingYear != nil {
		query = query.Where("billing_year = ?", *filter.BillingYear)
	}

	if filter.DueBefore != nil {
		query = query.Where("due_date < ?", filter.DueBefore)
	}

	if filter.DueAfter != nil {
		query = query.Where("due_date > ?", filter.DueAfter)
	}

	if filter.IssuedBefore != nil {
		query = query.Where("issue_date < ?", filter.IssuedBefore)
	}

	if filter.IssuedAfter != nil {
		query = query.Where("issue_date > ?", filter.IssuedAfter)
	}

	if filter.Search != "" {
		searchPattern := "%" + filter.Search + "%"
		query = query.Where("invoice_number ILIKE ? OR notes ILIKE ?",
			searchPattern, searchPattern)
	}

	// Exclude soft-deleted records
	query = query.Where("deleted_at IS NULL")

	// Order by issue date desc
	query = query.Order("issue_date DESC, created_at DESC")

	// Pagination
	if filter.Limit > 0 {
		query = query.Limit(filter.Limit)
		if filter.Page > 0 {
			offset := (filter.Page - 1) * filter.Limit
			query = query.Offset(offset)
		}
	}

	// Execute query
	err := query.Find(&invoices).Error
	if err != nil {
		return nil, err
	}

	return invoices, nil
}

func (adapter *invoiceAdapter) Update(invoice *model.Invoice) error {
	return adapter.db.Save(invoice).Error
}

func (adapter *invoiceAdapter) Delete(id string) error {
	// Soft delete
	return adapter.db.Model(&model.Invoice{}).
		Where("id = ?", id).
		Update("deleted_at", gorm.Expr("CURRENT_TIMESTAMP")).Error
}

func (adapter *invoiceAdapter) GetNextSequenceNumber(year int, month int) (int, error) {
	var count int64

	err := adapter.db.Model(&model.Invoice{}).
		Where("billing_year = ? AND billing_month = ?", year, month).
		Where("deleted_at IS NULL").
		Count(&count).Error

	if err != nil {
		return 0, err
	}

	return int(count) + 1, nil
}

func (adapter *invoiceAdapter) FindOverdueInvoices() ([]model.Invoice, error) {
	var invoices []model.Invoice

	now := time.Now()

	query := adapter.db.Model(&model.Invoice{}).
		Preload("Customer").
		Preload("Items").
		Where("due_date < ?", now).
		Where("payment_status IN ?", []string{
			string(model.PaymentStatusUnpaid),
			string(model.PaymentStatusPartial),
		}).
		Where("status NOT IN ?", []string{
			string(model.InvoiceStatusCancelled),
			string(model.InvoiceStatusRefunded),
		}).
		Where("deleted_at IS NULL")

	err := query.Find(&invoices).Error
	if err != nil {
		return nil, err
	}

	return invoices, nil
}

func (adapter *invoiceAdapter) FindInvoicesDueInDays(days int) ([]model.Invoice, error) {
	var invoices []model.Invoice

	now := time.Now()
	futureDate := now.AddDate(0, 0, days)

	query := adapter.db.Model(&model.Invoice{}).
		Preload("Customer").
		Preload("Items").
		Where("due_date BETWEEN ? AND ?", now, futureDate).
		Where("payment_status = ?", string(model.PaymentStatusUnpaid)).
		Where("status NOT IN ?", []string{
			string(model.InvoiceStatusCancelled),
			string(model.InvoiceStatusRefunded),
		}).
		Where("deleted_at IS NULL")

	err := query.Find(&invoices).Error
	if err != nil {
		return nil, err
	}

	return invoices, nil
}

func (adapter *invoiceAdapter) CountInvoicesByPeriod(year int, month int) (int64, error) {
	var count int64

	err := adapter.db.Model(&model.Invoice{}).
		Where("billing_year = ? AND billing_month = ?", year, month).
		Where("deleted_at IS NULL").
		Count(&count).Error

	if err != nil {
		return 0, err
	}

	return count, nil
}

func (adapter *invoiceAdapter) GetMonthlyRevenue(year int, month int) (map[string]interface{}, error) {
	type RevenueStats struct {
		TotalInvoices  int64
		TotalAmount    decimal.Decimal
		PaidAmount     decimal.Decimal
		PendingAmount  decimal.Decimal
		OverdueAmount  decimal.Decimal
	}

	var stats RevenueStats

	// Get total invoices
	err := adapter.db.Model(&model.Invoice{}).
		Where("billing_year = ? AND billing_month = ?", year, month).
		Where("deleted_at IS NULL").
		Count(&stats.TotalInvoices).Error

	if err != nil {
		return nil, err
	}

	// Get total amount
	err = adapter.db.Model(&model.Invoice{}).
		Select("COALESCE(SUM(total_amount), 0) as total_amount").
		Where("billing_year = ? AND billing_month = ?", year, month).
		Where("deleted_at IS NULL").
		Scan(&stats.TotalAmount).Error

	if err != nil {
		return nil, err
	}

	// Get paid amount
	err = adapter.db.Model(&model.Invoice{}).
		Select("COALESCE(SUM(paid_amount), 0) as paid_amount").
		Where("billing_year = ? AND billing_month = ?", year, month).
		Where("payment_status = ?", string(model.PaymentStatusPaid)).
		Where("deleted_at IS NULL").
		Scan(&stats.PaidAmount).Error

	if err != nil {
		return nil, err
	}

	// Get pending amount
	err = adapter.db.Model(&model.Invoice{}).
		Select("COALESCE(SUM(total_amount - paid_amount), 0) as pending_amount").
		Where("billing_year = ? AND billing_month = ?", year, month).
		Where("payment_status IN ?", []string{
			string(model.PaymentStatusUnpaid),
			string(model.PaymentStatusPartial),
		}).
		Where("due_date >= ?", time.Now()).
		Where("deleted_at IS NULL").
		Scan(&stats.PendingAmount).Error

	if err != nil {
		return nil, err
	}

	// Get overdue amount
	err = adapter.db.Model(&model.Invoice{}).
		Select("COALESCE(SUM(total_amount - paid_amount), 0) as overdue_amount").
		Where("billing_year = ? AND billing_month = ?", year, month).
		Where("payment_status IN ?", []string{
			string(model.PaymentStatusUnpaid),
			string(model.PaymentStatusPartial),
		}).
		Where("due_date < ?", time.Now()).
		Where("deleted_at IS NULL").
		Scan(&stats.OverdueAmount).Error

	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"total_invoices":  stats.TotalInvoices,
		"total_amount":    stats.TotalAmount.String(),
		"paid_amount":     stats.PaidAmount.String(),
		"pending_amount":  stats.PendingAmount.String(),
		"overdue_amount":  stats.OverdueAmount.String(),
	}, nil
}

func (adapter *invoiceAdapter) UpdatePaymentStatus(invoiceID string, paidAmount string, paymentDate time.Time, paymentMethod string) error {
	amount, err := decimal.NewFromString(paidAmount)
	if err != nil {
		return err
	}

	return adapter.db.Transaction(func(tx *gorm.DB) error {
		// Get invoice
		var invoice model.Invoice
		if err := tx.Where("id = ?", invoiceID).First(&invoice).Error; err != nil {
			return err
		}

		// Update paid amount
		invoice.PaidAmount = invoice.PaidAmount.Add(amount)
		invoice.PaymentDate = &paymentDate
		invoice.PaymentMethod = paymentMethod

		// Determine payment status
		if invoice.PaidAmount.GreaterThanOrEqual(invoice.TotalAmount) {
			invoice.PaymentStatus = model.PaymentStatusPaid
			invoice.Status = model.InvoiceStatusPaid
		} else if invoice.PaidAmount.GreaterThan(decimal.Zero) {
			invoice.PaymentStatus = model.PaymentStatusPartial
			invoice.Status = model.InvoiceStatusPartial
		}

		// Save invoice
		return tx.Save(&invoice).Error
	})
}
