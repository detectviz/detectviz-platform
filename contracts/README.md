# Contracts（SSOT）

本目錄是 Detectviz 的單一事實來源（Single Source of Truth, SSOT）。
- 定義跨語言 gRPC 介面（Proto）
- 定義平台組態（Config Schema）
- 定義模組卡規格（Module Card Schema）
- 提供生成器設定與驗證工具

所有語言（Go / Python）的型別、驗證與 API 皆由此目錄衍生，請勿在下游專案手動修改生成碼。

---

## 目錄結構
```
contracts/
  buf.yaml
  buf.gen.yaml
  buf.lock
  Makefile                      # 版本控制與生成管理
  proto/
    detectviz/contracts/v1/
      adk_bridge.proto          # ToolBridge / Health 等 gRPC 定義
  schemas/
    config.schema.json          # 平台組態規範
    module.card.schema.json     # 模組卡規範
    plugin.schema.json          # 插件規範
  samples/
    config.yaml                 # 樣本設定（建議複製到專案根）
    module.card.json            # 模組卡樣本
    plugin.yaml                 # 插件配置樣本
  specs/                        # 技術規格文檔
    error_model.md              # 錯誤模型規範
    memory_bank_contract.md     # 記憶體契約規範
    telemetry_conventions.md    # 遙測慣例
  tools/
    validate_module_card.py     # 模組卡驗證工具
    validate_config.py          # 配置驗證工具
    validate_plugin.py          # 插件驗證工具
    check_contracts.sh          # 一鍵檢查腳本
  gen/
    go/                         # buf 生成的 Go stub（請勿手改）
    python/                     # buf 生成的 Python stub（請勿手改）
    metadata/
      version.json              # 版本元數據（由 make gen 生成）
  profiles/                     # Alloy 配置樣板
    local/                      # 本地 LGTM 配置
    grafana-cloud/              # Grafana Cloud 配置
    gcp/                        # GCP 配置
```

---

## 生成流程（buf）
以 `buf` 為唯一的 Proto 生成工具，輸出至固定位置供各語言引用。

```bash
cd contracts
buf lint
buf generate
```

或使用 Makefile：
```bash
cd contracts
make gen            # 等同 buf lint + buf generate
make validate       # 執行 schema 與卡片等檢查（若已定義）
make validate-cards # 僅驗證 module.card.json（若已定義）
```

輸出路徑（由 `buf.gen.yaml` 決定）：
- Go：`contracts/gen/go/detectviz/contracts/v1`（`go_package` 固定）
- Python：`contracts/gen/python/detectviz/contracts/v1`

注意事項：
- 生成後請於 `go-platform` 執行 `go mod tidy`；於 `python-adk-runtime` 確保可匯入生成的 Python stub。
- 請勿手動編輯 `*.pb.go`、`*_pb2.py`、`*_pb2_grpc.py` 等生成檔。

---

## Proto 契約（`proto/detectviz/contracts/v1/adk_bridge.proto`）
- **HealthService**：健康檢查、版本與能力列舉（gRPC Health v1 另行註冊）。
- **ToolBridge**：
  - `Invoke(ToolInvokeRequest) returns (ToolInvokeReply)`：工具執行（目前採用單次回應）。
  - 預留串流與會話化 RPC（必要時新增 `ExecuteTool`/`OpenSession`/`CloseSession` 等）。
- **MemoryService（預留）**：供 ADK Runtime 透過平台代理特定記憶體操作。
- **主要訊息**：
  - `ToolInvokeRequest`：`tool_id`、`tool_version`、`payload`（`google.protobuf.Struct`）、`timeout_ms`、`metadata`（`map<string,string>`）
  - `ToolInvokeReply`：`status`（`code/message`）、`result`（`Struct`）、`exec_meta`（`attempt/duration_ms/plugin_id/route_id`）

版本策略：
- 遵循 SemVer；相容新增採 optional 欄位與新 RPC；禁止破壞性變更。
- 模組卡可填入 `contracts.min_proto` 以表達最小相容版本。

---

## 模組卡規格（`schemas/module.card.schema.json`）
用於描述 Agent、Tool、Capability、Plugin 等元件的身分、分類與相依。AI 擴增與審核一律依此 Schema 驗證。

必填欄位摘要：
- `name`：全域唯一名稱（建議使用域樣式，如 `detectviz.tools.http_request`）
- `version`：SemVer
- `entrypoint`：主程式或模組入口（相對於專案根目錄）
- `language`：`go` / `python` / 其他
- `role`（枚舉，對齊 ADK/Telegraf 分類）：
  - `agent.coordinator`，`agent.tool_exec`，`tool`，`capability`，
  - `plugin.gateway`，`memory.backend`，`security.module`，`observability.module`，`storage.module`
