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

type bandwidthProfileAdapter struct {
	domain domain.Domain
}

func NewBandwidthProfileAdapter(
	domain domain.Domain,
) inbound_port.BandwidthProfileHttpPort {
	return &bandwidthProfileAdapter{
		domain: domain,
	}
}

func (h *bandwidthProfileAdapter) Create(c *gin.Context) {
	ctx := activity.NewContext("http_bandwidth_profile_create")

	var input model.BandwidthProfileInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, model.Response{Success: false, Error: err.Error()})
		return
	}

	ctx = activity.WithPayload(ctx, input)

	profile, err := h.domain.BandwidthProfile().Create(ctx, input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{
			Success: false,
			Error:   stacktrace.RootCause(err).Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, model.Response{Success: true, Data: profile})
}

// CreateWithRouter handles POST /mikrotik/:router_id/bandwidth-profiles
// The MikroTik router is resolved by RouterAuth middleware and stored in gin context as "router".
func (h *bandwidthProfileAdapter) CreateWithRouter(c *gin.Context) {
	ctx := activity.NewContext("http_bandwidth_profile_create_with_router")

	// Get router from gin context (set by RouterAuth middleware)
	routerRaw, exists := c.Get("router")
	if !exists {
		c.JSON(http.StatusBadRequest, model.Response{Success: false, Error: "router context not found; use /mikrotik/:router_id/bandwidth-profiles"})
		return
	}
	router, ok := routerRaw.(*model.MikrotikRouter)
	if !ok {
		c.JSON(http.StatusInternalServerError, model.Response{Success: false, Error: "invalid router context"})
		return
	}

	var input model.BandwidthProfileInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, model.Response{Success: false, Error: err.Error()})
		return
	}

	ctx = activity.WithPayload(ctx, input)

	profile, err := h.domain.BandwidthProfile().CreateWithRouter(ctx, router.ID.String(), input)
	if err != nil {
		statusCode := http.StatusInternalServerError
		errMsg := stacktrace.RootCause(err).Error()
		// If MikroTik connection failed, return 502 to distinguish from server errors
		if errMsg != "" {
			statusCode = http.StatusBadGateway
		}
		c.JSON(statusCode, model.Response{Success: false, Error: errMsg})
		return
	}

	c.JSON(http.StatusCreated, model.Response{Success: true, Data: profile})
}

func (h *bandwidthProfileAdapter) GetByID(c *gin.Context) {
	ctx := activity.NewContext("http_bandwidth_profile_get_by_id")
	id := c.Param("id")

	if id == "" {
		c.JSON(http.StatusBadRequest, model.Response{
			Success: false,
			Error:   "id parameter is required",
		})
		return
	}

	ctx = activity.WithPayload(ctx, map[string]string{"id": id})

	profile, err := h.domain.BandwidthProfile().GetByID(ctx, id)
	if err != nil {
		c.JSON(http.StatusNotFound, model.Response{
			Success: false,
			Error:   stacktrace.RootCause(err).Error(),
		})
		return
	}

	c.JSON(http.StatusOK, model.Response{
		Success: true,
		Data:    profile,
	})
}

func (h *bandwidthProfileAdapter) GetByCode(c *gin.Context) {
	ctx := activity.NewContext("http_bandwidth_profile_get_by_code")
	code := c.Param("code")

	if code == "" {
		c.JSON(http.StatusBadRequest, model.Response{
			Success: false,
			Error:   "code parameter is required",
		})
		return
	}

	ctx = activity.WithPayload(ctx, map[string]string{"code": code})

	profile, err := h.domain.BandwidthProfile().GetByCode(ctx, code)
	if err != nil {
		c.JSON(http.StatusNotFound, model.Response{
			Success: false,
			Error:   stacktrace.RootCause(err).Error(),
		})
		return
	}

	c.JSON(http.StatusOK, model.Response{
		Success: true,
		Data:    profile,
	})
}

