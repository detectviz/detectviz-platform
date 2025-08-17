package pluginhost

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	v1 "github.com/detectviz/detectviz-platform/contracts/gen/go/detectviz/contracts/v1"
	"go.uber.org/zap"
)

// 插件健康狀態枚舉
type PluginHealth int

const (
	HealthUnknown      PluginHealth = iota // 0: 未知狀態
	HealthOk                               // 1: 正常運行
	HealthDegraded                         // 2: 部分功能可用
	HealthCritical                         // 3: 嚴重錯誤
	HealthShuttingDown                     // 4: 關閉中
)

// String 實作 Stringer 介面
func (h PluginHealth) String() string {
	switch h {
	case HealthUnknown:
		return "未知"
	case HealthOk:
		return "正常"
	case HealthDegraded:
		return "降級"
	case HealthCritical:
		return "異常"
	case HealthShuttingDown:
		return "關閉中"
	default:
		return "無效狀態"
	}
}

// Handler 為每個插件的處理器介面
type Handler interface {
	Invoke(ctx context.Context, req *v1.InvokeRequest) (*v1.InvokeResponse, error)
}

// ClosableHandler 支援資源釋放的處理器
type ClosableHandler interface {
	Handler
	Close() error
}

// HealthAwareHandler 支援健康檢查的處理器
type HealthAwareHandler interface {
	ClosableHandler
	HealthCheck() error
}

// PluginWrapper 插件封裝結構（包含健康狀態和負載追蹤）
type PluginWrapper struct {
	handler ClosableHandler
	name    string

	// 原子操作變數（無需鎖）
	health    atomic.Value // 存儲 PluginHealth
	load      int32        // 當前負載計數器
	lastCheck atomic.Value // 存儲 time.Time，最後檢查時間戳
}

// NewPluginWrapper 創建新的插件包裝器
func NewPluginWrapper(name string, handler ClosableHandler) *PluginWrapper {
	wrapper := &PluginWrapper{
		handler: handler,
		name:    name,
	}
	wrapper.health.Store(HealthUnknown)
	wrapper.lastCheck.Store(time.Now())
	return wrapper
}

// GetHealth 取得健康狀態
func (pw *PluginWrapper) GetHealth() PluginHealth {
	return pw.health.Load().(PluginHealth)
}

// SetHealth 設定健康狀態
func (pw *PluginWrapper) SetHealth(health PluginHealth) {
	pw.health.Store(health)
	pw.lastCheck.Store(time.Now())
}

// GetLoad 取得當前負載
func (pw *PluginWrapper) GetLoad() int32 {
	return atomic.LoadInt32(&pw.load)
}

// GetLastCheck 取得最後檢查時間
func (pw *PluginWrapper) GetLastCheck() time.Time {
	return pw.lastCheck.Load().(time.Time)
}

// Registry 插件註冊中心（核心數據結構）
type Registry struct {
	mu      sync.RWMutex              // 讀寫鎖保護 plugins map
	plugins map[string]*PluginWrapper // 插件映射表

	// 健康檢查控制
	healthTicker   *time.Ticker
	shutdownSignal chan struct{}
	wg             sync.WaitGroup // 確保協程完成

	// 配置參數
	maxLoad        int32         // 最大負載閾值
	healthInterval time.Duration // 健康檢查間隔
	logger         *zap.Logger   // 日誌記錄器
}

// RegistryConfig 註冊中心配置
type RegistryConfig struct {
	MaxLoad        int32         `yaml:"max_load" json:"max_load"`               // 最大負載，預設 100
	HealthInterval time.Duration `yaml:"health_interval" json:"health_interval"` // 健康檢查間隔，預設 30s
}

// DefaultRegistryConfig 預設配置
func DefaultRegistryConfig() *RegistryConfig {
	return &RegistryConfig{
		MaxLoad:        100,
		HealthInterval: 30 * time.Second,
	}
}

// NewRegistry 創建新註冊中心
func NewRegistry() *Registry {
	return NewRegistryWithConfig(DefaultRegistryConfig())
}

// NewRegistryWithConfig 使用配置創建註冊中心
func NewRegistryWithConfig(config *RegistryConfig) *Registry {
	if config == nil {
		config = DefaultRegistryConfig()
	}

	return &Registry{
		plugins:        make(map[string]*PluginWrapper),
		shutdownSignal: make(chan struct{}),
		maxLoad:        config.MaxLoad,
		healthInterval: config.HealthInterval,
		logger:         zap.L(),
	}
}

