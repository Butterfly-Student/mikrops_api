package cutoff

import (
	"context"
	"encoding/json"
	"time"

	"github.com/palantir/stacktrace"

	"mikrops/internal/model"
	outbound_port "mikrops/internal/port/outbound"
)

type CutoffDomain interface {
	RunDailyCutoff(ctx context.Context) error
	ScanTenantsForCutoff(ctx context.Context) ([]TenantCutoffInfo, error)
	IsolateCustomer(ctx context.Context, customerID string) error
	RestoreCustomer(ctx context.Context, customerID string) error
	OnInvoicePaid(ctx context.Context, invoiceID string) error
}

type TenantCutoffInfo struct {
	TenantID    string
	TenantName  string
	CutoffDay   int
	GracePeriod int
}

type cutoffDomain struct {
	databasePort outbound_port.DatabasePort
	messagePort  outbound_port.MessagePort
	cachePort    outbound_port.CachePort
	workflowPort outbound_port.WorkflowPort
}

func NewCutoffDomain(
	databasePort outbound_port.DatabasePort,
	messagePort outbound_port.MessagePort,
	cachePort outbound_port.CachePort,
	workflowPort outbound_port.WorkflowPort,
) CutoffDomain {
	return &cutoffDomain{
		databasePort: databasePort,
		messagePort:  messagePort,
		cachePort:    cachePort,
		workflowPort: workflowPort,
	}
}

// RunDailyCutoff is the main cron job that runs daily to isolate customers with overdue invoices
func (d *cutoffDomain) RunDailyCutoff(ctx context.Context) error {
	// Get all tenants that have cutoff day = today
	tenantsInfo, err := d.ScanTenantsForCutoff(ctx)
	if err != nil {
		return stacktrace.Propagate(err, "failed to scan tenants for cutoff")
	}

	if len(tenantsInfo) == 0 {
		// No tenants with cutoff today
		return nil
	}

	// For each tenant, find customers with overdue invoices
	for _, tenantInfo := range tenantsInfo {
		err := d.processTenantCutoff(ctx, tenantInfo)
		if err != nil {
			// Log error but continue with other tenants
			// TODO: Log to activity_logs
			continue
		}
	}

	return nil
}

// ScanTenantsForCutoff finds all tenants that have cutoff_day matching today
func (d *cutoffDomain) ScanTenantsForCutoff(ctx context.Context) ([]TenantCutoffInfo, error) {
	today := time.Now().Day()

	// Get all tenants (we'll filter by cutoff_day from tenant_settings)
	tenants, err := d.databasePort.Tenant().FindByFilter(model.TenantFilter{})
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to fetch tenants")
	}

	var tenantsInfo []TenantCutoffInfo
	for _, tenant := range tenants {
		// Get tenant settings
		settings, err := d.databasePort.TenantSetting().FindByTenantID(tenant.ID)
		if err != nil {
			// Skip if no settings found
			continue
		}

		// Check if cutoff_day matches today
		if settings.CutoffDay == today {
			tenantsInfo = append(tenantsInfo, TenantCutoffInfo{
				TenantID:    tenant.ID,
				TenantName:  tenant.Name,
				CutoffDay:   settings.CutoffDay,
				GracePeriod: settings.GracePeriod,
			})
		}
	}

	return tenantsInfo, nil
}

// processTenantCutoff processes cutoff for a specific tenant
func (d *cutoffDomain) processTenantCutoff(ctx context.Context, tenantInfo TenantCutoffInfo) error {
	// Find customers for this tenant with auto_cutoff enabled
	autoCutoff := true
	customers, err := d.databasePort.Customer().FindByFilter(model.CustomerFilter{
		TenantIDs:  []string{tenantInfo.TenantID},
		IsActive:   &autoCutoff,
		AutoCutoff: &autoCutoff,
	})
	if err != nil {
		return stacktrace.Propagate(err, "failed to fetch customers for tenant %s", tenantInfo.TenantID)
	}

	// For each customer, check if they have overdue invoices
	for _, customer := range customers {
		shouldIsolate, err := d.shouldIsolateCustomer(ctx, customer.ID, tenantInfo.GracePeriod)
		if err != nil {
			// Log error but continue
			continue
		}

		if shouldIsolate {
			err = d.IsolateCustomer(ctx, customer.ID)
			if err != nil {
				// Log error but continue with other customers
				continue
			}
		}
	}

	return nil
}

