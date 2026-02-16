package notification

import (
	"context"
	"fmt"
	"time"

	"go-template/internal/model"
	outbound_port "go-template/internal/port/outbound"
	"go-template/utils/email"
	"go-template/utils/gowa"
	"go-template/utils/log"
	"go-template/utils/settings"
	"go-template/utils/template"
)

type NotificationDomain interface {
	CreateNotification(ctx context.Context, input model.NotificationInput) (*model.Notification, error)
	GetNotification(ctx context.Context, id string) (*model.Notification, error)
	ListNotifications(ctx context.Context, filter model.NotificationFilter) ([]model.Notification, error)
	SendNotification(ctx context.Context, id string) error
	RetryFailedNotifications(ctx context.Context) ([]model.Notification, error)

	// Notification Templates
	CreateTemplate(ctx context.Context, input model.NotificationTemplateInput) (*model.NotificationTemplate, error)
	GetTemplate(ctx context.Context, id string) (*model.NotificationTemplate, error)
	ListTemplates(ctx context.Context) ([]model.NotificationTemplate, error)
	UpdateTemplate(ctx context.Context, id string, input model.NotificationTemplateInput) (*model.NotificationTemplate, error)
	DeleteTemplate(ctx context.Context, id string) error

	// Predefined Notification Types
	SendPaymentConfirmation(ctx context.Context, customerID string, paymentID string, amount float64) error
	SendInvoiceReminder(ctx context.Context, customerID string, invoiceID string, days int) error
	SendPaymentFailed(ctx context.Context, customerID string, invoiceID string, reason string) error
	SendInvoiceCreated(ctx context.Context, customerID string, invoiceID string) error

	// Customer Status Notifications
	SendActivationNotification(ctx context.Context, customerID string, message string) error
	SendIsolationNotification(ctx context.Context, customerID string, reason string) error
	SendReactivationNotification(ctx context.Context, customerID string, message string) error

	// Group Notifications
	SendToGroup(ctx context.Context, groupID string, message string) error
	SendPPPoEConnectionNotification(ctx context.Context, customerID string, status string, details map[string]string) error
}

type domain struct {
	dbPort    outbound_port.DatabasePort
	emailUtil *email.EmailUtil
	gowaUtil  *gowa.Client
}

func NewNotificationDomain(
	dbPort outbound_port.DatabasePort,
	emailUtil *email.EmailUtil,
	gowaUtil *gowa.Client,
) NotificationDomain {
	return &domain{
		dbPort:    dbPort,
		emailUtil: emailUtil,
		gowaUtil:  gowaUtil,
	}
}

func (d *domain) CreateNotification(ctx context.Context, input model.NotificationInput) (*model.Notification, error) {
	notification := &model.Notification{
		Type:        input.Type,
		Recipient:   input.Recipient,
		Subject:     input.Subject,
		Content:     input.Content,
		CustomerID:  input.CustomerID,
		InvoiceID:   input.InvoiceID,
		PaymentID:   input.PaymentID,
		Status:      "pending",
		RetryCount:  0,
		ScheduledAt: func() *time.Time { t := time.Now(); return &t }(),
	}

	err := d.dbPort.Notification().Create(notification)
	if err != nil {
		return nil, fmt.Errorf("failed to create notification: %w", err)
	}

	log.WithContext(ctx).Info(fmt.Sprintf("Notification created: ID=%s, Type=%s, Recipient=%s", notification.ID, notification.Type, notification.Recipient))

	return notification, nil
}

func (d *domain) GetNotification(ctx context.Context, id string) (*model.Notification, error) {
	return d.dbPort.Notification().FindByID(id)
}

func (d *domain) ListNotifications(ctx context.Context, filter model.NotificationFilter) ([]model.Notification, error) {
	if filter.IsEmpty() {
		return d.dbPort.Notification().FindAll()
	}
	return d.dbPort.Notification().Find(filter)
}

