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