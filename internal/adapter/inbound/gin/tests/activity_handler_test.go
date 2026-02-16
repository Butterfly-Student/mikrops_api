package gin_inbound_adapter_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	gin_inbound_adapter "go-template/internal/adapter/inbound/gin"
	"go-template/internal/domain"
	"go-template/internal/model"
	mock_outbound_port "go-template/tests/mocks/port"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestActivityHandler_ListLogs(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mock_outbound_port.NewMockDatabasePort(ctrl)
	mockActivityDB := mock_outbound_port.NewMockActivityLogDatabasePort(ctrl)

	mockDB.EXPECT().ActivityLog().Return(mockActivityDB).AnyTimes()

	dom := domain.NewDomain(mockDB, nil, nil, nil, nil, nil, nil, nil)
	handler := gin_inbound_adapter.NewActivityHandler(dom)

	t.Run("success - list all logs", func(t *testing.T) {
		userID := uuid.New()
		logs := []model.ActivityLog{
			{
				ID:          uuid.New(),
				UserID:      &userID,
				Action:      "customer.created",
				EntityType:  "customer",
				Description: "Customer created",
			},
		}

		mockActivityDB.EXPECT().
			FindAll().
			Return(logs, nil).
			Times(1)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/activity/logs?limit=10", nil)

		handler.ListLogs(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, float64(1), response["count"])
	})

	t.Run("success - list logs with filter", func(t *testing.T) {
		userID := uuid.New()
		logs := []model.ActivityLog{
			{
				ID:     uuid.New(),
				UserID: &userID,
				Action: "payment.received",
			},
		}

		mockActivityDB.EXPECT().
			Find(gomock.Any()).
			Return(logs, nil).
			Times(1)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/activity/logs?user_id="+userID.String(), nil)

		handler.ListLogs(c)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("error - domain error", func(t *testing.T) {
		mockActivityDB.EXPECT().
			FindAll().
			Return(nil, errors.New("database error")).
			Times(1)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/activity/logs", nil)

		handler.ListLogs(c)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("error - invalid query parameters", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/activity/logs?limit=invalid", nil)

		handler.ListLogs(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestActivityHandler_GetEntityHistory(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mock_outbound_port.NewMockDatabasePort(ctrl)
	mockActivityDB := mock_outbound_port.NewMockActivityLogDatabasePort(ctrl)

	mockDB.EXPECT().ActivityLog().Return(mockActivityDB).AnyTimes()

	dom := domain.NewDomain(mockDB, nil, nil, nil, nil, nil, nil, nil)
	handler := gin_inbound_adapter.NewActivityHandler(dom)

	t.Run("success - get entity history", func(t *testing.T) {
		entityID := uuid.New()
		entityIDStr := entityID.String()
		logs := []model.ActivityLog{
			{
				ID:         uuid.New(),
				EntityType: "customer",
				EntityID:   &entityID,
				Action:     "customer.created",
			},
			{
				ID:         uuid.New(),
				EntityType: "customer",
				EntityID:   &entityID,
				Action:     "customer.updated",
			},
		}

		mockActivityDB.EXPECT().
			FindByEntity("customer", entityIDStr).
			Return(logs, nil).
			Times(1)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/activity/entity-history?entity_type=customer&entity_id="+entityIDStr, nil)

		handler.GetEntityHistory(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, float64(2), response["count"])
	})

	t.Run("error - missing entity_type", func(t *testing.T) {
		entityID := uuid.New().String()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/activity/entity-history?entity_id="+entityID, nil)

		handler.GetEntityHistory(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("error - missing entity_id", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/activity/entity-history?entity_type=customer", nil)

		handler.GetEntityHistory(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("error - domain error", func(t *testing.T) {
		entityID := uuid.New().String()

		mockActivityDB.EXPECT().
			FindByEntity("customer", entityID).
			Return(nil, errors.New("database error")).
			Times(1)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/activity/entity-history?entity_type=customer&entity_id="+entityID, nil)

		handler.GetEntityHistory(c)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

func TestActivityHandler_GetUserLogs(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mock_outbound_port.NewMockDatabasePort(ctrl)
	mockActivityDB := mock_outbound_port.NewMockActivityLogDatabasePort(ctrl)

	mockDB.EXPECT().ActivityLog().Return(mockActivityDB).AnyTimes()

	dom := domain.NewDomain(mockDB, nil, nil, nil, nil, nil, nil, nil)
	handler := gin_inbound_adapter.NewActivityHandler(dom)

	t.Run("success - get user logs", func(t *testing.T) {
		userID := uuid.New()
		logs := []model.ActivityLog{
			{
				ID:     uuid.New(),
				UserID: &userID,
				Action: "customer.created",
			},
			{
				ID:     uuid.New(),
				UserID: &userID,
				Action: "payment.received",
			},
		}

		mockActivityDB.EXPECT().
			Find(gomock.Any()).
			Return(logs, nil).
			Times(1)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/activity/users/"+userID.String()+"/logs", nil)
		c.Params = gin.Params{{Key: "user_id", Value: userID.String()}}

		handler.GetUserLogs(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, float64(2), response["count"])
	})

	t.Run("error - invalid user ID", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/activity/users/invalid-uuid/logs", nil)
		c.Params = gin.Params{{Key: "user_id", Value: "invalid-uuid"}}

		handler.GetUserLogs(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("error - domain error", func(t *testing.T) {
		userID := uuid.New()

		mockActivityDB.EXPECT().
			Find(gomock.Any()).
			Return(nil, errors.New("database error")).
			Times(1)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/activity/users/"+userID.String()+"/logs", nil)
		c.Params = gin.Params{{Key: "user_id", Value: userID.String()}}

		handler.GetUserLogs(c)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}
