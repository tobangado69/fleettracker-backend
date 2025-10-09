package tracking

import (
	std_errors "errors" // Alias standard errors to avoid conflict
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"gorm.io/gorm"

	"github.com/tobangado69/fleettracker-pro/backend/internal/common/middleware"
	apperrors "github.com/tobangado69/fleettracker-pro/backend/pkg/errors"
	"github.com/tobangado69/fleettracker-pro/backend/pkg/models"
)

// Handler handles tracking HTTP requests
type Handler struct {
	service   *Service
	validator *validator.Validate
}

// NewHandler creates a new tracking handler
func NewHandler(service *Service) *Handler {
	return &Handler{
		service:   service,
		validator: validator.New(),
	}
}

// SuccessResponse represents a success response
type SuccessResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data"`
	Message string      `json:"message,omitempty"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Success bool   `json:"success" example:"false"`
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}

// PaginatedResponse represents a paginated response
type PaginatedResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data"`
	Meta    Meta        `json:"meta"`
}

// Meta represents pagination metadata
type Meta struct {
	Total       int64 `json:"total"`
	Page        int   `json:"page"`
	Limit       int   `json:"limit"`
	TotalPages  int   `json:"total_pages"`
	HasNext     bool  `json:"has_next"`
	HasPrevious bool  `json:"has_previous"`
}

// ProcessGPSData godoc
// @Summary Submit GPS data
// @Description Submit GPS data from mobile device for real-time tracking
// @Tags tracking
// @Accept json
// @Produce json
// @Param gps body GPSDataRequest true "GPS data"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/tracking/gps [post]
// @Security BearerAuth
func (h *Handler) ProcessGPSData(c *gin.Context) {
	var req GPSDataRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.AbortWithBadRequest(c, "invalid request data")
		return
	}

	// Validate request
	if err := h.validator.Struct(&req); err != nil {
		middleware.AbortWithValidation(c, err.Error())
		return
	}

	// Process GPS data
	gpsTrack, err := h.service.ProcessGPSData(req)
	if err != nil {
		if appErr, ok := err.(*apperrors.AppError); ok {
			middleware.AbortWithError(c, appErr)
		} else {
			middleware.AbortWithInternal(c, "failed to process GPS data", err)
		}
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Success: true,
		Data:    gpsTrack,
		Message: "GPS data processed successfully",
	})
}

// GetCurrentLocation godoc
// @Summary Get current vehicle location
// @Description Get the current location of a vehicle
// @Tags tracking
// @Produce json
// @Param id path string true "Vehicle ID"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/tracking/vehicles/{id}/current [get]
// @Security BearerAuth
func (h *Handler) GetCurrentLocation(c *gin.Context) {
	vehicleID := c.Param("id")
	if vehicleID == "" {
		middleware.AbortWithBadRequest(c, "vehicle ID is required")
		return
	}

	// Get company ID from JWT claims for authorization
	companyID, exists := c.Get("company_id")
	if !exists {
		middleware.AbortWithUnauthorized(c, "company ID not found in token")
		return
	}

	// Verify vehicle belongs to company
	var vehicle models.Vehicle
	if err := h.service.db.Where("id = ? AND company_id = ?", vehicleID, companyID).First(&vehicle).Error; err != nil {
		if std_errors.Is(err, gorm.ErrRecordNotFound) {
			middleware.AbortWithNotFound(c, "vehicle")
		} else {
			middleware.AbortWithInternal(c, "failed to verify vehicle", err)
		}
		return
	}

	// Get current location
	location, err := h.service.GetCurrentLocation(vehicleID)
	if err != nil {
		if appErr, ok := err.(*apperrors.AppError); ok {
			middleware.AbortWithError(c, appErr)
		} else {
			middleware.AbortWithInternal(c, "failed to get current location", err)
		}
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Success: true,
		Data:    location,
	})
}

