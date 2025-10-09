package monitoring

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
)

// CacheMetrics tracks cache hit/miss rates and statistics
type CacheMetrics struct {
	redis      *redis.Client
	mu         sync.RWMutex
	hits       int64
	misses     int64
	errors     int64
	totalCalls int64
	startTime  time.Time
}

// NewCacheMetrics creates a new cache metrics tracker
func NewCacheMetrics(redis *redis.Client) *CacheMetrics {
	return &CacheMetrics{
		redis:     redis,
		startTime: time.Now(),
	}
}

// RecordHit records a cache hit
func (cm *CacheMetrics) RecordHit() {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	cm.hits++
	cm.totalCalls++
}

// RecordMiss records a cache miss
func (cm *CacheMetrics) RecordMiss() {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	cm.misses++
	cm.totalCalls++
}

// RecordError records a cache error
func (cm *CacheMetrics) RecordError() {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	cm.errors++
}

// GetStats returns current cache statistics
func (cm *CacheMetrics) GetStats() CacheStats {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	hitRate := 0.0
	if cm.totalCalls > 0 {
		hitRate = float64(cm.hits) / float64(cm.totalCalls) * 100
	}

	missRate := 0.0
	if cm.totalCalls > 0 {
		missRate = float64(cm.misses) / float64(cm.totalCalls) * 100
	}

	uptime := time.Since(cm.startTime)

	return CacheStats{
		Hits:       cm.hits,
		Misses:     cm.misses,
		Errors:     cm.errors,
		TotalCalls: cm.totalCalls,
		HitRate:    hitRate,
		MissRate:   missRate,
		Uptime:     uptime.String(),
		StartTime:  cm.startTime,
	}
}

// Reset resets all metrics
func (cm *CacheMetrics) Reset() {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	cm.hits = 0
	cm.misses = 0
	cm.errors = 0
	cm.totalCalls = 0
	cm.startTime = time.Now()
}

// GetRedisInfo returns Redis server information
func (cm *CacheMetrics) GetRedisInfo(ctx context.Context) (map[string]interface{}, error) {
	info, err := cm.redis.Info(ctx).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get Redis info: %w", err)
	}

	// Get memory stats
	memoryInfo, err := cm.redis.Info(ctx, "memory").Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get Redis memory info: %w", err)
	}

	// Get stats
	statsInfo, err := cm.redis.Info(ctx, "stats").Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get Redis stats: %w", err)
	}

	// Get keyspace info
	keyspaceInfo, err := cm.redis.Info(ctx, "keyspace").Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get Redis keyspace info: %w", err)
	}

	return map[string]interface{}{
		"server":   info,
		"memory":   memoryInfo,
		"stats":    statsInfo,
		"keyspace": keyspaceInfo,
	}, nil
}

// GetKeyCount returns the number of keys matching a pattern
func (cm *CacheMetrics) GetKeyCount(ctx context.Context, pattern string) (int, error) {
	keys, err := cm.redis.Keys(ctx, pattern).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to get keys: %w", err)
	}
	return len(keys), nil
}

// GetMemoryUsage returns memory usage for keys matching a pattern
func (cm *CacheMetrics) GetMemoryUsage(ctx context.Context, pattern string) (int64, error) {
	keys, err := cm.redis.Keys(ctx, pattern).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to get keys: %w", err)
	}

	totalMemory := int64(0)
	for _, key := range keys {
		memory, err := cm.redis.MemoryUsage(ctx, key).Result()
		if err == nil {
			totalMemory += memory
		}
	}

	return totalMemory, nil
}

// GetTTLStats returns TTL statistics for keys matching a pattern
func (cm *CacheMetrics) GetTTLStats(ctx context.Context, pattern string) (TTLStats, error) {
	keys, err := cm.redis.Keys(ctx, pattern).Result()
	if err != nil {
		return TTLStats{}, fmt.Errorf("failed to get keys: %w", err)
	}

	stats := TTLStats{
		TotalKeys: len(keys),
	}

	for _, key := range keys {
		ttl, err := cm.redis.TTL(ctx, key).Result()
		if err != nil {
			continue
		}

		switch ttl {
case -1:
			stats.NoExpiry++
		case -2:
			stats.Expired++
		default:
			stats.WithExpiry++
			stats.TotalTTL += ttl
		}
	}

	if stats.WithExpiry > 0 {
		stats.AverageTTL = stats.TotalTTL / time.Duration(stats.WithExpiry)
	}

	return stats, nil
}

