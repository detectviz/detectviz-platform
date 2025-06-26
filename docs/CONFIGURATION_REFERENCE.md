# **Detectviz 平台配置指南 - Configuration Reference**

本文件旨在提供 Detectviz 平台所有配置檔案的詳細說明與指南，包括其結構、用途、可用配置項及其驗證方式。理解這些配置對於平台的部署、運行和擴展至關重要。

## **1. 配置管理概覽**

Detectviz 平台使用 **Viper** 進行配置管理，支持從多個來源（如 YAML 檔案、環境變數、命令行參數）讀取和合併配置。所有的配置都通過 **JSON Schema** 進行嚴格驗證，以確保配置的正確性與一致性。

### **1.1 主要配置檔案**

* configs/app_config.yaml：平台核心的全局應用程式配置。  
* configs/composition.yaml：定義平台中所有插件的組合與具體配置。

### **1.2 配置驗證流程**

Detectviz 平台採用嚴格的配置驗證機制，確保配置的正確性和一致性：

**啟動時驗證**：
- 在應用程式啟動時，`app_config.yaml` 和 `composition.yaml` 會被載入
- 配置會自動根據對應的 JSON Schema 進行驗證
- 驗證失敗將導致啟動失敗並提供詳細的錯誤訊息

**插件配置驗證**：
- 每個插件的配置會根據 `schemas/plugins/` 目錄下的對應 Schema 進行驗證
- 支援動態插件配置驗證，確保插件參數的合法性

**CI/CD 整合**：
- 建議在 CI/CD 流程中加入配置驗證步驟
- 使用 JSON Schema 驗證工具檢查配置文件
- 確保提交的配置檔案始終有效

**Schema 文件位置**：
- 主配置 Schema：`schemas/app_config.json`、`schemas/composition.json`  
- 插件 Schema：`schemas/plugins/{plugin_type}.json`

## **2. app_config.yaml - 全局應用程式配置**

app_config.yaml 包含了 Detectviz 平台的全局設定，影響整個應用程式的行為。

**路徑**：configs/app_config.yaml

**對應 Schema**：schemas/app_config.json

**範例結構 (部分)**：

server:  
  port: 8080  
  readTimeout: 5s  
  writeTimeout: 10s  
log:  
  level: info  
  encoding: json  
  outputPaths:  
    - stdout  
  errorOutputPaths:  
    - stderr  
database:  
  dsn: "user:password@tcp(127.0.0.1:3306)/detectviz?parseTime=true"  
  maxOpenConns: 100  
  maxIdleConns: 10  
  connMaxLifetime: 1h  
security:  
  jwtSecretEnvVar: APP_JWT_SECRET  
  csrfTokenLifeTime: 1h

**主要配置項說明**：

| 配置項 | 類型 | 預設值 | 說明 |
| :---- | :---- | :---- | :---- |
| server.port | integer | 8080 | HTTP 服務監聽的端口。 |
| server.readTimeout | string | 5s | 服務器讀取請求主體的最大超時時間。例如 5s, 1m。 |
| server.writeTimeout | string | 10s | 服務器寫入響應的最大超時時間。例如 10s, 1m。 |
| log.level | string | info | 日誌的最低記錄級別 (debug, info, warn, error, dpanic, panic, fatal)。 |
| log.encoding | string | json | 日誌輸出格式 (json 或 console)。 |
| log.outputPaths | array of string | ["stdout"] | 日誌寫入的路徑列表。可以是 stdout, stderr 或檔案路徑 (/var/log/app.log)。 |
| log.errorOutputPaths | array of string | ["stderr"] | 錯誤日誌寫入的路徑列表。 |
| database.dsn | string | "" | 資料庫連接字符串 (Data Source Name)。例如 MySQL 的格式為 user:password@tcp(host:port)/database_name?param=value。 **必填**。 |
| database.maxOpenConns | integer | 100 | 資料庫連接池中最大開啟連接數。 |
| database.maxIdleConns | integer | 10 | 資料庫連接池中最大空閒連接數。 |
| database.connMaxLifetime | string | 1h | 資料庫連接的最大生命週期。例如 1h, 30m。 |
| security.jwtSecretEnvVar | string | APP_JWT_SECRET | 環境變數名稱，用於獲取 JWT 簽名所需的秘密金鑰。實際值應從環境變數或 Secrets Provider 中獲取，**不應硬編碼**。 |
| security.csrfTokenLifeTime | string | 1h | CSRF Token 的生命週期。 |

