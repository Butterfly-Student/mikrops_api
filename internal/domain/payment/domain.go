package payment

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/palantir/stacktrace"

	"go-template/internal/domain/customer"
	"go-template/internal/domain/invoice"
	"go-template/internal/model"
	outbound_port "go-template/internal/port/outbound"
)

type PaymentDomain interface {
	Create(ctx context.Context, input model.PaymentInput) (*model.Payment, error)
	GetByID(ctx context.Context, id string) (*model.Payment, error)
	GetByNumber(ctx context.Context, number string) (*model.Payment, error)
	List(ctx context.Context, filter *model.PaymentFilter) ([]model.Payment, error)
	Update(ctx context.Context, id string, input model.PaymentInput) (*model.Payment, error)
	Delete(ctx context.Context, id string) error
	Confirm(ctx context.Context, id string, userID string) (*model.Payment, error)
	Reject(ctx context.Context, id string, userID string, reason string) (*model.Payment, error)
	AllocateToInvoice(ctx context.Context, paymentID string, invoiceID string, amount float64) (*model.PaymentAllocation, error)
	GeneratePaymentNumber(ctx context.Context) (string, error)
}

type domain struct {
	dbPort          outbound_port.DatabasePort
	invoiceDomain   invoice.InvoiceDomain
	customerDomain  customer.CustomerDomain
}

func NewPaymentDomain(
	dbPort outbound_port.DatabasePort,
	invoiceDomain invoice.InvoiceDomain,
	customerDomain customer.CustomerDomain,
) PaymentDomain {
	return &domain{
		dbPort:         dbPort,
		invoiceDomain:  invoiceDomain,
		customerDomain: customerDomain,
	}
}

func (d *domain) Create(ctx context.Context, input model.PaymentInput) (*model.Payment, error) {
	// Generate payment number if not provided
	paymentNumber := input.PaymentNumber
	if paymentNumber == "" {
		number, err := d.GeneratePaymentNumber(ctx)
		if err != nil {
			return nil, stacktrace.Propagate(err, "failed to generate payment number")
		}
		paymentNumber = number
		input.PaymentNumber = number
	}

	// Convert input to model
	payment := input.ToModel()
	payment.ID = uuid.New()

	// If invoice_id provided, validate and set allocation
	if payment.InvoiceID != nil {
		// Get invoice to validate
		inv, err := d.invoiceDomain.GetByID(ctx, payment.InvoiceID.String())
		if err != nil {
			return nil, stacktrace.Propagate(err, "failed to get invoice")
		}

		// Check if amount doesn't exceed invoice balance
		if payment.Amount > inv.GetBalance() {
			return nil, stacktrace.NewError("payment amount exceeds invoice balance")
		}

		// Set allocated amount
		payment.AllocatedAmount = payment.Amount

		// Create payment allocation (will be saved with payment via associations)
		payment.Allocations = []model.PaymentAllocation{
			{
				ID:              uuid.New(),
				PaymentID:       payment.ID,
				InvoiceID:       *payment.InvoiceID,
				AllocatedAmount: payment.Amount,
			},
		}
	}

	// Create payment (with allocations via GORM associations)
	if err := d.dbPort.Payment().Create(ctx, payment); err != nil {
		return nil, stacktrace.Propagate(err, "failed to create payment")
	}

	// Fetch created payment
	createdPayment, err := d.GetByID(ctx, payment.ID.String())
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to get created payment")
	}

	return createdPayment, nil
}

func (d *domain) GetByID(ctx context.Context, id string) (*model.Payment, error) {
	payment, err := d.dbPort.Payment().FindByID(ctx, id)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to get payment by id")
	}

	return payment, nil
}

func (d *domain) GetByNumber(ctx context.Context, number string) (*model.Payment, error) {
	payment, err := d.dbPort.Payment().FindByNumber(ctx, number)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to get payment by number")
	}

	return payment, nil
}

