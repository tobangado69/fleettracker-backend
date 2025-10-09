package logging

import (
	"bytes"
	"io"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// RequestLoggingMiddleware logs all HTTP requests and responses
func RequestLoggingMiddleware(logger *Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Generate request ID
		requestID := uuid.New().String()
		c.Set("request_id", requestID)

		// Record start time
		start := time.Now()

		// Capture request body (for POST/PUT)
		var requestBody []byte
		if c.Request.Body != nil && (c.Request.Method == "POST" || c.Request.Method == "PUT") {
			requestBody, _ = io.ReadAll(c.Request.Body)
			// Restore body for handler
			c.Request.Body = io.NopCloser(bytes.NewBuffer(requestBody))
		}

		// Create response writer wrapper
		writer := &responseWriter{
			ResponseWriter: c.Writer,
			body:           &bytes.Buffer{},
		}
		c.Writer = writer

		// Process request
		c.Next()

		// Calculate duration
		duration := time.Since(start)

		// Prepare log fields
		fields := map[string]interface{}{
			"request_id":   requestID,
			"method":       c.Request.Method,
			"path":         c.Request.URL.Path,
			"query":        c.Request.URL.RawQuery,
			"status":       c.Writer.Status(),
			"duration_ms":  duration.Milliseconds(),
			"client_ip":    c.ClientIP(),
			"user_agent":   c.Request.UserAgent(),
			"content_type": c.ContentType(),
		}

		// Add authenticated user info if available
		if userID, exists := c.Get("user_id"); exists {
			fields["user_id"] = userID
		}
		if companyID, exists := c.Get("company_id"); exists {
			fields["company_id"] = companyID
		}

		// Add request body for non-GET requests (exclude sensitive data)
		if len(requestBody) > 0 && len(requestBody) < 10240 { // Max 10KB
			if !containsSensitiveData(c.Request.URL.Path) {
				fields["request_body"] = string(requestBody)
			}
		}

		// Add response size
		fields["response_size"] = writer.body.Len()

		// Add error if present
		if len(c.Errors) > 0 {
			fields["errors"] = c.Errors.String()
		}

		// Log based on status code
		if c.Writer.Status() >= 500 {
			logger.WithFields(fields).Error("HTTP Request - Server Error")
		} else if c.Writer.Status() >= 400 {
			logger.WithFields(fields).Warn("HTTP Request - Client Error")
		} else {
			logger.WithFields(fields).Info("HTTP Request")
		}

		// Log slow requests (> 1 second)
		if duration > time.Second {
			logger.WithFields(fields).Warn("Slow HTTP Request detected")
		}
	}
}

// responseWriter wraps gin.ResponseWriter to capture response
type responseWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w *responseWriter) Write(data []byte) (int, error) {
	w.body.Write(data)
	return w.ResponseWriter.Write(data)
}

func (w *responseWriter) WriteString(s string) (int, error) {
	w.body.WriteString(s)
	return w.ResponseWriter.WriteString(s)
}

// containsSensitiveData checks if path contains sensitive endpoints
func containsSensitiveData(path string) bool {
	sensitiveEndpoints := []string{
		"/auth/login",
		"/auth/register",
		"/auth/change-password",
		"/auth/reset-password",
		"/auth/forgot-password",
		"/payment",
	}

	for _, endpoint := range sensitiveEndpoints {
		if contains(path, endpoint) {
			return true
		}
	}
	return false
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && s[:len(substr)] == substr || 
		   len(s) > len(substr) && containsRec(s[1:], substr)
}

func containsRec(s, substr string) bool {
	if len(s) < len(substr) {
		return false
	}
	if s[:len(substr)] == substr {
		return true
	}
	return containsRec(s[1:], substr)
}

// PerformanceLoggingMiddleware logs performance metrics
func PerformanceLoggingMiddleware(logger *Logger, slowThreshold time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		c.Next()

		duration := time.Since(start)

		// Log if request exceeded threshold
		if duration > slowThreshold {
			logger.Warn("Performance: Slow request",
				"method", c.Request.Method,
				"path", c.Request.URL.Path,
				"duration_ms", duration.Milliseconds(),
				"threshold_ms", slowThreshold.Milliseconds(),
				"status", c.Writer.Status(),
			)
		}
	}
}

// ErrorLoggingMiddleware logs all errors
func ErrorLoggingMiddleware(logger *Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// Log errors if present
		if len(c.Errors) > 0 {
			for _, err := range c.Errors {
				logger.Error("Request error",
					"error", err.Err,
					"type", err.Type,
					"meta", err.Meta,
					"method", c.Request.Method,
					"path", c.Request.URL.Path,
				)
			}
		}
	}
}

// RecoveryLoggingMiddleware logs panic recovery
func RecoveryLoggingMiddleware(logger *Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				logger.Error("Panic recovered",
					"error", err,
					"method", c.Request.Method,
					"path", c.Request.URL.Path,
					"client_ip", c.ClientIP(),
				)
				c.AbortWithStatus(500)
			}
		}()
		c.Next()
	}
}

