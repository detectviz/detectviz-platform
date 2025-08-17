package health_aggregator

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/detectviz/go-platform/internal/metrics/memory"
	pb "github.com/detectviz/go-platform/pkg/contracts/gen/go/detectviz/contracts/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

func TestPlugin_Initialize(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
	}{
		{
			name: "valid prometheus config",
			config: &Config{
				Provider: ProviderConfig{
					Type: "prometheus",
					Prometheus: &prometheus.Config{
						URL:     "http://localhost:9090",
						Timeout: 30 * time.Second,
					},
				},
				Query: QueryConfig{
					ParallelQueries: 10,
					Timeout:         30 * time.Second,
					DefaultStep:     1 * time.Minute,
				},
				Cache: CacheConfig{
					Enabled:  true,
					Duration: 5 * time.Minute,
				},
			},
			wantErr: false,
		},
		{
			name: "valid memory config",
			config: &Config{
				Provider: ProviderConfig{
					Type: "memory",
					Memory: &memory.Config{
						SeedData:   true,
						DataPoints: 100,
					},
				},
				Query: QueryConfig{
					ParallelQueries: 5,
					Timeout:         10 * time.Second,
				},
			},
			wantErr: false,
		},
		{
			name: "invalid provider type",
			config: &Config{
				Provider: ProviderConfig{
					Type: "invalid",
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := zaptest.NewLogger(t)
			plugin := NewPlugin()

			err := plugin.(*Plugin).Initialize(context.Background(), tt.config, logger)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				// For non-memory providers, this will fail in test
				// We're just testing initialization logic
			}
		})
	}
}

func TestPlugin_Invoke_MemoryProvider(t *testing.T) {
	// Setup plugin with memory provider
	logger := zaptest.NewLogger(t)
	plugin := &Plugin{
		metricsCache: make(map[string]*CachedMetrics),
	}

	config := &Config{
		Provider: ProviderConfig{
			Type: "memory",
			Memory: &memory.Config{
				SeedData:   true,
				DataPoints: 10,
				Metrics:    []string{"cpu_usage", "memory_usage"},
				Services:   []string{"test-service"},
			},
		},
		Query: QueryConfig{
			ParallelQueries: 5,
			Timeout:         10 * time.Second,
			DefaultStep:     1 * time.Minute,
		},
		Cache: CacheConfig{
			Enabled:  true,
			Duration: 1 * time.Minute,
		},
		Metrics: []MetricConfig{
			{
				Name:        "cpu_usage",
				Aggregation: "avg",
			},
			{
				Name:        "memory_usage",
				Aggregation: "max",
			},
		},
	}

	err := plugin.Initialize(context.Background(), config, logger)
	require.NoError(t, err)

	// Test cases
	tests := []struct {
		name    string
		request HealthQueryRequest
		wantErr bool
		checks  func(t *testing.T, response *HealthQueryResponse)
	}{
		{
			name: "single metric query",
			request: HealthQueryRequest{
				ServiceName: "test-service",
				TimeRange:   "1h",
				Metrics:     []string{"cpu_usage"},
			},
			wantErr: false,
			checks: func(t *testing.T, response *HealthQueryResponse) {
				assert.Equal(t, "test-service", response.ServiceName)
				assert.Contains(t, response.Metrics, "cpu_usage")
				assert.NotEmpty(t, response.Metrics["cpu_usage"].Values)
				assert.NotNil(t, response.Metrics["cpu_usage"].Statistics)
			},
		},
		{
			name: "multiple metrics query",
			request: HealthQueryRequest{
				ServiceName: "test-service",
				TimeRange:   "24h",
				Metrics:     []string{"cpu_usage", "memory_usage"},
			},
			wantErr: false,
			checks: func(t *testing.T, response *HealthQueryResponse) {
				assert.Len(t, response.Metrics, 2)
				assert.Contains(t, response.Metrics, "cpu_usage")
				assert.Contains(t, response.Metrics, "memory_usage")
			},
		},
		{
			name: "query with filters",
			request: HealthQueryRequest{
				ServiceName: "test-service",
				TimeRange:   "1h",
				Metrics:     []string{"cpu_usage"},
				Filters: map[string]string{
					"environment": "production",
				},
			},
			wantErr: false,
			checks: func(t *testing.T, response *HealthQueryResponse) {
				assert.NotNil(t, response.Metrics["cpu_usage"])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create request
			reqBytes, err := json.Marshal(tt.request)
			require.NoError(t, err)

			req := &pb.InvokeRequest{
				Payload: reqBytes,
			}

			// Invoke plugin
			resp, err := plugin.Invoke(context.Background(), req)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, resp)

			// Parse response
			var response HealthQueryResponse
			err = json.Unmarshal(resp.Payload, &response)
			require.NoError(t, err)

			// Run checks
			if tt.checks != nil {
				tt.checks(t, &response)
			}
		})
	}
}

