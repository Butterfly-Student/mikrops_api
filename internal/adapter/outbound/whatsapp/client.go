package whatsapp

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/palantir/stacktrace"

	"mikrops/internal/model"
	outbound_port "mikrops/internal/port/outbound"
)

type whatsappClient struct {
	apiURL    string
	apiKey    string
	sender    string
	httpClient *http.Client
}

// NewWhatsappClient creates a new WhatsApp gateway client
func NewWhatsappClient() outbound_port.WhatsappPort {
	return &whatsappClient{
		apiURL:    os.Getenv("WHATSAPP_GATEWAY_URL"),
		apiKey:    os.Getenv("WHATSAPP_API_KEY"),
		sender:    os.Getenv("WHATSAPP_SENDER"),
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

// SendCredentials sends PPPoE credentials to customer via WhatsApp
func (w *whatsappClient) SendCredentials(phone string, username string, password string) error {
	if w.apiURL == "" || w.apiKey == "" {
		// WhatsApp integration not configured, skip silently
		return nil
	}

	message := fmt.Sprintf(
		"✅ *Akun Internet Anda Telah Aktif*\n\n"+
			"Berikut adalah kredensial PPPoE Anda:\n"+
			"👤 Username: *%s*\n"+
			"🔑 Password: *%s*\n\n"+
			"Silakan hubungi customer service jika ada pertanyaan.\n"+
			"Terima kasih telah berlangganan! 🙏",
		username, password,
	)

	return w.sendMessage(phone, message)
}

// SendInvoiceReminder sends invoice reminder notification via WhatsApp
func (w *whatsappClient) SendInvoiceReminder(phone string, invoice model.Invoice) error {
	if w.apiURL == "" || w.apiKey == "" {
		return nil
	}

	message := fmt.Sprintf(
		"🔔 *Pengingat Tagihan*\n\n"+
			"Nomor Invoice: *%s*\n"+
			"Jumlah: *Rp %s*\n"+
			"Jatuh Tempo: *%s*\n\n"+
			"Mohon segera lakukan pembayaran untuk menghindari pemutusan layanan.\n"+
			"Terima kasih! 🙏",
		invoice.InvoiceNumber,
		formatCurrency(invoice.TotalAmount),
		invoice.DueDate.Format("02 Jan 2006"),
	)

	return w.sendMessage(phone, message)
}

// SendIsolationNotice sends isolation notice to customer via WhatsApp
func (w *whatsappClient) SendIsolationNotice(phone string, customer model.Customer) error {
	if w.apiURL == "" || w.apiKey == "" {
		return nil
	}

	message := fmt.Sprintf(
		"⚠️ *Pemberitahuan Isolasi Layanan*\n\n"+
			"Yth. %s,\n\n"+
			"Layanan internet Anda telah diisolasi karena tagihan yang belum dibayar.\n\n"+
			"Silakan hubungi customer service atau lakukan pembayaran untuk mengaktifkan kembali layanan Anda.\n\n"+
			"Terima kasih. 🙏",
		customer.FullName,
	)

	return w.sendMessage(phone, message)
}

// SendRestorationNotice sends restoration notice to customer via WhatsApp
func (w *whatsappClient) SendRestorationNotice(phone string, customer model.Customer) error {
	if w.apiURL == "" || w.apiKey == "" {
		return nil
	}

	message := fmt.Sprintf(
		"✅ *Layanan Anda Telah Dipulihkan*\n\n"+
			"Yth. %s,\n\n"+
			"Terima kasih atas pembayaran Anda. Layanan internet Anda telah diaktifkan kembali.\n\n"+
			"Selamat menikmati layanan kami! 🎉",
		customer.FullName,
	)

	return w.sendMessage(phone, message)
}

// sendMessage sends a message via WhatsApp Gateway API
func (w *whatsappClient) sendMessage(phone string, message string) error {
	// Generic WhatsApp gateway payload structure
	// Adjust based on your specific gateway API
	payload := map[string]interface{}{
		"api_key": w.apiKey,
		"sender":  w.sender,
		"number":  phone,
		"message": message,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return stacktrace.Propagate(err, "failed to marshal WhatsApp payload")
	}

	req, err := http.NewRequest("POST", w.apiURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return stacktrace.Propagate(err, "failed to create WhatsApp request")
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := w.httpClient.Do(req)
	if err != nil {
		return stacktrace.Propagate(err, "failed to send WhatsApp message")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return stacktrace.NewError("WhatsApp API returned non-success status: %d", resp.StatusCode)
	}

	return nil
}

// formatCurrency formats amount to Indonesian Rupiah format
func formatCurrency(amount int64) string {
	// Simple formatting, can be improved with proper locale formatting
	return fmt.Sprintf("%d", amount)
}
