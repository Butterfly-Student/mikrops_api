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

func TestNasAdapter(t *testing.T) {
	os.Setenv("ENCRYPTION_KEY", "01234567890123456789012345678901")

	Convey("Test NAS HTTP Adapter", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		mockDatabasePort := mock_outbound_port.NewMockDatabasePort(mockCtrl)
		mockMessagePort := mock_outbound_port.NewMockMessagePort(mockCtrl)
		mockCachePort := mock_outbound_port.NewMockCachePort(mockCtrl)
		mockWorkflowPort := mock_outbound_port.NewMockWorkflowPort(mockCtrl)
		mockHttpPort := mock_outbound_port.NewMockHttpPort(mockCtrl)

		mockNasDatabasePort := mock_outbound_port.NewMockNasDatabasePort(mockCtrl)
		mockTenantDatabasePort := mock_outbound_port.NewMockTenantDatabasePort(mockCtrl)
		mockMikrotikPort := mock_outbound_port.NewMockMikrotikPort(mockCtrl)

		mockDatabasePort.EXPECT().Nas().Return(mockNasDatabasePort).AnyTimes()
		mockDatabasePort.EXPECT().Tenant().Return(mockTenantDatabasePort).AnyTimes()
		mockHttpPort.EXPECT().Mikrotik().Return(mockMikrotikPort).AnyTimes()

		dom := domain.NewDomain(mockDatabasePort, mockMessagePort, mockCachePort, mockWorkflowPort, mockHttpPort)
		adapter := gin_inbound_adapter.NewAdapter(dom, mockHttpPort)

		nasOutput := model.Nas{
			ID: "nas-123",
			NasInput: model.NasInput{
				TenantID:  "tenant-123",
				Name:      "Test NAS",
				Host:      "192.168.1.1",
				ApiPort:   8728,
				RestPort:  80,
				Username:  "admin",
				IsActive:  true,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
		}

		router := gin.New()
		router.GET("/api/v1/nas", func(c *gin.Context) {
			c.Set("tenant_id", "tenant-123")
			adapter.Nas().List(c)
		})
		router.POST("/api/v1/nas", func(c *gin.Context) {
			c.Set("tenant_id", "tenant-123")
			adapter.Nas().Create(c)
		})
		router.GET("/api/v1/nas/:id", func(c *gin.Context) {
			adapter.Nas().Get(c)
		})
		router.PUT("/api/v1/nas/:id", func(c *gin.Context) {
			adapter.Nas().Update(c)
		})
		router.DELETE("/api/v1/nas/:id", func(c *gin.Context) {
			adapter.Nas().Delete(c)
		})
		router.POST("/api/v1/nas/:id/test", func(c *gin.Context) {
			adapter.Nas().TestConnection(c)
		})

		Convey("List", func() {
			Convey("Success", func() {
				mockNasDatabasePort.EXPECT().FindByFilter(gomock.Any()).Return([]model.Nas{nasOutput}, nil).Times(1)

				req := httptest.NewRequest(http.MethodGet, "/api/v1/nas", nil)
				w := httptest.NewRecorder()

				router.ServeHTTP(w, req)

				So(w.Code, ShouldEqual, http.StatusOK)

				var result model.Response
				json.Unmarshal(w.Body.Bytes(), &result)
				So(result.Success, ShouldBeTrue)
			})

			Convey("Domain error", func() {
				mockNasDatabasePort.EXPECT().FindByFilter(gomock.Any()).Return(nil, errors.New("error")).Times(1)

				req := httptest.NewRequest(http.MethodGet, "/api/v1/nas", nil)
				w := httptest.NewRecorder()

				router.ServeHTTP(w, req)

				So(w.Code, ShouldEqual, http.StatusInternalServerError)
			})
		})

		Convey("Create", func() {
			payload := model.NasInput{
				Name:     "New NAS",
				Host:     "192.168.1.2",
				ApiPort:  8728,
				RestPort: 80,
				Username: "admin",
				Password: "password123",
				IsActive: true,
			}

			Convey("Success", func() {
				mockNasDatabasePort.EXPECT().CountByTenantID(gomock.Any()).Return(1, nil).Times(1)
				mockTenantDatabasePort.EXPECT().FindByID(gomock.Any()).Return(model.Tenant{
					ID:          "tenant-123",
					TenantInput: model.TenantInput{MaxNas: 3},
				}, nil).Times(1)
				mockNasDatabasePort.EXPECT().Create(gomock.Any()).Return(nasOutput, nil).Times(1)

				body, _ := json.Marshal(payload)
				req := httptest.NewRequest(http.MethodPost, "/api/v1/nas", bytes.NewReader(body))
				req.Header.Set("Content-Type", "application/json")
				w := httptest.NewRecorder()

				router.ServeHTTP(w, req)

				So(w.Code, ShouldEqual, http.StatusCreated)

				var result model.Response
				json.Unmarshal(w.Body.Bytes(), &result)
				So(result.Success, ShouldBeTrue)
			})

			Convey("Invalid JSON", func() {
				req := httptest.NewRequest(http.MethodPost, "/api/v1/nas", bytes.NewReader([]byte("invalid")))
				req.Header.Set("Content-Type", "application/json")
				w := httptest.NewRecorder()

				router.ServeHTTP(w, req)

				So(w.Code, ShouldEqual, http.StatusBadRequest)
			})

			Convey("Domain error", func() {
				mockNasDatabasePort.EXPECT().CountByTenantID(gomock.Any()).Return(0, errors.New("error")).Times(1)

				body, _ := json.Marshal(payload)
				req := httptest.NewRequest(http.MethodPost, "/api/v1/nas", bytes.NewReader(body))
				req.Header.Set("Content-Type", "application/json")
				w := httptest.NewRecorder()

				router.ServeHTTP(w, req)

				So(w.Code, ShouldEqual, http.StatusInternalServerError)
			})
		})

		Convey("Get", func() {
			Convey("Success", func() {
				mockNasDatabasePort.EXPECT().FindByID(gomock.Any()).Return(nasOutput, nil).Times(1)

				req := httptest.NewRequest(http.MethodGet, "/api/v1/nas/nas-123", nil)
				w := httptest.NewRecorder()

				router.ServeHTTP(w, req)

				So(w.Code, ShouldEqual, http.StatusOK)

				var result model.Response
				json.Unmarshal(w.Body.Bytes(), &result)
				So(result.Success, ShouldBeTrue)
			})

			Convey("Not found", func() {
				mockNasDatabasePort.EXPECT().FindByID(gomock.Any()).Return(model.Nas{}, errors.New("not found")).Times(1)

				req := httptest.NewRequest(http.MethodGet, "/api/v1/nas/nas-999", nil)
				w := httptest.NewRecorder()

				router.ServeHTTP(w, req)

				So(w.Code, ShouldEqual, http.StatusInternalServerError)
			})
		})

		Convey("Update", func() {
			payload := model.NasInput{
				Name: "Updated NAS",
				Host: "192.168.1.3",
			}

			Convey("Success", func() {
				mockNasDatabasePort.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil).Times(1)

				body, _ := json.Marshal(payload)
				req := httptest.NewRequest(http.MethodPut, "/api/v1/nas/nas-123", bytes.NewReader(body))
				req.Header.Set("Content-Type", "application/json")
				w := httptest.NewRecorder()

				router.ServeHTTP(w, req)

				So(w.Code, ShouldEqual, http.StatusOK)

				var result model.Response
				json.Unmarshal(w.Body.Bytes(), &result)
				So(result.Success, ShouldBeTrue)
			})

			Convey("Invalid JSON", func() {
				req := httptest.NewRequest(http.MethodPut, "/api/v1/nas/nas-123", bytes.NewReader([]byte("invalid")))
				req.Header.Set("Content-Type", "application/json")
				w := httptest.NewRecorder()

				router.ServeHTTP(w, req)

				So(w.Code, ShouldEqual, http.StatusBadRequest)
			})
		})

		Convey("Delete", func() {
			Convey("Success", func() {
				mockNasDatabasePort.EXPECT().Delete(gomock.Any()).Return(nil).Times(1)

				req := httptest.NewRequest(http.MethodDelete, "/api/v1/nas/nas-123", nil)
				w := httptest.NewRecorder()

				router.ServeHTTP(w, req)

				So(w.Code, ShouldEqual, http.StatusOK)

				var result model.Response
				json.Unmarshal(w.Body.Bytes(), &result)
				So(result.Success, ShouldBeTrue)
			})
		})

		Convey("TestConnection", func() {
			Convey("NAS not found", func() {
				mockNasDatabasePort.EXPECT().FindByID(gomock.Any()).Return(model.Nas{}, errors.New("not found")).Times(1)

				req := httptest.NewRequest(http.MethodPost, "/api/v1/nas/nas-999/test", nil)
				w := httptest.NewRecorder()

				router.ServeHTTP(w, req)

				So(w.Code, ShouldEqual, http.StatusInternalServerError)
			})
		})
	})
}
