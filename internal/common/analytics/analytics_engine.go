package analytics

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"

	"github.com/tobangado69/fleettracker-pro/backend/pkg/models"
)

// AnalyticsEngine provides comprehensive analytics and reporting capabilities
type AnalyticsEngine struct {
	db    *gorm.DB
	redis *redis.Client
	cache *AnalyticsCache
}

// NewAnalyticsEngine creates a new analytics engine
func NewAnalyticsEngine(db *gorm.DB, redis *redis.Client) *AnalyticsEngine {
	return &AnalyticsEngine{
		db:    db,
		redis: redis,
		cache: NewAnalyticsCache(redis),
	}
}

// AnalyticsRequest represents a request for analytics data
type AnalyticsRequest struct {
	CompanyID    string                 `json:"company_id"`
	UserID       string                 `json:"user_id"`
	ReportType   string                 `json:"report_type"`
	DateRange    DateRange              `json:"date_range"`
	Filters      map[string]interface{} `json:"filters"`
	GroupBy      []string               `json:"group_by"`
	Metrics      []string               `json:"metrics"`
	Format       string                 `json:"format"` // json, csv, pdf
	IncludeCharts bool                  `json:"include_charts"`
}

// DateRange represents a date range for analytics
type DateRange struct {
	StartDate time.Time `json:"start_date"`
	EndDate   time.Time `json:"end_date"`
	Period    string    `json:"period"` // daily, weekly, monthly, quarterly, yearly
}

// AnalyticsResponse represents the response from analytics queries
type AnalyticsResponse struct {
	ReportType    string                 `json:"report_type"`
	DateRange     DateRange              `json:"date_range"`
	Data          interface{}            `json:"data"`
	Summary       AnalyticsSummary       `json:"summary"`
	Charts        []ChartData            `json:"charts,omitempty"`
	Metadata      AnalyticsMetadata      `json:"metadata"`
	GeneratedAt   time.Time              `json:"generated_at"`
	FromCache     bool                   `json:"from_cache"`
	CacheHit      bool                   `json:"cache_hit"`
}

// AnalyticsSummary provides summary statistics
type AnalyticsSummary struct {
	TotalRecords    int64   `json:"total_records"`
	TotalValue      float64 `json:"total_value"`
	AverageValue    float64 `json:"average_value"`
	MinValue        float64 `json:"min_value"`
	MaxValue        float64 `json:"max_value"`
	GrowthRate      float64 `json:"growth_rate"`
	Trend           string  `json:"trend"` // up, down, stable
	KeyInsights     []string `json:"key_insights"`
	Recommendations []string `json:"recommendations"`
}

// ChartData represents chart data for visualization
type ChartData struct {
	Type        string      `json:"type"` // line, bar, pie, area, scatter
	Title       string      `json:"title"`
	XAxis       []string    `json:"x_axis"`
	YAxis       []float64   `json:"y_axis"`
	Data        interface{} `json:"data"`
	Options     map[string]interface{} `json:"options"`
}

// AnalyticsMetadata provides metadata about the analytics
type AnalyticsMetadata struct {
	QueryTime     time.Duration `json:"query_time"`
	CacheTime     time.Duration `json:"cache_time"`
	DataPoints    int           `json:"data_points"`
	LastUpdated   time.Time     `json:"last_updated"`
	DataQuality   string        `json:"data_quality"` // high, medium, low
	Completeness  float64       `json:"completeness"` // percentage
}

// Report types
const (
	ReportTypeFleetOverview     = "fleet_overview"
	ReportTypeDriverPerformance = "driver_performance"
	ReportTypeFuelAnalytics     = "fuel_analytics"
	ReportTypeMaintenanceCosts  = "maintenance_costs"
	ReportTypeRouteEfficiency   = "route_efficiency"
	ReportTypeGeofenceActivity  = "geofence_activity"
	ReportTypeComplianceReport  = "compliance_report"
	ReportTypeCostAnalysis      = "cost_analysis"
	ReportTypeUtilizationReport = "utilization_report"
	ReportTypePredictiveInsights = "predictive_insights"
)

// GenerateAnalytics generates comprehensive analytics based on the request
func (ae *AnalyticsEngine) GenerateAnalytics(ctx context.Context, req *AnalyticsRequest) (*AnalyticsResponse, error) {
	startTime := time.Now()
	
	// Try to get from cache first
	cachedResponse, cacheHit, err := ae.cache.GetAnalyticsFromCache(ctx, req)
	if err != nil {
		fmt.Printf("Error getting analytics from cache: %v\n", err)
	}
	
	if cacheHit {
		fmt.Printf("Cache hit for analytics report: %s\n", req.ReportType)
		return cachedResponse, nil
	}
	
	fmt.Printf("Cache miss for analytics report: %s. Generating data...\n", req.ReportType)
	
	// Generate analytics data based on report type
	var data interface{}
	var summary AnalyticsSummary
	var charts []ChartData
	var metadata AnalyticsMetadata
	
	switch req.ReportType {
	case ReportTypeFleetOverview:
		data, summary, charts, err = ae.generateFleetOverviewAnalytics(context.Background(), req)
	case ReportTypeDriverPerformance:
		data, summary, charts, err = ae.generateDriverPerformanceAnalytics(context.Background(), req)
	case ReportTypeFuelAnalytics:
		data, summary, charts, err = ae.generateFuelAnalytics(context.Background(), req)
	case ReportTypeMaintenanceCosts:
		data, summary, charts, err = ae.generateMaintenanceCostsAnalytics(context.Background(), req)
	case ReportTypeRouteEfficiency:
		data, summary, charts, err = ae.generateRouteEfficiencyAnalytics(context.Background(), req)
	case ReportTypeGeofenceActivity:
		data, summary, charts, err = ae.generateGeofenceActivityAnalytics(context.Background(), req)
	case ReportTypeComplianceReport:
		data, summary, charts, err = ae.generateComplianceReportAnalytics(context.Background(), req)
	case ReportTypeCostAnalysis:
		data, summary, charts, err = ae.generateCostAnalysisAnalytics(context.Background(), req)
	case ReportTypeUtilizationReport:
		data, summary, charts, err = ae.generateUtilizationReportAnalytics(context.Background(), req)
	case ReportTypePredictiveInsights:
		data, summary, charts, err = ae.generatePredictiveInsightsAnalytics(context.Background(), req)
	default:
		return nil, fmt.Errorf("unsupported report type: %s", req.ReportType)
	}
	
	if err != nil {
		return nil, fmt.Errorf("failed to generate analytics: %w", err)
	}
	
	// Calculate metadata
	queryTime := time.Since(startTime)
	metadata = AnalyticsMetadata{
		QueryTime:    queryTime,
		DataPoints:   ae.calculateDataPoints(data),
		LastUpdated:  time.Now(),
		DataQuality:  ae.assessDataQuality(data),
		Completeness: ae.calculateCompleteness(data),
	}
	
	response := &AnalyticsResponse{
		ReportType:  req.ReportType,
		DateRange:   req.DateRange,
		Data:        data,
		Summary:     summary,
		Charts:      charts,
		Metadata:    metadata,
		GeneratedAt: time.Now(),
		FromCache:   false,
		CacheHit:    false,
	}
	
	// Store in cache asynchronously
	go func() {
		if err := ae.cache.SetAnalyticsInCache(ctx, req, response); err != nil {
			fmt.Printf("Failed to set analytics in cache: %v\n", err)
		}
	}()
	
	return response, nil
}

