# 介面定義

這是一個彙整 Detectviz 平台所有核心介面定義的程式碼區塊。實際專案中，這些介面會分佈在 'pkg/domain/interfaces' 和 'pkg/platform/contracts'下各自的 Go 檔案中，並使用各自的 package 聲明。

> AI 標籤重要性:
> 為了實現 AI 驅動的自動化腳手架和程式碼生成，本文件中的介面定義將包含特定的 AI 標籤（例如 AI_PLUGIN_TYPE, AI_IMPL_PACKAGE, AI_IMPL_CONSTRUCTOR）。
> 這些標籤是 AI 理解介面意圖、預期實現路徑和構造函數的強制性指令。
> AI 將嚴格依據這些標籤來生成符合平台規範的程式碼骨架和組裝邏輯。
> 開發者在新增或修改介面時，必須同時維護這些 AI 標籤，以確保 AI 輔助開發流程的順暢與正確性。
> 詳細的 AI 標籤規範和腳手架工作流程，請參考 docs/ai_scaffold/scaffold_workflow.md。

## 進度清單

### entities (5)

- [x] 1.User
- [x] 2.Detector
- [x] 3.AnalysisResult
- [x] 4.Detection
- [x] 5.DetectionResult

### interfaces (7)

- [x] 1.UserRepository
- [x] 2.DetectorRepository
- [ ] 3.AnalysisResultRepository
- [x] 4.AnalysisEngine
- [ ] 5.UserService
- [ ] 6.DetectorService
- [ ] 7.AnalysisService

### plugins (8)

- [x] 1.Plugin
- [x] 2.Importer
- [ ] 3.DetectorPlugin
- [ ] 4.AnalysisEnginePlugin
- [ ] 5.NotificationPlugin
- [ ] 6.AlertPlugin
- [x] 7.UIPagePlugin
- [ ] 8.CLIPlugin

### contracts (27)

#### 🎛 Platform I/O Providers
- [x] 1.ConfigProvider
- [x] 2.HttpServerProvider
- [x] 3.CliServerProvider

#### 🔐 Security & Identity
- [x] 4.AuthProvider
- [x] 5.KeycloakClientContract
- [ ] 6.SessionStore
- [ ] 7.CSRFTokenProvider

#### 📊 Observability & Stability
- [x] 8.Logger
- [ ] 9.MetricsProvider
- [ ] 10.TracingProvider
- [ ] 11.RateLimiterProvider
- [ ] 12.CircuitBreakerProvider

#### 🔌 Plugin / Registry / Metadata
- [x] 13.PluginRegistryProvider
- [ ] 14.PluginMetadataProvider

#### 💾 Storage & State
- [x] 15.DBClientProvider
- [x] 16.MigrationRunner
- [ ] 17.TransactionManager
- [ ] 18.CacheProvider
- [ ] 19.SecretsProvider

#### 📡 Event & Comms
- [ ] 20.EventBusProvider
- [ ] 21.AuditLogProvider

#### 🤖 AI / ML
- [ ] 22.LLMProvider
- [ ] 23.EmbeddingStoreProvider

#### 🔧 Platform Utility
- [ ] 24.MiddlewarePlugin
- [ ] 25.ErrorFactory
- [ ] 26.ServiceDiscoveryProvider
- [ ] 27.ServiceInstance


## --- 領域實體 (pkg/domain/entities) ---
> 定義領域內具有唯一標識和生命週期的核心業務物件。
> 對應目錄：`pkg/domain/entities/`

1. User 是 Detectviz 平台的核心用戶實體。
```go
// 定義領域內具有唯一標識和生命週期的核心業務物件。
// 檔案位置: pkg/domain/entities/user.go
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
```

