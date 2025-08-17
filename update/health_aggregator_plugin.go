// Package health_aggregator provides health metrics aggregation functionality
package health_aggregator

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/detectviz/go-platform/internal/metrics"
	"github.com/detectviz/go-platform/internal/metrics/memory"
	"github.com/detectviz/go-platform/internal/metrics/prometheus"
	"github.com/detectviz/go-platform/internal/pluginhost/framework"
	pb "github.com/detectviz/go-platform/pkg/contracts/gen/go/detectviz/contracts/v1"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

// Plugin implements the health aggregator plugin
type Plugin struct {
	framework.BasePlugin
	provider      metrics.MetricsProvider
	factory       *metrics.Factory
	config        *Config
	logger        *zap.Logger
	mu            sync.RWMutex
	metricsCache  map[string]*CachedMetrics
	cacheDuration time.Duration
}

// Config contains plugin configuration
type Config struct {
	// Provider configuration
	Provider ProviderConfig `json:"provider" yaml:"provider"`

	// Query configuration
	Query QueryConfig `json:"query" yaml:"query"`

	// Cache configuration
	Cache CacheConfig `json:"cache" yaml:"cache"`

	// Metrics to collect
	Metrics []MetricConfig `json:"metrics" yaml:"metrics"`
}

// ProviderConfig contains metrics provider configuration
type ProviderConfig struct {
	Type       string      `json:"type" yaml:"type"` // prometheus, memory
	Prometheus *prometheus.Config `json:"prometheus,omitempty" yaml:"prometheus,omitempty"`
	Memory     *memory.Config     `json:"memory,omitempty" yaml:"memory,omitempty"`
}

// QueryConfig contains query configuration
type QueryConfig struct {
	ParallelQueries int           `json:"parallel_queries" yaml:"parallel_queries"`
	Timeout         time.Duration `json:"timeout" yaml:"timeout"`
	DefaultStep     time.Duration `json:"default_step" yaml:"default_step"`
}

// CacheConfig contains cache configuration
type CacheConfig struct {
	Enabled  bool          `json:"enabled" yaml:"enabled"`
	Duration time.Duration `json:"duration" yaml:"duration"`
}

// MetricConfig defines a metric to collect
type MetricConfig struct {
	Name        string            `json:"name" yaml:"name"`
	Query       string            `json:"query" yaml:"query"`
	Labels      map[string]string `json:"labels" yaml:"labels"`
	Aggregation string            `json:"aggregation" yaml:"aggregation"`
}

// CachedMetrics holds cached metric data
type CachedMetrics struct {
	Data      interface{}
	Timestamp time.Time
}

// HealthQueryRequest represents a health metrics query request
type HealthQueryRequest struct {
	ServiceName string            `json:"service_name"`
	TimeRange   string            `json:"time_range"`
	Metrics     []string          `json:"metrics"`
	Provider    string            `json:"provider,omitempty"`
	Filters     map[string]string `json:"filters,omitempty"`
}

// HealthQueryResponse represents a health metrics query response
type HealthQueryResponse struct {
	ServiceName string                   `json:"service_name"`
	Metrics     map[string]*MetricData   `json:"metrics"`
	Timestamp   time.Time                `json:"timestamp"`
	Warnings    []string                 `json:"warnings,omitempty"`
}

// MetricData contains metric query results
type MetricData struct {
	Name       string       `json:"name"`
	Values     []DataPoint  `json:"values"`
	Statistics *Statistics  `json:"statistics,omitempty"`
}

// DataPoint represents a single metric data point
type DataPoint struct {
	Timestamp time.Time `json:"timestamp"`
	Value     float64   `json:"value"`
}

// Statistics contains metric statistics
type Statistics struct {
	Min     float64 `json:"min"`
	Max     float64 `json:"max"`
	Avg     float64 `json:"avg"`
	Count   int     `json:"count"`
	P50     float64 `json:"p50,omitempty"`
	P95     float64 `json:"p95,omitempty"`
	P99     float64 `json:"p99,omitempty"`
}

// NewPlugin creates a new health aggregator plugin
func NewPlugin() framework.Plugin {
	return &Plugin{
		metricsCache: make(map[string]*CachedMetrics),
	}
}

