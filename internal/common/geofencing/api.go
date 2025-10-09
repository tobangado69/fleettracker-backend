package geofencing

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/tobangado69/fleettracker-pro/backend/internal/common/middleware"
)

// GeofenceAPI provides HTTP API for geofencing operations
type GeofenceAPI struct {
	geofenceManager *GeofenceManager
}

// NewGeofenceAPI creates a new geofence API
func NewGeofenceAPI(geofenceManager *GeofenceManager) *GeofenceAPI {
	return &GeofenceAPI{
		geofenceManager: geofenceManager,
	}
}

// CreateGeofenceHandler handles geofence creation requests
func (ga *GeofenceAPI) CreateGeofenceHandler(c *gin.Context) {
	var geofence Geofence
	if err := c.ShouldBindJSON(&geofence); err != nil {
		middleware.AbortWithBadRequest(c, err.Error())
		return
	}

	// Get company information
	companyID, _ := c.Get("company_id")
	geofence.CompanyID = companyID.(string)

	// Get user information
	userID, _ := c.Get("user_id")
	geofence.CreatedBy = userID.(string)

	// Create geofence
	err := ga.geofenceManager.CreateGeofence(c.Request.Context(), &geofence)
	if err != nil {
		middleware.AbortWithInternal(c, "Operation failed", err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Geofence created successfully", "geofence": geofence})
}

// UpdateGeofenceHandler handles geofence update requests
func (ga *GeofenceAPI) UpdateGeofenceHandler(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		middleware.AbortWithBadRequest(c, "Geofence ID is required")
		return
	}

	var updates Geofence
	if err := c.ShouldBindJSON(&updates); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Update geofence
	err := ga.geofenceManager.UpdateGeofence(c.Request.Context(), id, &updates)
	if err != nil {
		middleware.AbortWithInternal(c, "Operation failed", err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Geofence updated successfully"})
}

// DeleteGeofenceHandler handles geofence deletion requests
func (ga *GeofenceAPI) DeleteGeofenceHandler(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		middleware.AbortWithBadRequest(c, "Geofence ID is required")
		return
	}

	// Delete geofence
	err := ga.geofenceManager.DeleteGeofence(c.Request.Context(), id)
	if err != nil {
		middleware.AbortWithInternal(c, "Operation failed", err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Geofence deleted successfully"})
}

// GetGeofencesHandler handles geofence retrieval requests
func (ga *GeofenceAPI) GetGeofencesHandler(c *gin.Context) {
	// Get company information
	companyID, _ := c.Get("company_id")

	// Get active only parameter
	activeOnlyStr := c.DefaultQuery("active_only", "true")
	activeOnly, err := strconv.ParseBool(activeOnlyStr)
	if err != nil {
		activeOnly = true
	}

	// Get geofences
	geofences, err := ga.geofenceManager.GetGeofences(c.Request.Context(), companyID.(string), activeOnly)
	if err != nil {
		middleware.AbortWithInternal(c, "Operation failed", err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"geofences": geofences})
}

// GetGeofenceHandler handles single geofence retrieval requests
func (ga *GeofenceAPI) GetGeofenceHandler(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		middleware.AbortWithBadRequest(c, "Geofence ID is required")
		return
	}

	// Get geofence
	geofence, err := ga.geofenceManager.GetGeofence(c.Request.Context(), id)
	if err != nil {
		middleware.AbortWithInternal(c, "Operation failed", err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"geofence": geofence})
}

// CheckGeofencesHandler handles geofence checking requests
func (ga *GeofenceAPI) CheckGeofencesHandler(c *gin.Context) {
	var req GeofenceCheckRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get company information
	companyID, _ := c.Get("company_id")
	req.CompanyID = companyID.(string)

	// Check geofences
	result, err := ga.geofenceManager.CheckGeofences(c.Request.Context(), &req)
	if err != nil {
		middleware.AbortWithInternal(c, "Operation failed", err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"geofence_check_result": result})
}

// GetGeofenceEventsHandler handles geofence events retrieval requests
func (ga *GeofenceAPI) GetGeofenceEventsHandler(c *gin.Context) {
	// Get company information
	companyID, _ := c.Get("company_id")

	// Build filters
	filters := make(map[string]interface{})

	if geofenceID := c.Query("geofence_id"); geofenceID != "" {
		filters["geofence_id"] = geofenceID
	}
	if vehicleID := c.Query("vehicle_id"); vehicleID != "" {
		filters["vehicle_id"] = vehicleID
	}
	if driverID := c.Query("driver_id"); driverID != "" {
		filters["driver_id"] = driverID
	}
	if eventType := c.Query("event_type"); eventType != "" {
		filters["event_type"] = eventType
	}
	if startDateStr := c.Query("start_date"); startDateStr != "" {
		if startDate, err := time.Parse("2006-01-02", startDateStr); err == nil {
			filters["start_date"] = startDate
		}
	}
	if endDateStr := c.Query("end_date"); endDateStr != "" {
		if endDate, err := time.Parse("2006-01-02", endDateStr); err == nil {
			filters["end_date"] = endDate
		}
	}

	// Get geofence events
	events, err := ga.geofenceManager.GetGeofenceEvents(c.Request.Context(), companyID.(string), filters)
	if err != nil {
		middleware.AbortWithInternal(c, "Operation failed", err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"geofence_events": events})
}

// GetGeofenceViolationsHandler handles geofence violations retrieval requests
func (ga *GeofenceAPI) GetGeofenceViolationsHandler(c *gin.Context) {
	// Get company information
	companyID, _ := c.Get("company_id")

	// Build filters
	filters := make(map[string]interface{})

	if geofenceID := c.Query("geofence_id"); geofenceID != "" {
		filters["geofence_id"] = geofenceID
	}
	if vehicleID := c.Query("vehicle_id"); vehicleID != "" {
		filters["vehicle_id"] = vehicleID
	}
	if driverID := c.Query("driver_id"); driverID != "" {
		filters["driver_id"] = driverID
	}
	if violationType := c.Query("violation_type"); violationType != "" {
		filters["violation_type"] = violationType
	}
	if severity := c.Query("severity"); severity != "" {
		filters["severity"] = severity
	}
	if isResolvedStr := c.Query("is_resolved"); isResolvedStr != "" {
		if isResolved, err := strconv.ParseBool(isResolvedStr); err == nil {
			filters["is_resolved"] = isResolved
		}
	}
	if startDateStr := c.Query("start_date"); startDateStr != "" {
		if startDate, err := time.Parse("2006-01-02", startDateStr); err == nil {
			filters["start_date"] = startDate
		}
	}
	if endDateStr := c.Query("end_date"); endDateStr != "" {
		if endDate, err := time.Parse("2006-01-02", endDateStr); err == nil {
			filters["end_date"] = endDate
		}
	}

	// Get geofence violations
	violations, err := ga.geofenceManager.GetGeofenceViolations(c.Request.Context(), companyID.(string), filters)
	if err != nil {
		middleware.AbortWithInternal(c, "Operation failed", err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"geofence_violations": violations})
}

// ResolveViolationHandler handles violation resolution requests
func (ga *GeofenceAPI) ResolveViolationHandler(c *gin.Context) {
	violationID := c.Param("violation_id")
	if violationID == "" {
		middleware.AbortWithBadRequest(c, "Violation ID is required")
		return
	}

	var req struct {
		Resolution string `json:"resolution" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get user information
	userID, _ := c.Get("user_id")

	// Resolve violation
	err := ga.geofenceManager.ResolveViolation(c.Request.Context(), violationID, userID.(string), req.Resolution)
	if err != nil {
		middleware.AbortWithInternal(c, "Operation failed", err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Violation resolved successfully"})
}

// GetGeofenceAnalyticsHandler handles geofence analytics requests
func (ga *GeofenceAPI) GetGeofenceAnalyticsHandler(c *gin.Context) {
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

	// Get geofence analytics
	analytics, err := ga.geofenceManager.GetGeofenceAnalytics(c.Request.Context(), companyID.(string), startDate, endDate)
	if err != nil {
		middleware.AbortWithInternal(c, "Operation failed", err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"geofence_analytics": analytics})
}

// CreateGeofenceFromMapHandler handles creating geofences from map coordinates
func (ga *GeofenceAPI) CreateGeofenceFromMapHandler(c *gin.Context) {
	var req struct {
		Name        string        `json:"name" binding:"required"`
		Description string        `json:"description"`
		Type        string        `json:"type" binding:"required"` // polygon, circle, rectangle
		Coordinates []Coordinate  `json:"coordinates" binding:"required"`
		Radius      float64       `json:"radius"` // for circular geofences
		Priority    int           `json:"priority"`
		Color       string        `json:"color"`
		AlertOnEntry bool         `json:"alert_on_entry"`
		AlertOnExit  bool         `json:"alert_on_exit"`
		AlertOnDwell bool         `json:"alert_on_dwell"`
		DwellTime    int          `json:"dwell_time"`
		SpeedLimit   float64      `json:"speed_limit"`
		AlertOnSpeed bool         `json:"alert_on_speed"`
		TimeRestrictions []TimeRestriction `json:"time_restrictions"`
		AllowedVehicles []string  `json:"allowed_vehicles"`
		AllowedDrivers  []string  `json:"allowed_drivers"`
		RestrictedVehicles []string `json:"restricted_vehicles"`
		RestrictedDrivers  []string `json:"restricted_drivers"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get company and user information
	companyID, _ := c.Get("company_id")
	userID, _ := c.Get("user_id")

	// Create geofence
	geofence := &Geofence{
		CompanyID:         companyID.(string),
		Name:              req.Name,
		Description:       req.Description,
		Type:              req.Type,
		Coordinates:       req.Coordinates,
		Radius:            req.Radius,
		Priority:          req.Priority,
		Color:             req.Color,
		AlertOnEntry:      req.AlertOnEntry,
		AlertOnExit:       req.AlertOnExit,
		AlertOnDwell:      req.AlertOnDwell,
		DwellTime:         req.DwellTime,
		SpeedLimit:        req.SpeedLimit,
		AlertOnSpeed:      req.AlertOnSpeed,
		TimeRestrictions:  req.TimeRestrictions,
		AllowedVehicles:   req.AllowedVehicles,
		AllowedDrivers:    req.AllowedDrivers,
		RestrictedVehicles: req.RestrictedVehicles,
		RestrictedDrivers:  req.RestrictedDrivers,
		IsActive:          true,
		CreatedBy:         userID.(string),
	}

	// Set defaults
	if geofence.Priority == 0 {
		geofence.Priority = 5
	}
	if geofence.Color == "" {
		geofence.Color = "#FF0000"
	}
	if geofence.AlertOnEntry {
		geofence.AlertOnEntry = true
	}
	if geofence.AlertOnExit {
		geofence.AlertOnExit = true
	}

	// Create geofence
	err := ga.geofenceManager.CreateGeofence(c.Request.Context(), geofence)
	if err != nil {
		middleware.AbortWithInternal(c, "Operation failed", err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Geofence created successfully", "geofence": geofence})
}

// BulkCheckGeofencesHandler handles bulk geofence checking for multiple vehicles
func (ga *GeofenceAPI) BulkCheckGeofencesHandler(c *gin.Context) {
	var req struct {
		Checks []GeofenceCheckRequest `json:"checks" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get company information
	companyID, _ := c.Get("company_id")

	var results []GeofenceCheckResult

	// Check geofences for each request
	for _, checkReq := range req.Checks {
		checkReq.CompanyID = companyID.(string)
		
		result, err := ga.geofenceManager.CheckGeofences(c.Request.Context(), &checkReq)
		if err != nil {
			// Log error but continue with other checks
			continue
		}
		
		results = append(results, *result)
	}

	c.JSON(http.StatusOK, gin.H{"bulk_check_results": results})
}

// GetGeofenceHeatmapHandler handles geofence heatmap data requests
func (ga *GeofenceAPI) GetGeofenceHeatmapHandler(c *gin.Context) {
	// Get company information
	companyID, _ := c.Get("company_id")

	// Get query parameters
	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")
	geofenceID := c.Query("geofence_id")

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
		startDate = time.Now().AddDate(0, 0, -7) // Default to last 7 days
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

	// Build filters
	filters := make(map[string]interface{})
	filters["start_date"] = startDate
	filters["end_date"] = endDate
	if geofenceID != "" {
		filters["geofence_id"] = geofenceID
	}

	// Get geofence events for heatmap
	events, err := ga.geofenceManager.GetGeofenceEvents(c.Request.Context(), companyID.(string), filters)
	if err != nil {
		middleware.AbortWithInternal(c, "Operation failed", err)
		return
	}

	// Create heatmap data
	heatmapData := make(map[string]interface{})
	heatmapData["events"] = events
	heatmapData["period"] = fmt.Sprintf("%s to %s", startDate.Format("2006-01-02"), endDate.Format("2006-01-02"))
	heatmapData["total_events"] = len(events)

	c.JSON(http.StatusOK, gin.H{"heatmap_data": heatmapData})
}

// SetupGeofenceRoutes sets up geofencing API routes
func SetupGeofenceRoutes(r *gin.RouterGroup, api *GeofenceAPI) {
	geofences := r.Group("/geofences")
	{
		// Geofence CRUD operations
		geofences.POST("", api.CreateGeofenceHandler)
		geofences.POST("/from-map", api.CreateGeofenceFromMapHandler)
		geofences.GET("", api.GetGeofencesHandler)
		geofences.GET("/:id", api.GetGeofenceHandler)
		geofences.PUT("/:id", api.UpdateGeofenceHandler)
		geofences.DELETE("/:id", api.DeleteGeofenceHandler)

		// Geofence checking and monitoring
		geofences.POST("/check", api.CheckGeofencesHandler)
		geofences.POST("/bulk-check", api.BulkCheckGeofencesHandler)

		// Events and violations
		geofences.GET("/events", api.GetGeofenceEventsHandler)
		geofences.GET("/violations", api.GetGeofenceViolationsHandler)
		geofences.PUT("/violations/:violation_id/resolve", api.ResolveViolationHandler)

		// Analytics and reporting
		geofences.GET("/analytics", api.GetGeofenceAnalyticsHandler)
		geofences.GET("/heatmap", api.GetGeofenceHeatmapHandler)
	}
}
