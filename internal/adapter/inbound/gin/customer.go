package gin_inbound_adapter

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/palantir/stacktrace"

	"go-template/internal/domain"
	"go-template/internal/model"
	inbound_port "go-template/internal/port/inbound"
	"go-template/utils/activity"
)

type customerAdapter struct {
	domain domain.Domain
}

func NewCustomerAdapter(
	domain domain.Domain,
) inbound_port.CustomerHttpPort {
	return &customerAdapter{
		domain: domain,
	}
}

func (h *customerAdapter) Create(c *gin.Context) {
	ctx := activity.NewContext("http_customer_create")

	// Get router from context (set by RouterAuth middleware)
	routerRaw, exists := c.Get("router")
	if !exists {
		c.JSON(http.StatusBadRequest, model.Response{Success: false, Error: "router context not found; use /mikrotik/:router_id/customers"})
		return
	}
	router, ok := routerRaw.(*model.MikrotikRouter)
	if !ok {
		c.JSON(http.StatusInternalServerError, model.Response{Success: false, Error: "invalid router context"})
		return
	}

	var input model.CustomerInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, model.Response{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	// Inject router_id from URL path
	input.RouterID = &router.ID

	ctx = activity.WithPayload(ctx, input)

	customer, err := h.domain.Customer().Create(ctx, input)
	if err != nil {
		statusCode := http.StatusInternalServerError
		errMsg := stacktrace.RootCause(err).Error()
		if errMsg != "" {
			statusCode = http.StatusBadGateway
		}
		c.JSON(statusCode, model.Response{
			Success: false,
			Error:   errMsg,
		})
		return
	}

	c.JSON(http.StatusCreated, model.Response{
		Success: true,
		Data:    customer,
	})
}

func (h *customerAdapter) GetByID(c *gin.Context) {
	ctx := activity.NewContext("http_customer_get_by_id")
	id := c.Param("id")

	if id == "" {
		c.JSON(http.StatusBadRequest, model.Response{
			Success: false,
			Error:   "id parameter is required",
		})
		return
	}

	ctx = activity.WithPayload(ctx, map[string]string{"id": id})

	customer, err := h.domain.Customer().GetByID(ctx, id)
	if err != nil {
		c.JSON(http.StatusNotFound, model.Response{
			Success: false,
			Error:   stacktrace.RootCause(err).Error(),
		})
		return
	}

	c.JSON(http.StatusOK, model.Response{
		Success: true,
		Data:    customer,
	})
}

func (h *customerAdapter) GetByCode(c *gin.Context) {
	ctx := activity.NewContext("http_customer_get_by_code")
	code := c.Param("code")

	if code == "" {
		c.JSON(http.StatusBadRequest, model.Response{
			Success: false,
			Error:   "code parameter is required",
		})
		return
	}

	ctx = activity.WithPayload(ctx, map[string]string{"code": code})

	customer, err := h.domain.Customer().GetByCode(ctx, code)
	if err != nil {
		c.JSON(http.StatusNotFound, model.Response{
			Success: false,
			Error:   stacktrace.RootCause(err).Error(),
		})
		return
	}

	c.JSON(http.StatusOK, model.Response{
		Success: true,
		Data:    customer,
	})
}

func (h *customerAdapter) List(c *gin.Context) {
	ctx := activity.NewContext("http_customer_list")

	var filter model.CustomerFilter
	if err := c.ShouldBindQuery(&filter); err != nil {
		c.JSON(http.StatusBadRequest, model.Response{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	ctx = activity.WithPayload(ctx, filter)

	customers, err := h.domain.Customer().List(ctx, &filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{
			Success: false,
			Error:   stacktrace.RootCause(err).Error(),
		})
		return
	}

	c.JSON(http.StatusOK, model.Response{
		Success: true,
		Data:    customers,
	})
}

func (h *customerAdapter) Update(c *gin.Context) {
	ctx := activity.NewContext("http_customer_update")
	id := c.Param("id")

	if id == "" {
		c.JSON(http.StatusBadRequest, model.Response{
			Success: false,
			Error:   "id parameter is required",
		})
		return
	}

	// Get router from context (set by RouterAuth middleware)
	routerRaw, exists := c.Get("router")
	if !exists {
		c.JSON(http.StatusBadRequest, model.Response{Success: false, Error: "router context not found; use /mikrotik/:router_id/customers"})
		return
	}
	router, ok := routerRaw.(*model.MikrotikRouter)
	if !ok {
		c.JSON(http.StatusInternalServerError, model.Response{Success: false, Error: "invalid router context"})
		return
	}

	var input model.CustomerInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, model.Response{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	// Inject router_id from URL path
	input.RouterID = &router.ID

	ctx = activity.WithPayload(ctx, map[string]interface{}{
		"id":    id,
		"input": input,
	})

	customer, err := h.domain.Customer().Update(ctx, id, input)
	if err != nil {
		c.JSON(http.StatusBadGateway, model.Response{
			Success: false,
			Error:   stacktrace.RootCause(err).Error(),
		})
		return
	}

	c.JSON(http.StatusOK, model.Response{
		Success: true,
		Data:    customer,
	})
}

func (h *customerAdapter) Delete(c *gin.Context) {
	ctx := activity.NewContext("http_customer_delete")
	id := c.Param("id")

	if id == "" {
		c.JSON(http.StatusBadRequest, model.Response{
			Success: false,
			Error:   "id parameter is required",
		})
		return
	}

	ctx = activity.WithPayload(ctx, map[string]string{"id": id})

	if err := h.domain.Customer().Delete(ctx, id); err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{
			Success: false,
			Error:   stacktrace.RootCause(err).Error(),
		})
		return
	}

	c.JSON(http.StatusOK, model.Response{
		Success: true,
		Data:    gin.H{"message": "Customer deleted successfully"},
	})
}

func (h *customerAdapter) ChangeStatus(c *gin.Context) {
	ctx := activity.NewContext("http_customer_change_status")
	id := c.Param("id")

	if id == "" {
		c.JSON(http.StatusBadRequest, model.Response{
			Success: false,
			Error:   "id parameter is required",
		})
		return
	}

	var input struct {
		Status string `json:"status" binding:"required,oneof=pending active suspended isolated terminated"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, model.Response{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	ctx = activity.WithPayload(ctx, map[string]interface{}{
		"id":     id,
		"status": input.Status,
	})

	status := model.CustomerStatus(input.Status)
	if err := h.domain.Customer().ChangeStatus(ctx, id, status); err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{
			Success: false,
			Error:   stacktrace.RootCause(err).Error(),
		})
		return
	}

	c.JSON(http.StatusOK, model.Response{
		Success: true,
		Data:    gin.H{"message": "Customer status changed successfully", "status": input.Status},
	})
}

func (h *customerAdapter) Isolate(c *gin.Context) {
	ctx := activity.NewContext("http_customer_isolate")
	id := c.Param("id")

	if id == "" {
		c.JSON(http.StatusBadRequest, model.Response{
			Success: false,
			Error:   "id parameter is required",
		})
		return
	}

	ctx = activity.WithPayload(ctx, map[string]string{"id": id})

	if err := h.domain.Customer().Isolate(ctx, id); err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{
			Success: false,
			Error:   stacktrace.RootCause(err).Error(),
		})
		return
	}

	c.JSON(http.StatusOK, model.Response{
		Success: true,
		Data:    gin.H{"message": "Customer isolated successfully"},
	})
}

func (h *customerAdapter) UnIsolate(c *gin.Context) {
	ctx := activity.NewContext("http_customer_unisolate")
	id := c.Param("id")

	if id == "" {
		c.JSON(http.StatusBadRequest, model.Response{
			Success: false,
			Error:   "id parameter is required",
		})
		return
	}

	ctx = activity.WithPayload(ctx, map[string]string{"id": id})

	if err := h.domain.Customer().UnIsolate(ctx, id); err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{
			Success: false,
			Error:   stacktrace.RootCause(err).Error(),
		})
		return
	}

	c.JSON(http.StatusOK, model.Response{
		Success: true,
		Data:    gin.H{"message": "Customer un-isolated successfully"},
	})
}

func (h *customerAdapter) SyncToMikrotik(c *gin.Context) {
	ctx := activity.NewContext("http_customer_sync_to_mikrotik")
	id := c.Param("id")

	if id == "" {
		c.JSON(http.StatusBadRequest, model.Response{
			Success: false,
			Error:   "id parameter is required",
		})
		return
	}

	ctx = activity.WithPayload(ctx, map[string]string{"id": id})

	if err := h.domain.Customer().SyncToMikrotik(ctx, id); err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{
			Success: false,
			Error:   stacktrace.RootCause(err).Error(),
		})
		return
	}

	c.JSON(http.StatusOK, model.Response{
		Success: true,
		Data:    gin.H{"message": "Customer synced to MikroTik successfully"},
	})
}
