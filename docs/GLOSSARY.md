# **GLOSSARY.md - 術語與概念詞彙表**

本詞彙表定義了 Detectviz 平台專案中常用的術語、概念和縮寫。理解這些術語對於參與專案開發、閱讀文檔以及與團隊成員溝通至關重要。

### **A**

* **AI-Driven Development (AIDD)**：**AI 驅動開發**。一種開發範式，其中人工智慧工具和模型被深度整合到軟體開發生命週期的各個階段，從需求分析、程式碼生成、測試到部署和運維。Detectviz 平台旨在實現高度 AIDD。  
* **AI Directives**：**AI 指令**。在 Detectviz 平台中，特指在 Go 介面定義（如 interface_spec.md）中嵌入的特殊註解標籤（例如 AI_PLUGIN_TYPE），用於向 AI 提供關於介面預期用途和實現細節的明確指示，以提高程式碼生成的精確性。  
* **Analysis Engine**：**分析引擎**。一個抽象介面，用於處理偵測器生成的結果，可能涉及數據聚合、歸因分析或與 LLM 互動以提供解釋。  
* **API Gateway**：**API 閘道**。位於客戶端和後端服務之間的一個單一入口點，負責請求路由、協議轉換、認證、授權、速率限制等。在 Detectviz 中，它可能是 HttpServerProvider 的一個實現。  
* **Application Layer**：**應用程式層**。Clean Architecture 中的一層，包含應用程式特有的業務邏輯和用例（Use Cases），協調領域層實體和基礎設施層的交互。

### **C**

* **CacheProvider**：**快取提供者**。一個平台合約介面，定義了快取操作的抽象，用於提高數據存取性能。  
* **CircuitBreakerProvider**：**熔斷器提供者**。一個平台合約介面，定義了熔斷器模式的抽象，用於在分散式系統中防止級聯故障。  
* **Clean Architecture**：**整潔架構**。一種軟體設計原則，強調將業務邏輯與實現細節（如資料庫、UI、框架）分離，通過依賴反轉原則實現高內聚、低耦合、易於測試和維護的系統。Detectviz 平台的核心架構遵循此原則。  
* **CLI**：**Command Line Interface**。命令列介面。Detectviz 平台提供 CLI 工具用於管理和交互。  
* **CNCF**：**Cloud Native Computing Foundation**。雲原生計算基金會。一個開源軟體基金會，推動雲原生技術的發展和採用。Detectviz 平台深度內化了 CNCF 的理念和專案。  
* **Composition Root**：**組裝根**。應用程式啟動時，所有依賴關係被解決和注入的地方。在 Detectviz 中，主要體現在 cmd/api/main.go 和 internal/app/initializer/platform_initializer.go。  
* **ConfigProvider**：**配置提供者**。一個平台合約介面，定義了配置管理的能力，支持從多來源載入和管理配置。  
* **Conventional Commits**：**常規提交**。一種提交訊息規範，透過標準化的前綴（如 feat:, fix:, docs:）來分類提交的類型，有助於自動化版本發布和生成變更日誌。  
* **Contracts**：**合約**。在 Detectviz 平台中，通常指 pkg/platform/contracts 目錄下的 Go 介面，它們定義了平台級服務的抽象功能。

### **D**

* **DBClientProvider**：**資料庫客戶端提供者**。一個平台合約介面，定義了獲取底層資料庫連接的能力。  
* **Detector**：**偵測器**。Detectviz 平台的核心領域實體，封裝了偵測器的配置、狀態及相關業務行為（例如啟用/禁用偵測）。  
* **DetectorPlugin**：**偵測器插件**。實現 pkg/domain/plugins/DetectorPlugin 介面的具體插件，負責執行特定的異常或模式偵測邏輯。  
* **DetectorRepository**：**偵測器儲存庫**。一個領域層介面，定義了對 Detector 實體的數據庫操作抽象。  
* **DevX**：**開發者體驗 (Developer Experience)**。指開發者在使用工具、平台或 API 時的整體感受和效率。Detectviz 平台致力於提供卓越的 DevX。  
* **Dependency Inversion Principle (DIP)**：**依賴反轉原則**。Clean Architecture 的核心原則之一，高層模組不應依賴低層模組，兩者都應依賴抽象；抽象不應依賴細節，細節應依賴抽象。  
* **Domain Layer**：**領域層**。Clean Architecture 最內層，包含核心業務概念、實體和抽象介面，不依賴任何外部框架或技術細節。

