package inbound_port

import "github.com/gin-gonic/gin"

//go:generate mockgen -source=middleware.go -destination=./../../../tests/mocks/port/mock_middleware.go
type MiddlewareHttpPort interface {
	InternalAuth() gin.HandlerFunc
	ClientAuth() gin.HandlerFunc
	UserAuth() gin.HandlerFunc
	RBAC() gin.HandlerFunc
}
