package tracking

import (
	"context"
	"fmt"
	"time"

	"github.com/tobangado69/fleettracker-pro/backend/pkg/models"
	"gorm.io/gorm"
)

// OptimizedQueryService provides optimized database queries for tracking
type OptimizedQueryService struct {
	db *gorm.DB
}

// NewOptimizedQueryService creates a new optimized query service
func NewOptimizedQueryService(db *gorm.DB) *OptimizedQueryService {
	return &OptimizedQueryService{db: db}
}

// GetCurrentLocationOptimized gets the current location with optimized query
func (oqs *OptimizedQueryService) GetCurrentLocationOptimized(ctx context.Context, vehicleID string) (*models.GPSTrack, error) {
	var gpsTrack models.GPSTrack
	
	// Use index on (vehicle_id, timestamp DESC) for optimal performance
	if err := oqs.db.WithContext(ctx).
		Select("id, vehicle_id, driver_id, latitude, longitude, speed, heading, timestamp, accuracy").
		Where("vehicle_id = ? AND accuracy <= ?", vehicleID, 50). // Filter out inaccurate readings
		Order("timestamp DESC").
		First(&gpsTrack).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("GPS data for vehicle %s not found", vehicleID)
		}
		return nil, fmt.Errorf("failed to get current location: %w", err)
	}
	
	return &gpsTrack, nil
}

// GetLocationHistoryOptimized gets location history with optimized pagination
func (oqs *OptimizedQueryService) GetLocationHistoryOptimized(ctx context.Context, vehicleID string, startTime, endTime time.Time, limit int) ([]*models.GPSTrack, error) {
	var gpsTracks []*models.GPSTrack
	
	// Use index on (vehicle_id, timestamp DESC) with time range filter
	query := oqs.db.WithContext(ctx).
		Select("id, vehicle_id, driver_id, latitude, longitude, speed, heading, timestamp, accuracy").
		Where("vehicle_id = ? AND timestamp BETWEEN ? AND ? AND accuracy <= ?", 
			vehicleID, startTime, endTime, 50).
		Order("timestamp DESC")
	
	if limit > 0 {
		query = query.Limit(limit)
	}
	
	if err := query.Find(&gpsTracks).Error; err != nil {
		return nil, fmt.Errorf("failed to get location history: %w", err)
	}
	
	return gpsTracks, nil
}

// GetSpeedViolationsOptimized gets speed violations with optimized query
func (oqs *OptimizedQueryService) GetSpeedViolationsOptimized(ctx context.Context, vehicleID string, minSpeed float64, startTime, endTime time.Time) ([]*models.GPSTrack, error) {
	var gpsTracks []*models.GPSTrack
	
	// Use index on speed with time range filter
	if err := oqs.db.WithContext(ctx).
		Select("id, vehicle_id, driver_id, latitude, longitude, speed, timestamp").
		Where("vehicle_id = ? AND speed > ? AND timestamp BETWEEN ? AND ?", 
			vehicleID, minSpeed, startTime, endTime).
		Order("timestamp DESC").
		Find(&gpsTracks).Error; err != nil {
		return nil, fmt.Errorf("failed to get speed violations: %w", err)
	}
	
	return gpsTracks, nil
}

// GetRouteDistanceOptimized calculates route distance using optimized query
func (oqs *OptimizedQueryService) GetRouteDistanceOptimized(ctx context.Context, vehicleID string, startTime, endTime time.Time) (float64, error) {
	var result struct {
		TotalDistance float64 `gorm:"column:total_distance"`
	}
	
	// Use window function for efficient distance calculation
	query := `
		WITH ordered_tracks AS (
			SELECT 
				latitude, 
				longitude,
				LAG(latitude) OVER (ORDER BY timestamp) as prev_lat,
				LAG(longitude) OVER (ORDER BY timestamp) as prev_lng
			FROM gps_tracks 
			WHERE vehicle_id = ? 
				AND timestamp BETWEEN ? AND ? 
				AND accuracy <= 50
			ORDER BY timestamp
		)
		SELECT COALESCE(SUM(
			ST_Distance(
				ST_Point(longitude, latitude)::geography,
				ST_Point(prev_lng, prev_lat)::geography
			)
		), 0) as total_distance
		FROM ordered_tracks 
		WHERE prev_lat IS NOT NULL AND prev_lng IS NOT NULL
	`
	
	if err := oqs.db.WithContext(ctx).Raw(query, vehicleID, startTime, endTime).Scan(&result).Error; err != nil {
		return 0, fmt.Errorf("failed to calculate route distance: %w", err)
	}
	
	return result.TotalDistance, nil
}

