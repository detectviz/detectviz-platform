package register

import (
	"github.com/detectviz/detectviz-platform/go-platform/internal/metrics"
	pluginhost "github.com/detectviz/detectviz-platform/go-platform/internal/pluginhost"
	httprequest "github.com/detectviz/detectviz-platform/go-platform/internal/pluginhost/plugins/gateway/http_request"
	knowledge "github.com/detectviz/detectviz-platform/go-platform/internal/pluginhost/plugins/knowledge"
	healthagg "github.com/detectviz/detectviz-platform/go-platform/internal/pluginhost/plugins/observability/health_aggregator"
	"go.uber.org/zap"
)

// RegisterAll 註冊內建插件到 Registry
func RegisterAll(reg *pluginhost.Registry, provider metrics.MetricsProvider) error {
	logger := zap.L().Named("plugin_register")

	// 註冊不依賴 metrics provider 的插件
	reg.Register("gateway/http_request", httprequest.New())
	reg.Register("knowledge/knowledge_provider", knowledge.New())

	// 註冊依賴 metrics provider 的插件
	healthAggregatorPlugin := healthagg.New(provider, logger.Named("health_aggregator"))
	reg.Register("observability/health_aggregator", healthAggregatorPlugin)

	return nil
}
