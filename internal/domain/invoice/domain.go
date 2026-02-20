package invoice

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/palantir/stacktrace"

	"go-template/internal/model"
	outbound_port "go-template/internal/port/outbound"
)

type InvoiceDomain interface {
	Create(ctx context.Context, input model.InvoiceInput) (*model.Invoice, error)
	GetByID(ctx context.Context, id string) (*model.Invoice, error)
	GetByNumber(ctx context.Context, number string) (*model.Invoice, error)
	List(ctx context.Context, filter *model.InvoiceFilter) ([]model.Invoice, error)
	Update(ctx context.Context, id string, input model.InvoiceInput) (*model.Invoice, error)
	Delete(ctx context.Context, id string) error
	GenerateInvoiceNumber(ctx context.Context) (string, error)
	GenerateMonthlyInvoice(ctx context.Context, customerID string, month, year int) (*model.Invoice, error)
	CalculateLateFee(ctx context.Context, id string) (*model.Invoice, error)
	MarkAsPaid(ctx context.Context, id string, paymentID *uuid.UUID) error
	AddPayment(ctx context.Context, id string, amount float64) (*model.Invoice, error)
}

type domain struct {
	dbPort outbound_port.DatabasePort
}

func NewInvoiceDomain(
	dbPort outbound_port.DatabasePort,
) InvoiceDomain {
	return &domain{
		dbPort: dbPort,
	}
}

func (d *domain) Create(ctx context.Context, input model.InvoiceInput) (*model.Invoice, error) {
	// Generate invoice number if not provided
	invoiceNumber := input.InvoiceNumber
	if invoiceNumber == "" {
		number, err := d.GenerateInvoiceNumber(ctx)
		if err != nil {
			return nil, stacktrace.Propagate(err, "failed to generate invoice number")
		}
		invoiceNumber = number
		input.InvoiceNumber = number
	}

	// Convert input to model
	invoice := input.ToModel()
	invoice.ID = uuid.New()

	// Calculate payment status based on paid amount
	invoice.UpdatePaymentStatus()

	// Create invoice (with items via GORM associations)
	if err := d.dbPort.Invoice().Create(ctx, invoice); err != nil {
		return nil, stacktrace.Propagate(err, "failed to create invoice")
	}

	// Fetch created invoice with relations
	createdInvoice, err := d.GetByID(ctx, invoice.ID.String())
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to get created invoice")
	}

	return createdInvoice, nil
}

func (d *domain) GetByID(ctx context.Context, id string) (*model.Invoice, error) {
	invoice, err := d.dbPort.Invoice().FindByID(ctx, id)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to get invoice by id")
	}

	return invoice, nil
}

func (d *domain) GetByNumber(ctx context.Context, number string) (*model.Invoice, error) {
	invoice, err := d.dbPort.Invoice().FindByNumber(ctx, number)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to get invoice by number")
	}

	return invoice, nil
}

func (d *domain) List(ctx context.Context, filter *model.InvoiceFilter) ([]model.Invoice, error) {
	invoices, err := d.dbPort.Invoice().FindAll(ctx, filter)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to list invoices")
	}

	return invoices, nil
}

func (d *domain) Update(ctx context.Context, id string, input model.InvoiceInput) (*model.Invoice, error) {
	// Get existing invoice
	invoice, err := d.dbPort.Invoice().FindByID(ctx, id)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to find invoice")
	}

	// Update fields
	invoice.InvoiceNumber = input.InvoiceNumber
	invoice.CustomerID = input.CustomerID
	invoice.BillingPeriodStart = input.BillingPeriodStart
	invoice.BillingPeriodEnd = input.BillingPeriodEnd
	invoice.BillingMonth = input.BillingMonth
	invoice.BillingYear = input.BillingYear
	invoice.DueDate = input.DueDate
	invoice.PaymentDeadline = input.PaymentDeadline
	invoice.Subtotal = input.Subtotal
	invoice.TaxAmount = input.TaxAmount
	invoice.DiscountAmount = input.DiscountAmount
	invoice.LateFee = input.LateFee
	invoice.TotalAmount = input.TotalAmount
	invoice.Notes = input.Notes
	invoice.InternalNotes = input.InternalNotes

	if input.IssueDate != nil {
		invoice.IssueDate = *input.IssueDate
	}

	if input.Status != nil {
		invoice.Status = model.InvoiceStatus(*input.Status)
	}

	if input.InvoiceType != nil {
		invoice.InvoiceType = model.InvoiceType(*input.InvoiceType)
	}

	// Update payment status
	invoice.UpdatePaymentStatus()

	// Create new items if provided
	if len(input.Items) > 0 {
		items := make([]model.InvoiceItem, len(input.Items))
		for idx, itemInput := range input.Items {
			item := itemInput.ToModel()
			item.InvoiceID = invoice.ID
			items[idx] = *item
		}
		invoice.Items = items
	}

	// Update invoice (GORM will handle associations)
	if err := d.dbPort.Invoice().Update(ctx, invoice); err != nil {
		return nil, stacktrace.Propagate(err, "failed to update invoice")
	}

	// Fetch updated invoice with relations
	updatedInvoice, err := d.GetByID(ctx, id)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to get updated invoice")
	}

	return updatedInvoice, nil
}

