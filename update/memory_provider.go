// Package memory implements an in-memory MetricsProvider for testing
package memory

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"sort"
	"sync"
	"time"

	"github.com/detectviz/go-platform/internal/metrics"
)

// Config contains configuration for the memory provider
type Config struct {
	// SeedData determines if the provider should be pre-populated with test data
	SeedData bool `json:"seed_data" yaml:"seed_data"`

	// DataPoints is the number of data points to generate per metric
	DataPoints int `json:"data_points" yaml:"data_points"`

	// Metrics to pre-populate
	Metrics []string `json:"metrics" yaml:"metrics"`

	// Services to simulate
	Services []string `json:"services" yaml:"services"`
}

// Provider implements an in-memory metrics provider for testing
type Provider struct {
	config *Config
	data   map[string][]metrics.TimeSeries
	mu     sync.RWMutex
}

// NewProvider creates a new memory metrics provider
func NewProvider(config *Config) *Provider {
	if config == nil {
		config = &Config{
			SeedData:   true,
			DataPoints: 100,
			Metrics: []string{
				"cpu_usage",
				"memory_usage",
				"http_request_duration_seconds",
				"http_requests_total",
				"error_rate",
				"disk_usage",
				"network_throughput",
			},
			Services: []string{
				"api-gateway",
				"auth-service",
				"payment-service",
				"notification-service",
				"database",
			},
		}
	}

	p := &Provider{
		config: config,
		data:   make(map[string][]metrics.TimeSeries),
	}

	if config.SeedData {
		p.seedData()
	}

	return p
}

// seedData populates the provider with test data
func (p *Provider) seedData() {
	now := time.Now()
	interval := 1 * time.Minute

	for _, metric := range p.config.Metrics {
		var seriesList []metrics.TimeSeries

		for _, service := range p.config.Services {
			series := metrics.TimeSeries{
				Labels: map[string]string{
					"service":     service,
					"environment": "production",
					"region":      "us-east-1",
				},
				Values: []metrics.DataPoint{},
			}

			// Generate data points
			for i := 0; i < p.config.DataPoints; i++ {
				timestamp := now.Add(-time.Duration(p.config.DataPoints-i) * interval)
				value := p.generateValue(metric, service, i)
				
				series.Values = append(series.Values, metrics.DataPoint{
					Timestamp: timestamp,
					Value:     value,
				})
			}

			seriesList = append(seriesList, series)
		}

		p.data[metric] = seriesList
	}
}

// generateValue generates a realistic value for a metric
func (p *Provider) generateValue(metric, service string, index int) float64 {
	// Base value with some randomness
	base := rand.Float64()
	
	// Add patterns based on metric type
	switch metric {
	case "cpu_usage":
		// CPU usage between 20-80% with spikes
		value := 0.2 + base*0.4
		if index%20 == 0 { // Occasional spike
			value += 0.3
		}
		// Add sine wave for daily pattern
		value += 0.1 * math.Sin(float64(index)*0.1)
		return math.Min(value, 1.0)

	case "memory_usage":
		// Memory usage with gradual increase (simulating memory leak)
		value := 0.3 + base*0.3 + float64(index)*0.001
		return math.Min(value, 0.95)

	case "http_request_duration_seconds":
		// Request duration in seconds (50ms to 500ms)
		value := 0.05 + base*0.3
		if service == "database" {
			value *= 2 // Database queries are slower
		}
		if index%30 == 0 { // Occasional slow request
			value *= 3
		}
		return value

	case "http_requests_total":
		// Request count (100-1000 per minute)
		value := 100 + base*900
		// Add daily pattern
		value *= (1 + 0.5*math.Sin(float64(index)*0.05))
		return value

	case "error_rate":
		// Error rate (0-5% normally, with occasional spikes)
		value := base * 0.02
		if index%50 == 0 { // Error spike
			value = 0.1 + base*0.1
		}
		return value

	case "disk_usage":
		// Disk usage with gradual increase
		value := 0.4 + float64(index)*0.002 + base*0.1
		return math.Min(value, 0.9)

	case "network_throughput":
		// Network throughput in MB/s
		value := 10 + base*50
		// Add daily pattern
		value *= (1 + 0.3*math.Sin(float64(index)*0.05))
		return value

	default:
		return base
	}
}

