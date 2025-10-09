package ratelimit

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// MonitoredRateLimitMiddleware creates a middleware that combines rate limiting with monitoring
func MonitoredRateLimitMiddleware(manager *RateLimitManager, monitor *RateLimitMonitor) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		
		// Get user and company information
		userID := ""
		companyID := ""
		
		if user, exists := c.Get("user_id"); exists {
			userID = user.(string)
		}
		if company, exists := c.Get("company_id"); exists {
			companyID = company.(string)
		}
		
		// Apply rate limiting
		manager.Middleware()(c)
		
		// Calculate response time
		responseTime := time.Since(start)
		
		// Determine if request was allowed (not aborted)
		allowed := !c.IsAborted()
		
		// Record metrics
		monitor.RecordRequest(
			c.Request.Context(),
			c.Request.URL.Path,
			c.Request.Method,
			userID,
			companyID,
			allowed,
			responseTime,
		)
	}
}

// MonitoredUserRateLimitMiddleware creates a middleware for user-specific rate limiting with monitoring
func MonitoredUserRateLimitMiddleware(manager *RateLimitManager, monitor *RateLimitMonitor) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		
		// Get user and company information
		userID := ""
		companyID := ""
		
		if user, exists := c.Get("user_id"); exists {
			userID = user.(string)
		}
		if company, exists := c.Get("company_id"); exists {
			companyID = company.(string)
		}
		
		// Apply user-specific rate limiting
		manager.UserSpecificMiddleware()(c)
		
		// Calculate response time
		responseTime := time.Since(start)
		
		// Determine if request was allowed (not aborted)
		allowed := !c.IsAborted()
		
		// Record metrics
		monitor.RecordRequest(
			c.Request.Context(),
			c.Request.URL.Path,
			c.Request.Method,
			userID,
			companyID,
			allowed,
			responseTime,
		)
	}
}

// MonitoredCompanyRateLimitMiddleware creates a middleware for company-specific rate limiting with monitoring
func MonitoredCompanyRateLimitMiddleware(manager *RateLimitManager, monitor *RateLimitMonitor) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		
		// Get user and company information
		userID := ""
		companyID := ""
		
		if user, exists := c.Get("user_id"); exists {
			userID = user.(string)
		}
		if company, exists := c.Get("company_id"); exists {
			companyID = company.(string)
		}
		
		// Apply company-specific rate limiting
		manager.CompanySpecificMiddleware()(c)
		
		// Calculate response time
		responseTime := time.Since(start)
		
		// Determine if request was allowed (not aborted)
		allowed := !c.IsAborted()
		
		// Record metrics
		monitor.RecordRequest(
			c.Request.Context(),
			c.Request.URL.Path,
			c.Request.Method,
			userID,
			companyID,
			allowed,
			responseTime,
		)
	}
}

// RateLimitMetricsHandler returns a handler for rate limit metrics
func RateLimitMetricsHandler(monitor *RateLimitMonitor) gin.HandlerFunc {
	return func(c *gin.Context) {
		metrics := monitor.GetMetrics()
		c.JSON(200, gin.H{
			"metrics": metrics,
			"uptime": monitor.GetUptime().String(),
		})
	}
}

// RateLimitHealthHandler returns a handler for rate limit health status
func RateLimitHealthHandler(monitor *RateLimitMonitor) gin.HandlerFunc {
	return func(c *gin.Context) {
		status := monitor.GetHealthStatus()
		c.JSON(200, status)
	}
}

// RateLimitStatsHandler returns a handler for rate limit statistics
func RateLimitStatsHandler(monitor *RateLimitMonitor) gin.HandlerFunc {
	return func(c *gin.Context) {
		limit := c.DefaultQuery("limit", "10")
		limitInt := 10
		if l, err := strconv.Atoi(limit); err == nil {
			limitInt = l
		}
		
		stats := gin.H{
			"top_blocked_endpoints": monitor.GetTopBlockedEndpoints(limitInt),
			"top_blocked_users":     monitor.GetTopBlockedUsers(limitInt),
			"top_blocked_companies": monitor.GetTopBlockedCompanies(limitInt),
		}
		
		c.JSON(200, stats)
	}
}

// RateLimitConfigHandler returns a handler for rate limit configuration management
func RateLimitConfigHandler(manager *RateLimitManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		switch c.Request.Method {
		case "GET":
			// Get all endpoint configurations
			configs := manager.GetEndpointConfigs()
			c.JSON(200, gin.H{
				"endpoint_configs": configs,
			})
			
		case "POST":
			// Update endpoint configuration
			var config EndpointConfig
			if err := c.ShouldBindJSON(&config); err != nil {
				c.JSON(400, gin.H{"error": err.Error()})
				return
			}
			
			manager.AddEndpointConfig(&config)
			c.JSON(200, gin.H{"message": "Configuration updated"})
			
		case "PUT":
			// Update specific endpoint configuration
			path := c.Param("path")
			method := c.Param("method")
			
			var config RateLimitConfig
			if err := c.ShouldBindJSON(&config); err != nil {
				c.JSON(400, gin.H{"error": err.Error()})
				return
			}
			
			manager.UpdateEndpointConfig(path, method, &config)
			c.JSON(200, gin.H{"message": "Configuration updated"})
			
		case "DELETE":
			// Remove endpoint configuration
			path := c.Param("path")
			method := c.Param("method")
			
			manager.RemoveEndpointConfig(path, method)
			c.JSON(200, gin.H{"message": "Configuration removed"})
		}
	}
}

// RateLimitResetHandler returns a handler for resetting rate limits
func RateLimitResetHandler(manager *RateLimitManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		path := c.Query("path")
		method := c.Query("method")
		key := c.Query("key")
		
		if path == "" || method == "" || key == "" {
			c.JSON(400, gin.H{"error": "path, method, and key are required"})
			return
		}
		
		err := manager.ResetRateLimit(c.Request.Context(), path, method, key)
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
		
		c.JSON(200, gin.H{"message": "Rate limit reset successfully"})
	}
}
