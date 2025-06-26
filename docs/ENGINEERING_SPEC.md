# **Detectviz 平台工程實作規範**

本文件旨在提供 Detectviz 平台各個技術層面的具體實作規範和技術選型，確保開發過程的一致性、效率與品質。

## **1. 平台核心技術棧 - Platform Core Technology Stack - 插件驅動**

Detectviz 平台的核心技術棧基於 Go 語言構建，旨在提供一個高效、可擴展且易於維護的平台工程基礎。我們的設計哲學是 **"一切皆插件" (Everything is a Plugin)**，即便是核心服務也以供應商 (Provider) 插件的形式呈現，確保高度模組化與可替換性。

### **1.1 語言與框架**

* **Go (Golang)**: 平台的首選開發語言，用於後端服務與所有插件。  
* **HTTP 框架**: Echo (或其他輕量級框架)  
* **配置管理**: Viper (用於多來源配置讀取與合併)  
* **日誌**: Zap (或 OtelZap for OpenTelemetry integration)  
* **ORM**: GORM  
* **資料庫遷移**: Atlas  
* **依賴注入/服務組裝**: 自定義的 Plugin Registry 機制  
* **YAML**: 主要的配置檔案格式 (app_config.yaml, composition.yaml)。  
* **JSON Schema**: 用於定義和驗證所有配置檔案的結構與內容。這是實現 AI 自動化配置和確保配置正確性的核心工具。**平台在啟動時會自動執行 Schema 驗證，確保配置的一致性和正確性。**  
* **Go JSON Schema 驗證**: 使用 `github.com/xeipuuv/gojsonschema` 庫實現運行時配置驗證。

### **1.2 資料儲存**

* **主資料庫**: MySQL (或其他關聯式資料庫)  
* **緩存**: Redis

## **2. 專案結構與檔案命名規範 - Directory Structure & File Naming Conventions**

清晰的專案結構對於可維護性和協作至關重要，尤其在 AI 輔助開發中，標準化能提高 AI 理解與操作的準確性。

