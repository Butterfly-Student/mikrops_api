package auth

import (
	"errors"
	"os"
	"testing"

	"go-template/internal/model"
	"go-template/internal/utils/hash"
	"go-template/internal/utils/token"
	mock_outbound_port "go-template/tests/mocks/port"

	"github.com/casbin/casbin/v2"
	casbinmodel "github.com/casbin/casbin/v2/model"
	"github.com/golang/mock/gomock"
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
		user := &model.User{ID: 1, Email: "test@example.com", Password: hashedPassword, Role: "user", Status: "active"}

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
		user := &model.User{ID: 1, Email: "test@example.com", Password: hashedPassword, Role: "user", Status: "active"}

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
		req := model.RegisterRequest{Name: "Test", Email: "new@example.com", Password: "password", Role: "user"}

		mockDB.EXPECT().User().Return(mockUserDB)
		mockUserDB.EXPECT().FindByEmail(req.Email).Return(nil, errors.New("not found"))

		mockDB.EXPECT().User().Return(mockUserDB)
		mockUserDB.EXPECT().Create(gomock.Any()).DoAndReturn(func(u *model.User) error {
			u.ID = 1
			return nil
		})

		err := domain.Register(req)
		assert.NoError(t, err)
	})

	t.Run("Register email exists", func(t *testing.T) {
		req := model.RegisterRequest{Name: "Test", Email: "existing@example.com", Password: "password"}

		mockDB.EXPECT().User().Return(mockUserDB)
		mockUserDB.EXPECT().FindByEmail(req.Email).Return(&model.User{}, nil)

		err := domain.Register(req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "email already exists")
	})

	t.Run("RefreshToken success", func(t *testing.T) {
		validToken, _ := token.GenerateRefreshToken(1)
		req := model.RefreshTokenRequest{RefreshToken: validToken}
		user := &model.User{ID: 1, Role: "user", Status: "active"}

		mockDB.EXPECT().User().Return(mockUserDB)
		mockUserDB.EXPECT().FindByID(uint(1)).Return(user, nil)

		res, err := domain.RefreshToken(req)
		assert.NoError(t, err)
		assert.NotEmpty(t, res.AccessToken)
		assert.NotEmpty(t, res.RefreshToken)
	})

	t.Run("ChangePassword success", func(t *testing.T) {
		oldHash, _ := hash.HashPassword("old_password")
		user := &model.User{ID: 1, Password: oldHash}
		req := model.ChangePasswordRequest{OldPassword: "old_password", NewPassword: "new_password"}

		mockDB.EXPECT().User().Return(mockUserDB)
		mockUserDB.EXPECT().FindByID(uint(1)).Return(user, nil)

		mockDB.EXPECT().User().Return(mockUserDB)
		mockUserDB.EXPECT().Update(gomock.Any()).Return(nil)

		err := domain.ChangePassword(1, req)
		assert.NoError(t, err)
	})
}
