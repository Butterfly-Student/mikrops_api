package customer

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/palantir/stacktrace"

	"go-template/internal/model"
	outbound_port "go-template/internal/port/outbound"
)

type CustomerDomain interface {
	Create(ctx context.Context, input model.CustomerInput) (*model.Customer, error)
	FindByID(ctx context.Context, id string) (*model.Customer, error)
	FindByFilter(ctx context.Context, filter model.CustomerFilter) ([]model.Customer, error)
	Update(ctx context.Context, id string, input model.CustomerInput) (*model.Customer, error)
	Delete(ctx context.Context, id string) error
	Isolate(ctx context.Context, customerID string) error
	Activate(ctx context.Context, customerID string) error
	GetBillingInfo(ctx context.Context, customerID string) (*BillingInfo, error)
	FindExpiredCustomers(ctx context.Context) ([]model.Customer, error)
}

type BillingInfo struct {
	Customer       *model.Customer         `json:"customer"`
	Profile        *model.BandwidthProfile `json:"profile"`
	MonthlyCharge  string                  `json:"monthly_charge"`
	NextBillingDate *time.Time             `json:"next_billing_date"`
	IsExpired      bool                    `json:"is_expired"`
	DaysUntilExpiry int                    `json:"days_until_expiry"`
}

type customerDomain struct {
	databasePort outbound_port.DatabasePort
	mikrotikPort outbound_port.MikrotikPort
}

func NewCustomerDomain(
	databasePort outbound_port.DatabasePort,
	mikrotikPort outbound_port.MikrotikPort,
) CustomerDomain {
	return &customerDomain{
		databasePort: databasePort,
		mikrotikPort: mikrotikPort,
	}
}

func (d *customerDomain) Create(ctx context.Context, input model.CustomerInput) (*model.Customer, error) {
	// Validate input
	if input.CustomerCode == "" {
		return nil, stacktrace.NewError("customer code is required")
	}
	if input.FullName == "" {
		return nil, stacktrace.NewError("customer name is required")
	}
	if input.ProfileID == nil {
		return nil, stacktrace.NewError("profile is required")
	}

	// Check if customer code already exists
	databaseCustomerPort := d.databasePort.Customer()
	existing, err := databaseCustomerPort.FindByFilter(model.CustomerFilter{
		CustomerCodes: []string{input.CustomerCode},
	})
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to check existing customer")
	}
	if len(existing) > 0 {
		return nil, stacktrace.NewError("customer code already exists")
	}

	// Check if PPP secret name already exists
	if input.PppSecretName != "" {
		existing, err := databaseCustomerPort.FindByFilter(model.CustomerFilter{
			Search: input.PppSecretName,
		})
		if err != nil {
			return nil, stacktrace.Propagate(err, "failed to check existing ppp secret")
		}
		for _, c := range existing {
			if c.PppSecretName == input.PppSecretName {
				return nil, stacktrace.NewError("ppp secret name already exists")
			}
		}
	}

	// Get profile
	databaseBandwidthProfilePort := d.databasePort.BandwidthProfile()
	profiles, err := databaseBandwidthProfilePort.FindByFilter(model.BandwidthProfileFilter{
		IDs: []uuid.UUID{*input.ProfileID},
	})
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to find bandwidth profile")
	}
	if len(profiles) == 0 {
		return nil, stacktrace.NewError("bandwidth profile not found")
	}

	// Create customer
	now := time.Now()
	customer := &model.Customer{
		ID:                      uuid.New(),
		CustomerCode:            input.CustomerCode,
		FullName:                input.FullName,
		Email:                   input.Email,
		Phone:                   input.Phone,
		Address:                 input.Address,
		Coordinates:             input.Coordinates,
		Status:                  model.CustomerStatusPending,
		RouterID:                input.RouterID,
		PppSecretName:           input.PppSecretName,
		PppSecretPassword:       input.PppSecretPassword,
		PppService:              input.PppService,
		StaticIP:                input.StaticIP,
		MacAddress:              input.MacAddress,
		ProfileID:               input.ProfileID,
		BillingCycle:            input.BillingCycle,
		BillingDay:              input.BillingDay,
		PaymentMethodPreference: input.PaymentMethodPreference,
		AutoIsolate:             input.AutoIsolate,
		GracePeriodDays:         input.GracePeriodDays,
		Notes:                   input.Notes,
		Tags:                    model.JSONBArray(input.Tags),
		CreatedAt:               now,
		UpdatedAt:               now,
	}

	// Set status and dates if provided
	if input.Status != "" {
		customer.Status = input.Status
	}
	if input.ActivationDate != nil {
		customer.ActivationDate = input.ActivationDate
	}
	if input.InstallationDate != nil {
		customer.InstallationDate = input.InstallationDate
	}
	if input.ExpiryDate != nil {
		customer.ExpiryDate = input.ExpiryDate
	}

	// Set default billing day if not provided
	if customer.BillingDay == 0 {
		customer.BillingDay = 1
	}

	// Set default grace period if not provided
	if customer.GracePeriodDays == 0 {
		customer.GracePeriodDays = 3
	}

	// Set default billing cycle if not provided
	if customer.BillingCycle == "" {
		customer.BillingCycle = model.BillingCycleMonthly
	}

	// Set default PPP service if not provided
	if customer.PppService == "" {
		customer.PppService = model.PPPServiceTypePPPoE
	}

	err = databaseCustomerPort.Create(customer)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to create customer")
	}

	// If customer is active and has router, create PPP secret on MikroTik
	if customer.Status == model.CustomerStatusActive && customer.RouterID != nil {
		err = d.syncToMikrotik(ctx, customer, &profiles[0])
		if err != nil {
			// Log error but don't fail the creation
			// The sync can be done manually later
		}
	}

	return customer, nil
}