detectviz-platform/  
├── .github/                       # GitHub Actions CI/CD workflows (或 .gitlab-ci.yml for GitLab)  
├── cmd/                           # 應用程式主入口點  
│   └── main.go                    # Detectviz 平台的啟動與核心組裝邏輯  
├── configs/                       # 應用程式運行時讀取的配置檔案  
│   ├── app_config.yaml            # 核心應用程式全局配置  
│   └── composition.yaml           # 平台插件的組合與具體配置  
├── internal/                      # 僅供內部使用，不應被外部專案直接引用  
│   ├── domain_logic/              # 核心業務邏輯的內部實現  
│   │   └── plugins/               # 業務插件的具體實現  
│   │       └── ui_page/  
│   │           └── hello_world_ui_page.go  
│   ├── platform/                  # 平台層面的內部實現  
│   │   ├── plugin_registry/       # 插件註冊與組裝的核心邏輯  
│   │   │   └── registry.go  
│   │   └── providers/             # 平台核心供應商的具體實現  
│   │       ├── config/  
│   │       │   └── viper_config.go  
│   │       ├── http_server/  
│   │       │   └── echo_server.go  
│   │       └── logger/  
│   │           └── otelzap_logger.go  
│   └── repositories/              # 資料庫操作的具體實現 (例如 mysql_user_repository.go)  
│   └── services/                  # 應用層業務邏輯的實現 (例如 user_service.go)  
├── pkg/                           # 可供外部專案引用的公共程式碼  
│   ├── domain/                    # 領域層定義，不依賴任何外部框架  
│   │   ├── entities/              # 領域實體定義 (`user.go`, `detector.go`)  
│   │   ├── interfaces/            # 領域層介面定義 (`user_repository.go`, `analysis_engine.go`)  
│   │   └── plugins/               # 插件領域介面（平台無關）  
│   │       ├── plugin.go          # 所有插件的基礎介面  
│   │       └── ui_page.go         # UI 頁面插件的介面  
│   ├── platform/                  # 平台級契約 (Contracts) 和通用類型  
│   │   └── contracts/             # 平台核心供應商介面定義  
│   │       ├── config.go  
│   │       ├── http_server.go  
│   │       ├── logger.go  
│   │       ├── plugin.go  
│   │       └── plugin_registry.go  
│   └── plugins/                   # 各類插件的公共介面和工廠定義 (如果需要更細化的分類)  
├── schemas/                       # **JSON Schema 定義檔案，用於驗證 `/configs` 目錄下的 YAML 配置。**  
│   ├── app_config.json            # `app_config.yaml` 的 Schema  
│   ├── composition.json           # `composition.yaml` 的 Schema  
│   └── plugins/                   # **各個插件的獨立配置 Schema**  
│       ├── gorm_mysql_client_provider.json  
│       ├── http_server_provider.json
│       ├── keycloak_auth_provider.json 
│       ├── otelzap_logger_provider.json  
│       └── ... (其他插件的 Schema)  
├── docs/                          # **專案所有文檔的集中存放地。這對 AI 理解專案結構和生成文檔至關重要。**  
│   ├── ARCHITECTURE.md            # 平台整體架構、核心設計原則、模組劃分  
│   ├── ENGINEERING_SPEC.md        # 技術棧、開發與實作細節、程式碼規範 (即本文件)  
│   ├── ROADMAP.md                 # 專案發展路線圖  
│   ├── CONFIGURATION_REFERENCE.md # 平台配置的詳細說明與指南  
│   ├── ai_scaffold/               # **AI 驅動開發的專屬文檔和指導，供 AI 參考**  
│   │   ├── ai_directives_spec.md    # 定義 AI 專用標籤/註解規範 (e.g., AI_SCAFFOLD_HINT)  
│   │   ├── scaffold_workflow.md     # AI Scaffold 總體流程與決策邏輯  
│   │   └── scaffolding_blueprints/  # 針對不同 Scaffold 類型的高層次藍圖/設計模式  
│   │       └── new_plugin_blueprint.md  
│   ├── templates/                 # AI 生成程式碼時可能需要參考的代碼範本或結構  
│   │   └── ai_scaffolding/        # 供 AI Scaffold 使用的程式碼骨架模板  
│   │       └── main_go_assembly.tmpl  
│   │       └── plugin_skeleton.go  
│   └── README.md -> ../README.md  # (軟連結或複製) 專案總覽的入口文件  
├── migrations/                    # 資料庫遷移腳本  
├── test/                          # 測試相關檔案，如測試數據、測試工具  
└── README.md                      # 專案概述、願景、核心特色、快速上手指南 (專案的總覽入口)

**檔案命名**：

* Go 檔案名使用 snake_case (例如 user_service.go)。  
* 測試檔案以 _test.go 結尾。  
* 文檔檔案名：使用清晰的 snake_case 或 kebab-case 命名，並使用 .md 擴展名。  
* Schema 檔案名：使用 snake_case 或 kebab-case 命名，並使用 .json 擴展名。

## **3. 程式碼實作與風格規範 - Code Implementation & Style Guidelines**

* **Go Modules**: 統一使用 Go Modules 進行依賴管理。  
* **Clean Architecture 實踐**:  
  * 嚴格分層：領域層 (Domain) 不應依賴平台層 (Platform)，Platform 層不應依賴 Internal 實現。  
  * 依賴反轉原則：高層模組不應依賴低層模組，它們都應該依賴抽象。  
* **錯誤處理**：  
  * 統一的錯誤返回機制：通常使用 error 類型返回，並利用 errors.Wrap 等庫進行錯誤鏈追溯。  
  * 避免裸露的 panic。  
* **日誌記錄**：  
  * 使用結構化日誌 (Zap/OtelZap)。  
  * 日誌級別的合理使用 (debug, info, warn, error)。  
  * 配置日誌級別應通過 app_config.yaml 或 composition.yaml，並遵循 schemas/plugins/otelzap_logger_provider.json 的規範。  
* **Context 使用**：  
  * 所有涉及請求或長時間操作的函數都必須接受 context.Context 作為第一個參數，用於傳遞上下文資訊和取消信號。  
* **命名慣例**：  
  * 變數、函數、類型命名遵循 Go 語言慣例。  
  * 介面命名以 Provider 或 Service 結尾，如 LoggerProvider, UserService。  
  * 實現命名應包含具體技術棧，如 OtelZapLogger, MySQLUserRepository。  
