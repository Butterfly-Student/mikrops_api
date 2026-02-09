package gin_inbound_adapter

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/palantir/stacktrace"

	"mikrops/internal/domain"
	"mikrops/internal/model"
	inbound_port "mikrops/internal/port/inbound"
	"mikrops/utils/activity"
)

type customerRegistrationAdapter struct {
	domain domain.Domain
}

func NewCustomerRegistrationAdapter(domain domain.Domain) inbound_port.CustomerRegistrationHttpPort {
	return &customerRegistrationAdapter{domain: domain}
}

func (h *customerRegistrationAdapter) List(a any) error {
	c := a.(*gin.Context)
	ctx := activity.NewContext("http_customer_registration_list")

	tenantID, _ := c.Get("tenant_id")
	tid, _ := tenantID.(string)

	filter := model.CustomerRegistrationFilter{
		TenantIDs:   []string{tid},
		WithPackage: true,
		WithNas:     true,
	}

	status := c.Query("status")
	if status != "" {
		filter.Statuses = []string{status}
	}

	results, err := h.domain.CustomerRegistration().FindByFilter(ctx, filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{Success: false, Error: stacktrace.RootCause(err).Error()})
		return nil
	}

	c.JSON(http.StatusOK, model.Response{Success: true, Data: results})
	return nil
}

func (h *customerRegistrationAdapter) Get(a any) error {
	c := a.(*gin.Context)
	ctx := activity.NewContext("http_customer_registration_get")

	id := c.Param("id")

	result, err := h.domain.CustomerRegistration().FindByID(ctx, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{Success: false, Error: stacktrace.RootCause(err).Error()})
		return nil
	}

	c.JSON(http.StatusOK, model.Response{Success: true, Data: result})
	return nil
}

func (h *customerRegistrationAdapter) Create(a any) error {
	c := a.(*gin.Context)
	ctx := activity.NewContext("http_customer_registration_create")

	tenantID, _ := c.Get("tenant_id")
	tid, _ := tenantID.(string)

	var payload model.CustomerRegistrationInput
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, model.Response{Success: false, Error: err.Error()})
		return nil
	}
	payload.TenantID = tid

	result, err := h.domain.CustomerRegistration().Create(ctx, payload)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{Success: false, Error: stacktrace.RootCause(err).Error()})
		return nil
	}

	c.JSON(http.StatusCreated, model.Response{Success: true, Data: result})
	return nil
}

func (h *customerRegistrationAdapter) Approve(a any) error {
	c := a.(*gin.Context)
	ctx := activity.NewContext("http_customer_registration_approve")

	id := c.Param("id")
	staffID, _ := c.Get("staff_id")
	sid, _ := staffID.(string)

	result, err := h.domain.CustomerRegistration().Approve(ctx, id, sid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{Success: false, Error: stacktrace.RootCause(err).Error()})
		return nil
	}

	c.JSON(http.StatusOK, model.Response{Success: true, Data: result})
	return nil
}

func (h *customerRegistrationAdapter) Reject(a any) error {
	c := a.(*gin.Context)
	ctx := activity.NewContext("http_customer_registration_reject")

	id := c.Param("id")

	var payload struct {
		Reason string `json:"reason"`
	}
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, model.Response{Success: false, Error: err.Error()})
		return nil
	}

	result, err := h.domain.CustomerRegistration().Reject(ctx, id, payload.Reason)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{Success: false, Error: stacktrace.RootCause(err).Error()})
		return nil
	}

	c.JSON(http.StatusOK, model.Response{Success: true, Data: result})
	return nil
}
