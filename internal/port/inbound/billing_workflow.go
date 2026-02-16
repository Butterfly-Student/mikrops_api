package inbound_port

import (
	"context"
	"time"
)

type BillingWorkflowPort interface {
	// Workers
	StartInvoiceGenerationWorker(ctx context.Context)
	StartPaymentProcessingWorker(ctx context.Context)

	// Scheduler
	StartScheduler(ctx context.Context)

	// Workflow Triggers
	TriggerGenerateMonthlyInvoices(ctx context.Context, year, month int) error
	TriggerProcessPayment(ctx context.Context, input ProcessPaymentWorkflowInput) error
	TriggerCheckOverdue(ctx context.Context, applyLateFees bool) error
	TriggerSendReminders(ctx context.Context, daysBefore, daysAfter []int) error
}

// ProcessPaymentWorkflowInput is the input for payment processing workflow
type ProcessPaymentWorkflowInput struct {
	InvoiceID      string
	Amount         float64
	Status         string
	PaymentMethod  string
	TransactionRef string
	PaymentDate    time.Time
}
