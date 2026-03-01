package outbound_port

//go:generate mockgen -source=registration.go -destination=./../../../tests/mocks/port/mock_registration.go

import (
	"context"

	"go-template/internal/model"
)

type RegistrationDatabasePort interface {
	Create(ctx context.Context, registration *model.CustomerRegistration) error
	FindByID(ctx context.Context, id string) (*model.CustomerRegistration, error)
	FindAll(ctx context.Context, filter *model.RegistrationFilter) ([]model.CustomerRegistration, error)
	SetApproved(ctx context.Context, id string, approverID string, customerID string) error
	SetRejected(ctx context.Context, id string, approverID string, reason string) error
}
