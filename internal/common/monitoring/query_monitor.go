package monitoring

import (
	"context"
	"fmt"
	"log"
	"time"

	"gorm.io/gorm"
)

// QueryMonitor provides database query performance monitoring
type QueryMonitor struct {
	slowQueryThreshold time.Duration
	logger            *log.Logger
}

// NewQueryMonitor creates a new query monitor
func NewQueryMonitor(slowQueryThreshold time.Duration, logger *log.Logger) *QueryMonitor {
	return &QueryMonitor{
		slowQueryThreshold: slowQueryThreshold,
		logger:            logger,
	}
}

// MonitorQuery wraps a database operation with performance monitoring
func (qm *QueryMonitor) MonitorQuery(ctx context.Context, operation string, fn func() error) error {
	start := time.Now()
	
	err := fn()
	
	duration := time.Since(start)
	
	// Log slow queries
	if duration > qm.slowQueryThreshold {
		qm.logger.Printf("SLOW QUERY: %s took %v", operation, duration)
	}
	
	// Log all queries in debug mode
	qm.logger.Printf("QUERY: %s took %v", operation, duration)
	
	return err
}

// MonitorQueryWithResult wraps a database operation that returns a result
func (qm *QueryMonitor) MonitorQueryWithResult(ctx context.Context, operation string, fn func() (interface{}, error)) (interface{}, error) {
	start := time.Now()
	
	result, err := fn()
	
	duration := time.Since(start)
	
	// Log slow queries
	if duration > qm.slowQueryThreshold {
		qm.logger.Printf("SLOW QUERY: %s took %v", operation, duration)
	}
	
	// Log all queries in debug mode
	qm.logger.Printf("QUERY: %s took %v", operation, duration)
	
	return result, err
}

// QueryStats represents query performance statistics
type QueryStats struct {
	Operation     string        `json:"operation"`
	Duration      time.Duration `json:"duration"`
	Timestamp     time.Time     `json:"timestamp"`
	IsSlow        bool          `json:"is_slow"`
	Error         string        `json:"error,omitempty"`
}

// QueryMetrics collects query performance metrics
type QueryMetrics struct {
	TotalQueries    int64         `json:"total_queries"`
	SlowQueries     int64         `json:"slow_queries"`
	AverageDuration time.Duration `json:"average_duration"`
	MaxDuration     time.Duration `json:"max_duration"`
	MinDuration     time.Duration `json:"min_duration"`
	ErrorCount      int64         `json:"error_count"`
}

// MetricsCollector collects and aggregates query metrics
type MetricsCollector struct {
	metrics map[string]*QueryMetrics
	stats   []QueryStats
}

// NewMetricsCollector creates a new metrics collector
func NewMetricsCollector() *MetricsCollector {
	return &MetricsCollector{
		metrics: make(map[string]*QueryMetrics),
		stats:   make([]QueryStats, 0),
	}
}

// RecordQuery records a query execution
func (mc *MetricsCollector) RecordQuery(operation string, duration time.Duration, isSlow bool, err error) {
	// Update metrics
	if mc.metrics[operation] == nil {
		mc.metrics[operation] = &QueryMetrics{}
	}
	
	metrics := mc.metrics[operation]
	metrics.TotalQueries++
	
	if isSlow {
		metrics.SlowQueries++
	}
	
	if err != nil {
		metrics.ErrorCount++
	}
	
	// Update duration statistics
	if metrics.TotalQueries == 1 {
		metrics.MinDuration = duration
		metrics.MaxDuration = duration
		metrics.AverageDuration = duration
	} else {
		if duration < metrics.MinDuration {
			metrics.MinDuration = duration
		}
		if duration > metrics.MaxDuration {
			metrics.MaxDuration = duration
		}
		
		// Calculate new average
		totalDuration := metrics.AverageDuration * time.Duration(metrics.TotalQueries-1)
		metrics.AverageDuration = (totalDuration + duration) / time.Duration(metrics.TotalQueries)
	}
	
	// Record individual stat
	stat := QueryStats{
		Operation: operation,
		Duration:  duration,
		Timestamp: time.Now(),
		IsSlow:    isSlow,
	}
	
	if err != nil {
		stat.Error = err.Error()
	}
	
	mc.stats = append(mc.stats, stat)
	
	// Keep only last 1000 stats to prevent memory issues
	if len(mc.stats) > 1000 {
		mc.stats = mc.stats[len(mc.stats)-1000:]
	}
}

// GetMetrics returns current metrics
func (mc *MetricsCollector) GetMetrics() map[string]*QueryMetrics {
	return mc.metrics
}

// GetStats returns recent query stats
func (mc *MetricsCollector) GetStats() []QueryStats {
	return mc.stats
}