// shouldIsolateCustomer checks if a customer has unpaid invoices past due date + grace period
func (d *cutoffDomain) shouldIsolateCustomer(ctx context.Context, customerID string, gracePeriod int) (bool, error) {
	// Get unpaid invoices for this customer
	invoices, err := d.databasePort.Invoice().FindByFilter(model.InvoiceFilter{
		CustomerIDs: []string{customerID},
		Statuses:    []string{model.InvoiceStatusUnpaid, model.InvoiceStatusOverdue},
	})
	if err != nil {
		return false, stacktrace.Propagate(err, "failed to fetch invoices for customer %s", customerID)
	}

	// Check if any invoice is past due date + grace period
	now := time.Now()
	for _, invoice := range invoices {
		graceDueDate := invoice.DueDate.AddDate(0, 0, gracePeriod)
		if now.After(graceDueDate) {
			return true, nil
		}
	}

	return false, nil
}

// IsolateCustomer isolates all PPPoE accounts for a customer
func (d *cutoffDomain) IsolateCustomer(ctx context.Context, customerID string) error {
	if customerID == "" {
		return stacktrace.NewError("customer_id is empty")
	}

	// Get all active PPPoE accounts for this customer
	accounts, err := d.databasePort.PppoeAccount().FindByFilter(model.PppoeAccountFilter{
		CustomerIDs: []string{customerID},
		Statuses:    []string{model.PppoeStatusActive},
	})
	if err != nil {
		return stacktrace.Propagate(err, "failed to fetch pppoe accounts for customer %s", customerID)
	}

	if len(accounts) == 0 {
		// No active accounts to isolate
		return nil
	}

	// Isolate each account
	for _, account := range accounts {
		err := d.isolatePppoeAccount(ctx, account)
		if err != nil {
			// Log error but continue with other accounts
			continue
		}
	}

	// Log activity
	detailsJSON, _ := json.Marshal(map[string]interface{}{
		"customer_id": customerID,
		"accounts_affected": len(accounts),
	})
	activityInput := model.ActivityLogInput{
		TenantID:     accounts[0].TenantID,
		ResourceType: "customer",
		ResourceID:   customerID,
		Action:       "isolate",
		Details:      detailsJSON,
		CreatedAt:    time.Now(),
	}
	_, _ = d.databasePort.ActivityLog().Create(activityInput)

	return nil
}

// isolatePppoeAccount isolates a single PPPoE account
func (d *cutoffDomain) isolatePppoeAccount(ctx context.Context, account model.PppoeAccount) error {
	// Get tenant settings for isolir profile name
	isolirProfile := "ISOLIR_MIKROPS" // default
	tenantSetting, err := d.databasePort.TenantSetting().FindByTenantID(account.TenantID)
	if err == nil && tenantSetting.IsolirProfileName != "" {
		isolirProfile = tenantSetting.IsolirProfileName
	}

	// Update PPPoE account status
	now := time.Now()
	updateInput := model.PppoeAccountInput{
		OriginalProfile: account.ProfileName,
		ProfileName:     isolirProfile,
		Status:          model.PppoeStatusIsolated,
		SyncStatus:      model.PppoeSyncStatusPending,
		LastSyncAt:      &now,
	}

	err = d.databasePort.PppoeAccount().Update(account.ID, updateInput)
	if err != nil {
		return stacktrace.Propagate(err, "failed to update pppoe account %s", account.ID)
	}

	// Also update subscription status to suspended
	if account.SubscriptionID != nil && *account.SubscriptionID != "" {
		subscriptionInput := model.SubscriptionInput{
			Status: model.SubscriptionStatusSuspended,
		}
		_ = d.databasePort.Subscription().Update(*account.SubscriptionID, subscriptionInput)
	}

	// TODO: Enqueue RabbitMQ message for MikroTik sync
	// message := map[string]interface{}{
	// 	"action":          "isolate",
	// 	"pppoe_account_id": account.ID,
	// 	"nas_id":          account.NasID,
	// 	"username":        account.Username,
	// 	"isolir_profile":  isolirProfile,
	// }
	// _ = d.messagePort.Publish(ctx, "mikrotik.ppp.isolate", message)

	// Log to mikrotik_sync_logs
	requestJSON, _ := json.Marshal(map[string]string{
		"username": account.Username,
		"profile":  isolirProfile,
	})
	syncLogInput := model.MikrotikSyncLogInput{
		TenantID:           account.TenantID,
		NasID:              account.NasID,
		Action:             "isolate",
		ResourceType:       "pppoe_account",
		ResourceIdentifier: account.Username,
		Status:             "pending",
		RequestPayload:     requestJSON,
		CreatedAt:          time.Now(),
	}
	_, _ = d.databasePort.MikrotikSyncLog().Create(syncLogInput)

	return nil
}

