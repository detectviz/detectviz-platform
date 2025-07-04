# Scaffold Prerequisites

本文件定義在執行任何由 AI Scaffold 工具（如 Cursor）自動擴展、修改、重構任務前，**必須參考與同步維護**的關鍵資源。其目的為確保平台一致性、語意正確性與向量資料庫（RAG）可用性。

---

## 執行前必讀文件

AI 工具在根據 `todo.md` 或其他指令執行任何 scaffold 任務前，請務必先讀取以下內容：

| 類別 | 檔案 | 說明 |
|------|------|------|
| Scaffold 任務清單 | `/todo.md` | 優化指令與未完成項目來源，原為 `CODE_REVIEW_REPORT.md` |
| 介面契約與分類 | `/docs/architecture/interface_spec.md` | 所有 interface 定義與 plugin 類型分類依據 |
| Scaffold 流程規則 | `/docs/ai_scaffold/scaffold_workflow.md` | AI_PLUGIN_TYPE 等標籤定義與組裝流程說明 |
| 平台總覽 | `/README.md` | Plugin 架構摘要與範例路徑參照 |
| 任務狀態 | `/todo.md` | 每次 scaffold 後需更新的完成標記（不可略過） |
| 組態參考 | `/configs/app_config.yaml` | Plugin 設定範例與欄位格式參照 |
| JSON Schema | `/schemas/plugins/*.schema.json` | 每個 plugin 的配置結構規範 |
| Plugin 文件 | `/docs/plugins/plugin-*.md` | 各插件說明、範例與使用情境 |

---

## 執行後必須同步更新的檔案

任何 scaffold 任務執行後，請同步更新下列檔案（如有變更）：

- `/docs/architecture/interface_spec.md`
- `/docs/plugins/plugin-*.md`
- `/schemas/plugins/*.schema.json`
- `/todo.md`
- `/README.md`
- `/docs/ai_scaffold/rag_ingest_plan.md`（如新加入知識源）

---

## 注意事項

- `todo.md` 為浮動內容，AI 每次執行時應重新解析當前內容，不應硬編寫死值。
- 若缺少任一必要欄位或未標註 AI_PLUGIN_TYPE 等資訊，請先修補再 scaffold。
- 若需新增 plugin，請依照 plugin scaffold 範本補齊 `NewFactory()`、schema、plugin doc、interface 實作。

---

## Checklist 動態產生原則

- `checklist.md` 為浮動文件，應依照當前最新的 `todo.md` 自動產出。
- 每次 AI Scaffold 執行前，請先比對 `todo.md` 與現況，重新建立對應的 `/docs/audit/checklist.md`。
- checklist 內容應包含：
  - todo 中每個項目的具體檢查子任務（例如：plugin 實作、test、schema、doc）
  - 通用檢查項目（interface 註解、plugin metadata、RAG 支援等）
- checklist 項目須可由 AI 自動比對達成狀態與產出補充建議。