func (d *customerDomain) FindByID(ctx context.Context, id string) (*model.Customer, error) {
	if id == "" {
		return nil, stacktrace.NewError("id is required")
	}

	customerID, err := uuid.Parse(id)
	if err != nil {
		return nil, stacktrace.Propagate(err, "invalid uuid format")
	}

	databaseCustomerPort := d.databasePort.Customer()
	customers, err := databaseCustomerPort.FindByFilter(model.CustomerFilter{
		IDs: []uuid.UUID{customerID},
	})
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to find customer")
	}

	if len(customers) == 0 {
		return nil, stacktrace.NewError("customer not found")
	}

	return &customers[0], nil
}

func (d *customerDomain) FindByFilter(ctx context.Context, filter model.CustomerFilter) ([]model.Customer, error) {
	databaseCustomerPort := d.databasePort.Customer()
	customers, err := databaseCustomerPort.FindByFilter(filter)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to find customers")
	}

	return customers, nil
}

func (d *customerDomain) Update(ctx context.Context, id string, input model.CustomerInput) (*model.Customer, error) {
	// Find existing customer
	customer, err := d.FindByID(ctx, id)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to find customer")
	}

	// Check if customer code is being changed and already exists
	if input.CustomerCode != customer.CustomerCode {
		databaseCustomerPort := d.databasePort.Customer()
		existing, err := databaseCustomerPort.FindByFilter(model.CustomerFilter{
			CustomerCodes: []string{input.CustomerCode},
		})
		if err != nil {
			return nil, stacktrace.Propagate(err, "failed to check existing customer")
		}
		if len(existing) > 0 {
			return nil, stacktrace.NewError("customer code already exists")
		}
	}

	// Update customer fields
	customer.CustomerCode = input.CustomerCode
	customer.FullName = input.FullName
	customer.Email = input.Email
	customer.Phone = input.Phone
	customer.Address = input.Address
	customer.Coordinates = input.Coordinates
	customer.RouterID = input.RouterID
	customer.PppSecretName = input.PppSecretName
	customer.PppSecretPassword = input.PppSecretPassword
	customer.PppService = input.PppService
	customer.StaticIP = input.StaticIP
	customer.MacAddress = input.MacAddress
	customer.ProfileID = input.ProfileID
	customer.BillingCycle = input.BillingCycle
	customer.BillingDay = input.BillingDay
	customer.PaymentMethodPreference = input.PaymentMethodPreference
	customer.AutoIsolate = input.AutoIsolate
	customer.GracePeriodDays = input.GracePeriodDays
	customer.Notes = input.Notes
	customer.Tags = model.JSONBArray(input.Tags)
	customer.UpdatedAt = time.Now()

	if input.Status != "" {
		customer.Status = input.Status
	}
	if input.ActivationDate != nil {
		customer.ActivationDate = input.ActivationDate
	}
	if input.InstallationDate != nil {
		customer.InstallationDate = input.InstallationDate
	}
	if input.TerminationDate != nil {
		customer.TerminationDate = input.TerminationDate
	}
	if input.ExpiryDate != nil {
		customer.ExpiryDate = input.ExpiryDate
	}

	databaseCustomerPort := d.databasePort.Customer()
	err = databaseCustomerPort.Update(customer)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to update customer")
	}

	return customer, nil
}

