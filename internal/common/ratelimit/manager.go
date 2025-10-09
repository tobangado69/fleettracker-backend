package ratelimit

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
)

// EndpointConfig holds rate limiting configuration for specific endpoints
type EndpointConfig struct {
	Path      string           `json:"path"`
	Method    string           `json:"method"`
	Config    *RateLimitConfig `json:"config"`
	UserLimit *RateLimitConfig `json:"user_limit,omitempty"` // User-specific limit
	CompanyLimit *RateLimitConfig `json:"company_limit,omitempty"` // Company-specific limit
}

// RateLimitManager manages multiple rate limiters for different endpoints
type RateLimitManager struct {
	redis           *redis.Client
	defaultConfig   *RateLimitConfig
	endpointConfigs map[string]*EndpointConfig
	limiters        map[string]*RateLimiter
}

// NewRateLimitManager creates a new rate limit manager
func NewRateLimitManager(redis *redis.Client, defaultConfig *RateLimitConfig) *RateLimitManager {
	if defaultConfig == nil {
		defaultConfig = &RateLimitConfig{
			Strategy: FixedWindow,
			Requests: 100,
			Window:   1 * time.Minute,
		}
	}
	
	manager := &RateLimitManager{
		redis:           redis,
		defaultConfig:   defaultConfig,
		endpointConfigs: make(map[string]*EndpointConfig),
		limiters:        make(map[string]*RateLimiter),
	}
	
	// Initialize with default configurations
	manager.initializeDefaultConfigs()
	
	return manager
}

// initializeDefaultConfigs sets up default rate limiting configurations
func (rm *RateLimitManager) initializeDefaultConfigs() {
	// Authentication endpoints - stricter limits
	rm.AddEndpointConfig(&EndpointConfig{
		Path:   "/api/v1/auth/login",
		Method: "POST",
		Config: &RateLimitConfig{
			Strategy: FixedWindow,
			Requests: 5,
			Window:   5 * time.Minute,
		},
	})
	
	rm.AddEndpointConfig(&EndpointConfig{
		Path:   "/api/v1/auth/register",
		Method: "POST",
		Config: &RateLimitConfig{
			Strategy: FixedWindow,
			Requests: 3,
			Window:   10 * time.Minute,
		},
	})
	
	rm.AddEndpointConfig(&EndpointConfig{
		Path:   "/api/v1/auth/forgot-password",
		Method: "POST",
		Config: &RateLimitConfig{
			Strategy: FixedWindow,
			Requests: 3,
			Window:   15 * time.Minute,
		},
	})
	
	// GPS tracking endpoints - higher limits for real-time data
	rm.AddEndpointConfig(&EndpointConfig{
		Path:   "/api/v1/tracking/gps",
		Method: "POST",
		Config: &RateLimitConfig{
			Strategy: TokenBucket,
			Requests: 1000,
			Window:   1 * time.Minute,
			Burst:    100,
			RefillRate: 50,
		},
	})
	
	// Vehicle management endpoints
	rm.AddEndpointConfig(&EndpointConfig{
		Path:   "/api/v1/vehicles",
		Method: "GET",
		Config: &RateLimitConfig{
			Strategy: FixedWindow,
			Requests: 200,
			Window:   1 * time.Minute,
		},
	})
	
	rm.AddEndpointConfig(&EndpointConfig{
		Path:   "/api/v1/vehicles",
		Method: "POST",
		Config: &RateLimitConfig{
			Strategy: FixedWindow,
			Requests: 20,
			Window:   1 * time.Minute,
		},
	})
	
	// Driver management endpoints
	rm.AddEndpointConfig(&EndpointConfig{
		Path:   "/api/v1/drivers",
		Method: "GET",
		Config: &RateLimitConfig{
			Strategy: FixedWindow,
			Requests: 200,
			Window:   1 * time.Minute,
		},
	})
	
	rm.AddEndpointConfig(&EndpointConfig{
		Path:   "/api/v1/drivers",
		Method: "POST",
		Config: &RateLimitConfig{
			Strategy: FixedWindow,
			Requests: 20,
			Window:   1 * time.Minute,
		},
	})
	
	// Analytics endpoints - moderate limits
	rm.AddEndpointConfig(&EndpointConfig{
		Path:   "/api/v1/analytics",
		Method: "GET",
		Config: &RateLimitConfig{
			Strategy: SlidingWindow,
			Requests: 100,
			Window:   1 * time.Minute,
		},
	})
	
	// Payment endpoints - stricter limits for security
	rm.AddEndpointConfig(&EndpointConfig{
		Path:   "/api/v1/payments",
		Method: "POST",
		Config: &RateLimitConfig{
			Strategy: FixedWindow,
			Requests: 10,
			Window:   1 * time.Minute,
		},
	})
	
	// WebSocket connections
	rm.AddEndpointConfig(&EndpointConfig{
		Path:   "/ws/tracking",
		Method: "GET",
		Config: &RateLimitConfig{
			Strategy: TokenBucket,
			Requests: 10,
			Window:   1 * time.Minute,
			Burst:    5,
			RefillRate: 2,
		},
	})
}

