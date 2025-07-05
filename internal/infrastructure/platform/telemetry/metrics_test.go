package telemetry

import (
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

// 為測試創建 MetricsCollector，使用獨立的 registry
func createTestMetricsCollector() *MetricsCollector {
	return &MetricsCollector{
		// HTTP 請求指標
		httpRequestsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "test_detectviz_http_requests_total",
				Help: "Total number of HTTP requests",
			},
			[]string{"method", "endpoint", "status_code"},
		),
		httpRequestDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "test_detectviz_http_request_duration_seconds",
				Help:    "HTTP request duration in seconds",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"method", "endpoint"},
		),
		httpRequestsInFlight: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "test_detectviz_http_requests_in_flight",
				Help: "Number of HTTP requests currently being processed",
			},
			[]string{"method", "endpoint"},
		),
		httpResponseSize: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "test_detectviz_http_response_size_bytes",
				Help:    "HTTP response size in bytes",
				Buckets: []float64{100, 1000, 10000, 100000, 1000000},
			},
			[]string{"method", "endpoint"},
		),

		// 插件指標
		pluginRequestsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "test_detectviz_plugin_requests_total",
				Help: "Total number of plugin requests",
			},
			[]string{"plugin_name", "plugin_type", "status"},
		),
		pluginRequestDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "test_detectviz_plugin_request_duration_seconds",
				Help:    "Plugin request duration in seconds",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"plugin_name", "plugin_type"},
		),
		pluginHealthStatus: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "test_detectviz_plugin_health_status",
				Help: "Plugin health status (1=healthy, 0=unhealthy)",
			},
			[]string{"plugin_name", "plugin_type"},
		),
		pluginErrors: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "test_detectviz_plugin_errors_total",
				Help: "Total number of plugin errors",
			},
			[]string{"plugin_name", "plugin_type", "error_type"},
		),

		// 系統指標
		systemMemoryUsage: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "test_detectviz_system_memory_usage_bytes",
				Help: "System memory usage in bytes",
			},
			[]string{"type"},
		),
		systemCPUUsage: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "test_detectviz_system_cpu_usage_percent",
				Help: "System CPU usage percentage",
			},
			[]string{"type"},
		),
		systemGoroutines: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name: "test_detectviz_system_goroutines",
				Help: "Number of goroutines",
			},
		),
		systemOpenFiles: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name: "test_detectviz_system_open_files",
				Help: "Number of open file descriptors",
			},
		),

		// 業務指標
		detectionsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "test_detectviz_detections_total",
				Help: "Total number of detections performed",
			},
			[]string{"detector_type", "result_type"},
		),
		detectionsLatency: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "test_detectviz_detection_latency_seconds",
				Help:    "Detection latency in seconds",
				Buckets: []float64{0.001, 0.01, 0.1, 1, 10, 60},
			},
			[]string{"detector_type"},
		),
		dataImportsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "test_detectviz_data_imports_total",
				Help: "Total number of data imports",
			},
			[]string{"importer_type", "status"},
		),
		dataImportSize: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "test_detectviz_data_import_size_bytes",
				Help:    "Data import size in bytes",
				Buckets: []float64{1024, 10240, 102400, 1048576, 10485760},
			},
			[]string{"importer_type"},
		),

		// 錯誤指標
		errorsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "test_detectviz_errors_total",
				Help: "Total number of errors",
			},
			[]string{"component", "error_type"},
		),
		panicTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "test_detectviz_panics_total",
				Help: "Total number of panics",
			},
			[]string{"component"},
		),

		// 配置指標
		configReloads: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "test_detectviz_config_reloads_total",
				Help: "Total number of configuration reloads",
			},
			[]string{"config_type", "status"},
		),
		configLoadDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "test_detectviz_config_load_duration_seconds",
				Help:    "Configuration load duration in seconds",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"config_type"},
		),
	}
}

func TestNewMetricsCollector(t *testing.T) {
	collector := NewMetricsCollector()

	if collector == nil {
		t.Error("NewMetricsCollector should return a non-nil collector")
	}

	if collector.GetName() != "metrics_collector" {
		t.Errorf("Expected name 'metrics_collector', got '%s'", collector.GetName())
	}
}

func TestMetricsCollector_RecordHTTPRequest(t *testing.T) {
	collector := createTestMetricsCollector()

	// 記錄 HTTP 請求
	collector.RecordHTTPRequest("GET", "/api/v1/users", "200", 100*time.Millisecond, 1024)

	// 測試不會 panic，實際的指標值需要通過 Prometheus 客戶端驗證
	// 這裡主要測試方法調用不會出錯
}

func TestMetricsCollector_RecordHTTPRequestInFlight(t *testing.T) {
	collector := createTestMetricsCollector()

	// 增加進行中的請求
	collector.RecordHTTPRequestInFlight("GET", "/api/v1/users", 1.0)

	// 減少進行中的請求
	collector.RecordHTTPRequestInFlight("GET", "/api/v1/users", -1.0)

	// 測試不會 panic
}

func TestMetricsCollector_RecordPluginRequest(t *testing.T) {
	collector := createTestMetricsCollector()

	// 記錄插件請求
	collector.RecordPluginRequest("csv_importer", "importer", "success", 50*time.Millisecond)
	collector.RecordPluginRequest("threshold_detector", "detector", "error", 200*time.Millisecond)

	// 測試不會 panic
}