// FleetOverviewData represents fleet overview analytics data
type FleetOverviewData struct {
	TotalVehicles     int                    `json:"total_vehicles"`
	ActiveVehicles    int                    `json:"active_vehicles"`
	TotalDrivers      int                    `json:"total_drivers"`
	ActiveDrivers     int                    `json:"active_drivers"`
	TotalTrips        int                    `json:"total_trips"`
	ActiveTrips       int                    `json:"active_trips"`
	TotalDistance     float64                `json:"total_distance"`
	TotalFuelUsed     float64                `json:"total_fuel_used"`
	AverageSpeed      float64                `json:"average_speed"`
	UtilizationRate   float64                `json:"utilization_rate"`
	CostPerKm         float64                `json:"cost_per_km"`
	EfficiencyScore   float64                `json:"efficiency_score"`
	VehicleBreakdown  []VehicleBreakdown     `json:"vehicle_breakdown"`
	DriverBreakdown   []DriverBreakdown      `json:"driver_breakdown"`
	PerformanceTrends []PerformanceTrend     `json:"performance_trends"`
	Alerts            []AnalyticsAlert       `json:"alerts"`
}

// VehicleBreakdown represents vehicle statistics breakdown
type VehicleBreakdown struct {
	Status     string  `json:"status"`
	Count      int     `json:"count"`
	Percentage float64 `json:"percentage"`
	AvgMileage float64 `json:"avg_mileage"`
	AvgFuel    float64 `json:"avg_fuel"`
}

// DriverBreakdown represents driver statistics breakdown
type DriverBreakdown struct {
	Status        string  `json:"status"`
	Count         int     `json:"count"`
	Percentage    float64 `json:"percentage"`
	AvgScore      float64 `json:"avg_score"`
	AvgTrips      float64 `json:"avg_trips"`
	AvgDistance   float64 `json:"avg_distance"`
}

// PerformanceTrend represents performance trend data
type PerformanceTrend struct {
	Date        time.Time `json:"date"`
	Efficiency  float64   `json:"efficiency"`
	Utilization float64   `json:"utilization"`
	Cost        float64   `json:"cost"`
	Distance    float64   `json:"distance"`
}

// AnalyticsAlert represents an analytics alert
type AnalyticsAlert struct {
	Type        string    `json:"type"`
	Severity    string    `json:"severity"`
	Message     string    `json:"message"`
	VehicleID   string    `json:"vehicle_id,omitempty"`
	DriverID    string    `json:"driver_id,omitempty"`
	Timestamp   time.Time `json:"timestamp"`
	Action      string    `json:"action,omitempty"`
}

// generateFleetOverviewAnalytics generates comprehensive fleet overview analytics
func (ae *AnalyticsEngine) generateFleetOverviewAnalytics(_ context.Context, req *AnalyticsRequest) (interface{}, AnalyticsSummary, []ChartData, error) {
	// Get fleet statistics
	var totalVehicles, activeVehicles int64
	ae.db.Model(&models.Vehicle{}).Where("company_id = ?", req.CompanyID).Count(&totalVehicles)
	ae.db.Model(&models.Vehicle{}).Where("company_id = ? AND status = 'active'", req.CompanyID).Count(&activeVehicles)
	
	var totalDrivers, activeDrivers int64
	ae.db.Model(&models.Driver{}).Where("company_id = ?", req.CompanyID).Count(&totalDrivers)
	ae.db.Model(&models.Driver{}).Where("company_id = ? AND status = 'active'", req.CompanyID).Count(&activeDrivers)
	
	var totalTrips, activeTrips int64
	ae.db.Model(&models.Trip{}).Where("company_id = ?", req.CompanyID).Count(&totalTrips)
	ae.db.Model(&models.Trip{}).Where("company_id = ? AND status = 'in_progress'", req.CompanyID).Count(&activeTrips)
	
	// Calculate total distance and fuel
	var totalDistance, totalFuel float64
	ae.db.Model(&models.Trip{}).Where("company_id = ? AND status = 'completed'", req.CompanyID).
		Select("COALESCE(SUM(total_distance), 0)").Scan(&totalDistance)
	ae.db.Model(&models.Trip{}).Where("company_id = ? AND status = 'completed'", req.CompanyID).
		Select("COALESCE(SUM(fuel_consumed), 0)").Scan(&totalFuel)
	
	// Calculate utilization rate
	utilizationRate := float64(0)
	if totalVehicles > 0 {
		utilizationRate = (float64(activeVehicles) / float64(totalVehicles)) * 100
	}
	
	// Calculate efficiency score
	efficiencyScore := ae.calculateEfficiencyScore(context.Background(), req.CompanyID)
	
	// Get vehicle breakdown
	vehicleBreakdown := ae.getVehicleBreakdown(context.Background(), req.CompanyID)
	
	// Get driver breakdown
	driverBreakdown := ae.getDriverBreakdown(context.Background(), req.CompanyID)
	
	// Get performance trends
	performanceTrends := ae.getPerformanceTrends(context.Background(), req.CompanyID, req.DateRange)
	
	// Get alerts
	alerts := ae.getAnalyticsAlerts(context.Background(), req.CompanyID)
	
	data := FleetOverviewData{
		TotalVehicles:     int(totalVehicles),
		ActiveVehicles:    int(activeVehicles),
		TotalDrivers:      int(totalDrivers),
		ActiveDrivers:     int(activeDrivers),
		TotalTrips:        int(totalTrips),
		ActiveTrips:       int(activeTrips),
		TotalDistance:     totalDistance,
		TotalFuelUsed:     totalFuel,
		AverageSpeed:      ae.calculateAverageSpeed(context.Background(), req.CompanyID),
		UtilizationRate:   utilizationRate,
		CostPerKm:         ae.calculateCostPerKm(context.Background(), req.CompanyID),
		EfficiencyScore:   efficiencyScore,
		VehicleBreakdown:  vehicleBreakdown,
		DriverBreakdown:   driverBreakdown,
		PerformanceTrends: performanceTrends,
		Alerts:            alerts,
	}
	
	// Generate summary
	summary := AnalyticsSummary{
		TotalRecords: totalTrips,
		TotalValue:   totalDistance,
		AverageValue: ae.calculateAverageValue(data),
		MinValue:     ae.calculateMinValue(data),
		MaxValue:     ae.calculateMaxValue(data),
		GrowthRate:   ae.calculateGrowthRate(performanceTrends),
		Trend:        ae.determineTrend(performanceTrends),
		KeyInsights:  ae.generateKeyInsights(data),
		Recommendations: ae.generateRecommendations(data),
	}
	
	// Generate charts
	charts := ae.generateFleetOverviewCharts(data)
	
	return data, summary, charts, nil
}

