package pluginhost

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// 插件負載指標
	pluginLoadGauge = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "detectviz_plugin_load",
			Help: "當前插件負載量",
		},
		[]string{"plugin_name"},
	)

	// 健康狀態指標
	pluginHealthGauge = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "detectviz_plugin_health",
			Help: "插件健康狀態 (0=未知,1=正常,2=降級,3=異常,4=關閉中)",
		},
		[]string{"plugin_name"},
	)

	// 插件調用次數計數器
	pluginInvocationCounter = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "detectviz_plugin_invocations_total",
			Help: "插件調用次數總計",
		},
		[]string{"plugin_name", "status"}, // status: success, error, overload, unavailable
	)

	// 插件調用耗時直方圖
	pluginInvocationDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "detectviz_plugin_invocation_duration_seconds",
			Help:    "插件調用耗時分佈",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"plugin_name"},
	)

	// 健康檢查耗時直方圖
	pluginHealthCheckDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "detectviz_plugin_health_check_duration_seconds",
			Help:    "插件健康檢查耗時分佈",
			Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1.0},
		},
		[]string{"plugin_name"},
	)
)

// UpdateMetrics 更新插件指標（由 Registry 定期調用）
func (r *Registry) UpdateMetrics() {
	status := r.GetPluginStatus()
	for name, data := range status {
		info := data.(map[string]interface{})

		// 更新負載指標
		load := info["load"].(int32)
		pluginLoadGauge.WithLabelValues(name).Set(float64(load))

		// 更新健康狀態指標
		health := info["health"].(PluginHealth)
		pluginHealthGauge.WithLabelValues(name).Set(float64(health))
	}
}
