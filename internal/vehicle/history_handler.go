package vehicle

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/tobangado69/fleettracker-pro/backend/internal/common/repository"
)

// VehicleHistoryHandler handles vehicle history HTTP requests
type VehicleHistoryHandler struct {
	service *VehicleHistoryService
}

// NewVehicleHistoryHandler creates a new vehicle history handler
func NewVehicleHistoryHandler(service *VehicleHistoryService) *VehicleHistoryHandler {
	return &VehicleHistoryHandler{
		service: service,
	}
}

// GetVehicleHistory handles GET /api/v1/vehicles/:id/history
func (h *VehicleHistoryHandler) GetVehicleHistory(c *gin.Context) {
	vehicleID := c.Param("id")
	companyID := c.GetString("company_id")
	
	// Parse query parameters
	filters := HistoryFilters{
		Page:      1,
		Limit:     20,
		SortBy:    "created_at",
		SortOrder: "desc",
	}
	
	// Parse pagination
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
	
	// Parse filters
	if eventType := c.Query("event_type"); eventType != "" {
		filters.EventType = &eventType
	}
	
	if eventCategory := c.Query("event_category"); eventCategory != "" {
		filters.EventCategory = &eventCategory
	}
	
	if serviceProvider := c.Query("service_provider"); serviceProvider != "" {
		filters.ServiceProvider = &serviceProvider
	}
	
	if search := c.Query("search"); search != "" {
		filters.Search = &search
	}
	
	if sortBy := c.Query("sort_by"); sortBy != "" {
		filters.SortBy = sortBy
	}
	
	if sortOrder := c.Query("sort_order"); sortOrder != "" {
		filters.SortOrder = sortOrder
	}
	
	// Parse date range
	if startDateStr := c.Query("start_date"); startDateStr != "" {
		if startDate, err := time.Parse("2006-01-02", startDateStr); err == nil {
			filters.StartDate = &startDate
		}
	}
	
	if endDateStr := c.Query("end_date"); endDateStr != "" {
		if endDate, err := time.Parse("2006-01-02", endDateStr); err == nil {
			filters.EndDate = &endDate
		}
	}
	
	// Parse cost range
	if minCostStr := c.Query("min_cost"); minCostStr != "" {
		if minCost, err := strconv.ParseFloat(minCostStr, 64); err == nil {
			filters.MinCost = &minCost
		}
	}
	
	if maxCostStr := c.Query("max_cost"); maxCostStr != "" {
		if maxCost, err := strconv.ParseFloat(maxCostStr, 64); err == nil {
			filters.MaxCost = &maxCost
		}
	}
	
	// Get vehicle history
	histories, err := h.service.GetVehicleHistory(c.Request.Context(), companyID, vehicleID, filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve vehicle history",
			"details": err.Error(),
		})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": histories,
		"meta": gin.H{
			"page": filters.Page,
			"limit": filters.Limit,
			"total": len(histories),
		},
	})
}

// AddVehicleHistory handles POST /api/v1/vehicles/:id/history
func (h *VehicleHistoryHandler) AddVehicleHistory(c *gin.Context) {
	vehicleID := c.Param("id")
	companyID := c.GetString("company_id")
	userID := c.GetString("user_id")
	
	var req AddHistoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request data",
			"details": err.Error(),
		})
		return
	}
	
	// Add vehicle history
	history, err := h.service.AddVehicleHistory(c.Request.Context(), companyID, vehicleID, userID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to add vehicle history",
			"details": err.Error(),
		})
		return
	}
	
	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data": history,
		"message": "Vehicle history added successfully",
	})
}

// GetVehicleHistoryByID handles GET /api/v1/vehicles/:id/history/:historyId
func (h *VehicleHistoryHandler) GetVehicleHistoryByID(c *gin.Context) {
	historyID := c.Param("historyId")
	companyID := c.GetString("company_id")
	
	// Get vehicle history by ID
	history, err := h.service.GetVehicleHistoryByID(c.Request.Context(), companyID, historyID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Vehicle history not found",
			"details": err.Error(),
		})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": history,
	})
}

// UpdateVehicleHistory handles PUT /api/v1/vehicles/:id/history/:historyId
func (h *VehicleHistoryHandler) UpdateVehicleHistory(c *gin.Context) {
	historyID := c.Param("historyId")
	companyID := c.GetString("company_id")
	userID := c.GetString("user_id")
	
	var req AddHistoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request data",
			"details": err.Error(),
		})
		return
	}
	
	// Update vehicle history
	history, err := h.service.UpdateVehicleHistory(c.Request.Context(), companyID, historyID, userID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to update vehicle history",
			"details": err.Error(),
		})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": history,
		"message": "Vehicle history updated successfully",
	})
}

