package metrics

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/detectviz/go-platform/internal/metrics/memory"
	"github.com/detectviz/go-platform/internal/metrics/prometheus"
	"go.uber.org/zap"
)

// Factory creates and manages metrics providers
type Factory struct {
	providers map[string]MetricsProvider
	configs   map[string]*ProviderConfig
	logger    *zap.Logger
	mu        sync.RWMutex
}

// NewFactory creates a new metrics provider factory
func NewFactory(logger *zap.Logger) *Factory {
	return &Factory{
		providers: make(map[string]MetricsProvider),
		configs:   make(map[string]*ProviderConfig),
		logger:    logger,
	}
}

// CreateProvider creates a new metrics provider based on configuration
func (f *Factory) CreateProvider(config *ProviderConfig) (MetricsProvider, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	// Check if provider already exists
	if provider, exists := f.providers[config.Type]; exists {
		return provider, nil
	}

	var provider MetricsProvider
	var err error

	switch config.Type {
	case "prometheus":
		promConfig, ok := config.Config.(*prometheus.Config)
		if !ok {
			return nil, fmt.Errorf("invalid prometheus configuration")
		}
		provider, err = prometheus.NewProvider(promConfig, f.logger)

	case "mimir":
		// Reserved for future implementation
		return nil, fmt.Errorf("mimir provider not yet implemented")

	case "memory":
		memConfig, ok := config.Config.(*memory.Config)
		if !ok {
			// Use default config if not provided
			memConfig = &memory.Config{
				SeedData:   true,
				DataPoints: 1000,
			}
		}
		provider = memory.NewProvider(memConfig)

	default:
		return nil, fmt.Errorf("unsupported provider type: %s", config.Type)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create %s provider: %w", config.Type, err)
	}

	// Store provider for reuse
	f.providers[config.Type] = provider
	f.configs[config.Type] = config

	f.logger.Info("Created metrics provider",
		zap.String("type", config.Type))

	return provider, nil
}

// GetProvider returns an existing provider by type
func (f *Factory) GetProvider(providerType string) (MetricsProvider, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()

	provider, exists := f.providers[providerType]
	if !exists {
		return nil, fmt.Errorf("provider %s not found", providerType)
	}

	return provider, nil
}

// CreateRoutingProvider creates a provider that routes queries based on time range
// This will be used when Mimir is implemented for long-term storage
func (f *Factory) CreateRoutingProvider(config *QueryRouterConfig) (MetricsProvider, error) {
	if config == nil {
		return nil, fmt.Errorf("routing config is required")
	}

	shortTermProvider, err := f.GetProvider(config.ShortTermProvider)
	if err != nil {
		return nil, fmt.Errorf("failed to get short-term provider: %w", err)
	}

	// For now, just return the short-term provider
	// When Mimir is implemented, this will create a RoutingProvider
	if config.LongTermProvider != "" && config.LongTermProvider != config.ShortTermProvider {
		f.logger.Warn("Long-term provider specified but not yet implemented",
			zap.String("provider", config.LongTermProvider))
	}

	return &RoutingProvider{
		shortTerm: shortTermProvider,
		longTerm:  nil, // Will be set when Mimir is implemented
		cutoff:    time.Duration(config.ShortTermDays) * 24 * time.Hour,
		logger:    f.logger,
	}, nil
}

// Close closes all managed providers
func (f *Factory) Close() error {
	f.mu.Lock()
	defer f.mu.Unlock()

	var errors []error
	for name, provider := range f.providers {
		if err := provider.Close(); err != nil {
			errors = append(errors, fmt.Errorf("failed to close %s provider: %w", name, err))
		}
	}

	// Clear the maps
	f.providers = make(map[string]MetricsProvider)
	f.configs = make(map[string]*ProviderConfig)

	if len(errors) > 0 {
		return fmt.Errorf("errors closing providers: %v", errors)
	}

	return nil
}

