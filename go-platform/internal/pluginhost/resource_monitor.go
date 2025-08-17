package pluginhost

import (
	"context"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.uber.org/zap"
)

// PluginResourceMetrics 插件級別資源指標
type PluginResourceMetrics struct {
	// 基本統計
	TotalRequests  int64 // 總請求數
	ActiveRequests int64 // 活躍請求數
	TotalErrors    int64 // 總錯誤數
	TotalDuration  int64 // 總執行時間（毫秒）

	// 資源使用
	MemoryUsageBytes int64 // 記憶體使用量（位元組）
	GoroutineCount   int32 // Goroutine 數量
	ConnectionCount  int32 // 連線數量

	// 性能指標
	AvgResponseTimeMs float64 // 平均回應時間
	RequestsPerSecond float64 // 每秒請求數
	ErrorRate         float64 // 錯誤率

	// 時間戳
	LastUpdateTime int64 // 最後更新時間（Unix毫秒）
	StartTime      int64 // 插件啟動時間
}

// ResourceMonitor 資源監控管理器
type ResourceMonitor struct {
	mu            sync.RWMutex
	pluginMetrics map[string]*PluginResourceMetrics

	// OpenTelemetry 指標
	meter          metric.Meter
	requestsTotal  metric.Int64Counter
	requestsActive metric.Int64UpDownCounter
	errorRate      metric.Float64Gauge
	responseTime   metric.Float64Histogram
	memoryUsage    metric.Int64Gauge
	goroutineCount metric.Int64Gauge

	// 監控控制
	ctx      context.Context
	cancel   context.CancelFunc
	interval time.Duration
	wg       sync.WaitGroup
}

// NewResourceMonitor 創建新的資源監控器
func NewResourceMonitor(interval time.Duration) (*ResourceMonitor, error) {
	ctx, cancel := context.WithCancel(context.Background())

	meter := otel.Meter("detectviz.pluginhost.resource_monitor")

	// 初始化 OpenTelemetry 指標
	requestsTotal, err := meter.Int64Counter("plugin_requests_total",
		metric.WithDescription("插件總請求數"))
	if err != nil {
		cancel()
		return nil, err
	}

	requestsActive, err := meter.Int64UpDownCounter("plugin_requests_active",
		metric.WithDescription("插件活躍請求數"))
	if err != nil {
		cancel()
		return nil, err
	}

	errorRate, err := meter.Float64Gauge("plugin_error_rate",
		metric.WithDescription("插件錯誤率"))
	if err != nil {
		cancel()
		return nil, err
	}

	responseTime, err := meter.Float64Histogram("plugin_response_time_ms",
		metric.WithDescription("插件回應時間（毫秒）"),
		metric.WithUnit("ms"))
	if err != nil {
		cancel()
		return nil, err
	}

	memoryUsage, err := meter.Int64Gauge("plugin_memory_usage_bytes",
		metric.WithDescription("插件記憶體使用量（位元組）"))
	if err != nil {
		cancel()
		return nil, err
	}

	goroutineCount, err := meter.Int64Gauge("plugin_goroutine_count",
		metric.WithDescription("插件 Goroutine 數量"))
	if err != nil {
		cancel()
		return nil, err
	}

	rm := &ResourceMonitor{
		pluginMetrics:  make(map[string]*PluginResourceMetrics),
		meter:          meter,
		requestsTotal:  requestsTotal,
		requestsActive: requestsActive,
		errorRate:      errorRate,
		responseTime:   responseTime,
		memoryUsage:    memoryUsage,
		goroutineCount: goroutineCount,
		ctx:            ctx,
		cancel:         cancel,
		interval:       interval,
	}

	// 啟動監控 goroutine
	rm.start()

	return rm, nil
}

// RegisterPlugin 註冊插件監控
func (rm *ResourceMonitor) RegisterPlugin(pluginID string) {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	now := time.Now().UnixMilli()
	rm.pluginMetrics[pluginID] = &PluginResourceMetrics{
		StartTime:      now,
		LastUpdateTime: now,
	}

	zap.L().Info("插件監控已註冊",
		zap.String("plugin_id", pluginID),
		zap.Int64("start_time", now))
}