2. Detector 是 Detectviz 平台的核心偵測器實體。
```go
// pkg/domain/entities/detector.go
// Detector 是 Detectviz 平台的核心偵測器實體。
// 職責: 封裝偵測器的配置、狀態及與偵測器相關的業務行為 (例如啟用/禁用偵測)。
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

3. AnalysisResult 封裝了偵測器執行後的分析結果。
```go
// pkg/domain/entities/analysis_result.go
// AnalysisResult 是一個表示數據分析結果的領域值物件。
// 職責: 捕獲並封裝分析過程產生的不可變結果數據。
type AnalysisResult struct{} // 佔位符類型，實際會包含詳細的分析數據結構
```

4. Detection 是表示一個特定偵測事件的領域實體。
```go
// pkg/domain/entities/detection.go
// Detection 是表示一個特定偵測事件的領域實體。
// 職責: 封裝偵測事件的上下文，例如觸發時間、來源數據等。
type Detection struct{} // 佔位符類型，實際會包含詳細的偵測事件數據結構
```

5. DetectionResult 是表示一個偵測事件處理後的最終結果的領域值物件。
```go
// pkg/domain/entities/detection_result.go
// DetectionResult 是表示一個偵測事件處理後的最終結果的領域值物件。
// 職責: 封裝偵測事件被處理後的輸出。
type DetectionResult struct{} // 佔位符類型，實際會包含詳細的偵測結果數據結構
```

## --- 抽象介面 (pkg/domain/interfaces) ---
> 定義領域業務邏輯的抽象操作介面，不依賴具體實現技術。
> 對應目錄：`pkg/domain/interfaces/`

1. UserRepository 定義了用戶數據的持久化操作介面。
```go
// pkg/domain/interfaces/user_repository.go
// UserRepository 定義了用戶數據的持久化操作介面。
// 職責: 提供用戶實體的 CRUD (創建、讀取、更新、刪除) 操作抽象。
// AI_PLUGIN_TYPE: "user_repository"
// AI_IMPL_PACKAGE: "detectviz-platform/internal/infrastructure/repositories/mysql_user_repository"
// AI_IMPL_CONSTRUCTOR: "NewMySQLUserRepository"
type UserRepository interface {
	Save(ctx context.Context, user *User) error
	FindByID(ctx context.Context, id string) (*User, error)
	FindByEmail(ctx context.Context, email string) (*User, error)
	Update(ctx context.Context, user *User) error
	Delete(ctx context.Context, id string) error
	GetName() string
}
```

2. DetectorRepository 定義了偵測器數據的持久化操作介面。
```go
// pkg/domain/interfaces/detector_repository.go
// DetectorRepository 定義了偵測器數據的持久化操作介面。
// 職責: 提供偵測器實體的 CRUD 操作抽象。
// AI_PLUGIN_TYPE: "detector_repository"
// AI_IMPL_PACKAGE: "detectviz-platform/internal/infrastructure/repositories/mysql_detector_repository"
// AI_IMPL_CONSTRUCTOR: "NewMySQLDetectorRepository"
type DetectorRepository interface {
	Save(ctx context.Context, detector *Detector) error
	FindByID(ctx context.Context, id string) (*Detector, error)
	FindAll(ctx context.Context) ([]*Detector, error)
	Update(ctx context.Context, detector *Detector) error
	Delete(ctx context.Context, id string) error
	GetName() string
}
```

3. AnalysisResultRepository 定義了分析結果數據的持久化操作介面。
```go
// pkg/domain/interfaces/analysis_result_repository.go
// AnalysisResultRepository 定義了分析結果數據的持久化操作介面。
// 職責: 提供分析結果實體的 CRUD 操作抽象。
// AI_PLUGIN_TYPE: "analysis_result_repository"
// AI_IMPL_PACKAGE: "detectviz-platform/internal/infrastructure/repositories/mysql_analysis_result_repository"
// AI_IMPL_CONSTRUCTOR: "NewMySQLAnalysisResultRepository"
type AnalysisResultRepository interface {
	Save(ctx context.Context, result *AnalysisResult) error
	FindByID(ctx context.Context, id string) (*AnalysisResult, error)
	FindByDetectorID(ctx context.Context, detectorID string) ([]*AnalysisResult, error)
	Update(ctx context.Context, result *AnalysisResult) error
	Delete(ctx context.Context, id string) error
	GetName() string
}
```

4. AnalysisEngine 定義了核心數據分析功能的介面。
```go
// pkg/domain/interfaces/analysis_engine.go
// AnalysisEngine 定義了核心數據分析功能的介面 (領域服務介面)。
// 職責: 執行複雜的數據分析演算法，不關心數據的來源或輸出格式。
// AI_PLUGIN_TYPE: "analysis_engine"
// AI_IMPL_PACKAGE: "detectviz-platform/internal/domain_logic/services/analysis_engine"
// AI_IMPL_CONSTRUCTOR: "NewAnalysisEngine"
type AnalysisEngine interface {
	AnalyzeData(ctx context.Context, data []byte) (entities.AnalysisResult, error)                         // 分析原始數據
	ProcessDetection(ctx context.Context, detection *entities.Detection) (entities.DetectionResult, error) // 處理偵測事件
}
```

5. UserService 定義了用戶相關的業務邏輯介面。
```go
// pkg/domain/interfaces/user_service.go
// UserService 定義了用戶相關的業務邏輯介面。
// 職責: 協調 UserRepository 和其他領域服務，處理用戶註冊、登入、資料更新等業務流程。
// AI_PLUGIN_TYPE: "user_service"
// AI_IMPL_PACKAGE: "detectviz-platform/internal/domain_logic/services/user_service"
// AI_IMPL_CONSTRUCTOR: "NewUserService"
// @See: internal/domain_logic/services/user_service/user_service.go
type UserService interface {
	RegisterUser(ctx context.Context, name, email, password string) (*User, error)
	AuthenticateUser(ctx context.Context, email, password string) (*User, error)
	GetUserByID(ctx context.Context, id string) (*User, error)
	UpdateUserProfile(ctx context.Context, id string, updates map[string]interface{}) (*User, error)
	DeleteUser(ctx context.Context, id string) error
	GetName() string
}
```

6. DetectorService 定義了偵測器相關的業務邏輯介面。
```go
// pkg/domain/interfaces/detector_service.go
// DetectorService 定義了偵測器相關的業務邏輯介面。
// 職責: 協調 DetectorRepository 和 DetectorPlugin，管理偵測器的生命週期和執行。
// AI_PLUGIN_TYPE: "detector_service"
// AI_IMPL_PACKAGE: "detectviz-platform/internal/domain_logic/services/detector_service"
// AI_IMPL_CONSTRUCTOR: "NewDetectorService"
// @See: internal/domain_logic/services/detector_service/detector_service.go
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

