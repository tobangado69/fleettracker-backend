package export

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/tobangado69/fleettracker-pro/backend/internal/common/middleware"
)

// ExportAPI provides HTTP API for data export operations
type ExportAPI struct {
	exportService *ExportService
}

// NewExportAPI creates a new export API
func NewExportAPI(exportService *ExportService) *ExportAPI {
	return &ExportAPI{
		exportService: exportService,
	}
}

// ExportDataHandler handles data export requests
func (ea *ExportAPI) ExportDataHandler(c *gin.Context) {
	var req ExportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.AbortWithBadRequest(c, err.Error())
		return
	}

	// Get user and company information
	userID, _ := c.Get("user_id")
	companyID, _ := c.Get("company_id")

	req.UserID = userID.(string)
	req.CompanyID = companyID.(string)

	// Validate export type
	if !ea.isValidExportType(req.ExportType) {
		middleware.AbortWithBadRequest(c, "Invalid export type")
		return
	}

	// Validate format
	if !ea.isValidFormat(req.Format) {
		middleware.AbortWithBadRequest(c, "Invalid format")
		return
	}

	// Export data
	response, err := ea.exportService.ExportData(c.Request.Context(), &req)
	if err != nil {
		middleware.AbortWithInternal(c, "Export operation failed", err)
		return
	}

	c.JSON(http.StatusOK, response)
}

// ExportVehiclesHandler handles vehicle export requests
func (ea *ExportAPI) ExportVehiclesHandler(c *gin.Context) {
	format := c.DefaultQuery("format", "json")
	
	// Get user and company information
	userID, _ := c.Get("user_id")
	companyID, _ := c.Get("company_id")

	// Build filters from query parameters
	filters := make(map[string]interface{})
	if status := c.Query("status"); status != "" {
		filters["status"] = status
	}
	if vehicleType := c.Query("vehicle_type"); vehicleType != "" {
		filters["vehicle_type"] = vehicleType
	}
	if year := c.Query("year"); year != "" {
		if yearInt, err := strconv.Atoi(year); err == nil {
			filters["year"] = float64(yearInt)
		}
	}

	req := &ExportRequest{
		ExportType: "vehicles",
		Format:     format,
		Filters:    filters,
		CompanyID:  companyID.(string),
		UserID:     userID.(string),
	}

	// Export data
	response, err := ea.exportService.ExportData(c.Request.Context(), req)
	if err != nil {
		middleware.AbortWithInternal(c, "Export operation failed", err)
		return
	}

	c.JSON(http.StatusOK, response)
}

// ExportDriversHandler handles driver export requests
func (ea *ExportAPI) ExportDriversHandler(c *gin.Context) {
	format := c.DefaultQuery("format", "json")
	
	// Get user and company information
	userID, _ := c.Get("user_id")
	companyID, _ := c.Get("company_id")

	// Build filters from query parameters
	filters := make(map[string]interface{})
	if status := c.Query("status"); status != "" {
		filters["status"] = status
	}
	if licenseType := c.Query("license_type"); licenseType != "" {
		filters["license_type"] = licenseType
	}

	req := &ExportRequest{
		ExportType: "drivers",
		Format:     format,
		Filters:    filters,
		CompanyID:  companyID.(string),
		UserID:     userID.(string),
	}

	// Export data
	response, err := ea.exportService.ExportData(c.Request.Context(), req)
	if err != nil {
		middleware.AbortWithInternal(c, "Export operation failed", err)
		return
	}

	c.JSON(http.StatusOK, response)
}

// ExportTripsHandler handles trip export requests
func (ea *ExportAPI) ExportTripsHandler(c *gin.Context) {
	format := c.DefaultQuery("format", "json")
	
	// Get user and company information
	userID, _ := c.Get("user_id")
	companyID, _ := c.Get("company_id")

	// Build filters from query parameters
	filters := make(map[string]interface{})
	if startDate := c.Query("start_date"); startDate != "" {
		filters["start_date"] = startDate
	}
	if endDate := c.Query("end_date"); endDate != "" {
		filters["end_date"] = endDate
	}
	if status := c.Query("status"); status != "" {
		filters["status"] = status
	}
	if vehicleID := c.Query("vehicle_id"); vehicleID != "" {
		filters["vehicle_id"] = vehicleID
	}
	if driverID := c.Query("driver_id"); driverID != "" {
		filters["driver_id"] = driverID
	}

	req := &ExportRequest{
		ExportType: "trips",
		Format:     format,
		Filters:    filters,
		CompanyID:  companyID.(string),
		UserID:     userID.(string),
	}

	// Export data
	response, err := ea.exportService.ExportData(c.Request.Context(), req)
	if err != nil {
		middleware.AbortWithInternal(c, "Export operation failed", err)
		return
	}

	c.JSON(http.StatusOK, response)
}