// GetSlowQueries returns slow query stats
func (mc *MetricsCollector) GetSlowQueries() []QueryStats {
	var slowQueries []QueryStats
	for _, stat := range mc.stats {
		if stat.IsSlow {
			slowQueries = append(slowQueries, stat)
		}
	}
	return slowQueries
}

// Reset resets all metrics
func (mc *MetricsCollector) Reset() {
	mc.metrics = make(map[string]*QueryMetrics)
	mc.stats = make([]QueryStats, 0)
}

// GORM plugin for query monitoring
type QueryMonitorPlugin struct {
	monitor *QueryMonitor
	collector *MetricsCollector
}

// NewQueryMonitorPlugin creates a new GORM plugin for query monitoring
func NewQueryMonitorPlugin(monitor *QueryMonitor, collector *MetricsCollector) *QueryMonitorPlugin {
	return &QueryMonitorPlugin{
		monitor:   monitor,
		collector: collector,
	}
}

// Name returns the plugin name
func (p *QueryMonitorPlugin) Name() string {
	return "query-monitor"
}

// Initialize initializes the plugin
func (p *QueryMonitorPlugin) Initialize(db *gorm.DB) error {
	// Register callbacks
	db.Callback().Query().Before("gorm:query").Register("query-monitor:before", p.beforeQuery)
	db.Callback().Query().After("gorm:query").Register("query-monitor:after", p.afterQuery)
	
	db.Callback().Create().Before("gorm:create").Register("query-monitor:before", p.beforeQuery)
	db.Callback().Create().After("gorm:create").Register("query-monitor:after", p.afterQuery)
	
	db.Callback().Update().Before("gorm:update").Register("query-monitor:before", p.beforeQuery)
	db.Callback().Update().After("gorm:update").Register("query-monitor:after", p.afterQuery)
	
	db.Callback().Delete().Before("gorm:delete").Register("query-monitor:before", p.beforeQuery)
	db.Callback().Delete().After("gorm:delete").Register("query-monitor:after", p.afterQuery)
	
	return nil
}

// beforeQuery callback
func (p *QueryMonitorPlugin) beforeQuery(db *gorm.DB) {
	db.InstanceSet("query_start_time", time.Now())
}

// afterQuery callback
func (p *QueryMonitorPlugin) afterQuery(db *gorm.DB) {
	startTime, ok := db.InstanceGet("query_start_time")
	if !ok {
		return
	}
	
	duration := time.Since(startTime.(time.Time))
	
	// Extract operation type from SQL
	operation := "unknown"
	if db.Statement != nil {
		sql := db.Statement.SQL.String()
		if len(sql) > 0 {
			// Simple operation detection
			switch {
			case sql[0] == 'S' || sql[0] == 's':
				operation = "SELECT"
			case sql[0] == 'I' || sql[0] == 'i':
				operation = "INSERT"
			case sql[0] == 'U' || sql[0] == 'u':
				operation = "UPDATE"
			case sql[0] == 'D' || sql[0] == 'd':
				operation = "DELETE"
			default:
				operation = "OTHER"
			}
		}
	}
	
	isSlow := duration > p.monitor.slowQueryThreshold
	
	// Record the query
	p.collector.RecordQuery(operation, duration, isSlow, db.Error)
	
	// Log if slow
	if isSlow {
		p.monitor.logger.Printf("SLOW QUERY: %s took %v", operation, duration)
	}
}

// QueryPerformanceReport generates a performance report
type QueryPerformanceReport struct {
	GeneratedAt time.Time                `json:"generated_at"`
	Summary     map[string]*QueryMetrics `json:"summary"`
	SlowQueries []QueryStats             `json:"slow_queries"`
	Recommendations []string             `json:"recommendations"`
}

// GenerateReport generates a performance report
func (mc *MetricsCollector) GenerateReport() *QueryPerformanceReport {
	report := &QueryPerformanceReport{
		GeneratedAt: time.Now(),
		Summary:     mc.GetMetrics(),
		SlowQueries: mc.GetSlowQueries(),
		Recommendations: []string{},
	}
	
	// Generate recommendations based on metrics
	for operation, metrics := range report.Summary {
		if metrics.SlowQueries > 0 {
			report.Recommendations = append(report.Recommendations, 
				fmt.Sprintf("Consider optimizing %s queries - %d slow queries detected", operation, metrics.SlowQueries))
		}
		
		if metrics.ErrorCount > 0 {
			report.Recommendations = append(report.Recommendations, 
				fmt.Sprintf("High error rate for %s queries - %d errors detected", operation, metrics.ErrorCount))
		}
		
		if metrics.AverageDuration > 100*time.Millisecond {
			report.Recommendations = append(report.Recommendations, 
				fmt.Sprintf("Consider adding indexes for %s queries - average duration: %v", operation, metrics.AverageDuration))
		}
	}
	
	return report
}
