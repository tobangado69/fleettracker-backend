package ratelimit

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
)

// RateLimitMetrics holds rate limiting metrics
type RateLimitMetrics struct {
	TotalRequests     int64             `json:"total_requests"`
	AllowedRequests   int64             `json:"allowed_requests"`
	BlockedRequests   int64             `json:"blocked_requests"`
	BlockRate         float64           `json:"block_rate"`
	AverageResponseTime time.Duration   `json:"average_response_time"`
	EndpointStats     map[string]*EndpointStats `json:"endpoint_stats"`
	UserStats         map[string]*UserStats     `json:"user_stats"`
	CompanyStats      map[string]*CompanyStats  `json:"company_stats"`
	LastUpdated       time.Time         `json:"last_updated"`
}

// EndpointStats holds statistics for a specific endpoint
type EndpointStats struct {
	Path              string    `json:"path"`
	Method            string    `json:"method"`
	TotalRequests     int64     `json:"total_requests"`
	AllowedRequests   int64     `json:"allowed_requests"`
	BlockedRequests   int64     `json:"blocked_requests"`
	BlockRate         float64   `json:"block_rate"`
	AverageResponseTime time.Duration `json:"average_response_time"`
	LastRequest       time.Time `json:"last_request"`
}

// UserStats holds statistics for a specific user
type UserStats struct {
	UserID            string    `json:"user_id"`
	TotalRequests     int64     `json:"total_requests"`
	AllowedRequests   int64     `json:"allowed_requests"`
	BlockedRequests   int64     `json:"blocked_requests"`
	BlockRate         float64   `json:"block_rate"`
	LastRequest       time.Time `json:"last_request"`
}

// CompanyStats holds statistics for a specific company
type CompanyStats struct {
	CompanyID         string    `json:"company_id"`
	TotalRequests     int64     `json:"total_requests"`
	AllowedRequests   int64     `json:"allowed_requests"`
	BlockedRequests   int64     `json:"blocked_requests"`
	BlockRate         float64   `json:"block_rate"`
	LastRequest       time.Time `json:"last_request"`
}

// RateLimitMonitor provides monitoring and metrics for rate limiting
type RateLimitMonitor struct {
	redis     *redis.Client
	metrics   *RateLimitMetrics
	mutex     sync.RWMutex
	startTime time.Time
}

// NewRateLimitMonitor creates a new rate limit monitor
func NewRateLimitMonitor(redis *redis.Client) *RateLimitMonitor {
	monitor := &RateLimitMonitor{
		redis: redis,
		metrics: &RateLimitMetrics{
			EndpointStats: make(map[string]*EndpointStats),
			UserStats:     make(map[string]*UserStats),
			CompanyStats:  make(map[string]*CompanyStats),
		},
		startTime: time.Now(),
	}
	
	// Attempt to load existing metrics from Redis
	ctx := context.Background()
	_ = monitor.loadMetricsFromRedis(ctx)
	
	return monitor
}

