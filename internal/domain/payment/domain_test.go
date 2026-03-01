package payment

import (
	"context"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"go-template/internal/model"
	mock_outbound_port "go-template/tests/mocks/port"
)

func TestPaymentDomain_Create(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mock_outbound_port.NewMockDatabasePort(ctrl)
	mockPaymentDB := mock_outbound_port.NewMockPaymentDatabasePort(ctrl)

	mockDB.EXPECT().Payment().Return(mockPaymentDB).AnyTimes()

	domain := NewPaymentDomain(mockDB, nil, nil)

	t.Run("success", func(t *testing.T) {
		customerID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440001")

		input := model.PaymentInput{
			PaymentNumber: "PAY-2024-0001",
			CustomerID:    customerID,
			Amount:        100000,
			PaymentMethod: "bank_transfer",
			PaymentDate:   time.Now(),
		}

		mockPaymentDB.EXPECT().Create(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, p *model.Payment) error {
			assert.Equal(t, input.PaymentNumber, p.PaymentNumber)
			assert.Equal(t, input.CustomerID, p.CustomerID)
			assert.Equal(t, input.Amount, p.Amount)
			assert.Equal(t, model.PaymentMethodBankTransfer, p.PaymentMethod)
			return nil
		})
		mockPaymentDB.EXPECT().FindByID(gomock.Any(), gomock.Any()).Return(&model.Payment{
			PaymentNumber: input.PaymentNumber,
			CustomerID:    input.CustomerID,
			Amount:        input.Amount,
		}, nil)

		payment, err := domain.Create(context.Background(), input)
		assert.NoError(t, err)
		assert.NotNil(t, payment)
		assert.Equal(t, input.PaymentNumber, payment.PaymentNumber)
	})
}

func TestPaymentDomain_GetByID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mock_outbound_port.NewMockDatabasePort(ctrl)
	mockPaymentDB := mock_outbound_port.NewMockPaymentDatabasePort(ctrl)

	mockDB.EXPECT().Payment().Return(mockPaymentDB).AnyTimes()

	domain := NewPaymentDomain(mockDB, nil, nil)

	t.Run("success", func(t *testing.T) {
		paymentID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440001")
		expectedPayment := &model.Payment{
			ID:            paymentID,
			PaymentNumber: "PAY-2024-0001",
			CustomerID:    uuid.MustParse("550e8400-e29b-41d4-a716-446655440002"),
			Amount:        100000,
			Status:        model.PaymentStatusTypePending,
		}

		mockPaymentDB.EXPECT().FindByID(gomock.Any(), paymentID.String()).Return(expectedPayment, nil)

		payment, err := domain.GetByID(context.Background(), paymentID.String())
		assert.NoError(t, err)
		assert.Equal(t, expectedPayment, payment)
	})
}

func TestPaymentDomain_List(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mock_outbound_port.NewMockDatabasePort(ctrl)
	mockPaymentDB := mock_outbound_port.NewMockPaymentDatabasePort(ctrl)

	mockDB.EXPECT().Payment().Return(mockPaymentDB).AnyTimes()

	domain := NewPaymentDomain(mockDB, nil, nil)

	t.Run("success", func(t *testing.T) {
		status := model.PaymentStatusTypeConfirmed
		filter := &model.PaymentFilter{
			Status: &status,
		}

		expectedPayments := []model.Payment{
			{
				ID:            uuid.MustParse("550e8400-e29b-41d4-a716-446655440001"),
				PaymentNumber: "PAY-2024-0001",
				Amount:        100000,
				Status:        model.PaymentStatusTypeConfirmed,
			},
			{
				ID:            uuid.MustParse("550e8400-e29b-41d4-a716-446655440002"),
				PaymentNumber: "PAY-2024-0002",
				Amount:        150000,
				Status:        model.PaymentStatusTypeConfirmed,
			},
		}

		mockPaymentDB.EXPECT().FindAll(gomock.Any(), filter).Return(expectedPayments, nil)

		payments, err := domain.List(context.Background(), filter)
		assert.NoError(t, err)
		assert.Len(t, payments, 2)
	})
}

func TestPayment_IsConfirmed(t *testing.T) {
	t.Run("payment is confirmed", func(t *testing.T) {
		payment := &model.Payment{
			ID:     uuid.New(),
			Status: model.PaymentStatusTypeConfirmed,
		}
		assert.True(t, payment.IsConfirmed())
	})

	t.Run("payment is not confirmed", func(t *testing.T) {
		payment := &model.Payment{
			ID:     uuid.New(),
			Status: model.PaymentStatusTypePending,
		}
		assert.False(t, payment.IsConfirmed())
	})
}

func TestPayment_IsPending(t *testing.T) {
	t.Run("payment is pending", func(t *testing.T) {
		payment := &model.Payment{
			ID:     uuid.New(),
			Status: model.PaymentStatusTypePending,
		}
		assert.True(t, payment.IsPending())
	})

	t.Run("payment is not pending", func(t *testing.T) {
		payment := &model.Payment{
			ID:     uuid.New(),
			Status: model.PaymentStatusTypeConfirmed,
		}
		assert.False(t, payment.IsPending())
	})
}

func TestPayment_GetRemainingAmount(t *testing.T) {
	payment := &model.Payment{
		ID:              uuid.New(),
		Amount:          100000,
		AllocatedAmount: 30000,
	}
	assert.Equal(t, 70000.0, payment.GetRemainingAmount())
}

func TestPayment_CanAllocate(t *testing.T) {
	t.Run("can allocate", func(t *testing.T) {
		payment := &model.Payment{
			ID:              uuid.New(),
			Amount:          100000,
			AllocatedAmount: 30000,
		}
		assert.True(t, payment.CanAllocate(50000))
	})

	t.Run("cannot allocate", func(t *testing.T) {
		payment := &model.Payment{
			ID:              uuid.New(),
			Amount:          100000,
			AllocatedAmount: 80000,
		}
		assert.False(t, payment.CanAllocate(30000))
	})
}

func TestPaymentInput_ToModel(t *testing.T) {
	t.Run("converts input to model", func(t *testing.T) {
		customerID := uuid.New()
		invoiceID := uuid.New()
		status := "confirmed"

		input := model.PaymentInput{
			PaymentNumber: "PAY-2024-0001",
			CustomerID:    customerID,
			InvoiceID:     &invoiceID,
			Amount:        100000,
			PaymentMethod: "bank_transfer",
			PaymentDate:   time.Now(),
			Status:        &status,
		}

		payment := input.ToModel()
		assert.Equal(t, input.PaymentNumber, payment.PaymentNumber)
		assert.Equal(t, input.CustomerID, payment.CustomerID)
		assert.Equal(t, input.Amount, payment.Amount)
		assert.Equal(t, model.PaymentStatusTypeConfirmed, payment.Status)
	})

	t.Run("default status is pending", func(t *testing.T) {
		input := model.PaymentInput{
			PaymentNumber: "PAY-2024-0001",
			CustomerID:    uuid.New(),
			Amount:        100000,
			PaymentMethod: "bank_transfer",
			PaymentDate:   time.Now(),
		}

		payment := input.ToModel()
		assert.Equal(t, model.PaymentStatusTypePending, payment.Status)
	})
}
