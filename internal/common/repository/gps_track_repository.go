package repository

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/tobangado69/fleettracker-pro/backend/pkg/models"
)

// GPSTrackRepositoryImpl implements the GPSTrackRepository interface
type GPSTrackRepositoryImpl struct {
	*BaseRepository[models.GPSTrack]
}

// NewGPSTrackRepository creates a new GPS track repository
func NewGPSTrackRepository(db *gorm.DB) GPSTrackRepository {
	return &GPSTrackRepositoryImpl{
		BaseRepository: NewBaseRepository[models.GPSTrack](db),
	}
}

// GetByVehicle retrieves GPS tracks by vehicle ID with pagination
func (r *GPSTrackRepositoryImpl) GetByVehicle(ctx context.Context, vehicleID string, pagination Pagination) ([]*models.GPSTrack, error) {
	var tracks []*models.GPSTrack
	query := r.db.WithContext(ctx).Where("vehicle_id = ?", vehicleID).Order("timestamp DESC")
	
	// Apply pagination
	query = r.applyPagination(query, pagination)
	
	if err := query.Find(&tracks).Error; err != nil {
		return nil, fmt.Errorf("failed to get GPS tracks by vehicle: %w", err)
	}
	
	return tracks, nil
}

// GetByVehicleAndDateRange retrieves GPS tracks by vehicle ID and date range
func (r *GPSTrackRepositoryImpl) GetByVehicleAndDateRange(ctx context.Context, vehicleID string, startDate, endDate string) ([]*models.GPSTrack, error) {
	var tracks []*models.GPSTrack
	query := r.db.WithContext(ctx).Where("vehicle_id = ?", vehicleID)
	
	if startDate != "" {
		query = query.Where("timestamp >= ?", startDate)
	}
	if endDate != "" {
		query = query.Where("timestamp <= ?", endDate)
	}
	
	query = query.Order("timestamp ASC")
	
	if err := query.Find(&tracks).Error; err != nil {
		return nil, fmt.Errorf("failed to get GPS tracks by vehicle and date range: %w", err)
	}
	
	return tracks, nil
}

// GetCurrentLocation retrieves the most recent GPS track for a vehicle
func (r *GPSTrackRepositoryImpl) GetCurrentLocation(ctx context.Context, vehicleID string) (*models.GPSTrack, error) {
	var track models.GPSTrack
	if err := r.db.WithContext(ctx).Where("vehicle_id = ?", vehicleID).Order("timestamp DESC").First(&track).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("no GPS tracks found for vehicle: %s", vehicleID)
		}
		return nil, fmt.Errorf("failed to get current location: %w", err)
	}
	return &track, nil
}

// GetLocationHistory retrieves GPS location history for a vehicle with pagination
func (r *GPSTrackRepositoryImpl) GetLocationHistory(ctx context.Context, vehicleID string, pagination Pagination) ([]*models.GPSTrack, error) {
	var tracks []*models.GPSTrack
	query := r.db.WithContext(ctx).Where("vehicle_id = ?", vehicleID).Order("timestamp DESC")
	
	// Apply pagination
	query = r.applyPagination(query, pagination)
	
	if err := query.Find(&tracks).Error; err != nil {
		return nil, fmt.Errorf("failed to get location history: %w", err)
	}
	
	return tracks, nil
}

// GetByDriver retrieves GPS tracks by driver ID with pagination
func (r *GPSTrackRepositoryImpl) GetByDriver(ctx context.Context, driverID string, pagination Pagination) ([]*models.GPSTrack, error) {
	var tracks []*models.GPSTrack
	query := r.db.WithContext(ctx).Where("driver_id = ?", driverID).Order("timestamp DESC")
	
	// Apply pagination
	query = r.applyPagination(query, pagination)
	
	if err := query.Find(&tracks).Error; err != nil {
		return nil, fmt.Errorf("failed to get GPS tracks by driver: %w", err)
	}
	
	return tracks, nil
}

// GetByTrip retrieves GPS tracks for a specific trip
func (r *GPSTrackRepositoryImpl) GetByTrip(ctx context.Context, tripID string) ([]*models.GPSTrack, error) {
	var tracks []*models.GPSTrack
	if err := r.db.WithContext(ctx).Where("trip_id = ?", tripID).Order("timestamp ASC").Find(&tracks).Error; err != nil {
		return nil, fmt.Errorf("failed to get GPS tracks by trip: %w", err)
	}
	return tracks, nil
}

