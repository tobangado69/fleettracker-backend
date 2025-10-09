package analytics

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"sort"
	"time"

	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"

	"github.com/tobangado69/fleettracker-pro/backend/internal/common/repository"
	apperrors "github.com/tobangado69/fleettracker-pro/backend/pkg/errors"
	"github.com/tobangado69/fleettracker-pro/backend/pkg/models"
)

// Service handles analytics and reporting operations
type Service struct {
	db          *gorm.DB
	redis       *redis.Client
	repoManager *repository.RepositoryManager
	cache       *CacheService
}

// CacheService provides caching functionality for analytics operations
type CacheService struct {
	redis *redis.Client
}

// NewCacheService creates a new cache service
func NewCacheService(redis *redis.Client) *CacheService {
	return &CacheService{redis: redis}
}

// GetFleetDashboardFromCache retrieves fleet dashboard data from cache
func (cs *CacheService) GetFleetDashboardFromCache(ctx context.Context, companyID string) (*FleetDashboard, error) {
	key := fmt.Sprintf("analytics:dashboard:%s", companyID)
	
	var dashboard FleetDashboard
	data, err := cs.redis.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil // Cache miss
		}
		return nil, fmt.Errorf("failed to get fleet dashboard from cache: %w", err)
	}
	
	if err := json.Unmarshal([]byte(data), &dashboard); err != nil {
		return nil, fmt.Errorf("failed to unmarshal fleet dashboard from cache: %w", err)
	}
	
	return &dashboard, nil
}

// SetFleetDashboardInCache stores fleet dashboard data in cache
func (cs *CacheService) SetFleetDashboardInCache(ctx context.Context, companyID string, dashboard *FleetDashboard, expiration time.Duration) error {
	key := fmt.Sprintf("analytics:dashboard:%s", companyID)
	
	data, err := json.Marshal(dashboard)
	if err != nil {
		return fmt.Errorf("failed to marshal fleet dashboard for cache: %w", err)
	}
	
	if err := cs.redis.Set(ctx, key, data, expiration).Err(); err != nil {
		return fmt.Errorf("failed to set fleet dashboard in cache: %w", err)
	}
	
	return nil
}

// GetFuelAnalyticsFromCache retrieves fuel analytics data from cache
func (cs *CacheService) GetFuelAnalyticsFromCache(ctx context.Context, companyID string, startDate, endDate time.Time) (*FuelAnalytics, error) {
	cacheKey := cs.generateFuelAnalyticsCacheKey(companyID, startDate, endDate)
	
	var analytics FuelAnalytics
	data, err := cs.redis.Get(ctx, cacheKey).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil // Cache miss
		}
		return nil, fmt.Errorf("failed to get fuel analytics from cache: %w", err)
	}
	
	if err := json.Unmarshal([]byte(data), &analytics); err != nil {
		return nil, fmt.Errorf("failed to unmarshal fuel analytics from cache: %w", err)
	}
	
	return &analytics, nil
}

// SetFuelAnalyticsInCache stores fuel analytics data in cache
func (cs *CacheService) SetFuelAnalyticsInCache(ctx context.Context, companyID string, startDate, endDate time.Time, analytics *FuelAnalytics, expiration time.Duration) error {
	cacheKey := cs.generateFuelAnalyticsCacheKey(companyID, startDate, endDate)
	
	data, err := json.Marshal(analytics)
	if err != nil {
		return fmt.Errorf("failed to marshal fuel analytics for cache: %w", err)
	}
	
	if err := cs.redis.Set(ctx, cacheKey, data, expiration).Err(); err != nil {
		return fmt.Errorf("failed to set fuel analytics in cache: %w", err)
	}
	
	return nil
}

// GetDriverPerformanceFromCache retrieves driver performance data from cache
func (cs *CacheService) GetDriverPerformanceFromCache(ctx context.Context, companyID, driverID, period string) (*DriverPerformance, error) {
	cacheKey := cs.generateDriverPerformanceCacheKey(companyID, driverID, period)
	
	var performance DriverPerformance
	data, err := cs.redis.Get(ctx, cacheKey).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil // Cache miss
		}
		return nil, fmt.Errorf("failed to get driver performance from cache: %w", err)
	}
	
	if err := json.Unmarshal([]byte(data), &performance); err != nil {
		return nil, fmt.Errorf("failed to unmarshal driver performance from cache: %w", err)
	}
	
	return &performance, nil
}

