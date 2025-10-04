package repository

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/tobangado69/fleettracker-pro/backend/pkg/models"
)

// TripRepositoryImpl implements the TripRepository interface
type TripRepositoryImpl struct {
	*BaseRepository[models.Trip]
}

// NewTripRepository creates a new trip repository
func NewTripRepository(db *gorm.DB) TripRepository {
	return &TripRepositoryImpl{
		BaseRepository: NewBaseRepository[models.Trip](db),
	}
}

// GetByCompany retrieves trips by company ID with pagination
func (r *TripRepositoryImpl) GetByCompany(ctx context.Context, companyID string, pagination Pagination) ([]*models.Trip, error) {
	var trips []*models.Trip
	query := r.db.WithContext(ctx).Where("company_id = ?", companyID).Order("start_time DESC")
	
	// Apply pagination
	query = r.applyPagination(query, pagination)
	
	if err := query.Find(&trips).Error; err != nil {
		return nil, fmt.Errorf("failed to get trips by company: %w", err)
	}
	
	return trips, nil
}

// GetByVehicle retrieves trips by vehicle ID with pagination
func (r *TripRepositoryImpl) GetByVehicle(ctx context.Context, vehicleID string, pagination Pagination) ([]*models.Trip, error) {
	var trips []*models.Trip
	query := r.db.WithContext(ctx).Where("vehicle_id = ?", vehicleID).Order("start_time DESC")
	
	// Apply pagination
	query = r.applyPagination(query, pagination)
	
	if err := query.Find(&trips).Error; err != nil {
		return nil, fmt.Errorf("failed to get trips by vehicle: %w", err)
	}
	
	return trips, nil
}

// GetByDriver retrieves trips by driver ID with pagination
func (r *TripRepositoryImpl) GetByDriver(ctx context.Context, driverID string, pagination Pagination) ([]*models.Trip, error) {
	var trips []*models.Trip
	query := r.db.WithContext(ctx).Where("driver_id = ?", driverID).Order("start_time DESC")
	
	// Apply pagination
	query = r.applyPagination(query, pagination)
	
	if err := query.Find(&trips).Error; err != nil {
		return nil, fmt.Errorf("failed to get trips by driver: %w", err)
	}
	
	return trips, nil
}

// GetByStatus retrieves trips by status within a company
func (r *TripRepositoryImpl) GetByStatus(ctx context.Context, companyID string, status string) ([]*models.Trip, error) {
	var trips []*models.Trip
	if err := r.db.WithContext(ctx).Where("company_id = ? AND status = ?", companyID, status).Order("start_time DESC").Find(&trips).Error; err != nil {
		return nil, fmt.Errorf("failed to get trips by status: %w", err)
	}
	return trips, nil
}

// GetByDateRange retrieves trips by date range within a company
func (r *TripRepositoryImpl) GetByDateRange(ctx context.Context, companyID string, startDate, endDate string) ([]*models.Trip, error) {
	var trips []*models.Trip
	query := r.db.WithContext(ctx).Where("company_id = ?", companyID)
	
	if startDate != "" {
		query = query.Where("start_time >= ?", startDate)
	}
	if endDate != "" {
		query = query.Where("start_time <= ?", endDate)
	}
	
	query = query.Order("start_time DESC")
	
	if err := query.Find(&trips).Error; err != nil {
		return nil, fmt.Errorf("failed to get trips by date range: %w", err)
	}
	
	return trips, nil
}

// GetActiveTrips retrieves currently active trips for a company
func (r *TripRepositoryImpl) GetActiveTrips(ctx context.Context, companyID string) ([]*models.Trip, error) {
	var trips []*models.Trip
	if err := r.db.WithContext(ctx).Where("company_id = ? AND status = ?", companyID, "active").Order("start_time DESC").Find(&trips).Error; err != nil {
		return nil, fmt.Errorf("failed to get active trips: %w", err)
	}
	return trips, nil
}

