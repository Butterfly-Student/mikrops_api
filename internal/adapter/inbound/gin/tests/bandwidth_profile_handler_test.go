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

func TestBandwidthProfileHandler(t *testing.T) {
	Convey("Test Bandwidth Profile HTTP Handler", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		mockDatabasePort := mock_outbound_port.NewMockDatabasePort(mockCtrl)
		mockMikrotikPort := mock_outbound_port.NewMockMikrotikPort(mockCtrl)
		mockBandwidthProfileDB := mock_outbound_port.NewMockBandwidthProfileDatabasePort(mockCtrl)

		mockDatabasePort.EXPECT().BandwidthProfile().Return(mockBandwidthProfileDB).AnyTimes()

		dom := domain.NewDomain(mockDatabasePort, nil, nil, nil, mockMikrotikPort, nil, nil, nil)
		handler := gin_inbound_adapter.NewBandwidthProfileHandler(dom)

		gin.SetMode(gin.TestMode)
		router := gin.New()
		router.POST("/profiles", handler.CreateProfile)
		router.GET("/profiles/:id", handler.GetProfile)
		router.GET("/profiles", handler.ListProfiles)
		router.PUT("/profiles/:id", handler.UpdateProfile)
		router.DELETE("/profiles/:id", handler.DeleteProfile)

		Convey("CreateProfile", func() {
			Convey("Success", func() {
				profileCode := "BW-100MBPS"
				name := "100 Mbps Package"
				downloadSpeed := int64(100000000)
				uploadSpeed := int64(100000000)

				reqBody := map[string]interface{}{
					"profile_code":   profileCode,
					"name":           name,
					"download_speed": downloadSpeed,
					"upload_speed":   uploadSpeed,
					"category":       "pppoe",
				}

				bodyBytes, _ := json.Marshal(reqBody)
				req := httptest.NewRequest("POST", "/profiles", bytes.NewBuffer(bodyBytes))
				req.Header.Set("Content-Type", "application/json")
				w := httptest.NewRecorder()

				expectedProfile := &model.BandwidthProfile{
					ID:            uuid.New(),
					ProfileCode:   profileCode,
					Name:          name,
					DownloadSpeed: downloadSpeed,
					UploadSpeed:   uploadSpeed,
				}

				mockDatabasePort.EXPECT().
					DoInTransaction(gomock.Any()).
					DoAndReturn(func(txFunc interface{}) (interface{}, error) {
						return expectedProfile, nil
					}).Times(1)

				router.ServeHTTP(w, req)

				So(w.Code, ShouldEqual, http.StatusCreated)
			})

			Convey("Invalid JSON", func() {
				req := httptest.NewRequest("POST", "/profiles", bytes.NewBuffer([]byte("invalid")))
				req.Header.Set("Content-Type", "application/json")
				w := httptest.NewRecorder()

				router.ServeHTTP(w, req)

				So(w.Code, ShouldEqual, http.StatusBadRequest)
			})
		})

		Convey("GetProfile", func() {
			Convey("Success", func() {
				profileID := uuid.New()

				expectedProfile := &model.BandwidthProfile{
					ID:            profileID,
					ProfileCode:   "BW-100MBPS",
					Name:          "100 Mbps Package",
					DownloadSpeed: 100000000,
					UploadSpeed:   100000000,
				}

				mockBandwidthProfileDB.EXPECT().
					FindByID(profileID.String()).
					Return(expectedProfile, nil).
					Times(1)

				req := httptest.NewRequest("GET", "/profiles/"+profileID.String(), nil)
				w := httptest.NewRecorder()

				router.ServeHTTP(w, req)

				So(w.Code, ShouldEqual, http.StatusOK)
			})

			Convey("Not Found", func() {
				profileID := uuid.New()

				mockBandwidthProfileDB.EXPECT().
					FindByID(profileID.String()).
					Return(nil, errors.New("profile not found")).
					Times(1)

				req := httptest.NewRequest("GET", "/profiles/"+profileID.String(), nil)
				w := httptest.NewRecorder()

				router.ServeHTTP(w, req)

				So(w.Code, ShouldEqual, http.StatusNotFound)
			})
		})

		Convey("ListProfiles", func() {
			Convey("Success", func() {
				profiles := []model.BandwidthProfile{
					{
						ID:            uuid.New(),
						ProfileCode:   "BW-100MBPS",
						Name:          "100 Mbps Package",
						DownloadSpeed: 100000000,
						UploadSpeed:   100000000,
					},
					{
						ID:            uuid.New(),
						ProfileCode:   "BW-50MBPS",
						Name:          "50 Mbps Package",
						DownloadSpeed: 50000000,
						UploadSpeed:   50000000,
					},
				}

				mockBandwidthProfileDB.EXPECT().
					FindAll().
					Return(profiles, nil).
					Times(1)

				req := httptest.NewRequest("GET", "/profiles", nil)
				w := httptest.NewRecorder()

				router.ServeHTTP(w, req)

				So(w.Code, ShouldEqual, http.StatusOK)

				var response []model.BandwidthProfile
				json.Unmarshal(w.Body.Bytes(), &response)
				So(len(response), ShouldEqual, 2)
			})
		})
	})
}
