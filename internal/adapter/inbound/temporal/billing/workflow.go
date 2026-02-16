package billing_workflow

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// GenerateMonthlyInvoicesWorkflow generates monthly invoices for all active customers
func GenerateMonthlyInvoicesWorkflow(ctx workflow.Context, input GenerateMonthlyInvoicesInput) (GenerateMonthlyInvoicesResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting GenerateMonthlyInvoicesWorkflow", "year", input.Year, "month", input.Month)

	result := GenerateMonthlyInvoicesResult{
		TotalCustomers:     0,
		GeneratedInvoices:  0,
		SkippedInvoices:    0,
		FailedInvoices:     0,
		NotificationSent:   0,
		NotificationFailed: 0,
		Errors:             []string{},
	}

	activityOptions := workflow.ActivityOptions{
		StartToCloseTimeout: 30 * time.Minute,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    time.Second,
			BackoffCoefficient: 2.0,
			MaximumInterval:    time.Minute,
			MaximumAttempts:    3,
		},
	}
	ctx = workflow.WithActivityOptions(ctx, activityOptions)

	// Step 1: Check if invoices already generated for this month/year
	if !input.Force {
		var checkResult CheckInvoicesGeneratedResult
		err := workflow.ExecuteActivity(ctx, "CheckInvoicesGeneratedActivity", CheckInvoicesGeneratedInput{
			Year:  input.Year,
			Month: input.Month,
		}).Get(ctx, &checkResult)
		if err != nil {
			logger.Error("Failed to check existing invoices", "error", err)
			return result, err
		}

		if checkResult.Exists {
			logger.Info("Invoices already generated for this month/year", "year", input.Year, "month", input.Month)
			return GenerateMonthlyInvoicesResult{
				SkippedInvoices:   checkResult.Count,
				GeneratedInvoices: 0,
			}, nil
		}
	}

	// Step 2: Generate invoices for all customers
	var genResult GenerateInvoicesResult
	err := workflow.ExecuteActivity(ctx, "GenerateInvoicesActivity", GenerateInvoicesActivityInput{
		Year:  input.Year,
		Month: input.Month,
	}).Get(ctx, &genResult)
	if err != nil {
		logger.Error("Failed to generate invoices", "error", err)
		return result, err
	}

	result.GeneratedInvoices = genResult.GeneratedCount
	result.SkippedInvoices = genResult.SkippedCount
	result.FailedInvoices = genResult.FailedCount
	result.TotalCustomers = genResult.TotalCustomers

	logger.Info("Invoice generation completed",
		"generated", result.GeneratedInvoices,
		"skipped", result.SkippedInvoices,
		"failed", result.FailedInvoices)

	if genResult.GeneratedCount == 0 {
		return result, nil
	}

	// Step 3: Send invoice notifications
	var notifyResult SendInvoiceNotificationsResult
	err = workflow.ExecuteActivity(ctx, "SendInvoiceNotificationsActivity", SendInvoiceNotificationsInput{
		Year:       input.Year,
		Month:      input.Month,
		InvoiceIDs: genResult.InvoiceIDs,
	}).Get(ctx, &notifyResult)
	if err != nil {
		logger.Error("Failed to send notifications", "error", err)
		// Don't fail workflow if notifications fail
		result.Errors = append(result.Errors, "Failed to send notifications: "+err.Error())
	} else {
		result.NotificationSent = notifyResult.SentCount
		result.NotificationFailed = notifyResult.FailedCount
		logger.Info("Notifications sent",
			"sent", result.NotificationSent,
			"failed", result.NotificationFailed)
	}

	return result, nil
}

