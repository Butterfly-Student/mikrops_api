package activity

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"go-template/internal/model"
	mock_outbound_port "go-template/tests/mocks/port"
)

func TestLogActivity(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mock_outbound_port.NewMockDatabasePort(ctrl)
	mockActivityDB := mock_outbound_port.NewMockActivityLogDatabasePort(ctrl)

	mockDB.EXPECT().ActivityLog().Return(mockActivityDB).AnyTimes()

	domain := NewActivityDomain(mockDB)
	ctx := context.Background()

	t.Run("success - log activity", func(t *testing.T) {
		userID := uint(1)
		action := "customer.created"
		entityType := "customer"
		entityID := uuid.New()
		description := "Customer created successfully"
		ipAddress := "192.168.1.1"
		userAgent := "Mozilla/5.0"

		input := model.ActivityLogInput{
			UserID:      &userID,
			Action:      action,
			EntityType:  entityType,
			EntityID:    &entityID,
			Description: description,
			IPAddress:   &ipAddress,
			UserAgent:   &userAgent,
		}

		mockActivityDB.EXPECT().
			Create(gomock.Any()).
			Return(nil).
			Times(1)

		err := domain.LogActivity(ctx, input)

		assert.NoError(t, err)
	})

	t.Run("error - database error", func(t *testing.T) {
		userID := uint(1)

		input := model.ActivityLogInput{
			UserID:      &userID,
			Action:      "customer.created",
			EntityType:  "customer",
			Description: "Test",
		}

		mockActivityDB.EXPECT().
			Create(gomock.Any()).
			Return(errors.New("database error")).
			Times(1)

		err := domain.LogActivity(ctx, input)

		assert.Error(t, err)
	})
}

func TestListLogs(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mock_outbound_port.NewMockDatabasePort(ctrl)
	mockActivityDB := mock_outbound_port.NewMockActivityLogDatabasePort(ctrl)

	mockDB.EXPECT().ActivityLog().Return(mockActivityDB).AnyTimes()

	domain := NewActivityDomain(mockDB)
	ctx := context.Background()

	t.Run("success - list all logs", func(t *testing.T) {
		expectedLogs := []model.ActivityLog{
			{
				ID:          uuid.New(),
				Action:      "customer.created",
				EntityType:  "customer",
				Description: "Customer created",
			},
			{
				ID:          uuid.New(),
				Action:      "payment.received",
				EntityType:  "payment",
				Description: "Payment received",
			},
		}

		mockActivityDB.EXPECT().
			FindAll().
			Return(expectedLogs, nil).
			Times(1)

		result, err := domain.ListLogs(ctx, model.ActivityLogFilter{})

		assert.NoError(t, err)
		assert.Len(t, result, 2)
	})

	t.Run("success - list logs with filter", func(t *testing.T) {
		filter := model.ActivityLogFilter{
			UserIDs: []uint{1},
			Limit:   10,
		}

		userID := uint(1)
		expectedLogs := []model.ActivityLog{
			{
				ID:     uuid.New(),
				UserID: &userID,
				Action: "customer.created",
			},
		}

		mockActivityDB.EXPECT().
			Find(filter).
			Return(expectedLogs, nil).
			Times(1)

		result, err := domain.ListLogs(ctx, filter)

		assert.NoError(t, err)
		assert.Len(t, result, 1)
	})

	t.Run("error - database error", func(t *testing.T) {
		mockActivityDB.EXPECT().
			FindAll().
			Return(nil, errors.New("database error")).
			Times(1)

		result, err := domain.ListLogs(ctx, model.ActivityLogFilter{})

		assert.Error(t, err)
		assert.Nil(t, result)
	})
}

func TestGetEntityHistory(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mock_outbound_port.NewMockDatabasePort(ctrl)
	mockActivityDB := mock_outbound_port.NewMockActivityLogDatabasePort(ctrl)

	mockDB.EXPECT().ActivityLog().Return(mockActivityDB).AnyTimes()

	domain := NewActivityDomain(mockDB)
	ctx := context.Background()

	t.Run("success - get entity history", func(t *testing.T) {
		entityType := "customer"
		entityID := uuid.New()
		entityIDStr := entityID.String()

		expectedLogs := []model.ActivityLog{
			{
				ID:         uuid.New(),
				EntityType: entityType,
				EntityID:   &entityID,
				Action:     "customer.created",
			},
			{
				ID:         uuid.New(),
				EntityType: entityType,
				EntityID:   &entityID,
				Action:     "customer.updated",
			},
		}

		mockActivityDB.EXPECT().
			FindByEntity(entityType, entityIDStr).
			Return(expectedLogs, nil).
			Times(1)

		result, err := domain.GetEntityHistory(ctx, entityType, entityIDStr)

		assert.NoError(t, err)
		assert.Len(t, result, 2)
	})

	t.Run("error - database error", func(t *testing.T) {
		entityType := "customer"
		entityID := uuid.New()
		entityIDStr := entityID.String()

		mockActivityDB.EXPECT().
			FindByEntity(entityType, entityIDStr).
			Return(nil, errors.New("database error")).
			Times(1)

		result, err := domain.GetEntityHistory(ctx, entityType, entityIDStr)

		assert.Error(t, err)
		assert.Nil(t, result)
	})
}
