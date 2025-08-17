# TODO.md - MVP 實作任務

> 文檔職責：記錄 MVP 階段具體實作任務和進度追蹤

## 專案狀態

> 當前專案狀態：參見 [專案狀態文檔](docs/status/PROJECT_STATUS.md)  
> 開發規範：參見 [AI協作指南](AGENT.md)  
> 技術規格：參見 [技術規格文檔](spec.md)

- **當前階段**: MVP Phase 3 - 事後複盤系統
- **時程**: 2 週交付（當前 Week 1）
- **技術棧**: Go + Python ADK + Prometheus + Grafana
- **架構模式**: Agent（決策）+ Tool（執行）分離

## MVP 核心交付目標

### 主要功能
1. 自動事故複盤分析
2. 結構化報告生成（Markdown）
3. 知識庫累積與檢索
4. Grafana Dashboard 自動生成（增強功能）

### 技術目標
- 端到端工作流程驗證
- MetricsProvider 抽象層實作
- Prometheus 整合
- 90%+ 測試覆蓋率

## 前置準備檢查清單

### 環境配置（用戶負責）
- [ ] **Prometheus 設置**
  ```yaml
  # prometheus.yml 基本配置
  global:
    scrape_interval: 15s
    evaluation_interval: 15s
  
  scrape_configs:
    - job_name: 'detectviz-services'
      static_configs:
        - targets: ['localhost:8080', 'localhost:8081']
    
    - job_name: 'node-exporter'
      static_configs:
        - targets: ['localhost:9100']
  ```

- [ ] **PostgreSQL 設置**（知識庫用）
  ```bash
  # 創建資料庫
  CREATE DATABASE detectviz;
  CREATE USER detectviz_user WITH PASSWORD 'secure_password';
  GRANT ALL PRIVILEGES ON DATABASE detectviz TO detectviz_user;
  ```

- [ ] **Grafana 配置**
  - API Key 生成
  - Prometheus 數據源配置
  - 資料夾權限設置

### 測試數據準備
- [ ] Prometheus 測試指標生成腳本
- [ ] 模擬故障場景數據
- [ ] Alert webhook payload 範例

## 實作任務清單

### Phase 1: 基礎架構（Week 1 前半）

#### Task 1.1: MetricsProvider 抽象層設計
**位置**: `go-platform/internal/metrics/`
**優先級**: P0
**預估時間**: 4 小時

```go
// 需要實作的介面
type MetricsProvider interface {
    // 基本查詢
    Query(ctx context.Context, query MetricQuery) (*QueryResult, error)
    
    // 批量查詢（並行優化）
    BatchQuery(ctx context.Context, queries []MetricQuery) ([]*QueryResult, error)
    
    // 聚合查詢
    GetAggregation(ctx context.Context, opts AggregationOptions) (*AggregationResult, error)
    
    // 健康檢查
    HealthCheck(ctx context.Context) error
}

type MetricQuery struct {
    Metric      string
    Labels      map[string]string
    TimeRange   TimeRange
    Step        time.Duration
    Aggregation string // avg, max, min, sum, count
}
```

**具體步驟**:
1. 創建 `provider.go` 定義介面
2. 創建 `types.go` 定義數據結構
3. 創建 `factory.go` 實作 Provider 工廠模式
4. 編寫介面測試用例

---

#### Task 1.2: Prometheus Provider 實作
**位置**: `go-platform/internal/metrics/prometheus/`
**優先級**: P0
**預估時間**: 8 小時

**實作內容**:
```go
// prometheus_provider.go
type PrometheusProvider struct {
    client     v1.API
    httpClient *http.Client
    cache      *cache.Cache
    config     PrometheusConfig
}

func (p *PrometheusProvider) Query(ctx context.Context, query MetricQuery) (*QueryResult, error) {
    // 1. 構建 PromQL 查詢
    promQL := p.buildPromQL(query)
    
    // 2. 檢查快取
    if cached := p.cache.Get(promQL); cached != nil {
        return cached.(*QueryResult), nil
    }
    
    // 3. 執行查詢
    result, warnings, err := p.client.Query(ctx, promQL, query.TimeRange.End)
    
    // 4. 轉換結果格式
    queryResult := p.convertResult(result)
    
    // 5. 寫入快取
    p.cache.Set(promQL, queryResult, 5*time.Minute)
    
    return queryResult, nil
}
```