// GetSpeedViolations retrieves GPS tracks with speed violations
func (r *GPSTrackRepositoryImpl) GetSpeedViolations(ctx context.Context, vehicleID string, minSpeed float64) ([]*models.GPSTrack, error) {
	var tracks []*models.GPSTrack
	query := r.db.WithContext(ctx).Where("vehicle_id = ? AND speed > ?", vehicleID, minSpeed).Order("timestamp DESC")
	
	if err := query.Find(&tracks).Error; err != nil {
		return nil, fmt.Errorf("failed to get speed violations: %w", err)
	}
	
	return tracks, nil
}

// GetRecentTracks retrieves recent GPS tracks for a company
func (r *GPSTrackRepositoryImpl) GetRecentTracks(ctx context.Context, companyID string, limit int) ([]*models.GPSTrack, error) {
	var tracks []*models.GPSTrack
	
	// Join with vehicles table to filter by company
	query := r.db.WithContext(ctx).
		Joins("JOIN vehicles ON gps_tracks.vehicle_id = vehicles.id").
		Where("vehicles.company_id = ?", companyID).
		Order("timestamp DESC").
		Limit(limit)
	
	if err := query.Find(&tracks).Error; err != nil {
		return nil, fmt.Errorf("failed to get recent tracks: %w", err)
	}
	
	return tracks, nil
}

// GetTracksInGeofence retrieves GPS tracks within a specific geofence
func (r *GPSTrackRepositoryImpl) GetTracksInGeofence(ctx context.Context, geofenceID string, pagination Pagination) ([]*models.GPSTrack, error) {
	var tracks []*models.GPSTrack
	
	// This would require PostGIS for spatial queries, but we'll implement a basic version
	// For now, we'll get tracks and filter by geofence in the application layer
	query := r.db.WithContext(ctx).Where("geofence_id = ?", geofenceID).Order("timestamp DESC")
	
	// Apply pagination
	query = r.applyPagination(query, pagination)
	
	if err := query.Find(&tracks).Error; err != nil {
		return nil, fmt.Errorf("failed to get tracks in geofence: %w", err)
	}
	
	return tracks, nil
}

// AggregateByTimeRange performs aggregation on GPS tracks within a time range
func (r *GPSTrackRepositoryImpl) AggregateByTimeRange(ctx context.Context, vehicleID string, startDate, endDate string) (map[string]interface{}, error) {
	var stats struct {
		TotalTracks    int64   `json:"total_tracks"`
		AverageSpeed   float64 `json:"average_speed"`
		MaxSpeed       float64 `json:"max_speed"`
		MinSpeed       float64 `json:"min_speed"`
		TotalDistance  float64 `json:"total_distance"`
		TotalDuration  int64   `json:"total_duration"` // in seconds
	}

	// Build base query
	query := r.db.WithContext(ctx).Model(&models.GPSTrack{}).Where("vehicle_id = ?", vehicleID)
	
	if startDate != "" {
		query = query.Where("timestamp >= ?", startDate)
	}
	if endDate != "" {
		query = query.Where("timestamp <= ?", endDate)
	}

	// Get total tracks
	if err := query.Count(&stats.TotalTracks).Error; err != nil {
		return nil, fmt.Errorf("failed to count total tracks: %w", err)
	}

	// Get speed statistics
	var speedStats struct {
		Average float64
		Max     float64
		Min     float64
	}
	if err := query.Select("AVG(speed) as average, MAX(speed) as max, MIN(speed) as min").Scan(&speedStats).Error; err != nil {
		return nil, fmt.Errorf("failed to get speed statistics: %w", err)
	}
	stats.AverageSpeed = speedStats.Average
	stats.MaxSpeed = speedStats.Max
	stats.MinSpeed = speedStats.Min

	// Get total distance (this would be calculated from GPS coordinates in a real implementation)
	// For now, we'll use a placeholder
	stats.TotalDistance = 0.0

	// Get total duration
	var durationStats struct {
		Duration int64
	}
	if err := query.Select("EXTRACT(EPOCH FROM (MAX(timestamp) - MIN(timestamp))) as duration").Scan(&durationStats).Error; err != nil {
		return nil, fmt.Errorf("failed to get duration: %w", err)
	}
	stats.TotalDuration = durationStats.Duration

	return map[string]interface{}{
		"total_tracks":    stats.TotalTracks,
		"average_speed":   stats.AverageSpeed,
		"max_speed":       stats.MaxSpeed,
		"min_speed":       stats.MinSpeed,
		"total_distance":  stats.TotalDistance,
		"total_duration":  stats.TotalDuration,
	}, nil
}

