# Comprehensive Logging System

Production-ready structured logging system with request tracking, performance monitoring, audit trails, and security event logging.

## Features

### ✅ Structured Logging
- JSON and text format support
- Contextual logging with request/user/company IDs
- Log levels: DEBUG, INFO, WARN, ERROR
- Source file and line numbers
- Customizable output destinations

### ✅ Request/Response Logging
- Automatic HTTP request logging
- Response size tracking
- Duration measurement
- Client IP and user agent capture
- Request ID generation
- Sensitive data filtering

### ✅ Performance Monitoring
- Slow query detection (>100ms by default)
- HTTP request performance tracking
- Operation timing
- Memory and goroutine monitoring
- Customizable thresholds

### ✅ Audit Trail
- Create/Update/Delete tracking
- User action logging
- Resource access logging
- Change detection (old vs new values)
- Payment event logging
- Driver event logging
- Geofence violation logging
- Database persistence

### ✅ Security Logging
- Authentication events (login, logout, register)
- Failed login attempts
- Authorization failures
- Security violations
- IP address tracking

## Quick Start

### 1. Initialize Logger

```go
import "github.com/tobangado69/fleettracker-pro/backend/internal/common/logging"

// Create logger with JSON format
loggerConfig := &logging.LoggerConfig{
    Level:      logging.LevelInfo,
    Format:     "json",
    Output:     os.Stdout,
    AddSource:  true,
    TimeFormat: time.RFC3339,
}

logger := logging.NewLogger(loggerConfig)
logging.InitDefaultLogger(loggerConfig)
```

### 2. Add Middleware

```go
// Request logging
r.Use(logging.RequestLoggingMiddleware(logger))

// Performance monitoring (1 second threshold)
r.Use(logging.PerformanceLoggingMiddleware(logger, 1*time.Second))

// Error logging
r.Use(logging.ErrorLoggingMiddleware(logger))

// Panic recovery
r.Use(logging.RecoveryLoggingMiddleware(logger))

// Audit trail (state-changing operations)
auditLogger := logging.NewAuditLogger(logger, db)
r.Use(logging.AuditMiddleware(auditLogger))
```

### 3. Configure Database Slow Query Logging

```go
// Log queries > 100ms
slowQueryLogger := logging.NewSlowQueryLogger(logger, 100*time.Millisecond)
db.Logger = slowQueryLogger
```

## Usage Examples

### Basic Logging

```go
// Simple logging
logging.Info("Server started", "port", 8080)
logging.Error("Failed to connect", "error", err)

// With fields
logger := logging.WithFields(map[string]interface{}{
    "user_id": "123",
    "action": "create",
})
logger.Info("User action performed")

// With context
ctx := context.WithValue(ctx, "request_id", "req-123")
logger.WithContext(ctx).Info("Processing request")
```

### HTTP Request Logging

```go
// Automatically logged by middleware
// Example output:
{
  "time": "2025-01-08T10:30:00Z",
  "level": "INFO",
  "msg": "HTTP Request",
  "request_id": "550e8400-e29b-41d4-a716-446655440000",
  "method": "POST",
  "path": "/api/v1/users",
  "status": 201,
  "duration_ms": 45,
  "client_ip": "192.168.1.1",
  "user_id": "user-123"
}
```

### Slow Query Logging

```go
// Automatically logged by GORM logger
// Example output:
{
  "time": "2025-01-08T10:30:00Z",
  "level": "WARN",
  "msg": "Slow query detected: SELECT * FROM users WHERE ...",
  "duration_ms": 150,
  "rows": 100,
  "slow_query": true,
  "threshold_ms": 100
}
```

### Audit Logging

