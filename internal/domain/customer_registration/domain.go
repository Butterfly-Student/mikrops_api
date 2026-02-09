package customer_registration

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/palantir/stacktrace"

	"mikrops/internal/model"
	outbound_port "mikrops/internal/port/outbound"
)

type CustomerRegistrationDomain interface {
	Create(ctx context.Context, input model.CustomerRegistrationInput) (model.CustomerRegistration, error)
	FindByFilter(ctx context.Context, filter model.CustomerRegistrationFilter) ([]model.CustomerRegistration, error)
	FindByID(ctx context.Context, id string) (model.CustomerRegistration, error)
	Approve(ctx context.Context, id string, staffID string) (model.CustomerRegistration, error)
	Reject(ctx context.Context, id string, reason string) (model.CustomerRegistration, error)
}

type customerRegistrationDomain struct {
	databasePort outbound_port.DatabasePort
	messagePort  outbound_port.MessagePort
	cachePort    outbound_port.CachePort
	workflowPort outbound_port.WorkflowPort
}

func NewCustomerRegistrationDomain(
	databasePort outbound_port.DatabasePort,
	messagePort outbound_port.MessagePort,
	cachePort outbound_port.CachePort,
	workflowPort outbound_port.WorkflowPort,
) CustomerRegistrationDomain {
	return &customerRegistrationDomain{
		databasePort: databasePort,
		messagePort:  messagePort,
		cachePort:    cachePort,
		workflowPort: workflowPort,
	}
}

func (d *customerRegistrationDomain) Create(ctx context.Context, input model.CustomerRegistrationInput) (model.CustomerRegistration, error) {
	if input.TenantID == "" {
		return model.CustomerRegistration{}, stacktrace.NewError("tenant_id is required")
	}
	if input.FullName == "" {
		return model.CustomerRegistration{}, stacktrace.NewError("full_name is required")
	}
	if input.RequestedPackageID == "" {
		return model.CustomerRegistration{}, stacktrace.NewError("requested_package_id is required")
	}

	input.Status = model.RegistrationStatusPending

	registration, err := d.databasePort.CustomerRegistration().Create(input)
	if err != nil {
		return model.CustomerRegistration{}, stacktrace.Propagate(err, "failed to create customer registration")
	}

	return registration, nil
}

func (d *customerRegistrationDomain) FindByFilter(ctx context.Context, filter model.CustomerRegistrationFilter) ([]model.CustomerRegistration, error) {
	if filter.IsEmpty() {
		return nil, stacktrace.NewError("filter is empty")
	}

	registrations, err := d.databasePort.CustomerRegistration().FindByFilter(filter)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to find customer registrations")
	}

	return registrations, nil
}

func (d *customerRegistrationDomain) FindByID(ctx context.Context, id string) (model.CustomerRegistration, error) {
	if id == "" {
		return model.CustomerRegistration{}, stacktrace.NewError("id is empty")
	}

	registration, err := d.databasePort.CustomerRegistration().FindByID(id)
	if err != nil {
		return model.CustomerRegistration{}, stacktrace.Propagate(err, "failed to find customer registration")
	}

	return registration, nil
}

