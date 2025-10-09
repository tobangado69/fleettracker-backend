package health

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
)

// Status represents health check status
type Status string

const (
	StatusHealthy   Status = "healthy"
	StatusUnhealthy Status = "unhealthy"
	StatusDegraded  Status = "degraded"
)

// HealthChecker provides health check functionality
type HealthChecker struct {
	db          *gorm.DB
	redis       *redis.Client
	startTime   time.Time
	version     string
	serviceName string
	mu          sync.RWMutex
}

// NewHealthChecker creates a new health checker
func NewHealthChecker(db *gorm.DB, redis *redis.Client, serviceName, version string) *HealthChecker {
	return &HealthChecker{
		db:          db,
		redis:       redis,
		startTime:   time.Now(),
		version:     version,
		serviceName: serviceName,
	}
}

// HealthResponse represents the health check response
type HealthResponse struct {
	Status       Status                 `json:"status"`
	Timestamp    time.Time              `json:"timestamp"`
	Service      string                 `json:"service"`
	Version      string                 `json:"version"`
	Uptime       string                 `json:"uptime"`
	Dependencies map[string]Dependency  `json:"dependencies,omitempty"`
	System       *SystemMetrics         `json:"system,omitempty"`
	Errors       []string               `json:"errors,omitempty"`
}

// Dependency represents a dependency health check
type Dependency struct {
	Status    Status  `json:"status"`
	LatencyMs int64   `json:"latency_ms"`
	Message   string  `json:"message,omitempty"`
	Error     string  `json:"error,omitempty"`
}

// SystemMetrics represents system health metrics
type SystemMetrics struct {
	MemoryUsageMB    uint64 `json:"memory_usage_mb"`
	MemoryAllocMB    uint64 `json:"memory_alloc_mb"`
	GoroutineCount   int    `json:"goroutine_count"`
	CPUCount         int    `json:"cpu_count"`
}

// Check performs a basic health check (liveness probe)
func (hc *HealthChecker) Check() HealthResponse {
	return HealthResponse{
		Status:    StatusHealthy,
		Timestamp: time.Now().UTC(),
		Service:   hc.serviceName,
		Version:   hc.version,
		Uptime:    hc.getUptime(),
	}
}

// CheckReadiness performs a comprehensive readiness check
func (hc *HealthChecker) CheckReadiness(ctx context.Context) HealthResponse {
	hc.mu.RLock()
	defer hc.mu.RUnlock()

	response := HealthResponse{
		Status:       StatusHealthy,
		Timestamp:    time.Now().UTC(),
		Service:      hc.serviceName,
		Version:      hc.version,
		Uptime:       hc.getUptime(),
		Dependencies: make(map[string]Dependency),
		System:       hc.getSystemMetrics(),
		Errors:       []string{},
	}

	// Check database
	if hc.db != nil {
		dbDep := hc.checkDatabase(ctx)
		response.Dependencies["database"] = dbDep
		if dbDep.Status != StatusHealthy {
			response.Status = StatusUnhealthy
			response.Errors = append(response.Errors, fmt.Sprintf("database: %s", dbDep.Error))
		}
	} else {
		response.Dependencies["database"] = Dependency{
			Status: StatusUnhealthy,
			Error:  "database not configured",
		}
		response.Status = StatusUnhealthy
		response.Errors = append(response.Errors, "database: not configured")
	}

	// Check Redis
	if hc.redis != nil {
		redisDep := hc.checkRedis(ctx)
		response.Dependencies["redis"] = redisDep
		if redisDep.Status != StatusHealthy {
			// Redis failure is degraded, not unhealthy (caching is optional)
			if response.Status == StatusHealthy {
				response.Status = StatusDegraded
			}
			response.Errors = append(response.Errors, fmt.Sprintf("redis: %s", redisDep.Error))
		}
	} else {
		response.Dependencies["redis"] = Dependency{
			Status: StatusUnhealthy,
			Error:  "redis not configured",
		}
		// Redis is optional, so degraded not unhealthy
		if response.Status == StatusHealthy {
			response.Status = StatusDegraded
		}
		response.Errors = append(response.Errors, "redis: not configured")
	}

	return response
}

