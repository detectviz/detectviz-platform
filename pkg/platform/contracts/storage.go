package contracts

import (
	"context"
	"time"
)

// Note: MigrationRunner moved to database.go

// TransactionManager 定義了事務管理服務的介面。
// 職責: 提供數據庫事務的開始、提交和回滾功能，以保證操作的原子性。
// AI_PLUGIN_TYPE: "transaction_manager_provider"
// AI_IMPL_PACKAGE: "detectviz-platform/internal/infrastructure/database/gorm_transaction_manager"
// AI_IMPL_CONSTRUCTOR: "NewGORMTransactionManager"
// @See: internal/infrastructure/database/gorm_transaction_manager.go
type TransactionManager interface {
	// BeginTx 開始一個新的事務並返回一個事務上下文，例如 *gorm.DB 或 *sql.Tx。
	BeginTx(ctx context.Context, opts *interface{}) (interface{}, error)
	// CommitTx 提交一個事務。
	CommitTx(tx interface{}) error
	// RollbackTx 回滾一個事務。
	RollbackTx(tx interface{}) error
	// GetName 返回事務管理器的名稱。
	GetName() string
}

// CacheProvider 定義了緩存服務的介面。
// 職責: 提供統一的鍵值對緩存操作，支持設置過期時間，以提高性能和降低後端負載。
// AI_PLUGIN_TYPE: "cache_provider"
// AI_IMPL_PACKAGE: "detectviz-platform/internal/platform/providers/cache/redis_cache"
// AI_IMPL_CONSTRUCTOR: "NewRedisCacheProvider"
// @See: internal/platform/providers/cache/redis_cache.go
type CacheProvider interface {
	// Set 將一個值存儲到緩存中，並可選地設置過期時間。
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error
	// Get 從緩存中根據鍵名檢索一個值。
	Get(ctx context.Context, key string) (interface{}, error)
	// Delete 從緩存中刪除一個值。
	Delete(ctx context.Context, key string) error
	// GetName 返回緩存提供者的名稱。
	GetName() string
}

// SecretsProvider 定義了秘密管理服務的介面。
// 職責: 安全地讀取和管理敏感資訊 (如 API 金鑰、數據庫憑證)，避免硬編碼在代碼中。
// AI_PLUGIN_TYPE: "secrets_provider"
// AI_IMPL_PACKAGE: "detectviz-platform/internal/platform/providers/secrets/env_secrets"
// AI_IMPL_CONSTRUCTOR: "NewEnvSecretsProvider"
// @See: internal/platform/providers/secrets/env_secrets.go
type SecretsProvider interface {
	// GetSecret 根據鍵名安全地檢索一個秘密值。
	GetSecret(ctx context.Context, key string) (string, error)
	// GetName 返回秘密管理提供者的名稱。
	GetName() string
}