func TestPlugin_Caching(t *testing.T) {
	// Setup plugin with memory provider and short cache
	logger := zaptest.NewLogger(t)
	plugin := &Plugin{
		metricsCache: make(map[string]*CachedMetrics),
	}

	config := &Config{
		Provider: ProviderConfig{
			Type: "memory",
			Memory: &memory.Config{
				SeedData:   true,
				DataPoints: 5,
			},
		},
		Query: QueryConfig{
			ParallelQueries: 5,
			Timeout:         10 * time.Second,
		},
		Cache: CacheConfig{
			Enabled:  true,
			Duration: 100 * time.Millisecond, // Short cache for testing
		},
	}

	err := plugin.Initialize(context.Background(), config, logger)
	require.NoError(t, err)

	request := HealthQueryRequest{
		ServiceName: "cache-test",
		TimeRange:   "1h",
		Metrics:     []string{"cpu_usage"},
	}

	reqBytes, err := json.Marshal(request)
	require.NoError(t, err)

	req := &pb.InvokeRequest{
		Payload: reqBytes,
	}

	// First call - should cache
	resp1, err := plugin.Invoke(context.Background(), req)
	require.NoError(t, err)

	// Second call immediately - should use cache
	resp2, err := plugin.Invoke(context.Background(), req)
	require.NoError(t, err)

	// Responses should be identical (from cache)
	assert.Equal(t, resp1.Payload, resp2.Payload)

	// Wait for cache to expire
	time.Sleep(200 * time.Millisecond)

	// Third call - cache expired, should fetch new data
	resp3, err := plugin.Invoke(context.Background(), req)
	require.NoError(t, err)

	// Parse responses to check timestamps
	var response1, response3 HealthQueryResponse
	json.Unmarshal(resp1.Payload, &response1)
	json.Unmarshal(resp3.Payload, &response3)

	// Timestamps should be different
	assert.NotEqual(t, response1.Timestamp, response3.Timestamp)
}

func TestPlugin_ParallelQueries(t *testing.T) {
	logger := zaptest.NewLogger(t)
	plugin := &Plugin{
		metricsCache: make(map[string]*CachedMetrics),
	}

	config := &Config{
		Provider: ProviderConfig{
			Type: "memory",
			Memory: &memory.Config{
				SeedData:   true,
				DataPoints: 10,
				Metrics: []string{
					"metric1", "metric2", "metric3",
					"metric4", "metric5", "metric6",
				},
			},
		},
		Query: QueryConfig{
			ParallelQueries: 3, // Limit parallel queries
			Timeout:         10 * time.Second,
		},
		Cache: CacheConfig{
			Enabled: false,
		},
	}

	err := plugin.Initialize(context.Background(), config, logger)
	require.NoError(t, err)

	request := HealthQueryRequest{
		ServiceName: "parallel-test",
		TimeRange:   "1h",
		Metrics: []string{
			"metric1", "metric2", "metric3",
			"metric4", "metric5", "metric6",
		},
	}

	reqBytes, err := json.Marshal(request)
	require.NoError(t, err)

	req := &pb.InvokeRequest{
		Payload: reqBytes,
	}

	// Execute query
	start := time.Now()
	resp, err := plugin.Invoke(context.Background(), req)
	duration := time.Since(start)

	require.NoError(t, err)
	require.NotNil(t, resp)

	// Parse response
	var response HealthQueryResponse
	err = json.Unmarshal(resp.Payload, &response)
	require.NoError(t, err)

	// Should have all metrics
	assert.Len(t, response.Metrics, 6)

	// Duration should be reasonable (parallel execution)
	assert.Less(t, duration, 5*time.Second)
}

