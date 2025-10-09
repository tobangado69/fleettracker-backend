package tracking

import (
	"context"
	"fmt"
	"time"

	"github.com/tobangado69/fleettracker-pro/backend/internal/common/cache"
	"github.com/tobangado69/fleettracker-pro/backend/pkg/models"
)

// CachedTrackingService provides cached tracking operations
type CachedTrackingService struct {
	optimizedQueries *OptimizedQueryService
	cache           *cache.RedisCache
}

// NewCachedTrackingService creates a new cached tracking service
func NewCachedTrackingService(optimizedQueries *OptimizedQueryService, cache *cache.RedisCache) *CachedTrackingService {
	return &CachedTrackingService{
		optimizedQueries: optimizedQueries,
		cache:           cache,
	}
}

// GetCurrentLocationCached gets current location with caching
func (cts *CachedTrackingService) GetCurrentLocationCached(ctx context.Context, vehicleID string) (*models.GPSTrack, error) {
	// Try cache first
	cacheKey := cts.cache.VehicleLocationKey(vehicleID)
	var cachedLocation models.GPSTrack
	
	if err := cts.cache.Get(ctx, cacheKey, &cachedLocation); err == nil {
		// Check if cached data is recent (within 1 minute)
		if time.Since(cachedLocation.Timestamp) < time.Minute {
			return &cachedLocation, nil
		}
		// Cache expired, remove it
		cts.cache.Delete(ctx, cacheKey)
	}
	
	// Get from database
	location, err := cts.optimizedQueries.GetCurrentLocationOptimized(ctx, vehicleID)
	if err != nil {
		return nil, err
	}
	
	// Cache the result
	cts.cache.Set(ctx, cacheKey, location, cache.LocationExpiration)
	
	return location, nil
}

// GetLocationHistoryCached gets location history with caching
func (cts *CachedTrackingService) GetLocationHistoryCached(ctx context.Context, vehicleID string, startTime, endTime time.Time, limit int) ([]*models.GPSTrack, error) {
	// Create cache key based on parameters
	cacheKey := fmt.Sprintf("location_history:%s:%d:%d:%d", 
		vehicleID, startTime.Unix(), endTime.Unix(), limit)
	
	// Try cache first
	var cachedHistory []*models.GPSTrack
	if err := cts.cache.Get(ctx, cacheKey, &cachedHistory); err == nil {
		return cachedHistory, nil
	}
	
	// Get from database
	history, err := cts.optimizedQueries.GetLocationHistoryOptimized(ctx, vehicleID, startTime, endTime, limit)
	if err != nil {
		return nil, err
	}
	
	// Cache the result (shorter expiration for historical data)
	cts.cache.Set(ctx, cacheKey, history, cache.ShortExpiration)
	
	return history, nil
}

// GetVehicleStatsCached gets vehicle statistics with caching
func (cts *CachedTrackingService) GetVehicleStatsCached(ctx context.Context, vehicleID string, startTime, endTime time.Time) (map[string]interface{}, error) {
	// Create cache key based on parameters
	cacheKey := fmt.Sprintf("vehicle_stats:%s:%d:%d", 
		vehicleID, startTime.Unix(), endTime.Unix())
	
	// Try cache first
	var cachedStats map[string]interface{}
	if err := cts.cache.Get(ctx, cacheKey, &cachedStats); err == nil {
		return cachedStats, nil
	}
	
	// Get from database
	stats, err := cts.optimizedQueries.GetVehicleStatsOptimized(ctx, vehicleID, startTime, endTime)
	if err != nil {
		return nil, err
	}
	
	// Cache the result
	cts.cache.Set(ctx, cacheKey, stats, cache.StatsExpiration)
	
	return stats, nil
}

// GetSpeedViolationsCached gets speed violations with caching
func (cts *CachedTrackingService) GetSpeedViolationsCached(ctx context.Context, vehicleID string, minSpeed float64, startTime, endTime time.Time) ([]*models.GPSTrack, error) {
	// Create cache key based on parameters
	cacheKey := fmt.Sprintf("speed_violations:%s:%.1f:%d:%d", 
		vehicleID, minSpeed, startTime.Unix(), endTime.Unix())
	
	// Try cache first
	var cachedViolations []*models.GPSTrack
	if err := cts.cache.Get(ctx, cacheKey, &cachedViolations); err == nil {
		return cachedViolations, nil
	}
	
	// Get from database
	violations, err := cts.optimizedQueries.GetSpeedViolationsOptimized(ctx, vehicleID, minSpeed, startTime, endTime)
	if err != nil {
		return nil, err
	}
	
	// Cache the result
	cts.cache.Set(ctx, cacheKey, violations, cache.MediumExpiration)
	
	return violations, nil
}

