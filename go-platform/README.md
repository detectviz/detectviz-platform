# go-platform（Platform Core / ToolBridge）

最小化的平台核心，提供跨語言可用的 gRPC **ToolBridge**、可插拔 **Plugin Host**、統一 **Observability** 與 **健康檢查／優雅關機**。本模組與 `python-adk-runtime` 僅以 gRPC/HTTP 通訊，嚴格解耦並對齊 `contracts/` 的 SSOT 契約。

> **AI 開發者注意**：進行任何變更前請先閱讀 [`../AGENT.md`](../AGENT.md) 中的 AI 開發守則與協作指南。

---

## 平台職責
- **ToolBridge（gRPC）**：供 Python/ADK 的 `RemoteTool` 呼叫 Go 插件（工具執行）。
- **Plugin Host**：優化的插件註冊中心，支援健康狀態管理、負載保護、熔斷機制與 Prometheus 監控。
- **契約版本檢查**：啟動時自動驗證 proto 生成碼版本一致性，確保 SSOT 合規。
- **安全邊界**：內建 allowlist/denylist、payload 大小限制、超時控制等安全機制。
- **Observability**：OpenTelemetry 統一匯出（Logs/Traces/Metrics），預設透過 Alloy 導至本地 Grafana 或 Grafana Cloud。
- **健康檢查與優雅關機**：提供 `/livez`、`/readyz` 與 gRPC Health；啟動就緒與 SIGTERM/SIGINT 的有序關閉。

---

## 目錄結構
- `cmd/detectviz/`：平台啟動器（CLI）
- `internal/configx/`：設定載入與驗證（統一優先序 + 環境變數覆蓋）
- `internal/contracts/`：契約版本檢查與一致性驗證（SSOT 合規）
- `internal/metrics/`：統一指標查詢抽象層（支援 Prometheus、Memory、預留 Mimir）
- `internal/observability/`：OpenTelemetry 初始化與結構化日誌
- `internal/health/`：HTTP 健康檢查服務（/livez、/readyz）
- `internal/pluginhost/`：插件宿主（registry、runtime、ToolBridge 伺服端）
  - `registry.go`：優化的插件註冊中心，支援健康狀態管理和負載保護
  - `metrics.go`：Prometheus 監控整合
  - `plugins/gateway/http_request/`：HTTP 請求插件（含安全邊界）
  - `plugins/observability/health_aggregator/`：健康指標聚合插件
  - `resource_monitor.go`：插件級資源使用監控
  - `monitored_handler.go`：監控處理器包裝
- `tools/`：開發工具（插件腳手架、代碼生成器）

---

## 系統需求
- Go 1.24+
- 已產生 gRPC 生成碼（於 repo 根目錄執行 `cd contracts && make gen`）
- 契約版本驗證通過（啟動時自動檢查）
- Alloy 已就緒（本地或 Grafana Cloud）

---

## 快速開始（優化版啟動流程）
```bash
# 於 repo 根目錄建立生效設定（依 SSOT 樣本）
cp contracts/samples/config.yaml ./config.yaml

# 可選：覆蓋常用環境變數
export DETECTVIZ__OBSERVABILITY__OTLP__ENDPOINT=127.0.0.1:4317
export DETECTVIZ_HEALTH_ADDR=":8081"

# 啟動 ToolBridge 與示範 HTTP（優化版：模組化啟動、結構化錯誤處理）
go run ./cmd/detectviz plugin serve --config ./config.yaml \
  --http-demo \
  --http-demo-listen :7777

# 驗證服務健康狀態
curl -sS http://127.0.0.1:8081/readyz  # 等待服務就緒（200 OK）
curl -sS http://127.0.0.1:7777/hello   # 產生示範 traces
```

**優化特性**：
- **模組化啟動**：清晰的啟動序列和錯誤處理
- **詳細日誌**：啟動時間追蹤和狀態記錄
- **優雅關機**：10秒超時的有序服務關閉
- **Panic 恢復**：啟動過程異常處理
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

## 健康檢查與優雅關機（已優化）
- HTTP 健康檢查服務（預設 `:8081`，可用 `DETECTVIZ_HEALTH_ADDR` 覆蓋）：
  - `GET /livez`：存活檢查（程序存活即 200）
  - `GET /readyz`：就緒檢查（ToolBridge 成功啟動後返回 200）
- gRPC Health：註冊 `grpc.health.v1.Health` 服務，便於 gRPC 層探測。
- **優化版優雅關機順序**（10秒超時保護）：
  1. 收到 SIGTERM/SIGINT 信號
  2. 健康服務標記為 not ready
  3. 停止 HTTP Demo 服務（若啟用）
  4. `ToolBridge` 執行 `GracefulStop`，超時保護
  5. 關閉可觀測性系統（OTel flush）
  6. 關閉健康檢查服務
  7. 記錄總運行時間並完成關機

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
- 以 `python-adk-runtime` 的 `RemoteTool` 連線：`DETECTVIZ_TOOLBRIDGE_ADDR=127.0.0.1:5002`
- 支援 metadata：`tenant_id`、`owner.root_agent_id`、`traceparent`、`tracestate`

