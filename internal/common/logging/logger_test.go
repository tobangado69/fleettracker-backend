package logging

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"
)

func TestNewLogger(t *testing.T) {
	tests := []struct {
		name   string
		config *LoggerConfig
	}{
		{
			name:   "default config",
			config: nil,
		},
		{
			name: "json format",
			config: &LoggerConfig{
				Level:     LevelInfo,
				Format:    "json",
				AddSource: true,
			},
		},
		{
			name: "text format",
			config: &LoggerConfig{
				Level:     LevelDebug,
				Format:    "text",
				AddSource: false,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := NewLogger(tt.config)
			if logger == nil {
				t.Error("Expected logger to be created")
			}
		})
	}
}

func TestLogger_WithContext(t *testing.T) {
	buf := &bytes.Buffer{}
	config := &LoggerConfig{
		Level:  LevelInfo,
		Format: "json",
		Output: buf,
	}
	logger := NewLogger(config)

	ctx := context.WithValue(context.Background(), "request_id", "test-123")
	ctx = context.WithValue(ctx, "user_id", "user-456")

	contextLogger := logger.WithContext(ctx)
	contextLogger.Info("test message")

	output := buf.String()
	if !strings.Contains(output, "test-123") {
		t.Error("Expected request_id in log output")
	}
	if !strings.Contains(output, "user-456") {
		t.Error("Expected user_id in log output")
	}
}

func TestLogger_WithFields(t *testing.T) {
	buf := &bytes.Buffer{}
	config := &LoggerConfig{
		Level:  LevelInfo,
		Format: "json",
		Output: buf,
	}
	logger := NewLogger(config)

	fields := map[string]interface{}{
		"key1": "value1",
		"key2": 123,
	}

	logger.WithFields(fields).Info("test message")

	output := buf.String()
	if !strings.Contains(output, "value1") {
		t.Error("Expected key1 in log output")
	}
	if !strings.Contains(output, "123") {
		t.Error("Expected key2 value in log output")
	}
}

func TestLogger_LogHTTPRequest(t *testing.T) {
	buf := &bytes.Buffer{}
	config := &LoggerConfig{
		Level:  LevelInfo,
		Format: "json",
		Output: buf,
	}
	logger := NewLogger(config)

	logger.LogHTTPRequest("GET", "/api/test", 200, 50*time.Millisecond, map[string]interface{}{
		"client_ip": "127.0.0.1",
	})

	output := buf.String()
	var logData map[string]interface{}
	if err := json.Unmarshal([]byte(output), &logData); err != nil {
		t.Fatalf("Failed to parse log output: %v", err)
	}

	if logData["method"] != "GET" {
		t.Errorf("Expected method GET, got %v", logData["method"])
	}
	if logData["path"] != "/api/test" {
		t.Errorf("Expected path /api/test, got %v", logData["path"])
	}
}

func TestLogger_LogError(t *testing.T) {
	buf := &bytes.Buffer{}
	config := &LoggerConfig{
		Level:  LevelError,
		Format: "json",
		Output: buf,
	}
	logger := NewLogger(config)

	testErr := errors.New("test error")
	logger.LogError(testErr, "operation failed", map[string]interface{}{
		"operation": "test_operation",
	})

	output := buf.String()
	if !strings.Contains(output, "test error") {
		t.Error("Expected error message in log output")
	}
	if !strings.Contains(output, "test_operation") {
		t.Error("Expected operation field in log output")
	}
}

func TestLogger_LogSlowQuery(t *testing.T) {
	buf := &bytes.Buffer{}
	config := &LoggerConfig{
		Level:  LevelWarn,
		Format: "json",
		Output: buf,
	}
	logger := NewLogger(config)

	logger.LogSlowQuery("SELECT * FROM users", 200*time.Millisecond, map[string]interface{}{
		"table": "users",
	})

	output := buf.String()
	if !strings.Contains(output, "SELECT * FROM users") {
		t.Error("Expected query in log output")
	}
	if !strings.Contains(output, "slow_query") {
		t.Error("Expected slow_query flag in log output")
	}
}

