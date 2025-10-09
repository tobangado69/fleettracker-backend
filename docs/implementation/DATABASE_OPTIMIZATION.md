# Database Optimization Guide

This document outlines the database optimization strategies implemented in the FleetTracker Pro backend to ensure optimal performance and scalability.

## Overview

The database optimization implementation includes:
- **Performance Indexes**: Strategic database indexes for frequently queried fields
- **Optimized Queries**: Efficient query patterns and best practices
- **Redis Caching**: Caching layer for expensive operations
- **Query Monitoring**: Performance monitoring and slow query detection
- **Pagination Optimization**: Efficient pagination with proper indexing

## Database Indexes

### Migration: `003_add_performance_indexes.up.sql`

The optimization includes comprehensive indexes for all major tables:

#### Vehicle Table Indexes
```sql
-- Composite indexes for common query patterns
CREATE INDEX CONCURRENTLY idx_vehicles_company_id_status ON vehicles(company_id, status);
CREATE INDEX CONCURRENTLY idx_vehicles_company_id_is_active ON vehicles(company_id, is_active);
CREATE INDEX CONCURRENTLY idx_vehicles_company_id_driver_id ON vehicles(company_id, driver_id);

-- Single column indexes for filtering
CREATE INDEX CONCURRENTLY idx_vehicles_make_model ON vehicles(make, model);
CREATE INDEX CONCURRENTLY idx_vehicles_year ON vehicles(year);
CREATE INDEX CONCURRENTLY idx_vehicles_fuel_type ON vehicles(fuel_type);
CREATE INDEX CONCURRENTLY idx_vehicles_license_plate_lower ON vehicles(LOWER(license_plate));
CREATE INDEX CONCURRENTLY idx_vehicles_vin ON vehicles(vin);
CREATE INDEX CONCURRENTLY idx_vehicles_inspection_date ON vehicles(inspection_date) WHERE inspection_date IS NOT NULL;
```

#### GPS Tracking Indexes (Most Critical)
```sql
-- Primary tracking indexes
CREATE INDEX CONCURRENTLY idx_gps_tracks_vehicle_id_timestamp ON gps_tracks(vehicle_id, timestamp DESC);
CREATE INDEX CONCURRENTLY idx_gps_tracks_driver_id_timestamp ON gps_tracks(driver_id, timestamp DESC) WHERE driver_id IS NOT NULL;
CREATE INDEX CONCURRENTLY idx_gps_tracks_trip_id_timestamp ON gps_tracks(trip_id, timestamp DESC) WHERE trip_id IS NOT NULL;

-- Performance optimization indexes
CREATE INDEX CONCURRENTLY idx_gps_tracks_vehicle_timestamp_range ON gps_tracks(vehicle_id, timestamp) WHERE timestamp >= NOW() - INTERVAL '30 days';
CREATE INDEX CONCURRENTLY idx_gps_tracks_speed ON gps_tracks(speed) WHERE speed > 0;
CREATE INDEX CONCURRENTLY idx_gps_tracks_accuracy ON gps_tracks(accuracy) WHERE accuracy > 0;
CREATE INDEX CONCURRENTLY idx_gps_tracks_moving ON gps_tracks(moving, timestamp) WHERE moving = true;
```

#### Driver Table Indexes
```sql
-- Composite indexes
CREATE INDEX CONCURRENTLY idx_drivers_company_id_status ON drivers(company_id, status);
CREATE INDEX CONCURRENTLY idx_drivers_company_id_is_active ON drivers(company_id, is_active);
CREATE INDEX CONCURRENTLY idx_drivers_vehicle_id ON drivers(vehicle_id) WHERE vehicle_id IS NOT NULL;

-- Single column indexes
CREATE INDEX CONCURRENTLY idx_drivers_nik ON drivers(nik);
CREATE INDEX CONCURRENTLY idx_drivers_sim_number ON drivers(sim_number);
CREATE INDEX CONCURRENTLY idx_drivers_medical_checkup_date ON drivers(medical_checkup_date) WHERE medical_checkup_date IS NOT NULL;
CREATE INDEX CONCURRENTLY idx_drivers_training_expiry ON drivers(training_expiry) WHERE training_expiry IS NOT NULL;
CREATE INDEX CONCURRENTLY idx_drivers_performance_score ON drivers(overall_score) WHERE overall_score > 0;
```

## Optimized Query Services

### Tracking Service Optimizations

#### `OptimizedQueryService` (`internal/tracking/optimized_queries.go`)

**Key Optimizations:**
- **Selective Field Loading**: Only load necessary fields to reduce memory usage
- **Index-Aware Queries**: Queries designed to leverage database indexes
- **Efficient Aggregations**: Use database functions for calculations
- **PostGIS Integration**: Leverage PostGIS for geospatial operations

