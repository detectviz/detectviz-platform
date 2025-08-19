# Knowledge Provider Plugin

知識庫提供者插件提供了強大的知識管理功能，用於儲存、檢索和搜索事後複盤知識。

## 功能特性

- **多種後端支持**：Memory Provider（測試用）和 PostgreSQL Provider（生產用）
- **智能搜索**：支持全文搜索和相似性搜索
- **類型安全**：使用 Protocol Buffers 確保跨語言通訊安全
- **高性能**：支持連接池、索引優化和快取
- **企業級功能**：完整的錯誤處理、日誌記錄和健康檢查

## 核心概念

### Provider 介面

```go
type Provider interface {
    Store(ctx context.Context, item *KnowledgeItem) error
    Retrieve(ctx context.Context, id string) (*KnowledgeItem, error)
    Search(ctx context.Context, query *SearchQuery) (*SearchResult, error)
    SimilaritySearch(ctx context.Context, content string, limit int) (*SearchResult, error)
    Delete(ctx context.Context, id string) error
    Close() error
}
```

### 知識項目結構

```go
type KnowledgeItem struct {
    ID          string            `json:"id"`
    Title       string            `json:"title"`
    Content     string            `json:"content"`
    Category    string            `json:"category"`
    Tags        []string          `json:"tags"`
    Metadata    map[string]string `json:"metadata"`
    CreatedAt   time.Time         `json:"created_at"`
    UpdatedAt   time.Time         `json:"updated_at"`
    CreatedBy   string            `json:"created_by"`
    Severity    string            `json:"severity,omitempty"`
    Status      string            `json:"status,omitempty"`
    
    // 事後複盤特有欄位
    IncidentID    string       `json:"incident_id,omitempty"`
    RootCause     string       `json:"root_cause,omitempty"`
    Resolution    string       `json:"resolution,omitempty"`
    LessonsLearned []string    `json:"lessons_learned,omitempty"`
    ActionItems   []ActionItem `json:"action_items,omitempty"`
}
```

## 支持的操作

### 1. 儲存知識項目

```json
{
  "method": "store",
  "data": {
    "item": {
      "title": "Database Connection Timeout",
      "content": "Investigation of database connection timeout issues",
      "category": "postmortem",
      "tags": ["database", "timeout", "performance"],
      "created_by": "sre-team",
      "severity": "high",
      "status": "published",
      "incident_id": "INC-DB-001",
      "root_cause": "Connection pool exhaustion",
      "resolution": "Optimized queries and increased pool size",
      "lessons_learned": ["Monitor connection pool metrics", "Set query timeouts"],
      "action_items": [
        {
          "id": "action-001",
          "description": "Implement connection pool monitoring",
          "assignee": "db-team",
          "status": "in-progress",
          "priority": "high"
        }
      ]
    }
  }
}
```

### 2. 檢索知識項目

```json
{
  "method": "retrieve",
  "data": {
    "item_id": "test-item-001"
  }
}
```

### 3. 搜索知識項目

```json
{
  "method": "search",
  "data": {
    "query": {
      "query": "database timeout",
      "category": "postmortem",
      "tags": ["database", "performance"],
      "filters": {
        "severity": "high",
        "status": "published"
      },
      "sort_by": "updated_at",
      "sort_order": "desc",
      "limit": 10,
      "offset": 0
    }
  }
}
```

### 4. 相似性搜索

```json
{
  "method": "similarity_search",
  "data": {
    "content": "connection pool exhaustion timeout",
    "limit": 5
  }
}
```

### 5. 刪除知識項目

```json
{
  "method": "delete",
  "data": {
    "item_id": "test-item-001"
  }
}
```

## Provider 配置

### Memory Provider（測試用）

```yaml
provider: "memory"
similarity:
  algorithm: "cosine"
  threshold: 0.3
  max_results: 10
cache:
  enabled: true
  ttl: 300000000000  # 5 minutes (nanoseconds)
  max_entries: 1000
```

### PostgreSQL Provider（生產用）