// SetDriverPerformanceInCache stores driver performance data in cache
func (cs *CacheService) SetDriverPerformanceInCache(ctx context.Context, companyID, driverID, period string, performance *DriverPerformance, expiration time.Duration) error {
	cacheKey := cs.generateDriverPerformanceCacheKey(companyID, driverID, period)
	
	data, err := json.Marshal(performance)
	if err != nil {
		return fmt.Errorf("failed to marshal driver performance for cache: %w", err)
	}
	
	if err := cs.redis.Set(ctx, cacheKey, data, expiration).Err(); err != nil {
		return fmt.Errorf("failed to set driver performance in cache: %w", err)
	}
	
	return nil
}

// GetComplianceReportFromCache retrieves compliance report data from cache
func (cs *CacheService) GetComplianceReportFromCache(ctx context.Context, companyID, period string) (*ComplianceReport, error) {
	cacheKey := cs.generateComplianceReportCacheKey(companyID, period)
	
	var report ComplianceReport
	data, err := cs.redis.Get(ctx, cacheKey).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil // Cache miss
		}
		return nil, fmt.Errorf("failed to get compliance report from cache: %w", err)
	}
	
	if err := json.Unmarshal([]byte(data), &report); err != nil {
		return nil, fmt.Errorf("failed to unmarshal compliance report from cache: %w", err)
	}
	
	return &report, nil
}

// SetComplianceReportInCache stores compliance report data in cache
func (cs *CacheService) SetComplianceReportInCache(ctx context.Context, companyID, period string, report *ComplianceReport, expiration time.Duration) error {
	cacheKey := cs.generateComplianceReportCacheKey(companyID, period)
	
	data, err := json.Marshal(report)
	if err != nil {
		return fmt.Errorf("failed to marshal compliance report for cache: %w", err)
	}
	
	if err := cs.redis.Set(ctx, cacheKey, data, expiration).Err(); err != nil {
		return fmt.Errorf("failed to set compliance report in cache: %w", err)
	}
	
	return nil
}

// InvalidateAnalyticsCache removes analytics cache for a company
func (cs *CacheService) InvalidateAnalyticsCache(ctx context.Context, companyID string) error {
	patterns := []string{
		fmt.Sprintf("analytics:dashboard:%s", companyID),
		fmt.Sprintf("analytics:fuel:%s:*", companyID),
		fmt.Sprintf("analytics:driver:%s:*", companyID),
		fmt.Sprintf("analytics:compliance:%s:*", companyID),
	}
	
	for _, pattern := range patterns {
		keys, err := cs.redis.Keys(ctx, pattern).Result()
		if err != nil {
			return fmt.Errorf("failed to get analytics cache keys for pattern %s: %w", pattern, err)
		}
		
		if len(keys) > 0 {
			if err := cs.redis.Del(ctx, keys...).Err(); err != nil {
				return fmt.Errorf("failed to invalidate analytics cache for pattern %s: %w", pattern, err)
			}
		}
	}
	
	return nil
}

// generateFuelAnalyticsCacheKey creates a cache key for fuel analytics queries
func (cs *CacheService) generateFuelAnalyticsCacheKey(companyID string, startDate, endDate time.Time) string {
	return fmt.Sprintf("analytics:fuel:%s:%s:%s", companyID, startDate.Format("2006-01-02"), endDate.Format("2006-01-02"))
}

// generateDriverPerformanceCacheKey creates a cache key for driver performance queries
func (cs *CacheService) generateDriverPerformanceCacheKey(companyID, driverID, period string) string {
	return fmt.Sprintf("analytics:driver:%s:%s:%s", companyID, driverID, period)
}

// generateComplianceReportCacheKey creates a cache key for compliance report queries
func (cs *CacheService) generateComplianceReportCacheKey(companyID, period string) string {
	return fmt.Sprintf("analytics:compliance:%s:%s", companyID, period)
}

// NewService creates a new analytics service
func NewService(db *gorm.DB, redis *redis.Client, repoManager *repository.RepositoryManager) *Service {
	return &Service{
		db:          db,
		redis:       redis,
		repoManager: repoManager,
		cache:       NewCacheService(redis),
	}
}