func (d *domain) SendNotification(ctx context.Context, id string) error {
	notification, err := d.dbPort.Notification().FindByID(id)
	if err != nil {
		return fmt.Errorf("notification not found: %w", err)
	}

	var sendErr error

	switch notification.Type {
	case "email":
		subject := "Notification"
		if notification.Subject != nil {
			subject = *notification.Subject
		}
		sendErr = d.emailUtil.SendEmail(notification.Recipient, subject, notification.Content)
	case "whatsapp":
		message := notification.Content
		_, sendErr = d.gowaUtil.SendTextMessage(notification.Recipient, message)
	case "sms":
		// SMS implementation can be added later
		sendErr = fmt.Errorf("SMS not implemented yet")
	default:
		sendErr = fmt.Errorf("unsupported notification type: %s", notification.Type)
	}

	now := time.Now()
	notification.SentAt = &now
	notification.RetryCount++

	if sendErr != nil {
		notification.Status = "failed"
		errMsg := sendErr.Error()
		notification.ErrorMsg = &errMsg
		if notification.RetryCount < 3 {
			notification.Status = "retrying"
		}
		log.WithContext(ctx).Error(fmt.Sprintf("Failed to send notification: ID=%s, Error: %v", notification.ID, sendErr))
	} else {
		notification.Status = "sent"
		log.WithContext(ctx).Info(fmt.Sprintf("Notification sent successfully: ID=%s, Type=%s, Recipient=%s", notification.ID, notification.Type, notification.Recipient))
	}

	err = d.dbPort.Notification().Update(notification)
	if err != nil {
		log.WithContext(ctx).Error(fmt.Sprintf("Failed to update notification status: ID=%s, Error: %v", notification.ID, err))
	}

	return sendErr
}

func (d *domain) RetryFailedNotifications(ctx context.Context) ([]model.Notification, error) {
	failedNotifications, err := d.dbPort.Notification().FindByStatus([]string{"failed", "retrying"})
	if err != nil {
		return nil, err
	}

	var retried []model.Notification

	for _, notification := range failedNotifications {
		if notification.RetryCount >= 3 {
			continue
		}

		err := d.SendNotification(ctx, notification.ID.String())
		if err == nil {
			retried = append(retried, notification)
		}
	}

	return retried, nil
}

func (d *domain) CreateTemplate(ctx context.Context, input model.NotificationTemplateInput) (*model.NotificationTemplate, error) {
	template := &model.NotificationTemplate{
		Name:      input.Name,
		Type:      input.Type,
		Subject:   input.Subject,
		Content:   input.Content,
		Variables: input.Variables,
	}

	if input.IsActive != nil {
		template.IsActive = *input.IsActive
	} else {
		template.IsActive = true
	}

	err := d.dbPort.NotificationTemplate().Create(template)
	if err != nil {
		return nil, fmt.Errorf("failed to create template: %w", err)
	}

	return template, nil
}

func (d *domain) GetTemplate(ctx context.Context, id string) (*model.NotificationTemplate, error) {
	return d.dbPort.NotificationTemplate().FindByID(id)
}

func (d *domain) ListTemplates(ctx context.Context) ([]model.NotificationTemplate, error) {
	return d.dbPort.NotificationTemplate().FindAll()
}

func (d *domain) UpdateTemplate(ctx context.Context, id string, input model.NotificationTemplateInput) (*model.NotificationTemplate, error) {
	template, err := d.dbPort.NotificationTemplate().FindByID(id)
	if err != nil {
		return nil, fmt.Errorf("template not found: %w", err)
	}

	template.Name = input.Name
	template.Type = input.Type
	template.Subject = input.Subject
	template.Content = input.Content
	template.Variables = input.Variables

	if input.IsActive != nil {
		template.IsActive = *input.IsActive
	}

	err = d.dbPort.NotificationTemplate().Update(template)
	if err != nil {
		return nil, err
	}

	return template, nil
}

func (d *domain) DeleteTemplate(ctx context.Context, id string) error {
	return d.dbPort.NotificationTemplate().Delete(id)
}

func (d *domain) RenderTemplate(ctx context.Context, templateName string, variables map[string]interface{}) (string, string, error) {
	tmpl, err := d.dbPort.NotificationTemplate().FindByName(templateName)
	if err != nil {
		return "", "", fmt.Errorf("template not found: %w", err)
	}

	if !tmpl.IsActive {
		return "", "", fmt.Errorf("template is inactive")
	}

	renderer := template.NewRenderer(tmpl.Content)
	renderedContent, err := renderer.RenderWithMap(variables)
	if err != nil {
		return "", "", fmt.Errorf("failed to render template content: %w", err)
	}

	renderedSubject := tmpl.Subject
	if tmpl.Subject != "" {
		subjectRenderer := template.NewRenderer(tmpl.Subject)
		renderedSubject, err = subjectRenderer.RenderWithMap(variables)
		if err != nil {
			log.WithContext(ctx).Warn(fmt.Sprintf("Failed to render template subject: %v", err))
		}
	}

	return renderedSubject, renderedContent, nil
}

