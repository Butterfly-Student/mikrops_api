package http_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
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

func TestSubscriptionAdapter(t *testing.T) {
	os.Setenv("ENCRYPTION_KEY", "01234567890123456789012345678901")

	Convey("Test Subscription HTTP Adapter", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		mockDatabasePort := mock_outbound_port.NewMockDatabasePort(mockCtrl)
		mockMessagePort := mock_outbound_port.NewMockMessagePort(mockCtrl)
		mockCachePort := mock_outbound_port.NewMockCachePort(mockCtrl)
		mockWorkflowPort := mock_outbound_port.NewMockWorkflowPort(mockCtrl)
		mockHttpPort := mock_outbound_port.NewMockHttpPort(mockCtrl)

		mockSubscriptionDatabasePort := mock_outbound_port.NewMockSubscriptionDatabasePort(mockCtrl)
		mockCustomerDatabasePort := mock_outbound_port.NewMockCustomerDatabasePort(mockCtrl)
		mockInternetPackageDatabasePort := mock_outbound_port.NewMockInternetPackageDatabasePort(mockCtrl)
		mockNasDatabasePort := mock_outbound_port.NewMockNasDatabasePort(mockCtrl)
		mockMikrotikPort := mock_outbound_port.NewMockMikrotikPort(mockCtrl)

		mockDatabasePort.EXPECT().Subscription().Return(mockSubscriptionDatabasePort).AnyTimes()
		mockDatabasePort.EXPECT().Customer().Return(mockCustomerDatabasePort).AnyTimes()
		mockDatabasePort.EXPECT().InternetPackage().Return(mockInternetPackageDatabasePort).AnyTimes()
		mockDatabasePort.EXPECT().Nas().Return(mockNasDatabasePort).AnyTimes()
		mockHttpPort.EXPECT().Mikrotik().Return(mockMikrotikPort).AnyTimes()

		dom := domain.NewDomain(mockDatabasePort, mockMessagePort, mockCachePort, mockWorkflowPort, mockHttpPort)
		adapter := gin_inbound_adapter.NewAdapter(dom, mockHttpPort)

		subOutput := model.Subscription{
			ID: "sub-123",
			SubscriptionInput: model.SubscriptionInput{
				TenantID:           "tenant-123",
				CustomerID:         "cust-123",
				PackageID:          "pkg-123",
				NasID:              "nas-123",
				Status:             model.SubscriptionStatusActive,
				StartDate:          time.Now(),
				EndDate:            time.Now().AddDate(0, 1, 0),
				MikrotikSecretName: "pppoe_john",
				MikrotikQueueName:  "queue_pppoe_john",
			},
		}

		router := gin.New()
		router.GET("/api/v1/subscriptions", func(c *gin.Context) {
			c.Set("tenant_id", "tenant-123")
			adapter.Subscription().List(c)
		})
		router.POST("/api/v1/subscriptions", func(c *gin.Context) {
			c.Set("tenant_id", "tenant-123")
			adapter.Subscription().Create(c)
		})
		router.GET("/api/v1/subscriptions/:id", func(c *gin.Context) {
			adapter.Subscription().Get(c)
		})
		router.PUT("/api/v1/subscriptions/:id", func(c *gin.Context) {
			adapter.Subscription().Update(c)
		})
		router.POST("/api/v1/subscriptions/:id/suspend", func(c *gin.Context) {
			adapter.Subscription().Suspend(c)
		})
		router.POST("/api/v1/subscriptions/:id/activate", func(c *gin.Context) {
			adapter.Subscription().Activate(c)
		})
		router.POST("/api/v1/subscriptions/:id/cancel", func(c *gin.Context) {
			adapter.Subscription().Cancel(c)
		})
		router.POST("/api/v1/subscriptions/:id/vacation", func(c *gin.Context) {
			adapter.Subscription().SetVacation(c)
		})

		Convey("List", func() {
			Convey("Success", func() {
				mockSubscriptionDatabasePort.EXPECT().FindByFilter(gomock.Any()).Return([]model.Subscription{subOutput}, nil).Times(1)

				req := httptest.NewRequest(http.MethodGet, "/api/v1/subscriptions", nil)
				w := httptest.NewRecorder()

				router.ServeHTTP(w, req)

				So(w.Code, ShouldEqual, http.StatusOK)

				var result model.Response
				json.Unmarshal(w.Body.Bytes(), &result)
				So(result.Success, ShouldBeTrue)
			})

			Convey("Domain error", func() {
				mockSubscriptionDatabasePort.EXPECT().FindByFilter(gomock.Any()).Return(nil, errors.New("error")).Times(1)

				req := httptest.NewRequest(http.MethodGet, "/api/v1/subscriptions", nil)
				w := httptest.NewRecorder()

				router.ServeHTTP(w, req)

				So(w.Code, ShouldEqual, http.StatusInternalServerError)
			})
		})

		Convey("Create", func() {
			payload := model.SubscriptionInput{
				CustomerID: "cust-123",
				PackageID:  "pkg-123",
				NasID:      "nas-123",
				StartDate:  time.Now(),
				EndDate:    time.Now().AddDate(0, 1, 0),
				Status:     model.SubscriptionStatusActive,
			}

			Convey("Invalid JSON", func() {
				req := httptest.NewRequest(http.MethodPost, "/api/v1/subscriptions", bytes.NewReader([]byte("invalid")))
				req.Header.Set("Content-Type", "application/json")
				w := httptest.NewRecorder()

				router.ServeHTTP(w, req)

				So(w.Code, ShouldEqual, http.StatusBadRequest)
			})

			Convey("Domain error - customer not found", func() {
				mockCustomerDatabasePort.EXPECT().FindByID(gomock.Any()).Return(model.Customer{}, errors.New("not found")).Times(1)

				body, _ := json.Marshal(payload)
				req := httptest.NewRequest(http.MethodPost, "/api/v1/subscriptions", bytes.NewReader(body))
				req.Header.Set("Content-Type", "application/json")
				w := httptest.NewRecorder()

				router.ServeHTTP(w, req)

				So(w.Code, ShouldEqual, http.StatusInternalServerError)
			})
		})

		Convey("Get", func() {
			Convey("Success", func() {
				mockSubscriptionDatabasePort.EXPECT().FindByID(gomock.Any()).Return(subOutput, nil).Times(1)

				req := httptest.NewRequest(http.MethodGet, "/api/v1/subscriptions/sub-123", nil)
				w := httptest.NewRecorder()

				router.ServeHTTP(w, req)

				So(w.Code, ShouldEqual, http.StatusOK)

				var result model.Response
				json.Unmarshal(w.Body.Bytes(), &result)
				So(result.Success, ShouldBeTrue)
			})

			Convey("Not found", func() {
				mockSubscriptionDatabasePort.EXPECT().FindByID(gomock.Any()).Return(model.Subscription{}, errors.New("not found")).Times(1)

				req := httptest.NewRequest(http.MethodGet, "/api/v1/subscriptions/sub-999", nil)
				w := httptest.NewRecorder()

				router.ServeHTTP(w, req)

				So(w.Code, ShouldEqual, http.StatusInternalServerError)
			})
		})

		Convey("Update", func() {
			payload := model.SubscriptionInput{
				Status: model.SubscriptionStatusActive,
			}

			Convey("Invalid JSON", func() {
				req := httptest.NewRequest(http.MethodPut, "/api/v1/subscriptions/sub-123", bytes.NewReader([]byte("invalid")))
				req.Header.Set("Content-Type", "application/json")
				w := httptest.NewRecorder()

				router.ServeHTTP(w, req)

				So(w.Code, ShouldEqual, http.StatusBadRequest)
			})

			Convey("Success", func() {
				mockSubscriptionDatabasePort.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil).Times(1)

				body, _ := json.Marshal(payload)
				req := httptest.NewRequest(http.MethodPut, "/api/v1/subscriptions/sub-123", bytes.NewReader(body))
				req.Header.Set("Content-Type", "application/json")
				w := httptest.NewRecorder()

				router.ServeHTTP(w, req)

				So(w.Code, ShouldEqual, http.StatusOK)

				var result model.Response
				json.Unmarshal(w.Body.Bytes(), &result)
				So(result.Success, ShouldBeTrue)
			})
		})

		Convey("Suspend", func() {
			Convey("Domain error", func() {
				mockSubscriptionDatabasePort.EXPECT().FindByID(gomock.Any()).Return(model.Subscription{}, errors.New("not found")).Times(1)

				req := httptest.NewRequest(http.MethodPost, "/api/v1/subscriptions/sub-999/suspend", nil)
				w := httptest.NewRecorder()

				router.ServeHTTP(w, req)

				So(w.Code, ShouldEqual, http.StatusInternalServerError)
			})
		})

		Convey("Activate", func() {
			Convey("Domain error", func() {
				mockSubscriptionDatabasePort.EXPECT().FindByID(gomock.Any()).Return(model.Subscription{}, errors.New("not found")).Times(1)

				req := httptest.NewRequest(http.MethodPost, "/api/v1/subscriptions/sub-999/activate", nil)
				w := httptest.NewRecorder()

				router.ServeHTTP(w, req)

				So(w.Code, ShouldEqual, http.StatusInternalServerError)
			})
		})

		Convey("Cancel", func() {
			Convey("Domain error", func() {
				mockSubscriptionDatabasePort.EXPECT().FindByID(gomock.Any()).Return(model.Subscription{}, errors.New("not found")).Times(1)

				req := httptest.NewRequest(http.MethodPost, "/api/v1/subscriptions/sub-999/cancel", nil)
				w := httptest.NewRecorder()

				router.ServeHTTP(w, req)

				So(w.Code, ShouldEqual, http.StatusInternalServerError)
			})
		})

		Convey("SetVacation", func() {
			Convey("Invalid JSON", func() {
				req := httptest.NewRequest(http.MethodPost, "/api/v1/subscriptions/sub-123/vacation", bytes.NewReader([]byte("invalid")))
				req.Header.Set("Content-Type", "application/json")
				w := httptest.NewRecorder()

				router.ServeHTTP(w, req)

				So(w.Code, ShouldEqual, http.StatusBadRequest)
			})
		})
	})
}