7. AnalysisService 定義了分析結果相關的業務邏輯介面。
```go
// pkg/domain/interfaces/analysis_service.go
// AnalysisService 定義了分析結果相關的業務邏輯介面。
// 職責: 協調 AnalysisResultRepository 和 AnalysisEnginePlugin，處理分析結果的查詢、確認和歸檔。
// AI_PLUGIN_TYPE: "analysis_service"
// AI_IMPL_PACKAGE: "detectviz-platform/internal/domain_logic/services/analysis_service"
// AI_IMPL_CONSTRUCTOR: "NewAnalysisService"
type AnalysisService interface {
	GetAnalysisResult(ctx context.Context, id string) (*AnalysisResult, error)
	ListAnalysisResultsByDetector(ctx context.Context, detectorID string) ([]*AnalysisResult, error)
	AcknowledgeResult(ctx context.Context, id, userID string) (*AnalysisResult, error)
	GetName() string
}
```

## --- 具體實現插件 (pkg/domain/plugins) ---
> 定義可插拔的領域功能介面，支援平台的擴展性和模組化。
> 對應目錄：`pkg/domain/plugins/`

1. Plugin 是所有 Detectviz 平台插件的基礎介面
```go
// 檔案位置: pkg/domain/plugins/
// pkg/domain/plugins/plugin.go
// Plugin 是所有 Detectviz 平台插件的基礎介面。
// 職責: 提供插件的通用方法，如獲取插件名稱。
type Plugin interface {
	GetName() string                                            // 返回插件的唯一名稱
	Init(ctx context.Context, cfg map[string]interface{}) error // 插件初始化，接收配置
	Start(ctx context.Context) error                            // 插件啟動，例如啟動背景任務
	Stop(ctx context.Context) error                             // 插件停止，清理資源
}
```

2. Importer 定義了數據導入插件的介面
```go
// pkg/domain/plugins/importer.go
// Importer 定義了數據導入功能的通用介面。
// 職責: 從不同來源（文件、API、數據庫）導入數據到平台。
// AI_PLUGIN_TYPE: "importer_plugin"
// AI_IMPL_PACKAGE: "detectviz-platform/internal/domain_logic/plugins/importer/csv_importer"
// AI_IMPL_CONSTRUCTOR: "NewCSVImporterPlugin"
type Importer interface {
	Plugin                                               // 繼承通用 Plugin 介面
	ImportData(ctx context.Context, source string) error // 根據來源導入數據
}
```

3. DetectorPlugin 定義了具體偵測器實現的介面
```go
// pkg/domain/plugins/detector.go
// DetectorPlugin 定義了具體偵測器實現的介面。
// 職責: 執行特定類型的數據偵測邏輯。
// AI_PLUGIN_TYPE: "detector_plugin"
// AI_IMPL_PACKAGE: "detectviz-platform/internal/domain_logic/plugins/detector/anomaly_detector"
// AI_IMPL_CONSTRUCTOR: "NewAnomalyDetectorPlugin"
type DetectorPlugin interface {
  Plugin
	Execute(ctx context.Context, data map[string]interface{}, detectorConfig map[string]interface{}) (*AnalysisResult, error)
}
```

4. AnalysisEnginePlugin 定義了數據分析引擎插件的介面
```go
// pkg/domain/plugins/analysis_engine.go
// AnalysisEnginePlugin 定義了數據分析引擎插件的介面。
// 職責: 對偵測結果進行深度分析和歸因。
// AI_PLUGIN_TYPE: "analysis_engine_plugin"
// AI_IMPL_PACKAGE: "detectviz-platform/internal/domain_logic/plugins/analysis_engine/llm_analysis_engine"
// AI_IMPL_CONSTRUCTOR: "NewLLMAnalysisEnginePlugin"
type AnalysisEnginePlugin interface {
  Plugin
	Analyze(ctx context.Context, result *AnalysisResult, analysisConfig map[string]interface{}) (map[string]interface{}, error)
}
```

5. NotificationPlugin 定義了通知發送插件的介面
```go
// pkg/domain/plugins/notification.go
// NotificationPlugin 定義了通知發送插件的介面。
// 職責: 負責通過不同渠道（如郵件、簡訊）發送通知。
// AI_PLUGIN_TYPE: "notification_plugin"
// AI_IMPL_PACKAGE: "detectviz-platform/internal/domain_logic/plugins/notification/email_notifier"
// AI_IMPL_CONSTRUCTOR: "NewEmailNotifierPlugin"
type NotificationPlugin interface {
  Plugin
	SendNotification(ctx context.Context, recipient, subject, body string, metadata map[string]interface{}) error
}
```

6. AlertPlugin 定義了告警觸發插件的介面
```go
// pkg/domain/plugins/alert.go
// AlertPlugin 定義了告警觸發插件的介面。
// 職責: 將偵測到的異常轉換為告警，並集成到告警系統。
// AI_PLUGIN_TYPE: "alert_plugin"
// AI_IMPL_PACKAGE: "detectviz-platform/internal/domain_logic/plugins/alert/slack_alerter"
// AI_IMPL_CONSTRUCTOR: "NewSlackAlerterPlugin"
type AlertPlugin interface {
  Plugin
	TriggerAlert(ctx context.Context, result *AnalysisResult, alertConfig map[string]interface{}) error
}
```

7. UIPagePlugin 定義了動態 UI 頁面插件的介面
```go
// pkg/domain/plugins/ui_page.go
// UIPagePlugin 定義了動態 UI 頁面插件的介面。
// 職責: 允許插件註冊新的前端頁面或組件，擴展平台 UI。
// AI_PLUGIN_TYPE: "ui_page_plugin"
// AI_IMPL_PACKAGE: "detectviz-platform/internal/domain_logic/plugins/ui_page/dashboard_page"
// AI_IMPL_CONSTRUCTOR: "NewDashboardPagePlugin"
type UIPagePlugin interface {
	Plugin
	GetRoutePath() string
	GetTemplateName() string
	GetData(ctx context.Context, params map[string]string) (map[string]interface{}, error)
}
```