func TestLogger_LogAudit(t *testing.T) {
	buf := &bytes.Buffer{}
	config := &LoggerConfig{
		Level:  LevelInfo,
		Format: "json",
		Output: buf,
	}
	logger := NewLogger(config)

	logger.LogAudit("create", "user", "user-123", "admin-456", map[string]interface{}{
		"ip_address": "192.168.1.1",
	})

	output := buf.String()
	var logData map[string]interface{}
	if err := json.Unmarshal([]byte(output), &logData); err != nil {
		t.Fatalf("Failed to parse log output: %v", err)
	}

	if logData["action"] != "create" {
		t.Errorf("Expected action create, got %v", logData["action"])
	}
	if logData["resource"] != "user" {
		t.Errorf("Expected resource user, got %v", logData["resource"])
	}
}

func TestLogger_LogSecurityEvent(t *testing.T) {
	buf := &bytes.Buffer{}
	config := &LoggerConfig{
		Level:  LevelWarn,
		Format: "json",
		Output: buf,
	}
	logger := NewLogger(config)

	logger.LogSecurityEvent("failed_login", "user-123", "192.168.1.1", map[string]interface{}{
		"attempts": 3,
	})

	output := buf.String()
	if !strings.Contains(output, "failed_login") {
		t.Error("Expected security event type in log output")
	}
	if !strings.Contains(output, "192.168.1.1") {
		t.Error("Expected IP address in log output")
	}
}

func TestLogger_LogJobExecution(t *testing.T) {
	buf := &bytes.Buffer{}
	config := &LoggerConfig{
		Level:  LevelInfo,
		Format: "json",
		Output: buf,
	}
	logger := NewLogger(config)

	tests := []struct {
		name     string
		status   string
		err      error
		wantLog  bool
		logLevel string
	}{
		{
			name:     "successful job",
			status:   "completed",
			err:      nil,
			wantLog:  true,
			logLevel: "INFO",
		},
		{
			name:     "failed job",
			status:   "failed",
			err:      errors.New("job error"),
			wantLog:  true,
			logLevel: "ERROR",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf.Reset()
			logger.LogJobExecution("job-123", "test_job", tt.status, 100*time.Millisecond, tt.err)

			output := buf.String()
			if tt.wantLog && output == "" {
				t.Error("Expected log output")
			}
			if tt.wantLog && !strings.Contains(output, tt.status) {
				t.Errorf("Expected status %s in log output", tt.status)
			}
		})
	}
}

func TestGetLogger(t *testing.T) {
	// Reset default logger
	defaultLogger = nil

	logger := GetLogger()
	if logger == nil {
		t.Error("Expected default logger to be created")
	}

	// Should return same instance
	logger2 := GetLogger()
	if logger != logger2 {
		t.Error("Expected same logger instance")
	}
}

func TestConvenienceFunctions(t *testing.T) {
	buf := &bytes.Buffer{}
	config := &LoggerConfig{
		Level:  LevelDebug,
		Format: "json",
		Output: buf,
	}
	InitDefaultLogger(config)

	tests := []struct {
		name     string
		logFunc  func()
		expected string
	}{
		{
			name: "Debug",
			logFunc: func() {
				Debug("debug message", "key", "value")
			},
			expected: "debug message",
		},
		{
			name: "Info",
			logFunc: func() {
				Info("info message", "key", "value")
			},
			expected: "info message",
		},
		{
			name: "Warn",
			logFunc: func() {
				Warn("warn message", "key", "value")
			},
			expected: "warn message",
		},
		{
			name: "Error",
			logFunc: func() {
				Error("error message", "key", "value")
			},
			expected: "error message",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf.Reset()
			tt.logFunc()
			output := buf.String()
			if !strings.Contains(output, tt.expected) {
				t.Errorf("Expected %s in log output, got: %s", tt.expected, output)
			}
		})
	}
}

