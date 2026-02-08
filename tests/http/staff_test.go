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

func TestStaffAdapter(t *testing.T) {
	Convey("Test Staff HTTP Adapter", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		mockDatabasePort := mock_outbound_port.NewMockDatabasePort(mockCtrl)
		mockMessagePort := mock_outbound_port.NewMockMessagePort(mockCtrl)
		mockCachePort := mock_outbound_port.NewMockCachePort(mockCtrl)
		mockWorkflowPort := mock_outbound_port.NewMockWorkflowPort(mockCtrl)
		mockHttpPort := mock_outbound_port.NewMockHttpPort(mockCtrl)

		mockStaffDatabasePort := mock_outbound_port.NewMockStaffDatabasePort(mockCtrl)
		mockDatabasePort.EXPECT().Staff().Return(mockStaffDatabasePort).AnyTimes()

		dom := domain.NewDomain(mockDatabasePort, mockMessagePort, mockCachePort, mockWorkflowPort, mockHttpPort)
		adapter := gin_inbound_adapter.NewAdapter(dom, mockHttpPort)

		staffOutput := model.Staff{
			ID: "staff-123",
			StaffInput: model.StaffInput{
				TenantID: "tenant-123",
				RoleID:   1,
				Email:    "admin@test.com",
				FullName: "Admin User",
				Phone:    "081234567890",
				IsActive: true,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
		}

		router := gin.New()
		router.GET("/api/v1/staffs", func(c *gin.Context) {
			c.Set("tenant_id", "tenant-123")
			adapter.Staff().List(c)
		})
		router.POST("/api/v1/staffs", func(c *gin.Context) {
			c.Set("tenant_id", "tenant-123")
			adapter.Staff().Create(c)
		})
		router.GET("/api/v1/staffs/:id", func(c *gin.Context) {
			adapter.Staff().Get(c)
		})
		router.PUT("/api/v1/staffs/:id", func(c *gin.Context) {
			adapter.Staff().Update(c)
		})
		router.DELETE("/api/v1/staffs/:id", func(c *gin.Context) {
			adapter.Staff().Delete(c)
		})

		Convey("List", func() {
			Convey("Success", func() {
				mockStaffDatabasePort.EXPECT().FindByFilter(gomock.Any()).Return([]model.Staff{staffOutput}, nil).Times(1)

				req := httptest.NewRequest(http.MethodGet, "/api/v1/staffs", nil)
				w := httptest.NewRecorder()

				router.ServeHTTP(w, req)

				So(w.Code, ShouldEqual, http.StatusOK)

				var result model.Response
				json.Unmarshal(w.Body.Bytes(), &result)
				So(result.Success, ShouldBeTrue)
			})

			Convey("Domain error", func() {
				mockStaffDatabasePort.EXPECT().FindByFilter(gomock.Any()).Return(nil, errors.New("error")).Times(1)

				req := httptest.NewRequest(http.MethodGet, "/api/v1/staffs", nil)
				w := httptest.NewRecorder()

				router.ServeHTTP(w, req)

				So(w.Code, ShouldEqual, http.StatusInternalServerError)
			})
		})

		Convey("Create", func() {
			payload := model.StaffInput{
				RoleID:   1,
				Email:    "new@test.com",
				Password: "password123",
				FullName: "New Staff",
			}

			Convey("Success", func() {
				mockStaffDatabasePort.EXPECT().Create(gomock.Any()).Return(staffOutput, nil).Times(1)

				body, _ := json.Marshal(payload)
				req := httptest.NewRequest(http.MethodPost, "/api/v1/staffs", bytes.NewReader(body))
				req.Header.Set("Content-Type", "application/json")
				w := httptest.NewRecorder()

				router.ServeHTTP(w, req)

				So(w.Code, ShouldEqual, http.StatusCreated)

				var result model.Response
				json.Unmarshal(w.Body.Bytes(), &result)
				So(result.Success, ShouldBeTrue)
			})

			Convey("Invalid JSON", func() {
				req := httptest.NewRequest(http.MethodPost, "/api/v1/staffs", bytes.NewReader([]byte("invalid")))
				req.Header.Set("Content-Type", "application/json")
				w := httptest.NewRecorder()

				router.ServeHTTP(w, req)

				So(w.Code, ShouldEqual, http.StatusBadRequest)
			})

			Convey("Domain error", func() {
				mockStaffDatabasePort.EXPECT().Create(gomock.Any()).Return(model.Staff{}, errors.New("error")).Times(1)

				body, _ := json.Marshal(payload)
				req := httptest.NewRequest(http.MethodPost, "/api/v1/staffs", bytes.NewReader(body))
				req.Header.Set("Content-Type", "application/json")
				w := httptest.NewRecorder()

				router.ServeHTTP(w, req)

				So(w.Code, ShouldEqual, http.StatusInternalServerError)
			})
		})

		Convey("Get", func() {
			Convey("Success", func() {
				mockStaffDatabasePort.EXPECT().FindByID(gomock.Any()).Return(staffOutput, nil).Times(1)

				req := httptest.NewRequest(http.MethodGet, "/api/v1/staffs/staff-123", nil)
				w := httptest.NewRecorder()

				router.ServeHTTP(w, req)

				So(w.Code, ShouldEqual, http.StatusOK)

				var result model.Response
				json.Unmarshal(w.Body.Bytes(), &result)
				So(result.Success, ShouldBeTrue)
			})

			Convey("Not found", func() {
				mockStaffDatabasePort.EXPECT().FindByID(gomock.Any()).Return(model.Staff{}, errors.New("not found")).Times(1)

				req := httptest.NewRequest(http.MethodGet, "/api/v1/staffs/staff-999", nil)
				w := httptest.NewRecorder()

				router.ServeHTTP(w, req)

				So(w.Code, ShouldEqual, http.StatusInternalServerError)
			})
		})

		Convey("Update", func() {
			payload := model.StaffInput{
				FullName: "Updated Name",
			}

			Convey("Success", func() {
				mockStaffDatabasePort.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil).Times(1)

				body, _ := json.Marshal(payload)
				req := httptest.NewRequest(http.MethodPut, "/api/v1/staffs/staff-123", bytes.NewReader(body))
				req.Header.Set("Content-Type", "application/json")
				w := httptest.NewRecorder()

				router.ServeHTTP(w, req)

				So(w.Code, ShouldEqual, http.StatusOK)

				var result model.Response
				json.Unmarshal(w.Body.Bytes(), &result)
				So(result.Success, ShouldBeTrue)
			})

			Convey("Invalid JSON", func() {
				req := httptest.NewRequest(http.MethodPut, "/api/v1/staffs/staff-123", bytes.NewReader([]byte("invalid")))
				req.Header.Set("Content-Type", "application/json")
				w := httptest.NewRecorder()

				router.ServeHTTP(w, req)

				So(w.Code, ShouldEqual, http.StatusBadRequest)
			})

			Convey("Domain error", func() {
				mockStaffDatabasePort.EXPECT().Update(gomock.Any(), gomock.Any()).Return(errors.New("error")).Times(1)

				body, _ := json.Marshal(payload)
				req := httptest.NewRequest(http.MethodPut, "/api/v1/staffs/staff-123", bytes.NewReader(body))
				req.Header.Set("Content-Type", "application/json")
				w := httptest.NewRecorder()

				router.ServeHTTP(w, req)

				So(w.Code, ShouldEqual, http.StatusInternalServerError)
			})
		})

		Convey("Delete", func() {
			Convey("Success", func() {
				mockStaffDatabasePort.EXPECT().Delete(gomock.Any()).Return(nil).Times(1)

				req := httptest.NewRequest(http.MethodDelete, "/api/v1/staffs/staff-123", nil)
				w := httptest.NewRecorder()

				router.ServeHTTP(w, req)

				So(w.Code, ShouldEqual, http.StatusOK)

				var result model.Response
				json.Unmarshal(w.Body.Bytes(), &result)
				So(result.Success, ShouldBeTrue)
			})

			Convey("Domain error", func() {
				mockStaffDatabasePort.EXPECT().Delete(gomock.Any()).Return(errors.New("error")).Times(1)

				req := httptest.NewRequest(http.MethodDelete, "/api/v1/staffs/staff-999", nil)
				w := httptest.NewRecorder()

				router.ServeHTTP(w, req)

				So(w.Code, ShouldEqual, http.StatusInternalServerError)
			})
		})
	})
}
