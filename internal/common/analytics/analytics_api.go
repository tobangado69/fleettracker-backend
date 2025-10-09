package analytics

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/tobangado69/fleettracker-pro/backend/internal/common/middleware"
)

// AnalyticsAPI provides HTTP API for analytics operations
type AnalyticsAPI struct {
	analyticsEngine *AnalyticsEngine
}

// NewAnalyticsAPI creates a new analytics API
func NewAnalyticsAPI(analyticsEngine *AnalyticsEngine) *AnalyticsAPI {
	return &AnalyticsAPI{
		analyticsEngine: analyticsEngine,
	}
}

// GenerateAnalyticsHandler handles analytics generation requests
func (api *AnalyticsAPI) GenerateAnalyticsHandler(c *gin.Context) {
	var req AnalyticsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.AbortWithBadRequest(c, err.Error())
		return
	}
	
	// Set company ID from context (from auth middleware)
	if companyID, exists := c.Get("company_id"); exists {
		req.CompanyID = companyID.(string)
	}
	
	// Set user ID from context (from auth middleware)
	if userID, exists := c.Get("user_id"); exists {
		req.UserID = userID.(string)
	}
	
	// Set default date range if not provided
	if req.DateRange.StartDate.IsZero() {
		req.DateRange.StartDate = time.Now().AddDate(0, 0, -30) // Last 30 days
	}
	if req.DateRange.EndDate.IsZero() {
		req.DateRange.EndDate = time.Now()
	}
	
	// Generate analytics
	response, err := api.analyticsEngine.GenerateAnalytics(c.Request.Context(), &req)
	if err != nil {
		middleware.AbortWithInternal(c, "Failed to generate analytics", err)
		return
	}
	
	c.JSON(http.StatusOK, response)
}

// GetFleetOverviewHandler handles fleet overview analytics requests
func (api *AnalyticsAPI) GetFleetOverviewHandler(c *gin.Context) {
	req := &AnalyticsRequest{
		ReportType: ReportTypeFleetOverview,
		DateRange: DateRange{
			StartDate: time.Now().AddDate(0, 0, -30),
			EndDate:   time.Now(),
			Period:    "daily",
		},
		IncludeCharts: true,
	}
	
	// Set company ID from context
	if companyID, exists := c.Get("company_id"); exists {
		req.CompanyID = companyID.(string)
	}
	
	// Set user ID from context
	if userID, exists := c.Get("user_id"); exists {
		req.UserID = userID.(string)
	}
	
	// Parse query parameters
	if startDate := c.Query("start_date"); startDate != "" {
		if parsed, err := time.Parse("2006-01-02", startDate); err == nil {
			req.DateRange.StartDate = parsed
		}
	}
	
	if endDate := c.Query("end_date"); endDate != "" {
		if parsed, err := time.Parse("2006-01-02", endDate); err == nil {
			req.DateRange.EndDate = parsed
		}
	}
	
	if period := c.Query("period"); period != "" {
		req.DateRange.Period = period
	}
	
	if includeCharts := c.Query("include_charts"); includeCharts == "true" {
		req.IncludeCharts = true
	}
	
	// Generate analytics
	response, err := api.analyticsEngine.GenerateAnalytics(c.Request.Context(), req)
	if err != nil {
		middleware.AbortWithInternal(c, "Failed to generate fleet overview analytics", err)
		return
	}
	
	c.JSON(http.StatusOK, response)
}

// GetDriverPerformanceHandler handles driver performance analytics requests
func (api *AnalyticsAPI) GetDriverPerformanceHandler(c *gin.Context) {
	req := &AnalyticsRequest{
		ReportType: ReportTypeDriverPerformance,
		DateRange: DateRange{
			StartDate: time.Now().AddDate(0, 0, -30),
			EndDate:   time.Now(),
			Period:    "daily",
		},
		IncludeCharts: true,
	}
	
	// Set company ID from context
	if companyID, exists := c.Get("company_id"); exists {
		req.CompanyID = companyID.(string)
	}
	
	// Set user ID from context
	if userID, exists := c.Get("user_id"); exists {
		req.UserID = userID.(string)
	}
	
	// Parse query parameters
	if startDate := c.Query("start_date"); startDate != "" {
		if parsed, err := time.Parse("2006-01-02", startDate); err == nil {
			req.DateRange.StartDate = parsed
		}
	}
	
	if endDate := c.Query("end_date"); endDate != "" {
		if parsed, err := time.Parse("2006-01-02", endDate); err == nil {
			req.DateRange.EndDate = parsed
		}
	}
	
	// Generate analytics
	response, err := api.analyticsEngine.GenerateAnalytics(c.Request.Context(), req)
	if err != nil {
		middleware.AbortWithInternal(c, "Failed to generate driver performance analytics", err)
		return
	}
	
	c.JSON(http.StatusOK, response)
}