8. CLIPlugin 定義了命令行界面擴展插件的介面
```go
// pkg/domain/plugins/cli.go
// CLIPlugin 定義了命令行界面擴展插件的介面。
// 職責: 允許插件向平台的 CLI 工具註冊新的命令。
// AI_PLUGIN_TYPE: "cli_plugin"
// AI_IMPL_PACKAGE: "detectviz-platform/internal/domain_logic/plugins/cli/detector_cli"
// AI_IMPL_CONSTRUCTOR: "NewDetectorCLIPlugin"
type CLIPlugin interface {
	Plugin
	GetCommandName() string
	GetDescription() string
	Execute(ctx context.Context, args []string) (string, error)
}
```

## --- 平台契約層 (pkg/platform/contracts) ---
> 定義 Detectviz 平台級基礎設施服務的抽象介面，這些介面是平台核心功能與其具體實現之間的契約。
> 對應目錄：`pkg/platform/contracts/`

### 🎛 Platform I/O Providers

1. ConfigProvider 定義了配置管理服務的介面
```go
// pkg/platform/contracts/contracts.go
// ConfigProvider 定義了平台統一的設定載入和存取介面。
// 職責: 支援讀取不同類型的配置值，並可將配置反序列化到結構體。
// AI_PLUGIN_TYPE: "config_provider"
// AI_IMPL_PACKAGE: "detectviz-platform/internal/infrastructure/platform/config/viper_config_provider"
// AI_IMPL_CONSTRUCTOR: "NewViperConfigProvider"
// @See: internal/infrastructure/platform/config/viper_config_provider.go
type ConfigProvider interface {
	GetString(key string) string
	GetInt(key string) int
	GetBool(key string) bool
	Unmarshal(rawVal interface{}) error // 將整個配置結構反序列化到 Go struct
	GetName() string
}
```

2. HttpServerProvider 定義了 HTTP 服務的介面
```go
// pkg/platform/contracts/contracts.go
// HttpServerProvider 定義了 HTTP 伺服器啟動和路由註冊的能力。
// 職責: 作為平台 Web 入口，處理 HTTP 請求並分發到 Handler。
// AI_PLUGIN_TYPE: "http_server_provider"
// AI_IMPL_PACKAGE: "detectviz-platform/internal/infrastructure/platform/http_server/echo_http_server_provider"
// AI_IMPL_CONSTRUCTOR: "NewEchoHttpServerProvider"
// @See: internal/infrastructure/platform/http_server/echo_http_server_provider.go
type HttpServerProvider interface {
	Start(port string) error        // 啟動 HTTP 服務
	Stop(ctx context.Context) error // 停止 HTTP 服務
	GetRouter() *echo.Echo          // 獲取底層路由實例，用於註冊路由和中介層 (這裡耦合 Echo，可考慮使用更通用的介面)
	GetName() string
}
```

3. CliServerProvider 定義了 CLI 服務的介面
```go
// pkg/platform/contracts/contracts.go
// CliServerProvider 定義了 CLI 服務啟動和命令註冊的能力。
// 職責: 作為平台命令行入口，處理命令解析和執行。
// AI_PLUGIN_TYPE: "cli_server_provider"
// AI_IMPL_PACKAGE: "detectviz-platform/internal/infrastructure/platform/cli_server/cobra_cli_server_provider"
// AI_IMPL_CONSTRUCTOR: "NewCobraCliServerProvider"
// @See: internal/infrastructure/platform/cli_server/cobra_cli_server_provider.go
type CliServerProvider interface {
	Execute() error
	AddCommand(cmd *cobra.Command) // 添加命令到 CLI 應用
	GetName() string
}
```

### 🔐 Security & Identity

4. AuthProvider 定義了身份驗證服務的介面
```go
// pkg/platform/contracts/contracts.go
// AuthProvider 定義了身份驗證與授權服務的通用介面。
// 職責: 負責驗證用戶身份並提供基礎授權判斷。
// AI_PLUGIN_TYPE: "keycloak_auth_provider"
// AI_IMPL_PACKAGE: "detectviz-platform/internal/infrastructure/platform/auth/keycloak_auth_provider"
// AI_IMPL_CONSTRUCTOR: "NewKeycloakAuthProvider"
// @See: internal/infrastructure/platform/auth/keycloak_auth_provider.go
type AuthProvider interface {
	Authenticate(ctx context.Context, credentials string) (userID string, err error)
	Authorize(ctx context.Context, userID string, resource string, action string) (bool, error)
	GetName() string
}
```

5. KeycloakClientContract 定義了與 Keycloak 外部服務互動的介面
```go
// pkg/platform/contracts/contracts.go
// KeycloakClientContract 定義了與 Keycloak 外部服務互動的抽象介面。
// 職責: 封裝與 Keycloak 服務進行底層 HTTP/gRPC 通訊的細節。
// AI_PLUGIN_TYPE: "keycloak_client_contract"
// AI_IMPL_PACKAGE: "detectviz-platform/internal/infrastructure/platform/external_services/keycloak_client"
// AI_IMPL_CONSTRUCTOR: "NewKeycloakClient"
// @See: internal/infrastructure/platform/external_services/keycloak_client.go
type KeycloakClientContract interface {
	VerifyToken(ctx context.Context, token string) (string, error)
	CheckPermissions(ctx context.Context, userID, resource, action string) (bool, error)
}
```