// GetLocationHistory godoc
// @Summary Get vehicle location history
// @Description Get historical GPS data for a vehicle with filtering and pagination
// @Tags tracking
// @Produce json
// @Param id path string true "Vehicle ID"
// @Param driver_id query string false "Driver ID"
// @Param start_time query string false "Start time (RFC3339)"
// @Param end_time query string false "End time (RFC3339)"
// @Param min_accuracy query number false "Minimum GPS accuracy"
// @Param max_speed query number false "Maximum speed"
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(100)
// @Param sort_by query string false "Sort field" default(timestamp)
// @Param sort_order query string false "Sort order" default(desc)
// @Success 200 {object} PaginatedResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/tracking/vehicles/{id}/history [get]
// @Security BearerAuth
func (h *Handler) GetLocationHistory(c *gin.Context) {
	vehicleID := c.Param("id")
	if vehicleID == "" {
		middleware.AbortWithBadRequest(c, "vehicle ID is required")
		return
	}

	// Get company ID from JWT claims for authorization
	companyID, exists := c.Get("company_id")
	if !exists {
		middleware.AbortWithUnauthorized(c, "company ID not found in token")
		return
	}

	// Verify vehicle belongs to company
	var vehicle models.Vehicle
	if err := h.service.db.Where("id = ? AND company_id = ?", vehicleID, companyID).First(&vehicle).Error; err != nil {
		if std_errors.Is(err, gorm.ErrRecordNotFound) {
			middleware.AbortWithNotFound(c, "vehicle")
		} else {
			middleware.AbortWithInternal(c, "failed to verify vehicle", err)
		}
		return
	}

	// Parse filters
	filters := GPSFilters{
		Page:      1,
		Limit:     100,
		SortBy:    "timestamp",
		SortOrder: "desc",
	}

	if driverID := c.Query("driver_id"); driverID != "" {
		filters.DriverID = &driverID
	}
	if startTimeStr := c.Query("start_time"); startTimeStr != "" {
		if startTime, err := time.Parse(time.RFC3339, startTimeStr); err == nil {
			filters.StartTime = &startTime
		}
	}
	if endTimeStr := c.Query("end_time"); endTimeStr != "" {
		if endTime, err := time.Parse(time.RFC3339, endTimeStr); err == nil {
			filters.EndTime = &endTime
		}
	}
	if minAccuracyStr := c.Query("min_accuracy"); minAccuracyStr != "" {
		if minAccuracy, err := strconv.ParseFloat(minAccuracyStr, 64); err == nil {
			filters.MinAccuracy = &minAccuracy
		}
	}
	if maxSpeedStr := c.Query("max_speed"); maxSpeedStr != "" {
		if maxSpeed, err := strconv.ParseFloat(maxSpeedStr, 64); err == nil {
			filters.MaxSpeed = &maxSpeed
		}
	}
	if pageStr := c.Query("page"); pageStr != "" {
		if page, err := strconv.Atoi(pageStr); err == nil && page > 0 {
			filters.Page = page
		}
	}
	if limitStr := c.Query("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 && limit <= 1000 {
			filters.Limit = limit
		}
	}
	if sortBy := c.Query("sort_by"); sortBy != "" {
		filters.SortBy = sortBy
	}
	if sortOrder := c.Query("sort_order"); sortOrder != "" {
		filters.SortOrder = sortOrder
	}

	// Get location history
	gpsTracks, total, err := h.service.GetLocationHistory(vehicleID, filters)
	if err != nil {
		if appErr, ok := err.(*apperrors.AppError); ok {
			middleware.AbortWithError(c, appErr)
		} else {
			middleware.AbortWithInternal(c, "failed to get location history", err)
		}
		return
	}

	// Calculate pagination metadata
	totalPages := int((total + int64(filters.Limit) - 1) / int64(filters.Limit))
	hasNext := filters.Page < totalPages
	hasPrevious := filters.Page > 1

	c.JSON(http.StatusOK, PaginatedResponse{
		Success: true,
		Data:    gpsTracks,
		Meta: Meta{
			Total:       total,
			Page:        filters.Page,
			Limit:       filters.Limit,
			TotalPages:  totalPages,
			HasNext:     hasNext,
			HasPrevious: hasPrevious,
		},
	})
}

