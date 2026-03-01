package outbound_port

import (
	"context"

	"go-template/internal/model"
)

//go:generate mockgen -source=customer.go -destination=./../../../tests/mocks/port/mock_customer.go
type CustomerDatabasePort interface {
	Create(ctx context.Context, customer *model.Customer) error
	FindByID(ctx context.Context, id string) (*model.Customer, error)
	FindByCode(ctx context.Context, code string) (*model.Customer, error)
	// FindByPortalIdentifier finds a customer by exact customer_code or phone match (for portal login)
	FindByPortalIdentifier(ctx context.Context, identifier string) (*model.Customer, error)
	FindAll(ctx context.Context, filter *model.CustomerFilter) ([]model.Customer, error)
	Update(ctx context.Context, customer *model.Customer) error
	Delete(ctx context.Context, id string) error
	UpdateStatus(ctx context.Context, id string, status model.CustomerStatus) error
}
