package postgres_outbound_adapter

import (
	"context"

	"github.com/google/uuid"
	"github.com/palantir/stacktrace"
	"gorm.io/gorm"

	"go-template/internal/model"
	outbound_port "go-template/internal/port/outbound"
)

const tableCustomer = "customers"

type customerAdapter struct {
	db *gorm.DB
}

func NewCustomerAdapter(
	db *gorm.DB,
) outbound_port.CustomerDatabasePort {
	return &customerAdapter{
		db: db,
	}
}

func (a *customerAdapter) Create(ctx context.Context, customer *model.Customer) error {
	if err := a.db.WithContext(ctx).Create(customer).Error; err != nil {
		return stacktrace.Propagate(err, "failed to create customer in database")
	}
	return nil
}

func (a *customerAdapter) FindByID(ctx context.Context, id string) (*model.Customer, error) {
	var customer model.Customer

	customerID, err := uuid.Parse(id)
	if err != nil {
		return nil, stacktrace.Propagate(err, "invalid customer id format")
	}

	if err := a.db.WithContext(ctx).
		Preload("Profile").
		Preload("Router").
		Where("id = ?", customerID).
		First(&customer).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, stacktrace.NewError("customer not found")
		}
		return nil, stacktrace.Propagate(err, "failed to find customer by id")
	}

	return &customer, nil
}

func (a *customerAdapter) FindByCode(ctx context.Context, code string) (*model.Customer, error) {
	var customer model.Customer

	if err := a.db.WithContext(ctx).
		Preload("Profile").
		Preload("Router").
		Where("customer_code = ?", code).
		First(&customer).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, stacktrace.NewError("customer not found")
		}
		return nil, stacktrace.Propagate(err, "failed to find customer by code")
	}

	return &customer, nil
}

func (a *customerAdapter) FindByPppSecretName(ctx context.Context, name string) (*model.Customer, error) {
	var customer model.Customer

	if err := a.db.WithContext(ctx).
		Preload("Profile").
		Preload("Router").
		Where("ppp_secret_name = ?", name).
		First(&customer).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, stacktrace.NewError("customer not found")
		}
		return nil, stacktrace.Propagate(err, "failed to find customer by ppp secret name")
	}

	return &customer, nil
}

func (a *customerAdapter) FindAll(ctx context.Context, filter *model.CustomerFilter) ([]model.Customer, error) {
	var customers []model.Customer

	query := a.db.WithContext(ctx).Model(&model.Customer{})

	// Preload relations
	query = query.Preload("Profile").Preload("Router")

	// Apply filters if provided
	if filter != nil {
		if filter.Status != nil {
			query = query.Where("status = ?", *filter.Status)
		}

		if filter.RouterID != nil {
			query = query.Where("router_id = ?", *filter.RouterID)
		}

		if filter.ProfileID != nil {
			query = query.Where("profile_id = ?", *filter.ProfileID)
		}

		if filter.BillingCycle != nil {
			query = query.Where("billing_cycle = ?", *filter.BillingCycle)
		}

		if filter.AutoIsolate != nil {
			query = query.Where("auto_isolate = ?", *filter.AutoIsolate)
		}

		if filter.ExpiryDateFrom != nil {
			query = query.Where("expiry_date >= ?", *filter.ExpiryDateFrom)
		}

		if filter.ExpiryDateTo != nil {
			query = query.Where("expiry_date <= ?", *filter.ExpiryDateTo)
		}

		if filter.Search != nil && *filter.Search != "" {
			searchPattern := "%" + *filter.Search + "%"
			query = query.Where(
				"full_name ILIKE ? OR customer_code ILIKE ? OR phone ILIKE ? OR email ILIKE ?",
				searchPattern, searchPattern, searchPattern, searchPattern,
			)
		}

		// Handle soft deletes
		if !filter.IncludeDeleted {
			query = query.Where("deleted_at IS NULL")
		} else {
			query = query.Unscoped()
		}
	} else {
		// Default: exclude deleted
		query = query.Where("deleted_at IS NULL")
	}

	// Order by creation date
	query = query.Order("created_at DESC")

	if err := query.Find(&customers).Error; err != nil {
		return nil, stacktrace.Propagate(err, "failed to find customers")
	}

	return customers, nil
}

func (a *customerAdapter) Update(ctx context.Context, customer *model.Customer) error {
	if err := a.db.WithContext(ctx).Save(customer).Error; err != nil {
		return stacktrace.Propagate(err, "failed to update customer in database")
	}
	return nil
}

func (a *customerAdapter) Delete(ctx context.Context, id string) error {
	customerID, err := uuid.Parse(id)
	if err != nil {
		return stacktrace.Propagate(err, "invalid customer id format")
	}

	// Soft delete using GORM
	if err := a.db.WithContext(ctx).Delete(&model.Customer{}, customerID).Error; err != nil {
		return stacktrace.Propagate(err, "failed to delete customer from database")
	}

	return nil
}

func (a *customerAdapter) UpdateStatus(ctx context.Context, id string, status model.CustomerStatus) error {
	customerID, err := uuid.Parse(id)
	if err != nil {
		return stacktrace.Propagate(err, "invalid customer id format")
	}

	if err := a.db.WithContext(ctx).
		Model(&model.Customer{}).
		Where("id = ?", customerID).
		Update("status", status).Error; err != nil {
		return stacktrace.Propagate(err, "failed to update customer status in database")
	}

	return nil
}
