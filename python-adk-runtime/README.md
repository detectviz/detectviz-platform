# python-adk-runtime（ADK Runtime）

以 **Google Agent Development Kit（ADK）** 為核心的 Python 執行環境。此 Runtime 嚴格遵循 ADK 模組邊界，並與 `go-platform` 透過 **gRPC ToolBridge** 解耦互通。

---

## 平台定位
- 對齊 ADK 模組邊界：Agent／Workflow／MemoryBank／Tools／Capabilities。
- 透過 **RemoteTool** 以 gRPC 呼叫 `go-platform` 插件，避免跨語言耦合。
- 支援多 Agent 協作（A2A）、共享記憶體 bank 抽換、版本化與依賴描述（Module Card）。
- 以 **SSOT（contracts）** 為唯一事實來源：proto／schema／samples 必須先於 `contracts/` 更新，兩端再同步。

---

## 目錄結構
- `src/detectviz_adk/config/loader.py`：統一設定載入（搜尋序／環境覆蓋／Schema 驗證）
- `src/detectviz_adk/tools/remote_tool.py`：遠端工具呼叫（gRPC → ToolBridge，支援 OTel Context）
- `src/detectviz_adk/memory/`：MemoryBank 介面與多後端實作
- `agents/`：範例與實作（含 coordinator／tool-exec）
- `templates/`：最小樣板（agent／tool／capability／workflow），含 `module.card.json` 範本

---

## 與 contracts 的對齊
- 模組卡 Schema：`contracts/schemas/module.card.schema.json`
  - `role` 範例：`agent.coordinator`、`agent.tool_exec`、`tool`、`capability`、`memory.backend`、`plugin.gateway`、`observability.module` 等
  - `category` 參考：`workflow`、`a2a`、`llm`、`retriever`、`capability`、`memory.backend`、`gateway` 等
- gRPC 生成碼：`contracts/gen/python/detectviz/contracts/v1`
- 組態樣本（SSOT 實例）：`contracts/samples/config.yaml`

> 原則：任何跨語言協議或分類調整，先改 `contracts/`，再同步此專案與 `go-platform`。

---

## 安裝與需求
- Python 3.11+
- 建議使用 `uv` 或 `poetry` 管理依賴
- 需要可讀取 `contracts/`（便於載入 Schema 與 gRPC 生成碼）

```bash
# 以 uv 為例
uv venv
uv pip install -e .

# 產生 gRPC 生成碼（於 repo 根目錄）
cd contracts
make gen
```

---

## 設定載入（統一搜尋順序）
**優先序（高 → 低）**：
1. 函式參數或 CLI：`--config /path/to/config.yaml`
2. 環境變數：`DETECTVIZ_CONFIG_FILE=/path/to/config.yaml`
3. 工作目錄：`./config.yaml`
4. 團隊覆蓋：`./contracts/config.yaml`
5. SSOT 樣本兜底：`./contracts/samples/config.yaml`

Python 與 Go 採用同一套鍵位與環境覆蓋規範（部分鍵位）：
- `DETECTVIZ_ENV` → `env`
- `DETECTVIZ__OBSERVABILITY__MODE` → `observability.mode`
- `DETECTVIZ__OBSERVABILITY__OTLP__{PROTOCOL,ENDPOINT,INSECURE,HEADERS}`
- `DETECTVIZ__OBSERVABILITY__RESOURCE__{SERVICE_NAME,SERVICE_VERSION,ENVIRONMENT}`
- `DETECTVIZ__OBSERVABILITY__SAMPLING__RATIO`
- `DETECTVIZ__PLUGIN__{PATHS,REGISTRY}`
- `DETECTVIZ__MEMORY__{BACKEND,DSN,DEFAULT_TTL_SECONDS}`

> Python 端多數觀測輸出委由 `go-platform`+Alloy 完成；此處主要傳遞 OTel Span Context。

---

## RemoteTool 使用指引
`RemoteTool` 透過 `ToolBridge.Invoke` 呼叫 Go 端工具。位址以 `DETECTVIZ_TOOLBRIDGE_ADDR` 設定（預設 `127.0.0.1:5002`）。

