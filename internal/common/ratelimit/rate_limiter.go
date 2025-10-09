package ratelimit

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
)

// RateLimitStrategy defines the rate limiting strategy
type RateLimitStrategy string

const (
	// FixedWindow limits requests within fixed time windows
	FixedWindow RateLimitStrategy = "fixed_window"
	// SlidingWindow limits requests within sliding time windows
	SlidingWindow RateLimitStrategy = "sliding_window"
	// TokenBucket implements token bucket algorithm
	TokenBucket RateLimitStrategy = "token_bucket"
	// LeakyBucket implements leaky bucket algorithm
	LeakyBucket RateLimitStrategy = "leaky_bucket"
)

// RateLimitConfig holds rate limiting configuration
type RateLimitConfig struct {
	Strategy    RateLimitStrategy `json:"strategy"`
	Requests    int               `json:"requests"`    // Number of requests allowed
	Window      time.Duration     `json:"window"`      // Time window for the limit
	Burst       int               `json:"burst"`       // Burst capacity (for token bucket)
	RefillRate  int               `json:"refill_rate"` // Refill rate per second (for token bucket)
	KeyFunc     KeyFunc           `json:"-"`           // Function to generate rate limit key
	SkipFunc    SkipFunc          `json:"-"`           // Function to skip rate limiting
	OnLimitFunc OnLimitFunc       `json:"-"`           // Function called when limit is exceeded
}

// KeyFunc generates a key for rate limiting
type KeyFunc func(c *gin.Context) string

// SkipFunc determines if rate limiting should be skipped
type SkipFunc func(c *gin.Context) bool

// OnLimitFunc is called when rate limit is exceeded
type OnLimitFunc func(c *gin.Context, info *RateLimitInfo)

// RateLimitInfo contains rate limit information
type RateLimitInfo struct {
	Limit     int           `json:"limit"`
	Remaining int           `json:"remaining"`
	Reset     time.Time     `json:"reset"`
	RetryAfter time.Duration `json:"retry_after"`
}

// RateLimiter provides rate limiting functionality
type RateLimiter struct {
	redis  *redis.Client
	config *RateLimitConfig
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(redis *redis.Client, config *RateLimitConfig) *RateLimiter {
	if config.KeyFunc == nil {
		config.KeyFunc = DefaultKeyFunc
	}
	if config.SkipFunc == nil {
		config.SkipFunc = DefaultSkipFunc
	}
	if config.OnLimitFunc == nil {
		config.OnLimitFunc = DefaultOnLimitFunc
	}
	
	return &RateLimiter{
		redis:  redis,
		config: config,
	}
}

// DefaultKeyFunc generates a default rate limit key based on IP and user
func DefaultKeyFunc(c *gin.Context) string {
	// Try to get user ID from context (set by auth middleware)
	userID, exists := c.Get("user_id")
	if exists {
		return fmt.Sprintf("rate_limit:user:%v", userID)
	}
	
	// Fallback to IP address
	ip := c.ClientIP()
	return fmt.Sprintf("rate_limit:ip:%s", ip)
}

// DefaultSkipFunc determines if rate limiting should be skipped
func DefaultSkipFunc(c *gin.Context) bool {
	// Skip rate limiting for health checks
	if strings.HasPrefix(c.Request.URL.Path, "/health") {
		return true
	}
	
	// Skip rate limiting for metrics endpoints
	if strings.HasPrefix(c.Request.URL.Path, "/metrics") {
		return true
	}
	
	return false
}

// DefaultOnLimitFunc handles rate limit exceeded
func DefaultOnLimitFunc(c *gin.Context, info *RateLimitInfo) {
	c.Header("X-RateLimit-Limit", strconv.Itoa(info.Limit))
	c.Header("X-RateLimit-Remaining", strconv.Itoa(info.Remaining))
	c.Header("X-RateLimit-Reset", strconv.FormatInt(info.Reset.Unix(), 10))
	c.Header("Retry-After", strconv.Itoa(int(info.RetryAfter.Seconds())))
	
	c.JSON(http.StatusTooManyRequests, gin.H{
		"error": "Rate limit exceeded",
		"message": "Too many requests. Please try again later.",
		"retry_after": info.RetryAfter.Seconds(),
		"limit": info.Limit,
		"remaining": info.Remaining,
		"reset": info.Reset.Unix(),
	})
}

// Middleware returns a Gin middleware for rate limiting
func (rl *RateLimiter) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip rate limiting if configured
		if rl.config.SkipFunc(c) {
			c.Next()
			return
		}
		
		// Generate rate limit key
		key := rl.config.KeyFunc(c)
		
		// Check rate limit based on strategy
		allowed, info, err := rl.checkRateLimit(c.Request.Context(), key)
		if err != nil {
			// Log error but don't block request
			fmt.Printf("Rate limit check error: %v\n", err)
			c.Next()
			return
		}
		
		// Set rate limit headers
		c.Header("X-RateLimit-Limit", strconv.Itoa(info.Limit))
		c.Header("X-RateLimit-Remaining", strconv.Itoa(info.Remaining))
		c.Header("X-RateLimit-Reset", strconv.FormatInt(info.Reset.Unix(), 10))
		
		if !allowed {
			rl.config.OnLimitFunc(c, info)
			c.Abort()
			return
		}
		
		c.Next()
	}
}

