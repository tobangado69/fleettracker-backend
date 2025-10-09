package analytics

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
)

// AnalyticsCache provides Redis-based caching for analytics data
type AnalyticsCache struct {
	redis *redis.Client
}

// NewAnalyticsCache creates a new analytics cache
func NewAnalyticsCache(redis *redis.Client) *AnalyticsCache {
	return &AnalyticsCache{
		redis: redis,
	}
}

// SetAnalyticsInCache stores analytics data in Redis cache
func (ac *AnalyticsCache) SetAnalyticsInCache(ctx context.Context, req *AnalyticsRequest, response *AnalyticsResponse) error {
	key := ac.generateCacheKey(req)
	ttl := ac.getTTLForReportType(req.ReportType)
	
	// Marshal the response
	jsonData, err := json.Marshal(response)
	if err != nil {
		return fmt.Errorf("failed to marshal analytics response: %w", err)
	}
	
	// Store in Redis
	if err := ac.redis.Set(ctx, key, jsonData, ttl).Err(); err != nil {
		return fmt.Errorf("failed to set analytics data in cache: %w", err)
	}
	
	return nil
}

// GetAnalyticsFromCache retrieves analytics data from Redis cache
func (ac *AnalyticsCache) GetAnalyticsFromCache(ctx context.Context, req *AnalyticsRequest) (*AnalyticsResponse, bool, error) {
	key := ac.generateCacheKey(req)
	
	// Get from Redis
	jsonData, err := ac.redis.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, false, nil // Cache miss
		}
		return nil, false, fmt.Errorf("failed to get analytics data from cache: %w", err)
	}
	
	// Unmarshal the response
	var response AnalyticsResponse
	if err := json.Unmarshal([]byte(jsonData), &response); err != nil {
		return nil, false, fmt.Errorf("failed to unmarshal cached analytics response: %w", err)
	}
	
	// Mark as from cache
	response.FromCache = true
	response.CacheHit = true
	
	return &response, true, nil
}

// InvalidateAnalyticsCache invalidates analytics cache for a specific request or company
func (ac *AnalyticsCache) InvalidateAnalyticsCache(ctx context.Context, req *AnalyticsRequest) error {
	// Invalidate specific analytics cache
	key := ac.generateCacheKey(req)
	if err := ac.redis.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("failed to invalidate specific analytics cache: %w", err)
	}
	
	// Optionally, invalidate all analytics for a company
	pattern := fmt.Sprintf("analytics:company:%s:*", req.CompanyID)
	keys, err := ac.redis.Keys(ctx, pattern).Result()
	if err != nil {
		return fmt.Errorf("failed to get analytics cache keys: %w", err)
	}
	
	if len(keys) > 0 {
		if err := ac.redis.Del(ctx, keys...).Err(); err != nil {
			return fmt.Errorf("failed to invalidate company analytics cache: %w", err)
		}
	}
	
	return nil
}

// InvalidateAllAnalyticsCache invalidates all analytics cache
func (ac *AnalyticsCache) InvalidateAllAnalyticsCache(ctx context.Context) error {
	pattern := "analytics:*"
	keys, err := ac.redis.Keys(ctx, pattern).Result()
	if err != nil {
		return fmt.Errorf("failed to get all analytics cache keys: %w", err)
	}
	
	if len(keys) > 0 {
		if err := ac.redis.Del(ctx, keys...).Err(); err != nil {
			return fmt.Errorf("failed to invalidate all analytics cache: %w", err)
		}
	}
	
	return nil
}

