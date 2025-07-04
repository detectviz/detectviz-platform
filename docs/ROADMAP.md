**Detectviz 平台發展路線圖**  
基於對 CNCF 平台工程核心概念的深度內化，Detectviz 平台的基礎規劃已然穩健清晰。本文件旨在將這份全面的藍圖，轉化為具體的 可實踐路線圖。我們將採用 迭代、增量 的方式，分階段進行規劃與實踐，並持續將 產品思維 (Product Thinking) 與 AI 驅動開發 (AI-Driven Development, AIDD) 貫穿其中，確保每個階段都能交付實際價值、收集用戶反饋，並最大化 AI 在開發、審查、測試與部署環節的效能。  
AI 協同與文檔體系：AI 的「心智模型」與「執行藍圖」  
在 Detectviz 平台中，AI 不僅僅是開發工具，更是核心的協同夥伴。AI 的高效率與高準確性，深度依賴於平台定義的清晰、一致的文檔體系。這些文檔共同構成了 AI 理解系統、生成程式碼、驗證配置和執行測試的「心智模型」與「執行藍圖」：

* **ARCHITECTURE.md (平台架構)**：為 AI 提供了系統的宏觀視角和高層次設計原則。AI 將解析此文件以理解模組職責、層次劃分、數據流與控制流，確保生成程式碼和設計方案符合整體架構規範。  
* **ENGINEERING_SPEC.md (平台工程實作規範)**：定義了開發者和 AI 都必須遵循的程式碼實作規範、技術選型和最佳實踐。AI 將依據此文件，生成符合規範的程式碼風格、錯誤處理機制和測試策略。  
* **interface_spec.md (核心介面定義)**：包含了所有核心服務和插件的 Go 介面定義，以及關鍵的 AI 指令標籤。這些介面是 AI 生成程式碼的「契約」，AI 將嚴格遵循這些定義來確保程式碼的功能正確性和介面兼容性。  
* **scaffold_workflow.md (AI 腳手架工作流程)**：這是直接指導 AI 進行開發和自動化的核心文件。它詳細定義了 AI 如何理解需求、生成程式碼、執行驗證的具體步驟和流程，包括對 AI 指令、工廠模式和 JSON Schema 驗證的強制性要求。  
* **main_go_assembly.tmpl (主程式組裝模板)**：是 AI 腳手架用來生成 main.go 檔案的核心模板。AI 將利用此模板，結合 composition.yaml 的配置和 interface_spec.md 的介面定義，自動化平台的組裝邏輯，實現服務和插件的動態組合。

透過這些文檔的協同作用，AI 能夠：

* **理解複雜性**：從高層次架構到具體實現細節，全面掌握系統。  
* **自動化生成**：根據需求和規範，自動生成程式碼、配置和測試用例。  
* **確保一致性**：保證所有生成內容符合平台定義的標準和最佳實踐。  
* **加速迭代**：大幅減少人工介入，加速開發週期。

**核心原則：迭代交付，價值先行，AI 協同**

* **最小可行平台 (MVP) 思維**：每個階段都將以交付最小可行功能集為目標，快速驗證假設，避免過度工程。  
* **用戶反饋驅動**：持續與產品團隊互動，將用戶需求和痛點作為功能優先級排序的依據。  
* **核心依賴優先**：優先實現其他插件所依賴的核心平台能力。  
* **自動化為本**：從一開始就將 CI/CD 和自動化部署納入考量。  
* **AI 深度協同**：在任務規劃、程式碼生成、配置管理 (特別是依賴 JSON Schema 進行推斷和驗證)、程式碼審查、測試與部署等環節深度整合 AI 角色，逐步提高自動化比例，實現從高人工干預到 AI 驅動的開發流程轉型。

### **Detectviz 平台分階段實踐路線圖**

Phase 0: 平台核心骨架與 AI 基礎建設  
核心目標：建立一個符合 Detectviz 平台架構文檔 和 Detectviz 平台工程規範 的可運行、可配置、可觀察的平台骨架，能夠組裝和運行最基礎的插件。同時，為 AI 驅動開發奠定堅實的基礎。

* **里程碑 0.1：基礎 Go 專案結構與核心組件介面**  
  * [x] 建立標準的 Go Modules 專案結構 (/cmd, /internal, /pkg, /configs, /docs 等)，遵循 ENGINEERING_SPEC.md 的專案結構規範。  
  * [x] 定義 pkg/platform/contracts 下的核心平台供應商介面 (LoggerProvider, ConfigProvider, HttpServerProvider, DBClientProvider, PluginRegistryProvider 等)，並在 interface_spec.md 中進行管理。  
  * [x] 定義 pkg/domain/plugins 下的通用插件介面 (Plugin, Importer 等)，並在 interface_spec.md 中進行管理。  
  * **AI 協同點**：AI 將依據高層次設計需求（參考 ARCHITECTURE.md），並結合 ENGINEERING_SPEC.md 的規範，協助生成符合 interface_spec.md 標準的介面定義和 GoDoc 註解。  
