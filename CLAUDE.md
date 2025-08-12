# AI 開發守則

## 目的
本文件提供給 AI 與人類協作者的操作與開發守則，確保在 Detectviz 平台上以 **SSOT（contracts）為唯一事實來源** 進行變更；同時規範雲端觀測（Grafana Cloud / GCP）、本地 LGTM、與程式碼變更的安全與品質要求。本文已全面對齊目前程式碼與 `spec.md`。

---

## SSOT 與提交規範
- **契約先行**：任何跨語言介面或設定變更，請先修改 `contracts/`（proto/schema/samples）。
  - Proto：`contracts/proto/detectviz/contracts/v1/adk_bridge.proto`
  - Schema：`contracts/schemas/{config.schema.json,module.card.schema.json}`
  - 樣本：`contracts/samples/config.yaml`
- 生成與驗證：
  ```bash
  cd contracts
  buf lint && buf generate
  make validate            # 若已提供
  make validate-cards      # 驗證所有 module.card.json
  ```
- **禁止**手動編輯任何生成碼（例如 `*.pb.go`、`*_pb2.py`、`*_pb2_grpc.py`）。
- 下游（Go / Python）在更新契約後，請同步調整 import 與 `go mod tidy`（Go）。

---

## 設定與啟動（統一搜尋優先序）
**搜尋順序（高→低）**：
1. 旗標/參數：`--config /path/to/config.yaml`
2. 環境變數：`DETECTVIZ_CONFIG_FILE=/path/to/config.yaml`
3. 工作目錄：`./config.yaml`
4. 合約覆蓋：`./contracts/config.yaml`
5. 樣本兜底：`./contracts/samples/config.yaml`

**建議用法**：
```bash
# 於 repo 根目錄
cp contracts/samples/config.yaml ./config.yaml
```

**環境變數覆蓋（節選，Go/Python 對齊）**：
- 一般：`DETECTVIZ_ENV`
- gRPC：`DETECTVIZ__GRPC__{LISTEN,MAX_RECV_BYTES,MAX_SEND_BYTES}`
- OTLP：`DETECTVIZ__OBSERVABILITY__OTLP__{PROTOCOL,ENDPOINT,INSECURE,HEADERS}`
- 資源：`DETECTVIZ__OBSERVABILITY__RESOURCE__{SERVICE_NAME,SERVICE_VERSION,ENVIRONMENT}`
- 取樣：`DETECTVIZ__OBSERVABILITY__SAMPLING__RATIO`
- 日誌：`DETECTVIZ__OBSERVABILITY__LOGS__FILE__{PATH,MAX_SIZE_MB,MAX_BACKUPS,MAX_AGE_DAYS,COMPRESS}`
- pprof：`DETECTVIZ__OBSERVABILITY__PROFILING__{ENABLED,PPROF_ADDRESS,APPLICATION_NAME,TAGS}`
- 插件：`DETECTVIZ__PLUGIN__{PATHS,REGISTRY}`
- 記憶體：`DETECTVIZ__MEMORY__{BACKEND,DSN,DEFAULT_TTL_SECONDS}`
- Python RemoteTool：`DETECTVIZ_TOOLBRIDGE_ADDR`、`DETECTVIZ_TOOLBRIDGE_INSECURE`、`DETECTVIZ_TOOLBRIDGE_TLS_{CERT,KEY,CA}`

`.env` 範例（Grafana Cloud）：
```bash
# OTLP（Tempo/Mimir 綜合 OTLP Gateway）使用 Basic Auth
export GF_CLOUD_OTEL_ID="__STACK_ID__"          # Basic Auth username（同 Stack ID）
export GCLOUD_RW_API_KEY="__OTEL_RW_TOKEN__"    # Basic Auth password

# Loki（Logs）
export GCLOUD_HOSTED_LOGS_URL="https://logs-prod-<region>.grafana.net/loki/api/v1/push"
export GCLOUD_HOSTED_LOGS_ID="__STACK_ID__"

# Pyroscope（Profiles）
export GF_CLOUD_PROFILES_URL="https://profiles-<region>.grafana.net"
export GF_CLOUD_PROFILES_ID="__STACK_ID__"

# 可選覆蓋（本機路徑/位址）
export DETECTVIZ_LOG_PATH="/abs/path/to/var/log/detectviz/detectviz.log"
export DETECTVIZ_PPROF_ADDR="127.0.0.1:6060"
```