// Helper methods for analytics calculations
func (ae *AnalyticsEngine) calculateEfficiencyScore(_ context.Context, companyID string) float64 {
	// Calculate efficiency based on multiple factors
	// This is a simplified calculation - in production, this would be more sophisticated
	
	var avgSpeed, avgFuel, utilization float64
	
	// Get average speed
	ae.db.Model(&models.Trip{}).Where("company_id = ? AND status = 'completed'", companyID).
		Select("COALESCE(AVG(average_speed), 0)").Scan(&avgSpeed)
	
	// Get average fuel efficiency
	ae.db.Model(&models.Trip{}).Where("company_id = ? AND status = 'completed'", companyID).
		Select("COALESCE(AVG(fuel_consumed/total_distance), 0)").Scan(&avgFuel)
	
	// Get utilization rate
	var totalVehicles, activeVehicles int64
	ae.db.Model(&models.Vehicle{}).Where("company_id = ?", companyID).Count(&totalVehicles)
	ae.db.Model(&models.Vehicle{}).Where("company_id = ? AND status = 'active'", companyID).Count(&activeVehicles)
	
	if totalVehicles > 0 {
		utilization = (float64(activeVehicles) / float64(totalVehicles)) * 100
	}
	
	// Calculate efficiency score (0-100)
	efficiency := (avgSpeed/100*30) + (utilization/100*40) + ((100-avgFuel)/100*30)
	return math.Min(100, math.Max(0, efficiency))
}

func (ae *AnalyticsEngine) getVehicleBreakdown(_ context.Context, companyID string) []VehicleBreakdown {
	var breakdown []VehicleBreakdown
	
	// Get vehicle counts by status
	var statuses []struct {
		Status string `json:"status"`
		Count  int    `json:"count"`
	}
	
	ae.db.Model(&models.Vehicle{}).Where("company_id = ?", companyID).
		Select("status, COUNT(*) as count").
		Group("status").
		Scan(&statuses)
	
	var totalVehicles int64
	ae.db.Model(&models.Vehicle{}).Where("company_id = ?", companyID).Count(&totalVehicles)
	
	for _, status := range statuses {
		percentage := float64(0)
		if totalVehicles > 0 {
			percentage = (float64(status.Count) / float64(totalVehicles)) * 100
		}
		
		breakdown = append(breakdown, VehicleBreakdown{
			Status:     status.Status,
			Count:      status.Count,
			Percentage: percentage,
			AvgMileage: ae.getAverageMileageByStatus(context.Background(), companyID, status.Status),
			AvgFuel:    ae.getAverageFuelByStatus(context.Background(), companyID, status.Status),
		})
	}
	
	return breakdown
}

func (ae *AnalyticsEngine) getDriverBreakdown(_ context.Context, companyID string) []DriverBreakdown {
	var breakdown []DriverBreakdown
	
	// Get driver counts by status
	var statuses []struct {
		Status string `json:"status"`
		Count  int    `json:"count"`
	}
	
	ae.db.Model(&models.Driver{}).Where("company_id = ?", companyID).
		Select("status, COUNT(*) as count").
		Group("status").
		Scan(&statuses)
	
	var totalDrivers int64
	ae.db.Model(&models.Driver{}).Where("company_id = ?", companyID).Count(&totalDrivers)
	
	for _, status := range statuses {
		percentage := float64(0)
		if totalDrivers > 0 {
			percentage = (float64(status.Count) / float64(totalDrivers)) * 100
		}
		
		breakdown = append(breakdown, DriverBreakdown{
			Status:      status.Status,
			Count:       status.Count,
			Percentage:  percentage,
			AvgScore:    ae.getAverageScoreByStatus(context.Background(), companyID, status.Status),
			AvgTrips:    ae.getAverageTripsByStatus(context.Background(), companyID, status.Status),
			AvgDistance: ae.getAverageDistanceByStatus(context.Background(), companyID, status.Status),
		})
	}
	
	return breakdown
}

func (ae *AnalyticsEngine) getPerformanceTrends(_ context.Context, companyID string, dateRange DateRange) []PerformanceTrend {
	var trends []PerformanceTrend
	
	// Generate trends based on date range
	startDate := dateRange.StartDate
	endDate := dateRange.EndDate
	
	// Create daily data points
	for d := startDate; d.Before(endDate); d = d.AddDate(0, 0, 1) {
		nextDay := d.AddDate(0, 0, 1)
		
		var efficiency, utilization, cost, distance float64
		
		// Calculate daily metrics
		ae.db.Model(&models.Trip{}).Where("company_id = ? AND created_at >= ? AND created_at < ?", companyID, d, nextDay).
			Select("COALESCE(AVG(average_speed), 0)").Scan(&efficiency)
		
		ae.db.Model(&models.Vehicle{}).Where("company_id = ? AND status = 'active'", companyID).
			Select("COUNT(*) * 100.0 / (SELECT COUNT(*) FROM vehicles WHERE company_id = ?)", companyID).Scan(&utilization)
		
		ae.db.Model(&models.Trip{}).Where("company_id = ? AND created_at >= ? AND created_at < ?", companyID, d, nextDay).
			Select("COALESCE(SUM(total_cost), 0)").Scan(&cost)
		
		ae.db.Model(&models.Trip{}).Where("company_id = ? AND created_at >= ? AND created_at < ?", companyID, d, nextDay).
			Select("COALESCE(SUM(total_distance), 0)").Scan(&distance)
		
		trends = append(trends, PerformanceTrend{
			Date:        d,
			Efficiency:  efficiency,
			Utilization: utilization,
			Cost:        cost,
			Distance:    distance,
		})
	}
	
	return trends
}

