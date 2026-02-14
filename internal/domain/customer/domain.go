package customer

import (
	"context"
	"errors"
	"fmt"

	"go-template/internal/domain/bandwidth_profile"
	"go-template/internal/domain/notification"
	"go-template/internal/model"
	outbound_port "go-template/internal/port/outbound"
	"go-template/utils/log"
)

type CustomerDomain interface {
	CreateCustomer(ctx context.Context, input model.CustomerInput) (*model.Customer, error)
	GetCustomer(ctx context.Context, id string) (*model.Customer, error)
	ListCustomers(ctx context.Context, filter model.CustomerFilter) ([]model.Customer, error)
	UpdateCustomer(ctx context.Context, id string, input model.CustomerInput) (*model.Customer, error)
	DeleteCustomer(ctx context.Context, id string) error
	SyncToMikrotik(ctx context.Context, customer *model.Customer) error
	IsolateCustomer(ctx context.Context, customerID string) error
	ActivateCustomer(ctx context.Context, customerID string) error
	FindExpired(ctx context.Context) ([]model.Customer, error)
	FindExpiringSoon(ctx context.Context, days int) ([]model.Customer, error)
	FindByPppSecretName(ctx context.Context, name string) (*model.Customer, error)
}

type domain struct {
	dbPort       outbound_port.DatabasePort
	mikrotikPort outbound_port.MikrotikPort
}

func (d *domain) getBandwidthProfile(ctx context.Context) bandwidth_profile.BandwidthProfileDomain {
	return bandwidth_profile.NewBandwidthProfileDomain(d.dbPort, d.mikrotikPort)
}

func (d *domain) getNotificationDomain(ctx context.Context) notification.NotificationDomain {
	return nil
}

func (d *domain) getPaymentPortalURL() string {
	return "portal.example.com"
}

func (d *domain) NewCustomerDomain(
	dbPort outbound_port.DatabasePort,
	mikrotikPort outbound_port.MikrotikPort,
) CustomerDomain {
	return &domain{
		dbPort:       dbPort,
		mikrotikPort: mikrotikPort,
	}
}

func (d *domain) CreateCustomer(ctx context.Context, input model.CustomerInput) (*model.Customer, error) {
	status := "pending"
	if input.Status != nil {
		status = *input.Status
	}

	var result interface{}
	var err error
	result, err = d.dbPort.DoInTransaction(func(repo outbound_port.DatabasePort) (interface{}, error) {
		customerCode := ""
		if input.CustomerCode != nil {
			customerCode = *input.CustomerCode
		}

		customer := &model.Customer{
			CustomerCode:      customerCode,
			FullName:          input.FullName,
			Email:             input.Email,
			Phone:             input.Phone,
			Address:           input.Address,
			Coordinates:       input.Coordinates,
			Status:            status,
			ActivationDate:    input.ActivationDate,
			InstallationDate:  input.InstallationDate,
			TerminationDate:   input.TerminationDate,
			ExpiryDate:        input.ExpiryDate,
			RouterID:          input.RouterID,
			PppSecretName:     input.PppSecretName,
			PppSecretPassword: input.PppSecretPassword,
			PppService:        input.PppService,
			StaticIP:          input.StaticIP,
			MACAddress:        input.MACAddress,
			ProfileID:         input.ProfileID,
			BillingCycle:      input.BillingCycle,
			BillingDay:        input.BillingDay,
			PaymentMethodPref: input.PaymentMethodPref,
			AutoIsolate:       input.AutoIsolate,
			GracePeriodDays:   input.GracePeriodDays,
			Notes:             input.Notes,
			Tags:              input.Tags,
		}

		model.CustomerPrepare(customer)

		err := repo.Customer().Create(customer)
		if err != nil {
			return nil, err
		}

		if input.PppSecretName != nil && input.PppSecretPassword != nil {
			router, err := repo.Mikrotik().FindByID(input.RouterID.String())
			if err != nil {
				return nil, err
			}

			if router.IsActive != nil && *router.IsActive {
				secret := &model.PppoeSecret{
					Name:     *input.PppSecretName,
					Password: *input.PppSecretPassword,
					Service:  "pppoe",
				}

				if customer.Profile != nil {
					secret.Profile = customer.Profile.PppProfileName
				}

				err = d.mikrotikPort.CreateSecret(router, secret)
				if err != nil {
					return nil, errors.New("failed to create PPP secret on Mikrotik")
				}
			}
		}

		return customer, nil
	})

	if err != nil {
		return nil, err
	}

	return result.(*model.Customer), nil
}