// RecordRequest records a rate limit request
func (rm *RateLimitMonitor) RecordRequest(ctx context.Context, path, method, userID, companyID string, allowed bool, responseTime time.Duration) {
	rm.mutex.Lock()
	defer rm.mutex.Unlock()
	
	// Update global metrics
	rm.metrics.TotalRequests++
	if allowed {
		rm.metrics.AllowedRequests++
	} else {
		rm.metrics.BlockedRequests++
	}
	
	// Update block rate
	if rm.metrics.TotalRequests > 0 {
		rm.metrics.BlockRate = float64(rm.metrics.BlockedRequests) / float64(rm.metrics.TotalRequests) * 100
	}
	
	// Update average response time
	if rm.metrics.TotalRequests == 1 {
		rm.metrics.AverageResponseTime = responseTime
	} else {
		rm.metrics.AverageResponseTime = (rm.metrics.AverageResponseTime*time.Duration(rm.metrics.TotalRequests-1) + responseTime) / time.Duration(rm.metrics.TotalRequests)
	}
	
	// Update endpoint stats
	endpointKey := fmt.Sprintf("%s:%s", method, path)
	if stats, exists := rm.metrics.EndpointStats[endpointKey]; exists {
		stats.TotalRequests++
		if allowed {
			stats.AllowedRequests++
		} else {
			stats.BlockedRequests++
		}
		stats.BlockRate = float64(stats.BlockedRequests) / float64(stats.TotalRequests) * 100
		stats.AverageResponseTime = (stats.AverageResponseTime*time.Duration(stats.TotalRequests-1) + responseTime) / time.Duration(stats.TotalRequests)
		stats.LastRequest = time.Now()
	} else {
		rm.metrics.EndpointStats[endpointKey] = &EndpointStats{
			Path:              path,
			Method:            method,
			TotalRequests:     1,
			AllowedRequests:   func() int64 { if allowed { return 1 } else { return 0 } }(),
			BlockedRequests:   func() int64 { if allowed { return 0 } else { return 1 } }(),
			BlockRate:         func() float64 { if allowed { return 0 } else { return 100 } }(),
			AverageResponseTime: responseTime,
			LastRequest:       time.Now(),
		}
	}
	
	// Update user stats
	if userID != "" {
		if stats, exists := rm.metrics.UserStats[userID]; exists {
			stats.TotalRequests++
			if allowed {
				stats.AllowedRequests++
			} else {
				stats.BlockedRequests++
			}
			stats.BlockRate = float64(stats.BlockedRequests) / float64(stats.TotalRequests) * 100
			stats.LastRequest = time.Now()
		} else {
			rm.metrics.UserStats[userID] = &UserStats{
				UserID:          userID,
				TotalRequests:   1,
				AllowedRequests: func() int64 { if allowed { return 1 } else { return 0 } }(),
				BlockedRequests: func() int64 { if allowed { return 0 } else { return 1 } }(),
				BlockRate:       func() float64 { if allowed { return 0 } else { return 100 } }(),
				LastRequest:     time.Now(),
			}
		}
	}
	
	// Update company stats
	if companyID != "" {
		if stats, exists := rm.metrics.CompanyStats[companyID]; exists {
			stats.TotalRequests++
			if allowed {
				stats.AllowedRequests++
			} else {
				stats.BlockedRequests++
			}
			stats.BlockRate = float64(stats.BlockedRequests) / float64(stats.TotalRequests) * 100
			stats.LastRequest = time.Now()
		} else {
			rm.metrics.CompanyStats[companyID] = &CompanyStats{
				CompanyID:       companyID,
				TotalRequests:   1,
				AllowedRequests: func() int64 { if allowed { return 1 } else { return 0 } }(),
				BlockedRequests: func() int64 { if allowed { return 0 } else { return 1 } }(),
				BlockRate:       func() float64 { if allowed { return 0 } else { return 100 } }(),
				LastRequest:     time.Now(),
			}
		}
	}
	
	rm.metrics.LastUpdated = time.Now()
	
	// Store metrics in Redis for persistence
	go rm.storeMetricsInRedis(ctx)
}

// GetMetrics returns current rate limiting metrics
func (rm *RateLimitMonitor) GetMetrics() *RateLimitMetrics {
	rm.mutex.RLock()
	defer rm.mutex.RUnlock()
	
	// Create a copy to avoid race conditions
	metricsCopy := *rm.metrics
	metricsCopy.EndpointStats = make(map[string]*EndpointStats)
	metricsCopy.UserStats = make(map[string]*UserStats)
	metricsCopy.CompanyStats = make(map[string]*CompanyStats)
	
	// Copy maps
	for k, v := range rm.metrics.EndpointStats {
		statsCopy := *v
		metricsCopy.EndpointStats[k] = &statsCopy
	}
	
	for k, v := range rm.metrics.UserStats {
		statsCopy := *v
		metricsCopy.UserStats[k] = &statsCopy
	}
	
	for k, v := range rm.metrics.CompanyStats {
		statsCopy := *v
		metricsCopy.CompanyStats[k] = &statsCopy
	}
	
	return &metricsCopy
}

// GetEndpointStats returns statistics for a specific endpoint
func (rm *RateLimitMonitor) GetEndpointStats(path, method string) *EndpointStats {
	rm.mutex.RLock()
	defer rm.mutex.RUnlock()
	
	endpointKey := fmt.Sprintf("%s:%s", method, path)
	if stats, exists := rm.metrics.EndpointStats[endpointKey]; exists {
		statsCopy := *stats
		return &statsCopy
	}
	
	return nil
}

// GetUserStats returns statistics for a specific user
func (rm *RateLimitMonitor) GetUserStats(userID string) *UserStats {
	rm.mutex.RLock()
	defer rm.mutex.RUnlock()
	
	if stats, exists := rm.metrics.UserStats[userID]; exists {
		statsCopy := *stats
		return &statsCopy
	}
	
	return nil
}

// GetCompanyStats returns statistics for a specific company
func (rm *RateLimitMonitor) GetCompanyStats(companyID string) *CompanyStats {
	rm.mutex.RLock()
	defer rm.mutex.RUnlock()
	
	if stats, exists := rm.metrics.CompanyStats[companyID]; exists {
		statsCopy := *stats
		return &statsCopy
	}
	
	return nil
}