// GetFuelAnalyticsHandler handles fuel analytics requests
func (api *AnalyticsAPI) GetFuelAnalyticsHandler(c *gin.Context) {
	req := &AnalyticsRequest{
		ReportType: ReportTypeFuelAnalytics,
		DateRange: DateRange{
			StartDate: time.Now().AddDate(0, 0, -30),
			EndDate:   time.Now(),
			Period:    "daily",
		},
		IncludeCharts: true,
	}
	
	// Set company ID from context
	if companyID, exists := c.Get("company_id"); exists {
		req.CompanyID = companyID.(string)
	}
	
	// Set user ID from context
	if userID, exists := c.Get("user_id"); exists {
		req.UserID = userID.(string)
	}
	
	// Parse query parameters
	if startDate := c.Query("start_date"); startDate != "" {
		if parsed, err := time.Parse("2006-01-02", startDate); err == nil {
			req.DateRange.StartDate = parsed
		}
	}
	
	if endDate := c.Query("end_date"); endDate != "" {
		if parsed, err := time.Parse("2006-01-02", endDate); err == nil {
			req.DateRange.EndDate = parsed
		}
	}
	
	// Generate analytics
	response, err := api.analyticsEngine.GenerateAnalytics(c.Request.Context(), req)
	if err != nil {
		middleware.AbortWithInternal(c, "Failed to generate fuel analytics", err)
		return
	}
	
	c.JSON(http.StatusOK, response)
}

// GetMaintenanceCostsHandler handles maintenance costs analytics requests
func (api *AnalyticsAPI) GetMaintenanceCostsHandler(c *gin.Context) {
	req := &AnalyticsRequest{
		ReportType: ReportTypeMaintenanceCosts,
		DateRange: DateRange{
			StartDate: time.Now().AddDate(0, -3, 0), // Last 3 months
			EndDate:   time.Now(),
			Period:    "monthly",
		},
		IncludeCharts: true,
	}
	
	// Set company ID from context
	if companyID, exists := c.Get("company_id"); exists {
		req.CompanyID = companyID.(string)
	}
	
	// Set user ID from context
	if userID, exists := c.Get("user_id"); exists {
		req.UserID = userID.(string)
	}
	
	// Parse query parameters
	if startDate := c.Query("start_date"); startDate != "" {
		if parsed, err := time.Parse("2006-01-02", startDate); err == nil {
			req.DateRange.StartDate = parsed
		}
	}
	
	if endDate := c.Query("end_date"); endDate != "" {
		if parsed, err := time.Parse("2006-01-02", endDate); err == nil {
			req.DateRange.EndDate = parsed
		}
	}
	
	// Generate analytics
	response, err := api.analyticsEngine.GenerateAnalytics(c.Request.Context(), req)
	if err != nil {
		middleware.AbortWithInternal(c, "Failed to generate maintenance costs analytics", err)
		return
	}
	
	c.JSON(http.StatusOK, response)
}

// GetRouteEfficiencyHandler handles route efficiency analytics requests
func (api *AnalyticsAPI) GetRouteEfficiencyHandler(c *gin.Context) {
	req := &AnalyticsRequest{
		ReportType: ReportTypeRouteEfficiency,
		DateRange: DateRange{
			StartDate: time.Now().AddDate(0, 0, -30),
			EndDate:   time.Now(),
			Period:    "daily",
		},
		IncludeCharts: true,
	}
	
	// Set company ID from context
	if companyID, exists := c.Get("company_id"); exists {
		req.CompanyID = companyID.(string)
	}
	
	// Set user ID from context
	if userID, exists := c.Get("user_id"); exists {
		req.UserID = userID.(string)
	}
	
	// Parse query parameters
	if startDate := c.Query("start_date"); startDate != "" {
		if parsed, err := time.Parse("2006-01-02", startDate); err == nil {
			req.DateRange.StartDate = parsed
		}
	}
	
	if endDate := c.Query("end_date"); endDate != "" {
		if parsed, err := time.Parse("2006-01-02", endDate); err == nil {
			req.DateRange.EndDate = parsed
		}
	}
	
	// Generate analytics
	response, err := api.analyticsEngine.GenerateAnalytics(c.Request.Context(), req)
	if err != nil {
		middleware.AbortWithInternal(c, "Failed to generate route efficiency analytics", err)
		return
	}
	
	c.JSON(http.StatusOK, response)
}

