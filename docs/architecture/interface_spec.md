# 介面定義

這是一個彙整 Detectviz 平台所有核心介面定義的程式碼區塊。

本文件的內容來源涵蓋整個 `/pkg` 目錄中所有與平台功能擴充、資料存取、業務邏輯、基礎設施介接相關的公開介面定義，包含但不限於：
- `pkg/domain/interfaces`（領域層與應用層）
- `pkg/domain/entities`, `pkg/domain/errors`, `pkg/domain/valueobjects`（實體與核心定義）
- `pkg/domain/interfaces/plugins`（可插拔插件介面）
- `pkg/platform/contracts`（平台契約與基礎設施抽象）
- 其他未來在 `/pkg` 目錄下新增的功能模組

介面來源目錄並無強制限制，重點在於其是否為平台核心功能或 AI 自動化腳手架所需的 API 定義。

> AI 標籤重要性:
> 為了實現 AI 驅動的自動化腳手架和程式碼生成，本文件中的介面定義將包含特定的 AI 標籤（例如 AI_PLUGIN_TYPE, AI_IMPL_PACKAGE, AI_IMPL_CONSTRUCTOR）。
> 這些標籤是 AI 理解介面意圖、預期實現路徑和構造函數的強制性指令。
> AI 將嚴格依據這些標籤來生成符合平台規範的程式碼骨架和組裝邏輯。
> 開發者在新增或修改介面時，必須同時維護這些 AI 標籤，以確保 AI 輔助開發流程的順暢與正確性。
> 詳細的 AI 標籤規範和腳手架工作流程，請參考 docs/ai_scaffold/scaffold_workflow.md。

## 清單

### 1. 領域實體 (entities)
1. [User](../../pkg/domain/entities/user.go)
2. [Detector](../../pkg/domain/entities/detector.go)
3. [AnalysisResult](../../pkg/domain/entities/analysis_result.go)
4. [Detection](../../pkg/domain/entities/detection.go)
5. [DetectionResult](../../pkg/domain/entities/detection_result.go) 

### 2. 領域值對象 (value objects)
1. [IDVO](../../pkg/domain/valueobjects/id.go)
2. [EmailVO](../../pkg/domain/valueobjects/email.go)

### 3. 領域錯誤 (domain errors)
1. [DomainError](../../pkg/domain/errors/errors.go)
2. [InfrastructureError](../../pkg/domain/errors/errors.go)

### 4. 領域介面 (interfaces)

**倉儲介面 (repositories)**
1. [UserRepository](../../pkg/domain/interfaces/repositories/user_repository.go)
2. [DetectorRepository](../../pkg/domain/interfaces/repositories/detector_repository.go)
3. [AnalysisResultRepository](../../pkg/domain/interfaces/repositories/analysis_result_repository.go)

**服務介面 (services)**
4. [UserService](../../pkg/domain/interfaces/services/user_service.go)
5. [DetectionService](../../pkg/domain/interfaces/services/detection_service.go)

**核心介面**
6. [AnalysisEngine](../../pkg/domain/interfaces/analysis_engine.go)

### 5. 插件介面 (plugins)
1. [Plugin](../../pkg/domain/interfaces/plugins/plugin.go)
2. [ImporterPlugin](../../pkg/domain/interfaces/plugins/importer.go)
3. [DetectorPlugin](../../pkg/domain/interfaces/plugins/detector.go)
4. [AnalysisPostProcessorPlugin](../../pkg/domain/interfaces/plugins/analysis_engine.go)
5. [NotificationPlugin](../../pkg/domain/interfaces/plugins/notification.go)
6. [AlertPlugin](../../pkg/domain/interfaces/plugins/alert.go)
7. [UIPagePlugin](../../pkg/domain/interfaces/plugins/uipage.go)
8. [CLIPlugin](../../pkg/domain/interfaces/plugins/cli.go)
9. [HealthCheckPlugin](../../pkg/domain/interfaces/plugins/health_check.go)
10. [HealthCheckCapablePlugin](../../pkg/domain/interfaces/plugins/health_check.go)
11. [MiddlewarePlugin](../../pkg/domain/interfaces/plugins/middleware.go)

**插件相關類型**
12. [HealthCheckResult](../../pkg/domain/interfaces/plugins/health_check.go)
13. [HealthStatus](../../pkg/domain/interfaces/plugins/health_check.go)

### 6. 平台契約 (contracts)

**核心基礎設施 (`config.go`, `logger.go`, `utility.go`)**
1. [ConfigProvider](../../pkg/platform/contracts/config.go)
2. [Logger](../../pkg/platform/contracts/logger.go)
3. [ErrorFactory](../../pkg/platform/contracts/utility.go)
4. [ServiceDiscoveryProvider](../../pkg/platform/contracts/utility.go)

**認證與授權 (`auth.go`)**
5. [AuthProvider](../../pkg/platform/contracts/auth.go) (整合後)
6. [AuthStorageProvider](../../pkg/platform/contracts/auth.go) (整合後)

**數據與存儲 (`database.go`, `storage.go`)**
7. [DBClientProvider](../../pkg/platform/contracts/database.go)
8. [MigrationRunner](../../pkg/platform/contracts/database.go)
9. [TransactionManager](../../pkg/platform/contracts/storage.go)
10. [CacheProvider](../../pkg/platform/contracts/storage.go)
11. [SecretsProvider](../../pkg/platform/contracts/storage.go)

