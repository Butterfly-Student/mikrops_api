package bandwidth_profile

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

func TestCreateProfile(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mock_outbound_port.NewMockDatabasePort(ctrl)
	mockMikrotikPort := mock_outbound_port.NewMockMikrotikPort(ctrl)

	domain := NewBandwidthProfileDomain(mockDB, mockMikrotikPort)
	ctx := context.Background()

	t.Run("success - create profile", func(t *testing.T) {
		profileCode := "BW-100MBPS"
		name := "100 Mbps Package"
		downloadSpeed := int64(100000000) // 100 Mbps in bps
		uploadSpeed := int64(100000000)

		input := model.BandwidthProfileInput{
			ProfileCode:   &profileCode,
			Name:          name,
			DownloadSpeed: downloadSpeed,
			UploadSpeed:   uploadSpeed,
		}

		mockDB.EXPECT().
			DoInTransaction(gomock.Any()).
			DoAndReturn(func(txFunc interface{}) (interface{}, error) {
				return &model.BandwidthProfile{
					ID:            uuid.New(),
					ProfileCode:   profileCode,
					Name:          name,
					DownloadSpeed: downloadSpeed,
					UploadSpeed:   uploadSpeed,
				}, nil
			}).Times(1)

		result, err := domain.CreateProfile(ctx, input)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, profileCode, result.ProfileCode)
	})

	t.Run("error - missing profile code", func(t *testing.T) {
		input := model.BandwidthProfileInput{}

		result, err := domain.CreateProfile(ctx, input)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "profile_code is required")
	})
}

func TestGetProfile(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mock_outbound_port.NewMockDatabasePort(ctrl)
	mockMikrotikPort := mock_outbound_port.NewMockMikrotikPort(ctrl)
	mockBandwidthProfileDB := mock_outbound_port.NewMockBandwidthProfileDatabasePort(ctrl)

	mockDB.EXPECT().BandwidthProfile().Return(mockBandwidthProfileDB).AnyTimes()

	domain := NewBandwidthProfileDomain(mockDB, mockMikrotikPort)
	ctx := context.Background()

	t.Run("success - get profile", func(t *testing.T) {
		profileID := uuid.New()
		profileCode := "BW-100MBPS"
		name := "100 Mbps Package"

		expectedProfile := &model.BandwidthProfile{
			ID:          profileID,
			ProfileCode: profileCode,
			Name:        name,
		}

		mockBandwidthProfileDB.EXPECT().
			FindByID(profileID.String()).
			Return(expectedProfile, nil).
			Times(1)

		result, err := domain.GetProfile(ctx, profileID.String())

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, profileID, result.ID)
		assert.Equal(t, profileCode, result.ProfileCode)
	})

	t.Run("error - profile not found", func(t *testing.T) {
		profileID := uuid.New()

		mockBandwidthProfileDB.EXPECT().
			FindByID(profileID.String()).
			Return(nil, errors.New("record not found")).
			Times(1)

		result, err := domain.GetProfile(ctx, profileID.String())

		assert.Error(t, err)
		assert.Nil(t, result)
	})
}

func TestGetIsolatedProfile(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mock_outbound_port.NewMockDatabasePort(ctrl)
	mockMikrotikPort := mock_outbound_port.NewMockMikrotikPort(ctrl)
	mockBandwidthProfileDB := mock_outbound_port.NewMockBandwidthProfileDatabasePort(ctrl)

	mockDB.EXPECT().BandwidthProfile().Return(mockBandwidthProfileDB).AnyTimes()

	domain := NewBandwidthProfileDomain(mockDB, mockMikrotikPort)
	ctx := context.Background()

	t.Run("success - get isolated profile", func(t *testing.T) {
		category := "isolated"
		expectedProfile := &model.BandwidthProfile{
			ID:          uuid.New(),
			ProfileCode: "ISOLATED",
			Category:    category,
		}

		mockBandwidthProfileDB.EXPECT().
			FindByCategory(category).
			Return([]model.BandwidthProfile{*expectedProfile}, nil).
			Times(1)

		result, err := domain.GetIsolatedProfile(ctx)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "ISOLATED", result.ProfileCode)
	})

	t.Run("error - no isolated profile found", func(t *testing.T) {
		category := "isolated"

		mockBandwidthProfileDB.EXPECT().
			FindByCategory(category).
			Return([]model.BandwidthProfile{}, nil).
			Times(1)

		result, err := domain.GetIsolatedProfile(ctx)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "isolated profile not found")
	})
}