// GetGeofenceActivityHandler handles geofence activity analytics requests
func (api *AnalyticsAPI) GetGeofenceActivityHandler(c *gin.Context) {
	req := &AnalyticsRequest{
		ReportType: ReportTypeGeofenceActivity,
		DateRange: DateRange{
			StartDate: time.Now().AddDate(0, 0, -7), // Last 7 days
			EndDate:   time.Now(),
			Period:    "hourly",
		},
		IncludeCharts: true,
	}
	
	// Set company ID from context
	if companyID, exists := c.Get("company_id"); exists {
		req.CompanyID = companyID.(string)
	}
	
	// Set user ID from context
	if userID, exists := c.Get("user_id"); exists {
		req.UserID = userID.(string)
	}
	
	// Parse query parameters
	if startDate := c.Query("start_date"); startDate != "" {
		if parsed, err := time.Parse("2006-01-02", startDate); err == nil {
			req.DateRange.StartDate = parsed
		}
	}
	
	if endDate := c.Query("end_date"); endDate != "" {
		if parsed, err := time.Parse("2006-01-02", endDate); err == nil {
			req.DateRange.EndDate = parsed
		}
	}
	
	// Generate analytics
	response, err := api.analyticsEngine.GenerateAnalytics(c.Request.Context(), req)
	if err != nil {
		middleware.AbortWithInternal(c, "Failed to generate geofence activity analytics", err)
		return
	}
	
	c.JSON(http.StatusOK, response)
}

// GetComplianceReportHandler handles compliance report analytics requests
func (api *AnalyticsAPI) GetComplianceReportHandler(c *gin.Context) {
	req := &AnalyticsRequest{
		ReportType: ReportTypeComplianceReport,
		DateRange: DateRange{
			StartDate: time.Now().AddDate(0, -1, 0), // Last month
			EndDate:   time.Now(),
			Period:    "weekly",
		},
		IncludeCharts: true,
	}
	
	// Set company ID from context
	if companyID, exists := c.Get("company_id"); exists {
		req.CompanyID = companyID.(string)
	}
	
	// Set user ID from context
	if userID, exists := c.Get("user_id"); exists {
		req.UserID = userID.(string)
	}
	
	// Parse query parameters
	if startDate := c.Query("start_date"); startDate != "" {
		if parsed, err := time.Parse("2006-01-02", startDate); err == nil {
			req.DateRange.StartDate = parsed
		}
	}
	
	if endDate := c.Query("end_date"); endDate != "" {
		if parsed, err := time.Parse("2006-01-02", endDate); err == nil {
			req.DateRange.EndDate = parsed
		}
	}
	
	// Generate analytics
	response, err := api.analyticsEngine.GenerateAnalytics(c.Request.Context(), req)
	if err != nil {
		middleware.AbortWithInternal(c, "Failed to generate compliance report analytics", err)
		return
	}
	
	c.JSON(http.StatusOK, response)
}

// GetCostAnalysisHandler handles cost analysis analytics requests
func (api *AnalyticsAPI) GetCostAnalysisHandler(c *gin.Context) {
	req := &AnalyticsRequest{
		ReportType: ReportTypeCostAnalysis,
		DateRange: DateRange{
			StartDate: time.Now().AddDate(0, -3, 0), // Last 3 months
			EndDate:   time.Now(),
			Period:    "monthly",
		},
		IncludeCharts: true,
	}
	
	// Set company ID from context
	if companyID, exists := c.Get("company_id"); exists {
		req.CompanyID = companyID.(string)
	}
	
	// Set user ID from context
	if userID, exists := c.Get("user_id"); exists {
		req.UserID = userID.(string)
	}
	
	// Parse query parameters
	if startDate := c.Query("start_date"); startDate != "" {
		if parsed, err := time.Parse("2006-01-02", startDate); err == nil {
			req.DateRange.StartDate = parsed
		}
	}
	
	if endDate := c.Query("end_date"); endDate != "" {
		if parsed, err := time.Parse("2006-01-02", endDate); err == nil {
			req.DateRange.EndDate = parsed
		}
	}
	
	// Generate analytics
	response, err := api.analyticsEngine.GenerateAnalytics(c.Request.Context(), req)
	if err != nil {
		middleware.AbortWithInternal(c, "Failed to generate cost analysis analytics", err)
		return
	}
	
	c.JSON(http.StatusOK, response)
}

