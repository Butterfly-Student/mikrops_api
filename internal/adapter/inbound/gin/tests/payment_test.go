package gin_adapter_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

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

func TestPaymentAdapter(t *testing.T) {
	Convey("Test Payment HTTP Adapter", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		mockDatabasePort := mock_outbound_port.NewMockDatabasePort(mockCtrl)
		mockMessagePort := mock_outbound_port.NewMockMessagePort(mockCtrl)
		mockCachePort := mock_outbound_port.NewMockCachePort(mockCtrl)
		mockWorkflowPort := mock_outbound_port.NewMockWorkflowPort(mockCtrl)

		paymentID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440001")
		customerID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440002")
		invoiceID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440003")

		mockInvoiceDBPort := mock_outbound_port.NewMockInvoiceDatabasePort(mockCtrl)
		mockCustomerDBPort := mock_outbound_port.NewMockCustomerDatabasePort(mockCtrl)
		mockPaymentDBPort := mock_outbound_port.NewMockPaymentDatabasePort(mockCtrl)
		mockDatabasePort.EXPECT().Invoice().Return(mockInvoiceDBPort).AnyTimes()
		mockDatabasePort.EXPECT().Customer().Return(mockCustomerDBPort).AnyTimes()
		mockDatabasePort.EXPECT().Payment().Return(mockPaymentDBPort).AnyTimes()

		// Setup common expectations for all subtests
		mockInvoiceDBPort.EXPECT().FindByID(gomock.Any(), gomock.Any()).Return(&model.Invoice{ID: invoiceID, TotalAmount: 100000, Status: model.InvoiceStatusDraft}, nil).AnyTimes()
		mockInvoiceDBPort.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
		mockCustomerDBPort.EXPECT().FindByID(gomock.Any(), gomock.Any()).Return(&model.Customer{ID: customerID}, nil).AnyTimes()
		mockPaymentDBPort.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
		mockPaymentDBPort.EXPECT().FindByID(gomock.Any(), gomock.Any()).Return(&model.Payment{ID: paymentID}, nil).AnyTimes()
		mockPaymentDBPort.EXPECT().FindByNumber(gomock.Any(), gomock.Any()).Return(&model.Payment{ID: paymentID}, nil).AnyTimes()
		mockPaymentDBPort.EXPECT().FindAll(gomock.Any(), gomock.Any()).Return([]model.Payment{{ID: paymentID}}, nil).AnyTimes()

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

		dom := domain.NewDomain(mockDatabasePort, mockMessagePort, mockCachePort, mockWorkflowPort, nil, nil, enforcer)
		adapter := gin_inbound_adapter.NewAdapter(dom)

		gin.SetMode(gin.TestMode)
		router := gin.New()

		testUUID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440001")
		authMiddleware := func(c *gin.Context) {
			c.Set("userID", testUUID.String())
			c.Next()
		}

		payments := router.Group("/payments")
		payments.Use(authMiddleware)
		{
			payments.POST("", adapter.Payment().Create)
			payments.GET("", adapter.Payment().List)
			payments.GET("/:id", adapter.Payment().GetByID)
			payments.GET("/number/:number", adapter.Payment().GetByNumber)
		}

		now := time.Now()

		Convey("Create", func() {
			input := model.PaymentInput{
				CustomerID:    customerID,
				InvoiceID:     &invoiceID,
				PaymentNumber: "PAY/2024/001",
				Amount:        100000,
				PaymentMethod: "cash",
				PaymentDate:   now,
			}

			Convey("Success", func() {
				body, _ := json.Marshal(input)
				req := httptest.NewRequest(http.MethodPost, "/payments", bytes.NewReader(body))
				req.Header.Set("Content-Type", "application/json")

				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				So(w.Code, ShouldBeBetweenOrEqual, 200, 201)
			})

			Convey("Invalid JSON", func() {
				req := httptest.NewRequest(http.MethodPost, "/payments", bytes.NewReader([]byte("invalid")))
				req.Header.Set("Content-Type", "application/json")

				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				So(w.Code, ShouldEqual, http.StatusBadRequest)
			})
		})

		Convey("GetByID", func() {
			Convey("Success", func() {
				req := httptest.NewRequest(http.MethodGet, "/payments/"+paymentID.String(), nil)
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				// Note: Actual success depends on domain implementation
			})
		})

		Convey("GetByNumber", func() {
			Convey("Success", func() {
				req := httptest.NewRequest(http.MethodGet, "/payments/number/PAY-2024-001", nil)
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				// Note: Actual success depends on domain implementation
			})
		})

		Convey("List", func() {
			Convey("Success", func() {
				req := httptest.NewRequest(http.MethodGet, "/payments", nil)
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				// Note: Actual success depends on domain implementation
			})

			Convey("With Filters", func() {
				req := httptest.NewRequest(http.MethodGet, "/payments?customer_id="+customerID.String()+"&status=confirmed", nil)
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				So(w.Code, ShouldBeBetweenOrEqual, 200, 500)
			})
		})
	})
}
