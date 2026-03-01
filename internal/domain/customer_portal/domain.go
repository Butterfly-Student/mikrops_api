package customer_portal

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/palantir/stacktrace"

	"go-template/internal/domain/customer"
	"go-template/internal/model"
	outbound_port "go-template/internal/port/outbound"
	"go-template/utils/hash"
	"go-template/utils/token"
)

// PortalLoginResponse holds the JWT tokens returned on portal login / refresh.
type PortalLoginResponse struct {
	AccessToken  string         `json:"access_token"`
	RefreshToken string         `json:"refresh_token"`
	Customer     *model.Customer `json:"customer"`
}

// UpdateProfileInput limits what a customer can update on their own profile.
type UpdateProfileInput struct {
	FullName  string   `json:"full_name" validate:"required,max=100"`
	Email     *string  `json:"email" validate:"omitempty,email,max=100"`
	Phone     string   `json:"phone" validate:"required,max=20"`
	Address   *string  `json:"address"`
	Latitude  *float64 `json:"latitude" validate:"omitempty,min=-90,max=90"`
	Longitude *float64 `json:"longitude" validate:"omitempty,min=-180,max=180"`
}

// ChangePppInput for updating PPPoE credentials.
// TODO: This should be moved to Subscription domain as it deals with service credentials.
type ChangePppInput struct {
	NewUsername *string `json:"new_username" validate:"omitempty,min=3,max=100"`
	NewPassword *string `json:"new_password" validate:"omitempty,min=6,max=255"`
}

type CustomerPortalDomain interface {
	Login(ctx context.Context, identifier, password string) (*PortalLoginResponse, error)
	RefreshToken(ctx context.Context, refreshTokenStr string) (*PortalLoginResponse, error)
	GetProfile(ctx context.Context, customerID string) (*model.Customer, error)
	UpdateProfile(ctx context.Context, customerID string, input UpdateProfileInput) (*model.Customer, error)
	ChangePassword(ctx context.Context, customerID, oldPassword, newPassword string) error
	// TODO: ChangePppCredentials should be moved to Subscription domain.
	// This method temporarily gets the active subscription and updates credentials there.
	ChangePppCredentials(ctx context.Context, customerID string, input ChangePppInput) (*model.Subscription, error)
	ListInvoices(ctx context.Context, customerID string) ([]model.Invoice, error)
	GetInvoice(ctx context.Context, customerID, invoiceID string) (*model.Invoice, error)
}

type domain struct {
	dbPort         outbound_port.DatabasePort
	customerDomain customer.CustomerDomain
}

func NewCustomerPortalDomain(
	dbPort outbound_port.DatabasePort,
	customerDomain customer.CustomerDomain,
) CustomerPortalDomain {
	return &domain{
		dbPort:         dbPort,
		customerDomain: customerDomain,
	}
}

func (d *domain) Login(ctx context.Context, identifier, password string) (*PortalLoginResponse, error) {
	// Find customer by customer_code or phone (exact match)
	c, err := d.dbPort.Customer().FindByPortalIdentifier(ctx, identifier)
	if err != nil {
		return nil, stacktrace.NewError("invalid credentials")
	}

	// Check customer has portal access
	if c.PortalPassword == nil {
		return nil, stacktrace.NewError("portal access not enabled for this customer")
	}

	// Verify password
	if !hash.CheckPasswordHash(password, *c.PortalPassword) {
		return nil, stacktrace.NewError("invalid credentials")
	}

	// Only active/pending/isolated customers can log in
	if c.Status == model.CustomerStatusTerminated || c.Status == model.CustomerStatusSuspended {
		return nil, stacktrace.NewError("customer account is not active")
	}

	// Generate tokens
	accessToken, err := token.GenerateCustomerAccessToken(c.ID.String())
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to generate access token")
	}
	refreshToken, err := token.GenerateCustomerRefreshToken(c.ID.String())
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to generate refresh token")
	}

	// Update last login timestamp
	now := time.Now()
	c.PortalLastLogin = &now
	_ = d.dbPort.Customer().Update(ctx, c)

	return &PortalLoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		Customer:     c,
	}, nil
}