// GetAverageSpeedOptimized calculates average speed with optimized query
func (oqs *OptimizedQueryService) GetAverageSpeedOptimized(ctx context.Context, vehicleID string, startTime, endTime time.Time) (float64, error) {
	var result struct {
		AverageSpeed float64 `gorm:"column:average_speed"`
	}
	
	// Use aggregate function for efficient calculation
	if err := oqs.db.WithContext(ctx).
		Model(&models.GPSTrack{}).
		Select("AVG(speed) as average_speed").
		Where("vehicle_id = ? AND timestamp BETWEEN ? AND ? AND speed > 0 AND accuracy <= ?", 
			vehicleID, startTime, endTime, 50).
		Scan(&result).Error; err != nil {
		return 0, fmt.Errorf("failed to calculate average speed: %w", err)
	}
	
	return result.AverageSpeed, nil
}

// GetMaxSpeedOptimized gets maximum speed with optimized query
func (oqs *OptimizedQueryService) GetMaxSpeedOptimized(ctx context.Context, vehicleID string, startTime, endTime time.Time) (float64, error) {
	var result struct {
		MaxSpeed float64 `gorm:"column:max_speed"`
	}
	
	// Use index on speed for efficient max calculation
	if err := oqs.db.WithContext(ctx).
		Model(&models.GPSTrack{}).
		Select("MAX(speed) as max_speed").
		Where("vehicle_id = ? AND timestamp BETWEEN ? AND ? AND speed > 0", 
			vehicleID, startTime, endTime).
		Scan(&result).Error; err != nil {
		return 0, fmt.Errorf("failed to get max speed: %w", err)
	}
	
	return result.MaxSpeed, nil
}

// GetIdleTimeOptimized calculates idle time with optimized query
func (oqs *OptimizedQueryService) GetIdleTimeOptimized(ctx context.Context, vehicleID string, startTime, endTime time.Time) (time.Duration, error) {
	var result struct {
		TotalIdleTime int `gorm:"column:total_idle_time"`
	}
	
	// Use aggregate function for efficient calculation
	if err := oqs.db.WithContext(ctx).
		Model(&models.GPSTrack{}).
		Select("SUM(idle_time) as total_idle_time").
		Where("vehicle_id = ? AND timestamp BETWEEN ? AND ?", 
			vehicleID, startTime, endTime).
		Scan(&result).Error; err != nil {
		return 0, fmt.Errorf("failed to calculate idle time: %w", err)
	}
	
	return time.Duration(result.TotalIdleTime) * time.Second, nil
}

// GetActiveTripsOptimized gets active trips with optimized query
func (oqs *OptimizedQueryService) GetActiveTripsOptimized(ctx context.Context, vehicleID string) ([]*models.Trip, error) {
	var trips []*models.Trip
	
	// Use index on (vehicle_id, status)
	if err := oqs.db.WithContext(ctx).
		Where("vehicle_id = ? AND status = ?", vehicleID, "active").
		Order("start_time DESC").
		Find(&trips).Error; err != nil {
		return nil, fmt.Errorf("failed to get active trips: %w", err)
	}
	
	return trips, nil
}

// GetTripsByVehicleOptimized gets trips with optimized pagination
func (oqs *OptimizedQueryService) GetTripsByVehicleOptimized(ctx context.Context, vehicleID string, startTime, endTime time.Time, page, limit int) ([]*models.Trip, int64, error) {
	var trips []*models.Trip
	var total int64
	
	// Build query with time range filter
	query := oqs.db.WithContext(ctx).Model(&models.Trip{}).
		Where("vehicle_id = ? AND start_time BETWEEN ? AND ?", vehicleID, startTime, endTime)
	
	// Get total count
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count trips: %w", err)
	}
	
	// Apply pagination and ordering
	offset := (page - 1) * limit
	if err := query.Order("start_time DESC").
		Offset(offset).
		Limit(limit).
		Find(&trips).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to get trips: %w", err)
	}
	
	return trips, total, nil
}