// FuelAnalytics represents fuel consumption analytics data
type FuelAnalytics struct {
	TotalConsumed     float64   `json:"total_consumed"`
	AverageEfficiency float64   `json:"average_efficiency"`
	CostSavings       float64   `json:"cost_savings"`
	IDRCost          float64   `json:"idr_cost"`
	PPN11Cost        float64   `json:"ppn_11_cost"`
	Trends           []Trend   `json:"trends"`
	TheftAlerts      []Alert   `json:"theft_alerts"`
	OptimizationTips []string  `json:"optimization_tips"`
}

// DriverPerformance represents driver performance analytics
type DriverPerformance struct {
	DriverID         string            `json:"driver_id"`
	DriverName       string            `json:"driver_name"`
	Score           float64           `json:"score"`
	BehaviorMetrics BehaviorMetrics   `json:"behavior_metrics"`
	Trends          []Trend           `json:"trends"`
	Recommendations []string          `json:"recommendations"`
	Ranking         int               `json:"ranking"`
	ImprovementAreas []string         `json:"improvement_areas"`
}

// BehaviorMetrics represents driver behavior analysis
type BehaviorMetrics struct {
	SpeedingViolations int     `json:"speeding_violations"`
	HarshBraking      int     `json:"harsh_braking"`
	ExcessiveIdling   int     `json:"excessive_idling"`
	GeofenceViolations int    `json:"geofence_violations"`
	AverageSpeed      float64 `json:"average_speed"`
	IdleTime          int     `json:"idle_time_minutes"`
}

// FleetDashboard represents fleet operations dashboard data
type FleetDashboard struct {
	ActiveVehicles      int                    `json:"active_vehicles"`
	TotalTrips         int                    `json:"total_trips"`
	DistanceTraveled   float64                `json:"distance_traveled"`
	FuelConsumed       float64                `json:"fuel_consumed"`
	DriverEvents       int                    `json:"driver_events"`
	GeofenceViolations int                    `json:"geofence_violations"`
	UtilizationRate    float64                `json:"utilization_rate"`
	CostPerKm          float64                `json:"cost_per_km"`
	MaintenanceAlerts  []MaintenanceAlert     `json:"maintenance_alerts"`
	TopPerformers      []DriverPerformance    `json:"top_performers"`
}

// Trend represents time-series trend data
type Trend struct {
	Date  string  `json:"date"`
	Value float64 `json:"value"`
}

// Alert represents system alerts
type Alert struct {
	ID          string    `json:"id"`
	Type        string    `json:"type"`
	Message     string    `json:"message"`
	Severity    string    `json:"severity"`
	VehicleID   string    `json:"vehicle_id"`
	DriverID    string    `json:"driver_id"`
	Timestamp   time.Time `json:"timestamp"`
}

// MaintenanceAlert represents maintenance scheduling alerts
type MaintenanceAlert struct {
	VehicleID     string    `json:"vehicle_id"`
	VehicleName   string    `json:"vehicle_name"`
	AlertType     string    `json:"alert_type"`
	Message       string    `json:"message"`
	DueDate       time.Time `json:"due_date"`
	Priority      string    `json:"priority"`
}

// ComplianceReport represents Indonesian compliance reporting
type ComplianceReport struct {
	CompanyID           string                 `json:"company_id"`
	ReportPeriod        string                 `json:"report_period"`
	DriverHours         []DriverHours          `json:"driver_hours"`
	VehicleInspections  []VehicleInspection    `json:"vehicle_inspections"`
	TaxReport           TaxReport              `json:"tax_report"`
	RegulatoryCompliance RegulatoryCompliance  `json:"regulatory_compliance"`
}

// DriverHours represents driver hours tracking
type DriverHours struct {
	DriverID     string  `json:"driver_id"`
	DriverName   string  `json:"driver_name"`
	TotalHours   float64 `json:"total_hours"`
	OvertimeHours float64 `json:"overtime_hours"`
	Compliance   bool    `json:"compliance"`
}

