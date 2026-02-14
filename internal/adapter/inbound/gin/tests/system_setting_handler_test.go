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

func TestSystemSettingHandler(t *testing.T) {
	Convey("Test System Setting HTTP Handler", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		mockDatabasePort := mock_outbound_port.NewMockDatabasePort(mockCtrl)
		mockSystemSettingDB := mock_outbound_port.NewMockSystemSettingDatabasePort(mockCtrl)

		mockDatabasePort.EXPECT().SystemSetting().Return(mockSystemSettingDB).AnyTimes()

		dom := domain.NewDomain(mockDatabasePort, nil, nil, nil, nil, nil, nil, nil)
		handler := gin_inbound_adapter.NewSystemSettingHandler(dom)

		gin.SetMode(gin.TestMode)
		router := gin.New()
		router.GET("/settings", handler.ListSettings)
		router.GET("/settings/:key", handler.GetSetting)
		router.PUT("/settings/:key", handler.UpdateSetting)

		Convey("GetSetting", func() {
			Convey("Success", func() {
				key := "company_name"
				value := "PT. Example ISP"

				expectedSetting := &model.SystemSetting{
					ID:    uuid.New(),
					Key:   key,
					Value: &value,
				}

				mockSystemSettingDB.EXPECT().
					FindByKey(key).
					Return(expectedSetting, nil).
					Times(1)

				req := httptest.NewRequest("GET", "/settings/"+key, nil)
				w := httptest.NewRecorder()

				router.ServeHTTP(w, req)

				So(w.Code, ShouldEqual, http.StatusOK)

				var response model.SystemSetting
				json.Unmarshal(w.Body.Bytes(), &response)
				So(response.Key, ShouldEqual, key)
			})

			Convey("Not Found", func() {
				key := "nonexistent_key"

				mockSystemSettingDB.EXPECT().
					FindByKey(key).
					Return(nil, errors.New("setting not found")).
					Times(1)

				req := httptest.NewRequest("GET", "/settings/"+key, nil)
				w := httptest.NewRecorder()

				router.ServeHTTP(w, req)

				So(w.Code, ShouldEqual, http.StatusNotFound)
			})
		})

		Convey("ListSettings", func() {
			Convey("Success", func() {
				companyName := "PT. Example ISP"
				companyEmail := "info@example.com"

				settings := []model.SystemSetting{
					{
						ID:    uuid.New(),
						Key:   "company_name",
						Value: &companyName,
					},
					{
						ID:    uuid.New(),
						Key:   "company_email",
						Value: &companyEmail,
					},
				}

				mockSystemSettingDB.EXPECT().
					FindAll().
					Return(settings, nil).
					Times(1)

				req := httptest.NewRequest("GET", "/settings", nil)
				w := httptest.NewRecorder()

				router.ServeHTTP(w, req)

				So(w.Code, ShouldEqual, http.StatusOK)

				var response []model.SystemSetting
				json.Unmarshal(w.Body.Bytes(), &response)
				So(len(response), ShouldEqual, 2)
			})
		})

		Convey("UpdateSetting", func() {
			Convey("Success", func() {
				key := "company_name"
				value := "PT. Updated ISP"

				reqBody := map[string]interface{}{
					"value": value,
				}

				bodyBytes, _ := json.Marshal(reqBody)
				req := httptest.NewRequest("PUT", "/settings/"+key, bytes.NewBuffer(bodyBytes))
				req.Header.Set("Content-Type", "application/json")
				w := httptest.NewRecorder()

				mockSystemSettingDB.EXPECT().
					UpdateByKey(key, value, gomock.Any()).
					Return(nil).
					Times(1)

				router.ServeHTTP(w, req)

				So(w.Code, ShouldEqual, http.StatusOK)
			})
		})
	})
}
