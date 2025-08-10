# go-platform（Platform Core / ToolBridge）

最小化的平台核心，提供跨語言可用的 gRPC **ToolBridge**、可插拔 **Plugin Host**、統一 **Observability** 與 **健康檢查／優雅關機**。本模組與 `python-adk-runtime` 僅以 gRPC/HTTP 通訊，嚴格解耦並對齊 `contracts/` 的 SSOT 契約。

---

## 平台職責
- **ToolBridge（gRPC）**：供 Python/ADK 的 `RemoteTool` 呼叫 Go 插件（工具執行）。
- **Plugin Host**：插件註冊、生命週期與執行；支援嚴格註冊（禁止覆蓋）。
- **Observability**：OpenTelemetry 統一匯出（Logs/Traces/Metrics），預設透過 Alloy 導至本地 Grafana 或 Grafana Cloud。
- **健康檢查與優雅關機**：提供 `/livez`、`/readyz` 與 gRPC Health；啟動就緒與 SIGTERM/SIGINT 的有序關閉。

---

## 目錄結構
- `cmd/detectviz/`：平台啟動器（CLI）
- `internal/configx/`：設定載入與驗證（統一優先序 + 環境變數覆蓋）
- `internal/observability/`：OTel 初始化與 zap 日誌
- `internal/health/`：HTTP 健康檢查服務（/livez、/readyz）
- `internal/pluginhost/`：插件宿主（registry、runtime、ToolBridge 伺服端）
  - `plugins/capability.gateway/http_request/`：內建範例插件

---

## 系統需求
- Go 1.22+
- 已產生 gRPC 生成碼（於 repo 根目錄執行 `cd contracts && make gen`）
- Alloy 已就緒（本地或 Grafana Cloud）

---

## 快速開始
```bash
# 於 repo 根目錄建立生效設定（依 SSOT 樣本）
cp contracts/samples/config.yaml ./config.yaml

# 可選：覆蓋常用環境變數
export DETECTVIZ__OBSERVABILITY__OTLP__ENDPOINT=127.0.0.1:4317
export DETECTVIZ_HEALTH_ADDR=":8081"

# 啟動 ToolBridge 與示範 HTTP（具 otelhttp 儀表化）
go run ./cmd/detectviz --config ./config.yaml \
  --http-demo \
  --http-demo-listen :8080
```
- 另開終端打流量：`curl -sS http://127.0.0.1:8080/hello`
- 檔案日誌輸出（供 Alloy 轉發至 Loki）：`./var/log/detectviz/detectviz.log`
- Profiles（pprof）：依 `observability.profiling` 自動啟動，預設 `127.0.0.1:6060`

---

## 設定載入與優先序（SSOT 對齊）
**搜尋順序（高 → 低）**：
1. 旗標：`--config /path/to/config.yaml`
2. 環境：`DETECTVIZ_CONFIG_FILE=/path/to/config.yaml`
3. 目前目錄：`./config.yaml`
4. 合約覆蓋：`./contracts/config.yaml`
5. 樣本兜底：`./contracts/samples/config.yaml`

載入後套用預設值，並執行「**環境變數覆蓋**」：
- `DETECTVIZ_ENV` → `env`
- `DETECTVIZ__GRPC__{LISTEN,MAX_RECV_BYTES,MAX_SEND_BYTES}`
- `DETECTVIZ__OBSERVABILITY__MODE`
- `DETECTVIZ__OBSERVABILITY__OTLP__{PROTOCOL,ENDPOINT,INSECURE,HEADERS}`
- `DETECTVIZ__OBSERVABILITY__LOGS__FILE__{PATH,MAX_SIZE_MB,MAX_BACKUPS,MAX_AGE_DAYS,COMPRESS}`
- `DETECTVIZ__OBSERVABILITY__PROFILING__{ENABLED,PPROF_ADDRESS,APPLICATION_NAME,TAGS}`
- `DETECTVIZ__OBSERVABILITY__RESOURCE__{SERVICE_NAME,SERVICE_VERSION,ENVIRONMENT}`
- `DETECTVIZ__OBSERVABILITY__SAMPLING__RATIO`
- `DETECTVIZ__PLUGIN__{PATHS,REGISTRY}`
- `DETECTVIZ__MEMORY__{BACKEND,DSN,DEFAULT_TTL_SECONDS}`

最終以 `contracts/schemas/config.schema.json` 驗證，確保鍵位與型別一致。

---

## 健康檢查與優雅關機
- HTTP 健康檢查服務（預設 `:8081`，可用 `DETECTVIZ_HEALTH_ADDR` 覆蓋）：
  - `GET /livez`：存活檢查（程序存活即 200）
  - `GET /readyz`：就緒檢查（ToolBridge 成功啟動後返回 200）