func (d *domain) GetCustomer(ctx context.Context, id string) (*model.Customer, error) {
	return d.dbPort.Customer().FindByID(id)
}

func (d *domain) ListCustomers(ctx context.Context, filter model.CustomerFilter) ([]model.Customer, error) {
	if filter.IsEmpty() {
		return d.dbPort.Customer().FindAll()
	}
	return d.dbPort.Customer().Find(filter)
}

func (d *domain) UpdateCustomer(ctx context.Context, id string, input model.CustomerInput) (*model.Customer, error) {
	var result interface{}
	var err error
	result, err = d.dbPort.DoInTransaction(func(repo outbound_port.DatabasePort) (interface{}, error) {
		customer, err := repo.Customer().FindByID(id)
		if err != nil {
			return nil, errors.New("customer not found")
		}

		if input.FullName != "" {
			customer.FullName = input.FullName
		}
		if input.Email != nil {
			customer.Email = input.Email
		}
		if input.Phone != "" {
			customer.Phone = input.Phone
		}
		if input.Address != nil {
			customer.Address = input.Address
		}
		if input.Coordinates != nil {
			customer.Coordinates = input.Coordinates
		}
		if input.ExpiryDate != nil {
			customer.ExpiryDate = input.ExpiryDate
		}
		if input.ProfileID != nil {
			customer.ProfileID = input.ProfileID
		}
		if input.Status != nil {
			customer.Status = *input.Status
		}

		err = repo.Customer().Update(customer)
		if err != nil {
			return nil, err
		}

		if customer.PppSecretName != nil && customer.PppSecretPassword != nil {
			router, err := repo.Mikrotik().FindByID(customer.RouterID.String())
			if err != nil {
				return nil, err
			}

			if router.IsActive != nil && *router.IsActive {
				secret := &model.PppoeSecret{
					Name:     *customer.PppSecretName,
					Password: *customer.PppSecretPassword,
					Service:  "pppoe",
				}

				if customer.Profile != nil {
					secret.Profile = customer.Profile.PppProfileName
				}

				err = d.mikrotikPort.UpdateSecret(router, secret)
				if err != nil {
					return nil, errors.New("failed to update PPP secret on Mikrotik")
				}
			}
		}

		return customer, nil
	})

	if err != nil {
		return nil, err
	}

	return result.(*model.Customer), nil
}

func (d *domain) DeleteCustomer(ctx context.Context, id string) error {
	customer, err := d.dbPort.Customer().FindByID(id)
	if err != nil {
		return errors.New("customer not found")
	}

	if customer.PppSecretName != nil {
		router, err := d.dbPort.Mikrotik().FindByID(customer.RouterID.String())
		if err != nil {
			return err
		}

		if router.IsActive != nil && *router.IsActive {
			err = d.mikrotikPort.DeleteSecret(router, *customer.PppSecretName)
			if err != nil {
				return errors.New("failed to delete PPP secret from Mikrotik")
			}
		}
	}

	return d.dbPort.Customer().Delete(id)
}