// GetTopBlockedEndpoints returns the top endpoints with highest block rates
func (rm *RateLimitMonitor) GetTopBlockedEndpoints(limit int) []*EndpointStats {
	rm.mutex.RLock()
	defer rm.mutex.RUnlock()
	
	var endpoints []*EndpointStats
	for _, stats := range rm.metrics.EndpointStats {
		endpoints = append(endpoints, stats)
	}
	
	// Sort by block rate (descending)
	for i := 0; i < len(endpoints)-1; i++ {
		for j := i + 1; j < len(endpoints); j++ {
			if endpoints[i].BlockRate < endpoints[j].BlockRate {
				endpoints[i], endpoints[j] = endpoints[j], endpoints[i]
			}
		}
	}
	
	if limit > 0 && limit < len(endpoints) {
		endpoints = endpoints[:limit]
	}
	
	return endpoints
}

// GetTopBlockedUsers returns the top users with highest block rates
func (rm *RateLimitMonitor) GetTopBlockedUsers(limit int) []*UserStats {
	rm.mutex.RLock()
	defer rm.mutex.RUnlock()
	
	var users []*UserStats
	for _, stats := range rm.metrics.UserStats {
		users = append(users, stats)
	}
	
	// Sort by block rate (descending)
	for i := 0; i < len(users)-1; i++ {
		for j := i + 1; j < len(users); j++ {
			if users[i].BlockRate < users[j].BlockRate {
				users[i], users[j] = users[j], users[i]
			}
		}
	}
	
	if limit > 0 && limit < len(users) {
		users = users[:limit]
	}
	
	return users
}

// GetTopBlockedCompanies returns the top companies with highest block rates
func (rm *RateLimitMonitor) GetTopBlockedCompanies(limit int) []*CompanyStats {
	rm.mutex.RLock()
	defer rm.mutex.RUnlock()
	
	var companies []*CompanyStats
	for _, stats := range rm.metrics.CompanyStats {
		companies = append(companies, stats)
	}
	
	// Sort by block rate (descending)
	for i := 0; i < len(companies)-1; i++ {
		for j := i + 1; j < len(companies); j++ {
			if companies[i].BlockRate < companies[j].BlockRate {
				companies[i], companies[j] = companies[j], companies[i]
			}
		}
	}
	
	if limit > 0 && limit < len(companies) {
		companies = companies[:limit]
	}
	
	return companies
}

// ResetMetrics resets all metrics
func (rm *RateLimitMonitor) ResetMetrics() {
	rm.mutex.Lock()
	defer rm.mutex.Unlock()
	
	rm.metrics = &RateLimitMetrics{
		EndpointStats: make(map[string]*EndpointStats),
		UserStats:     make(map[string]*UserStats),
		CompanyStats:  make(map[string]*CompanyStats),
	}
	rm.startTime = time.Now()
}

// storeMetricsInRedis stores metrics in Redis for persistence
func (rm *RateLimitMonitor) storeMetricsInRedis(ctx context.Context) {
	metrics := rm.GetMetrics()
	data, err := json.Marshal(metrics)
	if err != nil {
		return
	}
	
	rm.redis.Set(ctx, "rate_limit:metrics", data, 24*time.Hour)
}

// loadMetricsFromRedis loads metrics from Redis
func (rm *RateLimitMonitor) loadMetricsFromRedis(ctx context.Context) error {
	data, err := rm.redis.Get(ctx, "rate_limit:metrics").Result()
	if err != nil {
		if err == redis.Nil {
			return nil // No metrics stored yet
		}
		return err
	}
	
	var metrics RateLimitMetrics
	if err := json.Unmarshal([]byte(data), &metrics); err != nil {
		return err
	}
	
	rm.mutex.Lock()
	rm.metrics = &metrics
	rm.mutex.Unlock()
	
	return nil
}

// GetUptime returns the uptime of the rate limit monitor
func (rm *RateLimitMonitor) GetUptime() time.Duration {
	return time.Since(rm.startTime)
}

// GetHealthStatus returns the health status of rate limiting
func (rm *RateLimitMonitor) GetHealthStatus() map[string]interface{} {
	metrics := rm.GetMetrics()
	
	status := map[string]interface{}{
		"status": "healthy",
		"uptime": rm.GetUptime().String(),
		"total_requests": metrics.TotalRequests,
		"block_rate": metrics.BlockRate,
		"average_response_time": metrics.AverageResponseTime.String(),
		"endpoint_count": len(metrics.EndpointStats),
		"user_count": len(metrics.UserStats),
		"company_count": len(metrics.CompanyStats),
	}
	
	// Check if block rate is too high
	if metrics.BlockRate > 50 {
		status["status"] = "warning"
		status["warning"] = "High block rate detected"
	}
	
	// Check if response time is too high
	if metrics.AverageResponseTime > 100*time.Millisecond {
		status["status"] = "warning"
		status["warning"] = "High response time detected"
	}
	
	return status
}
