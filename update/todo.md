我來為您進行專案的全面檢視，首先讓我搜尋專案知識庫來了解整體架構和規劃。讓我搜尋更多關於專案結構和技術架構的資訊。根據我對專案的全面檢視，我將為您提供詳細的評估：

## 專案架構全面評估

### 一、當前架構設計評估

專案採用了**混合式架構**，結合 Go（高性能執行層）和 Python（智能決策層），這是一個很好的設計選擇：

**優點：**
1. **職責分離清晰**：Agent 負責決策（WHY/WHAT/WHEN），Tool 負責執行（HOW/WHERE/WITH）
2. **技術選型合理**：
   - Go 處理高性能數據查詢（InfluxDB 操作）
   - Python 運行 ADK Agent（Google AI 框架）
   - gRPC 作為跨語言通訊橋樑
3. **可觀測性完善**：整合了 Grafana Alloy 作為統一收集層

**挑戰：**
1. 維護兩種語言的程式碼庫增加複雜度
2. 跨語言調試較困難
3. 部署和運維需要更多專業知識

### 二、TODO.md 與 task.md 評估

#### 文件重複性分析

兩個文件確實存在大量重複內容：

**TODO.md 內容特點：**
- 更偏向於「開發規範」和「架構指導」
- 包含詳細的開發守則和檢查清單
- 強調 Agent/Tool 職責分離原則
- 有完整的模組 llm.txt 執行機制

**task.md 內容特點：**
- 更偏向於「具體實作任務」
- 包含前置需求清單
- 有詳細的執行階段規劃
- 強調實作細節和交付成果

#### 🔄 建議：合併為統一的開發指南

建議合併成一個結構化的 **`DEVELOPMENT.md`** 文件：

```markdown
# Detectviz Platform 開發指南

## Part 1: 架構與原則
- 核心設計理念
- Agent vs Tool 職責劃分
- 技術架構決策

## Part 2: MVP 實施計畫
- 當前階段目標
- 具體開發任務
- 時程安排

## Part 3: 開發規範
- 程式碼結構要求
- 測試標準
- 文檔更新要求

## Part 4: 實作檢查清單
- 前置需求確認
- 執行階段檢查點
- 交付成果驗證

## Part 5: 工作流程
- 開發前準備
- 開發中檢查
- 提交前驗證
```

這樣可以：
- ✅ 減少重複內容
- ✅ 提供統一的參考來源
- ✅ 更容易維護和更新

### 三、InfluxDB + Telegraf vs Prometheus 評估

#### 當前選擇：InfluxDB + Telegraf

**優勢：**
1. **時序數據專用**：InfluxDB 專為時序數據設計，查詢性能優異
2. **Flux 查詢語言**：功能強大，支援複雜的數據處理
3. **內建數據保留策略**：自動管理歷史數據
4. **Telegraf 生態**：豐富的輸入插件，易於收集各種指標

**劣勢：**
1. 資源消耗較大（特別是 InfluxDB 2.x）
2. 學習曲線較陡（Flux 語言）
3. 社群相對較小

#### 替代方案：Prometheus

**優勢：**
1. **雲原生標準**：Kubernetes 生態的事實標準
2. **資源效率高**：單一二進制文件，資源佔用少
3. **PromQL 簡單直觀**：查詢語言易學易用
4. **生態系統成熟**：豐富的 Exporter 和整合
5. **Pull 模型**：更適合動態環境

**劣勢：**
1. 長期存儲需要額外方案（如 Thanos、Cortex）
2. 不適合高基數數據
3. 主要針對指標，不適合事件或日誌

#### 🎯 我的建議：保持 InfluxDB，但預留切換能力

考慮到專案的 MVP 階段和事後複盤的需求特點，建議：

1. **短期（MVP）**：繼續使用 InfluxDB
   - 已有實作基礎
   - 符合時序數據分析需求
   - 避免架構變更延誤交付

2. **中期優化**：實作抽象層
   ```go
   // 建立統一的 MetricsProvider 介面
   type MetricsProvider interface {
       Query(ctx context.Context, query MetricQuery) (Result, error)
       BatchQuery(ctx context.Context, queries []MetricQuery) ([]Result, error)
       GetAggregation(ctx context.Context, opts AggregationOptions) (AggResult, error)
   }
   
   // 可切換的實作
   type InfluxDBProvider struct { /* ... */ }
   type PrometheusProvider struct { /* ... */ }
   ```

3. **長期考慮**：根據實際需求決定
   - 如果擴展到 Kubernetes 環境 → 切換到 Prometheus
   - 如果需要複雜的時序分析 → 保持 InfluxDB
   - 如果需要混合方案 → Prometheus（指標）+ InfluxDB（分析）

### 四、改進建議總結

