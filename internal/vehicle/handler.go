package vehicle

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/tobangado69/fleettracker-pro/backend/internal/common/middleware"
	customValidators "github.com/tobangado69/fleettracker-pro/backend/internal/common/validators"
)

// Handler handles vehicle HTTP requests
type Handler struct {
	service   *Service
	validator *validator.Validate
}

// NewHandler creates a new vehicle handler
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

// CreateVehicle godoc
// @Summary Create a new vehicle
// @Description Create a new vehicle with Indonesian compliance validation
// @Tags vehicles
// @Accept json
// @Produce json
// @Param vehicle body CreateVehicleRequest true "Vehicle data"
// @Success 201 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/vehicles [post]
// @Security BearerAuth
func (h *Handler) CreateVehicle(c *gin.Context) {
	// Get company ID from JWT claims
	companyID, exists := c.Get("company_id")
	if !exists {
		middleware.AbortWithUnauthorized(c, "company ID not found in token")
		return
	}

	var req CreateVehicleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.AbortWithBadRequest(c, err.Error())
		return
	}

	// Validate request
	if err := h.validator.Struct(&req); err != nil {
		middleware.AbortWithValidation(c, err.Error())
		return
	}

	// Validate and format license plate
	if req.LicensePlate != "" {
		if err := customValidators.ValidatePlateNumber(req.LicensePlate); err != nil {
			middleware.AbortWithBadRequest(c, "Invalid license plate: "+err.Error())
			return
		}
		req.LicensePlate = customValidators.FormatPlateNumber(req.LicensePlate)
	}

	// Validate VIN
	if req.VIN != "" {
		if err := customValidators.ValidateVIN(req.VIN); err != nil {
			middleware.AbortWithBadRequest(c, "Invalid VIN: "+err.Error())
			return
		}
	}

	// Validate year
	if req.Year != 0 {
		if err := customValidators.ValidateVehicleYear(req.Year); err != nil {
			middleware.AbortWithBadRequest(c, err.Error())
			return
		}
	}

	// Validate fuel type
	if req.FuelType != "" {
		if err := customValidators.ValidateFuelType(req.FuelType); err != nil {
			middleware.AbortWithBadRequest(c, err.Error())
			return
		}
	}


	// Create vehicle
	vehicle, err := h.service.CreateVehicle(companyID.(string), req)
	if err != nil {
		middleware.AbortWithBadRequest(c, err.Error())
		return
	}

	c.JSON(http.StatusCreated, SuccessResponse{
		Success: true,
		Data:    vehicle,
		Message: "Vehicle created successfully",
	})
}

// GetVehicle godoc
// @Summary Get vehicle by ID
// @Description Get vehicle details by ID
// @Tags vehicles
// @Produce json
// @Param id path string true "Vehicle ID"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/vehicles/{id} [get]
// @Security BearerAuth
func (h *Handler) GetVehicle(c *gin.Context) {
	// Get company ID from JWT claims
	companyID, exists := c.Get("company_id")
	if !exists {
		middleware.AbortWithUnauthorized(c, "company ID not found in token")
		return
	}

	vehicleID := c.Param("id")
	if vehicleID == "" {
		middleware.AbortWithBadRequest(c, "vehicle ID is required")
		return
	}

	// Get vehicle
	vehicle, err := h.service.GetVehicle(companyID.(string), vehicleID)
	if err != nil {
		if err.Error() == "vehicle not found" {
			middleware.AbortWithNotFound(c, err.Error())
			return
		}
		middleware.AbortWithInternal(c, err.Error(), err)
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Success: true,
		Data:    vehicle,
	})
}