---

## 快速啟動（本地 LGTM 或 Grafana Cloud）
1. **套用 SSOT 樣本組態**：`cp contracts/samples/config.yaml ./config.yaml`
2. **啟動 Alloy**：`./grafana-alloy/alloy run ./grafana-alloy/config.alloy`
3. **啟動 Go 平台（ToolBridge + http-demo）- 優化版**：
   ```bash
   # 優化後的啟動流程，包含詳細日誌和優雅關機
   go run ./go-platform/cmd/detectviz plugin serve --config ./config.yaml \
     --http-demo --http-demo-listen :7777
   
   # 驗證服務健康狀態
   curl -sS http://127.0.0.1:8081/readyz  # 等待服務就緒
   curl -sS http://127.0.0.1:7777/hello   # 產生 Traces
   ```
4. **在 Grafana 檢視**：Explore/Logs、Traces、Profiles 可 Drilldown。

---

## Profiles 與 Drilldown 完整性
- **僅支援 pprof**：Go 端以 `observability.profiling` 啟動 `net/http/pprof`，預設 `127.0.0.1:6060`。
- Alloy 使用 `pyroscope.scrape` → `pyroscope.write` 上傳（本地或雲端）。
- Logs ↔ Traces 關聯：在 Alloy 的 Loki 管線以 regex 萃取 `trace_id` → 標籤 `traceid`。

---

## AI 開發工作流程守則
**每次程式碼變更時，AI 必須遵循以下檢查清單**：

### 1. 變更前規劃
- [ ] 識別變更類型：SSOT 契約、核心邏輯、介面變更、內部重構
- [ ] 評估影響範圍：使用者介面、系統行為、文檔、範例
- [ ] 建立 TODO 清單，**必須包含文檔更新任務**

### 2. 實作過程中
- [ ] 遵循 SSOT 原則，契約變更優先
- [ ] 保持向後相容性，除非明確說明破壞性變更
- [ ] 記錄重要的設計決策和權衡考量

### 3. 完成後檢查
- [ ] 編譯和基本功能測試
- [ ] 檢查是否需要更新文檔（參考上述檢查清單）
- [ ] 驗證所有範例指令和配置仍然有效
- [ ] 確認變更符合平台設計原則

### 4. 文檔同步更新
- [ ] 根據變更類型更新相應文檔
- [ ] 更新快速開始指南中的指令
- [ ] 檢查所有文檔間的一致性

**重要提醒**：AI 在進行任何重大變更時，應主動詢問是否需要更新文檔，而不是等使用者提醒。

---

## Go 平台實作守則（AI 請遵守）
- **設定載入**：一律透過 `internal/configx/loader.go`；**禁止**在其他模組自行讀檔或解析 YAML。
- **日誌**：使用 `zap`，寫入純文字檔案（非 JSON），預設 `./var/log/detectviz/detectviz.log`；初始化時先建目錄並 `zap.ReplaceGlobals`。
- **OTEL**：依 `observability.otlp` 初始化 Traces/Metrics；Headers 以 `key=value` CSV 由 Loader 解析。
- **pprof**：僅在 `observability.profiling.enabled: true` 啟動；位址 `profiling.pprof_address`。
- **健康檢查**：提供 `/livez`、`/readyz` 與 gRPC Health；ToolBridge 就緒後再宣告 Ready。
- **優雅關機**：收到 SIGTERM/SIGINT → 先標記 not-ready → 停 HTTP demo → `GracefulStop` gRPC → 關閉 OTel。
- **錯誤處理**：關鍵服務（ToolBridge/OTLP 初始化/健康服務）失敗**不得靜默**；啟動期錯誤需回傳並中止。
- **Plugin Registry（並發安全）**：以 `sync.RWMutex` 保護；提供 `RegisterStrict(toolID, h) error`，同名即報錯，不覆蓋；`RegisterOrReplace` 允許熱替換並確保舊實例 `Close`。
- **ToolBridge 介面**：依 `adk_bridge.proto`；不得修改生成碼；如需串流或會話化，請先於 `contracts/` 設計。