func (d *domain) List(ctx context.Context, filter *model.PaymentFilter) ([]model.Payment, error) {
	payments, err := d.dbPort.Payment().FindAll(ctx, filter)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to list payments")
	}

	return payments, nil
}

func (d *domain) Update(ctx context.Context, id string, input model.PaymentInput) (*model.Payment, error) {
	// Get existing payment
	payment, err := d.dbPort.Payment().FindByID(ctx, id)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to find payment")
	}

	// Update fields
	payment.PaymentNumber = input.PaymentNumber
	payment.CustomerID = input.CustomerID
	payment.InvoiceID = input.InvoiceID
	payment.Amount = input.Amount
	payment.PaymentMethod = model.PaymentMethod(input.PaymentMethod)
	payment.PaymentDate = input.PaymentDate
	payment.BankName = input.BankName
	payment.BankAccountNumber = input.BankAccountNumber
	payment.BankAccountName = input.BankAccountName
	payment.TransactionReference = input.TransactionReference
	payment.EwalletProvider = input.EwalletProvider
	payment.EwalletNumber = input.EwalletNumber
	payment.XenditInvoiceID = input.XenditInvoiceID
	payment.XenditExternalID = input.XenditExternalID
	payment.XenditPaymentChannel = input.XenditPaymentChannel
	payment.ProofImage = input.ProofImage
	payment.ReceiptNumber = input.ReceiptNumber
	payment.Notes = input.Notes

	if input.Status != nil {
		payment.Status = model.PaymentStatusType(*input.Status)
	}

	// Update payment
	if err := d.dbPort.Payment().Update(ctx, payment); err != nil {
		return nil, stacktrace.Propagate(err, "failed to update payment")
	}

	// Fetch updated payment
	updatedPayment, err := d.GetByID(ctx, id)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to get updated payment")
	}

	return updatedPayment, nil
}

func (d *domain) Delete(ctx context.Context, id string) error {
	if err := d.dbPort.Payment().Delete(ctx, id); err != nil {
		return stacktrace.Propagate(err, "failed to delete payment")
	}

	return nil
}

func (d *domain) Confirm(ctx context.Context, id string, userID string) (*model.Payment, error) {
	// Get payment with allocations
	payment, err := d.dbPort.Payment().FindByID(ctx, id)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to get payment")
	}

	// Validate status
	if !payment.IsPending() {
		return nil, stacktrace.NewError("payment is not in pending status")
	}

	// Parse user ID
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, stacktrace.Propagate(err, "invalid user id")
	}

	// Update payment status
	payment.Status = model.PaymentStatusTypeConfirmed
	payment.ProcessedBy = &userUUID
	now := time.Now()
	payment.ProcessedAt = &now

	if err := d.dbPort.Payment().Update(ctx, payment); err != nil {
		return nil, stacktrace.Propagate(err, "failed to update payment status")
	}

	// Process allocations - update invoices
	for _, allocation := range payment.Allocations {
		// Add payment to invoice
		if _, err := d.invoiceDomain.AddPayment(ctx, allocation.InvoiceID.String(), allocation.AllocatedAmount); err != nil {
			return nil, stacktrace.Propagate(err, "failed to add payment to invoice")
		}
	}

	// Check customer status - un-isolate if all invoices paid
	customer, err := d.dbPort.Customer().FindByID(ctx, payment.CustomerID.String())
	if err == nil && customer.IsIsolated() {
		// Check if customer has any unpaid invoices
		customerUUID := customer.ID
		filter := &model.InvoiceFilter{
			CustomerID: &customerUUID,
		}
		invoices, err := d.invoiceDomain.List(ctx, filter)
		if err == nil {
			allPaid := true
			for _, inv := range invoices {
				if !inv.IsPaid() {
					allPaid = false
					break
				}
			}

			// If all paid, un-isolate customer
			if allPaid {
				if err := d.customerDomain.UnIsolate(ctx, customer.ID.String()); err != nil {
					// Log error but don't fail confirmation
					// Customer can be un-isolated manually
				}
			}
		}
	}

	// Fetch updated payment
	updatedPayment, err := d.GetByID(ctx, id)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to get updated payment")
	}

	return updatedPayment, nil
}

