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

func TestTenantAdapter(t *testing.T) {
	Convey("Test Tenant HTTP Adapter", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		mockDatabasePort := mock_outbound_port.NewMockDatabasePort(mockCtrl)
		mockMessagePort := mock_outbound_port.NewMockMessagePort(mockCtrl)
		mockCachePort := mock_outbound_port.NewMockCachePort(mockCtrl)
		mockWorkflowPort := mock_outbound_port.NewMockWorkflowPort(mockCtrl)
		mockHttpPort := mock_outbound_port.NewMockHttpPort(mockCtrl)

		mockTenantDatabasePort := mock_outbound_port.NewMockTenantDatabasePort(mockCtrl)
		mockDatabasePort.EXPECT().Tenant().Return(mockTenantDatabasePort).AnyTimes()

		dom := domain.NewDomain(mockDatabasePort, mockMessagePort, mockCachePort, mockWorkflowPort, mockHttpPort)
		adapter := gin_inbound_adapter.NewAdapter(dom, mockHttpPort)

		tenantOutput := model.Tenant{
			ID: "tenant-123",
			TenantInput: model.TenantInput{
				Name:             "Test ISP",
				Slug:             "test-isp",
				Email:            "admin@testisp.com",
				MaxNas:           3,
				SubscriptionPlan: "basic",
				IsActive:         true,
				CreatedAt:        time.Now(),
				UpdatedAt:        time.Now(),
			},
		}

		router := gin.New()
		router.GET("/api/v1/tenant", func(c *gin.Context) {
			c.Set("tenant_id", "tenant-123")
			adapter.Tenant().Get(c)
		})
		router.PUT("/api/v1/tenant", func(c *gin.Context) {
			c.Set("tenant_id", "tenant-123")
			adapter.Tenant().Update(c)
		})

		Convey("Get", func() {
			Convey("Success", func() {
				mockTenantDatabasePort.EXPECT().FindByID(gomock.Any()).Return(tenantOutput, nil).Times(1)

				req := httptest.NewRequest(http.MethodGet, "/api/v1/tenant", nil)
				w := httptest.NewRecorder()

				router.ServeHTTP(w, req)

				So(w.Code, ShouldEqual, http.StatusOK)

				var result model.Response
				json.Unmarshal(w.Body.Bytes(), &result)
				So(result.Success, ShouldBeTrue)
			})

			Convey("Domain error", func() {
				mockTenantDatabasePort.EXPECT().FindByID(gomock.Any()).Return(model.Tenant{}, errors.New("not found")).Times(1)

				req := httptest.NewRequest(http.MethodGet, "/api/v1/tenant", nil)
				w := httptest.NewRecorder()

				router.ServeHTTP(w, req)

				So(w.Code, ShouldEqual, http.StatusInternalServerError)
			})
		})

		Convey("Update", func() {
			payload := model.TenantInput{
				Name:  "Updated ISP",
				Email: "new@testisp.com",
			}

			Convey("Success", func() {
				mockTenantDatabasePort.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil).Times(1)

				body, _ := json.Marshal(payload)
				req := httptest.NewRequest(http.MethodPut, "/api/v1/tenant", bytes.NewReader(body))
				req.Header.Set("Content-Type", "application/json")
				w := httptest.NewRecorder()

				router.ServeHTTP(w, req)

				So(w.Code, ShouldEqual, http.StatusOK)

				var result model.Response
				json.Unmarshal(w.Body.Bytes(), &result)
				So(result.Success, ShouldBeTrue)
			})

			Convey("Invalid JSON", func() {
				req := httptest.NewRequest(http.MethodPut, "/api/v1/tenant", bytes.NewReader([]byte("invalid")))
				req.Header.Set("Content-Type", "application/json")
				w := httptest.NewRecorder()

				router.ServeHTTP(w, req)

				So(w.Code, ShouldEqual, http.StatusBadRequest)
			})

			Convey("Domain error", func() {
				mockTenantDatabasePort.EXPECT().Update(gomock.Any(), gomock.Any()).Return(errors.New("error")).Times(1)

				body, _ := json.Marshal(payload)
				req := httptest.NewRequest(http.MethodPut, "/api/v1/tenant", bytes.NewReader(body))
				req.Header.Set("Content-Type", "application/json")
				w := httptest.NewRecorder()

				router.ServeHTTP(w, req)

				So(w.Code, ShouldEqual, http.StatusInternalServerError)
			})
		})
	})
}