// checkRateLimit checks if the request is within rate limits
func (rl *RateLimiter) checkRateLimit(ctx context.Context, key string) (bool, *RateLimitInfo, error) {
	switch rl.config.Strategy {
	case FixedWindow:
		return rl.checkFixedWindow(ctx, key)
	case SlidingWindow:
		return rl.checkSlidingWindow(ctx, key)
	case TokenBucket:
		return rl.checkTokenBucket(ctx, key)
	case LeakyBucket:
		return rl.checkLeakyBucket(ctx, key)
	default:
		return rl.checkFixedWindow(ctx, key)
	}
}

// checkFixedWindow implements fixed window rate limiting
func (rl *RateLimiter) checkFixedWindow(ctx context.Context, key string) (bool, *RateLimitInfo, error) {
	now := time.Now()
	windowStart := now.Truncate(rl.config.Window)
	windowKey := fmt.Sprintf("%s:window:%d", key, windowStart.Unix())
	
	// Get current count
	count, err := rl.redis.Get(ctx, windowKey).Int()
	if err != nil && err != redis.Nil {
		return false, nil, err
	}
	
	// Check if limit exceeded
	if count >= rl.config.Requests {
		nextWindow := windowStart.Add(rl.config.Window)
		info := &RateLimitInfo{
			Limit:      rl.config.Requests,
			Remaining:  0,
			Reset:      nextWindow,
			RetryAfter: nextWindow.Sub(now),
		}
		return false, info, nil
	}
	
	// Increment counter
	pipe := rl.redis.Pipeline()
	pipe.Incr(ctx, windowKey)
	pipe.Expire(ctx, windowKey, rl.config.Window)
	_, err = pipe.Exec(ctx)
	if err != nil {
		return false, nil, err
	}
	
	// Get updated count
	count, err = rl.redis.Get(ctx, windowKey).Int()
	if err != nil {
		return false, nil, err
	}
	
	nextWindow := windowStart.Add(rl.config.Window)
	info := &RateLimitInfo{
		Limit:      rl.config.Requests,
		Remaining:  rl.config.Requests - count,
		Reset:      nextWindow,
		RetryAfter: 0,
	}
	
	return true, info, nil
}