// RoutingProvider routes queries to appropriate providers based on time range
type RoutingProvider struct {
	shortTerm MetricsProvider
	longTerm  MetricsProvider // Will be nil until Mimir is implemented
	cutoff    time.Duration
	logger    *zap.Logger
}

// Query routes the query to the appropriate provider based on time range
func (r *RoutingProvider) Query(ctx context.Context, query MetricQuery) (*QueryResult, error) {
	provider := r.selectProvider(query.TimeRange)
	return provider.Query(ctx, query)
}

// BatchQuery routes batch queries to the appropriate provider
func (r *RoutingProvider) BatchQuery(ctx context.Context, queries []MetricQuery) ([]*QueryResult, error) {
	// Group queries by provider
	shortTermQueries := []MetricQuery{}
	longTermQueries := []MetricQuery{}

	for _, query := range queries {
		if r.shouldUseShortTerm(query.TimeRange) {
			shortTermQueries = append(shortTermQueries, query)
		} else {
			longTermQueries = append(longTermQueries, query)
		}
	}

	results := make([]*QueryResult, 0, len(queries))

	// Execute short-term queries
	if len(shortTermQueries) > 0 {
		shortResults, err := r.shortTerm.BatchQuery(ctx, shortTermQueries)
		if err != nil {
			return nil, fmt.Errorf("short-term batch query failed: %w", err)
		}
		results = append(results, shortResults...)
	}

	// Execute long-term queries (when available)
	if len(longTermQueries) > 0 {
		if r.longTerm != nil {
			longResults, err := r.longTerm.BatchQuery(ctx, longTermQueries)
			if err != nil {
				return nil, fmt.Errorf("long-term batch query failed: %w", err)
			}
			results = append(results, longResults...)
		} else {
			// Fallback to short-term provider with warning
			r.logger.Warn("Long-term queries falling back to short-term provider",
				zap.Int("query_count", len(longTermQueries)))
			
			longResults, err := r.shortTerm.BatchQuery(ctx, longTermQueries)
			if err != nil {
				return nil, fmt.Errorf("fallback batch query failed: %w", err)
			}
			results = append(results, longResults...)
		}
	}

	return results, nil
}

// GetAggregation routes aggregation queries to the appropriate provider
func (r *RoutingProvider) GetAggregation(ctx context.Context, opts AggregationOptions) (*AggregationResult, error) {
	provider := r.selectProvider(opts.TimeRange)
	return provider.GetAggregation(ctx, opts)
}

// HealthCheck checks the health of all available providers
func (r *RoutingProvider) HealthCheck(ctx context.Context) error {
	// Check short-term provider
	if err := r.shortTerm.HealthCheck(ctx); err != nil {
		return fmt.Errorf("short-term provider health check failed: %w", err)
	}

	// Check long-term provider if available
	if r.longTerm != nil {
		if err := r.longTerm.HealthCheck(ctx); err != nil {
			// Log warning but don't fail if long-term is unhealthy
			r.logger.Warn("Long-term provider health check failed",
				zap.Error(err))
		}
	}

	return nil
}

// Close closes all providers
func (r *RoutingProvider) Close() error {
	var errors []error

	if err := r.shortTerm.Close(); err != nil {
		errors = append(errors, fmt.Errorf("failed to close short-term provider: %w", err))
	}

	if r.longTerm != nil {
		if err := r.longTerm.Close(); err != nil {
			errors = append(errors, fmt.Errorf("failed to close long-term provider: %w", err))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("errors closing providers: %v", errors)
	}

	return nil
}

// selectProvider selects the appropriate provider based on time range
func (r *RoutingProvider) selectProvider(timeRange TimeRange) MetricsProvider {
	if r.shouldUseShortTerm(timeRange) {
		return r.shortTerm
	}

	if r.longTerm != nil {
		return r.longTerm
	}

	// Fallback to short-term if long-term not available
	return r.shortTerm
}

// shouldUseShortTerm determines if the query should use the short-term provider
func (r *RoutingProvider) shouldUseShortTerm(timeRange TimeRange) bool {
	age := time.Since(timeRange.Start)
	return age <= r.cutoff
}