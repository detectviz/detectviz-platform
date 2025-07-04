package metrics

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"
)

func TestNewPrometheusMetricsProvider(t *testing.T) {
	tests := []struct {
		name   string
		config PrometheusConfig
	}{
		{
			name: "啟用配置",
			config: PrometheusConfig{
				Port:    9090,
				Path:    "/metrics",
				Enabled: true,
			},
		},
		{
			name: "禁用配置",
			config: PrometheusConfig{
				Port:    9091,
				Path:    "/metrics",
				Enabled: false,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := NewPrometheusMetricsProvider(tt.config)

			if provider == nil {
				t.Fatal("期望返回非空的提供者")
			}

			if provider.GetName() != "prometheus_metrics_provider" {
				t.Errorf("GetName() = %v, want prometheus_metrics_provider", provider.GetName())
			}

			// 清理
			if p, ok := provider.(*PrometheusMetricsProvider); ok {
				p.Shutdown(context.Background())
			}
		})
	}
}

func TestPrometheusMetricsProvider_IncCounter(t *testing.T) {
	config := PrometheusConfig{
		Port:    9092,
		Path:    "/metrics",
		Enabled: false, // 禁用 HTTP 服務器以避免端口衝突
	}

	provider := NewPrometheusMetricsProvider(config).(*PrometheusMetricsProvider)

	tests := []struct {
		name string
		tags map[string]string
	}{
		{
			name: "無標籤",
			tags: map[string]string{},
		},
		{
			name: "單個標籤",
			tags: map[string]string{"env": "test"},
		},
		{
			name: "多個標籤",
			tags: map[string]string{
				"env":     "test",
				"service": "api",
				"version": "1.0.0",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			counterName := fmt.Sprintf("test_counter_%s", tt.name)

			// 這些操作不應該崩潰
			provider.IncCounter(counterName, tt.tags)
			provider.IncCounter(counterName, tt.tags) // 再次增加

			// 驗證計數器已創建
			if _, exists := provider.counters[counterName]; !exists {
				t.Errorf("期望創建計數器 %s", counterName)
			}
		})
	}

	provider.Shutdown(context.Background())
}

func TestPrometheusMetricsProvider_ObserveHistogram(t *testing.T) {
	config := PrometheusConfig{
		Port:    9093,
		Path:    "/metrics",
		Enabled: false,
	}

	provider := NewPrometheusMetricsProvider(config).(*PrometheusMetricsProvider)

	tests := []struct {
		name  string
		value float64
		tags  map[string]string
	}{
		{
			name:  "小值",
			value: 0.1,
			tags:  map[string]string{"method": "GET"},
		},
		{
			name:  "中等值",
			value: 1.5,
			tags:  map[string]string{"method": "POST"},
		},
		{
			name:  "大值",
			value: 10.0,
			tags:  map[string]string{"method": "PUT"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			histogramName := fmt.Sprintf("test_histogram_%s", tt.name)

			// 這些操作不應該崩潰
			provider.ObserveHistogram(histogramName, tt.value, tt.tags)

			// 驗證直方圖已創建
			if _, exists := provider.histograms[histogramName]; !exists {
				t.Errorf("期望創建直方圖 %s", histogramName)
			}
		})
	}

	provider.Shutdown(context.Background())
}

func TestPrometheusMetricsProvider_SetGauge(t *testing.T) {
	config := PrometheusConfig{
		Port:    9094,
		Path:    "/metrics",
		Enabled: false,
	}

	provider := NewPrometheusMetricsProvider(config).(*PrometheusMetricsProvider)

	tests := []struct {
		name  string
		value float64
		tags  map[string]string
	}{
		{
			name:  "零值",
			value: 0.0,
			tags:  map[string]string{"type": "cpu"},
		},
		{
			name:  "正值",
			value: 85.5,
			tags:  map[string]string{"type": "memory"},
		},
		{
			name:  "負值",
			value: -10.0,
			tags:  map[string]string{"type": "temperature"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gaugeName := fmt.Sprintf("test_gauge_%s", tt.name)

			// 這些操作不應該崩潰
			provider.SetGauge(gaugeName, tt.value, tt.tags)

			// 驗證儀表盤已創建
			if _, exists := provider.gauges[gaugeName]; !exists {
				t.Errorf("期望創建儀表盤 %s", gaugeName)
			}
		})
	}

	provider.Shutdown(context.Background())
}

