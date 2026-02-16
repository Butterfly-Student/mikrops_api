package billing_workflow

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go-template/internal/domain"
	"go-template/internal/model"
	"go-template/utils/log"
)

// Activities struct holds dependencies for billing workflow activities
type Activities struct {
	domain domain.Domain
}

// NewActivities creates a new Activities instance
func NewActivities(domain domain.Domain) *Activities {
	return &Activities{
		domain: domain,
	}
}

// CheckInvoicesGeneratedActivity checks if invoices already exist for given month/year
func (a *Activities) CheckInvoicesGeneratedActivity(ctx context.Context, input CheckInvoicesGeneratedInput) (CheckInvoicesGeneratedResult, error) {
	log.WithContext(ctx).Info(fmt.Sprintf("Checking existing invoices for %d-%d", input.Year, input.Month))

	filter := model.InvoiceFilter{
		BillingMonth: &input.Month,
		BillingYear:  &input.Year,
	}

	invoices, err := a.domain.Billing().ListInvoices(ctx, filter)
	if err != nil {
		return CheckInvoicesGeneratedResult{}, fmt.Errorf("failed to check existing invoices: %w", err)
	}

	log.WithContext(ctx).Info(fmt.Sprintf("Found %d existing invoices for %d-%d", len(invoices), input.Year, input.Month))

	return CheckInvoicesGeneratedResult{
		Exists: len(invoices) > 0,
		Count:  len(invoices),
	}, nil
}

// GenerateInvoicesActivity generates monthly invoices for all active customers
func (a *Activities) GenerateInvoicesActivity(ctx context.Context, input GenerateInvoicesActivityInput) (GenerateInvoicesResult, error) {
	log.WithContext(ctx).Info(fmt.Sprintf("Generating invoices for %d-%d", input.Year, input.Month))

	invoices, err := a.domain.Billing().GenerateMonthlyInvoices(ctx, input.Year, input.Month)
	if err != nil {
		return GenerateInvoicesResult{}, fmt.Errorf("failed to generate invoices: %w", err)
	}

	invoiceIDs := make([]string, 0, len(invoices))
	for _, invoice := range invoices {
		invoiceIDs = append(invoiceIDs, invoice.ID.String())
	}

	log.WithContext(ctx).Info(fmt.Sprintf("Generated %d invoices for %d-%d", len(invoices), input.Year, input.Month))

	return GenerateInvoicesResult{
		GeneratedCount: len(invoices),
		SkippedCount:   0,
		FailedCount:    0,
		TotalCustomers: 0,
		InvoiceIDs:     invoiceIDs,
	}, nil
}

// SendInvoiceNotificationsActivity sends notifications for newly generated invoices
func (a *Activities) SendInvoiceNotificationsActivity(ctx context.Context, input SendInvoiceNotificationsInput) (SendInvoiceNotificationsResult, error) {
	log.WithContext(ctx).Info(fmt.Sprintf("Sending notifications for %d invoices", len(input.InvoiceIDs)))

	sentCount := 0
	failedCount := 0

	for _, invoiceID := range input.InvoiceIDs {
		invoice, err := a.domain.Billing().GetInvoice(ctx, invoiceID)
		if err != nil {
			log.WithContext(ctx).Warn(fmt.Sprintf("Failed to get invoice %s: %v", invoiceID, err))
			failedCount++
			continue
		}

		err = a.domain.Notification().SendInvoiceCreated(ctx, invoice.CustomerID.String(), invoiceID)
		if err != nil {
			log.WithContext(ctx).Warn(fmt.Sprintf("Failed to send notification for invoice %s: %v", invoiceID, err))
			failedCount++
		} else {
			sentCount++
			log.WithContext(ctx).Info(fmt.Sprintf("Notification sent for invoice %s", invoiceID))
		}
	}

	log.WithContext(ctx).Info(fmt.Sprintf("Notifications sent: %d, failed: %d", sentCount, failedCount))

	return SendInvoiceNotificationsResult{
		SentCount:   sentCount,
		FailedCount: failedCount,
	}, nil
}