**Example Optimized Query:**
```go
func (oqs *OptimizedQueryService) GetCurrentLocationOptimized(ctx context.Context, vehicleID string) (*models.GPSTrack, error) {
    var gpsTrack models.GPSTrack
    
    // Use index on (vehicle_id, timestamp DESC) for optimal performance
    if err := oqs.db.WithContext(ctx).
        Select("id, vehicle_id, driver_id, latitude, longitude, speed, heading, timestamp, accuracy").
        Where("vehicle_id = ? AND accuracy <= ?", vehicleID, 50). // Filter out inaccurate readings
        Order("timestamp DESC").
        First(&gpsTrack).Error; err != nil {
        // Handle error
    }
    
    return &gpsTrack, nil
}
```

**Route Distance Calculation:**
```go
func (oqs *OptimizedQueryService) GetRouteDistanceOptimized(ctx context.Context, vehicleID string, startTime, endTime time.Time) (float64, error) {
    // Use PostGIS window functions for efficient distance calculation
    query := `
        WITH ordered_tracks AS (
            SELECT 
                latitude, longitude,
                LAG(latitude) OVER (ORDER BY timestamp) as prev_lat,
                LAG(longitude) OVER (ORDER BY timestamp) as prev_lng
            FROM gps_tracks 
            WHERE vehicle_id = ? AND timestamp BETWEEN ? AND ? AND accuracy <= 50
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
    // Execute query...
}
```

### Vehicle Service Optimizations

#### `OptimizedVehicleQueries` (`internal/vehicle/optimized_queries.go`)

**Key Features:**
- **Composite Index Usage**: Leverage multi-column indexes for complex filters
- **Efficient Joins**: Optimized JOIN operations with selective field loading
- **Batch Operations**: Single-query operations for multiple records
- **Statistical Queries**: Efficient aggregation queries for dashboard data

**Example:**
```go
func (ovq *OptimizedVehicleQueries) GetVehicleStatsOptimized(ctx context.Context, companyID string) (map[string]interface{}, error) {
    var stats struct {
        TotalVehicles      int64 `gorm:"column:total_vehicles"`
        ActiveVehicles     int64 `gorm:"column:active_vehicles"`
        VehiclesWithDriver int64 `gorm:"column:vehicles_with_driver"`
        // ... more fields
    }
    
    // Use single query with conditional aggregation for better performance
    query := `
        SELECT 
            COUNT(*) as total_vehicles,
            COUNT(CASE WHEN is_active = true THEN 1 END) as active_vehicles,
            COUNT(CASE WHEN driver_id IS NOT NULL THEN 1 END) as vehicles_with_driver,
            -- ... more aggregations
        FROM vehicles 
        WHERE company_id = ?
    `
    // Execute and return results...
}
```

## Redis Caching Layer

### `RedisCache` (`internal/common/cache/redis_cache.go`)

**Features:**
- **Automatic Serialization**: JSON marshaling/unmarshaling
- **Hash Operations**: Support for Redis hash data structures
- **Expiration Management**: Configurable TTL for different data types
- **Key Generation**: Standardized cache key patterns

**Cache Key Patterns:**
```go
func (rc *RedisCache) VehicleKey(vehicleID string) string {
    return fmt.Sprintf("vehicle:%s", vehicleID)
}

func (rc *RedisCache) VehicleLocationKey(vehicleID string) string {
    return fmt.Sprintf("vehicle:location:%s", vehicleID)
}

func (rc *RedisCache) GPSTrackKey(vehicleID string, timestamp time.Time) string {
    return fmt.Sprintf("gps:track:%s:%d", vehicleID, timestamp.Unix())
}
```

### `CachedTrackingService` (`internal/tracking/cached_service.go`)

**Caching Strategy:**
- **Location Data**: 30-second cache for real-time location
- **Historical Data**: 1-minute cache for location history
- **Statistics**: 10-minute cache for calculated metrics
- **Route Data**: 1-hour cache for distance calculations

**Example Cached Operation:**
```go
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
    
    // Get from database and cache result
    location, err := cts.optimizedQueries.GetCurrentLocationOptimized(ctx, vehicleID)
    if err != nil {
        return nil, err
    }
    
    cts.cache.Set(ctx, cacheKey, location, cache.LocationExpiration)
    return location, nil
}
```

## Query Performance Monitoring

### `QueryMonitor` (`internal/common/monitoring/query_monitor.go`)

**Features:**
- **Slow Query Detection**: Configurable threshold for slow queries
- **Performance Metrics**: Collection of query execution statistics
- **GORM Integration**: Automatic monitoring of all database operations
- **Performance Reports**: Detailed analysis and recommendations

**Usage:**
```go
// Initialize monitoring
monitor := monitoring.NewQueryMonitor(100*time.Millisecond, logger)
collector := monitoring.NewMetricsCollector()

// Add GORM plugin
db.Use(monitoring.NewQueryMonitorPlugin(monitor, collector))