### **E**

* **Echo**：一個輕量級、高性能的 Go Web 框架，可能被 Detectviz 用作 HttpServerProvider 的實現。  
* **EmbeddingStoreProvider**：**嵌入向量儲存提供者**。一個平台合約介面，定義了向量資料庫操作的抽象，用於儲存和檢索機器學習模型生成的嵌入向量。  
* **Entities**：**實體**。在領域層中，指具有唯一標識和生命週期的核心業務物件（例如 User, Detector）。  
* **EventBusProvider**：**事件總線提供者**。一個平台合約介面，定義了發布和訂閱事件的抽象，用於實現服務間的異步通訊。

### **F**

* **Factory Pattern**：**工廠模式**。一種設計模式，用於創建物件而無需指定確切的類別。在 Detectviz 中，插件通常透過工廠模式創建。  
* **Feature Flag**：**功能開關**。一種軟體開發技術，允許在不重新部署程式碼的情況下，動態地開啟或關閉特定功能。

### **G**

* **GORM**：一個 Go 語言的 ORM (Object-Relational Mapping) 框架，可能被 Detectviz 用作主要資料庫操作工具。  
* **gRPC**：**Google Remote Procedure Call**。一種高性能、開源的通用 RPC 框架，可能被 Detectviz 用於服務間通訊。

### **H**

* **Hexagonal Architecture**：**六邊形架構**。與 Clean Architecture 類似，強調將應用程式核心與外部世界（UI、資料庫、外部服務）隔離，透過埠（Ports）和轉接器（Adapters）進行交互。  
* **HttpServerProvider**：**HTTP 伺服器提供者**。一個平台合約介面，定義了 HTTP 服務器的啟動、路由註冊和請求處理能力。

### **I**

* **ImporterPlugin**：**導入器插件**。實現 pkg/domain/plugins/ImporterPlugin 介面的具體插件，負責從外部來源導入數據到平台。  
* **Infrastructure Layer**：**基礎設施層**。Clean Architecture 最外層，包含所有外部框架和技術細節的實現，如資料庫驅動、Web 框架實現、外部 API 整合等。  
* **Interface Adapters Layer**：**介面轉接層**。Clean Architecture 中的一層，負責將領域層和應用程式層的抽象介面轉換為具體的實現，例如 Web 控制器、資料庫儲存庫實現、外部服務客戶端。  
* **interface_spec.md**：Detectviz 平台中一個關鍵的文檔，集中定義了所有核心 Go 介面，並包含 AI 專用指令標籤，是 AI 理解平台契約的基礎。  
* **Isolation Forest**：一種機器學習演算法，常用於異常偵測。

### **J**

* **JSON Schema**：一種用於描述和驗證 JSON 數據結構的規範。Detectviz 平台廣泛使用 JSON Schema 來定義和驗證所有配置檔的結構和內容。

### **K**

* **Keycloak**：一個開源的身份和訪問管理解決方案，可能被 Detectviz 用作 AuthProvider 的實現。  
* **Kubernetes (K8s)**：一個開源的容器編排平台，用於自動化部署、擴展和管理容器化應用程式。Detectviz 平台可能部署在 Kubernetes 上。  
* **Kyverno**：一個 Kubernetes 原生策略引擎，用於驗證、修改和生成 Kubernetes 資源配置。

### **L**

* **LLM**：**大型語言模型 (Large Language Model)**。指能夠理解和生成人類語言的深度學習模型。Detectviz 平台將整合 LLM 進行智能分析和交互。  
* **LLMProvider**：**LLM 提供者**。一個平台合約介面，定義了大型語言模型推論功能的通用介面，用於將 prompt 傳入 LLM 並取得模型輸出。  
* **LoggerProvider**：**日誌提供者**。一個平台合約介面，定義了日誌記錄能力的抽象，支持結構化日誌和可配置的日誌級別。

