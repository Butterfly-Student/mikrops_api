package main

import (
	"context"
	"fmt"

	"go-template/internal/model"
	"go-template/internal/port/outbound"
	"go-template/utils/log"
)

// SeedNotificationTemplates seeds default notification templates
func SeedNotificationTemplates(dbPort outbound_port.DatabasePort) error {
	ctx := context.Background()

	templates := []model.NotificationTemplate{
		{
			Name:      "invoice_created",
			Type:      "whatsapp",
			Subject:   "Invoice #[InvoiceNumber] - [CustomerName]",
			Content:   "Halo [CustomerName], invoice bulanan Anda telah dibuat.\n\nNo. Invoice: #[InvoiceNumber]\nPeriode: [BillingPeriod]\nTotal: Rp[TotalAmount]\nJatuh tempo: [DueDate]\n\nSilakan bayar melalui portal: [PaymentPortalURL]\n\nTerima kasih!",
			Variables: `["CustomerName", "InvoiceNumber", "BillingPeriod", "TotalAmount", "DueDate", "PaymentPortalURL"]`,
			IsActive:  true,
		},
		{
			Name:      "payment_received",
			Type:      "whatsapp",
			Subject:   "Pembayaran Diterima - [CustomerName]",
			Content:   "Halo [CustomerName],\n\nPembayaran Anda telah kami terima.\n\nNo. Invoice: #[InvoiceNumber]\nJumlah: Rp[Amount]\nMetode: [PaymentMethod]\nTanggal: [PaymentDate]\n\nTerima kasih atas pembayaran Anda! Layanan Anda tetap aktif.\n\nJika ada pertanyaan, hubungi: [SupportPhone]",
			Variables: `["CustomerName", "InvoiceNumber", "Amount", "PaymentMethod", "PaymentDate", "SupportPhone"]`,
			IsActive:  true,
		},
		{
			Name:      "overdue_warning",
			Type:      "whatsapp",
			Subject:   "Pengingat Pembayaran - Invoice #[InvoiceNumber]",
			Content:   "Halo [CustomerName],\n\nINVOICE ANDA SUDAH JATUH TEMPO\n\nNo. Invoice: #[InvoiceNumber]\nJatuh tempo: [DueDate]\nJumlah: Rp[Amount]\nHari terlambat: [DaysOverdue] hari\n\nMohon segera lakukan pembayaran untuk menghindari isolasi layanan.\n\nBayar melalui portal: [PaymentPortalURL]\n\nJika sudah dibayar, abaikan pesan ini.",
			Variables: `["CustomerName", "InvoiceNumber", "DueDate", "Amount", "DaysOverdue", "PaymentPortalURL"]`,
			IsActive:  true,
		},
		{
			Name:      "customer_isolated",
			Type:      "whatsapp",
			Subject:   "Layanan Terisolasi - [CustomerName]",
			Content:   "Halo [CustomerName],\n\nLAYANAN INTERNET ANDA TELAH DIISOLASI\n\nKarena pembayaran terlambat melewati masa tenggat, akses internet Anda telah dibatasi.\n\nInvoice terbayar: #[InvoiceNumber]\nJatuh tempo: [DueDate]\nJumlah: Rp[Amount]\n\nUntuk mengembalikan layanan normal, segera lakukan pembayaran:\n[PaymentPortalURL]\n\nSetelah pembayaran dikonfirmasi, layanan akan aktif kembali dalam 5-10 menit.\n\nHubungi support: [SupportPhone]",
			Variables: `["CustomerName", "InvoiceNumber", "DueDate", "Amount", "PaymentPortalURL", "SupportPhone"]`,
			IsActive:  true,
		},
		{
			Name:      "customer_reactivated",
			Type:      "whatsapp",
			Subject:   "Layanan Aktif Kembali - [CustomerName]",
			Content:   "Halo [CustomerName],\n\nLAYANAN ANDA TELAH AKTIF KEMBALI\n\nTerima kasih telah melakukan pembayaran. Layanan internet Anda telah dikembalikan ke profil semula.\n\nPembayaran diterima: [PaymentDate]\nJumlah: Rp[Amount]\nInvoice: #[InvoiceNumber]\n\nPeriode aktif hingga: [NewExpiryDate]\n\nJika mengalami kendala, coba restart router Anda. Jika masih bermasalah, hubungi: [SupportPhone]\n\nTerima kasih telah menjadi pelanggan kami!",
			Variables: `["CustomerName", "PaymentDate", "Amount", "InvoiceNumber", "NewExpiryDate", "SupportPhone"]`,
			IsActive:  true,
		},
		{
			Name:      "invoice_reminder_before_due",
			Type:      "whatsapp",
			Subject:   "Pengingat Invoice - [CustomerName]",
			Content:   "Halo [CustomerName],\n\nPENGINGAT INVOICE\n\nInvoice #[InvoiceNumber] akan jatuh tempo dalam [DaysUntilDue] hari.\n\nJumlah: Rp[Amount]\nJatuh tempo: [DueDate]\n\nSilakan lakukan pembayaran sebelum jatuh tempo:\n[PaymentPortalURL]\n\nTerima kasih!",
			Variables: `["CustomerName", "InvoiceNumber", "DaysUntilDue", "Amount", "DueDate", "PaymentPortalURL"]`,
			IsActive:  true,
		},
		{
			Name:      "payment_failed",
			Type:      "whatsapp",
			Subject:   "Pembayaran Gagal - [CustomerName]",
			Content:   "Halo [CustomerName],\n\nPEMBAYARAN ANDA GAGAL\n\nMohon maaf, pembayaran yang Anda lakukan tidak dapat kami proses.\n\nInvoice: #[InvoiceNumber]\nJumlah: Rp[Amount]\nMetode: [PaymentMethod]\nWaktu: [PaymentTime]\n\nSilakan coba lagi atau gunakan metode pembayaran lain.\n\nHubungi support jika masalah berlanjut: [SupportPhone]",
			Variables: `["CustomerName", "InvoiceNumber", "Amount", "PaymentMethod", "PaymentTime", "SupportPhone"]`,
			IsActive:  true,
		},
		{
			Name:      "new_customer_welcome",
			Type:      "whatsapp",
			Subject:   "Selamat Datang - [CustomerName]",
			Content:   "Halo [CustomerName],\n\nSELAMAT DATANG DI [CompanyName]!\n\nTerima kasih telah berlangganan layanan internet kami.\n\nPaket: [PackageName]\nKecepatan: [Speed]\nHarga: Rp[Price]/bulan\n\nLogin PPPoE:\nUsername: [PPPoEUsername]\nPassword: [PPPoEPassword]\n\nTagihan pertama Anda akan dibuat pada [FirstBillingDate].\n\nPortal pelanggan: [PaymentPortalURL]\nSupport: [SupportPhone]\n\nKami siap melayani Anda!",
			Variables: `["CustomerName", "CompanyName", "PackageName", "Speed", "Price", "PPPoEUsername", "PPPoEPassword", "FirstBillingDate", "PaymentPortalURL", "SupportPhone"]`,
			IsActive:  true,
		},
	}

	log.WithContext(ctx).Info(fmt.Sprintf("Seeding %d notification templates...", len(templates)))

	for _, template := range templates {
		// Check if template already exists
		existing, err := dbPort.NotificationTemplate().FindByName(template.Name)
		if err == nil && existing != nil {
			log.WithContext(ctx).Info(fmt.Sprintf("Template '%s' already exists, skipping...", template.Name))
			continue
		}

		// Create template
		err = dbPort.NotificationTemplate().Create(&template)
		if err != nil {
			log.WithContext(ctx).Error(fmt.Sprintf("Failed to create template '%s': %v", template.Name, err))
			return err
		}

		log.WithContext(ctx).Info(fmt.Sprintf("✓ Created notification template: %s", template.Name))
	}

	log.WithContext(ctx).Info("Notification templates seeding completed!")
	return nil
}