// GetDriverEventsOptimized gets driver events with optimized query
func (oqs *OptimizedQueryService) GetDriverEventsOptimized(ctx context.Context, driverID string, startTime, endTime time.Time, eventType string) ([]*models.DriverEvent, error) {
	var events []*models.DriverEvent
	
	// Build query with optional event type filter
	query := oqs.db.WithContext(ctx).
		Where("driver_id = ? AND timestamp BETWEEN ? AND ?", driverID, startTime, endTime)
	
	if eventType != "" {
		query = query.Where("event_type = ?", eventType)
	}
	
	if err := query.Order("timestamp DESC").Find(&events).Error; err != nil {
		return nil, fmt.Errorf("failed to get driver events: %w", err)
	}
	
	return events, nil
}

// GetGeofenceViolationsOptimized gets geofence violations with optimized query
func (oqs *OptimizedQueryService) GetGeofenceViolationsOptimized(ctx context.Context, vehicleID string, geofenceID string, startTime, endTime time.Time) ([]*models.GPSTrack, error) {
	var gpsTracks []*models.GPSTrack
	
	// Use PostGIS for efficient geofence violation detection
	query := `
		SELECT gt.id, gt.vehicle_id, gt.driver_id, gt.latitude, gt.longitude, gt.timestamp
		FROM gps_tracks gt
		JOIN geofences g ON g.id = ?
		WHERE gt.vehicle_id = ? 
			AND gt.timestamp BETWEEN ? AND ?
			AND ST_DWithin(
				ST_Point(gt.longitude, gt.latitude)::geography,
				ST_Point(g.center_longitude, g.center_latitude)::geography,
				g.radius
			)
		ORDER BY gt.timestamp DESC
	`
	
	if err := oqs.db.WithContext(ctx).Raw(query, geofenceID, vehicleID, startTime, endTime).Scan(&gpsTracks).Error; err != nil {
		return nil, fmt.Errorf("failed to get geofence violations: %w", err)
	}
	
	return gpsTracks, nil
}

// BatchInsertGPSTracksOptimized inserts GPS tracks in batches for better performance
func (oqs *OptimizedQueryService) BatchInsertGPSTracksOptimized(ctx context.Context, tracks []*models.GPSTrack) error {
	if len(tracks) == 0 {
		return nil
	}
	
	// Use batch insert for better performance
	if err := oqs.db.WithContext(ctx).CreateInBatches(tracks, 100).Error; err != nil {
		return fmt.Errorf("failed to batch insert GPS tracks: %w", err)
	}
	
	return nil
}

// GetVehicleStatsOptimized gets vehicle statistics with optimized query
func (oqs *OptimizedQueryService) GetVehicleStatsOptimized(ctx context.Context, vehicleID string, startTime, endTime time.Time) (map[string]interface{}, error) {
	var stats struct {
		TotalDistance    float64 `gorm:"column:total_distance"`
		AverageSpeed     float64 `gorm:"column:average_speed"`
		MaxSpeed         float64 `gorm:"column:max_speed"`
		TotalIdleTime    int     `gorm:"column:total_idle_time"`
		TrackCount       int64   `gorm:"column:track_count"`
		SpeedViolations  int64   `gorm:"column:speed_violations"`
	}
	
	query := `
		SELECT 
			COUNT(*) as track_count,
			AVG(CASE WHEN speed > 0 THEN speed END) as average_speed,
			MAX(speed) as max_speed,
			SUM(idle_time) as total_idle_time,
			COUNT(CASE WHEN speed > 80 THEN 1 END) as speed_violations
		FROM gps_tracks 
		WHERE vehicle_id = ? 
			AND timestamp BETWEEN ? AND ?
			AND accuracy <= 50
	`
	
	if err := oqs.db.WithContext(ctx).Raw(query, vehicleID, startTime, endTime).Scan(&stats).Error; err != nil {
		return nil, fmt.Errorf("failed to get vehicle stats: %w", err)
	}
	
	// Calculate total distance separately for better performance
	totalDistance, err := oqs.GetRouteDistanceOptimized(ctx, vehicleID, startTime, endTime)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate total distance: %w", err)
	}
	
	return map[string]interface{}{
		"total_distance":    totalDistance,
		"average_speed":     stats.AverageSpeed,
		"max_speed":         stats.MaxSpeed,
		"total_idle_time":   time.Duration(stats.TotalIdleTime) * time.Second,
		"track_count":       stats.TrackCount,
		"speed_violations":  stats.SpeedViolations,
	}, nil
}