- `category`（依角色細分）：
  - 插件：`collector.input` / `transform.processor` / `aggregate.aggregator` / `sink.output` / `gateway`
  - ADK/平台：`llm` / `retriever` / `workflow` / `a2a` / `capability`
  - 觀測/安全/儲存：`observability.exporter` / `observability.processor` / `security.authn` / `security.authz` /
    `memory.backend` / `storage.blob` / `storage.kv` / `storage.vector`
- `requires`：依賴清單（名稱 + 版本規則）
- `contracts.min_proto`：相容性宣告

驗證指令：
```bash
python3 contracts/tools/validate_module_card.py path/to/module.card.json
```

---

## Config Schema（`schemas/config.schema.json`）
供 `go-platform` 與 `python-adk-runtime` 共同載入與驗證的平台組態，`samples/config.yaml` 為對齊樣本。

重點欄位：
- `grpc.listen/max_recv_bytes/max_send_bytes/tls`（最小值限制 ≥ 1 MiB）
- `observability.mode`: `lgtm_local | grafana_cloud | gcp`
- `observability.otlp`: `protocol(grpc|http)/endpoint/insecure/headers`
- `observability.logs`: `mode(file|stdout|off)` + `file.path/max_*`
- `observability.profiling`: **僅支援 pprof** → `enabled/pprof_address/application_name/tags`（無 endpoint/username/password）
- `plugin.paths/registry`：`paths` 為陣列，不能為空值
- `memory.backend/dsn/default_ttl_seconds`

統一讀取路徑（Go 與 Python 一致）：
1. 旗標或函式參數：`--config /path/to/config.yaml`
2. 環境變數：`DETECTVIZ_CONFIG_FILE=/path/to/config.yaml`
3. 工作目錄：`./config.yaml`
4. 合約覆蓋：`./contracts/config.yaml`
5. 樣本兜底：`./contracts/samples/config.yaml`

常見驗證錯誤：
- `Additional property profiling.* is not allowed`：`config.yaml` 殘留舊欄位（`mode/endpoint/username/password`）。
- `plugin.paths: Invalid type. Expected: array`：請填入陣列，即使僅一個路徑也需以 `[...]` 表示。

---

## Alloy Profiles 樣板（選配）
- `profiles/alloy/` 內提供本地 LGTM、Grafana Cloud、GCP 的 Collector 管線示例。
- 原則：應用端不持有雲端憑證；由 Collector（Alloy）負責傳輸與認證。
- Profiles 僅使用 pprof scrape 路徑（`pyroscope.scrape` → `pyroscope.write`）。

環境變數慣例（示例）：
- OTLP：`GF_CLOUD_OTEL_ID`、`GCLOUD_RW_API_KEY`
- Loki：`GCLOUD_HOSTED_LOGS_URL`、`GCLOUD_HOSTED_LOGS_ID`
- Pyroscope：`GF_CLOUD_PROFILES_URL`、`GF_CLOUD_PROFILES_ID`

---

## 開發工作流（強制）
1. 任何跨語言介面或設定變更，**先改 `contracts/`**。
2. `buf lint && buf generate` 產生最新 Go/Python 生成碼。
3. 更新 `samples/config.yaml`，並確保下游載入程式可通過 Schema 驗證。
4. 編修或新增 `module.card.json`，以工具驗證分類與相依。
5. 在 `go-platform` 執行 `go mod tidy`，在 `python-adk-runtime` 確認 gRPC stub 可匯入。
6. 執行端到端測試（Logs/Traces/Metrics/Profiles）。

---

## 相容性與版本
- Proto 採 SemVer；新增以向後相容方式進行。
- Schema 變更需同步更新 `samples/config.yaml` 與下游 README。
- 嚴禁手動修改任何生成檔案；如需調整，請回到契約端更新。

---

## 常見問題（FAQ）
- Q：為何我的設定驗證失敗？
  - A：請對照 `schemas/config.schema.json`，移除未定義欄位，並確保陣列型別正確（如 `plugin.paths`）。
- Q：Python/Go 生成碼路徑不一致？
  - A：請檢查 `buf.gen.yaml` 與下游 import 路徑是否對齊；不要手動調整生成碼目錄。
- Q：Logs/Traces/Profiles 無法在雲端看到？
  - A：請使用 Alloy 樣板，並確認環境變數與憑證設定正確。Profiles 僅支援 pprof scrape 路徑。

---

## 變更紀錄（要點）
- 移除 `observability.profiling.mode`、`profiling.endpoint/username/password` 欄位；統一 pprof（scrape）。
- 明確化 `role`/`category` 枚舉與對齊關係，供 AI 擴增自動套用。
- 強制以 `buf` 產生跨語言生成碼，禁止下游手動變更。