// AddEndpointConfig adds a rate limiting configuration for an endpoint
func (rm *RateLimitManager) AddEndpointConfig(config *EndpointConfig) {
	key := rm.getEndpointKey(config.Path, config.Method)
	rm.endpointConfigs[key] = config
	
	// Create rate limiter for this endpoint
	limiter := NewRateLimiter(rm.redis, config.Config)
	rm.limiters[key] = limiter
}

// getEndpointKey generates a key for endpoint configuration
func (rm *RateLimitManager) getEndpointKey(path, method string) string {
	return fmt.Sprintf("%s:%s", method, path)
}

// matchesPattern checks if a path and method match a pattern
func (rm *RateLimitManager) matchesPattern(path, method, pattern string) bool {
	parts := strings.Split(pattern, ":")
	if len(parts) != 2 {
		return false
	}
	
	patternMethod := parts[0]
	patternPath := parts[1]
	
	// Check method
	if patternMethod != method && patternMethod != "*" {
		return false
	}
	
	// Check path pattern
	if strings.Contains(patternPath, "*") {
		// Simple wildcard matching
		patternParts := strings.Split(patternPath, "*")
		if len(patternParts) == 2 {
			return strings.HasPrefix(path, patternParts[0]) && strings.HasSuffix(path, patternParts[1])
		}
	}
	
	return path == patternPath
}

// getLimiterForEndpoint gets the rate limiter for an endpoint
func (rm *RateLimitManager) getLimiterForEndpoint(path, method string) *RateLimiter {
	key := rm.getEndpointKey(path, method)
	if limiter, exists := rm.limiters[key]; exists {
		return limiter
	}
	
	// Try pattern matching
	for endpointKey, limiter := range rm.limiters {
		if rm.matchesPattern(path, method, endpointKey) {
			return limiter
		}
	}
	
	// Return default limiter
	return NewRateLimiter(rm.redis, rm.defaultConfig)
}

// Middleware returns a Gin middleware that applies rate limiting based on endpoint
func (rm *RateLimitManager) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		path := c.Request.URL.Path
		method := c.Request.Method
		
		// Get rate limiter for this endpoint
		limiter := rm.getLimiterForEndpoint(path, method)
		
		// Apply rate limiting
		limiter.Middleware()(c)
	}
}

// UserSpecificMiddleware returns a middleware that applies user-specific rate limiting
func (rm *RateLimitManager) UserSpecificMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		path := c.Request.URL.Path
		method := c.Request.Method
		
		// Get endpoint configuration
		key := rm.getEndpointKey(path, method)
		config, exists := rm.endpointConfigs[key]
		if !exists {
			// Try pattern matching
			for endpointKey, endpointConfig := range rm.endpointConfigs {
				if rm.matchesPattern(path, method, endpointKey) {
					config = endpointConfig
					break
				}
			}
		}
		
		// Use user-specific limit if available
		if config != nil && config.UserLimit != nil {
			// Create custom key function for user-specific limiting
			userConfig := *config.UserLimit
			userConfig.KeyFunc = func(c *gin.Context) string {
				userID, exists := c.Get("user_id")
				if !exists {
					return fmt.Sprintf("rate_limit:ip:%s", c.ClientIP())
				}
				return fmt.Sprintf("rate_limit:user:%v", userID)
			}
			
			userLimiter := NewRateLimiter(rm.redis, &userConfig)
			userLimiter.Middleware()(c)
			return
		}
		
		// Fallback to regular rate limiting
		rm.Middleware()(c)
	}
}