// ProcessPaymentWorkflow processes payment from Xendit webhook
func ProcessPaymentWorkflow(ctx workflow.Context, input ProcessPaymentInput) (ProcessPaymentResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting ProcessPaymentWorkflow", "invoice_id", input.InvoiceID, "amount", input.Amount)

	result := ProcessPaymentResult{
		PaymentID:           "",
		InvoiceUpdated:      false,
		CashTransactionID:   "",
		NotificationSent:    false,
		CustomerReactivated: false,
	}

	activityOptions := workflow.ActivityOptions{
		StartToCloseTimeout: 5 * time.Minute,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    time.Second,
			BackoffCoefficient: 2.0,
			MaximumInterval:    time.Minute,
			MaximumAttempts:    3,
		},
	}
	ctx = workflow.WithActivityOptions(ctx, activityOptions)

	// Step 1: Validate payment webhook
	var validateResult ValidatePaymentWebhookResult
	err := workflow.ExecuteActivity(ctx, "ValidatePaymentWebhookActivity", ValidatePaymentWebhookInput{
		InvoiceID:            input.InvoiceID,
		Amount:               input.Amount,
		Status:               input.Status,
		TransactionReference: input.TransactionRef,
	}).Get(ctx, &validateResult)
	if err != nil {
		logger.Error("Failed to validate webhook", "error", err)
		return result, err
	}

	if !validateResult.Valid {
		logger.Error("Webhook validation failed", "reason", validateResult.Reason)
		return result, nil
	}

	// Step 2: Update invoice status
	var updateResult UpdateInvoiceStatusResult
	err = workflow.ExecuteActivity(ctx, "UpdateInvoiceStatusActivity", UpdateInvoiceStatusInput{
		InvoiceID:     input.InvoiceID,
		Status:        "paid",
		PaymentDate:   &input.PaymentDate,
		PaymentMethod: input.PaymentMethod,
	}).Get(ctx, &updateResult)
	if err != nil {
		logger.Error("Failed to update invoice status", "error", err)
		return result, err
	}

	result.InvoiceUpdated = true

	// Step 3: Create payment record
	var paymentResult CreatePaymentRecordResult
	err = workflow.ExecuteActivity(ctx, "CreatePaymentRecordActivity", CreatePaymentRecordInput{
		InvoiceID:      input.InvoiceID,
		Amount:         input.Amount,
		PaymentMethod:  input.PaymentMethod,
		TransactionRef: input.TransactionRef,
		PaymentDate:    input.PaymentDate,
	}).Get(ctx, &paymentResult)
	if err != nil {
		logger.Error("Failed to create payment record", "error", err)
		return result, err
	}

	result.PaymentID = paymentResult.PaymentID

	// Step 4: Record cash transaction
	var cashResult RecordCashTransactionResult
	err = workflow.ExecuteActivity(ctx, "RecordCashTransactionActivity", RecordCashTransactionInput{
		PaymentID: result.PaymentID,
		InvoiceID: input.InvoiceID,
		Amount:    input.Amount,
	}).Get(ctx, &cashResult)
	if err != nil {
		logger.Error("Failed to record cash transaction", "error", err)
		// Don't fail workflow if cash transaction fails
		result.Errors = append(result.Errors, "Failed to record cash transaction: "+err.Error())
	} else {
		result.CashTransactionID = cashResult.TransactionID
	}

	// Step 5: Send payment notification
	err = workflow.ExecuteActivity(ctx, "SendPaymentNotificationActivity", SendPaymentNotificationInput{
		PaymentID: result.PaymentID,
		InvoiceID: input.InvoiceID,
	}).Get(ctx, nil)
	if err != nil {
		logger.Warn("Failed to send payment notification", "error", err)
		// Don't fail workflow if notification fails
	} else {
		result.NotificationSent = true
	}

	// Step 6: Check if customer needs reactivation
	var reactivateResult CheckCustomerReactivationResult
	err = workflow.ExecuteActivity(ctx, "CheckCustomerReactivationActivity", CheckCustomerReactivationInput{
		InvoiceID: input.InvoiceID,
	}).Get(ctx, &reactivateResult)
	if err != nil {
		logger.Warn("Failed to check reactivation", "error", err)
	} else if reactivateResult.Needed {
		logger.Info("Customer needs reactivation, triggering workflow")
		err = workflow.ExecuteActivity(ctx, "TriggerReactivationWorkflowActivity", TriggerReactivationWorkflowInput{
			CustomerID: reactivateResult.CustomerID,
			InvoiceID:  input.InvoiceID,
		}).Get(ctx, nil)
		if err != nil {
			logger.Error("Failed to trigger reactivation", "error", err)
		} else {
			result.CustomerReactivated = true
		}
	}

	logger.Info("ProcessPaymentWorkflow completed successfully", "payment_id", result.PaymentID)

	return result, nil
}

