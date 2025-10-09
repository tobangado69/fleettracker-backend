package analytics

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tobangado69/fleettracker-pro/backend/internal/common/middleware"
)

// GetDashboard godoc
// @Summary Get fleet operations dashboard
// @Description Get real-time fleet operations dashboard data
// @Tags analytics
// @Produce json
// @Success 200 {object} SuccessResponse
// @Failure 401 {object} ErrorResponse
// @Router /api/v1/analytics/dashboard [get]
// @Security BearerAuth
func (h *Handler) GetDashboard(c *gin.Context) {
	// Get company ID from JWT claims
	companyID, exists := c.Get("company_id")
	if !exists {
		middleware.AbortWithUnauthorized(c, "company ID not found in token")
		return
	}

	dashboard, err := h.service.GetFleetDashboard(c.Request.Context(), companyID.(string))
	if err != nil {
		middleware.AbortWithInternal(c, "Failed to get dashboard data", err)
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Success: true,
		Data:    dashboard,
	})
}

// GetRealTimeDashboard godoc
// @Summary Get real-time dashboard updates
// @Description Get real-time fleet operations dashboard data
// @Tags analytics
// @Produce json
// @Success 200 {object} SuccessResponse
// @Failure 401 {object} ErrorResponse
// @Router /api/v1/analytics/dashboard/realtime [get]
// @Security BearerAuth
func (h *Handler) GetRealTimeDashboard(c *gin.Context) {
	// Get company ID from JWT claims
	companyID, exists := c.Get("company_id")
	if !exists {
		middleware.AbortWithUnauthorized(c, "company ID not found in token")
		return
	}

	dashboard, err := h.service.GetFleetDashboard(c.Request.Context(), companyID.(string))
	if err != nil {
		middleware.AbortWithInternal(c, "Failed to get real-time dashboard data", err)
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Success: true,
		Data:    dashboard,
	})
}

