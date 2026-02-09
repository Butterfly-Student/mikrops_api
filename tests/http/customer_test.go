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

func TestCustomerAdapter(t *testing.T) {
	Convey("Test Customer HTTP Adapter", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		mockDatabasePort := mock_outbound_port.NewMockDatabasePort(mockCtrl)
		mockMessagePort := mock_outbound_port.NewMockMessagePort(mockCtrl)
		mockCachePort := mock_outbound_port.NewMockCachePort(mockCtrl)
		mockWorkflowPort := mock_outbound_port.NewMockWorkflowPort(mockCtrl)
		mockHttpPort := mock_outbound_port.NewMockHttpPort(mockCtrl)

		mockCustomerDatabasePort := mock_outbound_port.NewMockCustomerDatabasePort(mockCtrl)
		mockDatabasePort.EXPECT().Customer().Return(mockCustomerDatabasePort).AnyTimes()

		dom := domain.NewDomain(mockDatabasePort, mockMessagePort, mockCachePort, mockWorkflowPort, mockHttpPort)
		adapter := gin_inbound_adapter.NewAdapter(dom, mockHttpPort)

		customerOutput := model.Customer{
			ID: "cust-123",
			CustomerInput: model.CustomerInput{
				TenantID:      "tenant-123",
				FullName:      "John Doe",
				Email:         "john@test.com",
				Phone:         "081234567890",
				IsActive:      true,
				CreatedAt:     time.Now(),
				UpdatedAt:     time.Now(),
			},
		}

		router := gin.New()
		router.GET("/api/v1/customers", func(c *gin.Context) {
			c.Set("tenant_id", "tenant-123")
			adapter.Customer().List(c)
		})
		router.POST("/api/v1/customers", func(c *gin.Context) {
			c.Set("tenant_id", "tenant-123")
			adapter.Customer().Create(c)
		})
		router.GET("/api/v1/customers/:id", func(c *gin.Context) {
			adapter.Customer().Get(c)
		})
		router.PUT("/api/v1/customers/:id", func(c *gin.Context) {
			adapter.Customer().Update(c)
		})
		router.DELETE("/api/v1/customers/:id", func(c *gin.Context) {
			adapter.Customer().Delete(c)
		})

		Convey("List", func() {
			Convey("Success", func() {
				mockCustomerDatabasePort.EXPECT().FindByFilter(gomock.Any()).Return([]model.Customer{customerOutput}, nil).Times(1)

				req := httptest.NewRequest(http.MethodGet, "/api/v1/customers", nil)
				w := httptest.NewRecorder()

				router.ServeHTTP(w, req)

				So(w.Code, ShouldEqual, http.StatusOK)

				var result model.Response
				json.Unmarshal(w.Body.Bytes(), &result)
				So(result.Success, ShouldBeTrue)
			})

			Convey("Domain error", func() {
				mockCustomerDatabasePort.EXPECT().FindByFilter(gomock.Any()).Return(nil, errors.New("error")).Times(1)

				req := httptest.NewRequest(http.MethodGet, "/api/v1/customers", nil)
				w := httptest.NewRecorder()

				router.ServeHTTP(w, req)

				So(w.Code, ShouldEqual, http.StatusInternalServerError)
			})
		})

		Convey("Create", func() {
			payload := model.CustomerInput{
				FullName:      "Jane Doe",
				Email:         "jane@test.com",
				Phone:         "081234567891",
			}

			Convey("Success", func() {
				mockCustomerDatabasePort.EXPECT().Create(gomock.Any()).Return(customerOutput, nil).Times(1)

				body, _ := json.Marshal(payload)
				req := httptest.NewRequest(http.MethodPost, "/api/v1/customers", bytes.NewReader(body))
				req.Header.Set("Content-Type", "application/json")
				w := httptest.NewRecorder()

				router.ServeHTTP(w, req)

				So(w.Code, ShouldEqual, http.StatusCreated)

				var result model.Response
				json.Unmarshal(w.Body.Bytes(), &result)
				So(result.Success, ShouldBeTrue)
			})

			Convey("Invalid JSON", func() {
				req := httptest.NewRequest(http.MethodPost, "/api/v1/customers", bytes.NewReader([]byte("invalid")))
				req.Header.Set("Content-Type", "application/json")
				w := httptest.NewRecorder()

				router.ServeHTTP(w, req)

				So(w.Code, ShouldEqual, http.StatusBadRequest)
			})

			Convey("Domain error", func() {
				mockCustomerDatabasePort.EXPECT().Create(gomock.Any()).Return(model.Customer{}, errors.New("error")).Times(1)

				body, _ := json.Marshal(payload)
				req := httptest.NewRequest(http.MethodPost, "/api/v1/customers", bytes.NewReader(body))
				req.Header.Set("Content-Type", "application/json")
				w := httptest.NewRecorder()

				router.ServeHTTP(w, req)

				So(w.Code, ShouldEqual, http.StatusInternalServerError)
			})
		})

		Convey("Get", func() {
			Convey("Success", func() {
				mockCustomerDatabasePort.EXPECT().FindByID(gomock.Any()).Return(customerOutput, nil).Times(1)

				req := httptest.NewRequest(http.MethodGet, "/api/v1/customers/cust-123", nil)
				w := httptest.NewRecorder()

				router.ServeHTTP(w, req)

				So(w.Code, ShouldEqual, http.StatusOK)

				var result model.Response
				json.Unmarshal(w.Body.Bytes(), &result)
				So(result.Success, ShouldBeTrue)
			})

			Convey("Not found", func() {
				mockCustomerDatabasePort.EXPECT().FindByID(gomock.Any()).Return(model.Customer{}, errors.New("not found")).Times(1)

				req := httptest.NewRequest(http.MethodGet, "/api/v1/customers/cust-999", nil)
				w := httptest.NewRecorder()

				router.ServeHTTP(w, req)

				So(w.Code, ShouldEqual, http.StatusInternalServerError)
			})
		})

		Convey("Update", func() {
			payload := model.CustomerInput{
				FullName: "Updated Name",
			}

			Convey("Success", func() {
				mockCustomerDatabasePort.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil).Times(1)

				body, _ := json.Marshal(payload)
				req := httptest.NewRequest(http.MethodPut, "/api/v1/customers/cust-123", bytes.NewReader(body))
				req.Header.Set("Content-Type", "application/json")
				w := httptest.NewRecorder()

				router.ServeHTTP(w, req)

				So(w.Code, ShouldEqual, http.StatusOK)

				var result model.Response
				json.Unmarshal(w.Body.Bytes(), &result)
				So(result.Success, ShouldBeTrue)
			})

			Convey("Invalid JSON", func() {
				req := httptest.NewRequest(http.MethodPut, "/api/v1/customers/cust-123", bytes.NewReader([]byte("invalid")))
				req.Header.Set("Content-Type", "application/json")
				w := httptest.NewRecorder()

				router.ServeHTTP(w, req)

				So(w.Code, ShouldEqual, http.StatusBadRequest)
			})
		})

		Convey("Delete", func() {
			Convey("Success", func() {
				mockCustomerDatabasePort.EXPECT().Delete(gomock.Any()).Return(nil).Times(1)

				req := httptest.NewRequest(http.MethodDelete, "/api/v1/customers/cust-123", nil)
				w := httptest.NewRecorder()

				router.ServeHTTP(w, req)

				So(w.Code, ShouldEqual, http.StatusOK)

				var result model.Response
				json.Unmarshal(w.Body.Bytes(), &result)
				So(result.Success, ShouldBeTrue)
			})

			Convey("Domain error", func() {
				mockCustomerDatabasePort.EXPECT().Delete(gomock.Any()).Return(errors.New("error")).Times(1)

				req := httptest.NewRequest(http.MethodDelete, "/api/v1/customers/cust-999", nil)
				w := httptest.NewRecorder()

				router.ServeHTTP(w, req)

				So(w.Code, ShouldEqual, http.StatusInternalServerError)
			})
		})
	})
}