func (d *domain) Delete(ctx context.Context, id string) error {
	if err := d.dbPort.Invoice().Delete(ctx, id); err != nil {
		return stacktrace.Propagate(err, "failed to delete invoice")
	}

	return nil
}

func (d *domain) GenerateInvoiceNumber(ctx context.Context) (string, error) {
	now := time.Now()
	year := now.Year()
	month := int(now.Month())

	// Get last invoice number for this month
	lastNumber, err := d.dbPort.Invoice().GetLastInvoiceNumber(ctx, year, month)
	if err != nil {
		// If no invoice found, start from 1
		lastNumber = ""
	}

	// Parse sequence number
	sequence := 1
	if lastNumber != "" {
		// Format: INV/YYYY/MM/XXXXX
		// Extract last part
		var seq int
		_, err := fmt.Sscanf(lastNumber, "INV/%d/%d/%d", &year, &month, &seq)
		if err == nil {
			sequence = seq + 1
		}
	}

	// Generate new invoice number
	invoiceNumber := fmt.Sprintf("INV/%d/%02d/%05d", year, month, sequence)
	return invoiceNumber, nil
}

func (d *domain) GenerateMonthlyInvoice(ctx context.Context, customerID string, month, year int) (*model.Invoice, error) {
	// Get customer with profile
	customer, err := d.dbPort.Customer().FindByID(ctx, customerID)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to get customer")
	}

	if customer.ProfileID == nil {
		return nil, stacktrace.NewError("customer has no profile assigned")
	}

	// Get profile
	profile, err := d.dbPort.BandwidthProfile().FindByID(ctx, customer.ProfileID.String())
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to get bandwidth profile")
	}

	// Calculate billing period
	billingPeriodStart := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.Local)
	billingPeriodEnd := billingPeriodStart.AddDate(0, 1, -1) // Last day of month

	// Generate invoice number
	invoiceNumber, err := d.GenerateInvoiceNumber(ctx)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to generate invoice number")
	}

	// Calculate due date (issue_date + 7 days default)
	issueDate := time.Now()
	dueDate := issueDate.AddDate(0, 0, 7)

	// Create invoice item for subscription
	itemType := model.InvoiceItemTypeSubscription
	taxRate := 0.0
	if profile.TaxRate != nil {
		taxRate = *profile.TaxRate
	}

	subtotal := profile.PriceMonthly
	taxAmount := subtotal * taxRate
	total := subtotal + taxAmount

	items := []model.InvoiceItem{
		{
			ID:          uuid.New(),
			ItemType:    &itemType,
			Description: fmt.Sprintf("Subscription - %s", profile.Name),
			ProfileID:   &profile.ID,
			Quantity:    1,
			UnitPrice:   profile.PriceMonthly,
			Subtotal:    subtotal,
			TaxRate:     taxRate,
			TaxAmount:   taxAmount,
			Total:       total,
		},
	}

	// Check if customer has unpaid invoices (add late fee)
	overdueInvoices, err := d.findOverdueInvoicesByCustomer(ctx, customerID)
	if err == nil && len(overdueInvoices) > 0 {
		// Add late fee (default 10000)
		lateFeeAmount := 10000.0
		lateFeeType := model.InvoiceItemTypeOther
		items = append(items, model.InvoiceItem{
			ID:          uuid.New(),
			ItemType:    &lateFeeType,
			Description: "Late Payment Fee",
			Quantity:    1,
			UnitPrice:   lateFeeAmount,
			Subtotal:    lateFeeAmount,
			TaxRate:     0,
			TaxAmount:   0,
			Total:       lateFeeAmount,
		})
		total += lateFeeAmount
	}

	// Create invoice
	paymentStatus := model.PaymentStatusUnpaid
	autoGenerated := true
	billingMonthInt := month
	billingYearInt := year

	invoice := &model.Invoice{
		ID:                 uuid.New(),
		InvoiceNumber:      invoiceNumber,
		CustomerID:         customer.ID,
		BillingPeriodStart: billingPeriodStart,
		BillingPeriodEnd:   billingPeriodEnd,
		BillingMonth:       &billingMonthInt,
		BillingYear:        &billingYearInt,
		IssueDate:          issueDate,
		DueDate:            dueDate,
		Subtotal:           subtotal,
		TaxAmount:          taxAmount,
		TotalAmount:        total,
		PaidAmount:         0,
		Status:             model.InvoiceStatusSent,
		PaymentStatus:      &paymentStatus,
		InvoiceType:        model.InvoiceTypeRecurring,
		IsAutoGenerated:    &autoGenerated,
		Items:              items,
	}

	// Create invoice with items (GORM will handle associations)
	if err := d.dbPort.Invoice().Create(ctx, invoice); err != nil {
		return nil, stacktrace.Propagate(err, "failed to create monthly invoice")
	}

	// Fetch created invoice
	createdInvoice, err := d.GetByID(ctx, invoice.ID.String())
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to get created invoice")
	}

	return createdInvoice, nil
}

