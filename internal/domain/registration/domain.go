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
	Approve(ctx context.Context, id string, approverID uint, input model.RegistrationApproveInput) (*model.RegistrationApprovalResult, error)
	Reject(ctx context.Context, id string, approverID uint, input model.RegistrationRejectInput) (*model.CustomerRegistration, error)
}

type domain struct {
	dbPort         outbound_port.DatabasePort
	customerDomain customer.CustomerDomain
}

func NewRegistrationDomain(
	dbPort outbound_port.DatabasePort,
	customerDomain customer.CustomerDomain,
) RegistrationDomain {
	return &domain{
		dbPort:         dbPort,
		customerDomain: customerDomain,
	}
}

func (d *domain) Submit(ctx context.Context, input model.RegistrationInput) (*model.CustomerRegistration, error) {
	// Validate bandwidth profile exists
	if input.BandwidthProfileID != nil {
		if _, err := d.dbPort.BandwidthProfile().FindByID(ctx, input.BandwidthProfileID.String()); err != nil {
			return nil, stacktrace.Propagate(err, "bandwidth profile not found")
		}
	}

	// Collect existing PPP secret names to ensure uniqueness
	existingNames, err := d.dbPort.Registration().ListPppSecretNames(ctx)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to check existing ppp secret names")
	}

	// Auto-generate PPP username from full_name
	pppSecretName := generatePppUsername(input.FullName, existingNames)

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
		PppSecretName:      &pppSecretName,
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

func (d *domain) Approve(ctx context.Context, id string, approverID uint, input model.RegistrationApproveInput) (*model.RegistrationApprovalResult, error) {
	// Get registration
	registration, err := d.dbPort.Registration().FindByID(ctx, id)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to get registration")
	}

	if registration.Status != model.RegistrationStatusPending {
		return nil, stacktrace.NewError("registration is not in pending status")
	}

	// Look up bandwidth profile
	var profileID *uuid.UUID
	if registration.BandwidthProfileID != nil {
		profileID = registration.BandwidthProfileID
	}

	// Auto-generate PPP password (16-char random alphanumeric)
	pppPassword := utils.GenerateSecureToken(8) // 8 bytes = 16 hex chars

	// Auto-generate portal password (10-char random)
	portalPasswordPlain := utils.GenerateSecureToken(5) // 5 bytes = 10 hex chars

	// Hash portal password
	portalPasswordHash, err := hash.HashPassword(portalPasswordPlain)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to hash portal password")
	}

	// Determine PPP username (use pre-generated or regenerate)
	pppUsername := ""
	if registration.PppSecretName != nil {
		pppUsername = *registration.PppSecretName
	} else {
		existingNames, _ := d.dbPort.Registration().ListPppSecretNames(ctx)
		pppUsername = generatePppUsername(registration.FullName, existingNames)
	}

	// Auto-generate customer code
	customerCode := generateCustomerCode(registration.FullName, registration.ID)

	// Build CustomerInput
	routerID := input.RouterID
	pendingStatus := string(model.CustomerStatusPending)
	pppService := string(model.PPPServicePPPoE)
	now := time.Now()

	customerInput := model.CustomerInput{
		CustomerCode:      customerCode,
		FullName:          registration.FullName,
		Email:             registration.Email,
		Phone:             registration.Phone,
		Address:           registration.Address,
		Latitude:          registration.Latitude,
		Longitude:         registration.Longitude,
		RouterID:          &routerID,
		PppSecretName:     &pppUsername,
		PppSecretPassword: &pppPassword,
		PppService:        &pppService,
		ProfileID:         profileID,
		Status:            &pendingStatus,
		InstallationDate:  &now,
	}

	// Create customer (MikroTik-first)
	createdCustomer, err := d.customerDomain.Create(ctx, customerInput)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to create customer during approval")
	}

	// Save hashed portal password to customer
	createdCustomer.PortalPassword = &portalPasswordHash
	if err := d.dbPort.Customer().Update(ctx, createdCustomer); err != nil {
		return nil, stacktrace.Propagate(err, "failed to save portal password to customer")
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
		InitialPortalPassword: portalPasswordPlain,
		PppSecretName:         pppUsername,
		PppSecretPassword:     pppPassword,
	}, nil
}

func (d *domain) Reject(ctx context.Context, id string, approverID uint, input model.RegistrationRejectInput) (*model.CustomerRegistration, error) {
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

// generatePppUsername creates a unique PPP username from a full name.
// "John Doe Jr." → "johndoejr", with numeric suffix if already taken.
func generatePppUsername(fullName string, existingNames []string) string {
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
