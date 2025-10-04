package payment

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// SuccessResponse represents a success response
type SuccessResponse struct {
	Success bool        `json:"success" example:"true"`
	Data    interface{} `json:"data,omitempty"`
	Message string      `json:"message,omitempty" example:"Operation successful"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error   string `json:"error" example:"Bad Request"`
	Message string `json:"message" example:"Invalid request data"`
	Details string `json:"details,omitempty" example:"Validation failed"`
}

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// GenerateInvoice generates a new invoice for manual bank transfer
// @Summary Generate invoice for manual bank transfer
// @Description Generate a new invoice with Indonesian PPN 11% tax calculation for manual bank transfer payment
// @Tags payments
// @Accept json
// @Produce json
// @Param request body InvoiceRequest true "Invoice generation data"
// @Success 201 {object} SuccessResponse{data=InvoiceResponse}
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/payments/invoices [post]
// @Security BearerAuth
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
// @Summary Confirm manual bank transfer payment
// @Description Confirm payment for an invoice via manual bank transfer with Indonesian banking system
// @Tags payments
// @Accept json
// @Produce json
// @Param request body PaymentConfirmationRequest true "Payment confirmation data"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/payments/invoices/{id}/confirm [post]
// @Security BearerAuth
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
// @Summary Get company invoices
// @Description Retrieve invoices for the authenticated company with filtering and pagination
// @Tags payments
// @Produce json
// @Param status query string false "Invoice status (pending, paid, overdue, cancelled)"
// @Param limit query int false "Number of invoices to return (default: 20)"
// @Param offset query int false "Number of invoices to skip (default: 0)"
// @Success 200 {object} SuccessResponse{data=[]InvoiceResponse}
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/payments/invoices [get]
// @Security BearerAuth
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
// @Summary Get payment instructions
// @Description Get detailed payment instructions for manual bank transfer with Indonesian banking details
// @Tags payments
// @Produce json
// @Param id path string true "Invoice ID"
// @Success 200 {object} SuccessResponse{data=PaymentInstructions}
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /api/v1/payments/invoices/{id}/instructions [get]
// @Security BearerAuth
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
// @Summary Generate subscription billing
// @Description Generate automatic billing for company subscription with Indonesian PPN 11% tax calculation
// @Tags payments
// @Accept json
// @Produce json
// @Param request body SubscriptionBillingRequest true "Subscription billing data"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/payments/subscriptions/billing [post]
// @Security BearerAuth
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
// CreateQRISPayment handles QRIS payment creation (not implemented)
// @Summary Create QRIS payment (Not Implemented)
// @Description Create QRIS payment - not supported, using manual bank transfer instead
// @Tags payments
// @Accept json
// @Produce json
// @Success 501 {object} ErrorResponse
// @Router /api/v1/payments/qris [post]
// @Security BearerAuth
func (h *Handler) CreateQRISPayment(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"message": "QRIS payment not supported - using manual bank transfer"})
}

// CreateBankTransfer handles direct bank transfer (not implemented)
// @Summary Create bank transfer payment (Not Implemented)
// @Description Create direct bank transfer - not implemented, use invoice generation and manual confirmation
// @Tags payments
// @Accept json
// @Produce json
// @Success 501 {object} ErrorResponse
// @Router /api/v1/payments/bank-transfer [post]
// @Security BearerAuth
func (h *Handler) CreateBankTransfer(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"message": "Direct bank transfer not implemented - use invoice generation and manual confirmation"})
}

// CreateEWalletPayment handles e-wallet payment (not implemented)
// @Summary Create e-wallet payment (Not Implemented)
// @Description Create e-wallet payment - not supported, using manual bank transfer instead
// @Tags payments
// @Accept json
// @Produce json
// @Success 501 {object} ErrorResponse
// @Router /api/v1/payments/e-wallet [post]
// @Security BearerAuth
func (h *Handler) CreateEWalletPayment(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"message": "E-wallet payment not supported - using manual bank transfer"})
}

// GetSubscriptions handles subscription listing (not implemented)
// @Summary Get subscriptions (Not Implemented)
// @Description Get company subscriptions - not implemented yet
// @Tags payments
// @Produce json
// @Success 501 {object} ErrorResponse
// @Router /api/v1/payments/subscriptions [get]
// @Security BearerAuth
func (h *Handler) GetSubscriptions(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"message": "Subscription listing not implemented yet"})
}

// CreateSubscription handles subscription creation (not implemented)
// @Summary Create subscription (Not Implemented)
// @Description Create new subscription - not implemented yet
// @Tags payments
// @Accept json
// @Produce json
// @Success 501 {object} ErrorResponse
// @Router /api/v1/payments/subscriptions [post]
// @Security BearerAuth
func (h *Handler) CreateSubscription(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"message": "Subscription creation not implemented yet"})
}
