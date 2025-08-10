package pluginhost

import (
	"context"
	"errors"
	"fmt"
	"io"
	"sync"
	"time"

	contractspb "github.com/detectviz/detectviz-platform/contracts/gen/go/detectviz/contracts/v1"
	"go.uber.org/zap"
)

// Handler 為每個 tool_id 的處理器介面
type Handler interface {
	Invoke(ctx context.Context, req *contractspb.ToolInvokeRequest) (*contractspb.ToolInvokeReply, error)
}

// 可選：具備釋放資源能力的處理器
// 若插件內部持有連線池/檔案/背景 goroutine，請實作 Close() 以便 Registry 在卸載或替換時釋放資源。
type ClosableHandler interface {
	Handler
	Close() error
}

// Registry 維護 tool_id → Handler 的映射（並發安全），支援資源監控
type Registry struct {
	rwmu            sync.RWMutex
	byToolID        map[string]Handler
	resourceMonitor *ResourceMonitor // 可選的資源監控器
}

func NewRegistry() *Registry {
	return &Registry{byToolID: map[string]Handler{}}
}

// NewRegistryWithMonitoring 創建帶資源監控的註冊表
func NewRegistryWithMonitoring(monitorInterval time.Duration) (*Registry, error) {
	monitor, err := NewResourceMonitor(monitorInterval)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource monitor: %w", err)
	}
	
	return &Registry{
		byToolID:        map[string]Handler{},
		resourceMonitor: monitor,
	}, nil
}

// ErrHandlerExists 代表在嚴格模式下該 toolID 已存在
var ErrHandlerExists = errors.New("handler already registered")

// RegisterStrict 嚴格註冊：若 toolID 已存在則直接回報錯誤，不覆蓋、也不關閉舊 handler。
// 適用於不允許同名覆蓋的場景，可避免無意間的替換造成語義混亂。
func (r *Registry) RegisterStrict(toolID string, h Handler) error {
	r.rwmu.Lock()
	defer r.rwmu.Unlock()
	
	if _, ok := r.byToolID[toolID]; ok {
		return fmt.Errorf("%w: tool_id=%s", ErrHandlerExists, toolID)
	}
	
	// 如果啟用監控，包裝處理器
	actualHandler := r.wrapWithMonitoring(toolID, h)
	r.byToolID[toolID] = actualHandler
	
	zap.L().Info("Handler registered (strict)", 
		zap.String("tool_id", toolID),
		zap.Bool("monitoring_enabled", r.resourceMonitor != nil))
	return nil
}

// Register 登錄或覆蓋指定 toolID 的處理器（寫鎖）
// - 若同一 toolID 已存在，會嘗試關閉舊的 handler 以避免資源洩漏，然後再覆蓋。
func (r *Registry) Register(toolID string, h Handler) {
	r.rwmu.Lock()
	defer r.rwmu.Unlock()
	
	if old, ok := r.byToolID[toolID]; ok && old != h {
		if err := closeIfPossible(old); err != nil {
			zap.L().Warn("Failed to close previous handler on re-register", 
				zap.String("tool_id", toolID), zap.Error(err))
		}
	}
	
	// 如果啟用監控，包裝處理器
	actualHandler := r.wrapWithMonitoring(toolID, h)
	r.byToolID[toolID] = actualHandler
	
	zap.L().Info("Handler registered", 
		zap.String("tool_id", toolID),
		zap.Bool("monitoring_enabled", r.resourceMonitor != nil))
}

// RegisterOrReplace 與 Register 相同語意；提供更語義化的名稱以利閱讀。
func (r *Registry) RegisterOrReplace(toolID string, h Handler) {
	r.Register(toolID, h)
}

// Lookup 讀取指定 toolID 的處理器（讀鎖）
func (r *Registry) Lookup(toolID string) (Handler, bool) {
	r.rwmu.RLock()
	h, ok := r.byToolID[toolID]
	r.rwmu.RUnlock()
	return h, ok
}

// Unregister 解除登錄，並嘗試釋放資源（寫鎖）
// 回傳是否存在且已移除。
func (r *Registry) Unregister(toolID string) bool {
	r.rwmu.Lock()
	old, ok := r.byToolID[toolID]
	if ok {
		delete(r.byToolID, toolID)
	}
	r.rwmu.Unlock()

	if ok {
		if err := closeIfPossible(old); err != nil {
			zap.L().Warn("Failed to close handler on unregister", zap.String("tool_id", toolID), zap.Error(err))
		} else {
			zap.L().Info("Handler unregistered", zap.String("tool_id", toolID))
		}
	}
	return ok
}

// List 取得目前已註冊的所有 toolID 快照（讀鎖）
func (r *Registry) List() []string {
	r.rwmu.RLock()
	ids := make([]string, 0, len(r.byToolID))
	for id := range r.byToolID {
		ids = append(ids, id)
	}
	r.rwmu.RUnlock()
	return ids
}

// Size 目前註冊數量（讀鎖）
func (r *Registry) Size() int {
	r.rwmu.RLock()
	n := len(r.byToolID)
	r.rwmu.RUnlock()
	return n
}

