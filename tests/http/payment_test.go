package http_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	. "github.com/smartystreets/goconvey/convey"

	gin_inbound_adapter "mikrops/internal/adapter/inbound/gin"
	"mikrops/internal/domain"
	"mikrops/internal/model"
	mock_outbound_port "mikrops/tests/mocks/port"
)

func TestPaymentAdapter(t *testing.T) {
	Convey("Test Payment HTTP Adapter", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		mockDatabasePort := mock_outbound_port.NewMockDatabasePort(mockCtrl)
		mockMessagePort := mock_outbound_port.NewMockMessagePort(mockCtrl)
		mockCachePort := mock_outbound_port.NewMockCachePort(mockCtrl)
		mockWorkflowPort := mock_outbound_port.NewMockWorkflowPort(mockCtrl)
		mockHttpPort := mock_outbound_port.NewMockHttpPort(mockCtrl)

		mockPaymentDatabasePort := mock_outbound_port.NewMockPaymentDatabasePort(mockCtrl)
		mockInvoiceDatabasePort := mock_outbound_port.NewMockInvoiceDatabasePort(mockCtrl)
		mockDatabasePort.EXPECT().Payment().Return(mockPaymentDatabasePort).AnyTimes()
		mockDatabasePort.EXPECT().Invoice().Return(mockInvoiceDatabasePort).AnyTimes()
		mockDatabasePort.EXPECT().DoInTransaction(gomock.Any()).DoAndReturn(func(fn func() error) error {
			return fn()
		}).AnyTimes()

		dom := domain.NewDomain(mockDatabasePort, mockMessagePort, mockCachePort, mockWorkflowPort, mockHttpPort)
		adapter := gin_inbound_adapter.NewAdapter(dom, mockHttpPort)

		paymentOutput := model.Payment{
			ID: "pay-123",
			PaymentInput: model.PaymentInput{
				TenantID:        "tenant-123",
				InvoiceID:       "inv-123",
				PaymentMethodID: "pm-123",
				Amount:          150000,
				PaymentDate:     time.Now(),
				Status:          model.PaymentStatusPending,
				CreatedAt:       time.Now(),
				UpdatedAt:       time.Now(),
			},
		}

		router := gin.New()
		router.GET("/api/v1/payments", func(c *gin.Context) {
			c.Set("tenant_id", "tenant-123")
			adapter.Payment().List(c)
		})
		router.GET("/api/v1/payments/:id", func(c *gin.Context) {
			adapter.Payment().Get(c)
		})
		router.POST("/api/v1/payments/:id/verify", func(c *gin.Context) {
			c.Set("staff_id", "staff-123")
			adapter.Payment().Verify(c)
		})
		router.POST("/api/v1/payments/:id/reject", func(c *gin.Context) {
			c.Set("staff_id", "staff-123")
			adapter.Payment().Reject(c)
		})

		Convey("List", func() {
			Convey("Success", func() {
				mockPaymentDatabasePort.EXPECT().FindByFilter(gomock.Any()).Return([]model.Payment{paymentOutput}, nil).Times(1)

				req := httptest.NewRequest(http.MethodGet, "/api/v1/payments", nil)
				w := httptest.NewRecorder()

				router.ServeHTTP(w, req)

				So(w.Code, ShouldEqual, http.StatusOK)

				var result model.Response
				json.Unmarshal(w.Body.Bytes(), &result)
				So(result.Success, ShouldBeTrue)
			})

			Convey("Domain error", func() {
				mockPaymentDatabasePort.EXPECT().FindByFilter(gomock.Any()).Return(nil, errors.New("error")).Times(1)

				req := httptest.NewRequest(http.MethodGet, "/api/v1/payments", nil)
				w := httptest.NewRecorder()

				router.ServeHTTP(w, req)

				So(w.Code, ShouldEqual, http.StatusInternalServerError)
			})
		})

		Convey("Get", func() {
			Convey("Success", func() {
				mockPaymentDatabasePort.EXPECT().FindByID(gomock.Any()).Return(paymentOutput, nil).Times(1)

				req := httptest.NewRequest(http.MethodGet, "/api/v1/payments/pay-123", nil)
				w := httptest.NewRecorder()

				router.ServeHTTP(w, req)

				So(w.Code, ShouldEqual, http.StatusOK)

				var result model.Response
				json.Unmarshal(w.Body.Bytes(), &result)
				So(result.Success, ShouldBeTrue)
			})

			Convey("Not found", func() {
				mockPaymentDatabasePort.EXPECT().FindByID(gomock.Any()).Return(model.Payment{}, errors.New("not found")).Times(1)

				req := httptest.NewRequest(http.MethodGet, "/api/v1/payments/pay-999", nil)
				w := httptest.NewRecorder()

				router.ServeHTTP(w, req)

				So(w.Code, ShouldEqual, http.StatusInternalServerError)
			})
		})

		Convey("Verify", func() {
			Convey("Success", func() {
				mockPaymentDatabasePort.EXPECT().FindByID(gomock.Any()).Return(paymentOutput, nil).AnyTimes()
				mockPaymentDatabasePort.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
				mockInvoiceDatabasePort.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

				req := httptest.NewRequest(http.MethodPost, "/api/v1/payments/pay-123/verify", nil)
				w := httptest.NewRecorder()

				router.ServeHTTP(w, req)

				So(w.Code, ShouldEqual, http.StatusOK)

				var result model.Response
				json.Unmarshal(w.Body.Bytes(), &result)
				So(result.Success, ShouldBeTrue)
			})

			Convey("Domain error", func() {
				mockPaymentDatabasePort.EXPECT().FindByID(gomock.Any()).Return(model.Payment{}, errors.New("not found")).Times(1)

				req := httptest.NewRequest(http.MethodPost, "/api/v1/payments/pay-999/verify", nil)
				w := httptest.NewRecorder()

				router.ServeHTTP(w, req)

				So(w.Code, ShouldEqual, http.StatusInternalServerError)
			})
		})

		Convey("Reject", func() {
			payload := map[string]string{"notes": "Invalid proof"}

			Convey("Success", func() {
				mockPaymentDatabasePort.EXPECT().FindByID(gomock.Any()).Return(paymentOutput, nil).Times(1)
				mockPaymentDatabasePort.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil).Times(1)

				body, _ := json.Marshal(payload)
				req := httptest.NewRequest(http.MethodPost, "/api/v1/payments/pay-123/reject", bytes.NewReader(body))
				req.Header.Set("Content-Type", "application/json")
				w := httptest.NewRecorder()

				router.ServeHTTP(w, req)

				So(w.Code, ShouldEqual, http.StatusOK)

				var result model.Response
				json.Unmarshal(w.Body.Bytes(), &result)
				So(result.Success, ShouldBeTrue)
			})

			Convey("Domain error", func() {
				mockPaymentDatabasePort.EXPECT().FindByID(gomock.Any()).Return(model.Payment{}, errors.New("not found")).Times(1)

				body, _ := json.Marshal(payload)
				req := httptest.NewRequest(http.MethodPost, "/api/v1/payments/pay-999/reject", bytes.NewReader(body))
				req.Header.Set("Content-Type", "application/json")
				w := httptest.NewRecorder()

				router.ServeHTTP(w, req)

				So(w.Code, ShouldEqual, http.StatusInternalServerError)
			})
		})
	})
}
