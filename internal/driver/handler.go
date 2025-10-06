package driver

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/tobangado69/fleettracker-pro/backend/internal/common/middleware"
	"github.com/tobangado69/fleettracker-pro/backend/pkg/errors"
)

// Handler handles driver HTTP requests
type Handler struct {
	service   *Service
	validator *validator.Validate
}

// NewHandler creates a new driver handler
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

// CreateDriver godoc
// @Summary Create a new driver
// @Description Create a new driver with Indonesian compliance validation
// @Tags drivers
// @Accept json
// @Produce json
// @Param driver body CreateDriverRequest true "Driver data"
// @Success 201 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/drivers [post]
// @Security BearerAuth
func (h *Handler) CreateDriver(c *gin.Context) {
	// Get company ID from JWT claims
	companyID, exists := c.Get("company_id")
	if !exists {
		middleware.AbortWithUnauthorized(c, "company ID not found in token")
		return
	}

	var req CreateDriverRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.AbortWithBadRequest(c, "invalid request data")
		return
	}

	// Validate request
	if err := h.validator.Struct(&req); err != nil {
		middleware.AbortWithValidation(c, err.Error())
		return
	}

	// Create driver
	driver, err := h.service.CreateDriver(companyID.(string), req)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			middleware.AbortWithError(c, appErr)
		} else {
			middleware.AbortWithInternal(c, "failed to create driver", err)
		}
		return
	}

		c.JSON(http.StatusCreated, SuccessResponse{
			Success: true,
			Data:    driver,
			Message: "Driver created successfully",
		})
}

// GetDriver godoc
// @Summary Get driver by ID
// @Description Get driver details by ID
// @Tags drivers
// @Produce json
// @Param id path string true "Driver ID"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/drivers/{id} [get]
// @Security BearerAuth
func (h *Handler) GetDriver(c *gin.Context) {
	// Get company ID from JWT claims
	companyID, exists := c.Get("company_id")
	if !exists {
		middleware.AbortWithUnauthorized(c, "company ID not found in token")
		return
	}

	driverID := c.Param("id")
	if driverID == "" {
		middleware.AbortWithBadRequest(c, "driver ID is required")
		return
	}

	// Get driver
	driver, err := h.service.GetDriver(companyID.(string), driverID)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			middleware.AbortWithError(c, appErr)
		} else {
			middleware.AbortWithInternal(c, "failed to get driver", err)
		}
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Success: true,
		Data:    driver,
	})
}

// UpdateDriver godoc
// @Summary Update driver
// @Description Update driver information
// @Tags drivers
// @Accept json
// @Produce json
// @Param id path string true "Driver ID"
// @Param driver body UpdateDriverRequest true "Driver update data"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/drivers/{id} [put]
// @Security BearerAuth
func (h *Handler) UpdateDriver(c *gin.Context) {
	// Get company ID from JWT claims
	companyID, exists := c.Get("company_id")
	if !exists {
		middleware.AbortWithUnauthorized(c, "company ID not found in token")
		return
	}

	driverID := c.Param("id")
	if driverID == "" {
		middleware.AbortWithBadRequest(c, "driver ID is required")
		return
	}

	var req UpdateDriverRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.AbortWithBadRequest(c, "invalid request data")
		return
	}

	// Validate request
	if err := h.validator.Struct(&req); err != nil {
		middleware.AbortWithValidation(c, err.Error())
		return
	}

	// Update driver
	driver, err := h.service.UpdateDriver(companyID.(string), driverID, req)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			middleware.AbortWithError(c, appErr)
		} else {
			middleware.AbortWithInternal(c, "failed to update driver", err)
		}
		return
	}

		c.JSON(http.StatusOK, SuccessResponse{
			Success: true,
			Data:    driver,
			Message: "Driver updated successfully",
		})
}

// DeleteDriver godoc
// @Summary Delete driver
// @Description Delete driver (soft delete)
// @Tags drivers
// @Produce json
// @Param id path string true "Driver ID"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/drivers/{id} [delete]
// @Security BearerAuth
func (h *Handler) DeleteDriver(c *gin.Context) {
	// Get company ID from JWT claims
	companyID, exists := c.Get("company_id")
	if !exists {
		middleware.AbortWithUnauthorized(c, "company ID not found in token")
		return
	}

	driverID := c.Param("id")
	if driverID == "" {
		middleware.AbortWithBadRequest(c, "driver ID is required")
		return
	}

	// Delete driver
	err := h.service.DeleteDriver(companyID.(string), driverID)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			middleware.AbortWithError(c, appErr)
		} else {
			middleware.AbortWithInternal(c, "failed to delete driver", err)
		}
		return
	}

		c.JSON(http.StatusOK, SuccessResponse{
			Success: true,
			Message: "Driver deleted successfully",
		})
}