func (ae *AnalyticsEngine) getAnalyticsAlerts(_ context.Context, _ string) []AnalyticsAlert {
	var alerts []AnalyticsAlert
	
	// Get recent alerts from various sources
	// This would integrate with the existing alert system
	
	// Example alerts based on data analysis
	alerts = append(alerts, AnalyticsAlert{
		Type:      "efficiency",
		Severity:  "medium",
		Message:   "Fleet efficiency is below target",
		Timestamp: time.Now(),
		Action:    "Review driver training and route optimization",
	})
	
	return alerts
}

// Additional helper methods for calculations
func (ae *AnalyticsEngine) calculateAverageSpeed(_ context.Context, companyID string) float64 {
	var avgSpeed float64
	ae.db.Model(&models.Trip{}).Where("company_id = ? AND status = 'completed'", companyID).
		Select("COALESCE(AVG(average_speed), 0)").Scan(&avgSpeed)
	return avgSpeed
}

func (ae *AnalyticsEngine) calculateCostPerKm(_ context.Context, companyID string) float64 {
	var totalCost, totalDistance float64
	
	ae.db.Model(&models.Trip{}).Where("company_id = ? AND status = 'completed'", companyID).
		Select("COALESCE(SUM(total_cost), 0)").Scan(&totalCost)
	
	ae.db.Model(&models.Trip{}).Where("company_id = ? AND status = 'completed'", companyID).
		Select("COALESCE(SUM(total_distance), 0)").Scan(&totalDistance)
	
	if totalDistance > 0 {
		return totalCost / totalDistance
	}
	return 0
}

func (ae *AnalyticsEngine) getAverageMileageByStatus(_ context.Context, companyID, status string) float64 {
	var avgMileage float64
	ae.db.Model(&models.Vehicle{}).Where("company_id = ? AND status = ?", companyID, status).
		Select("COALESCE(AVG(current_mileage), 0)").Scan(&avgMileage)
	return avgMileage
}

func (ae *AnalyticsEngine) getAverageFuelByStatus(_ context.Context, companyID, _ string) float64 {
	var avgFuel float64
	ae.db.Model(&models.Trip{}).Where("company_id = ? AND status = 'completed'", companyID).
		Select("COALESCE(AVG(fuel_consumed), 0)").Scan(&avgFuel)
	return avgFuel
}

func (ae *AnalyticsEngine) getAverageScoreByStatus(_ context.Context, companyID, status string) float64 {
	var avgScore float64
	ae.db.Model(&models.Driver{}).Where("company_id = ? AND status = ?", companyID, status).
		Select("COALESCE(AVG(performance_score), 0)").Scan(&avgScore)
	return avgScore
}

func (ae *AnalyticsEngine) getAverageTripsByStatus(_ context.Context, companyID, status string) float64 {
	var avgTrips float64
	ae.db.Model(&models.Driver{}).Where("company_id = ? AND status = ?", companyID, status).
		Select("COALESCE(AVG(trip_count), 0)").Scan(&avgTrips)
	return avgTrips
}

func (ae *AnalyticsEngine) getAverageDistanceByStatus(_ context.Context, companyID, status string) float64 {
	var avgDistance float64
	ae.db.Model(&models.Driver{}).Where("company_id = ? AND status = ?", companyID, status).
		Select("COALESCE(AVG(total_distance), 0)").Scan(&avgDistance)
	return avgDistance
}

// Summary calculation methods
func (ae *AnalyticsEngine) calculateAverageValue(data interface{}) float64 {
	// Simplified calculation - in production, this would be more sophisticated
	if fleetData, ok := data.(FleetOverviewData); ok {
		return (fleetData.TotalDistance + fleetData.TotalFuelUsed) / 2
	}
	return 0
}

func (ae *AnalyticsEngine) calculateMinValue(data interface{}) float64 {
	// Simplified calculation
	if fleetData, ok := data.(FleetOverviewData); ok {
		return math.Min(fleetData.TotalDistance, fleetData.TotalFuelUsed)
	}
	return 0
}

func (ae *AnalyticsEngine) calculateMaxValue(data interface{}) float64 {
	// Simplified calculation
	if fleetData, ok := data.(FleetOverviewData); ok {
		return math.Max(fleetData.TotalDistance, fleetData.TotalFuelUsed)
	}
	return 0
}

func (ae *AnalyticsEngine) calculateGrowthRate(trends []PerformanceTrend) float64 {
	if len(trends) < 2 {
		return 0
	}
	
	first := trends[0].Efficiency
	last := trends[len(trends)-1].Efficiency
	
	if first == 0 {
		return 0
	}
	
	return ((last - first) / first) * 100
}

func (ae *AnalyticsEngine) determineTrend(trends []PerformanceTrend) string {
	growthRate := ae.calculateGrowthRate(trends)
	
	if growthRate > 5 {
		return "up"
	} else if growthRate < -5 {
		return "down"
	}
	return "stable"
}

func (ae *AnalyticsEngine) generateKeyInsights(data interface{}) []string {
	var insights []string
	
	if fleetData, ok := data.(FleetOverviewData); ok {
		if fleetData.UtilizationRate < 70 {
			insights = append(insights, "Vehicle utilization is below optimal levels")
		}
		if fleetData.EfficiencyScore < 60 {
			insights = append(insights, "Fleet efficiency needs improvement")
		}
		if fleetData.CostPerKm > 2.0 {
			insights = append(insights, "Cost per kilometer is above industry average")
		}
	}
	
	return insights
}

func (ae *AnalyticsEngine) generateRecommendations(data interface{}) []string {
	var recommendations []string
	
	if fleetData, ok := data.(FleetOverviewData); ok {
		if fleetData.UtilizationRate < 70 {
			recommendations = append(recommendations, "Consider route optimization to improve vehicle utilization")
		}
		if fleetData.EfficiencyScore < 60 {
			recommendations = append(recommendations, "Implement driver training programs to improve efficiency")
		}
		if fleetData.CostPerKm > 2.0 {
			recommendations = append(recommendations, "Review fuel consumption and maintenance costs")
		}
	}
	
	return recommendations
}