func (d *domain) SendNotificationFromTemplate(ctx context.Context, templateName string, notificationType string, recipient string, variables map[string]interface{}) error {
	subject, content, err := d.RenderTemplate(ctx, templateName, variables)
	if err != nil {
		return err
	}

	input := model.NotificationInput{
		Type:      notificationType,
		Recipient: recipient,
		Subject:   &subject,
		Content:   content,
	}

	_, err = d.CreateNotification(ctx, input)
	if err != nil {
		return err
	}

	return nil
}

func (d *domain) SendPaymentConfirmation(ctx context.Context, customerID string, paymentID string, amount float64) error {
	customer, err := d.dbPort.Customer().FindByID(customerID)
	if err != nil {
		return fmt.Errorf("customer not found: %w", err)
	}

	payment, err := d.dbPort.Payment().FindByID(paymentID)
	if err != nil {
		return fmt.Errorf("payment not found: %w", err)
	}

	portalURL := d.getPaymentPortalURL()
	variables := map[string]interface{}{
		"CustomerName":  customer.FullName,
		"CustomerCode":  customer.CustomerCode,
		"PaymentNumber": payment.PaymentNumber,
		"PaymentAmount": fmt.Sprintf("%.2f", amount),
		"PaymentDate":   payment.PaymentDate.Format("2006-01-02"),
		"PaymentMethod": payment.PaymentMethod,
		"PortalURL":     portalURL,
	}

	var emailSubject, emailContent string

	if customer.Email != nil {
		subject, content, err := d.RenderTemplate(ctx, "payment_confirmation_email", variables)
		if err != nil {
			log.WithContext(ctx).Warn(fmt.Sprintf("Template not found or failed to render, using fallback: %v", err))
			emailSubject = fmt.Sprintf("Payment Confirmation - %s", payment.PaymentNumber)
			emailContent = fmt.Sprintf(`
Dear %s,

Thank you for your payment!

Payment Details:
- Payment Number: %s
- Amount: Rp %.2f
- Payment Date: %s
- Payment Method: %s

Your payment has been successfully received and applied to your invoice(s).

If you have any questions, please contact our support team.

Best regards,
Your Billing Team
			`, customer.FullName, payment.PaymentNumber, amount, payment.PaymentDate.Format("2006-01-02"), payment.PaymentMethod)
		} else {
			emailSubject = subject
			emailContent = content
		}

		emailInput := model.NotificationInput{
			Type:      "email",
			Recipient: *customer.Email,
			Subject:   &emailSubject,
			Content:   emailContent,
		}

		if customer.ID.String() != "" {
			emailInput.CustomerID = &customer.ID
			emailInput.PaymentID = &payment.ID
		}

		_, err = d.CreateNotification(ctx, emailInput)
		if err != nil {
			log.WithContext(ctx).Warn(fmt.Sprintf("Failed to create email notification: %v", err))
		}
	}

	if customer.Phone != "" {
		_, content, err := d.RenderTemplate(ctx, "payment_confirmation_whatsapp", variables)
		whatsappMessage := content
		if err != nil {
			log.WithContext(ctx).Warn(fmt.Sprintf("WhatsApp template not found or failed to render, using fallback: %v", err))
			whatsappMessage = fmt.Sprintf("Dear %s,\n\nThank you for your payment of Rp %.2f (Payment No: %s). Your payment has been successfully received.\n\nBest regards", customer.FullName, amount, payment.PaymentNumber)
		}

		whatsappInput := model.NotificationInput{
			Type:      "whatsapp",
			Recipient: customer.Phone,
			Content:   whatsappMessage,
		}

		if customer.ID.String() != "" {
			whatsappInput.CustomerID = &customer.ID
			whatsappInput.PaymentID = &payment.ID
		}

		_, err = d.CreateNotification(ctx, whatsappInput)
		if err != nil {
			log.WithContext(ctx).Warn(fmt.Sprintf("Failed to create WhatsApp notification: %v", err))
		}
	}

	return nil
}

