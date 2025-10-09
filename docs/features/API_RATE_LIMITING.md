# API Rate Limiting Documentation

## Overview

The FleetTracker Pro backend now includes a comprehensive API rate limiting system that provides multiple strategies, endpoint-specific configurations, user/company-specific limits, and detailed monitoring capabilities. This system is designed to protect the API from abuse while ensuring fair usage across different user types and endpoints.

## Architecture

### Core Components

1. **Rate Limiter** (`internal/common/ratelimit/rate_limiter.go`) - Core rate limiting logic
2. **Rate Limit Manager** (`internal/common/ratelimit/manager.go`) - Endpoint-specific configuration management
3. **Rate Limit Monitor** (`internal/common/ratelimit/monitoring.go`) - Metrics and monitoring
4. **Rate Limit Middleware** (`internal/common/ratelimit/middleware.go`) - Gin middleware integration

## Rate Limiting Strategies

### 1. Fixed Window

**Description**: Limits requests within fixed time windows (e.g., 100 requests per minute)

**Use Cases**: 
- General API endpoints
- Authentication endpoints
- Simple rate limiting scenarios

**Configuration**:
```go
&RateLimitConfig{
    Strategy: FixedWindow,
    Requests: 100,
    Window:   1 * time.Minute,
}
```

**How it works**:
- Divides time into fixed windows (e.g., 1-minute intervals)
- Counts requests within each window
- Resets counter at the start of each new window
- Simple and predictable behavior

### 2. Sliding Window

**Description**: Limits requests within sliding time windows for more precise control

**Use Cases**:
- Analytics endpoints
- Reporting endpoints
- Scenarios requiring precise rate control

**Configuration**:
```go
&RateLimitConfig{
    Strategy: SlidingWindow,
    Requests: 100,
    Window:   1 * time.Minute,
}
```

**How it works**:
- Uses Redis sorted sets to track request timestamps
- Removes expired entries automatically
- Provides smooth rate limiting without window boundary effects
- More memory intensive but more accurate

### 3. Token Bucket

**Description**: Implements token bucket algorithm with burst capacity and refill rate

**Use Cases**:
- GPS tracking endpoints (high frequency, bursty traffic)
- Real-time data endpoints
- Scenarios with burst capacity requirements

**Configuration**:
```go
&RateLimitConfig{
    Strategy:   TokenBucket,
    Requests:   1000,
    Window:     1 * time.Minute,
    Burst:      100,
    RefillRate: 50, // tokens per second
}
```

**How it works**:
- Maintains a bucket of tokens
- Each request consumes one token
- Tokens are refilled at a constant rate
- Allows bursts up to bucket capacity
- Ideal for handling traffic spikes

### 4. Leaky Bucket

**Description**: Implements leaky bucket algorithm for smooth traffic shaping

**Use Cases**:
- Payment endpoints
- Critical operations requiring smooth traffic
- Scenarios requiring traffic shaping

**Configuration**:
```go
&RateLimitConfig{
    Strategy:   LeakyBucket,
    Requests:   100,
    Window:     1 * time.Minute,
    Burst:      50,
    RefillRate: 10, // requests per second
}
```

**How it works**:
- Maintains a bucket that leaks at a constant rate
- Requests are added to the bucket
- If bucket is full, requests are rejected
- Provides smooth, predictable traffic flow

## Endpoint-Specific Configurations

### Authentication Endpoints

**Login** (`POST /api/v1/auth/login`):
- **Strategy**: Fixed Window
- **Limit**: 5 requests per 5 minutes
- **Purpose**: Prevent brute force attacks

**Registration** (`POST /api/v1/auth/register`):
- **Strategy**: Fixed Window
- **Limit**: 3 requests per 10 minutes
- **Purpose**: Prevent spam registrations

**Forgot Password** (`POST /api/v1/auth/forgot-password`):
- **Strategy**: Fixed Window
- **Limit**: 3 requests per 15 minutes
- **Purpose**: Prevent password reset abuse

### GPS Tracking Endpoints

**GPS Data Submission** (`POST /api/v1/tracking/gps`):
- **Strategy**: Token Bucket
- **Limit**: 1000 requests per minute
- **Burst**: 100 requests
- **Refill Rate**: 50 requests per second
- **Purpose**: Handle high-frequency GPS data with burst capacity

### Vehicle Management Endpoints