* **里程碑 0.2：配置管理與 JSON Schema 導入**  
  * [x] 實現 ConfigProvider，支持從 app_config.yaml 和 composition.yaml 載入配置，其實現將遵循 ENGINEERING_SPEC.md 的配置管理規範。  
  * [x] 為 app_config.yaml 和 composition.yaml 定義完整的 JSON Schema (schemas/app_config.json, schemas/composition.json)。  
  * [x] 為核心且穩定的平台供應商 (如 http_server_provider, otelzap_logger_provider, gorm_mysql_client_provider, keycloak_auth_provider) 撰寫獨立的 JSON Schema (schemas/plugins/*.json)。  
  * [x] 在 CONFIGURATION_REFERENCE.md 和 ENGINEERING_SPEC.md 中更新配置規範，明確引用 JSON Schema。  
  * [x] 實作配置載入後的 JSON Schema 驗證機制，確保配置的合法性。  
  * **AI 協同點**：AI 將依賴現有 Go 介面（來自 interface_spec.md）和平台文檔（來自 ARCHITECTURE.md, ENGINEERING_SPEC.md），推斷並自動生成 JSON Schema。未來 AI 將利用這些 Schema 來生成和驗證所有平台配置，嚴格遵循 scaffold_workflow.md 中定義的配置生成流程。  
* **里程碑 0.3：最小可行插件註冊與組裝**  
  * [x] 實作 PluginRegistryProvider 的核心邏輯，支持插件工廠註冊。  
  * [x] 實現一個基於 HttpServerProvider 的最小 Web 服務，可響應健康檢查。  
  * [x] 實現 OtelZapLoggerProvider，並在啟動時組裝。  
  * [x] 實現一個簡單的「Hello World」型 UIPagePlugin，並透過 composition.yaml 註冊和載入。  
  * **AI 協同點**：AI 將依據 composition.yaml 的配置，結合 interface_spec.md 中定義的插件介面，並參考 main_go_assembly.tmpl 模板，自動生成 main.go 中組裝程式碼的骨架與填充邏輯，實現平台組件的自動化組合。此過程將嚴格遵循 scaffold_workflow.md 中定義的 AI 腳手架工作流程，確保生成的組裝代碼符合 ARCHITECTURE.md 的組裝根原則。  
* **里程碑 0.4：CI/CD 與可觀察性基礎**  
  * [x] 設定基本的 Git Workflow 和 CI/CD Pipeline (例如，自動化測試、Go Build)，並將其設計原則納入 ENGINEERING_SPEC.md。  
  * [x] 整合 Prometheus/Grafana 進行基礎指標監控（CPU, Memory, Request Rate），並定義關鍵指標在 ENGINEERING_SPEC.md 中。  
  * [x] 整合 Jaeger/Zipkin 進行分散式追蹤，確保其與 OpenTelemetry 規範（來自 ARCHITECTURE.md）一致。  
  * **AI 協同點**：AI 將協助撰寫 CI/CD 腳本片段，並根據 ENGINEERING_SPEC.md 中定義的可觀察性規範，建議監控指標和追蹤配置。  
* **里程碑 0.5：資料導入與偵測基礎插件**  
  * [x] 實現 ImporterPlugin 的 CSV 導入器，可將 CSV 數據解析並儲存至資料庫。此實現將遵循 ENGINEERING_SPEC.md 的數據層細節規範。  
  * [x] 實現一個基於規則或簡單統計模型的 DetectorPlugin (例如，閾值告警)。此插件的開發將遵循 ENGINEERING_SPEC.md 的插件開發規範。  
  * **AI 協同點**：AI 將依據 ImporterPlugin 和 DetectorPlugin 介面（來自 interface_spec.md）和 CSV 數據格式定義，自動生成 CSV 解析邏輯、偵測邏輯和對應的資料庫儲存邏輯，並遵循 ENGINEERING_SPEC.md 的程式碼規範。

Phase 1: 核心偵測能力與用戶體驗 (MVP 功能)  
核心目標：實現 Detectviz 平台的核心數據導入、異常偵測與基礎數據展示能力，提供第一個可用的 MVP 版本。

* **里程碑 1.1：定義偵測實體與介面**  
  * [x] 定義 Detector 實體和 DetectorRepository 介面，並在 pkg/domain/entities 和 pkg/domain/interfaces 中定義，同時更新 interface_spec.md。  
  * **AI 協同點**：AI 將依據 DetectorPlugin 介面（來自 interface_spec.md）自動生成偵測邏輯的程式碼骨架和初步測試用例，並參考 ARCHITECTURE.md 中的事件驅動架構原則。  
* **里程碑 1.2：基礎偵測服務**  
  * [ ] 實現 AnalysisEngine 介面，作為偵測結果處理的抽象，並更新 interface_spec.md。  
  * **AI 協同點**：AI 將依據 AnalysisEngine 介面（來自 interface_spec.md），自動生成偵測結果處理邏輯的程式碼骨架和初步測試用例。  
* **里程碑 1.3：基礎 Web UI**  
  * [ ] 實現基礎的用戶登錄和註冊頁面 (依賴 keycloak_auth_provider)。前端與後端交互將遵循 ARCHITECTURE.md 中定義的 API Gateway / BFF 策略。  
  * [ ] 實現一個展示偵測結果列表的 UI 頁面 (依賴 ui_page_plugin)。  
  * **AI 協同點**：AI 將依據 UI 頁面需求和後端 API 定義，自動生成前端頁面的 HTML/JS 骨架和樣式，並建議符合 ENGINEERING_SPEC.md 的前端開發規範。  
* **里程碑 1.4：RAG 引擎初步整合**  
  * [ ] 實現 LLMProvider 介面，集成一個外部 LLM 服務（例如，Gemini API）。此實現將位於 internal/infrastructure/platform，並遵循 ENGINEERING_SPEC.md 的外部服務整合規範。  
  * [ ] 實現 EmbeddingStoreProvider 介面，集成向量資料庫。  
  * [ ] 升級 AnalysisEngine，使其能利用 LLM 對偵測結果進行自然語言解釋或歸因。  
  * **AI 協同點**：AI 將依據 LLMProvider 和 EmbeddingStoreProvider 介面（來自 interface_spec.md），自動生成 LLM 互動程式碼、嵌入向量生成邏輯，並確保其符合 ENGINEERING_SPEC.md 的安全和性能規範。  
* **里程碑 1.5：AI 指令規範與應用**  
  * [ ] 制定並發布 docs/ai_scaffold/ai_directives_spec.md，明確定義 AI 在程式碼生成、配置、文檔撰寫中應遵循的專用指令標籤和語義。  
  * [ ] 將這些 AI 指令應用於核心介面定義 [INTERFACE_SPEC.md](http://docs.google.com/architecture/interface_spec.md) 和工程規範 [ENGINEERING_SPEC.md](http://docs.google.com/ENGINEERING_SPEC.md) 中，並在相關的 AI 輔助開發流程中強制執行。  
  * **AI 協同點**：AI 將協助歸納和撰寫 ai_directives_spec 的內容，並在其後續生成行為中嚴格遵循這些指令，這將是 AI 自我優化和規範化的關鍵一步。

Phase 2: AI 增強與生態系統擴展  
核心目標：利用 LLM 和向量資料庫增強偵測與分析能力，並拓展平台的插件生態系統，同時深化 AI 輔助開發的自動化程度。

* **里程碑 2.1：AI Scaffold 藍圖與模板庫擴充**  
  * [ ] 擴充 docs/ai_scaffold/scaffolding_blueprints/ 目錄，為不同類型的插件、服務、或模組提供標準化的、高層次的 AI Scaffold 藍圖，這些藍圖將基於 ARCHITECTURE.md 的層次分解。  
  * [ ] 擴充 docs/templates/ai_scaffolding/ 目錄，納入 AI 生成程式碼時可參考的詳細代碼模板和結構，例如新插件的通用文件結構、Makefile 範本、README 骨架等，這些模板將遵循 ENGINEERING_SPEC.md 的程式碼風格。  
  * [ ] 實作 AI Scaffold 工具鏈，使其能夠基於選定的藍圖和模板，結合 AI 指令（來自 ai_directives_spec.md），自動生成具體的程式碼文件。  
  * **AI 協同點**：AI 將協助設計和組織藍圖與模板的內容，並利用這些資源進行更複雜、更規範的程式碼生成，極大提升開發效率。  
* **里程碑 2.2：第三方插件接入**  
  * [ ] 開發 SDK 或範例，展示如何開發和接入第三方插件。這將涉及更新 ENGINEERING_SPEC.md 中的插件開發指南。  
  * [ ] 建立插件市集或貢獻指南。  
  * **AI 協同點**：AI 將協助生成 SDK 範例程式碼，並根據現有插件介面（來自 interface_spec.md）自動生成新插件的骨架，加速第三方開發者的接入。

Phase 3: 穩定性、效能與操作化  
核心目標：提升平台的穩定性、效能和操作性，使其能應對生產環境的挑戰。

* **里程碑 3.1：效能最佳化**  
  * [ ] 對核心數據路徑進行效能分析和優化，並將優化策略納入 ENGINEERING_SPEC.md。  
  * [ ] 實作快取機制（例如，集成 CacheProvider 介面），並更新 interface_spec.md。  
  * **AI 協同點**：AI 可以分析監控數據（來自 OpenTelemetry 整合），識別性能瓶頸，並建議優化方案，包括生成快取實現的程式碼模板。  
* **里程碑 3.2：高可用與容錯**  
  * [ ] 實作資料庫主從複製，遵循 ARCHITECTURE.md 中的數據庫高可用原則。  
  * [ ] 引入消息佇列增強異步處理能力，深化 ARCHITECTURE.md 中的異步通訊策略。  
  * **AI 協同點**：AI 將協助生成數據庫複製的配置腳本，並根據 ARCHITECTURE.md 的容錯機制，建議熔斷器和重試邏輯的實現。  
* **里程碑 3.3：安全增強**  
  * [ ] 實作 CSRFTokenProvider 和 RateLimiter，並在 interface_spec.md 中定義相關介面。  
  * [ ] 進行安全審計和滲透測試，並將發現的問題和解決方案更新到 ENGINEERING_SPEC.md 的安全規範中。  
  * **AI 協同點**：AI 可以協助生成安全相關的程式碼骨架（例如 CSRF Token 驗證邏輯），並建議基於 ENGINEERING_SPEC.md 的安全審計點。

**路線圖進度追蹤與更新說明**

本路線圖是一個動態文檔，將根據專案的實際進度、團隊容量和新的發現進行調整。

**進度標記**：

* [ ]：待完成  
* [x]：已完成

**最新更新記錄**：
* **2025/01/04**：完成里程碑 0.4 CI/CD 與可觀察性基礎
  - ✅ 建立 GitHub Actions CI/CD Pipeline (測試、構建、安全掃描、Docker 構建)
  - ✅ 創建 golangci-lint 代碼質量檢查配置
  - ✅ 實現完整的可觀察性堆棧 (Prometheus + Grafana + Jaeger + OTEL Collector)
  - ✅ 實現 PrometheusMetricsProvider 指標收集服務
  - ✅ 實現 JaegerTracingProvider 分散式追蹤服務 (使用 OTLP)
  - ✅ 修復並更新 contracts.go 中的 MetricsProvider 和 TracingProvider 介面

* **2025/01/04**：完成里程碑 0.5 資料導入與偵測基礎插件
  - ✅ 實現 CSVImporterPlugin 完整功能 (支持自定義分隔符、批量插入、數據驗證)
  - ✅ 實現 ThresholdDetectorPlugin 閾值偵測器 (支持上下限檢測、可配置嚴重程度)
  - ✅ 創建完整的單元測試和集成測試
  - ✅ 創建插件配置示例和測試數據
  - ✅ 修復插件介面契約問題，確保實現正確性

**當前階段總結**：Phase 0 已全部完成 (5/5 里程碑)，平台核心骨架與 AI 基礎建設已建立完成。

**定期審查**：建議團隊成員至少每雙週審查一次路線圖，並根據實際完成情況更新任務標記。

AI 互動指南：  
為了最大化 AI 在完成里程碑任務中的效能，請在與 AI 互動時遵循以下原則：

* **明確上下文**：在請求 AI 協助時，請明確指出任務所屬的階段和里程碑，這有助於 AI 理解上下文和目標。  
* **文檔優先**：當某些任務涉及新的配置或程式碼結構時，**應首先確保相關的 JSON Schema 或 Go 介面已定義或已更新**（例如在 interface_spec.md 中），並在提示中明確提及這些文檔。AI 將優先從這些規範性文件中獲取信息。  
* **利用 AI 指令**：積極利用 ai_directives_spec.md 中定義的 AI 指令，指導 AI 進行精確的程式碼生成、配置撰寫和文檔更新。  
* **提供範例**：如果可能，提供相關的程式碼片段或文檔範例，作為 AI 生成的參考。  
* **迭代與反饋**：將 AI 生成的內容視為初稿，進行審閱並提供具體反饋，幫助 AI 學習和改進。

**溝通**：任何重大的路線圖變更都應在團隊內進行討論和確認。