func (d *domain) SendInvoiceReminder(ctx context.Context, customerID string, invoiceID string, days int) error {
	customer, err := d.dbPort.Customer().FindByID(customerID)
	if err != nil {
		return fmt.Errorf("customer not found: %w", err)
	}

	invoice, err := d.dbPort.Invoice().FindByID(invoiceID)
	if err != nil {
		return fmt.Errorf("invoice not found: %w", err)
	}

	portalURL := d.getPaymentPortalURL()
	outstandingBalance := invoice.TotalAmount - invoice.PaidAmount
	variables := map[string]interface{}{
		"CustomerName":       customer.FullName,
		"CustomerCode":       customer.CustomerCode,
		"InvoiceNumber":      invoice.InvoiceNumber,
		"InvoiceDate":        invoice.IssueDate.Format("2006-01-02"),
		"DueDate":            invoice.DueDate.Format("2006-01-02"),
		"TotalAmount":        fmt.Sprintf("%.2f", invoice.TotalAmount),
		"PaidAmount":         fmt.Sprintf("%.2f", invoice.PaidAmount),
		"OutstandingBalance": fmt.Sprintf("%.2f", outstandingBalance),
		"DaysOverdue":        days,
		"PortalURL":          portalURL,
	}

	var emailSubject, emailContent string

	if customer.Email != nil {
		subject, content, err := d.RenderTemplate(ctx, "invoice_reminder_email", variables)
		if err != nil {
			log.WithContext(ctx).Warn(fmt.Sprintf("Template not found or failed to render, using fallback: %v", err))
			emailSubject = fmt.Sprintf("Invoice Reminder - %s", invoice.InvoiceNumber)
			emailContent = fmt.Sprintf(`
Dear %s,

This is a friendly reminder that your invoice is %d days overdue.

Invoice Details:
- Invoice Number: %s
- Invoice Date: %s
- Due Date: %s
- Total Amount: Rp %.2f
- Outstanding Balance: Rp %.2f

Please complete your payment as soon as possible to avoid additional late fees.

You can pay your invoice through our payment portal: %s

If you have already made a payment, please disregard this notice.

Best regards,
Your Billing Team
			`, customer.FullName, days, invoice.InvoiceNumber, invoice.IssueDate.Format("2006-01-02"), invoice.DueDate.Format("2006-01-02"), invoice.TotalAmount, outstandingBalance, portalURL)
		} else {
			emailSubject = subject
			emailContent = content
		}

		emailInput := model.NotificationInput{
			Type:      "email",
			Recipient: *customer.Email,
			Subject:   &emailSubject,
			Content:   emailContent,
		}

		if customer.ID.String() != "" {
			emailInput.CustomerID = &customer.ID
			emailInput.InvoiceID = &invoice.ID
		}

		_, err = d.CreateNotification(ctx, emailInput)
		if err != nil {
			return err
		}
	}

	return nil
}

func (d *domain) SendPaymentFailed(ctx context.Context, customerID string, invoiceID string, reason string) error {
	customer, err := d.dbPort.Customer().FindByID(customerID)
	if err != nil {
		return fmt.Errorf("customer not found: %w", err)
	}

	invoice, err := d.dbPort.Invoice().FindByID(invoiceID)
	if err != nil {
		return fmt.Errorf("invoice not found: %w", err)
	}

	portalURL := d.getPaymentPortalURL()
	variables := map[string]interface{}{
		"CustomerName":  customer.FullName,
		"CustomerCode":  customer.CustomerCode,
		"InvoiceNumber": invoice.InvoiceNumber,
		"TotalAmount":   fmt.Sprintf("%.2f", invoice.TotalAmount),
		"FailedReason":  reason,
		"PortalURL":     portalURL,
	}

	var emailSubject, emailContent string

	if customer.Email != nil {
		subject, content, err := d.RenderTemplate(ctx, "payment_failed_email", variables)
		if err != nil {
			log.WithContext(ctx).Warn(fmt.Sprintf("Template not found or failed to render, using fallback: %v", err))
			emailSubject = fmt.Sprintf("Payment Failed - %s", invoice.InvoiceNumber)
			emailContent = fmt.Sprintf(`
Dear %s,

We regret to inform you that your payment for invoice %s has failed.

Payment Details:
- Invoice Number: %s
- Failed Reason: %s
- Total Amount: Rp %.2f

Please try again or use a different payment method. If issue persists, please contact our support team.

You can retry payment through our payment portal: %s

Best regards,
Your Billing Team
			`, customer.FullName, invoice.InvoiceNumber, invoice.InvoiceNumber, reason, invoice.TotalAmount, portalURL)
		} else {
			emailSubject = subject
			emailContent = content
		}

		emailInput := model.NotificationInput{
			Type:      "email",
			Recipient: *customer.Email,
			Subject:   &emailSubject,
			Content:   emailContent,
		}

		if customer.ID.String() != "" {
			emailInput.CustomerID = &customer.ID
			emailInput.InvoiceID = &invoice.ID
		}

		_, err = d.CreateNotification(ctx, emailInput)
		if err != nil {
			return err
		}
	}

	return nil
}

