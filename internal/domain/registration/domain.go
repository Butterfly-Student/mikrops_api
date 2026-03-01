package registration

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/palantir/stacktrace"

	"go-template/internal/domain/customer"
	"go-template/internal/model"
	outbound_port "go-template/internal/port/outbound"
	"go-template/utils"
	"go-template/utils/hash"
)

type RegistrationDomain interface {
	Submit(ctx context.Context, input model.RegistrationInput) (*model.CustomerRegistration, error)
	GetByID(ctx context.Context, id string) (*model.CustomerRegistration, error)
	List(ctx context.Context, filter *model.RegistrationFilter) ([]model.CustomerRegistration, error)
	Approve(ctx context.Context, id string, approverID string, input model.RegistrationApproveInput) (*model.RegistrationApprovalResult, error)
	Reject(ctx context.Context, id string, approverID string, input model.RegistrationRejectInput) (*model.CustomerRegistration, error)
}

type domain struct {
	dbPort            outbound_port.DatabasePort
	customerDomain    customer.CustomerDomain
	mikrotikSyncPort  outbound_port.MikrotikSyncMessagePort
}

func NewRegistrationDomain(
	dbPort outbound_port.DatabasePort,
	customerDomain customer.CustomerDomain,
	mikrotikSyncPort outbound_port.MikrotikSyncMessagePort,
) RegistrationDomain {
	return &domain{
		dbPort:           dbPort,
		customerDomain:   customerDomain,
		mikrotikSyncPort: mikrotikSyncPort,
	}
}

// TODO: This domain should use SubscriptionDomain once it's implemented.
// For now, we create subscriptions directly via database port within transactions.
// When SubscriptionDomain is ready, refactor to use:
//   subscriptionDomain subscription.SubscriptionDomain

func (d *domain) Submit(ctx context.Context, input model.RegistrationInput) (*model.CustomerRegistration, error) {
	// Validate bandwidth profile exists
	if input.BandwidthProfileID != nil {
		if _, err := d.dbPort.BandwidthProfile().FindByID(ctx, input.BandwidthProfileID.String()); err != nil {
			return nil, stacktrace.Propagate(err, "bandwidth profile not found")
		}
	}

	registration := &model.CustomerRegistration{
		ID:                 uuid.New(),
		FullName:           input.FullName,
		Email:              input.Email,
		Phone:              input.Phone,
		Address:            input.Address,
		Latitude:           input.Latitude,
		Longitude:          input.Longitude,
		Notes:              input.Notes,
		BandwidthProfileID: input.BandwidthProfileID,
		PreferredRouterID:  input.PreferredRouterID,
		Status:             model.RegistrationStatusPending,
	}

	if err := d.dbPort.Registration().Create(ctx, registration); err != nil {
		return nil, stacktrace.Propagate(err, "failed to save registration")
	}

	return registration, nil
}

func (d *domain) GetByID(ctx context.Context, id string) (*model.CustomerRegistration, error) {
	registration, err := d.dbPort.Registration().FindByID(ctx, id)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to get registration by id")
	}
	return registration, nil
}

func (d *domain) List(ctx context.Context, filter *model.RegistrationFilter) ([]model.CustomerRegistration, error) {
	registrations, err := d.dbPort.Registration().FindAll(ctx, filter)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to list registrations")
	}
	return registrations, nil
}

