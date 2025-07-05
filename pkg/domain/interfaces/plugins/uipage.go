package plugins

// UIPagePlugin 定義 Web UI 頁面插件介面。
// 職責: 允許在平台中動態註冊和提供新的 Web UI 頁面或組件。
// AI 擴展點: AI 可生成 `DashboardUIPagePlugin` 等具體實現，並自動註冊其路由。
type UIPagePlugin interface {
	Plugin                  // 繼承通用 Plugin 介面
	GetRoute() string       // 返回該 UI 頁面對應的 URL 路徑，例如 "/dashboard"
	GetHTMLContent() string // 返回該 UI 頁面所需的 HTML 內容 (可以是包含 WebComponent 的宿主 HTML)
	// RegisterRoute 是一個方便方法，用於將插件的 HTTP 路由註冊到平台 HTTP 伺服器實例。
	// router 和 logger 參數使用 interface{} 以避免 pkg 層對 internal/adapters 層的直接循環依賴。
	// 實際調用時會進行類型斷言。
	RegisterRoute(router interface{}, logger interface{}) error
}
