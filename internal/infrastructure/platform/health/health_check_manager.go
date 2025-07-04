package health

import (
	"context"
	"fmt"
	"sync"
	"time"

	"detectviz-platform/pkg/domain/plugins"
	"detectviz-platform/pkg/platform/contracts"
)

// HealthCheckManager 管理所有插件的健康檢查
// AI_PLUGIN_TYPE: "health_check_manager"
// AI_IMPL_PACKAGE: "detectviz-platform/internal/infrastructure/platform/health"
// AI_IMPL_CONSTRUCTOR: "NewHealthCheckManager"
type HealthCheckManager struct {
	plugins       map[string]plugins.HealthCheckablePlugin
	results       map[string]plugins.HealthCheckResult
	logger        contracts.Logger
	mu            sync.RWMutex
	ticker        *time.Ticker
	stopChan      chan struct{}
	checkInterval time.Duration
}

// NewHealthCheckManager 創建新的健康檢查管理器
func NewHealthCheckManager(logger contracts.Logger, checkInterval time.Duration) *HealthCheckManager {
	if checkInterval <= 0 {
		checkInterval = 30 * time.Second // 預設 30 秒檢查一次
	}

	return &HealthCheckManager{
		plugins:       make(map[string]plugins.HealthCheckablePlugin),
		results:       make(map[string]plugins.HealthCheckResult),
		logger:        logger,
		checkInterval: checkInterval,
		stopChan:      make(chan struct{}),
	}
}

// RegisterPlugin 註冊一個支援健康檢查的插件
func (hm *HealthCheckManager) RegisterPlugin(pluginName string, plugin plugins.HealthCheckablePlugin) {
	hm.mu.Lock()
	defer hm.mu.Unlock()

	hm.plugins[pluginName] = plugin
	hm.logger.Info("Registered plugin for health checking", "plugin", pluginName)
}

// UnregisterPlugin 取消註冊插件
func (hm *HealthCheckManager) UnregisterPlugin(pluginName string) {
	hm.mu.Lock()
	defer hm.mu.Unlock()

	delete(hm.plugins, pluginName)
	delete(hm.results, pluginName)
	hm.logger.Info("Unregistered plugin from health checking", "plugin", pluginName)
}

// Start 開始健康檢查監控
func (hm *HealthCheckManager) Start(ctx context.Context) error {
	hm.mu.Lock()
	defer hm.mu.Unlock()

	if hm.ticker != nil {
		return fmt.Errorf("health check manager already started")
	}

	hm.ticker = time.NewTicker(hm.checkInterval)
	hm.logger.Info("Starting health check manager", "interval", hm.checkInterval)

	go hm.runHealthChecks(ctx)

	return nil
}

// Stop 停止健康檢查監控
func (hm *HealthCheckManager) Stop(ctx context.Context) error {
	hm.mu.Lock()
	defer hm.mu.Unlock()

	if hm.ticker == nil {
		return fmt.Errorf("health check manager not started")
	}

	hm.ticker.Stop()
	hm.ticker = nil

	close(hm.stopChan)
	hm.stopChan = make(chan struct{})

	hm.logger.Info("Stopped health check manager")
	return nil
}

// runHealthChecks 執行健康檢查循環
func (hm *HealthCheckManager) runHealthChecks(ctx context.Context) {
	// 立即執行一次健康檢查
	hm.performHealthChecks(ctx)

	for {
		select {
		case <-hm.ticker.C:
			hm.performHealthChecks(ctx)
		case <-hm.stopChan:
			hm.logger.Info("Health check loop stopped")
			return
		case <-ctx.Done():
			hm.logger.Info("Health check loop cancelled")
			return
		}
	}
}

// performHealthChecks 執行所有插件的健康檢查
func (hm *HealthCheckManager) performHealthChecks(ctx context.Context) {
	hm.mu.RLock()
	pluginsCopy := make(map[string]plugins.HealthCheckablePlugin)
	for name, plugin := range hm.plugins {
		pluginsCopy[name] = plugin
	}
	hm.mu.RUnlock()

	for pluginName, plugin := range pluginsCopy {
		go func(name string, p plugins.HealthCheckablePlugin) {
			start := time.Now()
			result := p.HealthCheck(ctx)
			result.Duration = time.Since(start)
			result.LastChecked = time.Now()

			hm.mu.Lock()
			hm.results[name] = result
			hm.mu.Unlock()

			// 記錄不健康的插件
			if result.Status != plugins.HealthStatusHealthy {
				hm.logger.Warn("Plugin health check failed",
					"plugin", name,
					"status", result.Status,
					"message", result.Message,
					"duration", result.Duration)
			} else {
				hm.logger.Debug("Plugin health check passed",
					"plugin", name,
					"duration", result.Duration)
			}
		}(pluginName, plugin)
	}
}

// GetHealthStatus 獲取指定插件的健康狀態
func (hm *HealthCheckManager) GetHealthStatus(pluginName string) (plugins.HealthCheckResult, bool) {
	hm.mu.RLock()
	defer hm.mu.RUnlock()

	result, exists := hm.results[pluginName]
	return result, exists
}

// GetAllHealthStatus 獲取所有插件的健康狀態
func (hm *HealthCheckManager) GetAllHealthStatus() map[string]plugins.HealthCheckResult {
	hm.mu.RLock()
	defer hm.mu.RUnlock()

	results := make(map[string]plugins.HealthCheckResult)
	for name, result := range hm.results {
		results[name] = result
	}

	return results
}

// GetOverallHealthStatus 獲取整體健康狀態
func (hm *HealthCheckManager) GetOverallHealthStatus() plugins.HealthCheckResult {
	hm.mu.RLock()
	defer hm.mu.RUnlock()

	if len(hm.results) == 0 {
		return plugins.DefaultHealthCheckResult(plugins.HealthStatusUnknown, "No plugins registered")
	}

	healthyCount := 0
	unhealthyCount := 0
	degradedCount := 0

	for _, result := range hm.results {
		switch result.Status {
		case plugins.HealthStatusHealthy:
			healthyCount++
		case plugins.HealthStatusUnhealthy:
			unhealthyCount++
		case plugins.HealthStatusDegraded:
			degradedCount++
		}
	}

	totalCount := len(hm.results)

	// 決定整體狀態
	if unhealthyCount > 0 {
		return plugins.NewUnhealthyResult(
			fmt.Sprintf("%d of %d plugins are unhealthy", unhealthyCount, totalCount),
			map[string]interface{}{
				"healthy_count":   healthyCount,
				"unhealthy_count": unhealthyCount,
				"degraded_count":  degradedCount,
				"total_count":     totalCount,
			})
	}

	if degradedCount > 0 {
		return plugins.NewDegradedResult(
			fmt.Sprintf("%d of %d plugins are degraded", degradedCount, totalCount),
			map[string]interface{}{
				"healthy_count":   healthyCount,
				"unhealthy_count": unhealthyCount,
				"degraded_count":  degradedCount,
				"total_count":     totalCount,
			})
	}

	return plugins.NewHealthyResult(
		fmt.Sprintf("All %d plugins are healthy", totalCount))
}

// GetName 返回服務名稱
func (hm *HealthCheckManager) GetName() string {
	return "health_check_manager"
}
