// Package health_aggregator 實作健康指標聚合插件
package health_aggregator

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	pb "github.com/detectviz/detectviz-platform/contracts/gen/go/detectviz/contracts/v1"
	"github.com/detectviz/detectviz-platform/go-platform/internal/metrics"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/structpb"
)

// Plugin 實作健康聚合器插件，符合 Handler 介面
type Plugin struct {
	provider     metrics.MetricsProvider
	logger       *zap.Logger
	mu           sync.RWMutex
	metricsCache map[string]*CachedMetrics
}

// CachedMetrics 快取的指標數據
type CachedMetrics struct {
	Data      *HealthQueryResponse `json:"data"`
	Timestamp time.Time            `json:"timestamp"`
}

// HealthQueryRequest 健康查詢請求
type HealthQueryRequest struct {
	ServiceName string            `json:"service_name"`
	TimeRange   string            `json:"time_range"`
	Metrics     []string          `json:"metrics"`
	Filters     map[string]string `json:"filters,omitempty"`
}

// HealthQueryResponse 健康查詢回應
type HealthQueryResponse struct {
	ServiceName string                 `json:"service_name"`
	Metrics     map[string]*MetricData `json:"metrics"`
	Timestamp   time.Time              `json:"timestamp"`
	Warnings    []string               `json:"warnings,omitempty"`
}

// MetricData 指標數據
type MetricData struct {
	Name       string      `json:"name"`
	Values     []DataPoint `json:"values"`
	Statistics *Statistics `json:"statistics,omitempty"`
}

// DataPoint 數據點
type DataPoint struct {
	Timestamp time.Time `json:"timestamp"`
	Value     float64   `json:"value"`
}

// Statistics 統計資訊
type Statistics struct {
	Min   float64 `json:"min"`
	Max   float64 `json:"max"`
	Avg   float64 `json:"avg"`
	Count int64   `json:"count"`
}

// New 創建新的健康聚合器插件
func New() *Plugin {
	factory := metrics.NewSimpleFactory(zap.NewNop())
	provider := factory.CreateMemoryProvider()

	return &Plugin{
		provider:     provider,
		logger:       zap.NewNop(),
		metricsCache: make(map[string]*CachedMetrics),
	}
}

// Initialize 初始化插件（可選，用於配置）
func (p *Plugin) Initialize(logger *zap.Logger) {
	if logger != nil {
		p.logger = logger
	}
}

// Invoke 實作 Handler 介面 - 處理插件調用
func (p *Plugin) Invoke(ctx context.Context, req *pb.InvokeRequest) (*pb.InvokeResponse, error) {
	// 解析請求參數
	var queryReq HealthQueryRequest
	payloadBytes, err := req.Payload.MarshalJSON()
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload: %w", err)
	}

	if err := json.Unmarshal(payloadBytes, &queryReq); err != nil {
		return nil, fmt.Errorf("failed to parse request: %w", err)
	}

	p.logger.Debug("Processing health query",
		zap.String("service", queryReq.ServiceName),
		zap.String("time_range", queryReq.TimeRange),
		zap.Strings("metrics", queryReq.Metrics),
	)

	// 檢查快取
	if cached := p.getCachedMetrics(queryReq.ServiceName); cached != nil {
		p.logger.Debug("Returning cached metrics")
		return p.buildResponse(cached.Data)
	}

	// 執行查詢
	response, err := p.queryHealthMetrics(ctx, queryReq)
	if err != nil {
		return nil, fmt.Errorf("failed to query health metrics: %w", err)
	}

	// 快取結果
	p.cacheMetrics(queryReq.ServiceName, response)

	return p.buildResponse(response)
}

// Close 實作 ClosableHandler 介面 - 清理資源
func (p *Plugin) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	// 清理快取
	p.metricsCache = make(map[string]*CachedMetrics)

	// 關閉 provider
	if p.provider != nil {
		if err := p.provider.Close(); err != nil {
			p.logger.Warn("Failed to close metrics provider", zap.Error(err))
		}
	}

	p.logger.Info("Health aggregator plugin closed")
	return nil
}