// GetUtilizationReportHandler handles utilization report analytics requests
func (api *AnalyticsAPI) GetUtilizationReportHandler(c *gin.Context) {
	req := &AnalyticsRequest{
		ReportType: ReportTypeUtilizationReport,
		DateRange: DateRange{
			StartDate: time.Now().AddDate(0, 0, -30),
			EndDate:   time.Now(),
			Period:    "daily",
		},
		IncludeCharts: true,
	}
	
	// Set company ID from context
	if companyID, exists := c.Get("company_id"); exists {
		req.CompanyID = companyID.(string)
	}
	
	// Set user ID from context
	if userID, exists := c.Get("user_id"); exists {
		req.UserID = userID.(string)
	}
	
	// Parse query parameters
	if startDate := c.Query("start_date"); startDate != "" {
		if parsed, err := time.Parse("2006-01-02", startDate); err == nil {
			req.DateRange.StartDate = parsed
		}
	}
	
	if endDate := c.Query("end_date"); endDate != "" {
		if parsed, err := time.Parse("2006-01-02", endDate); err == nil {
			req.DateRange.EndDate = parsed
		}
	}
	
	// Generate analytics
	response, err := api.analyticsEngine.GenerateAnalytics(c.Request.Context(), req)
	if err != nil {
		middleware.AbortWithInternal(c, "Failed to generate utilization report analytics", err)
		return
	}
	
	c.JSON(http.StatusOK, response)
}

// GetPredictiveInsightsHandler handles predictive insights analytics requests
func (api *AnalyticsAPI) GetPredictiveInsightsHandler(c *gin.Context) {
	req := &AnalyticsRequest{
		ReportType: ReportTypePredictiveInsights,
		DateRange: DateRange{
			StartDate: time.Now().AddDate(0, -6, 0), // Last 6 months
			EndDate:   time.Now(),
			Period:    "monthly",
		},
		IncludeCharts: true,
	}
	
	// Set company ID from context
	if companyID, exists := c.Get("company_id"); exists {
		req.CompanyID = companyID.(string)
	}
	
	// Set user ID from context
	if userID, exists := c.Get("user_id"); exists {
		req.UserID = userID.(string)
	}
	
	// Parse query parameters
	if startDate := c.Query("start_date"); startDate != "" {
		if parsed, err := time.Parse("2006-01-02", startDate); err == nil {
			req.DateRange.StartDate = parsed
		}
	}
	
	if endDate := c.Query("end_date"); endDate != "" {
		if parsed, err := time.Parse("2006-01-02", endDate); err == nil {
			req.DateRange.EndDate = parsed
		}
	}
	
	// Generate analytics
	response, err := api.analyticsEngine.GenerateAnalytics(c.Request.Context(), req)
	if err != nil {
		middleware.AbortWithInternal(c, "Failed to generate predictive insights analytics", err)
		return
	}
	
	c.JSON(http.StatusOK, response)
}

// GetAnalyticsCacheStatsHandler handles analytics cache statistics requests
func (api *AnalyticsAPI) GetAnalyticsCacheStatsHandler(c *gin.Context) {
	stats, err := api.analyticsEngine.cache.GetAnalyticsCacheStats(c.Request.Context())
	if err != nil {
		middleware.AbortWithInternal(c, "Failed to get analytics cache statistics", err)
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"cache_stats": stats,
		"timestamp":   time.Now(),
	})
}

// InvalidateAnalyticsCacheHandler handles analytics cache invalidation requests
func (api *AnalyticsAPI) InvalidateAnalyticsCacheHandler(c *gin.Context) {
	var req struct {
		CompanyID  string `json:"company_id,omitempty"`
		ReportType string `json:"report_type,omitempty"`
		All        bool   `json:"all,omitempty"`
	}
	
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.AbortWithBadRequest(c, err.Error())
		return
	}
	
	// Set company ID from context if not provided
	if req.CompanyID == "" {
		if companyID, exists := c.Get("company_id"); exists {
			req.CompanyID = companyID.(string)
		}
	}
	
	var err error
	if req.All {
		err = api.analyticsEngine.cache.InvalidateAllAnalyticsCache(c.Request.Context())
	} else if req.ReportType != "" {
		// Invalidate specific report type
		analyticsReq := &AnalyticsRequest{
			CompanyID:  req.CompanyID,
			ReportType: req.ReportType,
		}
		err = api.analyticsEngine.cache.InvalidateAnalyticsCache(c.Request.Context(), analyticsReq)
	} else {
		// Invalidate all for company
		analyticsReq := &AnalyticsRequest{
			CompanyID: req.CompanyID,
		}
		err = api.analyticsEngine.cache.InvalidateAnalyticsCache(c.Request.Context(), analyticsReq)
	}
	
	if err != nil {
		middleware.AbortWithInternal(c, "Failed to invalidate analytics cache", err)
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"message": "Analytics cache invalidated successfully",
		"timestamp": time.Now(),
	})
}