// UpdateVehicle godoc
// @Summary Update vehicle
// @Description Update vehicle information
// @Tags vehicles
// @Accept json
// @Produce json
// @Param id path string true "Vehicle ID"
// @Param vehicle body UpdateVehicleRequest true "Vehicle update data"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/vehicles/{id} [put]
// @Security BearerAuth
func (h *Handler) UpdateVehicle(c *gin.Context) {
	// Get company ID from JWT claims
	companyID, exists := c.Get("company_id")
	if !exists {
		middleware.AbortWithUnauthorized(c, "company ID not found in token")
		return
	}

	vehicleID := c.Param("id")
	if vehicleID == "" {
		middleware.AbortWithBadRequest(c, "vehicle ID is required")
		return
	}

	var req UpdateVehicleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.AbortWithBadRequest(c, err.Error())
		return
	}

	// Validate request
	if err := h.validator.Struct(&req); err != nil {
		middleware.AbortWithValidation(c, err.Error())
		return
	}

	// Validate license plate if provided
	if req.LicensePlate != nil && *req.LicensePlate != "" {
		if err := customValidators.ValidatePlateNumber(*req.LicensePlate); err != nil {
			middleware.AbortWithBadRequest(c, "Invalid license plate: "+err.Error())
			return
		}
		formatted := customValidators.FormatPlateNumber(*req.LicensePlate)
		req.LicensePlate = &formatted
	}

	// Validate fuel type if provided
	if req.FuelType != nil && *req.FuelType != "" {
		if err := customValidators.ValidateFuelType(*req.FuelType); err != nil {
			middleware.AbortWithBadRequest(c, err.Error())
			return
		}
	}

	// Update vehicle
	vehicle, err := h.service.UpdateVehicle(companyID.(string), vehicleID, req)
	if err != nil {
		if err.Error() == "vehicle not found" {
			middleware.AbortWithNotFound(c, err.Error())
			return
		}
		middleware.AbortWithBadRequest(c, err.Error())
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Success: true,
		Data:    vehicle,
		Message: "Vehicle updated successfully",
	})
}

// DeleteVehicle godoc
// @Summary Delete vehicle
// @Description Delete vehicle (soft delete)
// @Tags vehicles
// @Produce json
// @Param id path string true "Vehicle ID"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/vehicles/{id} [delete]
// @Security BearerAuth
func (h *Handler) DeleteVehicle(c *gin.Context) {
	// Get company ID from JWT claims
	companyID, exists := c.Get("company_id")
	if !exists {
		middleware.AbortWithUnauthorized(c, "company ID not found in token")
		return
	}

	vehicleID := c.Param("id")
	if vehicleID == "" {
		middleware.AbortWithBadRequest(c, "vehicle ID is required")
		return
	}

	// Delete vehicle
	err := h.service.DeleteVehicle(companyID.(string), vehicleID)
	if err != nil {
		if err.Error() == "vehicle not found" {
			middleware.AbortWithNotFound(c, err.Error())
			return
		}
		middleware.AbortWithBadRequest(c, err.Error())
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Success: true,
		Message: "Vehicle deleted successfully",
	})
}