// GetRouteDistanceCached gets route distance with caching
func (cts *CachedTrackingService) GetRouteDistanceCached(ctx context.Context, vehicleID string, startTime, endTime time.Time) (float64, error) {
	// Create cache key based on parameters
	cacheKey := fmt.Sprintf("route_distance:%s:%d:%d", 
		vehicleID, startTime.Unix(), endTime.Unix())
	
	// Try cache first
	var cachedDistance float64
	if err := cts.cache.Get(ctx, cacheKey, &cachedDistance); err == nil {
		return cachedDistance, nil
	}
	
	// Get from database
	distance, err := cts.optimizedQueries.GetRouteDistanceOptimized(ctx, vehicleID, startTime, endTime)
	if err != nil {
		return 0, err
	}
	
	// Cache the result
	cts.cache.Set(ctx, cacheKey, distance, cache.LongExpiration)
	
	return distance, nil
}

// GetAverageSpeedCached gets average speed with caching
func (cts *CachedTrackingService) GetAverageSpeedCached(ctx context.Context, vehicleID string, startTime, endTime time.Time) (float64, error) {
	// Create cache key based on parameters
	cacheKey := fmt.Sprintf("average_speed:%s:%d:%d", 
		vehicleID, startTime.Unix(), endTime.Unix())
	
	// Try cache first
	var cachedSpeed float64
	if err := cts.cache.Get(ctx, cacheKey, &cachedSpeed); err == nil {
		return cachedSpeed, nil
	}
	
	// Get from database
	speed, err := cts.optimizedQueries.GetAverageSpeedOptimized(ctx, vehicleID, startTime, endTime)
	if err != nil {
		return 0, err
	}
	
	// Cache the result
	cts.cache.Set(ctx, cacheKey, speed, cache.MediumExpiration)
	
	return speed, nil
}

// GetMaxSpeedCached gets max speed with caching
func (cts *CachedTrackingService) GetMaxSpeedCached(ctx context.Context, vehicleID string, startTime, endTime time.Time) (float64, error) {
	// Create cache key based on parameters
	cacheKey := fmt.Sprintf("max_speed:%s:%d:%d", 
		vehicleID, startTime.Unix(), endTime.Unix())
	
	// Try cache first
	var cachedSpeed float64
	if err := cts.cache.Get(ctx, cacheKey, &cachedSpeed); err == nil {
		return cachedSpeed, nil
	}
	
	// Get from database
	speed, err := cts.optimizedQueries.GetMaxSpeedOptimized(ctx, vehicleID, startTime, endTime)
	if err != nil {
		return 0, err
	}
	
	// Cache the result
	cts.cache.Set(ctx, cacheKey, speed, cache.MediumExpiration)
	
	return speed, nil
}

// GetIdleTimeCached gets idle time with caching
func (cts *CachedTrackingService) GetIdleTimeCached(ctx context.Context, vehicleID string, startTime, endTime time.Time) (time.Duration, error) {
	// Create cache key based on parameters
	cacheKey := fmt.Sprintf("idle_time:%s:%d:%d", 
		vehicleID, startTime.Unix(), endTime.Unix())
	
	// Try cache first
	var cachedIdleTime int64
	if err := cts.cache.Get(ctx, cacheKey, &cachedIdleTime); err == nil {
		return time.Duration(cachedIdleTime), nil
	}
	
	// Get from database
	idleTime, err := cts.optimizedQueries.GetIdleTimeOptimized(ctx, vehicleID, startTime, endTime)
	if err != nil {
		return 0, err
	}
	
	// Cache the result
	cts.cache.Set(ctx, cacheKey, int64(idleTime), cache.MediumExpiration)
	
	return idleTime, nil
}

// GetActiveTripsCached gets active trips with caching
func (cts *CachedTrackingService) GetActiveTripsCached(ctx context.Context, vehicleID string) ([]*models.Trip, error) {
	// Create cache key
	cacheKey := fmt.Sprintf("active_trips:%s", vehicleID)
	
	// Try cache first
	var cachedTrips []*models.Trip
	if err := cts.cache.Get(ctx, cacheKey, &cachedTrips); err == nil {
		return cachedTrips, nil
	}
	
	// Get from database
	trips, err := cts.optimizedQueries.GetActiveTripsOptimized(ctx, vehicleID)
	if err != nil {
		return nil, err
	}
	
	// Cache the result (shorter expiration for active data)
	cts.cache.Set(ctx, cacheKey, trips, cache.ShortExpiration)
	
	return trips, nil
}

