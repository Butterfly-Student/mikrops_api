package payment

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
	"go-template/utils/xendit"
)

func TestCreatePayment(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mock_outbound_port.NewMockDatabasePort(ctrl)
	mockPaymentDB := mock_outbound_port.NewMockPaymentDatabasePort(ctrl)
	mockCustomerDB := mock_outbound_port.NewMockCustomerDatabasePort(ctrl)

	mockDB.EXPECT().Payment().Return(mockPaymentDB).AnyTimes()
	mockDB.EXPECT().Customer().Return(mockCustomerDB).AnyTimes()

	xenditClient := &xendit.Client{}
	domain := NewPaymentDomain(mockDB, xenditClient)
	ctx := context.Background()

	t.Run("success - create payment", func(t *testing.T) {
		customerID := uuid.New()
		invoiceID := uuid.New()
		amount := 100000.0
		paymentMethod := "va"
		paymentDate := time.Now()

		input := model.PaymentInput{
			CustomerID:    customerID,
			InvoiceID:     &invoiceID,
			Amount:        &amount,
			PaymentMethod: &paymentMethod,
			PaymentDate:   &paymentDate,
		}

		customer := &model.Customer{
			ID:           customerID,
			CustomerCode: "CUST-001",
			FullName:     "John Doe",
		}

		mockCustomerDB.EXPECT().
			FindByID(customerID.String()).
			Return(customer, nil).
			Times(1)

		mockPaymentDB.EXPECT().
			Create(gomock.Any()).
			DoAndReturn(func(payment *model.Payment) error {
				assert.Equal(t, customerID, payment.CustomerID)
				assert.Equal(t, amount, payment.Amount)
				assert.Equal(t, "pending", payment.Status)
				return nil
			}).
			Times(1)

		result, err := domain.CreatePayment(ctx, input)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, customerID, result.CustomerID)
		assert.Equal(t, amount, result.Amount)
	})

	t.Run("error - customer not found", func(t *testing.T) {
		customerID := uuid.New()
		amount := 100000.0

		input := model.PaymentInput{
			CustomerID: customerID,
			Amount:     &amount,
		}

		mockCustomerDB.EXPECT().
			FindByID(customerID.String()).
			Return(nil, errors.New("not found")).
			Times(1)

		result, err := domain.CreatePayment(ctx, input)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "customer not found")
	})
}

func TestGetPayment(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mock_outbound_port.NewMockDatabasePort(ctrl)
	mockPaymentDB := mock_outbound_port.NewMockPaymentDatabasePort(ctrl)

	mockDB.EXPECT().Payment().Return(mockPaymentDB).AnyTimes()

	domain := NewPaymentDomain(mockDB, nil)
	ctx := context.Background()

	t.Run("success - get payment", func(t *testing.T) {
		paymentID := uuid.New().String()
		expectedPayment := &model.Payment{
			ID:            uuid.MustParse(paymentID),
			PaymentNumber: "PAY-001",
			Amount:        100000.0,
			Status:        "confirmed",
		}

		mockPaymentDB.EXPECT().
			FindByID(paymentID).
			Return(expectedPayment, nil).
			Times(1)

		result, err := domain.GetPayment(ctx, paymentID)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "PAY-001", result.PaymentNumber)
	})

	t.Run("error - payment not found", func(t *testing.T) {
		paymentID := uuid.New().String()

		mockPaymentDB.EXPECT().
			FindByID(paymentID).
			Return(nil, errors.New("not found")).
			Times(1)

		result, err := domain.GetPayment(ctx, paymentID)

		assert.Error(t, err)
		assert.Nil(t, result)
	})

	t.Run("error - empty payment ID", func(t *testing.T) {
		result, err := domain.GetPayment(ctx, "")

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "payment ID is required")
	})
}

