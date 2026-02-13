package gin_inbound_adapter

import (
	"fmt"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"

	"go-template/internal/domain"
	"go-template/internal/model"
	inbound_port "go-template/internal/port/inbound"
	"go-template/internal/utils/token"
	"go-template/utils/activity"
	"go-template/utils/jwt"
)

const (
	authorizationHeader = "Authorization"
	bearerPrefix        = "Bearer "
	bearerPrefixLen     = 7
)

type middlewareAdapter struct {
	domain domain.Domain
}

func NewMiddlewareAdapter(
	domain domain.Domain,
) inbound_port.MiddlewareHttpPort {
	return &middlewareAdapter{
		domain: domain,
	}
}

func (h *middlewareAdapter) InternalAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader(authorizationHeader)
		var bearerToken string

		if len(authHeader) > bearerPrefixLen && authHeader[:bearerPrefixLen] == bearerPrefix {
			bearerToken = authHeader[bearerPrefixLen:]
		}

		if bearerToken == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Unauthorized",
			})
			return
		}

		if bearerToken != os.Getenv("INTERNAL_KEY") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Unauthorized",
			})
			return
		}

		c.Next()
	}
}

func (h *middlewareAdapter) UserAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader(authorizationHeader)
		var bearerToken string

		if len(authHeader) > bearerPrefixLen && authHeader[:bearerPrefixLen] == bearerPrefix {
			bearerToken = authHeader[bearerPrefixLen:]
		} else {
			bearerToken = authHeader
		}

		if bearerToken == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			return
		}

		claims, err := token.ValidateToken(bearerToken, false)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token: " + err.Error()})
			return
		}

		userIDFloat, ok := claims["sub"].(float64)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
			return
		}

		role, _ := claims["role"].(string)

		c.Set("userID", uint(userIDFloat))
		c.Set("role", role)
		c.Next()
	}
}

func (h *middlewareAdapter) RBAC() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("userID")
		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		sub := fmt.Sprintf("%d", userID.(uint))
		obj := c.Request.URL.Path
		act := c.Request.Method

		ok, err := h.domain.Auth().Enforce(sub, obj, act)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Error checking permissions: " + err.Error()})
			return
		}

		if !ok {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Forbidden"})
			return
		}

		c.Next()
	}
}

func (h *middlewareAdapter) ClientAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := activity.NewContext("http_client_auth")
		authHeader := c.GetHeader(authorizationHeader)
		var bearerToken string

		if len(authHeader) > bearerPrefixLen && authHeader[:bearerPrefixLen] == bearerPrefix {
			bearerToken = authHeader[bearerPrefixLen:]
		}

		if bearerToken == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, model.Response{
				Success: false,
				Error:   "Unauthorized",
			})
			return
		}

		authDriver := os.Getenv("AUTH_DRIVER")
		if authDriver == "jwt" {
			jwksURL := os.Getenv("AUTH_JWKS_URL")

			_, err := jwt.ValidateJWTWithURL(bearerToken, jwksURL)
			if err != nil {
				c.AbortWithStatusJSON(http.StatusUnauthorized, model.Response{
					Success: false,
					Error:   "Unauthorized: " + err.Error(),
				})
				return
			}
		} else {
			exists, err := h.domain.Client().IsExists(ctx, bearerToken)
			if err != nil {
				c.AbortWithStatusJSON(http.StatusInternalServerError, model.Response{
					Success: false,
					Error:   err.Error(),
				})
				return
			}

			if !exists {
				c.AbortWithStatusJSON(http.StatusUnauthorized, model.Response{
					Success: false,
					Error:   "Unauthorized",
				})
				return
			}
		}

		c.Next()
	}
}
