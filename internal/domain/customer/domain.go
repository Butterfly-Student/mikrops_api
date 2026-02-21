package customer

import (
	"context"

	"github.com/google/uuid"
	"github.com/palantir/stacktrace"

	"go-template/internal/model"
	outbound_port "go-template/internal/port/outbound"
)

type CustomerDomain interface {
	Create(ctx context.Context, input model.CustomerInput) (*model.Customer, error)
	GetByID(ctx context.Context, id string) (*model.Customer, error)
	GetByCode(ctx context.Context, code string) (*model.Customer, error)
	List(ctx context.Context, filter *model.CustomerFilter) ([]model.Customer, error)
	Update(ctx context.Context, id string, input model.CustomerInput) (*model.Customer, error)
	Delete(ctx context.Context, id string) error
	ChangeStatus(ctx context.Context, id string, status model.CustomerStatus) error
	Isolate(ctx context.Context, id string) error
	UnIsolate(ctx context.Context, id string) error
	SyncToMikrotik(ctx context.Context, id string) error
}

type domain struct {
	dbPort       outbound_port.DatabasePort
	mikrotikPort outbound_port.MikrotikPort
}

func NewCustomerDomain(
	dbPort outbound_port.DatabasePort,
	mikrotikPort outbound_port.MikrotikPort,
) CustomerDomain {
	return &domain{
		dbPort:       dbPort,
		mikrotikPort: mikrotikPort,
	}
}

func (d *domain) Create(ctx context.Context, input model.CustomerInput) (*model.Customer, error) {
	customer := input.ToModel()
	customer.ID = uuid.New()

	// MikroTik-first: if PPPoE fields are present, push to MikroTik BEFORE saving to DB
	if customer.RouterID != nil &&
		customer.PppSecretName != nil && *customer.PppSecretName != "" &&
		customer.PppSecretPassword != nil && *customer.PppSecretPassword != "" {

		if customer.Status == model.CustomerStatusActive || customer.Status == model.CustomerStatusPending {
			if err := d.createPppoeSecretInMikrotik(ctx, customer); err != nil {
				// MikroTik failed — do NOT save to DB
				return nil, stacktrace.Propagate(err, "failed to create ppp secret on mikrotik, customer not saved")
			}
		}
	}

	// MikroTik OK (or no MikroTik config) — save to DB
	if err := d.dbPort.Customer().Create(ctx, customer); err != nil {
		// DB failed — if we pushed to MikroTik, attempt rollback (best effort)
		if customer.RouterID != nil && customer.PppSecretName != nil {
			_ = d.deletePppoeSecretFromMikrotik(ctx, customer)
		}
		return nil, stacktrace.Propagate(err, "mikrotik ok but failed to save customer to database")
	}

	return customer, nil
}

func (d *domain) GetByID(ctx context.Context, id string) (*model.Customer, error) {
	customer, err := d.dbPort.Customer().FindByID(ctx, id)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to get customer by id")
	}

	return customer, nil
}

func (d *domain) GetByCode(ctx context.Context, code string) (*model.Customer, error) {
	customer, err := d.dbPort.Customer().FindByCode(ctx, code)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to get customer by code")
	}

	return customer, nil
}

func (d *domain) List(ctx context.Context, filter *model.CustomerFilter) ([]model.Customer, error) {
	customers, err := d.dbPort.Customer().FindAll(ctx, filter)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to list customers")
	}

	return customers, nil
}