func (d *domain) SyncToMikrotik(ctx context.Context, customer *model.Customer) error {
	if customer.PppSecretName == nil || customer.PppSecretPassword == nil {
		return errors.New("customer does not have PPP secret configured")
	}

	router, err := d.dbPort.Mikrotik().FindByID(customer.RouterID.String())
	if err != nil {
		return err
	}

	if router.IsActive == nil || !*router.IsActive {
		return errors.New("router is not active")
	}

	secret := &model.PppoeSecret{
		Name:     *customer.PppSecretName,
		Password: *customer.PppSecretPassword,
		Service:  "pppoe",
	}

	if customer.Profile != nil {
		secret.Profile = customer.Profile.PppProfileName
	}

	err = d.mikrotikPort.CreateSecret(router, secret)
	if err != nil {
		return errors.New("failed to sync customer to Mikrotik")
	}

	return nil
}

func (d *domain) FindExpired(ctx context.Context) ([]model.Customer, error) {
	return d.dbPort.Customer().FindExpired()
}

func (d *domain) FindExpiringSoon(ctx context.Context, days int) ([]model.Customer, error) {
	return d.dbPort.Customer().FindExpiringSoon(days)
}

func (d *domain) FindByPppSecretName(ctx context.Context, name string) (*model.Customer, error) {
	return d.dbPort.Customer().FindByPppSecretName(name)
}

func (d *domain) IsolateCustomer(ctx context.Context, customerID string) error {
	customer, err := d.dbPort.Customer().FindByID(customerID)
	if err != nil {
		return fmt.Errorf("customer not found: %w", err)
	}

	// Get isolation profile (limited speed)
	isolatedProfile, err := d.getBandwidthProfile(ctx).GetIsolatedProfile(ctx)
	if err != nil {
		return fmt.Errorf("failed to get isolation profile: %w", err)
	}

	// Update customer status
	customer.Status = "isolated"

	// Update PPP secret on Mikrotik
	if customer.PppSecretName != nil {
		router, err := d.dbPort.Mikrotik().FindByID(customer.RouterID.String())
		if err != nil {
			return fmt.Errorf("router not found: %w", err)
		}

		secret := &model.PppoeSecret{
			Name:     *customer.PppSecretName,
			Profile:  isolatedProfile.PppProfileName,
			Disabled: false, // Allow login but with limited speed
		}

		if customer.PppSecretPassword != nil {
			secret.Password = *customer.PppSecretPassword
		}

		err = d.mikrotikPort.UpdateSecret(router, secret)
		if err != nil {
			return fmt.Errorf("failed to update PPP secret on Mikrotik: %w", err)
		}

		log.WithContext(ctx).Info(fmt.Sprintf("Customer %s isolated on router %s with profile %s", customer.CustomerCode, router.Address, isolatedProfile.PppProfileName))
	}

	// Generate redirect script
	redirectScript := d.GenerateRedirectScript(customer)
	// Apply script to Mikrotik router
	router, err := d.dbPort.Mikrotik().FindByID(customer.RouterID.String())
	if err != nil {
		return fmt.Errorf("router not found: %w", err)
	}

	// Apply firewall rules for isolation
	firewallRules := d.GenerateIsolationFirewallRules(customer)
	for _, rule := range firewallRules {
		err = d.mikrotikPort.AddFirewallRule(router, rule)
		if err != nil {
			log.WithContext(ctx).Warn(fmt.Sprintf("Failed to add firewall rule: %v", err))
		}
	}

	log.WithContext(ctx).Info(fmt.Sprintf("Generated redirect script for customer %s: %s", customer.CustomerCode, redirectScript))

	// Save customer status
	err = d.dbPort.Customer().Update(customer)
	if err != nil {
		return fmt.Errorf("failed to update customer status: %w", err)
	}

	return nil
}