* **程式碼註解 (GoDoc)**：所有公共函數、結構體、介面必須有詳細的 GoDoc 註解，包括其職責、參數、返回值、錯誤等。 這對於 AI 理解程式碼並生成相關內容 (包括自動生成 Schema) 至關重要。  
* **API 設計**：  
  * RESTful API 設計原則。  
  * 清晰的請求/回應結構。
* **配置驗證最佳實踐**：  
  * **Schema 先行**: 所有配置檔案必須先定義 JSON Schema，後實作程式碼。  
  * **自動驗證**: 平台在啟動時自動驗證所有配置的合法性。  
  * **詳細錯誤**: 配置驗證失敗時提供明確的錯誤訊息和修正建議。  
  * **向後兼容**: Schema 更新應保持向後兼容性，避免破壞既有配置。  
  * **文檔同步**: Schema 變更時必須同步更新 CONFIGURATION_REFERENCE.md 文檔。

## **4. 數據層細節 (Data Layer Details)**

### **4.1 ORM/Query Builder 選擇**

* **選擇**: 將採用 **GORM** 作為主要的 ORM (Object-Relational Mapping) 框架。  
* **考量**:  
  * GORM 提供友好的 API，可大幅提升開發效率。  
  * 支援多種數據庫，便於未來擴展或切換。  
  * 內建數據庫遷移和模型同步功能。  
* **使用規範**:  
  * 所有數據庫操作應優先通過 GORM 模型進行，盡量避免直接使用純 SQL。  
  * 對於複雜查詢或性能敏感的場景，允許使用 GORM 的原生 SQL 接口 (db.Raw(), db.Exec())。  
  * 定義 GORM 模型時，應遵循其標籤規範 (gorm:"column:...")。

### **4.2 數據遷移工具**

* **選擇**: 採用 **Atlas** (由 Ariga 開發) 作為資料庫 Schema 遷移工具。  
* **考量**:  
  * **聲明式 Schema 管理**: 允許開發者定義目標 Schema 狀態，由 Atlas 自動計算並生成最小化遷移腳本，降低人為錯誤。  
  * **數據庫不可知**: 支持多種關係型資料庫 (MySQL, PostgreSQL, SQLite, SQL Server 等)。  
  * **CI/CD 整合友善**: 提供命令行工具，便於自動化。  
  * **靜態分析能力**: 可以檢查 Schema 變更的潛在風險。  
* **使用規範**:  
  * 所有 Schema 變更都必須透過 Atlas 的聲明式方式來管理，並生成遷移文件。  
  * 遷移文件應遵循 Atlas 的命名和格式規範。  
  * 在應用啟動時，應配置 Atlas 運行未完成的遷移，或在 CI/CD 流程中自動執行。  
  * 開發者應利用 Atlas 的 diff 命令來預覽和審查 Schema 變更。

### **4.3 事務管理 (Transaction Management)**

* **策略**: 事務將在 **Service 層** 統一管理。  
* **實作細節**:  
  * Service 層的方法如果需要跨多個 Repository 操作來保證數據一致性，應通過 contracts.TransactionManager 介面開啟、提交或回滾事務。  
  * Repository 層的方法不應直接管理事務，而應接收一個 *sql.Tx 或 GORM *gorm.DB 實例，以便在 Service 層的單一事務中協同操作。  
* **範例**:  
  // service/user_service.go  
  type UserService struct {  
      userRepo    domain.UserRepository  
      detectorRepo domain.DetectorRepository  
      txManager   contracts.TransactionManager  
  }

  func (s *UserService) CreateUserAndDetector(ctx context.Context, user *domain.User, detector *domain.Detector) error {  
      tx, err := s.txManager.BeginTx(ctx, nil)  
      if err != nil {  
          return err  
      }  
      // 確保事務在函數返回時關閉  
      defer func() {  
          if r := recover(); r != nil {  
              s.txManager.RollbackTx(tx)  
              panic(r)  
          } else if err != nil {  
              s.txManager.RollbackTx(tx)  
          } else {  
              err = s.txManager.CommitTx(tx)  
          }  
      }()

      // 將事務上下文傳遞給 Repository  
      ctxWithTx := context.WithValue(ctx, "tx", tx)  
      if err = s.userRepo.Save(ctxWithTx, user); err != nil {  
          return err  
      }  
      if err = s.detectorRepo.Save(ctxWithTx, detector); err != nil {  
          return err  
      }  
      return nil  
  }

