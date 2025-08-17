// Package prometheus implements the MetricsProvider interface for Prometheus
package prometheus

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/detectviz/go-platform/internal/metrics"
	"github.com/patrickmn/go-cache"
	"github.com/prometheus/client_golang/api"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/config"
	"github.com/prometheus/common/model"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

// Config contains configuration for the Prometheus provider
type Config struct {
	// Prometheus server URL
	URL string `json:"url" yaml:"url"`

	// Query timeout
	Timeout time.Duration `json:"timeout" yaml:"timeout"`

	// Retry configuration
	Retry RetryConfig `json:"retry" yaml:"retry"`

	// Query configuration
	Query QueryConfig `json:"query" yaml:"query"`

	// Basic authentication
	BasicAuth *BasicAuthConfig `json:"basic_auth,omitempty" yaml:"basic_auth,omitempty"`

	// TLS configuration
	TLS *TLSConfig `json:"tls,omitempty" yaml:"tls,omitempty"`
}

// RetryConfig contains retry configuration
type RetryConfig struct {
	Enabled     bool          `json:"enabled" yaml:"enabled"`
	MaxAttempts int           `json:"max_attempts" yaml:"max_attempts"`
	Backoff     time.Duration `json:"backoff" yaml:"backoff"`
}

// QueryConfig contains query-specific configuration
type QueryConfig struct {
	MaxSamples     int           `json:"max_samples" yaml:"max_samples"`
	MaxConcurrent  int           `json:"max_concurrent" yaml:"max_concurrent"`
	CacheEnabled   bool          `json:"cache_enabled" yaml:"cache_enabled"`
	CacheTTL       time.Duration `json:"cache_ttl" yaml:"cache_ttl"`
	CacheMaxSize   int           `json:"cache_max_size" yaml:"cache_max_size"`
}

// BasicAuthConfig contains basic authentication configuration
type BasicAuthConfig struct {
	Username string `json:"username" yaml:"username"`
	Password string `json:"password" yaml:"password"`
}

// TLSConfig contains TLS configuration
type TLSConfig struct {
	Enabled            bool   `json:"enabled" yaml:"enabled"`
	InsecureSkipVerify bool   `json:"insecure_skip_verify" yaml:"insecure_skip_verify"`
	CertFile           string `json:"cert_file,omitempty" yaml:"cert_file,omitempty"`
	KeyFile            string `json:"key_file,omitempty" yaml:"key_file,omitempty"`
	CAFile             string `json:"ca_file,omitempty" yaml:"ca_file,omitempty"`
}

// Provider implements the MetricsProvider interface for Prometheus
type Provider struct {
	client     v1.API
	httpClient *http.Client
	config     *Config
	cache      *cache.Cache
	logger     *zap.Logger
	semaphore  chan struct{} // For concurrent query limiting
	mu         sync.RWMutex
}

