# AGENTS.md

Detectviz 是一個基於 Clean Architecture 的可組合異常偵測平台，採用「一切皆插件」的設計理念，支援擴展與整合。平台通過 Registry、Lifecycle 與 Composition 系統實現插件的動態載入和配置管理。

本文件提供 AI 代理在開發與驗證過程中所需的架構說明與檢查指引，專注於問題診斷、插件協調與系統整合。

---

## AI 代理檢查任務說明

當 AI 代理接收到「請檢查 ROADMAP.md 里程碑完成狀況」或類似任務時，請依下列步驟執行：

### 檢查流程

1. **解析里程碑定義**：逐項分析 `docs/ROADMAP.md` 中定義的具體任務要求。
2. **驗證實現文件**：檢查對應的實作檔案、測試檔案和文檔是否存在且符合要求。
3. **架構合規性檢查**：確認實現是否遵循 Clean Architecture 原則和平台設計規範。
4. **功能完整性驗證**：測試核心功能是否可正常運行。

### 檢查範圍

**里程碑 0.1 - 基礎 Go 專案結構與核心組件介面**
- [ ] Go Modules 專案結構 (`/cmd`, `/internal`, `/pkg`, `/configs`, `/docs`)
- [ ] 平台契約介面定義 (`pkg/platform/contracts/`)
- [ ] 領域層插件介面 (`pkg/domain/plugins/`)
- [ ] 實體定義 (`pkg/domain/entities/`)

**里程碑 0.2 - 配置管理與 JSON Schema 導入**
- [ ] ConfigProvider 實現 (`internal/infrastructure/platform/config/`)
- [ ] JSON Schema 定義 (`schemas/app_config.json`, `schemas/composition.json`)
- [ ] 插件 Schema 定義 (`schemas/plugins/*.json`)
- [ ] 配置驗證機制
- [ ] 文檔更新狀況

**里程碑 0.3 - 最小可行插件註冊與組裝**
- [ ] PluginRegistryProvider 實現 (`internal/infrastructure/platform/registry/`)
- [ ] HttpServerProvider 實現 (`internal/infrastructure/platform/http_server/`)
- [ ] Logger 實現 (`internal/infrastructure/platform/logger/`)
- [ ] UI 插件範例 (`internal/adapters/plugins/web_ui/`)
- [ ] 主程序組裝邏輯 (`cmd/api/main.go`)

### 檢查結果分類

- ✅ **已完成**：功能實現完整，測試通過，文檔齊全
- ⚠️ **部分完成**：核心功能實現但缺少測試、文檔或驗證機制
- ❌ **未完成**：功能未實現或實現不完整

---

## Detectviz 核心架構（Clean Architecture）

```
detectviz-platform/
├── cmd/                    # 應用程式入口點
│   └── api/               # HTTP API 主程序
├── pkg/                   # 公共程式碼，可被外部引用
│   ├── domain/            # 領域層（核心業務邏輯）
│   │   ├── entities/      # 領域實體
│   │   ├── interfaces/    # 領域層介面
│   │   ├── plugins/       # 插件領域介面
│   │   └── errors/        # 領域錯誤定義
│   └── platform/          # 平台契約層
│       └── contracts/     # 平台核心契約介面
├── internal/              # 內部實現，外部不可引用
│   ├── infrastructure/    # 基礎設施層
│   │   └── platform/     # 平台基礎設施實現
│   ├── adapters/         # 適配器層
│   │   └── plugins/      # 插件適配器實現
│   ├── services/         # 應用服務層
│   └── repositories/     # 數據倉庫層
├── configs/              # 配置文件
├── schemas/              # JSON Schema 定義
│   └── plugins/         # 插件配置 Schema
└── docs/                # 文檔
```

### 分層職責

**領域層 (`pkg/domain/`)**
- 包含核心業務邏輯和規則
- 定義實體、值對象和領域服務
- 不依賴任何外部框架或基礎設施

**平台契約層 (`pkg/platform/contracts/`)**
- 定義平台級服務的介面契約
- 規範插件和平台組件的交互方式
- 支持依賴反轉原則

**基礎設施層 (`internal/infrastructure/`)**
- 實現平台契約介面
- 包含具體的技術實現（HTTP、數據庫、日誌等）
- 提供平台核心能力

