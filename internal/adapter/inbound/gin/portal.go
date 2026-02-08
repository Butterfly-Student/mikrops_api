package gin_inbound_adapter

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/palantir/stacktrace"
	"golang.org/x/crypto/bcrypt"

	"mikrops/internal/domain"
	"mikrops/internal/model"
	inbound_port "mikrops/internal/port/inbound"
	"mikrops/utils/activity"
)

type portalAdapter struct {
	domain domain.Domain
}

func NewPortalAdapter(domain domain.Domain) inbound_port.PortalHttpPort {
	return &portalAdapter{domain: domain}
}

func (h *portalAdapter) Dashboard(a any) error {
	c := a.(*gin.Context)
	ctx := activity.NewContext("http_portal_dashboard")

	customerID, _ := c.Get("customer_id")
	cid, _ := customerID.(string)
	tenantID, _ := c.Get("tenant_id")
	tid, _ := tenantID.(string)

	// Get customer info
	customer, err := h.domain.Customer().FindByID(ctx, cid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{Success: false, Error: stacktrace.RootCause(err).Error()})
		return nil
	}

	// Get active subscriptions
	subscriptions, err := h.domain.Subscription().FindByFilter(ctx, model.SubscriptionFilter{
		CustomerIDs:  []string{cid},
		TenantIDs:    []string{tid},
		Statuses:     []string{model.SubscriptionStatusActive},
		WithPackage:  true,
		WithCustomer: true,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{Success: false, Error: stacktrace.RootCause(err).Error()})
		return nil
	}

	// Get unpaid invoices
	invoices, err := h.domain.Invoice().FindByFilter(ctx, model.InvoiceFilter{
		CustomerIDs: []string{cid},
		TenantIDs:   []string{tid},
		Statuses:    []string{"unpaid", "overdue"},
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{Success: false, Error: stacktrace.RootCause(err).Error()})
		return nil
	}

	c.JSON(http.StatusOK, model.Response{
		Success: true,
		Data: gin.H{
			"customer":      customer,
			"subscriptions": subscriptions,
			"invoices":      invoices,
		},
	})
	return nil
}

func (h *portalAdapter) GetProfile(a any) error {
	c := a.(*gin.Context)
	ctx := activity.NewContext("http_portal_profile")

	customerID, _ := c.Get("customer_id")
	cid, _ := customerID.(string)

	result, err := h.domain.Customer().FindByID(ctx, cid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{Success: false, Error: stacktrace.RootCause(err).Error()})
		return nil
	}

	c.JSON(http.StatusOK, model.Response{Success: true, Data: result})
	return nil
}

func (h *portalAdapter) UpdateProfile(a any) error {
	c := a.(*gin.Context)
	ctx := activity.NewContext("http_portal_update_profile")

	customerID, _ := c.Get("customer_id")
	cid, _ := customerID.(string)

	var payload struct {
		FullName string `json:"full_name"`
		Email    string `json:"email"`
		Phone    string `json:"phone"`
		Address  string `json:"address"`
	}
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, model.Response{Success: false, Error: err.Error()})
		return nil
	}

	input := model.CustomerInput{
		FullName: payload.FullName,
		Email:    payload.Email,
		Phone:    payload.Phone,
		Address:  payload.Address,
	}

	err := h.domain.Customer().Update(ctx, cid, input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{Success: false, Error: stacktrace.RootCause(err).Error()})
		return nil
	}

	c.JSON(http.StatusOK, model.Response{Success: true})
	return nil
}

