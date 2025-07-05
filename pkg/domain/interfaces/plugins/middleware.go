package plugins

import "net/http"

// MiddlewarePlugin 定義了 HTTP 中介層插件的介面。
// 職責: 在 HTTP 請求處理鏈中插入可重用的通用邏輯 (如日誌、認證、CORS)。
// AI_PLUGIN_TYPE: "middleware_plugin"
// AI_IMPL_PACKAGE: "detectviz-platform/internal/platform/middleware/auth_middleware"
// AI_IMPL_CONSTRUCTOR: "NewAuthMiddlewarePlugin"
// @See: internal/platform/middleware/auth_middleware.go
type MiddlewarePlugin interface {
	// Handle 接收下一個處理程序並返回一個新的處理程序，實現中介層邏輯。
	Handle(next http.Handler) http.Handler
	// GetName 返回中介層插件的名稱。
	GetName() string
}
