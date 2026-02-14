package pdf

import (
	"bytes"
	"fmt"
	"time"

	"github.com/jung-kurt/gofpdf"
	"github.com/palantir/stacktrace"
)

// PDFGenerator handles PDF generation
type PDFGenerator struct {
	pdf *gofpdf.Fpdf
}

// NewPDFGenerator creates a new PDF generator
func NewPDFGenerator() *PDFGenerator {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetAutoPageBreak(true, 15)

	return &PDFGenerator{
		pdf: pdf,
	}
}

// GenerateReceipt generates a payment receipt PDF
func (g *PDFGenerator) GenerateReceipt(data ReceiptData) ([]byte, error) {
	g.pdf.AddPage()

	// Add company logo/header
	g.addHeader(data.CompanyInfo)

	// Add receipt title
	g.addTitle("PAYMENT RECEIPT")

	// Add receipt info
	g.addReceiptInfo(data)

	// Add customer info
	g.addCustomerInfo(data.CustomerInfo)

	// Add payment details
	g.addPaymentDetails(data.PaymentInfo)

	// Add invoice items
	g.addInvoiceItems(data.InvoiceItems)

	// Add totals
	g.addTotals(data.PaymentInfo)

	// Add footer
	g.addFooter()

	// Generate PDF bytes
	var buf bytes.Buffer
	err := g.pdf.Output(&buf)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to generate PDF")
	}

	return buf.Bytes(), nil
}

func (g *PDFGenerator) addHeader(company CompanyInfo) {
	g.pdf.SetFont("Arial", "B", 20)
	g.pdf.Cell(190, 10, company.Name)
	g.pdf.Ln(8)

	g.pdf.SetFont("Arial", "", 10)
	g.pdf.Cell(190, 5, company.Address)
	g.pdf.Ln(5)
	g.pdf.Cell(190, 5, fmt.Sprintf("Phone: %s | Email: %s", company.Phone, company.Email))
	g.pdf.Ln(5)
	if company.TaxID != "" {
		g.pdf.Cell(190, 5, fmt.Sprintf("Tax ID: %s", company.TaxID))
		g.pdf.Ln(5)
	}

	g.pdf.Ln(5)
}

func (g *PDFGenerator) addTitle(title string) {
	g.pdf.SetFont("Arial", "B", 16)
	g.pdf.SetFillColor(240, 240, 240)
	g.pdf.CellFormat(190, 10, title, "0", 1, "C", true, 0, "")
	g.pdf.Ln(8)
}

func (g *PDFGenerator) addReceiptInfo(data ReceiptData) {
	g.pdf.SetFont("Arial", "", 10)

	// Left column
	g.pdf.Cell(95, 5, fmt.Sprintf("Receipt Number: %s", data.ReceiptNumber))

	// Right column
	g.pdf.Cell(95, 5, fmt.Sprintf("Date: %s", data.ReceiptDate.Format("02 Jan 2006")))
	g.pdf.Ln(5)

	g.pdf.Cell(95, 5, fmt.Sprintf("Invoice Number: %s", data.InvoiceNumber))
	g.pdf.Cell(95, 5, fmt.Sprintf("Payment Date: %s", data.PaymentInfo.PaymentDate.Format("02 Jan 2006")))
	g.pdf.Ln(8)
}

func (g *PDFGenerator) addCustomerInfo(customer CustomerInfo) {
	g.pdf.SetFont("Arial", "B", 12)
	g.pdf.Cell(190, 6, "Bill To:")
	g.pdf.Ln(6)

	g.pdf.SetFont("Arial", "", 10)
	g.pdf.Cell(190, 5, customer.Name)
	g.pdf.Ln(5)
	if customer.Email != "" {
		g.pdf.Cell(190, 5, fmt.Sprintf("Email: %s", customer.Email))
		g.pdf.Ln(5)
	}
	if customer.Phone != "" {
		g.pdf.Cell(190, 5, fmt.Sprintf("Phone: %s", customer.Phone))
		g.pdf.Ln(5)
	}
	if customer.Address != "" {
		g.pdf.MultiCell(190, 5, fmt.Sprintf("Address: %s", customer.Address), "", "", false)
	}
	g.pdf.Ln(5)
}

func (g *PDFGenerator) addPaymentDetails(payment PaymentInfo) {
	g.pdf.SetFont("Arial", "B", 12)
	g.pdf.Cell(190, 6, "Payment Details:")
	g.pdf.Ln(6)

	g.pdf.SetFont("Arial", "", 10)
	g.pdf.Cell(95, 5, fmt.Sprintf("Payment Method: %s", payment.PaymentMethod))
	g.pdf.Cell(95, 5, fmt.Sprintf("Status: %s", payment.Status))
	g.pdf.Ln(5)

	if payment.BankName != "" {
		g.pdf.Cell(95, 5, fmt.Sprintf("Bank: %s", payment.BankName))
		g.pdf.Ln(5)
	}

	if payment.TransactionReference != "" {
		g.pdf.Cell(190, 5, fmt.Sprintf("Transaction Ref: %s", payment.TransactionReference))
		g.pdf.Ln(5)
	}

	g.pdf.Ln(5)
}

