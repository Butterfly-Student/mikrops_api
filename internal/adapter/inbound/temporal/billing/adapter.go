package billing_workflow

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

type billingAdapter struct {
	domain domain.Domain
}

func NewBillingAdapter(domain domain.Domain) inbound_port.BillingWorkflowPort {
	return &billingAdapter{domain: domain}
}

func (a *billingAdapter) StartInvoiceGenerationWorker(ctx context.Context) {
	go StartInvoiceGenerationWorker(ctx, a.domain)
}

func (a *billingAdapter) StartPaymentProcessingWorker(ctx context.Context) {
	go StartPaymentProcessingWorker(ctx, a.domain)
}

// StartScheduler starts the billing scheduler
func (a *billingAdapter) StartScheduler(ctx context.Context) {
	scheduler := NewScheduler(a.domain)
	scheduler.Start(ctx)
}

func (a *billingAdapter) TriggerGenerateMonthlyInvoices(ctx context.Context, year, month int) error {
	c, err := a.getTemporalClient()
	if err != nil {
		return fmt.Errorf("failed to create temporal client: %w", err)
	}
	defer c.Close()

	workflowOptions := client.StartWorkflowOptions{
		ID:        fmt.Sprintf("generate-monthly-invoices-%d-%d", year, month),
		TaskQueue: "GenerateMonthlyInvoicesTaskQueue",
	}

	_, err = c.ExecuteWorkflow(ctx, workflowOptions, GenerateMonthlyInvoicesWorkflow, GenerateMonthlyInvoicesInput{
		Year:  year,
		Month: month,
		Force: false,
	})

	if err != nil {
		return fmt.Errorf("failed to execute workflow: %w", err)
	}

	log.WithContext(ctx).Info(fmt.Sprintf("Triggered GenerateMonthlyInvoicesWorkflow for %d-%d", year, month))
	return nil
}

func (a *billingAdapter) TriggerProcessPayment(ctx context.Context, input inbound_port.ProcessPaymentWorkflowInput) error {
	c, err := a.getTemporalClient()
	if err != nil {
		return fmt.Errorf("failed to create temporal client: %w", err)
	}
	defer c.Close()

	workflowInput := ProcessPaymentInput{
		InvoiceID:      input.InvoiceID,
		Amount:         input.Amount,
		Status:         input.Status,
		PaymentMethod:  input.PaymentMethod,
		TransactionRef: input.TransactionRef,
		PaymentDate:    input.PaymentDate,
	}

	workflowOptions := client.StartWorkflowOptions{
		ID:        fmt.Sprintf("process-payment-%s", input.InvoiceID),
		TaskQueue: "ProcessPaymentTaskQueue",
	}

	_, err = c.ExecuteWorkflow(ctx, workflowOptions, ProcessPaymentWorkflow, workflowInput)

	if err != nil {
		return fmt.Errorf("failed to execute workflow: %w", err)
	}

	log.WithContext(ctx).Info(fmt.Sprintf("Triggered ProcessPaymentWorkflow for invoice %s", input.InvoiceID))
	return nil
}

func (a *billingAdapter) TriggerCheckOverdue(ctx context.Context, applyLateFees bool) error {
	c, err := a.getTemporalClient()
	if err != nil {
		return fmt.Errorf("failed to create temporal client: %w", err)
	}
	defer c.Close()

	workflowOptions := client.StartWorkflowOptions{
		ID:        fmt.Sprintf("check-overdue-%s", time.Now().Format("20060102-1504")),
		TaskQueue: "GenerateMonthlyInvoicesTaskQueue",
	}

	_, err = c.ExecuteWorkflow(ctx, workflowOptions, CheckOverdueInvoicesWorkflow, CheckOverdueInput{
		ApplyLateFees: applyLateFees,
	})

	if err != nil {
		return fmt.Errorf("failed to execute workflow: %w", err)
	}

	log.WithContext(ctx).Info("Triggered CheckOverdueInvoicesWorkflow")
	return nil
}

func (a *billingAdapter) TriggerSendReminders(ctx context.Context, daysBefore, daysAfter []int) error {
	c, err := a.getTemporalClient()
	if err != nil {
		return fmt.Errorf("failed to create temporal client: %w", err)
	}
	defer c.Close()

	workflowOptions := client.StartWorkflowOptions{
		ID:        fmt.Sprintf("send-reminders-%s", time.Now().Format("20060102-1504")),
		TaskQueue: "GenerateMonthlyInvoicesTaskQueue",
	}

	_, err = c.ExecuteWorkflow(ctx, workflowOptions, SendInvoiceRemindersWorkflow, SendRemindersWorkflowInput{
		DaysBeforeDue: daysBefore,
		DaysAfterDue:  daysAfter,
	})

	if err != nil {
		return fmt.Errorf("failed to execute workflow: %w", err)
	}

	log.WithContext(ctx).Info("Triggered SendInvoiceRemindersWorkflow")
	return nil
}

func (a *billingAdapter) getTemporalClient() (client.Client, error) {
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
