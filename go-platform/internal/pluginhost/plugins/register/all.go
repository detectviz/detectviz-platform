package register

import (
	pluginhost "github.com/detectviz/detectviz-platform/go-platform/internal/pluginhost"
	httprequest "github.com/detectviz/detectviz-platform/go-platform/internal/pluginhost/plugins/capability.gateway/http_request"
)

// RegisterAll registers built-in/commonly used plugins into the Registry.
func RegisterAll(reg *pluginhost.Registry) error {
	reg.Register("capability.gateway/http_request", httprequest.New())
	return nil
}