// GetCacheHealth performs health checks on the cache
func (cm *CacheMetrics) GetCacheHealth(ctx context.Context) CacheHealth {
	health := CacheHealth{
		Status:    "healthy",
		Timestamp: time.Now(),
	}

	// Check Redis ping
	_, err := cm.redis.Ping(ctx).Result()
	if err != nil {
		health.Status = "unhealthy"
		health.Errors = append(health.Errors, fmt.Sprintf("Redis ping failed: %v", err))
		return health
	}

	// Check memory usage
	info, err := cm.redis.Info(ctx, "memory").Result()
	if err != nil {
		health.Status = "degraded"
		health.Errors = append(health.Errors, fmt.Sprintf("Failed to get memory info: %v", err))
	} else {
		health.Details["memory"] = info
	}

	// Check stats
	stats := cm.GetStats()
	health.Details["stats"] = stats

	// Check hit rate
	if stats.HitRate < 50.0 && stats.TotalCalls > 100 {
		health.Status = "degraded"
		health.Errors = append(health.Errors, fmt.Sprintf("Low cache hit rate: %.2f%%", stats.HitRate))
	}

	return health
}

// ExportMetrics exports metrics in Prometheus format
func (cm *CacheMetrics) ExportMetrics() string {
	stats := cm.GetStats()
	
	return fmt.Sprintf(`# HELP cache_hits_total Total number of cache hits
# TYPE cache_hits_total counter
cache_hits_total %d

# HELP cache_misses_total Total number of cache misses
# TYPE cache_misses_total counter
cache_misses_total %d

# HELP cache_errors_total Total number of cache errors
# TYPE cache_errors_total counter
cache_errors_total %d

# HELP cache_hit_rate Cache hit rate percentage
# TYPE cache_hit_rate gauge
cache_hit_rate %.2f

# HELP cache_miss_rate Cache miss rate percentage
# TYPE cache_miss_rate gauge
cache_miss_rate %.2f

# HELP cache_total_calls_total Total number of cache calls
# TYPE cache_total_calls_total counter
cache_total_calls_total %d
`, stats.Hits, stats.Misses, stats.Errors, stats.HitRate, stats.MissRate, stats.TotalCalls)
}

// CacheStats represents cache statistics
type CacheStats struct {
	Hits       int64     `json:"hits"`
	Misses     int64     `json:"misses"`
	Errors     int64     `json:"errors"`
	TotalCalls int64     `json:"total_calls"`
	HitRate    float64   `json:"hit_rate"`
	MissRate   float64   `json:"miss_rate"`
	Uptime     string    `json:"uptime"`
	StartTime  time.Time `json:"start_time"`
}

// TTLStats represents TTL statistics
type TTLStats struct {
	TotalKeys  int           `json:"total_keys"`
	WithExpiry int           `json:"with_expiry"`
	NoExpiry   int           `json:"no_expiry"`
	Expired    int           `json:"expired"`
	AverageTTL time.Duration `json:"average_ttl"`
	TotalTTL   time.Duration `json:"total_ttl"`
}

// CacheHealth represents cache health status
type CacheHealth struct {
	Status    string                 `json:"status"`
	Timestamp time.Time              `json:"timestamp"`
	Errors    []string               `json:"errors,omitempty"`
	Details   map[string]interface{} `json:"details"`
}

func init() {
	// Initialize Details map
	_ = CacheHealth{
		Details: make(map[string]interface{}),
	}
}

// CacheMetricsHandler provides HTTP handlers for cache metrics
type CacheMetricsHandler struct {
	metrics *CacheMetrics
}

// NewCacheMetricsHandler creates a new cache metrics handler
func NewCacheMetricsHandler(metrics *CacheMetrics) *CacheMetricsHandler {
	return &CacheMetricsHandler{
		metrics: metrics,
	}
}