**適配器層 (`internal/adapters/`)**
- 實現外部系統的適配
- 將外部請求轉換為領域操作
- 包含插件的具體實現

---

## 插件系統架構

### 插件類型

| 類型 | 介面定義位置 | 功能說明 | 實現範例 |
|------|-------------|----------|----------|
| **平台提供者** | `pkg/platform/contracts/` | 核心平台服務 | Logger, Config, Registry |
| **領域插件** | `pkg/domain/plugins/` | 業務功能插件 | Importer, UIPage |
| **適配器插件** | `internal/adapters/plugins/` | 外部系統適配 | Web UI, Auth |

### 插件生命週期

```go
type Plugin interface {
    GetName() string
    Init(ctx context.Context, cfg map[string]interface{}) error
    Start(ctx context.Context) error
    Stop(ctx context.Context) error
}
```

### 插件註冊流程

1. **定義介面**：在適當的層級定義插件介面
2. **實現插件**：創建具體的插件實現
3. **註冊到 Registry**：通過 PluginRegistryProvider 註冊
4. **配置管理**：在 `composition.yaml` 中配置插件參數
5. **生命週期管理**：由平台自動管理插件的啟動和停止

---

## 配置管理系統

### 配置文件結構

**主配置 (`configs/app_config.yaml`)**
- 全局應用程式設定
- 服務器、數據庫、安全等基礎配置

**組合配置 (`configs/composition.yaml`)**
- 插件組合定義
- 每個插件實例的具體配置

### JSON Schema 驗證

**Schema 位置**
- `schemas/app_config.json` - 主配置 Schema
- `schemas/composition.json` - 組合配置 Schema  
- `schemas/plugins/*.json` - 各插件配置 Schema

**驗證流程**
1. 配置載入時進行 Schema 驗證
2. 確保配置格式和數據類型正確
3. 提供清晰的錯誤訊息

---

## 系統整合與診斷重點

### Registry 系統

**核心實現**：`internal/infrastructure/platform/registry/plugin_registry_provider.go`

**主要功能**：
- 插件實例註冊和查詢
- 插件元數據管理
- 線程安全的並發存取
- 插件生命週期協調

### HTTP 服務系統

**核心實現**：`internal/infrastructure/platform/http_server/echo_http_server_provider.go`

**內建端點**：
- `/health` - 健康檢查
- `/api/v1/info` - 平台資訊
- 插件自定義路由

### 日誌系統

**核心實現**：`internal/infrastructure/platform/logger/otelzap_logger.go`

**特點**：
- 結構化日誌輸出
- OpenTelemetry 整合準備
- 可配置的日誌級別和輸出

---

## 開發與測試指引

### 插件開發流程

1. **定義介面**：在對應層級定義插件介面
2. **實現插件**：創建具體實現，遵循介面規範
3. **編寫測試**：單元測試和整合測試
4. **創建 Schema**：定義配置的 JSON Schema
5. **更新組合配置**：在 `composition.yaml` 中添加插件配置
6. **註冊插件**：在主程序中註冊並啟動

### 測試策略

**單元測試**
- 測試個別組件的功能
- 模擬外部依賴
- 快速反馈循環

**整合測試**
- 測試組件間的交互
- 驗證配置載入和插件註冊
- 端到端的功能驗證

### 品質檢查清單

- [ ] 代碼遵循 Clean Architecture 原則
- [ ] 所有公共介面有完整的 GoDoc 註解
- [ ] 配置有對應的 JSON Schema 定義
- [ ] 實現了適當的錯誤處理和日誌記錄
- [ ] 單元測試覆蓋率達標
- [ ] 整合測試驗證主要功能
- [ ] 性能和安全性考量

---

## AI 協作指引

### 代碼生成

當需要 AI 協助生成代碼時，請提供：
1. **上下文資訊**：當前的架構層級和相關介面
2. **具體需求**：功能描述和預期行為
3. **約束條件**：性能、安全、相容性要求
4. **測試要求**：預期的測試覆蓋範圍

### 配置生成

AI 可協助生成：
- JSON Schema 定義
- 配置文件範例
- 驗證邏輯代碼
- 文檔說明

### 文檔更新

AI 應協助維護：
- API 文檔的即時更新
- 架構決策記錄 (ADR)
- 使用者指南和開發者文檔
- 故障排除指南

---

此文件將隨著平台發展持續更新，以反映最新的架構設計和開發實踐。