// NewProvider creates a new Prometheus metrics provider
func NewProvider(cfg *Config, logger *zap.Logger) (*Provider, error) {
	// Create HTTP client with custom transport if needed
	roundTripper := api.DefaultRoundTripper

	// Configure basic auth if provided
	if cfg.BasicAuth != nil {
		roundTripper = config.NewBasicAuthRoundTripper(
			cfg.BasicAuth.Username,
			config.Secret(cfg.BasicAuth.Password),
			"", // Empty string for password file
			roundTripper,
		)
	}

	// Create Prometheus client
	client, err := api.NewClient(api.Config{
		Address:      cfg.URL,
		RoundTripper: roundTripper,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create Prometheus client: %w", err)
	}

	// Create cache if enabled
	var c *cache.Cache
	if cfg.Query.CacheEnabled {
		c = cache.New(cfg.Query.CacheTTL, cfg.Query.CacheTTL*2)
	}

	// Create semaphore for concurrent query limiting
	semaphore := make(chan struct{}, cfg.Query.MaxConcurrent)
	for i := 0; i < cfg.Query.MaxConcurrent; i++ {
		semaphore <- struct{}{}
	}

	return &Provider{
		client:     v1.NewAPI(client),
		httpClient: &http.Client{Timeout: cfg.Timeout},
		config:     cfg,
		cache:      c,
		logger:     logger,
		semaphore:  semaphore,
	}, nil
}

// Query executes a single metric query
func (p *Provider) Query(ctx context.Context, query metrics.MetricQuery) (*metrics.QueryResult, error) {
	// Check cache if enabled
	if p.cache != nil {
		cacheKey := p.buildCacheKey(query)
		if cached, found := p.cache.Get(cacheKey); found {
			p.logger.Debug("Cache hit for query",
				zap.String("metric", query.Metric))
			return cached.(*metrics.QueryResult), nil
		}
	}

	// Build PromQL query
	promQL := p.buildPromQL(query)
	
	// Acquire semaphore
	select {
	case <-p.semaphore:
		defer func() { p.semaphore <- struct{}{} }()
	case <-ctx.Done():
		return nil, ctx.Err()
	}

	// Execute query
	start := time.Now()
	value, warnings, err := p.client.Query(ctx, promQL, query.TimeRange.End)
	if err != nil {
		return nil, fmt.Errorf("prometheus query failed: %w", err)
	}

	// Convert result
	result := p.convertResult(query, value, warnings)
	result.Stats = &metrics.QueryStats{
		Duration: time.Since(start),
	}

	// Cache result if enabled
	if p.cache != nil {
		cacheKey := p.buildCacheKey(query)
		p.cache.Set(cacheKey, result, cache.DefaultExpiration)
	}

	return result, nil
}

// BatchQuery executes multiple queries in parallel
func (p *Provider) BatchQuery(ctx context.Context, queries []metrics.MetricQuery) ([]*metrics.QueryResult, error) {
	results := make([]*metrics.QueryResult, len(queries))
	g, ctx := errgroup.WithContext(ctx)

	for i, query := range queries {
		i, query := i, query // Capture loop variables
		g.Go(func() error {
			result, err := p.Query(ctx, query)
			if err != nil {
				return fmt.Errorf("query %d failed: %w", i, err)
			}
			results[i] = result
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return nil, err
	}

	return results, nil
}

// GetAggregation performs aggregation queries
func (p *Provider) GetAggregation(ctx context.Context, opts metrics.AggregationOptions) (*metrics.AggregationResult, error) {
	result := &metrics.AggregationResult{
		Options: &opts,
		Values:  []metrics.AggregatedValue{},
	}

	for _, metric := range opts.Metrics {
		// Build aggregation query
		promQL := p.buildAggregationQuery(metric, opts)

		// Execute query
		value, warnings, err := p.client.Query(ctx, promQL, opts.TimeRange.End)
		if err != nil {
			return nil, fmt.Errorf("aggregation query failed for %s: %w", metric, err)
		}

		result.Warnings = append(result.Warnings, warnings...)

		// Convert to aggregated values
		aggValues := p.convertAggregationResult(metric, value, opts)
		result.Values = append(result.Values, aggValues...)
	}

	return result, nil
}

// HealthCheck verifies Prometheus is accessible
func (p *Provider) HealthCheck(ctx context.Context) error {
	// Use Prometheus targets API to check health
	targets, err := p.client.Targets(ctx)
	if err != nil {
		return fmt.Errorf("prometheus health check failed: %w", err)
	}

	// Check if any targets are up
	hasHealthyTargets := false
	for _, target := range targets.Active {
		if target.Health == v1.HealthGood {
			hasHealthyTargets = true
			break
		}
	}

	if !hasHealthyTargets {
		p.logger.Warn("No healthy Prometheus targets found")
	}

	return nil
}

// Close releases resources
func (p *Provider) Close() error {
	// Clear cache if exists
	if p.cache != nil {
		p.cache.Flush()
	}

	// Close semaphore channel
	close(p.semaphore)

	return nil
}

// buildPromQL constructs a PromQL query from MetricQuery
func (p *Provider) buildPromQL(query metrics.MetricQuery) string {
	// Start with metric name
	promQL := query.Metric

	// Add label selectors
	if len(query.Labels) > 0 {
		selectors := ""
		for k, v := range query.Labels {
			if selectors != "" {
				selectors += ","
			}
			selectors += fmt.Sprintf(`%s="%s"`, k, v)
		}
		promQL = fmt.Sprintf("%s{%s}", promQL, selectors)
	}

	// Add aggregation if specified
	if query.Aggregation != "" {
		switch query.Aggregation {
		case "avg":
			promQL = fmt.Sprintf("avg(%s)", promQL)
		case "max":
			promQL = fmt.Sprintf("max(%s)", promQL)
		case "min":
			promQL = fmt.Sprintf("min(%s)", promQL)
		case "sum":
			promQL = fmt.Sprintf("sum(%s)", promQL)
		case "count":
			promQL = fmt.Sprintf("count(%s)", promQL)
		case "p50":
			promQL = fmt.Sprintf("quantile(0.5, %s)", promQL)
		case "p95":
			promQL = fmt.Sprintf("quantile(0.95, %s)", promQL)
		case "p99":
			promQL = fmt.Sprintf("quantile(0.99, %s)", promQL)
		}
	}

	return promQL
}

// buildAggregationQuery builds a PromQL query for aggregation
func (p *Provider) buildAggregationQuery(metric string, opts metrics.AggregationOptions) string {
	base := metric

	// Add filters
	if len(opts.Filters) > 0 {
		selectors := ""
		for k, v := range opts.Filters {
			if selectors != "" {
				selectors += ","
			}
			selectors += fmt.Sprintf(`%s="%s"`, k, v)
		}
		base = fmt.Sprintf("%s{%s}", metric, selectors)
	}

	// Apply aggregation function
	var promQL string
	switch opts.Function {
	case metrics.AggregationAvg:
		promQL = fmt.Sprintf("avg(%s)", base)
	case metrics.AggregationMax:
		promQL = fmt.Sprintf("max(%s)", base)
	case metrics.AggregationMin:
		promQL = fmt.Sprintf("min(%s)", base)
	case metrics.AggregationSum:
		promQL = fmt.Sprintf("sum(%s)", base)
	case metrics.AggregationCount:
		promQL = fmt.Sprintf("count(%s)", base)
	case metrics.AggregationP50:
		promQL = fmt.Sprintf("quantile(0.5, %s)", base)
	case metrics.AggregationP95:
		promQL = fmt.Sprintf("quantile(0.95, %s)", base)
	case metrics.AggregationP99:
		promQL = fmt.Sprintf("quantile(0.99, %s)", base)
	default:
		promQL = base
	}

	// Add grouping if specified
	if len(opts.GroupBy) > 0 {
		grouping := ""
		for i, label := range opts.GroupBy {
			if i > 0 {
				grouping += ","
			}
			grouping += label
		}
		promQL = fmt.Sprintf("%s by (%s)", promQL, grouping)
	}

	return promQL
}

// convertResult converts Prometheus query result to QueryResult
func (p *Provider) convertResult(query metrics.MetricQuery, value model.Value, warnings []string) *metrics.QueryResult {
	result := &metrics.QueryResult{
		Query:    &query,
		Series:   []metrics.TimeSeries{},
		Warnings: warnings,
	}

	switch v := value.(type) {
	case model.Vector:
		for _, sample := range v {
			ts := metrics.TimeSeries{
				Labels: make(map[string]string),
				Values: []metrics.DataPoint{
					{
						Timestamp: sample.Timestamp.Time(),
						Value:     float64(sample.Value),
					},
				},
			}
			for k, v := range sample.Metric {
				ts.Labels[string(k)] = string(v)
			}
			result.Series = append(result.Series, ts)
		}

	case model.Matrix:
		for _, series := range v {
			ts := metrics.TimeSeries{
				Labels: make(map[string]string),
				Values: []metrics.DataPoint{},
			}
			for k, v := range series.Metric {
				ts.Labels[string(k)] = string(v)
			}
			for _, pair := range series.Values {
				ts.Values = append(ts.Values, metrics.DataPoint{
					Timestamp: pair.Timestamp.Time(),
					Value:     float64(pair.Value),
				})
			}
			result.Series = append(result.Series, ts)
		}
	}

	return result
}

// convertAggregationResult converts Prometheus result to aggregated values
func (p *Provider) convertAggregationResult(metric string, value model.Value, opts metrics.AggregationOptions) []metrics.AggregatedValue {
	var values []metrics.AggregatedValue

	switch v := value.(type) {
	case model.Vector:
		for _, sample := range v {
			aggValue := metrics.AggregatedValue{
				Metric: metric,
				Value:  float64(sample.Value),
				Labels: make(map[string]string),
			}
			for k, v := range sample.Metric {
				aggValue.Labels[string(k)] = string(v)
			}
			values = append(values, aggValue)
		}
	}

	return values
}

// buildCacheKey creates a cache key from a query
func (p *Provider) buildCacheKey(query metrics.MetricQuery) string {
	return fmt.Sprintf("%s_%v_%d_%d_%s",
		query.Metric,
		query.Labels,
		query.TimeRange.Start.Unix(),
		query.TimeRange.End.Unix(),
		query.Aggregation,
	)
}