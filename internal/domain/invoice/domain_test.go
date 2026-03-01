package invoice

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"go-template/internal/model"
	mock_outbound_port "go-template/tests/mocks/port"
)

func TestInvoiceDomain_Create(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mock_outbound_port.NewMockDatabasePort(ctrl)
	mockInvoiceDB := mock_outbound_port.NewMockInvoiceDatabasePort(ctrl)
	mockDB.EXPECT().Invoice().Return(mockInvoiceDB).AnyTimes()

	domain := NewInvoiceDomain(mockDB)

	t.Run("success", func(t *testing.T) {
		customerID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440001")
		subscriptionID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440002")

		input := model.InvoiceInput{
			InvoiceNumber:      "INV-2024-0001",
			CustomerID:         customerID,
			SubscriptionID:     &subscriptionID,
			BillingPeriodStart: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			BillingPeriodEnd:   time.Date(2024, 1, 31, 0, 0, 0, 0, time.UTC),
			DueDate:            time.Date(2024, 2, 5, 0, 0, 0, 0, time.UTC),
			TotalAmount:        100000,
		}

		mockInvoiceDB.EXPECT().Create(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, inv *model.Invoice) error {
			assert.Equal(t, input.InvoiceNumber, inv.InvoiceNumber)
			assert.Equal(t, input.CustomerID, inv.CustomerID)
			assert.Equal(t, *input.SubscriptionID, *inv.SubscriptionID)
			return nil
		})
		mockInvoiceDB.EXPECT().FindByID(gomock.Any(), gomock.Any()).Return(&model.Invoice{
			InvoiceNumber: input.InvoiceNumber,
			CustomerID:    input.CustomerID,
			TotalAmount:   input.TotalAmount,
		}, nil)

		invoice, err := domain.Create(context.Background(), input)
		assert.NoError(t, err)
		assert.NotNil(t, invoice)
		assert.Equal(t, input.InvoiceNumber, invoice.InvoiceNumber)
	})
}

func TestInvoiceDomain_GetByID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mock_outbound_port.NewMockDatabasePort(ctrl)
	mockInvoiceDB := mock_outbound_port.NewMockInvoiceDatabasePort(ctrl)
	mockDB.EXPECT().Invoice().Return(mockInvoiceDB).AnyTimes()

	domain := NewInvoiceDomain(mockDB)

	t.Run("success", func(t *testing.T) {
		invoiceID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440001")
		statusUnpaid := model.PaymentStatusUnpaid
		expectedInvoice := &model.Invoice{
			ID:             invoiceID,
			InvoiceNumber:  "INV-2024-0001",
			CustomerID:     uuid.MustParse("550e8400-e29b-41d4-a716-446655440002"),
			TotalAmount:    100000,
			Status:         model.InvoiceStatusDraft,
			PaymentStatus:  &statusUnpaid,
		}

		mockInvoiceDB.EXPECT().FindByID(gomock.Any(), invoiceID.String()).Return(expectedInvoice, nil)

		invoice, err := domain.GetByID(context.Background(), invoiceID.String())
		assert.NoError(t, err)
		assert.Equal(t, expectedInvoice, invoice)
	})

	t.Run("invoice not found", func(t *testing.T) {
		invoiceID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440003")

		mockInvoiceDB.EXPECT().FindByID(gomock.Any(), invoiceID.String()).Return(nil, errors.New("not found"))

		invoice, err := domain.GetByID(context.Background(), invoiceID.String())
		assert.Error(t, err)
		assert.Nil(t, invoice)
	})
}

func TestInvoiceDomain_List(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mock_outbound_port.NewMockDatabasePort(ctrl)
	mockInvoiceDB := mock_outbound_port.NewMockInvoiceDatabasePort(ctrl)
	mockDB.EXPECT().Invoice().Return(mockInvoiceDB).AnyTimes()

	domain := NewInvoiceDomain(mockDB)

	t.Run("success", func(t *testing.T) {
		status := model.InvoiceStatusDraft
		filter := &model.InvoiceFilter{
			Status: &status,
		}

		expectedInvoices := []model.Invoice{
			{
				ID:            uuid.MustParse("550e8400-e29b-41d4-a716-446655440001"),
				InvoiceNumber: "INV-2024-0001",
				TotalAmount:   100000,
				Status:        model.InvoiceStatusDraft,
			},
			{
				ID:            uuid.MustParse("550e8400-e29b-41d4-a716-446655440002"),
				InvoiceNumber: "INV-2024-0002",
				TotalAmount:   150000,
				Status:        model.InvoiceStatusDraft,
			},
		}

		mockInvoiceDB.EXPECT().FindAll(gomock.Any(), filter).Return(expectedInvoices, nil)

		invoices, err := domain.List(context.Background(), filter)
		assert.NoError(t, err)
		assert.Len(t, invoices, 2)
	})
}

