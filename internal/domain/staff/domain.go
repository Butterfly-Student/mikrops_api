package staff

import (
	"context"

	"github.com/palantir/stacktrace"
	"golang.org/x/crypto/bcrypt"

	"mikrops/internal/model"
	outbound_port "mikrops/internal/port/outbound"
)

type StaffDomain interface {
	Create(ctx context.Context, input model.StaffInput) (model.Staff, error)
	FindByFilter(ctx context.Context, filter model.StaffFilter) ([]model.Staff, error)
	FindByID(ctx context.Context, id string) (model.Staff, error)
	FindByEmail(ctx context.Context, email string) (model.Staff, error)
	Update(ctx context.Context, id string, input model.StaffInput) error
	Delete(ctx context.Context, id string) error
}

type staffDomain struct {
	databasePort outbound_port.DatabasePort
	messagePort  outbound_port.MessagePort
	cachePort    outbound_port.CachePort
	workflowPort outbound_port.WorkflowPort
}

func NewStaffDomain(
	databasePort outbound_port.DatabasePort,
	messagePort outbound_port.MessagePort,
	cachePort outbound_port.CachePort,
	workflowPort outbound_port.WorkflowPort,
) StaffDomain {
	return &staffDomain{
		databasePort: databasePort,
		messagePort:  messagePort,
		cachePort:    cachePort,
		workflowPort: workflowPort,
	}
}

func (d *staffDomain) Create(ctx context.Context, input model.StaffInput) (model.Staff, error) {
	// Hash password if provided
	if input.Password != "" {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
		if err != nil {
			return model.Staff{}, stacktrace.Propagate(err, "failed to hash password")
		}
		input.PasswordHash = string(hashedPassword)
		input.Password = "" // Clear plain password
	}

	staff, err := d.databasePort.Staff().Create(input)
	if err != nil {
		return model.Staff{}, stacktrace.Propagate(err, "failed to create staff")
	}

	return staff, nil
}

func (d *staffDomain) FindByFilter(ctx context.Context, filter model.StaffFilter) ([]model.Staff, error) {
	if filter.IsEmpty() {
		return nil, stacktrace.NewError("filter is empty")
	}

	staffs, err := d.databasePort.Staff().FindByFilter(filter)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to find staffs by filter")
	}

	return staffs, nil
}

func (d *staffDomain) FindByID(ctx context.Context, id string) (model.Staff, error) {
	if id == "" {
		return model.Staff{}, stacktrace.NewError("id is empty")
	}

	staff, err := d.databasePort.Staff().FindByID(id)
	if err != nil {
		return model.Staff{}, stacktrace.Propagate(err, "failed to find staff by id")
	}

	return staff, nil
}

func (d *staffDomain) FindByEmail(ctx context.Context, email string) (model.Staff, error) {
	if email == "" {
		return model.Staff{}, stacktrace.NewError("email is empty")
	}

	staff, err := d.databasePort.Staff().FindByEmail(email)
	if err != nil {
		return model.Staff{}, stacktrace.Propagate(err, "failed to find staff by email")
	}

	return staff, nil
}

func (d *staffDomain) Update(ctx context.Context, id string, input model.StaffInput) error {
	if id == "" {
		return stacktrace.NewError("id is empty")
	}

	// Hash password if changed
	if input.Password != "" {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
		if err != nil {
			return stacktrace.Propagate(err, "failed to hash password")
		}
		input.PasswordHash = string(hashedPassword)
		input.Password = "" // Clear plain password
	}

	err := d.databasePort.Staff().Update(id, input)
	if err != nil {
		return stacktrace.Propagate(err, "failed to update staff")
	}

	return nil
}

func (d *staffDomain) Delete(ctx context.Context, id string) error {
	if id == "" {
		return stacktrace.NewError("id is empty")
	}

	err := d.databasePort.Staff().Delete(id)
	if err != nil {
		return stacktrace.Propagate(err, "failed to delete staff")
	}

	return nil
}
