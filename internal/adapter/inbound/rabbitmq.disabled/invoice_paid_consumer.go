package rabbitmq_inbound_adapter

import (
	"context"
	"encoding/json"

	"mikrops/internal/domain"
	"mikrops/utils/log"
)

// InvoicePaidConsumer handles invoice.paid events to trigger restore flow
type InvoicePaidConsumer struct {
	domain domain.Domain
}

func NewInvoicePaidConsumer(d domain.Domain) *InvoicePaidConsumer {
	return &InvoicePaidConsumer{domain: d}
}

// Handle processes invoice.paid messages
func (c *InvoicePaidConsumer) Handle(msg []byte) bool {
	ctx := context.Background()
	
	var message struct {
		InvoiceID string `json:"invoice_id"`
	}
	
	err := json.Unmarshal(msg, &message)
	if err != nil {
		log.WithContext(ctx).Errorf("failed to unmarshal invoice.paid message: %v", err)
		return false
	}
	
	// Trigger restore flow via cutoff domain
	err = c.domain.Cutoff().OnInvoicePaid(ctx, message.InvoiceID)
	if err != nil {
		log.WithContext(ctx).Errorf("failed to process invoice paid event for invoice %s: %v", message.InvoiceID, err)
		return false
	}
	
	log.WithContext(ctx).Infof("Invoice paid event processed successfully: %s", message.InvoiceID)
	return true
}
