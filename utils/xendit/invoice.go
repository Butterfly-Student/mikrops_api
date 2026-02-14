package xendit

import (
	"encoding/json"
	"fmt"
	"time"
)

// InvoiceRequest represents a request to create Xendit invoice
type InvoiceRequest struct {
	ExternalID          string   `json:"external_id"`
	Amount              float64  `json:"amount"`
	PayerEmail          string   `json:"payer_email,omitempty"`
	Description         string   `json:"description"`
	InvoiceDuration     int      `json:"invoice_duration,omitempty"` // in seconds
	SuccessRedirectURL  string   `json:"success_redirect_url,omitempty"`
	FailureRedirectURL  string   `json:"failure_redirect_url,omitempty"`
	Currency            string   `json:"currency,omitempty"` // default: IDR
	PaymentMethods      []string `json:"payment_methods,omitempty"`
	ShouldSendEmail     bool     `json:"should_send_email,omitempty"`
	CustomerName        string   `json:"customer_name,omitempty"`
	CustomerPhone       string   `json:"customer_phone,omitempty"`
}

// InvoiceResponse represents Xendit invoice response
type InvoiceResponse struct {
	ID                 string    `json:"id"`
	ExternalID         string    `json:"external_id"`
	UserID             string    `json:"user_id"`
	Status             string    `json:"status"`
	MerchantName       string    `json:"merchant_name"`
	Amount             float64   `json:"amount"`
	PayerEmail         string    `json:"payer_email"`
	Description        string    `json:"description"`
	ExpiryDate         time.Time `json:"expiry_date"`
	InvoiceURL         string    `json:"invoice_url"`
	AvailableBanks     []Bank    `json:"available_banks"`
	AvailableRetailOutlets []RetailOutlet `json:"available_retail_outlets"`
	AvailableEwallets  []Ewallet `json:"available_ewallets"`
	ShouldExcludeCreditCard bool `json:"should_exclude_credit_card"`
	ShouldSendEmail    bool      `json:"should_send_email"`
	Created            time.Time `json:"created"`
	Updated            time.Time `json:"updated"`
	Currency           string    `json:"currency"`
}

// Bank represents available bank for VA
type Bank struct {
	BankCode          string  `json:"bank_code"`
	CollectionType    string  `json:"collection_type"`
	BankBranch        string  `json:"bank_branch"`
	AccountHolderName string  `json:"account_holder_name"`
	TransferAmount    float64 `json:"transfer_amount"`
}

// RetailOutlet represents available retail outlet
type RetailOutlet struct {
	RetailOutletName string `json:"retail_outlet_name"`
}

// Ewallet represents available e-wallet
type Ewallet struct {
	EwalletType string `json:"ewallet_type"`
}

// CreateInvoice creates a new Xendit invoice
func (c *Client) CreateInvoice(req InvoiceRequest) (*InvoiceResponse, error) {
	// Set defaults
	if req.Currency == "" {
		req.Currency = "IDR"
	}
	if req.InvoiceDuration == 0 {
		req.InvoiceDuration = 86400 // 24 hours
	}

	respBody, err := c.doRequest("POST", "/v2/invoices", req)
	if err != nil {
		return nil, err
	}

	var invoice InvoiceResponse
	if err := json.Unmarshal(respBody, &invoice); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &invoice, nil
}

// GetInvoice retrieves an invoice by ID
func (c *Client) GetInvoice(invoiceID string) (*InvoiceResponse, error) {
	respBody, err := c.doRequest("GET", "/v2/invoices/"+invoiceID, nil)
	if err != nil {
		return nil, err
	}

	var invoice InvoiceResponse
	if err := json.Unmarshal(respBody, &invoice); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &invoice, nil
}

// ExpireInvoice expires an invoice
func (c *Client) ExpireInvoice(invoiceID string) (*InvoiceResponse, error) {
	respBody, err := c.doRequest("POST", "/invoices/"+invoiceID+"/expire!", nil)
	if err != nil {
		return nil, err
	}

	var invoice InvoiceResponse
	if err := json.Unmarshal(respBody, &invoice); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &invoice, nil
}
