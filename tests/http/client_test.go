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

func TestClientAdapter(t *testing.T) {
	Convey("Test Client HTTP Adapter", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		mockDatabasePort := mock_outbound_port.NewMockDatabasePort(mockCtrl)
		mockMessagePort := mock_outbound_port.NewMockMessagePort(mockCtrl)
		mockCachePort := mock_outbound_port.NewMockCachePort(mockCtrl)
		mockWorkflowPort := mock_outbound_port.NewMockWorkflowPort(mockCtrl)
		mockHttpPort := mock_outbound_port.NewMockHttpPort(mockCtrl)

		mockClientDatabasePort := mock_outbound_port.NewMockClientDatabasePort(mockCtrl)
		mockClientCachePort := mock_outbound_port.NewMockClientCachePort(mockCtrl)
		mockClientWorkflowPort := mock_outbound_port.NewMockClientWorkflowPort(mockCtrl)

		mockDatabasePort.EXPECT().Client().Return(mockClientDatabasePort).AnyTimes()
		mockMessagePort.EXPECT().Client().Return(mock_outbound_port.NewMockClientMessagePort(mockCtrl)).AnyTimes()
		mockCachePort.EXPECT().Client().Return(mockClientCachePort).AnyTimes()
		mockWorkflowPort.EXPECT().Client().Return(mockClientWorkflowPort).AnyTimes()

		dom := domain.NewDomain(mockDatabasePort, mockMessagePort, mockCachePort, mockWorkflowPort, mockHttpPort)
		adapter := gin_inbound_adapter.NewAdapter(dom, mockHttpPort)

		router := gin.New()
		router.POST("/client-upsert", func(c *gin.Context) {
			adapter.Client().Upsert(c)
		})
		router.POST("/client-find", func(c *gin.Context) {
			adapter.Client().Find(c)
		})
		router.POST("/client-delete", func(c *gin.Context) {
			adapter.Client().Delete(c)
		})

		inputs := []model.ClientInput{
			{Name: "Test Client"},
		}

		outputs := []model.Client{
			{
				ID: 1,
				ClientInput: model.ClientInput{
					Name:      "Test Client",
					BearerKey: "test-bearer-key",
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				},
			},
		}

		filter := model.ClientFilter{
			IDs: []int{1},
		}

		Convey("Upsert", func() {
			Convey("Success", func() {
				mockClientDatabasePort.EXPECT().Upsert(gomock.Any()).Return(nil).Times(1)
				mockClientDatabasePort.EXPECT().FindByFilter(gomock.Any(), gomock.Any()).Return(outputs, nil).Times(1)

				body, _ := json.Marshal(inputs)
				req := httptest.NewRequest(http.MethodPost, "/client-upsert", bytes.NewReader(body))
				req.Header.Set("Content-Type", "application/json")
				w := httptest.NewRecorder()

				router.ServeHTTP(w, req)

				So(w.Code, ShouldEqual, http.StatusOK)

				var result model.Response
				json.Unmarshal(w.Body.Bytes(), &result)
				So(result.Success, ShouldBeTrue)
			})

			Convey("Invalid JSON", func() {
				req := httptest.NewRequest(http.MethodPost, "/client-upsert", bytes.NewReader([]byte("invalid json")))
				req.Header.Set("Content-Type", "application/json")
				w := httptest.NewRecorder()

				router.ServeHTTP(w, req)

				So(w.Code, ShouldEqual, http.StatusBadRequest)
			})

			Convey("Domain error", func() {
				mockClientDatabasePort.EXPECT().Upsert(gomock.Any()).Return(errors.New("database error")).Times(1)

				body, _ := json.Marshal(inputs)
				req := httptest.NewRequest(http.MethodPost, "/client-upsert", bytes.NewReader(body))
				req.Header.Set("Content-Type", "application/json")
				w := httptest.NewRecorder()

				router.ServeHTTP(w, req)

				So(w.Code, ShouldEqual, http.StatusInternalServerError)
			})
		})

		Convey("Find", func() {
			Convey("Success", func() {
				mockClientDatabasePort.EXPECT().FindByFilter(gomock.Any(), gomock.Any()).Return(outputs, nil).Times(1)

				body, _ := json.Marshal(filter)
				req := httptest.NewRequest(http.MethodPost, "/client-find", bytes.NewReader(body))
				req.Header.Set("Content-Type", "application/json")
				w := httptest.NewRecorder()

				router.ServeHTTP(w, req)

				So(w.Code, ShouldEqual, http.StatusOK)

				var result model.Response
				json.Unmarshal(w.Body.Bytes(), &result)
				So(result.Success, ShouldBeTrue)
			})

			Convey("Invalid JSON", func() {
				req := httptest.NewRequest(http.MethodPost, "/client-find", bytes.NewReader([]byte("invalid")))
				req.Header.Set("Content-Type", "application/json")
				w := httptest.NewRecorder()

				router.ServeHTTP(w, req)

				So(w.Code, ShouldEqual, http.StatusBadRequest)
			})

			Convey("Domain error", func() {
				mockClientDatabasePort.EXPECT().FindByFilter(gomock.Any(), gomock.Any()).Return(nil, errors.New("error")).Times(1)

				body, _ := json.Marshal(filter)
				req := httptest.NewRequest(http.MethodPost, "/client-find", bytes.NewReader(body))
				req.Header.Set("Content-Type", "application/json")
				w := httptest.NewRecorder()

				router.ServeHTTP(w, req)

				So(w.Code, ShouldEqual, http.StatusInternalServerError)
			})
		})

		Convey("Delete", func() {
			Convey("Success", func() {
				mockClientDatabasePort.EXPECT().DeleteByFilter(gomock.Any()).Return(nil).Times(1)

				body, _ := json.Marshal(filter)
				req := httptest.NewRequest(http.MethodPost, "/client-delete", bytes.NewReader(body))
				req.Header.Set("Content-Type", "application/json")
				w := httptest.NewRecorder()

				router.ServeHTTP(w, req)

				So(w.Code, ShouldEqual, http.StatusOK)

				var result model.Response
				json.Unmarshal(w.Body.Bytes(), &result)
				So(result.Success, ShouldBeTrue)
			})

			Convey("Invalid JSON", func() {
				req := httptest.NewRequest(http.MethodPost, "/client-delete", bytes.NewReader([]byte("invalid")))
				req.Header.Set("Content-Type", "application/json")
				w := httptest.NewRecorder()

				router.ServeHTTP(w, req)

				So(w.Code, ShouldEqual, http.StatusBadRequest)
			})

			Convey("Domain error", func() {
				mockClientDatabasePort.EXPECT().DeleteByFilter(gomock.Any()).Return(errors.New("error")).Times(1)

				body, _ := json.Marshal(filter)
				req := httptest.NewRequest(http.MethodPost, "/client-delete", bytes.NewReader(body))
				req.Header.Set("Content-Type", "application/json")
				w := httptest.NewRecorder()

				router.ServeHTTP(w, req)

				So(w.Code, ShouldEqual, http.StatusInternalServerError)
			})
		})
	})
}
