package health_aggregator

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	pb "github.com/detectviz/detectviz-platform/contracts/gen/go/detectviz/contracts/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
	"google.golang.org/protobuf/types/known/structpb"
)

func TestPlugin_New(t *testing.T) {
	plugin := New()

	assert.NotNil(t, plugin)
	assert.NotNil(t, plugin.provider)
	assert.NotNil(t, plugin.logger)
	assert.NotNil(t, plugin.metricsCache)
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

func TestPlugin_Cache(t *testing.T) {
	plugin := New()
	logger := zaptest.NewLogger(t)
	plugin.Initialize(logger)

	serviceName := "test-service"

	// 第一次查詢
	request := HealthQueryRequest{
		ServiceName: serviceName,
		TimeRange:   "1h",
		Metrics:     []string{"cpu_usage"},
	}

	reqBytes, err := json.Marshal(request)
	require.NoError(t, err)

	reqStruct := &structpb.Struct{}
	err = reqStruct.UnmarshalJSON(reqBytes)
	require.NoError(t, err)

	req := &pb.InvokeRequest{
		Payload: reqStruct,
	}

	// 執行第一次查詢
	resp1, err := plugin.Invoke(context.Background(), req)
	require.NoError(t, err)

	// 檢查快取是否存在
	cached := plugin.getCachedMetrics(serviceName)
	assert.NotNil(t, cached)

	// 執行第二次查詢（應該使用快取）
	resp2, err := plugin.Invoke(context.Background(), req)
	require.NoError(t, err)

	// 兩次回應應該相同（因為使用了快取）
	assert.Equal(t, resp1.Result, resp2.Result)
}

func TestPlugin_Close(t *testing.T) {
	plugin := New()
	logger := zaptest.NewLogger(t)
	plugin.Initialize(logger)

	// 添加一些快取數據
	response := &HealthQueryResponse{
		ServiceName: "test",
		Metrics:     make(map[string]*MetricData),
		Timestamp:   time.Now(),
	}
	plugin.cacheMetrics("test", response)

	// 驗證快取存在
	assert.NotEmpty(t, plugin.metricsCache)

	// 關閉插件
	err := plugin.Close()
	assert.NoError(t, err)

	// 驗證快取已清理
	assert.Empty(t, plugin.metricsCache)
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
