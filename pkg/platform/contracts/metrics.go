package contracts

import "context"

// MetricsProvider 定義了指標收集與導出的介面。
// 職責: 收集應用程式運行時指標（計數器、直方圖、儀表盤）並導出給監控系統（如 Prometheus）。
// AI_PLUGIN_TYPE: "prometheus_metrics_provider"
// AI_IMPL_PACKAGE: "detectviz-platform/internal/infrastructure/platform/metrics/prometheus_metrics_provider"
// AI_IMPL_CONSTRUCTOR: "NewPrometheusMetricsProvider"
// @See: internal/infrastructure/platform/metrics/prometheus_metrics_provider.go
type MetricsProvider interface {
	// IncCounter 增加一個計數器指標的值。
	IncCounter(name string, tags map[string]string)
	// ObserveHistogram 記錄一個直方圖指標的觀測值。
	ObserveHistogram(name string, value float64, tags map[string]string)
	// SetGauge 設置一個儀表盤指標的當前值。
	SetGauge(name string, value float64, tags map[string]string)
	// GetName 返回指標提供者的名稱。
	GetName() string
}

// TracingProvider 定義了分散式追蹤功能的介面。
// 職責: 提供請求追蹤、span 管理等功能。
// AI_PLUGIN_TYPE: "tracing_provider"
// AI_IMPL_PACKAGE: "detectviz-platform/internal/infrastructure/platform/telemetry"
// AI_IMPL_CONSTRUCTOR: "NewJaegerTracingProvider"
// @See: internal/infrastructure/platform/telemetry/tracing.go
type TracingProvider interface {
	// StartSpan 從上下文中開始一個新的追蹤 span。
	StartSpan(ctx context.Context, operationName string) (context.Context, Span)
	// GetName 返回追蹤提供者的名稱。
	GetName() string
}

// Span 定義了追蹤 span 的通用介面，封裝了底層追蹤庫的 span 對象。
type Span interface {
	// SetTag 為 span 添加一個標籤（鍵值對）。
	SetTag(key string, value interface{})
	// SetError 標記 span 為錯誤狀態並記錄錯誤信息。
	SetError(err error)
	// Finish 標記 span 的結束。
	Finish()
}
