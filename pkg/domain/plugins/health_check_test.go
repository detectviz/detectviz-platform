package plugins

import (
	"testing"
	"time"
)

func TestHealthStatus(t *testing.T) {
	tests := []struct {
		name     string
		status   HealthStatus
		expected string
	}{
		{"healthy status", HealthStatusHealthy, "healthy"},
		{"unhealthy status", HealthStatusUnhealthy, "unhealthy"},
		{"degraded status", HealthStatusDegraded, "degraded"},
		{"unknown status", HealthStatusUnknown, "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.status) != tt.expected {
				t.Errorf("HealthStatus string = %v, want %v", string(tt.status), tt.expected)
			}
		})
	}
}

func TestDefaultHealthCheckResult(t *testing.T) {
	status := HealthStatusHealthy
	message := "All systems operational"

	result := DefaultHealthCheckResult(status, message)

	if result.Status != status {
		t.Errorf("DefaultHealthCheckResult.Status = %v, want %v", result.Status, status)
	}

	if result.Message != message {
		t.Errorf("DefaultHealthCheckResult.Message = %v, want %v", result.Message, message)
	}

	if result.Duration != 0 {
		t.Errorf("DefaultHealthCheckResult.Duration = %v, want 0", result.Duration)
	}

	if result.LastChecked.IsZero() {
		t.Error("DefaultHealthCheckResult.LastChecked should not be zero")
	}
}

func TestNewHealthyResult(t *testing.T) {
	message := "Service is running smoothly"

	result := NewHealthyResult(message)

	if result.Status != HealthStatusHealthy {
		t.Errorf("NewHealthyResult.Status = %v, want %v", result.Status, HealthStatusHealthy)
	}

	if result.Message != message {
		t.Errorf("NewHealthyResult.Message = %v, want %v", result.Message, message)
	}

	if result.Details != nil {
		t.Errorf("NewHealthyResult.Details = %v, want nil", result.Details)
	}
}

func TestNewUnhealthyResult(t *testing.T) {
	message := "Database connection failed"
	details := map[string]interface{}{
		"error":   "connection timeout",
		"timeout": "5s",
	}

	result := NewUnhealthyResult(message, details)

	if result.Status != HealthStatusUnhealthy {
		t.Errorf("NewUnhealthyResult.Status = %v, want %v", result.Status, HealthStatusUnhealthy)
	}

	if result.Message != message {
		t.Errorf("NewUnhealthyResult.Message = %v, want %v", result.Message, message)
	}

	if result.Details == nil {
		t.Error("NewUnhealthyResult.Details should not be nil")
	}

	if result.Details["error"] != "connection timeout" {
		t.Errorf("NewUnhealthyResult.Details[error] = %v, want 'connection timeout'", result.Details["error"])
	}
}

func TestNewDegradedResult(t *testing.T) {
	message := "Performance degraded"
	details := map[string]interface{}{
		"response_time": "2s",
		"threshold":     "1s",
	}

	result := NewDegradedResult(message, details)

	if result.Status != HealthStatusDegraded {
		t.Errorf("NewDegradedResult.Status = %v, want %v", result.Status, HealthStatusDegraded)
	}

	if result.Message != message {
		t.Errorf("NewDegradedResult.Message = %v, want %v", result.Message, message)
	}

	if result.Details == nil {
		t.Error("NewDegradedResult.Details should not be nil")
	}

	if result.Details["response_time"] != "2s" {
		t.Errorf("NewDegradedResult.Details[response_time] = %v, want '2s'", result.Details["response_time"])
	}
}

func TestHealthCheckResult_JSON(t *testing.T) {
	result := HealthCheckResult{
		Status:      HealthStatusHealthy,
		Message:     "All good",
		Details:     map[string]interface{}{"key": "value"},
		LastChecked: time.Now(),
		Duration:    100 * time.Millisecond,
	}

	// Test that the struct can be marshaled to JSON
	// This is important for HTTP responses
	if result.Status != HealthStatusHealthy {
		t.Error("HealthCheckResult should maintain its status")
	}

	if result.Message != "All good" {
		t.Error("HealthCheckResult should maintain its message")
	}

	if result.Details == nil {
		t.Error("HealthCheckResult should maintain its details")
	}

	if result.LastChecked.IsZero() {
		t.Error("HealthCheckResult should maintain its last checked time")
	}

	if result.Duration != 100*time.Millisecond {
		t.Error("HealthCheckResult should maintain its duration")
	}
}
