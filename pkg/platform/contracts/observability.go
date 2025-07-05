package contracts

import "context"

// RateLimiterProvider 定義了速率限制服務的介面。
// 職責: 控制對資源的訪問速率，防止服務因過多請求而過載。
// AI_PLUGIN_TYPE: "rate_limiter_provider"
// AI_IMPL_PACKAGE: "detectviz-platform/internal/platform/providers/rate_limiter/uber_rate_limiter"
// AI_IMPL_CONSTRUCTOR: "NewUberRateLimiterProvider"
// @See: internal/platform/providers/rate_limiter/uber_rate_limiter.go
type RateLimiterProvider interface {
	// Allow 檢查基於給定的鍵是否應允許請求通過。
	Allow(ctx context.Context, key string) bool
	// GetName 返回速率限制器的名稱。
	GetName() string
}

// CircuitBreakerProvider 定義了熔斷器服務的介面。
// 職責: 在檢測到外部服務失敗時，快速失敗並提供降級處理，防止級聯故障。
// AI_PLUGIN_TYPE: "circuit_breaker_provider"
// AI_IMPL_PACKAGE: "detectviz-platform/internal/platform/providers/circuit_breaker/hystrix_breaker"
// AI_IMPL_CONSTRUCTOR: "NewHystrixCircuitBreakerProvider"
// @See: internal/platform/providers/circuit_breaker/hystrix_breaker.go
type CircuitBreakerProvider interface {
	// Execute 將一個函數包裝在熔斷器邏輯中執行。
	// 如果主函數(run)失敗達到閾值，熔斷器會打開並在一段時間內直接調用降級函數(fallback)。
	Execute(ctx context.Context, name string, run func() error, fallback func(error) error) error
	// GetName 返回熔斷器提供者的名稱。
	GetName() string
}
