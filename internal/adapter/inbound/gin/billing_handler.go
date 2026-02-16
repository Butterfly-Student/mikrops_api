package gin_inbound_adapter

import (
	"bytes"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jung-kurt/gofpdf"
	"go-template/internal/domain"
	"go-template/internal/model"
	inbound_port "go-template/internal/port/inbound"
)

type BillingHandler struct {
	domain domain.Domain
}

func NewBillingHandler(domain domain.Domain) inbound_port.BillingHttpPort {
	return &BillingHandler{domain: domain}
}

func (h *BillingHandler) CreateInvoice(c *gin.Context) {
	var input model.InvoiceInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	invoice, err := h.domain.Billing().CreateInvoice(c.Request.Context(), input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, invoice)
}

func (h *BillingHandler) GetInvoice(c *gin.Context) {
	id := c.Param("id")

	invoice, err := h.domain.Billing().GetInvoice(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, invoice)
}

func (h *BillingHandler) ListInvoices(c *gin.Context) {
	var filter model.InvoiceFilter
	if err := c.ShouldBindQuery(&filter); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	invoices, err := h.domain.Billing().ListInvoices(c.Request.Context(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, invoices)
}

func (h *BillingHandler) UpdateInvoice(c *gin.Context) {
	id := c.Param("id")
	var input model.InvoiceInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	invoice, err := h.domain.Billing().UpdateInvoice(c.Request.Context(), id, input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, invoice)
}

func (h *BillingHandler) DeleteInvoice(c *gin.Context) {
	id := c.Param("id")

	err := h.domain.Billing().DeleteInvoice(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "invoice deleted successfully"})
}

func (h *BillingHandler) GenerateMonthlyInvoices(c *gin.Context) {
	type GenerateRequest struct {
		Year  int `json:"year" binding:"required"`
		Month int `json:"month" binding:"required,min=1,max=12"`
	}

	var req GenerateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	results, err := h.domain.Billing().GenerateMonthlyInvoices(c.Request.Context(), req.Year, req.Month)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"count":   len(results),
		"results": results,
	})
}

func (h *BillingHandler) ApplyPayment(c *gin.Context) {
	id := c.Param("id")

	type PaymentRequest struct {
		Amount float64 `json:"amount" binding:"required,min=0"`
	}

	var req PaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.domain.Billing().ApplyPayment(c.Request.Context(), id, req.Amount)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "payment applied successfully"})
}

func (h *BillingHandler) CheckOverdueInvoices(c *gin.Context) {
	invoices, err := h.domain.Billing().CheckOverdueInvoices(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"count":    len(invoices),
		"invoices": invoices,
	})
}

func (h *BillingHandler) CalculateLateFee(c *gin.Context) {
	id := c.Param("id")

	invoice, err := h.domain.Billing().GetInvoice(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	lateFee := h.domain.Billing().CalculateLateFee(c.Request.Context(), invoice)

	c.JSON(http.StatusOK, gin.H{
		"invoice_id": invoice.ID,
		"late_fee":   lateFee,
	})
}

func (h *BillingHandler) CreateInvoiceItem(c *gin.Context) {
	var input model.InvoiceItemInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	item, err := h.domain.Billing().CreateInvoiceItem(c.Request.Context(), input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, item)
}

func (h *BillingHandler) GetInvoiceItem(c *gin.Context) {
	id := c.Param("id")

	item, err := h.domain.Billing().GetInvoiceItem(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, item)
}

func (h *BillingHandler) ListInvoiceItems(c *gin.Context) {
	invoiceID := c.Query("invoice_id")
	if invoiceID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invoice_id is required"})
		return
	}

	items, err := h.domain.Billing().ListInvoiceItems(c.Request.Context(), invoiceID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, items)
}