// CheckLiveness performs a liveness check (K8s liveness probe)
func (hc *HealthChecker) CheckLiveness() HealthResponse {
	// Simple check - just verify the service is responsive
	return HealthResponse{
		Status:    StatusHealthy,
		Timestamp: time.Now().UTC(),
		Service:   hc.serviceName,
		Version:   hc.version,
	}
}

// checkDatabase checks database connectivity
func (hc *HealthChecker) checkDatabase(ctx context.Context) Dependency {
	start := time.Now()
	
	// Create context with timeout
	checkCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	sqlDB, err := hc.db.DB()
	if err != nil {
		return Dependency{
			Status:    StatusUnhealthy,
			LatencyMs: time.Since(start).Milliseconds(),
			Error:     fmt.Sprintf("failed to get database: %v", err),
		}
	}

	// Ping database
	if err := sqlDB.PingContext(checkCtx); err != nil {
		return Dependency{
			Status:    StatusUnhealthy,
			LatencyMs: time.Since(start).Milliseconds(),
			Error:     fmt.Sprintf("database ping failed: %v", err),
		}
	}

	latency := time.Since(start).Milliseconds()

	// Check if database is too slow
	status := StatusHealthy
	message := "connected"
	if latency > 1000 {
		status = StatusDegraded
		message = "slow response"
	}

	return Dependency{
		Status:    status,
		LatencyMs: latency,
		Message:   message,
	}
}

// checkRedis checks Redis connectivity
func (hc *HealthChecker) checkRedis(ctx context.Context) Dependency {
	start := time.Now()

	// Create context with timeout
	checkCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	// Ping Redis
	if err := hc.redis.Ping(checkCtx).Err(); err != nil {
		return Dependency{
			Status:    StatusUnhealthy,
			LatencyMs: time.Since(start).Milliseconds(),
			Error:     fmt.Sprintf("redis ping failed: %v", err),
		}
	}

	latency := time.Since(start).Milliseconds()

	// Check if Redis is too slow
	status := StatusHealthy
	message := "connected"
	if latency > 500 {
		status = StatusDegraded
		message = "slow response"
	}

	return Dependency{
		Status:    status,
		LatencyMs: latency,
		Message:   message,
	}
}

// getSystemMetrics returns current system metrics
func (hc *HealthChecker) getSystemMetrics() *SystemMetrics {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	return &SystemMetrics{
		MemoryUsageMB:  m.Sys / 1024 / 1024,
		MemoryAllocMB:  m.Alloc / 1024 / 1024,
		GoroutineCount: runtime.NumGoroutine(),
		CPUCount:       runtime.NumCPU(),
	}
}

// getUptime returns the service uptime
func (hc *HealthChecker) getUptime() string {
	duration := time.Since(hc.startTime)
	
	hours := int(duration.Hours())
	minutes := int(duration.Minutes()) % 60
	seconds := int(duration.Seconds()) % 60
	
	if hours > 0 {
		return fmt.Sprintf("%dh %dm %ds", hours, minutes, seconds)
	} else if minutes > 0 {
		return fmt.Sprintf("%dm %ds", minutes, seconds)
	}
	return fmt.Sprintf("%ds", seconds)
}

// GetUptime returns the service uptime duration
func (hc *HealthChecker) GetUptime() time.Duration {
	return time.Since(hc.startTime)
}

// GetStartTime returns the service start time
func (hc *HealthChecker) GetStartTime() time.Time {
	return hc.startTime
}

// DetailedHealthCheck performs a comprehensive health check with all details
type DetailedHealthCheck struct {
	HealthResponse
	JobSystem    *JobSystemHealth    `json:"job_system,omitempty"`
	RateLimit    *RateLimitHealth    `json:"rate_limit,omitempty"`
}

// JobSystemHealth represents job system health
type JobSystemHealth struct {
	Status         Status `json:"status"`
	ActiveWorkers  int    `json:"active_workers"`
	QueueSize      int64  `json:"queue_size"`
	ProcessedJobs  int64  `json:"processed_jobs"`
	FailedJobs     int64  `json:"failed_jobs"`
}

// RateLimitHealth represents rate limiting health
type RateLimitHealth struct {
	Status        Status  `json:"status"`
	TotalRequests int64   `json:"total_requests"`
	BlockedCount  int64   `json:"blocked_count"`
	BlockRate     float64 `json:"block_rate"`
}