// UnregisterPlugin 取消註冊插件監控
func (rm *ResourceMonitor) UnregisterPlugin(pluginID string) {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	delete(rm.pluginMetrics, pluginID)

	zap.L().Info("插件監控已取消註冊",
		zap.String("plugin_id", pluginID))
}

// RecordRequest 記錄請求開始
func (rm *ResourceMonitor) RecordRequest(pluginID string) {
	rm.mu.RLock()
	metrics, exists := rm.pluginMetrics[pluginID]
	rm.mu.RUnlock()

	if !exists {
		return
	}

	atomic.AddInt64(&metrics.TotalRequests, 1)
	atomic.AddInt64(&metrics.ActiveRequests, 1)

	// 更新 OpenTelemetry 指標
	rm.requestsTotal.Add(rm.ctx, 1,
		metric.WithAttributes(attribute.String("plugin_id", pluginID)))
	rm.requestsActive.Add(rm.ctx, 1,
		metric.WithAttributes(attribute.String("plugin_id", pluginID)))
}

// RecordRequestComplete 記錄請求完成
func (rm *ResourceMonitor) RecordRequestComplete(pluginID string, durationMs int64, isError bool) {
	rm.mu.RLock()
	metrics, exists := rm.pluginMetrics[pluginID]
	rm.mu.RUnlock()

	if !exists {
		return
	}

	atomic.AddInt64(&metrics.ActiveRequests, -1)
	atomic.AddInt64(&metrics.TotalDuration, durationMs)

	if isError {
		atomic.AddInt64(&metrics.TotalErrors, 1)
	}

	// 更新平均回應時間和錯誤率（需要併發保護）
	totalReqs := atomic.LoadInt64(&metrics.TotalRequests)
	if totalReqs > 0 {
		totalDur := atomic.LoadInt64(&metrics.TotalDuration)
		totalErrors := atomic.LoadInt64(&metrics.TotalErrors)

		avgResponseTime := float64(totalDur) / float64(totalReqs)
		errorRate := float64(totalErrors) / float64(totalReqs) * 100

		// 使用鎖保護 float64 併發寫入
		rm.mu.Lock()
		metrics.AvgResponseTimeMs = avgResponseTime
		metrics.ErrorRate = errorRate
		rm.mu.Unlock()
	}

	attrs := metric.WithAttributes(attribute.String("plugin_id", pluginID))

	// 更新 OpenTelemetry 指標
	rm.requestsActive.Add(rm.ctx, -1, attrs)
	rm.responseTime.Record(rm.ctx, float64(durationMs), attrs)
	rm.errorRate.Record(rm.ctx, metrics.ErrorRate, attrs)
}

// UpdateResourceUsage 更新資源使用情況
func (rm *ResourceMonitor) UpdateResourceUsage(pluginID string, memoryBytes int64, goroutines int32, connections int32) {
	rm.mu.RLock()
	metrics, exists := rm.pluginMetrics[pluginID]
	rm.mu.RUnlock()

	if !exists {
		return
	}

	atomic.StoreInt64(&metrics.MemoryUsageBytes, memoryBytes)
	atomic.StoreInt32(&metrics.GoroutineCount, goroutines)
	atomic.StoreInt32(&metrics.ConnectionCount, connections)
	atomic.StoreInt64(&metrics.LastUpdateTime, time.Now().UnixMilli())

	attrs := metric.WithAttributes(attribute.String("plugin_id", pluginID))

	// 更新 OpenTelemetry 指標
	rm.memoryUsage.Record(rm.ctx, memoryBytes, attrs)
	rm.goroutineCount.Record(rm.ctx, int64(goroutines), attrs)
}

// GetPluginMetrics 獲取插件指標
func (rm *ResourceMonitor) GetPluginMetrics(pluginID string) *PluginResourceMetrics {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	metrics, exists := rm.pluginMetrics[pluginID]
	if !exists {
		return nil
	}

	// 返回副本，避免併發修改
	return &PluginResourceMetrics{
		TotalRequests:     atomic.LoadInt64(&metrics.TotalRequests),
		ActiveRequests:    atomic.LoadInt64(&metrics.ActiveRequests),
		TotalErrors:       atomic.LoadInt64(&metrics.TotalErrors),
		TotalDuration:     atomic.LoadInt64(&metrics.TotalDuration),
		MemoryUsageBytes:  atomic.LoadInt64(&metrics.MemoryUsageBytes),
		GoroutineCount:    atomic.LoadInt32(&metrics.GoroutineCount),
		ConnectionCount:   atomic.LoadInt32(&metrics.ConnectionCount),
		AvgResponseTimeMs: metrics.AvgResponseTimeMs,
		RequestsPerSecond: metrics.RequestsPerSecond,
		ErrorRate:         metrics.ErrorRate,
		LastUpdateTime:    atomic.LoadInt64(&metrics.LastUpdateTime),
		StartTime:         metrics.StartTime,
	}
}