// checkSlidingWindow implements sliding window rate limiting
func (rl *RateLimiter) checkSlidingWindow(ctx context.Context, key string) (bool, *RateLimitInfo, error) {
	now := time.Now()
	windowStart := now.Add(-rl.config.Window)
	
	// Use Redis sorted set for sliding window
	zsetKey := fmt.Sprintf("%s:sliding", key)
	
	// Remove expired entries
	rl.redis.ZRemRangeByScore(ctx, zsetKey, "0", strconv.FormatInt(windowStart.UnixNano(), 10))
	
	// Count current requests
	count, err := rl.redis.ZCard(ctx, zsetKey).Result()
	if err != nil {
		return false, nil, err
	}
	
	// Check if limit exceeded
	if int(count) >= rl.config.Requests {
		// Get oldest request to calculate reset time
		oldest, err := rl.redis.ZRangeWithScores(ctx, zsetKey, 0, 0).Result()
		if err != nil {
			return false, nil, err
		}
		
		var resetTime time.Time
		if len(oldest) > 0 {
			resetTime = time.Unix(0, int64(oldest[0].Score)).Add(rl.config.Window)
		} else {
			resetTime = now.Add(rl.config.Window)
		}
		
		info := &RateLimitInfo{
			Limit:      rl.config.Requests,
			Remaining:  0,
			Reset:      resetTime,
			RetryAfter: resetTime.Sub(now),
		}
		return false, info, nil
	}
	
	// Add current request
	rl.redis.ZAdd(ctx, zsetKey, &redis.Z{
		Score:  float64(now.UnixNano()),
		Member: now.UnixNano(),
	})
	
	// Set expiration
	rl.redis.Expire(ctx, zsetKey, rl.config.Window)
	
	info := &RateLimitInfo{
		Limit:      rl.config.Requests,
		Remaining:  rl.config.Requests - int(count) - 1,
		Reset:      now.Add(rl.config.Window),
		RetryAfter: 0,
	}
	
	return true, info, nil
}

// checkTokenBucket implements token bucket rate limiting
func (rl *RateLimiter) checkTokenBucket(ctx context.Context, key string) (bool, *RateLimitInfo, error) {
	now := time.Now()
	bucketKey := fmt.Sprintf("%s:bucket", key)
	
	// Get bucket state
	pipe := rl.redis.Pipeline()
	tokensCmd := pipe.HGet(ctx, bucketKey, "tokens")
	lastRefillCmd := pipe.HGet(ctx, bucketKey, "last_refill")
	_, err := pipe.Exec(ctx)
	
	if err != nil && err != redis.Nil {
		return false, nil, err
	}
	
	tokens := rl.config.Burst
	lastRefill := now
	
	if tokensCmd.Val() != "" {
		tokens, _ = strconv.Atoi(tokensCmd.Val())
	}
	if lastRefillCmd.Val() != "" {
		lastRefillUnix, _ := strconv.ParseInt(lastRefillCmd.Val(), 10, 64)
		lastRefill = time.Unix(lastRefillUnix, 0)
	}
	
	// Calculate tokens to add based on time elapsed
	timeElapsed := now.Sub(lastRefill)
	tokensToAdd := int(timeElapsed.Seconds()) * rl.config.RefillRate
	
	if tokensToAdd > 0 {
		tokens += tokensToAdd
		if tokens > rl.config.Burst {
			tokens = rl.config.Burst
		}
		lastRefill = now
	}
	
	// Check if tokens available
	if tokens <= 0 {
		// Calculate when next token will be available
		nextTokenTime := lastRefill.Add(time.Duration(1.0/float64(rl.config.RefillRate)) * time.Second)
		info := &RateLimitInfo{
			Limit:      rl.config.Burst,
			Remaining:  0,
			Reset:      nextTokenTime,
			RetryAfter: nextTokenTime.Sub(now),
		}
		return false, info, nil
	}
	
	// Consume token
	tokens--
	
	// Update bucket state
	pipe = rl.redis.Pipeline()
	pipe.HSet(ctx, bucketKey, "tokens", tokens)
	pipe.HSet(ctx, bucketKey, "last_refill", lastRefill.Unix())
	pipe.Expire(ctx, bucketKey, time.Hour) // Expire after 1 hour of inactivity
	_, err = pipe.Exec(ctx)
	
	if err != nil {
		return false, nil, err
	}
	
	info := &RateLimitInfo{
		Limit:      rl.config.Burst,
		Remaining:  tokens,
		Reset:      now.Add(time.Duration(1.0/float64(rl.config.RefillRate)) * time.Second),
		RetryAfter: 0,
	}
	
	return true, info, nil
}

