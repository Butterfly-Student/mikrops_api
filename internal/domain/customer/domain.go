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
	// Get customer
	customer, err := d.dbPort.Customer().FindByID(ctx, id)
	if err != nil {
		return stacktrace.Propagate(err, "failed to find customer")
	}

	// Update status
	if err := d.dbPort.Customer().UpdateStatus(ctx, id, status); err != nil {
		return stacktrace.Propagate(err, "failed to update customer status")
	}

	// Update customer object for sync
	customer.Status = status

	// Sync to MikroTik based on status
	if customer.RouterID != nil && customer.PppSecretName != nil {
		switch status {
		case model.CustomerStatusIsolated, model.CustomerStatusSuspended:
			// Disable PPPoE secret
			if err := d.syncToMikrotik(ctx, customer); err != nil {
				return stacktrace.Propagate(err, "status updated but failed to disable in mikrotik")
			}
		case model.CustomerStatusActive:
			// Enable PPPoE secret
			if err := d.syncToMikrotik(ctx, customer); err != nil {
				return stacktrace.Propagate(err, "status updated but failed to enable in mikrotik")
			}
		case model.CustomerStatusTerminated:
			// Disable and optionally delete
			if err := d.syncToMikrotik(ctx, customer); err != nil {
				return stacktrace.Propagate(err, "status updated but failed to disable in mikrotik")
			}
		}
	}

	return nil
}

func (d *domain) Isolate(ctx context.Context, id string) error {
	return d.ChangeStatus(ctx, id, model.CustomerStatusIsolated)
}

func (d *domain) UnIsolate(ctx context.Context, id string) error {
	return d.ChangeStatus(ctx, id, model.CustomerStatusActive)
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
	secret := &model.PppoeSecret{
		Name:     *customer.PppSecretName,
		Password: *customer.PppSecretPassword,
		Service:  string(customer.PppService),
		Disabled: customer.Status != model.CustomerStatusActive,
	}

	// Set caller ID (MAC address)
	if customer.MacAddress != nil && *customer.MacAddress != "" {
		secret.CallerID = *customer.MacAddress
	}

	// Set profile from BandwidthProfile
	if customer.Profile != nil {
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