// Initialize initializes the plugin
func (p *Plugin) Initialize(ctx context.Context, config interface{}, logger *zap.Logger) error {
	p.logger = logger

	// Parse configuration
	cfg, ok := config.(*Config)
	if !ok {
		// Try to parse from map
		configMap, ok := config.(map[string]interface{})
		if !ok {
			return fmt.Errorf("invalid configuration type")
		}
		
		configBytes, err := json.Marshal(configMap)
		if err != nil {
			return fmt.Errorf("failed to marshal config: %w", err)
		}
		
		cfg = &Config{}
		if err := json.Unmarshal(configBytes, cfg); err != nil {
			return fmt.Errorf("failed to unmarshal config: %w", err)
		}
	}
	
	p.config = cfg
	p.cacheDuration = cfg.Cache.Duration
	if p.cacheDuration == 0 {
		p.cacheDuration = 5 * time.Minute
	}

	// Create metrics factory
	p.factory = metrics.NewFactory(logger)

	// Create metrics provider
	if err := p.createProvider(); err != nil {
		return fmt.Errorf("failed to create metrics provider: %w", err)
	}

	p.logger.Info("Health aggregator plugin initialized",
		zap.String("provider", p.config.Provider.Type))

	return nil
}

// createProvider creates the configured metrics provider
func (p *Plugin) createProvider() error {
	var providerConfig *metrics.ProviderConfig

	switch p.config.Provider.Type {
	case "prometheus":
		if p.config.Provider.Prometheus == nil {
			return fmt.Errorf("prometheus configuration required")
		}
		providerConfig = &metrics.ProviderConfig{
			Type:   "prometheus",
			Config: p.config.Provider.Prometheus,
		}

	case "memory":
		if p.config.Provider.Memory == nil {
			// Use default memory config
			p.config.Provider.Memory = &memory.Config{
				SeedData:   true,
				DataPoints: 100,
			}
		}
		providerConfig = &metrics.ProviderConfig{
			Type:   "memory",
			Config: p.config.Provider.Memory,
		}

	default:
		return fmt.Errorf("unsupported provider type: %s", p.config.Provider.Type)
	}

	provider, err := p.factory.CreateProvider(providerConfig)
	if err != nil {
		return err
	}

	p.provider = provider
	return nil
}

// Invoke handles plugin invocation
func (p *Plugin) Invoke(ctx context.Context, req *pb.InvokeRequest) (*pb.InvokeResponse, error) {
	// Parse request
	var queryReq HealthQueryRequest
	if err := json.Unmarshal(req.Payload, &queryReq); err != nil {
		return nil, fmt.Errorf("failed to parse request: %w", err)
	}

	p.logger.Debug("Processing health query",
		zap.String("service", queryReq.ServiceName),
		zap.Strings("metrics", queryReq.Metrics))

	// Parse time range
	timeRange, err := p.parseTimeRange(queryReq.TimeRange)
	if err != nil {
		return nil, fmt.Errorf("invalid time range: %w", err)
	}

	// Check cache if enabled
	if p.config.Cache.Enabled {
		if cached := p.getCachedMetrics(queryReq.ServiceName); cached != nil {
			p.logger.Debug("Returning cached metrics",
				zap.String("service", queryReq.ServiceName))
			
			responseBytes, _ := json.Marshal(cached)
			return &pb.InvokeResponse{
				Payload: responseBytes,
			}, nil
		}
	}

	// Build metric queries
	queries := p.buildMetricQueries(queryReq, timeRange)

	// Execute queries in parallel
	results, warnings, err := p.executeQueries(ctx, queries)
	if err != nil {
		return nil, fmt.Errorf("failed to execute queries: %w", err)
	}

	// Build response
	response := &HealthQueryResponse{
		ServiceName: queryReq.ServiceName,
		Metrics:     results,
		Timestamp:   time.Now(),
		Warnings:    warnings,
	}

	// Cache response if enabled
	if p.config.Cache.Enabled {
		p.cacheMetrics(queryReq.ServiceName, response)
	}

	// Marshal response
	responseBytes, err := json.Marshal(response)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal response: %w", err)
	}

	return &pb.InvokeResponse{
		Payload: responseBytes,
	}, nil
}

