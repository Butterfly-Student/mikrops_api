package refund

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"go-template/internal/model"
	mock_outbound_port "go-template/tests/mocks/port"
)

func TestCreateRefund(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRefundDB := mock_outbound_port.NewMockRefundDatabasePort(ctrl)

	domain := NewRefundDomain(mockRefundDB)
	ctx := context.Background()

	t.Run("success - create refund", func(t *testing.T) {
		paymentID := uuid.New()
		refundAmount := 100000.0
		refundType := "partial"
		refundReason := "Customer request"
		refundMethod := "bank_transfer"

		input := model.RefundInput{
			PaymentID:    paymentID,
			RefundAmount: &refundAmount,
			RefundType:   &refundType,
			RefundReason: &refundReason,
			RefundMethod: &refundMethod,
		}

		expectedRefund := &model.Refund{
			ID:          uuid.New(),
			PaymentID:   paymentID,
			RefundAmount: refundAmount,
			RefundType:   refundType,
			RefundReason: &refundReason,
			RefundMethod: &refundMethod,
			Status:      "pending",
		}

		mockRefundDB.EXPECT().
			Create(&input).
			Return(expectedRefund, nil).
			Times(1)

		result, err := domain.CreateRefund(ctx, input)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, expectedRefund.ID, result.ID)
	})

	t.Run("error - database error", func(t *testing.T) {
		paymentID := uuid.New()
		refundAmount := 100000.0
		refundType := "partial"

		input := model.RefundInput{
			PaymentID:    paymentID,
			RefundAmount: &refundAmount,
			RefundType:   &refundType,
		}

		mockRefundDB.EXPECT().
			Create(&input).
			Return(nil, errors.New("database error")).
			Times(1)

		result, err := domain.CreateRefund(ctx, input)

		assert.Error(t, err)
		assert.Nil(t, result)
	})
}

func TestGetRefund(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRefundDB := mock_outbound_port.NewMockRefundDatabasePort(ctrl)

	domain := NewRefundDomain(mockRefundDB)
	ctx := context.Background()

	t.Run("success - get refund", func(t *testing.T) {
		refundID := uuid.New()

		expectedRefund := &model.Refund{
			ID:          refundID,
			PaymentID:   uuid.New(),
			RefundAmount: 100000.0,
			Status:      "pending",
		}

		mockRefundDB.EXPECT().
			FindByID(refundID.String()).
			Return(expectedRefund, nil).
			Times(1)

		result, err := domain.GetRefund(ctx, refundID.String())

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, refundID, result.ID)
	})

	t.Run("error - not found", func(t *testing.T) {
		refundID := uuid.New()

		mockRefundDB.EXPECT().
			FindByID(refundID.String()).
			Return(nil, errors.New("refund not found")).
			Times(1)

		result, err := domain.GetRefund(ctx, refundID.String())

		assert.Error(t, err)
		assert.Nil(t, result)
	})
}

func TestListRefunds(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRefundDB := mock_outbound_port.NewMockRefundDatabasePort(ctrl)

	domain := NewRefundDomain(mockRefundDB)
	ctx := context.Background()

	t.Run("success - list refunds", func(t *testing.T) {
		filter := model.RefundFilter{}

		refunds := []model.Refund{
			{
				ID:          uuid.New(),
				PaymentID:   uuid.New(),
				RefundAmount: 100000.0,
				Status:      "pending",
			},
			{
				ID:          uuid.New(),
				PaymentID:   uuid.New(),
				RefundAmount: 50000.0,
				Status:      "approved",
			},
		}

		mockRefundDB.EXPECT().
			FindByFilter(filter, false).
			Return(refunds, nil).
			Times(1)

		result, err := domain.ListRefunds(ctx, filter)

		assert.NoError(t, err)
		assert.Len(t, result, 2)
	})

	t.Run("error - database error", func(t *testing.T) {
		filter := model.RefundFilter{}

		mockRefundDB.EXPECT().
			FindByFilter(filter, false).
			Return(nil, errors.New("database error")).
			Times(1)

		result, err := domain.ListRefunds(ctx, filter)

		assert.Error(t, err)
		assert.Nil(t, result)
	})
}

