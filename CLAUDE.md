## 目的
本文件提供給 AI 與人類協作者的操作與開發守則，確保在 Detectviz 平台上以 **SSOT（contracts）為唯一事實來源** 進行變更；同時規範雲端觀測（Grafana Cloud / GCP）、本地 LGTM、以及程式碼變更的安全與品質要求。

## 安全與機密
- 請勿在任何檔案（含本文件、README、程式碼、日誌）中放入 **真實 Token / 密碼 / 秘鑰**。一律使用環境變數或 Secret 管理。
- 若曾將實際 Token 寫入 repo，請**立刻撤銷與輪替**，並改用 `.env` 或 CI Secret 注入。
- 避免在日誌中輸出敏感資訊；`zap` 的欄位請審核後再新增。

## 觀測環境變數（Grafana Cloud）
將下列內容放入 `.env`（僅示意，占位符請替換為你自己的值）並 `source .env` 後再啟動 Alloy：
```bash
# OTLP（Tempo/Mimir 綜合 OTLP Gateway）使用 Basic Auth
export GF_CLOUD_OTEL_ID="__STACK_ID__"          # Basic Auth username（同 Stack ID）
export GCLOUD_RW_API_KEY="__OTEL_RW_TOKEN__"    # Basic Auth password（可覆用於 Logs/Profiles）

# Loki（Logs）
export GCLOUD_HOSTED_LOGS_URL="https://logs-prod-<region>.grafana.net/loki/api/v1/push"
export GCLOUD_HOSTED_LOGS_ID="__STACK_ID__"      # Basic Auth username

# Pyroscope（Profiles）
export GF_CLOUD_PROFILES_URL="https://profiles-<region>.grafana.net"
export GF_CLOUD_PROFILES_ID="__STACK_ID__"       # Basic Auth username

# 可選覆蓋（本機路徑/位址）
export DETECTVIZ_LOG_PATH="/abs/path/to/var/log/detectviz/detectviz.log"
export DETECTVIZ_PPROF_ADDR="127.0.0.1:6060"
```

## 快速啟動（本地 LGTM 或 Grafana Cloud）
1. **使用 SSOT 樣本組態**  
   `contracts/samples/config.yaml` 已同步為 `go-platform/configs/config.yaml` 的預設樣本，任何調整請先改 **contracts** 再同步至各專案。
2. **啟動 Alloy（Grafana Cloud 例）**  
   `./grafana-alloy/alloy run ./grafana-alloy/config.alloy`
3. **啟動 Go 平台（ToolBridge + http-demo）**
   ```bash
   go run ./go-platform/cmd/detectviz plugin serve \
     --config ./go-platform/configs/config.yaml \
     --http-demo --http-demo-listen :7777
   # 驗證產生 Traces
   curl -sS http://127.0.0.1:7777/hello
   ```
4. **在 Grafana Cloud 檢視**  
   Explore/Logs、Traces、Profiles 皆可 Drilldown。

## SSOT 與設定（必讀）
- 任何跨語言契約或設定變更，**一律先改 `contracts/`**：
  - `proto/detectviz/contracts/v1/adk_bridge.proto`：gRPC 契約
  - `schemas/module.card.schema.json`：模組卡（請維持既有 `role`/`category` 枚舉）
  - `schemas/config.schema.json`：平台組態（含 `observability` 全部欄位）
  - `samples/config.yaml`：範例設定（必須與 Schema 對齊）
- 修改 `proto` 後：`buf lint && buf generate`；**請勿手改 `*.pb.go` 生成碼**。
- `go-platform` 僅讀取經 `configx/loader.go` 解析與驗證後的設定；請避免繞過 Loader。

## Profiles（Pyroscope）與 Drilldown 完整性
- **Schema 必須包含 profiling 區塊**（由 SSOT 管理）：
  ```yaml
  observability:
    profiling:
      enabled: true
      pprof_address: "127.0.0.1:6060"
      application_name: "go-platform"
      tags:
        service.name: go-platform
        deployment.environment: dev
  ```
- Go 端依設定啟動 `net/http/pprof`。Alloy 以 `pyroscope.scrape` 抓取並透過 `pyroscope.write` 上傳雲端。
- **Logs ↔ Traces 關聯**：在 Alloy 的 Loki 管線加入 regex 解析 `trace_id` → 標籤 `traceid`，確保從 Logs Drilldown 到 Tempo 穩定。

## Metrics 與匯流量控制
- 優先在 Collector（Alloy）端做 `batch` 與 `memory_limiter`；必要時再於 Go SDK 調整。
- 若需要過濾噪音 HTTP 指標，可在 Alloy 的 `otelcol.processor.filter` 處理。

## Go 平台實作守則（AI 請遵守）
- **zap 初始化**：`InitZapLoggerToFile()` 需先建立目錄，再 `zap.ReplaceGlobals(zl)`，並禁用 `log.Printf`。
- **檔案日誌**：ConsoleEncoder 純文字，預設 `./var/log/detectviz/detectviz.log`（供 Alloy file tail）。
- **OTEL 初始化**：依 `observability.otlp` 輸出（gRPC 4317 或 HTTP 4318）。
- **pprof 啟動**：僅在 `observability.profiling.enabled: true` 時啟動，位址 `pprof_address`。
- **CLI 旗標**：`--http-demo`、`--http-demo-listen` 可用；子命令旗標解析採 `fs.Parse(os.Args[3:])` 合理可行。
- **避免重複定義**：`GetObservabilityConfig()` 僅保留**一份**實作，變更時同步調整 Loader 與 Schema。
- **gRPC/Proto**：對 `go_package`、import 路徑保持穩定；**不要**手動編輯 `*.pb.go`。
- **Plugin 機制**：`internal/pluginhost/plugins/<category>/<name>/plugin.go`，介面 `Execute(ctx, *ToolRequest) -> stream ToolChunk`。

## Python ADK Runtime 提示
- 嚴格遵循 ADK 模組邊界：Agent / Workflow / MemoryBank / Tools / Capabilities。
- 遠端工具請使用 `RemoteTool` 經由 ToolBridge 呼叫 Go 插件，**禁止**跨語言直接耦合。
- 建立或擴增元件時，請先撰寫 `module.card.json` 並以 `contracts/tools/validate_module_card.py` 驗證。

## 典型錯誤與排查
- `config validation failed: Additional property profiling.*`：
  - 可能原因：Schema 未更新、或 `config.yaml` 仍含不被允許的欄位（例如舊的 `profiling.mode/endpoint/username/password`）。請先更新 `contracts/schemas/config.schema.json`，再同步 Loader 與樣本。
- `connect: connection refused :4317`：Alloy 未啟或端口不符；若改用 HTTP，請調整 `protocol: http` 與端點 4318。
- Logs 無輸出：檢查是否呼叫 `zap.ReplaceGlobals`、是否建立日誌目錄、是否仍殘留 `log.Printf`。
- Drilldown 失敗：確認 Loki 是否有 `traceid` 標籤（由 regex 萃取）。

## 一致性檢查（建議）
```bash
cd contracts
buf lint && buf generate
python3 tools/validate_module_card.py path/to/module.card.json
```

## 變更要求
- 任何對規格（spec.md）、契約（contracts）、或觀測管線（profiles）的修改，請在 PR 中：
  1. 說明對 SSOT 的變更點；
  2. 附上端到端驗證步驟（logs/traces/profiles）；
  3. 若影響到使用文件，**同步更新**本檔與 README。