// GetCompletedTrips retrieves completed trips for a company with pagination
func (r *TripRepositoryImpl) GetCompletedTrips(ctx context.Context, companyID string, pagination Pagination) ([]*models.Trip, error) {
	var trips []*models.Trip
	query := r.db.WithContext(ctx).Where("company_id = ? AND status = ?", companyID, "completed").Order("end_time DESC")
	
	// Apply pagination
	query = r.applyPagination(query, pagination)
	
	if err := query.Find(&trips).Error; err != nil {
		return nil, fmt.Errorf("failed to get completed trips: %w", err)
	}
	
	return trips, nil
}

// StartTrip creates a new trip
func (r *TripRepositoryImpl) StartTrip(ctx context.Context, trip *models.Trip) error {
	if err := r.db.WithContext(ctx).Create(trip).Error; err != nil {
		return fmt.Errorf("failed to start trip: %w", err)
	}
	return nil
}

// EndTrip updates a trip with end data
func (r *TripRepositoryImpl) EndTrip(ctx context.Context, tripID string, endData map[string]interface{}) error {
	if err := r.db.WithContext(ctx).Model(&models.Trip{}).Where("id = ?", tripID).Updates(endData).Error; err != nil {
		return fmt.Errorf("failed to end trip: %w", err)
	}
	return nil
}

// GetTripStatistics retrieves trip statistics for a company
func (r *TripRepositoryImpl) GetTripStatistics(ctx context.Context, companyID string, dateRange DateRange) (map[string]interface{}, error) {
	var stats struct {
		TotalTrips       int64   `json:"total_trips"`
		ActiveTrips      int64   `json:"active_trips"`
		CompletedTrips   int64   `json:"completed_trips"`
		CancelledTrips   int64   `json:"cancelled_trips"`
		TotalDistance    float64 `json:"total_distance"`
		TotalDuration    int64   `json:"total_duration"`
		AverageDistance  float64 `json:"average_distance"`
		AverageDuration  float64 `json:"average_duration"`
	}

	// Build base query with date range
	query := r.db.WithContext(ctx).Model(&models.Trip{}).Where("company_id = ?", companyID)
	
	if dateRange.Start != "" {
		query = query.Where("start_time >= ?", dateRange.Start)
	}
	if dateRange.End != "" {
		query = query.Where("start_time <= ?", dateRange.End)
	}

	// Get total trips
	if err := query.Count(&stats.TotalTrips).Error; err != nil {
		return nil, fmt.Errorf("failed to count total trips: %w", err)
	}

	// Get active trips
	if err := query.Where("status = ?", "active").Count(&stats.ActiveTrips).Error; err != nil {
		return nil, fmt.Errorf("failed to count active trips: %w", err)
	}

	// Get completed trips
	if err := query.Where("status = ?", "completed").Count(&stats.CompletedTrips).Error; err != nil {
		return nil, fmt.Errorf("failed to count completed trips: %w", err)
	}

	// Get cancelled trips
	if err := query.Where("status = ?", "cancelled").Count(&stats.CancelledTrips).Error; err != nil {
		return nil, fmt.Errorf("failed to count cancelled trips: %w", err)
	}

	// Get distance and duration statistics
	var distanceStats struct {
		Total   float64
		Average float64
	}
	if err := query.Where("status = ?", "completed").Select("SUM(distance) as total, AVG(distance) as average").Scan(&distanceStats).Error; err != nil {
		return nil, fmt.Errorf("failed to get distance statistics: %w", err)
	}
	stats.TotalDistance = distanceStats.Total
	stats.AverageDistance = distanceStats.Average

	// Get duration statistics
	var durationStats struct {
		Total   int64
		Average float64
	}
	if err := query.Where("status = ?", "completed").Select("SUM(EXTRACT(EPOCH FROM (end_time - start_time))) as total, AVG(EXTRACT(EPOCH FROM (end_time - start_time))) as average").Scan(&durationStats).Error; err != nil {
		return nil, fmt.Errorf("failed to get duration statistics: %w", err)
	}
	stats.TotalDuration = durationStats.Total
	stats.AverageDuration = durationStats.Average

	return map[string]interface{}{
		"total_trips":      stats.TotalTrips,
		"active_trips":     stats.ActiveTrips,
		"completed_trips":  stats.CompletedTrips,
		"cancelled_trips":  stats.CancelledTrips,
		"total_distance":   stats.TotalDistance,
		"total_duration":   stats.TotalDuration,
		"average_distance": stats.AverageDistance,
		"average_duration": stats.AverageDuration,
	}, nil
}

