package postgres_outbound_adapter

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/palantir/stacktrace"
	"gorm.io/gorm"

	"go-template/internal/model"
	outbound_port "go-template/internal/port/outbound"
)

const tablePayment = "payments"

type paymentAdapter struct {
	db *gorm.DB
}

func NewPaymentAdapter(
	db *gorm.DB,
) outbound_port.PaymentDatabasePort {
	return &paymentAdapter{
		db: db,
	}
}

func (a *paymentAdapter) Create(ctx context.Context, payment *model.Payment) error {
	// Create payment with allocations using GORM associations
	if err := a.db.WithContext(ctx).Session(&gorm.Session{FullSaveAssociations: true}).Create(payment).Error; err != nil {
		return stacktrace.Propagate(err, "failed to create payment in database")
	}
	return nil
}

func (a *paymentAdapter) FindByID(ctx context.Context, id string) (*model.Payment, error) {
	var payment model.Payment

	paymentID, err := uuid.Parse(id)
	if err != nil {
		return nil, stacktrace.Propagate(err, "invalid payment id format")
	}

	if err := a.db.WithContext(ctx).
		Preload("Allocations").
		Preload("Allocations.Invoice").
		Preload("Invoice").
		Preload("Customer").
		Where("id = ?", paymentID).
		First(&payment).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, stacktrace.NewError("payment not found")
		}
		return nil, stacktrace.Propagate(err, "failed to find payment by id")
	}

	return &payment, nil
}

func (a *paymentAdapter) FindByNumber(ctx context.Context, number string) (*model.Payment, error) {
	var payment model.Payment

	if err := a.db.WithContext(ctx).
		Preload("Allocations").
		Preload("Allocations.Invoice").
		Preload("Invoice").
		Preload("Customer").
		Where("payment_number = ?", number).
		First(&payment).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, stacktrace.NewError("payment not found")
		}
		return nil, stacktrace.Propagate(err, "failed to find payment by number")
	}

	return &payment, nil
}

func (a *paymentAdapter) FindAll(ctx context.Context, filter *model.PaymentFilter) ([]model.Payment, error) {
	var payments []model.Payment

	query := a.db.WithContext(ctx).Model(&model.Payment{})

	// Preload relations
	query = query.Preload("Allocations").Preload("Invoice").Preload("Customer")

	// Apply filters if provided
	if filter != nil {
		if filter.CustomerID != nil {
			query = query.Where("customer_id = ?", *filter.CustomerID)
		}

		if filter.InvoiceID != nil {
			query = query.Where("invoice_id = ?", *filter.InvoiceID)
		}

		if filter.Status != nil {
			query = query.Where("status = ?", *filter.Status)
		}

		if filter.PaymentMethod != nil {
			query = query.Where("payment_method = ?", *filter.PaymentMethod)
		}

		if filter.PaymentDateFrom != nil {
			query = query.Where("payment_date >= ?", *filter.PaymentDateFrom)
		}

		if filter.PaymentDateTo != nil {
			query = query.Where("payment_date <= ?", *filter.PaymentDateTo)
		}

		if filter.Search != nil && *filter.Search != "" {
			searchPattern := "%" + *filter.Search + "%"
			query = query.Where(
				"payment_number ILIKE ? OR transaction_reference ILIKE ?",
				searchPattern, searchPattern,
			)
		}

		// Handle soft deletes
		if !filter.IncludeDeleted {
			query = query.Where("deleted_at IS NULL")
		} else {
			query = query.Unscoped()
		}
	} else {
		// Default: exclude deleted
		query = query.Where("deleted_at IS NULL")
	}

	// Order by creation date
	query = query.Order("created_at DESC")

	if err := query.Find(&payments).Error; err != nil {
		return nil, stacktrace.Propagate(err, "failed to find payments")
	}

	return payments, nil
}

func (a *paymentAdapter) Update(ctx context.Context, payment *model.Payment) error {
	// Update payment (associations are handled separately)
	if err := a.db.WithContext(ctx).Save(payment).Error; err != nil {
		return stacktrace.Propagate(err, "failed to update payment in database")
	}
	return nil
}

func (a *paymentAdapter) Delete(ctx context.Context, id string) error {
	paymentID, err := uuid.Parse(id)
	if err != nil {
		return stacktrace.Propagate(err, "invalid payment id format")
	}

	// Soft delete using GORM
	if err := a.db.WithContext(ctx).Delete(&model.Payment{}, paymentID).Error; err != nil {
		return stacktrace.Propagate(err, "failed to delete payment from database")
	}

	return nil
}

func (a *paymentAdapter) CreateAllocation(ctx context.Context, allocation *model.PaymentAllocation) error {
	if err := a.db.WithContext(ctx).Create(allocation).Error; err != nil {
		return stacktrace.Propagate(err, "failed to create payment allocation in database")
	}
	return nil
}

func (a *paymentAdapter) GetLastPaymentNumber(ctx context.Context, year, month int) (string, error) {
	var payment model.Payment

	// Build pattern for this year and month: PAY/YYYY/MM/%
	pattern := fmt.Sprintf("PAY/%d/%02d/%%", year, month)

	if err := a.db.WithContext(ctx).
		Where("payment_number LIKE ?", pattern).
		Order("payment_number DESC").
		First(&payment).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return "", nil // No payment found for this month
		}
		return "", stacktrace.Propagate(err, "failed to get last payment number")
	}

	return payment.PaymentNumber, nil
}