// ValidatePaymentWebhookActivity validates payment webhook from Xendit
func (a *Activities) ValidatePaymentWebhookActivity(ctx context.Context, input ValidatePaymentWebhookInput) (ValidatePaymentWebhookResult, error) {
	log.WithContext(ctx).Info(fmt.Sprintf("Validating payment webhook for invoice %s, amount %f", input.InvoiceID, input.Amount))

	// Check if invoice exists
	invoice, err := a.domain.Billing().GetInvoice(ctx, input.InvoiceID)
	if err != nil {
		return ValidatePaymentWebhookResult{
			Valid:  false,
			Reason: "Invoice not found",
		}, nil
	}

	// Validate status
	if input.Status != "PAID" && input.Status != "COMPLETED" {
		return ValidatePaymentWebhookResult{
			Valid:  false,
			Reason: "Invalid payment status: " + input.Status,
		}, nil
	}

	// Validate amount (allow small difference due to rounding)
	diff := input.Amount - invoice.TotalAmount
	if diff < -100 || diff > 100 {
		return ValidatePaymentWebhookResult{
			Valid:  false,
			Reason: fmt.Sprintf("Amount mismatch: expected %.2f, got %.2f", invoice.TotalAmount, input.Amount),
		}, nil
	}

	log.WithContext(ctx).Info(fmt.Sprintf("Payment webhook validated for invoice %s", input.InvoiceID))

	return ValidatePaymentWebhookResult{
		Valid:  true,
		Reason: "",
	}, nil
}

// UpdateInvoiceStatusActivity updates invoice status to paid
func (a *Activities) UpdateInvoiceStatusActivity(ctx context.Context, input UpdateInvoiceStatusInput) (UpdateInvoiceStatusResult, error) {
	log.WithContext(ctx).Info(fmt.Sprintf("Updating invoice %s status to %s", input.InvoiceID, input.Status))

	invoice, err := a.domain.Billing().GetInvoice(ctx, input.InvoiceID)
	if err != nil {
		return UpdateInvoiceStatusResult{Updated: false}, fmt.Errorf("invoice not found: %w", err)
	}

	// Prepare update input
	invoiceInput := model.InvoiceInput{
		CustomerID:  invoice.CustomerID,
		Status:      &input.Status,
		PaymentDate: input.PaymentDate,
	}

	if input.PaymentMethod != "" {
		invoiceInput.PaymentMethod = &input.PaymentMethod
	}

	_, err = a.domain.Billing().UpdateInvoice(ctx, input.InvoiceID, invoiceInput)
	if err != nil {
		return UpdateInvoiceStatusResult{Updated: false}, fmt.Errorf("failed to update invoice: %w", err)
	}

	log.WithContext(ctx).Info(fmt.Sprintf("Invoice %s updated to %s", input.InvoiceID, input.Status))

	return UpdateInvoiceStatusResult{Updated: true}, nil
}

// CreatePaymentRecordActivity creates payment record
func (a *Activities) CreatePaymentRecordActivity(ctx context.Context, input CreatePaymentRecordInput) (CreatePaymentRecordResult, error) {
	log.WithContext(ctx).Info(fmt.Sprintf("Creating payment record for invoice %s", input.InvoiceID))

	invoice, err := a.domain.Billing().GetInvoice(ctx, input.InvoiceID)
	if err != nil {
		return CreatePaymentRecordResult{}, fmt.Errorf("invoice not found: %w", err)
	}

	paymentInput := model.PaymentInput{
		CustomerID:           invoice.CustomerID,
		InvoiceID:            &invoice.ID,
		Amount:               &input.Amount,
		PaymentMethod:        &input.PaymentMethod,
		PaymentDate:          &input.PaymentDate,
		TransactionReference: &input.TransactionRef,
	}

	payment, err := a.domain.Payment().CreatePayment(ctx, paymentInput)
	if err != nil {
		return CreatePaymentRecordResult{}, fmt.Errorf("failed to create payment: %w", err)
	}

	log.WithContext(ctx).Info(fmt.Sprintf("Payment %s created for invoice %s", payment.PaymentNumber, input.InvoiceID))

	return CreatePaymentRecordResult{
		PaymentID: payment.ID.String(),
	}, nil
}

