package pluginhost

import (
	"context"
	"runtime"
	"sync/atomic"
	"time"

	pb "github.com/detectviz/detectviz-platform/contracts/gen/go/detectviz/contracts/v1"
	"go.uber.org/zap"
)

// MonitoredHandler 具備資源監控功能的處理器包裝器
type MonitoredHandler struct {
	pluginID      string
	handler       Handler
	monitor       *ResourceMonitor
	
	// 資源計數器
	activeRequests int64
	goroutineBase  int
	memoryBase     uint64
}

// NewMonitoredHandler 創建新的監控處理器
func NewMonitoredHandler(pluginID string, handler Handler, monitor *ResourceMonitor) *MonitoredHandler {
	// 記錄基線指標
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	
	mh := &MonitoredHandler{
		pluginID:      pluginID,
		handler:       handler,
		monitor:       monitor,
		goroutineBase: runtime.NumGoroutine(),
		memoryBase:    memStats.Alloc,
	}
	
	// 註冊到監控系統
	monitor.RegisterPlugin(pluginID)
	
	zap.L().Info("監控處理器已創建",
		zap.String("plugin_id", pluginID),
		zap.Int("baseline_goroutines", mh.goroutineBase),
		zap.Uint64("baseline_memory_bytes", mh.memoryBase))
	
	return mh
}

// Invoke 執行請求並進行資源監控
func (mh *MonitoredHandler) Invoke(ctx context.Context, req *pb.ToolInvokeRequest) (*pb.ToolInvokeReply, error) {
	// 記錄請求開始
	start := time.Now()
	atomic.AddInt64(&mh.activeRequests, 1)
	mh.monitor.RecordRequest(mh.pluginID)
	
	// 採集資源使用情況
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	currentGoroutines := runtime.NumGoroutine()
	
	// 計算增量（相對於基線）
	goroutineDelta := int32(currentGoroutines - mh.goroutineBase)
	memoryDelta := int64(memStats.Alloc - mh.memoryBase)
	activeConns := int32(0) // TODO: 如果插件支援連接計數，可以添加
	
	// 更新資源監控
	mh.monitor.UpdateResourceUsage(mh.pluginID, memoryDelta, goroutineDelta, activeConns)
	
	// 執行實際處理邏輯
	reply, err := mh.handler.Invoke(ctx, req)
	
	// 記錄請求完成
	duration := time.Since(start)
	durationMs := duration.Milliseconds()
	isError := err != nil || (reply != nil && reply.Status != nil && reply.Status.Code != 0)
	
	atomic.AddInt64(&mh.activeRequests, -1)
	mh.monitor.RecordRequestComplete(mh.pluginID, durationMs, isError)
	
	// 記錄詳細日誌
	logger := zap.L().With(
		zap.String("plugin_id", mh.pluginID),
		zap.Duration("duration", duration),
		zap.Int64("memory_delta_bytes", memoryDelta),
		zap.Int32("goroutine_delta", goroutineDelta),
		zap.Bool("is_error", isError))
	
	if isError {
		logger.Warn("插件請求執行失敗",
			zap.Error(err))
	} else {
		logger.Debug("插件請求執行成功")
	}
	
	return reply, err
}

// Close 實作資源釋放
func (mh *MonitoredHandler) Close() error {
	// 等待所有活躍請求完成
	for atomic.LoadInt64(&mh.activeRequests) > 0 {
		time.Sleep(10 * time.Millisecond)
	}
	
	// 取消監控註冊
	mh.monitor.UnregisterPlugin(mh.pluginID)
	
	// 關閉底層處理器
	if closable, ok := mh.handler.(ClosableHandler); ok {
		return closable.Close()
	}
	
	zap.L().Info("監控處理器已關閉",
		zap.String("plugin_id", mh.pluginID))
	
	return nil
}

// GetActiveRequests 獲取活躍請求數量
func (mh *MonitoredHandler) GetActiveRequests() int64 {
	return atomic.LoadInt64(&mh.activeRequests)
}

// ResourceAwareHandler 支援資源感知的處理器接口
type ResourceAwareHandler interface {
	Handler
	
	// GetResourceUsage 返回當前資源使用情況
	GetResourceUsage() (memoryBytes int64, goroutines int32, connections int32)
	
	// SetResourceLimits 設置資源限制
	SetResourceLimits(maxMemoryBytes int64, maxGoroutines int32, maxConnections int32) error
}

// EnhancedMonitoredHandler 增強型監控處理器（支援資源感知）
type EnhancedMonitoredHandler struct {
	*MonitoredHandler
	resourceAware ResourceAwareHandler
}

// NewEnhancedMonitoredHandler 創建增強型監控處理器
func NewEnhancedMonitoredHandler(pluginID string, handler ResourceAwareHandler, monitor *ResourceMonitor) *EnhancedMonitoredHandler {
	base := NewMonitoredHandler(pluginID, handler, monitor)
	return &EnhancedMonitoredHandler{
		MonitoredHandler: base,
		resourceAware:    handler,
	}
}

// Invoke 執行增強型監控請求
func (emh *EnhancedMonitoredHandler) Invoke(ctx context.Context, req *pb.ToolInvokeRequest) (*pb.ToolInvokeReply, error) {
	// 使用插件自身的資源統計
	memBytes, goroutines, connections := emh.resourceAware.GetResourceUsage()
	emh.monitor.UpdateResourceUsage(emh.pluginID, memBytes, goroutines, connections)
	
	// 執行基礎監控邏輯
	return emh.MonitoredHandler.Invoke(ctx, req)
}

// SetResourceLimits 設置資源限制
func (emh *EnhancedMonitoredHandler) SetResourceLimits(maxMemoryBytes int64, maxGoroutines int32, maxConnections int32) error {
	return emh.resourceAware.SetResourceLimits(maxMemoryBytes, maxGoroutines, maxConnections)
}

// GetDetailedMetrics 獲取詳細的監控指標
func (emh *EnhancedMonitoredHandler) GetDetailedMetrics() map[string]interface{} {
	baseMetrics := emh.monitor.GetPluginMetrics(emh.pluginID)
	if baseMetrics == nil {
		return nil
	}
	
	memBytes, goroutines, connections := emh.resourceAware.GetResourceUsage()
	
	return map[string]interface{}{
		"plugin_id":           emh.pluginID,
		"total_requests":      baseMetrics.TotalRequests,
		"active_requests":     baseMetrics.ActiveRequests,
		"total_errors":        baseMetrics.TotalErrors,
		"avg_response_time":   baseMetrics.AvgResponseTimeMs,
		"requests_per_second": baseMetrics.RequestsPerSecond,
		"error_rate_percent":  baseMetrics.ErrorRate,
		"memory_usage_bytes":  memBytes,
		"goroutine_count":     goroutines,
		"connection_count":    connections,
		"uptime_ms":          time.Now().UnixMilli() - baseMetrics.StartTime,
		"last_update":        baseMetrics.LastUpdateTime,
	}
}