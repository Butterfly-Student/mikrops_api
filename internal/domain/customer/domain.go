package customer

import (
	"context"

	"github.com/palantir/stacktrace"
	"golang.org/x/crypto/bcrypt"

	"mikrops/internal/model"
	outbound_port "mikrops/internal/port/outbound"
)

type CustomerDomain interface {
	Create(ctx context.Context, input model.CustomerInput) (model.Customer, error)
	FindByFilter(ctx context.Context, filter model.CustomerFilter) ([]model.Customer, error)
	FindByID(ctx context.Context, id string) (model.Customer, error)
	FindByUsername(ctx context.Context, username string) (model.Customer, error)
	Update(ctx context.Context, id string, input model.CustomerInput) error
	Delete(ctx context.Context, id string) error
}

type customerDomain struct {
	databasePort outbound_port.DatabasePort
	messagePort  outbound_port.MessagePort
	cachePort    outbound_port.CachePort
	workflowPort outbound_port.WorkflowPort
}

func NewCustomerDomain(
	databasePort outbound_port.DatabasePort,
	messagePort outbound_port.MessagePort,
	cachePort outbound_port.CachePort,
	workflowPort outbound_port.WorkflowPort,
) CustomerDomain {
	return &customerDomain{
		databasePort: databasePort,
		messagePort:  messagePort,
		cachePort:    cachePort,
		workflowPort: workflowPort,
	}
}

func (d *customerDomain) Create(ctx context.Context, input model.CustomerInput) (model.Customer, error) {
	// Hash password if provided
	if input.Password != "" {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
		if err != nil {
			return model.Customer{}, stacktrace.Propagate(err, "failed to hash password")
		}
		input.PasswordHash = string(hashedPassword)
		input.Password = "" // Clear plain password
	}

	customer, err := d.databasePort.Customer().Create(input)
	if err != nil {
		return model.Customer{}, stacktrace.Propagate(err, "failed to create customer")
	}

	return customer, nil
}

func (d *customerDomain) FindByFilter(ctx context.Context, filter model.CustomerFilter) ([]model.Customer, error) {
	if filter.IsEmpty() {
		return nil, stacktrace.NewError("filter is empty")
	}

	customers, err := d.databasePort.Customer().FindByFilter(filter)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to find customers by filter")
	}

	return customers, nil
}

func (d *customerDomain) FindByID(ctx context.Context, id string) (model.Customer, error) {
	if id == "" {
		return model.Customer{}, stacktrace.NewError("id is empty")
	}

	customer, err := d.databasePort.Customer().FindByID(id)
	if err != nil {
		return model.Customer{}, stacktrace.Propagate(err, "failed to find customer by id")
	}

	return customer, nil
}

func (d *customerDomain) FindByUsername(ctx context.Context, username string) (model.Customer, error) {
	if username == "" {
		return model.Customer{}, stacktrace.NewError("username is empty")
	}

	customer, err := d.databasePort.Customer().FindByUsername(username)
	if err != nil {
		return model.Customer{}, stacktrace.Propagate(err, "failed to find customer by username")
	}

	return customer, nil
}

func (d *customerDomain) Update(ctx context.Context, id string, input model.CustomerInput) error {
	if id == "" {
		return stacktrace.NewError("id is empty")
	}

	// Hash password if changed
	if input.Password != "" {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
		if err != nil {
			return stacktrace.Propagate(err, "failed to hash password")
		}
		input.PasswordHash = string(hashedPassword)
		input.Password = "" // Clear plain password
	}

	err := d.databasePort.Customer().Update(id, input)
	if err != nil {
		return stacktrace.Propagate(err, "failed to update customer")
	}

	return nil
}

func (d *customerDomain) Delete(ctx context.Context, id string) error {
	if id == "" {
		return stacktrace.NewError("id is empty")
	}

	err := d.databasePort.Customer().Delete(id)
	if err != nil {
		return stacktrace.Propagate(err, "failed to delete customer")
	}

	return nil
}