// GetAnalyticsCacheStats gets analytics cache statistics
func (ac *AnalyticsCache) GetAnalyticsCacheStats(ctx context.Context) (map[string]interface{}, error) {
	stats := make(map[string]interface{})
	
	// Get total analytics cache keys
	pattern := "analytics:*"
	keys, err := ac.redis.Keys(ctx, pattern).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get analytics cache keys: %w", err)
	}
	
	stats["total_keys"] = len(keys)
	
	// Get cache size by report type
	reportTypes := []string{
		ReportTypeFleetOverview,
		ReportTypeDriverPerformance,
		ReportTypeFuelAnalytics,
		ReportTypeMaintenanceCosts,
		ReportTypeRouteEfficiency,
		ReportTypeGeofenceActivity,
		ReportTypeComplianceReport,
		ReportTypeCostAnalysis,
		ReportTypeUtilizationReport,
		ReportTypePredictiveInsights,
	}
	
	reportTypeStats := make(map[string]int)
	for _, reportType := range reportTypes {
		pattern := fmt.Sprintf("analytics:*:%s:*", reportType)
		keys, err := ac.redis.Keys(ctx, pattern).Result()
		if err != nil {
			continue
		}
		reportTypeStats[reportType] = len(keys)
	}
	
	stats["by_report_type"] = reportTypeStats
	
	// Get cache memory usage
	info, err := ac.redis.Info(ctx, "memory").Result()
	if err == nil {
		stats["memory_info"] = info
	}
	
	return stats, nil
}

// generateCacheKey creates a unique cache key based on analytics request parameters
func (ac *AnalyticsCache) generateCacheKey(req *AnalyticsRequest) string {
	// Create a hash of the request parameters to ensure uniqueness
	keyString := fmt.Sprintf("%s:%s:%s:%s:%s:%v:%v:%v",
		req.CompanyID,
		req.UserID,
		req.ReportType,
		req.DateRange.StartDate.Format("2006-01-02"),
		req.DateRange.EndDate.Format("2006-01-02"),
		req.DateRange.Period,
		req.Filters,
		req.GroupBy,
	)
	
	hash := md5.Sum([]byte(keyString))
	return fmt.Sprintf("analytics:company:%s:report:%s:%s", req.CompanyID, req.ReportType, hex.EncodeToString(hash[:]))
}

// getTTLForReportType returns the TTL for a given report type
func (ac *AnalyticsCache) getTTLForReportType(reportType string) time.Duration {
	switch reportType {
	case ReportTypeFleetOverview:
		return 15 * time.Minute // Fleet overview changes frequently
	case ReportTypeDriverPerformance:
		return 30 * time.Minute // Driver performance updates less frequently
	case ReportTypeFuelAnalytics:
		return 20 * time.Minute // Fuel data updates regularly
	case ReportTypeMaintenanceCosts:
		return 1 * time.Hour // Maintenance costs change less frequently
	case ReportTypeRouteEfficiency:
		return 10 * time.Minute // Route efficiency changes frequently
	case ReportTypeGeofenceActivity:
		return 5 * time.Minute // Geofence activity is real-time
	case ReportTypeComplianceReport:
		return 2 * time.Hour // Compliance reports are less dynamic
	case ReportTypeCostAnalysis:
		return 1 * time.Hour // Cost analysis updates periodically
	case ReportTypeUtilizationReport:
		return 15 * time.Minute // Utilization changes frequently
	case ReportTypePredictiveInsights:
		return 30 * time.Minute // Predictive insights update periodically
	default:
		return 15 * time.Minute // Default TTL
	}
}

// SetDashboardData stores dashboard data in cache
func (ac *AnalyticsCache) SetDashboardData(ctx context.Context, companyID string, data interface{}) error {
	key := fmt.Sprintf("analytics:dashboard:%s", companyID)
	ttl := 10 * time.Minute // Dashboard data updates frequently
	
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal dashboard data: %w", err)
	}
	
	if err := ac.redis.Set(ctx, key, jsonData, ttl).Err(); err != nil {
		return fmt.Errorf("failed to set dashboard data in cache: %w", err)
	}
	
	return nil
}