func TestInvoiceDomain_Update(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mock_outbound_port.NewMockDatabasePort(ctrl)
	mockInvoiceDB := mock_outbound_port.NewMockInvoiceDatabasePort(ctrl)
	mockDB.EXPECT().Invoice().Return(mockInvoiceDB).AnyTimes()

	domain := NewInvoiceDomain(mockDB)

	t.Run("success", func(t *testing.T) {
		invoiceID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440001")
		existingInvoice := &model.Invoice{
			ID:            invoiceID,
			InvoiceNumber: "INV-2024-0001",
			TotalAmount:   100000,
			Status:        model.InvoiceStatusDraft,
		}

		input := model.InvoiceInput{
			InvoiceNumber:      "INV-2024-0001",
			CustomerID:         uuid.MustParse("550e8400-e29b-41d4-a716-446655440002"),
			BillingPeriodStart: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			BillingPeriodEnd:   time.Date(2024, 1, 31, 0, 0, 0, 0, time.UTC),
			DueDate:            time.Date(2024, 2, 10, 0, 0, 0, 0, time.UTC),
			TotalAmount:        120000,
		}

		mockInvoiceDB.EXPECT().FindByID(gomock.Any(), invoiceID.String()).Return(existingInvoice, nil)
		mockInvoiceDB.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil)
		mockInvoiceDB.EXPECT().FindByID(gomock.Any(), invoiceID.String()).Return(&model.Invoice{
			ID:            invoiceID,
			InvoiceNumber: input.InvoiceNumber,
			TotalAmount:   input.TotalAmount,
		}, nil)

		invoice, err := domain.Update(context.Background(), invoiceID.String(), input)
		assert.NoError(t, err)
		assert.Equal(t, input.TotalAmount, invoice.TotalAmount)
	})
}

func TestInvoiceDomain_Delete(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mock_outbound_port.NewMockDatabasePort(ctrl)
	mockInvoiceDB := mock_outbound_port.NewMockInvoiceDatabasePort(ctrl)
	mockDB.EXPECT().Invoice().Return(mockInvoiceDB).AnyTimes()

	domain := NewInvoiceDomain(mockDB)

	t.Run("success", func(t *testing.T) {
		invoiceID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440001")

		mockInvoiceDB.EXPECT().Delete(gomock.Any(), invoiceID.String()).Return(nil)

		err := domain.Delete(context.Background(), invoiceID.String())
		assert.NoError(t, err)
	})
}

func TestInvoiceDomain_GenerateInvoiceNumber(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mock_outbound_port.NewMockDatabasePort(ctrl)
	mockInvoiceDB := mock_outbound_port.NewMockInvoiceDatabasePort(ctrl)
	mockDB.EXPECT().Invoice().Return(mockInvoiceDB).AnyTimes()

	domain := NewInvoiceDomain(mockDB)

	t.Run("success", func(t *testing.T) {
		mockInvoiceDB.EXPECT().GetLastInvoiceNumber(gomock.Any(), gomock.Any(), gomock.Any()).Return("INV-2024-0005", nil)

		number, err := domain.GenerateInvoiceNumber(context.Background())
		assert.NoError(t, err)
		assert.NotEmpty(t, number)
	})
}

func TestInvoice_IsOverdue(t *testing.T) {
	t.Run("invoice is overdue", func(t *testing.T) {
		invoice := &model.Invoice{
			ID:          uuid.New(),
			DueDate:     time.Now().AddDate(0, 0, -1),
			TotalAmount: 100000,
			PaidAmount:  0,
		}
		// Balance is a computed field in DB, set manually for test
		invoice.Balance = invoice.TotalAmount - invoice.PaidAmount
		assert.True(t, invoice.IsOverdue())
	})

	t.Run("invoice is not overdue", func(t *testing.T) {
		invoice := &model.Invoice{
			ID:          uuid.New(),
			DueDate:     time.Now().AddDate(0, 0, 1),
			TotalAmount: 100000,
			PaidAmount:  0,
		}
		invoice.Balance = invoice.TotalAmount - invoice.PaidAmount
		assert.False(t, invoice.IsOverdue())
	})

	t.Run("invoice is not overdue if paid", func(t *testing.T) {
		invoice := &model.Invoice{
			ID:          uuid.New(),
			DueDate:     time.Now().AddDate(0, 0, -1),
			TotalAmount: 100000,
			PaidAmount:  100000,
		}
		invoice.Balance = invoice.TotalAmount - invoice.PaidAmount
		assert.False(t, invoice.IsOverdue())
	})
}