// RestoreCustomer restores all isolated PPPoE accounts for a customer
func (d *cutoffDomain) RestoreCustomer(ctx context.Context, customerID string) error {
	if customerID == "" {
		return stacktrace.NewError("customer_id is empty")
	}

	// Get all isolated PPPoE accounts for this customer
	accounts, err := d.databasePort.PppoeAccount().FindByFilter(model.PppoeAccountFilter{
		CustomerIDs: []string{customerID},
		Statuses:    []string{model.PppoeStatusIsolated},
	})
	if err != nil {
		return stacktrace.Propagate(err, "failed to fetch isolated pppoe accounts for customer %s", customerID)
	}

	if len(accounts) == 0 {
		// No isolated accounts to restore
		return nil
	}

	// Restore each account
	for _, account := range accounts {
		err := d.restorePppoeAccount(ctx, account)
		if err != nil {
			// Log error but continue with other accounts
			continue
		}
	}

	// Log activity
	detailsJSON, _ := json.Marshal(map[string]interface{}{
		"customer_id": customerID,
		"accounts_affected": len(accounts),
	})
	activityInput := model.ActivityLogInput{
		TenantID:     accounts[0].TenantID,
		ResourceType: "customer",
		ResourceID:   customerID,
		Action:       "restore",
		Details:      detailsJSON,
		CreatedAt:    time.Now(),
	}
	_, _ = d.databasePort.ActivityLog().Create(activityInput)

	return nil
}

// restorePppoeAccount restores a single PPPoE account
func (d *cutoffDomain) restorePppoeAccount(ctx context.Context, account model.PppoeAccount) error {
	// Determine profile to restore
	profileToRestore := account.OriginalProfile
	if profileToRestore == "" {
		// Fallback: get profile from package
		if account.PackageID != "" {
			pkg, err := d.databasePort.InternetPackage().FindByID(account.PackageID)
			if err == nil && pkg.ProfileName != "" {
				profileToRestore = pkg.ProfileName
			}
		}
	}

	// Update PPPoE account status
	now := time.Now()
	updateInput := model.PppoeAccountInput{
		ProfileName:     profileToRestore,
		OriginalProfile: "",
		Status:          model.PppoeStatusActive,
		SyncStatus:      model.PppoeSyncStatusPending,
		LastSyncAt:      &now,
	}

	err := d.databasePort.PppoeAccount().Update(account.ID, updateInput)
	if err != nil {
		return stacktrace.Propagate(err, "failed to update pppoe account %s", account.ID)
	}

	// Also update subscription status back to active
	if account.SubscriptionID != nil && *account.SubscriptionID != "" {
		subscriptionInput := model.SubscriptionInput{
			Status: model.SubscriptionStatusActive,
		}
		_ = d.databasePort.Subscription().Update(*account.SubscriptionID, subscriptionInput)
	}

	// TODO: Enqueue RabbitMQ message for MikroTik sync
	// message := map[string]interface{}{
	// 	"action":           "restore",
	// 	"pppoe_account_id": account.ID,
	// 	"nas_id":           account.NasID,
	// 	"username":         account.Username,
	// 	"restore_profile":  profileToRestore,
	// }
	// _ = d.messagePort.Publish(ctx, "mikrotik.ppp.restore", message)

	// Log to mikrotik_sync_logs
	requestJSON, _ := json.Marshal(map[string]string{
		"username": account.Username,
		"profile":  profileToRestore,
	})
	syncLogInput := model.MikrotikSyncLogInput{
		TenantID:           account.TenantID,
		NasID:              account.NasID,
		Action:             "restore",
		ResourceType:       "pppoe_account",
		ResourceIdentifier: account.Username,
		Status:             "pending",
		RequestPayload:     requestJSON,
		CreatedAt:          time.Now(),
	}
	_, _ = d.databasePort.MikrotikSyncLog().Create(syncLogInput)

	return nil
}

// OnInvoicePaid is triggered when an invoice is paid to check if customer needs restoration
func (d *cutoffDomain) OnInvoicePaid(ctx context.Context, invoiceID string) error {
	if invoiceID == "" {
		return stacktrace.NewError("invoice_id is empty")
	}

	// Get invoice to find customer
	invoice, err := d.databasePort.Invoice().FindByID(invoiceID)
	if err != nil {
		return stacktrace.Propagate(err, "failed to fetch invoice %s", invoiceID)
	}

	// Check if customer has isolated accounts
	accounts, err := d.databasePort.PppoeAccount().FindByFilter(model.PppoeAccountFilter{
		CustomerIDs: []string{invoice.CustomerID},
		Statuses:    []string{model.PppoeStatusIsolated},
	})
	if err != nil {
		return stacktrace.Propagate(err, "failed to fetch isolated accounts for customer %s", invoice.CustomerID)
	}

	if len(accounts) == 0 {
		// No isolated accounts, nothing to restore
		return nil
	}

	// Restore customer
	err = d.RestoreCustomer(ctx, invoice.CustomerID)
	if err != nil {
		return stacktrace.Propagate(err, "failed to restore customer %s", invoice.CustomerID)
	}

	return nil
}