// DeleteVehicleHistory handles DELETE /api/v1/vehicles/:id/history/:historyId
func (h *VehicleHistoryHandler) DeleteVehicleHistory(c *gin.Context) {
	historyID := c.Param("historyId")
	companyID := c.GetString("company_id")
	
	// Delete vehicle history
	if err := h.service.DeleteVehicleHistory(c.Request.Context(), companyID, historyID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to delete vehicle history",
			"details": err.Error(),
		})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Vehicle history deleted successfully",
	})
}

// GetMaintenanceHistory handles GET /api/v1/vehicles/:id/maintenance
func (h *VehicleHistoryHandler) GetMaintenanceHistory(c *gin.Context) {
	vehicleID := c.Param("id")
	companyID := c.GetString("company_id")
	
	// Parse pagination
	page := 1
	limit := 20
	
	if pageStr := c.Query("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}
	
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}
	
	pagination := repository.Pagination{
		Page:     page,
		PageSize: limit,
		Offset:   (page - 1) * limit,
		Limit:    limit,
	}
	
	// Get maintenance history
	histories, err := h.service.GetMaintenanceHistory(c.Request.Context(), companyID, vehicleID, pagination)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve maintenance history",
			"details": err.Error(),
		})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": histories,
		"meta": gin.H{
			"page": page,
			"limit": limit,
			"total": len(histories),
		},
	})
}

// GetUpcomingMaintenance handles GET /api/v1/vehicles/maintenance/upcoming
func (h *VehicleHistoryHandler) GetUpcomingMaintenance(c *gin.Context) {
	companyID := c.GetString("company_id")
	
	// Parse days parameter
	days := 30
	if daysStr := c.Query("days"); daysStr != "" {
		if d, err := strconv.Atoi(daysStr); err == nil && d > 0 {
			days = d
		}
	}
	
	// Get upcoming maintenance
	histories, err := h.service.GetUpcomingMaintenance(c.Request.Context(), companyID, days)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve upcoming maintenance",
			"details": err.Error(),
		})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": histories,
		"meta": gin.H{
			"days": days,
			"total": len(histories),
		},
	})
}

// GetOverdueMaintenance handles GET /api/v1/vehicles/maintenance/overdue
func (h *VehicleHistoryHandler) GetOverdueMaintenance(c *gin.Context) {
	companyID := c.GetString("company_id")
	
	// Get overdue maintenance
	histories, err := h.service.GetOverdueMaintenance(c.Request.Context(), companyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve overdue maintenance",
			"details": err.Error(),
		})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": histories,
		"meta": gin.H{
			"total": len(histories),
		},
	})
}

// UpdateMaintenanceSchedule handles PUT /api/v1/vehicles/:id/maintenance/:historyId/schedule
func (h *VehicleHistoryHandler) UpdateMaintenanceSchedule(c *gin.Context) {
	historyID := c.Param("historyId")
	companyID := c.GetString("company_id")
	
	var req MaintenanceScheduleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request data",
			"details": err.Error(),
		})
		return
	}
	
	// Update maintenance schedule
	if err := h.service.UpdateMaintenanceSchedule(c.Request.Context(), companyID, historyID, req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to update maintenance schedule",
			"details": err.Error(),
		})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Maintenance schedule updated successfully",
	})
}

// GetCostSummary handles GET /api/v1/vehicles/:id/costs
func (h *VehicleHistoryHandler) GetCostSummary(c *gin.Context) {
	vehicleID := c.Param("id")
	companyID := c.GetString("company_id")
	
	// Parse date range (default to last 12 months)
	endDate := time.Now()
	startDate := endDate.AddDate(0, -12, 0)
	
	if startDateStr := c.Query("start_date"); startDateStr != "" {
		if sd, err := time.Parse("2006-01-02", startDateStr); err == nil {
			startDate = sd
		}
	}
	
	if endDateStr := c.Query("end_date"); endDateStr != "" {
		if ed, err := time.Parse("2006-01-02", endDateStr); err == nil {
			endDate = ed
		}
	}
	
	// Get cost summary
	summary, err := h.service.GetCostSummary(c.Request.Context(), companyID, vehicleID, startDate, endDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve cost summary",
			"details": err.Error(),
		})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": summary,
		"meta": gin.H{
			"start_date": startDate.Format("2006-01-02"),
			"end_date": endDate.Format("2006-01-02"),
		},
	})
}

// GetMaintenanceTrends handles GET /api/v1/vehicles/:id/trends
func (h *VehicleHistoryHandler) GetMaintenanceTrends(c *gin.Context) {
	vehicleID := c.Param("id")
	companyID := c.GetString("company_id")
	
	// Parse months parameter (default to 12 months)
	months := 12
	if monthsStr := c.Query("months"); monthsStr != "" {
		if m, err := strconv.Atoi(monthsStr); err == nil && m > 0 {
			months = m
		}
	}
	
	// Get maintenance trends
	trends, err := h.service.GetMaintenanceTrends(c.Request.Context(), companyID, vehicleID, months)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve maintenance trends",
			"details": err.Error(),
		})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": trends,
		"meta": gin.H{
			"months": months,
			"total": len(trends),
		},
	})
}
