package analytics

import (
	"context"
	"time"

	"github.com/tobangado69/fleettracker-pro/backend/pkg/models"
	"gorm.io/gorm"
)

// AnalyticsRepository defines the interface for analytics data operations
type AnalyticsRepository interface {
	// Dashboard operations
	GetFleetStats(ctx context.Context, companyID string) (map[string]interface{}, error)
	GetActiveVehicleCount(ctx context.Context, companyID string) (int64, error)
	GetTotalTripsCount(ctx context.Context, companyID string, startDate, endDate time.Time) (int64, error)
	GetTotalDistance(ctx context.Context, companyID string, startDate, endDate time.Time) (float64, error)
	GetTotalFuelConsumed(ctx context.Context, companyID string, startDate, endDate time.Time) (float64, error)
	
	// Fuel analytics
	GetFuelConsumptionData(ctx context.Context, companyID string, startDate, endDate time.Time) ([]*models.FuelLog, error)
	GetFuelTrendData(ctx context.Context, companyID string, startDate, endDate time.Time, period string) ([]Trend, error)
	GetFuelCostData(ctx context.Context, companyID string, startDate, endDate time.Time) (float64, error)
	
	// Driver analytics
	GetDriverPerformanceData(ctx context.Context, driverID string, startDate, endDate time.Time) (*models.PerformanceLog, error)
	GetTopPerformers(ctx context.Context, companyID string, limit int) ([]*models.Driver, error)
	GetDriverEvents(ctx context.Context, driverID string, startDate, endDate time.Time) ([]*models.DriverEvent, error)
	
	// Vehicle analytics
	GetVehicleUtilization(ctx context.Context, companyID string) (map[string]interface{}, error)
	GetMaintenanceAlerts(ctx context.Context, companyID string) ([]MaintenanceAlert, error)
	
	// Trip analytics
	GetTripData(ctx context.Context, companyID string, startDate, endDate time.Time) ([]*models.Trip, error)
	GetTripStats(ctx context.Context, companyID string, startDate, endDate time.Time) (map[string]interface{}, error)
	
	// Compliance
	GetComplianceData(ctx context.Context, companyID string, period string) (map[string]interface{}, error)
}

// analyticsRepository implements AnalyticsRepository interface
type analyticsRepository struct {
	db *gorm.DB
}

// NewAnalyticsRepository creates a new analytics repository
func NewAnalyticsRepository(db *gorm.DB) AnalyticsRepository {
	return &analyticsRepository{db: db}
}

// GetFleetStats gets fleet statistics
func (r *analyticsRepository) GetFleetStats(ctx context.Context, companyID string) (map[string]interface{}, error) {
	var stats struct {
		TotalVehicles  int64   `gorm:"column:total_vehicles"`
		ActiveVehicles int64   `gorm:"column:active_vehicles"`
		TotalDrivers   int64   `gorm:"column:total_drivers"`
		ActiveDrivers  int64   `gorm:"column:active_drivers"`
	}
	
	query := `
		SELECT 
			(SELECT COUNT(*) FROM vehicles WHERE company_id = ?) as total_vehicles,
			(SELECT COUNT(*) FROM vehicles WHERE company_id = ? AND is_active = true) as active_vehicles,
			(SELECT COUNT(*) FROM drivers WHERE company_id = ?) as total_drivers,
			(SELECT COUNT(*) FROM drivers WHERE company_id = ? AND is_active = true) as active_drivers
	`
	
	if err := r.db.WithContext(ctx).Raw(query, companyID, companyID, companyID, companyID).Scan(&stats).Error; err != nil {
		return nil, err
	}
	
	return map[string]interface{}{
		"total_vehicles":  stats.TotalVehicles,
		"active_vehicles": stats.ActiveVehicles,
		"total_drivers":   stats.TotalDrivers,
		"active_drivers":  stats.ActiveDrivers,
	}, nil
}

// GetActiveVehicleCount gets count of active vehicles
func (r *analyticsRepository) GetActiveVehicleCount(ctx context.Context, companyID string) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).
		Model(&models.Vehicle{}).
		Where("company_id = ? AND is_active = ?", companyID, true).
		Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

// GetTotalTripsCount gets total trips count
func (r *analyticsRepository) GetTotalTripsCount(ctx context.Context, companyID string, startDate, endDate time.Time) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).
		Model(&models.Trip{}).
		Where("company_id = ? AND start_time BETWEEN ? AND ?", companyID, startDate, endDate).
		Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

// GetTotalDistance gets total distance traveled
func (r *analyticsRepository) GetTotalDistance(ctx context.Context, companyID string, startDate, endDate time.Time) (float64, error) {
	var distance float64
	if err := r.db.WithContext(ctx).
		Model(&models.Trip{}).
		Where("company_id = ? AND start_time BETWEEN ? AND ?", companyID, startDate, endDate).
		Select("COALESCE(SUM(total_distance), 0)").
		Scan(&distance).Error; err != nil {
		return 0, err
	}
	return distance, nil
}

// GetTotalFuelConsumed gets total fuel consumed
func (r *analyticsRepository) GetTotalFuelConsumed(ctx context.Context, companyID string, startDate, endDate time.Time) (float64, error) {
	var fuel float64
	if err := r.db.WithContext(ctx).
		Model(&models.FuelLog{}).
		Where("company_id = ? AND date BETWEEN ? AND ?", companyID, startDate, endDate).
		Select("COALESCE(SUM(amount), 0)").
		Scan(&fuel).Error; err != nil {
		return 0, err
	}
	return fuel, nil
}

