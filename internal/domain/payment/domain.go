package payment

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	"go-template/internal/model"
	outbound_port "go-template/internal/port/outbound"
	"go-template/utils/log"
	"go-template/utils/pdf"
	"go-template/utils/xendit"
)

type PaymentDomain interface {
	CreatePayment(ctx context.Context, input model.PaymentInput) (*model.Payment, error)
	GetPayment(ctx context.Context, id string) (*model.Payment, error)
	ListPayments(ctx context.Context, filter model.PaymentFilter) ([]model.Payment, error)
	UpdatePayment(ctx context.Context, id string, input model.PaymentInput) (*model.Payment, error)
	DeletePayment(ctx context.Context, id string) error
	ProcessWebhook(ctx context.Context, webhookData *xendit.WebhookData) error
	AllocatePayment(ctx context.Context, paymentID string, invoiceID string, amount float64) error
	GetUnpaidInvoices(ctx context.Context, customerID string) ([]model.Invoice, error)
	CreateXenditInvoice(ctx context.Context, invoice *model.Invoice, customer *model.Customer) (*xendit.CreateInvoiceResponse, error)
	GenerateReceipt(ctx context.Context, paymentID string) ([]byte, error)
	GetPaymentHistory(ctx context.Context, customerID string, filter model.PaymentFilter) ([]model.Payment, int64, error)
	GetPaymentStatistics(ctx context.Context, customerID string) (*PaymentStatistics, error)
}

type domain struct {
	dbPort       outbound_port.DatabasePort
	xenditClient *xendit.Client
}

func NewPaymentDomain(
	dbPort outbound_port.DatabasePort,
	xenditClient *xendit.Client,
) PaymentDomain {
	return &domain{
		dbPort:       dbPort,
		xenditClient: xenditClient,
	}
}

func (d *domain) CreatePayment(ctx context.Context, input model.PaymentInput) (*model.Payment, error) {
	customer, err := d.dbPort.Customer().FindByID(input.CustomerID.String())
	if err != nil {
		return nil, fmt.Errorf("customer not found: %w", err)
	}

	payment := &model.Payment{
		PaymentNumber:        "",
		CustomerID:           input.CustomerID,
		InvoiceID:            input.InvoiceID,
		Amount:               *input.Amount,
		AllocatedAmount:      0,
		PaymentMethod:        *input.PaymentMethod,
		PaymentDate:          *input.PaymentDate,
		BankName:             input.BankName,
		BankAccountNumber:    input.BankAccountNumber,
		BankAccountName:      input.BankAccountName,
		TransactionReference: input.TransactionReference,
		EwalletProvider:      input.EwalletProvider,
		EwalletNumber:        input.EwalletNumber,
		ProofImage:           input.ProofImage,
		ReceiptNumber:        input.ReceiptNumber,
		Status:               "pending",
		Notes:                input.Notes,
	}

	model.PaymentPrepare(payment)

	err = d.dbPort.Payment().Create(payment)
	if err != nil {
		return nil, fmt.Errorf("failed to create payment: %w", err)
	}

	log.WithContext(ctx).Info(fmt.Sprintf("Created payment %s for customer %s", payment.PaymentNumber, customer.CustomerCode))

	return payment, nil
}

func (d *domain) GetPayment(ctx context.Context, id string) (*model.Payment, error) {
	if id == "" {
		return nil, errors.New("payment ID is required")
	}

	payment, err := d.dbPort.Payment().FindByID(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get payment: %w", err)
	}

	return payment, nil
}

func (d *domain) ListPayments(ctx context.Context, filter model.PaymentFilter) ([]model.Payment, error) {
	payments, err := d.dbPort.Payment().Find(filter)
	if err != nil {
		return nil, fmt.Errorf("failed to list payments: %w", err)
	}

	return payments, nil
}

