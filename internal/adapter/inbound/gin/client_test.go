package gin_inbound_adapter_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	. "github.com/smartystreets/goconvey/convey"

	gin_inbound_adapter "go-template/internal/adapter/inbound/gin"
	"go-template/internal/domain"
	"go-template/internal/model"
	mock_outbound_port "go-template/tests/mocks/port"
)

func TestClientAdapter(t *testing.T) {
	Convey("Test Client HTTP Adapter", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		mockDatabasePort := mock_outbound_port.NewMockDatabasePort(mockCtrl)
		mockMessagePort := mock_outbound_port.NewMockMessagePort(mockCtrl)
		mockCachePort := mock_outbound_port.NewMockCachePort(mockCtrl)
		mockWorkflowPort := mock_outbound_port.NewMockWorkflowPort(mockCtrl)

		mockClientDatabasePort := mock_outbound_port.NewMockClientDatabasePort(mockCtrl)
		mockClientCachePort := mock_outbound_port.NewMockClientCachePort(mockCtrl)
		mockClientWorkflowPort := mock_outbound_port.NewMockClientWorkflowPort(mockCtrl)

		mockDatabasePort.EXPECT().Client().Return(mockClientDatabasePort).AnyTimes()
		mockMessagePort.EXPECT().Client().Return(mock_outbound_port.NewMockClientMessagePort(mockCtrl)).AnyTimes()
		mockCachePort.EXPECT().Client().Return(mockClientCachePort).AnyTimes()
		mockWorkflowPort.EXPECT().Client().Return(mockClientWorkflowPort).AnyTimes()

		dom := domain.NewDomain(mockDatabasePort, mockMessagePort, mockCachePort, mockWorkflowPort, nil)
		adapter := gin_inbound_adapter.NewAdapter(dom)

		// Set Gin to test mode
		gin.SetMode(gin.TestMode)
		router := gin.New()
		router.POST("/client-upsert", adapter.Client().Upsert)
		router.POST("/client-find", adapter.Client().Find)
		router.POST("/client-delete", adapter.Client().Delete)

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

				respBody, _ := io.ReadAll(w.Body)
				var result model.Response
				json.Unmarshal(respBody, &result)
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

				respBody, _ := io.ReadAll(w.Body)
				var result model.Response
				json.Unmarshal(respBody, &result)
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

				respBody, _ := io.ReadAll(w.Body)
				var result model.Response
				json.Unmarshal(respBody, &result)
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