// GetRoute godoc
// @Summary Get vehicle route data
// @Description Get route information for a vehicle including distance and duration
// @Tags tracking
// @Produce json
// @Param id path string true "Vehicle ID"
// @Param start_time query string false "Start time (RFC3339)"
// @Param end_time query string false "End time (RFC3339)"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/tracking/vehicles/{id}/route [get]
// @Security BearerAuth
func (h *Handler) GetRoute(c *gin.Context) {
	vehicleID := c.Param("id")
	if vehicleID == "" {
		middleware.AbortWithBadRequest(c, "vehicle ID is required")
		return
	}

	// Get company ID from JWT claims for authorization
	companyID, exists := c.Get("company_id")
	if !exists {
		middleware.AbortWithUnauthorized(c, "company ID not found in token")
		return
	}

	// Verify vehicle belongs to company
	var vehicle models.Vehicle
	if err := h.service.db.Where("id = ? AND company_id = ?", vehicleID, companyID).First(&vehicle).Error; err != nil {
		if std_errors.Is(err, gorm.ErrRecordNotFound) {
			middleware.AbortWithNotFound(c, "vehicle")
		} else {
			middleware.AbortWithInternal(c, "failed to verify vehicle", err)
		}
		return
	}

	// Parse time filters
	var startTime, endTime time.Time
	if startTimeStr := c.Query("start_time"); startTimeStr != "" {
		var err error
		startTime, err = time.Parse(time.RFC3339, startTimeStr)
		if err != nil {
			middleware.AbortWithBadRequest(c, "invalid start_time format")
			return
		}
	} else {
		startTime = time.Now().Add(-24 * time.Hour) // Default to last 24 hours
	}

	if endTimeStr := c.Query("end_time"); endTimeStr != "" {
		var err error
		endTime, err = time.Parse(time.RFC3339, endTimeStr)
		if err != nil {
			middleware.AbortWithBadRequest(c, "invalid end_time format")
			return
		}
	} else {
		endTime = time.Now()
	}

	// Get route data using location history
	filters := GPSFilters{
		StartTime: &startTime,
		EndTime:   &endTime,
		Page:      1,
		Limit:     1000,
		SortBy:    "timestamp",
		SortOrder: "asc",
	}

	gpsTracks, _, err := h.service.GetLocationHistory(vehicleID, filters)
	if err != nil {
		if appErr, ok := err.(*apperrors.AppError); ok {
			middleware.AbortWithError(c, appErr)
		} else {
			middleware.AbortWithInternal(c, "failed to get location history for route", err)
		}
		return
	}

	// Calculate route metrics
	routeData := h.calculateRouteMetrics(gpsTracks)

	c.JSON(http.StatusOK, SuccessResponse{
		Success: true,
		Data:    routeData,
	})
}

// calculateRouteMetrics calculates route metrics from GPS tracks
func (h *Handler) calculateRouteMetrics(gpsTracks []models.GPSTrack) map[string]interface{} {
	if len(gpsTracks) == 0 {
		return map[string]interface{}{
			"distance":      0,
			"duration":      0,
			"average_speed": 0,
			"max_speed":     0,
			"points":        []interface{}{},
		}
	}

	var totalDistance float64
	var maxSpeed float64
	var totalSpeed float64
	var speedCount int
	var points []interface{}

	startTime := gpsTracks[0].Timestamp
	endTime := gpsTracks[len(gpsTracks)-1].Timestamp

	for i := 1; i < len(gpsTracks); i++ {
		prev := gpsTracks[i-1]
		curr := gpsTracks[i]

		// Calculate distance between two points
		distance := h.service.calculateDistance(prev.Latitude, prev.Longitude, curr.Latitude, curr.Longitude)
		totalDistance += distance

		// Track max speed
		if curr.Speed > maxSpeed {
			maxSpeed = curr.Speed
		}

		// Calculate average speed
		if curr.Speed > 0 {
			totalSpeed += curr.Speed
			speedCount++
		}

		// Add point to route
		points = append(points, map[string]interface{}{
			"latitude":  curr.Latitude,
			"longitude": curr.Longitude,
			"timestamp": curr.Timestamp,
			"speed":     curr.Speed,
		})
	}

	var averageSpeed float64
	if speedCount > 0 {
		averageSpeed = totalSpeed / float64(speedCount)
	}

	duration := int(endTime.Sub(startTime).Minutes())

	return map[string]interface{}{
		"distance":      totalDistance,
		"duration":      duration,
		"average_speed": averageSpeed,
		"max_speed":     maxSpeed,
		"start_time":    startTime,
		"end_time":      endTime,
		"points":        points,
	}
}

