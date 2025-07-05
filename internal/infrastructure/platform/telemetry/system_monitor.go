package telemetry

import (
	"context"
	"runtime"
	"time"

	"detectviz-platform/pkg/platform/contracts"
)

// SystemMonitor 監控系統資源和性能指標
// AI_PLUGIN_TYPE: "system_monitor"
// AI_IMPL_PACKAGE: "detectviz-platform/internal/infrastructure/platform/telemetry"
// AI_IMPL_CONSTRUCTOR: "NewSystemMonitor"
type SystemMonitor struct {
	metricsCollector *MetricsCollector
	logger           contracts.Logger
	ticker           *time.Ticker
	stopChan         chan struct{}
	interval         time.Duration
}

// NewSystemMonitor 創建新的系統監控器
func NewSystemMonitor(metricsCollector *MetricsCollector, logger contracts.Logger, interval time.Duration) *SystemMonitor {
	return &SystemMonitor{
		metricsCollector: metricsCollector,
		logger:           logger,
		interval:         interval,
		stopChan:         make(chan struct{}),
	}
}

// Start 開始監控系統指標
func (sm *SystemMonitor) Start(ctx context.Context) {
	sm.ticker = time.NewTicker(sm.interval)

	sm.logger.Info("系統監控器已啟動，監控間隔: %v", sm.interval)

	// 立即收集一次指標
	sm.collectMetrics()

	go func() {
		for {
			select {
			case <-ctx.Done():
				sm.logger.Info("系統監控器正在停止...")
				sm.Stop()
				return
			case <-sm.stopChan:
				sm.logger.Info("系統監控器已停止")
				return
			case <-sm.ticker.C:
				sm.collectMetrics()
			}
		}
	}()
}

// Stop 停止監控
func (sm *SystemMonitor) Stop() {
	if sm.ticker != nil {
		sm.ticker.Stop()
	}

	select {
	case sm.stopChan <- struct{}{}:
	default:
		// 如果通道已滿，忽略
	}
}

// collectMetrics 收集系統指標
func (sm *SystemMonitor) collectMetrics() {
	// 收集記憶體指標
	sm.collectMemoryMetrics()

	// 收集 Goroutine 指標
	sm.collectGoroutineMetrics()

	// 收集 GC 指標
	sm.collectGCMetrics()
}

// collectMemoryMetrics 收集記憶體指標
func (sm *SystemMonitor) collectMemoryMetrics() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	// 記錄各種記憶體指標
	sm.metricsCollector.RecordSystemMemory("heap_alloc", float64(m.HeapAlloc))
	sm.metricsCollector.RecordSystemMemory("heap_sys", float64(m.HeapSys))
	sm.metricsCollector.RecordSystemMemory("heap_idle", float64(m.HeapIdle))
	sm.metricsCollector.RecordSystemMemory("heap_inuse", float64(m.HeapInuse))
	sm.metricsCollector.RecordSystemMemory("heap_released", float64(m.HeapReleased))
	sm.metricsCollector.RecordSystemMemory("heap_objects", float64(m.HeapObjects))
	sm.metricsCollector.RecordSystemMemory("stack_inuse", float64(m.StackInuse))
	sm.metricsCollector.RecordSystemMemory("stack_sys", float64(m.StackSys))
	sm.metricsCollector.RecordSystemMemory("mspan_inuse", float64(m.MSpanInuse))
	sm.metricsCollector.RecordSystemMemory("mspan_sys", float64(m.MSpanSys))
	sm.metricsCollector.RecordSystemMemory("mcache_inuse", float64(m.MCacheInuse))
	sm.metricsCollector.RecordSystemMemory("mcache_sys", float64(m.MCacheSys))
	sm.metricsCollector.RecordSystemMemory("other_sys", float64(m.OtherSys))
	sm.metricsCollector.RecordSystemMemory("sys", float64(m.Sys))
	sm.metricsCollector.RecordSystemMemory("total_alloc", float64(m.TotalAlloc))
	sm.metricsCollector.RecordSystemMemory("lookups", float64(m.Lookups))
	sm.metricsCollector.RecordSystemMemory("mallocs", float64(m.Mallocs))
	sm.metricsCollector.RecordSystemMemory("frees", float64(m.Frees))
}

// collectGoroutineMetrics 收集 Goroutine 指標
func (sm *SystemMonitor) collectGoroutineMetrics() {
	numGoroutines := runtime.NumGoroutine()
	sm.metricsCollector.RecordSystemGoroutines(float64(numGoroutines))
}

// collectGCMetrics 收集垃圾回收指標
func (sm *SystemMonitor) collectGCMetrics() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	// 記錄 GC 相關指標
	sm.metricsCollector.RecordSystemMemory("gc_cpu_fraction", m.GCCPUFraction)
	sm.metricsCollector.RecordSystemMemory("num_gc", float64(m.NumGC))
	sm.metricsCollector.RecordSystemMemory("num_forced_gc", float64(m.NumForcedGC))
	sm.metricsCollector.RecordSystemMemory("gc_pause_total_ns", float64(m.PauseTotalNs))

	// 記錄最近的 GC 暫停時間
	if len(m.PauseNs) > 0 {
		lastGCPause := m.PauseNs[(m.NumGC+255)%256]
		sm.metricsCollector.RecordSystemMemory("last_gc_pause_ns", float64(lastGCPause))
	}
}

// GetName 返回監控器名稱
func (sm *SystemMonitor) GetName() string {
	return "system_monitor"
}

// GetMetrics 獲取當前系統指標快照
func (sm *SystemMonitor) GetMetrics() map[string]interface{} {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	return map[string]interface{}{
		"goroutines":      runtime.NumGoroutine(),
		"heap_alloc":      m.HeapAlloc,
		"heap_sys":        m.HeapSys,
		"heap_idle":       m.HeapIdle,
		"heap_inuse":      m.HeapInuse,
		"heap_released":   m.HeapReleased,
		"heap_objects":    m.HeapObjects,
		"stack_inuse":     m.StackInuse,
		"stack_sys":       m.StackSys,
		"sys":             m.Sys,
		"total_alloc":     m.TotalAlloc,
		"num_gc":          m.NumGC,
		"gc_cpu_fraction": m.GCCPUFraction,
		"pause_total_ns":  m.PauseTotalNs,
	}
}
