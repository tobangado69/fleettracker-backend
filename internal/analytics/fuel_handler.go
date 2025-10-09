package analytics

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/tobangado69/fleettracker-pro/backend/internal/common/middleware"
)

// GetFuelConsumption godoc
// @Summary Get fuel consumption analytics
// @Description Get fuel consumption analytics and reports
// @Tags analytics
// @Produce json
// @Param start_date query string false "Start date (YYYY-MM-DD)"
// @Param end_date query string false "End date (YYYY-MM-DD)"
// @Success 200 {object} SuccessResponse
// @Failure 401 {object} ErrorResponse
// @Router /api/v1/analytics/fuel/consumption [get]
// @Security BearerAuth
func (h *Handler) GetFuelConsumption(c *gin.Context) {
	// Get company ID from JWT claims
	companyID, exists := c.Get("company_id")
	if !exists {
		middleware.AbortWithUnauthorized(c, "company ID not found in token")
		return
	}

	// Parse date parameters
	startDateStr := c.DefaultQuery("start_date", time.Now().AddDate(0, 0, -30).Format("2006-01-02"))
	endDateStr := c.DefaultQuery("end_date", time.Now().Format("2006-01-02"))

	startDate, err := time.Parse("2006-01-02", startDateStr)
	if err != nil {
		middleware.AbortWithBadRequest(c, "Invalid start date format")
		return
	}

	endDate, err := time.Parse("2006-01-02", endDateStr)
	if err != nil {
		middleware.AbortWithBadRequest(c, "Invalid end date format")
		return
	}

	analytics, err := h.service.GetFuelConsumption(c.Request.Context(), companyID.(string), startDate, endDate)
	if err != nil {
		middleware.AbortWithInternal(c, "Failed to get fuel consumption analytics", err)
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Success: true,
		Data:    analytics,
	})
}

// GetFuelEfficiency godoc
// @Summary Get fuel efficiency metrics
// @Description Get fuel efficiency metrics and optimization recommendations
// @Tags analytics
// @Produce json
// @Param start_date query string false "Start date (YYYY-MM-DD)"
// @Param end_date query string false "End date (YYYY-MM-DD)"
// @Success 200 {object} SuccessResponse
// @Failure 401 {object} ErrorResponse
// @Router /api/v1/analytics/fuel/efficiency [get]
// @Security BearerAuth
func (h *Handler) GetFuelEfficiency(c *gin.Context) {
	// Get company ID from JWT claims
	companyID, exists := c.Get("company_id")
	if !exists {
		middleware.AbortWithUnauthorized(c, "company ID not found in token")
		return
	}

	// Parse date parameters
	startDateStr := c.DefaultQuery("start_date", time.Now().AddDate(0, 0, -30).Format("2006-01-02"))
	endDateStr := c.DefaultQuery("end_date", time.Now().Format("2006-01-02"))

	startDate, err := time.Parse("2006-01-02", startDateStr)
	if err != nil {
		middleware.AbortWithBadRequest(c, "Invalid start date format")
		return
	}

	endDate, err := time.Parse("2006-01-02", endDateStr)
	if err != nil {
		middleware.AbortWithBadRequest(c, "Invalid end date format")
		return
	}

	analytics, err := h.service.GetFuelConsumption(c.Request.Context(), companyID.(string), startDate, endDate)
	if err != nil {
		middleware.AbortWithInternal(c, "Failed to get fuel efficiency metrics", err)
		return
	}

	// Extract efficiency-specific data
	efficiencyData := map[string]interface{}{
		"average_efficiency": analytics.AverageEfficiency,
		"optimization_tips":  analytics.OptimizationTips,
		"cost_savings":       analytics.CostSavings,
		"trends":            analytics.Trends,
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Success: true,
		Data:    efficiencyData,
	})
}

// GetFuelTheftAlerts godoc
// @Summary Get fuel theft alerts
// @Description Get fuel theft detection alerts and suspicious activities
// @Tags analytics
// @Produce json
// @Param start_date query string false "Start date (YYYY-MM-DD)"
// @Param end_date query string false "End date (YYYY-MM-DD)"
// @Success 200 {object} SuccessResponse
// @Failure 401 {object} ErrorResponse
// @Router /api/v1/analytics/fuel/theft [get]
// @Security BearerAuth
func (h *Handler) GetFuelTheftAlerts(c *gin.Context) {
	// Get company ID from JWT claims
	companyID, exists := c.Get("company_id")
	if !exists {
		middleware.AbortWithUnauthorized(c, "company ID not found in token")
		return
	}

	// Parse date parameters
	startDateStr := c.DefaultQuery("start_date", time.Now().AddDate(0, 0, -30).Format("2006-01-02"))
	endDateStr := c.DefaultQuery("end_date", time.Now().Format("2006-01-02"))

	startDate, err := time.Parse("2006-01-02", startDateStr)
	if err != nil {
		middleware.AbortWithBadRequest(c, "Invalid start date format")
		return
	}

	endDate, err := time.Parse("2006-01-02", endDateStr)
	if err != nil {
		middleware.AbortWithBadRequest(c, "Invalid end date format")
		return
	}

	analytics, err := h.service.GetFuelConsumption(c.Request.Context(), companyID.(string), startDate, endDate)
	if err != nil {
		middleware.AbortWithInternal(c, "Failed to get fuel theft alerts", err)
		return
	}

	// Extract theft-specific data
	theftData := map[string]interface{}{
		"alerts": analytics.TheftAlerts,
		"count":  len(analytics.TheftAlerts),
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Success: true,
		Data:    theftData,
	})
}

// GetFuelOptimization godoc
// @Summary Get fuel optimization recommendations
// @Description Get fuel optimization recommendations and cost savings tips
// @Tags analytics
// @Produce json
// @Param start_date query string false "Start date (YYYY-MM-DD)"
// @Param end_date query string false "End date (YYYY-MM-DD)"
// @Success 200 {object} SuccessResponse
// @Failure 401 {object} ErrorResponse
// @Router /api/v1/analytics/fuel/optimization [get]
// @Security BearerAuth
func (h *Handler) GetFuelOptimization(c *gin.Context) {
	// Get company ID from JWT claims
	companyID, exists := c.Get("company_id")
	if !exists {
		middleware.AbortWithUnauthorized(c, "company ID not found in token")
		return
	}

	// Parse date parameters
	startDateStr := c.DefaultQuery("start_date", time.Now().AddDate(0, 0, -30).Format("2006-01-02"))
	endDateStr := c.DefaultQuery("end_date", time.Now().Format("2006-01-02"))

	startDate, err := time.Parse("2006-01-02", startDateStr)
	if err != nil {
		middleware.AbortWithBadRequest(c, "Invalid start date format")
		return
	}

	endDate, err := time.Parse("2006-01-02", endDateStr)
	if err != nil {
		middleware.AbortWithBadRequest(c, "Invalid end date format")
		return
	}

	analytics, err := h.service.GetFuelConsumption(c.Request.Context(), companyID.(string), startDate, endDate)
	if err != nil {
		middleware.AbortWithInternal(c, "Failed to get fuel optimization recommendations", err)
		return
	}

	// Extract optimization-specific data
	optimizationData := map[string]interface{}{
		"recommendations": analytics.OptimizationTips,
		"potential_savings": analytics.CostSavings,
		"efficiency_score":  analytics.AverageEfficiency,
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Success: true,
		Data:    optimizationData,
	})
}