## **5. API 層與服務間通訊 (API Layer & Inter-Service Communication)**

### **5.1 跨服務通訊**

* **同步通訊**: 優先使用 **gRPC** 進行內部微服務間的高效同步通訊。  
  * **考量**: Protocol Buffers (Protobuf) 提供強類型契約、高效的序列化和反序列化、多語言支持。  
  * **使用規範**: 定義 .proto 文件來描述服務接口和消息結構。使用 protoc 生成 Go 服務和客戶端代碼。  
* **異步通訊**: 優先使用 **NATS** 作為消息隊列和事件流平台。  
  * **考量**: 輕量級、高性能、易於部署和使用，支持 Pub/Sub 和 Request/Reply 模式。  
  * **使用規範**: 通過 contracts.EventBusProvider 和 contracts.EventFactory 介面進行事件的發布和訂閱。

## **6. 緩存策略 (Caching Strategy)**

### **6.1 緩存技術選型**

* **選擇**: 採用 **Redis** 作為主要的分佈式緩存服務。  
* **考量**:  
  * 高性能、支持多種數據結構。  
  * 廣泛應用於高併發場景。  
  * 支持數據持久化和高可用部署。  
* **使用規範**:  
  * 通過 contracts.CacheProvider 介面進行緩存操作。  
  * 緩存的鍵命名應清晰，包含模塊名、業務實體類型和唯一識別符（例如：user:id:<userID>, detector:name:<detectorName>）。  
  * 設置合理的 TTL (Time-To-Live) 以避免髒數據和內存溢出。  
  * 實施緩存穿透、緩存擊穿和緩存雪崩的防範措施（例如：熱點數據永不過期、使用互斥鎖）。  
* **引入層次**: 緩存主要在 **Service 層** 引入，用於緩存經常訪問且數據變化不頻繁的讀取操作，以減少對數據庫的壓力。

## **7. 安全實踐深入 (Deeper Security Practices)**

### **7.1 速率限制 (Rate Limiting)**

* **選擇**: 使用 uber-go/ratelimit 庫進行應用內部的速率限制。  
* **考量**:  
  * 簡單易用，適用於單個應用實例的流量控制。  
  * 提供令牌桶算法實現。  
* **使用規範**:  
  * 針對公共 API 接口、登入接口等敏感或易受攻擊的端點實施速率限制。  
  * 通過 contracts.RateLimiter 介面統一管理。  
  * 對於分佈式環境下的全局速率限制，考慮引入 API Gateway 層的限流能力。

### **7.2 熔斷 (Circuit Breaker)**

* **選擇**: 使用 afex/hystrix-go 庫實現熔斷器模式。  
* **考量**:  
  * 提供熔斷、降級、超時、隔離等容錯機制。  
  * 有助於防止級聯故障。  
* **使用規範**:  
  * 所有對外部服務（包括其他微服務、第三方 API、數據庫等）的調用都應使用熔斷器包裝。  
  * 通過 contracts.CircuitBreakerProvider 介面統一管理熔斷操作。  
  * 設置合理的閾值（例如：請求失敗率、請求延遲）和回退邏輯 (fallback)。

### **7.3 輸入驗證**

* **選擇**: 採用 go-playground/validator 庫進行結構化數據的輸入驗證。  
* **考量**:  
  * 支持豐富的驗證規則和自定義驗證。  
  * 基於結構體標籤 (struct tags) 定義驗證規則，清晰易讀。  
* **使用規範**:  
  * 在 API 層 (Handler) 和 Service 層對所有傳入的請求數據進行嚴格驗證。  
  * 對於複雜的業務邏輯驗證，在 Service 層實現。  
  * 驗證失敗應返回標準化的錯誤響應，包含明確的錯誤訊息。

### **7.4 防禦性編程規範**

* **SQL 注入**:  
  * **永遠** 使用 GORM 的安全 API 或 Go database/sql 的預準備語句 (PreparedStatement) 和參數化查詢，**嚴禁** 直接拼接用戶輸入到 SQL 語句中。  