**子任務**:
- [ ] Prometheus Go client 整合
- [ ] PromQL 查詢構建器
- [ ] 結果格式轉換
- [ ] 查詢快取實作
- [ ] 並行批量查詢優化
- [ ] 錯誤處理和重試機制

---

#### Task 1.3: HealthAggregator 插件改造
**位置**: `go-platform/internal/pluginhost/plugins/observability/health_aggregator/`
**優先級**: P0
**預估時間**: 6 小時

**改造內容**:
1. 移除 InfluxDB 直接依賴
2. 改用 MetricsProvider 介面
3. 支援 Provider 動態切換
4. 保持向後兼容的 gRPC 介面

```go
// plugin.go
type HealthAggregatorPlugin struct {
    provider metrics.MetricsProvider
    config   PluginConfig
}

func (p *HealthAggregatorPlugin) Invoke(ctx context.Context, req *pb.InvokeRequest) (*pb.InvokeResponse, error) {
    // 解析請求參數
    params := ParseHealthQueryParams(req.Payload)
    
    // 使用 MetricsProvider 查詢
    queries := p.buildMetricQueries(params)
    results, err := p.provider.BatchQuery(ctx, queries)
    
    // 聚合處理
    aggregated := p.aggregate(results, params.AggregationType)
    
    // 返回結果
    return &pb.InvokeResponse{
        Payload: p.marshalResults(aggregated),
    }, nil
}
```

---

### Phase 2: Python ADK 整合（Week 1 後半）

#### Task 2.1: RemoteTool 調用優化
**位置**: `python-adk-runtime/src/detectviz_adk/tools/`
**優先級**: P0
**預估時間**: 4 小時

**優化內容**:
- [ ] 移除所有模擬邏輯
- [ ] 實作健全的錯誤處理
- [ ] 添加重試機制
- [ ] 實作請求超時控制

```python
# adk_tools.py
async def get_health_metrics_func(
    ctx: ToolContext,
    service_name: str,
    time_range: str,
    metrics: List[str]
) -> Dict[str, Any]:
    """從 Prometheus 獲取健康指標（通過 Go HealthAggregator）"""
    
    # 構建請求
    request = {
        "service_name": service_name,
        "time_range": time_range,
        "metrics": metrics,
        "provider": "prometheus"  # 明確指定使用 Prometheus
    }
    
    # 調用 RemoteTool
    try:
        result = await ctx.remote_tool.invoke(
            plugin_id="observability.health_aggregator",
            payload=request,
            timeout=30.0
        )
        
        # 解析並驗證結果
        return parse_health_metrics(result)
        
    except TimeoutError:
        logger.error(f"Health metrics query timeout for {service_name}")
        raise
    except Exception as e:
        logger.error(f"Failed to get health metrics: {e}")
        raise
```

---

#### Task 2.2: Agent 狀態管理強化
**位置**: `python-adk-runtime/src/detectviz_adk/agents/postmortem/`
**優先級**: P0
**預估時間**: 6 小時

**實作內容**:
- [ ] Session State 管理優化
- [ ] Agent 間數據傳遞機制
- [ ] 狀態持久化（Redis）
- [ ] 狀態恢復機制

```python
# state_manager.py
class PostmortemStateManager:
    """管理複盤分析的狀態"""
    
    def __init__(self, redis_client: Redis):
        self.redis = redis_client
        self.session_id = None
    
    async def save_metrics(self, metrics: Dict[str, Any]):
        """保存收集到的指標數據"""
        key = f"postmortem:{self.session_id}:metrics"
        await self.redis.setex(
            key, 
            3600,  # 1 小時過期
            json.dumps(metrics)
        )
    
    async def get_metrics(self) -> Dict[str, Any]:
        """獲取指標數據"""
        key = f"postmortem:{self.session_id}:metrics"
        data = await self.redis.get(key)
        return json.loads(data) if data else {}
    
    async def save_analysis(self, analysis: Dict[str, Any]):
        """保存分析結果"""
        key = f"postmortem:{self.session_id}:analysis"
        await self.redis.setex(
            key,
            3600,
            json.dumps(analysis)
        )
```

