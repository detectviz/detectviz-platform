> 這是一個彙整 Detectviz 平台所有核心介面定義的程式碼區塊。
> 實際專案中，這些介面會分佈在 'pkg/domain/interfaces' 和 'pkg/platform/contracts'
> 下各自的 Go 檔案中，並使用各自的 package 聲明。
> AI 標籤重要性:
> 為了實現 AI 驅動的自動化腳手架和程式碼生成，本文件中的介面定義將包含特定的 AI 標籤（例如 AI_PLUGIN_TYPE, AI_IMPL_PACKAGE, AI_IMPL_CONSTRUCTOR）。
> 這些標籤是 AI 理解介面意圖、預期實現路徑和構造函數的強制性指令。
> AI 將嚴格依據這些標籤來生成符合平台規範的程式碼骨架和組裝邏輯。
> 開發者在新增或修改介面時，必須同時維護這些 AI 標籤，以確保 AI 輔助開發流程的順暢與正確性。
> 詳細的 AI 標籤規範和腳手架工作流程，請參考 docs/ai_scaffold/scaffold_workflow.md。

### --- pkg/domain (領域層) ---
> 此包定義了 Detectviz 平台的核心業務概念、實體和對應的抽象介面。
> 它是 Clean Architecture 最內層，不依賴任何外部框架或技術細節。
```go
package domain // 實際在各自的 package 中，例如 pkg/domain/entities, pkg/domain/interfaces, pkg/domain/plugins
import (
"context" // 領域層介面應接受 context.Context 以支持追蹤、取消和值傳遞
"time"
)
```

### --- 領域實體 (Entities) ---
> 定義領域內具有唯一標識和生命週期的核心業務物件。
> 檔案位置: pkg/domain/entities/
> pkg/domain/entities/user.go
> User 是 Detectviz 平台的核心用戶實體。
> 職責: 封裝用戶的基本資訊及與用戶身份相關的業務行為 (例如修改密碼的邏輯)。
```go
type User struct {
ID string
Name string
Email string
Password string // 在領域層，Password 通常指業務層的密碼概念，具體存儲形式(散列)由持久化層處理。
CreatedAt time.Time
UpdatedAt time.Time
}
```
> pkg/domain/entities/detector.go
> Detector 是 Detectviz 平台的核心偵測器實體。
> 職責: 封裝偵測器的配置、狀態及與偵測器相關的業務行為 (例如啟用/禁用偵測)。
```go
type Detector struct {
ID string
Name string
Type string // 例如 "anomaly_detection", "pattern_recognition"
Config map[string]interface{} // 偵測器特有的配置，由具體插件定義其 Schema
IsEnabled bool
CreatedAt time.Time
UpdatedAt time.Time
CreatedBy string // 創建者用戶ID
LastUpdatedBy string // 最後更新者用戶ID
}
```
> pkg/domain/entities/analysis_result.go
> AnalysisResult 封裝了偵測器執行後的分析結果。
> 職責: 儲存偵測到的異常、模式或洞察，以及相關的元數據。
```go
type AnalysisResult struct {
ID string
DetectorID string
Timestamp time.Time
Severity string // 例如 "low", "medium", "high", "critical"
Summary string // 簡要的分析結果概述
Details map[string]interface{} // 詳細的分析數據，例如異常點、相關指標等
Acknowledged bool // 是否已被確認處理
AcknowledgedBy string // 確認人用戶ID
AcknowledgedAt time.Time // 確認時間
}
```