func (ae *AnalyticsEngine) generateFleetOverviewCharts(data FleetOverviewData) []ChartData {
	var charts []ChartData
	
	// Vehicle status pie chart
	vehicleLabels := make([]string, len(data.VehicleBreakdown))
	vehicleValues := make([]float64, len(data.VehicleBreakdown))
	for i, breakdown := range data.VehicleBreakdown {
		vehicleLabels[i] = breakdown.Status
		vehicleValues[i] = breakdown.Percentage
	}
	
	charts = append(charts, ChartData{
		Type:  "pie",
		Title: "Vehicle Status Distribution",
		XAxis: vehicleLabels,
		YAxis: vehicleValues,
		Data:  data.VehicleBreakdown,
	})
	
	// Performance trends line chart
	if len(data.PerformanceTrends) > 0 {
		dates := make([]string, len(data.PerformanceTrends))
		efficiency := make([]float64, len(data.PerformanceTrends))
		
		for i, trend := range data.PerformanceTrends {
			dates[i] = trend.Date.Format("2006-01-02")
			efficiency[i] = trend.Efficiency
		}
		
		charts = append(charts, ChartData{
			Type:  "line",
			Title: "Efficiency Trends",
			XAxis: dates,
			YAxis: efficiency,
			Data:  data.PerformanceTrends,
		})
	}
	
	return charts
}

// Metadata calculation methods
func (ae *AnalyticsEngine) calculateDataPoints(_ interface{}) int {
	// Simplified calculation - in production, this would count actual data points
	return 100 // Placeholder
}

func (ae *AnalyticsEngine) assessDataQuality(_ interface{}) string {
	// Simplified assessment - in production, this would analyze data completeness and accuracy
	return "high" // Placeholder
}

func (ae *AnalyticsEngine) calculateCompleteness(_ interface{}) float64 {
	// Simplified calculation - in production, this would calculate actual completeness percentage
	return 95.0 // Placeholder
}

// Placeholder methods for other report types
func (ae *AnalyticsEngine) generateDriverPerformanceAnalytics(_ context.Context, _ *AnalyticsRequest) (interface{}, AnalyticsSummary, []ChartData, error) {
	// Implementation for driver performance analytics
	return nil, AnalyticsSummary{}, nil, fmt.Errorf("not implemented yet")
}

func (ae *AnalyticsEngine) generateFuelAnalytics(_ context.Context, _ *AnalyticsRequest) (interface{}, AnalyticsSummary, []ChartData, error) {
	// Implementation for fuel analytics
	return nil, AnalyticsSummary{}, nil, fmt.Errorf("not implemented yet")
}

func (ae *AnalyticsEngine) generateMaintenanceCostsAnalytics(ctx context.Context, req *AnalyticsRequest) (interface{}, AnalyticsSummary, []ChartData, error) {
	// Get maintenance logs for the period
	var maintenanceLogs []models.MaintenanceLog
	err := ae.db.WithContext(ctx).
		Where("company_id = ? AND created_at BETWEEN ? AND ?", req.CompanyID, req.DateRange.StartDate, req.DateRange.EndDate).
		Preload("Vehicle").
		Find(&maintenanceLogs).Error
	
	if err != nil {
		return nil, AnalyticsSummary{}, nil, fmt.Errorf("failed to get maintenance logs: %w", err)
	}
	
	// Calculate metrics
	totalCost := 0.0
	costByVehicle := make(map[string]float64)
	costByType := make(map[string]float64)
	vehicleNames := make(map[string]string)
	
	for _, log := range maintenanceLogs {
		totalCost += log.Cost
		costByVehicle[log.VehicleID] += log.Cost
		costByType[log.MaintenanceType] += log.Cost
		if log.Vehicle.ID != "" {
			vehicleNames[log.VehicleID] = log.Vehicle.LicensePlate
		}
	}
	
	// Calculate average cost
	avgCost := 0.0
	if len(maintenanceLogs) > 0 {
		avgCost = totalCost / float64(len(maintenanceLogs))
	}
	
	// Find min/max costs
	minCost, maxCost := math.MaxFloat64, 0.0
	for _, log := range maintenanceLogs {
		if log.Cost < minCost {
			minCost = log.Cost
		}
		if log.Cost > maxCost {
			maxCost = log.Cost
		}
	}
	if len(maintenanceLogs) == 0 {
		minCost = 0
	}
	
	// Calculate cost per km (total cost / total distance)
	var totalDistance float64
	ae.db.Model(&models.Trip{}).
		Where("company_id = ? AND start_time BETWEEN ? AND ?", req.CompanyID, req.DateRange.StartDate, req.DateRange.EndDate).
		Select("COALESCE(SUM(total_distance), 0)").
		Scan(&totalDistance)
	
	costPerKm := 0.0
	if totalDistance > 0 {
		costPerKm = totalCost / totalDistance
	}
	
	// Identify high-cost vehicles (above average)
	highCostVehicles := []map[string]interface{}{}
	for vehicleID, cost := range costByVehicle {
		if cost > avgCost {
			highCostVehicles = append(highCostVehicles, map[string]interface{}{
				"vehicle_id":     vehicleID,
				"license_plate":  vehicleNames[vehicleID],
				"cost":          cost,
				"percentage":    (cost / totalCost) * 100,
			})
		}
	}
	
	// Generate insights
	insights := []string{
		fmt.Sprintf("Total maintenance cost: IDR %.2f", totalCost),
		fmt.Sprintf("Average cost per service: IDR %.2f", avgCost),
		fmt.Sprintf("Cost per kilometer: IDR %.2f/km", costPerKm),
	}
	
	if len(highCostVehicles) > 0 {
		insights = append(insights, fmt.Sprintf("%d vehicles have above-average maintenance costs", len(highCostVehicles)))
	}
	
	// Generate recommendations
	recommendations := []string{}
	if costPerKm > 500 { // High cost per km threshold
		recommendations = append(recommendations, "Consider reviewing maintenance schedules - cost per km is high")
	}
	if len(highCostVehicles) > 0 {
		recommendations = append(recommendations, "Consider replacing or retiring high-cost vehicles")
	}
	
	// Create summary
	summary := AnalyticsSummary{
		TotalRecords:    int64(len(maintenanceLogs)),
		TotalValue:      totalCost,
		AverageValue:    avgCost,
		MinValue:        minCost,
		MaxValue:        maxCost,
		KeyInsights:     insights,
		Recommendations: recommendations,
		Trend:           "stable",
	}
	
	// Create chart data
	charts := []ChartData{
		{
			Type:  "bar",
			Title: "Maintenance Costs by Type",
			XAxis: getKeys(costByType),
			YAxis: getValues(costByType),
		},
	}
	
	// Return data
	data := map[string]interface{}{
		"total_cost":         totalCost,
		"average_cost":       avgCost,
		"cost_per_km":        costPerKm,
		"total_services":     len(maintenanceLogs),
		"cost_by_vehicle":    costByVehicle,
		"cost_by_type":       costByType,
		"high_cost_vehicles": highCostVehicles,
	}
	
	return data, summary, charts, nil
}

