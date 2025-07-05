# DetectViz AI Scaffold 結構導覽與說明

本說明文檔目的為協助開發者與 AI 工具（如 Cursor）在執行 Scaffold、重構與 Plugin 擴展時，能快速掌握目前專案的目錄結構與用途。所有路徑均已優化以配合 AI 自動化生成、對應與檔案定位。

---

## 根目錄

| 路徑                                   | 說明                       |
| ------------------------------------ | ------------------------ |
| `README.md`                          | 專案總覽說明                   |
| `todo.md`                            | Scaffold 任務管理，由 AI 或人工補充 |
| `go.mod`, `go.sum`                   | Go module 設定             |
| `Dockerfile`, `docker-compose.*.yml` | 部署設定與 Observability 套件整合 |
| `LICENSE`                            | 專案授權條款                   |

---

## AI Scaffold 專用文檔 `docs/ai_scaffold/`

| 檔案                          | 功能說明                                     |
| --------------------------- | ---------------------------------------- |
| `scaffold_prerequisites.md` | 定義 Scaffold 任務的基本規則與輸出要求                 |
| `cursor_prompt.md`          | 給 Cursor 使用的主 prompt，支援 @tree、@todo 指令語法 |
| `codex_prompt.md`           | 提供 AI 代碼審查與最佳化規則指引                       |
| `scaffold_workflow.md`      | 描述 Scaffold 的 AI 自動流程                    |
| `main_go_assembly.tmpl`     | 範本：主程式組裝模版，用於初始化 AI scaffold             |
| `rag_ingest_plan.md`        | RAG 文件資料注入策略與檔案規範說明                      |

---

## Plugin 文件 `docs/plugins/` + `schemas/plugins/`

| 類型          | 命名規範                      | 說明                                       |
| ----------- | ------------------------- | ---------------------------------------- |
| Plugin 文檔   | `plugin-{type}_{impl}.md` | 定義每個 plugin 的用途、介面對應與範例                  |
| Schema JSON | `{type}_{impl}.json`      | 對應 plugin 的配置 schema，供驗證與 AI scaffold 使用 |

例如：

- `plugin-detector_threshold.md` ↔ `detector_threshold.json`
- `plugin-importer_csv.md` ↔ `importer_csv.json`

---

## 專案邏輯結構說明

### `pkg/`

- `application/shared/`：放置共用層的 `DTO` 與 `Mapper`
- `common/utils/`：工具函式，AI 可直接重用
- `domain/`：Clean Architecture 的核心領域層（Entities、Errors、Interfaces）
- `platform/contracts/`：平台抽象契約（ConfigProvider, LoggerProvider 等）

### `internal/`

- `adapters/web/`：UI 與 Web handler，包含 Echo router handler 與 UI 元件頁面
- `application/`：應用邏輯實作層（與 `pkg/domain/interfaces` 對應）
- `bootstrap/`：初始化與平台配置載入（composition.yaml）
- `infrastructure/platform/`：基礎設施 Providers 的實作（如 Prometheus、Jaeger）
- `plugins/`：內建 plugin 實作，每類一目錄（如 `importers/`, `detectors/`）
- `repositories/`：資料儲存實作層（如 `mysql/user_repository.go`）
- `testdata/`：開發與測試使用的固定資料集

### `configs/` + `schemas/`

- `app_config.yaml`, `composition.yaml`：AI scaffold 初始化的主要設定來源
- `schemas/plugins/`：所有 plugin 的設定結構（與 plugin interface 對應）

---

## AI Scaffold 推薦結構設計原則（已實作）

| 原則              | 舉例                               | 說明                      |
| --------------- | -------------------------------- | ----------------------- |
| Plugin 統一命名     | `detector_threshold`             | 可對應 schema、doc、實作與 test |
| Interface 扁平化   | `interfaces/detector_service.go` | AI 不需理解多層目錄上下文          |
| DTO + Mapper 聚合 | `application/shared/`            | 避免分層過細、易於 scaffold 對應   |
| Utility 抽出      | `common/utils/`                  | 提供一致的共用邏輯               |
| Web UI 聚合       | `adapters/web/`                  | 所有 UI handler 與組件集中管理   |

---

## 輔助審查與品質追蹤文件

| 檔案                                      | 用途                  |
| --------------------------------------- | ------------------- |
| `interface_spec.md`                     | 所有介面定義規則與審查指引       |
| `interface_review_report.md`            | 最近一次介面審查結果（由 AI 生成） |
| `architecture_simplification_report.md` | 架構簡化與扁平化優化記錄        |
| `architecture_optimization_report.md`   | 架構重組與分層清理的策略與成果     |
| `checklist.md`                          | Scaffold 過程所有項目追蹤清單 |

---

## 最佳實踐建議

- 將 Plugin 命名、Schema、實作、測試、文檔使用統一命名規則
- 控制層級深度在 3 層以內，避免複雜目錄（`x/y/z/a/b.go`）
- 接口檔案獨立，不使用合併式 fat interface（單一 .go 管多類）
- 每層資料都可由 `AI_PLUGIN_TYPE`、`@tree`、`@todo` 推導出位置與用途

---

若有新目錄加入，請務必更新本文件與 `scaffold_prerequisites.md`，以確保 AI Scaffold 能正確對應所有檔案路徑與職責。

