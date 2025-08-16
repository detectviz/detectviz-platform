
### 提示詞：Detectviz 平台 MVP 核心功能實作指令

**角色**：
你是 Detectviz 平台的首席 AI 架構師。

**專案背景與當前狀態**：
專案的「空實作 MVP」已成功通過手動與 `adk web` 測試。`postmortem_orchestrator` Agent 團隊的協調流程、子代理間的數據傳遞以及工具的模擬調用均已驗證。我們現在的任務是將所有**模擬的佔位邏輯替換為真實的後端功能實作**，以達成 MVP 的核心交付目標。

**核心任務**：
你的任務是領導並完成從**真實數據獲取**到**最終報告生成**的完整端到端工作流程。你需要嚴格遵守已定義的架構原則，特別是「Agent 決策」與「Tool 執行」的分離。

**關鍵參考文件**：
在開始之前，你必須詳細閱讀並遵循以下文件的規範：
1.  **`AGENT.md`**: 統一的開發貢獻指南，包含所有開發流程與規範。
2.  **`spec.md`**: 平台技術規格，所有技術實作的單一事實來源 (SSOT)。
3.  **`docs/sre-services-map.md`**: 架構憲法，定義 Agent 與 Tool 的職責。

## 🔍 實作狀態評估 & 前置需求

### ✅ **已就緒的基礎設施**
- RemoteTool gRPC 橋接架構完整
- Agent 團隊協調機制已驗證
- Grafana Dashboard 和知識庫更新功能已規劃
- 空實作 MVP 測試通過

### ⚠️ **需要補強的關鍵組件**
1. **缺少 `health_aggregator` 插件目錄**
2. **Go 平台服務啟動機制未建立**
3. **InfluxDB 連接和查詢邏輯待實作**

### 📋 **前置需求清單（等待提供）**

#### 🔧 **技術配置信息**
- [ ] **InfluxDB 連接配置**：
  - 端點 URL（如：`http://localhost:8086`）
  - 認證 token 或 credentials
  - 資料庫/bucket 名稱
  - 預期的數據結構（measurement names, field keys）

- [ ] **PostgreSQL 連接配置**：
  - 端點 URL（如：`postgresql://localhost:5432/detectviz`）
  - 認證信息（username/password）
  - 資料庫名稱
  - 初始化 schema 腳本

- [ ] **Grafana Alert JSON 格式**：
  ```json
  // 實際的 webhook payload 格式範例
  {
    "incident_id": "?",
    "service_name": "?", 
    "alert_type": "?",
    "metrics": {
      // 實際的指標結構
    }
  }
  ```

#### 🏗️ **架構確認信息**
- [ ] **Go 平台啟動配置**：
  ```bash
  # 如何啟動 Go 平台服務？
  cd go-platform && go run cmd/detectviz/main.go --config=?
  ```

- [ ] **測試數據結構**：
  - InfluxDB 中的測試數據格式
  - 時間範圍和查詢模式
  - 預期的聚合結果範例

#### 📝 **MVP 範圍確認**
- [ ] **支援的指標類型**：CPU, Memory, Disk, Network？
- [ ] **分析邏輯複雜度**：簡單規則 vs 機器學習？
- [ ] **報告格式要求**：純 Markdown vs 圖表？
- [ ] **RAG 知識庫需求**：
  - ✅ **已確認**：使用 PostgreSQL 儲存結構化事件數據
  - ✅ **架構設計**：Provider 模式設計，支援未來彈性擴充替換
  - ✅ **檢索方式**：基於規則的相似事件匹配（服務名稱、指標類型、異常特徵）
  - [ ] **PostgreSQL 連接配置**：需要提供資料庫連接信息

### 🚀 **執行階段規劃**

#### 階段一：基礎設施準備（用戶負責）
```bash
1. 完成 Telegraf + InfluxDB + Grafana 設置
2. 設置 PostgreSQL 資料庫（知識庫用）
3. 提供 InfluxDB 和 PostgreSQL 連接配置
4. 確認 Alert JSON Payload 格式
5. 提供 Go 平台啟動配置
```

