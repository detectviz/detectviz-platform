package register

import (
	pluginhost "github.com/detectviz/detectviz-platform/go-platform/internal/pluginhost"
	httprequest "github.com/detectviz/detectviz-platform/go-platform/internal/pluginhost/plugins/gateway/http_request"
	healthagg "github.com/detectviz/detectviz-platform/go-platform/internal/pluginhost/plugins/observability/health_aggregator"
	knowledge "github.com/detectviz/detectviz-platform/go-platform/internal/pluginhost/plugins/knowledge"
)

// RegisterAll 註冊內建插件到 Registry
func RegisterAll(reg *pluginhost.Registry) error {
	reg.Register("gateway/http_request", httprequest.New())
	reg.Register("observability/health_aggregator", healthagg.New())
	reg.Register("knowledge/knowledge_provider", knowledge.New())
	return nil
}
