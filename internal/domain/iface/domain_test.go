package iface

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

func TestInterfaceDomain(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mock_outbound_port.NewMockDatabasePort(ctrl)
	mockMikrotikDB := mock_outbound_port.NewMockMikrotikDatabasePort(ctrl)
	mockCache := mock_outbound_port.NewMockCachePort(ctrl)
	mockPppoePubSub := mock_outbound_port.NewMockPppoeCachePort(ctrl)
	mockMikrotikPort := mock_outbound_port.NewMockMikrotikPort(ctrl)

	mockDB.EXPECT().Mikrotik().Return(mockMikrotikDB).AnyTimes()
	mockCache.EXPECT().PppoePubSub().Return(mockPppoePubSub).AnyTimes()
	domain := NewInterfaceDomain(mockDB, mockCache, mockMikrotikPort)

	t.Run("StartMonitoring success", func(t *testing.T) {
		ctx := context.Background()
		routerID := uuid.New().String()
		router := &model.MikrotikRouter{ID: uuid.MustParse(routerID)}
		statsCh := make(chan []model.InterfaceStats, 1)

		mockMikrotikDB.EXPECT().FindByID(routerID).Return(router, nil).Times(1)
		mockMikrotikPort.EXPECT().MonitorAllInterfaces(ctx, router).Return(statsCh, nil).Times(1)

		err := domain.StartMonitoring(ctx, routerID)
		assert.NoError(t, err)
	})

	t.Run("StartMonitoring router not found", func(t *testing.T) {
		ctx := context.Background()
		routerID := uuid.New().String()

		mockMikrotikDB.EXPECT().FindByID(routerID).Return(nil, errors.New("not found")).Times(1)

		err := domain.StartMonitoring(ctx, routerID)
		assert.Error(t, err)
	})

	t.Run("StartMonitoringByName success", func(t *testing.T) {
		ctx := context.Background()
		routerID := uuid.New().String()
		interfaceName := "ether1"
		router := &model.MikrotikRouter{ID: uuid.MustParse(routerID)}
		statsCh := make(chan model.InterfaceStats, 1)

		mockMikrotikDB.EXPECT().FindByID(routerID).Return(router, nil).Times(1)
		mockMikrotikPort.EXPECT().MonitorInterface(ctx, router, interfaceName).Return(statsCh, nil).Times(1)

		err := domain.StartMonitoringByName(ctx, routerID, interfaceName)
		assert.NoError(t, err)
	})

	t.Run("StopMonitoring success", func(t *testing.T) {
		routerID := uuid.New().String()

		err := domain.StopMonitoring(routerID)
		assert.NoError(t, err)
	})

	t.Run("StopMonitoringByName success", func(t *testing.T) {
		routerID := uuid.New().String()
		interfaceName := "ether1"

		err := domain.StopMonitoringByName(routerID, interfaceName)
		assert.NoError(t, err)
	})
}
