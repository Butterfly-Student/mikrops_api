package payment

import (
	"context"
	"time"

	"github.com/palantir/stacktrace"

	"mikrops/internal/model"
	outbound_port "mikrops/internal/port/outbound"
)

type PaymentDomain interface {
	Create(ctx context.Context, input model.PaymentInput) (model.Payment, error)
	FindByFilter(ctx context.Context, filter model.PaymentFilter) ([]model.Payment, error)
	FindByID(ctx context.Context, id string) (model.Payment, error)
	Update(ctx context.Context, id string, input model.PaymentInput) error
	Delete(ctx context.Context, id string) error
	Verify(ctx context.Context, id string, staffID string) error
	Reject(ctx context.Context, id string, staffID string, notes string) error
}

type paymentDomain struct {
	databasePort outbound_port.DatabasePort
	messagePort  outbound_port.MessagePort
	cachePort    outbound_port.CachePort
	workflowPort outbound_port.WorkflowPort
}

func NewPaymentDomain(
	databasePort outbound_port.DatabasePort,
	messagePort outbound_port.MessagePort,
	cachePort outbound_port.CachePort,
	workflowPort outbound_port.WorkflowPort,
) PaymentDomain {
	return &paymentDomain{
		databasePort: databasePort,
		messagePort:  messagePort,
		cachePort:    cachePort,
		workflowPort: workflowPort,
	}
}

func (d *paymentDomain) Create(ctx context.Context, input model.PaymentInput) (model.Payment, error) {
	// Set default status to pending
	if input.Status == "" {
		input.Status = model.PaymentStatusPending
	}

	payment, err := d.databasePort.Payment().Create(input)
	if err != nil {
		return model.Payment{}, stacktrace.Propagate(err, "failed to create payment")
	}

	return payment, nil
}

func (d *paymentDomain) FindByFilter(ctx context.Context, filter model.PaymentFilter) ([]model.Payment, error) {
	if filter.IsEmpty() {
		return nil, stacktrace.NewError("filter is empty")
	}

	payments, err := d.databasePort.Payment().FindByFilter(filter)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to find payments by filter")
	}

	return payments, nil
}

func (d *paymentDomain) FindByID(ctx context.Context, id string) (model.Payment, error) {
	if id == "" {
		return model.Payment{}, stacktrace.NewError("id is empty")
	}

	payment, err := d.databasePort.Payment().FindByID(id)
	if err != nil {
		return model.Payment{}, stacktrace.Propagate(err, "failed to find payment by id")
	}

	return payment, nil
}

func (d *paymentDomain) Update(ctx context.Context, id string, input model.PaymentInput) error {
	if id == "" {
		return stacktrace.NewError("id is empty")
	}

	err := d.databasePort.Payment().Update(id, input)
	if err != nil {
		return stacktrace.Propagate(err, "failed to update payment")
	}

	return nil
}

func (d *paymentDomain) Delete(ctx context.Context, id string) error {
	if id == "" {
		return stacktrace.NewError("id is empty")
	}

	err := d.databasePort.Payment().Delete(id)
	if err != nil {
		return stacktrace.Propagate(err, "failed to delete payment")
	}

	return nil
}

func (d *paymentDomain) Verify(ctx context.Context, id string, staffID string) error {
	if id == "" {
		return stacktrace.NewError("id is empty")
	}
	if staffID == "" {
		return stacktrace.NewError("staffID is empty")
	}

	// Use transaction to ensure consistency
	_, err := d.databasePort.DoInTransaction(func(txRepo outbound_port.DatabasePort) (interface{}, error) {
		// Get payment
		payment, err := txRepo.Payment().FindByID(id)
		if err != nil {
			return nil, stacktrace.Propagate(err, "failed to find payment")
		}

		// Update payment status
		now := time.Now()
		payment.Status = model.PaymentStatusVerified
		payment.VerifiedBy = &staffID
		payment.VerifiedAt = &now
		err = txRepo.Payment().Update(id, payment.PaymentInput)
		if err != nil {
			return nil, stacktrace.Propagate(err, "failed to update payment")
		}

		// Update invoice status to paid
		invoice, err := txRepo.Invoice().FindByID(payment.InvoiceID)
		if err != nil {
			return nil, stacktrace.Propagate(err, "failed to find invoice")
		}

		invoice.Status = model.InvoiceStatusPaid
		invoice.PaidAt = &now
		err = txRepo.Invoice().Update(payment.InvoiceID, invoice.InvoiceInput)
		if err != nil {
			return nil, stacktrace.Propagate(err, "failed to update invoice")
		}

		return nil, nil
	})

	if err != nil {
		return stacktrace.Propagate(err, "failed to verify payment")
	}

	// Trigger cutoff domain to check if customer needs restoration
	// For now,we're just logging this event. The cutoff consumer will handle it.
	// TODO: Implement messagePort.Publish when message broker is configured
	// message := map[string]interface{}{
	// 	"invoice_id": invoiceID,
	// 	"customer_id": invoice.CustomerID,
	// 	"tenant_id": invoice.TenantID,
	// }
	// _ = d.messagePort.Publish(ctx, "invoice.paid", message)
	return nil
}

func (d *paymentDomain) Reject(ctx context.Context, id string, staffID string, notes string) error {
	if id == "" {
		return stacktrace.NewError("id is empty")
	}
	if staffID == "" {
		return stacktrace.NewError("staffID is empty")
	}

	// Get payment
	payment, err := d.databasePort.Payment().FindByID(id)
	if err != nil {
		return stacktrace.Propagate(err, "failed to find payment")
	}

	// Update payment status
	now := time.Now()
	payment.Status = model.PaymentStatusRejected
	payment.VerifiedBy = &staffID
	payment.VerifiedAt = &now
	if notes != "" {
		payment.Notes = notes
	}

	err = d.databasePort.Payment().Update(id, payment.PaymentInput)
	if err != nil {
		return stacktrace.Propagate(err, "failed to reject payment")
	}

	return nil
}