// RecordCashTransactionActivity creates cash transaction for payment
func (a *Activities) RecordCashTransactionActivity(ctx context.Context, input RecordCashTransactionInput) (RecordCashTransactionResult, error) {
	log.WithContext(ctx).Info(fmt.Sprintf("Recording cash transaction for payment %s", input.PaymentID))

	// Get payment details
	payment, err := a.domain.Payment().GetPayment(ctx, input.PaymentID)
	if err != nil {
		return RecordCashTransactionResult{}, fmt.Errorf("payment not found: %w", err)
	}

	// Get invoice details to determine category
	invoice, err := a.domain.Billing().GetInvoice(ctx, input.InvoiceID)
	if err != nil {
		return RecordCashTransactionResult{}, fmt.Errorf("invoice not found: %w", err)
	}

	// Find or create subscription category
	typeStr := "income"
	categoryFilter := model.CashCategoryFilter{
		Type:     &typeStr,
		IsActive: func() *bool { b := true; return &b }(),
	}

	categories, err := a.domain.Cash().ListCategories(ctx, categoryFilter)
	if err != nil {
		return RecordCashTransactionResult{}, fmt.Errorf("failed to list categories: %w", err)
	}

	var categoryID uuid.UUID
	for _, cat := range categories {
		if cat.Code == "INC-SUB" {
			categoryID = cat.ID
			break
		}
	}

	if categoryID == uuid.Nil {
		return RecordCashTransactionResult{}, fmt.Errorf("subscription category not found")
	}

	// Create cash transaction
	transactionType := "income"
	transactionInput := model.CashTransactionInput{
		Type:          &transactionType,
		CategoryID:    categoryID,
		Amount:        &input.Amount,
		PaymentMethod: &payment.PaymentMethod,
		Description:   func() *string { s := fmt.Sprintf("Payment for invoice %s", invoice.InvoiceNumber); return &s }(),
		ReferenceType: func() *string { s := "payment"; return &s }(),
		ReferenceID:   &payment.ID,
		CustomerID:    &payment.CustomerID,
	}

	transaction, err := a.domain.Cash().CreateTransaction(ctx, transactionInput)
	if err != nil {
		return RecordCashTransactionResult{}, fmt.Errorf("failed to create cash transaction: %w", err)
	}

	log.WithContext(ctx).Info(fmt.Sprintf("Cash transaction %s recorded", transaction.TransactionNumber))

	return RecordCashTransactionResult{
		TransactionID: transaction.ID.String(),
	}, nil
}

// SendPaymentNotificationActivity sends payment confirmation notification
func (a *Activities) SendPaymentNotificationActivity(ctx context.Context, input SendPaymentNotificationInput) error {
	log.WithContext(ctx).Info(fmt.Sprintf("Sending payment notification for payment %s", input.PaymentID))

	payment, err := a.domain.Payment().GetPayment(ctx, input.PaymentID)
	if err != nil {
		return fmt.Errorf("payment not found: %w", err)
	}

	invoice, err := a.domain.Billing().GetInvoice(ctx, input.InvoiceID)
	if err != nil {
		return fmt.Errorf("invoice not found: %w", err)
	}

	err = a.domain.Notification().SendPaymentConfirmation(ctx, invoice.CustomerID.String(), input.PaymentID, payment.Amount)
	if err != nil {
		return fmt.Errorf("failed to send payment notification: %w", err)
	}

	log.WithContext(ctx).Info(fmt.Sprintf("Payment notification sent for %s", input.PaymentID))

	return nil
}

// CheckCustomerReactivationActivity checks if customer needs reactivation
func (a *Activities) CheckCustomerReactivationActivity(ctx context.Context, input CheckCustomerReactivationInput) (CheckCustomerReactivationResult, error) {
	log.WithContext(ctx).Info(fmt.Sprintf("Checking reactivation for invoice %s", input.InvoiceID))

	invoice, err := a.domain.Billing().GetInvoice(ctx, input.InvoiceID)
	if err != nil {
		return CheckCustomerReactivationResult{}, fmt.Errorf("invoice not found: %w", err)
	}

	customer, err := a.domain.Customer().GetCustomer(ctx, invoice.CustomerID.String())
	if err != nil {
		return CheckCustomerReactivationResult{}, fmt.Errorf("customer not found: %w", err)
	}

	// Check if customer was isolated
	if customer.Status != "isolated" {
		return CheckCustomerReactivationResult{
			Needed:     false,
			CustomerID: customer.ID.String(),
		}, nil
	}

	// Check if all overdue invoices are paid
	filter := model.InvoiceFilter{
		CustomerIDs: []uuid.UUID{customer.ID},
		Status:      []string{"overdue"},
	}

	overdueInvoices, err := a.domain.Billing().ListInvoices(ctx, filter)
	if err != nil {
		return CheckCustomerReactivationResult{}, fmt.Errorf("failed to check overdue invoices: %w", err)
	}

	// If there are still overdue invoices, don't reactivate
	if len(overdueInvoices) > 0 {
		log.WithContext(ctx).Info(fmt.Sprintf("Customer %s still has %d overdue invoices, skipping reactivation", customer.CustomerCode, len(overdueInvoices)))
		return CheckCustomerReactivationResult{
			Needed:     false,
			CustomerID: customer.ID.String(),
		}, nil
	}

	log.WithContext(ctx).Info(fmt.Sprintf("Customer %s needs reactivation", customer.CustomerCode))

	return CheckCustomerReactivationResult{
		Needed:     true,
		CustomerID: customer.ID.String(),
	}, nil
}