func (d *domain) Update(ctx context.Context, id string, input model.CustomerInput) (*model.Customer, error) {
	// Get existing customer
	customer, err := d.dbPort.Customer().FindByID(ctx, id)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to find customer")
	}

	// Store old values for comparison and rollback
	oldPppSecretName := customer.PppSecretName
	oldPppSecretPassword := customer.PppSecretPassword
	oldProfileID := customer.ProfileID
	oldStatus := customer.Status
	oldRouterID := customer.RouterID

	// Apply new values
	customer.CustomerCode = input.CustomerCode
	customer.FullName = input.FullName
	customer.Email = input.Email
	customer.Phone = input.Phone
	customer.Address = input.Address
	customer.Latitude = input.Latitude
	customer.Longitude = input.Longitude
	customer.ActivationDate = input.ActivationDate
	customer.InstallationDate = input.InstallationDate
	customer.TerminationDate = input.TerminationDate
	customer.ExpiryDate = input.ExpiryDate
	customer.RouterID = input.RouterID
	customer.PppSecretName = input.PppSecretName
	customer.PppSecretPassword = input.PppSecretPassword
	customer.StaticIP = input.StaticIP
	customer.MacAddress = input.MacAddress
	customer.ProfileID = input.ProfileID
	customer.BillingDay = input.BillingDay
	customer.PaymentMethodPreference = input.PaymentMethodPreference
	customer.AutoIsolate = input.AutoIsolate
	customer.GracePeriodDays = input.GracePeriodDays
	customer.Notes = input.Notes
	customer.Tags = input.Tags

	if input.Status != nil {
		customer.Status = model.CustomerStatus(*input.Status)
	}
	if input.PppService != nil {
		customer.PppService = model.PPPService(*input.PppService)
	}
	if input.BillingCycle != nil {
		customer.BillingCycle = model.BillingCycle(*input.BillingCycle)
	}

	// MikroTik-first: sync if PPPoE-related fields changed and router is configured
	pppChanged := oldPppSecretName != customer.PppSecretName ||
		oldPppSecretPassword != customer.PppSecretPassword ||
		oldProfileID != customer.ProfileID ||
		oldStatus != customer.Status

	if pppChanged && customer.RouterID != nil &&
		customer.PppSecretName != nil && *customer.PppSecretName != "" {

		// If router changed remove old secret from old router (best effort)
		if oldRouterID != nil && oldRouterID != customer.RouterID && oldPppSecretName != nil {
			oldCustomer := *customer
			oldCustomer.RouterID = oldRouterID
			oldCustomer.PppSecretName = oldPppSecretName
			_ = d.deletePppoeSecretFromMikrotik(ctx, &oldCustomer)
		}

		// Push updated secret to new/current router
		if err := d.syncToMikrotik(ctx, customer); err != nil {
			return nil, stacktrace.Propagate(err, "failed to update ppp secret on mikrotik, customer not updated")
		}
	}

	// MikroTik OK — update DB
	if err := d.dbPort.Customer().Update(ctx, customer); err != nil {
		// Best-effort rollback: restore old values on MikroTik
		if pppChanged && customer.RouterID != nil {
			restoredCustomer := *customer
			restoredCustomer.PppSecretName = oldPppSecretName
			restoredCustomer.PppSecretPassword = oldPppSecretPassword
			restoredCustomer.ProfileID = oldProfileID
			restoredCustomer.Status = oldStatus
			_ = d.syncToMikrotik(ctx, &restoredCustomer)
		}
		return nil, stacktrace.Propagate(err, "mikrotik ok but failed to update customer in database")
	}

	return customer, nil
}

func (d *domain) Delete(ctx context.Context, id string) error {
	// Get customer to check router and PPPoE fields
	customer, err := d.dbPort.Customer().FindByID(ctx, id)
	if err != nil {
		return stacktrace.Propagate(err, "failed to find customer")
	}

	// MikroTik-first: delete PPPoE secret BEFORE removing from DB
	if customer.RouterID != nil && customer.PppSecretName != nil && *customer.PppSecretName != "" {
		if err := d.deletePppoeSecretFromMikrotik(ctx, customer); err != nil {
			return stacktrace.Propagate(err, "failed to delete ppp secret from mikrotik, customer not deleted")
		}
	}

	// MikroTik OK — soft delete in DB
	if err := d.dbPort.Customer().Delete(ctx, id); err != nil {
		// Best-effort rollback: recreate secret on MikroTik
		if customer.RouterID != nil && customer.PppSecretName != nil {
			_ = d.createPppoeSecretInMikrotik(ctx, customer)
		}
		return stacktrace.Propagate(err, "mikrotik ok but failed to delete customer from database")
	}

	return nil
}