```yaml
provider: "postgresql"
database:
  host: "localhost"
  port: 5432
  database: "detectviz_knowledge"
  username: "detectviz"
  password: "secure_password"
  ssl_mode: "prefer"
  max_open_conns: 25
  max_idle_conns: 5
  conn_max_lifetime: 300000000000  # 5 minutes (nanoseconds)
similarity:
  algorithm: "postgresql_ts_rank"
  threshold: 0.1
  max_results: 20
cache:
  enabled: true
  ttl: 600000000000  # 10 minutes (nanoseconds)
  max_entries: 5000
```

## PostgreSQL Schema

插件會自動創建以下資料庫結構：

```sql
CREATE TABLE knowledge_items (
    id VARCHAR(255) PRIMARY KEY,
    title TEXT NOT NULL,
    content TEXT NOT NULL,
    category VARCHAR(100) NOT NULL,
    tags TEXT[],
    metadata JSONB,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    created_by VARCHAR(255) NOT NULL,
    severity VARCHAR(50),
    status VARCHAR(50),
    
    -- 事後複盤特有欄位
    incident_id VARCHAR(255),
    root_cause TEXT,
    resolution TEXT,
    lessons_learned TEXT[],
    action_items JSONB,
    
    -- 全文檢索索引
    content_vector tsvector GENERATED ALWAYS AS (
        setweight(to_tsvector('english', title), 'A') ||
        setweight(to_tsvector('english', content), 'B') ||
        setweight(to_tsvector('english', COALESCE(root_cause, '')), 'C')
    ) STORED
);
```

### 索引

- `idx_knowledge_category`: 類別索引
- `idx_knowledge_created_at`: 創建時間索引
- `idx_knowledge_incident_id`: 事故 ID 索引
- `idx_knowledge_severity`: 嚴重程度索引
- `idx_knowledge_status`: 狀態索引
- `idx_knowledge_tags`: 標籤 GIN 索引
- `idx_knowledge_metadata`: 元數據 GIN 索引
- `idx_knowledge_content_vector`: 全文檢索 GIN 索引

## 測試

運行單元測試：

```bash
go test ./internal/pluginhost/plugins/knowledge/ -v
```

運行特定測試：

```bash
go test ./internal/pluginhost/plugins/knowledge/ -v -run TestKnowledgePlugin_Store
```

## 性能考量

### Memory Provider
- 適用於測試和開發環境
- 數據不持久化
- 搜索性能隨數據量線性增長

### PostgreSQL Provider
- 適用於生產環境
- 支持完整的 ACID 特性
- 使用 PostgreSQL 全文檢索功能
- 支持連接池和索引優化
- 推薦配置：
  - 至少 25 個連接池
  - 啟用 shared_preload_libraries
  - 調整 work_mem 以支持複雜查詢

## 監控和觀察性

插件提供以下指標：

- `plugin_runs_total`: 插件執行總數
- `plugin_duration_seconds_bucket`: 插件執行時間分布
- `knowledge_provider_queries_total`: 查詢總數
- `knowledge_provider_query_duration_seconds`: 查詢執行時間
- `knowledge_provider_storage_size_bytes`: 存儲大小
- `knowledge_provider_similarity_search_duration_seconds`: 相似性搜索時間

## 故障排除

### 常見問題

1. **PostgreSQL 連接失敗**
   - 檢查資料庫連接參數
   - 確認 PostgreSQL 服務狀態
   - 檢查防火牆設置

2. **搜索結果為空**
   - 確認搜索詞拼寫
   - 檢查類別和標籤篩選條件
   - 驗證數據是否正確儲存

3. **性能問題**
   - 檢查索引是否正確創建
   - 調整連接池大小
   - 監控查詢執行計劃

### 日誌級別

- `DEBUG`: 詳細的操作日誌
- `INFO`: 重要操作成功記錄
- `WARN`: 非致命錯誤警告
- `ERROR`: 操作失敗錯誤

## 擴展性

### 添加新的 Provider

1. 實作 `Provider` 介面
2. 在 `ProviderFactory` 中註冊新的 provider 類型
3. 添加相應的配置驗證邏輯
4. 創建單元測試

### 自定義搜索算法

1. 在 `SimilarityConfig` 中添加新算法支持
2. 在相應的 Provider 中實作算法邏輯
3. 更新配置驗證器
4. 添加性能基準測試

## 安全考量

- 所有數據庫憑證必須通過環境變數管理
- 使用 TLS 加密數據庫連接
- 實施適當的輸入驗證和清理
- 定期更新依賴套件以修復安全漏洞