* **XSS (Cross-Site Scripting)**:  
  * 所有從用戶輸入或外部來源獲取的數據，在顯示到 Web 頁面時，必須進行適當的輸出編碼 (HTML Escaping)。  
  * Go 的 html/template 模版引擎會自動進行大部分編碼，但對於特定場景（如 JavaScript 注入到事件處理器），仍需額外注意。  
* **CSRF (Cross-Site Request Forgery)**:  
  * 利用 contracts.CSRFTokenProvider 介面在 Web 應用中實施 CSRF 防護。  
  * 對於所有會修改數據的 POST/PUT/DELETE 請求，要求在請求中包含 CSRF Token，並在服務端進行驗證。  
* **秘密管理 (Secret Management)**:  
  * 敏感資訊（數據庫憑證、API 金鑰等） **嚴禁** 硬編碼在程式碼中。  
  * 應通過 contracts.SecretsProvider 介面從環境變數、配置管理服務（如 Vault、Kubernetes Secrets）或安全配置檔案中讀取。  
  * 開發環境可以使用 .env 文件，生產環境必須使用更安全的秘密管理方案。  
* **錯誤處理**:  
  * 使用 contracts.ErrorFactory 統一錯誤創建和包裝，確保錯誤包含足夠的上下文訊息（錯誤碼、可讀訊息、原始錯誤等）。  
  * 錯誤應在底層詳細記錄，向上層傳播時進行適當的包裝或轉換，避免暴露敏感信息給終端用戶。

## **8. 消息隊列產品選型 (Message Queues)**

* **選擇**: 基於跨服務通訊的考量，消息隊列將主要採用 **NATS**。  
* **考量**:  
  * 高性能、低延遲的消息傳遞。  
  * 支持多種消息模式 (Pub/Sub, Request/Reply, Queue Groups)。  
  * 易於部署和擴展。  
  * JetStream 提供持久化和流處理能力，滿足事件驅動架構的需求。  
* **使用規範**:  
  * 所有異步任務和服務解耦場景都應通過 NATS 進行。  
  * 事件發布和訂閱應通過 contracts.EventBusProvider 介面抽象。  
  * 定義清晰的 Topic 命名規範，反映事件的業務含義和來源。  
  * 確保消息處理具備冪等性，以應對重複投遞。  
  * 考慮死信隊列 (Dead-Letter Queue) 機制處理無法成功處理的消息。

## **9. 可觀測性深入 (Deeper Observability)**

### **9.1 日誌收集與分析**

* **日誌庫**: 採用 **Zap** 作為日誌記錄庫。  
  * **考量**: 極致性能、結構化日誌、可配置的日誌級別。  
* **集中式日誌管理**: 日誌將收集到 **Grafana Loki**。  
  * **考量**: 輕量級、成本效益高、與 Grafana 深度整合。  
* **使用規範**:  
  * 所有日誌輸出應為結構化 JSON 格式，包含 timestamp, level, caller, message 等標準字段，並可添加業務相關的 fields。  
  * 通過 contracts.Logger 介面進行日誌操作。  
  * 生產環境日誌級別預設為 INFO，調試環境為 DEBUG。

### **9.2 指標可視化與警報**

* **指標庫**: 採用 **OpenTelemetry Metrics** 導出指標。  
* **指標收集**: 由 **Prometheus** 負責收集。  
* **可視化與警報**: 在 **Grafana** 中進行可視化和警報規則設定。  
* **使用規範**:  
  * 定義關鍵業務和系統指標（如：請求總量、錯誤率、響應時間、CPU/內存使用率）。  
  * 通過 contracts.MetricsProvider 介面記錄指標。  
  * 所有指標應具備統一的命名規範和標籤 (labels)，便於查詢和聚合。

### **9.3 追蹤可視化**

* **追蹤庫**: 採用 **OpenTelemetry Tracing** 導出追蹤數據。  
* **追蹤可視化**: 在 **Jaeger** 中進行追蹤數據的可視化和分析。  
* **使用規範**:  
  * 確保所有服務間調用、數據庫操作和外部 API 調用都能被追蹤。  
  * 通過 contracts.TracingProvider 介面管理 Span 的創建和結束。  
  * 重要的業務邏輯應添加自定義 Span，包含相關屬性，以便深入分析。

### **9.4 健康檢查與服務發現**