// ProcessDriverEvent godoc
// @Summary Submit driver event
// @Description Submit a driver behavior event (speed violation, harsh braking, etc.)
// @Tags tracking
// @Accept json
// @Produce json
// @Param event body DriverEventRequest true "Driver event data"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/tracking/events [post]
// @Security BearerAuth
func (h *Handler) ProcessDriverEvent(c *gin.Context) {
	var req DriverEventRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.AbortWithBadRequest(c, "invalid request data")
		return
	}

	// Validate request
	if err := h.validator.Struct(&req); err != nil {
		middleware.AbortWithValidation(c, err.Error())
		return
	}

	// Process driver event
	event, err := h.service.ProcessDriverEvent(req)
	if err != nil {
		if appErr, ok := err.(*apperrors.AppError); ok {
			middleware.AbortWithError(c, appErr)
		} else {
			middleware.AbortWithInternal(c, "failed to process driver event", err)
		}
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Success: true,
		Data:    event,
		Message: "Driver event processed successfully",
	})
}

// GetDriverEvents godoc
// @Summary Get driver events
// @Description Get driver behavior events with filtering
// @Tags tracking
// @Produce json
// @Param driver_id query string false "Driver ID"
// @Param vehicle_id query string false "Vehicle ID"
// @Param event_type query string false "Event type"
// @Param severity query string false "Severity level"
// @Param start_time query string false "Start time (RFC3339)"
// @Param end_time query string false "End time (RFC3339)"
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Success 200 {object} PaginatedResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/tracking/events [get]
// @Security BearerAuth
func (h *Handler) GetDriverEvents(c *gin.Context) {
	// Get company ID from JWT claims for authorization
	companyID, exists := c.Get("company_id")
	if !exists {
		middleware.AbortWithUnauthorized(c, "company ID not found in token")
		return
	}

	// Parse filters
	page := 1
	if pageStr := c.Query("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	limit := 20
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	// Build query
	query := h.service.db.Model(&models.DriverEvent{}).Where("company_id = ?", companyID)

	if driverID := c.Query("driver_id"); driverID != "" {
		query = query.Where("driver_id = ?", driverID)
	}
	if vehicleID := c.Query("vehicle_id"); vehicleID != "" {
		query = query.Where("vehicle_id = ?", vehicleID)
	}
	if eventType := c.Query("event_type"); eventType != "" {
		query = query.Where("event_type = ?", eventType)
	}
	if severity := c.Query("severity"); severity != "" {
		query = query.Where("severity = ?", severity)
	}
	if startTimeStr := c.Query("start_time"); startTimeStr != "" {
		if startTime, err := time.Parse(time.RFC3339, startTimeStr); err == nil {
			query = query.Where("timestamp >= ?", startTime)
		}
	}
	if endTimeStr := c.Query("end_time"); endTimeStr != "" {
		if endTime, err := time.Parse(time.RFC3339, endTimeStr); err == nil {
			query = query.Where("timestamp <= ?", endTime)
		}
	}

	// Get total count
	var total int64
	if err := query.Count(&total).Error; err != nil {
		middleware.AbortWithInternal(c, "failed to count driver events", err)
		return
	}

	// Get events with pagination
	var events []models.DriverEvent
	offset := (page - 1) * limit
	if err := query.Order("timestamp DESC").Offset(offset).Limit(limit).Find(&events).Error; err != nil {
		middleware.AbortWithInternal(c, "failed to get driver events", err)
		return
	}

	// Calculate pagination metadata
	totalPages := int((total + int64(limit) - 1) / int64(limit))
	hasNext := page < totalPages
	hasPrevious := page > 1

	c.JSON(http.StatusOK, PaginatedResponse{
		Success: true,
		Data:    events,
		Meta: Meta{
			Total:       total,
			Page:        page,
			Limit:       limit,
			TotalPages:  totalPages,
			HasNext:     hasNext,
			HasPrevious: hasPrevious,
		},
	})
}

// StartTrip godoc
// @Summary Start a trip
// @Description Start a new trip for a vehicle and driver
// @Tags tracking
// @Accept json
// @Produce json
// @Param trip body TripRequest true "Trip data"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/tracking/trips [post]
// @Security BearerAuth
func (h *Handler) StartTrip(c *gin.Context) {
	var req TripRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.AbortWithBadRequest(c, "invalid request data")
		return
	}

	// Validate request
	if err := h.validator.Struct(&req); err != nil {
		middleware.AbortWithValidation(c, err.Error())
		return
	}

	// Get company ID from JWT claims for authorization
	companyID, exists := c.Get("company_id")
	if !exists {
		middleware.AbortWithUnauthorized(c, "company ID not found in token")
		return
	}

	// Verify vehicle belongs to company
	var vehicle models.Vehicle
	if err := h.service.db.Where("id = ? AND company_id = ?", req.VehicleID, companyID).First(&vehicle).Error; err != nil {
		if std_errors.Is(err, gorm.ErrRecordNotFound) {
			middleware.AbortWithNotFound(c, "vehicle")
		} else {
			middleware.AbortWithInternal(c, "failed to verify vehicle", err)
		}
		return
	}

	// Process trip based on action
	var trip *models.Trip
	var err error

	switch req.Action {
	case "start":
		trip, err = h.service.StartTrip(req)
	case "end":
		trip, err = h.service.EndTrip(req)
	default:
		middleware.AbortWithBadRequest(c, "action must be 'start' or 'end'")
		return
	}

	if err != nil {
		if appErr, ok := err.(*apperrors.AppError); ok {
			middleware.AbortWithError(c, appErr)
		} else {
			middleware.AbortWithInternal(c, "trip operation failed", err)
		}
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Success: true,
		Data:    trip,
		Message: "Trip " + req.Action + " successful",
	})
}