// CompanySpecificMiddleware returns a middleware that applies company-specific rate limiting
func (rm *RateLimitManager) CompanySpecificMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		path := c.Request.URL.Path
		method := c.Request.Method
		
		// Get endpoint configuration
		key := rm.getEndpointKey(path, method)
		config, exists := rm.endpointConfigs[key]
		if !exists {
			// Try pattern matching
			for endpointKey, endpointConfig := range rm.endpointConfigs {
				if rm.matchesPattern(path, method, endpointKey) {
					config = endpointConfig
					break
				}
			}
		}
		
		// Use company-specific limit if available
		if config != nil && config.CompanyLimit != nil {
			// Create custom key function for company-specific limiting
			companyConfig := *config.CompanyLimit
			companyConfig.KeyFunc = func(c *gin.Context) string {
				companyID, exists := c.Get("company_id")
				if !exists {
					return fmt.Sprintf("rate_limit:ip:%s", c.ClientIP())
				}
				return fmt.Sprintf("rate_limit:company:%v", companyID)
			}
			
			companyLimiter := NewRateLimiter(rm.redis, &companyConfig)
			companyLimiter.Middleware()(c)
			return
		}
		
		// Fallback to regular rate limiting
		rm.Middleware()(c)
	}
}

// GetRateLimitInfo gets rate limit information for a specific endpoint and key
func (rm *RateLimitManager) GetRateLimitInfo(ctx context.Context, path, method, key string) (*RateLimitInfo, error) {
	limiter := rm.getLimiterForEndpoint(path, method)
	return limiter.GetRateLimitInfo(ctx, key)
}

// ResetRateLimit resets rate limit for a specific endpoint and key
func (rm *RateLimitManager) ResetRateLimit(ctx context.Context, path, method, key string) error {
	limiter := rm.getLimiterForEndpoint(path, method)
	return limiter.ResetRateLimit(ctx, key)
}

// GetEndpointConfigs returns all endpoint configurations
func (rm *RateLimitManager) GetEndpointConfigs() map[string]*EndpointConfig {
	return rm.endpointConfigs
}

// UpdateEndpointConfig updates the configuration for an endpoint
func (rm *RateLimitManager) UpdateEndpointConfig(path, method string, config *RateLimitConfig) {
	key := rm.getEndpointKey(path, method)
	if endpointConfig, exists := rm.endpointConfigs[key]; exists {
		endpointConfig.Config = config
		rm.limiters[key] = NewRateLimiter(rm.redis, config)
	}
}

// RemoveEndpointConfig removes rate limiting for an endpoint
func (rm *RateLimitManager) RemoveEndpointConfig(path, method string) {
	key := rm.getEndpointKey(path, method)
	delete(rm.endpointConfigs, key)
	delete(rm.limiters, key)
}

// GetRateLimitStats returns statistics about rate limiting
func (rm *RateLimitManager) GetRateLimitStats(ctx context.Context) (map[string]interface{}, error) {
	stats := make(map[string]interface{})
	
	// Get Redis info
	info, err := rm.redis.Info(ctx, "memory").Result()
	if err != nil {
		return nil, err
	}
	
	stats["redis_memory"] = info
	stats["endpoint_count"] = len(rm.endpointConfigs)
	stats["limiter_count"] = len(rm.limiters)
	
	// Get rate limit keys count
	keys, err := rm.redis.Keys(ctx, "rate_limit:*").Result()
	if err != nil {
		return nil, err
	}
	
	stats["active_rate_limits"] = len(keys)
	
	return stats, nil
}
