package cash

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

func TestCreateCategory(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mock_outbound_port.NewMockDatabasePort(ctrl)
	mockCashCategoryDB := mock_outbound_port.NewMockCashCategoryDatabasePort(ctrl)

	mockDB.EXPECT().CashCategory().Return(mockCashCategoryDB).AnyTimes()

	domain := NewCashDomain(mockDB)
	ctx := context.Background()

	t.Run("success - create category", func(t *testing.T) {
		code := "INC-001"
		name := "Payment Received"
		cashType := "income"

		input := model.CashCategoryInput{
			Code: &code,
			Name: &name,
			Type: &cashType,
		}

		mockCashCategoryDB.EXPECT().
			Create(gomock.Any()).
			DoAndReturn(func(category *model.CashCategory) error {
				category.ID = uuid.New()
				return nil
			}).Times(1)

		result, err := domain.CreateCategory(ctx, input)

		assert.NoError(t, err)
		assert.NotNil(t, result)
	})

	t.Run("error - database error", func(t *testing.T) {
		code := "INC-002"
		name := "Test Category"
		cashType := "income"

		input := model.CashCategoryInput{
			Code: &code,
			Name: &name,
			Type: &cashType,
		}

		mockCashCategoryDB.EXPECT().
			Create(gomock.Any()).
			Return(errors.New("database error")).
			Times(1)

		result, err := domain.CreateCategory(ctx, input)

		assert.Error(t, err)
		assert.Nil(t, result)
	})
}

func TestGetCategory(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mock_outbound_port.NewMockDatabasePort(ctrl)
	mockCashCategoryDB := mock_outbound_port.NewMockCashCategoryDatabasePort(ctrl)

	mockDB.EXPECT().CashCategory().Return(mockCashCategoryDB).AnyTimes()

	domain := NewCashDomain(mockDB)
	ctx := context.Background()

	t.Run("success - get category", func(t *testing.T) {
		categoryID := uuid.New()

		expectedCategory := &model.CashCategory{
			ID:   categoryID,
			Code: "INC-001",
			Name: "Payment Received",
			Type: "income",
		}

		mockCashCategoryDB.EXPECT().
			FindByID(categoryID.String()).
			Return(expectedCategory, nil).
			Times(1)

		result, err := domain.GetCategory(ctx, categoryID.String())

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, categoryID, result.ID)
	})
}

func TestCreateTransaction(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mock_outbound_port.NewMockDatabasePort(ctrl)
	mockCashTransactionDB := mock_outbound_port.NewMockCashTransactionDatabasePort(ctrl)

	mockDB.EXPECT().CashTransaction().Return(mockCashTransactionDB).AnyTimes()

	domain := NewCashDomain(mockDB)
	ctx := context.Background()

	t.Run("success - create transaction", func(t *testing.T) {
		categoryID := uuid.New()
		amount := 100000.0
		description := "Payment received"

		input := model.CashTransactionInput{
			CategoryID:  categoryID,
			Amount:      &amount,
			Description: &description,
		}

		mockCashTransactionDB.EXPECT().
			Create(gomock.Any()).
			DoAndReturn(func(transaction *model.CashTransaction) error {
				transaction.ID = uuid.New()
				return nil
			}).Times(1)

		result, err := domain.CreateTransaction(ctx, input)

		assert.NoError(t, err)
		assert.NotNil(t, result)
	})
}

func TestApproveTransaction(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mock_outbound_port.NewMockDatabasePort(ctrl)
	mockCashTransactionDB := mock_outbound_port.NewMockCashTransactionDatabasePort(ctrl)

	mockDB.EXPECT().CashTransaction().Return(mockCashTransactionDB).AnyTimes()

	domain := NewCashDomain(mockDB)
	ctx := context.Background()

	t.Run("success - approve transaction", func(t *testing.T) {
		transactionID := uuid.New()
		userID := uuid.New()

		existingTransaction := &model.CashTransaction{
			ID:     transactionID,
			Amount: 100000.0,
			ApprovalStatus: "pending",
		}

		mockCashTransactionDB.EXPECT().
			FindByID(transactionID.String()).
			Return(existingTransaction, nil).
			Times(1)

		mockCashTransactionDB.EXPECT().
			Update(gomock.Any()).
			Return(nil).
			Times(1)

		err := domain.ApproveTransaction(ctx, transactionID.String(), userID.String())

		assert.NoError(t, err)
	})

	t.Run("error - transaction not found", func(t *testing.T) {
		transactionID := uuid.New()
		userID := uuid.New()

		mockCashTransactionDB.EXPECT().
			FindByID(transactionID.String()).
			Return(nil, errors.New("record not found")).
			Times(1)

		err := domain.ApproveTransaction(ctx, transactionID.String(), userID.String())

		assert.Error(t, err)
	})
}

func TestGetBalance(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mock_outbound_port.NewMockDatabasePort(ctrl)
	mockCashTransactionDB := mock_outbound_port.NewMockCashTransactionDatabasePort(ctrl)
	mockCashCategoryDB := mock_outbound_port.NewMockCashCategoryDatabasePort(ctrl)

	mockDB.EXPECT().CashTransaction().Return(mockCashTransactionDB).AnyTimes()
	mockDB.EXPECT().CashCategory().Return(mockCashCategoryDB).AnyTimes()

	domain := NewCashDomain(mockDB)
	ctx := context.Background()

	t.Run("success - calculate balance", func(t *testing.T) {
		startDate := time.Now().Add(-30 * 24 * time.Hour)
		endDate := time.Now()

		incomeTotal := 250000.0
		expenseTotal := 50000.0

		mockCashTransactionDB.EXPECT().
			GetIncomeTotal(startDate, endDate).
			Return(incomeTotal, nil).
			Times(1)

		mockCashTransactionDB.EXPECT().
			GetExpenseTotal(startDate, endDate).
			Return(expenseTotal, nil).
			Times(1)

		balance, err := domain.GetBalance(ctx, startDate, endDate)

		assert.NoError(t, err)
		assert.Equal(t, 200000.0, balance)
	})
}