func (d *customerDomain) Delete(ctx context.Context, id string) error {
	// Find existing customer
	customer, err := d.FindByID(ctx, id)
	if err != nil {
		return stacktrace.Propagate(err, "failed to find customer")
	}

	// Delete PPP secret from MikroTik if customer is active
	if customer.Status == model.CustomerStatusActive && customer.RouterID != nil {
		err = d.deleteFromMikrotik(ctx, customer)
		if err != nil {
			// Log error but don't fail the deletion
		}
	}

	databaseCustomerPort := d.databasePort.Customer()
	err = databaseCustomerPort.Delete(customer.ID.String())
	if err != nil {
		return stacktrace.Propagate(err, "failed to delete customer")
	}

	return nil
}

func (d *customerDomain) Isolate(ctx context.Context, customerID string) error {
	// Find customer
	customer, err := d.FindByID(ctx, customerID)
	if err != nil {
		return stacktrace.Propagate(err, "failed to find customer")
	}

	if customer.Status == model.CustomerStatusIsolated {
		return stacktrace.NewError("customer is already isolated")
	}

	// Get isolated profile
	databaseBandwidthProfilePort := d.databasePort.BandwidthProfile()
	isolatedProfiles, err := databaseBandwidthProfilePort.FindByFilter(model.BandwidthProfileFilter{
		Categories: []model.BandwidthProfileCategory{model.BandwidthProfileCategoryIsolated},
	})
	if err != nil {
		return stacktrace.Propagate(err, "failed to find isolated profile")
	}
	if len(isolatedProfiles) == 0 {
		return stacktrace.NewError("isolated profile not found")
	}

	// Update customer status to isolated
	customer.Status = model.CustomerStatusIsolated
	customer.UpdatedAt = time.Now()

	// Sync to MikroTik with isolated profile
	if customer.RouterID != nil {
		err = d.syncToMikrotik(ctx, customer, &isolatedProfiles[0])
		if err != nil {
			return stacktrace.Propagate(err, "failed to sync isolated profile to mikrotik")
		}
	}

	databaseCustomerPort := d.databasePort.Customer()
	err = databaseCustomerPort.Update(customer)
	if err != nil {
		return stacktrace.Propagate(err, "failed to update customer status")
	}

	return nil
}

func (d *customerDomain) Activate(ctx context.Context, customerID string) error {
	// Find customer
	customer, err := d.FindByID(ctx, customerID)
	if err != nil {
		return stacktrace.Propagate(err, "failed to find customer")
	}

	if customer.Status == model.CustomerStatusActive {
		return stacktrace.NewError("customer is already active")
	}

	// Get customer's original profile
	if customer.ProfileID == nil {
		return stacktrace.NewError("customer has no profile assigned")
	}

	databaseBandwidthProfilePort := d.databasePort.BandwidthProfile()
	profiles, err := databaseBandwidthProfilePort.FindByFilter(model.BandwidthProfileFilter{
		IDs: []uuid.UUID{*customer.ProfileID},
	})
	if err != nil {
		return stacktrace.Propagate(err, "failed to find customer profile")
	}
	if len(profiles) == 0 {
		return stacktrace.NewError("customer profile not found")
	}

	// Update customer status to active
	customer.Status = model.CustomerStatusActive
	now := time.Now()
	customer.ActivationDate = &now
	customer.UpdatedAt = now

	// Calculate next expiry date based on billing cycle
	nextExpiry := calculateNextExpiryDate(now, customer.BillingCycle)
	customer.ExpiryDate = &nextExpiry

	// Sync to MikroTik with original profile
	if customer.RouterID != nil {
		err = d.syncToMikrotik(ctx, customer, &profiles[0])
		if err != nil {
			return stacktrace.Propagate(err, "failed to sync profile to mikrotik")
		}
	}

	databaseCustomerPort := d.databasePort.Customer()
	err = databaseCustomerPort.Update(customer)
	if err != nil {
		return stacktrace.Propagate(err, "failed to update customer status")
	}

	return nil
}

