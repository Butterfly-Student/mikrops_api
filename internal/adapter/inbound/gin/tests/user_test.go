package gin_adapter_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/casbin/casbin/v3"
	casbinmodel "github.com/casbin/casbin/v3/model"
	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	. "github.com/smartystreets/goconvey/convey"

	gin_inbound_adapter "go-template/internal/adapter/inbound/gin"
	"go-template/internal/domain"
	"go-template/internal/model"
	mock_outbound_port "go-template/tests/mocks/port"
)

func TestUserAdapter(t *testing.T) {
	Convey("Test User HTTP Adapter", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		mockDatabasePort := mock_outbound_port.NewMockDatabasePort(mockCtrl)
		mockUserDatabasePort := mock_outbound_port.NewMockUserDatabasePort(mockCtrl)
		mockMessagePort := mock_outbound_port.NewMockMessagePort(mockCtrl)
		mockCachePort := mock_outbound_port.NewMockCachePort(mockCtrl)
		mockWorkflowPort := mock_outbound_port.NewMockWorkflowPort(mockCtrl)

		mockDatabasePort.EXPECT().User().Return(mockUserDatabasePort).AnyTimes()

		// Setup Casbin (needed for domain creation)
		m, _ := casbinmodel.NewModelFromString(`
[request_definition]
r = sub, obj, act

[policy_definition]
p = sub, obj, act

[role_definition]
g = _, _

[policy_effect]
e = some(where (p.eft == allow))

[matchers]
m = g(r.sub, p.sub) && r.obj == p.obj && r.act == p.act
`)
		enforcer, _ := casbin.NewEnforcer(m)

		dom := domain.NewDomain(mockDatabasePort, mockMessagePort, mockCachePort, mockWorkflowPort, nil, nil, enforcer)
		adapter := gin_inbound_adapter.NewAdapter(dom)

		gin.SetMode(gin.TestMode)
		router := gin.New()

		testUUID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440001")
		// Mock auth middleware
		authMiddleware := func(c *gin.Context) {
			c.Set("userID", testUUID.String())
			c.Next()
		}

		router.GET("/user/profile", authMiddleware, adapter.User().GetProfile)
		router.PUT("/user/profile", authMiddleware, adapter.User().UpdateProfile)

		Convey("GetProfile", func() {
			isActive := true
			user := &model.User{ID: testUUID, FullName: "Test User", Email: "test@example.com", Role: model.AdminRoleAdmin, IsActive: &isActive}

			Convey("Success", func() {
				mockUserDatabasePort.EXPECT().FindByID(testUUID.String()).Return(user, nil).Times(1)

				req := httptest.NewRequest(http.MethodGet, "/user/profile", nil)
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				So(w.Code, ShouldEqual, http.StatusOK)

				var res model.User
				json.Unmarshal(w.Body.Bytes(), &res)
				So(res.FullName, ShouldEqual, user.FullName)
			})

			Convey("User Not Found", func() {
				mockUserDatabasePort.EXPECT().FindByID(testUUID.String()).Return(nil, errors.New("not found")).Times(1)

				req := httptest.NewRequest(http.MethodGet, "/user/profile", nil)
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				So(w.Code, ShouldEqual, http.StatusNotFound)
			})
		})

		Convey("UpdateProfile", func() {
			isActive := true
			user := &model.User{ID: testUUID, FullName: "Old Name", Email: "old@example.com", IsActive: &isActive}
			reqBody := model.UserInput{FullName: "New Name", Email: "new@example.com", Password: "password123", Role: string(model.AdminRoleAdmin)}

			Convey("Success", func() {
				mockUserDatabasePort.EXPECT().FindByID(testUUID.String()).Return(user, nil).Times(1)
				mockUserDatabasePort.EXPECT().FindByEmail(reqBody.Email).Return(nil, errors.New("not found")).Times(1) // Check email uniqueness
				mockUserDatabasePort.EXPECT().Update(gomock.Any()).Return(nil).Times(1)

				body, _ := json.Marshal(reqBody)
				req := httptest.NewRequest(http.MethodPut, "/user/profile", bytes.NewReader(body))
				req.Header.Set("Content-Type", "application/json")

				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				So(w.Code, ShouldEqual, http.StatusOK)
			})

			Convey("Email Taken", func() {
				otherUUID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440002")
				existingUser := &model.User{ID: otherUUID, Email: "new@example.com", IsActive: &isActive}
				mockUserDatabasePort.EXPECT().FindByID(testUUID.String()).Return(user, nil).Times(1)
				mockUserDatabasePort.EXPECT().FindByEmail(reqBody.Email).Return(existingUser, nil).Times(1)

				body, _ := json.Marshal(reqBody)
				req := httptest.NewRequest(http.MethodPut, "/user/profile", bytes.NewReader(body))
				req.Header.Set("Content-Type", "application/json")

				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				So(w.Code, ShouldEqual, http.StatusInternalServerError)
			})
		})
	})
}
