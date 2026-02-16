package billing

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

func TestCreateInvoice(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mock_outbound_port.NewMockDatabasePort(ctrl)
	mockSystemSetting := mock_outbound_port.NewMockSystemSettingDatabasePort(ctrl)
	mockCustomerDB := mock_outbound_port.NewMockCustomerDatabasePort(ctrl)
	mockBandwidthProfileDB := mock_outbound_port.NewMockBandwidthProfileDatabasePort(ctrl)

	mockDB.EXPECT().Customer().Return(mockCustomerDB).AnyTimes()
	mockDB.EXPECT().BandwidthProfile().Return(mockBandwidthProfileDB).AnyTimes()

	domain := NewBillingDomain(mockDB, mockSystemSetting)
	ctx := context.Background()

	t.Run("success - create invoice", func(t *testing.T) {
		customerID := uuid.New()
		profileID := uuid.New()
		priceMonthly := 100000.0

		customer := &model.Customer{
			ID:        customerID,
			ProfileID: &profileID,
		}

		profile := &model.BandwidthProfile{
			ID:           profileID,
			PriceMonthly: priceMonthly,
		}

		qty := 1
		input := model.InvoiceInput{
			CustomerID: customerID,
			Items: []model.InvoiceItemInput{
				{
					Description: "Monthly subscription",
					Quantity:    &qty,
					UnitPrice:   priceMonthly,
				},
			},
		}

		mockCustomerDB.EXPECT().
			FindByID(customerID.String()).
			Return(customer, nil).
			Times(1)

		mockBandwidthProfileDB.EXPECT().
			FindByID(profileID.String()).
			Return(profile, nil).
			Times(1)

		mockDB.EXPECT().
			DoInTransaction(gomock.Any()).
			DoAndReturn(func(txFunc interface{}) (interface{}, error) {
				return &model.Invoice{
					ID:         uuid.New(),
					CustomerID: customerID,
					TotalAmount: priceMonthly,
					Status:     "pending",
				}, nil
			}).Times(1)

		result, err := domain.CreateInvoice(ctx, input)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, customerID, result.CustomerID)
	})

	t.Run("error - no invoice items", func(t *testing.T) {
		input := model.InvoiceInput{
			CustomerID: uuid.New(),
			Items:      []model.InvoiceItemInput{},
		}

		result, err := domain.CreateInvoice(ctx, input)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "at least one invoice item is required")
	})
}

func TestGetInvoice(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mock_outbound_port.NewMockDatabasePort(ctrl)
	mockSystemSetting := mock_outbound_port.NewMockSystemSettingDatabasePort(ctrl)
	mockInvoiceDB := mock_outbound_port.NewMockInvoiceDatabasePort(ctrl)

	mockDB.EXPECT().Invoice().Return(mockInvoiceDB).AnyTimes()

	domain := NewBillingDomain(mockDB, mockSystemSetting)
	ctx := context.Background()

	t.Run("success - get invoice", func(t *testing.T) {
		invoiceID := uuid.New()
		customerID := uuid.New()

		expectedInvoice := &model.Invoice{
			ID:          invoiceID,
			CustomerID:  customerID,
			TotalAmount: 100000.0,
			Status:      "pending",
		}

		mockInvoiceDB.EXPECT().
			FindByID(invoiceID.String()).
			Return(expectedInvoice, nil).
			Times(1)

		result, err := domain.GetInvoice(ctx, invoiceID.String())

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, invoiceID, result.ID)
	})

	t.Run("error - invoice not found", func(t *testing.T) {
		invoiceID := uuid.New()

		mockInvoiceDB.EXPECT().
			FindByID(invoiceID.String()).
			Return(nil, errors.New("record not found")).
			Times(1)

		result, err := domain.GetInvoice(ctx, invoiceID.String())

		assert.Error(t, err)
		assert.Nil(t, result)
	})
}

func TestCheckOverdueInvoices(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mock_outbound_port.NewMockDatabasePort(ctrl)
	mockSystemSetting := mock_outbound_port.NewMockSystemSettingDatabasePort(ctrl)
	mockInvoiceDB := mock_outbound_port.NewMockInvoiceDatabasePort(ctrl)

	mockDB.EXPECT().Invoice().Return(mockInvoiceDB).AnyTimes()

	domain := NewBillingDomain(mockDB, mockSystemSetting)
	ctx := context.Background()

	t.Run("success - find overdue invoices", func(t *testing.T) {
		dueDate := time.Now().Add(-7 * 24 * time.Hour) // 7 days ago

		overdueInvoices := []model.Invoice{
			{
				ID:          uuid.New(),
				CustomerID:  uuid.New(),
				DueDate:     dueDate,
				Status:      "pending",
				TotalAmount: 100000.0,
			},
		}

		mockInvoiceDB.EXPECT().
			FindOverdue().
			Return(overdueInvoices, nil).
			Times(1)

		result, err := domain.CheckOverdueInvoices(ctx)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Greater(t, len(result), 0)
	})
}

func TestCalculateLateFee(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mock_outbound_port.NewMockDatabasePort(ctrl)
	mockSystemSetting := mock_outbound_port.NewMockSystemSettingDatabasePort(ctrl)

	domain := NewBillingDomain(mockDB, mockSystemSetting)
	ctx := context.Background()

	t.Run("success - calculate late fee", func(t *testing.T) {
		dueDate := time.Now().Add(-30 * 24 * time.Hour) // 30 days overdue
		invoice := &model.Invoice{
			ID:          uuid.New(),
			DueDate:     dueDate,
			TotalAmount: 100000.0,
			Status:      "pending",
		}

		lateFeeEnabledValue := "true"
		lateFeeAmountValue := "5000"
		lateFeeEnabledSetting := &model.SystemSetting{
			Key:   "invoice.late_fee_enabled",
			Value: &lateFeeEnabledValue,
		}
		lateFeeAmountSetting := &model.SystemSetting{
			Key:   "invoice.late_fee_amount",
			Value: &lateFeeAmountValue,
		}

		mockSystemSetting.EXPECT().
			FindByKey("invoice.late_fee_enabled").
			Return(lateFeeEnabledSetting, nil).
			Times(1)

		mockSystemSetting.EXPECT().
			FindByKey("invoice.late_fee_amount").
			Return(lateFeeAmountSetting, nil).
			Times(1)

		fee := domain.CalculateLateFee(ctx, invoice)

		assert.GreaterOrEqual(t, fee, 0.0)
	})
}
