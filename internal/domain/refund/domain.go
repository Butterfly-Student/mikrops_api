package refund

import (
	"context"
	"fmt"

	"go-template/internal/model"
	outbound_port "go-template/internal/port/outbound"
	"go-template/utils/log"
)

type RefundDomain interface {
	CreateRefund(ctx context.Context, input model.RefundInput) (*model.Refund, error)
	GetRefund(ctx context.Context, id string) (*model.Refund, error)
	ListRefunds(ctx context.Context, filter model.RefundFilter) ([]model.Refund, error)
	UpdateRefund(ctx context.Context, id string, input model.RefundInput) (*model.Refund, error)
	DeleteRefund(ctx context.Context, id string) error
	ApproveRefund(ctx context.Context, id string, approvedBy string) (*model.Refund, error)
	RejectRefund(ctx context.Context, id string, rejectedBy string, reason string) (*model.Refund, error)
	ProcessRefund(ctx context.Context, id string, processedBy string) (*model.Refund, error)
	CompleteRefund(ctx context.Context, id string, processedBy string, xenditRefundID string) (*model.Refund, error)
	GetPendingRefunds(ctx context.Context) ([]model.Refund, error)
}

type domain struct {
	dbPort outbound_port.RefundDatabasePort
}

func NewRefundDomain(dbPort outbound_port.RefundDatabasePort) RefundDomain {
	return &domain{
		dbPort: dbPort,
	}
}

func (d *domain) CreateRefund(ctx context.Context, input model.RefundInput) (*model.Refund, error) {
	log.WithContext(ctx).Info("creating refund")

	refund, err := d.dbPort.Create(&input)
	if err != nil {
		return nil, fmt.Errorf("failed to create refund: %w", err)
	}

	return refund, nil
}

func (d *domain) GetRefund(ctx context.Context, id string) (*model.Refund, error) {
	log.WithContext(ctx).Info(fmt.Sprintf("getting refund: %s", id))

	refund, err := d.dbPort.FindByID(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get refund: %w", err)
	}

	return refund, nil
}

func (d *domain) ListRefunds(ctx context.Context, filter model.RefundFilter) ([]model.Refund, error) {
	log.WithContext(ctx).Info("listing refunds")

	refunds, err := d.dbPort.FindByFilter(filter, false)
	if err != nil {
		return nil, fmt.Errorf("failed to list refunds: %w", err)
	}

	return refunds, nil
}

func (d *domain) UpdateRefund(ctx context.Context, id string, input model.RefundInput) (*model.Refund, error) {
	log.WithContext(ctx).Info(fmt.Sprintf("updating refund: %s", id))

	refund, err := d.dbPort.Update(id, &input)
	if err != nil {
		return nil, fmt.Errorf("failed to update refund: %w", err)
	}

	return refund, nil
}

func (d *domain) DeleteRefund(ctx context.Context, id string) error {
	log.WithContext(ctx).Info(fmt.Sprintf("deleting refund: %s", id))

	if err := d.dbPort.Delete(id); err != nil {
		return fmt.Errorf("failed to delete refund: %w", err)
	}

	return nil
}

func (d *domain) ApproveRefund(ctx context.Context, id string, approvedBy string) (*model.Refund, error) {
	log.WithContext(ctx).Info(fmt.Sprintf("approving refund: %s", id))

	refund, err := d.dbPort.Approve(id, approvedBy)
	if err != nil {
		return nil, fmt.Errorf("failed to approve refund: %w", err)
	}

	return refund, nil
}

func (d *domain) RejectRefund(ctx context.Context, id string, rejectedBy string, reason string) (*model.Refund, error) {
	log.WithContext(ctx).Info(fmt.Sprintf("rejecting refund: %s", id))

	refund, err := d.dbPort.Reject(id, rejectedBy, reason)
	if err != nil {
		return nil, fmt.Errorf("failed to reject refund: %w", err)
	}

	return refund, nil
}

func (d *domain) ProcessRefund(ctx context.Context, id string, processedBy string) (*model.Refund, error) {
	log.WithContext(ctx).Info(fmt.Sprintf("processing refund: %s", id))

	refund, err := d.dbPort.Process(id, processedBy)
	if err != nil {
		return nil, fmt.Errorf("failed to process refund: %w", err)
	}

	return refund, nil
}

func (d *domain) CompleteRefund(ctx context.Context, id string, processedBy string, xenditRefundID string) (*model.Refund, error) {
	log.WithContext(ctx).Info(fmt.Sprintf("completing refund: %s", id))

	refund, err := d.dbPort.Complete(id, processedBy, xenditRefundID)
	if err != nil {
		return nil, fmt.Errorf("failed to complete refund: %w", err)
	}

	return refund, nil
}

func (d *domain) GetPendingRefunds(ctx context.Context) ([]model.Refund, error) {
	log.WithContext(ctx).Info("getting pending refunds")

	refunds, err := d.dbPort.FindPendingRefunds()
	if err != nil {
		return nil, fmt.Errorf("failed to get pending refunds: %w", err)
	}

	return refunds, nil
}
