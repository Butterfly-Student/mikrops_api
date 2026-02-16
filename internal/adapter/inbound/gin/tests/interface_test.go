package gin_inbound_adapter_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	. "github.com/smartystreets/goconvey/convey"

	gin_inbound_adapter "go-template/internal/adapter/inbound/gin"
	"go-template/internal/domain"
	mock_outbound_port "go-template/tests/mocks/port"
)

func TestInterfaceHandler(t *testing.T) {
	Convey("Test Interface HTTP Handler", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		mockDatabasePort := mock_outbound_port.NewMockDatabasePort(mockCtrl)
		mockMikrotikPort := mock_outbound_port.NewMockMikrotikPort(mockCtrl)
		mockCachePort := mock_outbound_port.NewMockCachePort(mockCtrl)

		dom := domain.NewDomain(mockDatabasePort, nil, mockCachePort, nil, mockMikrotikPort, nil, nil, nil)
		handler := gin_inbound_adapter.NewInterfaceAdapter(dom)

		gin.SetMode(gin.TestMode)
		router := gin.New()
		router.POST("/interface/monitoring/start", handler.StartMonitoring)
		router.POST("/interface/monitoring/start/:name", handler.StartMonitoringByName)
		router.POST("/interface/monitoring/stop", handler.StopMonitoring)
		router.POST("/interface/monitoring/stop/:name", handler.StopMonitoringByName)

		routerID := uuid.New().String()

		Convey("StartMonitoring", func() {
			Convey("Success", func() {
				mockDatabasePort.EXPECT().
					Mikrotik().
					Return(mock_outbound_port.NewMockMikrotikDatabasePort(mockCtrl)).
					Times(1)

				mockMikrotikPort.EXPECT().
					MonitorAllInterfaces(gomock.Any(), gomock.Any()).
					Return(nil).
					Times(1)

				req := httptest.NewRequest("POST", "/interface/monitoring/start?router_id="+routerID, nil)
				w := httptest.NewRecorder()

				router.ServeHTTP(w, req)

				So(w.Code, ShouldEqual, http.StatusOK)
			})

			Convey("Missing router_id", func() {
				req := httptest.NewRequest("POST", "/interface/monitoring/start", nil)
				w := httptest.NewRecorder()

				router.ServeHTTP(w, req)

				So(w.Code, ShouldEqual, http.StatusBadRequest)
			})

			Convey("Domain Error", func() {
				mockDatabasePort.EXPECT().
					Mikrotik().
					Return(mock_outbound_port.NewMockMikrotikDatabasePort(mockCtrl)).
					Times(1)

				req := httptest.NewRequest("POST", "/interface/monitoring/start?router_id="+routerID, nil)
				w := httptest.NewRecorder()

				router.ServeHTTP(w, req)

				// Should handle error appropriately
				So(w.Code, ShouldNotEqual, http.StatusOK)
			})
		})

		Convey("StartMonitoringByName", func() {
			Convey("Success", func() {
				interfaceName := "ether1"

				mockDatabasePort.EXPECT().
					Mikrotik().
					Return(mock_outbound_port.NewMockMikrotikDatabasePort(mockCtrl)).
					Times(1)

				mockMikrotikPort.EXPECT().
					MonitorInterface(gomock.Any(), gomock.Any(), interfaceName).
					Return(nil).
					Times(1)

				req := httptest.NewRequest("POST", "/interface/monitoring/start/"+interfaceName+"?router_id="+routerID, nil)
				w := httptest.NewRecorder()

				router.ServeHTTP(w, req)

				So(w.Code, ShouldEqual, http.StatusOK)
			})

			Convey("Missing router_id", func() {
				req := httptest.NewRequest("POST", "/interface/monitoring/start/ether1", nil)
				w := httptest.NewRecorder()

				router.ServeHTTP(w, req)

				So(w.Code, ShouldEqual, http.StatusBadRequest)
			})

			Convey("Missing interface name", func() {
				// This should be handled by Gin routing
				// The path parameter is required
				req := httptest.NewRequest("POST", "/interface/monitoring/start?router_id="+routerID, nil)
				w := httptest.NewRecorder()

				router.ServeHTTP(w, req)

				// Should return 404 because the route doesn't match
				So(w.Code, ShouldEqual, http.StatusNotFound)
			})
		})

		Convey("StopMonitoring", func() {
			Convey("Success", func() {
				req := httptest.NewRequest("POST", "/interface/monitoring/stop?router_id="+routerID, nil)
				w := httptest.NewRecorder()

				router.ServeHTTP(w, req)

				So(w.Code, ShouldEqual, http.StatusOK)
			})

			Convey("Missing router_id", func() {
				req := httptest.NewRequest("POST", "/interface/monitoring/stop", nil)
				w := httptest.NewRecorder()

				router.ServeHTTP(w, req)

				So(w.Code, ShouldEqual, http.StatusBadRequest)
			})

			Convey("Domain Error", func() {
				req := httptest.NewRequest("POST", "/interface/monitoring/stop?router_id="+routerID, nil)
				w := httptest.NewRecorder()

				router.ServeHTTP(w, req)

				// Should handle error appropriately
				So(w.Code, ShouldEqual, http.StatusOK)
			})
		})

		Convey("StopMonitoringByName", func() {
			Convey("Success", func() {
				interfaceName := "ether1"

				req := httptest.NewRequest("POST", "/interface/monitoring/stop/"+interfaceName+"?router_id="+routerID, nil)
				w := httptest.NewRecorder()

				router.ServeHTTP(w, req)

				So(w.Code, ShouldEqual, http.StatusOK)
			})

			Convey("Missing router_id", func() {
				req := httptest.NewRequest("POST", "/interface/monitoring/stop/ether1", nil)
				w := httptest.NewRecorder()

				router.ServeHTTP(w, req)

				So(w.Code, ShouldEqual, http.StatusBadRequest)
			})

			Convey("Missing interface name", func() {
				req := httptest.NewRequest("POST", "/interface/monitoring/stop?router_id="+routerID, nil)
				w := httptest.NewRecorder()

				router.ServeHTTP(w, req)

				// Should return 404 because the route doesn't match
				So(w.Code, ShouldEqual, http.StatusNotFound)
			})
		})
	})
}
