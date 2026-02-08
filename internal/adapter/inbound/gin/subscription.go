package gin_inbound_adapter

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/palantir/stacktrace"

	"mikrops/internal/domain"
	"mikrops/internal/model"
	inbound_port "mikrops/internal/port/inbound"
	"mikrops/utils/activity"
)

type subscriptionAdapter struct {
	domain domain.Domain
}

func NewSubscriptionAdapter(domain domain.Domain) inbound_port.SubscriptionHttpPort {
	return &subscriptionAdapter{domain: domain}
}

func (h *subscriptionAdapter) List(a any) error {
	c := a.(*gin.Context)
	ctx := activity.NewContext("http_subscription_list")

	tenantID, _ := c.Get("tenant_id")
	tid, _ := tenantID.(string)

	filter := model.SubscriptionFilter{
		TenantIDs:    []string{tid},
		WithCustomer: true,
		WithPackage:  true,
	}

	results, err := h.domain.Subscription().FindByFilter(ctx, filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{Success: false, Error: stacktrace.RootCause(err).Error()})
		return nil
	}

	c.JSON(http.StatusOK, model.Response{Success: true, Data: results})
	return nil
}

func (h *subscriptionAdapter) Create(a any) error {
	c := a.(*gin.Context)
	ctx := activity.NewContext("http_subscription_create")

	tenantID, _ := c.Get("tenant_id")
	tid, _ := tenantID.(string)

	var payload model.SubscriptionInput
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, model.Response{Success: false, Error: err.Error()})
		return nil
	}
	payload.TenantID = tid

	result, err := h.domain.Subscription().Create(ctx, payload)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{Success: false, Error: stacktrace.RootCause(err).Error()})
		return nil
	}

	c.JSON(http.StatusCreated, model.Response{Success: true, Data: result})
	return nil
}

func (h *subscriptionAdapter) Get(a any) error {
	c := a.(*gin.Context)
	ctx := activity.NewContext("http_subscription_get")

	id := c.Param("id")

	result, err := h.domain.Subscription().FindByID(ctx, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{Success: false, Error: stacktrace.RootCause(err).Error()})
		return nil
	}

	c.JSON(http.StatusOK, model.Response{Success: true, Data: result})
	return nil
}

func (h *subscriptionAdapter) Update(a any) error {
	c := a.(*gin.Context)
	ctx := activity.NewContext("http_subscription_update")

	id := c.Param("id")

	var payload model.SubscriptionInput
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, model.Response{Success: false, Error: err.Error()})
		return nil
	}

	err := h.domain.Subscription().Update(ctx, id, payload)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{Success: false, Error: stacktrace.RootCause(err).Error()})
		return nil
	}

	c.JSON(http.StatusOK, model.Response{Success: true})
	return nil
}

func (h *subscriptionAdapter) Suspend(a any) error {
	c := a.(*gin.Context)
	ctx := activity.NewContext("http_subscription_suspend")

	id := c.Param("id")

	err := h.domain.Subscription().Suspend(ctx, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{Success: false, Error: stacktrace.RootCause(err).Error()})
		return nil
	}

	c.JSON(http.StatusOK, model.Response{Success: true})
	return nil
}

func (h *subscriptionAdapter) Activate(a any) error {
	c := a.(*gin.Context)
	ctx := activity.NewContext("http_subscription_activate")

	id := c.Param("id")

	err := h.domain.Subscription().Activate(ctx, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{Success: false, Error: stacktrace.RootCause(err).Error()})
		return nil
	}

	c.JSON(http.StatusOK, model.Response{Success: true})
	return nil
}

func (h *subscriptionAdapter) Cancel(a any) error {
	c := a.(*gin.Context)
	ctx := activity.NewContext("http_subscription_cancel")

	id := c.Param("id")

	err := h.domain.Subscription().Cancel(ctx, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{Success: false, Error: stacktrace.RootCause(err).Error()})
		return nil
	}

	c.JSON(http.StatusOK, model.Response{Success: true})
	return nil
}

func (h *subscriptionAdapter) SetVacation(a any) error {
	c := a.(*gin.Context)
	ctx := activity.NewContext("http_subscription_vacation")

	id := c.Param("id")

	var payload struct {
		Start time.Time `json:"start" binding:"required"`
		End   time.Time `json:"end" binding:"required"`
	}
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, model.Response{Success: false, Error: err.Error()})
		return nil
	}

	err := h.domain.Subscription().SetVacation(ctx, id, payload.Start, payload.End)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{Success: false, Error: stacktrace.RootCause(err).Error()})
		return nil
	}

	c.JSON(http.StatusOK, model.Response{Success: true})
	return nil
}