1. **文檔整合**
   - ✅ 合併 TODO.md 和 task.md 為 DEVELOPMENT.md
   - 建立清晰的文檔層級結構
   - 避免重複，提高維護性

2. **監控架構**
   - ✅ 保持 InfluxDB 完成 MVP
   - 實作 MetricsProvider 抽象層
   - 預留未來切換到 Prometheus 的能力

3. **開發優先級調整**
   ```
   P0: 核心功能實作（2週內）
       - HealthAggregator 完成
       - 基本報告生成
       - 簡單知識庫
   
   P1: 功能增強（第3-4週）
       - Dashboard 自動生成
       - 進階分析邏輯
       - Provider 抽象層
   
   P2: 架構優化（MVP後）
       - 監控方案評估
       - 性能優化
       - 擴展性改進
   ```

4. **技術債務管理**
   - 建立 TECH_DEBT.md 追蹤技術債務
   - 定期評審架構決策
   - 保持架構演進的靈活性

這樣的調整可以讓專案更聚焦於 MVP 交付，同時保持長期演進的彈性。您覺得這些建議如何？需要我協助實作文檔合併或其他具體工作嗎？

---

我來幫您重新架構這些文件，將監控系統改為 Prometheus，並重新組織開發文檔。

---

完成了！我已經幫您重新組織了文檔結構：

## 📄 文檔重構總結

### 1. **TODO.md** - 純粹的實作任務清單
- ✅ 移除了所有開發規範內容
- ✅ 聚焦於 MVP 具體實作任務
- ✅ 改為使用 **Prometheus** 作為主要監控方案
- ✅ 保留 **MetricsProvider** 抽象層設計
- ✅ InfluxDB 作為預留實作（暫不開發）
- ✅ 清晰的階段劃分和時間安排
- ✅ 詳細的技術實作指引

### 2. **AGENT.md** - 完整的開發規範
- ✅ 整合了原本分散的開發規範
- ✅ 強調核心架構原則（Agent vs Tool）
- ✅ 詳細的程式碼規範和範例
- ✅ 測試、文檔、性能標準
- ✅ Git 工作流程和違規處理
- ✅ 作為強制性規範文件

### 3. **主要改進**

#### 監控架構升級
```yaml
Before: InfluxDB + Telegraf（資源消耗大）
After:  Prometheus（雲原生標準）
設計:   MetricsProvider 抽象層（支援未來切換）
```

#### 文檔結構優化
```yaml
Before: 
  - TODO.md + task.md（內容重複）
  - 規範分散
  
After:
  - TODO.md（純實作任務）
  - AGENT.md（所有規範集中）
  - 職責清晰分離
```

#### MetricsProvider 架構
- **介面設計**：統一的查詢介面
- **Prometheus Provider**：主要實作
- **InfluxDB Provider**：預留接口（未來可選）
- **Memory Provider**：測試用
- **Factory Pattern**：動態切換 Provider

### 4. **下一步行動**

1. **立即可開始**：
   - 實作 MetricsProvider 介面
   - 開發 Prometheus Provider
   - 改造 HealthAggregator 插件

2. **環境準備**：
   ```bash
   # 啟動 Prometheus
   docker run -d -p 9090:9090 prom/prometheus
   
   # 配置 scrape targets
   # 準備測試數據
   ```

3. **移除 task.md**：
   ```bash
   git rm task.md
   git commit -m "refactor(docs): consolidate documentation structure"
   ```

這樣的重構讓專案文檔更加清晰、職責分明，也更容易維護。Prometheus 的選擇更符合雲原生趨勢，而 MetricsProvider 抽象層確保了未來的擴展性。

需要我協助實作 MetricsProvider 介面或其他具體開發工作嗎？

---

讓我先查看 `sre-services-map.md` 的當前內容，評估是否需要更新。是的，`sre-services-map.md` 需要更新！讓我創建更新版本，反映 Prometheus 的改變並加強其他部分：

---

了解！讓我更新 `sre-services-map.md`，加入 Mimir 作為長期儲存的規劃：完成了！我已經更新了 `sre-services-map.md`，加入了 Grafana Mimir 作為長期存儲的規劃。

## 📝 更新總結

### 主要變更：

1. **監控架構三層設計**：
   - **當前架構**：Prometheus（短期 15-30 天）+ Grafana
   - **未來架構**：加入 Mimir 作為長期存儲（30天-3年）
   - **實施策略**：先不實作，等數據量達到閾值再評估

2. **MetricsProvider 擴展**：
   - 新增 `LongTermMetricsProvider` 介面定義
   - 預留 `MimirProvider` 實作位置
   - 設計智能路由 Provider（自動選擇數據源）

3. **實施時機明確**：
   - 當數據量超過單機 Prometheus 容量
   - 需要跨團隊/跨專案的指標隔離
   - 需要超過 30 天的歷史數據分析