// ListDrivers godoc
// @Summary List drivers
// @Description List drivers with filters and pagination
// @Tags drivers
// @Produce json
// @Param status query string false "Driver status"
// @Param employment_status query string false "Employment status"
// @Param performance_grade query string false "Performance grade (A, B, C, D, F)"
// @Param city query string false "City"
// @Param province query string false "Province"
// @Param has_vehicle query bool false "Has vehicle assigned"
// @Param is_available query bool false "Is available"
// @Param is_compliant query bool false "Is compliant"
// @Param search query string false "Search term"
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Param sort_by query string false "Sort field" default(created_at)
// @Param sort_order query string false "Sort order" default(desc)
// @Success 200 {object} PaginatedResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/drivers [get]
// @Security BearerAuth
func (h *Handler) ListDrivers(c *gin.Context) {
	// Get company ID from JWT claims
	companyID, exists := c.Get("company_id")
	if !exists {
		middleware.AbortWithUnauthorized(c, "company ID not found in token")
		return
	}

	// Parse filters
	filters := DriverFilters{
		Page:      1,
		Limit:     20,
		SortBy:    "created_at",
		SortOrder: "desc",
	}

	if status := c.Query("status"); status != "" {
		filters.Status = &status
	}
	if employmentStatus := c.Query("employment_status"); employmentStatus != "" {
		filters.EmploymentStatus = &employmentStatus
	}
	if performanceGrade := c.Query("performance_grade"); performanceGrade != "" {
		filters.PerformanceGrade = &performanceGrade
	}
	if city := c.Query("city"); city != "" {
		filters.City = &city
	}
	if province := c.Query("province"); province != "" {
		filters.Province = &province
	}
	if hasVehicleStr := c.Query("has_vehicle"); hasVehicleStr != "" {
		if hasVehicle, err := strconv.ParseBool(hasVehicleStr); err == nil {
			filters.HasVehicle = &hasVehicle
		}
	}
	if isAvailableStr := c.Query("is_available"); isAvailableStr != "" {
		if isAvailable, err := strconv.ParseBool(isAvailableStr); err == nil {
			filters.IsAvailable = &isAvailable
		}
	}
	if isCompliantStr := c.Query("is_compliant"); isCompliantStr != "" {
		if isCompliant, err := strconv.ParseBool(isCompliantStr); err == nil {
			filters.IsCompliant = &isCompliant
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

	// List drivers
	drivers, total, err := h.service.ListDrivers(companyID.(string), filters)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			middleware.AbortWithError(c, appErr)
		} else {
			middleware.AbortWithInternal(c, "failed to list drivers", err)
		}
		return
	}

	// Calculate pagination metadata
	totalPages := int((total + int64(filters.Limit) - 1) / int64(filters.Limit))
	hasNext := filters.Page < totalPages
	hasPrevious := filters.Page > 1

	c.JSON(http.StatusOK, PaginatedResponse{
		Success: true,
		Data:    drivers,
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

// UpdateDriverStatus godoc
// @Summary Update driver status
// @Description Update driver status (available, busy, inactive, suspended, terminated)
// @Tags drivers
// @Accept json
// @Produce json
// @Param id path string true "Driver ID"
// @Param status body object{status=string,reason=string} true "Status update data"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/drivers/{id}/status [put]
// @Security BearerAuth
func (h *Handler) UpdateDriverStatus(c *gin.Context) {
	// Get company ID from JWT claims
	companyID, exists := c.Get("company_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "unauthorized",
			"message": "company ID not found in token",
		})
		return
	}

	driverID := c.Param("id")
	if driverID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": "driver ID is required",
		})
		return
	}

	var req struct {
		Status string `json:"status" binding:"required,oneof=available busy inactive suspended terminated"`
		Reason string `json:"reason"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": err.Error(),
		})
		return
	}

	// Update driver status
	err := h.service.UpdateDriverStatus(companyID.(string), driverID, DriverStatus(req.Status), req.Reason)
	if err != nil {
		if err.Error() == "driver not found" {
			c.JSON(http.StatusNotFound, gin.H{
				"error":   "not_found",
				"message": err.Error(),
			})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "update_failed",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Success: true,
		Message: "Driver status updated successfully",
	})
}

// GetDriverPerformance godoc
// @Summary Get driver performance
// @Description Get driver performance data and analytics
// @Tags drivers
// @Produce json
// @Param id path string true "Driver ID"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/drivers/{id}/performance [get]
// @Security BearerAuth
func (h *Handler) GetDriverPerformance(c *gin.Context) {
	// Get company ID from JWT claims
	companyID, exists := c.Get("company_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "unauthorized",
			"message": "company ID not found in token",
		})
		return
	}

	driverID := c.Param("id")
	if driverID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": "driver ID is required",
		})
		return
	}

	// Get driver performance
	driver, err := h.service.GetDriverPerformance(companyID.(string), driverID)
	if err != nil {
		if err.Error() == "driver not found" {
			c.JSON(http.StatusNotFound, gin.H{
				"error":   "not_found",
				"message": err.Error(),
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "internal_error",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Success: true,
		Data:    driver,
	})
}

// UpdateDriverPerformance godoc
// @Summary Update driver performance
// @Description Update driver performance scores
// @Tags drivers
// @Accept json
// @Produce json
// @Param id path string true "Driver ID"
// @Param performance body object{performance_score=number,safety_score=number,efficiency_score=number} true "Performance data"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/drivers/{id}/performance [put]
// @Security BearerAuth
func (h *Handler) UpdateDriverPerformance(c *gin.Context) {
	// Get company ID from JWT claims
	companyID, exists := c.Get("company_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "unauthorized",
			"message": "company ID not found in token",
		})
		return
	}

	driverID := c.Param("id")
	if driverID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": "driver ID is required",
		})
		return
	}

	var req struct {
		PerformanceScore float64 `json:"performance_score" binding:"required,min=0,max=100"`
		SafetyScore      float64 `json:"safety_score" binding:"required,min=0,max=100"`
		EfficiencyScore  float64 `json:"efficiency_score" binding:"required,min=0,max=100"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": err.Error(),
		})
		return
	}

	// Update driver performance
	err := h.service.UpdateDriverPerformance(companyID.(string), driverID, req.PerformanceScore, req.SafetyScore, req.EfficiencyScore)
	if err != nil {
		if err.Error() == "driver not found" {
			c.JSON(http.StatusNotFound, gin.H{
				"error":   "not_found",
				"message": err.Error(),
			})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "update_failed",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Success: true,
		Message: "Driver performance updated successfully",
	})
}

