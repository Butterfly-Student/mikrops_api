package notification

import (
	"context"
	"fmt"
	"time"

	"go-template/internal/model"
	outbound_port "go-template/internal/port/outbound"
	"go-template/utils/email"
	"go-template/utils/log"
	"go-template/utils/whatsapp"
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
}

type domain struct {
	dbPort       outbound_port.DatabasePort
	emailUtil    *email.EmailUtil
	whatsappUtil *whatsapp.WhatsAppUtil
}

func NewNotificationDomain(
	dbPort outbound_port.DatabasePort,
	emailUtil *email.EmailUtil,
	whatsappUtil *whatsapp.WhatsAppUtil,
) NotificationDomain {
	return &domain{
		dbPort:       dbPort,
		emailUtil:    emailUtil,
		whatsappUtil: whatsappUtil,
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
		sendErr = d.whatsappUtil.SendMessage(notification.Recipient, message)
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

func (d *domain) SendPaymentConfirmation(ctx context.Context, customerID string, paymentID string, amount float64) error {
	customer, err := d.dbPort.Customer().FindByID(customerID)
	if err != nil {
		return fmt.Errorf("customer not found: %w", err)
	}

	payment, err := d.dbPort.Payment().FindByID(paymentID)
	if err != nil {
		return fmt.Errorf("payment not found: %w", err)
	}

	// Send email
	emailSubject := fmt.Sprintf("Payment Confirmation - %s", payment.PaymentNumber)
	emailContent := fmt.Sprintf(`
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

	emailInput := model.NotificationInput{
		Type:      "email",
		Recipient: "",
		Subject:   &emailSubject,
		Content:   emailContent,
	}

	if customer.Email != nil {
		emailInput.Recipient = *customer.Email
	}

	if customer.ID.String() != "" {
		emailInput.CustomerID = &customer.ID
		emailInput.PaymentID = &payment.ID
	}

	_, err = d.CreateNotification(ctx, emailInput)
	if err != nil {
		return err
	}

	// Send WhatsApp if phone number is available
	if customer.Phone != "" {
		whatsappMessage := fmt.Sprintf("Dear %s,\n\nThank you for your payment of Rp %.2f (Payment No: %s). Your payment has been successfully received.\n\nBest regards", customer.FullName, amount, payment.PaymentNumber)
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

	emailSubject := fmt.Sprintf("Invoice Reminder - %s", invoice.InvoiceNumber)
	emailContent := fmt.Sprintf(`
Dear %s,

This is a friendly reminder that your invoice is %s days overdue.

Invoice Details:
- Invoice Number: %s
- Invoice Date: %s
- Due Date: %s
- Total Amount: Rp %.2f
- Outstanding Balance: Rp %.2f

Please complete your payment as soon as possible to avoid additional late fees.

You can pay your invoice through our payment portal: https://your-domain.com/payment/

If you have already made a payment, please disregard this notice.

Best regards,
Your Billing Team
	`, customer.FullName, days, invoice.InvoiceNumber, invoice.IssueDate.Format("2006-01-02"), invoice.DueDate.Format("2006-01-02"), invoice.TotalAmount, invoice.TotalAmount-invoice.PaidAmount)

	emailInput := model.NotificationInput{
		Type:      "email",
		Recipient: "",
		Subject:   &emailSubject,
		Content:   emailContent,
	}

	if customer.Email != nil {
		emailInput.Recipient = *customer.Email
	}

	if customer.ID.String() != "" {
		emailInput.CustomerID = &customer.ID
		emailInput.InvoiceID = &invoice.ID
	}

	_, err = d.CreateNotification(ctx, emailInput)
	if err != nil {
		return err
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

	emailSubject := fmt.Sprintf("Payment Failed - %s", invoice.InvoiceNumber)
	emailContent := fmt.Sprintf(`
Dear %s,

We regret to inform you that your payment for invoice %s has failed.

Payment Details:
- Invoice Number: %s
- Failed Reason: %s
- Total Amount: Rp %.2f

Please try again or use a different payment method. If the issue persists, please contact our support team.

You can retry payment through our payment portal: https://your-domain.com/payment/

Best regards,
Your Billing Team
	`, customer.FullName, invoice.InvoiceNumber, invoice.InvoiceNumber, reason, invoice.TotalAmount)

	emailInput := model.NotificationInput{
		Type:      "email",
		Recipient: "",
		Subject:   &emailSubject,
		Content:   emailContent,
	}

	if customer.Email != nil {
		emailInput.Recipient = *customer.Email
	}

	if customer.ID.String() != "" {
		emailInput.CustomerID = &customer.ID
		emailInput.InvoiceID = &invoice.ID
	}

	_, err = d.CreateNotification(ctx, emailInput)
	if err != nil {
		return err
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

	emailSubject := fmt.Sprintf("New Invoice - %s", invoice.InvoiceNumber)
	emailContent := fmt.Sprintf(`
Dear %s,

We are pleased to inform you that a new invoice has been generated for your account.

Invoice Details:
- Invoice Number: %s
- Invoice Date: %s
- Due Date: %s
- Total Amount: Rp %.2f

Please complete your payment before the due date to avoid late fees.

You can view and pay your invoice through our payment portal: https://your-domain.com/payment/

If you have any questions, please contact our support team.

Best regards,
Your Billing Team
	`, customer.FullName, invoice.InvoiceNumber, invoice.IssueDate.Format("2006-01-02"), invoice.DueDate.Format("2006-01-02"), invoice.TotalAmount)

	emailInput := model.NotificationInput{
		Type:      "email",
		Recipient: "",
		Subject:   &emailSubject,
		Content:   emailContent,
	}

	if customer.Email != nil {
		emailInput.Recipient = *customer.Email
	}

	if customer.ID.String() != "" {
		emailInput.CustomerID = &customer.ID
		emailInput.InvoiceID = &invoice.ID
	}

	_, err = d.CreateNotification(ctx, emailInput)
	if err != nil {
		return err
	}

	return nil
}
