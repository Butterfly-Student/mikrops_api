package auth

import (
	"errors"
	"os"
	"testing"

	"go-template/internal/model"
	mock_outbound_port "go-template/tests/mocks/port"
	"go-template/utils/hash"
	"go-template/utils/token"

	"github.com/casbin/casbin/v3"
	casbinmodel "github.com/casbin/casbin/v3/model"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestAuthDomain(t *testing.T) {
	os.Setenv("JWT_SECRET", "secret")
	os.Setenv("JWT_REFRESH_SECRET", "refresh_secret")
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mock_outbound_port.NewMockDatabasePort(ctrl)
	mockUserDB := mock_outbound_port.NewMockUserDatabasePort(ctrl)

	// Create memory enforcer
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
	domain := NewAuthDomain(mockDB, enforcer)

	t.Run("Login success", func(t *testing.T) {
		req := model.LoginRequest{Email: "test@example.com", Password: "password"}
		hashedPassword, _ := hash.HashPassword("password")
		testUUID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440001")
		isActive := true
		user := &model.AdminUser{
			ID:           testUUID,
			Email:        "test@example.com",
			PasswordHash: hashedPassword,
			Role:         model.AdminRoleCS,
			IsActive:     &isActive,
		}

		mockDB.EXPECT().User().Return(mockUserDB)
		mockUserDB.EXPECT().FindByEmail(req.Email).Return(user, nil)

		res, err := domain.Login(req)
		assert.NoError(t, err)
		assert.NotEmpty(t, res.AccessToken)
		assert.NotEmpty(t, res.RefreshToken)
	})

	t.Run("Login invalid credentials", func(t *testing.T) {
		req := model.LoginRequest{Email: "test@example.com", Password: "wrong_password"}
		hashedPassword, _ := hash.HashPassword("password")
		testUUID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440001")
		isActive := true
		user := &model.AdminUser{
			ID:           testUUID,
			Email:        "test@example.com",
			PasswordHash: hashedPassword,
			Role:         model.AdminRoleCS,
			IsActive:     &isActive,
		}

		mockDB.EXPECT().User().Return(mockUserDB)
		mockUserDB.EXPECT().FindByEmail(req.Email).Return(user, nil)

		res, err := domain.Login(req)
		assert.Error(t, err)
		assert.Nil(t, res)
	})

	t.Run("Login user not found", func(t *testing.T) {
		req := model.LoginRequest{Email: "test@example.com", Password: "password"}
		mockDB.EXPECT().User().Return(mockUserDB)
		mockUserDB.EXPECT().FindByEmail(req.Email).Return(nil, errors.New("not found"))

		res, err := domain.Login(req)
		assert.Error(t, err)
		assert.Nil(t, res)
	})

	t.Run("Register success", func(t *testing.T) {
		req := model.RegisterRequest{FullName: "Test", Email: "new@example.com", Password: "password", Role: "cs"}

		mockDB.EXPECT().User().Return(mockUserDB)
		mockUserDB.EXPECT().FindByEmail(req.Email).Return(nil, errors.New("not found"))

		mockDB.EXPECT().User().Return(mockUserDB)
		mockUserDB.EXPECT().Create(gomock.Any()).DoAndReturn(func(u *model.AdminUser) error {
			u.ID = uuid.MustParse("550e8400-e29b-41d4-a716-446655440001")
			return nil
		})

		err := domain.Register(req)
		assert.NoError(t, err)
	})

	t.Run("Register email exists", func(t *testing.T) {
		req := model.RegisterRequest{FullName: "Test", Email: "existing@example.com", Password: "password"}

		mockDB.EXPECT().User().Return(mockUserDB)
		mockUserDB.EXPECT().FindByEmail(req.Email).Return(&model.AdminUser{}, nil)

		err := domain.Register(req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "email already exists")
	})

	t.Run("RefreshToken success", func(t *testing.T) {
		testUUID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440001")
		validToken, _ := token.GenerateRefreshToken(testUUID.String())
		req := model.RefreshTokenRequest{RefreshToken: validToken}
		isActive := true
		user := &model.AdminUser{
			ID:       testUUID,
			Role:     model.AdminRoleCS,
			IsActive: &isActive,
		}

		mockDB.EXPECT().User().Return(mockUserDB)
		mockUserDB.EXPECT().FindByID(testUUID.String()).Return(user, nil)

		res, err := domain.RefreshToken(req)
		assert.NoError(t, err)
		assert.NotEmpty(t, res.AccessToken)
		assert.NotEmpty(t, res.RefreshToken)
	})

	t.Run("ChangePassword success", func(t *testing.T) {
		oldHash, _ := hash.HashPassword("old_password")
		testUUID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440001")
		user := &model.AdminUser{
			ID:           testUUID,
			PasswordHash: oldHash,
		}
		req := model.ChangePasswordRequest{OldPassword: "old_password", NewPassword: "new_password"}

		mockDB.EXPECT().User().Return(mockUserDB)
		mockUserDB.EXPECT().FindByID(testUUID.String()).Return(user, nil)

		mockDB.EXPECT().User().Return(mockUserDB)
		mockUserDB.EXPECT().Update(gomock.Any()).Return(nil)

		err := domain.ChangePassword(testUUID, req)
		assert.NoError(t, err)
	})
}
