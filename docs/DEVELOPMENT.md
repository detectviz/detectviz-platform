# **DEVELOPMENT.md - 開發環境設置與指南**

本文件旨在提供 Detectviz 平台的本地開發環境設置、常用開發工具、本地運行、測試編寫和除錯技巧的詳細指南。

## **1. 前提條件 (Prerequisites)**

在開始之前，請確保您的系統滿足以下要求：

* **Go 語言**：Go 1.21 或更高版本。  
  * [下載與安裝 Go](https://golang.org/doc/install)  
* **Git**：版本控制工具。  
  * [下載與安裝 Git](https://git-scm.com/downloads)  
* **Docker / Docker Compose (推薦)**：用於運行本地資料庫、消息佇列等外部依賴服務。  
  * [下載與安裝 Docker Desktop](https://www.docker.com/products/docker-desktop)  
* **Make (可選)**：如果專案中定義了 Makefile，make 命令將簡化許多開發任務。  
  * macOS/Linux 通常預裝。Windows 用戶可能需要安裝 [Chocolatey](https://chocolatey.org/) 並透過 choco install make 安裝。  
* **MySQL 客戶端 (可選)**：如果您需要直接與本地 MySQL 資料庫交互，例如使用 mysql CLI 或 GUI 工具。

## **2. 本地開發環境設置 (Local Development Setup)**

### **2.1 克隆專案儲存庫**

首先，將 Detectviz 平台的程式碼克隆到您的本地機器：

git clone https://github.com/detectviz/detectviz-platform.git # 請替換為實際的儲存庫 URL  
cd detectviz-platform

### **2.2 安裝 Go 模組依賴**

進入專案根目錄後，執行以下命令下載所有 Go 模組依賴：

go mod tidy

### **2.3 準備配置檔**

Detectviz 平台透過 configs/composition.yaml 來聲明式地定義要組裝的插件和服務。

* **configs/composition.yaml**：這是平台預設的組裝配置。  
* **configs/app_config.yaml**：包含應用程式的通用配置。

您可以創建一個 configs/composition.local.yaml 或 configs/app_config.local.yaml 來覆蓋預設配置，特別是用於本地開發環境的特定設定（例如連接本地資料庫、啟用除錯日誌等）。這些 .local.yaml 文件會被 Git 忽略，不會提交到儲存庫。

**AI 協同提示**：AI 在生成配置時會參考 [CONFIGURATION_REFERENCE.md](http://docs.google.com/docs/CONFIGURATION_REFERENCE.md) 中的 JSON Schema。在手動修改配置時，請確保其符合 Schema 規範，以避免啟動時的驗證錯誤。

### **2.4 啟動外部依賴服務 (使用 Docker Compose)**

Detectviz 平台可能依賴於外部服務，例如 MySQL 資料庫、NATS 消息佇列等。我們建議使用 Docker Compose 在本地快速啟動這些服務。

在專案根目錄下，如果存在 docker-compose.yaml，您可以透過以下命令啟動所有依賴服務：

docker-compose up -d # -d 表示在後台運行

當您完成開發時，可以使用以下命令停止並移除服務：

docker-compose down

## **3. 運行 Detectviz 平台 (Run Locally)**

### **3.1 啟動 API 服務**

要啟動 Detectviz 平台的 API 服務，請在專案根目錄下執行：

go run ./cmd/api

成功啟動後，您將在終端看到類似 Detectviz 平台核心 MVP 已準備就緒！ 的日誌輸出。

### **3.2 運行 CLI 工具**

Detectviz 也提供了 CLI 工具用於管理和交互。您可以使用以下方式運行 CLI 命令：

go run ./cmd/cli version  
# 預期輸出: Detectviz CLI Version: 0.0.1-MVP

## **4. 測試 (Testing)**

Detectviz 平台採用嚴格的測試策略，包括單元測試、集成測試和端到端測試。詳細的測試原則和規範請參考 [ENGINEERING_SPEC.md](http://docs.google.com/docs/ENGINEERING_SPEC.md) 中的「測試規範」章節。

### **4.1 運行所有測試**

在專案根目錄下，您可以運行所有測試：

go test ./...

### **4.2 運行特定套件的測試**

要運行特定 Go 套件的測試，例如 pkg/domain/services：

go test ./pkg/domain/services

### **4.3 運行特定測試函數**

要運行單個測試函數，例如 TestUserService_CreateUser：

go test ./pkg/domain/services -run TestUserService_CreateUser

### **4.4 運行測試並查看覆蓋率**

go test -cover ./...  
# 或生成 HTML 覆蓋率報告  
go test -coverprofile=coverage.out ./... && go tool cover -html=coverage.out

**AI 協同提示**：AI 可以根據您的程式碼自動生成單元測試的骨架和初步測試用例。請利用此功能加速測試的編寫，並確保其符合 ENGINEERING_SPEC.md 中的測試規範。

## **5. 除錯 (Debugging)**

### **5.1 使用 IDE 除錯**

推薦使用支援 Go 語言的 IDE (如 VS Code, GoLand) 進行除錯。

VS Code 配置範例：  
在 .vscode/launch.json 中，您可以配置除錯啟動項：  
{  
    "version": "0.2.0",  
    "configurations": [  
        {  
            "name": "Launch API Server",  
            "type": "go",  
            "request": "launch",  
            "mode": "debug",  
            "program": "${workspaceFolder}/cmd/api",  
            "env": {  
                // 您可以在這裡添加環境變數，例如：  
                // "APP_ENV": "development"  
            },  
            "args": []  
        },  
        {  
            "name": "Launch CLI Tool",  
            "type": "go",  
            "request": "launch",  
            "mode": "debug",  
            "program": "${workspaceFolder}/cmd/cli",  
            "env": {},  
            "args": ["version"] // 這裡可以修改為您要除錯的 CLI 命令參數  
        }  
    ]  
}

設置斷點後，啟動除錯器即可逐步執行程式碼。

### **5.2 日誌除錯**

Detectviz 平台使用結構化日誌（透過 OtelZapLoggerProvider 實現）。在開發環境中，您可以將日誌級別設置為 DEBUG，以獲取更詳細的運行資訊。

修改 configs/app_config.local.yaml (如果存在) 或 configs/app_config.yaml：

logging:  
  level: debug # 或 info, warn, error  
  format: console # 或 json

## **6. 常見開發工作流程 (Common Development Workflows)**

### **6.1 新增一個新的插件 (Plugin)**

如果您要新增一個新的平台插件（例如，一個新的 DetectorPlugin 或 ImporterPlugin），請遵循以下步驟：

1. **定義介面**：如果尚未存在，請在 pkg/domain/plugins 或 pkg/platform/contracts 中定義新的 Go 介面，並更新 [interface_spec.md](http://docs.google.com/docs/architecture/interface_spec.md)。  
2. **實現插件**：在 internal/platform/plugins 或 internal/platform/providers 下創建新的目錄和 Go 檔案來實現您的插件。  
3. **創建工廠**：為您的插件實現一個 NewFactory() 函數，使其可以被 PluginRegistry 發現。  
4. **更新 composition.yaml**：在 configs/composition.yaml 中註冊您的新插件及其配置。  
5. **更新 main.go 組裝邏輯 (如果需要)**：雖然 AI 腳手架會自動處理大部分組裝，但對於新的依賴或特殊組裝邏輯，您可能需要手動調整 cmd/api/main.go（或利用 AI 協助）。  
6. **編寫測試**：為您的新插件編寫單元測試和可能的集成測試。

**AI 協同提示**：利用 AI 腳手架 (scaffold_workflow.md) 來自動生成新插件的骨架、介面實現和組裝邏輯。AI 將根據您提供的介面定義和配置，生成符合 ENGINEERING_SPEC.md 規範的程式碼。

### **6.2 修改現有服務**

1. **理解現有程式碼**：查閱相關的 Go 檔案、GoDoc 註解和文檔（例如 ARCHITECTURE.md, ENGINEERING_SPEC.md）。  
2. **修改程式碼**：進行必要的修改。  
3. **更新測試**：如果您的修改影響了現有功能，請更新或新增測試。  
4. **運行測試**：確保所有相關測試通過。

## **7. AI 輔助開發 (AI-Assisted Development)**

Detectviz 平台的核心設計理念之一就是深度整合 AI 輔助開發。您可以利用 AI 進行：

* **程式碼生成**：根據介面定義、JSON Schema 和 AI 指令生成程式碼骨架、實現邏輯。  
* **配置管理**：自動生成和驗證配置檔。  
* **測試用例生成**：為函數和方法生成單元測試。  
* **文檔撰寫**：協助撰寫 GoDoc、README 或其他技術文檔。

請參考 [docs/ai_scaffold/scaffold_workflow.md](http://docs.google.com/docs/ai_scaffold/scaffold_workflow.md) 和 [AI 交互指南](http://docs.google.com/README.md#ai-交互指南) 以獲取更詳細的 AI 協同開發指南。

享受您的開發之旅！