```go
auditLogger := logging.NewAuditLogger(logger, db)

// Log resource creation
auditLogger.LogCreate(ctx, "user", userID, adminID, companyID, userData)

// Log resource update (with change tracking)
auditLogger.LogUpdate(ctx, "vehicle", vehicleID, userID, companyID, oldData, newData)

// Log resource deletion
auditLogger.LogDelete(ctx, "driver", driverID, userID, companyID)

// Log authentication events
auditLogger.LogAuthEvent("login", userID, email, ipAddress, true)

// Log payment events
auditLogger.LogPaymentEvent(ctx, "payment_completed", paymentID, invoiceID, userID, companyID, amount, metadata)

// Example output:
{
  "time": "2025-01-08T10:30:00Z",
  "level": "INFO",
  "msg": "Audit event recorded",
  "action": "update",
  "resource": "vehicle",
  "resource_id": "vehicle-123",
  "user_id": "user-456",
  "company_id": "company-789",
  "ip_address": "192.168.1.1",
  "changes": {
    "status": {"old": "active", "new": "maintenance"},
    "mileage": {"old": 50000, "new": 51000}
  }
}
```

### Security Event Logging

```go
logger.LogSecurityEvent("failed_login", userID, ipAddress, map[string]interface{}{
    "email": email,
    "attempts": 3,
    "reason": "invalid_password",
})

// Example output:
{
  "time": "2025-01-08T10:30:00Z",
  "level": "WARN",
  "msg": "Security event",
  "security_event": "failed_login",
  "user_id": "user-123",
  "ip_address": "192.168.1.1",
  "email": "user@example.com",
  "attempts": 3
}
```

### Performance Monitoring

```go
perfMonitor := logging.NewPerformanceMonitor(logger)

// Track operation performance
err := perfMonitor.TrackOperation("export_report", func() error {
    return exportService.GenerateReport(ctx)
})

// Track with result
result, err := perfMonitor.TrackOperationWithResult("fetch_analytics", func() (interface{}, error) {
    return analyticsService.GetData(ctx)
})
```

## Configuration

### Environment Variables

```bash
# Log level (debug, info, warn, error)
LOG_LEVEL=info

# Environment (development, staging, production)
ENVIRONMENT=production
```

### Log Levels

- **DEBUG**: Detailed information for debugging (all queries, cache operations)
- **INFO**: General information (HTTP requests, job completions)
- **WARN**: Warning conditions (slow queries, high error rates)
- **ERROR**: Error conditions (failed operations, exceptions)

### Performance Thresholds

```go
// HTTP requests
PerformanceLoggingMiddleware(logger, 1*time.Second)

// Database queries
NewSlowQueryLogger(logger, 100*time.Millisecond)

// Custom operations
TrackOperation("operation_name", func() error {
    // Warns if > 500ms
})
```

## Log Output Examples

### Successful Request
```json
{
  "time": "2025-01-08T10:30:00Z",
  "level": "INFO",
  "msg": "HTTP Request",
  "request_id": "550e8400-e29b-41d4-a716-446655440000",
  "method": "GET",
  "path": "/api/v1/vehicles",
  "status": 200,
  "duration_ms": 25,
  "client_ip": "192.168.1.1",
  "user_id": "user-123",
  "company_id": "company-456",
  "response_size": 1024
}
```

### Error Request
```json
{
  "time": "2025-01-08T10:30:00Z",
  "level": "ERROR",
  "msg": "HTTP Request - Server Error",
  "request_id": "550e8400-e29b-41d4-a716-446655440001",
  "method": "POST",
  "path": "/api/v1/payments",
  "status": 500,
  "duration_ms": 100,
  "client_ip": "192.168.1.1",
  "errors": "internal server error: database connection failed"
}
```

### Slow Query
```json
{
  "time": "2025-01-08T10:30:00Z",
  "level": "WARN",
  "msg": "Slow query detected: SELECT * FROM gps_tracks WHERE vehicle_id = $1 AND timestamp > $2",
  "duration_ms": 250,
  "rows": 10000,
  "slow_query": true,
  "threshold_ms": 100,
  "request_id": "550e8400-e29b-41d4-a716-446655440002"
}
```

