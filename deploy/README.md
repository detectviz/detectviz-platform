<<<<<<< HEAD
# 建置指南

## 📦 依賴服務總覽

### 1. **核心監控服務**
- **Prometheus** - 指標存儲（短期 15-30 天）
- **Grafana** - 視覺化和告警管理
- **Grafana Alloy** - 統一的可觀測性收集器

### 2. **數據存儲服務**
- **PostgreSQL** - 知識庫存儲（事故記錄、教訓學習）
- **Redis** - 狀態管理和快取

### 3. **可觀測性後端（LGTM Stack）**
- **Loki** - 日誌存儲
- **Tempo** - 分散式追蹤
- **Pyroscope** - 性能分析（CPU/Memory profiling）

### 4. **監控 Exporters**
- **Node Exporter** - 主機指標
- **Postgres Exporter** - PostgreSQL 指標
- **Redis Exporter** - Redis 指標

### 5. **未來擴展（預留）**
- **Mimir** - 長期指標存儲（暫時註解，未來需要時啟用）

## 🚀 快速開始

### 1. 初始化環境
```bash
# 賦予執行權限
chmod +x scripts/setup.sh

# 運行設置腳本（選擇 1 進行完整設置）
./scripts/setup.sh

# 或使用 Makefile
make setup
```

### 2. 啟動服務
```bash
# 使用 Docker Compose
docker-compose up -d

# 或使用 Makefile
make start
```

### 3. 驗證服務健康
```bash
# 檢查服務狀態
make status

# 查看服務日誌
make logs
```

## 🌐 服務訪問地址

| 服務 | 地址 | 預設帳密 |
|------|------|----------|
| **Grafana** | http://localhost:3000 | admin/admin123 |
| **Prometheus** | http://localhost:9090 | - |
| **Alloy UI** | http://localhost:12345 | - |
| **Loki** | http://localhost:3100 | - |
| **Tempo** | http://localhost:3200 | - |
| **Pyroscope** | http://localhost:4040 | - |
| **PostgreSQL** | localhost:5432 | detectviz/detectviz123 |
| **Redis** | localhost:6379 | - |

## 📁 目錄結構

```bash
detectviz-platform/
├── docker-compose.yml          # Docker Compose 配置
├── Makefile                    # 開發命令
├── scripts/
│   └── setup.sh                # 環境設置腳本
├── prometheus/
│   └── prometheus.yml          # Prometheus 配置
├── grafana/
│   ├── provisioning/
│   │   ├── datasources/        # 數據源配置
│   │   └── dashboards/         # Dashboard 配置
│   └── dashboards/             # Dashboard JSON 文件
├── loki/
│   └── loki-config.yaml        # Loki 配置
├── tempo/
│   └── tempo-config.yaml       # Tempo 配置
├── postgres/
│   └── init.sql                # PostgreSQL 初始化腳本
└── alloy/
    └── config.alloy            # Alloy 配置
```

## 🔧 開發工作流程

### Go 平台開發
```bash
# 構建
make build-go

# 運行
make run-go

# 測試
make test-go
make test-go-metrics      # 測試 MetricsProvider
make test-go-health       # 測試 HealthAggregator
```

### Python ADK 開發
```bash
# 構建
make build-python

# 運行
make run-python

# 測試
make test-python
```

### 整合測試
```bash
# 端到端測試
make test-e2e

# 整合測試
make test-integration

# 效能測試
make benchmark
make load-test
```

## 💡 重要提醒

1. **API Keys**: 記得在 `.env` 文件中設置您的 API Keys（如 Gemini API Key）

2. **資源需求**:
   - RAM: 建議至少 8GB
   - Disk: 預留 20GB 空間
   - CPU: 4 核心以上

3. **網路配置**: 所有服務使用 `detectviz` Docker 網路（172.28.0.0/16）

4. **數據持久化**: 所有數據存儲在 Docker volumes 中，使用 `make clean` 會清除所有數據

5. **監控配置**: Prometheus 已配置好所有服務的 scrape targets

## 🎯 下一步

1. **配置 Grafana Dashboards**: 導入或創建監控儀表板
2. **設置告警規則**: 在 Grafana 中配置 Unified Alerting
3. **整合測試**: 運行 `make test-integration` 確保所有服務正常
4. **開始開發**: 使用 `make dev` 進入開發模式

這套環境配置提供了完整的可觀測性和數據存儲支援，足以支撐您的 MVP 開發和測試需求！
=======
# Detectviz Platform 部署指南

## 📦 系統概覽

Detectviz Platform 是一個完整的可觀測性解決方案，其核心設計是採用 Grafana Alloy 作為統一的數據收集器，實現 **本地 LGTM Stack** + **Grafana Cloud** 的雙輸出監控架構。

### 🏗️ 核心組件