func (ae *AnalyticsEngine) generateRouteEfficiencyAnalytics(ctx context.Context, req *AnalyticsRequest) (interface{}, AnalyticsSummary, []ChartData, error) {
	// Get completed trips for the period
	var trips []models.Trip
	err := ae.db.WithContext(ctx).
		Where("company_id = ? AND status = ? AND start_time BETWEEN ? AND ?", 
			req.CompanyID, "completed", req.DateRange.StartDate, req.DateRange.EndDate).
		Preload("Vehicle").
		Preload("Driver").
		Find(&trips).Error
	
	if err != nil {
		return nil, AnalyticsSummary{}, nil, fmt.Errorf("failed to get trips: %w", err)
	}
	
	// Calculate route efficiency metrics
	totalDistance := 0.0
	totalDuration := 0.0
	efficientRoutes := 0
	inefficientRoutes := 0
	
	routeEfficiency := []map[string]interface{}{}
	
	for _, trip := range trips {
		if trip.EndTime != nil && trip.StartTime != nil {
			duration := trip.EndTime.Sub(*trip.StartTime).Hours()
			totalDistance += trip.TotalDistance
			totalDuration += duration
			
			// Calculate expected time (assume average speed 40 km/h)
			expectedTime := trip.TotalDistance / 40.0
			actualTime := duration
			efficiency := (expectedTime / actualTime) * 100
			
			if efficiency >= 80 {
				efficientRoutes++
			} else {
				inefficientRoutes++
			}
			
			routeEfficiency = append(routeEfficiency, map[string]interface{}{
				"trip_id":       trip.ID,
				"vehicle":       trip.Vehicle.LicensePlate,
				"driver":        trip.Driver.FirstName + " " + trip.Driver.LastName,
				"distance":      trip.TotalDistance,
				"duration":      duration,
				"efficiency":    efficiency,
				"is_efficient":  efficiency >= 80,
			})
		}
	}
	
	// Calculate averages
	avgDistance := 0.0
	avgSpeed := 0.0
	if len(trips) > 0 {
		avgDistance = totalDistance / float64(len(trips))
	}
	if totalDuration > 0 {
		avgSpeed = totalDistance / totalDuration
	}
	
	// Calculate efficiency rate
	efficiencyRate := 0.0
	if len(trips) > 0 {
		efficiencyRate = (float64(efficientRoutes) / float64(len(trips))) * 100
	}
	
	// Generate insights
	insights := []string{
		fmt.Sprintf("%.1f%% of routes are efficient (≥80%% efficiency)", efficiencyRate),
		fmt.Sprintf("Average speed: %.1f km/h", avgSpeed),
		fmt.Sprintf("%d trips analyzed", len(trips)),
	}
	
	// Generate recommendations
	recommendations := []string{}
	if efficiencyRate < 70 {
		recommendations = append(recommendations, "Route efficiency is below target - consider route optimization")
	}
	if avgSpeed < 30 {
		recommendations = append(recommendations, "Average speed is low - check for traffic or route issues")
	}
	if inefficientRoutes > efficientRoutes {
		recommendations = append(recommendations, "More inefficient routes than efficient - driver training recommended")
	}
	
	// Create summary
	summary := AnalyticsSummary{
		TotalRecords:    int64(len(trips)),
		TotalValue:      totalDistance,
		AverageValue:    avgDistance,
		KeyInsights:     insights,
		Recommendations: recommendations,
		Trend:           "stable",
	}
	
	// Create chart data
	charts := []ChartData{
		{
			Type:  "pie",
			Title: "Route Efficiency Distribution",
			Data: map[string]int{
				"efficient":   efficientRoutes,
				"inefficient": inefficientRoutes,
			},
		},
	}
	
	// Return data
	data := map[string]interface{}{
		"total_trips":        len(trips),
		"efficient_routes":   efficientRoutes,
		"inefficient_routes": inefficientRoutes,
		"efficiency_rate":    efficiencyRate,
		"average_speed":      avgSpeed,
		"total_distance":     totalDistance,
		"total_duration":     totalDuration,
		"route_details":      routeEfficiency,
	}
	
	return data, summary, charts, nil
}

