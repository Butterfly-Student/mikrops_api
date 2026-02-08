package gin_inbound_adapter

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/palantir/stacktrace"

	"mikrops/internal/domain"
	"mikrops/internal/model"
	inbound_port "mikrops/internal/port/inbound"
	"mikrops/utils/activity"
)

type authAdapter struct {
	domain domain.Domain
}

func NewAuthAdapter(domain domain.Domain) inbound_port.AuthHttpPort {
	return &authAdapter{domain: domain}
}

func (h *authAdapter) StaffLogin(a any) error {
	c := a.(*gin.Context)
	ctx := activity.NewContext("http_staff_login")

	var payload model.StaffAuthRequest
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, model.Response{Success: false, Error: err.Error()})
		return nil
	}

	result, err := h.domain.Auth().StaffLogin(ctx, payload)
	if err != nil {
		c.JSON(http.StatusUnauthorized, model.Response{Success: false, Error: stacktrace.RootCause(err).Error()})
		return nil
	}

	c.JSON(http.StatusOK, model.Response{Success: true, Data: result})
	return nil
}

func (h *authAdapter) StaffRefresh(a any) error {
	c := a.(*gin.Context)
	ctx := activity.NewContext("http_staff_refresh")

	var payload model.RefreshTokenRequest
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, model.Response{Success: false, Error: err.Error()})
		return nil
	}

	result, err := h.domain.Auth().StaffRefreshToken(ctx, payload.RefreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, model.Response{Success: false, Error: stacktrace.RootCause(err).Error()})
		return nil
	}

	c.JSON(http.StatusOK, model.Response{Success: true, Data: result})
	return nil
}

func (h *authAdapter) CustomerLogin(a any) error {
	c := a.(*gin.Context)
	ctx := activity.NewContext("http_customer_login")

	var payload model.CustomerAuthRequest
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, model.Response{Success: false, Error: err.Error()})
		return nil
	}

	result, err := h.domain.Auth().CustomerLogin(ctx, payload)
	if err != nil {
		c.JSON(http.StatusUnauthorized, model.Response{Success: false, Error: stacktrace.RootCause(err).Error()})
		return nil
	}

	c.JSON(http.StatusOK, model.Response{Success: true, Data: result})
	return nil
}

func (h *authAdapter) CustomerRefresh(a any) error {
	c := a.(*gin.Context)
	ctx := activity.NewContext("http_customer_refresh")

	var payload model.RefreshTokenRequest
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, model.Response{Success: false, Error: err.Error()})
		return nil
	}

	result, err := h.domain.Auth().CustomerRefreshToken(ctx, payload.RefreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, model.Response{Success: false, Error: stacktrace.RootCause(err).Error()})
		return nil
	}

	c.JSON(http.StatusOK, model.Response{Success: true, Data: result})
	return nil
}