#### 階段二：Go 端實作（AI 負責）
```bash
1. 創建 health_aggregator 插件目錄和基礎結構
2. 實作 InfluxDB 查詢邏輯和並行處理
3. 建立 Go 平台啟動和配置機制
4. 編寫單元測試
```

#### 階段三：Python 端整合（AI 負責）
```bash
1. 移除模擬邏輯，啟用真實 RemoteTool 調用
2. 實作狀態傳遞和 Session State 管理
3. 完善分析邏輯和 Markdown 報告生成
4. 驗證報告內容的正確性和完整性
```

#### 階段四：端到端測試與報告驗證（AI 負責）
```bash
1. 建立自動化整合測試
2. 驗證完整工作流程和報告質量
3. 性能和錯誤處理測試
4. 報告解讀正確性驗證
```

#### 階段五：RAG 知識庫實作（核心功能）
```bash
1. 設計 Provider 架構：統一接口 + 可插拔實作
2. 實作 PostgreSQL Provider（生產環境）+ Memory Provider（測試環境）
3. 建立事件數據模型和相似性匹配邏輯
4. 實作 Knowledge Base Go 插件（knowledge.knowledge_base）
5. 整合到 report_writer Agent 的知識庫更新流程
6. 實作歷史事件檢索和相似性比對功能
7. 預留向量數據庫擴展接口（未來升級路徑）
```

#### 階段六：Grafana Dashboard 自動生成（增強功能）
```bash
1. 基於已驗證的報告內容，分析指標數據結構
2. 實作 Grafana Dashboard Builder Go 插件
3. 根據事件類型和指標，AI 自動生成相應的 Dashboard 配置
4. 整合到 report_writer Agent 工作流程
5. 文檔更新和功能驗證
```

---

### 具體開發任務 (請依序完成)

#### **任務一：實作 `HealthAggregator` Go 插件**

**目標**：賦予平台從 InfluxDB 實際查詢指標數據的能力。這是整個複盤流程的數據來源。

* **位置**: `go-platform/internal/pluginhost/plugins/observability/health_aggregator/`
* **實作步驟**:
    1.  **編輯 `plugin.go`**: 在 `Invoke` 函式中，解析從 Python 端傳入的 `InvokeRequest` payload，獲取 `service_name`, `time_range`, `metrics` 等參數。
    2.  **實現 InfluxDB 查詢**: 引入 InfluxDB Go client，編寫 Flux 查詢語句，從 InfluxDB 獲取對應的時序數據。查詢應支援並行執行以提升效能。
    3.  **數據處理**: 對查詢結果進行初步的聚合與格式化，將其轉換為結構化的 Go struct。
    4.  **回傳結果**: 將處理後的數據序列化，並透過 `InvokeResponse` 回傳給 Python ADK Runtime。
    5.  **編寫單元測試**: 建立 `plugin_test.go`，使用 mock 的 InfluxDB client 來驗證查詢邏輯和數據處理的正確性。

---

#### **任務二：完善 `postmortem_orchestrator` Agent 團隊的真實數據流**

**目標**：將 Agent 團隊從依賴模擬數據轉換為處理來自 Go 插件的真實數據，並實現狀態在子代理間的正確傳遞。

* **位置**: `python-adk-runtime/src/detectviz_adk/`
* **實作步驟**:
    1.  **修改 `adk_tools.py`**: **移除 `get_health_metrics_func` 中的模擬數據邏輯**，確保它現在會透過 `RemoteTool` 實際呼叫 Go 端的 `HealthAggregator` 插件。
    2.  **實現狀態傳遞**:
        * 在 `data_collector_agent` 成功獲取數據後，使用 `ToolContext` 將數據存入會話狀態 (Session State)。
        * 修改 `root_cause_analyzer` 的 instruction，使其能夠理解並分析從會話狀態中讀取到的結構化指標數據，並將其分析結果（例如，識別出的根因摘要）再次寫入會話狀態。
        * 修改 `report_writer` 的 instruction，使其能夠從會話狀態中讀取分析結果，並將其作為生成報告的依據。
    3.  **完善 `analyzer.py`**: 為 `root_cause_analyzer` Agent 增加初步的分析邏輯。對於 MVP 階段，這可以是一個基於規則的簡單實現（例如：如果 CPU 使用率 > 90% 且 P99 延遲 > 2000ms，則將根因歸結為「資源瓶頸」）。