func (ae *AnalyticsEngine) generateGeofenceActivityAnalytics(ctx context.Context, req *AnalyticsRequest) (interface{}, AnalyticsSummary, []ChartData, error) {
	// Get geofence events for the period
	var events []struct {
		GeofenceID    string
		GeofenceName  string
		EventType     string
		Count         int64
	}
	
	err := ae.db.WithContext(ctx).
		Table("geofence_events").
		Select("geofence_id, event_type, COUNT(*) as count").
		Where("company_id = ? AND event_time BETWEEN ? AND ?", req.CompanyID, req.DateRange.StartDate, req.DateRange.EndDate).
		Group("geofence_id, event_type").
		Find(&events).Error
	
	if err != nil {
		return nil, AnalyticsSummary{}, nil, fmt.Errorf("failed to get geofence events: %w", err)
	}
	
	// Calculate metrics
	totalEvents := int64(0)
	entriesCount := int64(0)
	exitsCount := int64(0)
	violationsCount := int64(0)
	activityByGeofence := make(map[string]int64)
	
	for _, event := range events {
		totalEvents += event.Count
		activityByGeofence[event.GeofenceID] += event.Count
		
		switch event.EventType {
		case "entry":
			entriesCount += event.Count
		case "exit":
			exitsCount += event.Count
		case "violation":
			violationsCount += event.Count
		}
	}
	
	// Get violation count (geofence_violations table may not exist)
	var violationDetails []map[string]interface{}
	// Note: Using generic query as GeofenceViolation model may be in different package
	
	// Generate insights
	insights := []string{
		fmt.Sprintf("Total geofence events: %d", totalEvents),
		fmt.Sprintf("Entries: %d, Exits: %d", entriesCount, exitsCount),
	}
	
	if violationsCount > 0 {
		insights = append(insights, fmt.Sprintf("%d geofence violations detected", violationsCount))
	}
	
	// Generate recommendations
	recommendations := []string{}
	if violationsCount > 10 {
		recommendations = append(recommendations, "High number of geofence violations - review driver compliance")
	}
	if entriesCount != exitsCount {
		diff := entriesCount - exitsCount
		recommendations = append(recommendations, fmt.Sprintf("Mismatched entries/exits: %d vehicles may still be in geofences", diff))
	}
	
	// Create summary
	summary := AnalyticsSummary{
		TotalRecords:    totalEvents,
		TotalValue:      float64(totalEvents),
		KeyInsights:     insights,
		Recommendations: recommendations,
		Trend:           "stable",
	}
	
	// Create chart data
	charts := []ChartData{
		{
			Type:  "pie",
			Title: "Geofence Event Types",
			Data: map[string]int64{
				"entries":    entriesCount,
				"exits":      exitsCount,
				"violations": violationsCount,
			},
		},
	}
	
	// Return data
	data := map[string]interface{}{
		"total_events":         totalEvents,
		"entries":             entriesCount,
		"exits":               exitsCount,
		"violations":          violationsCount,
		"activity_by_geofence": activityByGeofence,
		"recent_violations":   violationDetails,
	}
	
	return data, summary, charts, nil
}

func (ae *AnalyticsEngine) generateComplianceReportAnalytics(_ context.Context, _ *AnalyticsRequest) (interface{}, AnalyticsSummary, []ChartData, error) {
	// Implementation for compliance report analytics
	return nil, AnalyticsSummary{}, nil, fmt.Errorf("not implemented yet")
}

func (ae *AnalyticsEngine) generateCostAnalysisAnalytics(_ context.Context, _ *AnalyticsRequest) (interface{}, AnalyticsSummary, []ChartData, error) {
	// Implementation for cost analysis analytics
	return nil, AnalyticsSummary{}, nil, fmt.Errorf("not implemented yet")
}

func (ae *AnalyticsEngine) generateUtilizationReportAnalytics(ctx context.Context, req *AnalyticsRequest) (interface{}, AnalyticsSummary, []ChartData, error) {
	// Get all vehicles for the company
	var vehicles []models.Vehicle
	err := ae.db.WithContext(ctx).
		Where("company_id = ?", req.CompanyID).
		Find(&vehicles).Error
	
	if err != nil {
		return nil, AnalyticsSummary{}, nil, fmt.Errorf("failed to get vehicles: %w", err)
	}
	
	// Calculate utilization for each vehicle
	totalPeriodHours := req.DateRange.EndDate.Sub(req.DateRange.StartDate).Hours()
	vehicleUtilization := []map[string]interface{}{}
	totalUtilizationRate := 0.0
	underutilizedCount := 0
	
	for _, vehicle := range vehicles {
		// Get total trip hours for this vehicle
		var tripDuration float64
		ae.db.Model(&models.Trip{}).
			Where("vehicle_id = ? AND status = ? AND start_time BETWEEN ? AND ?", 
				vehicle.ID, "completed", req.DateRange.StartDate, req.DateRange.EndDate).
			Select("COALESCE(SUM(EXTRACT(EPOCH FROM (end_time - start_time))/3600), 0)").
			Scan(&tripDuration)
		
		// Calculate utilization rate
		utilizationRate := 0.0
		if totalPeriodHours > 0 {
			utilizationRate = (tripDuration / totalPeriodHours) * 100
		}
		
		totalUtilizationRate += utilizationRate
		
		// Mark as underutilized if < 30%
		isUnderutilized := utilizationRate < 30
		if isUnderutilized {
			underutilizedCount++
		}
		
		vehicleUtilization = append(vehicleUtilization, map[string]interface{}{
			"vehicle_id":       vehicle.ID,
			"license_plate":    vehicle.LicensePlate,
			"utilization_rate": utilizationRate,
			"active_hours":     tripDuration,
			"idle_hours":       totalPeriodHours - tripDuration,
			"is_underutilized": isUnderutilized,
		})
	}
	
	// Calculate averages
	avgUtilization := 0.0
	if len(vehicles) > 0 {
		avgUtilization = totalUtilizationRate / float64(len(vehicles))
	}
	
	// Generate insights
	insights := []string{
		fmt.Sprintf("Average fleet utilization: %.1f%%", avgUtilization),
		fmt.Sprintf("%d vehicles analyzed", len(vehicles)),
	}
	
	if underutilizedCount > 0 {
		insights = append(insights, fmt.Sprintf("%d vehicles are underutilized (<30%% usage)", underutilizedCount))
	}
	
	// Generate recommendations
	recommendations := []string{}
	if avgUtilization < 50 {
		recommendations = append(recommendations, "Fleet utilization is low - consider reducing fleet size or increasing operations")
	}
	if underutilizedCount > len(vehicles)/2 {
		recommendations = append(recommendations, "More than half the fleet is underutilized - significant cost-saving opportunity")
	}
	if avgUtilization > 85 {
		recommendations = append(recommendations, "High utilization - consider expanding fleet to meet demand")
	}
	
	// Create summary
	summary := AnalyticsSummary{
		TotalRecords:    int64(len(vehicles)),
		AverageValue:    avgUtilization,
		KeyInsights:     insights,
		Recommendations: recommendations,
		Trend:           "stable",
	}
	
	// Create chart data
	charts := []ChartData{
		{
			Type:  "bar",
			Title: "Vehicle Utilization Rates",
			Data:  vehicleUtilization,
		},
	}
	
	// Return data
	data := map[string]interface{}{
		"average_utilization":  avgUtilization,
		"total_vehicles":       len(vehicles),
		"underutilized_count":  underutilizedCount,
		"vehicle_utilization":  vehicleUtilization,
	}
	
	return data, summary, charts, nil
}

