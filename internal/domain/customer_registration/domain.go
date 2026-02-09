package customer_registration

import (
	"context"
	"time"

	"github.com/palantir/stacktrace"

	"mikrops/internal/model"
	outbound_port "mikrops/internal/port/outbound"
)

type CustomerRegistrationDomain interface {
	Create(ctx context.Context, input model.CustomerRegistrationInput) (model.CustomerRegistration, error)
	FindByFilter(ctx context.Context, filter model.CustomerRegistrationFilter) ([]model.CustomerRegistration, error)
	FindByID(ctx context.Context, id string) (model.CustomerRegistration, error)
	Approve(ctx context.Context, id string, staffID string) (model.CustomerRegistration, error)
	Reject(ctx context.Context, id string, reason string) (model.CustomerRegistration, error)
}

type customerRegistrationDomain struct {
	databasePort outbound_port.DatabasePort
	messagePort  outbound_port.MessagePort
	cachePort    outbound_port.CachePort
	workflowPort outbound_port.WorkflowPort
}

func NewCustomerRegistrationDomain(
	databasePort outbound_port.DatabasePort,
	messagePort outbound_port.MessagePort,
	cachePort outbound_port.CachePort,
	workflowPort outbound_port.WorkflowPort,
) CustomerRegistrationDomain {
	return &customerRegistrationDomain{
		databasePort: databasePort,
		messagePort:  messagePort,
		cachePort:    cachePort,
		workflowPort: workflowPort,
	}
}

func (d *customerRegistrationDomain) Create(ctx context.Context, input model.CustomerRegistrationInput) (model.CustomerRegistration, error) {
	if input.TenantID == "" {
		return model.CustomerRegistration{}, stacktrace.NewError("tenant_id is required")
	}
	if input.FullName == "" {
		return model.CustomerRegistration{}, stacktrace.NewError("full_name is required")
	}
	if input.RequestedPackageID == "" {
		return model.CustomerRegistration{}, stacktrace.NewError("requested_package_id is required")
	}

	input.Status = model.RegistrationStatusPending

	registration, err := d.databasePort.CustomerRegistration().Create(input)
	if err != nil {
		return model.CustomerRegistration{}, stacktrace.Propagate(err, "failed to create customer registration")
	}

	return registration, nil
}

func (d *customerRegistrationDomain) FindByFilter(ctx context.Context, filter model.CustomerRegistrationFilter) ([]model.CustomerRegistration, error) {
	if filter.IsEmpty() {
		return nil, stacktrace.NewError("filter is empty")
	}

	registrations, err := d.databasePort.CustomerRegistration().FindByFilter(filter)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to find customer registrations")
	}

	return registrations, nil
}

func (d *customerRegistrationDomain) FindByID(ctx context.Context, id string) (model.CustomerRegistration, error) {
	if id == "" {
		return model.CustomerRegistration{}, stacktrace.NewError("id is empty")
	}

	registration, err := d.databasePort.CustomerRegistration().FindByID(id)
	if err != nil {
		return model.CustomerRegistration{}, stacktrace.Propagate(err, "failed to find customer registration")
	}

	return registration, nil
}

func (d *customerRegistrationDomain) Approve(ctx context.Context, id string, staffID string) (model.CustomerRegistration, error) {
	if id == "" {
		return model.CustomerRegistration{}, stacktrace.NewError("id is empty")
	}

	registration, err := d.databasePort.CustomerRegistration().FindByID(id)
	if err != nil {
		return model.CustomerRegistration{}, stacktrace.Propagate(err, "failed to find registration")
	}

	if registration.Status != model.RegistrationStatusPending {
		return model.CustomerRegistration{}, stacktrace.NewError("registration is not in pending status")
	}

	// Create customer from registration data
	result, err := d.databasePort.DoInTransaction(func(repo outbound_port.DatabasePort) (interface{}, error) {
		customerInput := model.CustomerInput{
			TenantID: registration.TenantID,
			FullName: registration.FullName,
			Email:    registration.Email,
			Phone:    registration.Phone,
			Address:  registration.Address,
		}
		if registration.RequestedNasID != nil {
			customerInput.NasID = registration.RequestedNasID
		}

		customer, err := repo.Customer().Create(customerInput)
		if err != nil {
			return nil, stacktrace.Propagate(err, "failed to create customer from registration")
		}

		now := time.Now()
		updateInput := model.CustomerRegistrationInput{
			Status:     model.RegistrationStatusApproved,
			CustomerID: &customer.ID,
			ApprovedBy: &staffID,
			ApprovedAt: &now,
		}
		err = repo.CustomerRegistration().Update(id, updateInput)
		if err != nil {
			return nil, stacktrace.Propagate(err, "failed to update registration status")
		}

		return customer, nil
	})
	if err != nil {
		return model.CustomerRegistration{}, stacktrace.Propagate(err, "failed to approve registration")
	}

	_ = result

	updatedRegistration, err := d.databasePort.CustomerRegistration().FindByID(id)
	if err != nil {
		return model.CustomerRegistration{}, stacktrace.Propagate(err, "failed to fetch updated registration")
	}

	return updatedRegistration, nil
}

func (d *customerRegistrationDomain) Reject(ctx context.Context, id string, reason string) (model.CustomerRegistration, error) {
	if id == "" {
		return model.CustomerRegistration{}, stacktrace.NewError("id is empty")
	}

	registration, err := d.databasePort.CustomerRegistration().FindByID(id)
	if err != nil {
		return model.CustomerRegistration{}, stacktrace.Propagate(err, "failed to find registration")
	}

	if registration.Status != model.RegistrationStatusPending {
		return model.CustomerRegistration{}, stacktrace.NewError("registration is not in pending status")
	}

	updateInput := model.CustomerRegistrationInput{
		Status:          model.RegistrationStatusRejected,
		RejectionReason: reason,
	}
	err = d.databasePort.CustomerRegistration().Update(id, updateInput)
	if err != nil {
		return model.CustomerRegistration{}, stacktrace.Propagate(err, "failed to reject registration")
	}

	updatedRegistration, err := d.databasePort.CustomerRegistration().FindByID(id)
	if err != nil {
		return model.CustomerRegistration{}, stacktrace.Propagate(err, "failed to fetch updated registration")
	}

	return updatedRegistration, nil
}