// CheckOverdueInvoicesWorkflow checks and updates overdue invoices
func CheckOverdueInvoicesWorkflow(ctx workflow.Context, input CheckOverdueInput) (CheckOverdueResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting CheckOverdueInvoicesWorkflow", "apply_late_fees", input.ApplyLateFees)

	result := CheckOverdueResult{
		TotalOverdue:      0,
		LateFeesApplied:   0,
		NotificationsSent: 0,
		Errors:            []string{},
	}

	activityOptions := workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Minute,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    time.Second,
			BackoffCoefficient: 2.0,
			MaximumInterval:    time.Minute,
			MaximumAttempts:    3,
		},
	}
	ctx = workflow.WithActivityOptions(ctx, activityOptions)

	// Step 1: Find overdue invoices
	var findResult FindOverdueInvoicesResult
	err := workflow.ExecuteActivity(ctx, "FindOverdueInvoicesActivity", FindOverdueInvoicesActivityInput{}).Get(ctx, &findResult)
	if err != nil {
		logger.Error("Failed to find overdue invoices", "error", err)
		return result, err
	}

	result.TotalOverdue = len(findResult.InvoiceIDs)
	logger.Info("Found overdue invoices", "count", result.TotalOverdue)

	if result.TotalOverdue == 0 {
		return result, nil
	}

	// Step 2: Apply late fees if enabled
	if input.ApplyLateFees {
		var applyResult ApplyLateFeesResult
		err = workflow.ExecuteActivity(ctx, "ApplyLateFeesActivity", ApplyLateFeesInput{
			InvoiceIDs: findResult.InvoiceIDs,
		}).Get(ctx, &applyResult)
		if err != nil {
			logger.Error("Failed to apply late fees", "error", err)
			result.Errors = append(result.Errors, "Failed to apply late fees: "+err.Error())
		} else {
			result.LateFeesApplied = applyResult.AppliedCount
			logger.Info("Late fees applied", "count", result.LateFeesApplied)
		}
	}

	// Step 3: Send overdue notifications
	var notifyResult SendInvoiceNotificationsResult
	err = workflow.ExecuteActivity(ctx, "SendInvoiceNotificationsActivity", SendInvoiceNotificationsInput{
		InvoiceIDs: findResult.InvoiceIDs,
	}).Get(ctx, &notifyResult)
	if err != nil {
		logger.Error("Failed to send overdue notifications", "error", err)
		result.Errors = append(result.Errors, "Failed to send notifications: "+err.Error())
	} else {
		result.NotificationsSent = notifyResult.SentCount
		logger.Info("Overdue notifications sent", "count", result.NotificationsSent)
	}

	return result, nil
}

// SendInvoiceRemindersWorkflow sends invoice reminders
func SendInvoiceRemindersWorkflow(ctx workflow.Context, input SendRemindersWorkflowInput) (SendRemindersWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting SendInvoiceRemindersWorkflow")

	result := SendRemindersWorkflowResult{
		TotalReminders: 0,
		Sent:           0,
		Failed:         0,
		Errors:         []string{},
	}

	activityOptions := workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Minute,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    time.Second,
			BackoffCoefficient: 2.0,
			MaximumInterval:    time.Minute,
			MaximumAttempts:    3,
		},
	}
	ctx = workflow.WithActivityOptions(ctx, activityOptions)

	// Step 1: Find invoices due soon
	var dueSoonResult FindInvoicesDueSoonResult
	err := workflow.ExecuteActivity(ctx, "FindInvoicesDueSoonActivity", FindInvoicesDueSoonInput{
		DaysBeforeDue: input.DaysBeforeDue,
	}).Get(ctx, &dueSoonResult)
	if err != nil {
		logger.Error("Failed to find invoices due soon", "error", err)
		result.Errors = append(result.Errors, "Failed to find invoices due soon: "+err.Error())
	} else if len(dueSoonResult.InvoiceIDs) > 0 {
		logger.Info("Found invoices due soon", "count", len(dueSoonResult.InvoiceIDs))

		var notifyResult SendInvoiceNotificationsResult
		err = workflow.ExecuteActivity(ctx, "SendInvoiceNotificationsActivity", SendInvoiceNotificationsInput{
			InvoiceIDs: dueSoonResult.InvoiceIDs,
		}).Get(ctx, &notifyResult)
		if err != nil {
			logger.Error("Failed to send reminders", "error", err)
			result.Errors = append(result.Errors, "Failed to send reminders: "+err.Error())
		} else {
			result.Sent += notifyResult.SentCount
			result.Failed += notifyResult.FailedCount
			result.TotalReminders += len(dueSoonResult.InvoiceIDs)
		}
	}

	// Step 2: Find invoices overdue but not yet isolated
	var overdueResult FindOverdueInvoicesResult
	err = workflow.ExecuteActivity(ctx, "FindOverdueInvoicesActivity", FindOverdueInvoicesActivityInput{}).Get(ctx, &overdueResult)
	if err != nil {
		logger.Error("Failed to find overdue invoices", "error", err)
		result.Errors = append(result.Errors, "Failed to find overdue invoices: "+err.Error())
	} else if len(overdueResult.InvoiceIDs) > 0 {
		logger.Info("Found overdue invoices", "count", len(overdueResult.InvoiceIDs))

		var notifyResult SendInvoiceNotificationsResult
		err = workflow.ExecuteActivity(ctx, "SendInvoiceNotificationsActivity", SendInvoiceNotificationsInput{
			InvoiceIDs: overdueResult.InvoiceIDs,
		}).Get(ctx, &notifyResult)
		if err != nil {
			logger.Error("Failed to send overdue reminders", "error", err)
			result.Errors = append(result.Errors, "Failed to send overdue reminders: "+err.Error())
		} else {
			result.Sent += notifyResult.SentCount
			result.Failed += notifyResult.FailedCount
			result.TotalReminders += len(overdueResult.InvoiceIDs)
		}
	}

	logger.Info("SendInvoiceRemindersWorkflow completed",
		"total", result.TotalReminders,
		"sent", result.Sent,
		"failed", result.Failed)

	return result, nil
}