func (d *domain) UpdatePayment(ctx context.Context, id string, input model.PaymentInput) (*model.Payment, error) {
	payment, err := d.dbPort.Payment().FindByID(id)
	if err != nil {
		return nil, fmt.Errorf("payment not found: %w", err)
	}

	if input.PaymentMethod != nil {
		payment.PaymentMethod = *input.PaymentMethod
	}
	if input.PaymentDate != nil {
		payment.PaymentDate = *input.PaymentDate
	}
	if input.BankName != nil {
		payment.BankName = input.BankName
	}
	if input.BankAccountNumber != nil {
		payment.BankAccountNumber = input.BankAccountNumber
	}
	if input.BankAccountName != nil {
		payment.BankAccountName = input.BankAccountName
	}
	if input.TransactionReference != nil {
		payment.TransactionReference = input.TransactionReference
	}
	if input.EwalletProvider != nil {
		payment.EwalletProvider = input.EwalletProvider
	}
	if input.EwalletNumber != nil {
		payment.EwalletNumber = input.EwalletNumber
	}
	if input.ProofImage != nil {
		payment.ProofImage = input.ProofImage
	}
	if input.ReceiptNumber != nil {
		payment.ReceiptNumber = input.ReceiptNumber
	}
	if input.Notes != nil {
		payment.Notes = input.Notes
	}

	err = d.dbPort.Payment().Update(payment)
	if err != nil {
		return nil, fmt.Errorf("failed to update payment: %w", err)
	}

	log.WithContext(ctx).Info(fmt.Sprintf("Updated payment %s", payment.PaymentNumber))

	return payment, nil
}

func (d *domain) DeletePayment(ctx context.Context, id string) error {
	if id == "" {
		return errors.New("payment ID is required")
	}

	err := d.dbPort.Payment().Delete(id)
	if err != nil {
		return fmt.Errorf("failed to delete payment: %w", err)
	}

	log.WithContext(ctx).Info(fmt.Sprintf("Deleted payment %s", id))

	return nil
}

func (d *domain) ProcessWebhook(ctx context.Context, webhookData *xendit.WebhookData) error {
	if webhookData == nil {
		return errors.New("webhook data is required")
	}

	if xendit.IsPaymentPaid(webhookData.Status) {
		payment, err := d.dbPort.Payment().FindByXenditExternalID(webhookData.ExternalID)
		if err != nil {
			return fmt.Errorf("payment not found for external ID %s: %w", webhookData.ExternalID, err)
		}

		if payment.Status == "pending" {
			payment.Status = "confirmed"
			payment.AllocatedAmount = webhookData.PaidAmount
			payment.TransactionReference = &webhookData.PaymentID
			payment.PaymentMethod = webhookData.PaymentMethod

			err = d.dbPort.Payment().Update(payment)
			if err != nil {
				return fmt.Errorf("failed to update payment status: %w", err)
			}

			if payment.InvoiceID != nil {
				err = d.AllocatePayment(ctx, payment.ID.String(), payment.InvoiceID.String(), webhookData.PaidAmount)
				if err != nil {
					log.WithContext(ctx).Error(fmt.Sprintf("Failed to allocate payment %s: %v", payment.PaymentNumber, err))
				}
			}

			log.WithContext(ctx).Info(fmt.Sprintf("Payment %s confirmed via webhook", payment.PaymentNumber))
		}
	}

	return nil
}

func (d *domain) AllocatePayment(ctx context.Context, paymentID string, invoiceID string, amount float64) error {
	if paymentID == "" || invoiceID == "" {
		return errors.New("payment ID and invoice ID are required")
	}

	if amount <= 0 {
		return errors.New("allocation amount must be greater than 0")
	}

	payment, err := d.dbPort.Payment().FindByID(paymentID)
	if err != nil {
		return fmt.Errorf("payment not found: %w", err)
	}

	if payment.AllocatedAmount+amount > payment.Amount {
		return errors.New("allocation amount exceeds payment amount")
	}

	invoiceIDUUID, err := uuid.Parse(invoiceID)
	if err != nil {
		return fmt.Errorf("invalid invoice ID: %w", err)
	}

	allocation := &model.PaymentAllocation{
		PaymentID:       payment.ID,
		InvoiceID:       invoiceIDUUID,
		AllocatedAmount: amount,
	}

	err = d.dbPort.PaymentAllocation().Create(allocation)
	if err != nil {
		return fmt.Errorf("failed to create payment allocation: %w", err)
	}

	payment.AllocatedAmount += amount
	err = d.dbPort.Payment().Update(payment)
	if err != nil {
		return fmt.Errorf("failed to update payment allocation: %w", err)
	}

	log.WithContext(ctx).Info(fmt.Sprintf("Allocated %f from payment %s to invoice %s", amount, payment.PaymentNumber, invoiceID))

	return nil
}

func (d *domain) GetUnpaidInvoices(ctx context.Context, customerID string) ([]model.Invoice, error) {
	if customerID == "" {
		return nil, errors.New("customer ID is required")
	}

	return []model.Invoice{}, nil
}

