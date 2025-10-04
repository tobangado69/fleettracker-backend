package analytics

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// Handler handles analytics HTTP requests
type Handler struct {
	service *Service
}

// NewHandler creates a new analytics handler
func NewHandler(service *Service) *Handler {
	return &Handler{
		service: service,
	}
}

// SuccessResponse represents a successful API response
type SuccessResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data"`
	Message string      `json:"message,omitempty"`
}

// ErrorResponse represents an error API response
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

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
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error:   "unauthorized",
			Message: "company ID not found in token",
		})
		return
	}

	dashboard, err := h.service.GetFleetDashboard(c.Request.Context(), companyID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to get dashboard data",
		})
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
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error:   "unauthorized",
			Message: "company ID not found in token",
		})
		return
	}

	dashboard, err := h.service.GetFleetDashboard(c.Request.Context(), companyID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to get real-time dashboard data",
		})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Success: true,
		Data:    dashboard,
	})
}

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
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error:   "unauthorized",
			Message: "company ID not found in token",
		})
		return
	}

	// Parse date parameters
	startDateStr := c.DefaultQuery("start_date", time.Now().AddDate(0, 0, -30).Format("2006-01-02"))
	endDateStr := c.DefaultQuery("end_date", time.Now().Format("2006-01-02"))

	startDate, err := time.Parse("2006-01-02", startDateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_date",
			Message: "Invalid start date format",
		})
		return
	}

	endDate, err := time.Parse("2006-01-02", endDateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_date",
			Message: "Invalid end date format",
		})
		return
	}

	analytics, err := h.service.GetFuelConsumption(c.Request.Context(), companyID.(string), startDate, endDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to get fuel consumption analytics",
		})
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
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error:   "unauthorized",
			Message: "company ID not found in token",
		})
		return
	}

	// Parse date parameters
	startDateStr := c.DefaultQuery("start_date", time.Now().AddDate(0, 0, -30).Format("2006-01-02"))
	endDateStr := c.DefaultQuery("end_date", time.Now().Format("2006-01-02"))

	startDate, err := time.Parse("2006-01-02", startDateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_date",
			Message: "Invalid start date format",
		})
		return
	}

	endDate, err := time.Parse("2006-01-02", endDateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_date",
			Message: "Invalid end date format",
		})
		return
	}

	analytics, err := h.service.GetFuelConsumption(c.Request.Context(), companyID.(string), startDate, endDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to get fuel efficiency metrics",
		})
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
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error:   "unauthorized",
			Message: "company ID not found in token",
		})
		return
	}

	// Parse date parameters
	startDateStr := c.DefaultQuery("start_date", time.Now().AddDate(0, 0, -30).Format("2006-01-02"))
	endDateStr := c.DefaultQuery("end_date", time.Now().Format("2006-01-02"))

	startDate, err := time.Parse("2006-01-02", startDateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_date",
			Message: "Invalid start date format",
		})
		return
	}

	endDate, err := time.Parse("2006-01-02", endDateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_date",
			Message: "Invalid end date format",
		})
		return
	}

	analytics, err := h.service.GetFuelConsumption(c.Request.Context(), companyID.(string), startDate, endDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to get fuel theft alerts",
		})
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
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error:   "unauthorized",
			Message: "company ID not found in token",
		})
		return
	}

	// Parse date parameters
	startDateStr := c.DefaultQuery("start_date", time.Now().AddDate(0, 0, -30).Format("2006-01-02"))
	endDateStr := c.DefaultQuery("end_date", time.Now().Format("2006-01-02"))

	startDate, err := time.Parse("2006-01-02", startDateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_date",
			Message: "Invalid start date format",
		})
		return
	}

	endDate, err := time.Parse("2006-01-02", endDateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_date",
			Message: "Invalid end date format",
		})
		return
	}

	analytics, err := h.service.GetFuelConsumption(c.Request.Context(), companyID.(string), startDate, endDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to get fuel optimization recommendations",
		})
		return
	}

	// Extract optimization-specific data
	optimizationData := map[string]interface{}{
		"optimization_tips": analytics.OptimizationTips,
		"cost_savings":      analytics.CostSavings,
		"current_efficiency": analytics.AverageEfficiency,
		"potential_savings": analytics.CostSavings,
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Success: true,
		Data:    optimizationData,
	})
}

// GetDriverPerformance godoc
// @Summary Get driver performance analytics
// @Description Get driver performance analytics and scoring
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
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error:   "unauthorized",
			Message: "company ID not found in token",
		})
		return
	}

	driverID := c.Query("driver_id")
	if driverID == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "missing_parameter",
			Message: "Driver ID is required",
		})
		return
	}

	period := c.DefaultQuery("period", "monthly")

	performance, err := h.service.GetDriverPerformance(c.Request.Context(), companyID.(string), driverID, period)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to get driver performance analytics",
		})
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
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error:   "unauthorized",
			Message: "company ID not found in token",
		})
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
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to get driver ranking",
		})
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
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error:   "unauthorized",
			Message: "company ID not found in token",
		})
		return
	}

	driverID := c.Query("driver_id")
	if driverID == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "missing_parameter",
			Message: "Driver ID is required",
		})
		return
	}

	performance, err := h.service.GetDriverPerformance(c.Request.Context(), companyID.(string), driverID, "monthly")
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to get driver behavior analysis",
		})
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
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error:   "unauthorized",
			Message: "company ID not found in token",
		})
		return
	}

	driverID := c.Query("driver_id")
	if driverID == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "missing_parameter",
			Message: "Driver ID is required",
		})
		return
	}

	performance, err := h.service.GetDriverPerformance(c.Request.Context(), companyID.(string), driverID, "monthly")
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to get driver recommendations",
		})
		return
	}

	// Extract recommendations-specific data
	recommendationsData := map[string]interface{}{
		"driver_id":      performance.DriverID,
		"driver_name":    performance.DriverName,
		"score":         performance.Score,
		"recommendations": performance.Recommendations,
		"improvement_areas": performance.ImprovementAreas,
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Success: true,
		Data:    recommendationsData,
	})
}