func TestAllocatePayment(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mock_outbound_port.NewMockDatabasePort(ctrl)
	mockPaymentDB := mock_outbound_port.NewMockPaymentDatabasePort(ctrl)
	mockAllocationDB := mock_outbound_port.NewMockPaymentAllocationDatabasePort(ctrl)

	mockDB.EXPECT().Payment().Return(mockPaymentDB).AnyTimes()
	mockDB.EXPECT().PaymentAllocation().Return(mockAllocationDB).AnyTimes()

	domain := NewPaymentDomain(mockDB, nil)
	ctx := context.Background()

	t.Run("success - allocate payment", func(t *testing.T) {
		paymentID := uuid.New().String()
		invoiceID := uuid.New().String()
		amount := 50000.0

		payment := &model.Payment{
			ID:              uuid.MustParse(paymentID),
			PaymentNumber:   "PAY-001",
			Amount:          100000.0,
			AllocatedAmount: 0,
			Status:          "confirmed",
		}

		mockPaymentDB.EXPECT().
			FindByID(paymentID).
			Return(payment, nil).
			Times(1)

		mockAllocationDB.EXPECT().
			Create(gomock.Any()).
			DoAndReturn(func(allocation *model.PaymentAllocation) error {
				assert.Equal(t, amount, allocation.AllocatedAmount)
				return nil
			}).
			Times(1)

		mockPaymentDB.EXPECT().
			Update(gomock.Any()).
			DoAndReturn(func(p *model.Payment) error {
				assert.Equal(t, amount, p.AllocatedAmount)
				return nil
			}).
			Times(1)

		err := domain.AllocatePayment(ctx, paymentID, invoiceID, amount)

		assert.NoError(t, err)
	})

	t.Run("error - allocation exceeds payment amount", func(t *testing.T) {
		paymentID := uuid.New().String()
		invoiceID := uuid.New().String()
		amount := 150000.0

		payment := &model.Payment{
			ID:              uuid.MustParse(paymentID),
			Amount:          100000.0,
			AllocatedAmount: 0,
		}

		mockPaymentDB.EXPECT().
			FindByID(paymentID).
			Return(payment, nil).
			Times(1)

		err := domain.AllocatePayment(ctx, paymentID, invoiceID, amount)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "allocation amount exceeds payment amount")
	})

	t.Run("error - invalid amount", func(t *testing.T) {
		paymentID := uuid.New().String()
		invoiceID := uuid.New().String()
		amount := 0.0

		err := domain.AllocatePayment(ctx, paymentID, invoiceID, amount)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "allocation amount must be greater than 0")
	})
}

func TestGenerateReceipt(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mock_outbound_port.NewMockDatabasePort(ctrl)
	mockPaymentDB := mock_outbound_port.NewMockPaymentDatabasePort(ctrl)
	mockCustomerDB := mock_outbound_port.NewMockCustomerDatabasePort(ctrl)
	mockInvoiceDB := mock_outbound_port.NewMockInvoiceDatabasePort(ctrl)
	mockInvoiceItemDB := mock_outbound_port.NewMockInvoiceItemDatabasePort(ctrl)
	mockSystemSettingDB := mock_outbound_port.NewMockSystemSettingDatabasePort(ctrl)

	mockDB.EXPECT().Payment().Return(mockPaymentDB).AnyTimes()
	mockDB.EXPECT().Customer().Return(mockCustomerDB).AnyTimes()
	mockDB.EXPECT().Invoice().Return(mockInvoiceDB).AnyTimes()
	mockDB.EXPECT().InvoiceItem().Return(mockInvoiceItemDB).AnyTimes()
	mockDB.EXPECT().SystemSetting().Return(mockSystemSettingDB).AnyTimes()

	domain := NewPaymentDomain(mockDB, nil)
	ctx := context.Background()

	t.Run("success - generate receipt with invoice", func(t *testing.T) {
		paymentID := uuid.New().String()
		customerID := uuid.New()
		invoiceID := uuid.New()
		email := "john@example.com"
		address := "Jl. Raya No. 123"

		payment := &model.Payment{
			ID:            uuid.MustParse(paymentID),
			PaymentNumber: "PAY-001",
			CustomerID:    customerID,
			InvoiceID:     &invoiceID,
			Amount:        100000.0,
			PaymentMethod: "va",
			PaymentDate:   time.Now(),
			Status:        "confirmed",
			CreatedAt:     time.Now(),
		}

		customer := &model.Customer{
			ID:           customerID,
			CustomerCode: "CUST-001",
			FullName:     "John Doe",
			Email:        &email,
			Phone:        "081234567890",
			Address:      &address,
		}

		invoice := &model.Invoice{
			ID:            invoiceID,
			InvoiceNumber: "INV-001",
			Subtotal:      90909.09,
			TaxAmount:     9090.91,
			TotalAmount:   100000.0,
		}

		invoiceItems := []model.InvoiceItem{
			{
				Description: "10 Mbps Package",
				Quantity:    1,
				UnitPrice:   90909.09,
				Total:       90909.09,
			},
		}

		companyName := "PT Internet Provider"
		mockSystemSettingDB.EXPECT().
			FindByKey("company.name").
			Return(&model.SystemSetting{Value: &companyName}, nil).
			AnyTimes()

		mockSystemSettingDB.EXPECT().
			FindByKey(gomock.Any()).
			Return(nil, errors.New("not found")).
			AnyTimes()

		mockPaymentDB.EXPECT().
			FindByID(paymentID).
			Return(payment, nil).
			Times(1)

		mockCustomerDB.EXPECT().
			FindByID(customerID.String()).
			Return(customer, nil).
			Times(1)

		mockInvoiceDB.EXPECT().
			FindByID(invoiceID.String()).
			Return(invoice, nil).
			Times(1)

		mockInvoiceItemDB.EXPECT().
			FindByInvoiceID(invoiceID.String()).
			Return(invoiceItems, nil).
			Times(1)

		result, err := domain.GenerateReceipt(ctx, paymentID)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Greater(t, len(result), 0) // PDF bytes should not be empty
	})

	t.Run("error - payment not found", func(t *testing.T) {
		paymentID := uuid.New().String()

		mockPaymentDB.EXPECT().
			FindByID(paymentID).
			Return(nil, errors.New("not found")).
			Times(1)

		result, err := domain.GenerateReceipt(ctx, paymentID)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "payment not found")
	})
}

