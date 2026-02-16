package isolation_workflow

import (
	"context"
	"fmt"
	"os"
	"time"

	"go-template/internal/domain"
	inbound_port "go-template/internal/port/inbound"
	"go-template/utils/log"
	"go.temporal.io/sdk/client"
)

type isolationAdapter struct {
	domain domain.Domain
}

func NewIsolationAdapter(domain domain.Domain) inbound_port.IsolationWorkflowPort {
	return &isolationAdapter{domain: domain}
}

func (a *isolationAdapter) StartIsolationWorker(ctx context.Context) {
	go StartIsolationWorker(ctx, a.domain)
}

// StartScheduler starts the isolation scheduler
func (a *isolationAdapter) StartScheduler(ctx context.Context) {
	scheduler := NewIsolationScheduler(a.domain)
	scheduler.Start(ctx)
}

func (a *isolationAdapter) TriggerIsolateExpiredCustomers(ctx context.Context, gracePeriodDays int) error {
	c, err := a.getTemporalClient()
	if err != nil {
		return fmt.Errorf("failed to create temporal client: %w", err)
	}
	defer c.Close()

	workflowOptions := client.StartWorkflowOptions{
		ID:        fmt.Sprintf("isolate-expired-customers-%s", time.Now().Format("20060102-1504")),
		TaskQueue: "IsolationTaskQueue",
	}

	_, err = c.ExecuteWorkflow(ctx, workflowOptions, IsolateExpiredCustomersWorkflow, IsolateExpiredCustomersInput{
		GracePeriodDays: gracePeriodDays,
	})

	if err != nil {
		return fmt.Errorf("failed to execute workflow: %w", err)
	}

	log.WithContext(ctx).Info("Triggered IsolateExpiredCustomersWorkflow")
	return nil
}

func (a *isolationAdapter) TriggerReactivateCustomer(ctx context.Context, customerID string, invoiceID string) error {
	c, err := a.getTemporalClient()
	if err != nil {
		return fmt.Errorf("failed to create temporal client: %w", err)
	}
	defer c.Close()

	workflowOptions := client.StartWorkflowOptions{
		ID:        fmt.Sprintf("reactivate-customer-%s-%s", customerID, time.Now().Format("20060102-1504")),
		TaskQueue: "IsolationTaskQueue",
	}

	_, err = c.ExecuteWorkflow(ctx, workflowOptions, ReactivateCustomerWorkflow, ReactivateCustomerInput{
		CustomerID: customerID,
		InvoiceID:  invoiceID,
	})

	if err != nil {
		return fmt.Errorf("failed to execute workflow: %w", err)
	}

	log.WithContext(ctx).Info(fmt.Sprintf("Triggered ReactivateCustomerWorkflow for customer %s", customerID))
	return nil
}

func (a *isolationAdapter) getTemporalClient() (client.Client, error) {
	hostPort := fmt.Sprintf("%s:%s", os.Getenv("WORKFLOW_HOST"), os.Getenv("WORKFLOW_PORT"))
	namespace := os.Getenv("WORKFLOW_NAMESPACE")
	if namespace == "" {
		namespace = "default"
	}

	return client.Dial(client.Options{
		HostPort:  hostPort,
		Namespace: namespace,
	})
}