func (d *domain) Reject(ctx context.Context, id string, userID string, reason string) (*model.Payment, error) {
	// Get payment
	payment, err := d.dbPort.Payment().FindByID(ctx, id)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to get payment")
	}

	// Validate status
	if !payment.IsPending() {
		return nil, stacktrace.NewError("payment is not in pending status")
	}

	// Parse user ID
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, stacktrace.Propagate(err, "invalid user id")
	}

	// Update payment status
	payment.Status = model.PaymentStatusTypeRejected
	payment.ProcessedBy = &userUUID
	now := time.Now()
	payment.ProcessedAt = &now
	payment.RejectionReason = &reason

	if err := d.dbPort.Payment().Update(ctx, payment); err != nil {
		return nil, stacktrace.Propagate(err, "failed to reject payment")
	}

	// Fetch updated payment
	updatedPayment, err := d.GetByID(ctx, id)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to get updated payment")
	}

	return updatedPayment, nil
}

func (d *domain) AllocateToInvoice(ctx context.Context, paymentID string, invoiceID string, amount float64) (*model.PaymentAllocation, error) {
	// Get payment
	payment, err := d.dbPort.Payment().FindByID(ctx, paymentID)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to get payment")
	}

	// Check remaining amount
	if !payment.CanAllocate(amount) {
		return nil, stacktrace.NewError("payment has insufficient remaining amount")
	}

	// Get invoice
	inv, err := d.invoiceDomain.GetByID(ctx, invoiceID)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to get invoice")
	}

	// Check invoice balance
	if amount > inv.GetBalance() {
		return nil, stacktrace.NewError("amount exceeds invoice balance")
	}

	// Parse IDs
	paymentUUID, err := uuid.Parse(paymentID)
	if err != nil {
		return nil, stacktrace.Propagate(err, "invalid payment id")
	}

	invoiceUUID, err := uuid.Parse(invoiceID)
	if err != nil {
		return nil, stacktrace.Propagate(err, "invalid invoice id")
	}

	// Create allocation
	allocation := &model.PaymentAllocation{
		ID:              uuid.New(),
		PaymentID:       paymentUUID,
		InvoiceID:       invoiceUUID,
		AllocatedAmount: amount,
	}

	if err := d.dbPort.Payment().CreateAllocation(ctx, allocation); err != nil {
		return nil, stacktrace.Propagate(err, "failed to create allocation")
	}

	// Update payment allocated amount
	payment.AllocatedAmount += amount
	if err := d.dbPort.Payment().Update(ctx, payment); err != nil {
		return nil, stacktrace.Propagate(err, "failed to update payment")
	}

	// Add payment to invoice
	if _, err := d.invoiceDomain.AddPayment(ctx, invoiceID, amount); err != nil {
		return nil, stacktrace.Propagate(err, "failed to add payment to invoice")
	}

	return allocation, nil
}

func (d *domain) GeneratePaymentNumber(ctx context.Context) (string, error) {
	now := time.Now()
	year := now.Year()
	month := int(now.Month())

	// Get last payment number for this month
	lastNumber, err := d.dbPort.Payment().GetLastPaymentNumber(ctx, year, month)
	if err != nil {
		// If no payment found, start from 1
		lastNumber = ""
	}

	// Parse sequence number
	sequence := 1
	if lastNumber != "" {
		// Format: PAY/YYYY/MM/XXXXX
		// Extract last part
		var seq int
		_, err := fmt.Sscanf(lastNumber, "PAY/%d/%d/%d", &year, &month, &seq)
		if err == nil {
			sequence = seq + 1
		}
	}

	// Generate new payment number
	paymentNumber := fmt.Sprintf("PAY/%d/%02d/%05d", year, month, sequence)
	return paymentNumber, nil
}
