package auth

import (
	"errors"
	"fmt"

	"go-template/internal/model"
	outbound_port "go-template/internal/port/outbound"
	"go-template/utils/hash"
	"go-template/utils/token"

	"github.com/casbin/casbin/v3"
)

type AuthDomain interface {
	Login(req model.LoginRequest) (*model.LoginResponse, error)
	Register(req model.RegisterRequest) error
	RefreshToken(req model.RefreshTokenRequest) (*model.LoginResponse, error)
	ChangePassword(userID uint, req model.ChangePasswordRequest) error
	Logout(userID uint) error
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

	if !hash.CheckPasswordHash(req.Password, user.Password) {
		return nil, errors.New("invalid credentials")
	}

	if user.Status != "active" {
		return nil, errors.New("account inactive")
	}

	accessToken, err := token.GenerateAccessToken(user.ID, user.Role)
	if err != nil {
		return nil, err
	}

	refreshToken, err := token.GenerateRefreshToken(user.ID)
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

	// Default role to "user" for public registration
	role := "user"

	user := model.User{
		Name:     req.Name,
		Email:    req.Email,
		Password: hashedPassword,
		Role:     role,
		Status:   "active",
	}

	if err = d.dbPort.User().Create(&user); err != nil {
		return err
	}

	// Assign role in Casbin
	// Sub: UserID (string), Role: role
	_, err = d.enforcer.AddGroupingPolicy(fmt.Sprintf("%d", user.ID), role)
	return err
}

func (d *domain) RefreshToken(req model.RefreshTokenRequest) (*model.LoginResponse, error) {
	claims, err := token.ValidateToken(req.RefreshToken, true)
	if err != nil {
		return nil, err
	}

	userIDFloat, ok := claims["sub"].(float64)
	if !ok {
		return nil, errors.New("invalid token sub")
	}
	userID := uint(userIDFloat)

	user, err := d.dbPort.User().FindByID(userID)
	if err != nil {
		return nil, errors.New("user not found")
	}

	if user.Status != "active" {
		return nil, errors.New("account inactive")
	}

	newAccessToken, err := token.GenerateAccessToken(user.ID, user.Role)
	if err != nil {
		return nil, err
	}

	newRefreshToken, err := token.GenerateRefreshToken(user.ID)
	if err != nil {
		return nil, err
	}

	return &model.LoginResponse{
		AccessToken:  newAccessToken,
		RefreshToken: newRefreshToken,
	}, nil
}

func (d *domain) ChangePassword(userID uint, req model.ChangePasswordRequest) error {
	user, err := d.dbPort.User().FindByID(userID)
	if err != nil {
		return err
	}

	if !hash.CheckPasswordHash(req.OldPassword, user.Password) {
		return errors.New("incorrect old password")
	}

	newHash, err := hash.HashPassword(req.NewPassword)
	if err != nil {
		return err
	}

	user.Password = newHash
	return d.dbPort.User().Update(*user)
}

func (d *domain) Logout(userID uint) error {
	return nil
}

func (d *domain) Enforce(sub, obj, act string) (bool, error) {
	return d.enforcer.Enforce(sub, obj, act)
}
