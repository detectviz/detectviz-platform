**Detectviz 平台**

![Detectviz Platform Banner Placeholder]

## **專案概述**

Detectviz 是一個現代、模組化且可擴展的 **內部應用平台**。我們將所有平台級功能抽象為可插拔的「軟體項目 (Software Items)」，旨在為產品開發者、數據科學家和 ML 工程師提供標準化、自服務、可組合且經過治理的工具與服務。Detectviz 的核心目標是降低認知負擔、加速價值交付，並為未來實現 **AI 驅動的自動化擴展 (AI-Powered Automation)** 奠定堅實基礎。

本平台深度內化 CNCF TAG App Delivery 所倡導的 **「平台工程 ++ (Platform Engineering ++) 」** 理念，將平台視為一個「虛擬圖書館」，以提供卓越的開發者體驗 (DevX)。

## **核心特色與能力**

Detectviz 透過豐富的插件系統提供以下核心能力，這些能力均可透過配置彈性組合：

* **平台基礎與核心服務**：統一設定、日誌、插件管理、功能開關等。  
* **應用程式運行與連接**：提供 HTTP/gRPC/CLI 接口、任務排程、通知、服務發現等運行時支持。  
* **數據管理與持久化**：抽象化資料庫、儲存、緩存、事件總線，並支持數據導入/匯出。  
* **安全與策略護欄**：內建認證、授權、秘密管理和通用策略執行。  
* **可觀測性與平台洞察**：全面日誌、指標、追蹤、健康檢查與成本管理。  
* **開發者工具與擴展**：提供可自定義的 Web UI 頁面、自動化部署與發布管理工具。

## **為什麼選擇 Detectviz？**

* **CNCF 最佳實踐對齊**：遵循 CNCF Platform Engineering 白皮書的指導原則，確保架構的前瞻性與兼容性。  
* **高度可擴展性**：**「一切皆插件」** 的設計理念，允許靈活地替換、新增和組合平台功能。  
* **自服務驅動**：提供豐富的 API 與 CLI 接口，賦能開發者自主管理應用生命週期。  
* **AI 賦能就緒**：標準化與結構化的設計，為未來 AI 自動化程式碼生成、平台配置與運維提供了理想的基礎，**具體詳見 AI 交互指南。**  
* **降低認知負擔**：抽象化底層複雜性，讓產品團隊專注於業務創新。

## **架構總覽**

Detectviz 平台基於 **Clean Architecture** 原則構建，嚴格劃分了領域層、應用程式層、介面轉接層和基礎設施層，並透過介面實現了各層之間的 **依賴反轉**。這種設計確保了核心業務邏輯的獨立性、可測試性與可維護性。

**「一切皆插件」的核心理念** 貫穿整個平台，所有關鍵功能都被抽象為可插拔的插件。

configs/composition.yaml 作為平台的 **「組裝根」核心配置檔**，聲明式地定義了哪些插件應被啟用，以及如何配置它們，這大大簡化了平台的啟動和維護。

**配置驗證機制**：平台採用 JSON Schema 對所有配置進行嚴格驗證，確保配置的正確性和一致性。在啟動時會自動驗證 app_config.yaml、composition.yaml 以及各插件的配置，提供明確的錯誤訊息以協助問題排查。

欲了解更多詳細的架構設計、層次職責及 AI 在各層次的協作方式，請查閱 [docs/ARCHITECTURE.md](./docs/ARCHITECTURE.md)。

## **文檔導航**

為了方便您快速找到所需的資訊，Detectviz 平台的所有重要文檔都集中存放在 docs/ 目錄下：

* [**ARCHITECTURE.md**](./docs/ARCHITECTURE.md)：平台整體架構、核心設計原則、模組劃分。  
* [**ENGINEERING_SPEC.md**](./docs/ENGINEERING_SPEC.md)：技術棧、程式碼實作規範、目錄與檔案命名慣例、程式碼風格、測試原則及插件開發規範。  
* [**ROADMAP.md**](./docs/ROADMAP.md)：平台未來發展藍圖、階段性目標和 AI 賦能路線圖。  
* [**CONFIGURATION_REFERENCE.md**](./docs/CONFIGURATION_REFERENCE.md)：平台所有可配置項的完整列表與詳細說明，包括 composition.yaml 的 Schema。  
* [**docs/architecture/interface_spec.md**](./docs/architecture/interface_spec.md)：**所有核心 Go 介面定義的規範性參考，對 AI 理解平台契約至關重要。**  
* [**docs/ai_scaffold/scaffold_workflow.md**](./docs/ai_scaffold/scaffold_workflow.md)：AI 輔助程式碼生成的核心工作流程，指導 AI 如何自動化任務。  
* [**CONTRIBUTING.md(待補)**](./CONTRIBUTING.md)：貢獻者指南，包含如何提交 Issue、Pull Request、程式碼規範等。  
* [**DEVELOPMENT.md(待補)**](./docs/DEVELOPMENT.md)：開發環境設置、常用開發工具、本地運行、測試編寫、除錯技巧。  
* [**DEPLOYMENT_GUIDE.md**](./docs/DEPLOYMENT_GUIDE.md)：如何將 Detectviz 平台部署到不同環境的詳細步驟和考量。  
* [**PLUGIN_GUIDE.md(待補)**](./docs/PLUGIN_GUIDE.md)：如何為 Detectviz 平台開發新插件的指南。  
* [**API_REFERENCE.md(待補)**](./docs/reference/API_REFERENCE.md)：平台所有公開 API 的詳細規範。  
* [**GLOSSARY.md(待補)**](./docs/GLOSSARY.md)：平台相關術語、概念和縮寫的定義。  
* [**TROUBLESHOOTING.md(待補)**](./docs/TROUBLESHOOTING.md)：常見問題及解決方案。  
* [**CHANGELOG.md(待補)**](./CHANGELOG.md)：平台所有發布版本的變更日誌。  
* [**SECURITY.md(待補)**](./SECURITY.md)：專案的安全策略與漏洞報告流程。

