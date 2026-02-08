package gin_inbound_adapter

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"

	"mikrops/internal/domain"
	"mikrops/internal/model"
	"mikrops/utils/activity"
	"mikrops/utils/jwt"
)

const (
	authorizationHeader = "Authorization"
	bearerPrefix        = "Bearer "
	bearerPrefixLen     = 7
)

type MiddlewareAdapter interface {
	InternalAuth(a any) error
	ClientAuth(a any) error
	StaffAuth(a any) error
	CustomerAuth(a any) error
	RequirePermission(resource, action string) func(a any) error
}

type middlewareAdapter struct {
	domain domain.Domain
}

func NewMiddlewareAdapter(
	domain domain.Domain,
) MiddlewareAdapter {
	return &middlewareAdapter{
		domain: domain,
	}
}

func (h *middlewareAdapter) extractBearerToken(c *gin.Context) string {
	authHeader := c.GetHeader(authorizationHeader)
	if len(authHeader) > bearerPrefixLen && authHeader[:bearerPrefixLen] == bearerPrefix {
		return authHeader[bearerPrefixLen:]
	}
	return ""
}

func (h *middlewareAdapter) InternalAuth(a any) error {
	c := a.(*gin.Context)
	bearerToken := h.extractBearerToken(c)

	if bearerToken == "" || bearerToken != os.Getenv("INTERNAL_KEY") {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		c.Abort()
		return nil
	}

	c.Next()
	return nil
}

func (h *middlewareAdapter) ClientAuth(a any) error {
	c := a.(*gin.Context)
	ctx := activity.NewContext("http_client_auth")
	bearerToken := h.extractBearerToken(c)

	if bearerToken == "" {
		c.JSON(http.StatusUnauthorized, model.Response{Success: false, Error: "Unauthorized"})
		c.Abort()
		return nil
	}

	authDriver := os.Getenv("AUTH_DRIVER")
	if authDriver == "jwt" {
		jwksURL := os.Getenv("AUTH_JWKS_URL")
		_, err := jwt.ValidateJWTWithURL(bearerToken, jwksURL)
		if err != nil {
			c.JSON(http.StatusUnauthorized, model.Response{Success: false, Error: "Unauthorized: " + err.Error()})
			c.Abort()
			return nil
		}
	} else {
		exists, err := h.domain.Client().IsExists(ctx, bearerToken)
		if err != nil {
			c.JSON(http.StatusInternalServerError, model.Response{Success: false, Error: err.Error()})
			c.Abort()
			return nil
		}
		if !exists {
			c.JSON(http.StatusUnauthorized, model.Response{Success: false, Error: "Unauthorized"})
			c.Abort()
			return nil
		}
	}

	c.Next()
	return nil
}

func (h *middlewareAdapter) StaffAuth(a any) error {
	c := a.(*gin.Context)
	bearerToken := h.extractBearerToken(c)

	if bearerToken == "" {
		c.JSON(http.StatusUnauthorized, model.Response{Success: false, Error: "Unauthorized"})
		c.Abort()
		return nil
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		c.JSON(http.StatusInternalServerError, model.Response{Success: false, Error: "JWT_SECRET not configured"})
		c.Abort()
		return nil
	}

	claims, err := jwt.ParseStaffToken(bearerToken, jwtSecret)
	if err != nil {
		c.JSON(http.StatusUnauthorized, model.Response{Success: false, Error: "Unauthorized: " + err.Error()})
		c.Abort()
		return nil
	}

	c.Set("staff_id", claims.StaffID)
	c.Set("tenant_id", claims.TenantID)
	c.Set("role", claims.Role)
	c.Set("permissions", claims.Permissions)

	c.Next()
	return nil
}

func (h *middlewareAdapter) CustomerAuth(a any) error {
	c := a.(*gin.Context)
	bearerToken := h.extractBearerToken(c)

	if bearerToken == "" {
		c.JSON(http.StatusUnauthorized, model.Response{Success: false, Error: "Unauthorized"})
		c.Abort()
		return nil
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		c.JSON(http.StatusInternalServerError, model.Response{Success: false, Error: "JWT_SECRET not configured"})
		c.Abort()
		return nil
	}

	claims, err := jwt.ParseCustomerToken(bearerToken, jwtSecret)
	if err != nil {
		c.JSON(http.StatusUnauthorized, model.Response{Success: false, Error: "Unauthorized: " + err.Error()})
		c.Abort()
		return nil
	}

	c.Set("customer_id", claims.CustomerID)
	c.Set("tenant_id", claims.TenantID)

	c.Next()
	return nil
}

func (h *middlewareAdapter) RequirePermission(resource, action string) func(a any) error {
	return func(a any) error {
		c := a.(*gin.Context)

		role, _ := c.Get("role")
		roleStr, _ := role.(string)

		// Admin has full access
		if roleStr == "admin" {
			c.Next()
			return nil
		}

		perms, exists := c.Get("permissions")
		if !exists {
			c.JSON(http.StatusForbidden, model.Response{Success: false, Error: "Forbidden: no permissions"})
			c.Abort()
			return nil
		}

		permissions, ok := perms.([]string)
		if !ok {
			c.JSON(http.StatusForbidden, model.Response{Success: false, Error: "Forbidden"})
			c.Abort()
			return nil
		}

		requiredPerm := resource + ":" + action
		managePerm := resource + ":manage"

		for _, p := range permissions {
			if p == requiredPerm || p == managePerm {
				c.Next()
				return nil
			}
		}

		c.JSON(http.StatusForbidden, model.Response{Success: false, Error: "Forbidden: insufficient permissions"})
		c.Abort()
		return nil
	}
}