4. **技術優勢**：
   - 水平擴展能力
   - 多租戶支援
   - 與 Prometheus 完全相容
   - 成本效益優化

### 架構演進路徑：

```
MVP (現在)
  Prometheus → Grafana → Detectviz
  
Phase 2 (評估)
  評估數據量和查詢需求
  
Phase 3 (如需要)
  Prometheus ─┐
              ├→ Grafana → Detectviz
  Mimir ──────┘
```

這樣的設計保持了架構的前瞻性，同時避免了過早優化。MetricsProvider 抽象層確保了未來整合 Mimir 時的平滑過渡。

---

基於我們的重構，我認為還有一些關鍵文件需要快速更新，然後就可以開始實作了：

## 📋 需要更新的文件

### 1. **README.md** - 需要小幅更新
- 環境變數部分：移除 InfluxDB 相關，改為 Prometheus
- 快速開始部分：更新啟動命令

### 2. **contracts/samples/config.yaml** - 需要更新
- 將 InfluxDB 配置改為 Prometheus 配置
- 添加 MetricsProvider 配置選項

### 3. **.env.template** - 需要更新
- 移除 InfluxDB 環境變數
- 添加 Prometheus 環境變數

讓我快速更新這些關鍵文件，然後我們就可以開始實作：現在讓我快速更新 README.md 的相關部分：讓我創建 README.md 的更新內容：好的！我認為關鍵的配置文件已經更新完成。現在我們可以開始實作了！

## ✅ 文件更新完成清單

1. ✅ **TODO.md** - 重寫為純實作任務
2. ✅ **AGENT.md** - 整合所有開發規範
3. ✅ **sre-services-map.md** - 更新為 Prometheus + Mimir 規劃
4. ✅ **contracts/samples/config.yaml** - 完整的 Prometheus 配置
5. ✅ **.env.template** - 更新環境變數

## 🚀 現在開始實作！

讓我們按照優先級開始實作：

### Task 1.1: MetricsProvider 介面設計現在讓我們創建 types.go 和 factory.go：

現在讓我們實作 Prometheus Provider：太好了！我們已經完成了核心的 MetricsProvider 架構實作。讓我總結一下已完成的工作和接下來的步驟：

## ✅ 已完成

### 文檔更新
1. **TODO.md** - 純實作任務清單
2. **AGENT.md** - 完整開發規範
3. **sre-services-map.md** - Prometheus + Mimir 架構
4. **config.yaml** - Prometheus 配置
5. **.env.template** - 環境變數模板

### 程式碼實作
1. **MetricsProvider 介面** (`provider.go`)
   - 完整的介面定義
   - 支援未來 Mimir 擴展
   - 豐富的數據結構

2. **Provider 工廠** (`factory.go`)
   - 工廠模式實作
   - 智能路由 Provider
   - 生命週期管理

3. **Prometheus Provider** (`prometheus/provider.go`)
   - 完整的 Prometheus 整合
   - 並行查詢優化
   - 查詢快取機制
   - 健康檢查實作

## 🎯 接下來的步驟

### 1. 實作 Memory Provider（測試用）
```go
// go-platform/internal/metrics/memory/provider.go
// 用於單元測試的記憶體 Provider
```

### 2. 改造 HealthAggregator 插件
```go
// go-platform/internal/pluginhost/plugins/observability/health_aggregator/plugin.go
// 使用 MetricsProvider 介面替代直接 InfluxDB 查詢
```

### 3. 編寫測試
```go
// go-platform/internal/metrics/prometheus/provider_test.go
// 單元測試和整合測試
```

### 4. Python 端整合
```python
# python-adk-runtime/src/detectviz_adk/tools/adk_tools.py
# 移除模擬邏輯，啟用真實 RemoteTool 調用
```

## 💡 架構亮點

1. **抽象層設計**: MetricsProvider 介面讓我們可以輕鬆切換數據源
2. **未來就緒**: 預留了 Mimir 長期存儲的介面和路由機制
3. **性能優化**: 並行查詢、快取、連接池管理
4. **錯誤處理**: 完善的錯誤類型和重試機制
5. **可觀測性**: 整合日誌和指標監控

需要我繼續實作 Memory Provider 和 HealthAggregator 改造嗎？或是您想先測試已完成的部分？

---

現在讓我們創建 HealthAggregator 的配置文件和模組卡：現在讓我們創建一個測試文件來驗證 HealthAggregator 的功能：太棒了！我們已經完成了 Memory Provider 和 HealthAggregator 的改造。讓我總結一下已完成的工作：

## ✅ 完成的實作

