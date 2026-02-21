package gin_inbound_adapter

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/palantir/stacktrace"

	customer_portal "go-template/internal/domain/customer_portal"
	"go-template/internal/domain"
	"go-template/internal/model"
	inbound_port "go-template/internal/port/inbound"
	"go-template/utils/activity"
)

type customerPortalAdapter struct {
	domain domain.Domain
}

func NewCustomerPortalAdapter(
	domain domain.Domain,
) inbound_port.CustomerPortalHttpPort {
	return &customerPortalAdapter{
		domain: domain,
	}
}

// Login — public, no auth required
func (h *customerPortalAdapter) Login(c *gin.Context) {
	ctx := activity.NewContext("http_portal_login")

	var input struct {
		Identifier string `json:"identifier" binding:"required"`
		Password   string `json:"password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, model.Response{Success: false, Error: err.Error()})
		return
	}

	resp, err := h.domain.CustomerPortal().Login(ctx, input.Identifier, input.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, model.Response{
			Success: false,
			Error:   stacktrace.RootCause(err).Error(),
		})
		return
	}

	c.JSON(http.StatusOK, model.Response{Success: true, Data: resp})
}

// RefreshToken — public, no auth required
func (h *customerPortalAdapter) RefreshToken(c *gin.Context) {
	ctx := activity.NewContext("http_portal_refresh")

	var input struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, model.Response{Success: false, Error: err.Error()})
		return
	}

	resp, err := h.domain.CustomerPortal().RefreshToken(ctx, input.RefreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, model.Response{
			Success: false,
			Error:   stacktrace.RootCause(err).Error(),
		})
		return
	}

	c.JSON(http.StatusOK, model.Response{Success: true, Data: resp})
}

// GetProfile — CustomerPortalAuth required
func (h *customerPortalAdapter) GetProfile(c *gin.Context) {
	ctx := activity.NewContext("http_portal_get_profile")
	customerID := c.GetString("customerID")

	profile, err := h.domain.CustomerPortal().GetProfile(ctx, customerID)
	if err != nil {
		c.JSON(http.StatusNotFound, model.Response{
			Success: false,
			Error:   stacktrace.RootCause(err).Error(),
		})
		return
	}

	c.JSON(http.StatusOK, model.Response{Success: true, Data: profile})
}

// UpdateProfile — CustomerPortalAuth required
func (h *customerPortalAdapter) UpdateProfile(c *gin.Context) {
	ctx := activity.NewContext("http_portal_update_profile")
	customerID := c.GetString("customerID")

	var input customer_portal.UpdateProfileInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, model.Response{Success: false, Error: err.Error()})
		return
	}

	updated, err := h.domain.CustomerPortal().UpdateProfile(ctx, customerID, input)
	if err != nil {
		c.JSON(http.StatusBadRequest, model.Response{
			Success: false,
			Error:   stacktrace.RootCause(err).Error(),
		})
		return
	}

	c.JSON(http.StatusOK, model.Response{Success: true, Data: updated})
}

// ChangePassword — CustomerPortalAuth required
func (h *customerPortalAdapter) ChangePassword(c *gin.Context) {
	ctx := activity.NewContext("http_portal_change_password")
	customerID := c.GetString("customerID")

	var input struct {
		OldPassword string `json:"old_password" binding:"required"`
		NewPassword string `json:"new_password" binding:"required,min=6"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, model.Response{Success: false, Error: err.Error()})
		return
	}

	if err := h.domain.CustomerPortal().ChangePassword(ctx, customerID, input.OldPassword, input.NewPassword); err != nil {
		c.JSON(http.StatusBadRequest, model.Response{
			Success: false,
			Error:   stacktrace.RootCause(err).Error(),
		})
		return
	}

	c.JSON(http.StatusOK, model.Response{Success: true, Data: "password updated successfully"})
}

// ChangePppCredentials — CustomerPortalAuth required
func (h *customerPortalAdapter) ChangePppCredentials(c *gin.Context) {
	ctx := activity.NewContext("http_portal_change_ppp")
	customerID := c.GetString("customerID")

	var input customer_portal.ChangePppInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, model.Response{Success: false, Error: err.Error()})
		return
	}

	updated, err := h.domain.CustomerPortal().ChangePppCredentials(ctx, customerID, input)
	if err != nil {
		c.JSON(http.StatusBadRequest, model.Response{
			Success: false,
			Error:   stacktrace.RootCause(err).Error(),
		})
		return
	}

	c.JSON(http.StatusOK, model.Response{Success: true, Data: updated})
}

// ListInvoices — CustomerPortalAuth required
func (h *customerPortalAdapter) ListInvoices(c *gin.Context) {
	ctx := activity.NewContext("http_portal_list_invoices")
	customerID := c.GetString("customerID")

	invoices, err := h.domain.CustomerPortal().ListInvoices(ctx, customerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{
			Success: false,
			Error:   stacktrace.RootCause(err).Error(),
		})
		return
	}

	c.JSON(http.StatusOK, model.Response{Success: true, Data: invoices})
}

// GetInvoice — CustomerPortalAuth required
func (h *customerPortalAdapter) GetInvoice(c *gin.Context) {
	ctx := activity.NewContext("http_portal_get_invoice")
	customerID := c.GetString("customerID")
	invoiceID := c.Param("id")

	invoice, err := h.domain.CustomerPortal().GetInvoice(ctx, customerID, invoiceID)
	if err != nil {
		c.JSON(http.StatusNotFound, model.Response{
			Success: false,
			Error:   stacktrace.RootCause(err).Error(),
		})
		return
	}

	c.JSON(http.StatusOK, model.Response{Success: true, Data: invoice})
}