// Query executes a single metric query
func (p *Provider) Query(ctx context.Context, query metrics.MetricQuery) (*metrics.QueryResult, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	// Check if metric exists
	seriesList, exists := p.data[query.Metric]
	if !exists {
		return &metrics.QueryResult{
			Query:  &query,
			Series: []metrics.TimeSeries{},
		}, nil
	}

	// Filter series based on labels
	var filteredSeries []metrics.TimeSeries
	for _, series := range seriesList {
		if p.matchLabels(series.Labels, query.Labels) {
			// Filter by time range
			filtered := p.filterByTimeRange(series, query.TimeRange)
			if len(filtered.Values) > 0 {
				// Apply aggregation if specified
				if query.Aggregation != "" {
					filtered = p.applyAggregation(filtered, query.Aggregation)
				}
				filteredSeries = append(filteredSeries, filtered)
			}
		}
	}

	return &metrics.QueryResult{
		Query:    &query,
		Series:   filteredSeries,
		Warnings: []string{},
		Stats: &metrics.QueryStats{
			Duration:       10 * time.Millisecond, // Simulate query time
			SamplesScanned: int64(len(filteredSeries) * p.config.DataPoints),
			SeriesScanned:  int64(len(seriesList)),
		},
	}, nil
}

// BatchQuery executes multiple queries in parallel
func (p *Provider) BatchQuery(ctx context.Context, queries []metrics.MetricQuery) ([]*metrics.QueryResult, error) {
	results := make([]*metrics.QueryResult, len(queries))
	
	// Simulate parallel execution
	var wg sync.WaitGroup
	for i, query := range queries {
		wg.Add(1)
		go func(idx int, q metrics.MetricQuery) {
			defer wg.Done()
			result, _ := p.Query(ctx, q)
			results[idx] = result
		}(i, query)
	}
	wg.Wait()

	return results, nil
}

// GetAggregation performs aggregation queries
func (p *Provider) GetAggregation(ctx context.Context, opts metrics.AggregationOptions) (*metrics.AggregationResult, error) {
	result := &metrics.AggregationResult{
		Options: &opts,
		Values:  []metrics.AggregatedValue{},
	}

	for _, metric := range opts.Metrics {
		query := metrics.MetricQuery{
			Metric:    metric,
			Labels:    opts.Filters,
			TimeRange: opts.TimeRange,
		}

		queryResult, err := p.Query(ctx, query)
		if err != nil {
			return nil, err
		}

		// Aggregate all series
		for _, series := range queryResult.Series {
			aggValue := p.aggregateSeries(series, opts.Function)
			result.Values = append(result.Values, metrics.AggregatedValue{
				Metric:      metric,
				Labels:      series.Labels,
				Value:       aggValue,
				SampleCount: int64(len(series.Values)),
			})
		}
	}

	return result, nil
}

// HealthCheck always returns healthy for memory provider
func (p *Provider) HealthCheck(ctx context.Context) error {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if len(p.data) == 0 {
		return fmt.Errorf("no data available")
	}

	return nil
}

// Close is a no-op for memory provider
func (p *Provider) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Clear data
	p.data = make(map[string][]metrics.TimeSeries)
	return nil
}

// AddMetric adds a new metric to the provider (for testing)
func (p *Provider) AddMetric(metric string, series metrics.TimeSeries) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if _, exists := p.data[metric]; !exists {
		p.data[metric] = []metrics.TimeSeries{}
	}
	p.data[metric] = append(p.data[metric], series)
}

// ClearData removes all data from the provider
func (p *Provider) ClearData() {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.data = make(map[string][]metrics.TimeSeries)
}

