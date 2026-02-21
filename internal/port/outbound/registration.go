package outbound_port

import (
	"context"

	"go-template/internal/model"
)

type RegistrationDatabasePort interface {
	Create(ctx context.Context, registration *model.CustomerRegistration) error
	FindByID(ctx context.Context, id string) (*model.CustomerRegistration, error)
	FindAll(ctx context.Context, filter *model.RegistrationFilter) ([]model.CustomerRegistration, error)
	SetApproved(ctx context.Context, id string, approverID uint, customerID string) error
	SetRejected(ctx context.Context, id string, approverID uint, reason string) error
	// ListPppSecretNames returns all non-rejected ppp_secret_name values (for uniqueness checking)
	ListPppSecretNames(ctx context.Context) ([]string, error)
}
