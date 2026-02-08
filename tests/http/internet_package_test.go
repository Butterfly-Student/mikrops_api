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

func TestInternetPackageAdapter(t *testing.T) {
	Convey("Test Internet Package HTTP Adapter", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		mockDatabasePort := mock_outbound_port.NewMockDatabasePort(mockCtrl)
		mockMessagePort := mock_outbound_port.NewMockMessagePort(mockCtrl)
		mockCachePort := mock_outbound_port.NewMockCachePort(mockCtrl)
		mockWorkflowPort := mock_outbound_port.NewMockWorkflowPort(mockCtrl)
		mockHttpPort := mock_outbound_port.NewMockHttpPort(mockCtrl)

		mockInternetPackageDatabasePort := mock_outbound_port.NewMockInternetPackageDatabasePort(mockCtrl)
		mockDatabasePort.EXPECT().InternetPackage().Return(mockInternetPackageDatabasePort).AnyTimes()

		dom := domain.NewDomain(mockDatabasePort, mockMessagePort, mockCachePort, mockWorkflowPort, mockHttpPort)
		adapter := gin_inbound_adapter.NewAdapter(dom, mockHttpPort)

		pkgOutput := model.InternetPackage{
			ID: "pkg-123",
			InternetPackageInput: model.InternetPackageInput{
				TenantID:     "tenant-123",
				Name:         "Basic 10Mbps",
				Type:         "pppoe",
				UploadRate:   "10M",
				DownloadRate: "10M",
				Price:        150000,
				BillingCycle: "monthly",
				ValidityDays: 30,
				IsActive:     true,
				CreatedAt:    time.Now(),
				UpdatedAt:    time.Now(),
			},
		}

		router := gin.New()
		router.GET("/api/v1/packages", func(c *gin.Context) {
			c.Set("tenant_id", "tenant-123")
			adapter.InternetPackage().List(c)
		})
		router.POST("/api/v1/packages", func(c *gin.Context) {
			c.Set("tenant_id", "tenant-123")
			adapter.InternetPackage().Create(c)
		})
		router.GET("/api/v1/packages/:id", func(c *gin.Context) {
			adapter.InternetPackage().Get(c)
		})
		router.PUT("/api/v1/packages/:id", func(c *gin.Context) {
			adapter.InternetPackage().Update(c)
		})
		router.DELETE("/api/v1/packages/:id", func(c *gin.Context) {
			adapter.InternetPackage().Delete(c)
		})

		Convey("List", func() {
			Convey("Success", func() {
				mockInternetPackageDatabasePort.EXPECT().FindByFilter(gomock.Any()).Return([]model.InternetPackage{pkgOutput}, nil).Times(1)

				req := httptest.NewRequest(http.MethodGet, "/api/v1/packages", nil)
				w := httptest.NewRecorder()

				router.ServeHTTP(w, req)

				So(w.Code, ShouldEqual, http.StatusOK)

				var result model.Response
				json.Unmarshal(w.Body.Bytes(), &result)
				So(result.Success, ShouldBeTrue)
			})

			Convey("Domain error", func() {
				mockInternetPackageDatabasePort.EXPECT().FindByFilter(gomock.Any()).Return(nil, errors.New("error")).Times(1)

				req := httptest.NewRequest(http.MethodGet, "/api/v1/packages", nil)
				w := httptest.NewRecorder()

				router.ServeHTTP(w, req)

				So(w.Code, ShouldEqual, http.StatusInternalServerError)
			})
		})

		Convey("Create", func() {
			payload := model.InternetPackageInput{
				Name:         "Premium 50Mbps",
				Type:         "pppoe",
				UploadRate:   "50M",
				DownloadRate: "50M",
				Price:        350000,
				BillingCycle: "monthly",
			}

			Convey("Success", func() {
				mockInternetPackageDatabasePort.EXPECT().Create(gomock.Any()).Return(pkgOutput, nil).Times(1)

				body, _ := json.Marshal(payload)
				req := httptest.NewRequest(http.MethodPost, "/api/v1/packages", bytes.NewReader(body))
				req.Header.Set("Content-Type", "application/json")
				w := httptest.NewRecorder()

				router.ServeHTTP(w, req)

				So(w.Code, ShouldEqual, http.StatusCreated)

				var result model.Response
				json.Unmarshal(w.Body.Bytes(), &result)
				So(result.Success, ShouldBeTrue)
			})

			Convey("Invalid JSON", func() {
				req := httptest.NewRequest(http.MethodPost, "/api/v1/packages", bytes.NewReader([]byte("invalid")))
				req.Header.Set("Content-Type", "application/json")
				w := httptest.NewRecorder()

				router.ServeHTTP(w, req)

				So(w.Code, ShouldEqual, http.StatusBadRequest)
			})

			Convey("Domain error", func() {
				mockInternetPackageDatabasePort.EXPECT().Create(gomock.Any()).Return(model.InternetPackage{}, errors.New("error")).Times(1)

				body, _ := json.Marshal(payload)
				req := httptest.NewRequest(http.MethodPost, "/api/v1/packages", bytes.NewReader(body))
				req.Header.Set("Content-Type", "application/json")
				w := httptest.NewRecorder()

				router.ServeHTTP(w, req)

				So(w.Code, ShouldEqual, http.StatusInternalServerError)
			})
		})

		Convey("Get", func() {
			Convey("Success", func() {
				mockInternetPackageDatabasePort.EXPECT().FindByID(gomock.Any()).Return(pkgOutput, nil).Times(1)

				req := httptest.NewRequest(http.MethodGet, "/api/v1/packages/pkg-123", nil)
				w := httptest.NewRecorder()

				router.ServeHTTP(w, req)

				So(w.Code, ShouldEqual, http.StatusOK)

				var result model.Response
				json.Unmarshal(w.Body.Bytes(), &result)
				So(result.Success, ShouldBeTrue)
			})

			Convey("Not found", func() {
				mockInternetPackageDatabasePort.EXPECT().FindByID(gomock.Any()).Return(model.InternetPackage{}, errors.New("not found")).Times(1)

				req := httptest.NewRequest(http.MethodGet, "/api/v1/packages/pkg-999", nil)
				w := httptest.NewRecorder()

				router.ServeHTTP(w, req)

				So(w.Code, ShouldEqual, http.StatusInternalServerError)
			})
		})

		Convey("Update", func() {
			payload := model.InternetPackageInput{
				Name:  "Updated Package",
				Price: 200000,
			}

			Convey("Success", func() {
				mockInternetPackageDatabasePort.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil).Times(1)

				body, _ := json.Marshal(payload)
				req := httptest.NewRequest(http.MethodPut, "/api/v1/packages/pkg-123", bytes.NewReader(body))
				req.Header.Set("Content-Type", "application/json")
				w := httptest.NewRecorder()

				router.ServeHTTP(w, req)

				So(w.Code, ShouldEqual, http.StatusOK)

				var result model.Response
				json.Unmarshal(w.Body.Bytes(), &result)
				So(result.Success, ShouldBeTrue)
			})

			Convey("Invalid JSON", func() {
				req := httptest.NewRequest(http.MethodPut, "/api/v1/packages/pkg-123", bytes.NewReader([]byte("invalid")))
				req.Header.Set("Content-Type", "application/json")
				w := httptest.NewRecorder()

				router.ServeHTTP(w, req)

				So(w.Code, ShouldEqual, http.StatusBadRequest)
			})
		})

		Convey("Delete", func() {
			Convey("Success", func() {
				mockInternetPackageDatabasePort.EXPECT().Delete(gomock.Any()).Return(nil).Times(1)

				req := httptest.NewRequest(http.MethodDelete, "/api/v1/packages/pkg-123", nil)
				w := httptest.NewRecorder()

				router.ServeHTTP(w, req)

				So(w.Code, ShouldEqual, http.StatusOK)

				var result model.Response
				json.Unmarshal(w.Body.Bytes(), &result)
				So(result.Success, ShouldBeTrue)
			})

			Convey("Domain error", func() {
				mockInternetPackageDatabasePort.EXPECT().Delete(gomock.Any()).Return(errors.New("error")).Times(1)

				req := httptest.NewRequest(http.MethodDelete, "/api/v1/packages/pkg-999", nil)
				w := httptest.NewRecorder()

				router.ServeHTTP(w, req)

				So(w.Code, ShouldEqual, http.StatusInternalServerError)
			})
		})
	})
}