## **3. composition.yaml - 插件組合與配置**

composition.yaml 是 Detectviz 平台的**核心配置檔案**，它定義了哪些插件會被載入、它們的類型、名稱以及它們各自的運行時配置。

**路徑**：configs/composition.yaml

**對應 Schema**：schemas/composition.json

**範例結構**：

plugins:  
  - type: http_server_provider  
    name: mainHttpServer  
    config:  
      port: 8080  
      readTimeout: 5s  
      writeTimeout: 10s  
  - type: otelzap_logger_provider  
    name: defaultLogger  
    config:  
      level: debug  
      encoding: console  
      outputPaths:  
        - stdout  
      errorOutputPaths:  
        - stderr  
      initialFields:  
        service: detectviz-platform  
  - type: gorm_mysql_client_provider  
    name: primaryDBClient  
    config:  
      dsn: "user:password@tcp(127.0.0.1:3306)/detectviz?parseTime=true"  
      maxOpenConns: 100  
      maxIdleConns: 10  
      connMaxLifetime: 1h  
  - type: importer_plugin  
    name: csvImporter  
    config:  
      sourceType: csv  
      sourceConfig:  
        filePath: /data/input.csv  
        delimiter: ","  
        hasHeader: true

**主要配置項說明**：

| 配置項 | 類型 | 說明 |
| :---- | :---- | :---- |
| plugins | array of object | 包含所有要載入的插件定義的列表。每個對象代表一個插件實例。 |
| plugins[].type | string | **必填**。插件的類型字符串，用於識別插件的種類並找到對應的插件工廠和 Schema。例如 http_server_provider, importer_plugin。 |
| plugins[].name | string | **必填**。插件實例的唯一名稱。在平台內部用於引用和查找特定的插件實例。 |
| plugins[].config | object | **必填**。該插件實例的特定配置。其結構會根據 plugins[].type 而變化，並由對應的插件 Schema 進行驗證。 |

## **4. 插件特定配置 (plugins[].config)**

plugins[].config 部分是每個插件實例的特定配置。這些配置的結構是高度動態的，並由位於 schemas/plugins/ 目錄下的對應 JSON Schema 檔案精確定義。

### **4.1 JSON Schema - 配置的權威定義**

每個插件類型（例如 http_server_provider 對應 schemas/plugins/http_server_provider.json）都有一個專屬的 JSON Schema 檔案，用於：

* **定義資料類型**：指定每個配置項的 type (例如 string, integer, boolean, array, object)。  
* **必填項**：通過 required 關鍵字指定哪些配置項是必須提供的。  
* **預設值**：通過 default 關鍵字提供配置項的預設值。  
* **約束**：  
  * enum：定義允許的枚舉值列表。  
  * pattern：對於字符串類型，定義正則表達式模式。  
  * minimum/maximum：對於數字類型，定義數值範圍。  
  * minItems/maxItems：對於數組類型，定義元素數量範圍。  
  * properties：定義對象中允許的屬性及其 Schema。  
  * description：提供配置項的詳細說明。

**Go 程式碼中的 Config 結構體**：

在 Go 程式碼中，每個插件實現的 Config 結構體應與其 JSON Schema 定義保持一致。應使用 yaml:"fieldName" 標籤來指定 YAML 欄位名，並為每個配置欄位提供詳細的 GoDoc 註解，這有助於自動生成文檔和 AI 理解。

### **4.2 常見插件配置項範例**

以下列出部分常見插件類型及其典型的配置項。詳細且完整的配置請參考 schemas/plugins/ 目錄下的 JSON Schema 檔案。

#### **4.2.1 http_server_provider**

* **類型字符串**：http_server_provider  
* **對應 Schema**：schemas/plugins/http_server_provider.json  
* **說明**：提供 HTTP 服務，處理路由和請求。  
* **配置範例**：  
  config:  
    port: 8080  
    readTimeout: 5s  
    writeTimeout: 10s

#### **4.2.2 otelzap_logger_provider**

* **類型字符串**：otelzap_logger_provider  
* **對應 Schema**：schemas/plugins/otelzap_logger_provider.json  
* **說明**：提供統一的日誌記錄功能，並可集成 OpenTelemetry。  
* **配置範例**：  
  config:  
    level: info # 日誌級別：debug, info, warn, error 等  
    encoding: json # 日誌格式：json 或 console  
    outputPaths: ["stdout"] # 日誌輸出目標  
    errorOutputPaths: ["stderr"] # 錯誤日誌輸出目標  
    initialFields: # 附加到所有日誌條目的初始字段  
      service: detectviz-platform

