package xendit

import (
	"fmt"
	"time"
)

const (
	invoiceEndpoint = "/v2/invoices"
)

func (c *Client) CreateInvoice(req CreateInvoiceRequest) (*CreateInvoiceResponse, error) {
	if req.ExternalID == "" {
		req.ExternalID = c.GenerateID()
	}

	if req.ExternalID == "" {
		req.ExternalID = fmt.Sprintf("INV-%d", time.Now().UnixNano())
	}

	if req.CustomerNotificationPreference == "" {
		req.CustomerNotificationPreference = "email_only"
	}

	var resp CreateInvoiceResponse
	err := c.makeRequest("POST", invoiceEndpoint, req, &resp)
	if err != nil {
		return nil, fmt.Errorf("failed to create invoice: %w", err)
	}

	return &resp, nil
}

func (c *Client) GetInvoice(invoiceID string) (*CreateInvoiceResponse, error) {
	if invoiceID == "" {
		return nil, fmt.Errorf("invoice ID is required")
	}

	endpoint := fmt.Sprintf("%s/%s", invoiceEndpoint, invoiceID)
	var resp CreateInvoiceResponse
	err := c.makeRequest("GET", endpoint, nil, &resp)
	if err != nil {
		return nil, fmt.Errorf("failed to get invoice: %w", err)
	}

	return &resp, nil
}

func (c *Client) GetInvoiceByExternalID(externalID string) (*CreateInvoiceResponse, error) {
	if externalID == "" {
		return nil, fmt.Errorf("external ID is required")
	}

	endpoint := fmt.Sprintf("%s?external_id=%s", invoiceEndpoint, externalID)
	var resp CreateInvoiceResponse
	err := c.makeRequest("GET", endpoint, nil, &resp)
	if err != nil {
		return nil, fmt.Errorf("failed to get invoice by external ID: %w", err)
	}

	return &resp, nil
}

func (c *Client) ExpireInvoice(invoiceID string) (*CreateInvoiceResponse, error) {
	if invoiceID == "" {
		return nil, fmt.Errorf("invoice ID is required")
	}

	endpoint := fmt.Sprintf("%s/%s/expire!", invoiceEndpoint, invoiceID)
	var resp CreateInvoiceResponse
	err := c.makeRequest("POST", endpoint, nil, &resp)
	if err != nil {
		return nil, fmt.Errorf("failed to expire invoice: %w", err)
	}

	return &resp, nil
}

func (c *Client) GetAllInvoices(params map[string]string) ([]CreateInvoiceResponse, error) {
	if params == nil {
		params = make(map[string]string)
	}

	var resp []CreateInvoiceResponse
	err := c.makeRequest("GET", invoiceEndpoint, nil, &resp)
	if err != nil {
		return nil, fmt.Errorf("failed to get all invoices: %w", err)
	}

	return resp, nil
}
