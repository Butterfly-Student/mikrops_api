package production

import (
	"go-template/internal/model"
	"go-template/internal/seeds/runner"
	"gorm.io/gorm"
)

type NotificationTemplateSeeder struct{}

func (s *NotificationTemplateSeeder) Name() string {
	return "Notification Templates"
}

func (s *NotificationTemplateSeeder) Seed(db *gorm.DB) error {
	templates := []model.NotificationTemplate{
		{
			Name:    "payment_confirmation_email",
			Type:    "email",
			Subject: "Payment Confirmation - [PaymentNumber]",
			Content: `Dear [CustomerName],

Thank you for your payment!

Payment Details:
- Payment Number: [PaymentNumber]
- Amount: Rp [PaymentAmount]
- Payment Date: [PaymentDate]
- Payment Method: [PaymentMethod]

Your payment has been successfully received and applied to your invoice(s).

If you have any questions, please contact our support team.

Best regards,
Your Billing Team`,
			IsActive: true,
		},
		{
			Name:    "payment_confirmation_whatsapp",
			Type:    "whatsapp",
			Subject: "",
			Content: `Dear [CustomerName],

Thank you for your payment of Rp [PaymentAmount] (Payment No: [PaymentNumber]). Your payment has been successfully received.

Best regards`,
			IsActive: true,
		},
		{
			Name:    "invoice_reminder_email",
			Type:    "email",
			Subject: "Invoice Reminder - [InvoiceNumber]",
			Content: `Dear [CustomerName],

This is a friendly reminder that your invoice is [DaysOverdue] days overdue.

Invoice Details:
- Invoice Number: [InvoiceNumber]
- Invoice Date: [InvoiceDate]
- Due Date: [DueDate]
- Total Amount: Rp [TotalAmount]
- Outstanding Balance: Rp [OutstandingBalance]

Please complete your payment as soon as possible to avoid additional late fees.

You can pay your invoice through our payment portal: [PortalURL]

If you have already made a payment, please disregard this notice.

Best regards,
Your Billing Team`,
			IsActive: true,
		},
		{
			Name:    "payment_failed_email",
			Type:    "email",
			Subject: "Payment Failed - [InvoiceNumber]",
			Content: `Dear [CustomerName],

We regret to inform you that your payment for invoice [InvoiceNumber] has failed.

Payment Details:
- Invoice Number: [InvoiceNumber]
- Failed Reason: [FailedReason]
- Total Amount: Rp [TotalAmount]

Please try again or use a different payment method. If the issue persists, please contact our support team.

You can retry payment through our payment portal: [PortalURL]

Best regards,
Your Billing Team`,
			IsActive: true,
		},
		{
			Name:    "invoice_created_email",
			Type:    "email",
			Subject: "New Invoice - [InvoiceNumber]",
			Content: `Dear [CustomerName],

We are pleased to inform you that a new invoice has been generated for your account.

Invoice Details:
- Invoice Number: [InvoiceNumber]
- Invoice Date: [InvoiceDate]
- Due Date: [DueDate]
- Total Amount: Rp [TotalAmount]

Please complete your payment before the due date to avoid late fees.

You can view and pay your invoice through our payment portal: [PortalURL]

If you have any questions, please contact our support team.

Best regards,
Your Billing Team`,
			IsActive: true,
		},
		{
			Name:    "customer_activated_whatsapp",
			Type:    "whatsapp",
			Subject: "",
			Content: `🟢 *AKUN DIAKTIFKAN*

Nama: [CustomerName]
Kode Pelanggan: [CustomerCode]
Status: Aktif

Layanan internet telah diaktifkan kembali.

Portal: [PortalURL]

Terima kasih!`,
			IsActive: true,
		},
		{
			Name:    "customer_isolated_whatsapp",
			Type:    "whatsapp",
			Subject: "",
			Content: `🔴 *AKUN TERISOLASI*

Nama: [CustomerName]
Kode Pelanggan: [CustomerCode]
Status: Terisolasi (terbatas)

Alasan: [Reason]

Silakan lakukan pembayaran melalui portal:
[PortalURL]

Portal: [PortalURL]

Terima kasih!`,
			IsActive: true,
		},
		{
			Name:    "invoice_created_whatsapp",
			Type:    "whatsapp",
			Subject: "",
			Content: `📄 *INVOICE BARU*

Halo [CustomerName],

Invoice baru telah dibuat untuk akun Anda:

- No. Invoice: [InvoiceNumber]
- Tanggal: [InvoiceDate]
- Jatuh Tempo: [DueDate]
- Total: Rp [TotalAmount]

Silakan lakukan pembayaran sebelum jatuh tempo.

Portal: [PortalURL]

Terima kasih!`,
			IsActive: true,
		},
		{
			Name:    "payment_failed_whatsapp",
			Type:    "whatsapp",
			Subject: "",
			Content: `❌ *PEMBAYARAN GAGAL*

Halo [CustomerName],

Maaf, pembayaran untuk invoice [InvoiceNumber] gagal.

Alasan: [FailedReason]
Total: Rp [TotalAmount]

Silakan coba lagi atau gunakan metode pembayaran lain.

Portal: [PortalURL]

Terima kasih!`,
			IsActive: true,
		},
		{
			Name:    "pppoe_connection_up_whatsapp",
			Type:    "whatsapp",
			Subject: "",
			Content: `🟢 *STATUS KONEKSI PPPoE*

Nama: [CustomerName]
Kode Pelanggan: [CustomerCode]
Status: TERHUBUNG

Router: [RouterName]
IP Address: [IPAddress]
Connected At: [ConnectedAt]`,
			IsActive: true,
		},
		{
			Name:    "pppoe_connection_down_whatsapp",
			Type:    "whatsapp",
			Subject: "",
			Content: `🔴 *STATUS KONEKSI PPPoE*

Nama: [CustomerName]
Kode Pelanggan: [CustomerCode]
Status: TERPUTUS

Router: [RouterName]
Disconnected At: [DisconnectedAt]
Duration: [Duration]

Silakan cek koneksi pelanggan.`,
			IsActive: true,
		},
	}

	for _, tmpl := range templates {
		var existing model.NotificationTemplate
		result := db.Where("name = ?", tmpl.Name).First(&existing)

		if result.Error != nil && result.Error == gorm.ErrRecordNotFound {
			if err := db.Create(&tmpl).Error; err != nil {
				return err
			}
		}
	}

	return nil
}

func init() {
	runner.RegisterSeeder(&NotificationTemplateSeeder{})
}
