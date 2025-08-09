## Templates（Agent / Tool / Capability / Workflow）

本模板集提供 ADK 對齊的最小骨架，以便快速擴增。所有模板均要求：
- 提供 `module.card.json`，并遵循 `contracts/schemas/module.card.schema.json`
- 明確標注 `role` 與 `category`（已與 ADK × Telegraf 對齊的枚舉）
- 指定 `requires`（依賴與版本規則），利於平台檢核與未來升級

### 三種重用模式決策樹（摘要）
- 若僅需對外部系統操作：實作 **Tool**（可改用 **RemoteTool** 呼叫 go-platform 插件）
- 若為可組合能力（模型/檢索/規則）：實作 **Capability**
- 若需任務協調/分解：實作 **Agent**（`agent.tool_exec` 或 `agent.coordinator`）

### 模板清單
- `agent.tool_exec/`：單一職責的工具執行 Agent
- `agent.coordinator/`：多 Agent 協作與決策路由
- `tool/`：Python 端工具（無外部副作用時優先落於 capability）
- `capability/`：可重用能力模組
- `workflow/`：順序/分支/並行/循環流程樣板

### Module Card 規範（關鍵欄位）
- `name` / `version`（SemVer）/ `language`
- `role`：`agent.coordinator`｜`agent.tool_exec`｜`tool`｜`capability`｜`memory.backend`…
- `category`：`workflow`｜`a2a`｜`llm`｜`retriever`｜`capability`｜`storage.vector`…
- `entrypoint`：模組進入點（可為 Python 模組路徑）
- `requires`：依賴（含 `contracts.min_proto` 等）
- `observability.tags`：建議填寫 `service`/`component` 等利於查詢之欄位

### 生成與驗證
- 建立後請以 `contracts/tools/validate_module_card.py` 驗證
- 端到端測試建議對接 `RemoteTool`（Go 插件 `http_request`）以驗證協作路徑
