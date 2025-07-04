package monitoring

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// MetricsCollector 收集平台指標
// AI_PLUGIN_TYPE: "metrics_collector"
// AI_IMPL_PACKAGE: "detectviz-platform/internal/infrastructure/platform/monitoring"
// AI_IMPL_CONSTRUCTOR: "NewMetricsCollector"
type MetricsCollector struct {
	// HTTP 請求指標
	httpRequestsTotal    *prometheus.CounterVec
	httpRequestDuration  *prometheus.HistogramVec
	httpRequestsInFlight *prometheus.GaugeVec
	httpResponseSize     *prometheus.HistogramVec

	// 插件指標
	pluginRequestsTotal   *prometheus.CounterVec
	pluginRequestDuration *prometheus.HistogramVec
	pluginHealthStatus    *prometheus.GaugeVec
	pluginErrors          *prometheus.CounterVec

	// 系統指標
	systemMemoryUsage *prometheus.GaugeVec
	systemCPUUsage    *prometheus.GaugeVec
	systemGoroutines  prometheus.Gauge
	systemOpenFiles   prometheus.Gauge

	// 業務指標
	detectionsTotal   *prometheus.CounterVec
	detectionsLatency *prometheus.HistogramVec
	dataImportsTotal  *prometheus.CounterVec
	dataImportSize    *prometheus.HistogramVec

	// 錯誤指標
	errorsTotal *prometheus.CounterVec
	panicTotal  *prometheus.CounterVec

	// 配置指標
	configReloads      *prometheus.CounterVec
	configLoadDuration *prometheus.HistogramVec
}

// NewMetricsCollector 創建新的指標收集器
func NewMetricsCollector() *MetricsCollector {
	return &MetricsCollector{
		// HTTP 請求指標
		httpRequestsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "detectviz_http_requests_total",
				Help: "Total number of HTTP requests",
			},
			[]string{"method", "endpoint", "status_code"},
		),
		httpRequestDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "detectviz_http_request_duration_seconds",
				Help:    "HTTP request duration in seconds",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"method", "endpoint"},
		),
		httpRequestsInFlight: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "detectviz_http_requests_in_flight",
				Help: "Number of HTTP requests currently being processed",
			},
			[]string{"method", "endpoint"},
		),
		httpResponseSize: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "detectviz_http_response_size_bytes",
				Help:    "HTTP response size in bytes",
				Buckets: []float64{100, 1000, 10000, 100000, 1000000},
			},
			[]string{"method", "endpoint"},
		),

		// 插件指標
		pluginRequestsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "detectviz_plugin_requests_total",
				Help: "Total number of plugin requests",
			},
			[]string{"plugin_name", "plugin_type", "status"},
		),
		pluginRequestDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "detectviz_plugin_request_duration_seconds",
				Help:    "Plugin request duration in seconds",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"plugin_name", "plugin_type"},
		),
		pluginHealthStatus: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "detectviz_plugin_health_status",
				Help: "Plugin health status (1=healthy, 0=unhealthy)",
			},
			[]string{"plugin_name", "plugin_type"},
		),
		pluginErrors: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "detectviz_plugin_errors_total",
				Help: "Total number of plugin errors",
			},
			[]string{"plugin_name", "plugin_type", "error_type"},
		),

		// 系統指標
		systemMemoryUsage: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "detectviz_system_memory_usage_bytes",
				Help: "System memory usage in bytes",
			},
			[]string{"type"},
		),
		systemCPUUsage: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "detectviz_system_cpu_usage_percent",
				Help: "System CPU usage percentage",
			},
			[]string{"type"},
		),
		systemGoroutines: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: "detectviz_system_goroutines",
				Help: "Number of goroutines",
			},
		),
		systemOpenFiles: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: "detectviz_system_open_files",
				Help: "Number of open file descriptors",
			},
		),

		// 業務指標
		detectionsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "detectviz_detections_total",
				Help: "Total number of detections performed",
			},
			[]string{"detector_type", "result_type"},
		),
		detectionsLatency: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "detectviz_detection_latency_seconds",
				Help:    "Detection latency in seconds",
				Buckets: []float64{0.001, 0.01, 0.1, 1, 10, 60},
			},
			[]string{"detector_type"},
		),
		dataImportsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "detectviz_data_imports_total",
				Help: "Total number of data imports",
			},
			[]string{"importer_type", "status"},
		),
		dataImportSize: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "detectviz_data_import_size_bytes",
				Help:    "Data import size in bytes",
				Buckets: []float64{1024, 10240, 102400, 1048576, 10485760},
			},
			[]string{"importer_type"},
		),

		// 錯誤指標
		errorsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "detectviz_errors_total",
				Help: "Total number of errors",
			},
			[]string{"component", "error_type"},
		),
		panicTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "detectviz_panics_total",
				Help: "Total number of panics",
			},
			[]string{"component"},
		),

		// 配置指標
		configReloads: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "detectviz_config_reloads_total",
				Help: "Total number of configuration reloads",
			},
			[]string{"config_type", "status"},
		),
		configLoadDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "detectviz_config_load_duration_seconds",
				Help:    "Configuration load duration in seconds",
				Buckets: []float64{0.001, 0.01, 0.1, 1, 5},
			},
			[]string{"config_type"},
		),
	}
}