**List Vehicles** (`GET /api/v1/vehicles`):
- **Strategy**: Fixed Window
- **Limit**: 200 requests per minute
- **Purpose**: Allow frequent dashboard updates

**Create Vehicle** (`POST /api/v1/vehicles`):
- **Strategy**: Fixed Window
- **Limit**: 20 requests per minute
- **Purpose**: Prevent bulk vehicle creation

### Driver Management Endpoints

**List Drivers** (`GET /api/v1/drivers`):
- **Strategy**: Fixed Window
- **Limit**: 200 requests per minute
- **Purpose**: Allow frequent driver list updates

**Create Driver** (`POST /api/v1/drivers`):
- **Strategy**: Fixed Window
- **Limit**: 20 requests per minute
- **Purpose**: Prevent bulk driver creation

### Analytics Endpoints

**Analytics Data** (`GET /api/v1/analytics`):
- **Strategy**: Sliding Window
- **Limit**: 100 requests per minute
- **Purpose**: Allow frequent analytics queries with precise control

### Payment Endpoints

**Payment Processing** (`POST /api/v1/payments`):
- **Strategy**: Fixed Window
- **Limit**: 10 requests per minute
- **Purpose**: Prevent payment abuse and ensure security

### WebSocket Connections

**WebSocket Connection** (`GET /ws/tracking`):
- **Strategy**: Token Bucket
- **Limit**: 10 connections per minute
- **Burst**: 5 connections
- **Refill Rate**: 2 connections per second
- **Purpose**: Control WebSocket connection rate

## User and Company-Specific Limits

### User-Specific Rate Limiting

Users can have individual rate limits that are more restrictive than global limits:

```go
UserLimit: &RateLimitConfig{
    Strategy: FixedWindow,
    Requests: 50,
    Window:   1 * time.Minute,
}
```

**Use Cases**:
- Premium users with higher limits
- Free tier users with lower limits
- Individual user abuse prevention

### Company-Specific Rate Limiting

Companies can have organization-wide rate limits:

```go
CompanyLimit: &RateLimitConfig{
    Strategy: FixedWindow,
    Requests: 1000,
    Window:   1 * time.Minute,
}
```

**Use Cases**:
- Enterprise customers with higher limits
- Small companies with lower limits
- Company-wide abuse prevention

## Rate Limit Key Generation

### Default Key Strategy

The system uses different key strategies based on available information:

1. **User-based**: `rate_limit:user:{user_id}` (when user is authenticated)
2. **IP-based**: `rate_limit:ip:{ip_address}` (fallback for anonymous users)
3. **Company-based**: `rate_limit:company:{company_id}` (for company-wide limits)

### Custom Key Functions

You can define custom key generation functions:

```go
config.KeyFunc = func(c *gin.Context) string {
    // Custom key generation logic
    return fmt.Sprintf("rate_limit:custom:%s", customIdentifier)
}
```

## Rate Limit Headers

The system automatically adds rate limit headers to all responses:

```
X-RateLimit-Limit: 100
X-RateLimit-Remaining: 95
X-RateLimit-Reset: 1640995200
Retry-After: 30
```

### Header Descriptions

- **X-RateLimit-Limit**: Maximum number of requests allowed
- **X-RateLimit-Remaining**: Number of requests remaining in current window
- **X-RateLimit-Reset**: Unix timestamp when the rate limit resets
- **Retry-After**: Number of seconds to wait before retrying (only on 429 responses)

## Rate Limit Response

When rate limit is exceeded, the system returns a 429 Too Many Requests response:

```json
{
    "error": "Rate limit exceeded",
    "message": "Too many requests. Please try again later.",
    "retry_after": 30,
    "limit": 100,
    "remaining": 0,
    "reset": 1640995200
}
```

## Monitoring and Metrics

### Real-Time Metrics

The system tracks comprehensive metrics:

```go
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
```

### Endpoint Statistics

```go
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
```

### User Statistics

```go
type UserStats struct {
    UserID            string    `json:"user_id"`
    TotalRequests     int64     `json:"total_requests"`
    AllowedRequests   int64     `json:"allowed_requests"`
    BlockedRequests   int64     `json:"blocked_requests"`
    BlockRate         float64   `json:"block_rate"`
    LastRequest       time.Time `json:"last_request"`
}
```

### Company Statistics