// GetFuelConsumptionData gets fuel consumption data
func (r *analyticsRepository) GetFuelConsumptionData(ctx context.Context, companyID string, startDate, endDate time.Time) ([]*models.FuelLog, error) {
	var fuelLogs []*models.FuelLog
	if err := r.db.WithContext(ctx).
		Preload("Vehicle").
		Where("company_id = ? AND date BETWEEN ? AND ?", companyID, startDate, endDate).
		Order("date DESC").
		Find(&fuelLogs).Error; err != nil {
		return nil, err
	}
	return fuelLogs, nil
}

// GetFuelTrendData gets fuel trend data
func (r *analyticsRepository) GetFuelTrendData(ctx context.Context, companyID string, startDate, endDate time.Time, period string) ([]Trend, error) {
	// Implementation would aggregate fuel data by period
	return []Trend{}, nil
}

// GetFuelCostData gets fuel cost data
func (r *analyticsRepository) GetFuelCostData(ctx context.Context, companyID string, startDate, endDate time.Time) (float64, error) {
	var cost float64
	if err := r.db.WithContext(ctx).
		Model(&models.FuelLog{}).
		Where("company_id = ? AND date BETWEEN ? AND ?", companyID, startDate, endDate).
		Select("COALESCE(SUM(cost), 0)").
		Scan(&cost).Error; err != nil {
		return 0, err
	}
	return cost, nil
}

// GetDriverPerformanceData gets driver performance data
func (r *analyticsRepository) GetDriverPerformanceData(ctx context.Context, driverID string, startDate, endDate time.Time) (*models.PerformanceLog, error) {
	var performanceLog models.PerformanceLog
	if err := r.db.WithContext(ctx).
		Where("driver_id = ? AND date BETWEEN ? AND ?", driverID, startDate, endDate).
		Order("date DESC").
		First(&performanceLog).Error; err != nil {
		return nil, err
	}
	return &performanceLog, nil
}

// GetTopPerformers gets top performing drivers
func (r *analyticsRepository) GetTopPerformers(ctx context.Context, companyID string, limit int) ([]*models.Driver, error) {
	var drivers []*models.Driver
	if err := r.db.WithContext(ctx).
		Where("company_id = ? AND is_active = ?", companyID, true).
		Order("performance_score DESC").
		Limit(limit).
		Find(&drivers).Error; err != nil {
		return nil, err
	}
	return drivers, nil
}

// GetDriverEvents gets driver events
func (r *analyticsRepository) GetDriverEvents(ctx context.Context, driverID string, startDate, endDate time.Time) ([]*models.DriverEvent, error) {
	var events []*models.DriverEvent
	if err := r.db.WithContext(ctx).
		Where("driver_id = ? AND timestamp BETWEEN ? AND ?", driverID, startDate, endDate).
		Order("timestamp DESC").
		Find(&events).Error; err != nil {
		return nil, err
	}
	return events, nil
}

// GetVehicleUtilization gets vehicle utilization data
func (r *analyticsRepository) GetVehicleUtilization(ctx context.Context, companyID string) (map[string]interface{}, error) {
	// Implementation would calculate utilization metrics
	return map[string]interface{}{}, nil
}

// GetMaintenanceAlerts gets maintenance alerts
func (r *analyticsRepository) GetMaintenanceAlerts(ctx context.Context, companyID string) ([]MaintenanceAlert, error) {
	// Implementation would get maintenance alerts
	return []MaintenanceAlert{}, nil
}

// GetTripData gets trip data
func (r *analyticsRepository) GetTripData(ctx context.Context, companyID string, startDate, endDate time.Time) ([]*models.Trip, error) {
	var trips []*models.Trip
	if err := r.db.WithContext(ctx).
		Preload("Vehicle").
		Preload("Driver").
		Where("company_id = ? AND start_time BETWEEN ? AND ?", companyID, startDate, endDate).
		Order("start_time DESC").
		Find(&trips).Error; err != nil {
		return nil, err
	}
	return trips, nil
}

// GetTripStats gets trip statistics
func (r *analyticsRepository) GetTripStats(ctx context.Context, companyID string, startDate, endDate time.Time) (map[string]interface{}, error) {
	var stats struct {
		TotalTrips    int64   `gorm:"column:total_trips"`
		TotalDistance float64 `gorm:"column:total_distance"`
		TotalDuration int64   `gorm:"column:total_duration"`
	}
	
	query := `
		SELECT 
			COUNT(*) as total_trips,
			COALESCE(SUM(total_distance), 0) as total_distance,
			COALESCE(SUM(EXTRACT(EPOCH FROM (end_time - start_time))/60), 0) as total_duration
		FROM trips
		WHERE company_id = ? AND start_time BETWEEN ? AND ?
	`
	
	if err := r.db.WithContext(ctx).Raw(query, companyID, startDate, endDate).Scan(&stats).Error; err != nil {
		return nil, err
	}
	
	return map[string]interface{}{
		"total_trips":    stats.TotalTrips,
		"total_distance": stats.TotalDistance,
		"total_duration": stats.TotalDuration,
	}, nil
}

// GetComplianceData gets compliance data
func (r *analyticsRepository) GetComplianceData(ctx context.Context, companyID string, period string) (map[string]interface{}, error) {
	// Implementation would get compliance-related data
	return map[string]interface{}{}, nil
}