func (d *domain) ChangeStatus(ctx context.Context, id string, status model.CustomerStatus) error {
	// Delegate isolation/un-isolation to specialized methods with profile-switching
	switch status {
	case model.CustomerStatusIsolated:
		return d.Isolate(ctx, id)
	case model.CustomerStatusActive:
		// Check if currently isolated — if so, use UnIsolate to restore profile
		customer, err := d.dbPort.Customer().FindByID(ctx, id)
		if err != nil {
			return stacktrace.Propagate(err, "failed to find customer")
		}
		if customer.Status == model.CustomerStatusIsolated {
			return d.UnIsolate(ctx, id)
		}
		// Otherwise, normal status change (e.g., pending → active)
		return d.changeStatusWithSync(ctx, id, customer, status)
	default:
		// For suspended, terminated, pending — use standard disable-based sync
		customer, err := d.dbPort.Customer().FindByID(ctx, id)
		if err != nil {
			return stacktrace.Propagate(err, "failed to find customer")
		}
		return d.changeStatusWithSync(ctx, id, customer, status)
	}
}

// changeStatusWithSync syncs disable state to MikroTik FIRST, then updates DB status.
func (d *domain) changeStatusWithSync(ctx context.Context, id string, customer *model.Customer, status model.CustomerStatus) error {
	oldStatus := customer.Status
	customer.Status = status

	// MikroTik-first: sync disabled/enabled state based on new status
	if customer.RouterID != nil && customer.PppSecretName != nil {
		if err := d.syncToMikrotik(ctx, customer); err != nil {
			return stacktrace.Propagate(err, "failed to sync status to mikrotik, status not changed")
		}
	}

	// MikroTik OK — update status in DB
	if err := d.dbPort.Customer().UpdateStatus(ctx, id, status); err != nil {
		// Best-effort rollback: restore old status on MikroTik
		if customer.RouterID != nil && customer.PppSecretName != nil {
			customer.Status = oldStatus
			_ = d.syncToMikrotik(ctx, customer)
		}
		return stacktrace.Propagate(err, "mikrotik synced but failed to update status in database")
	}

	return nil
}

func (d *domain) Isolate(ctx context.Context, id string) error {
	// Get customer with profile preloaded
	customer, err := d.dbPort.Customer().FindByID(ctx, id)
	if err != nil {
		return stacktrace.Propagate(err, "failed to find customer")
	}

	// Validate: must be active and have MikroTik config
	if customer.Status != model.CustomerStatusActive {
		return stacktrace.NewError("only active customers can be isolated, current status: %s", customer.Status)
	}
	if customer.RouterID == nil {
		return stacktrace.NewError("customer has no router assigned")
	}
	if customer.PppSecretName == nil || *customer.PppSecretName == "" {
		return stacktrace.NewError("customer has no ppp secret name")
	}

	// Get router
	router, err := d.dbPort.Mikrotik().FindByID(customer.RouterID.String())
	if err != nil {
		return stacktrace.Propagate(err, "failed to get mikrotik router")
	}

	// Ensure isolation infrastructure exists on router (idempotent)
	isolationSetup, _ := d.mikrotikPort.CheckIsolationSetup(router)
	if !isolationSetup {
		config := model.DefaultIsolationConfig()
		if err := d.mikrotikPort.SetupIsolation(router, config); err != nil {
			return stacktrace.Propagate(err, "failed to setup isolation on router")
		}
	}

	// MikroTik-first: switch PPP secret profile to "isolir"
	isolationProfileName := model.DefaultIsolationConfig().ProfileName
	secret := &model.PppoeSecret{
		Name:    *customer.PppSecretName,
		Profile: isolationProfileName,
	}
	existingSecret, _ := d.findSecretByName(router, *customer.PppSecretName)
	if existingSecret != nil {
		secret.ID = existingSecret.ID
	}
	if err := d.mikrotikPort.UpdateSecret(router, secret); err != nil {
		return stacktrace.Propagate(err, "failed to switch ppp secret to isolation profile")
	}

	// Kick active session so user reconnects with isolation profile
	_ = d.mikrotikPort.RemoveActiveSession(router, *customer.PppSecretName)

	// MikroTik OK — update DB: save previous profile and change status
	customer.PreviousProfileID = customer.ProfileID
	customer.Status = model.CustomerStatusIsolated
	if err := d.dbPort.Customer().Update(ctx, customer); err != nil {
		// Best-effort rollback: restore original profile on MikroTik
		if customer.Profile != nil {
			rollbackSecret := &model.PppoeSecret{
				Name:    *customer.PppSecretName,
				Profile: customer.Profile.PppProfileName,
			}
			if existingSecret != nil {
				rollbackSecret.ID = existingSecret.ID
			}
			_ = d.mikrotikPort.UpdateSecret(router, rollbackSecret)
		}
		return stacktrace.Propagate(err, "mikrotik isolated but failed to update database")
	}

	return nil
}