// VehicleInspection represents vehicle inspection data
type VehicleInspection struct {
	VehicleID     string    `json:"vehicle_id"`
	VehicleName   string    `json:"vehicle_name"`
	LastInspection time.Time `json:"last_inspection"`
	NextInspection time.Time `json:"next_inspection"`
	Status        string    `json:"status"`
	Compliance    bool      `json:"compliance"`
}

// TaxReport represents Indonesian tax reporting
type TaxReport struct {
	TotalRevenue    float64 `json:"total_revenue"`
	PPN11Amount     float64 `json:"ppn_11_amount"`
	TaxableAmount   float64 `json:"taxable_amount"`
	ReportPeriod    string  `json:"report_period"`
	Compliance      bool    `json:"compliance"`
}

// RegulatoryCompliance represents regulatory compliance status
type RegulatoryCompliance struct {
	MinistryTransport bool `json:"ministry_transport"`
	DataProtection    bool `json:"data_protection"`
	LaborLaw          bool `json:"labor_law"`
	TaxCompliance     bool `json:"tax_compliance"`
	OverallCompliance bool `json:"overall_compliance"`
}

// GetFuelConsumption calculates fuel consumption analytics
func (s *Service) GetFuelConsumption(ctx context.Context, companyID string, startDate, endDate time.Time) (*FuelAnalytics, error) {
	// Try to get from cache first
	cachedAnalytics, err := s.cache.GetFuelAnalyticsFromCache(ctx, companyID, startDate, endDate)
	if err != nil {
		// Log cache error but continue with database lookup
		fmt.Printf("Cache error for fuel analytics %s: %v\n", companyID, err)
	}
	
	if cachedAnalytics != nil {
		return cachedAnalytics, nil
	}

	// Get GPS tracks for the date range
	filters := repository.FilterOptions{
		CompanyID: companyID,
		DateRange: map[string]repository.DateRange{
			"timestamp": {
				Start: startDate.Format("2006-01-02"),
				End:   endDate.Format("2006-01-02"),
			},
		},
	}

	gpsTracks, err := s.repoManager.GetGPSTracks().List(ctx, filters, repository.Pagination{Page: 1, PageSize: 10000})
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to get GPS tracks")
	}

	// Calculate fuel consumption metrics
	totalDistance := 0.0
	totalFuel := 0.0
	fuelEfficiency := 0.0

	// Group by vehicle and calculate fuel consumption
	vehicleFuel := make(map[string]float64)
	vehicleDistance := make(map[string]float64)

	for _, track := range gpsTracks {
		if track.FuelLevel > 0 {
			vehicleFuel[track.VehicleID] += track.FuelLevel
		}
		if track.Distance > 0 {
			vehicleDistance[track.VehicleID] += track.Distance
		}
	}

	// Calculate total metrics
	for vehicleID := range vehicleFuel {
		totalFuel += vehicleFuel[vehicleID]
		totalDistance += vehicleDistance[vehicleID]
	}

	if totalDistance > 0 {
		fuelEfficiency = totalDistance / totalFuel // km/liter
	}

	// Calculate IDR costs (assuming 1 liter = 15,000 IDR)
	fuelPricePerLiter := 15000.0
	idrCost := totalFuel * fuelPricePerLiter
	ppn11Cost := idrCost * 0.11 // PPN 11%

	// Generate trends (simplified - in real implementation, group by date)
	trends := s.generateFuelTrends(gpsTracks)

	// Detect fuel theft (simplified algorithm)
	theftAlerts := s.detectFuelTheft(gpsTracks)

	// Generate optimization tips
	optimizationTips := s.generateFuelOptimizationTips(fuelEfficiency, totalFuel)

	analytics := &FuelAnalytics{
		TotalConsumed:     totalFuel,
		AverageEfficiency: fuelEfficiency,
		CostSavings:       s.calculateCostSavings(fuelEfficiency),
		IDRCost:          idrCost,
		PPN11Cost:        ppn11Cost,
		Trends:           trends,
		TheftAlerts:      theftAlerts,
		OptimizationTips: optimizationTips,
	}

	// Cache the result for 30 minutes (fuel analytics don't change frequently)
	if err := s.cache.SetFuelAnalyticsInCache(ctx, companyID, startDate, endDate, analytics, 30*time.Minute); err != nil {
		// Log cache error but don't fail the request
		fmt.Printf("Failed to cache fuel analytics %s: %v\n", companyID, err)
	}

	return analytics, nil
}