// Monitor specific operations
err := monitor.MonitorQuery(ctx, "GetVehicleLocation", func() error {
    return db.Where("id = ?", vehicleID).First(&vehicle).Error
})
```

**Performance Report:**
```go
report := collector.GenerateReport()
// Returns recommendations like:
// - "Consider optimizing SELECT queries - 15 slow queries detected"
// - "Consider adding indexes for UPDATE queries - average duration: 150ms"
```

## Pagination Optimization

### Efficient Pagination Patterns

**Cursor-Based Pagination (Recommended for large datasets):**
```go
func (ovq *OptimizedVehicleQueries) ListVehiclesCursorOptimized(ctx context.Context, companyID string, cursor string, limit int) ([]*models.Vehicle, string, error) {
    var vehicles []*models.Vehicle
    
    query := ovq.db.WithContext(ctx).Where("company_id = ?", companyID)
    
    if cursor != "" {
        // Use cursor for efficient pagination
        query = query.Where("created_at < ?", cursor)
    }
    
    if err := query.Order("created_at DESC").Limit(limit + 1).Find(&vehicles).Error; err != nil {
        return nil, "", err
    }
    
    // Determine next cursor
    nextCursor := ""
    if len(vehicles) > limit {
        nextCursor = vehicles[limit-1].CreatedAt.Format(time.RFC3339)
        vehicles = vehicles[:limit]
    }
    
    return vehicles, nextCursor, nil
}
```

**Offset-Based Pagination (For smaller datasets):**
```go
func (ovq *OptimizedVehicleQueries) ListVehiclesOptimized(ctx context.Context, companyID string, filters VehicleFilters) ([]*models.Vehicle, int64, error) {
    // Use optimized pagination with proper indexing
    offset := (filters.Page - 1) * filters.Limit
    query = query.Offset(offset).Limit(filters.Limit)
    // Execute query...
}
```

## Best Practices

### 1. Index Design
- **Composite Indexes**: Create indexes for common query patterns
- **Partial Indexes**: Use WHERE clauses to create smaller, more efficient indexes
- **Covering Indexes**: Include frequently accessed columns in indexes
- **Index Maintenance**: Monitor index usage and remove unused indexes

### 2. Query Optimization
- **Selective Loading**: Only load necessary fields
- **Avoid N+1 Queries**: Use Preload for related data
- **Use Database Functions**: Leverage SQL functions for calculations
- **Batch Operations**: Group multiple operations into single queries

### 3. Caching Strategy
- **Cache Expiration**: Set appropriate TTL based on data volatility
- **Cache Invalidation**: Implement proper cache invalidation strategies
- **Cache Warming**: Pre-populate frequently accessed data
- **Cache Monitoring**: Monitor cache hit rates and performance

### 4. Monitoring and Alerting
- **Slow Query Alerts**: Set up alerts for queries exceeding thresholds
- **Performance Baselines**: Establish performance baselines and track deviations
- **Resource Monitoring**: Monitor database CPU, memory, and I/O usage
- **Query Analysis**: Regular analysis of query execution plans

## Performance Metrics

### Target Performance Goals
- **API Response Time**: < 200ms for 95% of requests
- **Database Query Time**: < 100ms for 95% of queries
- **Cache Hit Rate**: > 80% for frequently accessed data
- **Slow Query Rate**: < 1% of total queries

### Monitoring Queries
```sql
-- Find slow queries
SELECT query, mean_time, calls, total_time
FROM pg_stat_statements
WHERE mean_time > 100
ORDER BY mean_time DESC;

-- Index usage analysis
SELECT schemaname, tablename, indexname, idx_scan, idx_tup_read, idx_tup_fetch
FROM pg_stat_user_indexes
ORDER BY idx_scan DESC;

-- Table size analysis
SELECT schemaname, tablename, pg_size_pretty(pg_total_relation_size(schemaname||'.'||tablename)) as size
FROM pg_tables
WHERE schemaname = 'public'
ORDER BY pg_total_relation_size(schemaname||'.'||tablename) DESC;
```

## Migration and Deployment

### Running the Optimization Migration
```bash
# Apply the performance indexes
psql -d fleettracker_pro -f migrations/003_add_performance_indexes.up.sql

# Verify indexes were created
psql -d fleettracker_pro -c "\di+ idx_*"
```

### Monitoring After Deployment
1. **Query Performance**: Monitor slow query logs
2. **Index Usage**: Check index utilization statistics
3. **Cache Performance**: Monitor Redis cache hit rates
4. **Application Metrics**: Track API response times

## Troubleshooting

### Common Issues

**1. Slow Queries After Index Creation**
- Check if queries are using the new indexes
- Analyze query execution plans
- Consider additional composite indexes

**2. High Memory Usage**
- Review cache TTL settings
- Monitor Redis memory usage
- Implement cache eviction policies

**3. Index Bloat**
- Regular VACUUM and REINDEX operations
- Monitor index size growth
- Consider partial indexes for large tables

### Performance Tuning Commands
```sql
-- Update table statistics
ANALYZE;

-- Rebuild indexes
REINDEX INDEX CONCURRENTLY idx_gps_tracks_vehicle_id_timestamp;

-- Vacuum tables
VACUUM ANALYZE gps_tracks;

-- Check query execution plan
EXPLAIN (ANALYZE, BUFFERS) SELECT * FROM gps_tracks WHERE vehicle_id = 'uuid' ORDER BY timestamp DESC LIMIT 1;
```

This optimization strategy ensures the FleetTracker Pro backend can handle high-volume GPS tracking data efficiently while maintaining fast response times for all operations.
