// Package metrics provides a unified interface for querying metrics from various sources
package metrics

import (
	"context"
	"fmt"
	"time"
)

// MetricsProvider defines the interface for querying metrics from various sources
type MetricsProvider interface {
	// Query executes a single metric query and returns the result
	Query(ctx context.Context, query MetricQuery) (*QueryResult, error)

	// BatchQuery executes multiple queries in parallel for efficiency
	BatchQuery(ctx context.Context, queries []MetricQuery) ([]*QueryResult, error)

	// GetAggregation performs aggregation queries on metrics
	GetAggregation(ctx context.Context, opts AggregationOptions) (*AggregationResult, error)

	// HealthCheck verifies the metrics provider is accessible and functional
	HealthCheck(ctx context.Context) error

	// Close releases any resources held by the provider
	Close() error
}

// LongTermMetricsProvider extends MetricsProvider with long-term storage capabilities
// This interface is reserved for future Mimir implementation
type LongTermMetricsProvider interface {
	MetricsProvider

	// QueryHistorical queries historical data beyond normal retention
	QueryHistorical(ctx context.Context, query HistoricalQuery) (*HistoricalResult, error)

	// QueryDownsampled queries downsampled data for better performance on long ranges
	QueryDownsampled(ctx context.Context, query DownsampledQuery) (*QueryResult, error)

	// QueryTenant queries metrics for a specific tenant (multi-tenancy support)
	QueryTenant(ctx context.Context, tenantID string, query MetricQuery) (*QueryResult, error)
}

// MetricQuery represents a query for metrics
type MetricQuery struct {
	// Metric name (e.g., "cpu_usage", "http_request_duration_seconds")
	Metric string `json:"metric"`

	// Labels for filtering (e.g., {"service": "api", "env": "prod"})
	Labels map[string]string `json:"labels"`

	// Time range for the query
	TimeRange TimeRange `json:"time_range"`

	// Step duration for range queries (e.g., 1m, 5m, 1h)
	Step time.Duration `json:"step"`

	// Aggregation function (avg, max, min, sum, count, p50, p95, p99)
	Aggregation string `json:"aggregation,omitempty"`

	// Additional provider-specific options
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
	Duration      time.Duration `json:"duration"`
	SamplesScanned int64        `json:"samples_scanned"`
	SeriesScanned  int64        `json:"series_scanned"`
}

// AggregationOptions defines options for aggregation queries
type AggregationOptions struct {
	// Metrics to aggregate
	Metrics []string `json:"metrics"`

	// Time range for aggregation
	TimeRange TimeRange `json:"time_range"`

	// Aggregation function
	Function AggregationFunc `json:"function"`

	// Group by labels
	GroupBy []string `json:"group_by,omitempty"`

	// Additional filters
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