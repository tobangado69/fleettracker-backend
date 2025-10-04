package middleware

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/time/rate"
	"gorm.io/gorm"

	"github.com/tobangado69/fleettracker-pro/backend/pkg/models"
)

// Claims represents JWT claims
type Claims struct {
	UserID    string `json:"user_id"`
	CompanyID string `json:"company_id"`
	Role      string `json:"role"`
	Username  string `json:"username"`
	jwt.RegisteredClaims
}

// AuthRequired middleware validates JWT token
func AuthRequired(jwtSecret string, db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "Authorization header required",
				"message": "Please provide a valid JWT token",
			})
			c.Abort()
			return
		}

		// Check if it starts with "Bearer "
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "Invalid authorization header format",
				"message": "Authorization header must start with 'Bearer '",
			})
			c.Abort()
			return
		}

		// Parse and validate token
		token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
			return []byte(jwtSecret), nil
		})

		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "Invalid token",
				"message": "Token validation failed",
			})
			c.Abort()
			return
		}

		claims, ok := token.Claims.(*Claims)
		if !ok || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "Invalid token claims",
				"message": "Token claims validation failed",
			})
			c.Abort()
			return
		}

		// Check if token is expired
		if claims.ExpiresAt != nil && claims.ExpiresAt.Before(time.Now()) {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "Token expired",
				"message": "Please refresh your token",
			})
			c.Abort()
			return
		}

		// Verify user still exists and is active
		var user models.User
		if err := db.Where("id = ? AND is_active = true", claims.UserID).First(&user).Error; err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "User not found or inactive",
				"message": "User account is no longer valid",
			})
			c.Abort()
			return
		}

		// Set user information in context
		c.Set("user_id", claims.UserID)
		c.Set("company_id", claims.CompanyID)
		c.Set("user_role", claims.Role)
		c.Set("username", claims.Username)
		c.Set("user", user)

		c.Next()
	}
}

// RoleRequired middleware checks if user has required role
func RoleRequired(requiredRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, exists := c.Get("user_role")
		if !exists {
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "Role information not found",
				"message": "User role could not be determined",
			})
			c.Abort()
			return
		}

		role := userRole.(string)
		hasRole := false
		for _, requiredRole := range requiredRoles {
			if role == requiredRole {
				hasRole = true
				break
			}
		}

		if !hasRole {
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "Insufficient permissions",
				"message": "This action requires one of the following roles: " + strings.Join(requiredRoles, ", "),
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// CompanyAccess middleware ensures user can only access their company data
func CompanyAccess() gin.HandlerFunc {
	return func(c *gin.Context) {
		companyID, exists := c.Get("company_id")
		if !exists {
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "Company information not found",
				"message": "User company could not be determined",
			})
			c.Abort()
			return
		}

		// Set company filter for database queries
		c.Set("filter_company_id", companyID)
		c.Next()
	}
}

// SecurityHeaders middleware adds security headers
func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		c.Header("Content-Security-Policy", "default-src 'self'")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Header("Permissions-Policy", "geolocation=(), microphone=(), camera=()")
		
		c.Next()
	}
}

// RateLimit middleware implements rate limiting
func RateLimit(requestsPerMinute int) gin.HandlerFunc {
	limiter := rate.NewLimiter(rate.Every(time.Minute/time.Duration(requestsPerMinute)), requestsPerMinute)
	
	return func(c *gin.Context) {
		if !limiter.Allow() {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":   "Too many requests",
				"message": "Rate limit exceeded. Please try again later.",
				"retry_after": 60,
			})
			c.Abort()
			return
		}
		c.Next()
	}
}

// GPSRateLimit middleware implements rate limiting specifically for GPS data
func GPSRateLimit(requestsPerMinute int) gin.HandlerFunc {
	limiter := rate.NewLimiter(rate.Every(time.Minute/time.Duration(requestsPerMinute)), requestsPerMinute)
	
	return func(c *gin.Context) {
		if !limiter.Allow() {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":   "GPS rate limit exceeded",
				"message": "Too many GPS updates. Please reduce frequency.",
				"retry_after": 30,
			})
			c.Abort()
			return
		}
		c.Next()
	}
}

// IndonesianDataResidency middleware logs requests for compliance
func IndonesianDataResidency() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Log request for Indonesian compliance audit
		// In production, this would send to a secure audit log
		// log.Printf("Request from IP: %s, Path: %s, User: %s", 
		//     c.ClientIP(), c.Request.URL.Path, c.GetString("user_id"))
		
		c.Next()
	}
}

// RequestLogger middleware logs requests with structured logging
func RequestLogger() gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		return fmt.Sprintf("%s - [%s] \"%s %s %s %d %s \"%s\" %s\"\n",
			param.ClientIP,
			param.TimeStamp.Format("02/Jan/2006:15:04:05 -0700"),
			param.Method,
			param.Path,
			param.Request.Proto,
			param.StatusCode,
			param.Latency,
			param.Request.UserAgent(),
			param.ErrorMessage,
		)
	})
}

// CORSMiddleware provides CORS configuration for Indonesian domains
func CORSMiddleware(allowedOrigins []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		
		// Check if origin is allowed
		allowed := false
		for _, allowedOrigin := range allowedOrigins {
			if origin == allowedOrigin || allowedOrigin == "*" {
				allowed = true
				break
			}
		}

		if allowed {
			c.Header("Access-Control-Allow-Origin", origin)
		}

		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Header("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

// ValidateIndonesianPhone middleware validates Indonesian phone numbers
func ValidateIndonesianPhone() gin.HandlerFunc {
	return func(c *gin.Context) {
		// This would validate Indonesian phone number format
		// For now, just pass through
		c.Next()
	}
}

// ValidateLicensePlate middleware validates Indonesian license plate format
func ValidateLicensePlate() gin.HandlerFunc {
	return func(c *gin.Context) {
		// This would validate Indonesian license plate format
		// For now, just pass through
		c.Next()
	}
}

// AuditLog middleware logs all requests for compliance
func AuditLog() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		
		c.Next()
		
		// Log request details for audit
		duration := time.Since(start)
		
		// In production, this would be sent to a secure audit log system
		auditData := map[string]interface{}{
			"timestamp":    start.UTC().Format(time.RFC3339),
			"method":       c.Request.Method,
			"path":         c.Request.URL.Path,
			"query":        c.Request.URL.RawQuery,
			"status":       c.Writer.Status(),
			"duration":     duration.String(),
			"user_id":      c.GetString("user_id"),
			"company_id":   c.GetString("company_id"),
			"ip_address":   c.ClientIP(),
			"user_agent":   c.Request.UserAgent(),
		}
		
		// TODO: Send to audit log system
		_ = auditData
	}
}