### 📊 Observability & Stability

6. Logger 定義了日誌記錄服務的介面
```go
// pkg/platform/contracts/logger.go
// Logger 定義了日誌服務的通用介面。
// 職責: 提供統一的日誌記錄功能，便於調試、監控和問題追蹤。
// AI_PLUGIN_TYPE: "logger_provider"
// AI_IMPL_PACKAGE: "detectviz-platform/internal/infrastructure/platform/logger/otelzap_logger"
// AI_IMPL_CONSTRUCTOR: "NewOtelZapLogger"
// @See: internal/infrastructure/platform/logger/otelzap_logger.go
type Logger interface {
	Debug(msg string, fields ...interface{})
	Info(msg string, fields ...interface{})
	Warn(msg string, fields ...interface{})
	Error(msg string, fields ...interface{})
	Fatal(msg string, fields ...interface{}) // Fatal 會導致程式終止
	WithFields(fields ...interface{}) Logger // 返回一個帶有附加字段的新 Logger 實例。
	WithContext(ctx interface{}) Logger      // 返回一個帶有上下文的新 Logger 實例。
	GetName() string
}
```

10. PluginMetadataProvider 定義了插件元資訊的查詢與註冊介面
```go
// pkg/platform/contracts/plugin_metadata.go
// PluginMetadataProvider 定義了插件元資訊的查詢與註冊介面。
// 職責: 提供插件名稱、版本、依賴等資訊，利於平台治理。
// AI_PLUGIN_TYPE: "plugin_metadata_provider"
// AI_IMPL_PACKAGE: "detectviz-platform/internal/platform/providers/plugin_metadata/in_memory_plugin_metadata"
// AI_IMPL_CONSTRUCTOR: "NewInMemoryPluginMetadataProvider"
// @See: internal/platform/providers/plugin_metadata/in_memory_plugin_metadata.go
type PluginMetadataProvider interface {
	GetMetadata(ctx context.Context, pluginName string) (map[string]any, error)
	RegisterMetadata(ctx context.Context, pluginName string, metadata map[string]any) error
	GetName() string
}
```

### 💾 Storage & State

11. DBClientProvider 定義了資料庫客戶端連接的介面
```go
// pkg/platform/contracts/contracts.go
// DBClientProvider 定義了資料庫連線能力的通用介面。
// 職責: 負責提供與特定資料庫類型（如 MySQL, PostgreSQL）的連線。
// AI_PLUGIN_TYPE: "gorm_mysql_client_provider"
// AI_IMPL_PACKAGE: "detectviz-platform/internal/infrastructure/database/gorm_mysql_client"
// AI_IMPL_CONSTRUCTOR: "NewGORMMySQLClientProvider"
// @See: internal/infrastructure/database/gorm_mysql_client.go
type DBClientProvider interface {
	GetDB(ctx context.Context) (*sql.DB, error) // 獲取底層 *sql.DB 連線實例
	GetName() string
}
```

12. MigrationRunner 定義了資料庫 Schema 遷移的通用介面
```go
// pkg/platform/contracts/contracts.go
// MigrationRunner 定義了資料庫 Schema 遷移的通用介面。
// 職責: 管理資料庫結構的版本化控制，確保應用程式與數據庫兼容。
// AI 擴展點: AI 可生成 `AtlasMigrationRunner` 或 `GoMigrateRunner`。
// AI_PLUGIN_TYPE: "migration_runner_provider"
// AI_IMPL_PACKAGE: "detectviz-platform/internal/infrastructure/database/atlas_migration_runner"
// AI_IMPL_CONSTRUCTOR: "NewAtlasMigrationRunner"
// @See: internal/infrastructure/database/atlas_migration_runner.go
type MigrationRunner interface {
	RunMigrations(ctx context.Context, db *sql.DB) error // 執行 Schema 遷移
	GetName() string
}
```

13. TransactionManager 定義了事務管理服務的介面
```go
// pkg/platform/contracts/transaction_manager.go
// TransactionManager 定義了事務管理服務的介面。
// 職責: 提供數據庫事務的開始、提交和回滾功能。
// AI_PLUGIN_TYPE: "transaction_manager_provider"
// AI_IMPL_PACKAGE: "detectviz-platform/internal/infrastructure/database/gorm_transaction_manager"
// AI_IMPL_CONSTRUCTOR: "NewGORMTransactionManager"
// @See: internal/infrastructure/database/gorm_transaction_manager.go
type TransactionManager interface {
	BeginTx(ctx context.Context, opts *interface{}) (interface{}, error) // 返回一個事務上下文，例如 *gorm.DB 或 *sql.Tx
	CommitTx(tx interface{}) error
	RollbackTx(tx interface{}) error
	GetName() string
}
```

14. CacheProvider 定義了緩存服務的介面
```go
// pkg/platform/contracts/cache.go
// CacheProvider 定義了緩存服務的介面。
// 職責: 提供鍵值對緩存操作，支持設置過期時間。
// AI_PLUGIN_TYPE: "cache_provider"
// AI_IMPL_PACKAGE: "detectviz-platform/internal/platform/providers/cache/redis_cache"
// AI_IMPL_CONSTRUCTOR: "NewRedisCacheProvider"
// @See: internal/platform/providers/cache/redis_cache.go
type CacheProvider interface {
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error
	Get(ctx context.Context, key string) (interface{}, error)
	Delete(ctx context.Context, key string) error
	GetName() string
}
```

