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

func TestIpPoolHandler(t *testing.T) {
	Convey("Test IP Pool HTTP Handler", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		mockDatabasePort := mock_outbound_port.NewMockDatabasePort(mockCtrl)
		mockMikrotikPort := mock_outbound_port.NewMockMikrotikPort(mockCtrl)

		dom := domain.NewDomain(mockDatabasePort, nil, nil, nil, mockMikrotikPort, nil, nil, nil)
		handler := gin_inbound_adapter.NewIpPoolAdapter(dom)

		gin.SetMode(gin.TestMode)
		router := gin.New()
		router.POST("/ippool", handler.CreateIpPool)
		router.GET("/ippool/:id", handler.GetIpPool)
		router.GET("/ippool", handler.ListIpPools)
		router.PUT("/ippool/:id", handler.UpdateIpPool)
		router.DELETE("/ippool/:id", handler.DeleteIpPool)

		routerID := uuid.New().String()

		Convey("CreateIpPool", func() {
			Convey("Success", func() {
				poolName := "pool-dhcp"
				ranges := "192.168.1.100-192.168.1.200"

				reqBody := model.IpPool{
					Name:    poolName,
					Ranges:  ranges,
				}

				bodyBytes, _ := json.Marshal(reqBody)
				req := httptest.NewRequest("POST", "/ippool?router_id="+routerID, bytes.NewBuffer(bodyBytes))
				req.Header.Set("Content-Type", "application/json")
				w := httptest.NewRecorder()

				mockMikrotikPort.EXPECT().
					CreateIpPool(routerID, gomock.Any()).
					Return(nil).
					Times(1)

				router.ServeHTTP(w, req)

				So(w.Code, ShouldEqual, http.StatusCreated)
			})

			Convey("Missing router_id", func() {
				reqBody := model.IpPool{
					Name:   "pool-dhcp",
					Ranges: "192.168.1.100-192.168.1.200",
				}

				bodyBytes, _ := json.Marshal(reqBody)
				req := httptest.NewRequest("POST", "/ippool", bytes.NewBuffer(bodyBytes))
				req.Header.Set("Content-Type", "application/json")
				w := httptest.NewRecorder()

				router.ServeHTTP(w, req)

				So(w.Code, ShouldEqual, http.StatusBadRequest)
			})

			Convey("Invalid JSON", func() {
				req := httptest.NewRequest("POST", "/ippool?router_id="+routerID, bytes.NewBuffer([]byte("invalid json")))
				req.Header.Set("Content-Type", "application/json")
				w := httptest.NewRecorder()

				router.ServeHTTP(w, req)

				So(w.Code, ShouldEqual, http.StatusBadRequest)
			})
		})

		Convey("GetIpPool", func() {
			Convey("Success", func() {
				poolID := "*1"

				expectedPool := &model.IpPool{
					ID:     poolID,
					Name:   "pool-dhcp",
					Ranges: "192.168.1.100-192.168.1.200",
				}

				mockMikrotikPort.EXPECT().
					GetIpPool(routerID, poolID).
					Return(expectedPool, nil).
					Times(1)

				req := httptest.NewRequest("GET", "/ippool/"+poolID+"?router_id="+routerID, nil)
				w := httptest.NewRecorder()

				router.ServeHTTP(w, req)

				So(w.Code, ShouldEqual, http.StatusOK)

				var response model.IpPool
				json.Unmarshal(w.Body.Bytes(), &response)
				So(response.ID, ShouldEqual, poolID)
			})

			Convey("Not Found", func() {
				poolID := "*1"

				mockMikrotikPort.EXPECT().
					GetIpPool(routerID, poolID).
					Return(nil, errors.New("pool not found")).
					Times(1)

				req := httptest.NewRequest("GET", "/ippool/"+poolID+"?router_id="+routerID, nil)
				w := httptest.NewRecorder()

				router.ServeHTTP(w, req)

				So(w.Code, ShouldEqual, http.StatusNotFound)
			})

			Convey("Missing router_id", func() {
				poolID := "*1"

				req := httptest.NewRequest("GET", "/ippool/"+poolID, nil)
				w := httptest.NewRecorder()

				router.ServeHTTP(w, req)

				So(w.Code, ShouldEqual, http.StatusBadRequest)
			})
		})

		Convey("ListIpPools", func() {
			Convey("Success", func() {
				pools := []model.IpPool{
					{
						ID:     "*1",
						Name:   "pool-dhcp",
						Ranges: "192.168.1.100-192.168.1.200",
					},
					{
						ID:     "*2",
						Name:   "pool-static",
						Ranges: "192.168.2.100-192.168.2.150",
					},
				}

				mockMikrotikPort.EXPECT().
					ListIpPools(routerID).
					Return(pools, nil).
					Times(1)

				req := httptest.NewRequest("GET", "/ippool?router_id="+routerID, nil)
				w := httptest.NewRecorder()

				router.ServeHTTP(w, req)

				So(w.Code, ShouldEqual, http.StatusOK)

				var response []model.IpPool
				json.Unmarshal(w.Body.Bytes(), &response)
				So(len(response), ShouldEqual, 2)
			})

			Convey("Missing router_id", func() {
				req := httptest.NewRequest("GET", "/ippool", nil)
				w := httptest.NewRecorder()

				router.ServeHTTP(w, req)

				So(w.Code, ShouldEqual, http.StatusBadRequest)
			})
		})

		Convey("UpdateIpPool", func() {
			Convey("Success", func() {
				poolID := "*1"
				poolName := "pool-dhcp-updated"
				ranges := "192.168.1.100-192.168.1.250"

				reqBody := model.IpPool{
					Name:   poolName,
					Ranges: ranges,
				}

				bodyBytes, _ := json.Marshal(reqBody)
				req := httptest.NewRequest("PUT", "/ippool/"+poolID+"?router_id="+routerID, bytes.NewBuffer(bodyBytes))
				req.Header.Set("Content-Type", "application/json")
				w := httptest.NewRecorder()

				mockMikrotikPort.EXPECT().
					UpdateIpPool(routerID, gomock.Any()).
					Return(nil).
					Times(1)

				router.ServeHTTP(w, req)

				So(w.Code, ShouldEqual, http.StatusOK)
			})
		})

		Convey("DeleteIpPool", func() {
			Convey("Success", func() {
				poolID := "*1"

				mockMikrotikPort.EXPECT().
					DeleteIpPool(routerID, poolID).
					Return(nil).
					Times(1)

				req := httptest.NewRequest("DELETE", "/ippool/"+poolID+"?router_id="+routerID, nil)
				w := httptest.NewRecorder()

				router.ServeHTTP(w, req)

				So(w.Code, ShouldEqual, http.StatusOK)
			})

			Convey("Error", func() {
				poolID := "*1"

				mockMikrotikPort.EXPECT().
					DeleteIpPool(routerID, poolID).
					Return(errors.New("failed to delete pool")).
					Times(1)

				req := httptest.NewRequest("DELETE", "/ippool/"+poolID+"?router_id="+routerID, nil)
				w := httptest.NewRecorder()

				router.ServeHTTP(w, req)

				So(w.Code, ShouldEqual, http.StatusInternalServerError)
			})
		})
	})
}
