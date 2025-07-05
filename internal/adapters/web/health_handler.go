package web

import (
	"net/http"
	"time"

	"detectviz-platform/internal/infrastructure/platform/health"
	"detectviz-platform/pkg/domain/interfaces/plugins"

	"github.com/labstack/echo/v4"
)

// HealthHandler 處理健康檢查相關的 HTTP 請求
type HealthHandler struct {
	healthManager *health.HealthCheckManager
}

// NewHealthHandler 創建新的健康檢查處理器
func NewHealthHandler(healthManager *health.HealthCheckManager) *HealthHandler {
	return &HealthHandler{
		healthManager: healthManager,
	}
}

// HealthResponse 健康檢查響應結構
type HealthResponse struct {
	Status    string                               `json:"status"`
	Timestamp time.Time                            `json:"timestamp"`
	Message   string                               `json:"message,omitempty"`
	Details   map[string]interface{}               `json:"details,omitempty"`
	Plugins   map[string]plugins.HealthCheckResult `json:"plugins,omitempty"`
}

// GetHealth 獲取整體健康狀態
func (h *HealthHandler) GetHealth(c echo.Context) error {
	overallStatus := h.healthManager.GetOverallHealthStatus()

	response := HealthResponse{
		Status:    string(overallStatus.Status),
		Timestamp: time.Now(),
		Message:   overallStatus.Message,
		Details:   overallStatus.Details,
	}

	// 根據健康狀態設置 HTTP 狀態碼
	var httpStatus int
	switch overallStatus.Status {
	case plugins.HealthStatusHealthy:
		httpStatus = http.StatusOK
	case plugins.HealthStatusDegraded:
		httpStatus = http.StatusOK // 降級但仍可用
	case plugins.HealthStatusUnhealthy:
		httpStatus = http.StatusServiceUnavailable
	default:
		httpStatus = http.StatusServiceUnavailable
	}

	return c.JSON(httpStatus, response)
}

// GetDetailedHealth 獲取詳細的健康狀態（包含所有插件）
func (h *HealthHandler) GetDetailedHealth(c echo.Context) error {
	overallStatus := h.healthManager.GetOverallHealthStatus()
	allPluginStatus := h.healthManager.GetAllHealthStatus()

	response := HealthResponse{
		Status:    string(overallStatus.Status),
		Timestamp: time.Now(),
		Message:   overallStatus.Message,
		Details:   overallStatus.Details,
		Plugins:   allPluginStatus,
	}

	// 根據健康狀態設置 HTTP 狀態碼
	var httpStatus int
	switch overallStatus.Status {
	case plugins.HealthStatusHealthy:
		httpStatus = http.StatusOK
	case plugins.HealthStatusDegraded:
		httpStatus = http.StatusOK // 降級但仍可用
	case plugins.HealthStatusUnhealthy:
		httpStatus = http.StatusServiceUnavailable
	default:
		httpStatus = http.StatusServiceUnavailable
	}

	return c.JSON(httpStatus, response)
}

// GetPluginHealth 獲取特定插件的健康狀態
func (h *HealthHandler) GetPluginHealth(c echo.Context) error {
	pluginName := c.Param("plugin")
	if pluginName == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "plugin name is required",
		})
	}

	status, exists := h.healthManager.GetHealthStatus(pluginName)
	if !exists {
		return c.JSON(http.StatusNotFound, map[string]string{
			"error": "plugin not found or not registered for health checking",
		})
	}

	response := HealthResponse{
		Status:    string(status.Status),
		Timestamp: time.Now(),
		Message:   status.Message,
		Details:   status.Details,
	}

	// 根據健康狀態設置 HTTP 狀態碼
	var httpStatus int
	switch status.Status {
	case plugins.HealthStatusHealthy:
		httpStatus = http.StatusOK
	case plugins.HealthStatusDegraded:
		httpStatus = http.StatusOK // 降級但仍可用
	case plugins.HealthStatusUnhealthy:
		httpStatus = http.StatusServiceUnavailable
	default:
		httpStatus = http.StatusServiceUnavailable
	}

	return c.JSON(httpStatus, response)
}

// RegisterRoutes 註冊健康檢查路由
func (h *HealthHandler) RegisterRoutes(e *echo.Echo) {
	healthGroup := e.Group("/health")

	// 基本健康檢查端點
	healthGroup.GET("", h.GetHealth)
	healthGroup.GET("/", h.GetHealth)

	// 詳細健康檢查端點
	healthGroup.GET("/detailed", h.GetDetailedHealth)

	// 特定插件健康檢查端點
	healthGroup.GET("/plugin/:plugin", h.GetPluginHealth)
}
