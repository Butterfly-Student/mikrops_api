package ippool

import (
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go-template/internal/model"
	mock_outbound_port "go-template/tests/mocks/port"
)

func TestIpPoolDomain(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mock_outbound_port.NewMockDatabasePort(ctrl)
	mockMikrotikDB := mock_outbound_port.NewMockMikrotikDatabasePort(ctrl)
	mockMikrotikPort := mock_outbound_port.NewMockMikrotikPort(ctrl)

	mockDB.EXPECT().Mikrotik().Return(mockMikrotikDB).AnyTimes()
	domain := NewIpPoolDomain(mockDB, mockMikrotikPort)

	t.Run("CreateIpPool success", func(t *testing.T) {
		routerID := uuid.New().String()
		pool := model.IpPool{Name: "test-pool", Ranges: "192.168.1.0/24"}
		router := &model.MikrotikRouter{ID: uuid.MustParse(routerID)}

		mockMikrotikDB.EXPECT().FindByID(routerID).Return(router, nil).Times(1)
		mockMikrotikPort.EXPECT().CreateIpPool(router, &pool).Return(nil).Times(1)

		err := domain.CreateIpPool(routerID, pool)
		assert.NoError(t, err)
	})

	t.Run("CreateIpPool router not found", func(t *testing.T) {
		routerID := uuid.New().String()
		pool := model.IpPool{Name: "test-pool"}

		mockMikrotikDB.EXPECT().FindByID(routerID).Return(nil, errors.New("not found")).Times(1)

		err := domain.CreateIpPool(routerID, pool)
		assert.Error(t, err)
	})

	t.Run("ListIpPools success", func(t *testing.T) {
		routerID := uuid.New().String()
		pools := []model.IpPool{{Name: "pool1"}, {Name: "pool2"}}
		router := &model.MikrotikRouter{ID: uuid.MustParse(routerID)}

		mockMikrotikDB.EXPECT().FindByID(routerID).Return(router, nil).Times(1)
		mockMikrotikPort.EXPECT().ListIpPools(router).Return(pools, nil).Times(1)

		result, err := domain.ListIpPools(routerID)
		assert.NoError(t, err)
		assert.Len(t, result, 2)
	})

	t.Run("UpdateIpPool success", func(t *testing.T) {
		routerID := uuid.New().String()
		pool := model.IpPool{Name: "updated-pool"}
		router := &model.MikrotikRouter{ID: uuid.MustParse(routerID)}

		mockMikrotikDB.EXPECT().FindByID(routerID).Return(router, nil).Times(1)
		mockMikrotikPort.EXPECT().UpdateIpPool(router, &pool).Return(nil).Times(1)

		err := domain.UpdateIpPool(routerID, pool)
		assert.NoError(t, err)
	})

	t.Run("DeleteIpPool success", func(t *testing.T) {
		routerID := uuid.New().String()
		poolID := "pool-123"
		router := &model.MikrotikRouter{ID: uuid.MustParse(routerID)}

		mockMikrotikDB.EXPECT().FindByID(routerID).Return(router, nil).Times(1)
		mockMikrotikPort.EXPECT().DeleteIpPool(router, poolID).Return(nil).Times(1)

		err := domain.DeleteIpPool(routerID, poolID)
		assert.NoError(t, err)
	})
}
