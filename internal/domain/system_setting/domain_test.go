package system_setting

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

func TestGetSetting(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mock_outbound_port.NewMockDatabasePort(ctrl)
	mockSystemSettingDB := mock_outbound_port.NewMockSystemSettingDatabasePort(ctrl)

	mockDB.EXPECT().SystemSetting().Return(mockSystemSettingDB).AnyTimes()

	domain := NewSystemSettingDomain(mockDB)
	ctx := context.Background()

	t.Run("success - get setting by key", func(t *testing.T) {
		key := "company_name"
		value := "PT. Example ISP"

		expectedSetting := &model.SystemSetting{
			ID:    uuid.New(),
			Key:   key,
			Value: &value,
		}

		mockSystemSettingDB.EXPECT().
			FindByKey(key).
			Return(expectedSetting, nil).
			Times(1)

		result, err := domain.GetSetting(ctx, key)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, key, result.Key)
		assert.Equal(t, value, *result.Value)
	})

	t.Run("error - setting not found", func(t *testing.T) {
		key := "non_existent_key"

		mockSystemSettingDB.EXPECT().
			FindByKey(key).
			Return(nil, errors.New("record not found")).
			Times(1)

		result, err := domain.GetSetting(ctx, key)

		assert.Error(t, err)
		assert.Nil(t, result)
	})
}

func TestListSettings(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mock_outbound_port.NewMockDatabasePort(ctrl)
	mockSystemSettingDB := mock_outbound_port.NewMockSystemSettingDatabasePort(ctrl)

	mockDB.EXPECT().SystemSetting().Return(mockSystemSettingDB).AnyTimes()

	domain := NewSystemSettingDomain(mockDB)
	ctx := context.Background()

	t.Run("success - list all settings", func(t *testing.T) {
		companyName := "PT. Example ISP"
		companyEmail := "info@example.com"

		expectedSettings := []model.SystemSetting{
			{
				ID:    uuid.New(),
				Key:   "company_name",
				Value: &companyName,
			},
			{
				ID:    uuid.New(),
				Key:   "company_email",
				Value: &companyEmail,
			},
		}

		filter := model.SystemSettingFilter{}

		mockSystemSettingDB.EXPECT().
			FindAll().
			Return(expectedSettings, nil).
			Times(1)

		result, err := domain.ListSettings(ctx, filter)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, 2, len(result))
	})

	t.Run("success - list settings by category", func(t *testing.T) {
		category := "general"
		companyName := "PT. Example ISP"

		expectedSettings := []model.SystemSetting{
			{
				ID:       uuid.New(),
				Key:      "company_name",
				Value:    &companyName,
				Category: &category,
			},
		}

		filter := model.SystemSettingFilter{
			Category: &category,
		}

		mockSystemSettingDB.EXPECT().
			FindByCategory(category).
			Return(expectedSettings, nil).
			Times(1)

		result, err := domain.ListSettings(ctx, filter)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, 1, len(result))
	})

	t.Run("success - list public settings", func(t *testing.T) {
		isPublic := true
		companyName := "PT. Example ISP"

		expectedSettings := []model.SystemSetting{
			{
				ID:       uuid.New(),
				Key:      "company_name",
				Value:    &companyName,
				IsPublic: &isPublic,
			},
		}

		filter := model.SystemSettingFilter{
			IsPublic: &isPublic,
		}

		mockSystemSettingDB.EXPECT().
			FindPublic().
			Return(expectedSettings, nil).
			Times(1)

		result, err := domain.ListSettings(ctx, filter)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Greater(t, len(result), 0)
	})
}

func TestUpdateSetting(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mock_outbound_port.NewMockDatabasePort(ctrl)
	mockSystemSettingDB := mock_outbound_port.NewMockSystemSettingDatabasePort(ctrl)

	mockDB.EXPECT().SystemSetting().Return(mockSystemSettingDB).AnyTimes()

	domain := NewSystemSettingDomain(mockDB)
	ctx := context.Background()

	t.Run("success - update setting", func(t *testing.T) {
		key := "company_name"
		value := "PT. Updated ISP"
		userID := uuid.New().String()

		mockSystemSettingDB.EXPECT().
			UpdateByKey(key, value, userID).
			Return(nil).
			Times(1)

		err := domain.UpdateSetting(ctx, key, value, userID)

		assert.NoError(t, err)
	})

	t.Run("error - update failed", func(t *testing.T) {
		key := "non_existent_key"
		value := "value"
		userID := uuid.New().String()

		mockSystemSettingDB.EXPECT().
			UpdateByKey(key, value, userID).
			Return(errors.New("setting not found")).
			Times(1)

		err := domain.UpdateSetting(ctx, key, value, userID)

		assert.Error(t, err)
	})
}

func TestGetPublicSettings(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mock_outbound_port.NewMockDatabasePort(ctrl)
	mockSystemSettingDB := mock_outbound_port.NewMockSystemSettingDatabasePort(ctrl)

	mockDB.EXPECT().SystemSetting().Return(mockSystemSettingDB).AnyTimes()

	domain := NewSystemSettingDomain(mockDB)
	ctx := context.Background()

	t.Run("success - get public settings", func(t *testing.T) {
		isPublic := true
		companyName := "PT. Example ISP"

		expectedSettings := []model.SystemSetting{
			{
				ID:       uuid.New(),
				Key:      "company_name",
				Value:    &companyName,
				IsPublic: &isPublic,
			},
		}

		mockSystemSettingDB.EXPECT().
			FindPublic().
			Return(expectedSettings, nil).
			Times(1)

		result, err := domain.GetPublicSettings(ctx)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Greater(t, len(result), 0)
	})
}