// ListVehicles godoc
// @Summary List vehicles
// @Description List vehicles with filters and pagination
// @Tags vehicles
// @Produce json
// @Param status query string false "Vehicle status"
// @Param make query string false "Vehicle make"
// @Param model query string false "Vehicle model"
// @Param year query int false "Vehicle year"
// @Param fuel_type query string false "Fuel type"
// @Param has_driver query bool false "Has driver assigned"
// @Param gps_enabled query bool false "GPS enabled"
// @Param search query string false "Search term"
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Param sort_by query string false "Sort field" default(created_at)
// @Param sort_order query string false "Sort order" default(desc)
// @Success 200 {object} PaginatedResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/vehicles [get]
// @Security BearerAuth
func (h *Handler) ListVehicles(c *gin.Context) {
	// Get company ID from JWT claims
	companyID, exists := c.Get("company_id")
	if !exists {
		middleware.AbortWithUnauthorized(c, "company ID not found in token")
		return
	}

	// Parse filters
	filters := VehicleFilters{
		Page:      1,
		Limit:     20,
		SortBy:    "created_at",
		SortOrder: "desc",
	}

	if status := c.Query("status"); status != "" {
		filters.Status = &status
	}
	if make := c.Query("make"); make != "" {
		filters.Make = &make
	}
	if model := c.Query("model"); model != "" {
		filters.Model = &model
	}
	if yearStr := c.Query("year"); yearStr != "" {
		if year, err := strconv.Atoi(yearStr); err == nil {
			filters.Year = &year
		}
	}
	if fuelType := c.Query("fuel_type"); fuelType != "" {
		filters.FuelType = &fuelType
	}
	if hasDriverStr := c.Query("has_driver"); hasDriverStr != "" {
		if hasDriver, err := strconv.ParseBool(hasDriverStr); err == nil {
			filters.HasDriver = &hasDriver
		}
	}
	if gpsEnabledStr := c.Query("gps_enabled"); gpsEnabledStr != "" {
		if gpsEnabled, err := strconv.ParseBool(gpsEnabledStr); err == nil {
			filters.GPSEnabled = &gpsEnabled
		}
	}
	if search := c.Query("search"); search != "" {
		filters.Search = &search
	}
	if pageStr := c.Query("page"); pageStr != "" {
		if page, err := strconv.Atoi(pageStr); err == nil && page > 0 {
			filters.Page = page
		}
	}
	if limitStr := c.Query("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 && limit <= 100 {
			filters.Limit = limit
		}
	}
	if sortBy := c.Query("sort_by"); sortBy != "" {
		filters.SortBy = sortBy
	}
	if sortOrder := c.Query("sort_order"); sortOrder != "" {
		filters.SortOrder = sortOrder
	}

	// List vehicles
	vehicles, total, err := h.service.ListVehicles(companyID.(string), filters)
	if err != nil {
		middleware.AbortWithInternal(c, err.Error(), err)
		return
	}

	// Calculate pagination metadata
	totalPages := int((total + int64(filters.Limit) - 1) / int64(filters.Limit))
	hasNext := filters.Page < totalPages
	hasPrevious := filters.Page > 1

	c.JSON(http.StatusOK, PaginatedResponse{
		Success: true,
		Data:    vehicles,
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

// UpdateVehicleStatus godoc
// @Summary Update vehicle status
// @Description Update vehicle status (active, maintenance, retired, inactive)
// @Tags vehicles
// @Accept json
// @Produce json
// @Param id path string true "Vehicle ID"
// @Param status body object{status=string,reason=string} true "Status update data"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/vehicles/{id}/status [put]
// @Security BearerAuth
func (h *Handler) UpdateVehicleStatus(c *gin.Context) {
	// Get company ID from JWT claims
	companyID, exists := c.Get("company_id")
	if !exists {
		middleware.AbortWithUnauthorized(c, "company ID not found in token")
		return
	}

	vehicleID := c.Param("id")
	if vehicleID == "" {
		middleware.AbortWithBadRequest(c, "vehicle ID is required")
		return
	}

	var req struct {
		Status string `json:"status" binding:"required,oneof=active maintenance retired inactive"`
		Reason string `json:"reason"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.AbortWithBadRequest(c, err.Error())
		return
	}

	// Update vehicle status
	err := h.service.UpdateVehicleStatus(companyID.(string), vehicleID, VehicleStatus(req.Status), req.Reason)
	if err != nil {
		if err.Error() == "vehicle not found" {
			middleware.AbortWithNotFound(c, err.Error())
			return
		}
		middleware.AbortWithBadRequest(c, err.Error())
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Success: true,
		Message: "Vehicle status updated successfully",
	})
}

// AssignDriver godoc
// @Summary Assign driver to vehicle
// @Description Assign a driver to a vehicle
// @Tags vehicles
// @Accept json
// @Produce json
// @Param id path string true "Vehicle ID"
// @Param assignment body object{driver_id=string} true "Driver assignment data"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/vehicles/{id}/assign-driver [post]
// @Security BearerAuth
func (h *Handler) AssignDriver(c *gin.Context) {
	// Get company ID from JWT claims
	companyID, exists := c.Get("company_id")
	if !exists {
		middleware.AbortWithUnauthorized(c, "company ID not found in token")
		return
	}

	vehicleID := c.Param("id")
	if vehicleID == "" {
		middleware.AbortWithBadRequest(c, "vehicle ID is required")
		return
	}

	var req struct {
		DriverID string `json:"driver_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.AbortWithBadRequest(c, err.Error())
		return
	}

	// Assign driver
	err := h.service.AssignDriver(companyID.(string), vehicleID, req.DriverID)
	if err != nil {
		middleware.AbortWithBadRequest(c, err.Error())
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Success: true,
		Message: "Driver assigned successfully",
	})
}

// UnassignDriver godoc
// @Summary Unassign driver from vehicle
// @Description Remove driver assignment from vehicle
// @Tags vehicles
// @Produce json
// @Param id path string true "Vehicle ID"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/vehicles/{id}/driver [delete]
// @Security BearerAuth
func (h *Handler) UnassignDriver(c *gin.Context) {
	// Get company ID from JWT claims
	companyID, exists := c.Get("company_id")
	if !exists {
		middleware.AbortWithUnauthorized(c, "company ID not found in token")
		return
	}

	vehicleID := c.Param("id")
	if vehicleID == "" {
		middleware.AbortWithBadRequest(c, "vehicle ID is required")
		return
	}

	// Unassign driver
	err := h.service.UnassignDriver(companyID.(string), vehicleID)
	if err != nil {
		middleware.AbortWithBadRequest(c, err.Error())
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Success: true,
		Message: "Driver unassigned successfully",
	})
}