// GetAvailableReportTypesHandler returns available report types
func (api *AnalyticsAPI) GetAvailableReportTypesHandler(c *gin.Context) {
	reportTypes := []map[string]interface{}{
		{
			"type":        ReportTypeFleetOverview,
			"name":        "Fleet Overview",
			"description": "Comprehensive fleet performance overview",
			"category":    "overview",
		},
		{
			"type":        ReportTypeDriverPerformance,
			"name":        "Driver Performance",
			"description": "Driver performance analytics and insights",
			"category":    "performance",
		},
		{
			"type":        ReportTypeFuelAnalytics,
			"name":        "Fuel Analytics",
			"description": "Fuel consumption and efficiency analysis",
			"category":    "efficiency",
		},
		{
			"type":        ReportTypeMaintenanceCosts,
			"name":        "Maintenance Costs",
			"description": "Maintenance cost analysis and trends",
			"category":    "costs",
		},
		{
			"type":        ReportTypeRouteEfficiency,
			"name":        "Route Efficiency",
			"description": "Route optimization and efficiency metrics",
			"category":    "efficiency",
		},
		{
			"type":        ReportTypeGeofenceActivity,
			"name":        "Geofence Activity",
			"description": "Geofence usage and violation analytics",
			"category":    "compliance",
		},
		{
			"type":        ReportTypeComplianceReport,
			"name":        "Compliance Report",
			"description": "Regulatory compliance and safety metrics",
			"category":    "compliance",
		},
		{
			"type":        ReportTypeCostAnalysis,
			"name":        "Cost Analysis",
			"description": "Comprehensive cost analysis and breakdown",
			"category":    "costs",
		},
		{
			"type":        ReportTypeUtilizationReport,
			"name":        "Utilization Report",
			"description": "Vehicle and driver utilization metrics",
			"category":    "utilization",
		},
		{
			"type":        ReportTypePredictiveInsights,
			"name":        "Predictive Insights",
			"description": "Predictive analytics and future insights",
			"category":    "predictive",
		},
	}
	
	c.JSON(http.StatusOK, gin.H{
		"report_types": reportTypes,
		"total":        len(reportTypes),
		"timestamp":    time.Now(),
	})
}

// SetupAnalyticsRoutes registers analytics API routes
func SetupAnalyticsRoutes(rg *gin.RouterGroup, api *AnalyticsAPI) {
	analytics := rg.Group("/analytics")
	{
		// General analytics endpoint
		analytics.POST("/generate", api.GenerateAnalyticsHandler)
		
		// Specific report endpoints
		analytics.GET("/fleet-overview", api.GetFleetOverviewHandler)
		analytics.GET("/driver-performance", api.GetDriverPerformanceHandler)
		analytics.GET("/fuel-analytics", api.GetFuelAnalyticsHandler)
		analytics.GET("/maintenance-costs", api.GetMaintenanceCostsHandler)
		analytics.GET("/route-efficiency", api.GetRouteEfficiencyHandler)
		analytics.GET("/geofence-activity", api.GetGeofenceActivityHandler)
		analytics.GET("/compliance-report", api.GetComplianceReportHandler)
		analytics.GET("/cost-analysis", api.GetCostAnalysisHandler)
		analytics.GET("/utilization-report", api.GetUtilizationReportHandler)
		analytics.GET("/predictive-insights", api.GetPredictiveInsightsHandler)
		
		// Cache management endpoints
		analytics.GET("/cache/stats", api.GetAnalyticsCacheStatsHandler)
		analytics.DELETE("/cache", api.InvalidateAnalyticsCacheHandler)
		
		// Utility endpoints
		analytics.GET("/report-types", api.GetAvailableReportTypesHandler)
	}
}