// Register 註冊插件（執行緒安全）
func (r *Registry) Register(name string, handler ClosableHandler) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.plugins[name]; exists {
		return fmt.Errorf("插件已註冊: %s", name)
	}

	wrapper := NewPluginWrapper(name, handler)
	r.plugins[name] = wrapper

	r.logger.Info("插件註冊成功",
		zap.String("name", name),
		zap.String("type", fmt.Sprintf("%T", handler)))
	return nil
}

// RegisterOrReplace 註冊或替換插件
func (r *Registry) RegisterOrReplace(name string, handler ClosableHandler) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// 如果存在舊插件，先關閉
	if oldWrapper, exists := r.plugins[name]; exists {
		oldWrapper.SetHealth(HealthShuttingDown)
		if err := oldWrapper.handler.Close(); err != nil {
			r.logger.Warn("關閉舊插件失敗",
				zap.String("name", name),
				zap.Error(err))
		}
	}

	wrapper := NewPluginWrapper(name, handler)
	r.plugins[name] = wrapper

	r.logger.Info("插件註冊或替換成功",
		zap.String("name", name),
		zap.String("type", fmt.Sprintf("%T", handler)))
	return nil
}

// Unregister 取消註冊插件
func (r *Registry) Unregister(name string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	wrapper, exists := r.plugins[name]
	if !exists {
		return fmt.Errorf("插件不存在: %s", name)
	}

	wrapper.SetHealth(HealthShuttingDown)
	if err := wrapper.handler.Close(); err != nil {
		r.logger.Warn("關閉插件失敗",
			zap.String("name", name),
			zap.Error(err))
	}

	delete(r.plugins, name)
	r.logger.Info("插件取消註冊成功", zap.String("name", name))
	return nil
}

// StartHealthChecks 啟動健康檢查協程
func (r *Registry) StartHealthChecks() {
	if r.healthTicker != nil {
		return // 已經啟動
	}

	r.healthTicker = time.NewTicker(r.healthInterval)
	r.wg.Add(1)

	go func() {
		defer r.wg.Done()
		r.logger.Info("健康檢查協程啟動",
			zap.Duration("interval", r.healthInterval))

		for {
			select {
			case <-r.healthTicker.C:
				r.performHealthChecks()
			case <-r.shutdownSignal:
				r.logger.Info("健康檢查協程終止")
				return
			}
		}
	}()
}

// performHealthChecks 執行全量健康檢查
func (r *Registry) performHealthChecks() {
	r.mu.RLock()
	pluginCount := len(r.plugins)
	plugins := make([]*PluginWrapper, 0, pluginCount)
	for _, wrapper := range r.plugins {
		plugins = append(plugins, wrapper)
	}
	r.mu.RUnlock()

	var wg sync.WaitGroup
	for _, wrapper := range plugins {
		wg.Add(1)
		go func(w *PluginWrapper) {
			defer wg.Done()
			r.checkPluginHealth(w)
		}(wrapper)
	}
	wg.Wait()

	r.logger.Debug("健康檢查完成", zap.Int("checked_plugins", pluginCount))
}

// checkPluginHealth 檢查單個插件的健康狀態
func (r *Registry) checkPluginHealth(wrapper *PluginWrapper) {
	// 跳過正在關閉的插件
	if wrapper.GetHealth() == HealthShuttingDown {
		return
	}

	start := time.Now()
	var err error

	// 如果插件支援健康檢查，使用自定義邏輯
	if healthAware, ok := wrapper.handler.(HealthAwareHandler); ok {
		err = healthAware.HealthCheck()
	} else {
		// 否則使用簡單的調用測試
		err = r.simpleHealthCheck(wrapper)
	}

	duration := time.Since(start)

	// 更新健康狀態
	var newHealth PluginHealth
	switch {
	case err != nil:
		newHealth = HealthCritical
		r.logger.Warn("插件健康檢查失敗",
			zap.String("name", wrapper.name),
			zap.Error(err),
			zap.Duration("duration", duration))
	case duration > 100*time.Millisecond:
		newHealth = HealthDegraded
		r.logger.Debug("插件回應緩慢",
			zap.String("name", wrapper.name),
			zap.Duration("duration", duration))
	default:
		newHealth = HealthOk
	}

	wrapper.SetHealth(newHealth)
}

// simpleHealthCheck 簡單的健康檢查（針對不支援 HealthAwareHandler 的插件）
func (r *Registry) simpleHealthCheck(wrapper *PluginWrapper) error {
	// 這裡可以實作基本的插件狀態檢查
	// 例如檢查是否還能正常回應等
	return nil
}

