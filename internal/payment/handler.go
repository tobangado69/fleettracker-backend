package payment

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// GenerateInvoice generates a new invoice for manual bank transfer
func (h *Handler) GenerateInvoice(c *gin.Context) {
	var req InvoiceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data", "details": err.Error()})
		return
	}

	// Get company ID from authenticated user context
	companyID, exists := c.Get("company_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Company ID not found in context"})
		return
	}

	req.CompanyID = companyID.(string)

	response, err := h.service.GenerateInvoice(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate invoice", "details": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data":    response,
	})
}

// ConfirmPayment confirms a manual bank transfer payment
func (h *Handler) ConfirmPayment(c *gin.Context) {
	var req PaymentConfirmationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data", "details": err.Error()})
		return
	}

	// Get user ID from authenticated user context
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found in context"})
		return
	}

	err := h.service.ConfirmPayment(c.Request.Context(), &req, userID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to confirm payment", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Payment confirmed successfully",
	})
}

// GetInvoices retrieves invoices for the authenticated company
func (h *Handler) GetInvoices(c *gin.Context) {
	// Get company ID from authenticated user context
	companyID, exists := c.Get("company_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Company ID not found in context"})
		return
	}

	// Get query parameters
	status := c.Query("status")
	limitStr := c.DefaultQuery("limit", "20")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		limit = 20
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil {
		offset = 0
	}

	invoices, err := h.service.GetInvoices(c.Request.Context(), companyID.(string), status, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve invoices", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    invoices,
	})
}

// GetPaymentInstructions generates payment instructions for an invoice
func (h *Handler) GetPaymentInstructions(c *gin.Context) {
	invoiceID := c.Param("id")
	if invoiceID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invoice ID is required"})
		return
	}

	instructions, err := h.service.GetPaymentInstructions(c.Request.Context(), invoiceID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get payment instructions", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    instructions,
	})
}

// GenerateSubscriptionBilling creates automatic billing for subscriptions
func (h *Handler) GenerateSubscriptionBilling(c *gin.Context) {
	var req SubscriptionBillingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data", "details": err.Error()})
		return
	}

	// Get company ID from authenticated user context
	companyID, exists := c.Get("company_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Company ID not found in context"})
		return
	}

	req.CompanyID = companyID.(string)

	err := h.service.GenerateSubscriptionBilling(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate subscription billing", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Subscription billing generated successfully",
	})
}

// Legacy endpoints for backward compatibility (not implemented for manual bank transfer)
func (h *Handler) CreateQRISPayment(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"message": "QRIS payment not supported - using manual bank transfer"})
}

func (h *Handler) CreateBankTransfer(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"message": "Direct bank transfer not implemented - use invoice generation and manual confirmation"})
}

func (h *Handler) CreateEWalletPayment(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"message": "E-wallet payment not supported - using manual bank transfer"})
}

func (h *Handler) GetSubscriptions(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"message": "Subscription listing not implemented yet"})
}

func (h *Handler) CreateSubscription(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"message": "Subscription creation not implemented yet"})
}