15. SecretsProvider 定義了秘密管理服務的介面
```go
// pkg/platform/contracts/secrets.go
// SecretsProvider 定義了秘密管理服務的介面。
// 職責: 安全地讀取和管理敏感資訊 (如 API 金鑰、數據庫憑證)。
// AI_PLUGIN_TYPE: "secrets_provider"
// AI_IMPL_PACKAGE: "detectviz-platform/internal/platform/providers/secrets/env_secrets"
// AI_IMPL_CONSTRUCTOR: "NewEnvSecretsProvider"
// @See: internal/platform/providers/secrets/env_secrets.go
type SecretsProvider interface {
	GetSecret(ctx context.Context, key string) (string, error)
	GetName() string
}
```

### 📊 Observability & Stability

16. MetricsProvider 定義了指標收集與導出的介面
```go
// pkg/platform/contracts/metrics_provider.go
// MetricsProvider 定義了指標收集與導出的介面。
// 職責: 提供應用程式運行時指標的記錄功能。
// AI_PLUGIN_TYPE: "metrics_provider"
// AI_IMPL_PACKAGE: "detectviz-platform/internal/platform/providers/metrics/otel_metrics"
// AI_IMPL_CONSTRUCTOR: "NewOtelMetricsProvider"
// @See: internal/platform/providers/metrics/otel_metrics.go
type MetricsProvider interface {
	IncCounter(name string, tags map[string]string)
	ObserveHistogram(name string, value float64, tags map[string]string)
	SetGauge(name string, value float64, tags map[string]string)
	GetName() string
}
```

17. TracingProvider 定義了分佈式追蹤的介面
```go
// pkg/platform/contracts/tracing_provider.go
// TracingProvider 定義了分佈式追蹤的介面。
// 職責: 提供 Span 的創建、管理和上下文傳播功能。
// AI_PLUGIN_TYPE: "tracing_provider"
// AI_IMPL_PACKAGE: "detectviz-platform/internal/platform/providers/tracing/otel_tracing"
// AI_IMPL_CONSTRUCTOR: "NewOtelTracingProvider"
// @See: internal/platform/providers/tracing/otel_tracing.go
type TracingProvider interface {
	StartSpan(ctx context.Context, name string, opts ...interface{}) (context.Context, interface{}) // 返回新的上下文和 Span
	EndSpan(span interface{})
	GetName() string
}
```

18. RateLimiterProvider 定義了速率限制服務的介面
```go
// pkg/platform/contracts/rate_limiter.go
// RateLimiterProvider 定義了速率限制服務的介面。
// 職責: 控制請求流量，防止服務過載。
// AI_PLUGIN_TYPE: "rate_limiter_provider"
// AI_IMPL_PACKAGE: "detectviz-platform/internal/platform/providers/rate_limiter/uber_rate_limiter"
// AI_IMPL_CONSTRUCTOR: "NewUberRateLimiterProvider"
// @See: internal/platform/providers/rate_limiter/uber_rate_limiter.go
type RateLimiterProvider interface {
	Allow(ctx context.Context, key string) bool
	GetName() string
}
```

19. CircuitBreakerProvider 定義了熔斷器服務的介面
```go
// pkg/platform/contracts/circuit_breaker.go
// CircuitBreakerProvider 定義了熔斷器服務的介面。
// 職責: 在外部服務失敗時，快速失敗並提供降級處理，防止級聯故障。
// AI_PLUGIN_TYPE: "circuit_breaker_provider"
// AI_IMPL_PACKAGE: "detectviz-platform/internal/platform/providers/circuit_breaker/hystrix_breaker"
// AI_IMPL_CONSTRUCTOR: "NewHystrixCircuitBreakerProvider"
// @See: internal/platform/providers/circuit_breaker/hystrix_breaker.go
type CircuitBreakerProvider interface {
	Execute(ctx context.Context, name string, run func() error, fallback func(error) error) error
	GetName() string
}
```

### 🔐 Security & Identity

4. AuthProvider 定義了身份驗證服務的介面
```go
// pkg/platform/contracts/contracts.go
// AuthProvider 定義了身份驗證與授權服務的通用介面。
// 職責: 負責驗證用戶身份並提供基礎授權判斷。
// AI_PLUGIN_TYPE: "keycloak_auth_provider"
// AI_IMPL_PACKAGE: "detectviz-platform/internal/infrastructure/platform/auth/keycloak_auth_provider"
// AI_IMPL_CONSTRUCTOR: "NewKeycloakAuthProvider"
// @See: internal/infrastructure/platform/auth/keycloak_auth_provider.go
type AuthProvider interface {
	Authenticate(ctx context.Context, credentials string) (userID string, err error)
	Authorize(ctx context.Context, userID string, resource string, action string) (bool, error)
	GetName() string
}
```

5. KeycloakClientContract 定義了與 Keycloak 外部服務互動的介面
```go
// pkg/platform/contracts/contracts.go
// KeycloakClientContract 定義了與 Keycloak 外部服務互動的抽象介面。
// 職責: 封裝與 Keycloak 服務進行底層 HTTP/gRPC 通訊的細節。
// AI_PLUGIN_TYPE: "keycloak_client_contract"
// AI_IMPL_PACKAGE: "detectviz-platform/internal/infrastructure/platform/external_services/keycloak_client"
// AI_IMPL_CONSTRUCTOR: "NewKeycloakClient"
// @See: internal/infrastructure/platform/external_services/keycloak_client.go
type KeycloakClientContract interface {
	VerifyToken(ctx context.Context, token string) (string, error)
	CheckPermissions(ctx context.Context, userID, resource, action string) (bool, error)
}
```