## **AI 交互指南**

Detectviz 平台設計之初就考慮了 AI 輔助開發，您可以利用 AI 進行程式碼生成、配置管理、測試用例生成甚至文檔撰寫。

### **如何利用 AI 進行 Scaffold (鷹架生成)？**

1. **理解需求：** 首先明確您希望 AI 完成什麼任務（例如：新增一個日誌提供者、一個新的數據導入插件）。  
2. **參考規範：** 確保您已熟悉 ARCHITECTURE.md 和 ENGINEERING_SPEC.md，這些是 AI 生成程式碼的基礎藍圖和規範。  
3. **準備配置：** 如果任務涉及插件組裝，請確保 configs/composition.yaml 已準備好，它將作為 AI 的核心輸入。  
4. **發出指令：** 根據您要生成的內容類型，向 AI 發出具體的指令。  
   * 核心功能入口： 對於自動化 main.go 中的插件工廠註冊和組裝邏輯，以及其他 AI 輔助腳手架功能，請參考：  
     AI Scaffold 工作流程詳情：docs/ai_scaffold/scaffold_workflow.md  
5. **集成與驗證：** 將 AI 生成的程式碼集成到您的專案中，並進行必要的測試和調整。

### **AI 指令標籤**

為了提高 AI 理解的精確性，我們在部分核心 Go 介面定義 (docs/architecture/interface_spec.md) 中加入了 AI 專用標籤，例如 AI_PLUGIN_TYPE、AI_IMPL_PACKAGE 等。這些標籤直接告訴 AI 介面的預期用途和實現細節，以確保生成程式碼的準確性。

## **快速上手 (開發環境)**

本指南將幫助您快速設置 Detectviz 的本地開發環境並啟動平台。

1. **前提條件**：  
   * **Go 語言**：確保您的開發環境已安裝 Go 1.21+。  
   * **Git**：用於克隆專案儲存庫。  
   * **Docker/Docker Compose (可選，用於本地資料庫等外部服務)**：建議安裝，以便啟動模擬的外部依賴服務。  
   * **MySQL 客戶端 (可選)**：如果您需要與本地 MySQL 資料庫交互。  
2. **啟動步驟**：  
   * **克隆專案**：  
     git clone https://github.com/your-org/detectviz.git # 將 'your-org' 替換為實際的組織/用戶名  
     cd detectviz

   * **安裝 Go 模組依賴**：  
     go mod tidy

   * **準備配置檔**：Detectviz 平台啟動時會讀取 configs/composition.yaml 來組裝服務。請確保該檔案存在。您可以根據需要創建一個 configs/composition.local.yaml 來覆蓋默認配置，例如連接本地資料庫。  
     * **提示**：如果您需要本地啟動資料庫或其他外部服務，建議參考 scripts/setup.sh 或專案根目錄下的 docker-compose.yaml (如果提供的話)。  
   * **啟動平台**：  
     go run ./cmd/api

3. **驗證 (Verify)**：  
   * **日誌輸出**：檢查終端輸出，確認平台是否成功啟動，並顯示類似 Detectviz 平台核心 MVP 已準備就緒 的日誌。  
   * **訪問 HTTP API**：打開瀏覽器或使用 curl 訪問平台提供的基礎 API，例如：  
     curl http://localhost:8080/api/v1/users # 檢查是否返回 400 Bad Request 或其他預期錯誤，表示路由已工作  
     curl http://localhost:8080/ui/dashboard # 如果 Dashboard UI 插件啟用，應返回 HTML 頁面

   * **CLI 命令**：打開另一個終端，測試 CLI 命令：  
     go run ./cmd/cli version

     應顯示 Detectviz CLI Version: 0.0.1-MVP。

## **貢獻指南**

我們歡迎任何形式的貢獻！請參考 [CONTRIBUTING.md](./CONTRIBUTING.md) 文件了解如何提交 Bug、功能請求或貢獻程式碼的詳細步驟。

## **架構重構說明**

本項目已從單一的 `main.go` 文件重構為符合 Clean Architecture 原則的多層結構。

### ✅ 已完成重構的組件

- **領域層 (pkg/domain)**
  - 實體定義 (User, Detector, AnalysisResult 等)
  - 領域介面 (UserRepository, DetectorRepository, AnalysisEngine)
  - 插件抽象 (Plugin, Importer, UIPagePlugin)
  - 領域錯誤定義

- **平台契約層 (pkg/platform/contracts)**
  - 核心平台服務介面 (Logger, ConfigProvider, HttpServerProvider 等)

- **基礎設施層 (internal/infrastructure)**
  - OtelZap 日誌實現
  - Viper 配置提供者實現

- **配置管理 (internal/config)**
  - 平台配置結構定義

- **應用程式入口 (cmd/api)**
  - 新的 main.go，展示基本的組裝流程

### 📝 重構參考

詳細的架構說明請參考：
- [docs/ARCHITECTURE.md](./docs/ARCHITECTURE.md) - 完整的平台架構說明
- [docs/ENGINEERING_SPEC.md](./docs/ENGINEERING_SPEC.md) - 詳細的目錄結構和實作規範
- 原始的 `main.go` 文件（保留作為完整實現參考）

### 🔄 使用重構後的結構

```bash
# 運行基本版本 (僅展示配置載入)
go run cmd/api/main.go

# 查看完整實現參考
cat main.go
```

## **授權條款**

本專案採用 MIT 授權。

