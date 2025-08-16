# 貢獻指南（Detectviz Platform）

歡迎貢獻！提交變更前，請遵循以下原則與檢查清單。

## 基本原則
- **SSOT 契約優先**：跨語言介面或設定變更，請先修改 `contracts/`（proto/schema/samples），再更新下游（Go/Python/Docs）。
- **文件同步**：任何影響使用者或行為的變更，必須同步更新對應文件（`detectviz-docs/`）。
- **安全**：禁止提交真實密鑰/Token；請使用環境變數或 Secret。

## llm 維護指南（必讀）
- 文件站：`detectviz-docs/llms.txt`
- 範例：`detectviz-examples/llms.txt`
- Go 平台：`go-platform/llm.txt`
- Python Runtime：`python-adk-runtime/llm.txt`
- 契約（SSOT）：`contracts/llm.txt`

> 提交前請勾選 PR 模板中的「已遵循 llm 指南」核取方塊。

## 提交前檢查清單（摘要）
- [ ] 若變更觸及導覽或文件內容，已更新 `detectviz-docs/` 並通過 `mkdocs build --strict`
- [ ] 若變更觸及契約，已在 `contracts/` 執行 `buf lint && buf generate` 與驗證腳本
- [ ] 若變更觸及 Go/Python，已對齊生成碼與設定載入，端到端測過（含健康檢查與觀測）
- [ ] 變更摘要、風險、驗證方式已寫入 PR 描述