6. SessionStore 定義了使用者登入狀態與會話的儲存抽象
```go
// pkg/platform/contracts/session_store.go
// SessionStore 定義了使用者登入狀態與會話的儲存抽象。
// 職責: 管理登入 Session 的生命週期與屬性。
// AI_PLUGIN_TYPE: "session_store_provider"
// AI_IMPL_PACKAGE: "detectviz-platform/internal/platform/providers/session_store/redis_session_store"
// AI_IMPL_CONSTRUCTOR: "NewRedisSessionStoreProvider"
// @See: internal/platform/providers/session_store/redis_session_store.go
type SessionStore interface {
	Set(ctx context.Context, sessionID string, data map[string]any) error
	Get(ctx context.Context, sessionID string) (map[string]any, error)
	Delete(ctx context.Context, sessionID string) error
	GetName() string
}
```

7. CSRFTokenProvider 定義了 CSRF Token 管理的介面
```go
// pkg/platform/contracts/csrf_token_provider.go
// CSRFTokenProvider 定義了 CSRF Token 管理的介面。
// 職責: 生成、驗證和管理用於防範 CSRF 攻擊的 Token。
// AI_PLUGIN_TYPE: "csrf_token_provider"
// AI_IMPL_PACKAGE: "detectviz-platform/internal/platform/providers/csrf_token/default_csrf_token"
// AI_IMPL_CONSTRUCTOR: "NewDefaultCSRFTokenProvider"
// @See: internal/platform/providers/csrf_token/default_csrf_token.go
type CSRFTokenProvider interface {
	GenerateToken(ctx context.Context) (string, error)
	ValidateToken(ctx context.Context, token string) error
	GetName() string
}
```

### 🔌 Plugin / Registry / Metadata

13. PluginRegistryProvider 定義了插件註冊與查詢的介面
```go
// pkg/platform/contracts/contracts.go
// PluginRegistryProvider 定義了平台 plugin 的註冊與 metadata 查詢能力。
// 職責: 管理已載入和可用的插件，提供插件查詢和元數據獲取功能。
// AI_PLUGIN_TYPE: "plugin_registry_provider"
// AI_IMPL_PACKAGE: "detectviz-platform/internal/infrastructure/platform/registry/plugin_registry_provider"
// AI_IMPL_CONSTRUCTOR: "NewPluginRegistryProvider"
// @See: internal/infrastructure/platform/registry/plugin_registry_provider.go
type PluginRegistryProvider interface {
	Register(name string, provider any) error        // 註冊一個具名的插件實例
	Get(name string) (any, error)                    // 獲取指定名稱的插件實例
	List() []string                                  // 列出所有已註冊插件的名稱
	GetMetadata(name string) (map[string]any, error) // 回傳特定插件的描述資訊（版本、作者、狀態等）
	GetName() string                                 // 例如 "core_registry"
}
```

14. PluginMetadataProvider 定義了插件元資訊的查詢與註冊介面
```go
// pkg/platform/contracts/plugin_metadata.go
// PluginMetadataProvider 定義了插件元資訊的查詢與註冊介面。
// 職責: 提供插件名稱、版本、依賴等資訊，利於平台治理。
// AI_PLUGIN_TYPE: "plugin_metadata_provider"
// AI_IMPL_PACKAGE: "detectviz-platform/internal/platform/providers/plugin_metadata/in_memory_plugin_metadata"
// AI_IMPL_CONSTRUCTOR: "NewInMemoryPluginMetadataProvider"
// @See: internal/platform/providers/plugin_metadata/in_memory_plugin_metadata.go
type PluginMetadataProvider interface {
	GetMetadata(ctx context.Context, pluginName string) (map[string]any, error)
	RegisterMetadata(ctx context.Context, pluginName string, metadata map[string]any) error
	GetName() string
}
```

### 📡 Event & Comms

20. EventBusProvider 定義了事件總線服務的介面
```go
// pkg/platform/contracts/event_bus.go
// EventBusProvider 定義了事件總線服務的介面。
// 職責: 提供異步事件的發布和訂閱機制。
// AI_PLUGIN_TYPE: "event_bus_provider"
// AI_IMPL_PACKAGE: "detectviz-platform/internal/platform/providers/event_bus/nats_event_bus"
// AI_IMPL_CONSTRUCTOR: "NewNATSEventBusProvider"
// @See: internal/platform/providers/event_bus/nats_event_bus.go
type EventBusProvider interface {
	Publish(ctx context.Context, topic string, event interface{}) error
	Subscribe(ctx context.Context, topic string, handler func(event interface{})) error
	GetName() string
}
```

21. AuditLogProvider 定義了審計記錄的儲存與查詢功能
```go
// pkg/platform/contracts/audit_log.go
// AuditLogProvider 定義了審計記錄的儲存與查詢功能。
// 職責: 記錄關鍵操作、身份與時間資訊，支援合規需求。
// AI_PLUGIN_TYPE: "audit_log_provider"
// AI_IMPL_PACKAGE: "detectviz-platform/internal/platform/providers/audit_log/db_audit_log"
// AI_IMPL_CONSTRUCTOR: "NewDBAuditLogProvider"
// @See: internal/platform/providers/audit_log/db_audit_log.go
type AuditLogProvider interface {
	LogAction(ctx context.Context, userID, action, resource string, metadata map[string]any) error
	GetName() string
}
```