| 組件 | 端口 (對外) | 用途 | 狀態 |
|:--- |:--- |:--- |:--- |
| **Grafana** | `3001` | 統一視覺化介面 | ✅ 就緒 |
| **Prometheus** | `9090` | 指標存儲和查詢 | ✅ 就緒 |
| **Grafana Alloy** | `12345` (UI) | 遙測數據收集器 | ✅ 就緒 |
| **Tempo** | `3200` (UI) | 分散式追蹤存儲 | ✅ 就緒 |
| **Loki** | `3100` | 日誌聚合和查詢 | ✅ 就緒 |
| **Pyroscope** | `4040` (UI) | 持續性能分析 | ✅ 就緒 |
| **PostgreSQL** | `5432` | 業務數據存儲 | ✅ 就緒 |
| **Redis** | `6379` | 快取和會話管理 | ✅ 就緒 |

### 📊 數據流向 (雙輸出架構)

Alloy 作為數據收集的閘道，將從應用程式收到的數據，同時轉發到兩處：

```
                      ┌──────────────────┐
DetectViz App  ───>   │  Grafana Alloy   │  ───>   本地 LGTM Stack (Loki, Tempo, etc.)
(OTLP & Logs)         │ (統一收集與轉發)    │  ───>   Grafana Cloud
                      └──────────────────┘
                            │
                            │ (需在 .env 中配置憑證)
                            ↓
                      Grafana Cloud
```

## 🚀 快速開始 (5 分鐘)

### 步驟 1: 配置環境變數 (非常重要！)

專案的啟動依賴 `.env` 檔案中的環境變數。

```bash
# 1. 進入部署目錄
cd /path/to/detectviz-platform/deploy

# 2. 從範本複製一份 .env 檔案
cp ../.env.example ./.env
```

**3. 編輯 `.env` 檔案**：
*   **本地開發**：您暫時不需要修改任何內容。
*   **雲端數據流**：如果您希望將數據發送到 Grafana Cloud，請務必填寫檔案下方的 `GF_CLOUD_*` 和 `GCLOUD_*` 相關的 ID 和 API Key。

### 步驟 2: 啟動所有服務

我們強烈建議使用 `Makefile` 來管理服務。

```bash
# 進入部署目錄
cd /path/to/detectviz-platform/deploy

# 啟動所有服務 (推薦)
make start

# 驗證服務狀態 (所有服務應顯示 "Up" 或 "healthy")
make status
```

### 步驟 3: 啟動 DetectViz 應用程式

在**另一個終端機視窗**中，從**專案根目錄**啟動您的應用程式。

```bash
# 進入專案根目錄
cd /path/to/detectviz-platform

# 讀取 .env 檔案並啟動應用 (包含 HTTP Demo)
source .env && \
DETECTVIZ_HTTP_DEMO=1 \
./detectviz plugin serve --config config.yaml
```

### 步驟 4: 驗證本地數據流

```bash
# 1. 發送測試請求到應用程式
curl http://localhost:7777/business

# 2. 直接查詢 Tempo API，確認是否收到 trace
curl "http://localhost:3200/api/search"
# 預期輸出: {"traces":[{"traceID":"...", ...}]}
```

### 步驟 5: 訪問本地 Grafana

1.  打開瀏覽器: [http://localhost:3001](http://localhost:3001)
2.  登入: `admin` / `admin123`
3.  導航到 **Explore**，選擇 **Loki**, **Tempo**, **Pyroscope** 等數據源，您應該能看到對應的數據。

## 🔧 Makefile 常用指令

在 `deploy` 目錄下，您可以使用以下指令：

*   `make start`: 啟動所有服務。
*   `make stop`: 停止所有服務。
*   `make restart`: 重啟所有服務。
*   `make status`: 檢查所有服務的狀態。
*   `make logs`: 查看所有服務的日誌。
*   `make logs-alloy`: 只查看 Alloy 的日誌。

## 🔍 故障排除

### 問題 1: Alloy 日誌顯示 `Unauthenticated` 或 `401` 錯誤
*   **原因**: Grafana Cloud 憑證錯誤。
*   **解決**: 仔細檢查 `.env` 檔案中的 `GF_CLOUD_*` 和 `GCLOUD_*` 變數是否從您的 Grafana Cloud 帳戶中正確複製。

### 問題 2: Loki 日誌顯示 `entry too far behind`
*   **原因**: Loki 預設拒絕時間戳過舊的日誌。
*   **解決**: 我們已將 `loki-config.yaml` 中的 `reject_old_samples` 設為 `false`，允許接收舊日誌，此問題通常不會再出現。

### 問題 3: 容器無法啟動，提示 `port is already allocated`
*   **原因**: 您主機上的某個連接埠已被其他應用程式佔用。
*   **解決**: 使用 `netstat -tlpn | grep <端口號>` 或其他工具找出佔用連接埠的程式並停止它，或修改 `docker-compose.yml` 將衝突的對外連接埠更換為另一個。

### 問題 4: Alloy 崩潰，日誌出現 `panic`
*   **原因**: `config.alloy` 設定檔語法錯誤。
*   **解決**: 我們已經修復了此問題。如果再次出現，請仔細檢查 `config.alloy` 的語法，特別是陣列和物件中的逗號。

--- 

> **版本**: v2.2.0 | **更新**: 2025-08-18 | **狀態**: 本地與雲端雙輸出
>>>>>>> 08fa581 (update)