func TestInvoice_IsPaid(t *testing.T) {
	t.Run("invoice is fully paid", func(t *testing.T) {
		invoice := &model.Invoice{
			ID:          uuid.New(),
			TotalAmount: 100000,
			PaidAmount:  100000,
		}
		assert.True(t, invoice.IsPaid())
	})

	t.Run("invoice is partially paid", func(t *testing.T) {
		invoice := &model.Invoice{
			ID:          uuid.New(),
			TotalAmount: 100000,
			PaidAmount:  50000,
		}
		assert.False(t, invoice.IsPaid())
	})
}

func TestInvoice_GetBalance(t *testing.T) {
	invoice := &model.Invoice{
		ID:          uuid.New(),
		TotalAmount: 100000,
		PaidAmount:  30000,
	}
	assert.Equal(t, 70000.0, invoice.GetBalance())
}

func TestInvoice_UpdatePaymentStatus(t *testing.T) {
	t.Run("fully paid", func(t *testing.T) {
		invoice := &model.Invoice{
			ID:          uuid.New(),
			TotalAmount: 100000,
			PaidAmount:  100000,
		}
		invoice.UpdatePaymentStatus()
		assert.Equal(t, model.InvoiceStatusPaid, invoice.Status)
		assert.Equal(t, model.PaymentStatusPaid, *invoice.PaymentStatus)
	})

	t.Run("partially paid", func(t *testing.T) {
		invoice := &model.Invoice{
			ID:          uuid.New(),
			TotalAmount: 100000,
			PaidAmount:  50000,
		}
		invoice.UpdatePaymentStatus()
		assert.Equal(t, model.InvoiceStatusPartial, invoice.Status)
		assert.Equal(t, model.PaymentStatusPartial, *invoice.PaymentStatus)
	})

	t.Run("unpaid", func(t *testing.T) {
		invoice := &model.Invoice{
			ID:          uuid.New(),
			TotalAmount: 100000,
			PaidAmount:  0,
		}
		invoice.UpdatePaymentStatus()
		status := model.PaymentStatusUnpaid
		assert.Equal(t, &status, invoice.PaymentStatus)
	})
}

func TestInvoiceInput_ToModel(t *testing.T) {
	t.Run("converts input to model", func(t *testing.T) {
		customerID := uuid.New()
		subscriptionID := uuid.New()
		status := "sent"
		invoiceType := "recurring"

		input := model.InvoiceInput{
			InvoiceNumber:      "INV-2024-0001",
			CustomerID:         customerID,
			SubscriptionID:     &subscriptionID,
			BillingPeriodStart: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			BillingPeriodEnd:   time.Date(2024, 1, 31, 0, 0, 0, 0, time.UTC),
			DueDate:            time.Date(2024, 2, 5, 0, 0, 0, 0, time.UTC),
			TotalAmount:        100000,
			Status:             &status,
			InvoiceType:        &invoiceType,
		}

		invoice := input.ToModel()
		assert.Equal(t, input.InvoiceNumber, invoice.InvoiceNumber)
		assert.Equal(t, input.CustomerID, invoice.CustomerID)
		assert.Equal(t, *input.SubscriptionID, *invoice.SubscriptionID)
		assert.Equal(t, model.InvoiceStatusSent, invoice.Status)
		assert.Equal(t, model.InvoiceTypeRecurring, invoice.InvoiceType)
	})

	t.Run("default status is draft", func(t *testing.T) {
		input := model.InvoiceInput{
			InvoiceNumber:      "INV-2024-0001",
			CustomerID:         uuid.New(),
			BillingPeriodStart: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			BillingPeriodEnd:   time.Date(2024, 1, 31, 0, 0, 0, 0, time.UTC),
			DueDate:            time.Date(2024, 2, 5, 0, 0, 0, 0, time.UTC),
			TotalAmount:        100000,
		}

		invoice := input.ToModel()
		assert.Equal(t, model.InvoiceStatusDraft, invoice.Status)
	})
}

func strPtr(s string) *string {
	return &s
}