// GetStats returns cache statistics
func (h *CacheMetricsHandler) GetStats(c *gin.Context) {
	stats := h.metrics.GetStats()
	c.JSON(200, gin.H{
		"success": true,
		"data":    stats,
	})
}

// GetHealth returns cache health status
func (h *CacheMetricsHandler) GetHealth(c *gin.Context) {
	health := h.metrics.GetCacheHealth(c.Request.Context())
	
	status := 200
	switch health.Status {
case "degraded":
		status = 207 // Multi-Status
	case "unhealthy":
		status = 503 // Service Unavailable
	}
	
	c.JSON(status, gin.H{
		"success": health.Status == "healthy",
		"data":    health,
	})
}

// GetRedisInfo returns Redis server information
func (h *CacheMetricsHandler) GetRedisInfo(c *gin.Context) {
	info, err := h.metrics.GetRedisInfo(c.Request.Context())
	if err != nil {
		c.JSON(500, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}
	
	c.JSON(200, gin.H{
		"success": true,
		"data":    info,
	})
}

// GetKeyStats returns key statistics for a pattern
func (h *CacheMetricsHandler) GetKeyStats(c *gin.Context) {
	pattern := c.DefaultQuery("pattern", "*")
	
	keyCount, err := h.metrics.GetKeyCount(c.Request.Context(), pattern)
	if err != nil {
		c.JSON(500, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}
	
	memoryUsage, err := h.metrics.GetMemoryUsage(c.Request.Context(), pattern)
	if err != nil {
		c.JSON(500, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}
	
	ttlStats, err := h.metrics.GetTTLStats(c.Request.Context(), pattern)
	if err != nil {
		c.JSON(500, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}
	
	c.JSON(200, gin.H{
		"success": true,
		"data": gin.H{
			"pattern":      pattern,
			"key_count":    keyCount,
			"memory_bytes": memoryUsage,
			"ttl_stats":    ttlStats,
		},
	})
}

// GetPrometheusMetrics returns metrics in Prometheus format
func (h *CacheMetricsHandler) GetPrometheusMetrics(c *gin.Context) {
	metrics := h.metrics.ExportMetrics()
	c.String(200, metrics)
}

// ResetMetrics resets all metrics
func (h *CacheMetricsHandler) ResetMetrics(c *gin.Context) {
	h.metrics.Reset()
	c.JSON(200, gin.H{
		"success": true,
		"message": "Metrics reset successfully",
	})
}

// GetDashboard returns a comprehensive cache dashboard
func (h *CacheMetricsHandler) GetDashboard(c *gin.Context) {
	stats := h.metrics.GetStats()
	health := h.metrics.GetCacheHealth(c.Request.Context())
	
	// Get key stats for common patterns
	vehicleKeys, _ := h.metrics.GetKeyCount(c.Request.Context(), "vehicle:*")
	driverKeys, _ := h.metrics.GetKeyCount(c.Request.Context(), "driver:*")
	trackingKeys, _ := h.metrics.GetKeyCount(c.Request.Context(), "gps:*")
	
	dashboard := gin.H{
		"stats":  stats,
		"health": health,
		"keys_by_type": gin.H{
			"vehicles": vehicleKeys,
			"drivers":  driverKeys,
			"tracking": trackingKeys,
		},
	}
	
	c.JSON(200, gin.H{
		"success": true,
		"data":    dashboard,
	})
}

// TrackedCacheMiddleware wraps cache operations with metrics tracking
func TrackedCacheMiddleware(metrics *CacheMetrics) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check for cache status header from cache middleware
		c.Next()
		
		cacheStatus := c.Writer.Header().Get("X-Cache-Status")
		switch cacheStatus {
		case "HIT":
			metrics.RecordHit()
		case "MISS":
			metrics.RecordMiss()
		}
	}
}

// CacheStatsToJSON converts cache stats to JSON
func CacheStatsToJSON(stats CacheStats) ([]byte, error) {
	return json.Marshal(stats)
}

// CacheHealthToJSON converts cache health to JSON
func CacheHealthToJSON(health CacheHealth) ([]byte, error) {
	return json.Marshal(health)
}