// GetDriverPerformance calculates driver performance analytics
func (s *Service) GetDriverPerformance(ctx context.Context, companyID string, driverID string, period string) (*DriverPerformance, error) {
	// Try to get from cache first
	cachedPerformance, err := s.cache.GetDriverPerformanceFromCache(ctx, companyID, driverID, period)
	if err != nil {
		// Log cache error but continue with database lookup
		fmt.Printf("Cache error for driver performance %s: %v\n", driverID, err)
	}
	
	if cachedPerformance != nil {
		return cachedPerformance, nil
	}

	// Get driver information
	driver, err := s.repoManager.GetDrivers().GetByID(ctx, driverID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.NewNotFoundError("driver")
		}
		return nil, apperrors.Wrap(err, "failed to get driver")
	}

	// Get GPS tracks for the driver
	filters := repository.FilterOptions{
		CompanyID: companyID,
		Where: map[string]interface{}{
			"driver_id": driverID,
		},
	}

	gpsTracks, err := s.repoManager.GetGPSTracks().List(ctx, filters, repository.Pagination{Page: 1, PageSize: 10000})
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to get GPS tracks")
	}

	// Calculate behavior metrics
	behaviorMetrics := s.calculateBehaviorMetrics(gpsTracks)

	// Calculate performance score (0-100)
	score := s.calculateDriverScore(behaviorMetrics)

	// Generate recommendations
	recommendations := s.generateDriverRecommendations(behaviorMetrics, score)

	// Generate improvement areas
	improvementAreas := s.identifyImprovementAreas(behaviorMetrics)

	// Generate trends
	trends := s.generateDriverTrends(gpsTracks)

	performance := &DriverPerformance{
		DriverID:         driverID,
		DriverName:       driver.FirstName + " " + driver.LastName,
		Score:           score,
		BehaviorMetrics: behaviorMetrics,
		Trends:          trends,
		Recommendations: recommendations,
		ImprovementAreas: improvementAreas,
	}

	// Cache the result for 20 minutes (driver performance changes moderately)
	if err := s.cache.SetDriverPerformanceInCache(ctx, companyID, driverID, period, performance, 20*time.Minute); err != nil {
		// Log cache error but don't fail the request
		fmt.Printf("Failed to cache driver performance %s: %v\n", driverID, err)
	}

	return performance, nil
}

// GetFleetDashboard generates fleet operations dashboard data
func (s *Service) GetFleetDashboard(ctx context.Context, companyID string) (*FleetDashboard, error) {
	// Try to get from cache first
	cachedDashboard, err := s.cache.GetFleetDashboardFromCache(ctx, companyID)
	if err != nil {
		// Log cache error but continue with database lookup
		fmt.Printf("Cache error for fleet dashboard %s: %v\n", companyID, err)
	}
	
	if cachedDashboard != nil {
		return cachedDashboard, nil
	}

	// Get active vehicles
	vehicleFilters := repository.FilterOptions{
		CompanyID: companyID,
		Where: map[string]interface{}{
			"status": "active",
		},
	}
	vehicles, err := s.repoManager.GetVehicles().List(ctx, vehicleFilters, repository.Pagination{Page: 1, PageSize: 1000})
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to get vehicles")
	}

	// Get recent trips
	tripFilters := repository.FilterOptions{
		CompanyID: companyID,
	}
	trips, err := s.repoManager.GetTrips().List(ctx, tripFilters, repository.Pagination{Page: 1, PageSize: 1000})
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to get trips")
	}

	// Calculate metrics
	activeVehicles := len(vehicles)
	totalTrips := len(trips)
	
	distanceTraveled := 0.0
	fuelConsumed := 0.0
	driverEvents := 0

	for _, trip := range trips {
		if trip.TotalDistance > 0 {
			distanceTraveled += trip.TotalDistance
		}
		if trip.FuelConsumed > 0 {
			fuelConsumed += trip.FuelConsumed
		}
	}

	// Calculate utilization rate
	utilizationRate := 0.0
	if activeVehicles > 0 {
		utilizationRate = float64(totalTrips) / float64(activeVehicles) * 100
	}

	// Calculate cost per km (simplified)
	costPerKm := 0.0
	if distanceTraveled > 0 {
		costPerKm = (fuelConsumed * 15000) / distanceTraveled // IDR per km
	}

	// Get maintenance alerts
	maintenanceAlerts := s.getMaintenanceAlerts(ctx, companyID)

	// Get top performers
	topPerformers := s.getTopPerformers(ctx, companyID)

	dashboard := &FleetDashboard{
		ActiveVehicles:      activeVehicles,
		TotalTrips:         totalTrips,
		DistanceTraveled:   distanceTraveled,
		FuelConsumed:       fuelConsumed,
		DriverEvents:       driverEvents,
		GeofenceViolations: 0, // TODO: Calculate from GPS data
		UtilizationRate:    utilizationRate,
		CostPerKm:          costPerKm,
		MaintenanceAlerts:  maintenanceAlerts,
		TopPerformers:      topPerformers,
	}

	// Cache the result for 10 minutes (dashboard data changes frequently)
	if err := s.cache.SetFleetDashboardInCache(ctx, companyID, dashboard, 10*time.Minute); err != nil {
		// Log cache error but don't fail the request
		fmt.Printf("Failed to cache fleet dashboard %s: %v\n", companyID, err)
	}

	return dashboard, nil
}

