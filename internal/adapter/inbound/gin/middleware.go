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

func (h *middlewareAdapter) InternalAuth(a any) error {
	c := a.(*gin.Context)
	authHeader := c.GetHeader(authorizationHeader)
	var bearerToken string
	if len(authHeader) > bearerPrefixLen && authHeader[:bearerPrefixLen] == bearerPrefix {
		bearerToken = authHeader[bearerPrefixLen:]
	}

	if bearerToken == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Unauthorized",
		})
		c.Abort()
		return nil
	}

	if bearerToken != os.Getenv("INTERNAL_KEY") {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Unauthorized",
		})
		c.Abort()
		return nil
	}

	c.Next()
	return nil
}

func (h *middlewareAdapter) ClientAuth(a any) error {
	c := a.(*gin.Context)
	ctx := activity.NewContext("http_client_auth")
	authHeader := c.GetHeader(authorizationHeader)
	var bearerToken string
	if len(authHeader) > bearerPrefixLen && authHeader[:bearerPrefixLen] == bearerPrefix {
		bearerToken = authHeader[bearerPrefixLen:]
	}

	if bearerToken == "" {
		c.JSON(http.StatusUnauthorized, model.Response{
			Success: false,
			Error:   "Unauthorized",
		})
		c.Abort()
		return nil
	}

	authDriver := os.Getenv("AUTH_DRIVER")
	if authDriver == "jwt" {
		jwksURL := os.Getenv("AUTH_JWKS_URL")

		_, err := jwt.ValidateJWTWithURL(bearerToken, jwksURL)
		if err != nil {
			c.JSON(http.StatusUnauthorized, model.Response{
				Success: false,
				Error:   "Unauthorized: " + err.Error(),
			})
			c.Abort()
			return nil
		}
	} else {
		exists, err := h.domain.Client().IsExists(ctx, bearerToken)
		if err != nil {
			c.JSON(http.StatusInternalServerError, model.Response{
				Success: false,
				Error:   err.Error(),
			})
			c.Abort()
			return nil
		}

		if !exists {
			c.JSON(http.StatusUnauthorized, model.Response{
				Success: false,
				Error:   "Unauthorized",
			})
			c.Abort()
			return nil
		}
	}

	c.Next()
	return nil
}
