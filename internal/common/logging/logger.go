package logging

import (
	"context"
	"io"
	"log/slog"
	"os"
	"time"
)

// LogLevel represents logging level
type LogLevel string

const (
	LevelDebug LogLevel = "debug"
	LevelInfo  LogLevel = "info"
	LevelWarn  LogLevel = "warn"
	LevelError LogLevel = "error"
)

// LoggerConfig holds logger configuration
type LoggerConfig struct {
	Level      LogLevel
	Format     string // "json" or "text"
	Output     io.Writer
	AddSource  bool // Add source file and line number
	TimeFormat string
}

// DefaultLoggerConfig returns default logger configuration
func DefaultLoggerConfig() *LoggerConfig {
	return &LoggerConfig{
		Level:      LevelInfo,
		Format:     "json",
		Output:     os.Stdout,
		AddSource:  true,
		TimeFormat: time.RFC3339,
	}
}

// Logger wraps slog.Logger with additional functionality
type Logger struct {
	*slog.Logger
	config *LoggerConfig
}

// NewLogger creates a new structured logger
func NewLogger(config *LoggerConfig) *Logger {
	if config == nil {
		config = DefaultLoggerConfig()
	}

	// Convert log level
	var level slog.Level
	switch config.Level {
	case LevelDebug:
		level = slog.LevelDebug
	case LevelInfo:
		level = slog.LevelInfo
	case LevelWarn:
		level = slog.LevelWarn
	case LevelError:
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}

	// Create handler options
	opts := &slog.HandlerOptions{
		Level:     level,
		AddSource: config.AddSource,
	}

	// Create handler based on format
	var handler slog.Handler
	if config.Format == "json" {
		handler = slog.NewJSONHandler(config.Output, opts)
	} else {
		handler = slog.NewTextHandler(config.Output, opts)
	}

	return &Logger{
		Logger: slog.New(handler),
		config: config,
	}
}

// WithContext returns a logger with context values
func (l *Logger) WithContext(ctx context.Context) *Logger {
	return &Logger{
		Logger: l.Logger.With(contextFields(ctx)...),
		config: l.config,
	}
}

// WithFields returns a logger with additional fields
func (l *Logger) WithFields(fields map[string]interface{}) *Logger {
	args := make([]interface{}, 0, len(fields)*2)
	for k, v := range fields {
		args = append(args, k, v)
	}
	return &Logger{
		Logger: l.Logger.With(args...),
		config: l.config,
	}
}

// WithField returns a logger with an additional field
func (l *Logger) WithField(key string, value interface{}) *Logger {
	return &Logger{
		Logger: l.Logger.With(key, value),
		config: l.config,
	}
}

// LogHTTPRequest logs HTTP request details
func (l *Logger) LogHTTPRequest(method, path string, statusCode int, duration time.Duration, fields map[string]interface{}) {
	attrs := []slog.Attr{
		slog.String("method", method),
		slog.String("path", path),
		slog.Int("status", statusCode),
		slog.Duration("duration", duration),
	}

	for k, v := range fields {
		attrs = append(attrs, slog.Any(k, v))
	}

	l.LogAttrs(context.Background(), slog.LevelInfo, "HTTP Request", attrs...)
}

// LogError logs error with stack trace
func (l *Logger) LogError(err error, message string, fields map[string]interface{}) {
	args := []interface{}{"error", err}
	for k, v := range fields {
		args = append(args, k, v)
	}
	l.Error(message, args...)
}

// LogSlowQuery logs slow database queries
func (l *Logger) LogSlowQuery(query string, duration time.Duration, fields map[string]interface{}) {
	args := []interface{}{
		"query", query,
		"duration", duration,
		"slow_query", true,
	}
	for k, v := range fields {
		args = append(args, k, v)
	}
	l.Warn("Slow query detected", args...)
}

// LogAudit logs audit trail events
func (l *Logger) LogAudit(action, resource, resourceID, userID string, fields map[string]interface{}) {
	args := []interface{}{
		"audit_type", "security",
		"action", action,
		"resource", resource,
		"resource_id", resourceID,
		"user_id", userID,
	}
	for k, v := range fields {
		args = append(args, k, v)
	}
	l.Info("Audit event", args...)
}

// LogSecurityEvent logs security-related events
func (l *Logger) LogSecurityEvent(eventType, userID, ipAddress string, fields map[string]interface{}) {
	args := []interface{}{
		"security_event", eventType,
		"user_id", userID,
		"ip_address", ipAddress,
	}
	for k, v := range fields {
		args = append(args, k, v)
	}
	l.Warn("Security event", args...)
}

// LogCacheOperation logs cache operations
func (l *Logger) LogCacheOperation(operation, key string, hit bool, duration time.Duration) {
	l.Debug("Cache operation",
		"operation", operation,
		"key", key,
		"hit", hit,
		"duration", duration,
	)
}

// LogJobExecution logs job execution
func (l *Logger) LogJobExecution(jobID, jobType, status string, duration time.Duration, err error) {
	args := []interface{}{
		"job_id", jobID,
		"job_type", jobType,
		"status", status,
		"duration", duration,
	}
	if err != nil {
		args = append(args, "error", err)
	}
	
	if status == "failed" {
		l.Error("Job execution failed", args...)
	} else {
		l.Info("Job execution completed", args...)
	}
}

// LogDatabaseOperation logs database operations
func (l *Logger) LogDatabaseOperation(operation, table string, rowsAffected int64, duration time.Duration) {
	l.Debug("Database operation",
		"operation", operation,
		"table", table,
		"rows_affected", rowsAffected,
		"duration", duration,
	)
}

// Helper function to extract context fields
func contextFields(ctx context.Context) []interface{} {
	fields := make([]interface{}, 0)

	if requestID := ctx.Value("request_id"); requestID != nil {
		fields = append(fields, "request_id", requestID)
	}

	if userID := ctx.Value("user_id"); userID != nil {
		fields = append(fields, "user_id", userID)
	}

	if companyID := ctx.Value("company_id"); companyID != nil {
		fields = append(fields, "company_id", companyID)
	}

	return fields
}

// Global logger instance
var defaultLogger *Logger

// InitDefaultLogger initializes the global logger
func InitDefaultLogger(config *LoggerConfig) {
	defaultLogger = NewLogger(config)
}

// GetLogger returns the global logger
func GetLogger() *Logger {
	if defaultLogger == nil {
		defaultLogger = NewLogger(DefaultLoggerConfig())
	}
	return defaultLogger
}

// Convenience functions using global logger

// Debug logs a debug message
func Debug(msg string, args ...interface{}) {
	GetLogger().Debug(msg, args...)
}

// Info logs an info message
func Info(msg string, args ...interface{}) {
	GetLogger().Info(msg, args...)
}

// Warn logs a warning message
func Warn(msg string, args ...interface{}) {
	GetLogger().Warn(msg, args...)
}

// Error logs an error message
func Error(msg string, args ...interface{}) {
	GetLogger().Error(msg, args...)
}

// WithFields returns a logger with fields
func WithFields(fields map[string]interface{}) *Logger {
	return GetLogger().WithFields(fields)
}

// WithField returns a logger with a field
func WithField(key string, value interface{}) *Logger {
	return GetLogger().WithField(key, value)
}