func (d *domain) SendInvoiceCreated(ctx context.Context, customerID string, invoiceID string) error {
	customer, err := d.dbPort.Customer().FindByID(customerID)
	if err != nil {
		return fmt.Errorf("customer not found: %w", err)
	}

	invoice, err := d.dbPort.Invoice().FindByID(invoiceID)
	if err != nil {
		return fmt.Errorf("invoice not found: %w", err)
	}

	portalURL := d.getPaymentPortalURL()
	variables := map[string]interface{}{
		"CustomerName":  customer.FullName,
		"CustomerCode":  customer.CustomerCode,
		"InvoiceNumber": invoice.InvoiceNumber,
		"InvoiceDate":   invoice.IssueDate.Format("2006-01-02"),
		"DueDate":       invoice.DueDate.Format("2006-01-02"),
		"TotalAmount":   fmt.Sprintf("%.2f", invoice.TotalAmount),
		"PortalURL":     portalURL,
	}

	var emailSubject, emailContent string

	if customer.Email != nil {
		subject, content, err := d.RenderTemplate(ctx, "invoice_created_email", variables)
		if err != nil {
			log.WithContext(ctx).Warn(fmt.Sprintf("Template not found or failed to render, using fallback: %v", err))
			emailSubject = fmt.Sprintf("New Invoice - %s", invoice.InvoiceNumber)
			emailContent = fmt.Sprintf(`
Dear %s,

We are pleased to inform you that a new invoice has been generated for your account.

Invoice Details:
- Invoice Number: %s
- Invoice Date: %s
- Due Date: %s
- Total Amount: Rp %.2f

Please complete your payment before the due date to avoid late fees.

You can view and pay your invoice through our payment portal: %s

If you have any questions, please contact our support team.

Best regards,
Your Billing Team
			`, customer.FullName, invoice.InvoiceNumber, invoice.IssueDate.Format("2006-01-02"), invoice.DueDate.Format("2006-01-02"), invoice.TotalAmount, portalURL)
		} else {
			emailSubject = subject
			emailContent = content
		}

		emailInput := model.NotificationInput{
			Type:      "email",
			Recipient: *customer.Email,
			Subject:   &emailSubject,
			Content:   emailContent,
		}

		if customer.ID.String() != "" {
			emailInput.CustomerID = &customer.ID
			emailInput.InvoiceID = &invoice.ID
		}

		_, err = d.CreateNotification(ctx, emailInput)
		if err != nil {
			return err
		}
	}

	return nil
}

// SendActivationNotification sends a notification when customer is activated/reactivated
func (d *domain) SendActivationNotification(ctx context.Context, customerID string, message string) error {
	customer, err := d.dbPort.Customer().FindByID(customerID)
	if err != nil {
		return fmt.Errorf("customer not found: %w", err)
	}

	portalURL := d.getPaymentPortalURL()
	whatsappMessage := fmt.Sprintf(
		"🟢 *AKUN DIAKTIFKAN*\n\n"+
			"Nama: %s\n"+
			"Kode Pelanggan: %s\n"+
			"Status: Aktif\n"+
			"Layanan internet telah diaktifkan kembali.\n\n"+
			"%s\n\n"+
			"Terima kasih!\n\n"+
			"Portal: %s",
		customer.FullName,
		customer.CustomerCode,
		message,
		portalURL,
	)

	if customer.Phone != "" {
		// Send directly to customer's WhatsApp
		_, err = d.gowaUtil.SendTextMessage(customer.Phone, whatsappMessage)
		if err != nil {
			log.WithContext(ctx).Error(fmt.Sprintf("Failed to send WhatsApp activation notification: %v", err))
			// Create pending notification for retry
			whatsappInput := model.NotificationInput{
				Type:      "whatsapp",
				Recipient: customer.Phone,
				Content:   whatsappMessage,
			}
			if customer.ID.String() != "" {
				whatsappInput.CustomerID = &customer.ID
			}
			d.CreateNotification(ctx, whatsappInput)
		} else {
			log.WithContext(ctx).Info(fmt.Sprintf("Activation notification sent to %s", customer.Phone))
		}
	}

	return nil
}

