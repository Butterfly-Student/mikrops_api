package outbound_port

import (
	"go-template/internal/model"
)

//go:generate mockgen -source=refund.go -destination=./../../../tests/mocks/port/mock_refund.go
type RefundDatabasePort interface {
	Create(input *model.RefundInput) (*model.Refund, error)
	FindByID(id string) (*model.Refund, error)
	FindByFilter(filter model.RefundFilter, lock bool) ([]model.Refund, error)
	Update(id string, input *model.RefundInput) (*model.Refund, error)
	Delete(id string) error
	Approve(id string, approvedBy string) (*model.Refund, error)
	Reject(id string, rejectedBy string, reason string) (*model.Refund, error)
	Process(id string, processedBy string) (*model.Refund, error)
	Complete(id string, processedBy string, xenditRefundID string) (*model.Refund, error)
	FindPendingRefunds() ([]model.Refund, error)
}