func (g *PDFGenerator) addInvoiceItems(items []InvoiceItem) {
	// Table header
	g.pdf.SetFont("Arial", "B", 10)
	g.pdf.SetFillColor(220, 220, 220)
	g.pdf.CellFormat(80, 7, "Description", "1", 0, "L", true, 0, "")
	g.pdf.CellFormat(30, 7, "Quantity", "1", 0, "C", true, 0, "")
	g.pdf.CellFormat(40, 7, "Unit Price", "1", 0, "R", true, 0, "")
	g.pdf.CellFormat(40, 7, "Total", "1", 1, "R", true, 0, "")

	// Table rows
	g.pdf.SetFont("Arial", "", 10)
	g.pdf.SetFillColor(255, 255, 255)

	for _, item := range items {
		g.pdf.CellFormat(80, 6, item.Description, "1", 0, "L", false, 0, "")
		g.pdf.CellFormat(30, 6, fmt.Sprintf("%d", item.Quantity), "1", 0, "C", false, 0, "")
		g.pdf.CellFormat(40, 6, formatCurrency(item.UnitPrice), "1", 0, "R", false, 0, "")
		g.pdf.CellFormat(40, 6, formatCurrency(item.Total), "1", 1, "R", false, 0, "")
	}

	g.pdf.Ln(3)
}

func (g *PDFGenerator) addTotals(payment PaymentInfo) {
	// Subtotal
	g.pdf.SetFont("Arial", "", 10)
	g.pdf.Cell(150, 6, "Subtotal:")
	g.pdf.CellFormat(40, 6, formatCurrency(payment.Subtotal), "0", 1, "R", false, 0, "")

	// Tax
	if payment.TaxAmount > 0 {
		g.pdf.Cell(150, 6, fmt.Sprintf("Tax (%s):", payment.TaxLabel))
		g.pdf.CellFormat(40, 6, formatCurrency(payment.TaxAmount), "0", 1, "R", false, 0, "")
	}

	// Discount
	if payment.DiscountAmount > 0 {
		g.pdf.Cell(150, 6, "Discount:")
		g.pdf.CellFormat(40, 6, fmt.Sprintf("-%s", formatCurrency(payment.DiscountAmount)), "0", 1, "R", false, 0, "")
	}

	// Total
	g.pdf.SetFont("Arial", "B", 12)
	g.pdf.SetFillColor(240, 240, 240)
	g.pdf.CellFormat(150, 8, "Total Amount Paid:", "1", 0, "L", true, 0, "")
	g.pdf.CellFormat(40, 8, formatCurrency(payment.TotalAmount), "1", 1, "R", true, 0, "")

	g.pdf.Ln(5)
}

func (g *PDFGenerator) addFooter() {
	g.pdf.SetY(-30)
	g.pdf.SetFont("Arial", "I", 8)
	g.pdf.SetTextColor(128, 128, 128)
	g.pdf.Cell(190, 5, "Thank you for your payment!")
	g.pdf.Ln(5)
	g.pdf.Cell(190, 5, "This is a computer-generated receipt and does not require a signature.")
	g.pdf.Ln(5)
	g.pdf.Cell(190, 5, fmt.Sprintf("Generated on: %s", time.Now().Format("02 Jan 2006 15:04:05")))
}

func formatCurrency(amount float64) string {
	return fmt.Sprintf("Rp %.2f", amount)
}

// ReceiptData contains all data needed for receipt generation
type ReceiptData struct {
	ReceiptNumber  string
	ReceiptDate    time.Time
	InvoiceNumber  string
	CompanyInfo    CompanyInfo
	CustomerInfo   CustomerInfo
	PaymentInfo    PaymentInfo
	InvoiceItems   []InvoiceItem
}

// CompanyInfo contains company information
type CompanyInfo struct {
	Name    string
	Address string
	Phone   string
	Email   string
	TaxID   string
}

// CustomerInfo contains customer information
type CustomerInfo struct {
	Name    string
	Email   string
	Phone   string
	Address string
}

// PaymentInfo contains payment information
type PaymentInfo struct {
	PaymentDate          time.Time
	PaymentMethod        string
	Status               string
	BankName             string
	TransactionReference string
	Subtotal             float64
	TaxAmount            float64
	TaxLabel             string
	DiscountAmount       float64
	TotalAmount          float64
}

// InvoiceItem represents a line item in the invoice
type InvoiceItem struct {
	Description string
	Quantity    int
	UnitPrice   float64
	Total       float64
}