// parseTimeRange parses a time range string
func (p *Plugin) parseTimeRange(timeRangeStr string) (metrics.TimeRange, error) {
	// Support various time range formats
	// Examples: "1h", "24h", "7d", "2024-01-01T00:00:00Z/2024-01-02T00:00:00Z"
	
	now := time.Now()
	
	// Check for duration format (e.g., "1h", "24h", "7d")
	if duration, err := time.ParseDuration(timeRangeStr); err == nil {
		return metrics.TimeRange{
			Start: now.Add(-duration),
			End:   now,
		}, nil
	}
	
	// Check for absolute time range (ISO format)
	// Format: "start/end"
	if len(timeRangeStr) > 0 && timeRangeStr[0] != '-' {
		// Simple parsing for demonstration
		// In production, use proper time parsing
		return metrics.TimeRange{
			Start: now.Add(-1 * time.Hour),
			End:   now,
		}, nil
	}
	
	// Default to last hour
	return metrics.TimeRange{
		Start: now.Add(-1 * time.Hour),
		End:   now,
	}, nil
}

// buildMetricQueries builds metric queries from request
func (p *Plugin) buildMetricQueries(req HealthQueryRequest, timeRange metrics.TimeRange) []metrics.MetricQuery {
	queries := make([]metrics.MetricQuery, 0, len(req.Metrics))
	
	for _, metric := range req.Metrics {
		// Build labels
		labels := make(map[string]string)
		if req.ServiceName != "" {
			labels["service"] = req.ServiceName
		}
		for k, v := range req.Filters {
			labels[k] = v
		}
		
		// Find metric configuration
		var metricConfig *MetricConfig
		for _, mc := range p.config.Metrics {
			if mc.Name == metric {
				metricConfig = &mc
				break
			}
		}
		
		// Build query
		query := metrics.MetricQuery{
			Metric:    metric,
			Labels:    labels,
			TimeRange: timeRange,
			Step:      p.config.Query.DefaultStep,
		}
		
		// Apply metric-specific configuration
		if metricConfig != nil {
			if metricConfig.Query != "" {
				query.Metric = metricConfig.Query
			}
			if metricConfig.Aggregation != "" {
				query.Aggregation = metricConfig.Aggregation
			}
			for k, v := range metricConfig.Labels {
				query.Labels[k] = v
			}
		}
		
		queries = append(queries, query)
	}
	
	return queries
}

// executeQueries executes metric queries in parallel
func (p *Plugin) executeQueries(ctx context.Context, queries []metrics.MetricQuery) (map[string]*MetricData, []string, error) {
	results := make(map[string]*MetricData)
	warnings := []string{}
	mu := sync.Mutex{}
	
	// Create context with timeout
	queryCtx, cancel := context.WithTimeout(ctx, p.config.Query.Timeout)
	defer cancel()
	
	// Execute queries in parallel with limit
	g, gCtx := errgroup.WithContext(queryCtx)
	g.SetLimit(p.config.Query.ParallelQueries)
	
	for _, query := range queries {
		query := query // Capture loop variable
		g.Go(func() error {
			// Execute query
			result, err := p.provider.Query(gCtx, query)
			if err != nil {
				p.logger.Warn("Query failed",
					zap.String("metric", query.Metric),
					zap.Error(err))
				
				mu.Lock()
				warnings = append(warnings, fmt.Sprintf("Query failed for %s: %v", query.Metric, err))
				mu.Unlock()
				return nil // Don't fail entire batch
			}
			
			// Convert result to MetricData
			metricData := p.convertToMetricData(query.Metric, result)
			
			mu.Lock()
			results[query.Metric] = metricData
			if len(result.Warnings) > 0 {
				warnings = append(warnings, result.Warnings...)
			}
			mu.Unlock()
			
			return nil
		})
	}
	
	if err := g.Wait(); err != nil {
		return nil, warnings, err
	}
	
	return results, warnings, nil
}

// convertToMetricData converts QueryResult to MetricData
func (p *Plugin) convertToMetricData(name string, result *metrics.QueryResult) *MetricData {
	data := &MetricData{
		Name:   name,
		Values: []DataPoint{},
	}
	
	// Aggregate all series into a single time series
	allValues := []float64{}
	
	for _, series := range result.Series {
		for _, point := range series.Values {
			data.Values = append(data.Values, DataPoint{
				Timestamp: point.Timestamp,
				Value:     point.Value,
			})
			allValues = append(allValues, point.Value)
		}
	}
	
	// Calculate statistics if we have values
	if len(allValues) > 0 {
		data.Statistics = p.calculateStatistics(allValues)
	}
	
	return data
}

