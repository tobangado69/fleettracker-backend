package fleet

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/tobangado69/fleettracker-pro/backend/internal/common/middleware"
)

// FleetAPI provides HTTP API for fleet management operations
type FleetAPI struct {
	fleetManager *FleetManager
}

// NewFleetAPI creates a new fleet API
func NewFleetAPI(fleetManager *FleetManager) *FleetAPI {
	return &FleetAPI{
		fleetManager: fleetManager,
	}
}

// GetFleetOverviewHandler handles fleet overview requests
func (fa *FleetAPI) GetFleetOverviewHandler(c *gin.Context) {
	// Get company information
	companyID, _ := c.Get("company_id")

	// Get fleet overview
	overview, err := fa.fleetManager.GetFleetOverview(c.Request.Context(), companyID.(string))
	if err != nil {
		middleware.AbortWithInternal(c, "Operation failed", err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"fleet_overview": overview})
}

// OptimizeFleetHandler handles fleet optimization requests
func (fa *FleetAPI) OptimizeFleetHandler(c *gin.Context) {
	var req FleetOptimizationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.AbortWithBadRequest(c, err.Error())
		return
	}

	// Get company information
	companyID, _ := c.Get("company_id")
	req.CompanyID = companyID.(string)

	// Perform optimization
	result, err := fa.fleetManager.OptimizeFleet(c.Request.Context(), &req)
	if err != nil {
		middleware.AbortWithInternal(c, "Operation failed", err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"optimization_result": result})
}

// OptimizeRoutesHandler handles route optimization requests
func (fa *FleetAPI) OptimizeRoutesHandler(c *gin.Context) {
	var req RouteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.AbortWithBadRequest(c, err.Error())
		return
	}

	// Get company information
	companyID, _ := c.Get("company_id")
	req.CompanyID = companyID.(string)

	// Optimize route
	route, err := fa.fleetManager.routeOptimizer.OptimizeRoute(c.Request.Context(), &req)
	if err != nil {
		middleware.AbortWithInternal(c, "Operation failed", err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"optimized_route": route})
}