func (ae *AnalyticsEngine) generatePredictiveInsightsAnalytics(_ context.Context, req *AnalyticsRequest) (interface{}, AnalyticsSummary, []ChartData, error) {
	// Get historical data for predictions
	historicalPeriod := req.DateRange.EndDate.Sub(req.DateRange.StartDate)
	historicalStart := req.DateRange.StartDate.Add(-historicalPeriod)
	
	// Predict fuel consumption
	var currentFuel, historicalFuel float64
	ae.db.Model(&models.FuelLog{}).
		Where("company_id = ? AND date BETWEEN ? AND ?", req.CompanyID, req.DateRange.StartDate, req.DateRange.EndDate).
		Select("COALESCE(SUM(amount), 0)").
		Scan(&currentFuel)
	
	ae.db.Model(&models.FuelLog{}).
		Where("company_id = ? AND date BETWEEN ? AND ?", req.CompanyID, historicalStart, req.DateRange.StartDate).
		Select("COALESCE(SUM(amount), 0)").
		Scan(&historicalFuel)
	
	fuelTrend := "stable"
	fuelChange := 0.0
	if historicalFuel > 0 {
		fuelChange = ((currentFuel - historicalFuel) / historicalFuel) * 100
		if fuelChange > 10 {
			fuelTrend = "increasing"
		} else if fuelChange < -10 {
			fuelTrend = "decreasing"
		}
	}
	
	// Predict next period fuel consumption (simple linear projection)
	predictedFuel := currentFuel * (1 + (fuelChange / 100))
	
	// Predict maintenance costs
	var currentMaintCost, historicalMaintCost float64
	ae.db.Model(&models.MaintenanceLog{}).
		Where("company_id = ? AND created_at BETWEEN ? AND ?", req.CompanyID, req.DateRange.StartDate, req.DateRange.EndDate).
		Select("COALESCE(SUM(cost), 0)").
		Scan(&currentMaintCost)
	
	ae.db.Model(&models.MaintenanceLog{}).
		Where("company_id = ? AND created_at BETWEEN ? AND ?", req.CompanyID, historicalStart, req.DateRange.StartDate).
		Select("COALESCE(SUM(cost), 0)").
		Scan(&historicalMaintCost)
	
	maintCostChange := 0.0
	if historicalMaintCost > 0 {
		maintCostChange = ((currentMaintCost - historicalMaintCost) / historicalMaintCost) * 100
	}
	
	predictedMaintCost := currentMaintCost * (1 + (maintCostChange / 100))
	
	// Identify vehicles needing attention
	vehiclesNeedingAttention := []map[string]interface{}{}
	
	// Get vehicle count for threshold calculation
	var vehicleCount int64
	ae.db.Model(&models.Vehicle{}).Where("company_id = ?", req.CompanyID).Count(&vehicleCount)
	
	if vehicleCount > 0 && currentMaintCost > 0 {
		// Vehicles with high maintenance costs
		var highCostVehicles []struct {
			VehicleID    string
			LicensePlate string
			TotalCost    float64
		}
		avgMaintCost := currentMaintCost / float64(vehicleCount)
		
		ae.db.Table("maintenance_logs").
			Select("vehicle_id, SUM(cost) as total_cost").
			Joins("JOIN vehicles ON maintenance_logs.vehicle_id = vehicles.id").
			Where("maintenance_logs.company_id = ? AND maintenance_logs.created_at BETWEEN ? AND ?", 
				req.CompanyID, req.DateRange.StartDate, req.DateRange.EndDate).
			Group("vehicle_id, vehicles.license_plate").
			Having("SUM(cost) > ?", avgMaintCost*1.5). // 1.5x average
			Scan(&highCostVehicles)
		
		for _, v := range highCostVehicles {
			vehiclesNeedingAttention = append(vehiclesNeedingAttention, map[string]interface{}{
				"vehicle_id":     v.VehicleID,
				"license_plate":  v.LicensePlate,
				"reason":        "high_maintenance_cost",
				"cost":          v.TotalCost,
			})
		}
	}
	
	// Calculate total cost forecast
	predictedTotalCost := predictedFuel*15000 + predictedMaintCost // Fuel + Maintenance
	
	// Generate insights
	insights := []string{
		fmt.Sprintf("Fuel consumption trend: %s (%.1f%% change)", fuelTrend, fuelChange),
		fmt.Sprintf("Predicted next period fuel: %.2f liters", predictedFuel),
		fmt.Sprintf("Predicted maintenance cost: IDR %.2f", predictedMaintCost),
		fmt.Sprintf("Predicted total cost: IDR %.2f", predictedTotalCost),
	}
	
	if len(vehiclesNeedingAttention) > 0 {
		insights = append(insights, fmt.Sprintf("%d vehicles need attention", len(vehiclesNeedingAttention)))
	}
	
	// Generate recommendations
	recommendations := []string{}
	if fuelChange > 20 {
		recommendations = append(recommendations, "Fuel consumption increasing significantly - investigate causes")
	}
	if maintCostChange > 30 {
		recommendations = append(recommendations, "Maintenance costs rising rapidly - review fleet age and condition")
	}
	if len(vehiclesNeedingAttention) > 0 {
		recommendations = append(recommendations, "Consider preventive maintenance for high-cost vehicles")
	}
	
	// Create summary
	summary := AnalyticsSummary{
		TotalValue:      predictedTotalCost,
		KeyInsights:     insights,
		Recommendations: recommendations,
		Trend:           fuelTrend,
	}
	
	// Create chart data
	charts := []ChartData{
		{
			Type:  "line",
			Title: "Fuel Consumption Trend & Forecast",
			Data: map[string]interface{}{
				"historical": historicalFuel,
				"current":    currentFuel,
				"predicted":  predictedFuel,
			},
		},
	}
	
	// Return data
	data := map[string]interface{}{
		"fuel_prediction": map[string]interface{}{
			"current":   currentFuel,
			"predicted": predictedFuel,
			"change":    fuelChange,
			"trend":     fuelTrend,
		},
		"cost_prediction": map[string]interface{}{
			"current_maintenance": currentMaintCost,
			"predicted_maintenance": predictedMaintCost,
			"change":              maintCostChange,
			"total_predicted_cost": predictedTotalCost,
		},
		"vehicles_needing_attention": vehiclesNeedingAttention,
	}
	
	return data, summary, charts, nil
}

// Helper functions for chart data generation

// getKeys extracts keys from a map[string]float64
func getKeys(m map[string]float64) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

// getValues extracts values from a map[string]float64
func getValues(m map[string]float64) []float64 {
	values := make([]float64, 0, len(m))
	for _, v := range m {
		values = append(values, v)
	}
	return values
}