---

#### **任務三：實作 Markdown 報告生成功能**

**目標**：將 `root_cause_analyzer` 的分析結果轉化為高質量的 Markdown 複盤報告，確保內容完整性和可讀性。

* **位置**: `python-adk-runtime/src/detectviz_adk/tools/adk_tools.py`
* **實作步驟**:

    **3.1 移除模擬邏輯並啟用真實數據處理**:
    1.  **清理模擬程式碼**: 移除 `generate_report_func` 中的所有模擬邏輯。
    2.  **啟用狀態讀取**: 從 Session State 中讀取真實的分析數據。
    3.  **數據驗證**: 確保輸入數據的完整性和格式正確性。

    **3.2 建立專業的報告模板系統**:
    1.  **模板檔案創建**: 在 `python-adk-runtime/templates/` 建立 `postmortem_report.md.j2` 模板。
    2.  **模板結構設計**: 包含執行摘要、事件時間線、根因分析、影響評估、改善建議、學習重點等章節。
    3.  **依賴管理**: 將 Jinja2 加入 `requirements.txt`。

    **3.3 實現智能模板渲染**:
    1.  **數據映射**: 將 Agent 分析結果映射到模板變數。
    2.  **條件渲染**: 根據數據可用性動態顯示報告章節。
    3.  **格式化處理**: 確保數值、時間、列表等格式正確顯示。

    **3.4 報告儲存和回傳**:
    1.  **檔案管理**: 將報告儲存到指定目錄，使用事件 ID 命名。
    2.  **元數據記錄**: 記錄報告生成時間、版本等元數據。
    3.  **回傳格式**: 提供檔案路徑和報告內容的結構化回傳。

---

#### **任務四：建立端到端 (E2E) 整合測試**

**目標**：建立一個自動化測試，模擬一次完整的複盤請求，驗證從 Agent 觸發到 Go 插件查詢，再回到 Agent 生成報告的整個流程。

* **位置**: `python-adk-runtime/tests/`
* **實作步驟**:
    1.  **建立 `test_e2e_postmortem.py`**: 新增一個整合測試檔案。
    2.  **準備測試環境**: 測試需要一個可用的 InfluxDB 實例（可以透過 Docker 啟動）並預先填入一些測試數據。
    3.  **編寫測試案例**:
        * 在測試開始時，以子進程方式啟動 `go-platform` 服務。
        * 使用 `PostmortemRunner` 或 `adk web` 入口，發送一個針對測試數據的複盤請求。
        * 斷言 (Assert) 最終產生的報告檔案是否存在，且其內容是否符合預期（例如，包含了基於測試數據分析得出的正確根因）。
        * 在測試結束時，確保 Go 平台服務被正常關閉。

---

## 🏗️ **Provider 架構設計說明**

### 知識庫 Provider 統一接口
```go
// providers/interface.go
type KnowledgeProvider interface {
    // 事件管理
    StoreIncident(ctx context.Context, incident *models.Incident) error
    GetIncident(ctx context.Context, incidentID string) (*models.Incident, error)
    
    // 教訓學習
    StoreLessons(ctx context.Context, lessons []models.Lesson) error
    GetLessons(ctx context.Context, incidentID string) ([]models.Lesson, error)
    
    // 相似性檢索
    FindSimilarIncidents(ctx context.Context, criteria models.SearchCriteria) ([]models.Incident, error)
    
    // 健康檢查
    HealthCheck(ctx context.Context) error
}
```

### PostgreSQL Provider 實作
```go
// providers/postgresql/provider.go
type PostgreSQLProvider struct {
    db *sql.DB
    config PostgreSQLConfig
}

// 支援的檢索維度
type SearchCriteria struct {
    ServiceName   string
    MetricTypes   []string  // ["cpu", "memory", "network"]
    AnomalyTypes  []string  // ["spike", "drop", "high_variance"]
    TimeWindow    time.Duration
    Similarity    float64   // 相似度閾值
}
```

