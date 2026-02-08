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

func TestPaymentMethodAdapter(t *testing.T) {
	Convey("Test Payment Method HTTP Adapter", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		mockDatabasePort := mock_outbound_port.NewMockDatabasePort(mockCtrl)
		mockMessagePort := mock_outbound_port.NewMockMessagePort(mockCtrl)
		mockCachePort := mock_outbound_port.NewMockCachePort(mockCtrl)
		mockWorkflowPort := mock_outbound_port.NewMockWorkflowPort(mockCtrl)
		mockHttpPort := mock_outbound_port.NewMockHttpPort(mockCtrl)

		mockPaymentMethodDatabasePort := mock_outbound_port.NewMockPaymentMethodDatabasePort(mockCtrl)
		mockDatabasePort.EXPECT().PaymentMethod().Return(mockPaymentMethodDatabasePort).AnyTimes()

		dom := domain.NewDomain(mockDatabasePort, mockMessagePort, mockCachePort, mockWorkflowPort, mockHttpPort)
		adapter := gin_inbound_adapter.NewAdapter(dom, mockHttpPort)

		pmOutput := model.PaymentMethod{
			ID: "pm-123",
			PaymentMethodInput: model.PaymentMethodInput{
				TenantID:      "tenant-123",
				Name:          "BCA Transfer",
				Type:          "bank_transfer",
				AccountName:   "PT Test ISP",
				AccountNumber: "1234567890",
				BankName:      "BCA",
				IsActive:      true,
				CreatedAt:     time.Now(),
				UpdatedAt:     time.Now(),
			},
		}

		router := gin.New()
		router.GET("/api/v1/payment-methods", func(c *gin.Context) {
			c.Set("tenant_id", "tenant-123")
			adapter.PaymentMethod().List(c)
		})
		router.POST("/api/v1/payment-methods", func(c *gin.Context) {
			c.Set("tenant_id", "tenant-123")
			adapter.PaymentMethod().Create(c)
		})
		router.GET("/api/v1/payment-methods/:id", func(c *gin.Context) {
			adapter.PaymentMethod().Get(c)
		})
		router.PUT("/api/v1/payment-methods/:id", func(c *gin.Context) {
			adapter.PaymentMethod().Update(c)
		})
		router.DELETE("/api/v1/payment-methods/:id", func(c *gin.Context) {
			adapter.PaymentMethod().Delete(c)
		})

		Convey("List", func() {
			Convey("Success", func() {
				mockPaymentMethodDatabasePort.EXPECT().FindByFilter(gomock.Any()).Return([]model.PaymentMethod{pmOutput}, nil).Times(1)

				req := httptest.NewRequest(http.MethodGet, "/api/v1/payment-methods", nil)
				w := httptest.NewRecorder()

				router.ServeHTTP(w, req)

				So(w.Code, ShouldEqual, http.StatusOK)

				var result model.Response
				json.Unmarshal(w.Body.Bytes(), &result)
				So(result.Success, ShouldBeTrue)
			})

			Convey("Domain error", func() {
				mockPaymentMethodDatabasePort.EXPECT().FindByFilter(gomock.Any()).Return(nil, errors.New("error")).Times(1)

				req := httptest.NewRequest(http.MethodGet, "/api/v1/payment-methods", nil)
				w := httptest.NewRecorder()

				router.ServeHTTP(w, req)

				So(w.Code, ShouldEqual, http.StatusInternalServerError)
			})
		})

		Convey("Create", func() {
			payload := model.PaymentMethodInput{
				Name:          "Mandiri Transfer",
				Type:          "bank_transfer",
				AccountName:   "PT Test ISP",
				AccountNumber: "0987654321",
				BankName:      "Mandiri",
			}

			Convey("Success", func() {
				mockPaymentMethodDatabasePort.EXPECT().Create(gomock.Any()).Return(pmOutput, nil).Times(1)

				body, _ := json.Marshal(payload)
				req := httptest.NewRequest(http.MethodPost, "/api/v1/payment-methods", bytes.NewReader(body))
				req.Header.Set("Content-Type", "application/json")
				w := httptest.NewRecorder()

				router.ServeHTTP(w, req)

				So(w.Code, ShouldEqual, http.StatusCreated)

				var result model.Response
				json.Unmarshal(w.Body.Bytes(), &result)
				So(result.Success, ShouldBeTrue)
			})

			Convey("Invalid JSON", func() {
				req := httptest.NewRequest(http.MethodPost, "/api/v1/payment-methods", bytes.NewReader([]byte("invalid")))
				req.Header.Set("Content-Type", "application/json")
				w := httptest.NewRecorder()

				router.ServeHTTP(w, req)

				So(w.Code, ShouldEqual, http.StatusBadRequest)
			})

			Convey("Domain error", func() {
				mockPaymentMethodDatabasePort.EXPECT().Create(gomock.Any()).Return(model.PaymentMethod{}, errors.New("error")).Times(1)

				body, _ := json.Marshal(payload)
				req := httptest.NewRequest(http.MethodPost, "/api/v1/payment-methods", bytes.NewReader(body))
				req.Header.Set("Content-Type", "application/json")
				w := httptest.NewRecorder()

				router.ServeHTTP(w, req)

				So(w.Code, ShouldEqual, http.StatusInternalServerError)
			})
		})

		Convey("Get", func() {
			Convey("Success", func() {
				mockPaymentMethodDatabasePort.EXPECT().FindByID(gomock.Any()).Return(pmOutput, nil).Times(1)

				req := httptest.NewRequest(http.MethodGet, "/api/v1/payment-methods/pm-123", nil)
				w := httptest.NewRecorder()

				router.ServeHTTP(w, req)

				So(w.Code, ShouldEqual, http.StatusOK)

				var result model.Response
				json.Unmarshal(w.Body.Bytes(), &result)
				So(result.Success, ShouldBeTrue)
			})

			Convey("Not found", func() {
				mockPaymentMethodDatabasePort.EXPECT().FindByID(gomock.Any()).Return(model.PaymentMethod{}, errors.New("not found")).Times(1)

				req := httptest.NewRequest(http.MethodGet, "/api/v1/payment-methods/pm-999", nil)
				w := httptest.NewRecorder()

				router.ServeHTTP(w, req)

				So(w.Code, ShouldEqual, http.StatusInternalServerError)
			})
		})

		Convey("Update", func() {
			payload := model.PaymentMethodInput{
				Name: "Updated Payment Method",
			}

			Convey("Success", func() {
				mockPaymentMethodDatabasePort.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil).Times(1)

				body, _ := json.Marshal(payload)
				req := httptest.NewRequest(http.MethodPut, "/api/v1/payment-methods/pm-123", bytes.NewReader(body))
				req.Header.Set("Content-Type", "application/json")
				w := httptest.NewRecorder()

				router.ServeHTTP(w, req)

				So(w.Code, ShouldEqual, http.StatusOK)

				var result model.Response
				json.Unmarshal(w.Body.Bytes(), &result)
				So(result.Success, ShouldBeTrue)
			})

			Convey("Invalid JSON", func() {
				req := httptest.NewRequest(http.MethodPut, "/api/v1/payment-methods/pm-123", bytes.NewReader([]byte("invalid")))
				req.Header.Set("Content-Type", "application/json")
				w := httptest.NewRecorder()

				router.ServeHTTP(w, req)

				So(w.Code, ShouldEqual, http.StatusBadRequest)
			})
		})

		Convey("Delete", func() {
			Convey("Success", func() {
				mockPaymentMethodDatabasePort.EXPECT().Delete(gomock.Any()).Return(nil).Times(1)

				req := httptest.NewRequest(http.MethodDelete, "/api/v1/payment-methods/pm-123", nil)
				w := httptest.NewRecorder()

				router.ServeHTTP(w, req)

				So(w.Code, ShouldEqual, http.StatusOK)

				var result model.Response
				json.Unmarshal(w.Body.Bytes(), &result)
				So(result.Success, ShouldBeTrue)
			})

			Convey("Domain error", func() {
				mockPaymentMethodDatabasePort.EXPECT().Delete(gomock.Any()).Return(errors.New("error")).Times(1)

				req := httptest.NewRequest(http.MethodDelete, "/api/v1/payment-methods/pm-999", nil)
				w := httptest.NewRecorder()

				router.ServeHTTP(w, req)

				So(w.Code, ShouldEqual, http.StatusInternalServerError)
			})
		})
	})
}
