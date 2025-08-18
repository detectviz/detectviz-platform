package health_aggregator

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"testing"
	"time"

	pb "github.com/detectviz/detectviz-platform/contracts/gen/go/detectviz/contracts/v1"
	"github.com/detectviz/detectviz-platform/go-platform/internal/metrics"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
	"google.golang.org/protobuf/types/known/structpb"
)

// mockMetricsProvider is a mock implementation of MetricsProvider for testing.
type mockMetricsProvider struct {
	metrics.MetricsProvider
	queryFunc func(ctx context.Context, query metrics.MetricQuery) (*metrics.QueryResult, error)
}

func (m *mockMetricsProvider) Query(ctx context.Context, query metrics.MetricQuery) (*metrics.QueryResult, error) {
	if m.queryFunc != nil {
		return m.queryFunc(ctx, query)
	}
	return nil, errors.New("mock not implemented")
}

func (m *mockMetricsProvider) BatchQuery(ctx context.Context, queries []metrics.MetricQuery) ([]*metrics.QueryResult, error) {
	results := make([]*metrics.QueryResult, len(queries))
	var errs []error
	for i, q := range queries {
		res, err := m.Query(ctx, q)
		if err != nil {
			// In a real scenario, we might want to handle multiple errors.
			// For this test, the first error is enough to signal failure.
			errs = append(errs, err)
		}
		results[i] = res
	}
	if len(errs) > 0 {
		return results, errs[0]
	}
	return results, nil
}

func (m *mockMetricsProvider) Close() error {
	return nil
}

func TestPlugin_New(t *testing.T) {
	plugin := New()

	assert.NotNil(t, plugin)
	assert.NotNil(t, plugin.provider)
	assert.NotNil(t, plugin.logger)
}

func TestPlugin_Initialize(t *testing.T) {
	plugin := New()
	logger := zaptest.NewLogger(t)

	plugin.Initialize(logger)

	assert.Equal(t, logger, plugin.logger)
}

func TestPlugin_Invoke(t *testing.T) {
	plugin := New()
	logger := zaptest.NewLogger(t)
	plugin.Initialize(logger)

	tests := []struct {
		name    string
		request HealthQueryRequest
		wantErr bool
	}{
		{
			name: "valid request",
			request: HealthQueryRequest{
				ServiceName: "test-service",
				TimeRange:   "1h",
				Metrics:     []string{"cpu_usage", "memory_usage"},
				Filters:     map[string]string{"env": "test"},
			},
			wantErr: false,
		},
		{
			name: "request with 6h time range",
			request: HealthQueryRequest{
				ServiceName: "another-service",
				TimeRange:   "6h",
				Metrics:     []string{"http_request_duration_seconds"},
			},
			wantErr: false,
		},
		{
			name: "request with 1d time range",
			request: HealthQueryRequest{
				ServiceName: "daily-service",
				TimeRange:   "1d",
				Metrics:     []string{"up"},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 創建請求
			reqBytes, err := json.Marshal(tt.request)
			require.NoError(t, err)

			reqStruct := &structpb.Struct{}
			err = reqStruct.UnmarshalJSON(reqBytes)
			require.NoError(t, err)

			req := &pb.InvokeRequest{
				Payload: reqStruct,
			}

			// 執行調用
			resp, err := plugin.Invoke(context.Background(), req)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.NotNil(t, resp)
			assert.NotNil(t, resp.Result)

			// 解析回應
			respBytes, err := resp.Result.MarshalJSON()
			require.NoError(t, err)

			var healthResp HealthQueryResponse
			err = json.Unmarshal(respBytes, &healthResp)
			require.NoError(t, err)

			// 驗證回應
			assert.Equal(t, tt.request.ServiceName, healthResp.ServiceName)
			assert.NotEmpty(t, healthResp.Metrics)
			assert.NotZero(t, healthResp.Timestamp)

			// 驗證請求的每個指標都有回應
			for _, metric := range tt.request.Metrics {
				metricData, exists := healthResp.Metrics[metric]
				assert.True(t, exists, "Metric %s should exist in response", metric)
				assert.Equal(t, metric, metricData.Name)
				assert.NotNil(t, metricData.Values)
			}
		})
	}
}