// GetDashboardData retrieves dashboard data from cache
func (ac *AnalyticsCache) GetDashboardData(ctx context.Context, companyID string) (interface{}, bool, error) {
	key := fmt.Sprintf("analytics:dashboard:%s", companyID)
	
	jsonData, err := ac.redis.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, false, nil // Cache miss
		}
		return nil, false, fmt.Errorf("failed to get dashboard data from cache: %w", err)
	}
	
	var data interface{}
	if err := json.Unmarshal([]byte(jsonData), &data); err != nil {
		return nil, false, fmt.Errorf("failed to unmarshal cached dashboard data: %w", err)
	}
	
	return data, true, nil
}

// SetReportData stores report data in cache
func (ac *AnalyticsCache) SetReportData(ctx context.Context, companyID, reportType string, data interface{}) error {
	key := fmt.Sprintf("analytics:report:%s:%s", companyID, reportType)
	ttl := ac.getTTLForReportType(reportType)
	
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal report data: %w", err)
	}
	
	if err := ac.redis.Set(ctx, key, jsonData, ttl).Err(); err != nil {
		return fmt.Errorf("failed to set report data in cache: %w", err)
	}
	
	return nil
}

// GetReportData retrieves report data from cache
func (ac *AnalyticsCache) GetReportData(ctx context.Context, companyID, reportType string) (interface{}, bool, error) {
	key := fmt.Sprintf("analytics:report:%s:%s", companyID, reportType)
	
	jsonData, err := ac.redis.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, false, nil // Cache miss
		}
		return nil, false, fmt.Errorf("failed to get report data from cache: %w", err)
	}
	
	var data interface{}
	if err := json.Unmarshal([]byte(jsonData), &data); err != nil {
		return nil, false, fmt.Errorf("failed to unmarshal cached report data: %w", err)
	}
	
	return data, true, nil
}

// SetMetricsData stores metrics data in cache
func (ac *AnalyticsCache) SetMetricsData(ctx context.Context, companyID, metricType string, data interface{}) error {
	key := fmt.Sprintf("analytics:metrics:%s:%s", companyID, metricType)
	ttl := 5 * time.Minute // Metrics update frequently
	
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal metrics data: %w", err)
	}
	
	if err := ac.redis.Set(ctx, key, jsonData, ttl).Err(); err != nil {
		return fmt.Errorf("failed to set metrics data in cache: %w", err)
	}
	
	return nil
}

// GetMetricsData retrieves metrics data from cache
func (ac *AnalyticsCache) GetMetricsData(ctx context.Context, companyID, metricType string) (interface{}, bool, error) {
	key := fmt.Sprintf("analytics:metrics:%s:%s", companyID, metricType)
	
	jsonData, err := ac.redis.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, false, nil // Cache miss
		}
		return nil, false, fmt.Errorf("failed to get metrics data from cache: %w", err)
	}
	
	var data interface{}
	if err := json.Unmarshal([]byte(jsonData), &data); err != nil {
		return nil, false, fmt.Errorf("failed to unmarshal cached metrics data: %w", err)
	}
	
	return data, true, nil
}

// CleanupExpiredCache removes expired cache entries
func (ac *AnalyticsCache) CleanupExpiredCache(ctx context.Context) error {
	// Redis automatically handles TTL expiration, but we can clean up manually if needed
	pattern := "analytics:*"
	keys, err := ac.redis.Keys(ctx, pattern).Result()
	if err != nil {
		return fmt.Errorf("failed to get analytics cache keys: %w", err)
	}
	
	// Check TTL for each key and remove if expired
	for _, key := range keys {
		ttl, err := ac.redis.TTL(ctx, key).Result()
		if err != nil {
			continue
		}
		
		// If TTL is -1 (no expiration) or -2 (key doesn't exist), skip
		if ttl <= 0 {
			continue
		}
		
		// If TTL is very low (less than 1 minute), we can consider it expired
		if ttl < time.Minute {
			ac.redis.Del(ctx, key)
		}
	}
	
	return nil
}
