package middleware

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
)

// CacheMiddleware provides HTTP response caching
type CacheMiddleware struct {
	redis  *redis.Client
	prefix string
}

// NewCacheMiddleware creates a new cache middleware
func NewCacheMiddleware(redis *redis.Client, prefix string) *CacheMiddleware {
	return &CacheMiddleware{
		redis:  redis,
		prefix: prefix,
	}
}

// CachedResponse represents a cached HTTP response
type CachedResponse struct {
	Status      int                 `json:"status"`
	ContentType string              `json:"content_type"`
	Headers     map[string][]string `json:"headers"`
	Body        []byte              `json:"body"`
}

// responseWriter wraps gin.ResponseWriter to capture response body
type responseWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

// Write captures the response body
func (w *responseWriter) Write(data []byte) (int, error) {
	w.body.Write(data)
	return w.ResponseWriter.Write(data)
}

// WriteString captures the response body
func (w *responseWriter) WriteString(s string) (int, error) {
	w.body.WriteString(s)
	return w.ResponseWriter.WriteString(s)
}

// CacheResponse caches GET request responses for the specified duration
func (cm *CacheMiddleware) CacheResponse(duration time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Only cache GET requests
		if c.Request.Method != "GET" {
			c.Next()
			return
		}

		// Generate cache key from request
		cacheKey := cm.generateCacheKey(c)

		// Try to get from cache
		cached, err := cm.redis.Get(c.Request.Context(), cacheKey).Result()
		if err == nil {
			// Cache hit - return cached response
			var response CachedResponse
			if err := json.Unmarshal([]byte(cached), &response); err == nil {
				c.Header("X-Cache-Status", "HIT")
				c.Header("X-Cache-Key", cacheKey)
				
				// Set cached headers
				for key, values := range response.Headers {
					for _, value := range values {
						c.Header(key, value)
					}
				}
				
				c.Data(response.Status, response.ContentType, response.Body)
				c.Abort()
				return
			}
		}

		// Cache miss - continue with request
		c.Header("X-Cache-Status", "MISS")
		c.Header("X-Cache-Key", cacheKey)

		// Create a response writer wrapper to capture the response
		writer := &responseWriter{
			ResponseWriter: c.Writer,
			body:           &bytes.Buffer{},
		}
		c.Writer = writer

		// Process request
		c.Next()

		// Only cache successful responses
		if c.Writer.Status() == 200 {
			// Capture response
			response := CachedResponse{
				Status:      c.Writer.Status(),
				ContentType: c.Writer.Header().Get("Content-Type"),
				Headers:     c.Writer.Header(),
				Body:        writer.body.Bytes(),
			}

			// Store in cache
			data, err := json.Marshal(response)
			if err == nil {
				cm.redis.Set(c.Request.Context(), cacheKey, data, duration)
			}
		}
	}
}

// generateCacheKey generates a cache key from the request
func (cm *CacheMiddleware) generateCacheKey(c *gin.Context) string {
	// Include path, query params, and relevant headers in cache key
	keyData := fmt.Sprintf("%s?%s", c.Request.URL.Path, c.Request.URL.RawQuery)
	
	// Include company_id from context if present (for multi-tenancy)
	if companyID, exists := c.Get("company_id"); exists {
		keyData = fmt.Sprintf("%s|company:%v", keyData, companyID)
	}
	
	// Include user_id from context if present (for user-specific caching)
	if userID, exists := c.Get("user_id"); exists {
		keyData = fmt.Sprintf("%s|user:%v", keyData, userID)
	}

	// Hash the key data to keep keys short
	hash := sha256.Sum256([]byte(keyData))
	hashStr := hex.EncodeToString(hash[:])

	return fmt.Sprintf("%s:response:%s", cm.prefix, hashStr)
}

// CacheShort provides 1-minute caching (for frequently changing data)
func (cm *CacheMiddleware) CacheShort() gin.HandlerFunc {
	return cm.CacheResponse(1 * time.Minute)
}

// CacheMedium provides 5-minute caching (default for most endpoints)
func (cm *CacheMiddleware) CacheMedium() gin.HandlerFunc {
	return cm.CacheResponse(5 * time.Minute)
}

// CacheLong provides 30-minute caching (for relatively static data)
func (cm *CacheMiddleware) CacheLong() gin.HandlerFunc {
	return cm.CacheResponse(30 * time.Minute)
}

// NoCache disables caching for the request
func NoCache() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Cache-Control", "no-cache, no-store, must-revalidate")
		c.Header("Pragma", "no-cache")
		c.Header("Expires", "0")
		c.Next()
	}
}
