package gin_inbound_adapter

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/palantir/stacktrace"

	"go-template/internal/domain"
	"go-template/internal/model"
	inbound_port "go-template/internal/port/inbound"
	"go-template/utils/activity"
)

type mikrotikRouterAdapter struct {
	domain domain.Domain
}

func NewMikrotikRouterAdapter(domain domain.Domain) inbound_port.MikrotikRouterHttpPort {
	return &mikrotikRouterAdapter{domain: domain}
}

func (h *mikrotikRouterAdapter) Create(c *gin.Context) {
	ctx := activity.NewContext("http_mikrotik_router_create")

	var input model.MikrotikRouterInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, model.Response{Success: false, Error: err.Error()})
		return
	}

	ctx = activity.WithPayload(ctx, input)

	router, err := h.domain.MikrotikRouter().Create(ctx, input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{
			Success: false,
			Error:   stacktrace.RootCause(err).Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, model.Response{Success: true, Data: router})
}

func (h *mikrotikRouterAdapter) List(c *gin.Context) {
	ctx := activity.NewContext("http_mikrotik_router_list")

	routers, err := h.domain.MikrotikRouter().List(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{
			Success: false,
			Error:   stacktrace.RootCause(err).Error(),
		})
		return
	}

	c.JSON(http.StatusOK, model.Response{Success: true, Data: routers})
}

func (h *mikrotikRouterAdapter) GetByID(c *gin.Context) {
	ctx := activity.NewContext("http_mikrotik_router_get_by_id")
	id := c.Param("id")

	if _, err := uuid.Parse(id); err != nil {
		c.JSON(http.StatusBadRequest, model.Response{Success: false, Error: "invalid id format"})
		return
	}

	router, err := h.domain.MikrotikRouter().GetByID(ctx, id)
	if err != nil {
		c.JSON(http.StatusNotFound, model.Response{
			Success: false,
			Error:   stacktrace.RootCause(err).Error(),
		})
		return
	}

	c.JSON(http.StatusOK, model.Response{Success: true, Data: router})
}

func (h *mikrotikRouterAdapter) Update(c *gin.Context) {
	ctx := activity.NewContext("http_mikrotik_router_update")
	id := c.Param("id")

	if _, err := uuid.Parse(id); err != nil {
		c.JSON(http.StatusBadRequest, model.Response{Success: false, Error: "invalid id format"})
		return
	}

	var input model.MikrotikRouterInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, model.Response{Success: false, Error: err.Error()})
		return
	}

	router, err := h.domain.MikrotikRouter().Update(ctx, id, input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{
			Success: false,
			Error:   stacktrace.RootCause(err).Error(),
		})
		return
	}

	c.JSON(http.StatusOK, model.Response{Success: true, Data: router})
}

func (h *mikrotikRouterAdapter) Delete(c *gin.Context) {
	ctx := activity.NewContext("http_mikrotik_router_delete")
	id := c.Param("id")

	if _, err := uuid.Parse(id); err != nil {
		c.JSON(http.StatusBadRequest, model.Response{Success: false, Error: "invalid id format"})
		return
	}

	if err := h.domain.MikrotikRouter().Delete(ctx, id); err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{
			Success: false,
			Error:   stacktrace.RootCause(err).Error(),
		})
		return
	}

	c.JSON(http.StatusOK, model.Response{
		Success: true,
		Data:    gin.H{"message": "MikroTik router deleted successfully"},
	})
}

func (h *mikrotikRouterAdapter) TestConnection(c *gin.Context) {
	ctx := activity.NewContext("http_mikrotik_router_test_connection")
	id := c.Param("id")

	if _, err := uuid.Parse(id); err != nil {
		c.JSON(http.StatusBadRequest, model.Response{Success: false, Error: "invalid id format"})
		return
	}

	result, err := h.domain.MikrotikRouter().TestConnection(ctx, id)
	if err != nil {
		c.JSON(http.StatusNotFound, model.Response{
			Success: false,
			Error:   stacktrace.RootCause(err).Error(),
		})
		return
	}

	c.JSON(http.StatusOK, model.Response{Success: result.Success, Data: result})
}
