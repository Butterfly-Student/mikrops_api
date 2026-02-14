package ping

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

func TestPingDomain(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mock_outbound_port.NewMockDatabasePort(ctrl)
	mockMikrotikDB := mock_outbound_port.NewMockMikrotikDatabasePort(ctrl)
	mockCache := mock_outbound_port.NewMockCachePort(ctrl)
	mockMikrotikPort := mock_outbound_port.NewMockMikrotikPort(ctrl)

	mockDB.EXPECT().Mikrotik().Return(mockMikrotikDB).AnyTimes()
	domain := NewPingDomain(mockDB, mockCache, mockMikrotikPort)

	t.Run("StartPing success", func(t *testing.T) {
		ctx := context.Background()
		routerID := uuid.New().String()
		req := model.PingRequest{Address: "192.168.1.1"}
		router := &model.MikrotikRouter{
			ID:       uuid.MustParse(routerID),
			Name:     "TestRouter",
			Address:  "192.168.88.1",
			Username: "admin",
			Password: "admin",
		}

		mockMikrotikDB.EXPECT().FindByID(routerID).Return(router, nil).Times(1)
		mockMikrotikPort.EXPECT().Ping(ctx, router, req).Return(nil, nil).Times(1)

		err := domain.StartPing(ctx, routerID, req)
		assert.NoError(t, err)
	})

	t.Run("StartPing router not found", func(t *testing.T) {
		ctx := context.Background()
		routerID := uuid.New().String()
		req := model.PingRequest{Address: "192.168.1.1"}

		mockMikrotikDB.EXPECT().FindByID(routerID).Return(nil, errors.New("not found")).Times(1)

		err := domain.StartPing(ctx, routerID, req)
		assert.Error(t, err)
	})

	t.Run("StopPing success", func(t *testing.T) {
		routerID := uuid.New().String()

		err := domain.StopPing(routerID, "192.168.1.1")
		assert.NoError(t, err)
	})
}
