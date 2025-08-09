
## python-adk-runtime（ADK Runtime）

以 **Google Agent Development Kit（ADK）** 為核心的 Python 執行環境。此 Runtime 嚴格遵循 ADK 模組邊界，並與 `go-platform` 以 **gRPC ToolBridge** 解耦互通。

### 平台定位
- 嚴格遵循 ADK 模組邊界（Agent / Workflow / MemoryBank / Tools / Capabilities）。
- 以 **RemoteTool** 透過 gRPC 呼叫 `go-platform` 的插件，避免跨語言耦合。
- 支援多 Agent 協作（A2A）、共享記憶體 bank 抽換、版本化與依賴描述（Module Card）。
- 設計以 **SSOT（contracts）** 為唯一事實來源：proto / schema / samples 必須先於 `contracts/` 更新。

---

## 目錄結構
- `src/detectviz_adk/runtime/`：Runtime 啟動、協作與決策路由
- `src/detectviz_adk/tools/remote_tool.py`：遠端工具呼叫（gRPC → ToolBridge）
- `src/detectviz_adk/memory/`：MemoryBank 介面與多後端實作
- `agents/`：範例與實作（含 coordinator / tool-exec）
- `templates/`：最小樣板（agent / tool / capability / workflow），含 `module.card.json` 範本

---

## 與 contracts 的對齊
- 模組卡 Schema：`contracts/schemas/module.card.schema.json`
  - `role` 範例：`agent.coordinator`、`agent.tool_exec`、`tool`、`capability`、`memory.backend`、`plugin.gateway`、`observability.module` 等
  - `category` 參考：`workflow`、`a2a`、`llm`、`retriever`、`capability`、`memory.backend`、`gateway` 等
- gRPC 生成碼：`contracts/gen/python/detectviz/contracts/v1`
- 組態樣本（供 Go 使用）：`contracts/samples/config.yaml`（Python 端僅需 ToolBridge 位址）

> 原則：任何跨語言協議或分類調整，先改 `contracts/`，再同步此專案與 `go-platform`。

---

## 安裝與需求
- Python 3.11+
- 建議使用 `uv` 或 `poetry` 管理依賴
- 需要可讀取 `contracts/`（本專案相鄰路徑或已生成的 Python gRPC stub）

```bash
# 以 uv 為例
uv venv
uv pip install -e .
```

---

## RemoteTool 使用指引（最小概念範例）
RemoteTool 透過 gRPC 呼叫 `go-platform` 的 ToolBridge。範例示意如何串流接收工具輸出：

```python
from typing import Iterator
import grpc

# 生成碼：contracts/gen/python/detectviz/contracts/v1
from detectviz.contracts.v1 import adk_bridge_pb2 as pb
from detectviz.contracts.v1 import adk_bridge_pb2_grpc as pb_grpc

DEFAULT_BRIDGE_ADDR = "127.0.0.1:6606"

class RemoteToolClient:
    def __init__(self, address: str = DEFAULT_BRIDGE_ADDR):
        self._channel = grpc.insecure_channel(address)
        self._stub = pb_grpc.ToolBridgeServiceStub(self._channel)

    def execute(self, name: str, version: str, args: dict, metadata: dict | None = None) -> Iterator[pb.ToolChunk]:
        req = pb.ToolRequest(
            name=name,
            version=version,
            # args 建議以 JSON 字串傳遞，或依 proto 定義包裝
            args_json=json.dumps(args),
            metadata=metadata or {},
        )
        return self._stub.ExecuteTool(req)
```

在 Agent/Tool 中使用（概念）：

```python
from opentelemetry import trace

tracer = trace.get_tracer("detectviz_adk.remote_tool")

def call_http():
    client = RemoteToolClient()
    with tracer.start_as_current_span("tool.http_request"):
        chunks = client.execute(
            name="detectviz.tools.http_request",
            version="0.1.0",
            args={"method": "GET", "url": "https://example.com"},
        )
        for ch in chunks:
            # ch 可能分段回傳結果/日誌/進度
            process_chunk(ch)
```

> 注意：請保持工具名稱與版本對齊其 `module.card.json`；如需傳遞 trace context，請在 gRPC metadata 中附帶 W3C `traceparent`。

---

## Tools 與 Capabilities 的區分
- **Tools**：對外部系統的具體操作介面（HTTP、gRPC、DB、Shell 等）。RemoteTool 可橋接到 Go 插件。
- **Capabilities**：可複用的能力單元（模型、檢索、規則、策略、資料連接等），不直接握外部副作用。

此分離可確保：
- 能力可重用與測試更單純（無 I/O 依賴）。
- 工具權限最小化與審計更清晰（外部操作集中於 Tools）。

---

## 樣板與擴增流程
1. 於 `templates/` 選擇樣板（`agent.coordinator`、`agent.tool_exec`、`tool`、`capability`、`workflow`）。
2. 複製至對應子目錄並填寫 `module.card.json`：
   - `name`、`version`（SemVer）、`role`、`category`
   - `requires`（依賴名稱/版本規則）、`contracts.min_proto`
3. 以 `contracts/tools/validate_module_card.py` 驗證模組卡。
4. 在 Agent 中以 RemoteTool 或 ADK 原生工具拼裝工作流；必要時設定 MemoryBank 命名空間與存取策略。

---

## 記憶體（MemoryBank）
- 以抽象介面支援 In-Memory、Redis、向量庫（Weaviate/Chroma/Milvus）或雲端（Vertex）。
- 允許多個 root_agent 共用 sub_agent 與工具：透過命名空間與授權策略控制讀寫與可見性。
- MemoryBank 後端替換不影響上層 Agent/Workflow（遵循 ADK 的 Memory 介面）。

---

## 觀測（Observability）
- Python/ADK 端建議沿用 ADK 的 OpenTelemetry 整合，傳遞 span context 至 ToolBridge（確保 Logs/Traces Drilldown 一致）。
- Profiling：由 Go 端 `pprof` + Alloy `pyroscope.scrape/write` 統一上送 Grafana Cloud，Python 端無需持有雲端憑證。

---

## 執行與測試（端到端）
1. 啟動 `go-platform`（ToolBridge + http-demo），並確認 Alloy 已啟動。
2. 在此專案中執行你的 Agent（可參考 `agents/` 範例）。
3. 於 Grafana Cloud 檢視 Logs/Traces/Profiles；確保從 Logs 可 Drilldown 至對應 Trace 與 Profile。

---

## 安全準則
- 禁止在程式碼或樣板中硬編密鑰/Token；請使用環境變數或 Secret。
- 外部服務憑證優先透過 Go 插件或 Alloy 端管理；Python 端僅持有業務所需的最少資訊（例如 ToolBridge 位址）。

---

## 常見問題
- 無法呼叫 ToolBridge：請確認 `go-platform` 已啟動且位址（預設 `127.0.0.1:6606`）正確。
- Drilldown 失敗：檢查是否在 Python 端正確傳遞 trace context 至 RemoteTool（W3C TraceContext）。
- 模組卡驗證失敗：請對照 `contracts/schemas/module.card.schema.json` 的 `role/category` 與必填欄位。

---

## 參考
- `../spec.md`（整體平台規格）
- `../contracts/`（SSOT：proto / schema / samples）
- `../go-platform/`（ToolBridge 與觀測初始化）
