package analytics

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/tobangado69/fleettracker-pro/backend/internal/common/middleware"
)

// GetDriverPerformance godoc
// @Summary Get driver performance analytics
// @Description Get comprehensive driver performance analytics
// @Tags analytics
// @Produce json
// @Param driver_id query string false "Driver ID"
// @Param period query string false "Period (daily, weekly, monthly)"
// @Success 200 {object} SuccessResponse
// @Failure 401 {object} ErrorResponse
// @Router /api/v1/analytics/drivers/performance [get]
// @Security BearerAuth
func (h *Handler) GetDriverPerformance(c *gin.Context) {
	// Get company ID from JWT claims
	companyID, exists := c.Get("company_id")
	if !exists {
		middleware.AbortWithUnauthorized(c, "company ID not found in token")
		return
	}

	driverID := c.Query("driver_id")
	if driverID == "" {
		middleware.AbortWithBadRequest(c, "Driver ID is required")
		return
	}

	period := c.DefaultQuery("period", "monthly")

	performance, err := h.service.GetDriverPerformance(c.Request.Context(), companyID.(string), driverID, period)
	if err != nil {
		middleware.AbortWithInternal(c, "Failed to get driver performance analytics", err)
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Success: true,
		Data:    performance,
	})
}

// GetDriverRanking godoc
// @Summary Get driver ranking
// @Description Get driver performance ranking and leaderboard
// @Tags analytics
// @Produce json
// @Param limit query int false "Number of drivers to return"
// @Success 200 {object} SuccessResponse
// @Failure 401 {object} ErrorResponse
// @Router /api/v1/analytics/drivers/ranking [get]
// @Security BearerAuth
func (h *Handler) GetDriverRanking(c *gin.Context) {
	// Get company ID from JWT claims
	companyID, exists := c.Get("company_id")
	if !exists {
		middleware.AbortWithUnauthorized(c, "company ID not found in token")
		return
	}

	limitStr := c.DefaultQuery("limit", "10")
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		limit = 10
	}

	// Get dashboard data which includes top performers
	dashboard, err := h.service.GetFleetDashboard(c.Request.Context(), companyID.(string))
	if err != nil {
		middleware.AbortWithInternal(c, "Failed to get driver ranking", err)
		return
	}

	// Limit the number of top performers
	topPerformers := dashboard.TopPerformers
	if len(topPerformers) > limit {
		topPerformers = topPerformers[:limit]
	}

	rankingData := map[string]interface{}{
		"drivers": topPerformers,
		"total":   len(dashboard.TopPerformers),
		"limit":   limit,
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Success: true,
		Data:    rankingData,
	})
}

// GetDriverBehavior godoc
// @Summary Get driver behavior analysis
// @Description Get detailed driver behavior analysis and metrics
// @Tags analytics
// @Produce json
// @Param driver_id query string false "Driver ID"
// @Success 200 {object} SuccessResponse
// @Failure 401 {object} ErrorResponse
// @Router /api/v1/analytics/drivers/behavior [get]
// @Security BearerAuth
func (h *Handler) GetDriverBehavior(c *gin.Context) {
	// Get company ID from JWT claims
	companyID, exists := c.Get("company_id")
	if !exists {
		middleware.AbortWithUnauthorized(c, "company ID not found in token")
		return
	}

	driverID := c.Query("driver_id")
	if driverID == "" {
		middleware.AbortWithBadRequest(c, "Driver ID is required")
		return
	}

	performance, err := h.service.GetDriverPerformance(c.Request.Context(), companyID.(string), driverID, "monthly")
	if err != nil {
		middleware.AbortWithInternal(c, "Failed to get driver behavior analysis", err)
		return
	}

	// Extract behavior-specific data
	behaviorData := map[string]interface{}{
		"driver_id":         performance.DriverID,
		"driver_name":       performance.DriverName,
		"behavior_metrics":  performance.BehaviorMetrics,
		"improvement_areas": performance.ImprovementAreas,
		"trends":           performance.Trends,
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Success: true,
		Data:    behaviorData,
	})
}

// GetDriverRecommendations godoc
// @Summary Get driver training recommendations
// @Description Get personalized training recommendations for drivers
// @Tags analytics
// @Produce json
// @Param driver_id query string false "Driver ID"
// @Success 200 {object} SuccessResponse
// @Failure 401 {object} ErrorResponse
// @Router /api/v1/analytics/drivers/recommendations [get]
// @Security BearerAuth
func (h *Handler) GetDriverRecommendations(c *gin.Context) {
	// Get company ID from JWT claims
	companyID, exists := c.Get("company_id")
	if !exists {
		middleware.AbortWithUnauthorized(c, "company ID not found in token")
		return
	}

	driverID := c.Query("driver_id")
	if driverID == "" {
		middleware.AbortWithBadRequest(c, "Driver ID is required")
		return
	}

	performance, err := h.service.GetDriverPerformance(c.Request.Context(), companyID.(string), driverID, "monthly")
	if err != nil {
		middleware.AbortWithInternal(c, "Failed to get driver recommendations", err)
		return
	}

	// Extract recommendations
	recommendations := map[string]interface{}{
		"driver_id":          performance.DriverID,
		"driver_name":        performance.DriverName,
		"training_recommendations": performance.ImprovementAreas,
		"priority_areas":    performance.ImprovementAreas,
		"behavior_metrics":  performance.BehaviorMetrics,
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Success: true,
		Data:    recommendations,
	})
}