// Shutdown 關閉所有可關閉的 handler 並清空註冊表（寫鎖 + 批次釋放）
func (r *Registry) Shutdown() error {
	r.rwmu.Lock()
	defer r.rwmu.Unlock()

	var errs error
	for id, h := range r.byToolID {
		if err := closeIfPossible(h); err != nil {
			errs = errors.Join(errs, fmt.Errorf("tool_id=%s: %w", id, err))
		}
		delete(r.byToolID, id)
	}
	
	// 關閉資源監控器
	if r.resourceMonitor != nil {
		r.resourceMonitor.Stop()
	}
	
	if errs != nil {
		zap.L().Warn("Registry shutdown with errors", zap.Error(errs))
	} else {
		zap.L().Info("Registry shutdown complete")
	}
	return errs
}

// closeIfPossible 嘗試呼叫 Close()；支援 ClosableHandler 或 io.Closer
func closeIfPossible(h Handler) error {
	switch x := h.(type) {
	case ClosableHandler:
		return x.Close()
	case io.Closer:
		return x.Close()
	default:
		return nil
	}
}

// wrapWithMonitoring 根據監控配置包裝處理器
func (r *Registry) wrapWithMonitoring(toolID string, h Handler) Handler {
	if r.resourceMonitor == nil {
		return h // 沒有監控器，直接回傳原始處理器
	}
	
	// 檢查處理器是否支援資源感知
	if resourceAware, ok := h.(ResourceAwareHandler); ok {
		// 使用增強型監控包裝器
		return NewEnhancedMonitoredHandler(toolID, resourceAware, r.resourceMonitor)
	} else {
		// 使用基本監控包裝器
		return NewMonitoredHandler(toolID, h, r.resourceMonitor)
	}
}

// GetResourceMonitor 獲取資源監控器實例
func (r *Registry) GetResourceMonitor() *ResourceMonitor {
	r.rwmu.RLock()
	defer r.rwmu.RUnlock()
	return r.resourceMonitor
}

// GetPluginMetrics 獲取指定插件的監控指標
func (r *Registry) GetPluginMetrics(toolID string) *PluginResourceMetrics {
	if r.resourceMonitor == nil {
		return nil
	}
	return r.resourceMonitor.GetPluginMetrics(toolID)
}

// GetAllPluginMetrics 獲取所有插件的監控指標
func (r *Registry) GetAllPluginMetrics() map[string]*PluginResourceMetrics {
	if r.resourceMonitor == nil {
		return nil
	}
	return r.resourceMonitor.GetAllMetrics()
}

// GetMonitoringHealthStatus 獲取監控系統健康狀態
func (r *Registry) GetMonitoringHealthStatus() map[string]interface{} {
	if r.resourceMonitor == nil {
		return map[string]interface{}{
			"monitoring_enabled": false,
			"message":           "資源監控未啟用",
		}
	}
	
	health := r.resourceMonitor.GetHealthStatus()
	health["monitoring_enabled"] = true
	return health
}

// SetPluginResourceLimits 為指定插件設置資源限制
func (r *Registry) SetPluginResourceLimits(toolID string, maxMemoryBytes int64, maxGoroutines int32, maxConnections int32) error {
	r.rwmu.RLock()
	handler, exists := r.byToolID[toolID]
	r.rwmu.RUnlock()
	
	if !exists {
		return fmt.Errorf("plugin not found: %s", toolID)
	}
	
	// 嘗試找到增強監控處理器
	if enhancedHandler, ok := handler.(*EnhancedMonitoredHandler); ok {
		return enhancedHandler.SetResourceLimits(maxMemoryBytes, maxGoroutines, maxConnections)
	}
	
	// 嘗試直接訪問資源感知處理器（可能是原始處理器）
	if monitoredHandler, ok := handler.(*MonitoredHandler); ok {
		if resourceAwareHandler, ok := monitoredHandler.handler.(ResourceAwareHandler); ok {
			return resourceAwareHandler.SetResourceLimits(maxMemoryBytes, maxGoroutines, maxConnections)
		}
	}
	
	return fmt.Errorf("plugin %s does not support resource limits", toolID)
}

// GetPluginDetailedMetrics 獲取插件的詳細監控指標
func (r *Registry) GetPluginDetailedMetrics(toolID string) map[string]interface{} {
	r.rwmu.RLock()
	handler, exists := r.byToolID[toolID]
	r.rwmu.RUnlock()
	
	if !exists {
		return nil
	}
	
	// 嘗試獲取增強監控處理器的詳細指標
	if enhancedHandler, ok := handler.(*EnhancedMonitoredHandler); ok {
		return enhancedHandler.GetDetailedMetrics()
	}
	
	// 回退到基本監控指標
	if r.resourceMonitor != nil {
		basicMetrics := r.resourceMonitor.GetPluginMetrics(toolID)
		if basicMetrics != nil {
			return map[string]interface{}{
				"plugin_id":           toolID,
				"total_requests":      basicMetrics.TotalRequests,
				"active_requests":     basicMetrics.ActiveRequests,
				"total_errors":        basicMetrics.TotalErrors,
				"avg_response_time":   basicMetrics.AvgResponseTimeMs,
				"requests_per_second": basicMetrics.RequestsPerSecond,
				"error_rate_percent":  basicMetrics.ErrorRate,
				"memory_usage_bytes":  basicMetrics.MemoryUsageBytes,
				"goroutine_count":     basicMetrics.GoroutineCount,
				"connection_count":    basicMetrics.ConnectionCount,
				"uptime_ms":          time.Now().UnixMilli() - basicMetrics.StartTime,
				"last_update":        basicMetrics.LastUpdateTime,
			}
		}
	}
	
	return nil
}