func TestGetPaymentHistory(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mock_outbound_port.NewMockDatabasePort(ctrl)
	mockPaymentDB := mock_outbound_port.NewMockPaymentDatabasePort(ctrl)

	mockDB.EXPECT().Payment().Return(mockPaymentDB).AnyTimes()

	domain := NewPaymentDomain(mockDB, nil)
	ctx := context.Background()

	t.Run("success - get payment history", func(t *testing.T) {
		customerID := uuid.New().String()

		payments := []model.Payment{
			{
				ID:            uuid.New(),
				PaymentNumber: "PAY-001",
				Amount:        100000.0,
				Status:        "confirmed",
			},
			{
				ID:            uuid.New(),
				PaymentNumber: "PAY-002",
				Amount:        150000.0,
				Status:        "confirmed",
			},
		}

		mockPaymentDB.EXPECT().
			Find(gomock.Any()).
			Return(payments, nil).
			Times(1)

		mockPaymentDB.EXPECT().
			Count(gomock.Any()).
			Return(int64(2), nil).
			Times(1)

		result, count, err := domain.GetPaymentHistory(ctx, customerID, model.PaymentFilter{})

		assert.NoError(t, err)
		assert.Len(t, result, 2)
		assert.Equal(t, int64(2), count)
	})

	t.Run("error - empty customer ID", func(t *testing.T) {
		result, count, err := domain.GetPaymentHistory(ctx, "", model.PaymentFilter{})

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, int64(0), count)
		assert.Contains(t, err.Error(), "customer ID is required")
	})
}

func TestGetPaymentStatistics(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mock_outbound_port.NewMockDatabasePort(ctrl)
	mockPaymentDB := mock_outbound_port.NewMockPaymentDatabasePort(ctrl)

	mockDB.EXPECT().Payment().Return(mockPaymentDB).AnyTimes()

	domain := NewPaymentDomain(mockDB, nil)
	ctx := context.Background()

	t.Run("success - get payment statistics", func(t *testing.T) {
		customerID := uuid.New().String()
		lastPaymentDate := time.Now()

		payments := []model.Payment{
			{
				ID:            uuid.New(),
				PaymentNumber: "PAY-001",
				Amount:        100000.0,
				Status:        "confirmed",
				PaymentDate:   lastPaymentDate.Add(-5 * 24 * time.Hour),
			},
			{
				ID:            uuid.New(),
				PaymentNumber: "PAY-002",
				Amount:        150000.0,
				Status:        "confirmed",
				PaymentDate:   lastPaymentDate,
			},
			{
				ID:            uuid.New(),
				PaymentNumber: "PAY-003",
				Amount:        200000.0,
				Status:        "confirmed",
				PaymentDate:   lastPaymentDate.Add(-10 * 24 * time.Hour),
			},
		}

		mockPaymentDB.EXPECT().
			Find(gomock.Any()).
			Return(payments, nil).
			Times(1)

		result, err := domain.GetPaymentStatistics(ctx, customerID)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, int64(3), result.TotalPayments)
		assert.Equal(t, 450000.0, result.TotalAmount)
		assert.Equal(t, 150000.0, result.AverageAmount)
		assert.Equal(t, 150000.0, result.LastPaymentAmount)
		assert.NotNil(t, result.LastPaymentDate)
	})

	t.Run("success - no payments", func(t *testing.T) {
		customerID := uuid.New().String()

		mockPaymentDB.EXPECT().
			Find(gomock.Any()).
			Return([]model.Payment{}, nil).
			Times(1)

		result, err := domain.GetPaymentStatistics(ctx, customerID)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, int64(0), result.TotalPayments)
		assert.Equal(t, 0.0, result.TotalAmount)
		assert.Nil(t, result.LastPaymentDate)
	})
}
