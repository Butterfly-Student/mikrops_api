package activity

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"

	"go-template/internal/model"
	outbound_port "go-template/internal/port/outbound"
)

type ActivityDomain interface {
	LogActivity(ctx context.Context, input model.ActivityLogInput) error
	ListLogs(ctx context.Context, filter model.ActivityLogFilter) ([]model.ActivityLog, error)
	GetEntityHistory(ctx context.Context, entityType string, entityID string) ([]model.ActivityLog, error)
}

type domain struct {
	dbPort outbound_port.DatabasePort
}

func NewActivityDomain(dbPort outbound_port.DatabasePort) ActivityDomain {
	return &domain{dbPort: dbPort}
}

func (d *domain) LogActivity(ctx context.Context, input model.ActivityLogInput) error {
	activityLog := &model.ActivityLog{
		UserID:      input.UserID,
		Action:      input.Action,
		EntityType:  input.EntityType,
		EntityID:    input.EntityID,
		Description: input.Description,
		OldValues:   input.OldValues,
		NewValues:   input.NewValues,
		IPAddress:   input.IPAddress,
		UserAgent:   input.UserAgent,
	}

	err := d.dbPort.ActivityLog().Create(activityLog)
	if err != nil {
		return err
	}

	return nil
}

func (d *domain) ListLogs(ctx context.Context, filter model.ActivityLogFilter) ([]model.ActivityLog, error) {
	if filter.IsEmpty() {
		return d.dbPort.ActivityLog().FindAll()
	}
	return d.dbPort.ActivityLog().Find(filter)
}

func (d *domain) GetEntityHistory(ctx context.Context, entityType string, entityID string) ([]model.ActivityLog, error) {
	return d.dbPort.ActivityLog().FindByEntity(entityType, entityID)
}

func (d *domain) LogAction(ctx context.Context, userID *uint, action string, entityType string, entityID *uuid.UUID, description string, oldValue interface{}, newValue interface{}) error {
	var oldValuesJSON *string
	if oldValue != nil {
		data, err := json.Marshal(oldValue)
		if err == nil {
			s := string(data)
			oldValuesJSON = &s
		}
	}

	var newValuesJSON *string
	if newValue != nil {
		data, err := json.Marshal(newValue)
		if err == nil {
			s := string(data)
			newValuesJSON = &s
		}
	}

	input := model.ActivityLogInput{
		UserID:      userID,
		Action:      action,
		EntityType:  entityType,
		EntityID:    entityID,
		Description: description,
		OldValues:   oldValuesJSON,
		NewValues:   newValuesJSON,
	}

	return d.LogActivity(ctx, input)
}