// GetTrips godoc
// @Summary Get trip history
// @Description Get trip history for vehicles with filtering
// @Tags tracking
// @Produce json
// @Param driver_id query string false "Driver ID"
// @Param vehicle_id query string false "Vehicle ID"
// @Param status query string false "Trip status"
// @Param start_time query string false "Start time (RFC3339)"
// @Param end_time query string false "End time (RFC3339)"
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Success 200 {object} PaginatedResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/tracking/trips [get]
// @Security BearerAuth
func (h *Handler) GetTrips(c *gin.Context) {
	// Get company ID from JWT claims for authorization
	companyID, exists := c.Get("company_id")
	if !exists {
		middleware.AbortWithUnauthorized(c, "company ID not found in token")
		return
	}

	// Parse filters
	page := 1
	if pageStr := c.Query("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	limit := 20
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	// Build query
	query := h.service.db.Model(&models.Trip{}).Joins("JOIN vehicles ON trips.vehicle_id = vehicles.id").Where("vehicles.company_id = ?", companyID)

	if driverID := c.Query("driver_id"); driverID != "" {
		query = query.Where("trips.driver_id = ?", driverID)
	}
	if vehicleID := c.Query("vehicle_id"); vehicleID != "" {
		query = query.Where("trips.vehicle_id = ?", vehicleID)
	}
	if status := c.Query("status"); status != "" {
		query = query.Where("trips.status = ?", status)
	}
	if startTimeStr := c.Query("start_time"); startTimeStr != "" {
		if startTime, err := time.Parse(time.RFC3339, startTimeStr); err == nil {
			query = query.Where("trips.start_time >= ?", startTime)
		}
	}
	if endTimeStr := c.Query("end_time"); endTimeStr != "" {
		if endTime, err := time.Parse(time.RFC3339, endTimeStr); err == nil {
			query = query.Where("trips.start_time <= ?", endTime)
		}
	}

	// Get total count
	var total int64
	if err := query.Count(&total).Error; err != nil {
		middleware.AbortWithInternal(c, "failed to count trips", err)
		return
	}

	// Get trips with pagination
	var trips []models.Trip
	offset := (page - 1) * limit
	if err := query.Order("trips.start_time DESC").Offset(offset).Limit(limit).Find(&trips).Error; err != nil {
		middleware.AbortWithInternal(c, "failed to get trips", err)
		return
	}

	// Calculate pagination metadata
	totalPages := int((total + int64(limit) - 1) / int64(limit))
	hasNext := page < totalPages
	hasPrevious := page > 1

	c.JSON(http.StatusOK, PaginatedResponse{
		Success: true,
		Data:    trips,
		Meta: Meta{
			Total:       total,
			Page:        page,
			Limit:       limit,
			TotalPages:  totalPages,
			HasNext:     hasNext,
			HasPrevious: hasPrevious,
		},
	})
}

// CreateGeofence godoc
// @Summary Create geofence
// @Description Create a new geofence for monitoring
// @Tags tracking
// @Accept json
// @Produce json
// @Param geofence body GeofenceRequest true "Geofence data"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/tracking/geofences [post]
// @Security BearerAuth
func (h *Handler) CreateGeofence(c *gin.Context) {
	var req GeofenceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.AbortWithBadRequest(c, "invalid request data")
		return
	}

	// Validate request
	if err := h.validator.Struct(&req); err != nil {
		middleware.AbortWithValidation(c, err.Error())
		return
	}

	// Get company ID from JWT claims
	companyID, exists := c.Get("company_id")
	if !exists {
		middleware.AbortWithUnauthorized(c, "company ID not found in token")
		return
	}

	// Set company ID from JWT
	req.CompanyID = companyID.(string)

	// Create geofence
	geofence, err := h.service.CreateGeofence(req)
	if err != nil {
		if appErr, ok := err.(*apperrors.AppError); ok {
			middleware.AbortWithError(c, appErr)
		} else {
			middleware.AbortWithInternal(c, "failed to create geofence", err)
		}
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Success: true,
		Data:    geofence,
		Message: "Geofence created successfully",
	})
}

