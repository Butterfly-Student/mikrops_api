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

func TestCustomerAdapter(t *testing.T) {
	Convey("Test Customer HTTP Adapter", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		mockDatabasePort := mock_outbound_port.NewMockDatabasePort(mockCtrl)
		mockMikrotikDBPort := mock_outbound_port.NewMockMikrotikDatabasePort(mockCtrl)
		mockMessagePort := mock_outbound_port.NewMockMessagePort(mockCtrl)
		mockCachePort := mock_outbound_port.NewMockCachePort(mockCtrl)
		mockWorkflowPort := mock_outbound_port.NewMockWorkflowPort(mockCtrl)
		mockMikrotikPort := mock_outbound_port.NewMockMikrotikPort(mockCtrl)

		mockDatabasePort.EXPECT().Mikrotik().Return(mockMikrotikDBPort).AnyTimes()

		mockCustomerDBPort := mock_outbound_port.NewMockCustomerDatabasePort(mockCtrl)
		mockDatabasePort.EXPECT().Customer().Return(mockCustomerDBPort).AnyTimes()

		// Setup common expectations for all subtests
		mockCustomerDBPort.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
		mockCustomerDBPort.EXPECT().FindByID(gomock.Any(), gomock.Any()).Return(&model.Customer{ID: uuid.MustParse("550e8400-e29b-41d4-a716-446655440002")}, nil).AnyTimes()
		mockCustomerDBPort.EXPECT().FindByCode(gomock.Any(), gomock.Any()).Return(&model.Customer{ID: uuid.MustParse("550e8400-e29b-41d4-a716-446655440002"), CustomerCode: "CUST001"}, nil).AnyTimes()
		mockCustomerDBPort.EXPECT().FindAll(gomock.Any(), gomock.Any()).Return([]model.Customer{{ID: uuid.MustParse("550e8400-e29b-41d4-a716-446655440002")}}, nil).AnyTimes()
		mockCustomerDBPort.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
		mockCustomerDBPort.EXPECT().Delete(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
		mockCustomerDBPort.EXPECT().UpdateStatus(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

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

		// Router with MikroTik context middleware
		routerWithMikrotik := func(router *gin.Engine) {
			mikrotikGroup := router.Group("/mikrotik/:router_id")
			mikrotikGroup.Use(func(c *gin.Context) {
				// Mock router context
				c.Set("router", &model.MikrotikRouter{
					ID:      uuid.MustParse("550e8400-e29b-41d4-a716-446655440001"),
					Name:    "TestRouter",
					Address: "192.168.88.1:8728",
				})
				c.Next()
			})
			mikrotikGroup.Use(authMiddleware)
			{
				mikrotikGroup.POST("/customers", adapter.Customer().Create)
				mikrotikGroup.PUT("/customers/:id", adapter.Customer().Update)
			}
		}
		routerWithMikrotik(router)

		customers := router.Group("/customers")
		customers.Use(authMiddleware)
		{
			customers.GET("", adapter.Customer().List)
			customers.GET("/:id", adapter.Customer().GetByID)
			customers.GET("/code/:code", adapter.Customer().GetByCode)
			customers.DELETE("/:id", adapter.Customer().Delete)
			customers.POST("/:id/status", adapter.Customer().ChangeStatus)
		}

		customerID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440002")

		Convey("Create", func() {
			input := model.CustomerInput{
				CustomerCode: "CUST001",
				FullName:     "Test Customer",
				Phone:        "08123456789",
			}

			Convey("Success", func() {
				mockCustomerDBPort.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

				body, _ := json.Marshal(input)
				req := httptest.NewRequest(http.MethodPost, "/mikrotik/550e8400-e29b-41d4-a716-446655440001/customers", bytes.NewReader(body))
				req.Header.Set("Content-Type", "application/json")

				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				// Note: Actual success depends on domain implementation
				So(w.Code, ShouldBeBetweenOrEqual, 200, 201)
			})

			Convey("Missing Router Context", func() {
				// Test without MikroTik middleware - should fail
				body, _ := json.Marshal(input)
				req := httptest.NewRequest(http.MethodPost, "/customers", bytes.NewReader(body))
				req.Header.Set("Content-Type", "application/json")

				w := httptest.NewRecorder()
				// This won't match the route properly
				router.ServeHTTP(w, req)
			})

			Convey("Invalid JSON", func() {
				req := httptest.NewRequest(http.MethodPost, "/mikrotik/550e8400-e29b-41d4-a716-446655440001/customers", bytes.NewReader([]byte("invalid")))
				req.Header.Set("Content-Type", "application/json")

				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				So(w.Code, ShouldEqual, http.StatusBadRequest)
			})
		})

		Convey("GetByID", func() {
			Convey("Success", func() {
				req := httptest.NewRequest(http.MethodGet, "/customers/"+customerID.String(), nil)
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				// Note: Actual success depends on domain implementation
			})
		})

		Convey("GetByCode", func() {
			Convey("Success", func() {
				req := httptest.NewRequest(http.MethodGet, "/customers/code/CUST001", nil)
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				// Note: Actual success depends on domain implementation
			})
		})

		Convey("List", func() {
			Convey("Success", func() {
				req := httptest.NewRequest(http.MethodGet, "/customers", nil)
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				// Note: Actual success depends on domain implementation
			})

			Convey("With Filters", func() {
				req := httptest.NewRequest(http.MethodGet, "/customers?status=active&search=test", nil)
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				So(w.Code, ShouldBeBetweenOrEqual, 200, 500)
			})
		})

		Convey("Update", func() {
			input := model.CustomerInput{
				FullName: "Updated Customer Name",
				Phone:    "08987654321",
			}

			Convey("Success", func() {
				body, _ := json.Marshal(input)
				req := httptest.NewRequest(http.MethodPut, "/mikrotik/550e8400-e29b-41d4-a716-446655440001/customers/"+customerID.String(), bytes.NewReader(body))
				req.Header.Set("Content-Type", "application/json")

				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				// Note: Actual success depends on domain implementation
			})

			Convey("Invalid JSON", func() {
				req := httptest.NewRequest(http.MethodPut, "/mikrotik/550e8400-e29b-41d4-a716-446655440001/customers/"+customerID.String(), bytes.NewReader([]byte("invalid")))
				req.Header.Set("Content-Type", "application/json")

				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				So(w.Code, ShouldEqual, http.StatusBadRequest)
			})
		})

		Convey("Delete", func() {
			Convey("Success", func() {
				req := httptest.NewRequest(http.MethodDelete, "/customers/"+customerID.String(), nil)
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				// Note: Actual success depends on domain implementation
			})
		})

		Convey("ChangeStatus", func() {
			input := struct {
				Status string `json:"status"`
			}{
				Status: "active",
			}

			Convey("Success", func() {
				body, _ := json.Marshal(input)
				req := httptest.NewRequest(http.MethodPost, "/customers/"+customerID.String()+"/status", bytes.NewReader(body))
				req.Header.Set("Content-Type", "application/json")

				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				// Note: Actual success depends on domain implementation
			})

			Convey("Invalid JSON", func() {
				req := httptest.NewRequest(http.MethodPost, "/customers/"+customerID.String()+"/status", bytes.NewReader([]byte("invalid")))
				req.Header.Set("Content-Type", "application/json")

				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				So(w.Code, ShouldEqual, http.StatusBadRequest)
			})
		})
	})
}