---

## Plugin 機制
- Registry 並發安全，提供：
  - `RegisterStrict(toolID, handler)`：**嚴格模式**，同名即回錯，不覆蓋
  - `RegisterOrReplace(toolID, handler)`：允許熱替換（會關閉舊 handler）
- 內建插件：
  - `gateway/http_request` → 工具 ID：`detectviz.tools.http_request`
    - 支援 `method/url/headers/query/body/json/form/timeout_ms/max_response_bytes`
    - 以 `otelhttp` 攔截器自動產生外呼 span
  - `observability/health_aggregator` → 工具 ID：`observability.health_aggregator`
    - 整合 MetricsProvider 抽象層，支援多種監控數據源
    - 提供指標查詢、聚合、快取和統計計算功能
- 模組卡（Module Card）：
  - 位置：`internal/pluginhost/plugins/<category>/<name>/module.card.json`
  - 驗證：`cd contracts && make validate-cards`

### 可靠性與負載保護 (Reliability and Load Protection)

Plugin Host 內建多種機制以確保系統穩定性與資源隔離：

- **Prometheus 指標提供者熔斷 (Circuit Breaker)**：
  - **目的**：當後端 Prometheus 服務不穩定或超載時，自動「熔斷」，暫停發送查詢，避免連鎖故障。
  - **機制**：採用狀態機（關閉 `Closed` → 開啟 `Open` → 半開 `Half-Open`）管理。
    - `Closed`：正常狀態，允許查詢。
    - `Open`：連續查詢失敗達到閾值後進入此狀態，所有查詢會立即失敗，直到恢復超時。
    - `Half-Open`：超時後進入，允許少量測試查詢。若成功則轉換回 `Closed`，若失敗則再次進入 `Open`。
  - **設定 (`config.yaml`)**：
    ```yaml
    metrics_provider:
      prometheus:
        circuit_breaker:
          enabled: true           # 是否啟用
          failure_threshold: 5    # 連續失敗幾次後開啟熔斷
          recovery_timeout: 30s   # 開啟後多久嘗試恢復 (進入半開狀態)
    ```

- **並發控制 (Concurrency Limiter)**：限制對 Prometheus 的同時查詢數量，防止其過載。
- **查詢快取 (Query Caching)**：快取重複的指標查詢結果，降低延遲並減少後端負載。

---

## 插件開發

### 創建新插件
使用內建腳手架工具快速創建插件骨架：

```bash
# 在 go-platform 目錄下執行
go run tools/scaffold.go <category>/<name>

# 範例：創建健康檢查插件
go run tools/scaffold.go observability/system_health

# 範例：創建 API 調用插件
go run tools/scaffold.go gateway/api_client
```

### 支援的插件類別
- `gateway`：能力閘道插件（HTTP 請求、API 調用）
- `observability`：可觀測性插件（健康檢查、指標聚合）
- `collector.input`：數據收集插件（日誌收集、指標採集）
- `transform.processor`：數據處理插件（數據轉換、過濾）
- `sink.output`：數據輸出插件（數據庫寫入、消息發送）

### 插件開發流程
1. **生成腳手架**：`go run tools/scaffold.go <category>/<name>`
2. **實作邏輯**：編輯 `plugin.go` 實作 `Invoke` 方法
3. **配置模組**：更新 `module.card.json` 設定
4. **註冊插件**：在 `internal/pluginhost/plugins/register/all.go` 中註冊
5. **編寫測試**：完善單元測試和整合測試
6. **更新文檔**：補充 README 和使用範例

詳細開發指南請參考：[`tools/README.md`](tools/README.md)

---

## CLI 參數（節選）
- `--config`：指定設定檔路徑（預設依優先序自動尋找）。
- `--http-demo`：啟用示範 HTTP 服務（含 otelhttp 儀表化）。
- `--http-demo-listen`：示範 HTTP 監聽位址（預設 `:7777`）。

**注意**：TLS/mTLS 相關設定（憑證路徑等）已移至 `config.yaml` 的 `grpc.tls` 區塊中進行統一管理，不再支援透過 CLI 旗標設定。

常用環境變數：
```bash
# 健康服務位置
export DETECTVIZ_HEALTH_ADDR=":8081"

# ToolBridge gRPC
export DETECTVIZ__GRPC__LISTEN=":5002"
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

---

## 開發與維護

### AI 協作開發
- **開發守則**：請參考 [`../AGENT.md`](../AGENT.md) 中的完整 AI 開發守則
- **提交前檢查**：請參考 `go-platform/llm.txt`（技術特定檢查清單）
- **契約變更**：任何 gRPC 介面變更需先在 `../contracts/` 進行

### 故障排查
詳細故障排查指南請參考：[`../AGENT.md - 常見錯誤與排查`](../AGENT.md#常見錯誤與排查)