**網絡與服務器 (`server.go`)**
12. [HttpServerProvider](../../pkg/platform/contracts/server.go)
13. [CliServerProvider](../../pkg/platform/contracts/server.go)

**可觀測性 (`metrics.go`, `observability.go`)**
14. [MetricsProvider](../../pkg/platform/contracts/metrics.go)
15. [TracingProvider](../../pkg/platform/contracts/metrics.go)
16. [Span](../../pkg/platform/contracts/metrics.go)
17. [RateLimiterProvider](../../pkg/platform/contracts/observability.go)
18. [CircuitBreakerProvider](../../pkg/platform/contracts/observability.go)

**插件與事件 (`plugin.go`, `events.go`)**
19. [PluginRegistryProvider](../../pkg/platform/contracts/plugin.go)
20. [PluginMetadataProvider](../../pkg/platform/contracts/plugin.go)
21. [EventBusProvider](../../pkg/platform/contracts/events.go)
22. [AuditLogProvider](../../pkg/platform/contracts/events.go)

**AI/ML (`llm_provider.go`, `embedding_store.go`)**
23. [LLMProvider](../../pkg/platform/contracts/llm_provider.go)
24. [EmbeddingStoreProvider](../../pkg/platform/contracts/embedding_store.go)

**類型定義 (`types.go`)**
25. [ServiceInstance](../../pkg/platform/contracts/types.go)

## 統計

- **領域實體**: 5 個
- **領域值對象**: 2 個
- **領域錯誤**: 2 個 (簡化後)
- **領域介面**: 6 個 (3 個倉儲 + 2 個服務 + 1 個核心)
- **插件介面**: 13 個 (11 個介面 + 2 個相關類型)
- **平台契約**: 25 個 (重新分類後)
- **總計**: 53 個定義 (減少 6 個)

## Clean Architecture 分層說明

### 領域層 (Domain Layer)
- **實體 (Entities)**: 核心業務邏輯和規則
- **值對象 (Value Objects)**: 不可變的值類型
- **領域介面**: 定義業務操作的抽象

### 應用層 (Application Layer)
- **服務介面**: 協調多個實體的業務流程
- **倉儲介面**: 數據持久化的抽象

### 介面層 (Interface Layer)
- **插件介面**: 可擴展功能的抽象定義

### 基礎設施層 (Infrastructure Layer)
- **平台契約**: 外部依賴和技術實現的抽象

這種分層結構確保了：
1. **依賴反轉**: 高層模組不依賴低層模組
2. **關注點分離**: 每個介面都有明確的職責
3. **可測試性**: 所有依賴都可以被模擬
4. **可擴展性**: 通過插件系統支持功能擴展
5. **AI 友好**: 完整的 AI 標籤支持自動化腳手架生成

## 完整性驗證

以下是 `/pkg` 目錄下所有定義的完整清單，確保沒有遺漏：

### pkg/domain/entities/ (5 個實體)
✅ User, Detector, AnalysisResult, Detection, DetectionResult

### pkg/domain/valueobjects/ (2 個值對象)
✅ IDVO, EmailVO

### pkg/domain/errors/ (2 個錯誤類型)
✅ DomainError, InfrastructureError

### pkg/domain/interfaces/ (6 個核心介面)
✅ 3 個倉儲介面 + 2 個服務介面 + 1 個分析引擎介面

### pkg/domain/interfaces/plugins/ (13 個插件定義)
✅ 11 個插件介面 + 2 個相關類型 (HealthCheckResult, HealthStatus)

### pkg/platform/contracts/ (25 個平台契約)
✅ 所有基礎設施層的抽象介面和類型定義

**總計**: 53 個定義，涵蓋 `/pkg` 目錄下的所有公開 API 定義。

---

## Interface 規劃審查用途與檢查方式

本文件作為 Detectviz 平台唯一可信的 interface 清單，具備以下目標：

- 作為 Clean Architecture 合理性與結構設計的審查依據（例如是否過度擴展或分類模糊）
- 驗證 interface 是否出現重複定義、語意衝突或命名模糊
- 作為 AI scaffold 系統與 Cursor 編輯輔助的唯一參考來源
- 確保每個介面皆有對應實作與測試，並對應 plugin doc、schema、AI metadata

### 快速檢查提示語（Prompt）

> 請依照 `interface_spec.md` 中「Interface 規劃審查用途與檢查方式」段落，逐項檢查目前介面定義是否合理，並回報所有可優化項目，產出 interface_review_report.md。完成檢查後，請立即根據報告內容，逐條執行優化項目並產生對應 patch。

AI Agent 將自動依據以下 4 點進行檢查，並在 `/docs/audit/interface_review_report.md` 中輸出檢查報告，同時依據回報逐條執行修正：

1. **是否有重複定義**：
   - 職責重疊或命名過於相似的 interface

2. **是否過度設計**：
   - interface 過度切割、分類不清，導致管理困難

3. **plugin 命名與分類是否一致**：
   - 所有 plugin 是否皆以 Plugin 結尾？
   - 是否與 AI_PLUGIN_TYPE 對應一致？

4. **平台契約是否合理分層**：
   - 是否有 contracts interface 應移入 domain 層的情況？

建議將每次分析結果存放於 `docs/audit/interface_review_report.md`，並作為後續 Cursor Scaffold 的依據與自動 patch 執行觸發點。