// GetTripsByVehicleCached gets trips with caching
func (cts *CachedTrackingService) GetTripsByVehicleCached(ctx context.Context, vehicleID string, startTime, endTime time.Time, page, limit int) ([]*models.Trip, int64, error) {
	// Create cache key based on parameters
	cacheKey := fmt.Sprintf("trips:%s:%d:%d:%d:%d", 
		vehicleID, startTime.Unix(), endTime.Unix(), page, limit)
	
	// Try cache first
	var cachedResult struct {
		Trips []*models.Trip `json:"trips"`
		Total int64          `json:"total"`
	}
	if err := cts.cache.Get(ctx, cacheKey, &cachedResult); err == nil {
		return cachedResult.Trips, cachedResult.Total, nil
	}
	
	// Get from database
	trips, total, err := cts.optimizedQueries.GetTripsByVehicleOptimized(ctx, vehicleID, startTime, endTime, page, limit)
	if err != nil {
		return nil, 0, err
	}
	
	// Cache the result
	result := struct {
		Trips []*models.Trip `json:"trips"`
		Total int64          `json:"total"`
	}{
		Trips: trips,
		Total: total,
	}
	cts.cache.Set(ctx, cacheKey, result, cache.MediumExpiration)
	
	return trips, total, nil
}

// GetDriverEventsCached gets driver events with caching
func (cts *CachedTrackingService) GetDriverEventsCached(ctx context.Context, driverID string, startTime, endTime time.Time, eventType string) ([]*models.DriverEvent, error) {
	// Create cache key based on parameters
	cacheKey := fmt.Sprintf("driver_events:%s:%d:%d:%s", 
		driverID, startTime.Unix(), endTime.Unix(), eventType)
	
	// Try cache first
	var cachedEvents []*models.DriverEvent
	if err := cts.cache.Get(ctx, cacheKey, &cachedEvents); err == nil {
		return cachedEvents, nil
	}
	
	// Get from database
	events, err := cts.optimizedQueries.GetDriverEventsOptimized(ctx, driverID, startTime, endTime, eventType)
	if err != nil {
		return nil, err
	}
	
	// Cache the result
	cts.cache.Set(ctx, cacheKey, events, cache.MediumExpiration)
	
	return events, nil
}

// GetGeofenceViolationsCached gets geofence violations with caching
func (cts *CachedTrackingService) GetGeofenceViolationsCached(ctx context.Context, vehicleID string, geofenceID string, startTime, endTime time.Time) ([]*models.GPSTrack, error) {
	// Create cache key based on parameters
	cacheKey := fmt.Sprintf("geofence_violations:%s:%s:%d:%d", 
		vehicleID, geofenceID, startTime.Unix(), endTime.Unix())
	
	// Try cache first
	var cachedViolations []*models.GPSTrack
	if err := cts.cache.Get(ctx, cacheKey, &cachedViolations); err == nil {
		return cachedViolations, nil
	}
	
	// Get from database
	violations, err := cts.optimizedQueries.GetGeofenceViolationsOptimized(ctx, vehicleID, geofenceID, startTime, endTime)
	if err != nil {
		return nil, err
	}
	
	// Cache the result
	cts.cache.Set(ctx, cacheKey, violations, cache.MediumExpiration)
	
	return violations, nil
}

// InvalidateVehicleCache invalidates all cache entries for a vehicle
func (cts *CachedTrackingService) InvalidateVehicleCache(ctx context.Context, vehicleID string) error {
	// Note: This is a simplified approach. In production, you might want to maintain
	// a list of cache keys per vehicle for more efficient invalidation.
	
	// For now, we'll invalidate the most common keys
	keys := []string{
		cts.cache.VehicleLocationKey(vehicleID),
		cts.cache.VehicleStatsKey(vehicleID),
	}
	
	for _, key := range keys {
		cts.cache.Delete(ctx, key)
	}
	
	return nil
}

// InvalidateDriverCache invalidates all cache entries for a driver
func (cts *CachedTrackingService) InvalidateDriverCache(ctx context.Context, driverID string) error {
	// Invalidate driver-related cache entries
	keys := []string{
		cts.cache.DriverKey(driverID),
	}
	
	for _, key := range keys {
		cts.cache.Delete(ctx, key)
	}
	
	return nil
}
