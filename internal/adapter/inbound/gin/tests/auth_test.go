package gin_adapter_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
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
	"go-template/utils/hash"
	"go-template/utils/token"
)

func TestAuthAdapter(t *testing.T) {
	os.Setenv("JWT_SECRET", "secret")
	os.Setenv("JWT_REFRESH_SECRET", "refresh_secret")

	Convey("Test Auth HTTP Adapter", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		mockDatabasePort := mock_outbound_port.NewMockDatabasePort(mockCtrl)
		mockUserDatabasePort := mock_outbound_port.NewMockUserDatabasePort(mockCtrl)
		mockMessagePort := mock_outbound_port.NewMockMessagePort(mockCtrl)
		mockCachePort := mock_outbound_port.NewMockCachePort(mockCtrl)
		mockWorkflowPort := mock_outbound_port.NewMockWorkflowPort(mockCtrl)

		mockDatabasePort.EXPECT().User().Return(mockUserDatabasePort).AnyTimes()

		// Setup Casbin
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
		router.POST("/auth/login", adapter.Auth().Login)
		router.POST("/auth/register", adapter.Auth().Register)
		router.POST("/auth/refresh-token", adapter.Auth().RefreshToken)

		testUUID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440001")
		authMiddleware := func(c *gin.Context) {
			c.Set("userID", testUUID.String())
			c.Next()
		}
		router.POST("/auth/change-password", authMiddleware, adapter.Auth().ChangePassword)
		router.POST("/auth/logout", authMiddleware, adapter.Auth().Logout)

		Convey("Login", func() {
			reqBody := model.LoginRequest{Email: "test@example.com", Password: "password"}
			hashedPassword, _ := hash.HashPassword("password")
			isActive := true
			user := &model.User{
				ID:           testUUID,
				Email:        "test@example.com",
				PasswordHash: hashedPassword,
				Role:         model.AdminRoleAdmin,
				IsActive:     &isActive,
			}

			Convey("Success", func() {
				mockUserDatabasePort.EXPECT().FindByEmail(reqBody.Email).Return(user, nil).Times(1)

				body, _ := json.Marshal(reqBody)
				req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewReader(body))
				req.Header.Set("Content-Type", "application/json")

				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				So(w.Code, ShouldEqual, http.StatusOK)

				var res model.LoginResponse
				json.Unmarshal(w.Body.Bytes(), &res)
				So(res.AccessToken, ShouldNotBeEmpty)
			})

			Convey("Invalid Credentials", func() {
				mockUserDatabasePort.EXPECT().FindByEmail(reqBody.Email).Return(user, nil).Times(1)

				invalidReq := reqBody
				invalidReq.Password = "wrong"
				body, _ := json.Marshal(invalidReq)
				req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewReader(body))
				req.Header.Set("Content-Type", "application/json")

				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				So(w.Code, ShouldEqual, http.StatusUnauthorized)
			})
		})

		Convey("Register", func() {
			reqBody := model.RegisterRequest{FullName: "New User", Email: "new@example.com", Password: "password"}

			Convey("Success", func() {
				mockUserDatabasePort.EXPECT().FindByEmail(reqBody.Email).Return(nil, errors.New("not found")).Times(1)
				mockUserDatabasePort.EXPECT().Create(gomock.Any()).Return(nil).Times(1)

				body, _ := json.Marshal(reqBody)
				req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewReader(body))
				req.Header.Set("Content-Type", "application/json")

				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				So(w.Code, ShouldEqual, http.StatusCreated)
			})
		})

		Convey("RefreshToken", func() {
			testUUID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440001")
			validToken, _ := token.GenerateRefreshToken(testUUID.String())
			reqBody := model.RefreshTokenRequest{RefreshToken: validToken}
			isActive := true
			user := &model.User{ID: testUUID, Role: model.AdminRoleAdmin, IsActive: &isActive}

			Convey("Success", func() {
				mockUserDatabasePort.EXPECT().FindByID(testUUID.String()).Return(user, nil).Times(1)

				body, _ := json.Marshal(reqBody)
				req := httptest.NewRequest(http.MethodPost, "/auth/refresh-token", bytes.NewReader(body))
				req.Header.Set("Content-Type", "application/json")

				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				So(w.Code, ShouldEqual, http.StatusOK)
			})
		})

		Convey("ChangePassword", func() {
			oldHash, _ := hash.HashPassword("old_password")
			testUUID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440001")
			user := &model.User{ID: testUUID, PasswordHash: oldHash}
			reqBody := model.ChangePasswordRequest{OldPassword: "old_password", NewPassword: "new_password"}

			Convey("Success", func() {
				mockUserDatabasePort.EXPECT().FindByID(testUUID.String()).Return(user, nil).Times(1)
				mockUserDatabasePort.EXPECT().Update(gomock.Any()).Return(nil).Times(1)

				body, _ := json.Marshal(reqBody)
				req := httptest.NewRequest(http.MethodPost, "/auth/change-password", bytes.NewReader(body))
				req.Header.Set("Content-Type", "application/json")

				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				So(w.Code, ShouldEqual, http.StatusOK)
			})
		})
	})
}