// GetTripsByDistanceRange retrieves trips within a distance range
func (r *TripRepositoryImpl) GetTripsByDistanceRange(ctx context.Context, companyID string, minDistance, maxDistance float64, pagination Pagination) ([]*models.Trip, error) {
	var trips []*models.Trip
	query := r.db.WithContext(ctx).Where("company_id = ? AND status = ?", companyID, "completed")
	
	if minDistance >= 0 {
		query = query.Where("distance >= ?", minDistance)
	}
	if maxDistance > 0 {
		query = query.Where("distance <= ?", maxDistance)
	}
	
	query = query.Order("start_time DESC")
	
	// Apply pagination
	query = r.applyPagination(query, pagination)
	
	if err := query.Find(&trips).Error; err != nil {
		return nil, fmt.Errorf("failed to get trips by distance range: %w", err)
	}
	
	return trips, nil
}

// GetTripsByDurationRange retrieves trips within a duration range
func (r *TripRepositoryImpl) GetTripsByDurationRange(ctx context.Context, companyID string, minDuration, maxDuration int64, pagination Pagination) ([]*models.Trip, error) {
	var trips []*models.Trip
	query := r.db.WithContext(ctx).Where("company_id = ? AND status = ?", companyID, "completed")
	
	if minDuration >= 0 {
		query = query.Where("EXTRACT(EPOCH FROM (end_time - start_time)) >= ?", minDuration)
	}
	if maxDuration > 0 {
		query = query.Where("EXTRACT(EPOCH FROM (end_time - start_time)) <= ?", maxDuration)
	}
	
	query = query.Order("start_time DESC")
	
	// Apply pagination
	query = r.applyPagination(query, pagination)
	
	if err := query.Find(&trips).Error; err != nil {
		return nil, fmt.Errorf("failed to get trips by duration range: %w", err)
	}
	
	return trips, nil
}

// GetTripsByFuelConsumption retrieves trips within a fuel consumption range
func (r *TripRepositoryImpl) GetTripsByFuelConsumption(ctx context.Context, companyID string, minFuel, maxFuel float64, pagination Pagination) ([]*models.Trip, error) {
	var trips []*models.Trip
	query := r.db.WithContext(ctx).Where("company_id = ? AND status = ?", companyID, "completed")
	
	if minFuel >= 0 {
		query = query.Where("fuel_consumed >= ?", minFuel)
	}
	if maxFuel > 0 {
		query = query.Where("fuel_consumed <= ?", maxFuel)
	}
	
	query = query.Order("start_time DESC")
	
	// Apply pagination
	query = r.applyPagination(query, pagination)
	
	if err := query.Find(&trips).Error; err != nil {
		return nil, fmt.Errorf("failed to get trips by fuel consumption: %w", err)
	}
	
	return trips, nil
}

// GetTopTripsByDistance retrieves the top trips by distance
func (r *TripRepositoryImpl) GetTopTripsByDistance(ctx context.Context, companyID string, limit int) ([]*models.Trip, error) {
	var trips []*models.Trip
	if err := r.db.WithContext(ctx).Where("company_id = ? AND status = ?", companyID, "completed").Order("distance DESC").Limit(limit).Find(&trips).Error; err != nil {
		return nil, fmt.Errorf("failed to get top trips by distance: %w", err)
	}
	return trips, nil
}

