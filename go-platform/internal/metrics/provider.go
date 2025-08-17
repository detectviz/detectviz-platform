// Package metrics 提供統一的指標查詢介面，支援多種資料來源
package metrics

import (
	"context"
	"fmt"
	"time"
)

// MetricsProvider 定義從各種來源查詢指標的統一介面
type MetricsProvider interface {
	// Query 執行單一指標查詢並返回結果
	Query(ctx context.Context, query MetricQuery) (*QueryResult, error)

	// BatchQuery 並行執行多個查詢以提高效率
	BatchQuery(ctx context.Context, queries []MetricQuery) ([]*QueryResult, error)

	// GetAggregation 對指標執行聚合查詢
	GetAggregation(ctx context.Context, opts AggregationOptions) (*AggregationResult, error)

	// HealthCheck 驗證指標提供者是否可訪問且功能正常
	HealthCheck(ctx context.Context) error

	// Close 釋放提供者持有的任何資源
	Close() error
}

// LongTermMetricsProvider 擴展 MetricsProvider，支援長期儲存功能
// 此介面預留給未來的 Mimir 實作
type LongTermMetricsProvider interface {
	MetricsProvider

	// QueryHistorical 查詢超出正常保留期的歷史數據
	QueryHistorical(ctx context.Context, query HistoricalQuery) (*HistoricalResult, error)

	// QueryDownsampled 查詢降採樣數據，在長時間範圍內獲得更好的性能
	QueryDownsampled(ctx context.Context, query DownsampledQuery) (*QueryResult, error)

	// QueryTenant 查詢特定租戶的指標（多租戶支援）
	QueryTenant(ctx context.Context, tenantID string, query MetricQuery) (*QueryResult, error)
}

// MetricQuery 表示指標查詢
type MetricQuery struct {
	// 指標名稱（例如："cpu_usage", "http_request_duration_seconds"）
	Metric string `json:"metric"`

	// 用於過濾的標籤（例如：{"service": "api", "env": "prod"}）
	Labels map[string]string `json:"labels"`

	// 查詢的時間範圍
	TimeRange TimeRange `json:"time_range"`

	// 範圍查詢的步長（例如：1m, 5m, 1h）
	Step time.Duration `json:"step"`

	// 聚合函數（avg, max, min, sum, count, p50, p95, p99）
	Aggregation string `json:"aggregation,omitempty"`

	// 額外的提供者特定選項
	Options map[string]interface{} `json:"options,omitempty"`
}

// TimeRange represents a time range for queries
type TimeRange struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}

// QueryResult represents the result of a metric query
type QueryResult struct {
	// Query that produced this result
	Query *MetricQuery `json:"query"`

	// Series contains the time series data
	Series []TimeSeries `json:"series"`

	// Warnings from the query execution
	Warnings []string `json:"warnings,omitempty"`

	// Stats contains query statistics
	Stats *QueryStats `json:"stats,omitempty"`

	// Truncated indicates if results were truncated due to size limits
	Truncated bool `json:"truncated,omitempty"`

	// MaxSeriesExceeded indicates if max series limit was exceeded
	MaxSeriesExceeded bool `json:"max_series_exceeded,omitempty"`

	// MaxDataPointsExceeded indicates if max data points limit was exceeded
	MaxDataPointsExceeded bool `json:"max_data_points_exceeded,omitempty"`
}

// TimeSeries represents a single time series
type TimeSeries struct {
	// Labels identifying this series
	Labels map[string]string `json:"labels"`

	// Values contains the data points
	Values []DataPoint `json:"values"`
}

// DataPoint represents a single data point in a time series
type DataPoint struct {
	Timestamp time.Time `json:"timestamp"`
	Value     float64   `json:"value"`
}

// QueryStats contains statistics about query execution
type QueryStats struct {
	Duration       time.Duration `json:"duration"`
	SamplesScanned int64         `json:"samples_scanned"`
	SeriesScanned  int64         `json:"series_scanned"`
}

// AggregationOptions 定義聚合查詢的選項
type AggregationOptions struct {
	// 要聚合的指標
	Metrics []string `json:"metrics"`

	// 聚合的時間範圍
	TimeRange TimeRange `json:"time_range"`

	// 聚合函數
	Function AggregationFunc `json:"function"`

	// 按標籤分組
	GroupBy []string `json:"group_by,omitempty"`

	// 額外的篩選條件
	Filters map[string]string `json:"filters,omitempty"`
}