* **健康檢查端點**:  
  * 提供標準的 HTTP 健康檢查端點（例如 /health），返回服務狀態。  
  * 可以包含 Liveness Probe (服務是否存活) 和 Readiness Probe (服務是否準備好處理請求) 的邏輯。  
* **服務發現**:  
  * 在 Kubernetes 環境下，利用 **Kubernetes 內建的 Service Discovery** 機制（DNS）。  
  * 對於非 Kubernetes 或跨集群場景，可利用 contracts.ServiceDiscoveryProvider 介面實現對 Consul 或 Etcd 等服務發現工具的集成。  
* **使用規範**:  
  * 確保健康檢查響應快速，不包含敏感信息。  
  * Readiness Probe 應檢查所有關鍵依賴項（如數據庫、消息隊列）是否可用。

## **10. 插件開發規範詳表 - Detailed Plugin Development Specification Table**

Detectviz 平台的核心擴展機制是插件。本節詳細說明不同類型的插件如何開發與集成。

### **10.1 平台供應商 (Platform Core & Governance)**

這類插件負責提供平台的核心能力，通常在平台啟動時被組裝。

| 插件類型 (Type String) | 介面定義 (Go Interface) | 職責簡述 | 建議配置項與 Schema (參考 schemas/plugins/) |
| :---- | :---- | :---- | :---- |
| http_server_provider | pkg/platform/contracts/http_server.go:HttpServerProvider | 提供 HTTP 服務，處理路由和請求。 | port, readTimeout, writeTimeout (http_server_provider.json) |
| otelzap_logger_provider | pkg/platform/contracts/logger.go:LoggerProvider | 提供統一的日誌記錄功能。 | level, encoding, outputPaths, errorOutputPaths, initialFields (otelzap_logger_provider.json) |
| gorm_mysql_client_provider | pkg/platform/contracts/database.go:DBClientProvider | 提供資料庫連線池與操作介面。 | dsn, maxOpenConns, maxIdleConns, connMaxLifetime (gorm_mysql_client_provider.json) |
| keycloak_auth_provider | pkg/platform/contracts/auth.go:AuthProvider | 處理用戶身份驗證和 JWT 驗證。 | url, clientId, clientSecretEnvVar (keycloak_auth_provider.json) |
| config_provider | pkg/platform/contracts/config.go:ConfigProvider | 載入與管理應用程式配置。 | paths (配置檔案路徑) |
| plugin_registry_provider | pkg/platform/contracts/plugin_registry.go:PluginRegistry | 管理平台內所有插件的註冊與查詢。 | 無特定配置項 |
| audit_log_provider | pkg/platform/contracts/audit_log.go:AuditLogProvider | 記錄平台操作的審計日誌。 | storageType (e.g., db, file), retentionDays |
| secrets_provider | pkg/platform/contracts/secrets.go:SecretsProvider | 安全地管理敏感資訊。 | type (e.g., vault, env), backendConfig |
| rate_limiter_provider | pkg/platform/contracts/rate_limiter.go:RateLimiterProvider | 提供請求限流功能。 | algorithm (e.g., token_bucket), capacity, rate |
| plugin_metadata_provider | pkg/platform/contracts/plugin_metadata.go:PluginMetadataProvider | 管理插件元資料的儲存與查詢。 | 無特定配置項 |
| llm_provider | pkg/platform/contracts/llm_provider.go:LLMProvider | 提供 LLM 推理功能（如文本生成）。 | modelName, apiKeyEnvVar, temperature, maxTokens |
| embedding_store_provider | pkg/platform/contracts/embedding_store.go:EmbeddingStoreProvider | 提供向量嵌入的儲存與查詢功能。 | dbConnection, collectionName, vectorDimension |

### **10.2 核心業務插件 (Core Business Plugins)**

這類插件實現 Detectviz 的具體業務功能，通常會依賴於平台供應商。