func TestUpdateRefund(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRefundDB := mock_outbound_port.NewMockRefundDatabasePort(ctrl)

	domain := NewRefundDomain(mockRefundDB)
	ctx := context.Background()

	t.Run("success - update refund", func(t *testing.T) {
		refundID := uuid.New()
		refundReason := "Updated reason"

		input := model.RefundInput{
			RefundReason: &refundReason,
		}

		updatedRefund := &model.Refund{
			ID:          refundID,
			PaymentID:   uuid.New(),
			RefundAmount: 100000.0,
			RefundReason: refundReason,
			Status:      "pending",
		}

		mockRefundDB.EXPECT().
			Update(refundID.String(), &input).
			Return(updatedRefund, nil).
			Times(1)

		result, err := domain.UpdateRefund(ctx, refundID.String(), input)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, refundReason, result.RefundReason)
	})

	t.Run("error - not found", func(t *testing.T) {
		refundID := uuid.New()
		refundReason := "Updated reason"

		input := model.RefundInput{
			RefundReason: &refundReason,
		}

		mockRefundDB.EXPECT().
			Update(refundID.String(), &input).
			Return(nil, errors.New("refund not found")).
			Times(1)

		result, err := domain.UpdateRefund(ctx, refundID.String(), input)

		assert.Error(t, err)
		assert.Nil(t, result)
	})
}

func TestDeleteRefund(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRefundDB := mock_outbound_port.NewMockRefundDatabasePort(ctrl)

	domain := NewRefundDomain(mockRefundDB)
	ctx := context.Background()

	t.Run("success - delete refund", func(t *testing.T) {
		refundID := uuid.New()

		mockRefundDB.EXPECT().
			Delete(refundID.String()).
			Return(nil).
			Times(1)

		err := domain.DeleteRefund(ctx, refundID.String())

		assert.NoError(t, err)
	})

	t.Run("error - database error", func(t *testing.T) {
		refundID := uuid.New()

		mockRefundDB.EXPECT().
			Delete(refundID.String()).
			Return(errors.New("database error")).
			Times(1)

		err := domain.DeleteRefund(ctx, refundID.String())

		assert.Error(t, err)
	})
}

func TestApproveRefund(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRefundDB := mock_outbound_port.NewMockRefundDatabasePort(ctrl)

	domain := NewRefundDomain(mockRefundDB)
	ctx := context.Background()

	t.Run("success - approve refund", func(t *testing.T) {
		refundID := uuid.New()
		approvedBy := "admin@example.com"
		approvedByUUID := uuid.New()

		approvedRefund := &model.Refund{
			ID:          refundID,
			PaymentID:   uuid.New(),
			RefundAmount: 100000.0,
			Status:      "approved",
			ApprovedBy:  &approvedByUUID,
		}

		mockRefundDB.EXPECT().
			Approve(refundID.String(), approvedBy).
			Return(approvedRefund, nil).
			Times(1)

		result, err := domain.ApproveRefund(ctx, refundID.String(), approvedBy)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "approved", result.Status)
		assert.Equal(t, approvedByUUID, *result.ApprovedBy)
	})

	t.Run("error - not found", func(t *testing.T) {
		refundID := uuid.New()
		approvedBy := "admin@example.com"

		mockRefundDB.EXPECT().
			Approve(refundID.String(), approvedBy).
			Return(nil, errors.New("refund not found")).
			Times(1)

		result, err := domain.ApproveRefund(ctx, refundID.String(), approvedBy)

		assert.Error(t, err)
		assert.Nil(t, result)
	})
}

func TestRejectRefund(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRefundDB := mock_outbound_port.NewMockRefundDatabasePort(ctrl)

	domain := NewRefundDomain(mockRefundDB)
	ctx := context.Background()

	t.Run("success - reject refund", func(t *testing.T) {
		refundID := uuid.New()
		rejectedBy := "admin@example.com"
		rejectedByUUID := uuid.New()
		reason := "Duplicate request"

		rejectedRefund := &model.Refund{
			ID:              refundID,
			PaymentID:       uuid.New(),
			RefundAmount:    100000.0,
			Status:          "rejected",
			ApprovedBy:      &rejectedByUUID,
			RejectionReason: &reason,
		}

		mockRefundDB.EXPECT().
			Reject(refundID.String(), rejectedBy, reason).
			Return(rejectedRefund, nil).
			Times(1)

		result, err := domain.RejectRefund(ctx, refundID.String(), rejectedBy, reason)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "rejected", result.Status)
		assert.Equal(t, rejectedByUUID, *result.ApprovedBy)
		assert.Equal(t, reason, *result.RejectionReason)
	})

	t.Run("error - not found", func(t *testing.T) {
		refundID := uuid.New()
		rejectedBy := "admin@example.com"
		reason := "Not found"

		mockRefundDB.EXPECT().
			Reject(refundID.String(), rejectedBy, reason).
			Return(nil, errors.New("refund not found")).
			Times(1)

		result, err := domain.RejectRefund(ctx, refundID.String(), rejectedBy, reason)

		assert.Error(t, err)
		assert.Nil(t, result)
	})
}

