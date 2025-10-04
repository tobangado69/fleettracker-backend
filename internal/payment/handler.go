package payment

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) CreateQRISPayment(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "QRIS payment not implemented yet"})
}
func (h *Handler) CreateBankTransfer(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Bank transfer not implemented yet"})
}
func (h *Handler) CreateEWalletPayment(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "E-wallet payment not implemented yet"})
}
func (h *Handler) GetSubscriptions(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Subscriptions not implemented yet"})
}
func (h *Handler) CreateSubscription(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Subscription creation not implemented yet"})
}
func (h *Handler) GetInvoices(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Invoices not implemented yet"})
}
