package inbound_port

import "context"

type IsolationWorkflowPort interface {
	// Worker
	StartIsolationWorker(ctx context.Context)

	// Scheduler
	StartScheduler(ctx context.Context)

	// Workflow Triggers
	TriggerIsolateExpiredCustomers(ctx context.Context, gracePeriodDays int) error
	TriggerReactivateCustomer(ctx context.Context, customerID string, invoiceID string) error
}

// IsolateExpiredCustomersInput is input for isolation workflow
type IsolateExpiredCustomersInput struct {
	GracePeriodDays int
}

// ReactivateCustomerInput is input for reactivation workflow
type ReactivateCustomerInput struct {
	CustomerID string
	InvoiceID  string
}
