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