func (d *domain) CreateXenditInvoice(ctx context.Context, invoice *model.Invoice, customer *model.Customer) (*xendit.CreateInvoiceResponse, error) {
	if d.xenditClient == nil {
		return nil, errors.New("xendit client is not configured")
	}

	customerEmail := ""
	if customer.Email != nil {
		customerEmail = *customer.Email
	}
	if customerEmail == "" {
		return nil, errors.New("customer email is required for payment")
	}

	customerPhone := customer.Phone

	req := xendit.CreateInvoiceRequest{
		ExternalID:      invoice.ID.String(),
		Amount:          invoice.TotalAmount,
		InvoiceDuration: 86400,
		Description:     fmt.Sprintf("Invoice %s - %s", invoice.InvoiceNumber, customer.FullName),
		Customer: xendit.CustomerDetail{
			GivenNames:   customer.FullName,
			Email:        customerEmail,
			MobileNumber: customerPhone,
		},
		CustomerNotificationPreference: "email_only",
		PaymentMethods:                 []string{"VA", "BCA", "MANDIRI", "BRI", "GOPAY", "OVO", "DANA", "QRIS"},
		ShouldSendEmail:                true,
		ShouldSendSMS:                  false,
	}

	response, err := d.xenditClient.CreateInvoice(req)
	if err != nil {
		return nil, fmt.Errorf("failed to create xendit invoice: %w", err)
	}

	log.WithContext(ctx).Info(fmt.Sprintf("Created xendit invoice for %s: %s", invoice.InvoiceNumber, response.ID))

	return response, nil
}

func (d *domain) GenerateReceipt(ctx context.Context, paymentID string) ([]byte, error) {
	if paymentID == "" {
		return nil, errors.New("payment ID is required")
	}

	// Get payment
	payment, err := d.dbPort.Payment().FindByID(paymentID)
	if err != nil {
		return nil, fmt.Errorf("payment not found: %w", err)
	}

	// Get customer
	customer, err := d.dbPort.Customer().FindByID(payment.CustomerID.String())
	if err != nil {
		return nil, fmt.Errorf("customer not found: %w", err)
	}

	// Get invoice if available
	var invoice *model.Invoice
	var invoiceItems []model.InvoiceItem
	if payment.InvoiceID != nil {
		invoice, err = d.dbPort.Invoice().FindByID(payment.InvoiceID.String())
		if err == nil {
			invoiceItems, err = d.dbPort.InvoiceItem().FindByInvoiceID(invoice.ID.String())
			if err != nil {
				log.WithContext(ctx).Warn(fmt.Sprintf("Failed to get invoice items: %v", err))
			}
		}
	}

	// Get company info from system settings
	companyName := d.getSystemSetting("company.name", "PT Internet Provider")
	companyAddress := d.getSystemSetting("company.address", "Jl. Raya No. 123")
	companyPhone := d.getSystemSetting("company.phone", "021-12345678")
	companyEmail := d.getSystemSetting("company.email", "info@isp.com")
	companyTaxID := d.getSystemSetting("company.tax_id", "")

	// Prepare receipt data
	receiptData := pdf.ReceiptData{
		ReceiptNumber: payment.PaymentNumber,
		ReceiptDate:   payment.CreatedAt,
		InvoiceNumber: "",
		CompanyInfo: pdf.CompanyInfo{
			Name:    companyName,
			Address: companyAddress,
			Phone:   companyPhone,
			Email:   companyEmail,
			TaxID:   companyTaxID,
		},
		CustomerInfo: pdf.CustomerInfo{
			Name:    customer.FullName,
			Email:   getStringValue(customer.Email),
			Phone:   customer.Phone,
			Address: getStringValue(customer.Address),
		},
		PaymentInfo: pdf.PaymentInfo{
			PaymentDate:          payment.PaymentDate,
			PaymentMethod:        payment.PaymentMethod,
			Status:               payment.Status,
			BankName:             getStringValue(payment.BankName),
			TransactionReference: getStringValue(payment.TransactionReference),
			Subtotal:             payment.Amount,
			TaxAmount:            0,
			TaxLabel:             "PPN 11%",
			DiscountAmount:       0,
			TotalAmount:          payment.Amount,
		},
		InvoiceItems: []pdf.InvoiceItem{},
	}

	if invoice != nil {
		receiptData.InvoiceNumber = invoice.InvoiceNumber
		receiptData.PaymentInfo.Subtotal = invoice.Subtotal
		receiptData.PaymentInfo.TaxAmount = invoice.TaxAmount
		receiptData.PaymentInfo.DiscountAmount = invoice.DiscountAmount
		receiptData.PaymentInfo.TotalAmount = invoice.TotalAmount

		for _, item := range invoiceItems {
			receiptData.InvoiceItems = append(receiptData.InvoiceItems, pdf.InvoiceItem{
				Description: item.Description,
				Quantity:    item.Quantity,
				UnitPrice:   item.UnitPrice,
				Total:       item.Total,
			})
		}
	} else {
		// If no invoice, create a simple item
		receiptData.InvoiceItems = append(receiptData.InvoiceItems, pdf.InvoiceItem{
			Description: "Payment",
			Quantity:    1,
			UnitPrice:   payment.Amount,
			Total:       payment.Amount,
		})
	}

	// Generate PDF
	generator := pdf.NewPDFGenerator()
	pdfBytes, err := generator.GenerateReceipt(receiptData)
	if err != nil {
		return nil, fmt.Errorf("failed to generate receipt PDF: %w", err)
	}

	log.WithContext(ctx).Info(fmt.Sprintf("Generated receipt for payment %s", payment.PaymentNumber))

	return pdfBytes, nil
}