// RecordFuelConsumptionHandler handles fuel consumption recording
func (fa *FleetAPI) RecordFuelConsumptionHandler(c *gin.Context) {
	var record FuelRecord
	if err := c.ShouldBindJSON(&record); err != nil {
		middleware.AbortWithBadRequest(c, err.Error())
		return
	}

	// Get company information
	companyID, _ := c.Get("company_id")
	record.CompanyID = companyID.(string)

	// Record fuel consumption
	err := fa.fleetManager.fuelManager.RecordFuelConsumption(c.Request.Context(), &record)
	if err != nil {
		middleware.AbortWithInternal(c, "Operation failed", err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Fuel consumption recorded successfully"})
}

// GetFuelAnalyticsHandler handles fuel analytics requests
func (fa *FleetAPI) GetFuelAnalyticsHandler(c *gin.Context) {
	// Get company information
	companyID, _ := c.Get("company_id")

	// Get query parameters
	period := c.DefaultQuery("period", "monthly")
	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")

	// Parse dates
	var startDate, endDate time.Time
	var err error

	if startDateStr != "" {
		startDate, err = time.Parse("2006-01-02", startDateStr)
		if err != nil {
			middleware.AbortWithBadRequest(c, "Invalid start_date format")
			return
		}
	} else {
		startDate = time.Now().AddDate(0, 0, -30) // Default to last 30 days
	}

	if endDateStr != "" {
		endDate, err = time.Parse("2006-01-02", endDateStr)
		if err != nil {
			middleware.AbortWithBadRequest(c, "Invalid end_date format")
			return
		}
	} else {
		endDate = time.Now()
	}

	// Get fuel analytics
	analytics, err := fa.fleetManager.fuelManager.GetFuelAnalytics(c.Request.Context(), companyID.(string), period, startDate, endDate)
	if err != nil {
		middleware.AbortWithInternal(c, "Operation failed", err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"fuel_analytics": analytics})
}

// GetVehicleFuelHistoryHandler handles vehicle fuel history requests
func (fa *FleetAPI) GetVehicleFuelHistoryHandler(c *gin.Context) {
	vehicleID := c.Param("vehicle_id")
	if vehicleID == "" {
		middleware.AbortWithBadRequest(c, "Vehicle ID is required")
		return
	}

	// Get limit parameter
	limitStr := c.DefaultQuery("limit", "50")
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid limit parameter"})
		return
	}

	// Get fuel history
	history, err := fa.fleetManager.fuelManager.GetVehicleFuelHistory(c.Request.Context(), vehicleID, limit)
	if err != nil {
		middleware.AbortWithInternal(c, "Operation failed", err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"fuel_history": history})
}

// GetFuelAlertsHandler handles fuel alerts requests
func (fa *FleetAPI) GetFuelAlertsHandler(c *gin.Context) {
	// Get company information
	companyID, _ := c.Get("company_id")

	// Get severity parameter
	severity := c.Query("severity")

	// Get fuel alerts
	alerts, err := fa.fleetManager.fuelManager.GetFuelAlerts(c.Request.Context(), companyID.(string), severity)
	if err != nil {
		middleware.AbortWithInternal(c, "Operation failed", err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"fuel_alerts": alerts})
}

// PredictFuelConsumptionHandler handles fuel consumption prediction
func (fa *FleetAPI) PredictFuelConsumptionHandler(c *gin.Context) {
	vehicleID := c.Param("vehicle_id")
	if vehicleID == "" {
		middleware.AbortWithBadRequest(c, "Vehicle ID is required")
		return
	}

	// Get query parameters
	distanceStr := c.Query("distance")
	routeType := c.DefaultQuery("route_type", "mixed")

	distance, err := strconv.ParseFloat(distanceStr, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid distance parameter"})
		return
	}

	// Predict fuel consumption
	predictedFuel, predictedCost, err := fa.fleetManager.fuelManager.PredictFuelConsumption(c.Request.Context(), vehicleID, distance, routeType)
	if err != nil {
		middleware.AbortWithInternal(c, "Operation failed", err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"predicted_fuel": predictedFuel,
		"predicted_cost": predictedCost,
		"distance":       distance,
		"route_type":     routeType,
	})
}

// CreateMaintenanceRuleHandler handles maintenance rule creation
func (fa *FleetAPI) CreateMaintenanceRuleHandler(c *gin.Context) {
	var rule MaintenanceRule
	if err := c.ShouldBindJSON(&rule); err != nil {
		middleware.AbortWithBadRequest(c, err.Error())
		return
	}

	// Get company information
	companyID, _ := c.Get("company_id")
	rule.CompanyID = companyID.(string)

	// Create maintenance rule
	err := fa.fleetManager.maintenanceScheduler.CreateMaintenanceRule(c.Request.Context(), &rule)
	if err != nil {
		middleware.AbortWithInternal(c, "Operation failed", err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Maintenance rule created successfully", "rule": rule})
}

// ScheduleMaintenanceHandler handles maintenance scheduling
func (fa *FleetAPI) ScheduleMaintenanceHandler(c *gin.Context) {
	var schedule MaintenanceSchedule
	if err := c.ShouldBindJSON(&schedule); err != nil {
		middleware.AbortWithBadRequest(c, err.Error())
		return
	}

	// Get company information
	companyID, _ := c.Get("company_id")
	schedule.CompanyID = companyID.(string)

	// Schedule maintenance
	err := fa.fleetManager.maintenanceScheduler.ScheduleMaintenance(c.Request.Context(), &schedule)
	if err != nil {
		middleware.AbortWithInternal(c, "Operation failed", err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Maintenance scheduled successfully", "schedule": schedule})
}

// CompleteMaintenanceHandler handles maintenance completion
func (fa *FleetAPI) CompleteMaintenanceHandler(c *gin.Context) {
	scheduleID := c.Param("schedule_id")
	if scheduleID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Schedule ID is required"})
		return
	}

	var req struct {
		ActualCost      float64            `json:"actual_cost"`
		CompletionNotes string             `json:"completion_notes"`
		PartsUsed       []MaintenancePart  `json:"parts_used"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.AbortWithBadRequest(c, err.Error())
		return
	}

	// Complete maintenance
	err := fa.fleetManager.maintenanceScheduler.CompleteMaintenance(c.Request.Context(), scheduleID, req.ActualCost, req.CompletionNotes, req.PartsUsed)
	if err != nil {
		middleware.AbortWithInternal(c, "Operation failed", err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Maintenance completed successfully"})
}

// GetUpcomingMaintenanceHandler handles upcoming maintenance requests
func (fa *FleetAPI) GetUpcomingMaintenanceHandler(c *gin.Context) {
	// Get company information
	companyID, _ := c.Get("company_id")

	// Get days parameter
	daysStr := c.DefaultQuery("days", "7")
	days, err := strconv.Atoi(daysStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid days parameter"})
		return
	}

	// Get upcoming maintenance
	maintenance, err := fa.fleetManager.maintenanceScheduler.GetUpcomingMaintenance(c.Request.Context(), companyID.(string), days)
	if err != nil {
		middleware.AbortWithInternal(c, "Operation failed", err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"upcoming_maintenance": maintenance})
}

// GetMaintenanceAlertsHandler handles maintenance alerts requests
func (fa *FleetAPI) GetMaintenanceAlertsHandler(c *gin.Context) {
	// Get company information
	companyID, _ := c.Get("company_id")

	// Get severity parameter
	severity := c.Query("severity")

	// Get maintenance alerts
	alerts, err := fa.fleetManager.maintenanceScheduler.GetMaintenanceAlerts(c.Request.Context(), companyID.(string), severity)
	if err != nil {
		middleware.AbortWithInternal(c, "Operation failed", err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"maintenance_alerts": alerts})
}

// GetMaintenanceAnalyticsHandler handles maintenance analytics requests
func (fa *FleetAPI) GetMaintenanceAnalyticsHandler(c *gin.Context) {
	// Get company information
	companyID, _ := c.Get("company_id")

	// Get query parameters
	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")

	// Parse dates
	var startDate, endDate time.Time
	var err error

	if startDateStr != "" {
		startDate, err = time.Parse("2006-01-02", startDateStr)
		if err != nil {
			middleware.AbortWithBadRequest(c, "Invalid start_date format")
			return
		}
	} else {
		startDate = time.Now().AddDate(0, 0, -30) // Default to last 30 days
	}

	if endDateStr != "" {
		endDate, err = time.Parse("2006-01-02", endDateStr)
		if err != nil {
			middleware.AbortWithBadRequest(c, "Invalid end_date format")
			return
		}
	} else {
		endDate = time.Now()
	}

	// Get maintenance analytics
	analytics, err := fa.fleetManager.maintenanceScheduler.GetMaintenanceAnalytics(c.Request.Context(), companyID.(string), startDate, endDate)
	if err != nil {
		middleware.AbortWithInternal(c, "Operation failed", err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"maintenance_analytics": analytics})
}

// AssignDriverHandler handles driver assignment requests
func (fa *FleetAPI) AssignDriverHandler(c *gin.Context) {
	var req AssignmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.AbortWithBadRequest(c, err.Error())
		return
	}

	// Get company information
	companyID, _ := c.Get("company_id")
	req.CompanyID = companyID.(string)

	// Assign driver
	assignment, err := fa.fleetManager.driverAssigner.AssignDriver(c.Request.Context(), &req)
	if err != nil {
		middleware.AbortWithInternal(c, "Operation failed", err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"driver_assignment": assignment})
}

// GetDriverRecommendationsHandler handles driver recommendation requests
func (fa *FleetAPI) GetDriverRecommendationsHandler(c *gin.Context) {
	var req AssignmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.AbortWithBadRequest(c, err.Error())
		return
	}

	// Get company information
	companyID, _ := c.Get("company_id")
	req.CompanyID = companyID.(string)

	// Get limit parameter
	limitStr := c.DefaultQuery("limit", "5")
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid limit parameter"})
		return
	}

	// Get driver recommendations
	recommendations, err := fa.fleetManager.driverAssigner.GetDriverRecommendations(c.Request.Context(), &req, limit)
	if err != nil {
		middleware.AbortWithInternal(c, "Operation failed", err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"driver_recommendations": recommendations})
}

// GetAssignmentAnalyticsHandler handles assignment analytics requests
func (fa *FleetAPI) GetAssignmentAnalyticsHandler(c *gin.Context) {
	// Get company information
	companyID, _ := c.Get("company_id")

	// Get query parameters
	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")

	// Parse dates
	var startDate, endDate time.Time
	var err error

	if startDateStr != "" {
		startDate, err = time.Parse("2006-01-02", startDateStr)
		if err != nil {
			middleware.AbortWithBadRequest(c, "Invalid start_date format")
			return
		}
	} else {
		startDate = time.Now().AddDate(0, 0, -30) // Default to last 30 days
	}

	if endDateStr != "" {
		endDate, err = time.Parse("2006-01-02", endDateStr)
		if err != nil {
			middleware.AbortWithBadRequest(c, "Invalid end_date format")
			return
		}
	} else {
		endDate = time.Now()
	}

	// Get assignment analytics
	analytics, err := fa.fleetManager.driverAssigner.GetAssignmentAnalytics(c.Request.Context(), companyID.(string), startDate, endDate)
	if err != nil {
		middleware.AbortWithInternal(c, "Operation failed", err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"assignment_analytics": analytics})
}

// CheckMaintenanceTriggersHandler handles maintenance trigger checking
func (fa *FleetAPI) CheckMaintenanceTriggersHandler(c *gin.Context) {
	// Get company information
	companyID, _ := c.Get("company_id")

	// Check maintenance triggers
	err := fa.fleetManager.ScheduleMaintenanceCheck(c.Request.Context(), companyID.(string))
	if err != nil {
		middleware.AbortWithInternal(c, "Operation failed", err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Maintenance triggers checked successfully"})
}

// ProcessFuelConsumptionHandler handles fuel consumption processing
func (fa *FleetAPI) ProcessFuelConsumptionHandler(c *gin.Context) {
	// Get company information
	companyID, _ := c.Get("company_id")

	// Process fuel consumption
	err := fa.fleetManager.ProcessFuelConsumption(c.Request.Context(), companyID.(string))
	if err != nil {
		middleware.AbortWithInternal(c, "Operation failed", err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Fuel consumption processed successfully"})
}

// SetupFleetRoutes sets up fleet management API routes
func SetupFleetRoutes(r *gin.RouterGroup, api *FleetAPI) {
	fleet := r.Group("/fleet")
	{
		// Fleet Overview
		fleet.GET("/overview", api.GetFleetOverviewHandler)
		fleet.POST("/optimize", api.OptimizeFleetHandler)

		// Route Optimization
		routes := fleet.Group("/routes")
		{
			routes.POST("/optimize", api.OptimizeRoutesHandler)
		}

		// Fuel Management
		fuel := fleet.Group("/fuel")
		{
			fuel.POST("/consumption", api.RecordFuelConsumptionHandler)
			fuel.GET("/analytics", api.GetFuelAnalyticsHandler)
			fuel.GET("/history/:vehicle_id", api.GetVehicleFuelHistoryHandler)
			fuel.GET("/alerts", api.GetFuelAlertsHandler)
			fuel.GET("/predict/:vehicle_id", api.PredictFuelConsumptionHandler)
		}

		// Maintenance Management
		maintenance := fleet.Group("/maintenance")
		{
			maintenance.POST("/rules", api.CreateMaintenanceRuleHandler)
			maintenance.POST("/schedule", api.ScheduleMaintenanceHandler)
			maintenance.PUT("/complete/:schedule_id", api.CompleteMaintenanceHandler)
			maintenance.GET("/upcoming", api.GetUpcomingMaintenanceHandler)
			maintenance.GET("/alerts", api.GetMaintenanceAlertsHandler)
			maintenance.GET("/analytics", api.GetMaintenanceAnalyticsHandler)
		}

		// Driver Assignment
		drivers := fleet.Group("/drivers")
		{
			drivers.POST("/assign", api.AssignDriverHandler)
			drivers.POST("/recommendations", api.GetDriverRecommendationsHandler)
			drivers.GET("/analytics", api.GetAssignmentAnalyticsHandler)
		}

		// System Operations
		system := fleet.Group("/system")
		{
			system.POST("/check-maintenance", api.CheckMaintenanceTriggersHandler)
			system.POST("/process-fuel", api.ProcessFuelConsumptionHandler)
		}
	}
}
