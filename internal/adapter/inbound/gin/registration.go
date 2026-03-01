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

type registrationAdapter struct {
	domain domain.Domain
}

func NewRegistrationAdapter(
	domain domain.Domain,
) inbound_port.RegistrationHttpPort {
	return &registrationAdapter{
		domain: domain,
	}
}

// Submit — public, no auth required
func (h *registrationAdapter) Submit(c *gin.Context) {
	ctx := activity.NewContext("http_registration_submit")

	var input model.RegistrationInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, model.Response{Success: false, Error: err.Error()})
		return
	}

	ctx = activity.WithPayload(ctx, input)

	reg, err := h.domain.Registration().Submit(ctx, input)
	if err != nil {
		c.JSON(http.StatusBadRequest, model.Response{
			Success: false,
			Error:   stacktrace.RootCause(err).Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, model.Response{Success: true, Data: reg})
}

// GetByID — admin/user auth
func (h *registrationAdapter) GetByID(c *gin.Context) {
	ctx := activity.NewContext("http_registration_get_by_id")
	id := c.Param("id")

	reg, err := h.domain.Registration().GetByID(ctx, id)
	if err != nil {
		c.JSON(http.StatusNotFound, model.Response{
			Success: false,
			Error:   stacktrace.RootCause(err).Error(),
		})
		return
	}

	c.JSON(http.StatusOK, model.Response{Success: true, Data: reg})
}

// List — admin/user auth
func (h *registrationAdapter) List(c *gin.Context) {
	ctx := activity.NewContext("http_registration_list")

	var filter model.RegistrationFilter
	_ = c.ShouldBindQuery(&filter)

	regs, err := h.domain.Registration().List(ctx, &filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{
			Success: false,
			Error:   stacktrace.RootCause(err).Error(),
		})
		return
	}

	c.JSON(http.StatusOK, model.Response{Success: true, Data: regs})
}

// Approve — admin/user auth; reads approverID from context
func (h *registrationAdapter) Approve(c *gin.Context) {
	ctx := activity.NewContext("http_registration_approve")
	id := c.Param("id")

	approverRaw, _ := c.Get("userID")
	approverID, _ := approverRaw.(string)

	var input model.RegistrationApproveInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, model.Response{Success: false, Error: err.Error()})
		return
	}

	result, err := h.domain.Registration().Approve(ctx, id, approverID, input)
	if err != nil {
		c.JSON(http.StatusBadRequest, model.Response{
			Success: false,
			Error:   stacktrace.RootCause(err).Error(),
		})
		return
	}

	c.JSON(http.StatusOK, model.Response{Success: true, Data: result})
}

// Reject — admin/user auth
func (h *registrationAdapter) Reject(c *gin.Context) {
	ctx := activity.NewContext("http_registration_reject")
	id := c.Param("id")

	approverRaw, _ := c.Get("userID")
	approverID, _ := approverRaw.(string)

	var input model.RegistrationRejectInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, model.Response{Success: false, Error: err.Error()})
		return
	}

	reg, err := h.domain.Registration().Reject(ctx, id, approverID, input)
	if err != nil {
		c.JSON(http.StatusBadRequest, model.Response{
			Success: false,
			Error:   stacktrace.RootCause(err).Error(),
		})
		return
	}

	c.JSON(http.StatusOK, model.Response{Success: true, Data: reg})
}
