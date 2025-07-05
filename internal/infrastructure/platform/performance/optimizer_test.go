package performance

import (
	"context"
	"testing"

	"detectviz-platform/internal/infrastructure/platform/telemetry"
)

func TestNewPerformanceOptimizer(t *testing.T) {
	mockLogger := telemetry.NewOtelZapLogger(map[string]interface{}{
		"level":    "info",
		"encoding": "json",
	})

	optimizer := NewPerformanceOptimizer(mockLogger, 10)

	if optimizer == nil {
		t.Error("NewPerformanceOptimizer should return a non-nil optimizer")
	}

	if optimizer.GetName() != "performance_optimizer" {
		t.Errorf("Expected name 'performance_optimizer', got '%s'", optimizer.GetName())
	}
}

func TestPerformanceOptimizer_OptimizeSystem(t *testing.T) {
	mockLogger := telemetry.NewOtelZapLogger(map[string]interface{}{
		"level":    "info",
		"encoding": "json",
	})

	optimizer := NewPerformanceOptimizer(mockLogger, 10)
	ctx := context.Background()

	err := optimizer.OptimizeSystem(ctx)
	if err != nil {
		t.Errorf("OptimizeSystem should not return error, got: %v", err)
	}
}