func (d *domain) Approve(ctx context.Context, id string, approverID string, input model.RegistrationApproveInput) (*model.RegistrationApprovalResult, error) {
	// Get registration
	registration, err := d.dbPort.Registration().FindByID(ctx, id)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to get registration")
	}

	if registration.Status != model.RegistrationStatusPending {
		return nil, stacktrace.NewError("registration is not in pending status")
	}

	// Get bandwidth profile to determine service type
	var serviceType model.ServiceType
	var planID uuid.UUID
	var plan *model.BandwidthProfile
	if registration.BandwidthProfileID != nil {
		var err error
		plan, err = d.dbPort.BandwidthProfile().FindByID(ctx, registration.BandwidthProfileID.String())
		if err != nil {
			return nil, stacktrace.Propagate(err, "failed to get bandwidth profile")
		}
		planID = plan.ID
		serviceType = plan.ServiceType
	} else {
		return nil, stacktrace.NewError("bandwidth profile is required for approval")
	}

	// Generate service credentials based on service type
	existingUsernames, err := d.dbPort.Subscription().ListUsernames(ctx)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to check existing usernames")
	}

	serviceUsername := generateServiceUsername(registration.FullName, existingUsernames)
	servicePassword := utils.GenerateSecureToken(8) // 16 hex chars

	// Auto-generate portal password (10-char random)
	portalPasswordPlain := utils.GenerateSecureToken(5) // 5 bytes = 10 hex chars

	// Hash portal password
	portalPasswordHash, err := hash.HashPassword(portalPasswordPlain)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to hash portal password")
	}

	// Auto-generate customer code
	customerCode := generateCustomerCode(registration.FullName, registration.ID)

	// Build CustomerInput - only identity fields (no technical fields)
	pendingStatus := string(model.CustomerStatusPending)
	now := time.Now()

	customerInput := model.CustomerInput{
		CustomerCode:     customerCode,
		FullName:         registration.FullName,
		Email:            registration.Email,
		Phone:            registration.Phone,
		Address:          registration.Address,
		Latitude:         registration.Latitude,
		Longitude:        registration.Longitude,
		Status:           &pendingStatus,
		InstallationDate: &now,
	}

	// Create customer (identity only)
	createdCustomer, err := d.customerDomain.Create(ctx, customerInput)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to create customer during approval")
	}

	// Save hashed portal password to customer
	createdCustomer.PortalPassword = &portalPasswordHash
	if err := d.dbPort.Customer().Update(ctx, createdCustomer); err != nil {
		return nil, stacktrace.Propagate(err, "failed to save portal password to customer")
	}

	// Create Subscription with technical details
	var createdSubscription *model.Subscription
	subscriptionStatus := string(model.SubscriptionStatusPending)

	subscriptionInput := model.SubscriptionInput{
		CustomerID:  createdCustomer.ID,
		PlanID:      planID,
		RouterID:    input.RouterID,
		ServiceType: string(serviceType),
		Username:    serviceUsername,
		Password:    servicePassword,
		Status:      &subscriptionStatus,
	}

	// Create subscription
	subscription := subscriptionInput.ToModel()
	if err := d.dbPort.Subscription().Create(ctx, subscription); err != nil {
		return nil, stacktrace.Propagate(err, "failed to create subscription during approval")
	}
	createdSubscription = subscription

	// Publish MikroTik sync message for PPPoE and Hotspot
	if serviceType == model.ServiceTypePPPoE || serviceType == model.ServiceTypeHotspot {
		if err := d.publishMikrotikSync(ctx, subscription, plan, serviceUsername, servicePassword); err != nil {
			// Log error but don't fail the approval - sync can be retried later
			// TODO: Add to dead letter queue or retry mechanism
			// For now, we continue with the approval
		}
	}

	// Mark registration as approved
	if err := d.dbPort.Registration().SetApproved(ctx, id, approverID, createdCustomer.ID.String()); err != nil {
		return nil, stacktrace.Propagate(err, "failed to mark registration as approved")
	}

	// Reload registration with relations
	updatedRegistration, _ := d.dbPort.Registration().FindByID(ctx, id)

	return &model.RegistrationApprovalResult{
		Registration:          updatedRegistration,
		Customer:              createdCustomer,
		Subscription:          createdSubscription,
		InitialPortalPassword: portalPasswordPlain,
		ServiceUsername:       serviceUsername,
		ServicePassword:       servicePassword,
	}, nil
}

func (d *domain) Reject(ctx context.Context, id string, approverID string, input model.RegistrationRejectInput) (*model.CustomerRegistration, error) {
	// Get registration
	registration, err := d.dbPort.Registration().FindByID(ctx, id)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to get registration")
	}

	if registration.Status != model.RegistrationStatusPending {
		return nil, stacktrace.NewError("registration is not in pending status")
	}

	if err := d.dbPort.Registration().SetRejected(ctx, id, approverID, input.Reason); err != nil {
		return nil, stacktrace.Propagate(err, "failed to reject registration")
	}

	registration.Status = model.RegistrationStatusRejected
	registration.RejectionReason = &input.Reason

	return registration, nil
}