func (d *domain) CalculateLateFee(ctx context.Context, id string) (*model.Invoice, error) {
	// Get invoice
	invoice, err := d.dbPort.Invoice().FindByID(ctx, id)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to get invoice")
	}

	// Check if overdue
	if !invoice.IsOverdue() {
		return invoice, nil // Not overdue, no late fee needed
	}

	// Check if late fee already added
	if invoice.LateFee > 0 {
		return invoice, nil // Late fee already added
	}

	// Get late fee amount (default 10000)
	lateFeeAmount := 10000.0

	// Add late fee to invoice
	invoice.LateFee = lateFeeAmount
	invoice.TotalAmount += lateFeeAmount

	// Update payment status
	invoice.UpdatePaymentStatus()

	// Update invoice
	if err := d.dbPort.Invoice().Update(ctx, invoice); err != nil {
		return nil, stacktrace.Propagate(err, "failed to update invoice with late fee")
	}

	// Fetch updated invoice
	updatedInvoice, err := d.GetByID(ctx, id)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to get updated invoice")
	}

	return updatedInvoice, nil
}

func (d *domain) MarkAsPaid(ctx context.Context, id string, paymentID *uuid.UUID) error {
	// Get invoice
	invoice, err := d.dbPort.Invoice().FindByID(ctx, id)
	if err != nil {
		return stacktrace.Propagate(err, "failed to get invoice")
	}

	// Mark as paid
	invoice.PaidAmount = invoice.TotalAmount
	now := time.Now()
	invoice.PaymentDate = &now
	invoice.UpdatePaymentStatus()

	// Update invoice
	if err := d.dbPort.Invoice().Update(ctx, invoice); err != nil {
		return stacktrace.Propagate(err, "failed to mark invoice as paid")
	}

	return nil
}

func (d *domain) AddPayment(ctx context.Context, id string, amount float64) (*model.Invoice, error) {
	// Get invoice
	invoice, err := d.dbPort.Invoice().FindByID(ctx, id)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to get invoice")
	}

	// Add payment amount
	invoice.PaidAmount += amount

	// If fully paid, set payment date
	if invoice.PaidAmount >= invoice.TotalAmount {
		now := time.Now()
		invoice.PaymentDate = &now
	}

	// Update payment status
	invoice.UpdatePaymentStatus()

	// Update invoice
	if err := d.dbPort.Invoice().Update(ctx, invoice); err != nil {
		return nil, stacktrace.Propagate(err, "failed to add payment to invoice")
	}

	// Fetch updated invoice
	updatedInvoice, err := d.GetByID(ctx, id)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to get updated invoice")
	}

	return updatedInvoice, nil
}

// Helper methods

func (d *domain) findOverdueInvoicesByCustomer(ctx context.Context, customerID string) ([]model.Invoice, error) {
	customerUUID, err := uuid.Parse(customerID)
	if err != nil {
		return nil, stacktrace.Propagate(err, "invalid customer id")
	}

	filter := &model.InvoiceFilter{
		CustomerID: &customerUUID,
	}

	invoices, err := d.dbPort.Invoice().FindAll(ctx, filter)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to find invoices")
	}

	// Filter overdue invoices
	var overdueInvoices []model.Invoice
	for _, inv := range invoices {
		if inv.IsOverdue() {
			overdueInvoices = append(overdueInvoices, inv)
		}
	}

	return overdueInvoices, nil
}