// AssignVehicle godoc
// @Summary Assign vehicle to driver
// @Description Assign a vehicle to a driver
// @Tags drivers
// @Accept json
// @Produce json
// @Param id path string true "Driver ID"
// @Param assignment body object{vehicle_id=string} true "Vehicle assignment data"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/drivers/{id}/assign-vehicle [post]
// @Security BearerAuth
func (h *Handler) AssignVehicle(c *gin.Context) {
	// Get company ID from JWT claims
	companyID, exists := c.Get("company_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "unauthorized",
			"message": "company ID not found in token",
		})
		return
	}

	driverID := c.Param("id")
	if driverID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": "driver ID is required",
		})
		return
	}

	var req struct {
		VehicleID string `json:"vehicle_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": err.Error(),
		})
		return
	}

	// Assign vehicle
	err := h.service.AssignVehicle(companyID.(string), driverID, req.VehicleID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "assignment_failed",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Success: true,
		Message: "Vehicle assigned successfully",
	})
}

// UnassignVehicle godoc
// @Summary Unassign vehicle from driver
// @Description Remove vehicle assignment from driver
// @Tags drivers
// @Produce json
// @Param id path string true "Driver ID"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/drivers/{id}/vehicle [delete]
// @Security BearerAuth
func (h *Handler) UnassignVehicle(c *gin.Context) {
	// Get company ID from JWT claims
	companyID, exists := c.Get("company_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "unauthorized",
			"message": "company ID not found in token",
		})
		return
	}

	driverID := c.Param("id")
	if driverID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": "driver ID is required",
		})
		return
	}

	// Unassign vehicle
	err := h.service.UnassignVehicle(companyID.(string), driverID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "unassignment_failed",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Success: true,
		Message: "Vehicle unassigned successfully",
	})
}

// GetDriverVehicle godoc
// @Summary Get driver vehicle
// @Description Get the vehicle assigned to a driver
// @Tags drivers
// @Produce json
// @Param id path string true "Driver ID"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/drivers/{id}/vehicle [get]
// @Security BearerAuth
func (h *Handler) GetDriverVehicle(c *gin.Context) {
	// Get company ID from JWT claims
	companyID, exists := c.Get("company_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "unauthorized",
			"message": "company ID not found in token",
		})
		return
	}

	driverID := c.Param("id")
	if driverID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": "driver ID is required",
		})
		return
	}

	// Get driver vehicle
	vehicle, err := h.service.GetDriverVehicle(companyID.(string), driverID)
	if err != nil {
		if err.Error() == "no vehicle assigned to this driver" {
			c.JSON(http.StatusNotFound, gin.H{
				"error":   "not_found",
				"message": err.Error(),
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "internal_error",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Success: true,
		Data:    vehicle,
	})
}

// UpdateMedicalCheckup godoc
// @Summary Update driver medical checkup
// @Description Update driver medical checkup date
// @Tags drivers
// @Accept json
// @Produce json
// @Param id path string true "Driver ID"
// @Param medical body object{medical_checkup_date=string} true "Medical checkup data"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/drivers/{id}/medical [put]
// @Security BearerAuth
func (h *Handler) UpdateMedicalCheckup(c *gin.Context) {
	// Get company ID from JWT claims
	companyID, exists := c.Get("company_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "unauthorized",
			"message": "company ID not found in token",
		})
		return
	}

	driverID := c.Param("id")
	if driverID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": "driver ID is required",
		})
		return
	}

	var req struct {
		MedicalCheckupDate string `json:"medical_checkup_date" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": err.Error(),
		})
		return
	}

	// Parse medical checkup date
	checkupDate, err := time.Parse("2006-01-02", req.MedicalCheckupDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_date",
			"message": "medical checkup date must be in YYYY-MM-DD format",
		})
		return
	}

	// Update medical checkup
	err = h.service.UpdateMedicalCheckup(companyID.(string), driverID, checkupDate)
	if err != nil {
		if err.Error() == "driver not found" {
			c.JSON(http.StatusNotFound, gin.H{
				"error":   "not_found",
				"message": err.Error(),
			})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "update_failed",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Success: true,
		Message: "Medical checkup updated successfully",
	})
}

