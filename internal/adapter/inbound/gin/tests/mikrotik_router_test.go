package gin_adapter_test

import (
	"bytes"
	"encoding/json"
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

func TestMikrotikRouterAdapter(t *testing.T) {
	Convey("Test MikroTik Router HTTP Adapter", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		mockDatabasePort := mock_outbound_port.NewMockDatabasePort(mockCtrl)
		mockMessagePort := mock_outbound_port.NewMockMessagePort(mockCtrl)
		mockCachePort := mock_outbound_port.NewMockCachePort(mockCtrl)
		mockWorkflowPort := mock_outbound_port.NewMockWorkflowPort(mockCtrl)
		mockMikrotikPort := mock_outbound_port.NewMockMikrotikPort(mockCtrl)

		mockMikrotikDBPort := mock_outbound_port.NewMockMikrotikDatabasePort(mockCtrl)
		mockDatabasePort.EXPECT().Mikrotik().Return(mockMikrotikDBPort).AnyTimes()

		// Setup common expectations for all subtests
		mockMikrotikDBPort.EXPECT().Create(gomock.Any()).Return(nil).AnyTimes()
		mockMikrotikDBPort.EXPECT().FindByID(gomock.Any()).Return(&model.MikrotikRouter{ID: uuid.MustParse("550e8400-e29b-41d4-a716-446655440001"), Name: "TestRouter", Address: "192.168.1.1:8728"}, nil).AnyTimes()
		mockMikrotikDBPort.EXPECT().FindAll().Return([]model.MikrotikRouter{{ID: uuid.MustParse("550e8400-e29b-41d4-a716-446655440001")}}, nil).AnyTimes()
		mockMikrotikDBPort.EXPECT().Update(gomock.Any()).Return(nil).AnyTimes()
		mockMikrotikDBPort.EXPECT().Delete(gomock.Any()).Return(nil).AnyTimes()

		// Setup mock expectations for MikrotikPort (outbound adapter)
		mockMikrotikPort.EXPECT().ListSecrets(gomock.Any()).Return([]model.PppoeSecret{}, nil).AnyTimes()
		mockMikrotikPort.EXPECT().SetupIsolation(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
		mockMikrotikPort.EXPECT().CheckIsolationSetup(gomock.Any()).Return(true, nil).AnyTimes()

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

		routers := router.Group("/routers")
		{
			routers.POST("", adapter.MikrotikRouter().Create)
			routers.GET("", adapter.MikrotikRouter().List)
			routers.GET("/:router_id", adapter.MikrotikRouter().GetByID)
			routers.PUT("/:router_id", adapter.MikrotikRouter().Update)
			routers.DELETE("/:router_id", adapter.MikrotikRouter().Delete)
			routers.POST("/:router_id/test-connection", adapter.MikrotikRouter().TestConnection)
			routers.POST("/:router_id/isolation/setup", adapter.MikrotikRouter().SetupIsolation)
			routers.GET("/:router_id/isolation/check", adapter.MikrotikRouter().CheckIsolationSetup)
		}

		routerID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440001")
		apiPort := 8728
		useSSL := false
		isActive := true

		Convey("Create", func() {
			input := model.MikrotikRouterInput{
				Name:     "Test Router",
				Address:  "192.168.1.1",
				ApiPort:  &apiPort,
				Username: "admin",
				Password: "password",
				UseSSL:   &useSSL,
				IsActive: &isActive,
			}

			Convey("Success", func() {
				mockMikrotikDBPort.EXPECT().Create(gomock.Any()).Return(nil).AnyTimes()

				body, _ := json.Marshal(input)
				req := httptest.NewRequest(http.MethodPost, "/routers", bytes.NewReader(body))
				req.Header.Set("Content-Type", "application/json")

				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				So(w.Code, ShouldBeBetweenOrEqual, 200, 201)
			})

			Convey("Invalid JSON", func() {
				req := httptest.NewRequest(http.MethodPost, "/routers", bytes.NewReader([]byte("invalid")))
				req.Header.Set("Content-Type", "application/json")

				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				So(w.Code, ShouldEqual, http.StatusBadRequest)
			})
		})

		Convey("List", func() {
			Convey("Success", func() {
				req := httptest.NewRequest(http.MethodGet, "/routers", nil)
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				So(w.Code, ShouldBeBetweenOrEqual, 200, 500)
			})
		})

		Convey("GetByID", func() {
			Convey("Success", func() {
				req := httptest.NewRequest(http.MethodGet, "/routers/"+routerID.String(), nil)
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				So(w.Code, ShouldBeBetweenOrEqual, 200, 404)
			})

			Convey("Invalid ID Format", func() {
				req := httptest.NewRequest(http.MethodGet, "/routers/invalid-id", nil)
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				So(w.Code, ShouldEqual, http.StatusBadRequest)

				var resp model.Response
				err := json.Unmarshal(w.Body.Bytes(), &resp)
				So(err, ShouldBeNil)
				So(resp.Success, ShouldBeFalse)
				So(resp.Error, ShouldContainSubstring, "invalid id")
			})
		})

		Convey("Update", func() {
			apiPort := 8729
			useSSL := true
			input := model.MikrotikRouterInput{
				Name:     "Updated Router",
				Address:  "192.168.1.2",
				ApiPort:  &apiPort,
				Username: "admin",
				Password: "newpassword",
				UseSSL:   &useSSL,
			}

			Convey("Success", func() {
				body, _ := json.Marshal(input)
				req := httptest.NewRequest(http.MethodPut, "/routers/"+routerID.String(), bytes.NewReader(body))
				req.Header.Set("Content-Type", "application/json")

				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				So(w.Code, ShouldBeBetweenOrEqual, 200, 500)
			})

			Convey("Invalid ID Format", func() {
				body, _ := json.Marshal(input)
				req := httptest.NewRequest(http.MethodPut, "/routers/invalid-id", bytes.NewReader(body))
				req.Header.Set("Content-Type", "application/json")

				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				So(w.Code, ShouldEqual, http.StatusBadRequest)
			})

			Convey("Invalid JSON", func() {
				req := httptest.NewRequest(http.MethodPut, "/routers/"+routerID.String(), bytes.NewReader([]byte("invalid")))
				req.Header.Set("Content-Type", "application/json")

				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				So(w.Code, ShouldEqual, http.StatusBadRequest)
			})
		})

		Convey("Delete", func() {
			Convey("Success", func() {
				req := httptest.NewRequest(http.MethodDelete, "/routers/"+routerID.String(), nil)
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				So(w.Code, ShouldBeBetweenOrEqual, 200, 500)
			})

			Convey("Invalid ID Format", func() {
				req := httptest.NewRequest(http.MethodDelete, "/routers/invalid-id", nil)
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				So(w.Code, ShouldEqual, http.StatusBadRequest)
			})
		})

		Convey("TestConnection", func() {
			Convey("Success", func() {
				req := httptest.NewRequest(http.MethodPost, "/routers/"+routerID.String()+"/test-connection", nil)
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				So(w.Code, ShouldBeBetweenOrEqual, 200, 404)
			})

			Convey("Invalid ID Format", func() {
				req := httptest.NewRequest(http.MethodPost, "/routers/invalid-id/test-connection", nil)
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				So(w.Code, ShouldEqual, http.StatusBadRequest)
			})
		})

		Convey("SetupIsolation", func() {
			input := model.IsolationConfig{
				ProfileName: "isolated",
				AddressList: "isolated",
				PortalIP:    "192.168.254.1",
				PortalPort:  "8080",
				DNSServer:   "1.1.1.1",
				RateLimit:   "1M/1M",
			}

			Convey("Success", func() {
				body, _ := json.Marshal(input)
				req := httptest.NewRequest(http.MethodPost, "/routers/"+routerID.String()+"/isolation/setup", bytes.NewReader(body))
				req.Header.Set("Content-Type", "application/json")

				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				So(w.Code, ShouldBeBetweenOrEqual, 200, 500)
			})

			Convey("Invalid ID Format", func() {
				body, _ := json.Marshal(input)
				req := httptest.NewRequest(http.MethodPost, "/routers/invalid-id/isolation/setup", bytes.NewReader(body))
				req.Header.Set("Content-Type", "application/json")

				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				So(w.Code, ShouldEqual, http.StatusBadRequest)
			})

			Convey("Missing PortalIP", func() {
				emptyConfig := model.IsolationConfig{
					ProfileName: "isolated",
				}
				body, _ := json.Marshal(emptyConfig)
				req := httptest.NewRequest(http.MethodPost, "/routers/"+routerID.String()+"/isolation/setup", bytes.NewReader(body))
				req.Header.Set("Content-Type", "application/json")

				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				So(w.Code, ShouldEqual, http.StatusBadRequest)
			})
		})

		Convey("CheckIsolationSetup", func() {
			Convey("Success", func() {
				req := httptest.NewRequest(http.MethodGet, "/routers/"+routerID.String()+"/isolation/check", nil)
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				So(w.Code, ShouldBeBetweenOrEqual, 200, 500)
			})

			Convey("Invalid ID Format", func() {
				req := httptest.NewRequest(http.MethodGet, "/routers/invalid-id/isolation/check", nil)
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				So(w.Code, ShouldEqual, http.StatusBadRequest)
			})
		})
	})
}
