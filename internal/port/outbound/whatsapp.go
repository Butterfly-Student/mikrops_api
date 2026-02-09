package outbound_port

import "mikrops/internal/model"

// WhatsappPort defines the interface for WhatsApp Gateway integration
type WhatsappPort interface {
	// SendCredentials sends PPPoE credentials to customer via WhatsApp
	SendCredentials(phone string, username string, password string) error

	// SendInvoiceReminder sends invoice reminder notification via WhatsApp
	SendInvoiceReminder(phone string, invoice model.Invoice) error

	// SendIsolationNotice sends isolation notice to customer via WhatsApp
	SendIsolationNotice(phone string, customer model.Customer) error

	// SendRestorationNotice sends restoration notice to customer via WhatsApp
	SendRestorationNotice(phone string, customer model.Customer) error
}
