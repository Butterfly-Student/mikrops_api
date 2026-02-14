package queue

import (
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go-template/internal/model"
	mock_outbound_port "go-template/tests/mocks/port"
)

func TestQueueDomain(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mock_outbound_port.NewMockDatabasePort(ctrl)
	mockMikrotikDB := mock_outbound_port.NewMockMikrotikDatabasePort(ctrl)
	mockMikrotikPort := mock_outbound_port.NewMockMikrotikPort(ctrl)
	mockCache := mock_outbound_port.NewMockCachePort(ctrl)
	mockPppoePubSub := mock_outbound_port.NewMockPppoeCachePort(ctrl)

	mockDB.EXPECT().Mikrotik().Return(mockMikrotikDB).AnyTimes()
	mockCache.EXPECT().PppoePubSub().Return(mockPppoePubSub).AnyTimes()
	domain := NewQueueDomain(mockDB, mockCache, mockMikrotikPort)

	t.Run("CreateQueue success", func(t *testing.T) {
		routerID := uuid.New().String()
		queue := model.PppoeQueue{Name: "testqueue", Target: "192.168.1.10"}
		router := &model.MikrotikRouter{ID: uuid.MustParse(routerID)}

		mockMikrotikDB.EXPECT().FindByID(routerID).Return(router, nil).Times(1)
		mockMikrotikPort.EXPECT().CreateQueue(router, &queue).Return(nil).Times(1)

		err := domain.CreateQueue(routerID, queue)
		assert.NoError(t, err)
	})

	t.Run("CreateQueue router not found", func(t *testing.T) {
		routerID := uuid.New().String()
		queue := model.PppoeQueue{Name: "testqueue"}

		mockMikrotikDB.EXPECT().FindByID(routerID).Return(nil, errors.New("not found")).Times(1)

		err := domain.CreateQueue(routerID, queue)
		assert.Error(t, err)
	})

	t.Run("ListQueues success", func(t *testing.T) {
		routerID := uuid.New().String()
		queues := []model.PppoeQueue{{Name: "queue1"}, {Name: "queue2"}}
		router := &model.MikrotikRouter{ID: uuid.MustParse(routerID)}

		mockMikrotikDB.EXPECT().FindByID(routerID).Return(router, nil).Times(1)
		mockMikrotikPort.EXPECT().ListQueues(router).Return(queues, nil).Times(1)

		result, err := domain.ListQueues(routerID)
		assert.NoError(t, err)
		assert.Len(t, result, 2)
	})

	t.Run("UpdateQueue success", func(t *testing.T) {
		routerID := uuid.New().String()
		queue := model.PppoeQueue{Name: "updated-queue"}
		router := &model.MikrotikRouter{ID: uuid.MustParse(routerID)}

		mockMikrotikDB.EXPECT().FindByID(routerID).Return(router, nil).Times(1)
		mockMikrotikPort.EXPECT().UpdateQueue(router, &queue).Return(nil).Times(1)

		err := domain.UpdateQueue(routerID, queue)
		assert.NoError(t, err)
	})

	t.Run("DeleteQueue success", func(t *testing.T) {
		routerID := uuid.New().String()
		queueID := "queue-123"
		router := &model.MikrotikRouter{ID: uuid.MustParse(routerID)}

		mockMikrotikDB.EXPECT().FindByID(routerID).Return(router, nil).Times(1)
		mockMikrotikPort.EXPECT().DeleteQueue(router, queueID).Return(nil).Times(1)

		err := domain.DeleteQueue(routerID, queueID)
		assert.NoError(t, err)
	})
}