func (h *portalAdapter) UpdatePassword(a any) error {
	c := a.(*gin.Context)
	ctx := activity.NewContext("http_portal_update_password")

	customerID, _ := c.Get("customer_id")
	cid, _ := customerID.(string)

	var payload struct {
		OldPassword string `json:"old_password" binding:"required"`
		NewPassword string `json:"new_password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, model.Response{Success: false, Error: err.Error()})
		return nil
	}

	// Verify old password
	customer, err := h.domain.Customer().FindByID(ctx, cid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{Success: false, Error: stacktrace.RootCause(err).Error()})
		return nil
	}

	if err := bcrypt.CompareHashAndPassword([]byte(customer.PasswordHash), []byte(payload.OldPassword)); err != nil {
		c.JSON(http.StatusBadRequest, model.Response{Success: false, Error: "invalid old password"})
		return nil
	}

	// Hash new password
	hash, err := bcrypt.GenerateFromPassword([]byte(payload.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{Success: false, Error: "failed to hash password"})
		return nil
	}

	input := model.CustomerInput{PasswordHash: string(hash)}
	err = h.domain.Customer().Update(ctx, cid, input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{Success: false, Error: stacktrace.RootCause(err).Error()})
		return nil
	}

	c.JSON(http.StatusOK, model.Response{Success: true})
	return nil
}

func (h *portalAdapter) GetSubscription(a any) error {
	c := a.(*gin.Context)
	ctx := activity.NewContext("http_portal_subscription")

	customerID, _ := c.Get("customer_id")
	cid, _ := customerID.(string)
	tenantID, _ := c.Get("tenant_id")
	tid, _ := tenantID.(string)

	results, err := h.domain.Subscription().FindByFilter(ctx, model.SubscriptionFilter{
		CustomerIDs: []string{cid},
		TenantIDs:   []string{tid},
		WithPackage: true,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{Success: false, Error: stacktrace.RootCause(err).Error()})
		return nil
	}

	c.JSON(http.StatusOK, model.Response{Success: true, Data: results})
	return nil
}

func (h *portalAdapter) GetConnectionStatus(a any) error {
	c := a.(*gin.Context)
	ctx := activity.NewContext("http_portal_connection_status")

	customerID, _ := c.Get("customer_id")
	cid, _ := customerID.(string)
	tenantID, _ := c.Get("tenant_id")
	tid, _ := tenantID.(string)

	// Get customer's active subscription to find the NAS
	subs, err := h.domain.Subscription().FindByFilter(ctx, model.SubscriptionFilter{
		CustomerIDs: []string{cid},
		TenantIDs:   []string{tid},
		Statuses:    []string{model.SubscriptionStatusActive},
	})
	if err != nil || len(subs) == 0 {
		c.JSON(http.StatusOK, model.Response{Success: true, Data: gin.H{"connected": false}})
		return nil
	}

	nasID := subs[0].NasID
	connections, err := h.domain.Mikrotik().GetActiveConnections(ctx, nasID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{Success: false, Error: stacktrace.RootCause(err).Error()})
		return nil
	}

	// Find customer's connection by PPPoE username
	customer, _ := h.domain.Customer().FindByID(ctx, cid)
	for _, conn := range connections {
		if conn.Name == customer.PppoeUsername {
			c.JSON(http.StatusOK, model.Response{Success: true, Data: gin.H{
				"connected": true,
				"connection": conn,
			}})
			return nil
		}
	}

	c.JSON(http.StatusOK, model.Response{Success: true, Data: gin.H{"connected": false}})
	return nil
}

func (h *portalAdapter) GetBandwidth(a any) error {
	c := a.(*gin.Context)
	ctx := activity.NewContext("http_portal_bandwidth")

	customerID, _ := c.Get("customer_id")
	cid, _ := customerID.(string)
	tenantID, _ := c.Get("tenant_id")
	tid, _ := tenantID.(string)

	subs, err := h.domain.Subscription().FindByFilter(ctx, model.SubscriptionFilter{
		CustomerIDs: []string{cid},
		TenantIDs:   []string{tid},
		Statuses:    []string{model.SubscriptionStatusActive},
	})
	if err != nil || len(subs) == 0 {
		c.JSON(http.StatusNotFound, model.Response{Success: false, Error: "no active subscription found"})
		return nil
	}

	customer, _ := h.domain.Customer().FindByID(ctx, cid)
	target := "<pppoe-" + customer.PppoeUsername + ">"

	bandwidth, err := h.domain.Mikrotik().GetBandwidth(ctx, subs[0].NasID, target)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{Success: false, Error: stacktrace.RootCause(err).Error()})
		return nil
	}

	c.JSON(http.StatusOK, model.Response{Success: true, Data: bandwidth})
	return nil
}

func (h *portalAdapter) ListInvoices(a any) error {
	c := a.(*gin.Context)
	ctx := activity.NewContext("http_portal_invoices")

	customerID, _ := c.Get("customer_id")
	cid, _ := customerID.(string)
	tenantID, _ := c.Get("tenant_id")
	tid, _ := tenantID.(string)

	results, err := h.domain.Invoice().FindByFilter(ctx, model.InvoiceFilter{
		CustomerIDs:      []string{cid},
		TenantIDs:        []string{tid},
		WithSubscription: true,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{Success: false, Error: stacktrace.RootCause(err).Error()})
		return nil
	}

	c.JSON(http.StatusOK, model.Response{Success: true, Data: results})
	return nil
}

func (h *portalAdapter) GetInvoice(a any) error {
	c := a.(*gin.Context)
	ctx := activity.NewContext("http_portal_invoice_get")

	id := c.Param("id")

	result, err := h.domain.Invoice().FindByID(ctx, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{Success: false, Error: stacktrace.RootCause(err).Error()})
		return nil
	}

	c.JSON(http.StatusOK, model.Response{Success: true, Data: result})
	return nil
}

func (h *portalAdapter) ListPayments(a any) error {
	c := a.(*gin.Context)
	ctx := activity.NewContext("http_portal_payments")

	tenantID, _ := c.Get("tenant_id")
	tid, _ := tenantID.(string)

	// Get customer's invoices first, then find payments
	customerID, _ := c.Get("customer_id")
	cid, _ := customerID.(string)

	invoices, err := h.domain.Invoice().FindByFilter(ctx, model.InvoiceFilter{
		CustomerIDs: []string{cid},
		TenantIDs:   []string{tid},
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{Success: false, Error: stacktrace.RootCause(err).Error()})
		return nil
	}

	var invoiceIDs []string
	for _, inv := range invoices {
		invoiceIDs = append(invoiceIDs, inv.ID)
	}

	if len(invoiceIDs) == 0 {
		c.JSON(http.StatusOK, model.Response{Success: true, Data: []model.Payment{}})
		return nil
	}

	results, err := h.domain.Payment().FindByFilter(ctx, model.PaymentFilter{
		InvoiceIDs:        invoiceIDs,
		TenantIDs:         []string{tid},
		WithInvoice:       true,
		WithPaymentMethod: true,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{Success: false, Error: stacktrace.RootCause(err).Error()})
		return nil
	}

	c.JSON(http.StatusOK, model.Response{Success: true, Data: results})
	return nil
}

func (h *portalAdapter) CreatePayment(a any) error {
	c := a.(*gin.Context)
	ctx := activity.NewContext("http_portal_payment_create")

	tenantID, _ := c.Get("tenant_id")
	tid, _ := tenantID.(string)

	var payload model.PaymentInput
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, model.Response{Success: false, Error: err.Error()})
		return nil
	}
	payload.TenantID = tid
	payload.Status = "pending"

	result, err := h.domain.Payment().Create(ctx, payload)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{Success: false, Error: stacktrace.RootCause(err).Error()})
		return nil
	}

	c.JSON(http.StatusCreated, model.Response{Success: true, Data: result})
	return nil
}

func (h *portalAdapter) GetPayment(a any) error {
	c := a.(*gin.Context)
	ctx := activity.NewContext("http_portal_payment_get")

	id := c.Param("id")

	result, err := h.domain.Payment().FindByID(ctx, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{Success: false, Error: stacktrace.RootCause(err).Error()})
		return nil
	}

	c.JSON(http.StatusOK, model.Response{Success: true, Data: result})
	return nil
}

func (h *portalAdapter) UploadProof(a any) error {
	c := a.(*gin.Context)
	ctx := activity.NewContext("http_portal_upload_proof")

	id := c.Param("id")

	file, err := c.FormFile("proof")
	if err != nil {
		c.JSON(http.StatusBadRequest, model.Response{Success: false, Error: "proof file is required"})
		return nil
	}

	// Save file
	dst := "uploads/proofs/" + id + "_" + file.Filename
	if err := c.SaveUploadedFile(file, dst); err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{Success: false, Error: "failed to save file"})
		return nil
	}

	// Update payment with proof URL
	input := model.PaymentInput{ProofURL: dst}
	err = h.domain.Payment().Update(ctx, id, input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{Success: false, Error: stacktrace.RootCause(err).Error()})
		return nil
	}

	c.JSON(http.StatusOK, model.Response{Success: true, Data: gin.H{"proof_url": dst}})
	return nil
}

func (h *portalAdapter) ListPackages(a any) error {
	c := a.(*gin.Context)
	ctx := activity.NewContext("http_portal_packages")

	tenantID, _ := c.Get("tenant_id")
	tid, _ := tenantID.(string)

	active := true
	filter := model.InternetPackageFilter{
		TenantIDs: []string{tid},
		IsActive:  &active,
	}

	results, err := h.domain.InternetPackage().FindByFilter(ctx, filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{Success: false, Error: stacktrace.RootCause(err).Error()})
		return nil
	}

	c.JSON(http.StatusOK, model.Response{Success: true, Data: results})
	return nil
}

func (h *portalAdapter) ListPaymentMethods(a any) error {
	c := a.(*gin.Context)
	ctx := activity.NewContext("http_portal_payment_methods")

	tenantID, _ := c.Get("tenant_id")
	tid, _ := tenantID.(string)

	active := true
	filter := model.PaymentMethodFilter{
		TenantIDs: []string{tid},
		IsActive:  &active,
	}

	results, err := h.domain.PaymentMethod().FindByFilter(ctx, filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{Success: false, Error: stacktrace.RootCause(err).Error()})
		return nil
	}

	c.JSON(http.StatusOK, model.Response{Success: true, Data: results})
	return nil
}