func (d *customerRegistrationDomain) Approve(ctx context.Context, id string, staffID string) (model.CustomerRegistration, error) {
	if id == "" {
		return model.CustomerRegistration{}, stacktrace.NewError("id is empty")
	}

	registration, err := d.databasePort.CustomerRegistration().FindByID(id)
	if err != nil {
		return model.CustomerRegistration{}, stacktrace.Propagate(err, "failed to find registration")
	}

	if registration.Status != model.RegistrationStatusPending {
		return model.CustomerRegistration{}, stacktrace.NewError("registration is not in pending status")
	}

	// Get tenant settings for PPPoE credential generation and other configs
	tenantSettings, err := d.databasePort.TenantSetting().FindByTenantID(registration.TenantID)
	if err != nil {
		// Use default settings if not found
		tenantSettings = model.TenantSetting{
			TenantSettingInput: model.TenantSettingInput{
				TenantID:                 registration.TenantID,
				CutoffDay:                1,
				GracePeriod:              3,
				IsolirProfileName:        "ISOLIR_MIKROPS",
				DefaultPppPasswordLength: 8,
				DefaultPppPasswordType:   "random",
			},
		}
	}

	// Create customer and all related resources in transaction
	result, err := d.databasePort.DoInTransaction(func(repo outbound_port.DatabasePort) (interface{}, error) {
		// 1. Create customer from registration data
		customerInput := model.CustomerInput{
			TenantID: registration.TenantID,
			FullName: registration.FullName,
			Email:    registration.Email,
			Phone:    registration.Phone,
			Address:  registration.Address,
		}
		if registration.RequestedNasID != nil {
			customerInput.NasID = registration.RequestedNasID
		}

		customer, err := repo.Customer().Create(customerInput)
		if err != nil {
			return nil, stacktrace.Propagate(err, "failed to create customer from registration")
		}

		// 2. Generate PPPoE credentials based on tenant settings
		pppoeUsername := d.generatePPPoEUsername(customer.ID)
		pppoePassword := d.generatePPPoEPassword(tenantSettings.DefaultPppPasswordLength)

		// 3. Get requested package and NAS
		requestedPackage, err := repo.InternetPackage().FindByID(registration.RequestedPackageID)
		if err != nil {
			return nil, stacktrace.Propagate(err, "failed to fetch requested package")
		}

		nasID := registration.RequestedNasID
		if nasID == nil || *nasID == "" {
			return nil, stacktrace.NewError("nas_id is required")
		}

		// 4. Create PPPoE account
		pppoeInput := model.PppoeAccountInput{
			TenantID:          registration.TenantID,
			CustomerID:        customer.ID,
			NasID:             *nasID,
			PackageID:         requestedPackage.ID,
			Username:          pppoeUsername,
			PasswordEncrypted: pppoePassword,
			ProfileName:       requestedPackage.ProfileName,
			Status:            model.PppoeStatusActive,
			SyncStatus:        model.PppoeSyncStatusPending,
		}

		pppoeAccount, err := repo.PppoeAccount().Create(pppoeInput)
		if err != nil {
			return nil, stacktrace.Propagate(err, "failed to create pppoe account")
		}

		// 5. Create subscription with pppoe_account_id
		now := time.Now()
		endDate := now.AddDate(0, 1, 0) // 1 month from now
		subscriptionInput := model.SubscriptionInput{
			TenantID:       registration.TenantID,
			CustomerID:     customer.ID,
			PackageID:      requestedPackage.ID,
			NasID:          *nasID,
			PppoeAccountID: &pppoeAccount.ID,
			Status:         model.SubscriptionStatusActive,
			StartDate:      now,
			EndDate:        endDate,
			AutoRenew:      true,
		}

		subscription, err := repo.Subscription().Create(subscriptionInput)
		if err != nil {
			return nil, stacktrace.Propagate(err, "failed to create subscription")
		}

		// 6. Generate first invoice
		dueDate := now.AddDate(0, 0, tenantSettings.CutoffDay) // Due on cutoff day
		if dueDate.Before(now) {
			// If cutoff day already passed this month, set to next month
			dueDate = dueDate.AddDate(0, 1, 0)
		}

		invoiceNumber := d.generateInvoiceNumber(registration.TenantID, now)
		invoiceInput := model.InvoiceInput{
			TenantID:       registration.TenantID,
			CustomerID:     customer.ID,
			SubscriptionID: subscription.ID,
			InvoiceNumber:  invoiceNumber,
			Amount:         requestedPackage.Price,
			TaxAmount:      0,
			TotalAmount:    requestedPackage.Price,
			Status:         model.InvoiceStatusUnpaid,
			DueDate:        dueDate,
			PeriodStart:    now,
			PeriodEnd:      endDate,
			Notes:          "First invoice for new customer registration",
		}

		invoice, err := repo.Invoice().Create(invoiceInput)
		if err != nil {
			return nil, stacktrace.Propagate(err, "failed to create first invoice")
		}

		// 7. Update registration status
		updateInput := model.CustomerRegistrationInput{
			Status:     model.RegistrationStatusApproved,
			CustomerID: &customer.ID,
			ApprovedBy: &staffID,
			ApprovedAt: &now,
		}
		err = repo.CustomerRegistration().Update(id, updateInput)
		if err != nil {
			return nil, stacktrace.Propagate(err, "failed to update registration status")
		}

		return map[string]interface{}{
			"customer":      customer,
			"pppoe_account": pppoeAccount,
			"subscription":  subscription,
			"invoice":       invoice,
		}, nil
	})
	if err != nil {
		return model.CustomerRegistration{}, stacktrace.Propagate(err, "failed to approve registration")
	}

	resultMap := result.(map[string]interface{})
	customer := resultMap["customer"].(model.Customer)
	pppoeAccount := resultMap["pppoe_account"].(model.PppoeAccount)

	// 8. Enqueue RabbitMQ for MikroTik sync
	// TODO: messagePort.Publish not implemented yet
	// message := map[string]interface{}{
	// 	"action":           "create",
	// 	"pppoe_account_id": pppoeAccount.ID,
	// 	"nas_id":           pppoeAccount.NasID,
	// 	"username":         pppoeAccount.Username,
	// 	"password":         pppoeAccount.PasswordEncrypted,
	// 	"profile":          pppoeAccount.ProfileName,
	// }
	// _ = d.messagePort.Publish(ctx, "mikrotik.ppp.create", message)

	// 9. Send WhatsApp notification with credentials (placeholder)
	// TODO: Implement WhatsApp gateway integration
	// whatsappMessage := fmt.Sprintf("Welcome! Your internet credentials:\nUsername: %s\nPassword: %s", pppoeAccount.Username, pppoeAccount.PasswordEncrypted)
	// _ = d.whatsappPort.SendCredentials(customer.Phone, whatsappMessage)

	// 10. Log activity
	detailsJSON, _ := json.Marshal(map[string]string{
		"staff_id":    staffID,
		"customer_id": customer.ID,
		"pppoe_id":    pppoeAccount.ID,
	})
	activityInput := model.ActivityLogInput{
		TenantID:     registration.TenantID,
		Action:       "approve",
		ResourceType: "customer_registration",
		ResourceID:   id,
		Details:      detailsJSON,
		CreatedAt:    time.Now(),
	}
	_, _ = d.databasePort.ActivityLog().Create(activityInput)

	updatedRegistration, err := d.databasePort.CustomerRegistration().FindByID(id)
	if err != nil {
		return model.CustomerRegistration{}, stacktrace.Propagate(err, "failed to fetch updated registration")
	}

	return updatedRegistration, nil
}

