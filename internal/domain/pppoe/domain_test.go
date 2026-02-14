package pppoe

import (
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go-template/internal/model"
	mock_outbound_port "go-template/tests/mocks/port"
)

func TestPppoeDomain(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mock_outbound_port.NewMockDatabasePort(ctrl)
	mockMikrotikDB := mock_outbound_port.NewMockMikrotikDatabasePort(ctrl)
	mockMikrotikPort := mock_outbound_port.NewMockMikrotikPort(ctrl)
	mockCache := mock_outbound_port.NewMockCachePort(ctrl)
	mockPppoePubSub := mock_outbound_port.NewMockPppoeCachePort(ctrl)

	mockDB.EXPECT().Mikrotik().Return(mockMikrotikDB).AnyTimes()
	mockCache.EXPECT().PppoePubSub().Return(mockPppoePubSub).AnyTimes()
	domain := NewPppoeDomain(mockDB, mockCache, mockMikrotikPort)

	t.Run("CreateSecret success", func(t *testing.T) {
		routerID := uuid.New().String()
		secret := model.PppoeSecret{Name: "testuser", Password: "password"}
		router := &model.MikrotikRouter{ID: uuid.MustParse(routerID)}

		mockMikrotikDB.EXPECT().FindByID(routerID).Return(router, nil).Times(1)
		mockMikrotikPort.EXPECT().CreateSecret(router, &secret).Return(nil).Times(1)

		err := domain.CreateSecret(routerID, secret)
		assert.NoError(t, err)
	})

	t.Run("CreateSecret router not found", func(t *testing.T) {
		routerID := uuid.New().String()
		secret := model.PppoeSecret{Name: "testuser", Password: "password"}

		mockMikrotikDB.EXPECT().FindByID(routerID).Return(nil, errors.New("not found")).Times(1)

		err := domain.CreateSecret(routerID, secret)
		assert.Error(t, err)
	})

	t.Run("ListSecrets success", func(t *testing.T) {
		routerID := uuid.New().String()
		secrets := []model.PppoeSecret{{Name: "user1"}, {Name: "user2"}}
		router := &model.MikrotikRouter{ID: uuid.MustParse(routerID)}

		mockMikrotikDB.EXPECT().FindByID(routerID).Return(router, nil).Times(1)
		mockMikrotikPort.EXPECT().ListSecrets(router).Return(secrets, nil).Times(1)

		result, err := domain.ListSecrets(routerID)
		assert.NoError(t, err)
		assert.Len(t, result, 2)
	})

	t.Run("UpdateSecret success", func(t *testing.T) {
		routerID := uuid.New().String()
		secret := model.PppoeSecret{Name: "updated-user"}
		router := &model.MikrotikRouter{ID: uuid.MustParse(routerID)}

		mockMikrotikDB.EXPECT().FindByID(routerID).Return(router, nil).Times(1)
		mockMikrotikPort.EXPECT().UpdateSecret(router, &secret).Return(nil).Times(1)

		err := domain.UpdateSecret(routerID, secret)
		assert.NoError(t, err)
	})

	t.Run("DeleteSecret success", func(t *testing.T) {
		routerID := uuid.New().String()
		secretID := "secret-123"
		router := &model.MikrotikRouter{ID: uuid.MustParse(routerID)}

		mockMikrotikDB.EXPECT().FindByID(routerID).Return(router, nil).Times(1)
		mockMikrotikPort.EXPECT().DeleteSecret(router, secretID).Return(nil).Times(1)

		err := domain.DeleteSecret(routerID, secretID)
		assert.NoError(t, err)
	})
}