// GetTopTripsByDuration retrieves the top trips by duration
func (r *TripRepositoryImpl) GetTopTripsByDuration(ctx context.Context, companyID string, limit int) ([]*models.Trip, error) {
	var trips []*models.Trip
	if err := r.db.WithContext(ctx).Where("company_id = ? AND status = ?", companyID, "completed").Order("EXTRACT(EPOCH FROM (end_time - start_time)) DESC").Limit(limit).Find(&trips).Error; err != nil {
		return nil, fmt.Errorf("failed to get top trips by duration: %w", err)
	}
	return trips, nil
}

// GetTripsByAverageSpeed retrieves trips within an average speed range
func (r *TripRepositoryImpl) GetTripsByAverageSpeed(ctx context.Context, companyID string, minSpeed, maxSpeed float64, pagination Pagination) ([]*models.Trip, error) {
	var trips []*models.Trip
	query := r.db.WithContext(ctx).Where("company_id = ? AND status = ?", companyID, "completed")
	
	if minSpeed >= 0 {
		query = query.Where("average_speed >= ?", minSpeed)
	}
	if maxSpeed > 0 {
		query = query.Where("average_speed <= ?", maxSpeed)
	}
	
	query = query.Order("start_time DESC")
	
	// Apply pagination
	query = r.applyPagination(query, pagination)
	
	if err := query.Find(&trips).Error; err != nil {
		return nil, fmt.Errorf("failed to get trips by average speed: %w", err)
	}
	
	return trips, nil
}

// GetTripsByMaxSpeed retrieves trips within a max speed range
func (r *TripRepositoryImpl) GetTripsByMaxSpeed(ctx context.Context, companyID string, minSpeed, maxSpeed float64, pagination Pagination) ([]*models.Trip, error) {
	var trips []*models.Trip
	query := r.db.WithContext(ctx).Where("company_id = ? AND status = ?", companyID, "completed")
	
	if minSpeed >= 0 {
		query = query.Where("max_speed >= ?", minSpeed)
	}
	if maxSpeed > 0 {
		query = query.Where("max_speed <= ?", maxSpeed)
	}
	
	query = query.Order("start_time DESC")
	
	// Apply pagination
	query = r.applyPagination(query, pagination)
	
	if err := query.Find(&trips).Error; err != nil {
		return nil, fmt.Errorf("failed to get trips by max speed: %w", err)
	}
	
	return trips, nil
}

// GetTripsWithViolations retrieves trips with speed violations
func (r *TripRepositoryImpl) GetTripsWithViolations(ctx context.Context, companyID string, maxSpeedLimit float64, pagination Pagination) ([]*models.Trip, error) {
	var trips []*models.Trip
	query := r.db.WithContext(ctx).Where("company_id = ? AND status = ? AND max_speed > ?", companyID, "completed", maxSpeedLimit).Order("start_time DESC")
	
	// Apply pagination
	query = r.applyPagination(query, pagination)
	
	if err := query.Find(&trips).Error; err != nil {
		return nil, fmt.Errorf("failed to get trips with violations: %w", err)
	}
	
	return trips, nil
}

// GetTripsByDriverPerformance retrieves trips by driver performance score
func (r *TripRepositoryImpl) GetTripsByDriverPerformance(ctx context.Context, companyID string, minScore, maxScore float64, pagination Pagination) ([]*models.Trip, error) {
	var trips []*models.Trip
	query := r.db.WithContext(ctx).Where("company_id = ? AND status = ?", companyID, "completed")
	
	if minScore >= 0 {
		query = query.Where("driver_performance_score >= ?", minScore)
	}
	if maxScore <= 100 {
		query = query.Where("driver_performance_score <= ?", maxScore)
	}
	
	query = query.Order("start_time DESC")
	
	// Apply pagination
	query = r.applyPagination(query, pagination)
	
	if err := query.Find(&trips).Error; err != nil {
		return nil, fmt.Errorf("failed to get trips by driver performance: %w", err)
	}
	
	return trips, nil
}