// queryHealthMetrics 查詢健康指標
func (p *Plugin) queryHealthMetrics(ctx context.Context, req HealthQueryRequest) (*HealthQueryResponse, error) {
	// 解析時間範圍
	timeRange, err := p.parseTimeRange(req.TimeRange)
	if err != nil {
		return nil, fmt.Errorf("invalid time range: %w", err)
	}

	response := &HealthQueryResponse{
		ServiceName: req.ServiceName,
		Metrics:     make(map[string]*MetricData),
		Timestamp:   time.Now(),
	}

	// 查詢每個指標
	for _, metricName := range req.Metrics {
		query := metrics.MetricQuery{
			Metric: metricName,
			Labels: map[string]string{
				"service": req.ServiceName,
			},
			TimeRange: timeRange,
			Step:      time.Minute,
		}

		// 加入額外的篩選條件
		for k, v := range req.Filters {
			query.Labels[k] = v
		}

		result, err := p.provider.Query(ctx, query)
		if err != nil {
			p.logger.Warn("Failed to query metric",
				zap.String("metric", metricName),
				zap.Error(err),
			)
			continue
		}

		// 轉換結果
		metricData := p.convertToMetricData(metricName, result)
		response.Metrics[metricName] = metricData
	}

	return response, nil
}

// parseTimeRange 解析時間範圍
func (p *Plugin) parseTimeRange(timeRangeStr string) (metrics.TimeRange, error) {
	now := time.Now()

	switch timeRangeStr {
	case "1h":
		return metrics.TimeRange{
			Start: now.Add(-time.Hour),
			End:   now,
		}, nil
	case "6h":
		return metrics.TimeRange{
			Start: now.Add(-6 * time.Hour),
			End:   now,
		}, nil
	case "1d":
		return metrics.TimeRange{
			Start: now.Add(-24 * time.Hour),
			End:   now,
		}, nil
	default:
		return metrics.TimeRange{
			Start: now.Add(-time.Hour),
			End:   now,
		}, nil
	}
}

// convertToMetricData 轉換查詢結果為指標數據
func (p *Plugin) convertToMetricData(name string, result *metrics.QueryResult) *MetricData {
	if len(result.Series) == 0 {
		return &MetricData{
			Name:   name,
			Values: []DataPoint{},
		}
	}

	// 使用第一個序列的數據
	series := result.Series[0]
	values := make([]DataPoint, len(series.Values))

	for i, dp := range series.Values {
		values[i] = DataPoint{
			Timestamp: dp.Timestamp,
			Value:     dp.Value,
		}
	}

	// 計算統計資訊
	stats := p.calculateStatistics(values)

	return &MetricData{
		Name:       name,
		Values:     values,
		Statistics: stats,
	}
}

// calculateStatistics 計算統計資訊
func (p *Plugin) calculateStatistics(values []DataPoint) *Statistics {
	if len(values) == 0 {
		return &Statistics{}
	}

	min := values[0].Value
	max := values[0].Value
	sum := 0.0

	for _, v := range values {
		if v.Value < min {
			min = v.Value
		}
		if v.Value > max {
			max = v.Value
		}
		sum += v.Value
	}

	avg := sum / float64(len(values))

	return &Statistics{
		Min:   min,
		Max:   max,
		Avg:   avg,
		Count: int64(len(values)),
	}
}

// getCachedMetrics 獲取快取的指標
func (p *Plugin) getCachedMetrics(serviceName string) *CachedMetrics {
	p.mu.RLock()
	defer p.mu.RUnlock()

	cached, exists := p.metricsCache[serviceName]
	if !exists {
		return nil
	}

	// 檢查快取是否過期（5分鐘）
	if time.Since(cached.Timestamp) > 5*time.Minute {
		return nil
	}

	return cached
}

// cacheMetrics 快取指標數據
func (p *Plugin) cacheMetrics(serviceName string, response *HealthQueryResponse) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.metricsCache[serviceName] = &CachedMetrics{
		Data:      response,
		Timestamp: time.Now(),
	}
}

// buildResponse 建構回應
func (p *Plugin) buildResponse(response *HealthQueryResponse) (*pb.InvokeResponse, error) {
	// 序列化回應
	responseBytes, err := json.Marshal(response)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal response: %w", err)
	}

	// 轉換為 structpb.Struct
	result := &structpb.Struct{}
	if err := result.UnmarshalJSON(responseBytes); err != nil {
		return nil, fmt.Errorf("failed to convert response to struct: %w", err)
	}

	return &pb.InvokeResponse{
		Result: result,
	}, nil
}

// HealthCheck 實作 HealthAwareHandler 介面 - 健康檢查
func (p *Plugin) HealthCheck() error {
	// 檢查 provider 是否可用
	if p.provider == nil {
		return fmt.Errorf("metrics provider 未初始化")
	}

	// 檢查快取大小是否過大
	p.mu.RLock()
	cacheSize := len(p.metricsCache)
	p.mu.RUnlock()

	if cacheSize > 1000 {
		return fmt.Errorf("快取過大: %d entries", cacheSize)
	}

	return nil
}