func (d *domain) UnIsolate(ctx context.Context, id string) error {
	// Get customer
	customer, err := d.dbPort.Customer().FindByID(ctx, id)
	if err != nil {
		return stacktrace.Propagate(err, "failed to find customer")
	}

	// Validate: must be isolated
	if customer.Status != model.CustomerStatusIsolated {
		return stacktrace.NewError("only isolated customers can be un-isolated, current status: %s", customer.Status)
	}
	if customer.RouterID == nil {
		return stacktrace.NewError("customer has no router assigned")
	}
	if customer.PppSecretName == nil || *customer.PppSecretName == "" {
		return stacktrace.NewError("customer has no ppp secret name")
	}
	if customer.PreviousProfileID == nil {
		return stacktrace.NewError("customer has no previous profile to restore")
	}

	// Get the previous bandwidth profile to restore
	previousProfile, err := d.dbPort.BandwidthProfile().FindByID(ctx, customer.PreviousProfileID.String())
	if err != nil {
		return stacktrace.Propagate(err, "failed to get previous bandwidth profile")
	}

	// Get router
	router, err := d.dbPort.Mikrotik().FindByID(customer.RouterID.String())
	if err != nil {
		return stacktrace.Propagate(err, "failed to get mikrotik router")
	}

	// MikroTik-first: restore PPP secret profile to original
	secret := &model.PppoeSecret{
		Name:    *customer.PppSecretName,
		Profile: previousProfile.PppProfileName,
	}
	existingSecret, _ := d.findSecretByName(router, *customer.PppSecretName)
	if existingSecret != nil {
		secret.ID = existingSecret.ID
	}
	if err := d.mikrotikPort.UpdateSecret(router, secret); err != nil {
		return stacktrace.Propagate(err, "failed to restore ppp secret profile")
	}

	// Kick active session so user reconnects with restored profile
	_ = d.mikrotikPort.RemoveActiveSession(router, *customer.PppSecretName)

	// MikroTik OK — update DB: restore profile and clear previous
	customer.ProfileID = customer.PreviousProfileID
	customer.PreviousProfileID = nil
	customer.Status = model.CustomerStatusActive
	if err := d.dbPort.Customer().Update(ctx, customer); err != nil {
		// Best-effort rollback: switch back to isolation profile
		isolationProfileName := model.DefaultIsolationConfig().ProfileName
		rollbackSecret := &model.PppoeSecret{
			Name:    *customer.PppSecretName,
			Profile: isolationProfileName,
		}
		if existingSecret != nil {
			rollbackSecret.ID = existingSecret.ID
		}
		_ = d.mikrotikPort.UpdateSecret(router, rollbackSecret)
		return stacktrace.Propagate(err, "mikrotik restored but failed to update database")
	}

	return nil
}

func (d *domain) SyncToMikrotik(ctx context.Context, id string) error {
	// Get customer with profile and router
	customer, err := d.dbPort.Customer().FindByID(ctx, id)
	if err != nil {
		return stacktrace.Propagate(err, "failed to get customer")
	}

	return d.syncToMikrotik(ctx, customer)
}

// Helper methods

