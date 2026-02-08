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
	"mikrops/utils/crypto"
	mock_outbound_port "mikrops/tests/mocks/port"
)

func TestMikrotikAdapter(t *testing.T) {
	os.Setenv("ENCRYPTION_KEY", "01234567890123456789012345678901")

	Convey("Test Mikrotik HTTP Adapter", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		mockDatabasePort := mock_outbound_port.NewMockDatabasePort(mockCtrl)
		mockMessagePort := mock_outbound_port.NewMockMessagePort(mockCtrl)
		mockCachePort := mock_outbound_port.NewMockCachePort(mockCtrl)
		mockWorkflowPort := mock_outbound_port.NewMockWorkflowPort(mockCtrl)
		mockHttpPort := mock_outbound_port.NewMockHttpPort(mockCtrl)

		mockNasDatabasePort := mock_outbound_port.NewMockNasDatabasePort(mockCtrl)
		mockSubscriptionDatabasePort := mock_outbound_port.NewMockSubscriptionDatabasePort(mockCtrl)
		mockCustomerDatabasePort := mock_outbound_port.NewMockCustomerDatabasePort(mockCtrl)
		mockInternetPackageDatabasePort := mock_outbound_port.NewMockInternetPackageDatabasePort(mockCtrl)
		mockMikrotikPort := mock_outbound_port.NewMockMikrotikPort(mockCtrl)

		mockDatabasePort.EXPECT().Nas().Return(mockNasDatabasePort).AnyTimes()
		mockDatabasePort.EXPECT().Subscription().Return(mockSubscriptionDatabasePort).AnyTimes()
		mockDatabasePort.EXPECT().Customer().Return(mockCustomerDatabasePort).AnyTimes()
		mockDatabasePort.EXPECT().InternetPackage().Return(mockInternetPackageDatabasePort).AnyTimes()
		mockHttpPort.EXPECT().Mikrotik().Return(mockMikrotikPort).AnyTimes()

		dom := domain.NewDomain(mockDatabasePort, mockMessagePort, mockCachePort, mockWorkflowPort, mockHttpPort)
		adapter := gin_inbound_adapter.NewAdapter(dom, mockHttpPort)

		encryptedPassword, _ := crypto.Encrypt("password123")

		nasOutput := model.Nas{
			ID: "nas-123",
			NasInput: model.NasInput{
				TenantID:          "tenant-123",
				Name:              "Test NAS",
				Host:              "192.168.1.1",
				ApiPort:           8728,
				RestPort:          80,
				Username:          "admin",
				PasswordEncrypted: encryptedPassword,
				IsActive:          true,
				CreatedAt:         time.Now(),
				UpdatedAt:         time.Now(),
			},
		}

		router := gin.New()
		router.GET("/api/v1/nas/:id/pppoe/secrets", func(c *gin.Context) {
			adapter.Mikrotik().ListPPPoESecrets(c)
		})
		router.POST("/api/v1/nas/:id/pppoe/secrets", func(c *gin.Context) {
			adapter.Mikrotik().CreatePPPoESecret(c)
		})
		router.GET("/api/v1/nas/:id/queues", func(c *gin.Context) {
			adapter.Mikrotik().ListSimpleQueues(c)
		})
		router.POST("/api/v1/nas/:id/queues", func(c *gin.Context) {
			adapter.Mikrotik().CreateSimpleQueue(c)
		})
		router.GET("/api/v1/nas/:id/connections", func(c *gin.Context) {
			adapter.Mikrotik().GetActiveConnections(c)
		})

		Convey("ListPPPoESecrets", func() {
			Convey("Success", func() {
				mockNasDatabasePort.EXPECT().FindByID(gomock.Any()).Return(nasOutput, nil).Times(1)
				mockMikrotikPort.EXPECT().ListPPPoESecrets(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return([]model.MikrotikPPPoESecret{
					{ID: "s1", Name: "pppoe_test"},
				}, nil).Times(1)

				req := httptest.NewRequest(http.MethodGet, "/api/v1/nas/nas-123/pppoe/secrets", nil)
				w := httptest.NewRecorder()

				router.ServeHTTP(w, req)

				So(w.Code, ShouldEqual, http.StatusOK)

				var result model.Response
				json.Unmarshal(w.Body.Bytes(), &result)
				So(result.Success, ShouldBeTrue)
			})

			Convey("NAS not found", func() {
				mockNasDatabasePort.EXPECT().FindByID(gomock.Any()).Return(model.Nas{}, errors.New("not found")).Times(1)

				req := httptest.NewRequest(http.MethodGet, "/api/v1/nas/nas-999/pppoe/secrets", nil)
				w := httptest.NewRecorder()

				router.ServeHTTP(w, req)

				So(w.Code, ShouldEqual, http.StatusInternalServerError)
			})
		})

		Convey("CreatePPPoESecret", func() {
			payload := model.MikrotikPPPoESecretInput{
				Name:     "new_secret",
				Password: "secret123",
				Service:  "pppoe",
				Profile:  "default",
			}

			Convey("Success", func() {
				mockNasDatabasePort.EXPECT().FindByID(gomock.Any()).Return(nasOutput, nil).Times(1)
				mockMikrotikPort.EXPECT().CreatePPPoESecret(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(1)

				body, _ := json.Marshal(payload)
				req := httptest.NewRequest(http.MethodPost, "/api/v1/nas/nas-123/pppoe/secrets", bytes.NewReader(body))
				req.Header.Set("Content-Type", "application/json")
				w := httptest.NewRecorder()

				router.ServeHTTP(w, req)

				So(w.Code, ShouldEqual, http.StatusCreated)

				var result model.Response
				json.Unmarshal(w.Body.Bytes(), &result)
				So(result.Success, ShouldBeTrue)
			})

			Convey("Invalid JSON", func() {
				mockNasDatabasePort.EXPECT().FindByID(gomock.Any()).Return(nasOutput, nil).Times(1)

				req := httptest.NewRequest(http.MethodPost, "/api/v1/nas/nas-123/pppoe/secrets", bytes.NewReader([]byte("invalid")))
				req.Header.Set("Content-Type", "application/json")
				w := httptest.NewRecorder()

				router.ServeHTTP(w, req)

				So(w.Code, ShouldEqual, http.StatusBadRequest)
			})
		})

		Convey("ListSimpleQueues", func() {
			Convey("Success", func() {
				mockNasDatabasePort.EXPECT().FindByID(gomock.Any()).Return(nasOutput, nil).Times(1)
				mockMikrotikPort.EXPECT().ListSimpleQueues(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return([]model.MikrotikSimpleQueue{
					{ID: "q1", Name: "queue_test"},
				}, nil).Times(1)

				req := httptest.NewRequest(http.MethodGet, "/api/v1/nas/nas-123/queues", nil)
				w := httptest.NewRecorder()

				router.ServeHTTP(w, req)

				So(w.Code, ShouldEqual, http.StatusOK)

				var result model.Response
				json.Unmarshal(w.Body.Bytes(), &result)
				So(result.Success, ShouldBeTrue)
			})
		})

		Convey("CreateSimpleQueue", func() {
			payload := model.MikrotikSimpleQueueInput{
				Name:   "queue_new",
				Target: "192.168.1.100/32",
			}

			Convey("Success", func() {
				mockNasDatabasePort.EXPECT().FindByID(gomock.Any()).Return(nasOutput, nil).Times(1)
				mockMikrotikPort.EXPECT().CreateSimpleQueue(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(1)

				body, _ := json.Marshal(payload)
				req := httptest.NewRequest(http.MethodPost, "/api/v1/nas/nas-123/queues", bytes.NewReader(body))
				req.Header.Set("Content-Type", "application/json")
				w := httptest.NewRecorder()

				router.ServeHTTP(w, req)

				So(w.Code, ShouldEqual, http.StatusCreated)

				var result model.Response
				json.Unmarshal(w.Body.Bytes(), &result)
				So(result.Success, ShouldBeTrue)
			})
		})

		Convey("GetActiveConnections", func() {
			Convey("Success", func() {
				mockNasDatabasePort.EXPECT().FindByID(gomock.Any()).Return(nasOutput, nil).Times(1)
				mockMikrotikPort.EXPECT().GetActiveConnections(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return([]model.MikrotikConnection{}, nil).Times(1)

				req := httptest.NewRequest(http.MethodGet, "/api/v1/nas/nas-123/connections", nil)
				w := httptest.NewRecorder()

				router.ServeHTTP(w, req)

				So(w.Code, ShouldEqual, http.StatusOK)

				var result model.Response
				json.Unmarshal(w.Body.Bytes(), &result)
				So(result.Success, ShouldBeTrue)
			})

			Convey("NAS not found", func() {
				mockNasDatabasePort.EXPECT().FindByID(gomock.Any()).Return(model.Nas{}, errors.New("not found")).Times(1)

				req := httptest.NewRequest(http.MethodGet, "/api/v1/nas/nas-999/connections", nil)
				w := httptest.NewRecorder()

				router.ServeHTTP(w, req)

				So(w.Code, ShouldEqual, http.StatusInternalServerError)
			})
		})
	})
}
