package pluginhost

import "context"

// StartSpan 為保留位，後續可換成 OpenTelemetry 具體實作。
func StartSpan(ctx context.Context, name string) (context.Context, func()) {
	// TODO: 導入 OTel Tracer，回傳 span 與結束函式
	f := func() {}
	return ctx, f
}