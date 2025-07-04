


# RAG Ingest Plan for Detectviz Platform

本文件定義 Detectviz Platform 導入 RAG（Retrieval-Augmented Generation）所需的文件結構、收錄來源、更新策略與資料標準格式，作為 AI Scaffold 輔助開發的基礎知識索引。

---

## RAG 資料來源分類

| 類別 | 說明 | 範例路徑 |
|------|------|----------|
| Interface 定義 | 所有 `*.go` 中的 interface 說明與用途註解 | `pkg/domain/plugins/*.go`, `pkg/platform/contracts/*.go` |
| Plugin 實作 | 每個 plugin 的工廠函式與註解 | `internal/domain_logic/plugins/**/factory.go` |
| Plugin 說明文件 | Markdown 格式的 plugin 說明與範例 | `docs/plugins/plugin-*.md` |
| 配置檔 | 組裝平台的 YAML 結構 | `configs/composition.yaml`, `configs/app_config.yaml` |
| Schema 驗證 | 用於 plugin 配置校驗的 JSON Schema | `schemas/plugins/*.json` |
| Scaffold Workflow | AI scaffold 建構邏輯與語意標籤規範 | `docs/ai_scaffold/scaffold_workflow.md` |
| 核心架構文檔 | Clean Architecture 說明與平台總覽 | `ARCHITECTURE.md`, `ENGINEERING_SPEC.md` |

---

## 📁 建議資料目錄與分類

```
rag_index/
├── interfaces/
│   ├── plugin_ui_page.md
│   ├── config_provider.md
├── plugins/
│   ├── mysql_importer.md
│   ├── llm_analysis_engine.md
├── examples/
│   ├── plugin_factory_example.go
│   ├── plugin_schema_example.json
├── configs/
│   ├── composition.yaml
│   ├── app_config.yaml
├── core_docs/
│   ├── scaffold_workflow.md
│   ├── architecture.md
```

---

## 資料更新策略

| 資料類型 | 更新頻率 | 更新方式 |
|----------|----------|----------|
| Go interface 註解 | 每次 PR 提交前自動擷取 | 使用 script + `golang.org/x/tools/go/packages` |
| Plugin 實作 & 說明 | 每次 scaffold 時一併產生 | Scaffold generator 自動產生對應 MD |
| JSON Schema | 每次變更 plugin 配置時同步更新 | 應與 plugin 建構 script 綁定 |
| 核心文檔 | 版本里程碑前手動審核更新 | 維護者人工審查後更新 RAG |

---

## Index 建議標準格式（供嵌入器使用）

```json
{
  "source_path": "docs/plugins/plugin-mysql_importer.md",
  "type": "plugin_doc",
  "plugin_type": "importer",
  "interface": "ImporterPlugin",
  "embedding_tags": ["plugin", "importer", "factory", "config_schema"],
  "content": "..."
}
```

---

## 未來可擴展方向

- 導入 langchain-go 或 chroma 作為嵌入與查詢工具
- 支援 RAG 增強的開發者 Q&A 工作台
- 與 `plugin_metadata.yaml` 自動對應，用於查詢版本、範例、依賴等資訊
