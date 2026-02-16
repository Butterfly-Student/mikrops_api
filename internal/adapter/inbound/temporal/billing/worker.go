package billing_workflow

import (
	"context"

	"go-template/internal/domain"
	"go-template/utils/activity"
	"go-template/utils/log"
	"go-template/utils/temporal"
	"go.temporal.io/sdk/worker"
)

// StartInvoiceGenerationWorker starts worker for invoice generation workflow
func StartInvoiceGenerationWorker(ctx context.Context, domain domain.Domain) {
	activityCtx := activity.NewContext("billing-invoice-generation-worker")

	w, err := temporal.NewWorker(activityCtx, "GenerateMonthlyInvoicesTaskQueue")
	if err != nil {
		log.WithContext(activityCtx).Error("Unable to create invoice generation worker", err)
		return
	}

	activities := NewActivities(domain)

	// Register workflows
	w.RegisterWorkflow(GenerateMonthlyInvoicesWorkflow)
	w.RegisterWorkflow(CheckOverdueInvoicesWorkflow)
	w.RegisterWorkflow(SendInvoiceRemindersWorkflow)

	// Register activities
	w.RegisterActivity(activities.CheckInvoicesGeneratedActivity)
	w.RegisterActivity(activities.GenerateInvoicesActivity)
	w.RegisterActivity(activities.SendInvoiceNotificationsActivity)
	w.RegisterActivity(activities.FindOverdueInvoicesActivity)
	w.RegisterActivity(activities.ApplyLateFeesActivity)
	w.RegisterActivity(activities.FindInvoicesDueSoonActivity)

	// Start worker
	err = w.Run(worker.InterruptCh())
	if err != nil {
		log.WithContext(activityCtx).Error("Unable to start invoice generation worker", err)
	}
}

// StartPaymentProcessingWorker starts worker for payment processing workflow
func StartPaymentProcessingWorker(ctx context.Context, domain domain.Domain) {
	activityCtx := activity.NewContext("billing-payment-processing-worker")

	w, err := temporal.NewWorker(activityCtx, "ProcessPaymentTaskQueue")
	if err != nil {
		log.WithContext(activityCtx).Error("Unable to create payment processing worker", err)
		return
	}

	activities := NewActivities(domain)

	// Register workflows
	w.RegisterWorkflow(ProcessPaymentWorkflow)

	// Register activities
	w.RegisterActivity(activities.ValidatePaymentWebhookActivity)
	w.RegisterActivity(activities.UpdateInvoiceStatusActivity)
	w.RegisterActivity(activities.CreatePaymentRecordActivity)
	w.RegisterActivity(activities.RecordCashTransactionActivity)
	w.RegisterActivity(activities.SendPaymentNotificationActivity)
	w.RegisterActivity(activities.CheckCustomerReactivationActivity)
	w.RegisterActivity(activities.TriggerReactivationWorkflowActivity)

	// Start worker
	err = w.Run(worker.InterruptCh())
	if err != nil {
		log.WithContext(activityCtx).Error("Unable to start payment processing worker", err)
	}
}
