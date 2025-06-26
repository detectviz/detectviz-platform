package contracts

import (
	"context"
	"database/sql"

	"github.com/labstack/echo/v4"
	"github.com/spf13/cobra"
)

// ConfigProvider 定義了平台統一的設定載入和存取介面。
// 職責: 支援讀取不同類型的配置值，並可將配置反序列化到結構體。
// AI 擴展點: AI 可根據部署環境，生成 `ViperConfigProvider` 或 `ConsulConfigProvider` 等實現。
type ConfigProvider interface {
	GetString(key string) string
	GetInt(key string) int
	GetBool(key string) bool
	Unmarshal(rawVal interface{}) error // 將整個配置結構反序列化到 Go struct
	GetName() string
}

// DBClientProvider 定義了資料庫連線能力的通用介面。
// 職責: 負責提供與特定資料庫類型（如 MySQL, PostgreSQL）的連線。
// AI 擴展點: AI 可根據數據庫選擇，生成 `GormMySQLClientProvider` 或 `PgxPostgreSQLClientProvider`。
type DBClientProvider interface {
	GetDB(ctx context.Context) (*sql.DB, error) // 獲取底層 *sql.DB 連線實例
	GetName() string
}

// HttpServerProvider 定義了 HTTP 伺服器啟動和路由註冊的能力。
// 職責: 作為平台 Web 入口，處理 HTTP 請求並分發到 Handler。
// AI 擴展點: AI 可生成 `EchoServerProvider` 或 `GinServerProvider`。
type HttpServerProvider interface {
	Start(port string) error        // 啟動 HTTP 服務
	Stop(ctx context.Context) error // 停止 HTTP 服務
	GetRouter() *echo.Echo          // 獲取底層路由實例，用於註冊路由和中介層 (這裡耦合 Echo，可考慮使用更通用的介面)
	GetName() string
}

// AuthProvider 定義了身份驗證與授權服務的通用介面。
// 職責: 負責驗證用戶身份並提供基礎授權判斷。
// AI 擴展點: AI 可生成 `KeycloakAuthProvider`、`OAuth2Provider` 等具體實現。
type AuthProvider interface {
	Authenticate(ctx context.Context, credentials string) (userID string, err error)
	Authorize(ctx context.Context, userID string, resource string, action string) (bool, error)
	GetName() string
}

// CliServerProvider 定義了 CLI 服務啟動和命令註冊的能力。
// 職責: 作為平台命令行入口，處理命令解析和執行。
// AI 擴展點: AI 可根據業務需求，生成新的 CLI 命令註冊邏輯。
type CliServerProvider interface {
	Execute() error
	AddCommand(cmd *cobra.Command) // 添加命令到 CLI 應用
	GetName() string
}

// MigrationRunner 定義了資料庫 Schema 遷移的通用介面。
// 職責: 管理資料庫結構的版本化控制，確保應用程式與數據庫兼容。
// AI 擴展點: AI 可生成 `AtlasMigrationRunner` 或 `GoMigrateRunner`。
type MigrationRunner interface {
	RunMigrations(ctx context.Context, db *sql.DB) error // 執行 Schema 遷移
	GetName() string
}

// PluginRegistryProvider 定義了平台 plugin 的註冊與 metadata 查詢能力。
// 職責: 管理已載入和可用的插件，提供插件查詢和元數據獲取功能。
// AI 擴展點: AI 可協助維護和擴展核心插件註冊邏輯。
type PluginRegistryProvider interface {
	Register(name string, provider any) error        // 註冊一個具名的插件實例
	Get(name string) (any, error)                    // 獲取指定名稱的插件實例
	List() []string                                  // 列出所有已註冊插件的名稱
	GetMetadata(name string) (map[string]any, error) // 回傳特定插件的描述資訊（版本、作者、狀態等）
	GetName() string                                 // 例如 "core_registry"
}

// KeycloakClientContract 定義了與 Keycloak 外部服務互動的抽象介面。
// 職責: 封裝與 Keycloak 服務進行底層 HTTP/gRPC 通訊的細節。
// AI 擴展點: AI 可生成 `HTTPKeycloakClient` 實現。
type KeycloakClientContract interface {
	VerifyToken(ctx context.Context, token string) (string, error)
	CheckPermissions(ctx context.Context, userID, resource, action string) (bool, error)
}
