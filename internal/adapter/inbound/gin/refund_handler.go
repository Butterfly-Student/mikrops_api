package gin_inbound_adapter

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go-template/internal/domain"
	"go-template/internal/model"
)

type RefundHttpHandler struct {
	domain domain.Domain
}

func NewRefundHttpHandler(domain domain.Domain) *RefundHttpHandler {
	return &RefundHttpHandler{
		domain: domain,
	}
}

type CreateRefundRequest struct {
	PaymentID         uuid.UUID  `json:"payment_id" validate:"required"`
	InvoiceID         *uuid.UUID `json:"invoice_id"`
	RefundAmount      *float64   `json:"refund_amount" validate:"required,gt=0"`
	RefundType        *string    `json:"refund_type" validate:"required,oneof=full partial"`
	RefundReason      *string    `json:"refund_reason" validate:"required"`
	RefundMethod      *string    `json:"refund_method" validate:"required,oneof=original bank_transfer cash"`
	BankName          *string    `json:"bank_name"`
	BankAccountName   *string    `json:"bank_account_name"`
	BankAccountNumber *string    `json:"bank_account_number"`
	Notes             *string    `json:"notes"`
}

type UpdateRefundRequest struct {
	RefundReason      *string `json:"refund_reason"`
	RefundMethod      *string `json:"refund_method" validate:"omitempty,oneof=original bank_transfer cash"`
	BankName          *string `json:"bank_name"`
	BankAccountName   *string `json:"bank_account_name"`
	BankAccountNumber *string `json:"bank_account_number"`
	Notes             *string `json:"notes"`
}

type ApproveRefundRequest struct {
	ApprovedBy string `json:"approved_by" validate:"required"`
}

type RejectRefundRequest struct {
	RejectedBy string `json:"rejected_by" validate:"required"`
	Reason     string `json:"reason" validate:"required"`
}

type ProcessRefundRequest struct {
	ProcessedBy string `json:"processed_by" validate:"required"`
}

type CompleteRefundRequest struct {
	ProcessedBy    string `json:"processed_by" validate:"required"`
	XenditRefundID string `json:"xendit_refund_id" validate:"required"`
}

func (h *RefundHttpHandler) CreateRefund(c *gin.Context) {
	var req CreateRefundRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	input := model.RefundInput{
		PaymentID:         req.PaymentID,
		InvoiceID:         req.InvoiceID,
		RefundAmount:      req.RefundAmount,
		RefundType:        req.RefundType,
		RefundReason:      req.RefundReason,
		RefundMethod:      req.RefundMethod,
		BankName:          req.BankName,
		BankAccountName:   req.BankAccountName,
		BankAccountNumber: req.BankAccountNumber,
		Notes:             req.Notes,
	}

	refund, err := h.domain.Refund().CreateRefund(c.Request.Context(), input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, refund)
}

func (h *RefundHttpHandler) GetRefund(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id is required"})
		return
	}

	refund, err := h.domain.Refund().GetRefund(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if refund == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "refund not found"})
		return
	}

	c.JSON(http.StatusOK, refund)
}

func (h *RefundHttpHandler) ListRefunds(c *gin.Context) {
	var filter model.RefundFilter

	if ids := c.QueryArray("ids"); len(ids) > 0 {
		filter.IDs = make([]uuid.UUID, len(ids))
		for i, id := range ids {
			filter.IDs[i], _ = uuid.Parse(id)
		}
	}

	if refundNumbers := c.QueryArray("refund_numbers"); len(refundNumbers) > 0 {
		filter.RefundNumbers = refundNumbers
	}

	if paymentIDs := c.QueryArray("payment_ids"); len(paymentIDs) > 0 {
		filter.PaymentIDs = make([]uuid.UUID, len(paymentIDs))
		for i, id := range paymentIDs {
			filter.PaymentIDs[i], _ = uuid.Parse(id)
		}
	}

	if invoiceIDs := c.QueryArray("invoice_ids"); len(invoiceIDs) > 0 {
		filter.InvoiceIDs = make([]uuid.UUID, len(invoiceIDs))
		for i, id := range invoiceIDs {
			filter.InvoiceIDs[i], _ = uuid.Parse(id)
		}
	}

	if customerID := c.Query("customer_id"); customerID != "" {
		parsedID, err := uuid.Parse(customerID)
		if err == nil {
			filter.CustomerID = &parsedID
		}
	}

	if status := c.QueryArray("status"); len(status) > 0 {
		filter.Status = status
	}

	if refundType := c.QueryArray("refund_type"); len(refundType) > 0 {
		filter.RefundType = refundType
	}

	if refundMethod := c.QueryArray("refund_method"); len(refundMethod) > 0 {
		filter.RefundMethod = refundMethod
	}

	if search := c.Query("search"); search != "" {
		filter.Search = &search
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	refunds, err := h.domain.Refund().ListRefunds(c.Request.Context(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"refunds": refunds,
		"page":    page,
		"limit":   limit,
		"total":   len(refunds),
	})
}

func (h *RefundHttpHandler) UpdateRefund(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id is required"})
		return
	}

	var req UpdateRefundRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	input := model.RefundInput{
		RefundReason:      req.RefundReason,
		RefundMethod:      req.RefundMethod,
		BankName:          req.BankName,
		BankAccountName:   req.BankAccountName,
		BankAccountNumber: req.BankAccountNumber,
		Notes:             req.Notes,
	}

	refund, err := h.domain.Refund().UpdateRefund(c.Request.Context(), id, input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if refund == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "refund not found"})
		return
	}

	c.JSON(http.StatusOK, refund)
}

func (h *RefundHttpHandler) DeleteRefund(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id is required"})
		return
	}

	if err := h.domain.Refund().DeleteRefund(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "refund deleted successfully"})
}

func (h *RefundHttpHandler) ApproveRefund(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id is required"})
		return
	}

	var req ApproveRefundRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	refund, err := h.domain.Refund().ApproveRefund(c.Request.Context(), id, req.ApprovedBy)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if refund == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "refund not found"})
		return
	}

	c.JSON(http.StatusOK, refund)
}

func (h *RefundHttpHandler) RejectRefund(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id is required"})
		return
	}

	var req RejectRefundRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	refund, err := h.domain.Refund().RejectRefund(c.Request.Context(), id, req.RejectedBy, req.Reason)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if refund == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "refund not found"})
		return
	}

	c.JSON(http.StatusOK, refund)
}

func (h *RefundHttpHandler) ProcessRefund(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id is required"})
		return
	}

	var req ProcessRefundRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	refund, err := h.domain.Refund().ProcessRefund(c.Request.Context(), id, req.ProcessedBy)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if refund == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "refund not found"})
		return
	}

	c.JSON(http.StatusOK, refund)
}

func (h *RefundHttpHandler) CompleteRefund(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id is required"})
		return
	}

	var req CompleteRefundRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	refund, err := h.domain.Refund().CompleteRefund(c.Request.Context(), id, req.ProcessedBy, req.XenditRefundID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if refund == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "refund not found"})
		return
	}

	c.JSON(http.StatusOK, refund)
}

func (h *RefundHttpHandler) GetPendingRefunds(c *gin.Context) {
	refunds, err := h.domain.Refund().GetPendingRefunds(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"refunds": refunds,
		"total":   len(refunds),
	})
}