// checkLeakyBucket implements leaky bucket rate limiting
func (rl *RateLimiter) checkLeakyBucket(ctx context.Context, key string) (bool, *RateLimitInfo, error) {
	now := time.Now()
	bucketKey := fmt.Sprintf("%s:leaky", key)
	
	// Get bucket state
	pipe := rl.redis.Pipeline()
	levelCmd := pipe.HGet(ctx, bucketKey, "level")
	lastLeakCmd := pipe.HGet(ctx, bucketKey, "last_leak")
	_, err := pipe.Exec(ctx)
	
	if err != nil && err != redis.Nil {
		return false, nil, err
	}
	
	level := 0
	lastLeak := now
	
	if levelCmd.Val() != "" {
		level, _ = strconv.Atoi(levelCmd.Val())
	}
	if lastLeakCmd.Val() != "" {
		lastLeakUnix, _ := strconv.ParseInt(lastLeakCmd.Val(), 10, 64)
		lastLeak = time.Unix(lastLeakUnix, 0)
	}
	
	// Calculate leaked amount based on time elapsed
	timeElapsed := now.Sub(lastLeak)
	leaked := int(timeElapsed.Seconds()) * rl.config.RefillRate
	
	if leaked > 0 {
		level -= leaked
		if level < 0 {
			level = 0
		}
		lastLeak = now
	}
	
	// Check if bucket is full
	if level >= rl.config.Burst {
		// Calculate when bucket will have space
		nextSpaceTime := lastLeak.Add(time.Duration(1.0/float64(rl.config.RefillRate)) * time.Second)
		info := &RateLimitInfo{
			Limit:      rl.config.Burst,
			Remaining:  0,
			Reset:      nextSpaceTime,
			RetryAfter: nextSpaceTime.Sub(now),
		}
		return false, info, nil
	}
	
	// Add request to bucket
	level++
	
	// Update bucket state
	pipe = rl.redis.Pipeline()
	pipe.HSet(ctx, bucketKey, "level", level)
	pipe.HSet(ctx, bucketKey, "last_leak", lastLeak.Unix())
	pipe.Expire(ctx, bucketKey, time.Hour) // Expire after 1 hour of inactivity
	_, err = pipe.Exec(ctx)
	
	if err != nil {
		return false, nil, err
	}
	
	info := &RateLimitInfo{
		Limit:      rl.config.Burst,
		Remaining:  rl.config.Burst - level,
		Reset:      now.Add(time.Duration(1.0/float64(rl.config.RefillRate)) * time.Second),
		RetryAfter: 0,
	}
	
	return true, info, nil
}

// GetRateLimitInfo gets current rate limit information for a key
func (rl *RateLimiter) GetRateLimitInfo(ctx context.Context, key string) (*RateLimitInfo, error) {
	_, info, err := rl.checkRateLimit(ctx, key)
	return info, err
}

// ResetRateLimit resets the rate limit for a key
func (rl *RateLimiter) ResetRateLimit(ctx context.Context, key string) error {
	patterns := []string{
		fmt.Sprintf("%s:window:*", key),
		fmt.Sprintf("%s:sliding", key),
		fmt.Sprintf("%s:bucket", key),
		fmt.Sprintf("%s:leaky", key),
	}
	
	for _, pattern := range patterns {
		keys, err := rl.redis.Keys(ctx, pattern).Result()
		if err != nil {
			continue
		}
		
		if len(keys) > 0 {
			rl.redis.Del(ctx, keys...)
		}
	}
	
	return nil
}
