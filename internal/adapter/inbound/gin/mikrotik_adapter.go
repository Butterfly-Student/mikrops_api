package gin_inbound_adapter

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"go-template/internal/domain"
	"go-template/internal/domain/mikrotik"
)

type mikrotikAdapter struct {
	domain domain.Domain
}

func NewMikrotikHandler(domain domain.Domain) *mikrotikAdapter {
	return &mikrotikAdapter{
		domain: domain,
	}
}

func (h *mikrotikAdapter) CreateRouter(c *gin.Context) {
	var req struct {
		Name     string `json:"name" binding:"required"`
		Address  string `json:"address" binding:"required"`
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
		ApiPort  *int   `json:"api_port"`
		RestPort *int   `json:"rest_port"`
		UseSSL   *bool  `json:"use_ssl"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	router, err := h.domain.Mikrotik().Create(req.Name, req.Address, req.Username, req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, router)
}

func (h *mikrotikAdapter) FindRouterByID(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "router id is required"})
		return
	}

	router, err := h.domain.Mikrotik().FindByID(id)
	if err != nil {
		if err == mikrotik.ErrRouterNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "router not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, router)
}

func (h *mikrotikAdapter) ListRouters(c *gin.Context) {
	routers, err := h.domain.Mikrotik().FindAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, routers)
}

func (h *mikrotikAdapter) UpdateRouter(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "router id is required"})
		return
	}

	var req struct {
		Name              *string `json:"name"`
		Address           *string `json:"address"`
		Username          *string `json:"username"`
		Password          *string `json:"password"`
		ApiPort           *int    `json:"api_port"`
		RestPort          *int    `json:"rest_port"`
		UseSSL            *bool   `json:"use_ssl"`
		PasswordEncrypted *string `json:"password_encrypted"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updates := make(map[string]interface{})
	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.Address != nil {
		updates["address"] = *req.Address
	}
	if req.Username != nil {
		updates["username"] = *req.Username
	}
	if req.Password != nil {
		updates["password"] = *req.Password
	}
	if req.ApiPort != nil {
		updates["api_port"] = *req.ApiPort
	}
	if req.RestPort != nil {
		updates["rest_port"] = *req.RestPort
	}
	if req.UseSSL != nil {
		updates["use_ssl"] = *req.UseSSL
	}
	if req.PasswordEncrypted != nil {
		updates["password_encrypted"] = *req.PasswordEncrypted
	}

	router, err := h.domain.Mikrotik().Update(id, updates)
	if err != nil {
		if err == mikrotik.ErrRouterNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "router not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, router)
}

func (h *mikrotikAdapter) DeleteRouter(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "router id is required"})
		return
	}

	if err := h.domain.Mikrotik().Delete(id); err != nil {
		if err == mikrotik.ErrRouterNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "router not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "router deleted successfully"})
}

func (h *mikrotikAdapter) SetActiveRouter(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "router id is required"})
		return
	}

	router, err := h.domain.Mikrotik().SetActive(id)
	if err != nil {
		if err == mikrotik.ErrRouterNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "router not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, router)
}

func (h *mikrotikAdapter) GetActiveRouter(c *gin.Context) {
	router, err := h.domain.Mikrotik().GetActive()
	if err != nil {
		if err == mikrotik.ErrNoActiveRouter {
			c.JSON(http.StatusNotFound, gin.H{"error": "no active router found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, router)
}

func (h *mikrotikAdapter) TestRouterConnection(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "router id is required"})
		return
	}

	if err := h.domain.Mikrotik().TestConnection(id); err != nil {
		if err == mikrotik.ErrRouterNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "router not found"})
			return
		}
		if err == mikrotik.ErrConnectionFailed {
			c.JSON(http.StatusBadGateway, gin.H{"error": "failed to connect to router"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "connection successful"})
}