- gRPC Health：註冊 `grpc.health.v1.Health` 服務，便於 gRPC 層探測。
- 優雅關機順序：
  1. 健康服務標記為 not ready
  2. 停止 HTTP Demo（若啟用）
  3. `ToolBridge` 執行 `GracefulStop`，超時再 `Stop`
  4. 關閉 listener 與 OTel provider

---

## Observability 與 Alloy 對齊
- 預設以 OTLP 匯出至 Alloy，再由 Alloy 導流到本地 Grafana Stack 或 Grafana Cloud。
- 建議：將應用日誌寫入檔案（zap）或 stdout，Alloy 以檔案/STDIN 介面收集並上送。
- 可於 Grafana 進行 Drilldown：Logs ↔ Traces ↔ Profiles。

關鍵設定（環境變數）：
```bash
export DETECTVIZ__OBSERVABILITY__MODE=lgtm_local   # 或 grafana_cloud / gcp
export DETECTVIZ__OBSERVABILITY__OTLP__ENDPOINT=127.0.0.1:4317
export DETECTVIZ__OBSERVABILITY__OTLP__PROTOCOL=grpc
export DETECTVIZ__OBSERVABILITY__OTLP__INSECURE=true
```

---

## ToolBridge（gRPC）
- 介面來源：`contracts/proto/detectviz/contracts/v1/adk_bridge.proto`
- 服務：`ToolBridge.Invoke(ToolInvokeRequest) returns (ToolInvokeReply)`
- 已註冊 gRPC Health 服務，便於 gRPC 層存活/就緒探測。

Python 端呼叫方式（示意）：
- 以 `python-adk-runtime` 的 `RemoteTool` 連線：`DETECTVIZ_TOOLBRIDGE_ADDR=127.0.0.1:6606`
- 支援 metadata：`tenant_id`、`owner.root_agent_id`、`traceparent`、`tracestate`

---

## Plugin 機制
- Registry 並發安全，提供：
  - `RegisterStrict(toolID, handler)`：**嚴格模式**，同名即回錯，不覆蓋
  - `RegisterOrReplace(toolID, handler)`：允許熱替換（會關閉舊 handler）
- 內建插件：
  - `capability.gateway/http_request` → 工具 ID：`detectviz.tools.http_request`
    - 支援 `method/url/headers/query/body/json/form/timeout_ms/max_response_bytes`
    - 以 `otelhttp` 攔截器自動產生外呼 span
- 模組卡（Module Card）：
  - 位置：`internal/pluginhost/plugins/capability.gateway/http_request/module.card.json`
  - 驗證：`cd contracts && make validate-cards`

---

## CLI 參數（節選）
- `--config`：指定設定檔路徑（預設依優先序自動尋找）
- `--http-demo`：啟用示範 HTTP 服務（含 otelhttp 儀表化）
- `--http-demo-listen`：示範 HTTP 監聽位址（預設 `:8080`）

常用環境變數：
```bash
# 健康服務位置
export DETECTVIZ_HEALTH_ADDR=":8081"

# ToolBridge gRPC
export DETECTVIZ__GRPC__LISTEN=":6606"
export DETECTVIZ__GRPC__MAX_RECV_BYTES=10485760
export DETECTVIZ__GRPC__MAX_SEND_BYTES=10485760

# Observability（OTLP 端點）
export DETECTVIZ__OBSERVABILITY__OTLP__ENDPOINT=127.0.0.1:4317
export DETECTVIZ__OBSERVABILITY__OTLP__PROTOCOL=grpc
export DETECTVIZ__OBSERVABILITY__OTLP__INSECURE=true
```

---

## 與 contracts 對齊
- SSOT：`contracts/`（proto／schemas／samples）。
- 修改跨語言契約 → 先改 `contracts/`，再於兩端重新產生生成碼：
  ```bash
  cd contracts
  make gen && make validate
  ```

---

## 疑難排解
- 無法連線 ToolBridge：
  - 檢查 `DETECTVIZ_TOOLBRIDGE_ADDR` 與 `DETECTVIZ__GRPC__LISTEN` 是否一致
  - 確認防火牆或容器網路對應端口已開放
- Logs 可見但無 Traces：
  - 檢查 OTLP 端點與 Alloy 設定；確認 `protocol` 與 `insecure` 是否匹配
- `/readyz` 一直 503：
  - 檢查 ToolBridge 是否成功啟動；查看應用日誌中的就緒訊息

---

## 安全注意
- 請勿在設定檔硬編密鑰；請使用環境變數或 Secret 管理。
- 生產環境建議啟用 TLS/mTLS（ToolBridge 與 Alloy 端）。

---

## 編譯與執行
```bash
go build -o ./bin/detectviz ./go-platform/cmd/detectviz
./bin/detectviz --config ./config.yaml --http-demo --http-demo-listen :8080
```