// GetComplianceReport generates Indonesian compliance report
func (s *Service) GetComplianceReport(ctx context.Context, companyID string, period string) (*ComplianceReport, error) {
	// Try to get from cache first
	cachedReport, err := s.cache.GetComplianceReportFromCache(ctx, companyID, period)
	if err != nil {
		// Log cache error but continue with database lookup
		fmt.Printf("Cache error for compliance report %s: %v\n", companyID, err)
	}
	
	if cachedReport != nil {
		return cachedReport, nil
	}

	// Get driver hours
	driverHours := s.calculateDriverHours(ctx, companyID, period)

	// Get vehicle inspections
	vehicleInspections := s.getVehicleInspections(ctx, companyID)

	// Generate tax report
	taxReport := s.generateTaxReport(ctx, companyID, period)

	// Check regulatory compliance
	regulatoryCompliance := s.checkRegulatoryCompliance(driverHours, vehicleInspections, taxReport)

	report := &ComplianceReport{
		CompanyID:           companyID,
		ReportPeriod:        period,
		DriverHours:         driverHours,
		VehicleInspections:  vehicleInspections,
		TaxReport:           taxReport,
		RegulatoryCompliance: regulatoryCompliance,
	}

	// Cache the result for 1 hour (compliance reports don't change frequently)
	if err := s.cache.SetComplianceReportInCache(ctx, companyID, period, report, 1*time.Hour); err != nil {
		// Log cache error but don't fail the request
		fmt.Printf("Failed to cache compliance report %s: %v\n", companyID, err)
	}

	return report, nil
}

// Helper methods

func (s *Service) calculateBehaviorMetrics(gpsTracks []*models.GPSTrack) BehaviorMetrics {
	speedingViolations := 0
	harshBraking := 0
	excessiveIdling := 0
	totalSpeed := 0.0
	idleTime := 0

	for _, track := range gpsTracks {
		if track.Speed > 0 {
			totalSpeed += track.Speed
			if track.Speed > 80 { // Speeding threshold
				speedingViolations++
			}
		}
		// Note: Acceleration field not available in current model
		// if track.Acceleration != nil && *track.Acceleration < -5 { // Harsh braking
		//	harshBraking++
		// }
		if track.Speed > 0 && track.Speed < 5 { // Idling
			idleTime++
		}
	}

	averageSpeed := 0.0
	if len(gpsTracks) > 0 {
		averageSpeed = totalSpeed / float64(len(gpsTracks))
	}

	return BehaviorMetrics{
		SpeedingViolations:  speedingViolations,
		HarshBraking:       harshBraking,
		ExcessiveIdling:    excessiveIdling,
		GeofenceViolations: 0, // TODO: Calculate from geofence data
		AverageSpeed:       averageSpeed,
		IdleTime:           idleTime,
	}
}