### 未來擴展路徑
- **向量數據庫 Provider**：ChromaDB、Weaviate、Pinecone
- **混合 Provider**：PostgreSQL（結構化）+ ChromaDB（語意搜尋）
- **雲端 Provider**：AWS OpenSearch、GCP Vertex AI Matching Engine

---

## 📦 **交付成果檢查清單**

### 🔧 **Go 端組件**
- [ ] `go-platform/internal/pluginhost/plugins/observability/health_aggregator/`
  - [ ] `plugin.go` - InfluxDB 查詢和數據聚合
  - [ ] `plugin_test.go` - 單元測試
  - [ ] `module.card.json` - 插件元數據
- [ ] `go-platform/internal/pluginhost/plugins/reporting/dashboard_builder/`
  - [ ] `plugin.go` - Grafana Dashboard 自動創建
  - [ ] `grafana_client.go` - Grafana API 客戶端
- [ ] `go-platform/internal/pluginhost/plugins/knowledge/knowledge_base/`
  - [ ] `plugin.go` - 知識庫插件主入口
  - [ ] `providers/` - Provider 架構目錄
    - [ ] `interface.go` - 知識庫 Provider 統一接口
    - [ ] `postgresql/` - PostgreSQL Provider 實作
      - [ ] `provider.go` - PostgreSQL 知識庫實作
      - [ ] `schema.sql` - 資料庫 Schema 定義
      - [ ] `queries.sql` - 預定義查詢語句
    - [ ] `memory/` - 記憶體 Provider（測試用）
      - [ ] `provider.go` - 記憶體實作
  - [ ] `models/` - 數據模型定義
    - [ ] `incident.go` - 事件模型
    - [ ] `lesson.go` - 教訓模型
  - [ ] `similarity/` - 相似性計算
    - [ ] `matcher.go` - 基於規則的事件匹配
  - [ ] `knowledge_store_test.go` - 單元測試

### 🐍 **Python 端組件**
- [ ] **移除所有模擬邏輯**：
  - [ ] `python-adk-runtime/src/detectviz_adk/tools/adk_tools.py`
  - [ ] 啟用真實的 RemoteTool 調用
- [ ] **增強 Agent 功能**：
  - [ ] `analyzer.py` - 基於規則的分析邏輯
  - [ ] 狀態傳遞和 Session State 管理
- [ ] **報告模板**：
  - [ ] Markdown 報告模板文件
  - [ ] Jinja2 模板渲染邏輯

### 🧪 **測試基礎設施**
- [ ] `python-adk-runtime/tests/test_e2e_postmortem.py` - 端到端整合測試
- [ ] Go 端各插件的單元測試
- [ ] 測試用的 InfluxDB 數據準備腳本
- [ ] 自動化測試執行文檔

### 📚 **文檔更新**
- [ ] **執行說明**：描述如何運行新的端到端測試
- [ ] **配置指南**：InfluxDB、Grafana 連接配置
- [ ] **API 文檔**：新增插件的 API 規格
- [ ] **故障排除**：常見問題和解決方案

### 🎯 **功能驗證項目**
- [ ] **數據流驗證**：InfluxDB → HealthAggregator → Agent 分析
- [ ] **報告生成**：完整的 Markdown 報告輸出
- [ ] **知識庫功能**：
  - [ ] 事件儲存到 PostgreSQL
  - [ ] 教訓學習結構化儲存
  - [ ] 相似事件檢索和匹配
  - [ ] Provider 切換功能（PostgreSQL ↔ Memory）
- [ ] **Dashboard 創建**：自動生成的 Grafana Dashboard（階段六）
- [ ] **錯誤處理**：網路失敗、數據缺失等異常情況

### 🔄 **整合測試目標**
- [ ] **端到端流程**：從 Alert Webhook → 最終報告的完整鏈路
- [ ] **並行處理**：多個指標同時查詢的性能驗證
- [ ] **時間範圍**：不同時間窗口的數據聚合正確性
- [ ] **多服務**：同時處理多個服務的事件數據

---

**⚠️ 重要提醒**：
在開始實作前，請確保前置需求清單中的所有項目都已完成。特別是 InfluxDB 連接配置和測試數據準備，這是整個實作的基礎。

**✅ 準備就緒後**，請立即開始實作。