```go
type CompanyStats struct {
    CompanyID         string    `json:"company_id"`
    TotalRequests     int64     `json:"total_requests"`
    AllowedRequests   int64     `json:"allowed_requests"`
    BlockedRequests   int64     `json:"blocked_requests"`
    BlockRate         float64   `json:"block_rate"`
    LastRequest       time.Time `json:"last_request"`
}
```

## Management Endpoints

### Admin-Only Endpoints

All rate limit management endpoints require admin role:

#### Get Rate Limit Metrics
```
GET /api/v1/admin/rate-limit/metrics
```

**Response**:
```json
{
    "metrics": {
        "total_requests": 10000,
        "allowed_requests": 9500,
        "blocked_requests": 500,
        "block_rate": 5.0,
        "average_response_time": "50ms",
        "endpoint_stats": {...},
        "user_stats": {...},
        "company_stats": {...}
    },
    "uptime": "2h30m15s"
}
```

#### Get Rate Limit Health Status
```
GET /api/v1/admin/rate-limit/health
```

**Response**:
```json
{
    "status": "healthy",
    "uptime": "2h30m15s",
    "total_requests": 10000,
    "block_rate": 5.0,
    "average_response_time": "50ms",
    "endpoint_count": 15,
    "user_count": 150,
    "company_count": 25
}
```

#### Get Rate Limit Statistics
```
GET /api/v1/admin/rate-limit/stats?limit=10
```

**Response**:
```json
{
    "top_blocked_endpoints": [
        {
            "path": "/api/v1/auth/login",
            "method": "POST",
            "block_rate": 15.5,
            "total_requests": 1000,
            "blocked_requests": 155
        }
    ],
    "top_blocked_users": [
        {
            "user_id": "user123",
            "block_rate": 25.0,
            "total_requests": 200,
            "blocked_requests": 50
        }
    ],
    "top_blocked_companies": [
        {
            "company_id": "company456",
            "block_rate": 10.0,
            "total_requests": 500,
            "blocked_requests": 50
        }
    ]
}
```

#### Get Rate Limit Configuration
```
GET /api/v1/admin/rate-limit/config
```

**Response**:
```json
{
    "endpoint_configs": {
        "POST:/api/v1/auth/login": {
            "path": "/api/v1/auth/login",
            "method": "POST",
            "config": {
                "strategy": "fixed_window",
                "requests": 5,
                "window": "5m"
            }
        }
    }
}
```

#### Update Rate Limit Configuration
```
POST /api/v1/admin/rate-limit/config
```

**Request Body**:
```json
{
    "path": "/api/v1/custom/endpoint",
    "method": "POST",
    "config": {
        "strategy": "fixed_window",
        "requests": 100,
        "window": "1m"
    }
}
```

#### Update Specific Endpoint Configuration
```
PUT /api/v1/admin/rate-limit/config/{path}/{method}
```

**Request Body**:
```json
{
    "strategy": "token_bucket",
    "requests": 1000,
    "window": "1m",
    "burst": 100,
    "refill_rate": 50
}
```

#### Remove Endpoint Configuration
```
DELETE /api/v1/admin/rate-limit/config/{path}/{method}
```

#### Reset Rate Limit
```
POST /api/v1/admin/rate-limit/reset?path=/api/v1/auth/login&method=POST&key=user123
```

## Configuration

### Environment Variables

```bash
# Redis configuration for rate limiting
REDIS_URL=redis://localhost:6379

# Default rate limiting configuration
DEFAULT_RATE_LIMIT_REQUESTS=100
DEFAULT_RATE_LIMIT_WINDOW=1m
DEFAULT_RATE_LIMIT_STRATEGY=fixed_window
```

### Default Configuration

```go
defaultConfig := &RateLimitConfig{
    Strategy: FixedWindow,
    Requests: 100,
    Window:   1 * time.Minute,
}
```

## Integration

### Basic Integration

```go
// Initialize rate limiting system
rateLimitManager := ratelimit.NewRateLimitManager(redisClient, nil)
rateLimitMonitor := ratelimit.NewRateLimitMonitor(redisClient)

// Apply to all routes
r.Use(ratelimit.MonitoredRateLimitMiddleware(rateLimitManager, rateLimitMonitor))
```

### User-Specific Rate Limiting

```go
// Apply user-specific rate limiting
r.Use(ratelimit.MonitoredUserRateLimitMiddleware(rateLimitManager, rateLimitMonitor))
```

