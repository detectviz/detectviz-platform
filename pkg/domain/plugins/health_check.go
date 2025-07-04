package plugins

import (
	"context"
	"time"
)

// HealthStatus 表示插件的健康狀態
type HealthStatus string

const (
	HealthStatusHealthy   HealthStatus = "healthy"
	HealthStatusUnhealthy HealthStatus = "unhealthy"
	HealthStatusDegraded  HealthStatus = "degraded"
	HealthStatusUnknown   HealthStatus = "unknown"
)

// HealthCheckResult 表示健康檢查的結果
type HealthCheckResult struct {
	Status      HealthStatus           `json:"status"`
	Message     string                 `json:"message,omitempty"`
	Details     map[string]interface{} `json:"details,omitempty"`
	LastChecked time.Time              `json:"last_checked"`
	Duration    time.Duration          `json:"duration"`
}

// HealthChecker 定義了插件健康檢查的介面
// 職責: 提供插件健康狀態的檢查和報告功能
// AI_PLUGIN_TYPE: "health_checker"
// AI_IMPL_PACKAGE: "detectviz-platform/internal/adapters/plugins/health"
// AI_IMPL_CONSTRUCTOR: "NewHealthChecker"
type HealthChecker interface {
	// HealthCheck 執行健康檢查並返回結果
	HealthCheck(ctx context.Context) HealthCheckResult

	// GetHealthCheckInterval 返回建議的健康檢查間隔
	GetHealthCheckInterval() time.Duration

	// IsHealthy 快速檢查插件是否健康
	IsHealthy(ctx context.Context) bool
}

// HealthCheckablePlugin 表示支援健康檢查的插件
// 插件可以選擇性地實現此介面來提供健康檢查功能
type HealthCheckablePlugin interface {
	Plugin
	HealthChecker
}

// DefaultHealthCheckResult 創建一個預設的健康檢查結果
func DefaultHealthCheckResult(status HealthStatus, message string) HealthCheckResult {
	return HealthCheckResult{
		Status:      status,
		Message:     message,
		LastChecked: time.Now(),
		Duration:    0,
	}
}

// NewHealthyResult 創建一個健康狀態的結果
func NewHealthyResult(message string) HealthCheckResult {
	return DefaultHealthCheckResult(HealthStatusHealthy, message)
}

// NewUnhealthyResult 創建一個不健康狀態的結果
func NewUnhealthyResult(message string, details map[string]interface{}) HealthCheckResult {
	result := DefaultHealthCheckResult(HealthStatusUnhealthy, message)
	result.Details = details
	return result
}

// NewDegradedResult 創建一個降級狀態的結果
func NewDegradedResult(message string, details map[string]interface{}) HealthCheckResult {
	result := DefaultHealthCheckResult(HealthStatusDegraded, message)
	result.Details = details
	return result
}