| 插件類型 (Type String) | 介面定義 (Go Interface) | 職責簡述 | 建議配置項與 Schema (參考 schemas/plugins/) |
| :---- | :---- | :---- | :---- |
| importer_plugin | pkg/domain/plugins/importer.go:ImporterPlugin | 負責從外部數據源導入數據。 | sourceType (e.g., csv, api), sourceConfig (具體格式依 sourceType 而定) |
| detector_plugin | pkg/domain/plugins/detector.go:DetectorPlugin | 執行數據異常偵測或模式識別。 | model (e.g., isolation_forest, rnn), threshold, parameters (模型的具體參數) |
| analysis_engine_plugin | pkg/domain/plugins/analysis_engine.go:AnalysisEnginePlugin | 基於 LLM 或其他模型進行數據分析與歸因。 | llmProviderName, embeddingStoreProviderName, promptTemplate |
| user_service_plugin | pkg/domain/interfaces/user_service.go:UserService | 管理用戶帳戶和相關業務邏輯。 | defaultRole (新用戶預設角色), passwordPolicy (密碼複雜度規則) |
| notification_plugin | pkg/domain/plugins/notification.go:NotificationPlugin | 處理系統通知（郵件、簡訊等）。 | provider (e.g., smtp, sms), config (郵件伺服器地址、簡訊服務商 API Key 等) |
| alert_plugin | pkg/domain/plugins/alert.go:AlertPlugin | 觸發告警並集成告警系統。 | severityMapping, target (e.g., slack_webhook, opsgenie_api) |
| ui_page_plugin | pkg/domain/plugins/ui_page.go:UIPagePlugin | 動態註冊新的前端 UI 頁面或組件。 | routePath, templateName, jsBundlePath (對應前端資源路徑) |
| middleware_plugin | pkg/platform/contracts/middleware.go:MiddlewarePlugin | 提供 HTTP 請求中介層邏輯（如 CORS, 認證）。 | priority, handlers (要應用此中介層的路由) |
| cli_plugin | pkg/domain/plugins/cli.go:CLIPlugin | 擴展命令行界面功能。 | commandName, description, arguments (命令的參數定義) |

### **10.3 插件配置的實作與驗證**

所有插件的配置都必須在 composition.yaml 中定義，並通過 **JSON Schema** (schemas/plugins/{plugin_type}.json) 進行嚴格驗證。

* **Go 程式碼中的 Config 結構體**：  
  * 每個插件實現的 Config 結構體應清晰定義所有可配置的參數。  
  * 必須使用 yaml:"fieldName" 標籤來指定 YAML 欄位名。  
  * 強烈建議為每個配置欄位提供 GoDoc 註解，詳細說明其目的、類型、預設值、合法範圍等。這些註解是未來 AI 自動生成 Schema 或文檔的基礎。

// Example: OtelZapLoggerConfig defines the configuration for the OtelZap logger provider.  
type OtelZapLoggerConfig struct {  
    // Level specifies the minimum log level to record (debug, info, warn, error, dpanic, panic, fatal).  
    // Default: "info"  
    Level string `yaml:"level"`  
    // Encoding specifies the log output format ("json" or "console").  
    // Default: "json"  
    Encoding string `yaml:"encoding"`  
    // OutputPaths specifies where logs should be written (e.g., ["stdout", "/var/log/app.log"]).  
    OutputPaths []string `yaml:"outputPaths"`  
    // ErrorOutputPaths specifies where error logs should be written.  
    ErrorOutputPaths []string `yaml:"errorOutputPaths"`  
    // InitialFields specifies initial fields to be attached to all log entries.  
    InitialFields map[string]interface{} `yaml:"initialFields"`  
}

* **JSON Schema (規範性文檔)**：  
  * 每個插件的 Config 結構體應有一個對應的 JSON Schema 檔案放置在 schemas/plugins/ 目錄下。  
  * Schema 應精確定義配置欄位的資料類型 (type), 是否必填 (required), 預設值 (default), 列舉值 (enum), 格式 (pattern), 範圍 (minimum/maximum) 等。  
  * **AI 驅動**: 這些 Schema 是 AI 生成和驗證配置的核心規則。AI 將利用這些 Schema 來確保生成的 composition.yaml 片段是有效且符合規範的。  
* **配置驗證流程**:  
  * 在應用程式啟動時，應載入 app_config.yaml 和 composition.yaml。  
  * 必須在配置載入後，使用 Schema 驗證這些配置檔案的結構和內容。任何不符合 Schema 的配置應導致啟動失敗並提供清晰的錯誤訊息。  
  * CI/CD 流程中應包含配置 Schema 驗證步驟，確保提交的配置檔案始終有效。

