# Detectviz 平台 SPEC

本規格針對 Detectviz 以 Google Agent Development Kit（ADK）為核心的混合語言平台（Go + Python），定義目錄結構、跨語言契約、外部觀測、插件/Agent 擴增模式與最小可運行（MVP）組態。設計目標是讓 AI 與人類開發者依此文件即可產出品質穩定、可長期演進的程式碼。

* * *

## 1. 架構總覽

- **Platform Core（Go）**：最小化的平台層，負責 CLI、gRPC ToolBridge、Plugin Registry、觀測初始化、配置驗證。
- **Agent Runtime（Python/ADK）**：依 ADK 模組邊界實作 Agent、Workflow、MemoryBank、Tools/Capabilities，支援多 Agent 協作與 A2A。
- **Contracts（SSOT）**：以 Proto + JSON Schema 為單一事實來源（SSOT），所有語言的型別、驗證、與 API 皆由此生成。
- **Observability**：Alloy 為收集/轉發層；支援本地 LGTM、Grafana Cloud、GCP（Cloud Trace/Logging/Profiler）。
- **解耦準則**：Go 與 Python 僅以 gRPC/HTTP 溝通；支援 Local 或 Cloud 部署無需改動程式碼（以組態切換）。

* * *

## 2. 目錄結構（統一倉庫）

```bash
detectviz-platform/                 # 統一平台倉庫
├── contracts/                      # SSOT 契約
│   ├── proto/detectviz/contracts/v1/
│   │   └── adk_bridge.proto        # ToolBridge / Health APIs
│   ├── schemas/
│   │   ├── config.schema.json      # 平台組態規範
│   │   ├── module.card.schema.json # 模組卡規範
│   │   └── plugin.schema.json      # 插件規範
│   ├── gen/                        # 生成碼目錄
│   │   ├── go/detectviz/contracts/v1/
│   │   ├── python/detectviz/contracts/v1/
│   │   └── metadata/version.json   # 版本元數據
│   ├── samples/                    # 範例配置
│   ├── specs/                      # 技術規格
│   ├── tools/                      # 驗證工具
│   ├── buf.yaml                    # buf 工作區
│   ├── buf.gen.yaml               # 生成配置
│   └── Makefile                   # 版本控制與生成

├── go-platform/                   # Go 平台核心
│   ├── cmd/detectviz/main.go      # CLI 入口
│   ├── internal/
│   │   ├── configx/               # 配置載入與驗證
│   │   ├── contracts/             # 契約版本檢查
│   │   ├── health/                # 健康檢查服務
│   │   ├── observability/         # 日誌與 OTel 初始化
│   │   ├── pluginhost/           # 插件託管
│   │   │   ├── plugins/          # 插件實作目錄
│   │   │   │   └── capability.gateway/http_request/
│   │   │   │       ├── plugin.go
│   │   │   │       ├── security.go      # 安全邊界
│   │   │   │       └── secure_plugin.go # 安全增強版
│   │   │   ├── registry.go           # 插件註冊（並發安全）
│   │   │   ├── resource_monitor.go   # 資源監控
│   │   │   └── monitored_handler.go  # 監控包裝器
│   │   └── pluginnew/             # 腳手架生成
│   └── go.mod

├── python-adk-runtime/            # Python ADK Runtime
│   ├── src/detectviz_adk/
│   │   ├── config/                # 配置載入（與 Go 對齊）
│   │   ├── tools/remote_tool.py   # RemoteTool gRPC 客戶端
│   │   └── services/              # gRPC 服務實作
│   ├── templates/                 # ADK 樣板
│   └── requirements.txt

├── .github/workflows/             # CI/CD 流程
│   └── contracts-validation.yml   # 契約驗證與安全掃描
├── grafana-alloy/config.alloy     # 可觀測性收集配置
└── config.yaml                    # 預設平台配置
```

* * *

## 3. SSOT 契約

### 3.1 Proto（contracts/proto/detectviz/contracts/v1/adk_bridge.proto）

- **HealthService**：健康檢查、版本/能力列舉。
- **ToolBridgeService**：
  - `ExecuteTool(request) -> stream ToolChunk`：標準工具呼叫（支援串流回傳）。
  - `OpenSession/CloseSession`：會話化工具執行（可選）。