// SendIsolationNotification sends a notification when customer is isolated
func (d *domain) SendIsolationNotification(ctx context.Context, customerID string, reason string) error {
	customer, err := d.dbPort.Customer().FindByID(customerID)
	if err != nil {
		return fmt.Errorf("customer not found: %w", err)
	}

	portalURL := d.getPaymentPortalURL()
	whatsappMessage := fmt.Sprintf(
		"🔴 *AKUN TERISOLASI*\n\n"+
			"Nama: %s\n"+
			"Kode Pelanggan: %s\n"+
			"Status: Terisolasi (terbatas)\n\n"+
			"Alasan: %s\n\n"+
			"Silakan lakukan pembayaran melalui portal:\n%s\n\n"+
			"Terima kasih!\n\n"+
			"Portal: %s",
		customer.FullName,
		customer.CustomerCode,
		reason,
		portalURL,
	)

	if customer.Phone != "" {
		// Send directly to customer's WhatsApp
		_, err = d.gowaUtil.SendTextMessage(customer.Phone, whatsappMessage)
		if err != nil {
			log.WithContext(ctx).Error(fmt.Sprintf("Failed to send WhatsApp isolation notification: %v", err))
			// Create pending notification for retry
			whatsappInput := model.NotificationInput{
				Type:      "whatsapp",
				Recipient: customer.Phone,
				Content:   whatsappMessage,
			}
			if customer.ID.String() != "" {
				whatsappInput.CustomerID = &customer.ID
			}
			d.CreateNotification(ctx, whatsappInput)
		} else {
			log.WithContext(ctx).Info(fmt.Sprintf("Isolation notification sent to %s", customer.Phone))
		}
	}

	return nil
}

// SendReactivationNotification sends a notification when customer is reactivated after isolation
func (d *domain) SendReactivationNotification(ctx context.Context, customerID string, message string) error {
	return d.SendActivationNotification(ctx, customerID, message)
}

// SendToGroup sends a notification to a WhatsApp group
func (d *domain) SendToGroup(ctx context.Context, groupID string, message string) error {
	if !d.gowaUtil.IsEnabled() {
		return fmt.Errorf("Gowa is disabled")
	}

	_, err := d.gowaUtil.SendToGroup(groupID, message)
	if err != nil {
		log.WithContext(ctx).Error(fmt.Sprintf("Failed to send group notification: %v", err))
		return err
	}

	log.WithContext(ctx).Info(fmt.Sprintf("Group notification sent to %s", groupID))
	return nil
}

// SendPPPoEConnectionNotification sends notification about PPPoE connection status
func (d *domain) SendPPPoEConnectionNotification(ctx context.Context, customerID string, status string, details map[string]string) error {
	customer, err := d.dbPort.Customer().FindByID(customerID)
	if err != nil {
		return fmt.Errorf("customer not found: %w", err)
	}

	groupID := d.getNotificationGroupID()
	if groupID == "" {
		return fmt.Errorf("notification group ID not configured")
	}

	// Build message based on status
	var emoji, statusText string
	if status == "up" || status == "active" {
		emoji = "🟢"
		statusText = "TERHUBUNG"
	} else {
		emoji = "🔴"
		statusText = "TERPUTUS"
	}

	message := fmt.Sprintf(
		"%s *STATUS KONEKSI PPPoE*\n\n"+
			"Nama: %s\n"+
			"Kode Pelanggan: %s\n"+
			"Status: %s\n\n"+
			"Detail:\n",
		emoji,
		customer.FullName,
		customer.CustomerCode,
		statusText,
	)

	// Add additional details
	for key, value := range details {
		message += fmt.Sprintf("- %s: %s\n", key, value)
	}

	// Send to group
	return d.SendToGroup(ctx, groupID, message)
}

// Helper method to get payment portal URL
func (d *domain) getPaymentPortalURL() string {
	url, err := settings.GetStringSetting(d.dbPort, "payment.portal_url")
	if err != nil {
		return "https://portal.example.com"
	}
	return url
}

// Helper method to get notification group ID
func (d *domain) getNotificationGroupID() string {
	groupID, err := settings.GetStringSetting(d.dbPort, "whatsapp.group_notifications")
	if err != nil {
		return ""
	}
	return groupID
}