#### **4.2.3 gorm_mysql_client_provider**

* **類型字符串**：gorm_mysql_client_provider  
* **對應 Schema**：schemas/plugins/gorm_mysql_client_provider.json  
* **說明**：提供基於 GORM 的 MySQL 資料庫連接池與操作介面。  
* **配置範例**：  
  config:  
    dsn: "user:password@tcp(127.0.0.1:3306)/detectviz?parseTime=true"  
    maxOpenConns: 100  
    maxIdleConns: 10  
    connMaxLifetime: 1h

#### **4.2.4 keycloak_auth_provider**

* **類型字符串**：keycloak_auth_provider  
* **對應 Schema**：schemas/plugins/keycloak_auth_provider.json  
* **說明**：處理用戶身份驗證和 JWT 驗證，集成 Keycloak。  
* **配置範例**：  
  config:  
    url: "http://localhost:8080/auth" # Keycloak 認證服務地址  
    realm: "detectviz" # Keycloak Realm 名稱  
    clientId: "detectviz-client" # 在 Keycloak 中註冊的客戶端 ID  
    clientSecretEnvVar: "KEYCLOAK_CLIENT_SECRET" # 存放客戶端秘密的環境變數名稱  
    # 其他 Keycloak 相關配置，例如 JWKS URL, token 驗證選項等

#### **4.2.5 llm_provider**

* **類型字符串**：llm_provider  
* **對應 Schema**：schemas/plugins/llm_provider.json  
* **說明**：提供大型語言模型推理功能（如文本生成）。  
* **配置範例**：  
  config:  
    modelName: "gemini-1.5-pro" # 使用的 LLM 模型名稱  
    apiKeyEnvVar: "GEMINI_API_KEY" # 存放 LLM API Key 的環境變數名稱  
    temperature: 0.7 # 模型生成文本的隨機性 (0.0 - 1.0)  
    maxTokens: 1024 # 生成文本的最大令牌數

#### **4.2.6 importer_plugin**

* **類型字符串**：importer_plugin  
* **對應 Schema**：schemas/plugins/importer_plugin.json  
* **說明**：負責從外部數據源導入數據。  
* **配置範例**：  
  config:  
    sourceType: csv # 數據源類型 (e.g., csv, api, database)  
    sourceConfig: # 具體數據源的配置，依 sourceType 而定  
      filePath: "/data/input.csv"  
      delimiter: ","  
      hasHeader: true

#### **4.2.7 detector_plugin**

* **類型字符串**：detector_plugin  
* **對應 Schema**：schemas/plugins/detector_plugin.json  
* **說明**：執行數據異常偵測或模式識別。  
* **配置範例**：  
  config:  
    model: isolation_forest # 偵測模型 (e.g., isolation_forest, rnn, deep_learning)  
    threshold: 0.5 # 異常分數閾值  
    parameters: # 模型的具體參數  
      n_estimators: 100  
      contamination: auto

#### **4.2.8 ui_page_plugin**

* **類型字符串**：ui_page_plugin  
* **對應 Schema**：schemas/plugins/ui_page_plugin.json  
* **說明**：動態註冊新的前端 UI 頁面或組件。  
* **配置範例**：  
  config:  
    routePath: "/dashboard/metrics" # UI 頁面或組件對應的路由路徑  
    templateName: "metrics-dashboard" # 後端渲染的模板名稱  
    jsBundlePath: "/static/js/metrics-dashboard.bundle.js" # 前端 JavaScript 資源路徑

## **5. 配置最佳實踐**

* **環境變數**：敏感資訊（如 API 金鑰、資料庫密碼）應從環境變數或專用的 Secrets Provider 中獲取，**嚴禁硬編碼在配置檔案中**。  
* **版本控制**：所有配置檔案（.yaml 和 .json Schema）都應納入版本控制系統。  
* **命名慣例**：遵循一致的命名慣例，使配置項易於理解和查找。  
* **註解**：在配置檔案中添加清晰的註解，解釋複雜或非顯而易見的配置項。  
* **分層配置**：利用 Viper 的多來源配置能力，實現不同環境（開發、測試、生產）的分層配置。例如，基本配置在 app_config.yaml，環境特定覆蓋則通過環境變數或特定環境的配置檔案載入。

本指南將作為 Detectviz 平台配置的核心參考文件，並會根據插件的增長和功能的完善持續更新。

