package payment_method

import (
	"context"

	"github.com/palantir/stacktrace"

	"mikrops/internal/model"
	outbound_port "mikrops/internal/port/outbound"
)

type PaymentMethodDomain interface {
	Create(ctx context.Context, input model.PaymentMethodInput) (model.PaymentMethod, error)
	FindByFilter(ctx context.Context, filter model.PaymentMethodFilter) ([]model.PaymentMethod, error)
	FindByID(ctx context.Context, id string) (model.PaymentMethod, error)
	Update(ctx context.Context, id string, input model.PaymentMethodInput) error
	Delete(ctx context.Context, id string) error
}

type paymentMethodDomain struct {
	databasePort outbound_port.DatabasePort
	messagePort  outbound_port.MessagePort
	cachePort    outbound_port.CachePort
	workflowPort outbound_port.WorkflowPort
}

func NewPaymentMethodDomain(
	databasePort outbound_port.DatabasePort,
	messagePort outbound_port.MessagePort,
	cachePort outbound_port.CachePort,
	workflowPort outbound_port.WorkflowPort,
) PaymentMethodDomain {
	return &paymentMethodDomain{
		databasePort: databasePort,
		messagePort:  messagePort,
		cachePort:    cachePort,
		workflowPort: workflowPort,
	}
}

func (d *paymentMethodDomain) Create(ctx context.Context, input model.PaymentMethodInput) (model.PaymentMethod, error) {
	method, err := d.databasePort.PaymentMethod().Create(input)
	if err != nil {
		return model.PaymentMethod{}, stacktrace.Propagate(err, "failed to create payment method")
	}

	return method, nil
}

func (d *paymentMethodDomain) FindByFilter(ctx context.Context, filter model.PaymentMethodFilter) ([]model.PaymentMethod, error) {
	if filter.IsEmpty() {
		return nil, stacktrace.NewError("filter is empty")
	}

	methods, err := d.databasePort.PaymentMethod().FindByFilter(filter)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to find payment methods by filter")
	}

	return methods, nil
}

func (d *paymentMethodDomain) FindByID(ctx context.Context, id string) (model.PaymentMethod, error) {
	if id == "" {
		return model.PaymentMethod{}, stacktrace.NewError("id is empty")
	}

	method, err := d.databasePort.PaymentMethod().FindByID(id)
	if err != nil {
		return model.PaymentMethod{}, stacktrace.Propagate(err, "failed to find payment method by id")
	}

	return method, nil
}

func (d *paymentMethodDomain) Update(ctx context.Context, id string, input model.PaymentMethodInput) error {
	if id == "" {
		return stacktrace.NewError("id is empty")
	}

	err := d.databasePort.PaymentMethod().Update(id, input)
	if err != nil {
		return stacktrace.Propagate(err, "failed to update payment method")
	}

	return nil
}

func (d *paymentMethodDomain) Delete(ctx context.Context, id string) error {
	if id == "" {
		return stacktrace.NewError("id is empty")
	}

	err := d.databasePort.PaymentMethod().Delete(id)
	if err != nil {
		return stacktrace.Propagate(err, "failed to delete payment method")
	}

	return nil
}