---

## Python ADK Runtime 守則
- **ADK 對齊**：遵守 Agent / Workflow / MemoryBank / Tools / Capabilities 模組邊界。
- **RemoteTool**：透過 `grpc.aio` 呼叫 ToolBridge；端點以 `DETECTVIZ_TOOLBRIDGE_ADDR` 設定；支援 `traceparent/tracestate` 注入。
- **設定載入**：使用 `detectviz_adk/config/loader.py`，與 Go 相同的搜尋序與環境覆蓋。
- **模組卡**：新增/擴增元件需附 `module.card.json` 並通過 `contracts/tools/validate_module_card.py`。
- **安全**：Python 不持有雲端憑證；外部系統交互集中於 Go 插件（Tools）。

---

## 測試與驗證清單
- 契約：`cd contracts && buf lint && buf generate && make validate-cards`
- 啟動：`go run ./go-platform/cmd/detectviz --config ./config.yaml --http-demo`
- 健康：檢查 `GET /livez` 與 `GET /readyz`（Ready 需等 ToolBridge 成功啟動）
- Drilldown：Grafana Explore 檢查 Logs ↔ Traces ↔ Profiles 是否關聯
- 關機：發送 SIGTERM，確認優雅關機順序與 OTel flush 成功
- 插件：測試 `RegisterStrict` 重複註冊回錯；`RegisterOrReplace` 能熱替換且釋放資源

---

## 常見錯誤與排查
- `config validation failed: ... profiling.* not allowed`：`config.yaml` 殘留舊欄位（如 `mode/endpoint/username/password`）；請以最新 Schema 修正。
- `plugin.paths: Invalid type. Expected: array`：請確認是 YAML 陣列（即使只有一個路徑也需 `[...]`）。
- `could not import grpc/codes` 或 `no required module provides package ...`：在 `contracts/` 重新 `buf generate`，並於 Go 端 `go mod tidy`。
- Logs 無輸出：確認 `zap.ReplaceGlobals`、日誌目錄存在、且未使用 `log.Printf`。
- Traces 缺失：檢查 OTLP endpoint/協定（grpc/http）與 Alloy 狀態。

---

## 安全與機密
- **嚴禁**在任何檔案（含 README、程式碼、日誌）寫入真實 Token / 密碼 / 秘鑰；請使用環境變數或 Secret。
- 若曾誤提交，請立刻撤銷、輪替並於 PR 附上修復步驟。

---

## 變更要求（PR 模板建議）
1. 說明對 SSOT 的變更（proto/schema/samples）與影響面
2. 附端到端驗證步驟（Logs/Traces/Profiles 與健康/關機）
3. **強制文檔更新檢查清單**：
   - [ ] 檢查是否影響使用者介面（CLI 參數、環境變數、啟動流程）
   - [ ] 檢查是否影響架構或系統行為（啟動序列、錯誤處理、關機流程）
   - [ ] 檢查是否需要更新：`spec.md`、根 `README.md`、子專案 `README.md`、本檔
   - [ ] 驗證所有文檔中的範例指令和配置是否仍然有效
   - [ ] 確認新功能在快速開始指南中有適當說明

**文檔更新優先級**：
- **P0（必須）**：影響使用者操作的變更（CLI、環境變數、啟動指令）
- **P1（重要）**：架構或行為變更（錯誤處理、關機流程、性能優化）
- **P2（建議）**：內部實作細節（代碼結構、函數重構）

---

## 文檔與溝通風格規範
- **禁用 Emoji**：所有文檔、程式碼註解、提示詞和溝通內容都不使用 emoji 符號
- **專業文字表達**：使用清晰的文字描述代替視覺符號，保持專業和正式的技術文檔風格
- **標題標記**：使用傳統的標題層級（#、##、###）和項目符號（-、*）來組織內容結構
- **狀態表示**：使用文字描述狀態，如「已完成」、「進行中」、「待處理」而非符號標記
- **重點強調**：使用 **粗體** 和 `程式碼格式` 來突出重要內容，而非 emoji 裝飾