func (s *Service) calculateDriverScore(metrics BehaviorMetrics) float64 {
	score := 100.0

	// Deduct points for violations
	score -= float64(metrics.SpeedingViolations) * 2
	score -= float64(metrics.HarshBraking) * 3
	score -= float64(metrics.ExcessiveIdling) * 0.5
	score -= float64(metrics.GeofenceViolations) * 5

	// Ensure score is between 0 and 100
	if score < 0 {
		score = 0
	}
	if score > 100 {
		score = 100
	}

	return math.Round(score*100) / 100
}

func (s *Service) generateDriverRecommendations(metrics BehaviorMetrics, score float64) []string {
	recommendations := []string{}

	if metrics.SpeedingViolations > 5 {
		recommendations = append(recommendations, "Focus on maintaining speed limits to improve safety score")
	}
	if metrics.HarshBraking > 3 {
		recommendations = append(recommendations, "Practice smoother braking techniques to reduce wear and improve fuel efficiency")
	}
	if metrics.ExcessiveIdling > 10 {
		recommendations = append(recommendations, "Reduce idling time to improve fuel efficiency and reduce emissions")
	}
	if score < 70 {
		recommendations = append(recommendations, "Consider additional driver training to improve overall performance")
	}

	return recommendations
}

func (s *Service) identifyImprovementAreas(metrics BehaviorMetrics) []string {
	areas := []string{}

	if metrics.SpeedingViolations > 0 {
		areas = append(areas, "Speed Management")
	}
	if metrics.HarshBraking > 0 {
		areas = append(areas, "Braking Technique")
	}
	if metrics.ExcessiveIdling > 5 {
		areas = append(areas, "Idle Time Management")
	}
	if metrics.GeofenceViolations > 0 {
		areas = append(areas, "Route Compliance")
	}

	return areas
}

func (s *Service) generateFuelTrends(gpsTracks []*models.GPSTrack) []Trend {
	// Simplified trend generation - in real implementation, group by date
	trends := []Trend{}
	
	// Group tracks by date and calculate daily fuel consumption
	dateFuel := make(map[string]float64)
	for _, track := range gpsTracks {
		if track.FuelLevel > 0 {
			date := track.Timestamp.Format("2006-01-02")
			dateFuel[date] += track.FuelLevel
		}
	}

	// Convert to trends
	for date, fuel := range dateFuel {
		trends = append(trends, Trend{
			Date:  date,
			Value: fuel,
		})
	}

	// Sort by date
	sort.Slice(trends, func(i, j int) bool {
		return trends[i].Date < trends[j].Date
	})

	return trends
}

func (s *Service) detectFuelTheft(gpsTracks []*models.GPSTrack) []Alert {
	alerts := []Alert{}
	
	// Simplified fuel theft detection
	// In real implementation, this would analyze fuel level drops without corresponding distance
	vehicleFuel := make(map[string][]float64)
	
	for _, track := range gpsTracks {
		if track.FuelLevel > 0 {
			vehicleFuel[track.VehicleID] = append(vehicleFuel[track.VehicleID], track.FuelLevel)
		}
	}

	// Check for suspicious fuel drops
	for vehicleID, fuelLevels := range vehicleFuel {
		if len(fuelLevels) > 1 {
			for i := 1; i < len(fuelLevels); i++ {
				if fuelLevels[i-1] - fuelLevels[i] > 20 { // Significant fuel drop
					alerts = append(alerts, Alert{
						ID:        fmt.Sprintf("fuel-theft-%s-%d", vehicleID, i),
						Type:      "fuel_theft",
						Message:   "Suspicious fuel level drop detected",
						Severity:  "high",
						VehicleID: vehicleID,
						Timestamp: time.Now(),
					})
				}
			}
		}
	}

	return alerts
}

func (s *Service) generateFuelOptimizationTips(efficiency float64, totalFuel float64) []string {
	tips := []string{}

	if efficiency < 10 {
		tips = append(tips, "Consider vehicle maintenance to improve fuel efficiency")
		tips = append(tips, "Train drivers on fuel-efficient driving techniques")
	}
	if efficiency < 15 {
		tips = append(tips, "Monitor tire pressure regularly")
		tips = append(tips, "Plan routes to avoid traffic congestion")
	}
	if totalFuel > 1000 {
		tips = append(tips, "Consider fleet optimization to reduce fuel consumption")
	}

	return tips
}

