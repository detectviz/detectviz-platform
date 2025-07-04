package performance

import (
	"context"
	"runtime"
	"sync"
	"time"

	"detectviz-platform/pkg/platform/contracts"
)

// PerformanceOptimizer 性能優化器
type PerformanceOptimizer struct {
	logger           contracts.Logger
	cache            *CacheManager
	poolManager      *PoolManager
	gcOptimizer      *GCOptimizer
	concurrencyLimit int
	semaphore        chan struct{}
	mu               sync.RWMutex
}

// NewPerformanceOptimizer 創建新的性能優化器
func NewPerformanceOptimizer(logger contracts.Logger, concurrencyLimit int) *PerformanceOptimizer {
	return &PerformanceOptimizer{
		logger:           logger,
		cache:            NewCacheManager(logger),
		poolManager:      NewPoolManager(logger),
		gcOptimizer:      NewGCOptimizer(logger),
		concurrencyLimit: concurrencyLimit,
		semaphore:        make(chan struct{}, concurrencyLimit),
	}
}

// OptimizeSystem 優化系統性能
func (po *PerformanceOptimizer) OptimizeSystem(ctx context.Context) error {
	po.logger.Info("開始系統性能優化...")

	// 優化 GC 設置
	if err := po.gcOptimizer.OptimizeGC(); err != nil {
		po.logger.Error("GC 優化失敗: %v", err)
		return err
	}

	// 優化對象池
	po.poolManager.OptimizePools()

	// 清理緩存
	po.cache.CleanupExpiredEntries()

	po.logger.Info("系統性能優化完成")
	return nil
}

// GetName 返回優化器名稱
func (po *PerformanceOptimizer) GetName() string {
	return "performance_optimizer"
}

// CacheManager 緩存管理器
type CacheManager struct {
	logger contracts.Logger
	cache  sync.Map
	stats  CacheStats
	mu     sync.RWMutex
}

// CacheStats 緩存統計
type CacheStats struct {
	Hits        int64
	Misses      int64
	Evictions   int64
	TotalSize   int64
	LastCleanup time.Time
}

// NewCacheManager 創建新的緩存管理器
func NewCacheManager(logger contracts.Logger) *CacheManager {
	return &CacheManager{
		logger: logger,
		stats:  CacheStats{LastCleanup: time.Now()},
	}
}

// CacheEntry 緩存條目
type CacheEntry struct {
	Value     interface{}
	ExpiresAt time.Time
}

// Set 設置緩存
func (cm *CacheManager) Set(key string, value interface{}, ttl time.Duration) {
	entry := CacheEntry{
		Value:     value,
		ExpiresAt: time.Now().Add(ttl),
	}
	cm.cache.Store(key, entry)

	cm.mu.Lock()
	cm.stats.TotalSize++
	cm.mu.Unlock()
}

// Get 獲取緩存
func (cm *CacheManager) Get(key string) (interface{}, bool) {
	value, exists := cm.cache.Load(key)
	if !exists {
		cm.mu.Lock()
		cm.stats.Misses++
		cm.mu.Unlock()
		return nil, false
	}

	entry := value.(CacheEntry)
	if time.Now().After(entry.ExpiresAt) {
		cm.cache.Delete(key)
		cm.mu.Lock()
		cm.stats.Misses++
		cm.stats.Evictions++
		cm.stats.TotalSize--
		cm.mu.Unlock()
		return nil, false
	}

	cm.mu.Lock()
	cm.stats.Hits++
	cm.mu.Unlock()
	return entry.Value, true
}

// CleanupExpiredEntries 清理過期條目
func (cm *CacheManager) CleanupExpiredEntries() {
	now := time.Now()
	evicted := 0

	cm.cache.Range(func(key, value interface{}) bool {
		entry := value.(CacheEntry)
		if now.After(entry.ExpiresAt) {
			cm.cache.Delete(key)
			evicted++
		}
		return true
	})

	cm.mu.Lock()
	cm.stats.Evictions += int64(evicted)
	cm.stats.TotalSize -= int64(evicted)
	cm.stats.LastCleanup = now
	cm.mu.Unlock()

	cm.logger.Info("緩存清理完成，清理了 %d 個過期條目", evicted)
}

// GetStats 獲取緩存統計
func (cm *CacheManager) GetStats() CacheStats {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return cm.stats
}

// PoolManager 對象池管理器
type PoolManager struct {
	logger contracts.Logger
	pools  map[string]*sync.Pool
	stats  PoolStats
	mu     sync.RWMutex
}

// PoolStats 對象池統計
type PoolStats struct {
	PoolCount    int
	TotalGets    int64
	TotalPuts    int64
	LastOptimize time.Time
}

// NewPoolManager 創建新的對象池管理器
func NewPoolManager(logger contracts.Logger) *PoolManager {
	return &PoolManager{
		logger: logger,
		pools:  make(map[string]*sync.Pool),
		stats:  PoolStats{LastOptimize: time.Now()},
	}
}

// GetPool 獲取對象池
func (pm *PoolManager) GetPool(name string, newFunc func() interface{}) *sync.Pool {
	pm.mu.RLock()
	pool, exists := pm.pools[name]
	pm.mu.RUnlock()

	if exists {
		return pool
	}

	pm.mu.Lock()
	defer pm.mu.Unlock()

	// 雙重檢查
	if pool, exists := pm.pools[name]; exists {
		return pool
	}

	pool = &sync.Pool{New: newFunc}
	pm.pools[name] = pool
	pm.stats.PoolCount++

	return pool
}

// OptimizePools 優化對象池
func (pm *PoolManager) OptimizePools() {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	// 清理未使用的池
	for name, pool := range pm.pools {
		// 嘗試從池中獲取對象，如果為空則考慮清理
		obj := pool.Get()
		if obj == nil {
			pm.logger.Debug("對象池 %s 為空，考慮清理", name)
		} else {
			pool.Put(obj)
		}
	}

	pm.stats.LastOptimize = time.Now()
	pm.logger.Info("對象池優化完成")
}

// GetStats 獲取對象池統計
func (pm *PoolManager) GetStats() PoolStats {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	return pm.stats
}

// GCOptimizer GC 優化器
type GCOptimizer struct {
	logger contracts.Logger
}

// NewGCOptimizer 創建新的 GC 優化器
func NewGCOptimizer(logger contracts.Logger) *GCOptimizer {
	return &GCOptimizer{
		logger: logger,
	}
}

// OptimizeGC 優化垃圾回收
func (gco *GCOptimizer) OptimizeGC() error {
	// 手動觸發 GC
	runtime.GC()
	gco.logger.Info("手動觸發 GC 完成")

	return nil
}