func (d *domain) GetPaymentHistory(ctx context.Context, customerID string, filter model.PaymentFilter) ([]model.Payment, int64, error) {
	if customerID == "" {
		return nil, 0, errors.New("customer ID is required")
	}

	// Set customer ID filter
	customerUUID, err := uuid.Parse(customerID)
	if err != nil {
		return nil, 0, fmt.Errorf("invalid customer ID: %w", err)
	}
	filter.CustomerID = &customerUUID

	// Get payments
	payments, err := d.dbPort.Payment().Find(filter)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get payment history: %w", err)
	}

	// Get total count
	totalCount, err := d.dbPort.Payment().Count(filter)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count payments: %w", err)
	}

	log.WithContext(ctx).Info(fmt.Sprintf("Retrieved %d payment history records for customer %s", len(payments), customerID))

	return payments, totalCount, nil
}

func (d *domain) GetPaymentStatistics(ctx context.Context, customerID string) (*PaymentStatistics, error) {
	if customerID == "" {
		return nil, errors.New("customer ID is required")
	}

	customerUUID, err := uuid.Parse(customerID)
	if err != nil {
		return nil, fmt.Errorf("invalid customer ID: %w", err)
	}

	// Get all confirmed payments
	filter := model.PaymentFilter{
		CustomerID: &customerUUID,
		Status:     []string{"confirmed"},
	}

	payments, err := d.dbPort.Payment().Find(filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get payments for statistics: %w", err)
	}

	// Calculate statistics
	stats := &PaymentStatistics{
		TotalPayments:     int64(len(payments)),
		TotalAmount:       0,
		AverageAmount:     0,
		LastPaymentDate:   nil,
		LastPaymentAmount: 0,
	}

	if len(payments) > 0 {
		for _, payment := range payments {
			stats.TotalAmount += payment.Amount
		}
		stats.AverageAmount = stats.TotalAmount / float64(len(payments))

		// Get last payment
		lastPayment := payments[0]
		for _, p := range payments {
			if p.PaymentDate.After(lastPayment.PaymentDate) {
				lastPayment = p
			}
		}
		stats.LastPaymentDate = &lastPayment.PaymentDate
		stats.LastPaymentAmount = lastPayment.Amount
	}

	log.WithContext(ctx).Info(fmt.Sprintf("Generated payment statistics for customer %s", customerID))

	return stats, nil
}

// Helper functions

func (d *domain) getSystemSetting(key string, defaultValue string) string {
	setting, err := d.dbPort.SystemSetting().FindByKey(key)
	if err != nil || setting == nil {
		return defaultValue
	}
	if setting.Value == nil {
		return defaultValue
	}
	return *setting.Value
}

func getStringValue(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}


// PaymentStatistics represents payment statistics for a customer
type PaymentStatistics struct {
	TotalPayments     int64     `json:"total_payments"`
	TotalAmount       float64   `json:"total_amount"`
	AverageAmount     float64   `json:"average_amount"`
	LastPaymentDate   *time.Time `json:"last_payment_date"`
	LastPaymentAmount float64   `json:"last_payment_amount"`
}