### --- 領域介面 (Interfaces) ---
> 定義領域服務和倉儲的抽象契約。
> 檔案位置: pkg/domain/interfaces/
> pkg/domain/interfaces/user_repository.go
> UserRepository 定義了用戶數據的持久化操作介面。
> 職責: 提供用戶實體的 CRUD (創建、讀取、更新、刪除) 操作抽象。
> AI_PLUGIN_TYPE: "user_repository"
> AI_IMPL_PACKAGE: "detectviz-platform/internal/infrastructure/repositories/mysql_user_repository"
> AI_IMPL_CONSTRUCTOR: "NewMySQLUserRepository"
```go
type UserRepository interface {
Save(ctx context.Context, user *User) error
FindByID(ctx context.Context, id string) (*User, error)
FindByEmail(ctx context.Context, email string) (*User, error)
Update(ctx context.Context, user *User) error
Delete(ctx context.Context, id string) error
GetName() string
}
```
> pkg/domain/interfaces/detector_repository.go
> DetectorRepository 定義了偵測器數據的持久化操作介面。
> 職責: 提供偵測器實體的 CRUD 操作抽象。
> AI_PLUGIN_TYPE: "detector_repository"
> AI_IMPL_PACKAGE: "detectviz-platform/internal/infrastructure/repositories/mysql_detector_repository"
> AI_IMPL_CONSTRUCTOR: "NewMySQLDetectorRepository"
```go
type DetectorRepository interface {
Save(ctx context.Context, detector *Detector) error
FindByID(ctx context.Context, id string) (*Detector, error)
FindAll(ctx context.Context) ([]*Detector, error)
Update(ctx context.Context, detector *Detector) error
Delete(ctx context.Context, id string) error
GetName() string
}
```
> pkg/domain/interfaces/analysis_result_repository.go
> AnalysisResultRepository 定義了分析結果數據的持久化操作介面。
> 職責: 提供分析結果實體的 CRUD 操作抽象。
> AI_PLUGIN_TYPE: "analysis_result_repository"
> AI_IMPL_PACKAGE: "detectviz-platform/internal/infrastructure/repositories/mysql_analysis_result_repository"
> AI_IMPL_CONSTRUCTOR: "NewMySQLAnalysisResultRepository"
```go
type AnalysisResultRepository interface {
Save(ctx context.Context, result *AnalysisResult) error
FindByID(ctx context.Context, id string) (*AnalysisResult, error)
FindByDetectorID(ctx context.Context, detectorID string) ([]*AnalysisResult, error)
Update(ctx context.Context, result *AnalysisResult) error
Delete(ctx context.Context, id string) error
GetName() string
}
```
> pkg/domain/interfaces/user_service.go
> UserService 定義了用戶相關的業務邏輯介面。
> 職責: 協調 UserRepository 和其他領域服務，處理用戶註冊、登入、資料更新等業務流程。
> AI_PLUGIN_TYPE: "user_service"
> AI_IMPL_PACKAGE: "detectviz-platform/internal/domain_logic/services/user_service"
> AI_IMPL_CONSTRUCTOR: "NewUserService"
```go
type UserService interface {
RegisterUser(ctx context.Context, name, email, password string) (*User, error)
AuthenticateUser(ctx context.Context, email, password string) (*User, error)
GetUserByID(ctx context.Context, id string) (*User, error)
UpdateUserProfile(ctx context.Context, id string, updates map[string]interface{}) (*User, error)
DeleteUser(ctx context.Context, id string) error
GetName() string
}
```
> pkg/domain/interfaces/detector_service.go
> DetectorService 定義了偵測器相關的業務邏輯介面。
> 職責: 協調 DetectorRepository 和 DetectorPlugin，管理偵測器的生命週期和執行。
> AI_PLUGIN_TYPE: "detector_service"
> AI_IMPL_PACKAGE: "detectviz-platform/internal/domain_logic/services/detector_service"
> AI_IMPL_CONSTRUCTOR: "NewDetectorService"
```go
type DetectorService interface {
CreateDetector(ctx context.Context, name, detectorType string, config map[string]interface{}) (*Detector, error)
GetDetector(ctx context.Context, id string) (*Detector, error)
ListDetectors(ctx context.Context) ([]*Detector, error)
UpdateDetector(ctx context.Context, id string, updates map[string]interface{}) (*Detector, error)
DeleteDetector(ctx context.Context, id string) error
ExecuteDetector(ctx context.Context, id string, data map[string]interface{}) (*AnalysisResult, error) // 執行偵測器
GetName() string
}
```
> pkg/domain/interfaces/analysis_service.go
> AnalysisService 定義了分析結果相關的業務邏輯介面。
> 職責: 協調 AnalysisResultRepository 和 AnalysisEnginePlugin，處理分析結果的查詢、確認和歸檔。
> AI_PLUGIN_TYPE: "analysis_service"
> AI_IMPL_PACKAGE: "detectviz-platform/internal/domain_logic/services/analysis_service"
> AI_IMPL_CONSTRUCTOR: "NewAnalysisService"
```go
type AnalysisService interface {
GetAnalysisResult(ctx context.Context, id string) (*AnalysisResult, error)
ListAnalysisResultsByDetector(ctx context.Context, detectorID string) ([]*AnalysisResult, error)
AcknowledgeResult(ctx context.Context, id, userID string) (*AnalysisResult, error)
GetName() string
}
```
### --- 領域插件 (Plugins) ---
> 定義可插拔的領域功能介面。
> 檔案位置: pkg/domain/plugins/
> pkg/domain/plugins/plugin.go
> Plugin 是所有 Detectviz 平台插件的基礎介面。
> 職責: 提供插件的通用方法，如獲取插件名稱。
```go
type Plugin interface {
GetName() string
}
```
> pkg/domain/plugins/importer.go
> ImporterPlugin 定義了數據導入插件的介面。
> 職責: 從不同的數據源導入數據到平台。
> AI_PLUGIN_TYPE: "importer_plugin"
> AI_IMPL_PACKAGE: "detectviz-platform/internal/domain_logic/plugins/importer/csv_importer"
> AI_IMPL_CONSTRUCTOR: "NewCSVImporterPlugin"
```go
type ImporterPlugin interface {
Plugin
ImportData(ctx context.Context, sourceConfig map[string]interface{}) (map[string]interface{}, error)
}
```
> pkg/domain/plugins/detector.go
> DetectorPlugin 定義了具體偵測器實現的介面。
> 職責: 執行特定類型的數據偵測邏輯。
> AI_PLUGIN_TYPE: "detector_plugin"
> AI_IMPL_PACKAGE: "detectviz-platform/internal/domain_logic/plugins/detector/anomaly_detector"
> AI_IMPL_CONSTRUCTOR: "NewAnomalyDetectorPlugin"
```go
type DetectorPlugin interface {
Plugin
Execute(ctx context.Context, data map[string]interface{}, detectorConfig map[string]interface{}) (*AnalysisResult, error)
}
```
> pkg/domain/plugins/analysis_engine.go
> AnalysisEnginePlugin 定義了數據分析引擎插件的介面。
> 職責: 對偵測結果進行深度分析和歸因。
> AI_PLUGIN_TYPE: "analysis_engine_plugin"
> AI_IMPL_PACKAGE: "detectviz-platform/internal/domain_logic/plugins/analysis_engine/llm_analysis_engine"
> AI_IMPL_CONSTRUCTOR: "NewLLMAnalysisEnginePlugin"
```go
type AnalysisEnginePlugin interface {
Plugin
Analyze(ctx context.Context, result *AnalysisResult, analysisConfig map[string]interface{}) (map[string]interface{}, error)
}
```
> pkg/domain/plugins/notification.go
> NotificationPlugin 定義了通知發送插件的介面。
> 職責: 負責通過不同渠道（如郵件、簡訊）發送通知。
> AI_PLUGIN_TYPE: "notification_plugin"
> AI_IMPL_PACKAGE: "detectviz-platform/internal/domain_logic/plugins/notification/email_notifier"
> AI_IMPL_CONSTRUCTOR: "NewEmailNotifierPlugin"
```go
type NotificationPlugin interface {
Plugin
SendNotification(ctx context.Context, recipient, subject, body string, metadata map[string]interface{}) error
}
```
> pkg/domain/plugins/alert.go
> AlertPlugin 定義了告警觸發插件的介面。
> 職責: 將偵測到的異常轉換為告警，並集成到告警系統。
> AI_PLUGIN_TYPE: "alert_plugin"
> AI_IMPL_PACKAGE: "detectviz-platform/internal/domain_logic/plugins/alert/slack_alerter"
> AI_IMPL_CONSTRUCTOR: "NewSlackAlerterPlugin"
```go
type AlertPlugin interface {
Plugin
TriggerAlert(ctx context.Context, result *AnalysisResult, alertConfig map[string]interface{}) error
}
```
> pkg/domain/plugins/ui_page.go
> UIPagePlugin 定義了動態 UI 頁面插件的介面。
> 職責: 允許插件註冊新的前端頁面或組件，擴展平台 UI。
> AI_PLUGIN_TYPE: "ui_page_plugin"
> AI_IMPL_PACKAGE: "detectviz-platform/internal/domain_logic/plugins/ui_page/dashboard_page"
> AI_IMPL_CONSTRUCTOR: "NewDashboardPagePlugin"
```go
type UIPagePlugin interface {
Plugin
GetRoutePath() string
GetTemplateName() string
GetData(ctx context.Context, params map[string]string) (map[string]interface{}, error)
}
```
> pkg/domain/plugins/cli.go
> CLIPlugin 定義了命令行界面擴展插件的介面。
> 職責: 允許插件向平台的 CLI 工具註冊新的命令。
> AI_PLUGIN_TYPE: "cli_plugin"
> AI_IMPL_PACKAGE: "detectviz-platform/internal/domain_logic/plugins/cli/detector_cli"
> AI_IMPL_CONSTRUCTOR: "NewDetectorCLIPlugin"
```go
type CLIPlugin interface {
Plugin
GetCommandName() string
GetDescription() string
Execute(ctx context.Context, args []string) (string, error)
}
```
### --- pkg/platform/contracts (平台契約層) ---
> 此包定義了 Detectviz 平台級基礎設施服務的抽象介面。
> 這些介面是平台核心功能與其具體實現之間的契約。
package contracts // 實際在各自的 package 中，例如 pkg/platform/contracts
import (
"context"
"time"
)
> pkg/platform/contracts/config.go
> ConfigProvider 定義了配置管理服務的介面。
> 職責: 提供應用程式和插件配置的讀取功能。
> AI_PLUGIN_TYPE: "config_provider"
> AI_IMPL_PACKAGE: "detectviz-platform/internal/platform/providers/config/viper_config"
> AI_IMPL_CONSTRUCTOR: "NewViperConfigProvider"
```go
type ConfigProvider interface {
LoadAppConfig(ctx context.Context, configPath string) (map[string]interface{}, error)
LoadPluginConfig(ctx context.Context, pluginName string) (map[string]interface{}, error)
GetString(key string) string
GetInt(key string) int
GetBool(key string) bool
Get(key string) interface{} // 通用獲取任意類型配置
Unmarshal(key string, rawVal interface{}) error // 將配置反序列化到結構體
GetName() string
}
```
> pkg/platform/contracts/http_server.go
> HttpServerProvider 定義了 HTTP 服務的介面。
> 職責: 啟動 HTTP 伺服器，註冊路由，處理請求。
> AI_PLUGIN_TYPE: "http_server_provider"
> AI_IMPL_PACKAGE: "detectviz-platform/internal/platform/providers/http_server/echo_server"
> AI_IMPL_CONSTRUCTOR: "NewEchoServerProvider"
```go
type HttpServerProvider interface {
Start(ctx context.Context) error
Stop(ctx context.Context) error
RegisterRoute(method, path string, handler interface{}) error // handler 可以是 func(echo.Context) error 或 http.HandlerFunc
GetName() string
}
```
> pkg/platform/contracts/logger.go
> LoggerProvider 定義了日誌記錄服務的介面。
> 職責: 提供不同日誌級別的記錄功能，支持結構化日誌。
> AI_PLUGIN_TYPE: "logger_provider"
> AI_IMPL_PACKAGE: "detectviz-platform/internal/platform/providers/logger/otelzap_logger"
> AI_IMPL_CONSTRUCTOR: "NewOtelZapLoggerProvider"
```go
type LoggerProvider interface {
Debug(msg string, fields ...interface{})
Info(msg string, fields ...interface{})
Warn(msg string, fields ...interface{})
Error(msg string, fields ...interface{})
Fatal(msg string, fields ...interface{})
WithName(name string) LoggerProvider // 為日誌添加名稱或上下文
GetName() string
}
```
> pkg/platform/contracts/database.go
> DBClientProvider 定義了資料庫客戶端連接的介面。
> 職責: 提供資料庫連接池，並允許獲取底層數據庫實例。
> AI_PLUGIN_TYPE: "gorm_mysql_client_provider"
> AI_IMPL_PACKAGE: "detectviz-platform/internal/infrastructure/database/gorm_mysql_client"
> AI_IMPL_CONSTRUCTOR: "NewGORMMySQLClientProvider"
```go
type DBClientProvider interface {
Connect(ctx context.Context) error
Disconnect(ctx context.Context) error
GetDB(ctx context.Context) interface{} // 返回底層數據庫實例，例如 *gorm.DB 或 *sql.DB
GetName() string
}
```
> pkg/platform/contracts/auth.go
> AuthProvider 定義了身份驗證服務的介面。
> 職責: 處理用戶認證、授權，生成和驗證令牌。
> AI_PLUGIN_TYPE: "keycloak_auth_provider"
> AI_IMPL_PACKAGE: "detectviz-platform/internal/platform/providers/auth/keycloak_auth"
> AI_IMPL_CONSTRUCTOR: "NewKeycloakAuthProvider"
```go
type AuthProvider interface {
VerifyToken(ctx context.Context, token string) (map[string]interface{}, error) // 驗證 JWT 並返回 claims
GenerateToken(ctx context.Context, userID string, roles []string) (string, error)
GetName() string
}
```
> pkg/platform/contracts/plugin_registry.go
> PluginRegistry 定義了插件註冊與查詢的介面。
> 職責: 管理平台中所有已註冊的插件實例。
> AI_PLUGIN_TYPE: "plugin_registry_provider"
> AI_IMPL_PACKAGE: "detectviz-platform/internal/platform/plugin_registry/registry"
> AI_IMPL_CONSTRUCTOR: "NewPluginRegistry"
```go
type PluginRegistry interface {
RegisterPlugin(pluginType string, factory func(ctx context.Context, configProvider ConfigProvider) (Plugin, error)) error
GetPlugin(pluginType string, name string) (Plugin, error) // 根據類型和名稱獲取特定插件
ListPlugins(pluginType string) ([]Plugin, error) // 列出某類所有插件
GetName() string
}
```
> pkg/platform/contracts/transaction_manager.go
> TransactionManager 定義了事務管理服務的介面。
> 職責: 提供數據庫事務的開始、提交和回滾功能。
> AI_PLUGIN_TYPE: "transaction_manager_provider"
> AI_IMPL_PACKAGE: "detectviz-platform/internal/infrastructure/database/gorm_transaction_manager"
> AI_IMPL_CONSTRUCTOR: "NewGORMTransactionManager"
```go
type TransactionManager interface {
BeginTx(ctx context.Context, opts *interface{}) (interface{}, error) // 返回一個事務上下文，例如 *gorm.DB 或 *sql.Tx
CommitTx(tx interface{}) error
RollbackTx(tx interface{}) error
GetName() string
}
```
> pkg/platform/contracts/cache.go
> CacheProvider 定義了緩存服務的介面。
> 職責: 提供鍵值對緩存操作，支持設置過期時間。
> AI_PLUGIN_TYPE: "cache_provider"
> AI_IMPL_PACKAGE: "detectviz-platform/internal/platform/providers/cache/redis_cache"
> AI_IMPL_CONSTRUCTOR: "NewRedisCacheProvider"
```go
type CacheProvider interface {
Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error
Get(ctx context.Context, key string) (interface{}, error)
Delete(ctx context.Context, key string) error
GetName() string
}
```
> pkg/platform/contracts/rate_limiter.go
> RateLimiterProvider 定義了速率限制服務的介面。
> 職責: 控制請求流量，防止服務過載。
> AI_PLUGIN_TYPE: "rate_limiter_provider"
> AI_IMPL_PACKAGE: "detectviz-platform/internal/platform/providers/rate_limiter/uber_rate_limiter"
> AI_IMPL_CONSTRUCTOR: "NewUberRateLimiterProvider"
```go
type RateLimiterProvider interface {
Allow(ctx context.Context, key string) bool
GetName() string
}
```
> pkg/platform/contracts/circuit_breaker.go
> CircuitBreakerProvider 定義了熔斷器服務的介面。
> 職責: 在外部服務失敗時，快速失敗並提供降級處理，防止級聯故障。
> AI_PLUGIN_TYPE: "circuit_breaker_provider"
> AI_IMPL_PACKAGE: "detectviz-platform/internal/platform/providers/circuit_breaker/hystrix_breaker"
> AI_IMPL_CONSTRUCTOR: "NewHystrixCircuitBreakerProvider"
```go
type CircuitBreakerProvider interface {
Execute(ctx context.Context, name string, run func() error, fallback func(error) error) error
GetName() string
}
```
> pkg/platform/contracts/event_bus.go
> EventBusProvider 定義了事件總線服務的介面。
> 職責: 提供異步事件的發布和訂閱機制。
> AI_PLUGIN_TYPE: "event_bus_provider"
> AI_IMPL_PACKAGE: "detectviz-platform/internal/platform/providers/event_bus/nats_event_bus"
> AI_IMPL_CONSTRUCTOR: "NewNATSEventBusProvider"
```go
type EventBusProvider interface {
Publish(ctx context.Context, topic string, event interface{}) error
Subscribe(ctx context.Context, topic string, handler func(event interface{})) error
GetName() string
}
```
> pkg/platform/contracts/secrets.go
> SecretsProvider 定義了秘密管理服務的介面。
> 職責: 安全地讀取和管理敏感資訊 (如 API 金鑰、數據庫憑證)。
> AI_PLUGIN_TYPE: "secrets_provider"
> AI_IMPL_PACKAGE: "detectviz-platform/internal/platform/providers/secrets/env_secrets"
> AI_IMPL_CONSTRUCTOR: "NewEnvSecretsProvider"
```go
type SecretsProvider interface {
GetSecret(ctx context.Context, key string) (string, error)
GetName() string
}
```
> pkg/platform/contracts/middleware.go
> MiddlewarePlugin 定義了 HTTP 中介層插件的介面。
> 職責: 在 HTTP 請求處理鏈中插入通用邏輯 (如日誌、認證、CORS)。
> AI_PLUGIN_TYPE: "middleware_plugin"
> AI_IMPL_PACKAGE: "detectviz-platform/internal/platform/middleware/auth_middleware"
> AI_IMPL_CONSTRUCTOR: "NewAuthMiddlewarePlugin"
```go
type MiddlewarePlugin interface {
Plugin
Apply(handler interface{}) interface{} // 應用中介層到給定的 HTTP 處理器
}
```
> pkg/platform/contracts/migration_runner.go
> MigrationRunner 定義了資料庫遷移執行器的介面。
> 職責: 執行資料庫 Schema 遷移。
> AI_PLUGIN_TYPE: "migration_runner_provider"
> AI_IMPL_PACKAGE: "detectviz-platform/internal/infrastructure/database/atlas_migration_runner"
> AI_IMPL_CONSTRUCTOR: "NewAtlasMigrationRunner"
```go
type MigrationRunner interface {
RunMigrations(ctx context.Context) error
GetName() string
}
```
> pkg/platform/contracts/audit_log.go
> AuditLogProvider 定義了審計記錄的儲存與查詢功能。
> 職責: 記錄關鍵操作、身份與時間資訊，支援合規需求。
> AI_PLUGIN_TYPE: "audit_log_provider"
> AI_IMPL_PACKAGE: "detectviz-platform/internal/platform/providers/audit_log/db_audit_log"
> AI_IMPL_CONSTRUCTOR: "NewDBAuditLogProvider"
```go
type AuditLogProvider interface {
LogAction(ctx context.Context, userID, action, resource string, metadata map[string]any) error
GetName() string
}
```
> pkg/platform/contracts/session_store.go
> SessionStore 定義了使用者登入狀態與會話的儲存抽象。
> 職責: 管理登入 Session 的生命週期與屬性。
> AI_PLUGIN_TYPE: "session_store_provider"
> AI_IMPL_PACKAGE: "detectviz-platform/internal/platform/providers/session_store/redis_session_store"
> AI_IMPL_CONSTRUCTOR: "NewRedisSessionStoreProvider"
```go
type SessionStore interface {
Set(ctx context.Context, sessionID string, data map[string]any) error
Get(ctx context.Context, sessionID string) (map[string]any, error)
Delete(ctx context.Context, sessionID string) error
GetName() string
}
```
> pkg/platform/contracts/plugin_metadata.go
> PluginMetadataProvider 定義了插件元資訊的查詢與註冊介面。
> 職責: 提供插件名稱、版本、依賴等資訊，利於平台治理。
> AI_PLUGIN_TYPE: "plugin_metadata_provider"
> AI_IMPL_PACKAGE: "detectviz-platform/internal/platform/providers/plugin_metadata/in_memory_plugin_metadata"
> AI_IMPL_CONSTRUCTOR: "NewInMemoryPluginMetadataProvider"
```go
type PluginMetadataProvider interface {
GetMetadata(ctx context.Context, pluginName string) (map[string]any, error)
RegisterMetadata(ctx context.Context, pluginName string, metadata map[string]any) error
GetName() string
}
```
> pkg/platform/contracts/llm_provider.go
> LLMProvider 定義了大型語言模型推論功能的通用介面。
> 職責: 將 prompt 傳入 LLM 並取得模型輸出。
> AI_PLUGIN_TYPE: "llm_provider"
> AI_IMPL_PACKAGE: "detectviz-platform/internal/platform/providers/llm/gemini_llm"
> AI_IMPL_CONSTRUCTOR: "NewGeminiLLMProvider"
```go
type LLMProvider interface {
GenerateText(ctx context.Context, prompt string, options map[string]any) (string, error)
GetName() string
}
```
> pkg/platform/contracts/embedding_store.go
> EmbeddingStoreProvider 定義了向量嵌入儲存與查詢功能的介面。
> 職責: 儲存和檢索高維向量，支持相似性搜索。
> AI_PLUGIN_TYPE: "embedding_store_provider"
> AI_IMPL_PACKAGE: "detectviz-platform/internal/platform/providers/embedding_store/chroma_embedding_store"
> AI_IMPL_CONSTRUCTOR: "NewChromaEmbeddingStoreProvider"
```go
type EmbeddingStoreProvider interface {
StoreEmbedding(ctx context.Context, id string, vector []float32, metadata map[string]any) error
QueryNearest(ctx context.Context, queryVector []float32, topK int, filter map[string]any) ([]string, error) // 返回最相似的 ID
GetName() string
}
```
> pkg/platform/contracts/metrics_provider.go
> MetricsProvider 定義了指標收集與導出的介面。
> 職責: 提供應用程式運行時指標的記錄功能。
> AI_PLUGIN_TYPE: "metrics_provider"
> AI_IMPL_PACKAGE: "detectviz-platform/internal/platform/providers/metrics/otel_metrics"
> AI_IMPL_CONSTRUCTOR: "NewOtelMetricsProvider"
```go
type MetricsProvider interface {
IncCounter(name string, tags map[string]string)
ObserveHistogram(name string, value float64, tags map[string]string)
SetGauge(name string, value float64, tags map[string]string)
GetName() string
}
```
> pkg/platform/contracts/tracing_provider.go
> TracingProvider 定義了分佈式追蹤的介面。
> 職責: 提供 Span 的創建、管理和上下文傳播功能。
> AI_PLUGIN_TYPE: "tracing_provider"
> AI_IMPL_PACKAGE: "detectviz-platform/internal/platform/providers/tracing/otel_tracing"
> AI_IMPL_CONSTRUCTOR: "NewOtelTracingProvider"
```go
type TracingProvider interface {
StartSpan(ctx context.Context, name string, opts ...interface{}) (context.Context, interface{}) // 返回新的上下文和 Span
EndSpan(span interface{})
GetName() string
}
```
> pkg/platform/contracts/error_factory.go
> ErrorFactory 定義了錯誤創建和標準化的介面。
> 職責: 提供統一的錯誤創建機制，包含錯誤碼和可讀訊息。
> AI_PLUGIN_TYPE: "error_factory_provider"
> AI_IMPL_PACKAGE: "detectviz-platform/internal/platform/providers/error_factory/standard_error_factory"
> AI_IMPL_CONSTRUCTOR: "NewStandardErrorFactory"
```go
type ErrorFactory interface {
NewBadRequestError(message string, details ...map[string]any) error
NewNotFoundError(message string, details ...map[string]any) error
NewUnauthorizedError(message string, details ...map[string]any) error
NewInternalServerError(message string, details ...map[string]any) error
NewErrorf(format string, args ...any) error // 類似 fmt.Errorf 但返回標準錯誤類型
GetName() string
}
```
> pkg/platform/contracts/csrf_token_provider.go
> CSRFTokenProvider 定義了 CSRF Token 管理的介面。
> 職責: 生成、驗證和管理用於防範 CSRF 攻擊的 Token。
> AI_PLUGIN_TYPE: "csrf_token_provider"
> AI_IMPL_PACKAGE: "detectviz-platform/internal/platform/providers/csrf_token/default_csrf_token"
> AI_IMPL_CONSTRUCTOR: "NewDefaultCSRFTokenProvider"
```go
type CSRFTokenProvider interface {
GenerateToken(ctx context.Context) (string, error)
ValidateToken(ctx context.Context, token string) error
GetName() string
}
```
> pkg/platform/contracts/service_discovery.go
> ServiceDiscoveryProvider 定義了服務發現的介面。
> 職責: 註冊、註銷服務實例，並查詢可用服務實例的地址。
> AI_PLUGIN_TYPE: "service_discovery_provider"
> AI_IMPL_PACKAGE: "detectviz-platform/internal/platform/providers/service_discovery/k8s_discovery" // 假設 Kubernetes 服務發現實現
> AI_IMPL_CONSTRUCTOR: "NewK8sServiceDiscoveryProvider" // 假設其構造函數
```go
type ServiceDiscoveryProvider interface {
RegisterService(ctx context.Context, serviceName string, instanceID string, address string, port int, metadata map[string]string) error
DeregisterService(ctx context.Context, serviceName string, instanceID string) error
GetInstances(ctx context.Context, serviceName string) ([]ServiceInstance, error)
GetName() string
}
```
> ServiceInstance 代表一個服務的實例。
```go
type ServiceInstance struct {
ID string
Address string
Port int
Metadata map[string]string
}
```