// GetGeofences godoc
// @Summary Get geofences
// @Description Get list of geofences for the company
// @Tags tracking
// @Produce json
// @Param type query string false "Geofence type"
// @Param is_active query bool false "Active status"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/tracking/geofences [get]
// @Security BearerAuth
func (h *Handler) GetGeofences(c *gin.Context) {
	// Get company ID from JWT claims
	companyID, exists := c.Get("company_id")
	if !exists {
		middleware.AbortWithUnauthorized(c, "company ID not found in token")
		return
	}

	// Build query
	query := h.service.db.Where("company_id = ?", companyID)

	if geofenceType := c.Query("type"); geofenceType != "" {
		query = query.Where("type = ?", geofenceType)
	}
	if isActiveStr := c.Query("is_active"); isActiveStr != "" {
		if isActive, err := strconv.ParseBool(isActiveStr); err == nil {
			query = query.Where("is_active = ?", isActive)
		}
	}

	// Get geofences
	var geofences []models.Geofence
	if err := query.Order("created_at DESC").Find(&geofences).Error; err != nil {
		middleware.AbortWithInternal(c, "failed to get geofences", err)
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Success: true,
		Data:    geofences,
	})
}

// UpdateGeofence godoc
// @Summary Update geofence
// @Description Update an existing geofence
// @Tags tracking
// @Accept json
// @Produce json
// @Param id path string true "Geofence ID"
// @Param geofence body GeofenceRequest true "Geofence update data"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/tracking/geofences/{id} [put]
// @Security BearerAuth
func (h *Handler) UpdateGeofence(c *gin.Context) {
	geofenceID := c.Param("id")
	if geofenceID == "" {
		middleware.AbortWithBadRequest(c, "geofence ID is required")
		return
	}

	var req GeofenceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.AbortWithBadRequest(c, "invalid request data")
		return
	}

	// Get company ID from JWT claims
	companyID, exists := c.Get("company_id")
	if !exists {
		middleware.AbortWithUnauthorized(c, "company ID not found in token")
		return
	}

	// Verify geofence belongs to company
	var geofence models.Geofence
	if err := h.service.db.Where("id = ? AND company_id = ?", geofenceID, companyID).First(&geofence).Error; err != nil {
		if std_errors.Is(err, gorm.ErrRecordNotFound) {
			middleware.AbortWithNotFound(c, "geofence")
		} else {
			middleware.AbortWithInternal(c, "failed to verify geofence", err)
		}
		return
	}

	// Update geofence
	if req.Name != "" {
		geofence.Name = req.Name
	}
	if req.Type != "" {
		geofence.Type = req.Type
	}
	if req.CenterLat != 0 {
		geofence.CenterLatitude = req.CenterLat
	}
	if req.CenterLng != 0 {
		geofence.CenterLongitude = req.CenterLng
	}
	if req.Radius != 0 {
		geofence.Radius = req.Radius
	}
	geofence.AlertOnEnter = req.AlertOnEntry
	geofence.AlertOnExit = req.AlertOnExit
	geofence.IsActive = req.IsActive
	if req.Description != "" {
		geofence.Description = req.Description
	}

	// Save changes
	if err := h.service.db.Save(&geofence).Error; err != nil {
		middleware.AbortWithInternal(c, "failed to update geofence", err)
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Success: true,
		Data:    geofence,
		Message: "Geofence updated successfully",
	})
}

