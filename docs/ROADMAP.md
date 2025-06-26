**Detectviz 平台發展路線圖**  
基於對 CNCF 平台工程核心概念的深度內化，Detectviz 平台的基礎規劃已然穩健清晰。本文件旨在將這份全面的藍圖，轉化為具體的 **可實踐路線圖**。我們將採用 **迭代、增量** 的方式，分階段進行規劃與實踐，並持續將 **產品思維 (Product Thinking)** 與 **AI 驅動開發 (AI-Driven Development, AIDD)** 貫穿其中，確保每個階段都能交付實際價值、收集用戶反饋，並最大化 AI 在開發、審查、測試與部署環節的效能。

**核心原則：迭代交付，價值先行，AI 協同**

* **最小可行平台 (MVP) 思維**：每個階段都將以交付最小可行功能集為目標，快速驗證假設，避免過度工程。  
* **用戶反饋驅動**：持續與產品團隊互動，將用戶需求和痛點作為功能優先級排序的依據。  
* **核心依賴優先**：優先實現其他插件所依賴的核心平台能力。  
* **自動化為本**：從一開始就將 CI/CD 和自動化部署納入考量。  
* **AI 深度協同**：在任務規劃、程式碼生成、配置管理 (特別是依賴 JSON Schema 進行推斷和驗證)、程式碼審查、測試與部署等環節深度整合 AI 角色，逐步提高自動化比例，實現從高人工干預到 AI 驅動的開發流程轉型。

### **Detectviz 平台分階段實踐路線圖**

Phase 0: 平台核心骨架與 AI 基礎建設 ✅ **已完成 85%**  
核心目標：建立一個符合 Detectviz 平台架構文檔 和 Detectviz 平台工程規範 的可運行、可配置、可觀察的平台骨架，能夠組裝和運行最基礎的插件。同時，為 AI 驅動開發奠定堅實的基礎。

* **里程碑 0.1：基礎 Go 專案結構與核心組件介面**  
  * [x] 建立標準的 Go Modules 專案結構 (/cmd, /internal, /pkg, /configs, /docs 等)。  
  * [x] 定義 pkg/platform/contracts 下的核心平台供應商介面 (LoggerProvider, ConfigProvider, HttpServerProvider, DBClientProvider, PluginRegistryProvider 等)。  
  * [x] 定義 pkg/domain/plugins 下的通用插件介面 (Plugin, Importer 等)。  
  * **AI 協同點**：AI 協助生成介面定義和 GoDoc 註解。  
