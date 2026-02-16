package gin_inbound_adapter_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	. "github.com/smartystreets/goconvey/convey"

	gin_inbound_adapter "go-template/internal/adapter/inbound/gin"
	"go-template/internal/domain"
	"go-template/internal/model"
	mock_outbound_port "go-template/tests/mocks/port"
)

func TestPingHandler(t *testing.T) {
	Convey("Test Ping HTTP Handler", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		mockDatabasePort := mock_outbound_port.NewMockDatabasePort(mockCtrl)
		mockMikrotikPort := mock_outbound_port.NewMockMikrotikPort(mockCtrl)
		mockMikrotikDB := mock_outbound_port.NewMockMikrotikDatabasePort(mockCtrl)

		mockDatabasePort.EXPECT().Mikrotik().Return(mockMikrotikDB).AnyTimes()

		dom := domain.NewDomain(mockDatabasePort, nil, nil, nil, mockMikrotikPort, nil, nil, nil)
		handler := gin_inbound_adapter.NewPingAdapter(dom)

		gin.SetMode(gin.TestMode)
		router := gin.New()
		router.GET("/ping/resource", handler.GetResource)
		router.POST("/ping/start", handler.StartPing)
		router.POST("/ping/stop", handler.StopPing)
		router.GET("/ping/ws", handler.HandleWebSocket)

		routerID := uuid.New().String()

		Convey("GetResource", func() {
			Convey("Success", func() {
				req := httptest.NewRequest("GET", "/ping/resource", nil)
				w := httptest.NewRecorder()

				router.ServeHTTP(w, req)

				So(w.Code, ShouldEqual, http.StatusOK)
			})
		})

		Convey("StartPing", func() {
			Convey("Success", func() {
				address := "192.168.1.1"
				count := 5
				interval := "1s"

				reqBody := model.PingRequest{
					Address:  address,
					Count:    count,
					Interval: interval,
				}

				bodyBytes, _ := json.Marshal(reqBody)
				req := httptest.NewRequest("POST", "/ping/start?router_id="+routerID, bytes.NewBuffer(bodyBytes))
				req.Header.Set("Content-Type", "application/json")
				w := httptest.NewRecorder()

				mockMikrotikDB.EXPECT().
					FindByID(routerID).
					Return(&model.MikrotikRouter{}, nil).
					Times(1)

				mockMikrotikPort.EXPECT().
					Ping(gomock.Any(), gomock.Any(), reqBody).
					Return(nil, nil).
					Times(1)

				router.ServeHTTP(w, req)

				So(w.Code, ShouldEqual, http.StatusOK)
			})

			Convey("Missing router_id", func() {
				reqBody := model.PingRequest{
					Address: "192.168.1.1",
				}

				bodyBytes, _ := json.Marshal(reqBody)
				req := httptest.NewRequest("POST", "/ping/start", bytes.NewBuffer(bodyBytes))
				req.Header.Set("Content-Type", "application/json")
				w := httptest.NewRecorder()

				router.ServeHTTP(w, req)

				So(w.Code, ShouldEqual, http.StatusBadRequest)
			})

			Convey("Invalid JSON", func() {
				req := httptest.NewRequest("POST", "/ping/start?router_id="+routerID, bytes.NewBuffer([]byte("invalid json")))
				req.Header.Set("Content-Type", "application/json")
				w := httptest.NewRecorder()

				router.ServeHTTP(w, req)

				So(w.Code, ShouldEqual, http.StatusBadRequest)
			})

			Convey("Domain Error", func() {
				address := "192.168.1.1"
				count := 5

				reqBody := model.PingRequest{
					Address: address,
					Count:   count,
				}

				bodyBytes, _ := json.Marshal(reqBody)
				req := httptest.NewRequest("POST", "/ping/start?router_id="+routerID, bytes.NewBuffer(bodyBytes))
				req.Header.Set("Content-Type", "application/json")
				w := httptest.NewRecorder()

				mockMikrotikDB.EXPECT().
					FindByID(routerID).
					Return(nil, errors.New("router not found")).
					Times(1)

				router.ServeHTTP(w, req)

				// Should handle error appropriately
				So(w.Code, ShouldNotEqual, http.StatusOK)
			})
		})

		Convey("StopPing", func() {
			Convey("Success", func() {
				address := "192.168.1.1"

				req := httptest.NewRequest("POST", "/ping/stop?router_id="+routerID+"&address="+address, nil)
				w := httptest.NewRecorder()

				router.ServeHTTP(w, req)

				So(w.Code, ShouldEqual, http.StatusOK)
			})

			Convey("Missing router_id", func() {
				req := httptest.NewRequest("POST", "/ping/stop?address=192.168.1.1", nil)
				w := httptest.NewRecorder()

				router.ServeHTTP(w, req)

				So(w.Code, ShouldEqual, http.StatusBadRequest)
			})

			Convey("Missing address", func() {
				req := httptest.NewRequest("POST", "/ping/stop?router_id="+routerID, nil)
				w := httptest.NewRecorder()

				router.ServeHTTP(w, req)

				So(w.Code, ShouldEqual, http.StatusBadRequest)
			})

			Convey("Domain Error", func() {
				address := "192.168.1.1"

				req := httptest.NewRequest("POST", "/ping/stop?router_id="+routerID+"&address="+address, nil)
				w := httptest.NewRecorder()

				router.ServeHTTP(w, req)

				So(w.Code, ShouldEqual, http.StatusOK)
			})
		})

		Convey("HandleWebSocket", func() {
			Convey("Missing address", func() {
				req := httptest.NewRequest("GET", "/ping/ws", nil)
				w := httptest.NewRecorder()

				router.ServeHTTP(w, req)

				So(w.Code, ShouldEqual, http.StatusBadRequest)
			})

			Convey("Not Implemented", func() {
				req := httptest.NewRequest("GET", "/ping/ws?address=192.168.1.1", nil)
				w := httptest.NewRecorder()

				router.ServeHTTP(w, req)

				So(w.Code, ShouldEqual, http.StatusNotImplemented)
			})
		})
	})
}