// GetAllMetrics 獲取所有插件指標
func (rm *ResourceMonitor) GetAllMetrics() map[string]*PluginResourceMetrics {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	result := make(map[string]*PluginResourceMetrics, len(rm.pluginMetrics))
	for pluginID := range rm.pluginMetrics {
		result[pluginID] = rm.GetPluginMetrics(pluginID)
	}

	return result
}

// start 啟動監控循環
func (rm *ResourceMonitor) start() {
	rm.wg.Add(1)
	go func() {
		defer rm.wg.Done()

		ticker := time.NewTicker(rm.interval)
		defer ticker.Stop()

		for {
			select {
			case <-rm.ctx.Done():
				return
			case <-ticker.C:
				rm.collectSystemMetrics()
			}
		}
	}()
}

// collectSystemMetrics 收集系統級指標
func (rm *ResourceMonitor) collectSystemMetrics() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	goroutines := runtime.NumGoroutine()

	rm.mu.RLock()
	pluginIDs := make([]string, 0, len(rm.pluginMetrics))
	for pluginID := range rm.pluginMetrics {
		pluginIDs = append(pluginIDs, pluginID)
	}
	rm.mu.RUnlock()

	// 更新每個插件的系統資源指標
	for _, pluginID := range pluginIDs {
		rm.mu.RLock()
		metrics, exists := rm.pluginMetrics[pluginID]
		rm.mu.RUnlock()

		if !exists {
			continue
		}

		now := time.Now()
		lastUpdate := atomic.LoadInt64(&metrics.LastUpdateTime)

		// 計算 RPS（基於最近的更新間隔）
		if lastUpdate > 0 {
			timeDiffSec := float64(now.UnixMilli()-lastUpdate) / 1000.0
			if timeDiffSec > 0 {
				activeReqs := atomic.LoadInt64(&metrics.ActiveRequests)
				// 使用 atomic 操作更新 RPS
				rps := float64(activeReqs) / timeDiffSec
				// 由於 float64 的併發寫入不是原子的，我們需要保護它
				rm.mu.Lock()
				metrics.RequestsPerSecond = rps
				rm.mu.Unlock()
			}
		}

		// 估算插件級記憶體使用（簡化實作）
		totalPlugins := len(pluginIDs)
		if totalPlugins > 0 {
			estimatedMemory := int64(m.Alloc) / int64(totalPlugins)
			rm.UpdateResourceUsage(pluginID, estimatedMemory, int32(goroutines/totalPlugins), 0)
		}
	}
}

// Stop 停止監控器
func (rm *ResourceMonitor) Stop() {
	rm.cancel()
	rm.wg.Wait()

	zap.L().Info("資源監控器已停止")
}

// GetHealthStatus 獲取監控健康狀態
func (rm *ResourceMonitor) GetHealthStatus() map[string]interface{} {
	allMetrics := rm.GetAllMetrics()

	totalPlugins := len(allMetrics)
	totalActiveRequests := int64(0)
	totalErrors := int64(0)
	totalRequests := int64(0)

	for _, metrics := range allMetrics {
		totalActiveRequests += metrics.ActiveRequests
		totalErrors += metrics.TotalErrors
		totalRequests += metrics.TotalRequests
	}

	overallErrorRate := float64(0)
	if totalRequests > 0 {
		overallErrorRate = float64(totalErrors) / float64(totalRequests) * 100
	}

	return map[string]interface{}{
		"total_plugins":         totalPlugins,
		"total_active_requests": totalActiveRequests,
		"total_requests":        totalRequests,
		"total_errors":          totalErrors,
		"overall_error_rate":    overallErrorRate,
		"monitoring_interval":   rm.interval.String(),
		"last_collection_time":  time.Now().Format(time.RFC3339),
	}
}