### **M**

* **main.go assembly**：指 Detectviz 平台主程式 cmd/api/main.go 中負責組裝和初始化所有插件和服務的邏輯。AI 腳手架會自動生成這部分代碼。  
* **Message Queues**：**消息佇列**。一種用於服務間異步通訊的技術，例如 NATS。

### **N**

* **NATS**：一個高性能、輕量級的開源消息系統，可能被 Detectviz 用作消息佇列。

### **O**

* **OpenTelemetry**：一個開源的可觀測性框架，提供統一的 API、SDK 和工具來收集和匯出遙測數據（日誌、指標、追蹤）。Detectviz 平台整合 OpenTelemetry 實現可觀測性。  
* **ORM**：**物件關係映射 (Object-Relational Mapping)**。一種程式設計技術，用於在物件導向程式語言和關係型資料庫之間轉換數據。GORM 是一個 Go 語言的 ORM。  
* **OtelZapLoggerProvider**：一個基於 Zap 和 OpenTelemetry 的日誌提供者實現。

### **P**

* **Platform Engineering**：**平台工程**。一種學科，旨在構建和維護內部開發者平台，以提高開發效率、加速價值交付和降低認知負擔。  
* **Plugin**：**插件**。在 Detectviz 平台中，「一切皆插件」的核心理念體現。指可插拔、可替換的功能模組，透過實現特定的介面來擴展平台能力。  
* **PluginRegistryProvider**：**插件註冊中心提供者**。一個平台合約介面，定義了插件工廠的註冊和獲取能力，是平台動態組裝的關鍵。  
* **Policy as Code**：**策略即程式碼**。透過程式碼定義和管理安全、合規性和最佳實踐策略，通常與 Kubernetes 原生策略引擎結合使用。  
* **Ports and Adapters Architecture**：**埠與轉接器架構**。同 Hexagonal Architecture，強調將應用程式核心與外部世界隔離。  
* **Provider**：**供應商**。在 Detectviz 平台中，通常指實現 pkg/platform/contracts 下介面的具體服務，例如 HttpServerProvider 的實現。

### **R**

* **Repository**：**儲存庫**。一種設計模式，用於抽象數據持久化邏輯，將領域層與資料庫細節解耦（例如 UserRepository）。  
* **ROADMAP.md**：Detectviz 平台的發展路線圖，定義了階段性目標和 AI 賦能的進程。

### **S**

* **Scaffold**：**鷹架**。指自動生成程式碼骨架、文件結構或配置文件的過程或工具。在 Detectviz 中，AI 腳手架是核心的自動化能力。  
* **scaffold_workflow.md**：Detectviz 平台中一個關鍵的文檔，詳細定義了 AI 輔助程式碼生成的核心工作流程和指令。  
* **Service Discovery Provider**：**服務發現提供者**。一個平台合約介面，定義了服務註冊、註銷和實例查詢的能力，用於微服務架構中的服務間通訊。  
* **SessionStore**：**會話儲存**。一個平台合約介面，定義了使用者登入狀態和會話的儲存抽象。  
* **Software Items**：**軟體項目**。Detectviz 平台中對所有平台級功能模組的抽象稱謂，強調其可組合和可插拔的特性。

### **T**

* **TransactionManager**：**事務管理器**。一個平台合約介面，提供了事務管理的能力，確保多個資料庫操作的原子性。

### **U**

* **UIPagePlugin**：**UI 頁面插件**。實現 pkg/domain/plugins/UIPagePlugin 介面的具體插件，用於提供可自定義的 Web UI 頁面。  
* **User**：**用戶**。Detectviz 平台的核心用戶實體，封裝了用戶的基本資訊及與用戶身份相關的業務行為。  
* **UserRepository**：**用戶儲存庫**。一個領域層介面，定義了對 User 實體的數據庫操作抽象。

### **V**

* **Viper**：一個 Go 語言的配置管理庫，可能被 Detectviz 用於多來源配置讀取與合併。

### **Z**

* **Zap**：一個高性能的 Go 結構化日誌庫，可能被 Detectviz 用作日誌記錄庫。