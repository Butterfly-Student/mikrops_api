package postgres_outbound_adapter

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"go-template/internal/model"
	"go-template/utils/log"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type refundPostgresAdapter struct {
	db *gorm.DB
}

func NewRefundAdapter(db *gorm.DB) *refundPostgresAdapter {
	return &refundPostgresAdapter{db: db}
}

func (a *refundPostgresAdapter) Create(input *model.RefundInput) (*model.Refund, error) {
	ctx := context.Background()

	var payment model.Payment
	if err := a.db.WithContext(ctx).Where("id = ?", input.PaymentID).First(&payment).Error; err != nil {
		log.WithContext(ctx).Error(fmt.Sprintf("payment not found: %v", input.PaymentID), err)
		return nil, err
	}

	refund := &model.Refund{
		RefundNumber:      model.GenerateRefundNumber(),
		PaymentID:         input.PaymentID,
		InvoiceID:         input.InvoiceID,
		CustomerID:        payment.CustomerID,
		RefundAmount:      *input.RefundAmount,
		RefundType:        *input.RefundType,
		RefundReason:      input.RefundReason,
		RefundMethod:      input.RefundMethod,
		BankName:          input.BankName,
		BankAccountName:   input.BankAccountName,
		BankAccountNumber: input.BankAccountNumber,
		Status:            "pending",
		Notes:             input.Notes,
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}

	model.RefundPrepare(refund)

	if err := a.db.WithContext(ctx).Create(refund).Error; err != nil {
		log.WithContext(ctx).Error("failed to create refund", err)
		return nil, err
	}

	log.WithContext(ctx).Info(fmt.Sprintf("refund created: %s", refund.ID))
	return refund, nil
}

func (a *refundPostgresAdapter) FindByID(id string) (*model.Refund, error) {
	ctx := context.Background()

	var refund model.Refund
	if err := a.db.WithContext(ctx).Where("id = ?", id).First(&refund).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		log.WithContext(ctx).Error(fmt.Sprintf("failed to find refund: %s", id), err)
		return nil, err
	}

	return &refund, nil
}

func (a *refundPostgresAdapter) FindByFilter(filter model.RefundFilter, lock bool) ([]model.Refund, error) {
	ctx := context.Background()

	query := a.db.WithContext(ctx).Model(&model.Refund{})

	if len(filter.IDs) > 0 {
		query = query.Where("id IN ?", filter.IDs)
	}
	if len(filter.RefundNumbers) > 0 {
		query = query.Where("refund_number IN ?", filter.RefundNumbers)
	}
	if len(filter.PaymentIDs) > 0 {
		query = query.Where("payment_id IN ?", filter.PaymentIDs)
	}
	if len(filter.InvoiceIDs) > 0 {
		query = query.Where("invoice_id IN ?", filter.InvoiceIDs)
	}
	if len(filter.CustomerIDs) > 0 {
		query = query.Where("customer_id IN ?", filter.CustomerIDs)
	}
	if filter.CustomerID != nil {
		query = query.Where("customer_id = ?", *filter.CustomerID)
	}
	if len(filter.Status) > 0 {
		query = query.Where("status IN ?", filter.Status)
	}
	if len(filter.RefundType) > 0 {
		query = query.Where("refund_type IN ?", filter.RefundType)
	}
	if len(filter.RefundMethod) > 0 {
		query = query.Where("refund_method IN ?", filter.RefundMethod)
	}
	if filter.CreatedStart != nil {
		query = query.Where("created_at >= ?", *filter.CreatedStart)
	}
	if filter.CreatedEnd != nil {
		query = query.Where("created_at <= ?", *filter.CreatedEnd)
	}
	if filter.ApprovedStart != nil {
		query = query.Where("approved_at >= ?", *filter.ApprovedStart)
	}
	if filter.ApprovedEnd != nil {
		query = query.Where("approved_at <= ?", *filter.ApprovedEnd)
	}
	if filter.ProcessedStart != nil {
		query = query.Where("processed_at >= ?", *filter.ProcessedStart)
	}
	if filter.ProcessedEnd != nil {
		query = query.Where("processed_at <= ?", *filter.ProcessedEnd)
	}
	if filter.AmountMin != nil {
		query = query.Where("refund_amount >= ?", *filter.AmountMin)
	}
	if filter.AmountMax != nil {
		query = query.Where("refund_amount <= ?", *filter.AmountMax)
	}
	if filter.Search != nil {
		search := "%" + *filter.Search + "%"
		query = query.Where("refund_number ILIKE ? OR refund_reason ILIKE ? OR notes ILIKE ?", search, search, search)
	}

	query = query.Order("created_at DESC")

	if lock {
		query = query.Clauses(clause.Locking{Strength: "UPDATE"})
	}

	var refunds []model.Refund
	if err := query.Find(&refunds).Error; err != nil {
		log.WithContext(ctx).Error("failed to find refunds", err)
		return nil, err
	}

	return refunds, nil
}

func (a *refundPostgresAdapter) Update(id string, input *model.RefundInput) (*model.Refund, error) {
	ctx := context.Background()

	var refund model.Refund
	if err := a.db.WithContext(ctx).Where("id = ?", id).First(&refund).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		log.WithContext(ctx).Error(fmt.Sprintf("failed to find refund: %s", id), err)
		return nil, err
	}

	if input.RefundReason != nil {
		refund.RefundReason = *input.RefundReason
	}
	if input.RefundMethod != nil {
		refund.RefundMethod = *input.RefundMethod
	}
	if input.BankName != nil {
		refund.BankName = input.BankName
	}
	if input.BankAccountName != nil {
		refund.BankAccountName = input.BankAccountName
	}
	if input.BankAccountNumber != nil {
		refund.BankAccountNumber = input.BankAccountNumber
	}
	if input.Notes != nil {
		refund.Notes = input.Notes
	}

	refund.UpdatedAt = time.Now()

	if err := a.db.WithContext(ctx).Save(&refund).Error; err != nil {
		log.WithContext(ctx).Error(fmt.Sprintf("failed to update refund: %s", id), err)
		return nil, err
	}

	log.WithContext(ctx).Info(fmt.Sprintf("refund updated: %s", refund.ID))
	return &refund, nil
}

