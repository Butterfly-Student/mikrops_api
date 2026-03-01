package postgres_outbound_adapter

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/palantir/stacktrace"
	"gorm.io/gorm"

	"go-template/internal/model"
	outbound_port "go-template/internal/port/outbound"
)

type registrationAdapter struct {
	db *gorm.DB
}

func NewRegistrationAdapter(db *gorm.DB) outbound_port.RegistrationDatabasePort {
	return &registrationAdapter{db: db}
}

func (a *registrationAdapter) Create(ctx context.Context, registration *model.CustomerRegistration) error {
	if err := a.db.WithContext(ctx).Create(registration).Error; err != nil {
		return stacktrace.Propagate(err, "failed to create customer registration in database")
	}
	return nil
}

func (a *registrationAdapter) FindByID(ctx context.Context, id string) (*model.CustomerRegistration, error) {
	var registration model.CustomerRegistration
	registrationID, err := uuid.Parse(id)
	if err != nil {
		return nil, stacktrace.Propagate(err, "invalid registration id format")
	}

	if err := a.db.WithContext(ctx).
		Preload("BandwidthProfile").
		Preload("Customer").
		Where("id = ?", registrationID).
		First(&registration).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, stacktrace.NewError("registration not found")
		}
		return nil, stacktrace.Propagate(err, "failed to find registration by id")
	}

	return &registration, nil
}

func (a *registrationAdapter) FindAll(ctx context.Context, filter *model.RegistrationFilter) ([]model.CustomerRegistration, error) {
	var registrations []model.CustomerRegistration

	query := a.db.WithContext(ctx).
		Preload("BandwidthProfile").
		Order("created_at DESC")

	if filter != nil {
		if filter.Status != nil {
			query = query.Where("status = ?", *filter.Status)
		}
		if filter.BandwidthProfileID != nil {
			query = query.Where("bandwidth_profile_id = ?", *filter.BandwidthProfileID)
		}
		if filter.Search != nil && *filter.Search != "" {
			search := "%" + *filter.Search + "%"
			query = query.Where("full_name ILIKE ? OR phone ILIKE ? OR email ILIKE ?", search, search, search)
		}
	}

	if err := query.Find(&registrations).Error; err != nil {
		return nil, stacktrace.Propagate(err, "failed to list registrations")
	}

	return registrations, nil
}

func (a *registrationAdapter) SetApproved(ctx context.Context, id string, approverID string, customerID string) error {
	registrationID, err := uuid.Parse(id)
	if err != nil {
		return stacktrace.Propagate(err, "invalid registration id")
	}

	customerUUID, err := uuid.Parse(customerID)
	if err != nil {
		return stacktrace.Propagate(err, "invalid customer id")
	}

	approverUUID, err := uuid.Parse(approverID)
	if err != nil {
		return stacktrace.Propagate(err, "invalid approver id")
	}

	now := time.Now()
	updates := map[string]interface{}{
		"status":      model.RegistrationStatusApproved,
		"approved_by": approverUUID,
		"approved_at": now,
		"customer_id": customerUUID,
		"updated_at":  now,
	}

	if err := a.db.WithContext(ctx).
		Model(&model.CustomerRegistration{}).
		Where("id = ?", registrationID).
		Updates(updates).Error; err != nil {
		return stacktrace.Propagate(err, "failed to set registration as approved")
	}

	return nil
}

func (a *registrationAdapter) SetRejected(ctx context.Context, id string, approverID string, reason string) error {
	registrationID, err := uuid.Parse(id)
	if err != nil {
		return stacktrace.Propagate(err, "invalid registration id")
	}

	approverUUID, err := uuid.Parse(approverID)
	if err != nil {
		return stacktrace.Propagate(err, "invalid approver id")
	}

	now := time.Now()
	updates := map[string]interface{}{
		"status":           model.RegistrationStatusRejected,
		"approved_by":      approverUUID,
		"rejection_reason": reason,
		"updated_at":       now,
	}

	if err := a.db.WithContext(ctx).
		Model(&model.CustomerRegistration{}).
		Where("id = ?", registrationID).
		Updates(updates).Error; err != nil {
		return stacktrace.Propagate(err, "failed to set registration as rejected")
	}

	return nil
}
