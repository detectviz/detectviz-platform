package metrics

import (
	"context"
	"math/rand"
	"sync"
	"time"

	"go.uber.org/zap"
)

// SimpleFactory 提供簡化的 metrics provider 工廠
type SimpleFactory struct {
	logger *zap.Logger
}

// NewSimpleFactory 創建簡化的工廠
func NewSimpleFactory(logger *zap.Logger) *SimpleFactory {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &SimpleFactory{
		logger: logger,
	}
}

// CreateMemoryProvider 創建記憶體測試 provider
func (f *SimpleFactory) CreateMemoryProvider() MetricsProvider {
	return &MemoryProvider{
		data:   make(map[string][]DataPoint),
		mu:     sync.RWMutex{},
		logger: f.logger,
	}
}

// MemoryProvider 簡化的記憶體 provider
type MemoryProvider struct {
	data   map[string][]DataPoint
	mu     sync.RWMutex
	logger *zap.Logger
}

// Query 實作 MetricsProvider.Query
func (p *MemoryProvider) Query(ctx context.Context, query MetricQuery) (*QueryResult, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	// 生成測試數據
	dataPoints := p.generateTestData(query.Metric, query.TimeRange)

	series := []TimeSeries{
		{
			Labels: query.Labels,
			Values: dataPoints,
		},
	}

	return &QueryResult{
		Query:  &query,
		Series: series,
	}, nil
}

// BatchQuery 實作 MetricsProvider.BatchQuery
func (p *MemoryProvider) BatchQuery(ctx context.Context, queries []MetricQuery) ([]*QueryResult, error) {
	results := make([]*QueryResult, len(queries))
	for i, query := range queries {
		result, err := p.Query(ctx, query)
		if err != nil {
			return nil, err
		}
		results[i] = result
	}
	return results, nil
}

// GetAggregation 實作 MetricsProvider.GetAggregation
func (p *MemoryProvider) GetAggregation(ctx context.Context, opts AggregationOptions) (*AggregationResult, error) {
	// 簡化實作
	return &AggregationResult{
		Options: &opts,
		Values: []AggregatedValue{
			{
				Metric: opts.Metrics[0],
				Value:  rand.Float64() * 100,
			},
		},
	}, nil
}

// HealthCheck 實作 MetricsProvider.HealthCheck
func (p *MemoryProvider) HealthCheck(ctx context.Context) error {
	return nil
}

// Close 實作 MetricsProvider.Close
func (p *MemoryProvider) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.data = make(map[string][]DataPoint)
	return nil
}

// generateTestData 生成測試數據
func (p *MemoryProvider) generateTestData(metric string, timeRange TimeRange) []DataPoint {
	var points []DataPoint
	current := timeRange.Start
	step := time.Minute

	for current.Before(timeRange.End) || current.Equal(timeRange.End) {
		value := p.generateTestValue(metric)
		points = append(points, DataPoint{
			Timestamp: current,
			Value:     value,
		})
		current = current.Add(step)
	}

	return points
}

// generateTestValue 生成測試值
func (p *MemoryProvider) generateTestValue(metric string) float64 {
	base := 0.0
	amplitude := 1.0

	switch metric {
	case "cpu_usage":
		base = 0.3
		amplitude = 0.4
	case "memory_usage":
		base = 0.6
		amplitude = 0.3
	case "http_request_duration_seconds":
		base = 0.1
		amplitude = 0.2
	case "up":
		return 1.0
	default:
		base = 50
		amplitude = 25
	}

	// 添加隨機性
	noise := (rand.Float64() - 0.5) * 0.1
	return base + amplitude*rand.Float64() + noise
}