func (a *refundPostgresAdapter) Delete(id string) error {
	ctx := context.Background()

	if err := a.db.WithContext(ctx).Where("id = ?", id).Delete(&model.Refund{}).Error; err != nil {
		log.WithContext(ctx).Error(fmt.Sprintf("failed to delete refund: %s", id), err)
		return err
	}

	log.WithContext(ctx).Info(fmt.Sprintf("refund deleted: %s", id))
	return nil
}

func (a *refundPostgresAdapter) Approve(id string, approvedBy string) (*model.Refund, error) {
	ctx := context.Background()

	var refund model.Refund
	if err := a.db.WithContext(ctx).Where("id = ?", id).First(&refund).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	now := time.Now()
	approvedByID, err := strconv.ParseUint(approvedBy, 10, 64)
	if err != nil {
		log.WithContext(ctx).Error(fmt.Sprintf("invalid approved by id: %s", approvedBy), err)
		return nil, err
	}
	approvedByUint := uint(approvedByID)

	refund.Status = "approved"
	refund.ApprovedBy = &approvedByUint
	refund.ApprovedAt = &now
	refund.UpdatedAt = now

	if err := a.db.WithContext(ctx).Save(&refund).Error; err != nil {
		log.WithContext(ctx).Error(fmt.Sprintf("failed to approve refund: %s", id), err)
		return nil, err
	}

	log.WithContext(ctx).Info(fmt.Sprintf("refund approved: %s", id))
	return &refund, nil
}

func (a *refundPostgresAdapter) Reject(id string, rejectedBy string, reason string) (*model.Refund, error) {
	ctx := context.Background()

	var refund model.Refund
	if err := a.db.WithContext(ctx).Where("id = ?", id).First(&refund).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	now := time.Now()
	rejectedByID, err := strconv.ParseUint(rejectedBy, 10, 64)
	if err != nil {
		log.WithContext(ctx).Error(fmt.Sprintf("invalid rejected by id: %s", rejectedBy), err)
		return nil, err
	}
	rejectedByUint := uint(rejectedByID)

	refund.Status = "rejected"
	refund.ProcessedBy = &rejectedByUint
	refund.ProcessedAt = &now
	refund.RejectionReason = &reason
	refund.UpdatedAt = now

	if err := a.db.WithContext(ctx).Save(&refund).Error; err != nil {
		log.WithContext(ctx).Error(fmt.Sprintf("failed to reject refund: %s", id), err)
		return nil, err
	}

	log.WithContext(ctx).Info(fmt.Sprintf("refund rejected: %s", id))
	return &refund, nil
}

func (a *refundPostgresAdapter) Process(id string, processedBy string) (*model.Refund, error) {
	ctx := context.Background()

	var refund model.Refund
	if err := a.db.WithContext(ctx).Where("id = ?", id).First(&refund).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	now := time.Now()
	processedByID, err := strconv.ParseUint(processedBy, 10, 64)
	if err != nil {
		log.WithContext(ctx).Error(fmt.Sprintf("invalid processed by id: %s", processedBy), err)
		return nil, err
	}
	processedByUint := uint(processedByID)

	refund.Status = "processed"
	refund.ProcessedBy = &processedByUint
	refund.ProcessedAt = &now
	refund.UpdatedAt = now

	if err := a.db.WithContext(ctx).Save(&refund).Error; err != nil {
		log.WithContext(ctx).Error(fmt.Sprintf("failed to process refund: %s", id), err)
		return nil, err
	}

	log.WithContext(ctx).Info(fmt.Sprintf("refund processed: %s", id))
	return &refund, nil
}

func (a *refundPostgresAdapter) Complete(id string, processedBy string, xenditRefundID string) (*model.Refund, error) {
	ctx := context.Background()

	var refund model.Refund
	if err := a.db.WithContext(ctx).Where("id = ?", id).First(&refund).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	now := time.Now()
	completedByID, err := strconv.ParseUint(processedBy, 10, 64)
	if err != nil {
		log.WithContext(ctx).Error(fmt.Sprintf("invalid processed by id: %s", processedBy), err)
		return nil, err
	}
	completedByUint := uint(completedByID)

	refund.Status = "completed"
	refund.ProcessedBy = &completedByUint
	refund.ProcessedAt = &now
	refund.XenditRefundID = &xenditRefundID
	refund.UpdatedAt = now

	if err := a.db.WithContext(ctx).Save(&refund).Error; err != nil {
		log.WithContext(ctx).Error(fmt.Sprintf("failed to complete refund: %s", id), err)
		return nil, err
	}

	log.WithContext(ctx).Info(fmt.Sprintf("refund completed: %s", id))
	return &refund, nil
}

func (a *refundPostgresAdapter) FindPendingRefunds() ([]model.Refund, error) {
	ctx := context.Background()

	var refunds []model.Refund
	if err := a.db.WithContext(ctx).Where("status = ?", "pending").Order("created_at ASC").Find(&refunds).Error; err != nil {
		log.WithContext(ctx).Error("failed to find pending refunds", err)
		return nil, err
	}

	return refunds, nil
}
