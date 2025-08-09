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
  proto/
    detectviz/contracts/v1/
      adk_bridge.proto          # ToolBridge / Health 等 gRPC 定義
  schemas/
    module.card.schema.json     # 模組卡（Agent/Tool/Capability/Plugin）規範
    config.schema.json          # 平台組態規範（go-platform 載入與驗證）
  samples/
    config.yaml                 # 樣本設定（與 Schema 對齊）
  profiles/                     # （選配）Alloy 管線樣板：lgtm_local / grafana_cloud / gcp
    alloy/...
  tools/
    validate_module_card.py     # 模組卡驗證工具
    check_contracts.sh          # 一鍵檢查腳本（可選）
```

---

## 生成流程（buf）
以 `buf` 為唯一的 Proto 生成工具，輸出至固定位置供各語言引用。

```bash
cd contracts
buf lint
buf generate
```

輸出路徑（由 `buf.gen.yaml` 決定）：
- Go：`contracts/gen/go/detectviz/contracts/v1`（`go_package` 固定）
- Python：`contracts/gen/python/detectviz/contracts/v1`

注意事項：
- 生成後請於 `go-platform` 執行 `go mod tidy`；於 `python-adk-runtime` 確保可匯入生成的 Python stub。
- 請勿手動編輯 `*.pb.go`、`*_grpc.py` 等生成檔。

---

## Proto 契約（`proto/detectviz/contracts/v1/adk_bridge.proto`）
- **HealthService**：健康檢查、版本與能力列舉。
- **ToolBridgeService**：
  - `ExecuteTool(request) -> stream ToolChunk`：工具執行（支援串流回傳）。
  - `OpenSession/CloseSession`：會話化工具執行（可選）。
- **MemoryService（預留）**：供 ADK Runtime 透過 Go 平台代理必要操作。
- **共通訊息**：`ToolRequest`（name/version/args/metadata/trace）、`ToolChunk`（data/status/progress/logs）。

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
供 `go-platform` 載入與驗證的平台組態，`samples/config.yaml` 為對齊樣本。

重點欄位：
- `grpc.listen/max_recv_bytes/max_send_bytes/tls`（最大值限制 ≥ 1 MiB）
- `observability.mode`: `lgtm_local | grafana_cloud | gcp`
- `observability.otlp`: `protocol(grpc|http)/endpoint/insecure/headers`
- `observability.logs`: `mode(file|stdout|off)` + `file.path/max_*`
- `observability.profiling`: **僅支援 pprof** → `enabled/pprof_address/application_name/tags`（無 endpoint/username/password）
- `plugin.paths/registry`：`paths` 為陣列，不能為空值
- `memory.backend/dsn/default_ttl_seconds`

常見驗證錯誤：
- `Additional property profiling.* is not allowed`：`config.yaml` 殘留舊欄位（`mode/endpoint/username/password`）。
- `plugin.paths: Invalid type. Expected: array`：請填入陣列，即使僅一個路徑也需以 `[...]` 表示。

---

## Alloy Profiles 樣板（選配）
- `profiles/alloy/` 內提供本地 LGTM、Grafana Cloud、GCP 的 Collector 管線示例。
- 原則：應用端不持有雲端憑證；由 Collector（Alloy）負責傳輸與認證。
- Profiles 以 pprof scrape 為唯一路徑（`pyroscope.scrape` → `pyroscope.write`）。

環境變數慣例（示例）：
- OTLP：`GF_CLOUD_OTEL_ID`、`GCLOUD_RW_API_KEY`
- Loki：`GCLOUD_HOSTED_LOGS_URL`、`GCLOUD_HOSTED_LOGS_ID`
- Pyroscope：`GF_CLOUD_PROFILES_URL`、`GF_CLOUD_PROFILES_ID`

---

## 開發工作流（強制）
1. 任何跨語言介面或設定變更，**先改 `contracts/`**。
2. `buf lint && buf generate` 產生最新 Go/Python 生成碼。
3. 更新 `samples/config.yaml`，並確保 `go-platform` 經 `configx/loader.go` 可通過驗證。
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