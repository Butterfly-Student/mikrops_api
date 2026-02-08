package internet_package

import (
	"context"

	"github.com/palantir/stacktrace"

	"mikrops/internal/model"
	outbound_port "mikrops/internal/port/outbound"
)

type InternetPackageDomain interface {
	Create(ctx context.Context, input model.InternetPackageInput) (model.InternetPackage, error)
	FindByFilter(ctx context.Context, filter model.InternetPackageFilter) ([]model.InternetPackage, error)
	FindByID(ctx context.Context, id string) (model.InternetPackage, error)
	Update(ctx context.Context, id string, input model.InternetPackageInput) error
	Delete(ctx context.Context, id string) error
}

type internetPackageDomain struct {
	databasePort outbound_port.DatabasePort
	messagePort  outbound_port.MessagePort
	cachePort    outbound_port.CachePort
	workflowPort outbound_port.WorkflowPort
}

func NewInternetPackageDomain(
	databasePort outbound_port.DatabasePort,
	messagePort outbound_port.MessagePort,
	cachePort outbound_port.CachePort,
	workflowPort outbound_port.WorkflowPort,
) InternetPackageDomain {
	return &internetPackageDomain{
		databasePort: databasePort,
		messagePort:  messagePort,
		cachePort:    cachePort,
		workflowPort: workflowPort,
	}
}

func (d *internetPackageDomain) Create(ctx context.Context, input model.InternetPackageInput) (model.InternetPackage, error) {
	pkg, err := d.databasePort.InternetPackage().Create(input)
	if err != nil {
		return model.InternetPackage{}, stacktrace.Propagate(err, "failed to create internet package")
	}

	return pkg, nil
}

func (d *internetPackageDomain) FindByFilter(ctx context.Context, filter model.InternetPackageFilter) ([]model.InternetPackage, error) {
	if filter.IsEmpty() {
		return nil, stacktrace.NewError("filter is empty")
	}

	packages, err := d.databasePort.InternetPackage().FindByFilter(filter)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to find internet packages by filter")
	}

	return packages, nil
}

func (d *internetPackageDomain) FindByID(ctx context.Context, id string) (model.InternetPackage, error) {
	if id == "" {
		return model.InternetPackage{}, stacktrace.NewError("id is empty")
	}

	pkg, err := d.databasePort.InternetPackage().FindByID(id)
	if err != nil {
		return model.InternetPackage{}, stacktrace.Propagate(err, "failed to find internet package by id")
	}

	return pkg, nil
}

func (d *internetPackageDomain) Update(ctx context.Context, id string, input model.InternetPackageInput) error {
	if id == "" {
		return stacktrace.NewError("id is empty")
	}

	err := d.databasePort.InternetPackage().Update(id, input)
	if err != nil {
		return stacktrace.Propagate(err, "failed to update internet package")
	}

	return nil
}

func (d *internetPackageDomain) Delete(ctx context.Context, id string) error {
	if id == "" {
		return stacktrace.NewError("id is empty")
	}

	err := d.databasePort.InternetPackage().Delete(id)
	if err != nil {
		return stacktrace.Propagate(err, "failed to delete internet package")
	}

	return nil
}