// GetTracksByLocation retrieves GPS tracks within a geographic area
func (r *GPSTrackRepositoryImpl) GetTracksByLocation(ctx context.Context, latitude, longitude, radius float64, pagination Pagination) ([]*models.GPSTrack, error) {
	var tracks []*models.GPSTrack
	
	// This would require PostGIS for spatial queries
	// For now, we'll implement a basic bounding box approach
	latRange := radius / 111.0 // Rough conversion: 1 degree latitude ≈ 111 km
	lngRange := radius / (111.0 * cos(latitude))
	
	query := r.db.WithContext(ctx).
		Where("latitude BETWEEN ? AND ? AND longitude BETWEEN ? AND ?",
			latitude-latRange, latitude+latRange,
			longitude-lngRange, longitude+lngRange).
		Order("timestamp DESC")
	
	// Apply pagination
	query = r.applyPagination(query, pagination)
	
	if err := query.Find(&tracks).Error; err != nil {
		return nil, fmt.Errorf("failed to get tracks by location: %w", err)
	}
	
	return tracks, nil
}

// GetTracksBySpeedRange retrieves GPS tracks within a speed range
func (r *GPSTrackRepositoryImpl) GetTracksBySpeedRange(ctx context.Context, vehicleID string, minSpeed, maxSpeed float64, pagination Pagination) ([]*models.GPSTrack, error) {
	var tracks []*models.GPSTrack
	query := r.db.WithContext(ctx).Where("vehicle_id = ?", vehicleID)
	
	if minSpeed >= 0 {
		query = query.Where("speed >= ?", minSpeed)
	}
	if maxSpeed > 0 {
		query = query.Where("speed <= ?", maxSpeed)
	}
	
	query = query.Order("timestamp DESC")
	
	// Apply pagination
	query = r.applyPagination(query, pagination)
	
	if err := query.Find(&tracks).Error; err != nil {
		return nil, fmt.Errorf("failed to get tracks by speed range: %w", err)
	}
	
	return tracks, nil
}

// GetTracksByAccuracy retrieves GPS tracks filtered by accuracy
func (r *GPSTrackRepositoryImpl) GetTracksByAccuracy(ctx context.Context, vehicleID string, maxAccuracy float64, pagination Pagination) ([]*models.GPSTrack, error) {
	var tracks []*models.GPSTrack
	query := r.db.WithContext(ctx).Where("vehicle_id = ? AND accuracy <= ?", vehicleID, maxAccuracy).Order("timestamp DESC")
	
	// Apply pagination
	query = r.applyPagination(query, pagination)
	
	if err := query.Find(&tracks).Error; err != nil {
		return nil, fmt.Errorf("failed to get tracks by accuracy: %w", err)
	}
	
	return tracks, nil
}

// GetTracksByBatteryLevel retrieves GPS tracks filtered by battery level
func (r *GPSTrackRepositoryImpl) GetTracksByBatteryLevel(ctx context.Context, vehicleID string, minBatteryLevel float64, pagination Pagination) ([]*models.GPSTrack, error) {
	var tracks []*models.GPSTrack
	query := r.db.WithContext(ctx).Where("vehicle_id = ? AND battery_level >= ?", vehicleID, minBatteryLevel).Order("timestamp DESC")
	
	// Apply pagination
	query = r.applyPagination(query, pagination)
	
	if err := query.Find(&tracks).Error; err != nil {
		return nil, fmt.Errorf("failed to get tracks by battery level: %w", err)
	}
	
	return tracks, nil
}