func (s *Service) calculateCostSavings(efficiency float64) float64 {
	// Simplified cost savings calculation
	// In real implementation, this would compare against industry benchmarks
	benchmarkEfficiency := 15.0 // km/liter
	if efficiency > benchmarkEfficiency {
		return (efficiency - benchmarkEfficiency) * 1000 * 15000 // IDR savings
	}
	return 0
}

func (s *Service) generateDriverTrends(gpsTracks []*models.GPSTrack) []Trend {
	// Simplified trend generation for driver performance
	trends := []Trend{}
	
	// Group by date and calculate daily performance score
	dateScores := make(map[string][]float64)
	for _, track := range gpsTracks {
		date := track.Timestamp.Format("2006-01-02")
		score := 100.0 // Base score
		if track.Speed > 80 {
			score -= 2
		}
		dateScores[date] = append(dateScores[date], score)
	}

	// Calculate average daily scores
	for date, scores := range dateScores {
		avgScore := 0.0
		for _, score := range scores {
			avgScore += score
		}
		avgScore /= float64(len(scores))
		
		trends = append(trends, Trend{
			Date:  date,
			Value: avgScore,
		})
	}

	// Sort by date
	sort.Slice(trends, func(i, j int) bool {
		return trends[i].Date < trends[j].Date
	})

	return trends
}

func (s *Service) getMaintenanceAlerts(_ context.Context, _ string) []MaintenanceAlert {
	// Simplified maintenance alerts
	// In real implementation, this would check actual maintenance schedules
	alerts := []MaintenanceAlert{
		{
			VehicleID:   "vehicle-1",
			VehicleName: "Truck-001",
			AlertType:   "inspection_due",
			Message:     "Annual inspection due in 30 days",
			DueDate:     time.Now().AddDate(0, 0, 30),
			Priority:    "medium",
		},
	}
	return alerts
}

func (s *Service) getTopPerformers(_ context.Context, _ string) []DriverPerformance {
	// Simplified top performers
	// In real implementation, this would query actual driver performance data
	performers := []DriverPerformance{
		{
			DriverID:   "driver-1",
			DriverName: "John Doe",
			Score:      95.5,
		},
		{
			DriverID:   "driver-2",
			DriverName: "Jane Smith",
			Score:      92.3,
		},
	}
	return performers
}

func (s *Service) calculateDriverHours(_ context.Context, _ string, _ string) []DriverHours {
	// Simplified driver hours calculation
	// In real implementation, this would calculate actual working hours
	hours := []DriverHours{
		{
			DriverID:     "driver-1",
			DriverName:   "John Doe",
			TotalHours:   40.0,
			OvertimeHours: 5.0,
			Compliance:   true,
		},
	}
	return hours
}

func (s *Service) getVehicleInspections(_ context.Context, _ string) []VehicleInspection {
	// Simplified vehicle inspections
	// In real implementation, this would check actual inspection records
	inspections := []VehicleInspection{
		{
			VehicleID:      "vehicle-1",
			VehicleName:    "Truck-001",
			LastInspection: time.Now().AddDate(0, -6, 0),
			NextInspection: time.Now().AddDate(0, 6, 0),
			Status:         "valid",
			Compliance:     true,
		},
	}
	return inspections
}

func (s *Service) generateTaxReport(_ context.Context, _ string, period string) TaxReport {
	// Simplified tax report
	// In real implementation, this would calculate actual revenue and taxes
	totalRevenue := 100000000.0 // 100M IDR
	ppn11Amount := totalRevenue * 0.11
	
	return TaxReport{
		TotalRevenue:  totalRevenue,
		PPN11Amount:   ppn11Amount,
		TaxableAmount: totalRevenue,
		ReportPeriod:  period,
		Compliance:    true,
	}
}

func (s *Service) checkRegulatoryCompliance(_ []DriverHours, _ []VehicleInspection, taxReport TaxReport) RegulatoryCompliance {
	// Simplified compliance check
	// In real implementation, this would check actual compliance status
	return RegulatoryCompliance{
		MinistryTransport: true,
		DataProtection:    true,
		LaborLaw:          true,
		TaxCompliance:     taxReport.Compliance,
		OverallCompliance: true,
	}
}