// DeleteGeofence godoc
// @Summary Delete geofence
// @Description Delete a geofence
// @Tags tracking
// @Produce json
// @Param id path string true "Geofence ID"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/tracking/geofences/{id} [delete]
// @Security BearerAuth
func (h *Handler) DeleteGeofence(c *gin.Context) {
	geofenceID := c.Param("id")
	if geofenceID == "" {
		middleware.AbortWithBadRequest(c, "geofence ID is required")
		return
	}

	// Get company ID from JWT claims
	companyID, exists := c.Get("company_id")
	if !exists {
		middleware.AbortWithUnauthorized(c, "company ID not found in token")
		return
	}

	// Verify geofence belongs to company and delete
	result := h.service.db.Where("id = ? AND company_id = ?", geofenceID, companyID).Delete(&models.Geofence{})
	if result.Error != nil {
		middleware.AbortWithInternal(c, "failed to delete geofence", result.Error)
		return
	}

	if result.RowsAffected == 0 {
		middleware.AbortWithNotFound(c, "geofence")
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Success: true,
		Message: "Geofence deleted successfully",
	})
}

// HandleWebSocket godoc
// @Summary WebSocket connection for real-time tracking
// @Description WebSocket endpoint for real-time GPS tracking updates
// @Tags tracking
// @Param vehicle_id path string true "Vehicle ID"
// @Success 101 "Switching Protocols"
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /api/v1/tracking/ws/{vehicle_id} [get]
// @Security BearerAuth
func (h *Handler) HandleWebSocket(c *gin.Context) {
	// Use the service's WebSocket handler
	h.service.HandleWebSocket(c)
}

// GetDashboardStats godoc
// @Summary Get dashboard statistics
// @Description Get dashboard statistics for tracking
// @Tags tracking
// @Produce json
// @Success 200 {object} SuccessResponse
// @Failure 401 {object} ErrorResponse
// @Router /api/v1/tracking/dashboard/stats [get]
// @Security BearerAuth
func (h *Handler) GetDashboardStats(c *gin.Context) {
	// Get company ID from JWT claims
	companyID, exists := c.Get("company_id")
	if !exists {
		middleware.AbortWithUnauthorized(c, "company ID not found in token")
		return
	}

	// TODO: Implement dashboard statistics
	stats := map[string]interface{}{
		"company_id":         companyID,
		"active_vehicles":    0,
		"total_trips":        0,
		"distance_traveled":  0,
		"fuel_consumed":      0,
		"driver_events":      0,
		"geofence_violations": 0,
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Success: true,
		Data:    stats,
	})
}

