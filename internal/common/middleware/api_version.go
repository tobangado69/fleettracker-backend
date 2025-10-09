package middleware

import (
	"github.com/gin-gonic/gin"
)

// APIVersionConfig holds API version configuration
type APIVersionConfig struct {
	Version    string
	Deprecated bool
}

// DefaultAPIVersionConfig returns default API version configuration
func DefaultAPIVersionConfig() *APIVersionConfig {
	return &APIVersionConfig{
		Version:    "1.0.0",
		Deprecated: false,
	}
}

// APIVersionMiddleware adds API version headers to all responses
func APIVersionMiddleware(config *APIVersionConfig) gin.HandlerFunc {
	if config == nil {
		config = DefaultAPIVersionConfig()
	}

	return func(c *gin.Context) {
		// Add API version header
		c.Header("X-API-Version", config.Version)
		
		// Add deprecation warning if deprecated
		if config.Deprecated {
			c.Header("X-API-Deprecated", "true")
			c.Header("X-API-Deprecation-Info", "This API version is deprecated. Please upgrade to the latest version.")
		}
		
		// Add service name
		c.Header("X-Service-Name", "FleetTracker Pro API")
		
		// Add response timestamp
		c.Header("X-Response-Time", c.GetHeader("Date"))
		
		c.Next()
	}
}

