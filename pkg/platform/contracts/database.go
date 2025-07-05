package contracts

import (
	"context"
	"database/sql"
)

// DBClientProvider 定義了資料庫客戶端連接的介面。
// 職責: 負責提供與特定資料庫類型（如 MySQL, PostgreSQL）的連線池和底層連接實例。
// AI_PLUGIN_TYPE: "gorm_mysql_client_provider"
// AI_IMPL_PACKAGE: "detectviz-platform/internal/infrastructure/database/gorm_mysql_client"
// AI_IMPL_CONSTRUCTOR: "NewGORMMySQLClientProvider"
// @See: internal/infrastructure/database/gorm_mysql_client.go
type DBClientProvider interface {
	// GetDB 獲取底層的 *sql.DB 連線實例，用於執行原生 SQL 或傳遞給其他庫。
	GetDB(ctx context.Context) (*sql.DB, error)
	// GetName 返回數據庫客戶端提供者的名稱。
	GetName() string
}

// MigrationRunner 定義了資料庫 Schema 遷移的通用介面。
// 職責: 管理資料庫結構的版本化控制，確保應用程式與數據庫 Schema 的兼容性。
// AI_PLUGIN_TYPE: "migration_runner_provider"
// AI_IMPL_PACKAGE: "detectviz-platform/internal/infrastructure/database/atlas_migration_runner"
// AI_IMPL_CONSTRUCTOR: "NewAtlasMigrationRunner"
// @See: internal/infrastructure/database/atlas_migration_runner.go
type MigrationRunner interface {
	// RunMigrations 根據遷移文件執行必要的 Schema 變更。
	RunMigrations(ctx context.Context, db *sql.DB) error
	// GetName 返回遷移執行器的名稱。
	GetName() string
}