func (d *domain) RefreshToken(ctx context.Context, refreshTokenStr string) (*PortalLoginResponse, error) {
	claims, err := token.ValidateCustomerToken(refreshTokenStr, true)
	if err != nil {
		return nil, stacktrace.Propagate(err, "invalid refresh token")
	}

	customerID, ok := claims["sub"].(string)
	if !ok || customerID == "" {
		return nil, stacktrace.NewError("invalid token claims")
	}

	c, err := d.dbPort.Customer().FindByID(ctx, customerID)
	if err != nil {
		return nil, stacktrace.Propagate(err, "customer not found")
	}

	accessToken, err := token.GenerateCustomerAccessToken(c.ID.String())
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to generate access token")
	}
	newRefreshToken, err := token.GenerateCustomerRefreshToken(c.ID.String())
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to generate refresh token")
	}

	return &PortalLoginResponse{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
		Customer:     c,
	}, nil
}

func (d *domain) GetProfile(ctx context.Context, customerID string) (*model.Customer, error) {
	c, err := d.dbPort.Customer().FindByID(ctx, customerID)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to get customer profile")
	}
	return c, nil
}

func (d *domain) UpdateProfile(ctx context.Context, customerID string, input UpdateProfileInput) (*model.Customer, error) {
	c, err := d.dbPort.Customer().FindByID(ctx, customerID)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to get customer")
	}

	// Update only personal/contact info — never touch service/billing/MikroTik fields
	c.FullName = input.FullName
	c.Email = input.Email
	c.Phone = input.Phone
	c.Address = input.Address
	c.Latitude = input.Latitude
	c.Longitude = input.Longitude

	if err := d.dbPort.Customer().Update(ctx, c); err != nil {
		return nil, stacktrace.Propagate(err, "failed to update customer profile")
	}

	return c, nil
}

func (d *domain) ChangePassword(ctx context.Context, customerID, oldPassword, newPassword string) error {
	c, err := d.dbPort.Customer().FindByID(ctx, customerID)
	if err != nil {
		return stacktrace.Propagate(err, "failed to get customer")
	}

	if c.PortalPassword == nil || !hash.CheckPasswordHash(oldPassword, *c.PortalPassword) {
		return stacktrace.NewError("current password is incorrect")
	}

	hashedNew, err := hash.HashPassword(newPassword)
	if err != nil {
		return stacktrace.Propagate(err, "failed to hash new password")
	}

	c.PortalPassword = &hashedNew
	if err := d.dbPort.Customer().Update(ctx, c); err != nil {
		return stacktrace.Propagate(err, "failed to update portal password")
	}

	return nil
}

func (d *domain) ChangePppCredentials(ctx context.Context, customerID string, input ChangePppInput) (*model.Subscription, error) {
	if input.NewUsername == nil && input.NewPassword == nil {
		return nil, stacktrace.NewError("at least one of new_username or new_password must be provided")
	}

	// Get active subscriptions for this customer
	subscriptions, err := d.dbPort.Subscription().FindActiveByCustomerID(ctx, customerID)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to get customer subscriptions")
	}

	if len(subscriptions) == 0 {
		return nil, stacktrace.NewError("no active subscription found for this customer")
	}

	// Use the first active subscription
	activeSub := &subscriptions[0]

	// Check if customer has router assigned
	if activeSub.RouterID == uuid.Nil {
		return nil, stacktrace.NewError("subscription has no router assigned; cannot update PPPoE credentials")
	}

	// Update only the credential fields
	if input.NewUsername != nil {
		activeSub.Username = *input.NewUsername
	}
	if input.NewPassword != nil {
		activeSub.Password = *input.NewPassword
	}

	// Save the updated subscription
	if err := d.dbPort.Subscription().Update(ctx, activeSub); err != nil {
		return nil, stacktrace.Propagate(err, "failed to update PPPoE credentials")
	}

	return activeSub, nil
}

func (d *domain) ListInvoices(ctx context.Context, customerID string) ([]model.Invoice, error) {
	customerUUID, err := uuid.Parse(customerID)
	if err != nil {
		return nil, stacktrace.Propagate(err, "invalid customer id")
	}

	filter := &model.InvoiceFilter{
		CustomerID: &customerUUID,
	}

	invoices, err := d.dbPort.Invoice().FindAll(ctx, filter)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to list invoices")
	}

	return invoices, nil
}

func (d *domain) GetInvoice(ctx context.Context, customerID, invoiceID string) (*model.Invoice, error) {
	invoice, err := d.dbPort.Invoice().FindByID(ctx, invoiceID)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to get invoice")
	}

	// Ensure the invoice belongs to this customer
	if invoice.CustomerID.String() != customerID {
		return nil, stacktrace.NewError("invoice not found")
	}

	return invoice, nil
}