---

### Phase 3: 報告生成與知識庫（Week 2 前半）

#### Task 3.1: Markdown 報告模板系統
**位置**: `python-adk-runtime/templates/`
**優先級**: P0
**預估時間**: 4 小時

**實作內容**:
- [ ] 創建 Jinja2 模板
- [ ] 實作模板渲染引擎
- [ ] 支援多語言（中文/英文）
- [ ] 圖表和表格生成

```python
# templates/postmortem_report.md.j2
# 事故複盤報告

## 執行摘要
- **事件 ID**: {{ incident.id }}
- **發生時間**: {{ incident.start_time }}
- **恢復時間**: {{ incident.end_time }}
- **持續時間**: {{ incident.duration }}
- **影響等級**: {{ incident.severity }}

## 事件時間線
{% for event in timeline %}
- **{{ event.time }}**: {{ event.description }}
{% endfor %}

## 根因分析
### 直接原因
{{ analysis.direct_cause }}

### 根本原因
{{ analysis.root_cause }}

### 貢獻因素
{% for factor in analysis.contributing_factors %}
- {{ factor }}
{% endfor %}

## 影響評估
- **受影響用戶**: {{ impact.affected_users }}
- **業務影響**: {{ impact.business_impact }}
- **SLA 違反**: {{ impact.sla_violation }}

## 改進建議
{% for recommendation in recommendations %}
### {{ recommendation.title }}
- **優先級**: {{ recommendation.priority }}
- **負責人**: {{ recommendation.owner }}
- **截止日期**: {{ recommendation.deadline }}
- **行動項目**: {{ recommendation.action }}
{% endfor %}

## 學習重點
{% for lesson in lessons_learned %}
- {{ lesson }}
{% endfor %}
```

---

#### Task 3.2: 知識庫 Provider 架構
**位置**: `go-platform/internal/pluginhost/plugins/knowledge/`
**優先級**: P0
**預估時間**: 8 小時

**架構設計**:
```go
// provider/interface.go
type KnowledgeProvider interface {
    // 儲存事件
    StoreIncident(ctx context.Context, incident *Incident) error
    
    // 儲存教訓
    StoreLessons(ctx context.Context, lessons []*Lesson) error
    
    // 檢索相似事件
    FindSimilarIncidents(ctx context.Context, query SimilarityQuery) ([]*Incident, error)
    
    // 獲取歷史教訓
    GetLessonsLearned(ctx context.Context, filter LessonFilter) ([]*Lesson, error)
    
    // 更新事件狀態
    UpdateIncidentStatus(ctx context.Context, id string, status IncidentStatus) error
}

// PostgreSQL Provider 實作
type PostgreSQLProvider struct {
    db     *sql.DB
    config PostgreSQLConfig
}

// Memory Provider (測試用)
type MemoryProvider struct {
    incidents map[string]*Incident
    lessons   map[string]*Lesson
    mu        sync.RWMutex
}
```

**子任務**:
- [ ] Provider 介面定義
- [ ] PostgreSQL Provider 實作
- [ ] Memory Provider 實作
- [ ] 資料庫 Schema 設計
- [ ] 相似性匹配算法
- [ ] 單元測試和整合測試

---

### Phase 4: Dashboard 自動生成（Week 2 後半）

#### Task 4.1: Grafana Dashboard Builder
**位置**: `go-platform/internal/pluginhost/plugins/reporting/dashboard_builder/`
**優先級**: P1
**預估時間**: 6 小時