// generateServiceUsername creates a unique service username from a full name.
// Works for all service types: PPPoE, Hotspot, VPN, etc.
// "John Doe Jr." → "johndoejr", with numeric suffix if already taken.
func generateServiceUsername(fullName string, existingNames []string) string {
	re := regexp.MustCompile(`[^a-z0-9]`)
	slug := re.ReplaceAllString(strings.ToLower(fullName), "")
	if slug == "" {
		slug = "user"
	}

	existingSet := make(map[string]bool, len(existingNames))
	for _, n := range existingNames {
		existingSet[strings.ToLower(n)] = true
	}

	candidate := slug
	for i := 1; existingSet[candidate]; i++ {
		candidate = fmt.Sprintf("%s%d", slug, i)
	}
	return candidate
}

// publishMikrotikSync publishes a sync message to RabbitMQ for MikroTik provisioning
func (d *domain) publishMikrotikSync(ctx context.Context, subscription *model.Subscription, plan *model.BandwidthProfile, username, password string) error {
	if d.mikrotikSyncPort == nil {
		return nil // Skip if sync port not configured
	}

	var syncMessage model.MikrotikSyncMessage

	switch plan.ServiceType {
	case model.ServiceTypePPPoE:
		localAddr := ""
		if plan.LocalAddress != nil {
			localAddr = *plan.LocalAddress
		}
		remoteAddr := ""
		if plan.RemoteAddress != nil {
			remoteAddr = *plan.RemoteAddress
		}
		syncMessage = model.MikrotikSyncMessage{
			SubscriptionID: subscription.ID,
			RouterID:       subscription.RouterID,
			Action:         model.MikrotikSyncActionCreate,
			ServiceType:    model.MikrotikServiceTypePPPoE,
			PPPoEData: &model.PPPoESyncData{
				Username:      username,
				Password:      password,
				Profile:       plan.PppProfileName,
				LocalAddress:  localAddr,
				RemoteAddress: remoteAddr,
				RateLimit:     plan.GetRateLimit(),
				Comment:       "Created via registration approval",
			},
		}
	case model.ServiceTypeHotspot:
		sharedUsers := 1
		if plan.SharedUsers != nil {
			sharedUsers = *plan.SharedUsers
		}
		syncMessage = model.MikrotikSyncMessage{
			SubscriptionID: subscription.ID,
			RouterID:       subscription.RouterID,
			Action:         model.MikrotikSyncActionCreate,
			ServiceType:    model.MikrotikServiceTypeHotspot,
			HotspotData: &model.HotspotSyncData{
				Username:    username,
				Password:    password,
				Profile:     plan.PppProfileName, // Hotspot profile uses same field
				SharedUsers: sharedUsers,
				RateLimit:   plan.GetRateLimit(),
				Comment:     "Created via registration approval",
			},
		}
	default:
		return fmt.Errorf("unsupported service type for MikroTik sync: %s", plan.ServiceType)
	}

	if err := d.mikrotikSyncPort.PublishSyncMessage(syncMessage); err != nil {
		return stacktrace.Propagate(err, "failed to publish MikroTik sync message")
	}

	return nil
}

// generateCustomerCode creates a unique customer code from name and registration UUID.
func generateCustomerCode(fullName string, id uuid.UUID) string {
	re := regexp.MustCompile(`[^A-Z0-9]`)
	slug := re.ReplaceAllString(strings.ToUpper(fullName), "")
	if len(slug) > 6 {
		slug = slug[:6]
	}
	if slug == "" {
		slug = "CUST"
	}
	// Append last 4 chars of UUID for uniqueness
	suffix := strings.ToUpper(strings.ReplaceAll(id.String(), "-", ""))
	suffix = suffix[len(suffix)-4:]
	return fmt.Sprintf("%s-%s", slug, suffix)
}
