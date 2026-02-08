package gin_inbound_adapter_test

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/redis/go-redis/v9"
	. "github.com/smartystreets/goconvey/convey"

	gin_inbound_adapter "mikrops/internal/adapter/inbound/gin"
	"mikrops/internal/domain"
	"mikrops/internal/model"
	mock_outbound_port "mikrops/tests/mocks/port"
)

func TestMiddlewareAdapter(t *testing.T) {
	Convey("Test Middleware Adapter", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		mockDatabasePort := mock_outbound_port.NewMockDatabasePort(mockCtrl)
		mockMessagePort := mock_outbound_port.NewMockMessagePort(mockCtrl)
		mockCachePort := mock_outbound_port.NewMockCachePort(mockCtrl)
		mockWorkflowPort := mock_outbound_port.NewMockWorkflowPort(mockCtrl)
		mockHttpPort := mock_outbound_port.NewMockHttpPort(mockCtrl)

		mockClientDatabasePort := mock_outbound_port.NewMockClientDatabasePort(mockCtrl)
		mockClientMessagePort := mock_outbound_port.NewMockClientMessagePort(mockCtrl)
		mockClientCachePort := mock_outbound_port.NewMockClientCachePort(mockCtrl)
		mockClientWorkflowPort := mock_outbound_port.NewMockClientWorkflowPort(mockCtrl)

		mockDatabasePort.EXPECT().Client().Return(mockClientDatabasePort).AnyTimes()
		mockCachePort.EXPECT().Client().Return(mockClientCachePort).AnyTimes()
		mockMessagePort.EXPECT().Client().Return(mockClientMessagePort).AnyTimes()
		mockWorkflowPort.EXPECT().Client().Return(mockClientWorkflowPort).AnyTimes()

		dom := domain.NewDomain(mockDatabasePort, mockMessagePort, mockCachePort, mockWorkflowPort, mockHttpPort)
		adapter := gin_inbound_adapter.NewAdapter(dom, mockHttpPort)

		Convey("InternalAuth", func() {
			router := gin.New()
			router.Use(func(c *gin.Context) {
				adapter.Middleware().InternalAuth(c)
			})
			router.GET("/test", func(c *gin.Context) {
				c.String(http.StatusOK, "OK")
			})

			Convey("Missing Authorization header", func() {
				req := httptest.NewRequest(http.MethodGet, "/test", nil)
				w := httptest.NewRecorder()

				router.ServeHTTP(w, req)

				So(w.Code, ShouldEqual, http.StatusUnauthorized)
			})

			Convey("Empty bearer token", func() {
				req := httptest.NewRequest(http.MethodGet, "/test", nil)
				req.Header.Set("Authorization", "Bearer ")
				w := httptest.NewRecorder()

				router.ServeHTTP(w, req)

				So(w.Code, ShouldEqual, http.StatusUnauthorized)
			})

			Convey("Invalid bearer token", func() {
				os.Setenv("INTERNAL_KEY", "valid-key")
				defer os.Unsetenv("INTERNAL_KEY")

				req := httptest.NewRequest(http.MethodGet, "/test", nil)
				req.Header.Set("Authorization", "Bearer invalid-key")
				w := httptest.NewRecorder()

				router.ServeHTTP(w, req)

				So(w.Code, ShouldEqual, http.StatusUnauthorized)
			})

			Convey("Valid bearer token", func() {
				os.Setenv("INTERNAL_KEY", "valid-key")
				defer os.Unsetenv("INTERNAL_KEY")

				req := httptest.NewRequest(http.MethodGet, "/test", nil)
				req.Header.Set("Authorization", "Bearer valid-key")
				w := httptest.NewRecorder()

				router.ServeHTTP(w, req)

				So(w.Code, ShouldEqual, http.StatusOK)
			})

			Convey("Malformed authorization header", func() {
				req := httptest.NewRequest(http.MethodGet, "/test", nil)
				req.Header.Set("Authorization", "Basic abc123")
				w := httptest.NewRecorder()

				router.ServeHTTP(w, req)

				So(w.Code, ShouldEqual, http.StatusUnauthorized)
			})
		})

		Convey("ClientAuth", func() {
			router := gin.New()
			router.Use(func(c *gin.Context) {
				adapter.Middleware().ClientAuth(c)
			})
			router.GET("/test", func(c *gin.Context) {
				c.String(http.StatusOK, "OK")
			})

			clientOutput := model.Client{
				ID: 1,
				ClientInput: model.ClientInput{
					Name:      "Test Client",
					BearerKey: "valid-client-key",
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				},
			}

			Convey("Missing Authorization header", func() {
				req := httptest.NewRequest(http.MethodGet, "/test", nil)
				w := httptest.NewRecorder()

				router.ServeHTTP(w, req)

				So(w.Code, ShouldEqual, http.StatusUnauthorized)
			})

			Convey("Client exists in cache", func() {
				mockClientCachePort.EXPECT().Get(gomock.Any()).Return(clientOutput, nil).Times(1)

				req := httptest.NewRequest(http.MethodGet, "/test", nil)
				req.Header.Set("Authorization", "Bearer valid-client-key")
				w := httptest.NewRecorder()

				router.ServeHTTP(w, req)

				So(w.Code, ShouldEqual, http.StatusOK)
			})

			Convey("Client exists in database (cache miss)", func() {
				mockClientCachePort.EXPECT().Get(gomock.Any()).Return(model.Client{}, redis.Nil).Times(1)
				mockClientDatabasePort.EXPECT().IsExists(gomock.Any()).Return(true, nil).Times(1)
				mockClientDatabasePort.EXPECT().FindByFilter(gomock.Any(), gomock.Any()).Return([]model.Client{clientOutput}, nil).Times(1)
				mockClientCachePort.EXPECT().Set(gomock.Any()).Return(nil).Times(1)

				req := httptest.NewRequest(http.MethodGet, "/test", nil)
				req.Header.Set("Authorization", "Bearer valid-client-key")
				w := httptest.NewRecorder()

				router.ServeHTTP(w, req)

				So(w.Code, ShouldEqual, http.StatusOK)
			})

			Convey("Client does not exist", func() {
				mockClientCachePort.EXPECT().Get(gomock.Any()).Return(model.Client{}, redis.Nil).Times(1)
				mockClientDatabasePort.EXPECT().IsExists(gomock.Any()).Return(false, nil).Times(1)

				req := httptest.NewRequest(http.MethodGet, "/test", nil)
				req.Header.Set("Authorization", "Bearer nonexistent-key")
				w := httptest.NewRecorder()

				router.ServeHTTP(w, req)

				So(w.Code, ShouldEqual, http.StatusUnauthorized)
			})
		})
	})
}