func (d *domain) syncToMikrotik(ctx context.Context, customer *model.Customer) error {
	// Validate required fields
	if customer.RouterID == nil {
		return stacktrace.NewError("customer has no router assigned")
	}
	if customer.PppSecretName == nil || *customer.PppSecretName == "" {
		return stacktrace.NewError("customer has no ppp secret name")
	}
	if customer.PppSecretPassword == nil || *customer.PppSecretPassword == "" {
		return stacktrace.NewError("customer has no ppp secret password")
	}

	// Get router
	router, err := d.dbPort.Mikrotik().FindByID(customer.RouterID.String())
	if err != nil {
		return stacktrace.Propagate(err, "failed to get mikrotik router")
	}

	// Build PPPoE secret
	pppSecret := d.convertToPppoeSecret(customer)

	// Check if secret exists
	existingSecret, err := d.findSecretByName(router, *customer.PppSecretName)
	if err != nil || existingSecret == nil {
		// Secret doesn't exist, create it
		if err := d.mikrotikPort.CreateSecret(router, pppSecret); err != nil {
			return stacktrace.Propagate(err, "failed to create secret on mikrotik")
		}
	} else {
		// Secret exists, update it
		pppSecret.ID = existingSecret.ID
		if err := d.mikrotikPort.UpdateSecret(router, pppSecret); err != nil {
			return stacktrace.Propagate(err, "failed to update secret on mikrotik")
		}
	}

	return nil
}

func (d *domain) createPppoeSecretInMikrotik(ctx context.Context, customer *model.Customer) error {
	// Validate required fields
	if customer.RouterID == nil {
		return nil // Skip if no router assigned
	}
	if customer.PppSecretName == nil || *customer.PppSecretName == "" {
		return nil // Skip if no secret name
	}
	if customer.PppSecretPassword == nil || *customer.PppSecretPassword == "" {
		return nil // Skip if no password
	}

	// Get router
	router, err := d.dbPort.Mikrotik().FindByID(customer.RouterID.String())
	if err != nil {
		return stacktrace.Propagate(err, "failed to get mikrotik router")
	}

	// Build PPPoE secret
	pppSecret := d.convertToPppoeSecret(customer)

	// Create secret
	if err := d.mikrotikPort.CreateSecret(router, pppSecret); err != nil {
		return stacktrace.Propagate(err, "failed to create secret on mikrotik")
	}

	return nil
}

func (d *domain) deletePppoeSecretFromMikrotik(ctx context.Context, customer *model.Customer) error {
	// Get router
	router, err := d.dbPort.Mikrotik().FindByID(customer.RouterID.String())
	if err != nil {
		return stacktrace.Propagate(err, "failed to get mikrotik router")
	}

	// Find secret by name
	existingSecret, err := d.findSecretByName(router, *customer.PppSecretName)
	if err != nil || existingSecret == nil {
		// Secret doesn't exist, nothing to delete
		return nil
	}

	// Delete secret
	if err := d.mikrotikPort.DeleteSecret(router, existingSecret.ID); err != nil {
		return stacktrace.Propagate(err, "failed to delete secret from mikrotik")
	}

	return nil
}

func (d *domain) findSecretByName(router *model.MikrotikRouter, name string) (*model.PppoeSecret, error) {
	// List all secrets
	secrets, err := d.mikrotikPort.ListSecrets(router)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to list secrets from mikrotik")
	}

	// Find by name
	for _, secret := range secrets {
		if secret.Name == name {
			return &secret, nil
		}
	}

	return nil, stacktrace.NewError("secret not found")
}

func (d *domain) convertToPppoeSecret(customer *model.Customer) *model.PppoeSecret {
	// Isolated and active users stay connected (not disabled)
	// Suspended and terminated users are disabled
	disabled := customer.Status == model.CustomerStatusSuspended ||
		customer.Status == model.CustomerStatusTerminated

	secret := &model.PppoeSecret{
		Name:     *customer.PppSecretName,
		Password: *customer.PppSecretPassword,
		Service:  string(customer.PppService),
		Disabled: disabled,
	}

	// Set caller ID (MAC address)
	if customer.MacAddress != nil && *customer.MacAddress != "" {
		secret.CallerID = *customer.MacAddress
	}

	// For isolated customers, use the isolation profile instead of the bandwidth profile
	if customer.Status == model.CustomerStatusIsolated {
		secret.Profile = model.DefaultIsolationConfig().ProfileName
	} else if customer.Profile != nil {
		secret.Profile = customer.Profile.PppProfileName
	}

	// Set static IP
	if customer.StaticIP != nil && *customer.StaticIP != "" {
		secret.RemoteAddress = *customer.StaticIP
	}

	// Set comment
	comment := customer.FullName + " - " + customer.CustomerCode
	secret.Comment = comment

	return secret
}
