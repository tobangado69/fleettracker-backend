package analytics

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tobangado69/fleettracker-pro/backend/internal/common/middleware"
)

// GetFleetUtilization godoc
// @Summary Get fleet utilization metrics
// @Description Get fleet utilization metrics and statistics
// @Tags analytics
// @Produce json
// @Success 200 {object} SuccessResponse
// @Failure 401 {object} ErrorResponse
// @Router /api/v1/analytics/fleet/utilization [get]
// @Security BearerAuth
func (h *Handler) GetFleetUtilization(c *gin.Context) {
	// Get company ID from JWT claims
	companyID, exists := c.Get("company_id")
	if !exists {
		middleware.AbortWithUnauthorized(c, "company ID not found in token")
		return
	}

	dashboard, err := h.service.GetFleetDashboard(c.Request.Context(), companyID.(string))
	if err != nil {
		middleware.AbortWithInternal(c, "Failed to get fleet utilization data", err)
		return
	}

	// Extract utilization-specific data
	utilizationData := map[string]interface{}{
		"active_vehicles":   dashboard.ActiveVehicles,
		"total_trips":      dashboard.TotalTrips,
		"utilization_rate": dashboard.UtilizationRate,
		"distance_traveled": dashboard.DistanceTraveled,
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Success: true,
		Data:    utilizationData,
	})
}

// GetFleetCosts godoc
// @Summary Get fleet cost analysis
// @Description Get fleet cost analysis and financial metrics
// @Tags analytics
// @Produce json
// @Success 200 {object} SuccessResponse
// @Failure 401 {object} ErrorResponse
// @Router /api/v1/analytics/fleet/costs [get]
// @Security BearerAuth
func (h *Handler) GetFleetCosts(c *gin.Context) {
	// Get company ID from JWT claims
	companyID, exists := c.Get("company_id")
	if !exists {
		middleware.AbortWithUnauthorized(c, "company ID not found in token")
		return
	}

	dashboard, err := h.service.GetFleetDashboard(c.Request.Context(), companyID.(string))
	if err != nil {
		middleware.AbortWithInternal(c, "Failed to get fleet cost data", err)
		return
	}

	// Extract cost-specific data
	costData := map[string]interface{}{
		"cost_per_km":     dashboard.CostPerKm,
		"fuel_consumed":   dashboard.FuelConsumed,
		"total_distance":  dashboard.DistanceTraveled,
		"fuel_cost_idr":   dashboard.FuelConsumed * 15000, // IDR cost
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Success: true,
		Data:    costData,
	})
}

// GetMaintenanceInsights godoc
// @Summary Get maintenance insights
// @Description Get maintenance scheduling insights and alerts
// @Tags analytics
// @Produce json
// @Success 200 {object} SuccessResponse
// @Failure 401 {object} ErrorResponse
// @Router /api/v1/analytics/fleet/maintenance [get]
// @Security BearerAuth
func (h *Handler) GetMaintenanceInsights(c *gin.Context) {
	// Get company ID from JWT claims
	companyID, exists := c.Get("company_id")
	if !exists {
		middleware.AbortWithUnauthorized(c, "company ID not found in token")
		return
	}

	dashboard, err := h.service.GetFleetDashboard(c.Request.Context(), companyID.(string))
	if err != nil {
		middleware.AbortWithInternal(c, "Failed to get maintenance insights", err)
		return
	}

	// Extract maintenance-specific data
	maintenanceData := map[string]interface{}{
		"active_vehicles":     dashboard.ActiveVehicles,
		"maintenance_alerts":  dashboard.MaintenanceAlerts,
		"total_alerts":       len(dashboard.MaintenanceAlerts),
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Success: true,
		Data:    maintenanceData,
	})
}

