// 這是一個彙整 Detectviz 平台所有核心介面定義的程式碼區塊。
// 實際專案中，這些介面會分佈在 'pkg/domain/interfaces' 和 'pkg/platform/contracts'
// 下各自的 Go 檔案中，並使用各自的 package 聲明。
// --- pkg/domain (領域層) ---
// 此包定義了 Detectviz 平台的核心業務概念、實體和對應的抽象介面。
// 它是 Clean Architecture 最內層，不依賴任何外部框架或技術細節。
package domain // 實際在各自的 package 中，例如 pkg/domain/entities, pkg/domain/interfaces, pkg/domain/plugins
import (
"context" // 領域層介面應接受 context.Context 以支持追蹤、取消和值傳遞
"time"
)
// --- 領域實體 (Entities) ---
// 定義領域內具有唯一標識和生命週期的核心業務物件。
// 檔案位置: pkg/domain/entities/
// pkg/domain/entities/user.go
// User 是 Detectviz 平台的核心用戶實體。
// 職責: 封裝用戶的基本資訊及與用戶身份相關的業務行為 (例如修改密碼的邏輯)。
type User struct {
ID string
Name string
Email string
Password string // 在領域層，Password 通常指業務層的密碼概念，具體存儲形式(散列)由持久化層處理。
CreatedAt time.Time
UpdatedAt time.Time
}
// pkg/domain/entities/detector.go
// Detector 是 Detectviz 平台的核心偵測器實體。
// 職責: 封裝偵測器的配置、狀態及與偵測器相關的業務行為 (例如啟用/禁用偵測)。
type Detector struct {
ID string
Name string
Config map[string]any
IsEnabled bool
CreatedAt time.Time
UpdatedAt time.Time
}
// --- 領域層介面 (Interfaces) ---
// 定義領域服務和倉儲的抽象契約，這些介面定義了「能做什麼」，而不是「如何做」。
// 檔案位置: pkg/domain/interfaces/
// pkg/domain/interfaces/user_repository.go
// UserRepository 定義了用戶數據的持久化介面。
// 職責: 提供用戶實體的 CRUD 操作。
type UserRepository interface {
GetUserByID(ctx context.Context, id string) (*User, error)
GetUserByEmail(ctx context.Context, email string) (*User, error)
CreateUser(ctx context.Context, user *User) error
UpdateUser(ctx context.Context, user *User) error
DeleteUser(ctx context.Context, id string) error
}
// pkg/domain/interfaces/detector_repository.go
// DetectorRepository 定義了偵測器數據的持久化介面。
// 職責: 提供偵測器實體的 CRUD 操作。
type DetectorRepository interface {
GetDetectorByID(ctx context.Context, id string) (*Detector, error)
CreateDetector(ctx context.Context, detector *Detector) error
UpdateDetector(ctx context.Context, detector *Detector) error
DeleteDetector(ctx context.Context, id string) error
ListDetectors(ctx context.Context, userID string) ([]*Detector, error)
}
// pkg/domain/interfaces/user_service.go
// UserService 定義了與用戶相關的業務邏輯服務介面。
// 職責: 處理用戶註冊、登入、個人資料管理等高層次業務操作。
type UserService interface {
RegisterUser(ctx context.Context, name, email, password string) (*User, error)
AuthenticateUser(ctx context.Context, email, password string) (*User, error)
UpdateUserProfile(ctx context.Context, userID string, updates map[string]any) error
}
// pkg/domain/interfaces/detector_service.go
// DetectorService 定義了與偵測器相關的業務邏輯服務介面。
// 職責: 處理偵測器的創建、配置、啟用/禁用等高層次業務操作。
type DetectorService interface {
CreateDetector(ctx context.Context, name string, config map[string]any, userID string) (*Detector, error)
UpdateDetectorConfig(ctx context.Context, detectorID string, config map[string]any) error
EnableDetector(ctx context.Context, detectorID string) error
DisableDetector(ctx context.Context, detectorID string) error
GetDetectorDetails(ctx context.Context, detectorID string) (*Detector, error)
}
// --- pkg/domain/plugins (領域插件介面) ---
// 定義可擴展的領域層插件介面。
// pkg/domain/plugins/data_processor.go
// DataProcessor 定義了數據處理插件的介面。
// 職責: 對輸入數據進行預處理、轉換、清洗等操作。
type DataProcessor interface {
ProcessData(ctx context.Context, rawData []byte, options map[string]any) ([]byte, error)
GetName() string
}
// pkg/domain/plugins/alert_notifier.go
// AlertNotifier 定義了警報通知插件的介面。
// 職責: 將警報信息發送給指定渠道（例如：郵件、Slack、Webhook）。
type AlertNotifier interface {
Notify(ctx context.Context, alertType string, message string, details map[string]any) error
GetName() string
}
// pkg/domain/plugins/event_handler.go
// EventHandler 定義了事件處理插件的介面。
// 職責: 響應特定的系統事件，執行相應的業務邏輯。
type EventHandler interface {
Handle(ctx context.Context, event Event) error
EventType() string // 返回此處理器關注的事件類型
GetName() string
}
// pkg/domain/interfaces/validator.go
// Validator 定義了通用的數據驗證介面。
// 職責: 對數據結構或輸入進行業務規則驗證。
type Validator interface {
Validate(ctx context.Context, data any) error
GetName() string
}
// --- pkg/platform/contracts (平台契約層) ---
// 此包定義了與外部基礎設施或通用服務互動的抽象介面。
// 它位於 Clean Architecture 的外部層，但通過這些介面保持與領域層的解耦。
package contracts
import (
"context"
"io"
"net/http"
"time"
)
// pkg/platform/contracts/logger.go
// Logger 定義了日誌服務的通用介面。
// 職責: 提供不同日誌級別的記錄功能。
// AI_PLUGIN_TYPE: "logger"
// AI_IMPL_PACKAGE: "detectviz-platform/internal/platform/providers/logger/otelzap_logger" // 假設 OtelZap 實現
// AI_IMPL_CONSTRUCTOR: "NewOtelZapLogger" // 假設其構造函數
type Logger interface {
Debug(ctx context.Context, msg string, fields ...any)
Info(ctx context.Context, msg string, fields ...any)
Warn(ctx context.Context, msg string, fields ...any)
Error(ctx context.Context, err error, msg string, fields ...any)
Fatal(ctx context.Context, err error, msg string, fields ...any)
WithContext(ctx context.Context) Logger // 為日誌添加上下文
GetName() string
}
// pkg/platform/contracts/metrics_provider.go
// MetricsProvider 定義了指標記錄的介面。
// 職責: 提供計數器、測量儀等指標操作，支持可觀測性。
// AI_PLUGIN_TYPE: "metrics_provider"
// AI_IMPL_PACKAGE: "detectviz-platform/internal/platform/providers/metrics/opentelemetry_metrics" // 假設 OpenTelemetry 實現
// AI_IMPL_CONSTRUCTOR: "NewOpenTelemetryMetricsProvider" // 假設其構造函數
type MetricsProvider interface {
IncCounter(ctx context.Context, name string, labels map[string]string)
ObserveGauge(ctx context.Context, name string, value float64, labels map[string]string)
ObserveHistogram(ctx context.Context, name string, value float64, labels map[string]string)
GetName() string
}
// pkg/platform/contracts/tracing_provider.go
// TracingProvider 定義了分散式追蹤的介面。
// 職責: 提供 Span 的開始、結束和屬性設置，支持故障排查。
// AI_PLUGIN_TYPE: "tracing_provider"
// AI_IMPL_PACKAGE: "detectviz-platform/internal/platform/providers/tracing/opentelemetry_tracing" // 假設 OpenTelemetry 實現
// AI_IMPL_CONSTRUCTOR: "NewOpenTelemetryTracingProvider" // 假設其構造函數
type TracingProvider interface {
StartSpan(ctx context.Context, spanName string, opts ...any) (context.Context, any) // returns (context.Context, opentelemetry.Span)
EndSpan(span any) // expects opentelemetry.Span
AddSpanEvent(span any, eventName string, attributes ...any) // expects opentelemetry.Span
GetName() string
}
// pkg/platform/contracts/config_provider.go
// ConfigProvider 定義了配置讀取的介面。
// 職責: 從不同來源讀取應用程式配置。
// AI_PLUGIN_TYPE: "config_provider"
// AI_IMPL_PACKAGE: "detectviz-platform/internal/platform/providers/config/viper_config" // 假設 Viper 實現
// AI_IMPL_CONSTRUCTOR: "NewViperConfigProvider" // 假設其構造函數
type ConfigProvider interface {
GetString(key string) string
GetInt(key string) int
GetBool(key string) bool
Get(key string) any
Unmarshal(rawVal any) error
GetName() string
}
// pkg/platform/contracts/secrets_provider.go
// SecretsProvider 定義了秘密管理服務的介面。
// 職責: 安全地存取敏感配置（例如資料庫密碼、API 金鑰）。
// AI_PLUGIN_TYPE: "secrets_provider"
// AI_IMPL_PACKAGE: "detectviz-platform/internal/platform/providers/secrets/k8s_secrets" // 假設 K8s Secrets 實現
// AI_IMPL_CONSTRUCTOR: "NewK8sSecretsProvider" // 假設其構造函數
type SecretsProvider interface {
GetSecret(ctx context.Context, key string) (string, error)
GetName() string
}
// pkg/platform/contracts/db_client.go
// DBClientProvider 定義了資料庫客戶端獲取介面。
// 職責: 提供獲取底層資料庫連接的能力，供 Repository 實現使用。
// AI_PLUGIN_TYPE: "db_client_provider"
// AI_IMPL_PACKAGE: "detectviz-platform/internal/platform/providers/db/gorm_client" // 假設 GORM 實現
// AI_IMPL_CONSTRUCTOR: "NewGORMClientProvider" // 假設其構造函數
type DBClientProvider interface {
GetDB(ctx context.Context) (any, error) // 返回底層的 *gorm.DB 或 *sql.DB 實例
GetName() string
}
// pkg/platform/contracts/migration_runner.go
// MigrationRunner 定義了資料庫遷移工具的介面。
// 職責: 執行資料庫 Schema 的升級和降級。
// AI_PLUGIN_TYPE: "migration_runner"
// AI_IMPL_PACKAGE: "detectviz-platform/internal/platform/providers/migration/atlas_runner" // 假設 Atlas 實現
// AI_IMPL_CONSTRUCTOR: "NewAtlasMigrationRunner" // 假設其構造函數
type MigrationRunner interface {
Up(ctx context.Context) error
Down(ctx context.Context) error
GetName() string
}
// pkg/platform/contracts/transaction_manager.go
// TransactionManager 定義了事務管理的介面。
// 職責: 提供事務的開始、提交和回滾機制。
// AI_PLUGIN_TYPE: "transaction_manager"
// AI_IMPL_PACKAGE: "detectviz-platform/internal/platform/providers/transaction/gorm_transaction" // 假設 GORM 實現
// AI_IMPL_CONSTRUCTOR: "NewGORMTransactionManager" // 假設其構造函數
type TransactionManager interface {
BeginTx(ctx context.Context) (context.Context, error)
CommitTx(ctx context.Context) error
RollbackTx(ctx context.Context) error
GetName() string
}
// pkg/platform/contracts/http_server.go
// HttpServerProvider 定義了 HTTP 伺服器框架的介面。
// 職責: 啟動 HTTP 伺服器，註冊路由和中介層。
// AI_PLUGIN_TYPE: "http_server_provider"
// AI_IMPL_PACKAGE: "detectviz-platform/internal/platform/providers/http/echo_server" // 假設 Echo 實現
// AI_IMPL_CONSTRUCTOR: "NewEchoHttpServerProvider" // 假設其構造函數
type HttpServerProvider interface {
Start(ctx context.Context, addr string) error
Stop(ctx context.Context) error
RegisterRoutes(routes func(r any)) // r is the underlying router instance (e.g., *echo.Echo)
GetName() string
}
// pkg/platform/contracts/grpc_server.go
// GrpcServerProvider 定義了 gRPC 伺服器框架的介面。
// 職責: 啟動 gRPC 伺服器，註冊服務。
// AI_PLUGIN_TYPE: "grpc_server_provider"
// AI_IMPL_PACKAGE: "detectviz-platform/internal/platform/providers/grpc/grpc_server" // 假設標準 gRPC 實現
// AI_IMPL_CONSTRUCTOR: "NewGrpcServerProvider" // 假設其構造函數
type GrpcServerProvider interface {
Start(ctx context.Context, addr string) error
Stop(ctx context.Context) error
RegisterService(registerFunc func(s any)) // s is the underlying grpc.Server instance
GetName() string
}
// pkg/platform/contracts/event_bus.go
// EventBusProvider 定義了事件總線服務的介面。
// 職責: 提供事件的發布和訂閱機制。
// AI_PLUGIN_TYPE: "event_bus_provider"
// AI_IMPL_PACKAGE: "detectviz-platform/internal/platform/providers/event_bus/nats_event_bus" // 假設 NATS 實現
// AI_IMPL_CONSTRUCTOR: "NewNATSEventBusProvider" // 假設其構造函數
type EventBusProvider interface {
Publish(ctx context.Context, topic string, event Event) error
Subscribe(ctx context.Context, topic string, handler EventHandler) error
GetName() string
}
// Event 是一個通用事件接口，所有事件都應實現此接口。
type Event interface {
EventType() string
EventID() string
OccurredAt() time.Time
Payload() []byte
}
// pkg/platform/contracts/event_factory.go
// EventFactory 定義了事件創建的工廠介面。
// 職責: 根據類型創建具體的事件實例。
// AI_PLUGIN_TYPE: "event_factory"
// AI_IMPL_PACKAGE: "detectviz-platform/internal/platform/providers/event_factory/default_factory" // 假設預設實現
// AI_IMPL_CONSTRUCTOR: "NewDefaultEventFactory" // 假設其構造函數
type EventFactory interface {
NewEvent(eventType string, payload []byte) (Event, error)
GetName() string
}
// pkg/platform/contracts/access_control.go
// AccessControlProvider 定義了權限控制的介面。
// 職責: 檢查用戶是否具有執行特定操作的權限。
// AI_PLUGIN_TYPE: "access_control_provider"
// AI_IMPL_PACKAGE: "detectviz-platform/internal/platform/providers/authz/opa_access_control" // 假設 OPA 實現
// AI_IMPL_CONSTRUCTOR: "NewOPAAccessControlProvider" // 假設其構造函數
type AccessControlProvider interface {
CheckPermission(ctx context.Context, userID, action, resource string) (bool, error)
GetName() string
}
// pkg/platform/contracts/auth_provider.go
// AuthProvider 定義了身份驗證服務的介面。
// 職責: 處理用戶的登入、登出和會話管理。
// AI_PLUGIN_TYPE: "auth_provider"
// AI_IMPL_PACKAGE: "detectviz-platform/internal/platform/providers/auth/keycloak_auth" // 假設 Keycloak 實現
// AI_IMPL_CONSTRUCTOR: "NewKeycloakAuthProvider" // 假設其構造函數
type AuthProvider interface {
Login(ctx context.Context, username, password string) (string, error) // Returns token
ValidateToken(ctx context.Context, token string) (string, error) // Returns userID
GetName() string
}
// pkg/platform/contracts/csrf_token.go
// CSRFTokenProvider 定義了 CSRF token 管理的介面。
// 職責: 生成和驗證 CSRF token，防止跨站請求偽造。
// AI_PLUGIN_TYPE: "csrf_token_provider"
// AI_IMPL_PACKAGE: "detectviz-platform/internal/platform/providers/csrf/default_csrf" // 假設預設實現
// AI_IMPL_CONSTRUCTOR: "NewDefaultCSRFTokenProvider" // 假設其構造函數
type CSRFTokenProvider interface {
GenerateToken(ctx context.Context) (string, error)
ValidateToken(ctx context.Context, token string) (bool, error)
GetName() string
}
// pkg/platform/contracts/rate_limiter.go
// RateLimiter 定義了速率限制服務的介面。
// 職責: 限制對資源或 API 的請求頻率。
// AI_PLUGIN_TYPE: "rate_limiter"
// AI_IMPL_PACKAGE: "detectviz-platform/internal/platform/providers/rate_limiter/redis_rate_limiter" // 假設 Redis 實現
// AI_IMPL_CONSTRUCTOR: "NewRedisRateLimiter" // 假設其構造函數
type RateLimiter interface {
Allow(ctx context.Context, key string, limit int, duration time.Duration) (bool, error)
GetName() string
}
// pkg/platform/contracts/health_checker.go
// HealthChecker 定義了健康檢查的介面。
// 職責: 檢查應用程式或其依賴的健康狀態。
// AI_PLUGIN_TYPE: "health_checker"
// AI_IMPL_PACKAGE: "detectviz-platform/internal/platform/providers/health/default_health" // 假設預設實現
// AI_IMPL_CONSTRUCTOR: "NewDefaultHealthChecker" // 假設其構造函數
type HealthChecker interface {
CheckLiveness(ctx context.Context) error // 檢查應用程式自身是否存活
CheckReadiness(ctx context.Context) error // 檢查應用程式是否準備好處理請求（包括依賴）
GetName() string
}
// pkg/platform/contracts/audit_log.go
// AuditLogProvider 定義了審計記錄的儲存與查詢功能。
// 職責: 記錄關鍵操作、身份與時間資訊，支援合規需求。
type AuditLogProvider interface {
LogAction(ctx context.Context, userID, action, resource string, metadata map[string]any) error
GetName() string
}
// pkg/platform/contracts/session_store.go
// SessionStore 定義了使用者登入狀態與會話的儲存抽象。
// 職責: 管理登入 Session 的生命週期與屬性。
type SessionStore interface {
Set(ctx context.Context, sessionID string, data map[string]any) error
Get(ctx context.Context, sessionID string) (map[string]any, error)
Delete(ctx context.Context, sessionID string) error
GetName() string
}
// pkg/platform/contracts/plugin_metadata.go
// PluginMetadataProvider 定義了插件元資訊的查詢與註冊介面。
// 職責: 提供插件名稱、版本、依賴等資訊，利於平台治理。
type PluginMetadataProvider interface {
GetMetadata(ctx context.Context, pluginName string) (map[string]any, error)
RegisterMetadata(ctx context.Context, pluginName string, metadata map[string]any) error
GetName() string
}
// pkg/platform/contracts/llm_provider.go
// LLMProvider 定義了大型語言模型推論功能的通用介面。
// 職責: 將 prompt 傳入 LLM 並取得模型輸出。
type LLMProvider interface {
GenerateText(ctx context.Context, prompt string, options map[string]any) (string, error)
GetName() string
}
// pkg/platform/contracts/embedding_store.go
// EmbeddingStore 定義了嵌入向量儲存與查詢的介面。
// 職責: 儲存和檢索透過 LLM 生成的嵌入向量。
type EmbeddingStore interface {
StoreEmbedding(ctx context.Context, id string, embedding []float32, metadata map[string]any) error
GetNearest(ctx context.Context, queryEmbedding []float32, k int, filter map[string]any) ([]EmbeddingResult, error)
GetName() string
}
// EmbeddingResult 結構用於表示嵌入向量查詢的結果。
type EmbeddingResult struct {
ID string
Score float32
Data map[string]any
}
// pkg/platform/contracts/file_storage.go
// FileStorageProvider 定義了檔案儲存服務的通用介面。
// 職責: 提供檔案的上傳、下載和刪除功能。
type FileStorageProvider interface {
UploadFile(ctx context.Context, bucket, objectName string, reader io.Reader) error
DownloadFile(ctx context.Context, bucket, objectName string) (io.ReadCloser, error)
DeleteFile(ctx context.Context, bucket, objectName string) error
GetName() string
}
// pkg/platform/contracts/template_renderer.go
// TemplateRenderer 定義了模板渲染服務的介面。
// 職責: 將數據填充到模板中，生成最終的輸出（例如 HTML、郵件內容）。
type TemplateRenderer interface {
Render(ctx context.Context, templateName string, data map[string]any) (string, error)
GetName() string
}
// pkg/platform/contracts/http_client.go
// HTTPClientProvider 定義了 HTTP 客戶端功能，用於進行外部 HTTP 請求。
// 職責: 提供統一的 HTTP 請求介面，便於集成追蹤、指標等通用功能。
// AI_PLUGIN_TYPE: "http_client_provider"
// AI_IMPL_PACKAGE: "detectviz-platform/internal/platform/providers/http_client/default_http_client" // 假設標準庫 http.Client 實現
// AI_IMPL_CONSTRUCTOR: "NewDefaultHTTPClientProvider" // 假設其構造函數
type HTTPClientProvider interface {
Do(req *http.Request) (*http.Response, error)
GetName() string
}
// --- 新增介面 ---
// pkg/platform/contracts/cache.go
// CacheProvider 定義了平台通用緩存服務的介面。
// 職責: 提供鍵值對緩存的存取、設置和刪除功能。
// AI_PLUGIN_TYPE: "cache_provider"
// AI_IMPL_PACKAGE: "detectviz-platform/internal/platform/providers/cache/redis_cache" // 假設 Redis 實現
// AI_IMPL_CONSTRUCTOR: "NewRedisCacheProvider" // 假設其構造函數
type CacheProvider interface {
Set(ctx context.Context, key string, value []byte, ttl time.Duration) error
Get(ctx context.Context, key string) ([]byte, error)
Delete(ctx context.Context, key string) error
GetName() string
}
// pkg/platform/contracts/circuit_breaker.go
// CircuitBreakerProvider 定義了熔斷器服務的介面。
// 職責: 提供對外部服務調用進行保護的熔斷機制。
// AI_PLUGIN_TYPE: "circuit_breaker_provider"
// AI_IMPL_PACKAGE: "detectviz-platform/internal/platform/providers/circuit_breaker/go_hystrix" // 假設 Hystrix Go 實現
// AI_IMPL_CONSTRUCTOR: "NewHystrixCircuitBreakerProvider" // 假設其構造函數
type CircuitBreakerProvider interface {
Execute(name string, run func() error, fallback func(error) error) error
GetName() string
}
// pkg/platform/contracts/service_discovery.go
// ServiceDiscoveryProvider 定義了服務發現能力的介面。
// 職責: 允許服務註冊自身，並發現其他服務的實例。
// AI_PLUGIN_TYPE: "service_discovery_provider"
// AI_IMPL_PACKAGE: "detectviz-platform/internal/platform/providers/service_discovery/k8s_discovery" // 假設 Kubernetes 服務發現實現
// AI_IMPL_CONSTRUCTOR: "NewK8sServiceDiscoveryProvider" // 假設其構造函數
type ServiceDiscoveryProvider interface {
RegisterService(ctx context.Context, serviceName string, instanceID string, address string, port int, metadata map[string]string) error
DeregisterService(ctx context.Context, serviceName string, instanceID string) error
GetInstances(ctx context.Context, serviceName string) ([]ServiceInstance, error)
GetName() string
}
// ServiceInstance 代表一個服務的實例。
type ServiceInstance struct {
ID string
Address string
Port int
Metadata map[string]string
}