// UpdateTrainingStatus godoc
// @Summary Update driver training status
// @Description Update driver training completion and expiry
// @Tags drivers
// @Accept json
// @Produce json
// @Param id path string true "Driver ID"
// @Param training body object{training_completed=bool,training_expiry=string} true "Training data"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/drivers/{id}/training [put]
// @Security BearerAuth
func (h *Handler) UpdateTrainingStatus(c *gin.Context) {
	// Get company ID from JWT claims
	companyID, exists := c.Get("company_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "unauthorized",
			"message": "company ID not found in token",
		})
		return
	}

	driverID := c.Param("id")
	if driverID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": "driver ID is required",
		})
		return
	}

	var req struct {
		TrainingCompleted bool   `json:"training_completed" binding:"required"`
		TrainingExpiry    string `json:"training_expiry"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": err.Error(),
		})
		return
	}

	var expiryDate *time.Time
	if req.TrainingExpiry != "" {
		parsedDate, err := time.Parse("2006-01-02", req.TrainingExpiry)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "invalid_date",
				"message": "training expiry date must be in YYYY-MM-DD format",
			})
			return
		}
		expiryDate = &parsedDate
	}

	// Update training status
	err := h.service.UpdateTrainingStatus(companyID.(string), driverID, req.TrainingCompleted, expiryDate)
	if err != nil {
		if err.Error() == "driver not found" {
			c.JSON(http.StatusNotFound, gin.H{
				"error":   "not_found",
				"message": err.Error(),
			})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "update_failed",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Success: true,
		Message: "Training status updated successfully",
	})
}

// GetDriverTrips handles getting driver trips (legacy method for compatibility)
func (h *Handler) GetDriverTrips(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "Driver trips not implemented yet",
	})
}

// GetDrivers handles getting drivers (legacy method for compatibility)
func (h *Handler) GetDrivers(c *gin.Context) {
	// Redirect to ListDrivers for now
	h.ListDrivers(c)
}