// TriggerReactivationWorkflowActivity triggers reactivation workflow
func (a *Activities) TriggerReactivationWorkflowActivity(ctx context.Context, input TriggerReactivationWorkflowInput) error {
	log.WithContext(ctx).Info(fmt.Sprintf("Triggering reactivation workflow for customer %s", input.CustomerID))

	// This activity should trigger the reactivation workflow
	// For now, we'll call the domain method directly
	err := a.domain.Customer().ActivateCustomer(ctx, input.CustomerID)
	if err != nil {
		return fmt.Errorf("failed to activate customer: %w", err)
	}

	log.WithContext(ctx).Info(fmt.Sprintf("Customer %s reactivated", input.CustomerID))

	return nil
}

// FindOverdueInvoicesActivity finds all overdue invoices
func (a *Activities) FindOverdueInvoicesActivity(ctx context.Context, input FindOverdueInvoicesActivityInput) (FindOverdueInvoicesResult, error) {
	log.WithContext(ctx).Info("Finding overdue invoices")

	isOverdue := true
	filter := model.InvoiceFilter{
		Status:    []string{"sent", "partial"},
		IsOverdue: &isOverdue,
	}

	invoices, err := a.domain.Billing().ListInvoices(ctx, filter)
	if err != nil {
		return FindOverdueInvoicesResult{}, fmt.Errorf("failed to find overdue invoices: %w", err)
	}

	invoiceIDs := make([]string, 0, len(invoices))
	for _, invoice := range invoices {
		invoiceIDs = append(invoiceIDs, invoice.ID.String())
	}

	log.WithContext(ctx).Info(fmt.Sprintf("Found %d overdue invoices", len(invoiceIDs)))

	return FindOverdueInvoicesResult{
		InvoiceIDs: invoiceIDs,
	}, nil
}

// ApplyLateFeesActivity applies late fees to overdue invoices
func (a *Activities) ApplyLateFeesActivity(ctx context.Context, input ApplyLateFeesInput) (ApplyLateFeesResult, error) {
	log.WithContext(ctx).Info(fmt.Sprintf("Applying late fees to %d invoices", len(input.InvoiceIDs)))

	appliedCount := 0

	for _, invoiceID := range input.InvoiceIDs {
		invoice, err := a.domain.Billing().GetInvoice(ctx, invoiceID)
		if err != nil {
			log.WithContext(ctx).Warn(fmt.Sprintf("Failed to get invoice %s: %v", invoiceID, err))
			continue
		}

		lateFee := a.domain.Billing().CalculateLateFee(ctx, invoice)
		if lateFee > 0 {
			newLateFee := invoice.LateFee + lateFee
			newTotalAmount := invoice.TotalAmount + lateFee

			invoiceInput := model.InvoiceInput{
				CustomerID:  invoice.CustomerID,
				LateFee:     &newLateFee,
				TotalAmount: &newTotalAmount,
			}

			_, err = a.domain.Billing().UpdateInvoice(ctx, invoiceID, invoiceInput)
			if err != nil {
				log.WithContext(ctx).Warn(fmt.Sprintf("Failed to update invoice %s with late fee: %v", invoiceID, err))
				continue
			}

			appliedCount++
			log.WithContext(ctx).Info(fmt.Sprintf("Applied late fee %.2f to invoice %s", lateFee, invoiceID))
		}
	}

	log.WithContext(ctx).Info(fmt.Sprintf("Late fees applied to %d invoices", appliedCount))

	return ApplyLateFeesResult{
		AppliedCount: appliedCount,
	}, nil
}

// FindInvoicesDueSoonActivity finds invoices due within specified days
func (a *Activities) FindInvoicesDueSoonActivity(ctx context.Context, input FindInvoicesDueSoonInput) (FindInvoicesDueSoonResult, error) {
	log.WithContext(ctx).Info(fmt.Sprintf("Finding invoices due soon within %v days", input.DaysBeforeDue))

	filter := model.InvoiceFilter{
		Status: []string{"sent", "partial"},
	}

	invoices, err := a.domain.Billing().ListInvoices(ctx, filter)
	if err != nil {
		return FindInvoicesDueSoonResult{}, fmt.Errorf("failed to find invoices: %w", err)
	}

	now := time.Now()
	dueSoonInvoices := make([]string, 0)

	for _, invoice := range invoices {
		for _, days := range input.DaysBeforeDue {
			dueDate := invoice.DueDate
			diff := dueDate.Sub(now)
			diffDays := int(diff.Hours() / 24)

			if diffDays == days && diffDays >= 0 {
				// Check if reminder already sent
				if invoice.ReminderSentCount == nil || *invoice.ReminderSentCount < len(input.DaysBeforeDue) {
					dueSoonInvoices = append(dueSoonInvoices, invoice.ID.String())
					break
				}
			}
		}
	}

	log.WithContext(ctx).Info(fmt.Sprintf("Found %d invoices due soon", len(dueSoonInvoices)))

	return FindInvoicesDueSoonResult{
		InvoiceIDs: dueSoonInvoices,
	}, nil
}
