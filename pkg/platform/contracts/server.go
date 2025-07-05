package contracts

import (
	"context"

	"github.com/labstack/echo/v4"
	"github.com/spf13/cobra"
)

// HttpServerProvider 定義了 HTTP 服務的介面。
// 職責: 作為平台 Web 入口，抽象化 HTTP 伺服器的具體實現，處理請求的生命週期。
// AI_PLUGIN_TYPE: "http_server_provider"
// AI_IMPL_PACKAGE: "detectviz-platform/internal/infrastructure/platform/http_server/echo_http_server_provider"
// AI_IMPL_CONSTRUCTOR: "NewEchoHttpServerProvider"
// @See: internal/infrastructure/platform/http_server/echo_http_server_provider.go
type HttpServerProvider interface {
	// Start 在指定的端口上啟動 HTTP 服務。
	Start(port string) error
	// Stop 優雅地停止 HTTP 服務。
	Stop(ctx context.Context) error
	// GetRouter 獲取底層路由實例，用於註冊路由和中介層。
	// 注意: 此處耦合了 Echo，未來可考慮使用更通用的介面以支持其他框架。
	GetRouter() *echo.Echo
	// GetName 返回 HTTP 服務提供者的名稱。
	GetName() string
}

// CliServerProvider 定義了 CLI 服務的介面。
// 職責: 作為平台命令行入口，抽象化 CLI 應用的具體實現，處理命令的解析和執行。
// AI_PLUGIN_TYPE: "cli_server_provider"
// AI_IMPL_PACKAGE: "detectviz-platform/internal/infrastructure/platform/cli_server/cobra_cli_server_provider"
// AI_IMPL_CONSTRUCTOR: "NewCobraCliServerProvider"
// @See: internal/infrastructure/platform/cli_server/cobra_cli_server_provider.go
type CliServerProvider interface {
	// Execute 開始執行 CLI 應用。
	Execute() error
	// AddCommand 將一個新的命令添加到 CLI 應用中。
	AddCommand(cmd *cobra.Command)
	// GetName 返回 CLI 服務提供者的名稱。
	GetName() string
}