* **里程碑 0.2：配置管理與 JSON Schema 導入** ✅  
  * [x] 實現 ConfigProvider，支持從 app_config.yaml 和 composition.yaml 載入配置。  
  * [x] 為 app_config.yaml 和 composition.yaml 定義完整的 JSON Schema (schemas/app_config.json, schemas/composition.json)。  
  * [x] 為核心且穩定的平台供應商 (如 http_server_provider, otelzap_logger_provider, gorm_mysql_client_provider, keycloak_auth_provider) 撰寫獨立的 JSON Schema (schemas/plugins/*.json)。  
  * [x] 在 [CONFIGURATION_REFERENCE.md](./docs/CONFIGURATION_REFERENCE.md) 和 [ENGINEERING_SPEC.md](./docs/ENGINEERING_SPEC.md) 中更新配置規範，明確引用 JSON Schema。  
  * [x] 實作配置載入後的 JSON Schema 驗證機制，確保配置的合法性。  
  * **AI 協同點**：AI 依賴現有 Go 介面和文檔，推斷並生成 JSON Schema；未來 AI 將利用這些 Schema 來生成和驗證配置。  
* **里程碑 0.3：最小可行插件註冊與組裝**  
  * [x] 實作 PluginRegistryProvider 的核心邏輯，支持插件工廠註冊。  
  * [x] 實現一個基於 HttpServerProvider 的最小 Web 服務，可響應健康檢查。  
  * [x] 實現 OtelZapLoggerProvider，並在啟動時組裝。  
  * [x] 實現一個簡單的「Hello World」型 UIPagePlugin，並透過 composition.yaml 註冊和載入。  
  * **AI 協同點**：AI 協助生成插件工廠和初始化邏輯，並根據 composition.yaml **向 AI 發出具體的指令以生成 main.go 中組裝程式碼的骨架與填充邏輯**，實現平台組件的自動化組合。  
* **里程碑 0.4：CI/CD 與可觀察性基礎**  
  * [ ] 設定基本的 Git Workflow 和 CI/CD Pipeline (例如，自動化測試、Go Build)。  
  * [ ] 整合 Prometheus/Grafana 進行基礎指標監控（CPU, Memory, Request Rate）。  
  * [ ] 整合 Jaeger/Zipkin 進行分散式追蹤。  
  * **AI 協同點**：AI 協助撰寫 CI/CD 腳本片段，並建議監控指標。

Phase 1: 核心偵測能力與用戶體驗 (MVP 功能)  
核心目標：實現 Detectviz 平台的核心數據導入、異常偵測與基礎數據展示能力，提供第一個可用的 MVP 版本。

* **里程碑 1.1：數據導入引擎**  
  * [ ] 實現 ImporterPlugin 的 CSV 導入器，可將 CSV 數據解析並儲存至資料庫。  
  * [x] 定義 Detector 實體和 DetectorRepository 介面。  
  * **AI 協同點**：AI 協助撰寫 CSV 解析邏輯和資料庫儲存邏輯。  
* **里程碑 1.2：基礎偵測服務**  
  * [ ] 實現一個基於規則或簡單統計模型的 DetectorPlugin (例如，閾值告警)。  
  * [ ] 實現 AnalysisEngine 介面，作為偵測結果處理的抽象。  
  * **AI 協同點**：AI 協助生成偵測邏輯的程式碼和測試用例。  
* **里程碑 1.3：基礎 Web UI**  
  * [ ] 實現基礎的用戶登錄和註冊頁面 (依賴 keycloak_auth_provider)。  
  * [ ] 實現一個展示偵測結果列表的 UI 頁面 (依賴 ui_page_plugin)。  
  * **AI 協同點**：AI 協助生成前端頁面的 HTML/JS 骨架和樣式。

Phase 2: AI 增強與生態系統擴展  
核心目標：利用 LLM 和向量資料庫增強偵測與分析能力，並拓展平台的插件生態系統，同時深化 AI 輔助開發的自動化程度。

* **里程碑 2.1：LLM 整合**  
  * [ ] 實現 LLMProvider 介面，集成一個外部 LLM 服務（例如，Gemini API）。  
  * [ ] 實現 EmbeddingStoreProvider 介面，集成向量資料庫。  
  * [ ] 升級 AnalysisEngine，使其能利用 LLM 對偵測結果進行自然語言解釋或歸因。  
  * **AI 協同點**：AI 協助撰寫 LLM 互動程式碼、嵌入向量生成邏輯。  
* **里程碑 2.2：高級偵測與分析插件**  
  * [ ] 實現一個基於機器學習模型（例如，Isolation Forest）的 DetectorPlugin。  
  * [ ] 開發一個結合 LLM 和數據庫查詢的「智慧問答」插件，允許用戶自然語言查詢數據。  
  * **AI 協同點**：AI 協助生成複雜的分析演算法和業務邏輯。  
* **里程碑 2.3：AI 指令規範與實踐 (ai_directives_spec)**  
  * [ ] 制定並發布 docs/ai_scaffold/ai_directives_spec.md，明確定義 AI 在程式碼生成、配置、文檔撰寫中應遵循的專用指令標籤和語義。  
  * [ ] 將這些 AI 指令應用於核心介面定義 [INTERFACE_SPEC.md](./architecture/interface_spec.md) 和工程規範 [ENGINEERING_SPEC.md](ENGINEERING_SPEC.md) 中，並在相關的 AI 輔助開發流程中強制執行。  
  * **AI 協同點**：AI 協助歸納和撰寫 ai_directives_spec 的內容，並在其後續生成行為中遵循這些指令。  
* **里程碑 2.4：AI Scaffold 藍圖與模板庫擴充**  
  * [ ] 擴充 docs/ai_scaffold/scaffolding_blueprints/ 目錄，為不同類型的插件、服務、或模組提供標準化的、高層次的 AI Scaffold 藍圖。  
  * [ ] 擴充 docs/templates/ai_scaffolding/ 目錄，納入 AI 生成程式碼時可參考的詳細代碼模板和結構，例如新插件的通用文件結構、Makefile 範本、README 骨架等。  
  * [ ] 實作 AI Scaffold 工具鏈，使其能夠基於選定的藍圖和模板，結合 AI 指令，自動生成具體的程式碼文件。  
  * **AI 協同點**：AI 協助設計和組織藍圖與模板的內容，並利用這些資源進行更複雜、更規範的程式碼生成。  
* **里程碑 2.5：第三方插件接入**  
  * [ ] 開發 SDK 或範例，展示如何開發和接入第三方插件。  
  * [ ] 建立插件市集或貢獻指南。

Phase 3: 穩定性、效能與操作化  
核心目標：提升平台的穩定性、效能和操作性，使其能應對生產環境的挑戰。

* **里程碑 3.1：效能最佳化**  
  * [ ] 對核心數據路徑進行效能分析和優化。  
  * [ ] 實作快取機制。  
* **里程碑 3.2：高可用與容錯**  
  * [ ] 實作資料庫主從複製。  
  * [ ] 引入消息佇列增強異步處理能力。  
* **里程碑 3.3：安全增強**  
  * [ ] 實作 CSRFTokenProvider 和 RateLimiter。  
  * [ ] 進行安全審計和滲透測試。

**路線圖進度追蹤與更新說明**

本路線圖是一個動態文檔，將根據專案的實際進度、團隊容量和新的發現進行調整。

**進度標記**：

* [ ]：待完成  
* [ ]：已完成

**定期審查**：建議團隊成員至少每雙週審查一次路線圖，並根據實際完成情況更新任務標記。

**AI 互動**：

* 在請求 AI 協助時，請明確指出任務所屬的階段和里程碑，這有助於 AI 理解上下文。  
* 當某些任務涉及新的配置或程式碼結構時，**應首先確保相關的 JSON Schema 或 Go 介面已定義或已更新**，並在提示中提及。  
* **積極利用 ai_directives_spec 中定義的 AI 指令**，指導 AI 進行精確的程式碼生成和文件撰寫。

**溝通**：任何重大的路線圖變更都應在團隊內進行討論和確認。