func (d *customerRegistrationDomain) Reject(ctx context.Context, id string, reason string) (model.CustomerRegistration, error) {
	if id == "" {
		return model.CustomerRegistration{}, stacktrace.NewError("id is empty")
	}

	registration, err := d.databasePort.CustomerRegistration().FindByID(id)
	if err != nil {
		return model.CustomerRegistration{}, stacktrace.Propagate(err, "failed to find registration")
	}

	if registration.Status != model.RegistrationStatusPending {
		return model.CustomerRegistration{}, stacktrace.NewError("registration is not in pending status")
	}

	updateInput := model.CustomerRegistrationInput{
		Status:          model.RegistrationStatusRejected,
		RejectionReason: reason,
	}
	err = d.databasePort.CustomerRegistration().Update(id, updateInput)
	if err != nil {
		return model.CustomerRegistration{}, stacktrace.Propagate(err, "failed to reject registration")
	}

	// TODO: Send WhatsApp rejection notification
	// _ = d.whatsappPort.SendRejectionNotice(registration.Phone, reason)

	updatedRegistration, err := d.databasePort.CustomerRegistration().FindByID(id)
	if err != nil {
		return model.CustomerRegistration{}, stacktrace.Propagate(err, "failed to fetch updated registration")
	}

	return updatedRegistration, nil
}

// Helper methods

// generatePPPoEUsername generates a PPPoE username based on customer ID
func (d *customerRegistrationDomain) generatePPPoEUsername(customerID string) string {
	// Use first 8 chars of customer ID
	if len(customerID) >= 8 {
		return fmt.Sprintf("pppoe_%s", customerID[:8])
	}
	// Generate random username
	b := make([]byte, 4)
	rand.Read(b)
	return fmt.Sprintf("pppoe_%s", hex.EncodeToString(b))
}

// generatePPPoEPassword generates a random password with specified length
func (d *customerRegistrationDomain) generatePPPoEPassword(length int) string {
	if length < 6 {
		length = 8 // minimum length
	}

	bytes := make([]byte, length/2+1)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)[:length]
}

// generateInvoiceNumber generates a unique invoice number
func (d *customerRegistrationDomain) generateInvoiceNumber(tenantID string, timestamp time.Time) string {
	// Format: INV-YYYYMMDD-XXXXXX
	datePart := timestamp.Format("20060102")
	randomBytes := make([]byte, 3)
	rand.Read(randomBytes)
	randomPart := hex.EncodeToString(randomBytes)

	return fmt.Sprintf("INV-%s-%s", datePart, strings.ToUpper(randomPart))
}
