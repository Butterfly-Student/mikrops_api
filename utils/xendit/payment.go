package xendit

import (
	"fmt"
	"time"
)

const (
	vaEndpoint = "/callback_virtual_accounts"
)

func (c *Client) CreateVAPayment(req CreateVAPaymentRequest) (*CreateVAPaymentResponse, error) {
	if req.ExternalID == "" {
		req.ExternalID = c.GenerateID()
	}

	if req.ExternalID == "" {
		req.ExternalID = fmt.Sprintf("VA-%d", time.Now().UnixNano())
	}

	if req.ExpirationDate == nil {
		expiry := time.Now().Add(24 * time.Hour)
		req.ExpirationDate = &expiry
	}

	var resp CreateVAPaymentResponse
	err := c.makeRequest("POST", vaEndpoint, req, &resp)
	if err != nil {
		return nil, fmt.Errorf("failed to create VA payment: %w", err)
	}

	return &resp, nil
}

func (c *Client) GetVAPayment(paymentID string) (*CreateVAPaymentResponse, error) {
	if paymentID == "" {
		return nil, fmt.Errorf("payment ID is required")
	}

	endpoint := fmt.Sprintf("%s/%s", vaEndpoint, paymentID)
	var resp CreateVAPaymentResponse
	err := c.makeRequest("GET", endpoint, nil, &resp)
	if err != nil {
		return nil, fmt.Errorf("failed to get VA payment: %w", err)
	}

	return &resp, nil
}

func (c *Client) GetVAPaymentByExternalID(externalID string) (*CreateVAPaymentResponse, error) {
	if externalID == "" {
		return nil, fmt.Errorf("external ID is required")
	}

	endpoint := fmt.Sprintf("%s?external_id=%s", vaEndpoint, externalID)
	var resp CreateVAPaymentResponse
	err := c.makeRequest("GET", endpoint, nil, &resp)
	if err != nil {
		return nil, fmt.Errorf("failed to get VA payment by external ID: %w", err)
	}

	return &resp, nil
}