func TestPlugin_Close(t *testing.T) {
	plugin := New()
	logger := zaptest.NewLogger(t)
	plugin.Initialize(logger)

	// 關閉插件
	err := plugin.Close()
	assert.NoError(t, err)
}

func TestPlugin_TimeRangeParsing(t *testing.T) {
	plugin := New()

	tests := []struct {
		name        string
		timeRange   string
		expectedDur time.Duration
	}{
		{
			name:        "1 hour",
			timeRange:   "1h",
			expectedDur: time.Hour,
		},
		{
			name:        "6 hours",
			timeRange:   "6h",
			expectedDur: 6 * time.Hour,
		},
		{
			name:        "1 day",
			timeRange:   "1d",
			expectedDur: 24 * time.Hour,
		},
		{
			name:        "unknown defaults to 1h",
			timeRange:   "unknown",
			expectedDur: time.Hour,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			timeRange, err := plugin.parseTimeRange(tt.timeRange)
			require.NoError(t, err)

			actualDur := timeRange.End.Sub(timeRange.Start)
			assert.Equal(t, tt.expectedDur, actualDur)
		})
	}
}

func TestPlugin_StatisticsCalculation(t *testing.T) {
	plugin := New()

	values := []DataPoint{
		{Timestamp: time.Now(), Value: 10.0},
		{Timestamp: time.Now(), Value: 20.0},
		{Timestamp: time.Now(), Value: 30.0},
		{Timestamp: time.Now(), Value: 40.0},
		{Timestamp: time.Now(), Value: 50.0},
	}

	stats := plugin.calculateStatistics(values)

	assert.NotNil(t, stats)
	assert.Equal(t, 10.0, stats.Min)
	assert.Equal(t, 50.0, stats.Max)
	assert.Equal(t, 30.0, stats.Avg)
	assert.Equal(t, int64(5), stats.Count)
}

func TestPlugin_StatisticsEmptyValues(t *testing.T) {
	plugin := New()

	stats := plugin.calculateStatistics([]DataPoint{})

	assert.NotNil(t, stats)
	assert.Equal(t, int64(0), stats.Count)
}

func TestHealthAggregator_ProviderFailure(t *testing.T) {
	logger := zaptest.NewLogger(t)

	// 1. Create a mock provider that fails for one metric
	mockProvider := &mockMetricsProvider{
		queryFunc: func(ctx context.Context, query metrics.MetricQuery) (*metrics.QueryResult, error) {
			if query.Metric == "cpu_usage" {
				// Succeed for cpu_usage
				return &metrics.QueryResult{
					Query: &query,
					Series: []metrics.TimeSeries{
						{
							Values: []metrics.DataPoint{{Value: 0.5}},
						},
					},
				}, nil
			}
			if query.Metric == "memory_usage" {
				// Fail for memory_usage
				return nil, errors.New("provider failed for memory_usage")
			}
			return nil, fmt.Errorf("unexpected metric: %s", query.Metric)
		},
	}

	// 2. Create the plugin and inject the mock provider
	plugin := New()
	plugin.provider = mockProvider
	plugin.logger = logger

	// 3. Call Invoke with a request for both metrics
	request := HealthQueryRequest{
		ServiceName: "test-service",
		Metrics:     []string{"cpu_usage", "memory_usage"},
	}
	reqBytes, err := json.Marshal(request)
	require.NoError(t, err)
	reqStruct := &structpb.Struct{}
	err = reqStruct.UnmarshalJSON(reqBytes)
	require.NoError(t, err)
	req := &pb.InvokeRequest{Payload: reqStruct}

	resp, err := plugin.Invoke(context.Background(), req)

	// 4. Assert that the call succeeds overall
	require.NoError(t, err)
	assert.NotNil(t, resp.Result)

	// 5. Assert the response contains partial data
	var healthResp HealthQueryResponse
	respBytes, err := resp.Result.MarshalJSON()
	require.NoError(t, err)
	err = json.Unmarshal(respBytes, &healthResp)
	require.NoError(t, err)

	// Should contain the successful metric
	assert.Contains(t, healthResp.Metrics, "cpu_usage")
	assert.NotNil(t, healthResp.Metrics["cpu_usage"])
	assert.Len(t, healthResp.Metrics["cpu_usage"].Values, 1)

	// Should NOT contain the failed metric
	assert.NotContains(t, healthResp.Metrics, "memory_usage")
}