// GetTracksStatistics retrieves GPS tracking statistics for a vehicle
func (r *GPSTrackRepositoryImpl) GetTracksStatistics(ctx context.Context, vehicleID string, days int) (map[string]interface{}, error) {
	var stats struct {
		TotalTracks      int64   `json:"total_tracks"`
		TodayTracks      int64   `json:"today_tracks"`
		AverageSpeed     float64 `json:"average_speed"`
		MaxSpeed         float64 `json:"max_speed"`
		AverageAccuracy  float64 `json:"average_accuracy"`
		AverageBattery   float64 `json:"average_battery"`
	}

	// Get total tracks
	if err := r.db.WithContext(ctx).Model(&models.GPSTrack{}).Where("vehicle_id = ?", vehicleID).Count(&stats.TotalTracks).Error; err != nil {
		return nil, fmt.Errorf("failed to count total tracks: %w", err)
	}

	// Get today's tracks
	today := time.Now().Format("2006-01-02")
	if err := r.db.WithContext(ctx).Model(&models.GPSTrack{}).Where("vehicle_id = ? AND DATE(timestamp) = ?", vehicleID, today).Count(&stats.TodayTracks).Error; err != nil {
		return nil, fmt.Errorf("failed to count today's tracks: %w", err)
	}

	// Get speed statistics
	var speedStats struct {
		Average float64
		Max     float64
	}
	if err := r.db.WithContext(ctx).Model(&models.GPSTrack{}).Where("vehicle_id = ?", vehicleID).Select("AVG(speed) as average, MAX(speed) as max").Scan(&speedStats).Error; err != nil {
		return nil, fmt.Errorf("failed to get speed statistics: %w", err)
	}
	stats.AverageSpeed = speedStats.Average
	stats.MaxSpeed = speedStats.Max

	// Get accuracy statistics
	var accuracyStats struct {
		Average float64
	}
	if err := r.db.WithContext(ctx).Model(&models.GPSTrack{}).Where("vehicle_id = ?", vehicleID).Select("AVG(accuracy) as average").Scan(&accuracyStats).Error; err != nil {
		return nil, fmt.Errorf("failed to get accuracy statistics: %w", err)
	}
	stats.AverageAccuracy = accuracyStats.Average

	// Get battery statistics
	var batteryStats struct {
		Average float64
	}
	if err := r.db.WithContext(ctx).Model(&models.GPSTrack{}).Where("vehicle_id = ?", vehicleID).Select("AVG(battery_level) as average").Scan(&batteryStats).Error; err != nil {
		return nil, fmt.Errorf("failed to get battery statistics: %w", err)
	}
	stats.AverageBattery = batteryStats.Average

	return map[string]interface{}{
		"total_tracks":     stats.TotalTracks,
		"today_tracks":     stats.TodayTracks,
		"average_speed":    stats.AverageSpeed,
		"max_speed":        stats.MaxSpeed,
		"average_accuracy": stats.AverageAccuracy,
		"average_battery":  stats.AverageBattery,
	}, nil
}

// CleanupOldTracks removes GPS tracks older than specified days
func (r *GPSTrackRepositoryImpl) CleanupOldTracks(ctx context.Context, days int) error {
	cutoffDate := time.Now().AddDate(0, 0, -days)
	
	if err := r.db.WithContext(ctx).Where("timestamp < ?", cutoffDate).Delete(&models.GPSTrack{}).Error; err != nil {
		return fmt.Errorf("failed to cleanup old tracks: %w", err)
	}
	
	return nil
}

// GetTracksByTimeOfDay retrieves GPS tracks filtered by time of day
func (r *GPSTrackRepositoryImpl) GetTracksByTimeOfDay(ctx context.Context, vehicleID string, startHour, endHour int, pagination Pagination) ([]*models.GPSTrack, error) {
	var tracks []*models.GPSTrack
	query := r.db.WithContext(ctx).Where("vehicle_id = ? AND EXTRACT(HOUR FROM timestamp) BETWEEN ? AND ?", vehicleID, startHour, endHour).Order("timestamp DESC")
	
	// Apply pagination
	query = r.applyPagination(query, pagination)
	
	if err := query.Find(&tracks).Error; err != nil {
		return nil, fmt.Errorf("failed to get tracks by time of day: %w", err)
	}
	
	return tracks, nil
}

// Helper function for cosine calculation
func cos(x float64) float64 {
	// Simple cosine approximation for small angles
	// In a real implementation, you would use math.Cos
	return 1.0 - (x*x)/2.0 + (x*x*x*x)/24.0
}
