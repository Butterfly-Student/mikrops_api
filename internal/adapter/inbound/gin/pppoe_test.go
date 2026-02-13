package gin_inbound_adapter_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/casbin/casbin/v2"
	casbinmodel "github.com/casbin/casbin/v2/model"
	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	. "github.com/smartystreets/goconvey/convey"

	gin_inbound_adapter "go-template/internal/adapter/inbound/gin"
	"go-template/internal/domain"
	"go-template/internal/model"
	mock_outbound_port "go-template/tests/mocks/port"
)

func TestPppoeAdapter(t *testing.T) {
	Convey("Test PPPoE HTTP Adapter", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		mockDatabasePort := mock_outbound_port.NewMockDatabasePort(mockCtrl)
		mockMikrotikDBPort := mock_outbound_port.NewMockMikrotikDatabasePort(mockCtrl)
		mockMessagePort := mock_outbound_port.NewMockMessagePort(mockCtrl)
		mockCachePort := mock_outbound_port.NewMockCachePort(mockCtrl)
		mockWorkflowPort := mock_outbound_port.NewMockWorkflowPort(mockCtrl)
		mockMikrotikPort := mock_outbound_port.NewMockMikrotikPort(mockCtrl)

		// Fix: Use EXPECT() on the mock interface, but DatabasePort interface has been updated
		// and the mock needs to reflect that.
		mockDatabasePort.EXPECT().Mikrotik().Return(mockMikrotikDBPort).AnyTimes()

		// Setup Casbin
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

		dom := domain.NewDomain(mockDatabasePort, mockMessagePort, mockCachePort, mockWorkflowPort, mockMikrotikPort, enforcer)
		adapter := gin_inbound_adapter.NewAdapter(dom)

		gin.SetMode(gin.TestMode)
		router := gin.New()

		// Setup middleware mock
		authMiddleware := func(c *gin.Context) {
			c.Set("userID", uint(1))
			c.Next()
		}

		pppoe := router.Group("/pppoe")
		pppoe.Use(authMiddleware)
		{
			pppoe.POST("/secrets", adapter.Pppoe().CreateSecret)
			pppoe.GET("/secrets/:id", adapter.Pppoe().GetSecret)
			pppoe.GET("/sessions/inactive", adapter.Pppoe().ListInactiveSessions)
		}

		routerID := uint(1)
		routerModel := &model.MikrotikRouter{ID: 1, Name: "TestRouter", Address: "192.168.88.1:8728"}

		Convey("CreateSecret", func() {
			secret := model.PppoeSecret{
				Name:     "testuser",
				Password: "password",
				Service:  "pppoe",
			}

			Convey("Success", func() {
				mockMikrotikDBPort.EXPECT().FindByID(routerID).Return(routerModel, nil).Times(1)
				mockMikrotikPort.EXPECT().CreateSecret(routerModel, gomock.Any()).Return(nil).Times(1)

				body, _ := json.Marshal(secret)
				req := httptest.NewRequest(http.MethodPost, "/pppoe/secrets?router_id=1", bytes.NewReader(body))
				req.Header.Set("Content-Type", "application/json")

				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				So(w.Code, ShouldEqual, http.StatusCreated)
			})

			Convey("Router Not Found", func() {
				mockMikrotikDBPort.EXPECT().FindByID(routerID).Return(nil, errors.New("not found")).Times(1)

				body, _ := json.Marshal(secret)
				req := httptest.NewRequest(http.MethodPost, "/pppoe/secrets?router_id=1", bytes.NewReader(body))
				req.Header.Set("Content-Type", "application/json")

				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				So(w.Code, ShouldEqual, http.StatusInternalServerError)
			})
		})

		Convey("ListInactiveSessions", func() {
			Convey("Success", func() {
				// Mock domain logic indirectly via port calls
				// 1. Get Router
				mockMikrotikDBPort.EXPECT().FindByID(routerID).Return(routerModel, nil).Times(1)

				// 2. List Secrets
				secrets := []model.PppoeSecret{
					{Name: "active_user", Disabled: false},
					{Name: "inactive_user", Disabled: false},
				}
				mockMikrotikPort.EXPECT().ListSecrets(routerModel).Return(secrets, nil).Times(1)

				// 3. List Active Sessions
				activeSessions := []model.PppoeActive{
					{Name: "active_user"},
				}
				mockMikrotikPort.EXPECT().ListActiveSessions(routerModel).Return(activeSessions, nil).Times(1)

				req := httptest.NewRequest(http.MethodGet, "/pppoe/sessions/inactive?router_id=1", nil)
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				So(w.Code, ShouldEqual, http.StatusOK)

				var res []model.PppoeSecret
				json.Unmarshal(w.Body.Bytes(), &res)
				So(len(res), ShouldEqual, 1)
				So(res[0].Name, ShouldEqual, "inactive_user")
			})
		})
	})
}
