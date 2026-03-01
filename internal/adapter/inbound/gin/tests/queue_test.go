package gin_adapter_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/casbin/casbin/v3"
	casbinmodel "github.com/casbin/casbin/v3/model"
	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	. "github.com/smartystreets/goconvey/convey"

	gin_inbound_adapter "go-template/internal/adapter/inbound/gin"
	"go-template/internal/domain"
	"go-template/internal/model"
	mock_outbound_port "go-template/tests/mocks/port"
)

func TestQueueAdapter(t *testing.T) {
	Convey("Test Queue HTTP Adapter", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		mockDatabasePort := mock_outbound_port.NewMockDatabasePort(mockCtrl)
		mockMikrotikDBPort := mock_outbound_port.NewMockMikrotikDatabasePort(mockCtrl)
		mockMessagePort := mock_outbound_port.NewMockMessagePort(mockCtrl)
		mockCachePort := mock_outbound_port.NewMockCachePort(mockCtrl)
		mockWorkflowPort := mock_outbound_port.NewMockWorkflowPort(mockCtrl)
		mockMikrotikPort := mock_outbound_port.NewMockMikrotikPort(mockCtrl)

		mockDatabasePort.EXPECT().Mikrotik().Return(mockMikrotikDBPort).AnyTimes()

		m, _ := casbinmodel.NewModelFromString(`
[request_definition]
r = sub, obj, act

[policy_definition]
p = sub, obj, act

[role_definition]
g = _, _

[policy_effect]
e = some(where (p.eft == allow))

[matchers]
m = g(r.sub, p.sub) && r.obj == p.obj && r.act == p.act
`)
		enforcer, _ := casbin.NewEnforcer(m)

		dom := domain.NewDomain(mockDatabasePort, mockMessagePort, mockCachePort, mockWorkflowPort, mockMikrotikPort, nil, enforcer)
		adapter := gin_inbound_adapter.NewAdapter(dom)

		gin.SetMode(gin.TestMode)
		router := gin.New()

		authMiddleware := func(c *gin.Context) {
			c.Set("userID", uint(1))
			c.Next()
		}

		queue := router.Group("/queues")
		queue.Use(authMiddleware)
		{
			queue.POST("", adapter.Queue().CreateQueue)
			queue.GET("", adapter.Queue().ListQueues)
		}

		routerID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
		routerModel := &model.MikrotikRouter{ID: routerID, Name: "TestRouter", Address: "192.168.88.1:8728"}

		Convey("CreateQueue", func() {
			queueItem := model.PppoeQueue{
				Name:     "testqueue",
				Target:   "192.168.1.10",
				MaxLimit: "10M/10M",
			}

			Convey("Success", func() {
				mockMikrotikDBPort.EXPECT().FindByID(routerID.String()).Return(routerModel, nil).Times(1)
				mockMikrotikPort.EXPECT().CreateQueue(routerModel, gomock.Any()).Return(nil).Times(1)

				body, _ := json.Marshal(queueItem)
				req := httptest.NewRequest(http.MethodPost, "/queues?router_id="+routerID.String(), bytes.NewReader(body))
				req.Header.Set("Content-Type", "application/json")

				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				So(w.Code, ShouldEqual, http.StatusCreated)
			})

			Convey("Router Not Found", func() {
				mockMikrotikDBPort.EXPECT().FindByID(routerID.String()).Return(nil, errors.New("not found")).Times(1)

				body, _ := json.Marshal(queueItem)
				req := httptest.NewRequest(http.MethodPost, "/queues?router_id="+routerID.String(), bytes.NewReader(body))
				req.Header.Set("Content-Type", "application/json")

				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				So(w.Code, ShouldEqual, http.StatusInternalServerError)
			})
		})
	})
}