// GetFuelConsumption godoc
// @Summary Get fuel consumption analytics
// @Description Get fuel consumption analytics and reports
// @Tags tracking
// @Produce json
// @Param start_date query string false "Start date"
// @Param end_date query string false "End date"
// @Success 200 {object} SuccessResponse
// @Failure 401 {object} ErrorResponse
// @Router /api/v1/tracking/analytics/fuel [get]
// @Security BearerAuth
func (h *Handler) GetFuelConsumption(c *gin.Context) {
	// Get company ID from JWT claims
	companyID, exists := c.Get("company_id")
	if !exists {
		middleware.AbortWithUnauthorized(c, "company ID not found in token")
		return
	}

	// TODO: Implement fuel consumption analytics
	analytics := map[string]interface{}{
		"company_id": companyID,
		"total_consumed": 0,
		"average_efficiency": 0,
		"cost_savings": 0,
		"trends": []interface{}{},
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Success: true,
		Data:    analytics,
	})
}

// GetDriverPerformance godoc
// @Summary Get driver performance analytics
// @Description Get driver performance analytics and scoring
// @Tags tracking
// @Produce json
// @Param driver_id query string false "Driver ID"
// @Param period query string false "Period (daily, weekly, monthly)"
// @Success 200 {object} SuccessResponse
// @Failure 401 {object} ErrorResponse
// @Router /api/v1/tracking/analytics/drivers [get]
// @Security BearerAuth
func (h *Handler) GetDriverPerformance(c *gin.Context) {
	// Get company ID from JWT claims
	companyID, exists := c.Get("company_id")
	if !exists {
		middleware.AbortWithUnauthorized(c, "company ID not found in token")
		return
	}

	// TODO: Implement driver performance analytics
	performance := map[string]interface{}{
		"company_id": companyID,
		"average_score": 0,
		"top_drivers": []interface{}{},
		"improvement_areas": []interface{}{},
		"trends": []interface{}{},
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Success: true,
		Data:    performance,
	})
}

// GenerateReport godoc
// @Summary Generate tracking report
// @Description Generate comprehensive tracking report
// @Tags tracking
// @Produce json
// @Param type query string false "Report type"
// @Param format query string false "Report format (json, csv, pdf)"
// @Success 200 {object} SuccessResponse
// @Failure 401 {object} ErrorResponse
// @Router /api/v1/tracking/reports/generate [post]
// @Security BearerAuth
func (h *Handler) GenerateReport(c *gin.Context) {
	// Get company ID from JWT claims
	companyID, exists := c.Get("company_id")
	if !exists {
		middleware.AbortWithUnauthorized(c, "company ID not found in token")
		return
	}

	// TODO: Implement report generation
	report := map[string]interface{}{
		"company_id": companyID,
		"report_id": "temp-report-id",
		"status": "generating",
		"estimated_completion": time.Now().Add(5 * time.Minute),
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Success: true,
		Data:    report,
	})
}

// GetComplianceReport godoc
// @Summary Get compliance report
// @Description Get regulatory compliance report for tracking
// @Tags tracking
// @Produce json
// @Param period query string false "Report period"
// @Success 200 {object} SuccessResponse
// @Failure 401 {object} ErrorResponse
// @Router /api/v1/tracking/reports/compliance [get]
// @Security BearerAuth
func (h *Handler) GetComplianceReport(c *gin.Context) {
	// Get company ID from JWT claims
	companyID, exists := c.Get("company_id")
	if !exists {
		middleware.AbortWithUnauthorized(c, "company ID not found in token")
		return
	}

	// TODO: Implement compliance report
	compliance := map[string]interface{}{
		"company_id": companyID,
		"compliance_score": 100,
		"violations": []interface{}{},
		"recommendations": []interface{}{},
		"next_audit": time.Now().Add(30 * 24 * time.Hour),
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Success: true,
		Data:    compliance,
	})
}