// AggregationFunc represents an aggregation function
type AggregationFunc string

const (
	AggregationAvg   AggregationFunc = "avg"
	AggregationMax   AggregationFunc = "max"
	AggregationMin   AggregationFunc = "min"
	AggregationSum   AggregationFunc = "sum"
	AggregationCount AggregationFunc = "count"
	AggregationP50   AggregationFunc = "p50"
	AggregationP95   AggregationFunc = "p95"
	AggregationP99   AggregationFunc = "p99"
)

// AggregationResult represents the result of an aggregation query
type AggregationResult struct {
	// Aggregation options used
	Options *AggregationOptions `json:"options"`

	// Aggregated values
	Values []AggregatedValue `json:"values"`

	// Warnings from the aggregation
	Warnings []string `json:"warnings,omitempty"`
}

// AggregatedValue represents an aggregated metric value
type AggregatedValue struct {
	// Metric name
	Metric string `json:"metric"`

	// Group labels (if grouped)
	Labels map[string]string `json:"labels,omitempty"`

	// Aggregated value
	Value float64 `json:"value"`

	// Number of samples aggregated
	SampleCount int64 `json:"sample_count"`
}

// HistoricalQuery represents a query for historical data (future Mimir support)
type HistoricalQuery struct {
	MetricQuery

	// Downsampling resolution for long-term data
	Resolution time.Duration `json:"resolution"`

	// Whether to include raw samples
	IncludeRaw bool `json:"include_raw"`
}

// HistoricalResult represents the result of a historical query
type HistoricalResult struct {
	QueryResult

	// Data retention information
	RetentionInfo *RetentionInfo `json:"retention_info"`
}

// RetentionInfo contains information about data retention
type RetentionInfo struct {
	// Oldest available data point
	OldestDataPoint time.Time `json:"oldest_data_point"`

	// Retention policy applied
	RetentionPolicy string `json:"retention_policy"`

	// Data completeness percentage
	Completeness float64 `json:"completeness"`
}

// DownsampledQuery represents a query for downsampled data
type DownsampledQuery struct {
	MetricQuery

	// Downsampling method (avg, max, min)
	Method string `json:"method"`

	// Target resolution
	Resolution time.Duration `json:"resolution"`
}

// ProviderConfig contains configuration for metrics providers
type ProviderConfig struct {
	// Provider type: "prometheus", "mimir", "memory"
	Type string `json:"type"`

	// Provider-specific configuration
	Config interface{} `json:"config"`

	// Query routing configuration
	QueryRouter *QueryRouterConfig `json:"query_router,omitempty"`

	// Result limits configuration
	Limits *ResultLimits `json:"limits,omitempty"`
}

// ResultLimits defines limits for query results
type ResultLimits struct {
	// Maximum number of time series per query
	MaxSeries int `json:"max_series"`

	// Maximum number of data points per series
	MaxDataPointsPerSeries int `json:"max_data_points_per_series"`

	// Maximum total data points per query
	MaxTotalDataPoints int `json:"max_total_data_points"`

	// Maximum query result size in bytes
	MaxResultSizeBytes int64 `json:"max_result_size_bytes"`
}

// QueryRouterConfig defines routing rules for queries
type QueryRouterConfig struct {
	// Short-term query threshold in days
	ShortTermDays int `json:"short_term_days"`

	// Provider for short-term queries
	ShortTermProvider string `json:"short_term_provider"`

	// Provider for long-term queries
	LongTermProvider string `json:"long_term_provider"`
}

// Error types for metrics operations
var (
	// ErrProviderNotAvailable indicates the metrics provider is not accessible
	ErrProviderNotAvailable = fmt.Errorf("metrics provider not available")

	// ErrQueryTimeout indicates a query exceeded the timeout
	ErrQueryTimeout = fmt.Errorf("query timeout")

	// ErrInvalidQuery indicates an invalid query format
	ErrInvalidQuery = fmt.Errorf("invalid query")

	// ErrNoData indicates no data was found for the query
	ErrNoData = fmt.Errorf("no data found")

	// ErrTooManyDataPoints indicates the query would return too many data points
	ErrTooManyDataPoints = fmt.Errorf("too many data points")
)
