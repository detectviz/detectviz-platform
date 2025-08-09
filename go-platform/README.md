
## go-platform（Platform Core / ToolBridge）

Detectviz 平台的最小平台層，負責：
- CLI 與 gRPC ToolBridge 服務（跨語言調用 Go 插件供 Python/ADK 使用）
- Plugin Registry（telegraf-like 分類與註冊）
- 組態載入與 Schema 驗證（SSOT 對齊 `contracts/schemas/config.schema.json`）
- 可觀測性初始化（zap + OpenTelemetry；Logs/Traces/Metrics/Profiles）

---

## 目錄結構（精簡）
- `cmd/detectviz/`：CLI 入口（`plugin|config` 子命令）
- `internal/configx/`：設定載入＋Schema 驗證
- `internal/pluginhost/`：ToolBridge 伺服端、Registry、插件目錄
- `internal/observability/`：zap 與 OTLP 初始化（含 pprof 啟動）
- `configs/config.yaml`：預設組態（與 `contracts/samples/config.yaml` 同步）

---

## 系統需求
- Go 1.21+（建議 1.22）
- 本機或遠端 Alloy（Grafana Alloy 作為收集/轉送層）
- `contracts/` 倉庫（SSOT）可讀（同一專案或相鄰路徑）

---

## 快速開始（本機 LGTM / Grafana Cloud 均適用）
1. 使用 SSOT 樣本組態（已同步至 `go-platform/configs/config.yaml`）：
   ```bash
   detectviz config validate -f ./go-platform/configs/config.yaml
   ```

2. 啟動 Alloy（請先於 `.env` 設定必要環境變數）。若使用本專案提供的檔案：
   ```bash
   ./grafana-alloy/alloy run ./grafana-alloy/config.alloy
   ```

3. 啟動 ToolBridge 並啟用示範 HTTP（產生 Traces/Metrics/Logs）：
   ```bash
   go run ./go-platform/cmd/detectviz plugin serve \
     --config ./go-platform/configs/config.yaml \
     --http-demo --http-demo-listen :7777
   ```
   另開終端打流量：
   ```bash
   curl -sS http://127.0.0.1:7777/hello
   ```

4. 檔案日誌輸出（供 Alloy 轉發至 Loki）：
   - `./var/log/detectviz/detectviz.log`（zap ConsoleEncoder 純文字）

5. Profiles（pprof）：
   - go-platform 依 `observability.profiling` 自動啟動 pprof（預設 `127.0.0.1:6060`）
   - Alloy 以 `pyroscope.scrape` 擷取並上傳至 Grafana Cloud

---

## 組態說明（與 SSOT 對齊）
- 單一來源：`contracts/schemas/config.schema.json`。範例檔：`contracts/samples/config.yaml`。
- 觀測端點：
  - OTLP 預設 gRPC → `127.0.0.1:4317`（若用 HTTP：`protocol: http` 並改 `endpoint: http://127.0.0.1:4318`）
- 日誌輸出：`observability.logs.mode: file`，檔案路徑 `./var/log/detectviz/detectviz.log`
- Profiling：**僅支援 pprof**，欄位為 `enabled / pprof_address / application_name / tags`，無任何雲端憑證欄位
- gRPC（ToolBridge）：`grpc.listen` 預設 `:6606`，最大封包限制 ≥ 1MB（schema 已限制）

節錄（與 SSOT 同步）：
```yaml
observability:
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
```

---

## CLI 用法
```bash
detectviz plugin serve [--listen :6606] [--config ./go-platform/configs/config.yaml] \
                       [--mtls-cert path --mtls-key path --mtls-ca path] \
                       [--http-demo] [--http-demo-listen :7777]

detectviz plugin new <category>/<name>         # 產生 Go 插件骨架
detectviz plugin validate <path>               # 導引到 contracts 工具/驗證
detectviz config validate -f <config.yaml>
```
- `--http-demo`：啟動 otelhttp 包裝的示範 HTTP 伺服，產生 span 與 metrics
- pprof 由 `observability.profiling.enabled` 控制；位址 `pprof_address`

---

## 插件開發（telegraf-like）
1. 產生骨架：
   ```bash
   detectviz plugin new capability.gateway/http_request
   ```
2. 實作 `internal/pluginhost/plugins/<category>/<name>/plugin.go`：
   - 介面：`Execute(ctx, *pb.ToolRequest) (<-chan *pb.ToolChunk, error)`（支援串流回傳）
3. 在 `internal/pluginhost/registry.go` 註冊

4. 撰寫 `module.card.json` 並以 `contracts/tools/validate_module_card.py` 驗證：
   - `role`: `plugin.gateway`（或依實際分類）
   - `category`: `gateway` / `collector.input` 等
5. 啟動後，Python/ADK 端可透過 `RemoteTool` 以名稱路由到該插件

---

## 觀測到 Grafana Cloud / GCP
- 建議使用 Alloy 作為收集與轉送層（本專案提供 `grafana-alloy/config.alloy`）
- 需要的環境變數（示意，放 `.env`）：
  - OTLP：`GF_CLOUD_OTEL_ID`、`GCLOUD_RW_API_KEY`
  - Loki：`GCLOUD_HOSTED_LOGS_URL`、`GCLOUD_HOSTED_LOGS_ID`
  - Profiles：`GF_CLOUD_PROFILES_URL`、`GF_CLOUD_PROFILES_ID`
  - 可選覆蓋：`DETECTVIZ_LOG_PATH`、`DETECTVIZ_PPROF_ADDR`
- 若改本地 LGTM：請切換 `contracts/profiles/lgtm_local/` 對應設定

---

## 疑難排解
- `configuration validation failed`：`config.yaml` 與 `config.schema.json` 不一致（常見：殘留 `profiling.mode/endpoint/username/password`）。請改用範例檔覆寫。
- `connect: connection refused :4317`：Alloy 未啟或端口不符；或改用 HTTP（4318）。
- 無 Profiles：確認 pprof 監聽與 Alloy `pyroscope.scrape` 目標一致（`__address__=127.0.0.1:6060`）。
- 無 Logs：檢查 `./var/log/detectviz/detectviz.log` 是否存在，並確認 Alloy `loki.source.file` 目標路徑。
- Drilldown 失敗：請在 Alloy 的 Loki 管線確保解析出 `traceid` 標籤。

---

## 注意事項
- **請勿**在任何程式碼或設定中放入雲端憑證。憑證應只存在 `.env` 與 `config.alloy`。
- SSOT 為 `contracts/`，修改 proto/schema 後請以 buf/gojsonschema 驗證並同步生成碼，勿手改 `*.pb.go`。

---

## 參考
- `../spec.md`（整體平台規格）
- `../contracts/`（SSOT：proto / schema / samples）
- `../python-adk-runtime/`（ADK Runtime 與 RemoteTool）
- `../grafana-alloy/config.alloy`（收集/轉送設定）

---

## 編譯與執行
```bash
go build -o ./bin/detectviz ./go-platform/cmd/detectviz
./bin/detectviz plugin serve --config ./go-platform/configs/config.yaml
```