- **MemoryService（預留）**：由 Python/ADK 主導，Go 僅代理必要操作。
- **共通訊息**：`ToolRequest`（name/version/args/metadata/trace）、`ToolChunk`（data/status/progress/logs）。

生成策略（buf.gen.yaml）：

- Go：`contracts/gen/go/detectviz/contracts/v1`（go_package 設定固定）。
- Python：`contracts/gen/python/detectviz/contracts/v1`。

### 3.2 模組卡（contracts/schemas/module.card.schema.json）

- **識別**：`name`、`version`（SemVer）、`entrypoint`、`language`。
- **分類**：
  - `role`：`agent.coordinator`、`agent.tool_exec`、`tool`、`capability`、`plugin.gateway`、`memory.backend`、`security.module`、`observability.module`、`storage.module`。
  - `category`（借鏡 Telegraf 與 ADK）：如 `input`, `processor`, `output`（插件）；`llm`, `retriever`, `workflow`, `a2a`（ADK）。
- **依賴**：`requires`（名稱/版本規則）、`contracts.min_proto`。
- **運行**：`resources`、`config`（schema-uri）、`observability.tags`。
- **對齊關係**：已內嵌於 `role`/`category` 的枚舉與說明，AI 擴增時自動套用。

### 3.3 Config（contracts/schemas/config.schema.json）

- `env`、`grpc.listen/max_recv_bytes/max_send_bytes/tls`
- `observability.mode`: `lgtm_local|grafana_cloud|gcp`
- `observability.otlp`: `protocol(grpc|http)/endpoint/insecure/headers`
- `observability.logs`: `mode(file|stdout|off)` + `file.path/max_*`
- `observability.profiling`: 僅支援 `pprof`，欄位為 `enabled/pprof_address/application_name/tags`
- `plugin.paths/registry`、`memory.backend/dsn/default_ttl_seconds`
- 所有欄位皆有預設值策略（由 go-platform/internal/configx/loader.go 套用）。

* * *

## 4. Go 平台（go-platform）

### 4.1 CLI 介面

```bash
detectviz plugin serve [--listen :5002] [--config ./config.yaml]
                       [--mtls-cert path --mtls-key path --mtls-ca path]
                       [--http-demo] [--http-demo-listen :7777]

detectviz plugin new <category>/<name>      # 產生 Go 插件骨架
detectviz plugin validate <path>            # 導引到 contracts/tools 驗證
detectviz config validate -f <config.yaml>
```

- `--http-demo`：啟動 otelhttp 包裝的示範 HTTP 服務（預設路徑 `/hello`）以產生 span。
- `pprof` 由 `config.yaml` 的 `observability.profiling` 控制並由 `otel_init.go` 啟動，預設 `127.0.0.1:6060`。

### 4.2 ToolBridge 服務

- 以 `pluginhost/runtime.go` 啟動 gRPC 服務，使用 `contracts` 生成碼。
- Registry 註冊 Go 插件（telegraf-like 分類），Python 的 `RemoteTool` 以工具名稱路由。
- mTLS 可選（CLI 參數）；服務健康度由 HealthService 提供。

### 4.3 插件機制（telegraf-like）

- 目錄：`internal/pluginhost/plugins/<category>/<name>/plugin.go`。
- 介面：`Invoke(ctx, *pb.ToolInvokeRequest) (*pb.ToolInvokeReply, error)`；支援同步回傳。
- 最小範例：`capability.gateway/http_request`（HTTP 呼叫器，具備完整安全邊界）。
- 插件生命週期：`Invoke -> Close`（可選）；支援 ResourceAwareHandler 進行資源監控。
- 安全機制：內建 allowlist/denylist、payload 大小限制、超時控制等安全檢查。

### 4.4 組態載入與驗證

- `configx/loader.go`：YAML 讀取 → 預設值套用 → 以 `gojsonschema` 驗證 `config.schema.json`。
- 驗證成功後以結構化 `zap` 輸出摘要。

### 4.5 契約版本檢查（Critical）

- `internal/contracts/version_check.go`：啟動時強制檢查 proto 生成碼版本對齊。
- 讀取 `contracts/gen/metadata/version.json` 驗證 buf 版本、生成時間、proto hash。
- 版本不一致時拒絕啟動並提供清晰錯誤訊息，確保 SSOT 合規性。
- 支援跨語言一致性檢查（Go 和 Python 生成碼）。