```python
import asyncio
from detectviz_adk.tools.remote_tool import RemoteTool

async def main():
    tool = RemoteTool(tool_id="detectviz.tools.http_request", tool_version="0.1.0")
    res = await tool.invoke({
        "method": "GET",
        "url": "https://example.com",
        # 可選：headers/query/json/form/body/timeout_ms/max_response_bytes
    })
    print(res)

if __name__ == "__main__":
    asyncio.run(main())
```

### 連線與安全
- 端點：`DETECTVIZ_TOOLBRIDGE_ADDR`（預設 127.0.0.1:5002）
- 明文：`DETECTVIZ_TOOLBRIDGE_INSECURE=true`
- TLS／mTLS：`DETECTVIZ_TOOLBRIDGE_TLS_{CERT,KEY,CA}`
- OTel：若安裝 OpenTelemetry，`RemoteTool` 會嘗試自動注入 `traceparent`／`tracestate`

---

## Tools 與 Capabilities 的區分
- **Tools**：對外部系統的具體操作介面（HTTP／gRPC／DB／Shell 等）。`RemoteTool` 可橋接到 Go 插件。
- **Capabilities**：可複用的能力單元（模型、檢索、規則、策略、資料連接等），不直接握外部副作用。

此分離帶來：
- 能力可重用與測試更單純（無 I/O 依賴）。
- 工具權限最小化與審計清晰（外部操作集中於 Tools）。

---

## 樣板與擴增流程
1. 於 `templates/` 選擇樣板（`agent.coordinator`／`agent.tool_exec`／`tool`／`capability`／`workflow`）。
2. 複製至對應子目錄並填寫 `module.card.json`：
   - `name`、`version`（SemVer）、`role`、`category`
   - `requires`（依賴名稱／版本規則）、`contracts.min_proto`
3. 以 `contracts/tools/validate_module_card.py` 驗證模組卡。
4. 在 Agent 中以 RemoteTool 或 ADK 原生工具拼裝工作流；必要時設定 MemoryBank 命名空間與存取策略。

---

## 記憶體（MemoryBank）
- 抽象介面支援 In-Memory、Redis、向量庫（Weaviate／Chroma／Milvus）或雲端（Vertex）。
- 允許多個 root_agent 共用 sub_agent 與工具：以命名空間與授權策略控管讀寫與可見性。
- MemoryBank 後端替換不影響上層 Agent／Workflow（遵循 ADK 的 Memory 介面）。

---

## 觀測（Observability）
- Python 端主要傳遞 OTel Trace Context 至 ToolBridge；
- Logs／Metrics／Profiling 由 Go 端與 Alloy 匯聚上送 Grafana Cloud，可從 Logs Drilldown 至 Traces／Profiles。

---

## 執行與測試（端到端）
1. 啟動 `go-platform`（ToolBridge + http-demo），並確認 Alloy 已啟動。
2. 在此專案中執行你的 Agent（可參考 `agents/` 範例）。
3. 於 Grafana Cloud 檢視 Logs／Traces／Profiles；確認 Drilldown 正常。

---

## 安全準則
- 禁止在程式碼或樣板中硬編密鑰／Token；請使用環境變數或 Secret。
- 外部服務憑證優先透過 Go 插件或 Alloy 端管理；Python 僅持有必要資訊（例如 ToolBridge 位址）。

---

## 常見問題
- 無法呼叫 ToolBridge：確認 `go-platform` 已啟動且 `DETECTVIZ_TOOLBRIDGE_ADDR` 正確。
- Proto 生成碼缺失：於 `contracts/` 執行 `make gen`，並確保 Python 可匯入 `contracts/gen/python/...`。
- Drilldown 失敗：確認是否正確傳遞 `traceparent`／`tracestate` 至 RemoteTool。

---

## 參考
- `../spec.md`（整體平台規格）
- `../contracts/`（SSOT：proto／schema／samples）
- `../go-platform/`（ToolBridge 與觀測初始化）
