package invoice

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/palantir/stacktrace"

	"mikrops/internal/model"
	outbound_port "mikrops/internal/port/outbound"
)

type InvoiceDomain interface {
	Create(ctx context.Context, input model.InvoiceInput) (model.Invoice, error)
	FindByFilter(ctx context.Context, filter model.InvoiceFilter) ([]model.Invoice, error)
	FindByID(ctx context.Context, id string) (model.Invoice, error)
	Update(ctx context.Context, id string, input model.InvoiceInput) error
	Delete(ctx context.Context, id string) error
	GenerateBulk(ctx context.Context, tenantID string, periodStart, periodEnd time.Time) error
}

type invoiceDomain struct {
	databasePort outbound_port.DatabasePort
	messagePort  outbound_port.MessagePort
	cachePort    outbound_port.CachePort
	workflowPort outbound_port.WorkflowPort
}

func NewInvoiceDomain(
	databasePort outbound_port.DatabasePort,
	messagePort outbound_port.MessagePort,
	cachePort outbound_port.CachePort,
	workflowPort outbound_port.WorkflowPort,
) InvoiceDomain {
	return &invoiceDomain{
		databasePort: databasePort,
		messagePort:  messagePort,
		cachePort:    cachePort,
		workflowPort: workflowPort,
	}
}

func (d *invoiceDomain) Create(ctx context.Context, input model.InvoiceInput) (model.Invoice, error) {
	// Generate invoice number if not provided
	if input.InvoiceNumber == "" {
		input.InvoiceNumber = generateInvoiceNumber(input.TenantID)
	}

	// Set default status
	if input.Status == "" {
		input.Status = model.InvoiceStatusUnpaid
	}

	// Calculate total amount if not provided
	if input.TotalAmount == 0 {
		input.TotalAmount = input.Amount + input.TaxAmount
	}

	invoice, err := d.databasePort.Invoice().Create(input)
	if err != nil {
		return model.Invoice{}, stacktrace.Propagate(err, "failed to create invoice")
	}

	return invoice, nil
}

func (d *invoiceDomain) FindByFilter(ctx context.Context, filter model.InvoiceFilter) ([]model.Invoice, error) {
	if filter.IsEmpty() {
		return nil, stacktrace.NewError("filter is empty")
	}

	invoices, err := d.databasePort.Invoice().FindByFilter(filter)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to find invoices by filter")
	}

	return invoices, nil
}

func (d *invoiceDomain) FindByID(ctx context.Context, id string) (model.Invoice, error) {
	if id == "" {
		return model.Invoice{}, stacktrace.NewError("id is empty")
	}

	invoice, err := d.databasePort.Invoice().FindByID(id)
	if err != nil {
		return model.Invoice{}, stacktrace.Propagate(err, "failed to find invoice by id")
	}

	return invoice, nil
}

func (d *invoiceDomain) Update(ctx context.Context, id string, input model.InvoiceInput) error {
	if id == "" {
		return stacktrace.NewError("id is empty")
	}

	err := d.databasePort.Invoice().Update(id, input)
	if err != nil {
		return stacktrace.Propagate(err, "failed to update invoice")
	}

	return nil
}

func (d *invoiceDomain) Delete(ctx context.Context, id string) error {
	if id == "" {
		return stacktrace.NewError("id is empty")
	}

	err := d.databasePort.Invoice().Delete(id)
	if err != nil {
		return stacktrace.Propagate(err, "failed to delete invoice")
	}

	return nil
}

func (d *invoiceDomain) GenerateBulk(ctx context.Context, tenantID string, periodStart, periodEnd time.Time) error {
	if tenantID == "" {
		return stacktrace.NewError("tenantID is empty")
	}

	// Fetch tenant settings to get cutoff_day
	tenantSettings, err := d.databasePort.TenantSetting().FindByTenantID(tenantID)
	var cutoffDay int
	if err != nil {
		// Fallback to default 7 days if tenant settings not found
		cutoffDay = 0 // will use fallback logic below
	} else {
		cutoffDay = tenantSettings.CutoffDay
	}

	// Find all active subscriptions for this tenant
	subscriptions, err := d.databasePort.Subscription().FindByFilter(model.SubscriptionFilter{
		TenantIDs: []string{tenantID},
		Statuses:  []string{model.SubscriptionStatusActive},
	})
	if err != nil {
		return stacktrace.Propagate(err, "failed to find active subscriptions")
	}

	// Generate invoices for each subscription
	var invoiceInputs []model.InvoiceInput
	for _, sub := range subscriptions {
		// Get package to determine amount
		pkg, err := d.databasePort.InternetPackage().FindByID(sub.PackageID)
		if err != nil {
			// Log error and skip this subscription
			continue
		}

		// Calculate due date based on cutoff_day
		var dueDate time.Time
		if cutoffDay > 0 && cutoffDay <= 28 {
			// Use cutoff_day from tenant settings
			now := time.Now()
			currentDay := now.Day()
			
			// If current day < cutoff_day: due_date = current_month + cutoff_day
			// If current day >= cutoff_day: due_date = next_month + cutoff_day
			if currentDay < cutoffDay {
				dueDate = time.Date(now.Year(), now.Month(), cutoffDay, 23, 59, 59, 0, now.Location())
			} else {
				// Next month
				nextMonth := now.AddDate(0, 1, 0)
				dueDate = time.Date(nextMonth.Year(), nextMonth.Month(), cutoffDay, 23, 59, 59, 0, nextMonth.Location())
			}
		} else {
			// Fallback to default 7 days after period end
			dueDate = periodEnd.AddDate(0, 0, 7)
		}

		// Generate invoice
		invoiceInput := model.InvoiceInput{
			TenantID:       tenantID,
			CustomerID:     sub.CustomerID,
			SubscriptionID: sub.ID,
			InvoiceNumber:  generateInvoiceNumber(tenantID),
			Amount:         pkg.Price,
			TaxAmount:      0, // TODO: Calculate tax if needed
			TotalAmount:    pkg.Price,
			Status:         model.InvoiceStatusUnpaid,
			DueDate:        dueDate,
			PeriodStart:    periodStart,
			PeriodEnd:      periodEnd,
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		}

		invoiceInputs = append(invoiceInputs, invoiceInput)
	}

	// Bulk create invoices
	if len(invoiceInputs) > 0 {
		err = d.databasePort.Invoice().BulkCreate(invoiceInputs)
		if err != nil {
			return stacktrace.Propagate(err, "failed to bulk create invoices")
		}
	}

	return nil
}

// generateInvoiceNumber generates a unique invoice number
func generateInvoiceNumber(tenantID string) string {
	// Get short tenant ID (first 8 chars)
	shortTenant := tenantID
	if len(tenantID) > 8 {
		shortTenant = tenantID[:8]
	}

	// Generate timestamp
	timestamp := time.Now().Format("20060102")

	// Generate random suffix
	b := make([]byte, 3)
	rand.Read(b)
	random := hex.EncodeToString(b)

	return fmt.Sprintf("INV-%s-%s-%s", shortTenant, timestamp, random)
}
