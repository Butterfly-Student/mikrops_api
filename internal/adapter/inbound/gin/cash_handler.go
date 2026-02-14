package gin_inbound_adapter

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"go-template/internal/domain"
	"go-template/internal/model"
	inbound_port "go-template/internal/port/inbound"
)

type CashHandler struct {
	domain domain.Domain
}

func NewCashHandler(
	domain domain.Domain,
) inbound_port.CashHttpPort {
	return &CashHandler{
		domain: domain,
	}
}

func (h *CashHandler) CreateCashCategory(c *gin.Context) {
	var input model.CashCategoryInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	category, err := h.domain.Cash().CreateCategory(c.Request.Context(), input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, category)
}

func (h *CashHandler) GetCashCategory(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "category ID is required"})
		return
	}

	category, err := h.domain.Cash().GetCategory(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, category)
}

func (h *CashHandler) ListCashCategories(c *gin.Context) {
	filter := model.CashCategoryFilter{}

	if typeFilter := c.Query("type"); typeFilter != "" {
		filter.Type = &typeFilter
	}
	if isActive := c.Query("is_active"); isActive != "" {
		isActiveBool := isActive == "true"
		filter.IsActive = &isActiveBool
	}
	if isSystem := c.Query("is_system"); isSystem != "" {
		isSystemBool := isSystem == "true"
		filter.IsSystem = &isSystemBool
	}
	if search := c.Query("search"); search != "" {
		filter.Search = &search
	}

	categories, err := h.domain.Cash().ListCategories(c.Request.Context(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, categories)
}

func (h *CashHandler) UpdateCashCategory(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "category ID is required"})
		return
	}

	var input model.CashCategoryInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	category, err := h.domain.Cash().UpdateCategory(c.Request.Context(), id, input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, category)
}

func (h *CashHandler) DeleteCashCategory(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "category ID is required"})
		return
	}

	err := h.domain.Cash().DeleteCategory(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "category deleted successfully"})
}

func (h *CashHandler) CreateCashTransaction(c *gin.Context) {
	var input model.CashTransactionInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	transaction, err := h.domain.Cash().CreateTransaction(c.Request.Context(), input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, transaction)
}

func (h *CashHandler) GetCashTransaction(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "transaction ID is required"})
		return
	}

	transaction, err := h.domain.Cash().GetTransaction(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, transaction)
}

func (h *CashHandler) ListCashTransactions(c *gin.Context) {
	filter := model.CashTransactionFilter{}

	if typeFilter := c.Query("type"); typeFilter != "" {
		filter.Type = &typeFilter
	}
	if approvalStatus := c.Query("approval_status"); approvalStatus != "" {
		filter.ApprovalStatus = &approvalStatus
	}
	if requiresApproval := c.Query("requires_approval"); requiresApproval != "" {
		requiresApprovalBool := requiresApproval == "true"
		filter.RequiresApproval = &requiresApprovalBool
	}
	if startDate := c.Query("start_date"); startDate != "" {
		t, err := time.Parse(time.RFC3339, startDate)
		if err == nil {
			filter.DateStart = &t
		}
	}
	if endDate := c.Query("end_date"); endDate != "" {
		t, err := time.Parse(time.RFC3339, endDate)
		if err == nil {
			filter.DateEnd = &t
		}
	}
	if search := c.Query("search"); search != "" {
		filter.Search = &search
	}

	transactions, err := h.domain.Cash().ListTransactions(c.Request.Context(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, transactions)
}

func (h *CashHandler) UpdateCashTransaction(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "transaction ID is required"})
		return
	}

	var input model.CashTransactionInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	transaction, err := h.domain.Cash().UpdateTransaction(c.Request.Context(), id, input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, transaction)
}

func (h *CashHandler) DeleteCashTransaction(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "transaction ID is required"})
		return
	}

	err := h.domain.Cash().DeleteTransaction(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "transaction deleted successfully"})
}

func (h *CashHandler) ApproveCashTransaction(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "transaction ID is required"})
		return
	}

	err := h.domain.Cash().ApproveTransaction(c.Request.Context(), id, "")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "transaction approved successfully"})
}

func (h *CashHandler) RejectCashTransaction(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "transaction ID is required"})
		return
	}

	var input struct {
		Reason string `json:"reason"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.domain.Cash().RejectTransaction(c.Request.Context(), id, "", input.Reason)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "transaction rejected successfully"})
}

func (h *CashHandler) GetCashBalance(c *gin.Context) {
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")

	start := time.Now().AddDate(0, 0, -30)
	if startDate != "" {
		t, err := time.Parse(time.RFC3339, startDate)
		if err == nil {
			start = t
		}
	}

	end := time.Now()
	if endDate != "" {
		t, err := time.Parse(time.RFC3339, endDate)
		if err == nil {
			end = t
		}
	}

	balance, err := h.domain.Cash().GetBalance(c.Request.Context(), start, end)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"start_date": start,
		"end_date":   end,
		"balance":    balance,
	})
}