// matchLabels checks if series labels match query labels
func (p *Provider) matchLabels(seriesLabels, queryLabels map[string]string) bool {
	for k, v := range queryLabels {
		if seriesLabels[k] != v {
			return false
		}
	}
	return true
}

// filterByTimeRange filters data points by time range
func (p *Provider) filterByTimeRange(series metrics.TimeSeries, timeRange metrics.TimeRange) metrics.TimeSeries {
	filtered := metrics.TimeSeries{
		Labels: series.Labels,
		Values: []metrics.DataPoint{},
	}

	for _, point := range series.Values {
		if point.Timestamp.After(timeRange.Start) && point.Timestamp.Before(timeRange.End) {
			filtered.Values = append(filtered.Values, point)
		}
	}

	return filtered
}

// applyAggregation applies an aggregation function to a series
func (p *Provider) applyAggregation(series metrics.TimeSeries, aggregation string) metrics.TimeSeries {
	if len(series.Values) == 0 {
		return series
	}

	value := p.aggregateSeries(series, metrics.AggregationFunc(aggregation))
	
	// Return single aggregated value at the end timestamp
	return metrics.TimeSeries{
		Labels: series.Labels,
		Values: []metrics.DataPoint{
			{
				Timestamp: series.Values[len(series.Values)-1].Timestamp,
				Value:     value,
			},
		},
	}
}

// aggregateSeries calculates an aggregate value for a series
func (p *Provider) aggregateSeries(series metrics.TimeSeries, function metrics.AggregationFunc) float64 {
	if len(series.Values) == 0 {
		return 0
	}

	values := make([]float64, len(series.Values))
	for i, point := range series.Values {
		values[i] = point.Value
	}

	switch function {
	case metrics.AggregationAvg:
		sum := 0.0
		for _, v := range values {
			sum += v
		}
		return sum / float64(len(values))

	case metrics.AggregationMax:
		max := values[0]
		for _, v := range values {
			if v > max {
				max = v
			}
		}
		return max

	case metrics.AggregationMin:
		min := values[0]
		for _, v := range values {
			if v < min {
				min = v
			}
		}
		return min

	case metrics.AggregationSum:
		sum := 0.0
		for _, v := range values {
			sum += v
		}
		return sum

	case metrics.AggregationCount:
		return float64(len(values))

	case metrics.AggregationP50:
		return p.percentile(values, 0.5)

	case metrics.AggregationP95:
		return p.percentile(values, 0.95)

	case metrics.AggregationP99:
		return p.percentile(values, 0.99)

	default:
		// Default to average
		sum := 0.0
		for _, v := range values {
			sum += v
		}
		return sum / float64(len(values))
	}
}

// percentile calculates the percentile value
func (p *Provider) percentile(values []float64, percentile float64) float64 {
	if len(values) == 0 {
		return 0
	}

	// Sort values
	sorted := make([]float64, len(values))
	copy(sorted, values)
	sort.Float64s(sorted)

	// Calculate percentile index
	index := int(float64(len(sorted)-1) * percentile)
	return sorted[index]
}

// GetMetrics returns all available metrics (for testing)
func (p *Provider) GetMetrics() []string {
	p.mu.RLock()
	defer p.mu.RUnlock()

	metrics := make([]string, 0, len(p.data))
	for metric := range p.data {
		metrics = append(metrics, metric)
	}
	sort.Strings(metrics)
	return metrics
}

// GetServices returns all available services (for testing)
func (p *Provider) GetServices() []string {
	p.mu.RLock()
	defer p.mu.RUnlock()

	serviceMap := make(map[string]bool)
	for _, seriesList := range p.data {
		for _, series := range seriesList {
			if service, ok := series.Labels["service"]; ok {
				serviceMap[service] = true
			}
		}
	}

	services := make([]string, 0, len(serviceMap))
	for service := range serviceMap {
		services = append(services, service)
	}
	sort.Strings(services)
	return services
}