### 🤖 AI / ML

22. LLMProvider 定義了大型語言模型推論功能的通用介面
```go
// pkg/platform/contracts/llm_provider.go
// LLMProvider 定義了大型語言模型推論功能的通用介面。
// 職責: 將 prompt 傳入 LLM 並取得模型輸出。
// AI_PLUGIN_TYPE: "llm_provider"
// AI_IMPL_PACKAGE: "detectviz-platform/internal/platform/providers/llm/gemini_llm"
// AI_IMPL_CONSTRUCTOR: "NewGeminiLLMProvider"
// @See: internal/platform/providers/llm/gemini_llm.go
type LLMProvider interface {
	GenerateText(ctx context.Context, prompt string, options map[string]any) (string, error)
	GetName() string
}
```

23. EmbeddingStoreProvider 定義了向量嵌入儲存與查詢功能的介面
```go
// pkg/platform/contracts/embedding_store.go
// EmbeddingStoreProvider 定義了向量嵌入儲存與查詢功能的介面。
// 職責: 儲存和檢索高維向量，支持相似性搜索。
// AI_PLUGIN_TYPE: "embedding_store_provider"
// AI_IMPL_PACKAGE: "detectviz-platform/internal/platform/providers/embedding_store/chroma_embedding_store"
// AI_IMPL_CONSTRUCTOR: "NewChromaEmbeddingStoreProvider"
// @See: internal/platform/providers/embedding_store/chroma_embedding_store.go
type EmbeddingStoreProvider interface {
	StoreEmbedding(ctx context.Context, id string, vector []float32, metadata map[string]any) error
	QueryNearest(ctx context.Context, queryVector []float32, topK int, filter map[string]any) ([]string, error) // 返回最相似的 ID
	GetName() string
}
```

### 🔧 Platform Utility

24. MiddlewarePlugin 定義了 HTTP 中介層插件的介面
```go
// pkg/platform/contracts/middleware.go
// MiddlewarePlugin 定義了 HTTP 中介層插件的介面。
// 職責: 在 HTTP 請求處理鏈中插入通用邏輯 (如日誌、認證、CORS)。
// AI_PLUGIN_TYPE: "middleware_plugin"
// AI_IMPL_PACKAGE: "detectviz-platform/internal/platform/middleware/auth_middleware"
// AI_IMPL_CONSTRUCTOR: "NewAuthMiddlewarePlugin"
// @See: internal/platform/middleware/auth_middleware.go
type MiddlewarePlugin interface {
	Handle(next http.Handler) http.Handler
	GetName() string
}
```

25. ErrorFactory 定義了錯誤創建和標準化的介面
```go
// pkg/platform/contracts/error_factory.go
// ErrorFactory 定義了錯誤創建和標準化的介面。
// 職責: 提供統一的錯誤創建機制，包含錯誤碼和可讀訊息。
// AI_PLUGIN_TYPE: "error_factory_provider"
// AI_IMPL_PACKAGE: "detectviz-platform/internal/platform/providers/error_factory/standard_error_factory"
// AI_IMPL_CONSTRUCTOR: "NewStandardErrorFactory"
// @See: internal/platform/providers/error_factory/standard_error_factory.go
type ErrorFactory interface {
	NewBadRequestError(message string, details ...map[string]any) error
	NewNotFoundError(message string, details ...map[string]any) error
	NewUnauthorizedError(message string, details ...map[string]any) error
	NewInternalServerError(message string, details ...map[string]any) error
	NewErrorf(format string, args ...any) error // 類似 fmt.Errorf 但返回標準錯誤類型
	GetName() string
}
```

26. ServiceDiscoveryProvider 定義了服務發現的介面
```go
// pkg/platform/contracts/service_discovery.go
// ServiceDiscoveryProvider 定義了服務發現的介面。
// 職責: 註冊、註銷服務實例，並查詢可用服務實例的地址。
// AI_PLUGIN_TYPE: "service_discovery_provider"
// AI_IMPL_PACKAGE: "detectviz-platform/internal/platform/providers/service_discovery/k8s_discovery"
// AI_IMPL_CONSTRUCTOR: "NewK8sServiceDiscoveryProvider"
// @See: internal/platform/providers/service_discovery/k8s_discovery.go
type ServiceDiscoveryProvider interface {
	RegisterService(ctx context.Context, serviceName string, instanceID string, address string, port int, metadata map[string]string) error
	DeregisterService(ctx context.Context, serviceName string, instanceID string) error
	GetInstances(ctx context.Context, serviceName string) ([]ServiceInstance, error)
	GetName() string
}
```

27. ServiceInstance 定義了服務實例的結構
```go
// pkg/platform/contracts/types.go
// ServiceInstance 定義了服務實例的結構。
// 職責: 封裝服務的基本資訊 (名稱、地址、端口、健康狀態等)。
// @See: pkg/platform/contracts/types.go
type ServiceInstance struct {
	ID       string            `json:"id"`
	Name     string            `json:"name"`
	Address  string            `json:"address"`
	Port     int               `json:"port"`
	Metadata map[string]string `json:"metadata"`
	Healthy  bool              `json:"healthy"`
}
```

## 進度統計

**總計完成進度：9/47 項目 (19%)**

- **entities**: 5/5 完成 (100%)
- **interfaces**: 4/7 完成 (57%)  
- **plugins**: 3/8 完成 (38%)
- **contracts**: 9/27 完成 (33%)
