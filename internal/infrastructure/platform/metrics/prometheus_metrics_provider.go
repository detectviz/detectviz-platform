package metrics

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"detectviz-platform/pkg/platform/contracts"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// PrometheusMetricsProvider 實現 MetricsProvider 介面，提供 Prometheus 指標收集功能
// 職責: 收集應用程式運行時指標並通過 HTTP endpoint 導出給 Prometheus
type PrometheusMetricsProvider struct {
	registry   *prometheus.Registry
	counters   map[string]*prometheus.CounterVec
	histograms map[string]*prometheus.HistogramVec
	gauges     map[string]*prometheus.GaugeVec
	httpServer *http.Server
}

// PrometheusConfig 定義 Prometheus 指標提供者的配置
type PrometheusConfig struct {
	Port    int    `yaml:"port" json:"port"`
	Path    string `yaml:"path" json:"path"`
	Enabled bool   `yaml:"enabled" json:"enabled"`
}

// NewPrometheusMetricsProvider 創建新的 Prometheus 指標提供者實例
func NewPrometheusMetricsProvider(config PrometheusConfig) contracts.MetricsProvider {
	registry := prometheus.NewRegistry()

	provider := &PrometheusMetricsProvider{
		registry:   registry,
		counters:   make(map[string]*prometheus.CounterVec),
		histograms: make(map[string]*prometheus.HistogramVec),
		gauges:     make(map[string]*prometheus.GaugeVec),
	}

	if config.Enabled {
		provider.startHTTPServer(config.Port, config.Path)
	}

	return provider
}

// IncCounter 增加計數器指標
func (p *PrometheusMetricsProvider) IncCounter(name string, tags map[string]string) {
	counter := p.getOrCreateCounter(name, getTagKeys(tags))
	counter.With(prometheus.Labels(tags)).Inc()
}

// ObserveHistogram 記錄直方圖指標
func (p *PrometheusMetricsProvider) ObserveHistogram(name string, value float64, tags map[string]string) {
	histogram := p.getOrCreateHistogram(name, getTagKeys(tags))
	histogram.With(prometheus.Labels(tags)).Observe(value)
}

// SetGauge 設置儀表盤指標
func (p *PrometheusMetricsProvider) SetGauge(name string, value float64, tags map[string]string) {
	gauge := p.getOrCreateGauge(name, getTagKeys(tags))
	gauge.With(prometheus.Labels(tags)).Set(value)
}

// GetName 返回提供者名稱
func (p *PrometheusMetricsProvider) GetName() string {
	return "prometheus_metrics_provider"
}

// getOrCreateCounter 獲取或創建計數器指標
func (p *PrometheusMetricsProvider) getOrCreateCounter(name string, labelNames []string) *prometheus.CounterVec {
	if counter, exists := p.counters[name]; exists {
		return counter
	}

	counter := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: name,
		Help: fmt.Sprintf("Counter metric for %s", name),
	}, labelNames)

	p.registry.MustRegister(counter)
	p.counters[name] = counter
	return counter
}

// getOrCreateHistogram 獲取或創建直方圖指標
func (p *PrometheusMetricsProvider) getOrCreateHistogram(name string, labelNames []string) *prometheus.HistogramVec {
	if histogram, exists := p.histograms[name]; exists {
		return histogram
	}

	histogram := prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    name,
		Help:    fmt.Sprintf("Histogram metric for %s", name),
		Buckets: prometheus.DefBuckets,
	}, labelNames)

	p.registry.MustRegister(histogram)
	p.histograms[name] = histogram
	return histogram
}

// getOrCreateGauge 獲取或創建儀表盤指標
func (p *PrometheusMetricsProvider) getOrCreateGauge(name string, labelNames []string) *prometheus.GaugeVec {
	if gauge, exists := p.gauges[name]; exists {
		return gauge
	}

	gauge := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: name,
		Help: fmt.Sprintf("Gauge metric for %s", name),
	}, labelNames)

	p.registry.MustRegister(gauge)
	p.gauges[name] = gauge
	return gauge
}

// startHTTPServer 啟動 HTTP 服務器以導出指標
func (p *PrometheusMetricsProvider) startHTTPServer(port int, path string) {
	mux := http.NewServeMux()
	mux.Handle(path, promhttp.HandlerFor(p.registry, promhttp.HandlerOpts{
		EnableOpenMetrics: true,
	}))

	p.httpServer = &http.Server{
		Addr:         fmt.Sprintf(":%d", port),
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  15 * time.Second,
	}

	go func() {
		if err := p.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			// 這裡應該使用適當的日誌記錄
			fmt.Printf("Metrics server error: %v\n", err)
		}
	}()
}

// Shutdown 優雅關閉指標提供者
func (p *PrometheusMetricsProvider) Shutdown(ctx context.Context) error {
	if p.httpServer != nil {
		return p.httpServer.Shutdown(ctx)
	}
	return nil
}

// getTagKeys 從 tags map 中提取所有鍵
func getTagKeys(tags map[string]string) []string {
	keys := make([]string, 0, len(tags))
	for key := range tags {
		keys = append(keys, key)
	}
	return keys
}
