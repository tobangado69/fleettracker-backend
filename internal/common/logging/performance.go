package logging

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm/logger"
)

// SlowQueryLogger logs slow database queries
type SlowQueryLogger struct {
	logger        *Logger
	slowThreshold time.Duration
	logLevel      logger.LogLevel
}

// NewSlowQueryLogger creates a new slow query logger
func NewSlowQueryLogger(log *Logger, slowThreshold time.Duration) *SlowQueryLogger {
	return &SlowQueryLogger{
		logger:        log,
		slowThreshold: slowThreshold,
		logLevel:      logger.Warn,
	}
}

// LogMode sets log mode
func (l *SlowQueryLogger) LogMode(level logger.LogLevel) logger.Interface {
	newLogger := *l
	newLogger.logLevel = level
	return &newLogger
}

// Info logs info level messages
func (l *SlowQueryLogger) Info(_ context.Context, msg string, data ...interface{}) {
	if l.logLevel >= logger.Info {
		l.logger.Info(fmt.Sprintf(msg, data...))
	}
}

// Warn logs warning level messages
func (l *SlowQueryLogger) Warn(_ context.Context, msg string, data ...interface{}) {
	if l.logLevel >= logger.Warn {
		l.logger.Warn(fmt.Sprintf(msg, data...))
	}
}

// Error logs error level messages
func (l *SlowQueryLogger) Error(_ context.Context, msg string, data ...interface{}) {
	if l.logLevel >= logger.Error {
		l.logger.Error(fmt.Sprintf(msg, data...))
	}
}

// Trace logs SQL queries
func (l *SlowQueryLogger) Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	if l.logLevel <= logger.Silent {
		return
	}

	elapsed := time.Since(begin)
	sql, rows := fc()

	fields := map[string]interface{}{
		"duration_ms": elapsed.Milliseconds(),
		"rows":        rows,
	}

	// Add context fields
	if requestID := ctx.Value("request_id"); requestID != nil {
		fields["request_id"] = requestID
	}

	// Log errors
	if err != nil && l.logLevel >= logger.Error {
		fields["error"] = err
		l.logger.WithFields(fields).Error("Database error: " + sql)
		return
	}

	// Log slow queries
	if elapsed > l.slowThreshold {
		fields["slow_query"] = true
		fields["threshold_ms"] = l.slowThreshold.Milliseconds()
		l.logger.WithFields(fields).Warn("Slow query detected: " + sql)
		return
	}

	// Log all queries in debug mode
	if l.logLevel >= logger.Info {
		l.logger.WithFields(fields).Debug("Query executed: " + sql)
	}
}

// PerformanceMonitor tracks performance metrics
type PerformanceMonitor struct {
	logger *Logger
}

// NewPerformanceMonitor creates a new performance monitor
func NewPerformanceMonitor(logger *Logger) *PerformanceMonitor {
	return &PerformanceMonitor{
		logger: logger,
	}
}

// TrackOperation tracks an operation's performance
func (pm *PerformanceMonitor) TrackOperation(name string, operation func() error) error {
	start := time.Now()
	err := operation()
	duration := time.Since(start)

	fields := map[string]interface{}{
		"operation":   name,
		"duration_ms": duration.Milliseconds(),
	}

	if err != nil {
		fields["error"] = err
		pm.logger.WithFields(fields).Error("Operation failed")
		return err
	}

	if duration > 500*time.Millisecond {
		pm.logger.WithFields(fields).Warn("Slow operation detected")
	} else {
		pm.logger.WithFields(fields).Debug("Operation completed")
	}

	return nil
}

// TrackOperationWithResult tracks an operation and returns result
func (pm *PerformanceMonitor) TrackOperationWithResult(name string, operation func() (interface{}, error)) (interface{}, error) {
	start := time.Now()
	result, err := operation()
	duration := time.Since(start)

	fields := map[string]interface{}{
		"operation":   name,
		"duration_ms": duration.Milliseconds(),
	}

	if err != nil {
		fields["error"] = err
		pm.logger.WithFields(fields).Error("Operation failed")
		return nil, err
	}

	if duration > 500*time.Millisecond {
		pm.logger.WithFields(fields).Warn("Slow operation detected")
	} else {
		pm.logger.WithFields(fields).Debug("Operation completed")
	}

	return result, nil
}

// LogMemoryUsage logs current memory usage
func (pm *PerformanceMonitor) LogMemoryUsage(memStats interface{}) {
	pm.logger.Info("Memory usage",
		"stats", memStats,
	)
}

// LogGoroutineCount logs current goroutine count
func (pm *PerformanceMonitor) LogGoroutineCount(count int) {
	if count > 1000 {
		pm.logger.Warn("High goroutine count",
			"count", count,
			"threshold", 1000,
		)
	} else {
		pm.logger.Debug("Goroutine count",
			"count", count,
		)
	}
}