### 4.6 Observability（Go）

- `logging.go`：以 `zap` + `otelzap` 初始化；ConsoleEncoder；可雙寫 stdout + 檔案（`./var/log/detectviz/detectviz.log`）。
- `otel_init.go`：
  - 以 `observability.otlp` 初始化 OTLP Traces/Metrics Exporter（gRPC 或 HTTP）。
  - 啟用 Go runtime metrics（`go.opentelemetry.io/contrib/instrumentation/runtime`）。
  - 依 `observability.profiling` 啟動 `net/http/pprof`（僅 pprof，無應用端 Pyroscope push）。

* * *

## 5. Python ADK Runtime（python-adk-runtime）

### 5.1 對齊 ADK

- 嚴格遵守 ADK 的 Agent/Memory/Workflow/Tools/Capabilities 模組邊界。
- Multi-Agent 與 A2A 以 ADK 原生機制；Coordinator-Agent 與 Tool-Execution-Agent 分工清楚。

### 5.2 Tools vs Capabilities

- **Tools**：對外部系統的具體操作介面（HTTP、gRPC、DB、Shell 等），可經由 `RemoteTool` 轉呼 Go 插件。
- **Capabilities**：模型、資料存取、檢索、規則、策略等可組合的能力單元（不直接握外部副作用）。
- 目錄分離讓可重用/測試與授權最小化更明確；模型升級或向量庫切換歸類為 Capabilities。

### 5.3 RemoteTool

- 以 `contracts` 生成的 Python gRPC Client 連接 Go ToolBridge。
- 工具標識以 `module.card.json` 之 `name/version` 對齊；傳遞 trace context 與 metadata。

### 5.4 記憶體（MemoryBank）

- 以抽象介面對接 In-Memory、Redis、向量庫（Weaviate/Chroma/Milvus）或雲端（Vertex/Bigtable）。
- 命名空間、權限與 Cache 策略可由 Capability/Workflow 控制；支援共享記憶體 bank 抽換。

* * *

## 6. Observability 與部署

### 6.1 Alloy 管線（最小可行）

- **Logs → Grafana Cloud Loki**：file tail `./var/log/detectviz/*.log` → `loki.write`。
- **Traces/Metrics → Grafana Cloud OTLP**：`otelcol.receiver.otlp(4317/4318)` → `otelcol.exporter.otlphttp`（basic_auth）。
- **Profiles → Grafana Cloud Pyroscope**：
  - 應用啟 `pprof`：`pyroscope.scrape` → `pyroscope.write`。
  - 僅採用 scrape 模式；應用端不持有雲端憑證。

支援模式切換：

- `observability.mode: lgtm_local`：目標為本地 LGTM。
- `grafana_cloud`：使用上述 `write` exporter。
- `gcp`：可切換到 Google Cloud 原生端點（Cloud Trace/Logging/Profiler）。

### 6.2 環境變數（Grafana Cloud）

- OTLP/Tempo：`GF_CLOUD_OTEL_ID`、`GCLOUD_RW_API_KEY`、`OTLP_GATEWAY_URL`
- Pyroscope：`GF_CLOUD_PROFILES_ID`、`GF_CLOUD_PROFILES_URL`、`GCLOUD_RW_API_KEY`
- Loki：`GCLOUD_HOSTED_LOGS_ID`、`GCLOUD_HOSTED_LOGS_URL`、`GCLOUD_RW_API_KEY`

上述命名用於 Alloy `config.alloy`，僅作為參考實作；實際名稱可在 `.env` 層統一管理。

* * *

## 7. 擴增流程（清單化）

### 7.1 新增 Go 插件（platform-side）

1. `detectviz plugin new <category>/<name>` 生成骨架至 `internal/pluginhost/plugins/...`。
2. 補齊 `module.card.json`（role=`plugin.gateway` 或對應分類），`plugin.go` 實作 `Execute`。
3. 在 `registry.go` 註冊；`plugin serve` 啟動後由 Python 端可見。
4. `contracts/tools/validate_module_card.py` 驗證模組卡。

### 7.2 新增 Python Agent/Tool/Capability（runtime-side）