### Audit Event
```json
{
  "time": "2025-01-08T10:30:00Z",
  "level": "INFO",
  "msg": "Audit event recorded",
  "action": "create",
  "resource": "driver",
  "resource_id": "driver-123",
  "user_id": "admin-456",
  "company_id": "company-789",
  "ip_address": "192.168.1.1",
  "changes": {
    "name": "John Doe",
    "sim_number": "1234567890",
    "status": "active"
  }
}
```

## Features

### Sensitive Data Filtering

The logging system automatically filters sensitive data from:
- `/auth/login`, `/auth/register` (passwords)
- `/auth/change-password`, `/auth/reset-password` (passwords)
- `/payment` endpoints (payment details)

### Request ID Tracking

Every request gets a unique UUID that can be tracked across:
- HTTP requests
- Database queries
- Cache operations
- Job executions
- Audit events

### Context Propagation

Pass context through layers:
```go
ctx = context.WithValue(ctx, "request_id", requestID)
ctx = context.WithValue(ctx, "user_id", userID)
ctx = context.WithValue(ctx, "company_id", companyID)

logger.WithContext(ctx).Info("Processing")
```

## Best Practices

### 1. Use Appropriate Log Levels
```go
logger.Debug("Cache hit", "key", cacheKey)           // Development debugging
logger.Info("User created", "user_id", userID)       // Important events
logger.Warn("Slow query", "duration", duration)      // Performance issues
logger.Error("Failed to process", "error", err)      // Errors requiring attention
```

### 2. Add Context
```go
// Good: Contextual information
logger.Info("Payment processed",
    "payment_id", paymentID,
    "amount", amount,
    "method", method,
)

// Bad: Vague logging
logger.Info("Payment successful")
```

### 3. Use Structured Fields
```go
// Good: Structured
logger.WithFields(map[string]interface{}{
    "user_id": userID,
    "action": "login",
    "ip": ipAddress,
}).Info("User logged in")

// Bad: String concatenation
logger.Info(fmt.Sprintf("User %s logged in from %s", userID, ipAddress))
```

### 4. Log Security Events
```go
// Always log security-related events
auditLogger.LogAuthEvent("failed_login", userID, email, ipAddress, false)
auditLogger.LogSecurityEvent("unauthorized_access", userID, ipAddress, metadata)
```

## Performance Impact

- **Negligible overhead** for INFO level in production
- **JSON marshaling** ~1-2ms per log entry
- **Async database writes** for audit logs (non-blocking)
- **Context propagation** <1μs

## Integration

The logging system is fully integrated with:
- ✅ Gin HTTP framework
- ✅ GORM database ORM
- ✅ Redis cache
- ✅ Background job system
- ✅ Rate limiting
- ✅ Authentication/Authorization

## Testing

```bash
# Run logging tests
go test ./internal/common/logging/... -v -cover

# Test coverage: 21.3%+
```

## Production Deployment

### Docker Logging
```bash
# JSON logs to stdout (captured by Docker)
docker logs fleettracker-backend --tail 100 -f
```

### Log Aggregation
Works seamlessly with:
- **ELK Stack** (Elasticsearch, Logstash, Kibana)
- **Grafana Loki**
- **AWS CloudWatch**
- **Google Cloud Logging**
- **Datadog**

### Log Rotation
Handled by container orchestration or log aggregation services.

## Troubleshooting

### High Log Volume
```go
// Reduce log level in production
LOG_LEVEL=warn

// Disable slow query logging for specific queries
db.Session(&gorm.Session{Logger: logger.LogMode(logger.Silent)})
```

### Missing Request IDs
```go
// Ensure middleware is registered before routes
r.Use(logging.RequestLoggingMiddleware(logger))
```

### Audit Logs Not Persisted
```go
// Check database connection in AuditLogger
auditLogger := logging.NewAuditLogger(logger, db)
```

## License

MIT License - Part of FleetTracker Pro