**實作內容**:
```go
type DashboardBuilder struct {
    grafanaClient *grafana.Client
    templates     map[string]*DashboardTemplate
}

func (b *DashboardBuilder) CreatePostmortemDashboard(
    ctx context.Context,
    incident *Incident,
    metrics []MetricData,
) (*Dashboard, error) {
    // 1. 選擇模板
    template := b.selectTemplate(incident.Type)
    
    // 2. 生成 Panel 配置
    panels := b.generatePanels(metrics, incident.TimeRange)
    
    // 3. 創建 Dashboard
    dashboard := &Dashboard{
        Title:  fmt.Sprintf("Postmortem: %s", incident.ID),
        Panels: panels,
        Tags:   []string{"postmortem", "automated"},
    }
    
    // 4. 上傳到 Grafana
    return b.grafanaClient.CreateDashboard(ctx, dashboard)
}
```

---

### Phase 5: 測試與驗證（持續進行）

#### Task 5.1: 單元測試完善
**優先級**: P0
**預估時間**: 持續

**測試覆蓋**:
- [ ] MetricsProvider 介面測試
- [ ] Prometheus Provider 測試
- [ ] Agent 決策邏輯測試
- [ ] 報告生成測試
- [ ] 知識庫操作測試

#### Task 5.2: 端到端整合測試
**優先級**: P0
**預估時間**: 4 小時

**測試場景**:
```python
# tests/e2e/test_postmortem_flow.py
async def test_complete_postmortem_flow():
    """測試完整的事後複盤流程"""
    
    # 1. 準備測試數據
    incident = create_test_incident()
    
    # 2. 觸發複盤分析
    result = await postmortem_orchestrator.analyze(incident)
    
    # 3. 驗證各階段輸出
    assert result.metrics_collected
    assert result.root_cause_identified
    assert result.report_generated
    assert result.knowledge_stored
    assert result.dashboard_created
    
    # 4. 驗證報告質量
    report = result.report
    assert "根因分析" in report
    assert "改進建議" in report
    assert len(result.recommendations) >= 3
```

---

## 進度追蹤

### Week 1 (當前)
- Day 1-2: MetricsProvider 架構 + Prometheus Provider
- Day 3-4: HealthAggregator 改造 + Registry 優化 + 生產級功能
- Day 5: 完整測試驗證與文檔更新

### Week 2
- [ ] Day 6-7: 報告生成 + 知識庫
- [ ] Day 8-9: Dashboard 自動生成
- [ ] Day 10: 端到端測試 + 文檔

## 快速開始

### 1. 環境設置
```bash
# 啟動 Prometheus
docker run -d \
  -p 9090:9090 \
  -v $(pwd)/prometheus.yml:/etc/prometheus/prometheus.yml \
  prom/prometheus

# 啟動 PostgreSQL
docker run -d \
  -p 5432:5432 \
  -e POSTGRES_DB=detectviz \
  -e POSTGRES_USER=detectviz \
  -e POSTGRES_PASSWORD=secure_password \
  postgres:15

# 啟動 Redis（狀態管理）
docker run -d -p 6379:6379 redis:7
```

### 2. 依賴安裝
```bash
# Go 依賴
cd go-platform
go mod download

# Python 依賴
cd python-adk-runtime
pip install -r requirements.txt
```

### 3. 執行測試
```bash
# Go 測試
make test-go

# Python 測試
make test-python

# 端到端測試
make test-e2e
```

## 交付標準

### 必須完成
- Prometheus 數據查詢功能
- 完整的複盤報告生成
- 基本知識庫功能
- 90% 測試覆蓋率

### 加分項
- Dashboard 自動生成
- 多語言報告支援
- 性能優化（<3s 報告生成）

## 問題追蹤

遇到問題時的處理流程：
1. 查閱 [`AGENT.md`](./AGENT.md) 確認是否違反架構原則
2. 查看 [`docs/troubleshooting.md`](./docs/troubleshooting.md)
3. 在 GitHub Issues 中搜尋相關問題
4. 創建新 Issue 並標記優先級

---

*最後更新: 2025-08-17*
*版本: 1.0.0*