// calculateStatistics calculates statistics for a set of values
func (p *Plugin) calculateStatistics(values []float64) *Statistics {
	if len(values) == 0 {
		return nil
	}
	
	stats := &Statistics{
		Count: len(values),
	}
	
	// Calculate min, max, sum
	sum := 0.0
	stats.Min = values[0]
	stats.Max = values[0]
	
	for _, v := range values {
		sum += v
		if v < stats.Min {
			stats.Min = v
		}
		if v > stats.Max {
			stats.Max = v
		}
	}
	
	stats.Avg = sum / float64(len(values))
	
	// Calculate percentiles if enough data points
	if len(values) >= 10 {
		// For percentiles, we need a sorted copy
		sortedValues := make([]float64, len(values))
		copy(sortedValues, values)
		sort.Float64s(sortedValues)
		
		stats.P50 = p.percentile(sortedValues, 0.5)
		stats.P95 = p.percentile(sortedValues, 0.95)
		stats.P99 = p.percentile(sortedValues, 0.99)
	}
	
	return stats
}

// percentile calculates the percentile value from sorted values
func (p *Plugin) percentile(sortedValues []float64, percentile float64) float64 {
	if len(sortedValues) == 0 {
		return 0
	}
	
	index := int(float64(len(sortedValues)-1) * percentile)
	return sortedValues[index]
}

// getCachedMetrics retrieves cached metrics if available and not expired
func (p *Plugin) getCachedMetrics(serviceName string) *HealthQueryResponse {
	p.mu.RLock()
	defer p.mu.RUnlock()
	
	cached, exists := p.metricsCache[serviceName]
	if !exists {
		return nil
	}
	
	// Check if cache is expired
	if time.Since(cached.Timestamp) > p.cacheDuration {
		return nil
	}
	
	// Type assert to response
	response, ok := cached.Data.(*HealthQueryResponse)
	if !ok {
		return nil
	}
	
	return response
}

// cacheMetrics stores metrics in cache
func (p *Plugin) cacheMetrics(serviceName string, response *HealthQueryResponse) {
	p.mu.Lock()
	defer p.mu.Unlock()
	
	p.metricsCache[serviceName] = &CachedMetrics{
		Data:      response,
		Timestamp: time.Now(),
	}
	
	// Clean up old cache entries
	p.cleanupCache()
}

// cleanupCache removes expired cache entries
func (p *Plugin) cleanupCache() {
	now := time.Now()
	for key, cached := range p.metricsCache {
		if now.Sub(cached.Timestamp) > p.cacheDuration*2 {
			delete(p.metricsCache, key)
		}
	}
}

// Close cleans up resources
func (p *Plugin) Close(ctx context.Context) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	
	// Clear cache
	p.metricsCache = make(map[string]*CachedMetrics)
	
	// Close provider
	if p.provider != nil {
		if err := p.provider.Close(); err != nil {
			p.logger.Warn("Failed to close metrics provider",
				zap.Error(err))
		}
	}
	
	// Close factory
	if p.factory != nil {
		if err := p.factory.Close(); err != nil {
			p.logger.Warn("Failed to close metrics factory",
				zap.Error(err))
		}
	}
	
	p.logger.Info("Health aggregator plugin closed")
	return nil
}

// GetInfo returns plugin information
func (p *Plugin) GetInfo() framework.PluginInfo {
	return framework.PluginInfo{
		Name:        "health_aggregator",
		Version:     "1.0.0",
		Description: "Aggregates health metrics from various sources",
		Category:    "observability",
		Capabilities: []string{
			"metrics_query",
			"batch_query",
			"aggregation",
			"caching",
		},
	}
}

// HealthCheck checks the health of the plugin
func (p *Plugin) HealthCheck(ctx context.Context) error {
	// Check if provider is healthy
	if p.provider == nil {
		return fmt.Errorf("metrics provider not initialized")
	}
	
	if err := p.provider.HealthCheck(ctx); err != nil {
		return fmt.Errorf("metrics provider unhealthy: %w", err)
	}
	
	return nil
}