### Company-Specific Rate Limiting

```go
// Apply company-specific rate limiting
r.Use(ratelimit.MonitoredCompanyRateLimitMiddleware(rateLimitManager, rateLimitMonitor))
```

### Custom Configuration

```go
// Create custom rate limiter
customConfig := &RateLimitConfig{
    Strategy: TokenBucket,
    Requests: 1000,
    Window:   1 * time.Minute,
    Burst:    100,
    RefillRate: 50,
    KeyFunc: func(c *gin.Context) string {
        return fmt.Sprintf("custom:%s", c.ClientIP())
    },
}

customLimiter := ratelimit.NewRateLimiter(redisClient, customConfig)
r.Use(customLimiter.Middleware())
```

## Best Practices

### 1. Strategy Selection

- **Fixed Window**: Use for simple, predictable rate limiting
- **Sliding Window**: Use when precise control is needed
- **Token Bucket**: Use for bursty traffic with burst capacity
- **Leaky Bucket**: Use for smooth traffic shaping

### 2. Limit Configuration

- **Authentication endpoints**: Low limits (3-10 requests per 5-15 minutes)
- **GPS tracking**: High limits with burst capacity (1000+ requests per minute)
- **Analytics**: Moderate limits (100-200 requests per minute)
- **Payment endpoints**: Low limits for security (5-10 requests per minute)

### 3. Monitoring

- Monitor block rates regularly
- Set up alerts for high block rates (>20%)
- Track top blocked endpoints and users
- Review and adjust limits based on usage patterns

### 4. Error Handling

- Always provide clear error messages
- Include retry-after information
- Log rate limit violations for analysis
- Implement graceful degradation

### 5. Performance

- Use Redis for distributed rate limiting
- Implement efficient key generation
- Monitor Redis memory usage
- Set appropriate TTL values for rate limit keys

## Troubleshooting

### Common Issues

1. **High Block Rates**
   - Review endpoint limits
   - Check for legitimate high-frequency usage
   - Consider increasing limits for specific endpoints

2. **Redis Memory Usage**
   - Monitor rate limit key expiration
   - Implement key cleanup strategies
   - Consider using Redis memory optimization

3. **Performance Issues**
   - Check Redis connection health
   - Monitor rate limiting overhead
   - Optimize key generation functions

### Debugging

```go
// Get rate limit information for debugging
info, err := rateLimitManager.GetRateLimitInfo(ctx, path, method, key)
if err != nil {
    log.Printf("Rate limit check error: %v", err)
}

// Reset rate limit for testing
err = rateLimitManager.ResetRateLimit(ctx, path, method, key)
if err != nil {
    log.Printf("Rate limit reset error: %v", err)
}
```

## Security Considerations

### 1. Key Generation

- Use secure, unpredictable keys
- Include user/company context in keys
- Avoid IP-only rate limiting for authenticated users

### 2. Limit Bypass Prevention

- Implement multiple rate limiting layers
- Use different strategies for different attack vectors
- Monitor for unusual patterns

### 3. Data Protection

- Don't log sensitive information in rate limit keys
- Implement proper access controls for management endpoints
- Use HTTPS for all rate limit management

### 4. Abuse Prevention

- Implement progressive rate limiting
- Use different limits for different user types
- Monitor and alert on suspicious activity

## Performance Impact

### Expected Overhead

- **Redis Operations**: 1-2ms per request
- **Memory Usage**: ~1KB per active rate limit key
- **CPU Usage**: Minimal impact on request processing

### Optimization Tips

1. **Use Connection Pooling**: Reuse Redis connections
2. **Batch Operations**: Use Redis pipelines when possible
3. **Key Expiration**: Set appropriate TTL values
4. **Monitoring**: Track performance metrics

## Future Enhancements

### Planned Features

1. **Dynamic Rate Limiting**: Adjust limits based on system load
2. **Machine Learning**: Intelligent rate limit adjustment
3. **Geographic Rate Limiting**: Different limits by region
4. **API Key Rate Limiting**: Per-API-key limits
5. **Rate Limit Analytics**: Advanced reporting and insights
6. **Webhook Integration**: Real-time rate limit notifications
7. **Rate Limit Templates**: Predefined configurations for common scenarios

The FleetTracker Pro API rate limiting system provides comprehensive protection against abuse while maintaining excellent performance and flexibility for different use cases and user types.