// RecordHTTPRequest 記錄 HTTP 請求指標
func (m *MetricsCollector) RecordHTTPRequest(method, endpoint, statusCode string, duration time.Duration, responseSize int64) {
	m.httpRequestsTotal.WithLabelValues(method, endpoint, statusCode).Inc()
	m.httpRequestDuration.WithLabelValues(method, endpoint).Observe(duration.Seconds())
	m.httpResponseSize.WithLabelValues(method, endpoint).Observe(float64(responseSize))
}

// RecordHTTPRequestInFlight 記錄進行中的 HTTP 請求
func (m *MetricsCollector) RecordHTTPRequestInFlight(method, endpoint string, delta float64) {
	m.httpRequestsInFlight.WithLabelValues(method, endpoint).Add(delta)
}

// RecordPluginRequest 記錄插件請求指標
func (m *MetricsCollector) RecordPluginRequest(pluginName, pluginType, status string, duration time.Duration) {
	m.pluginRequestsTotal.WithLabelValues(pluginName, pluginType, status).Inc()
	m.pluginRequestDuration.WithLabelValues(pluginName, pluginType).Observe(duration.Seconds())
}

// RecordPluginHealth 記錄插件健康狀態
func (m *MetricsCollector) RecordPluginHealth(pluginName, pluginType string, healthy bool) {
	value := 0.0
	if healthy {
		value = 1.0
	}
	m.pluginHealthStatus.WithLabelValues(pluginName, pluginType).Set(value)
}

// RecordPluginError 記錄插件錯誤
func (m *MetricsCollector) RecordPluginError(pluginName, pluginType, errorType string) {
	m.pluginErrors.WithLabelValues(pluginName, pluginType, errorType).Inc()
}

// RecordSystemMemory 記錄系統記憶體使用
func (m *MetricsCollector) RecordSystemMemory(memType string, bytes float64) {
	m.systemMemoryUsage.WithLabelValues(memType).Set(bytes)
}

// RecordSystemCPU 記錄系統 CPU 使用
func (m *MetricsCollector) RecordSystemCPU(cpuType string, percent float64) {
	m.systemCPUUsage.WithLabelValues(cpuType).Set(percent)
}

// RecordSystemGoroutines 記錄 Goroutine 數量
func (m *MetricsCollector) RecordSystemGoroutines(count float64) {
	m.systemGoroutines.Set(count)
}

// RecordSystemOpenFiles 記錄打開的文件描述符數量
func (m *MetricsCollector) RecordSystemOpenFiles(count float64) {
	m.systemOpenFiles.Set(count)
}

// RecordDetection 記錄檢測指標
func (m *MetricsCollector) RecordDetection(detectorType, resultType string, latency time.Duration) {
	m.detectionsTotal.WithLabelValues(detectorType, resultType).Inc()
	m.detectionsLatency.WithLabelValues(detectorType).Observe(latency.Seconds())
}

// RecordDataImport 記錄數據導入指標
func (m *MetricsCollector) RecordDataImport(importerType, status string, size int64) {
	m.dataImportsTotal.WithLabelValues(importerType, status).Inc()
	m.dataImportSize.WithLabelValues(importerType).Observe(float64(size))
}

// RecordError 記錄錯誤指標
func (m *MetricsCollector) RecordError(component, errorType string) {
	m.errorsTotal.WithLabelValues(component, errorType).Inc()
}

// RecordPanic 記錄 panic 指標
func (m *MetricsCollector) RecordPanic(component string) {
	m.panicTotal.WithLabelValues(component).Inc()
}

// RecordConfigReload 記錄配置重載指標
func (m *MetricsCollector) RecordConfigReload(configType, status string, duration time.Duration) {
	m.configReloads.WithLabelValues(configType, status).Inc()
	m.configLoadDuration.WithLabelValues(configType).Observe(duration.Seconds())
}

// GetName 返回指標收集器名稱
func (m *MetricsCollector) GetName() string {
	return "metrics_collector"
}