func TestPlugin_Statistics(t *testing.T) {
	logger := zaptest.NewLogger(t)
	plugin := &Plugin{
		metricsCache: make(map[string]*CachedMetrics),
	}

	config := &Config{
		Provider: ProviderConfig{
			Type: "memory",
			Memory: &memory.Config{
				SeedData:   true,
				DataPoints: 100, // Enough for percentile calculations
			},
		},
		Query: QueryConfig{
			ParallelQueries: 5,
			Timeout:         10 * time.Second,
		},
		Cache: CacheConfig{
			Enabled: false,
		},
	}

	err := plugin.Initialize(context.Background(), config, logger)
	require.NoError(t, err)

	request := HealthQueryRequest{
		ServiceName: "stats-test",
		TimeRange:   "1h",
		Metrics:     []string{"cpu_usage"},
	}

	reqBytes, err := json.Marshal(request)
	require.NoError(t, err)

	req := &pb.InvokeRequest{
		Payload: reqBytes,
	}

	resp, err := plugin.Invoke(context.Background(), req)
	require.NoError(t, err)

	var response HealthQueryResponse
	err = json.Unmarshal(resp.Payload, &response)
	require.NoError(t, err)

	// Check statistics
	stats := response.Metrics["cpu_usage"].Statistics
	require.NotNil(t, stats)

	assert.Greater(t, stats.Count, 0)
	assert.GreaterOrEqual(t, stats.Max, stats.Min)
	assert.GreaterOrEqual(t, stats.Avg, stats.Min)
	assert.LessOrEqual(t, stats.Avg, stats.Max)

	// Percentiles should be calculated for 100 data points
	assert.Greater(t, stats.P50, 0.0)
	assert.Greater(t, stats.P95, 0.0)
	assert.Greater(t, stats.P99, 0.0)
	assert.GreaterOrEqual(t, stats.P99, stats.P95)
	assert.GreaterOrEqual(t, stats.P95, stats.P50)
}

func TestPlugin_HealthCheck(t *testing.T) {
	logger := zaptest.NewLogger(t)
	plugin := &Plugin{
		metricsCache: make(map[string]*CachedMetrics),
	}

	// Before initialization
	err := plugin.HealthCheck(context.Background())
	assert.Error(t, err)

	// After initialization with memory provider
	config := &Config{
		Provider: ProviderConfig{
			Type: "memory",
		},
		Query: QueryConfig{
			Timeout: 10 * time.Second,
		},
	}

	err = plugin.Initialize(context.Background(), config, logger)
	require.NoError(t, err)

	// Should be healthy
	err = plugin.HealthCheck(context.Background())
	assert.NoError(t, err)
}

func TestPlugin_Close(t *testing.T) {
	logger := zaptest.NewLogger(t)
	plugin := &Plugin{
		metricsCache: make(map[string]*CachedMetrics),
	}

	config := &Config{
		Provider: ProviderConfig{
			Type: "memory",
		},
		Query: QueryConfig{
			Timeout: 10 * time.Second,
		},
	}

	err := plugin.Initialize(context.Background(), config, logger)
	require.NoError(t, err)

	// Add some cache entries
	plugin.metricsCache["test"] = &CachedMetrics{
		Data:      "test data",
		Timestamp: time.Now(),
	}

	// Close plugin
	err = plugin.Close(context.Background())
	assert.NoError(t, err)

	// Cache should be cleared
	assert.Empty(t, plugin.metricsCache)
}