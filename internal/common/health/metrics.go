package health

import (
	"fmt"
	"net/http"
	"runtime"
	"time"

	"github.com/gin-gonic/gin"
)

// MetricsHandler provides Prometheus-compatible metrics
type MetricsHandler struct {
	checker *HealthChecker
}

// NewMetricsHandler creates a new metrics handler
func NewMetricsHandler(checker *HealthChecker) *MetricsHandler {
	return &MetricsHandler{
		checker: checker,
	}
}

// HandleMetrics handles Prometheus metrics endpoint
// @Summary Prometheus metrics
// @Description Prometheus-compatible metrics endpoint
// @Tags health
// @Produce text/plain
// @Success 200 {string} string "Prometheus metrics"
// @Router /metrics [get]
func (mh *MetricsHandler) HandleMetrics(c *gin.Context) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	
	uptime := mh.checker.GetUptime().Seconds()
	
	metrics := fmt.Sprintf(`# HELP fleettracker_up Service up status (1 = up, 0 = down)
# TYPE fleettracker_up gauge
fleettracker_up 1

# HELP fleettracker_uptime_seconds Service uptime in seconds
# TYPE fleettracker_uptime_seconds counter
fleettracker_uptime_seconds %f

# HELP fleettracker_memory_usage_bytes Memory usage in bytes
# TYPE fleettracker_memory_usage_bytes gauge
fleettracker_memory_usage_bytes %d

# HELP fleettracker_memory_alloc_bytes Allocated memory in bytes
# TYPE fleettracker_memory_alloc_bytes gauge
fleettracker_memory_alloc_bytes %d

# HELP fleettracker_goroutines Current number of goroutines
# TYPE fleettracker_goroutines gauge
fleettracker_goroutines %d

# HELP fleettracker_cpu_count Number of CPUs
# TYPE fleettracker_cpu_count gauge
fleettracker_cpu_count %d

# HELP fleettracker_gc_pause_seconds GC pause duration in seconds
# TYPE fleettracker_gc_pause_seconds gauge
fleettracker_gc_pause_seconds %f

# HELP fleettracker_heap_objects Number of allocated heap objects
# TYPE fleettracker_heap_objects gauge
fleettracker_heap_objects %d
`,
		uptime,
		m.Sys,
		m.Alloc,
		runtime.NumGoroutine(),
		runtime.NumCPU(),
		float64(m.PauseTotalNs)/1e9,
		m.HeapObjects,
	)
	
	c.Data(http.StatusOK, "text/plain; version=0.0.4; charset=utf-8", []byte(metrics))
}

// HandleMetricsJSON handles metrics in JSON format
// @Summary Metrics (JSON)
// @Description System metrics in JSON format
// @Tags health
// @Produce json
// @Success 200 {object} MetricsResponse
// @Router /metrics/json [get]
func (mh *MetricsHandler) HandleMetricsJSON(c *gin.Context) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	
	response := MetricsResponse{
		Timestamp: time.Now().UTC(),
		Service:   mh.checker.serviceName,
		Version:   mh.checker.version,
		Uptime:    mh.checker.getUptime(),
		Memory: MemoryMetrics{
			AllocMB:      m.Alloc / 1024 / 1024,
			TotalAllocMB: m.TotalAlloc / 1024 / 1024,
			SysMB:        m.Sys / 1024 / 1024,
			NumGC:        m.NumGC,
		},
		Goroutines: runtime.NumGoroutine(),
		CPUCount:   runtime.NumCPU(),
	}
	
	c.JSON(http.StatusOK, response)
}

// MetricsResponse represents metrics in JSON format
type MetricsResponse struct {
	Timestamp  time.Time      `json:"timestamp"`
	Service    string         `json:"service"`
	Version    string         `json:"version"`
	Uptime     string         `json:"uptime"`
	Memory     MemoryMetrics  `json:"memory"`
	Goroutines int            `json:"goroutines"`
	CPUCount   int            `json:"cpu_count"`
}

// MemoryMetrics represents memory metrics
type MemoryMetrics struct {
	AllocMB      uint64 `json:"alloc_mb"`
	TotalAllocMB uint64 `json:"total_alloc_mb"`
	SysMB        uint64 `json:"sys_mb"`
	NumGC        uint32 `json:"num_gc"`
}

// SetupMetricsRoutes sets up metrics routes
func SetupMetricsRoutes(r *gin.Engine, handler *MetricsHandler) {
	// Prometheus metrics (text format)
	r.GET("/metrics", handler.HandleMetrics)
	
	// JSON metrics (for dashboards)
	r.GET("/metrics/json", handler.HandleMetricsJSON)
}