func (h *bandwidthProfileAdapter) List(c *gin.Context) {
	ctx := activity.NewContext("http_bandwidth_profile_list")

	var filter model.BandwidthProfileFilter
	if err := c.ShouldBindQuery(&filter); err != nil {
		c.JSON(http.StatusBadRequest, model.Response{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	ctx = activity.WithPayload(ctx, filter)

	profiles, err := h.domain.BandwidthProfile().List(ctx, &filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{
			Success: false,
			Error:   stacktrace.RootCause(err).Error(),
		})
		return
	}

	c.JSON(http.StatusOK, model.Response{
		Success: true,
		Data:    profiles,
	})
}

func (h *bandwidthProfileAdapter) Update(c *gin.Context) {
	ctx := activity.NewContext("http_bandwidth_profile_update")
	id := c.Param("id")

	if id == "" {
		c.JSON(http.StatusBadRequest, model.Response{
			Success: false,
			Error:   "id parameter is required",
		})
		return
	}

	var input model.BandwidthProfileInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, model.Response{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	ctx = activity.WithPayload(ctx, map[string]interface{}{
		"id":    id,
		"input": input,
	})

	profile, err := h.domain.BandwidthProfile().Update(ctx, id, input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{
			Success: false,
			Error:   stacktrace.RootCause(err).Error(),
		})
		return
	}

	c.JSON(http.StatusOK, model.Response{
		Success: true,
		Data:    profile,
	})
}

func (h *bandwidthProfileAdapter) Delete(c *gin.Context) {
	ctx := activity.NewContext("http_bandwidth_profile_delete")
	id := c.Param("id")

	if id == "" {
		c.JSON(http.StatusBadRequest, model.Response{
			Success: false,
			Error:   "id parameter is required",
		})
		return
	}

	ctx = activity.WithPayload(ctx, map[string]string{"id": id})

	if err := h.domain.BandwidthProfile().Delete(ctx, id); err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{
			Success: false,
			Error:   stacktrace.RootCause(err).Error(),
		})
		return
	}

	c.JSON(http.StatusOK, model.Response{
		Success: true,
		Data:    gin.H{"message": "Bandwidth profile deleted successfully"},
	})
}

func (h *bandwidthProfileAdapter) SyncToMikrotik(c *gin.Context) {
	ctx := activity.NewContext("http_bandwidth_profile_sync_to_mikrotik")
	id := c.Param("id")

	if id == "" {
		c.JSON(http.StatusBadRequest, model.Response{
			Success: false,
			Error:   "id parameter is required",
		})
		return
	}

	routerID := c.Query("router_id")
	if routerID == "" {
		c.JSON(http.StatusBadRequest, model.Response{
			Success: false,
			Error:   "router_id query parameter is required",
		})
		return
	}

	ctx = activity.WithPayload(ctx, map[string]string{
		"id":        id,
		"router_id": routerID,
	})

	if err := h.domain.BandwidthProfile().SyncToMikrotik(ctx, id, routerID); err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{
			Success: false,
			Error:   stacktrace.RootCause(err).Error(),
		})
		return
	}

	c.JSON(http.StatusOK, model.Response{
		Success: true,
		Data:    gin.H{"message": "Bandwidth profile synced to MikroTik successfully"},
	})
}

// UpdateWithRouter handles PUT /mikrotik/:router_id/bandwidth-profiles/:id (MikroTik-first)
func (h *bandwidthProfileAdapter) UpdateWithRouter(c *gin.Context) {
	ctx := activity.NewContext("http_bandwidth_profile_update_with_router")

	routerRaw, exists := c.Get("router")
	if !exists {
		c.JSON(http.StatusBadRequest, model.Response{Success: false, Error: "router context not found"})
		return
	}
	router, ok := routerRaw.(*model.MikrotikRouter)
	if !ok {
		c.JSON(http.StatusInternalServerError, model.Response{Success: false, Error: "invalid router context"})
		return
	}

	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, model.Response{Success: false, Error: "id parameter is required"})
		return
	}

	var input model.BandwidthProfileInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, model.Response{Success: false, Error: err.Error()})
		return
	}

	ctx = activity.WithPayload(ctx, input)

	profile, err := h.domain.BandwidthProfile().UpdateWithRouter(ctx, router.ID.String(), id, input)
	if err != nil {
		c.JSON(http.StatusBadGateway, model.Response{
			Success: false,
			Error:   stacktrace.RootCause(err).Error(),
		})
		return
	}

	c.JSON(http.StatusOK, model.Response{Success: true, Data: profile})
}

// DeleteWithRouter handles DELETE /mikrotik/:router_id/bandwidth-profiles/:id (MikroTik-first)
func (h *bandwidthProfileAdapter) DeleteWithRouter(c *gin.Context) {
	ctx := activity.NewContext("http_bandwidth_profile_delete_with_router")

	routerRaw, exists := c.Get("router")
	if !exists {
		c.JSON(http.StatusBadRequest, model.Response{Success: false, Error: "router context not found"})
		return
	}
	router, ok := routerRaw.(*model.MikrotikRouter)
	if !ok {
		c.JSON(http.StatusInternalServerError, model.Response{Success: false, Error: "invalid router context"})
		return
	}

	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, model.Response{Success: false, Error: "id parameter is required"})
		return
	}

	if err := h.domain.BandwidthProfile().DeleteWithRouter(ctx, router.ID.String(), id); err != nil {
		c.JSON(http.StatusBadGateway, model.Response{
			Success: false,
			Error:   stacktrace.RootCause(err).Error(),
		})
		return
	}

	c.JSON(http.StatusOK, model.Response{Success: true})
}