// DownloadInvoicePDF generates and downloads a PDF for an invoice
func (h *BillingHandler) DownloadInvoicePDF(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invoice ID is required"})
		return
	}

	// Get invoice details
	invoice, err := h.domain.Billing().GetInvoice(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	// Get invoice items
	items, err := h.domain.Billing().ListInvoiceItems(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Get customer details
	customer, err := h.domain.Customer().GetCustomer(c.Request.Context(), invoice.CustomerID.String())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Generate PDF
	pdfBytes, err := h.generateInvoicePDF(*invoice, items, *customer)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	filename := "invoice_" + invoice.InvoiceNumber + ".pdf"

	c.Header("Content-Type", "application/pdf")
	c.Header("Content-Disposition", "attachment; filename="+filename)
	c.Header("Content-Length", strconv.Itoa(len(pdfBytes)))

	c.Data(http.StatusOK, "application/pdf", pdfBytes)
}

// generateInvoicePDF generates a PDF document for an invoice
func (h *BillingHandler) generateInvoicePDF(invoice model.Invoice, items []model.InvoiceItem, customer model.Customer) ([]byte, error) {
	// Import gofpdf package
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()

	// Set fonts - use built-in fonts for simplicity
	pdf.SetFont("Arial", "B", 24)

	// Company Header
	pdf.SetXY(10, 20)
	pdf.Cell(190, 10, "INVOICE")

	// Invoice details
	pdf.SetFont("Arial", "", 12)
	pdf.SetXY(10, 40)
	pdf.Cell(95, 8, "Invoice Number: "+invoice.InvoiceNumber)

	pdf.SetXY(105, 40)
	pdf.Cell(95, 8, "Date: "+invoice.IssueDate.Format("2006-01-02"))

	pdf.SetXY(10, 48)
	pdf.Cell(95, 8, "Due Date: "+invoice.DueDate.Format("2006-01-02"))

	pdf.SetXY(105, 48)
	status := string(invoice.Status)
	pdf.Cell(95, 8, "Status: "+status)

	// Customer Information
	pdf.SetFont("Arial", "B", 14)
	pdf.SetXY(10, 65)
	pdf.Cell(190, 8, "Bill To:")

	pdf.SetFont("Arial", "", 12)
	pdf.SetXY(10, 73)
	pdf.Cell(95, 8, customer.FullName)

	if customer.Phone != "" {
		pdf.SetXY(10, 81)
		pdf.Cell(95, 8, "Phone: "+customer.Phone)
	}

	if customer.Email != nil && *customer.Email != "" {
		pdf.SetXY(10, 89)
		pdf.Cell(95, 8, "Email: "+*customer.Email)
	}

	if customer.Address != nil && *customer.Address != "" {
		pdf.SetXY(10, 97)
		pdf.MultiCell(190, 8, "Address: "+*customer.Address, "", "", false)
	}

	// Table Header
	yPos := 120.0
	pdf.SetFont("Arial", "B", 12)
	pdf.SetXY(10, yPos)
	pdf.Cell(100, 10, "Description")
	pdf.Cell(30, 10, "Qty")
	pdf.Cell(30, 10, "Unit Price")
	pdf.Cell(30, 10, "Amount")
	pdf.Ln(-1)
	yPos += 10

	// Line
	pdf.Line(10, yPos, 200, yPos)
	yPos += 5

	// Items
	pdf.SetFont("Arial", "", 11)
	for _, item := range items {
		pdf.SetXY(10, yPos)
		pdf.Cell(100, 8, item.Description)
		pdf.Cell(30, 8, strconv.Itoa(int(item.Quantity)))
		pdf.Cell(30, 8, formatCurrency(item.UnitPrice))
		pdf.Cell(30, 8, formatCurrency(item.Total))
		pdf.Ln(-1)
		yPos += 8
	}

	// Totals
	yPos += 10
	pdf.Line(10, yPos, 200, yPos)
	yPos += 10

	pdf.SetFont("Arial", "", 12)
	pdf.SetXY(120, yPos)
	pdf.Cell(30, 8, "Subtotal:")
	pdf.Cell(40, 8, formatCurrency(invoice.Subtotal))

	yPos += 8
	pdf.SetXY(120, yPos)
	pdf.Cell(30, 8, "Tax (11%):")
	pdf.Cell(40, 8, formatCurrency(invoice.TaxAmount))

	if invoice.LateFee > 0 {
		yPos += 8
		pdf.SetXY(120, yPos)
		pdf.Cell(30, 8, "Late Fee:")
		pdf.Cell(40, 8, formatCurrency(invoice.LateFee))
	}

	yPos += 10
	pdf.SetFont("Arial", "B", 14)
	pdf.SetXY(120, yPos)
	pdf.Cell(30, 10, "TOTAL:")
	pdf.Cell(40, 10, formatCurrency(invoice.TotalAmount))

	// Generate PDF bytes
	var buf bytes.Buffer
	err := pdf.Output(&buf)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// formatCurrency formats a float64 to Indonesian Rupiah string
func formatCurrency(amount float64) string {
	return "Rp " + strconv.FormatFloat(amount, 'f', 0, 64)
}