// Invoke 調用插件（含熔斷和負載保護）
func (r *Registry) Invoke(ctx context.Context, pluginName string, req *v1.InvokeRequest) (*v1.InvokeResponse, error) {
	// 讀取階段（使用讀鎖）
	r.mu.RLock()
	wrapper, exists := r.plugins[pluginName]
	r.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("插件未找到: %s", pluginName)
	}

	// 健康狀態檢查（熔斷機制）
	health := wrapper.GetHealth()
	if health == HealthCritical || health == HealthShuttingDown {
		return nil, fmt.Errorf("插件不可用 (%s): %s", health, pluginName)
	}

	// 負載保護（避免單一插件過載）
	currentLoad := atomic.AddInt32(&wrapper.load, 1)
	defer atomic.AddInt32(&wrapper.load, -1)

	if currentLoad > r.maxLoad {
		return nil, fmt.Errorf("插件過載 (%d/%d): %s", currentLoad, r.maxLoad, pluginName)
	}

	// 執行插件邏輯
	start := time.Now()
	resp, err := wrapper.handler.Invoke(ctx, req)
	duration := time.Since(start)

	// 記錄執行統計
	if err != nil {
		r.logger.Warn("插件調用失敗",
			zap.String("name", pluginName),
			zap.Error(err),
			zap.Duration("duration", duration),
			zap.Int32("load", currentLoad))
	} else {
		r.logger.Debug("插件調用成功",
			zap.String("name", pluginName),
			zap.Duration("duration", duration),
			zap.Int32("load", currentLoad))
	}

	return resp, err
}

// Lookup 查找插件處理器（保持相容性）
func (r *Registry) Lookup(pluginName string) (Handler, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if wrapper, exists := r.plugins[pluginName]; exists {
		return wrapper.handler, true
	}
	return nil, false
}

// GetPluginStatus 獲取插件狀態（用於監控）
func (r *Registry) GetPluginStatus() map[string]interface{} {
	status := make(map[string]interface{})

	r.mu.RLock()
	defer r.mu.RUnlock()

	for name, wrapper := range r.plugins {
		status[name] = map[string]interface{}{
			"health":     wrapper.GetHealth(),
			"health_str": wrapper.GetHealth().String(),
			"load":       wrapper.GetLoad(),
			"last_check": wrapper.GetLastCheck(),
		}
	}
	return status
}

// GetPluginNames 獲取所有插件名稱
func (r *Registry) GetPluginNames() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	names := make([]string, 0, len(r.plugins))
	for name := range r.plugins {
		names = append(names, name)
	}
	return names
}

// GetPluginCount 獲取插件數量
func (r *Registry) GetPluginCount() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.plugins)
}

// Shutdown 優雅關閉流程
func (r *Registry) Shutdown() {
	r.logger.Info("啟動插件註冊中心關閉程序")

	// 1. 停止健康檢查
	close(r.shutdownSignal)
	if r.healthTicker != nil {
		r.healthTicker.Stop()
	}

	// 2. 標記所有插件為關閉中
	r.mu.RLock()
	shutdownCount := len(r.plugins)
	for _, wrapper := range r.plugins {
		wrapper.SetHealth(HealthShuttingDown)
	}
	r.mu.RUnlock()

	r.logger.Info("標記插件關閉中", zap.Int("count", shutdownCount))

	// 3. 等待進行中請求完成
	r.logger.Info("等待進行中請求完成...")
	waitDuration := 2 * time.Second

	// 檢查是否還有負載
	ticker := time.NewTicker(100 * time.Millisecond)
	timeout := time.After(waitDuration)

checkLoop:
	for {
		select {
		case <-ticker.C:
			hasLoad := false
			r.mu.RLock()
			for _, wrapper := range r.plugins {
				if wrapper.GetLoad() > 0 {
					hasLoad = true
					break
				}
			}
			r.mu.RUnlock()

			if !hasLoad {
				break checkLoop
			}
		case <-timeout:
			r.logger.Warn("等待超時，強制關閉")
			break checkLoop
		}
	}
	ticker.Stop()

	// 4. 關閉所有插件
	r.mu.Lock()
	defer r.mu.Unlock()

	var closeErrors []error
	for name, wrapper := range r.plugins {
		r.logger.Debug("關閉插件", zap.String("name", name))
		if err := wrapper.handler.Close(); err != nil {
			closeErrors = append(closeErrors, fmt.Errorf("關閉插件 %s 失敗: %w", name, err))
		}
	}

	if len(closeErrors) > 0 {
		for _, err := range closeErrors {
			r.logger.Warn("插件關閉錯誤", zap.Error(err))
		}
	}

	// 清空插件表
	r.plugins = make(map[string]*PluginWrapper)

	// 5. 等待協程退出
	r.wg.Wait()
	r.logger.Info("插件註冊中心安全關閉完成")
}