func TestMetricsCollector_RecordPluginHealth(t *testing.T) {
	collector := createTestMetricsCollector()

	// 記錄插件健康狀態
	collector.RecordPluginHealth("csv_importer", "importer", true)
	collector.RecordPluginHealth("threshold_detector", "detector", false)

	// 測試不會 panic
}

func TestMetricsCollector_RecordPluginError(t *testing.T) {
	collector := createTestMetricsCollector()

	// 記錄插件錯誤
	collector.RecordPluginError("csv_importer", "importer", "validation_error")
	collector.RecordPluginError("threshold_detector", "detector", "runtime_error")

	// 測試不會 panic
}

func TestMetricsCollector_RecordSystemMemory(t *testing.T) {
	collector := createTestMetricsCollector()

	// 記錄系統記憶體
	collector.RecordSystemMemory("heap_alloc", 1024*1024)
	collector.RecordSystemMemory("heap_sys", 2048*1024)

	// 測試不會 panic
}

func TestMetricsCollector_RecordSystemCPU(t *testing.T) {
	collector := createTestMetricsCollector()

	// 記錄系統 CPU
	collector.RecordSystemCPU("user", 45.5)
	collector.RecordSystemCPU("system", 15.2)

	// 測試不會 panic
}

func TestMetricsCollector_RecordSystemGoroutines(t *testing.T) {
	collector := createTestMetricsCollector()

	// 記錄 Goroutine 數量
	collector.RecordSystemGoroutines(100)

	// 測試不會 panic
}

func TestMetricsCollector_RecordSystemOpenFiles(t *testing.T) {
	collector := createTestMetricsCollector()

	// 記錄打開的文件數量
	collector.RecordSystemOpenFiles(50)

	// 測試不會 panic
}

func TestMetricsCollector_RecordDetection(t *testing.T) {
	collector := createTestMetricsCollector()

	// 記錄檢測
	collector.RecordDetection("threshold_detector", "anomaly", 300*time.Millisecond)
	collector.RecordDetection("pattern_detector", "normal", 150*time.Millisecond)

	// 測試不會 panic
}

func TestMetricsCollector_RecordDataImport(t *testing.T) {
	collector := createTestMetricsCollector()

	// 記錄數據導入
	collector.RecordDataImport("csv_importer", "success", 1024*1024)
	collector.RecordDataImport("json_importer", "error", 512*1024)

	// 測試不會 panic
}

func TestMetricsCollector_RecordError(t *testing.T) {
	collector := createTestMetricsCollector()

	// 記錄錯誤
	collector.RecordError("http_handler", "validation_error")
	collector.RecordError("plugin_manager", "load_error")

	// 測試不會 panic
}

func TestMetricsCollector_RecordPanic(t *testing.T) {
	collector := createTestMetricsCollector()

	// 記錄 panic
	collector.RecordPanic("http_handler")
	collector.RecordPanic("plugin_manager")

	// 測試不會 panic
}

func TestMetricsCollector_RecordConfigReload(t *testing.T) {
	collector := createTestMetricsCollector()

	// 記錄配置重載
	collector.RecordConfigReload("app_config", "success", 10*time.Millisecond)
	collector.RecordConfigReload("plugin_config", "error", 5*time.Millisecond)

	// 測試不會 panic
}

func TestMetricsCollector_AllMethods(t *testing.T) {
	collector := createTestMetricsCollector()

	// 測試所有方法的組合調用
	collector.RecordHTTPRequest("POST", "/api/v1/detectors", "201", 200*time.Millisecond, 2048)
	collector.RecordHTTPRequestInFlight("POST", "/api/v1/detectors", 1.0)
	collector.RecordPluginRequest("threshold_detector", "detector", "success", 100*time.Millisecond)
	collector.RecordPluginHealth("threshold_detector", "detector", true)
	collector.RecordSystemMemory("heap_alloc", 1024*1024)
	collector.RecordSystemGoroutines(150)
	collector.RecordDetection("threshold_detector", "anomaly", 250*time.Millisecond)
	collector.RecordDataImport("csv_importer", "success", 2048*1024)
	collector.RecordError("detector", "threshold_exceeded")
	collector.RecordConfigReload("detector_config", "success", 15*time.Millisecond)
	collector.RecordHTTPRequestInFlight("POST", "/api/v1/detectors", -1.0)

	// 測試不會 panic，所有方法都能正常調用
}

// 基準測試
func BenchmarkMetricsCollector_RecordHTTPRequest(b *testing.B) {
	collector := createTestMetricsCollector()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		collector.RecordHTTPRequest("GET", "/api/v1/users", "200", 100*time.Millisecond, 1024)
	}
}

func BenchmarkMetricsCollector_RecordPluginRequest(b *testing.B) {
	collector := createTestMetricsCollector()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		collector.RecordPluginRequest("csv_importer", "importer", "success", 50*time.Millisecond)
	}
}

func BenchmarkMetricsCollector_RecordSystemMemory(b *testing.B) {
	collector := createTestMetricsCollector()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		collector.RecordSystemMemory("heap_alloc", 1024*1024)
	}
}