1. 從 `python-adk-runtime/templates` 選擇樣板（`agent.tool_exec`、`agent.coordinator`、`tool`、`capability`）。
2. 撰寫 `module.card.json`（role/category 對齊）；如需外部呼叫，使用 `RemoteTool`。
3. 將 Workflow 與 MemoryBank 設定透過 ADK 組裝；加上單元/整合測試。
4. 以 `spec.md` 的 A2A 溝通契約驗證互通。

* * *

## 8. 組態與啟動（MVP）

### 8.1 go-platform/configs/config.yaml

```yaml
env: dev
grcc:
  listen: ":6606"
  max_recv_bytes: 4194304
  max_send_bytes: 4194304
observability:
  mode: lgtm_local            # 或 grafana_cloud / gcp
  otlp:
    protocol: grpc
    endpoint: "127.0.0.1:4317"
    insecure: true
  logs:
    mode: file
    file:
      path: ./var/log/detectviz/detectviz.log
  profiling:
    enabled: true
    pprof_address: "127.0.0.1:6060"
    application_name: "go-platform"
    tags:
      service.name: "go-platform"
      deployment.environment: "dev"
plugin:
  paths: ["./go-platform/internal/pluginhost/plugins"]
  registry: file
memory:
  backend: inmem
  default_ttl_seconds: 3600
```

### 8.2 啟動指令

```bash
# 檢查組態
detectviz config validate -f ./go-platform/configs/config.yaml

# 啟動 ToolBridge（本機）
detectviz plugin serve --config ./go-platform/configs/config.yaml --http-demo --http-demo-listen :7777

# 產生 Traces
curl -sS http://127.0.0.1:7777/hello
```

* * *

## 9. 測試與驗證

- **契約測試**：contracts 生成碼能在 Go/Python 成功編譯；proto 版本與 go_package/python package 固定。
- **模組卡驗證**：`validate_module_card.py` 應對 role/category/semver/依賴做靜態檢查。
- **端對端**：Python Agent 透過 `RemoteTool` 呼叫 Go 插件，返回串流 chunk；trace context 傳遞正確。
- **觀測驗證**：
  - Logs：Grafana Cloud Loki 可看到 `service=go-platform` 行。
  - Traces：Tempo 出現 `/hello` 的 span。
  - Metrics：可見 Go runtime metrics 與自訂度量；若使用 Alloy `filter` 過濾 `http.server.*` 指標，需依需求調整規則。
  - Profiles：Pyroscope 以 `pyroscope.scrape` 擷取 pprof 樣本；應與 Traces/Logs Drilldown 關聯。
- Collector 未啟動時，Go 端不得 panic（Exporter 失敗須以警告處理）。

* * *

## 10. 安全與授權

- mTLS（可選）保護 ToolBridge；TLS 憑證由 CLI 參數提供。
- 插件與工具的外部呼叫權限由模組卡與平台策略限制（最小權限）。
- 憑證與 Token 經環境變數與 Secret 管理，不落地於 repo。

* * *

## 11. 路線圖（漸進擴充）

- **UI 控台**：Agent/Plugin/Workflow 管理；基於 gRPC 定義生成 SDK。
- **Workflow Orchestration**：增強協調器（重試/補償/超時/並行控制）。
- **記憶體一致性**：引入因果/會話一致策略與衝突合併。
- **多語言 Agent**：保持 gRPC 契約穩定，擴充到 JVM/.NET。
- **雲端原生**：Kubernetes Helm/Operators，支援水平擴展與自動化部署。

* * *

## 12. 關鍵設計準則回顧

- **ADK 為核心**：Python 端遵守 ADK 邊界；Go 僅作橋接與觀測。
- **SSOT 契約**：proto/schema 為唯一真相，生成各語言型別與驗證。
- **插件/Agent 分類一致**：`role`/`category` 由 contracts 枚舉統一，利於 AI 擴增。
- **雲地一致配置**：以 `observability.mode` 切換目標（LGTM / Grafana Cloud / GCP），程式碼零變更。
- **可觀測性優先**：zap+otel、Logs/Traces/Metrics/Profiles 一致關聯；初始化須不阻塞主流程。
- **MVP 先行**：先打通 ToolBridge、RemoteTool、最小插件與範例 Agent，再逐步加值。