func (d *domain) ActivateCustomer(ctx context.Context, customerID string) error {
	customer, err := d.dbPort.Customer().FindByID(customerID)
	if err != nil {
		return fmt.Errorf("customer not found: %w", err)
	}

	// Get original profile
	if customer.Profile == nil {
		return errors.New("customer has no profile assigned")
	}

	originalProfile, err := d.bandwidthProfile.GetProfile(ctx, customer.ProfileID.String())
	if err != nil {
		return fmt.Errorf("failed to get original profile: %w", err)
	}

	// Update customer status
	customer.Status = "active"

	// Update PPP secret on Mikrotik to restore original profile
	if customer.PppSecretName != nil {
		router, err := d.dbPort.Mikrotik().FindByID(customer.RouterID.String())
		if err != nil {
			return fmt.Errorf("router not found: %w", err)
		}

		secret := &model.PppoeSecret{
			Name:     *customer.PppSecretName,
			Profile:  originalProfile.PppProfileName,
			Disabled: false,
		}

		if customer.PppSecretPassword != nil {
			secret.Password = *customer.PppSecretPassword
		}

		err = d.mikrotikPort.UpdateSecret(router, secret)
		if err != nil {
			return fmt.Errorf("failed to restore PPP secret on Mikrotik: %w", err)
		}

		log.WithContext(ctx).Info(fmt.Sprintf("Customer %s activated on router %s with profile %s", customer.CustomerCode, router.Host, originalProfile.PppProfileName))
	}

	// Remove isolation firewall rules
	router, err := d.dbPort.Mikrotik().FindByID(customer.RouterID.String())
	if err == nil {
		firewallRules := d.GenerateIsolationFirewallRules(customer)
		for _, rule := range firewallRules {
			err = d.mikrotikPort.RemoveFirewallRule(router, rule)
			if err != nil {
				log.WithContext(ctx).Warn(fmt.Sprintf("Failed to remove firewall rule: %v", err))
			}
		}
	}

	// Save customer status
	err = d.dbPort.Customer().Update(customer)
	if err != nil {
		return fmt.Errorf("failed to update customer status: %w", err)
	}

	// Send notification if available
	if d.notificationPort != nil {
		// Send activation notification to customer
		// notificationPort.SendActivationNotification(ctx, customerID, "Customer activated, internet restored")
	}

	return nil
}

func (d *domain) GenerateRedirectScript(customer *model.Customer, router *model.MikrotikRouter) string {
	portalURL := d.getPaymentPortalURL()
	customerCode := customer.CustomerCode
	routerAddress := ""

	if router != nil {
		routerAddress = router.Address
	}

	// RouterOS script to redirect to portal
	script := fmt.Sprintf("# Customer: %s\n:local portalURL \"%s\"\n:local customerCode \"%s\"\n# Redirect HTTP to portal\n/ip firewall nat add chain=dstnat protocol=tcp dst-port=80 src-address=%s action=redirect to-ports=8080\n# Walled garden (allow payment portal)\n/ip firewall nat add chain=dstnat protocol=tcp dst-port=80,443 dst-address-list=portal-allowed action=accept",
		customerCode, portalURL, customerCode, routerAddress)

	return script
}

func (d *domain) GenerateIsolationFirewallRules(customer *model.Customer, router *model.MikrotikRouter) []model.FirewallRule {
	portalURL := d.getPaymentPortalURL()

	rules := []model.FirewallRule{
		{
			Chain:      "dstnat",
			Protocol:   "tcp",
			SrcAddress: router.Address,
			DstPort:     80,
			Action:     "redirect",
			ToPorts:     "8080",
			Comment:    fmt.Sprintf("Redirect %s to payment portal", customer.CustomerCode),
		},
		{
			Chain:      "dstnat",
			Protocol:   "tcp",
			SrcAddress: router.Address,
			DstPort:     443,
			Action:     "redirect",
			ToPorts:     "8080",
			Comment:    fmt.Sprintf("Redirect %s HTTPS to payment portal", customer.CustomerCode),
		},
		{
			Chain:      "dstnat",
			Protocol:   "tcp",
			DstPort:    80,
			Action:     "accept",
			DstAddress: portalURL,
			Comment:    "Allow payment portal access",
		},
		{
			Chain:      "dstnat",
			Protocol:   "tcp",
			DstPort:    443,
			Action:     "accept",
			DstAddress: portalURL,
			Comment:    "Allow payment portal HTTPS access",
			},
	}

	return rules
}