func TestProcessRefund(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRefundDB := mock_outbound_port.NewMockRefundDatabasePort(ctrl)

	domain := NewRefundDomain(mockRefundDB)
	ctx := context.Background()

	t.Run("success - process refund", func(t *testing.T) {
		refundID := uuid.New()
		processedBy := "admin@example.com"
		processedByUUID := uuid.New()

		processedRefund := &model.Refund{
			ID:          refundID,
			PaymentID:   uuid.New(),
			RefundAmount: 100000.0,
			Status:      "processing",
			ProcessedBy: &processedByUUID,
		}

		mockRefundDB.EXPECT().
			Process(refundID.String(), processedBy).
			Return(processedRefund, nil).
			Times(1)

		result, err := domain.ProcessRefund(ctx, refundID.String(), processedBy)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "processing", result.Status)
		assert.Equal(t, processedByUUID, *result.ProcessedBy)
	})

	t.Run("error - not found", func(t *testing.T) {
		refundID := uuid.New()
		processedBy := "admin@example.com"

		mockRefundDB.EXPECT().
			Process(refundID.String(), processedBy).
			Return(nil, errors.New("refund not found")).
			Times(1)

		result, err := domain.ProcessRefund(ctx, refundID.String(), processedBy)

		assert.Error(t, err)
		assert.Nil(t, result)
	})
}

func TestCompleteRefund(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRefundDB := mock_outbound_port.NewMockRefundDatabasePort(ctrl)

	domain := NewRefundDomain(mockRefundDB)
	ctx := context.Background()

	t.Run("success - complete refund", func(t *testing.T) {
		refundID := uuid.New()
		processedBy := "admin@example.com"
		processedByUUID := uuid.New()
		xenditRefundID := "xendit-refund-123"

		completedRefund := &model.Refund{
			ID:            refundID,
			PaymentID:     uuid.New(),
			RefundAmount:  100000.0,
			Status:        "completed",
			ProcessedBy:   &processedByUUID,
			XenditRefundID: &xenditRefundID,
		}

		mockRefundDB.EXPECT().
			Complete(refundID.String(), processedBy, xenditRefundID).
			Return(completedRefund, nil).
			Times(1)

		result, err := domain.CompleteRefund(ctx, refundID.String(), processedBy, xenditRefundID)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "completed", result.Status)
		assert.Equal(t, processedByUUID, *result.ProcessedBy)
		assert.Equal(t, xenditRefundID, *result.XenditRefundID)
	})

	t.Run("error - not found", func(t *testing.T) {
		refundID := uuid.New()
		processedBy := "admin@example.com"
		xenditRefundID := "xendit-refund-123"

		mockRefundDB.EXPECT().
			Complete(refundID.String(), processedBy, xenditRefundID).
			Return(nil, errors.New("refund not found")).
			Times(1)

		result, err := domain.CompleteRefund(ctx, refundID.String(), processedBy, xenditRefundID)

		assert.Error(t, err)
		assert.Nil(t, result)
	})
}

func TestGetPendingRefunds(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRefundDB := mock_outbound_port.NewMockRefundDatabasePort(ctrl)

	domain := NewRefundDomain(mockRefundDB)
	ctx := context.Background()

	t.Run("success - get pending refunds", func(t *testing.T) {
		refunds := []model.Refund{
			{
				ID:          uuid.New(),
				PaymentID:   uuid.New(),
				RefundAmount: 100000.0,
				Status:      "pending",
			},
			{
				ID:          uuid.New(),
				PaymentID:   uuid.New(),
				RefundAmount: 50000.0,
				Status:      "pending",
			},
		}

		mockRefundDB.EXPECT().
			FindPendingRefunds().
			Return(refunds, nil).
			Times(1)

		result, err := domain.GetPendingRefunds(ctx)

		assert.NoError(t, err)
		assert.Len(t, result, 2)
		for _, refund := range result {
			assert.Equal(t, "pending", refund.Status)
		}
	})

	t.Run("error - database error", func(t *testing.T) {
		mockRefundDB.EXPECT().
			FindPendingRefunds().
			Return(nil, errors.New("database error")).
			Times(1)

		result, err := domain.GetPendingRefunds(ctx)

		assert.Error(t, err)
		assert.Nil(t, result)
	})
}