// Workflow input/output types

type GenerateMonthlyInvoicesInput struct {
	Year  int
	Month int
	Force bool
}

type GenerateMonthlyInvoicesResult struct {
	TotalCustomers     int
	GeneratedInvoices  int
	SkippedInvoices    int
	FailedInvoices     int
	NotificationSent   int
	NotificationFailed int
	Errors             []string
}

type ProcessPaymentInput struct {
	InvoiceID      string
	Amount         float64
	Status         string
	PaymentMethod  string
	TransactionRef string
	PaymentDate    time.Time
}

type ProcessPaymentResult struct {
	PaymentID           string
	InvoiceUpdated      bool
	CashTransactionID   string
	NotificationSent    bool
	CustomerReactivated bool
	Errors              []string
}

type CheckOverdueInput struct {
	ApplyLateFees bool
}

type CheckOverdueResult struct {
	TotalOverdue      int
	LateFeesApplied   int
	NotificationsSent int
	Errors            []string
}

type SendRemindersWorkflowInput struct {
	DaysBeforeDue []int
	DaysAfterDue  []int
}

type SendRemindersWorkflowResult struct {
	TotalReminders int
	Sent           int
	Failed         int
	Errors         []string
}

// Activity input/output types

type CheckInvoicesGeneratedInput struct {
	Year  int
	Month int
}

type CheckInvoicesGeneratedResult struct {
	Exists bool
	Count  int
}

type GenerateInvoicesActivityInput struct {
	Year  int
	Month int
}

type GenerateInvoicesResult struct {
	GeneratedCount int
	SkippedCount   int
	FailedCount    int
	TotalCustomers int
	InvoiceIDs     []string
}

type SendInvoiceNotificationsInput struct {
	Year       int
	Month      int
	InvoiceIDs []string
}

type SendInvoiceNotificationsResult struct {
	SentCount   int
	FailedCount int
}

type ValidatePaymentWebhookInput struct {
	InvoiceID            string
	Amount               float64
	Status               string
	TransactionReference string
}

type ValidatePaymentWebhookResult struct {
	Valid  bool
	Reason string
}

type UpdateInvoiceStatusInput struct {
	InvoiceID     string
	Status        string
	PaymentDate   *time.Time
	PaymentMethod string
}

type UpdateInvoiceStatusResult struct {
	Updated bool
}

type CreatePaymentRecordInput struct {
	InvoiceID      string
	Amount         float64
	PaymentMethod  string
	TransactionRef string
	PaymentDate    time.Time
}

type CreatePaymentRecordResult struct {
	PaymentID string
}

type RecordCashTransactionInput struct {
	PaymentID string
	InvoiceID string
	Amount    float64
}

type RecordCashTransactionResult struct {
	TransactionID string
}

type SendPaymentNotificationInput struct {
	PaymentID string
	InvoiceID string
}

type CheckCustomerReactivationInput struct {
	InvoiceID string
}

type CheckCustomerReactivationResult struct {
	Needed     bool
	CustomerID string
}

type TriggerReactivationWorkflowInput struct {
	CustomerID string
	InvoiceID  string
}

type FindOverdueInvoicesActivityInput struct {
	DaysAfterDue []int
}

type FindOverdueInvoicesResult struct {
	InvoiceIDs []string
}

type ApplyLateFeesInput struct {
	InvoiceIDs []string
}

type ApplyLateFeesResult struct {
	AppliedCount int
}

type FindInvoicesDueSoonInput struct {
	DaysBeforeDue []int
}

type FindInvoicesDueSoonResult struct {
	InvoiceIDs []string
}
