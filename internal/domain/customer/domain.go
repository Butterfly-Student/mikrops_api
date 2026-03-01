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
	// TODO: Isolate - Moved to Subscription domain. Remove from interface once subscription domain is implemented.
	Isolate(ctx context.Context, id string) error
	// TODO: UnIsolate - Moved to Subscription domain. Remove from interface once subscription domain is implemented.
	UnIsolate(ctx context.Context, id string) error
	// TODO: SyncToMikrotik - Moved to Subscription domain. Remove from interface once subscription domain is implemented.
	SyncToMikrotik(ctx context.Context, id string) error
}

type domain struct {
	dbPort outbound_port.DatabasePort
	// TODO: mikrotikPort should be moved to Subscription domain
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

	// Create customer in DB - technical config (PPPoE, Router, etc) is now in Subscription
	if err := d.dbPort.Customer().Create(ctx, customer); err != nil {
		return nil, stacktrace.Propagate(err, "failed to save customer to database")
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

	// Apply new values - only identity fields, not technical config
	customer.CustomerCode = input.CustomerCode
	customer.FullName = input.FullName
	customer.Email = input.Email
	customer.Phone = input.Phone
	customer.IDCardNumber = input.IDCardNumber
	customer.Address = input.Address
	customer.Latitude = input.Latitude
	customer.Longitude = input.Longitude
	customer.ActivationDate = input.ActivationDate
	customer.InstallationDate = input.InstallationDate
	customer.TerminationDate = input.TerminationDate
	customer.AutoIsolate = input.AutoIsolate
	customer.GracePeriodDays = input.GracePeriodDays
	customer.Notes = input.Notes
	customer.Tags = input.Tags

	if input.Status != nil {
		customer.Status = model.CustomerStatus(*input.Status)
	}

	// Update in DB
	if err := d.dbPort.Customer().Update(ctx, customer); err != nil {
		return nil, stacktrace.Propagate(err, "failed to update customer in database")
	}

	return customer, nil
}

func (d *domain) Delete(ctx context.Context, id string) error {
	// Get customer to verify existence
	_, err := d.dbPort.Customer().FindByID(ctx, id)
	if err != nil {
		return stacktrace.Propagate(err, "failed to find customer")
	}

	// TODO: Check for active subscriptions before deletion
	// If customer has active subscriptions, they should be terminated first
	// This logic should be coordinated with Subscription domain

	// Soft delete in DB
	if err := d.dbPort.Customer().Delete(ctx, id); err != nil {
		return stacktrace.Propagate(err, "failed to delete customer from database")
	}

	return nil
}

func (d *domain) ChangeStatus(ctx context.Context, id string, status model.CustomerStatus) error {
	// Get existing customer
	customer, err := d.dbPort.Customer().FindByID(ctx, id)
	if err != nil {
		return stacktrace.Propagate(err, "failed to find customer")
	}

	customer.Status = status

	// Update status in DB
	if err := d.dbPort.Customer().UpdateStatus(ctx, id, status); err != nil {
		return stacktrace.Propagate(err, "failed to update customer status in database")
	}

	// TODO: When status changes, sync to related subscriptions
	// This should be handled by Subscription domain:
	// - If customer is suspended/isolated, all active subscriptions should also be suspended/isolated
	// - If customer is reactivated, subscriptions should follow their own status logic

	return nil
}

// Isolate isolates a customer by switching their subscriptions to isolation profile
// TODO: This method should be moved to Subscription domain.
// Customer.Isolate should delegate to Subscription.Isolate for each active subscription.
func (d *domain) Isolate(ctx context.Context, id string) error {
	// Get customer
	_, err := d.dbPort.Customer().FindByID(ctx, id)
	if err != nil {
		return stacktrace.Propagate(err, "failed to find customer")
	}

	// TODO: Implement isolation at subscription level
	// This method should:
	// 1. Find all active subscriptions for this customer
	// 2. Call Subscription.Isolate() for each subscription
	// 3. Update customer status to Isolated
	//
	// For now, just update the customer status
	return stacktrace.NewError("Isolate operation has been moved to Subscription domain. Please use Subscription.Isolate() instead")
}

// UnIsolate restores a customer from isolation
// TODO: This method should be moved to Subscription domain.
// Customer.UnIsolate should delegate to Subscription.UnIsolate for each isolated subscription.
func (d *domain) UnIsolate(ctx context.Context, id string) error {
	// Get customer
	_, err := d.dbPort.Customer().FindByID(ctx, id)
	if err != nil {
		return stacktrace.Propagate(err, "failed to find customer")
	}

	// TODO: Implement un-isolation at subscription level
	// This method should:
	// 1. Find all isolated subscriptions for this customer
	// 2. Call Subscription.UnIsolate() for each subscription
	// 3. Update customer status to Active
	//
	// For now, return error indicating this needs to be implemented at subscription level
	return stacktrace.NewError("UnIsolate operation has been moved to Subscription domain. Please use Subscription.UnIsolate() instead")
}

// SyncToMikrotik syncs customer subscriptions to MikroTik
// TODO: This method should be moved to Subscription domain.
func (d *domain) SyncToMikrotik(ctx context.Context, id string) error {
	// Get customer
	_, err := d.dbPort.Customer().FindByID(ctx, id)
	if err != nil {
		return stacktrace.Propagate(err, "failed to get customer")
	}

	// TODO: Implement sync at subscription level
	// This method should:
	// 1. Find all subscriptions for this customer
	// 2. Call Subscription.SyncToMikrotik() for each subscription
	//
	// For now, return error indicating this needs to be implemented at subscription level
	return stacktrace.NewError("SyncToMikrotik operation has been moved to Subscription domain. Please use Subscription.SyncToMikrotik() instead")
}