func (d *customerDomain) GetBillingInfo(ctx context.Context, customerID string) (*BillingInfo, error) {
	// Find customer
	customer, err := d.FindByID(ctx, customerID)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to find customer")
	}

	// Get profile
	var profile *model.BandwidthProfile
	if customer.ProfileID != nil {
		databaseBandwidthProfilePort := d.databasePort.BandwidthProfile()
		profiles, err := databaseBandwidthProfilePort.FindByFilter(model.BandwidthProfileFilter{
			IDs: []uuid.UUID{*customer.ProfileID},
		})
		if err != nil {
			return nil, stacktrace.Propagate(err, "failed to find profile")
		}
		if len(profiles) > 0 {
			profile = &profiles[0]
		}
	}

	// Calculate days until expiry
	daysUntilExpiry := 0
	if customer.ExpiryDate != nil {
		daysUntilExpiry = int(time.Until(*customer.ExpiryDate).Hours() / 24)
	}

	// Calculate next billing date
	var nextBillingDate *time.Time
	if customer.ExpiryDate != nil {
		nextBilling := calculateNextExpiryDate(*customer.ExpiryDate, customer.BillingCycle)
		nextBillingDate = &nextBilling
	}

	billingInfo := &BillingInfo{
		Customer:        customer,
		Profile:         profile,
		MonthlyCharge:   "",
		NextBillingDate: nextBillingDate,
		IsExpired:       customer.IsExpired(),
		DaysUntilExpiry: daysUntilExpiry,
	}

	if profile != nil {
		billingInfo.MonthlyCharge = profile.PriceMonthly.String()
	}

	return billingInfo, nil
}

func (d *customerDomain) FindExpiredCustomers(ctx context.Context) ([]model.Customer, error) {
	databaseCustomerPort := d.databasePort.Customer()
	customers, err := databaseCustomerPort.FindExpiredCustomers()
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to find expired customers")
	}

	return customers, nil
}

// Helper functions

func (d *customerDomain) syncToMikrotik(ctx context.Context, customer *model.Customer, profile *model.BandwidthProfile) error {
	if customer.RouterID == nil {
		return nil
	}

	// Get router
	databaseMikrotikPort := d.databasePort.Mikrotik()
	router, err := databaseMikrotikPort.FindByID(customer.RouterID.String())
	if err != nil {
		return stacktrace.Propagate(err, "failed to find mikrotik router")
	}

	// Create or update PPP secret
	pppSecret := &model.PppoeSecret{
		Name:     customer.PppSecretName,
		Password: customer.PppSecretPassword,
		Service:  string(customer.PppService),
		Profile:  profile.PppProfileName,
		Comment:  customer.FullName + " (" + customer.CustomerCode + ")",
	}

	mikrotikPppoePort := d.mikrotikPort.Pppoe()

	// Try to update first, if not exists then create
	err = mikrotikPppoePort.UpdateSecret(router, pppSecret)
	if err != nil {
		// If update fails, try to create
		err = mikrotikPppoePort.CreateSecret(router, pppSecret)
		if err != nil {
			return stacktrace.Propagate(err, "failed to create ppp secret on mikrotik")
		}
	}

	return nil
}

func (d *customerDomain) deleteFromMikrotik(ctx context.Context, customer *model.Customer) error {
	if customer.RouterID == nil {
		return nil
	}

	// Get router
	databaseMikrotikPort := d.databasePort.Mikrotik()
	router, err := databaseMikrotikPort.FindByID(customer.RouterID.String())
	if err != nil {
		return stacktrace.Propagate(err, "failed to find mikrotik router")
	}

	// Delete PPP secret
	mikrotikPppoePort := d.mikrotikPort.Pppoe()
	err = mikrotikPppoePort.DeleteSecret(router, customer.PppSecretName)
	if err != nil {
		return stacktrace.Propagate(err, "failed to delete ppp secret from mikrotik")
	}

	return nil
}

func calculateNextExpiryDate(currentDate time.Time, billingCycle model.BillingCycleType) time.Time {
	switch billingCycle {
	case model.BillingCycleMonthly:
		return currentDate.AddDate(0, 1, 0)
	case model.BillingCycleQuarterly:
		return currentDate.AddDate(0, 3, 0)
	case model.BillingCycleSemiAnnually:
		return currentDate.AddDate(0, 6, 0)
	case model.BillingCycleYearly:
		return currentDate.AddDate(1, 0, 0)
	default:
		return currentDate.AddDate(0, 1, 0)
	}
}