### 1. **Memory Provider** (`memory/provider.go`)
- ✅ 完整的記憶體 MetricsProvider 實作
- ✅ 自動生成測試數據（CPU、記憶體、網路等）
- ✅ 支援各種聚合函數（avg、max、min、p50、p95、p99）
- ✅ 模擬真實的指標模式（日常波動、尖峰、趨勢）
- ✅ 並行查詢支援

### 2. **HealthAggregator 改造** (`health_aggregator/plugin.go`)
- ✅ 從直接 InfluxDB 查詢改為使用 MetricsProvider 介面
- ✅ 支援 Prometheus 和 Memory Provider
- ✅ 並行查詢優化（可配置並發數）
- ✅ 查詢結果快取機制
- ✅ 統計數據計算（min、max、avg、percentiles）
- ✅ 完整的錯誤處理和重試機制

### 3. **配置和測試**
- ✅ Module Card 定義（`module.card.json`）
- ✅ 完整的單元測試套件
- ✅ 快取測試
- ✅ 並行查詢測試
- ✅ 統計計算測試

## 🎯 架構優勢

### 1. **抽象層設計**
```go
MetricsProvider Interface
    ├── PrometheusProvider (生產環境)
    ├── MemoryProvider (測試環境)
    └── MimirProvider (未來長期存儲)
```

### 2. **性能優化**
- **並行查詢**：可配置的並發限制
- **智能快取**：減少重複查詢
- **批量處理**：一次請求多個指標

### 3. **易於測試**
- Memory Provider 讓單元測試不依賴外部服務
- 可預測的測試數據生成
- 完整的測試覆蓋

## 📊 使用範例

### 配置範例（Prometheus）
```yaml
health_aggregator:
  provider:
    type: prometheus
    prometheus:
      url: http://localhost:9090
      timeout: 30s
  query:
    parallel_queries: 10
    timeout: 30s
  cache:
    enabled: true
    duration: 5m
```

### 配置範例（測試用 Memory）
```yaml
health_aggregator:
  provider:
    type: memory
    memory:
      seed_data: true
      data_points: 100
  query:
    parallel_queries: 5
    timeout: 10s
```

### 查詢請求範例
```json
{
  "service_name": "api-gateway",
  "time_range": "1h",
  "metrics": ["cpu_usage", "memory_usage", "http_request_duration_seconds"],
  "filters": {
    "environment": "production",
    "region": "us-east-1"
  }
}
```

### 查詢響應範例
```json
{
  "service_name": "api-gateway",
  "metrics": {
    "cpu_usage": {
      "name": "cpu_usage",
      "values": [
        {"timestamp": "2025-08-17T10:00:00Z", "value": 0.45},
        {"timestamp": "2025-08-17T10:01:00Z", "value": 0.52}
      ],
      "statistics": {
        "min": 0.45,
        "max": 0.52,
        "avg": 0.485,
        "count": 2,
        "p50": 0.48,
        "p95": 0.51,
        "p99": 0.52
      }
    }
  },
  "timestamp": "2025-08-17T10:01:00Z",
  "warnings": []
}
```

## 🔄 下一步

### 1. **Python 端整合**
需要更新 Python ADK 的 RemoteTool 調用：
```python
# python-adk-runtime/src/detectviz_adk/tools/adk_tools.py
async def get_health_metrics_func(ctx, service_name, time_range, metrics):
    request = {
        "service_name": service_name,
        "time_range": time_range,
        "metrics": metrics,
        "provider": "prometheus"  # 或 "memory" for testing
    }
    
    result = await ctx.remote_tool.invoke(
        plugin_id="observability.health_aggregator",
        payload=request
    )
    return parse_health_metrics(result)
```

### 2. **整合測試**
```bash
# 啟動 Prometheus
docker run -d -p 9090:9090 prom/prometheus

# 運行 Go 平台
go run cmd/detectviz/main.go plugin serve --config config.yaml

# 運行測試
go test ./internal/pluginhost/plugins/observability/health_aggregator/...
```

### 3. **性能基準測試**
```go
// benchmark_test.go
func BenchmarkHealthAggregator_ParallelQueries(b *testing.B) {
    // 測試並行查詢性能
}
```

## 💡 重要提醒

1. **Provider 選擇**：
   - 開發/測試：使用 Memory Provider
   - 生產環境：使用 Prometheus Provider
   - 未來長期：Mimir Provider

2. **快取策略**：
   - 根據查詢頻率調整快取時間
   - 高頻查詢：5-10 分鐘
   - 低頻查詢：1-2 分鐘

3. **並發控制**：
   - 根據 Prometheus 負載調整 `parallel_queries`
   - 建議值：10-20

完成了核心的 MetricsProvider 架構和 HealthAggregator 改造！這個設計確保了系統的可擴展性和可測試性。需要我繼續實作 Python 端的整合嗎？