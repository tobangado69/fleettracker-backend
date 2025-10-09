package health

import (
	"context"
	"testing"
	"time"
)

func TestNewHealthChecker(t *testing.T) {
	checker := NewHealthChecker(nil, nil, "TestService", "1.0.0")
	
	if checker == nil {
		t.Fatal("Expected health checker to be created")
	}
	
	if checker.serviceName != "TestService" {
		t.Errorf("Expected service name TestService, got %s", checker.serviceName)
	}
	
	if checker.version != "1.0.0" {
		t.Errorf("Expected version 1.0.0, got %s", checker.version)
	}
}

func TestHealthChecker_Check(t *testing.T) {
	checker := NewHealthChecker(nil, nil, "TestService", "1.0.0")
	
	response := checker.Check()
	
	if response.Status != StatusHealthy {
		t.Errorf("Expected status healthy, got %s", response.Status)
	}
	
	if response.Service != "TestService" {
		t.Errorf("Expected service TestService, got %s", response.Service)
	}
	
	if response.Version != "1.0.0" {
		t.Errorf("Expected version 1.0.0, got %s", response.Version)
	}
}

func TestHealthChecker_CheckLiveness(t *testing.T) {
	checker := NewHealthChecker(nil, nil, "TestService", "1.0.0")
	
	response := checker.CheckLiveness()
	
	if response.Status != StatusHealthy {
		t.Errorf("Expected status healthy, got %s", response.Status)
	}
}

func TestHealthChecker_GetUptime(t *testing.T) {
	checker := NewHealthChecker(nil, nil, "TestService", "1.0.0")
	
	// Wait a bit
	time.Sleep(100 * time.Millisecond)
	
	uptime := checker.getUptime()
	if uptime == "" {
		t.Error("Expected uptime to be non-empty")
	}
	
	duration := checker.GetUptime()
	if duration < 100*time.Millisecond {
		t.Errorf("Expected uptime >= 100ms, got %v", duration)
	}
}

func TestHealthChecker_GetSystemMetrics(t *testing.T) {
	checker := NewHealthChecker(nil, nil, "TestService", "1.0.0")
	
	metrics := checker.getSystemMetrics()
	
	if metrics == nil {
		t.Fatal("Expected metrics to be non-nil")
		return
	}
	
	if metrics.CPUCount <= 0 {
		t.Errorf("Expected CPU count > 0, got %d", metrics.CPUCount)
	}
	
	if metrics.GoroutineCount <= 0 {
		t.Errorf("Expected goroutine count > 0, got %d", metrics.GoroutineCount)
	}
	
	if metrics.MemoryUsageMB == 0 {
		t.Error("Expected memory usage > 0")
	}
}

func TestHealthChecker_CheckReadiness_NoDependencies(t *testing.T) {
	// Test with nil dependencies (should handle gracefully)
	checker := NewHealthChecker(nil, nil, "TestService", "1.0.0")
	
	ctx := context.Background()
	response := checker.CheckReadiness(ctx)
	
	// Should return unhealthy since dependencies are nil
	if response.Status != StatusUnhealthy {
		t.Errorf("Expected status unhealthy with nil dependencies, got %s", response.Status)
	}
	
	if response.System == nil {
		t.Error("Expected system metrics to be present")
	}
	
	if len(response.Errors) == 0 {
		t.Error("Expected errors to be present with nil dependencies")
	}
}

func TestStatus_Types(t *testing.T) {
	tests := []struct {
		status   Status
		expected string
	}{
		{StatusHealthy, "healthy"},
		{StatusUnhealthy, "unhealthy"},
		{StatusDegraded, "degraded"},
	}
	
	for _, tt := range tests {
		t.Run(string(tt.status), func(t *testing.T) {
			if string(tt.status) != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, string(tt.status))
			}
		})
	}
}

func TestHealthResponse_Structure(t *testing.T) {
	now := time.Now()
	response := HealthResponse{
		Status:    StatusHealthy,
		Timestamp: now,
		Service:   "TestService",
		Version:   "1.0.0",
		Uptime:    "1h 30m 15s",
		Dependencies: map[string]Dependency{
			"database": {
				Status:    StatusHealthy,
				LatencyMs: 5,
				Message:   "connected",
			},
		},
		System: &SystemMetrics{
			MemoryUsageMB:  256,
			GoroutineCount: 10,
			CPUCount:       4,
		},
	}
	
	if response.Status != StatusHealthy {
		t.Error("Expected healthy status")
	}
	
	if response.Service != "TestService" {
		t.Errorf("Expected service TestService, got %s", response.Service)
	}
	
	if response.Version != "1.0.0" {
		t.Errorf("Expected version 1.0.0, got %s", response.Version)
	}
	
	if response.Uptime != "1h 30m 15s" {
		t.Errorf("Expected uptime 1h 30m 15s, got %s", response.Uptime)
	}
	
	if response.Timestamp.IsZero() {
		t.Error("Expected non-zero timestamp")
	}
	
	if len(response.Dependencies) != 1 {
		t.Errorf("Expected 1 dependency, got %d", len(response.Dependencies))
	}
	
	if response.System == nil {
		t.Error("Expected system metrics")
	}
}

func TestDependency_HealthyCheck(t *testing.T) {
	dep := Dependency{
		Status:    StatusHealthy,
		LatencyMs: 10,
		Message:   "connected",
	}
	
	if dep.Status != StatusHealthy {
		t.Errorf("Expected healthy status, got %s", dep.Status)
	}
	
	if dep.LatencyMs != 10 {
		t.Errorf("Expected latency 10ms, got %d", dep.LatencyMs)
	}
	
	if dep.Message != "connected" {
		t.Errorf("Expected message 'connected', got %s", dep.Message)
	}
}

func TestDependency_UnhealthyCheck(t *testing.T) {
	dep := Dependency{
		Status:    StatusUnhealthy,
		LatencyMs: 1000,
		Error:     "connection timeout",
	}
	
	if dep.Status != StatusUnhealthy {
		t.Errorf("Expected unhealthy status, got %s", dep.Status)
	}
	
	if dep.LatencyMs != 1000 {
		t.Errorf("Expected latency 1000ms, got %d", dep.LatencyMs)
	}
	
	if dep.Error == "" {
		t.Error("Expected error message")
	}
}

func BenchmarkHealthChecker_Check(b *testing.B) {
	checker := NewHealthChecker(nil, nil, "TestService", "1.0.0")
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		checker.Check()
	}
}

func BenchmarkHealthChecker_GetSystemMetrics(b *testing.B) {
	checker := NewHealthChecker(nil, nil, "TestService", "1.0.0")
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		checker.getSystemMetrics()
	}
}