// GetVehicleDriver godoc
// @Summary Get vehicle driver
// @Description Get the driver assigned to a vehicle
// @Tags vehicles
// @Produce json
// @Param id path string true "Vehicle ID"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/vehicles/{id}/driver [get]
// @Security BearerAuth
func (h *Handler) GetVehicleDriver(c *gin.Context) {
	// Get company ID from JWT claims
	companyID, exists := c.Get("company_id")
	if !exists {
		middleware.AbortWithUnauthorized(c, "company ID not found in token")
		return
	}

	vehicleID := c.Param("id")
	if vehicleID == "" {
		middleware.AbortWithBadRequest(c, "vehicle ID is required")
		return
	}

	// Get vehicle driver
	driver, err := h.service.GetVehicleDriver(companyID.(string), vehicleID)
	if err != nil {
		if err.Error() == "no driver assigned to this vehicle" {
			middleware.AbortWithNotFound(c, err.Error())
			return
		}
		middleware.AbortWithInternal(c, err.Error(), err)
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Success: true,
		Data:    driver,
	})
}

// UpdateInspectionDate godoc
// @Summary Update vehicle inspection date
// @Description Update vehicle inspection date and calculate next inspection
// @Tags vehicles
// @Accept json
// @Produce json
// @Param id path string true "Vehicle ID"
// @Param inspection body object{inspection_date=string} true "Inspection date data"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/vehicles/{id}/inspection [put]
// @Security BearerAuth
func (h *Handler) UpdateInspectionDate(c *gin.Context) {
	// Get company ID from JWT claims
	companyID, exists := c.Get("company_id")
	if !exists {
		middleware.AbortWithUnauthorized(c, "company ID not found in token")
		return
	}

	vehicleID := c.Param("id")
	if vehicleID == "" {
		middleware.AbortWithBadRequest(c, "vehicle ID is required")
		return
	}

	var req struct {
		InspectionDate string `json:"inspection_date" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.AbortWithBadRequest(c, err.Error())
		return
	}

	// Parse inspection date
	inspectionDate, err := time.Parse("2006-01-02", req.InspectionDate)
	if err != nil {
		middleware.AbortWithBadRequest(c, "inspection date must be in YYYY-MM-DD format")
		return
	}

	// Update inspection date
	err = h.service.UpdateInspectionDate(companyID.(string), vehicleID, inspectionDate)
	if err != nil {
		if err.Error() == "vehicle not found" {
			middleware.AbortWithNotFound(c, err.Error())
			return
		}
		middleware.AbortWithBadRequest(c, err.Error())
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Success: true,
		Message: "Inspection date updated successfully",
	})
}

// GetVehicleStatus handles getting vehicle status (legacy method for compatibility)
func (h *Handler) GetVehicleStatus(c *gin.Context) {
	// Redirect to GetVehicle for now
	h.GetVehicle(c)
}

// GetVehicles handles getting vehicles (legacy method for compatibility)
func (h *Handler) GetVehicles(c *gin.Context) {
	// Redirect to ListVehicles for now
	h.ListVehicles(c)
}