package auth

import (
	"errors"

	"go-template/internal/model"
	outbound_port "go-template/internal/port/outbound"
	"go-template/utils/hash"
	"go-template/utils/token"

	"github.com/casbin/casbin/v3"
	"github.com/google/uuid"
)

type AuthDomain interface {
	Login(req model.LoginRequest) (*model.LoginResponse, error)
	Register(req model.RegisterRequest) error
	RefreshToken(req model.RefreshTokenRequest) (*model.LoginResponse, error)
	ChangePassword(userID uuid.UUID, req model.ChangePasswordRequest) error
	Logout(userID uuid.UUID) error
	Enforce(sub, obj, act string) (bool, error)
}

type domain struct {
	dbPort   outbound_port.DatabasePort
	enforcer *casbin.Enforcer
}

func NewAuthDomain(dbPort outbound_port.DatabasePort, enforcer *casbin.Enforcer) AuthDomain {
	return &domain{
		dbPort:   dbPort,
		enforcer: enforcer,
	}
}

func (d *domain) Login(req model.LoginRequest) (*model.LoginResponse, error) {
	user, err := d.dbPort.User().FindByEmail(req.Email)
	if err != nil {
		return nil, errors.New("invalid credentials")
	}

	if !hash.CheckPasswordHash(req.Password, user.PasswordHash) {
		return nil, errors.New("invalid credentials")
	}

	if user.IsActive != nil && !*user.IsActive {
		return nil, errors.New("account inactive")
	}

	accessToken, err := token.GenerateAccessToken(user.ID.String(), string(user.Role))
	if err != nil {
		return nil, err
	}

	refreshToken, err := token.GenerateRefreshToken(user.ID.String())
	if err != nil {
		return nil, err
	}

	return &model.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (d *domain) Register(req model.RegisterRequest) error {
	if _, err := d.dbPort.User().FindByEmail(req.Email); err == nil {
		return errors.New("email already exists")
	}

	hashedPassword, err := hash.HashPassword(req.Password)
	if err != nil {
		return err
	}

	role := model.AdminRoleCS
	if req.Role != "" {
		role = model.AdminUserRole(req.Role)
	}

	isActive := true
	user := model.AdminUser{
		FullName:     req.FullName,
		Email:        req.Email,
		PasswordHash: hashedPassword,
		Role:         role,
		IsActive:     &isActive,
	}

	if err = d.dbPort.User().Create(&user); err != nil {
		return err
	}

	_, err = d.enforcer.AddGroupingPolicy(user.ID.String(), string(user.Role))
	return err
}

func (d *domain) RefreshToken(req model.RefreshTokenRequest) (*model.LoginResponse, error) {
	claims, err := token.ValidateToken(req.RefreshToken, true)
	if err != nil {
		return nil, err
	}

	subStr, ok := claims["sub"].(string)
	if !ok {
		return nil, errors.New("invalid token sub")
	}
	userID, err := uuid.Parse(subStr)
	if err != nil {
		return nil, errors.New("invalid token sub")
	}

	user, err := d.dbPort.User().FindByID(userID.String())
	if err != nil {
		return nil, errors.New("user not found")
	}

	if user.IsActive != nil && !*user.IsActive {
		return nil, errors.New("account inactive")
	}

	newAccessToken, err := token.GenerateAccessToken(user.ID.String(), string(user.Role))
	if err != nil {
		return nil, err
	}

	newRefreshToken, err := token.GenerateRefreshToken(user.ID.String())
	if err != nil {
		return nil, err
	}

	return &model.LoginResponse{
		AccessToken:  newAccessToken,
		RefreshToken: newRefreshToken,
	}, nil
}

func (d *domain) ChangePassword(userID uuid.UUID, req model.ChangePasswordRequest) error {
	user, err := d.dbPort.User().FindByID(userID.String())
	if err != nil {
		return err
	}

	if !hash.CheckPasswordHash(req.OldPassword, user.PasswordHash) {
		return errors.New("incorrect old password")
	}

	newHash, err := hash.HashPassword(req.NewPassword)
	if err != nil {
		return err
	}

	user.PasswordHash = newHash
	return d.dbPort.User().Update(*user)
}

func (d *domain) Logout(userID uuid.UUID) error {
	return nil
}

func (d *domain) Enforce(sub, obj, act string) (bool, error) {
	return d.enforcer.Enforce(sub, obj, act)
}