// GetFleetUtilization godoc
// @Summary Get fleet utilization analytics
// @Description Get fleet utilization rates and efficiency metrics
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
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error:   "unauthorized",
			Message: "company ID not found in token",
		})
		return
	}

	dashboard, err := h.service.GetFleetDashboard(c.Request.Context(), companyID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to get fleet utilization data",
		})
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
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error:   "unauthorized",
			Message: "company ID not found in token",
		})
		return
	}

	dashboard, err := h.service.GetFleetDashboard(c.Request.Context(), companyID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to get fleet cost data",
		})
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
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error:   "unauthorized",
			Message: "company ID not found in token",
		})
		return
	}

	dashboard, err := h.service.GetFleetDashboard(c.Request.Context(), companyID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to get maintenance insights",
		})
		return
	}

	// Extract maintenance-specific data
	maintenanceData := map[string]interface{}{
		"alerts":         dashboard.MaintenanceAlerts,
		"alert_count":    len(dashboard.MaintenanceAlerts),
		"active_vehicles": dashboard.ActiveVehicles,
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Success: true,
		Data:    maintenanceData,
	})
}

// GenerateReport godoc
// @Summary Generate analytics report
// @Description Generate comprehensive analytics report in various formats
// @Tags analytics
// @Produce json
// @Param type query string false "Report type (fuel, driver, fleet, compliance)"
// @Param format query string false "Report format (json, csv, pdf)"
// @Success 200 {object} SuccessResponse
// @Failure 401 {object} ErrorResponse
// @Router /api/v1/analytics/reports/generate [post]
// @Security BearerAuth
func (h *Handler) GenerateReport(c *gin.Context) {
	// Get company ID from JWT claims
	companyID, exists := c.Get("company_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error:   "unauthorized",
			Message: "company ID not found in token",
		})
		return
	}

	reportType := c.DefaultQuery("type", "fleet")
	format := c.DefaultQuery("format", "json")

	// Generate report based on type
	var reportData interface{}

	switch reportType {
	case "fuel":
		analytics, err := h.service.GetFuelConsumption(c.Request.Context(), companyID.(string), 
			time.Now().AddDate(0, 0, -30), time.Now())
		if err != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error:   "internal_error",
				Message: "Failed to generate fuel report",
			})
			return
		}
		reportData = analytics
	case "compliance":
		report, err := h.service.GetComplianceReport(c.Request.Context(), companyID.(string), "monthly")
		if err != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error:   "internal_error",
				Message: "Failed to generate compliance report",
			})
			return
		}
		reportData = report
	default: // fleet
		dashboard, err := h.service.GetFleetDashboard(c.Request.Context(), companyID.(string))
		if err != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error:   "internal_error",
				Message: "Failed to generate fleet report",
			})
			return
		}
		reportData = dashboard
	}

	// For now, return JSON format
	// In real implementation, this would generate PDF/CSV based on format parameter
	responseData := map[string]interface{}{
		"report_type": reportType,
		"format":     format,
		"data":       reportData,
		"generated_at": time.Now().Format(time.RFC3339),
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Success: true,
		Data:    responseData,
		Message: "Report generated successfully",
	})
}

// GetComplianceReport godoc
// @Summary Get compliance report
// @Description Get Indonesian regulatory compliance report
// @Tags analytics
// @Produce json
// @Param period query string false "Report period (monthly, quarterly, yearly)"
// @Success 200 {object} SuccessResponse
// @Failure 401 {object} ErrorResponse
// @Router /api/v1/analytics/reports/compliance [get]
// @Security BearerAuth
func (h *Handler) GetComplianceReport(c *gin.Context) {
	// Get company ID from JWT claims
	companyID, exists := c.Get("company_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error:   "unauthorized",
			Message: "company ID not found in token",
		})
		return
	}

	period := c.DefaultQuery("period", "monthly")

	report, err := h.service.GetComplianceReport(c.Request.Context(), companyID.(string), period)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to get compliance report",
		})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Success: true,
		Data:    report,
	})
}

// ExportReport godoc
// @Summary Export analytics report
// @Description Export analytics report in various formats
// @Tags analytics
// @Produce json
// @Param id path string true "Report ID"
// @Param format query string false "Export format (csv, pdf, excel)"
// @Success 200 {object} SuccessResponse
// @Failure 401 {object} ErrorResponse
// @Router /api/v1/analytics/reports/export/{id} [get]
// @Security BearerAuth
func (h *Handler) ExportReport(c *gin.Context) {
	// Get company ID from JWT claims
	companyID, exists := c.Get("company_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error:   "unauthorized",
			Message: "company ID not found in token",
		})
		return
	}

	reportID := c.Param("id")
	_ = companyID // Use companyID for future implementation
	format := c.DefaultQuery("format", "csv")

	// For now, return a placeholder response
	// In real implementation, this would generate and return the actual file
	exportData := map[string]interface{}{
		"report_id":    reportID,
		"format":      format,
		"download_url": "/api/v1/analytics/reports/download/" + reportID + "." + format,
		"expires_at":  time.Now().Add(24 * time.Hour).Format(time.RFC3339),
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Success: true,
		Data:    exportData,
		Message: "Export link generated successfully",
	})
}
