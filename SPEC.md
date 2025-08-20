# Detectviz 平台技術規格

> **文檔職責**：定義技術實作規格、契約定義、配置管理和開發者實作指南
> **⚠️ 重要提醒**：本文檔正在重構中，請優先參考 contracts/ 目錄的 proto 定義和各模組的 README.md

## 文檔定位

- **目標受眾**：開發者、實作者、系統整合工程師
- **更新頻率**：隨重大功能變更同步更新
- **版本**：2.0.0
- **最後更新**：2025-08-19

## 文檔關係

```bash
README.md → AGENT.md → ARCHITECTURE.md → [SPEC.md] → TASKS.md
```

**閱讀路徑**：
- **前置閱讀**：[ARCHITECTURE.md - 系統架構設計](ARCHITECTURE.md#系統架構設計)
- **後續閱讀**：[TASKS.md - 開發任務進度](TASKS.md#實作任務清單)
- **相關參考**：[AGENT.md - AI協作指南](AGENT.md#ai協作原則)
- **實際契約**：[contracts/ - Protocol Buffers SSOT](contracts/)

> **架構決策與設計原則**：參見 [ARCHITECTURE.md - 系統架構設計](ARCHITECTURE.md#系統架構設計)

## 當前實作狀態

根據 [TASKS.md](TASKS.md) 的最新稽核結果：
- **MVP Phase 3 完成度**：95% ✅ (所有核心功能已驗證)
- **專案狀態**：準備交付階段
- **技術實作**：Go 平台核心 + Python ADK Runtime 雙架構已穩定運行

## 技術棧現狀

> **完整技術棧概覽**：參見 [README.md - 技術棧](README.md#技術棧)

### 核心技術組件
- **Go Platform Core** (go-platform/)
- **Python ADK Runtime** (python-adk-runtime/) 
- **Contracts SSOT** (contracts/)
- **依賴服務** (deploy/)

### 關鍵實作成果
- ✅ **MetricsProvider 抽象層** - 統一指標查詢介面
- ✅ **gRPC 跨語言通訊** - ToolBridge 服務
- ✅ **知識庫管理** - PostgreSQL 基礎的持久化存儲
- ✅ **智能 Agent 協作** - 多層級決策架構

* * *

## 實際專案結構

### 根目錄結構
```bash
detectviz-platform/
├── contracts/           # 📋 SSOT - 跨語言契約定義 (重要)
├── go-platform/         # ⚡ Go 平台核心
├── python-adk-runtime/  # 🧠 Python ADK Runtime
├── deploy/              # 🐳 依賴服務 Docker Compose
├── docs/                # 📚 文檔體系
└── scripts/             # 🔧 自動化腳本
```

### 契約與配置目錄（contracts/）

**實際結構** (基於 contracts/ 的存在)：
```bash
contracts/
├── proto/detectviz/contracts/v1/  # Protocol Buffers 定義
│   └── adk_bridge.proto           # ToolBridge 核心服務定義
├── schemas/                       # JSON Schema 驗證規範
├── gen/                           # 自動生成的程式碼
│   ├── go/                        # Go 生成碼
│   └── python/                    # Python 生成碼
├── buf.yaml                       # Buf 工作區配置
├── buf.gen.yaml                   # 程式碼生成配置
└── Makefile                       # 建置與版本控制
```

> **重要提醒**：詳細的契約定義請直接參考 `contracts/` 目錄，該目錄為 SSOT (Single Source of Truth)

### 實際目錄結構說明

```bash
├── docs/                           # 📚 技術文檔目錄
│   ├── development/                # 開發相關文檔
│   ├── guides/                     # 使用指南
│   └── status/                     # 專案狀態文檔
├── deploy/                         # 🐳 Docker Compose 服務
│   ├── grafana/                    # Grafana 配置
│   ├── prometheus/                 # Prometheus 配置
│   └── docker-compose.yaml         # 服務編排
├── scripts/                        # 🔧 自動化腳本
└── assets/                         # 靜態資源 (如存在)
```

> **注意**：具體的目錄結構請參考各組件的 README.md 文檔

### Go 平台核心（go-platform/）

**當前已驗證的結構**：
```bash
go-platform/
├── cmd/detectviz/                     # 主程式入口
├── internal/                          # 內部套件
│   ├── configx/                       # 配置管理
│   ├── health/                        # 健康檢查服務
│   ├── metrics/                       # ✅ 指標管理 (MetricsProvider)
│   ├── observability/                 # 可觀測性整合
│   └── pluginhost/                    # ✅ 插件託管系統
│       ├── bridge_server.go           # gRPC ToolBridge 服務
│       └── plugins/                   # 插件實作
├── configs/                           # 配置檔案
└── tools/                             # 開發工具
```

> **實際詳情**：請參考 [go-platform/README.md](go-platform/README.md) 獲取最新結構說明

### Python ADK 運行時（python-adk-runtime/）

**當前已驗證的結構**：
```bash
python-adk-runtime/
├── src/detectviz_adk/               # ✅ 核心 ADK 套件
│   ├── agents/                      # ✅ Agent 實作
│   ├── tools/                       # ✅ 工具抽象層
│   ├── memory/                      # ✅ 記憶體與知識管理
│   ├── config/                      # 配置管理
│   ├── runners/                     # 執行器
│   └── sessions/                    # 會話管理
├── tests/                           # ✅ 測試目錄 (已驗證)
├── agents/                          # Agent 實作範例
├── requirements.txt                 # Python 依賴清單
├── example_usage.py                 # 使用範例
└── web_server.py                    # 開發用 Web 服務
```

**關鍵實作成果** (根據 TASKS.md 稽核)：
- ✅ **RemoteTool** - gRPC 跨語言工具調用
- ✅ **Agent 狀態管理** - Redis + 記憶體雙重後備
- ✅ **報告生成引擎** - Jinja2 模板系統
- ✅ **知識庫整合** - PostgreSQL 持久化存儲

> **實際詳情**：請參考 [python-adk-runtime/README.md](python-adk-runtime/README.md) 獲取最新結構說明

* * *

## 契約定義 (SSOT)

> **重要提醒**：所有 API 定義以 `contracts/` 目錄為準，此處僅提供概要說明

### 實際 gRPC 服務

根據當前 contracts/proto/detectviz/contracts/v1/adk_bridge.proto：

- **HealthService** - 健康檢查服務
- **ToolBridgeService** - 跨語言工具調用

### HTTP 服務

根據 go-platform/internal/health/ 的實作：

```http
GET /health       # 基本健康檢查
GET /health/ready # 就緒檢查  
GET /health/live  # 存活檢查
GET /version      # 版本資訊
```

**使用方式**：
```bash
# 檢查服務狀態
curl http://localhost:8080/health

# 檢查 gRPC 服務
grpcurl -plaintext localhost:5002 detectviz.contracts.v1.HealthService/Check
```

### 契約管理

**實際設定** (位於 contracts/)：
```bash
# 生成所有契約程式碼
make gen

# 驗證契約一致性
make health-check-proto

# 維護契約品質
make maintain-proto
```

**產生的程式碼**：
- **Go 版本**：`contracts/gen/go/detectviz/contracts/v1/`
- **Python 版本**：`contracts/gen/python/detectviz/contracts/v1/`

> **詳細配置**：請參考 [contracts/README.md](contracts/README.md) 和 `contracts/buf.gen.yaml`

## 配置管理

### 實際配置文件

**主要配置位置**：
- `config.yaml` - 主要平台配置
- `.env` / `deploy/.env` - 環境變數
- `deploy/` - Docker Compose 服務配置

### 核心配置項目

```yaml
# config.yaml 簡化版本
grpc:
  listen: ":5002"
  
http:
  listen: ":8080"
  
observability:
  mode: "lgtm_local"  # 或 grafana_cloud
  
# HTTP Demo 功能 (用於測試)
http:
  demo:
    enabled: true
    path: "/business"
    port: ":7777"
```

### 環境變數範例

```bash
# 基本配置
DETECTVIZ_ENV=development

# Demo 功能
DETECTVIZ_HTTP_DEMO=1  # 啟用 HTTP Demo

# 可觀測性配置
DETECTVIZ__OBSERVABILITY__MODE=lgtm_local
DETECTVIZ__GRPC__LISTEN=:5002
DETECTVIZ__OBSERVABILITY__OTLP__ENDPOINT=localhost:4317
```

> **完整配置範例**：請參考 [.env.example](.env.example) 文件

### 配置驗證

**主要驗證指令** (在專案根目錄執行)：

```bash
# 驗證全部實作
./detectviz config validate -f config.yaml

# 驗證契約一致性
make health-check-proto

# 完整驗證流程
make validate-implementation
```

> **進階工具**：請參考 `contracts/tools/` 目錄中的驗證工具

* * *

## MVP 現狀 (Phase 3 完成)

> **當前進度**：根據 [TASKS.md](TASKS.md) 稿核結果，**MVP Phase 3 已 95% 完成並驗證**，系統處於可交付狀態

### 已驗證的核心功能

#### 1. 事後複盤協調器（Python ADK Agent）

**實作狀態**：✅ 已完成並驗證 (根據 [TASKS.md 稽核結果](TASKS.md#專案狀態概覽))

**檔案位置**：[python-adk-runtime/src/detectviz_adk/agents/postmortem/orchestrator.py](python-adk-runtime/src/detectviz_adk/agents/postmortem/orchestrator.py)

**核心功能**：
- 多層級決策架構，符合 [ARCHITECTURE.md Agent 協作模式](ARCHITECTURE.md#agent-協作模式)
- 委派式任務執行：數據收集 → 根因分析 → 報告生成
- 智能策略決策：根據事件嚴重程度調整分析深度
- 完整錯誤處理與狀態追蹤

#### 2. 健康數據聚合器（Go Plugin）

**實作狀態**：✅ 已完成並驗證 (根據 [TASKS.md 稽核結果](TASKS.md#專案狀態概覽))

**檔案位置**：[go-platform/internal/pluginhost/plugins/observability/health_aggregator/plugin.go](go-platform/internal/pluginhost/plugins/observability/health_aggregator/plugin.go)

**核心功能**：
- 使用 [MetricsProvider 抽象層](ARCHITECTURE.md#metricsprovider-架構)，已成功移除 InfluxDB 依賴
- Prometheus 指標查詢與聚合，支援多種指標類型 (CPU, Memory, Network 等)
- 並行查詢優化，支援大規模服務監控
- 完整的錯誤處理、快取機制和斷路器模式
- 符合 [ARCHITECTURE.md 工具層規範](ARCHITECTURE.md#工具層規範)

### 實作最佳實踐

#### 1. 錯誤處理策略

**實作位置**：[go-platform/internal/pluginhost/plugins/](go-platform/internal/pluginhost/plugins/) - 所有插件均已實作統一錯誤處理

**標準化特性**：
- 結構化錯誤類型 (PluginError)
- 錯誤代碼分類 (INVALID_CONFIG, QUERY_FAILED, TIMEOUT, INTERNAL_ERROR)
- 結合 [ARCHITECTURE.md 開發指導原則](ARCHITECTURE.md#開發指導原則)

#### 2. 效能優化

**實作狀態**：✅ 已在 [HealthAggregatorPlugin](go-platform/internal/pluginhost/plugins/observability/health_aggregator/) 中實作

**優化特性**：
- 並行查詢支援 (queryMetricsParallel)
- Connection Pooling 管理
- 查詢結果快取 (CacheEntry TTL)
- 斷路器模式，防止級聯失敗

#### 3. 快取策略

**實作狀態**：✅ 已整合在 [MetricsProvider](go-platform/internal/metrics/) 結構中

**快取特性**：
- TTL 基朮的結果快取 (CacheEntry)
- 自動過期清理機制
- 查詢結果優先使用快取，減少 Prometheus 負載
- 符合 [MetricsProvider 抽象層設計](ARCHITECTURE.md#metricsprovider-架構)

* * *

## 工具與指令

### 主要指令 (已驗證)

```bash
# 啟動平台服務
./detectviz plugin serve --config config.yaml

# 配置驗證 
./detectviz config validate -f config.yaml

# 契約管理
make gen                    # 生成所有契約程式碼
make health-check-proto     # 驗證契約一致性
make maintain-proto         # 維護契約品質

# 開發環境
make setup-development      # 設置開發環境  
make validate-implementation # 完整實作驗證
make fix-common-issues      # 修復常見問題
```

> **完整指令清單**：請參考 [AGENT.md - 自動化檢查工具](AGENT.md#自動化檢查工具)

### 服務啟動選項

```bash
detectviz plugin serve \
  --config ./config.yaml \              # 配置檔案路徑
  --listen :5002 \                      # gRPC 監聽位址
  --http-listen :8080 \                 # HTTP 監聽位址
  --log-level info \                    # 日誌等級
  --enable-pprof \                      # 啟用效能分析
  --mtls-cert ./certs/server.crt \      # mTLS 憑證（可選）
  --mtls-key ./certs/server.key \       # mTLS 私鑰（可選）
  --mtls-ca ./certs/ca.crt              # mTLS CA 憑證（可選）
```

### 啟動流程優化

**模組化啟動序列** (已在 [main.go](go-platform/cmd/detectviz/main.go) 實作)：
1. **參數解析** - CLI 參數和環境變數驗證
2. **配置載入** - [configx](go-platform/internal/configx/) 配置檔案驗證
3. **契約檢查** - Protocol Buffer 版本一致性檢查
4. **日誌初始化** - 結構化日誌系統 (uber/zap)
5. **可觀測性** - [observability](go-platform/internal/observability/) OpenTelemetry 整合
6. **插件系統** - [pluginhost](go-platform/internal/pluginhost/) 載入和註冊
7. **服務啟動** - gRPC (port 5002) + HTTP (port 8080) 服務
8. **健康檢查** - [health](go-platform/internal/health/) 端點

**錯誤處理**: 已在 [Server.Start()](go-platform/cmd/detectviz/main.go) 中實作完整的階段性啟動與 panic 恢復機制

## 部署指南 (簡化版)

### 快速開始

**方法 1：体驗完整平台** (推薦)
```bash
cd deploy
cp ../.env.example .env
make start  # 啟動所有服務

# 啟動主程式 (另一個終端機)
source deploy/.env && DETECTVIZ_HTTP_DEMO=1 ./detectviz plugin serve --config config.yaml
```

**方法 2：本地開發**
```bash
make setup-development      # 設置開發環境
make health-check-proto     # 驗證 proto 檔案
make gen                    # 生成程式碼
make validate-implementation # 完整驗證
```

> **詳細指南**：請參考 [README.md 快速開始](README.md#快速開始-quick-start)

### 生產環境部署

**實際部署方式** (已驗證)：
```bash
# 使用 Docker Compose (推薦)
cd deploy
cp ../.env.example .env
make start

# 本地二進制部署
make setup-development
./detectviz plugin serve --config config.yaml
```

> **詳細部署指南**：請參考 [README.md 快速開始](README.md#快速開始-quick-start) 和 `deploy/` 目錄

## 故障排除 (簡化版)

### 常見問題與解決方案

#### 1. 服務無法啟動
```bash
# 檢查 proto 版本一致性
make health-check-proto

# 驗證完整實作
make validate-implementation
```

#### 2. gRPC 連線問題  
```bash
# 檢查服務狀態
curl http://localhost:8080/health
grpcurl -plaintext localhost:5002 detectviz.contracts.v1.HealthService/Check
```

#### 3. 環境相關問題
```bash
# 修復常見問題
make fix-common-issues
```

> **完整故障排除**：請參考各模組的 README.md 和 [AGENT.md 故障排除](AGENT.md#故障排除)

## 開發擴展指南 (簡化版)

### 核心擴展原則

**基於已驗證架構擴展**：
1. **Agent 開發** - 遵循 [ARCHITECTURE.md Agent 協作模式](ARCHITECTURE.md#agent-協作模式)
2. **Tool 開發** - 實作 ToolBridge 介面，參考現有插件
3. **契約擴展** - 優先更新 `contracts/` 目錄中的 proto 定義
4. **測試驗證** - 使用 `make validate-implementation` 確保品質

### 新增功能步驟

**標準開發流程** (參考 [AGENT.md 工作流程規範](AGENT.md#工作流程規範))：

1. **契約優先**: 更新 [contracts/](contracts/) proto 定義 → `make gen`
2. **Go/Python 實作**: 參考現有模組 [go-platform/](go-platform/) 與 [python-adk-runtime/](python-adk-runtime/)
3. **完整驗證**: `make validate-implementation` 確保品質
4. **文檔同步**: 更新相應 README.md 與設計文檔

> **詳細開發指南**：請參考 [AGENT.md 開發指導原則](AGENT.md#開發指導原則)

---

## 重要提醒

**本文檔正在重構中**，請優先參考：

1. **契約 SSOT**：[contracts/ 目錄](contracts/) - 最權威的 API 定義
2. **實際實作**：各組件的 README.md - 最新的實作狀態
3. **開發指南**：[AGENT.md](AGENT.md) - AI 協作規範
4. **系統架構**：[ARCHITECTURE.md](ARCHITECTURE.md) - 設計決策
5. **任務狀態**：[TASKS.md](TASKS.md) - 最新進度