func TestPrometheusMetricsProvider_HTTPServer(t *testing.T) {
	config := PrometheusConfig{
		Port:    9095,
		Path:    "/metrics",
		Enabled: true,
	}

	provider := NewPrometheusMetricsProvider(config).(*PrometheusMetricsProvider)

	// 等待 HTTP 服務器啟動
	time.Sleep(100 * time.Millisecond)

	// 測試指標端點
	resp, err := http.Get(fmt.Sprintf("http://localhost:%d%s", config.Port, config.Path))
	if err != nil {
		t.Fatalf("無法訪問指標端點: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("期望狀態碼 200，實際為 %d", resp.StatusCode)
	}

	// 添加一些指標
	provider.IncCounter("http_requests_total", map[string]string{"method": "GET"})
	provider.ObserveHistogram("request_duration_seconds", 0.5, map[string]string{"endpoint": "/api"})
	provider.SetGauge("active_connections", 42, map[string]string{"server": "main"})

	// 再次請求指標端點
	resp2, err := http.Get(fmt.Sprintf("http://localhost:%d%s", config.Port, config.Path))
	if err != nil {
		t.Fatalf("無法訪問指標端點: %v", err)
	}
	defer resp2.Body.Close()

	if resp2.StatusCode != http.StatusOK {
		t.Errorf("期望狀態碼 200，實際為 %d", resp2.StatusCode)
	}

	provider.Shutdown(context.Background())
}

func TestPrometheusMetricsProvider_Shutdown(t *testing.T) {
	config := PrometheusConfig{
		Port:    9096,
		Path:    "/metrics",
		Enabled: true,
	}

	provider := NewPrometheusMetricsProvider(config).(*PrometheusMetricsProvider)

	// 等待 HTTP 服務器啟動
	time.Sleep(100 * time.Millisecond)

	// 測試正常關閉
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := provider.Shutdown(ctx)
	if err != nil {
		t.Errorf("Shutdown() error = %v", err)
	}

	// 驗證服務器已關閉
	_, err = http.Get(fmt.Sprintf("http://localhost:%d%s", config.Port, config.Path))
	if err == nil {
		t.Error("期望服務器關閉後無法訪問")
	}

	// 測試重複關閉（不應該崩潰）
	err = provider.Shutdown(ctx)
	if err != nil {
		t.Errorf("Second Shutdown() error = %v", err)
	}
}

func TestPrometheusMetricsProvider_DisabledHTTPServer(t *testing.T) {
	config := PrometheusConfig{
		Port:    9097,
		Path:    "/metrics",
		Enabled: false,
	}

	provider := NewPrometheusMetricsProvider(config).(*PrometheusMetricsProvider)

	// 驗證沒有 HTTP 服務器
	if provider.httpServer != nil {
		t.Error("期望禁用時沒有 HTTP 服務器")
	}

	// 測試關閉（不應該崩潰）
	err := provider.Shutdown(context.Background())
	if err != nil {
		t.Errorf("Shutdown() error = %v", err)
	}
}

func TestGetTagKeys(t *testing.T) {
	tests := []struct {
		name string
		tags map[string]string
		want int
	}{
		{
			name: "空標籤",
			tags: map[string]string{},
			want: 0,
		},
		{
			name: "單個標籤",
			tags: map[string]string{"env": "test"},
			want: 1,
		},
		{
			name: "多個標籤",
			tags: map[string]string{
				"env":     "test",
				"service": "api",
				"version": "1.0.0",
			},
			want: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			keys := getTagKeys(tt.tags)
			if len(keys) != tt.want {
				t.Errorf("getTagKeys() 返回 %d 個鍵，期望 %d 個", len(keys), tt.want)
			}

			// 驗證所有鍵都存在
			keyMap := make(map[string]bool)
			for _, key := range keys {
				keyMap[key] = true
			}

			for expectedKey := range tt.tags {
				if !keyMap[expectedKey] {
					t.Errorf("期望鍵 %s 存在於結果中", expectedKey)
				}
			}
		})
	}
}

// 基準測試
func BenchmarkPrometheusMetricsProvider_IncCounter(b *testing.B) {
	config := PrometheusConfig{
		Port:    9098,
		Path:    "/metrics",
		Enabled: false,
	}

	provider := NewPrometheusMetricsProvider(config).(*PrometheusMetricsProvider)
	tags := map[string]string{"method": "GET", "status": "200"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		provider.IncCounter("benchmark_counter", tags)
	}

	provider.Shutdown(context.Background())
}

func BenchmarkPrometheusMetricsProvider_ObserveHistogram(b *testing.B) {
	config := PrometheusConfig{
		Port:    9099,
		Path:    "/metrics",
		Enabled: false,
	}

	provider := NewPrometheusMetricsProvider(config).(*PrometheusMetricsProvider)
	tags := map[string]string{"endpoint": "/api", "method": "POST"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		provider.ObserveHistogram("benchmark_histogram", float64(i%1000)/1000.0, tags)
	}

	provider.Shutdown(context.Background())
}

func BenchmarkPrometheusMetricsProvider_SetGauge(b *testing.B) {
	config := PrometheusConfig{
		Port:    9100,
		Path:    "/metrics",
		Enabled: false,
	}

	provider := NewPrometheusMetricsProvider(config).(*PrometheusMetricsProvider)
	tags := map[string]string{"instance": "server1"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		provider.SetGauge("benchmark_gauge", float64(i%100), tags)
	}

	provider.Shutdown(context.Background())
}
