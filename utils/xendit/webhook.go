package xendit

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
)

func (c *Client) VerifyWebhookSignature(payload []byte, signatureHeader string) (bool, error) {
	if c.webhookToken == "" {
		return false, fmt.Errorf("webhook token is not configured")
	}

	if signatureHeader == "" {
		return false, fmt.Errorf("signature header is empty")
	}

	expectedSignature := c.calculateHMAC(payload)

	if !hmac.Equal([]byte(signatureHeader), []byte(expectedSignature)) {
		return false, fmt.Errorf("invalid webhook signature")
	}

	return true, nil
}

func (c *Client) calculateHMAC(payload []byte) string {
	h := hmac.New(sha256.New, []byte(c.webhookToken))
	h.Write(payload)
	return hex.EncodeToString(h.Sum(nil))
}

func (c *Client) ParseWebhookData(payload []byte) (*WebhookData, error) {
	if len(payload) == 0 {
		return nil, fmt.Errorf("payload is empty")
	}

	var webhookData WebhookData
	if err := json.Unmarshal(payload, &webhookData); err != nil {
		return nil, fmt.Errorf("failed to unmarshal webhook data: %w", err)
	}

	return &webhookData, nil
}

func (c *Client) ValidateWebhook(payload []byte, signatureHeader string) (*WebhookData, error) {
	isValid, err := c.VerifyWebhookSignature(payload, signatureHeader)
	if err != nil {
		return nil, fmt.Errorf("failed to verify webhook signature: %w", err)
	}

	if !isValid {
		return nil, fmt.Errorf("invalid webhook signature")
	}

	webhookData, err := c.ParseWebhookData(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to parse webhook data: %w", err)
	}

	return webhookData, nil
}

func IsPaymentPaid(status string) bool {
	return status == "PAID" || status == "SUCCEEDED"
}

func IsPaymentPending(status string) bool {
	return status == "PENDING" || status == "AWAITING_PAYMENT"
}

func IsPaymentFailed(status string) bool {
	return status == "FAILED" || status == "CANCELLED" || status == "EXPIRED"
}