// ExportGPSTracksHandler handles GPS track export requests
func (ea *ExportAPI) ExportGPSTracksHandler(c *gin.Context) {
	format := c.DefaultQuery("format", "json")
	
	// Get user and company information
	userID, _ := c.Get("user_id")
	companyID, _ := c.Get("company_id")

	// Build filters from query parameters
	filters := make(map[string]interface{})
	if startDate := c.Query("start_date"); startDate != "" {
		filters["start_date"] = startDate
	}
	if endDate := c.Query("end_date"); endDate != "" {
		filters["end_date"] = endDate
	}
	if vehicleID := c.Query("vehicle_id"); vehicleID != "" {
		filters["vehicle_id"] = vehicleID
	}
	if driverID := c.Query("driver_id"); driverID != "" {
		filters["driver_id"] = driverID
	}
	if limit := c.Query("limit"); limit != "" {
		if limitInt, err := strconv.Atoi(limit); err == nil {
			filters["limit"] = float64(limitInt)
		}
	}

	req := &ExportRequest{
		ExportType: "gps_tracks",
		Format:     format,
		Filters:    filters,
		CompanyID:  companyID.(string),
		UserID:     userID.(string),
	}

	// Export data
	response, err := ea.exportService.ExportData(c.Request.Context(), req)
	if err != nil {
		middleware.AbortWithInternal(c, "Export operation failed", err)
		return
	}

	c.JSON(http.StatusOK, response)
}

// InvalidateCacheHandler handles cache invalidation requests
func (ea *ExportAPI) InvalidateCacheHandler(c *gin.Context) {
	exportType := c.Param("type")
	format := c.Query("format")
	
	// Get user and company information
	userID, _ := c.Get("user_id")
	companyID, _ := c.Get("company_id")

	// Build filters from query parameters
	filters := make(map[string]interface{})
	for key, values := range c.Request.URL.Query() {
		if key != "format" && len(values) > 0 {
			filters[key] = values[0]
		}
	}

	err := ea.exportService.InvalidateCache(c.Request.Context(), exportType, format, filters, companyID.(string), userID.(string))
	if err != nil {
		middleware.AbortWithInternal(c, "Export operation failed", err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Cache invalidated successfully"})
}

// InvalidateCompanyCacheHandler handles company cache invalidation
func (ea *ExportAPI) InvalidateCompanyCacheHandler(c *gin.Context) {
	// Get company information
	companyID, _ := c.Get("company_id")

	err := ea.exportService.InvalidateCompanyCache(c.Request.Context(), companyID.(string))
	if err != nil {
		middleware.AbortWithInternal(c, "Export operation failed", err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Company cache invalidated successfully"})
}

// InvalidateUserCacheHandler handles user cache invalidation
func (ea *ExportAPI) InvalidateUserCacheHandler(c *gin.Context) {
	// Get user information
	userID, _ := c.Get("user_id")

	err := ea.exportService.InvalidateUserCache(c.Request.Context(), userID.(string))
	if err != nil {
		middleware.AbortWithInternal(c, "Export operation failed", err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User cache invalidated successfully"})
}

// GetCacheStatsHandler handles cache statistics requests
func (ea *ExportAPI) GetCacheStatsHandler(c *gin.Context) {
	stats, err := ea.exportService.GetCacheStats(c.Request.Context())
	if err != nil {
		middleware.AbortWithInternal(c, "Export operation failed", err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"cache_stats": stats})
}

// GetCacheHitRateHandler handles cache hit rate requests
func (ea *ExportAPI) GetCacheHitRateHandler(c *gin.Context) {
	hitRate, err := ea.exportService.GetCacheHitRate(c.Request.Context())
	if err != nil {
		middleware.AbortWithInternal(c, "Export operation failed", err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"cache_hit_rate": hitRate})
}

// CleanupExpiredCacheHandler handles expired cache cleanup
func (ea *ExportAPI) CleanupExpiredCacheHandler(c *gin.Context) {
	err := ea.exportService.CleanupExpiredCache(c.Request.Context())
	if err != nil {
		middleware.AbortWithInternal(c, "Export operation failed", err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Expired cache cleaned up successfully"})
}

// isValidExportType validates export type
func (ea *ExportAPI) isValidExportType(exportType string) bool {
	validTypes := []string{"vehicles", "drivers", "trips", "gps_tracks", "reports"}
	for _, validType := range validTypes {
		if exportType == validType {
			return true
		}
	}
	return false
}

// isValidFormat validates export format
func (ea *ExportAPI) isValidFormat(format string) bool {
	validFormats := []string{"json", "csv"}
	for _, validFormat := range validFormats {
		if format == validFormat {
			return true
		}
	}
	return false
}

// SetupExportRoutes sets up export API routes
func SetupExportRoutes(r *gin.RouterGroup, api *ExportAPI) {
	exports := r.Group("/exports")
	{
		// General export endpoint
		exports.POST("/data", api.ExportDataHandler)
		
		// Specific export endpoints
		exports.GET("/vehicles", api.ExportVehiclesHandler)
		exports.GET("/drivers", api.ExportDriversHandler)
		exports.GET("/trips", api.ExportTripsHandler)
		exports.GET("/gps-tracks", api.ExportGPSTracksHandler)
		
		// Cache management
		cache := exports.Group("/cache")
		{
			cache.DELETE("/:type", api.InvalidateCacheHandler)
			cache.DELETE("/company", api.InvalidateCompanyCacheHandler)
			cache.DELETE("/user", api.InvalidateUserCacheHandler)
			cache.GET("/stats", api.GetCacheStatsHandler)
			cache.GET("/hit-rate", api.GetCacheHitRateHandler)
			cache.POST("/cleanup", api.CleanupExpiredCacheHandler)
		}
	}
}