## **11. 插件註冊與生命週期 - Plugin Registration & Lifecycle - AI 輔助優化**

本節描述插件如何被 Detectviz 平台發現、載入、初始化和管理。

### **11.1 插件發現與註冊機制**

* **工廠模式 (Factory Pattern)**: 每個插件類型都應該有一個對應的工廠函數，用於創建插件實例。  
* **集中註冊**: 在 main.go 或專門的 init 模組中，將所有可用的插件工廠註冊到 PluginRegistry。  
* **AI 輔助**: AI 可以協助生成新的插件工廠骨架，並自動在 main.go 中添加註冊邏輯。

### **11.2 插件生命週期**

* **初始化 (Initialization)**: 平台在啟動時，根據 composition.yaml 的指示，透過其工廠創建並初始化插件實例。  
* **啟動 (Start)**: 插件可能需要實現 Start() 方法來啟動其內部服務（例如 HTTP 伺服器、消費者）。  
* **停止 (Stop)**: 插件可能需要實現 Stop() 方法來優雅地關閉資源、停止服務。

### **11.3 依賴注入 (Dependency Injection)**

* 插件之間或插件與核心服務之間的依賴應透過介面進行，並在組裝階段進行注入。  
* PluginRegistry 應能解析並提供必要的依賴給插件。  
* **AI 輔助**: AI 可以分析插件的介面依賴，並建議如何在 main.go 的組裝邏輯中正確注入這些依賴。

## **12. 測試規範 - Testing Specification**

### **12.1 單元測試 (Unit Tests)**

* 針對 Go 函數和方法編寫，覆蓋獨立的邏輯單元。  
* 遵循 Go 測試慣例 (_test.go 檔案，TestXxx 函數)。  
* **AI 輔助**: AI 可以根據程式碼和 GoDoc 註解，自動生成單元測試的骨架和初步測試用例。

### **12.2 集成測試 (Integration Tests)**

* 測試多個模組或服務協同工作的場景 (例如，HTTP 請求流經多個中介層)。  
* 可使用 Testcontainers 進行外部依賴 (資料庫、消息佇列) 的測試。

### **12.3 端到端測試 (End-to-End Tests)**

* 模擬真實用戶場景，驗證整個系統的功能。  
* 應包含對平台 API 和 UI 的測試。

## **13. 文檔標準與註解 - Doc Strings & Comments**

* **清晰、一致**的文檔是協作和 AI 驅動開發的基石。  
* **GoDoc**: 所有公共 (大寫開頭) 的 package, type, func, var 必須包含 GoDoc 註解。  
  * 應清晰說明其目的、行為、參數、返回值和任何潛在的副作用或錯誤。  
* **AI 友好註解**: 積極使用如 AI_SCAFFOLD_HINT, AI_TEMPLATE_PATH 等自定義標籤，為 AI 提供額外的指導信息。  
* **內部註解**: 對於複雜的邏輯、演算法或非顯而易見的設計決策，應在程式碼內部添加行級或區塊註解。  
* **Markdown 文件**: 所有的 *.md 文件都應該結構清晰、易於閱讀，並提供足夠的細節。

## **14. MVP 階段的實作策略 - MVP Plugin Focus & Implementation Strategy**

在 MVP 階段，我們將聚焦於構建核心平台骨架和關鍵插件，以驗證基本功能。

* **核心組裝流程**: 確保 main.go 能夠正確載入配置，組裝平台供應商，並初始化插件註冊表。  
* **基礎設施供應商優先**: 優先實現 HttpServerProvider, LoggerProvider, DBClientProvider, ConfigProvider。  
* **首個業務插件**: 實現一個簡單的 ImporterPlugin 或 UIPagePlugin，以驗證插件機制端到端的工作流程。  
* **AI 優先任務**: 讓 AI Scaffold 在此階段主要負責生成介面骨架、Config 結構體以及對應的 GoDoc 註解，並依據 Schema 生成配置檔案的初始內容。

## **15. 總結 - Conclusion**

本工程規範旨在為 Detectviz 平台提供一套全面的開發準則。透過嚴格遵循這些規範，並結合 AI 驅動開發的強大能力，我們將能夠構建一個高品質、可擴展且易於維護的平台，加速產品的迭代與創新。

