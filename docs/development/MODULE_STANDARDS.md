# 模組開發標準 - SSOT

> 📌 **文檔職責**：本文檔定義統一的模組開發標準，確保代碼品質、架構一致性和可維護性。

## 🎯 模組分類標準

### 角色定義 (Role)
```
agent.coordinator      - 協調型Agent，負責工作流編排
agent.tool_exec       - 工具執行型Agent，負責具體任務執行
tool                  - 工具組件，提供特定功能接口
capability            - 能力組件，提供可重用的業務邏輯
plugin.gateway        - 插件網關，處理外部系統集成
memory.backend        - 記憶後端，提供數據存儲能力
security.module       - 安全模組，提供認證授權功能
observability.module  - 可觀察性模組，提供監控指標
storage.module        - 存儲模組，提供數據持久化
```

### 分類細分 (Category)
#### 插件分類
- `collector.input` - 數據收集器
- `transform.processor` - 數據轉換器  
- `aggregate.aggregator` - 數據聚合器
- `sink.output` - 數據輸出器
- `gateway` - 網關類插件

#### ADK/平台分類
- `llm` - 語言模型相關
- `retriever` - 檢索器
- `workflow` - 工作流程
- `a2a` - Agent間通訊
- `capability` - 通用能力

#### 觀測/安全/儲存分類
- `observability.exporter` - 可觀察性導出器
- `observability.processor` - 可觀察性處理器
- `security.authn` - 身份認證
- `security.authz` - 權限授權
- `memory.backend` - 記憶後端
- `storage.blob` - 對象存儲
- `storage.kv` - 鍵值存儲
- `storage.vector` - 向量存儲

## 📋 代碼品質要求

### Go 模組標準
```go
// 1. 包註釋必須清晰說明功能
// Package health provides health monitoring and aggregation capabilities
// for the Detectviz platform.
package health

// 2. 結構體定義必須有完整註釋
// HealthAggregatorPlugin implements health data aggregation
// from multiple monitoring sources.
type HealthAggregatorPlugin struct {
    client influxdb2.Client  // InfluxDB客戶端
    logger *zap.Logger       // 結構化日誌記錄器
    config PluginConfig      // 插件配置
}

// 3. 公開方法必須有詳細文檔
// Invoke executes health data aggregation based on the provided request.
// 
// Parameters:
//   - ctx: 上下文，用於取消和超時控制
//   - req: 包含查詢參數的請求對象
//
// Returns:
//   - *pb.InvokeResponse: 聚合後的健康數據
//   - error: 執行過程中的錯誤
func (h *HealthAggregatorPlugin) Invoke(ctx context.Context, req *pb.InvokeRequest) (*pb.InvokeResponse, error) {
    // 實現細節...
}

// 4. 錯誤處理標準
func (h *HealthAggregatorPlugin) queryMetrics(ctx context.Context, query string) (*QueryResult, error) {
    result, err := h.client.Query(ctx, query)
    if err != nil {
        // 使用structured logging記錄錯誤
        h.logger.Error("查詢指標失敗",
            zap.String("query", query),
            zap.Error(err),
        )
        return nil, fmt.Errorf("查詢指標失敗: %w", err)
    }
    return result, nil
}
```

### Python 模組標準
```python
"""
模組級別文檔：清晰說明模組用途和主要組件
"""
from typing import Dict, List, Optional, Any, Union
import logging
from dataclasses import dataclass

# 1. 類型註釋必須完整
@dataclass
class AgentConfig:
    """Agent配置類，定義Agent運行所需的所有參數"""
    name: str
    model_provider: str
    temperature: float = 0.7
    max_tokens: int = 1000
    tools: List[str] = None

    def __post_init__(self):
        if self.tools is None:
            self.tools = []

# 2. 類定義必須有清晰的文檔
class PostmortemOrchestrator:
    """事後複盤協調器，負責管理整個複盤分析流程
    
    這個類作為ADK Root Agent，協調子Agent完成事故分析任務。
    不直接執行工具操作，而是通過委派給專門的子Agent來完成。
    
    Attributes:
        name: Agent名稱
        sub_agents: 子Agent列表
        session_manager: 會話狀態管理器
    """
    
    def __init__(self, config: AgentConfig) -> None:
        self.name = config.name
        self.config = config
        self.logger = logging.getLogger(f"detectviz.agent.{config.name}")
        
    async def analyze_incident(
        self,
        incident_data: Dict[str, Any],
        context: Optional[Dict[str, Any]] = None
    ) -> Dict[str, Any]:
        """分析事故並生成複盤報告
        
        Args:
            incident_data: 事故基本信息
            context: 額外的上下文信息
            
        Returns:
            包含分析結果的字典
            
        Raises:
            AnalysisError: 分析過程中發生錯誤
        """
        try:
            # 實現細節...
            pass
        except Exception as e:
            self.logger.error(f"事故分析失敗: {e}", exc_info=True)
            raise AnalysisError(f"Agent {self.name} 分析失敗") from e

# 3. 錯誤處理標準
class AnalysisError(Exception):
    """分析過程中的錯誤"""
    pass

class DataCollectionError(Exception):
    """數據收集過程中的錯誤"""
    pass
```

## 🔍 可觀察性整合標準

### 分散式追蹤
```go
// Go端OpenTelemetry整合
import (
    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/trace"
)

func (h *HealthAggregatorPlugin) Invoke(ctx context.Context, req *pb.InvokeRequest) (*pb.InvokeResponse, error) {
    tracer := otel.Tracer("detectviz.plugin.health_aggregator")
    
    ctx, span := tracer.Start(ctx, "health.aggregate")
    defer span.End()
    
    span.SetAttributes(
        attribute.String("plugin.name", "health_aggregator"),
        attribute.String("plugin.version", "0.1.0"),
    )
    
    // 業務邏輯實現
    result, err := h.processRequest(ctx, req)
    if err != nil {
        span.SetStatus(codes.Error, err.Error())
        return nil, err
    }
    
    span.SetStatus(codes.Ok, "健康數據聚合完成")
    return result, nil
}
```

```python
# Python端分散式追蹤
from opentelemetry import trace
from opentelemetry.trace import Status, StatusCode

tracer = trace.get_tracer("detectviz.agent.postmortem")

async def analyze_incident(self, incident_data: Dict[str, Any]) -> Dict[str, Any]:
    with tracer.start_as_current_span("postmortem.analyze") as span:
        span.set_attribute("agent.name", self.name)
        span.set_attribute("incident.type", incident_data.get("type", "unknown"))
        
        try:
            result = await self._analyze_internal(incident_data)
            span.set_status(Status(StatusCode.OK, "分析完成"))
            return result
        except Exception as e:
            span.set_status(Status(StatusCode.ERROR, str(e)))
            raise
```

### 指標收集
```go
// Go端Prometheus指標
import (
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promauto"
)

var (
    pluginInvocations = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "detectviz_plugin_invocations_total",
            Help: "插件調用總次數",
        },
        []string{"plugin_name", "status"},
    )
    
    pluginDuration = promauto.NewHistogramVec(
        prometheus.HistogramOpts{
            Name: "detectviz_plugin_duration_seconds",
            Help: "插件執行時間",
        },
        []string{"plugin_name"},
    )
)

func (h *HealthAggregatorPlugin) Invoke(ctx context.Context, req *pb.InvokeRequest) (*pb.InvokeResponse, error) {
    timer := prometheus.NewTimer(pluginDuration.WithLabelValues("health_aggregator"))
    defer timer.ObserveDuration()
    
    // 業務邏輯
    result, err := h.processRequest(ctx, req)
    
    status := "success"
    if err != nil {
        status = "error"
    }
    pluginInvocations.WithLabelValues("health_aggregator", status).Inc()
    
    return result, err
}
```

```python
# Python端指標收集
from prometheus_client import Counter, Histogram, Gauge

agent_requests_total = Counter(
    'detectviz_agent_requests_total',
    'Agent請求總數',
    ['agent_name', 'status']
)

agent_response_time = Histogram(
    'detectviz_agent_response_time_seconds',
    'Agent響應時間',
    ['agent_name']
)

async def analyze_incident(self, incident_data: Dict[str, Any]) -> Dict[str, Any]:
    with agent_response_time.labels(agent_name=self.name).time():
        try:
            result = await self._analyze_internal(incident_data)
            agent_requests_total.labels(
                agent_name=self.name, 
                status='success'
            ).inc()
            return result
        except Exception as e:
            agent_requests_total.labels(
                agent_name=self.name, 
                status='error'
            ).inc()
            raise
```

## 🧪 測試標準

### 單元測試要求
- 測試覆蓋率 > 90%
- 每個公開方法都有測試
- 包含正常流程和異常流程測試
- 使用模擬對象隔離外部依賴

### Go測試範例
```go
func TestHealthAggregatorPlugin_Invoke(t *testing.T) {
    tests := []struct {
        name        string
        request     *pb.InvokeRequest
        mockSetup   func(*MockInfluxClient)
        wantErr     bool
        wantStatus  string
    }{
        {
            name: "成功聚合健康數據",
            request: &pb.InvokeRequest{
                Payload: mustMarshalStruct(map[string]interface{}{
                    "service_name": "test-service",
                    "time_range": "1h",
                }),
            },
            mockSetup: func(m *MockInfluxClient) {
                m.EXPECT().Query(gomock.Any(), gomock.Any()).
                    Return(&QueryResult{Values: []interface{}{1.0}}, nil)
            },
            wantErr: false,
            wantStatus: "success",
        },
        {
            name: "數據庫查詢失敗",
            request: &pb.InvokeRequest{
                Payload: mustMarshalStruct(map[string]interface{}{
                    "service_name": "test-service",
                }),
            },
            mockSetup: func(m *MockInfluxClient) {
                m.EXPECT().Query(gomock.Any(), gomock.Any()).
                    Return(nil, errors.New("連接失敗"))
            },
            wantErr: true,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            ctrl := gomock.NewController(t)
            defer ctrl.Finish()
            
            mockClient := NewMockInfluxClient(ctrl)
            tt.mockSetup(mockClient)
            
            plugin := &HealthAggregatorPlugin{
                client: mockClient,
                logger: zap.NewNop(),
            }
            
            resp, err := plugin.Invoke(context.Background(), tt.request)
            
            if tt.wantErr {
                assert.Error(t, err)
            } else {
                assert.NoError(t, err)
                assert.Equal(t, tt.wantStatus, resp.Status)
            }
        })
    }
}
```

### Python測試範例
```python
import pytest
from unittest.mock import AsyncMock, Mock
from detectviz_adk.agents.postmortem.orchestrator import PostmortemOrchestrator

@pytest.mark.asyncio
class TestPostmortemOrchestrator:
    
    async def test_analyze_incident_success(self):
        """測試成功的事故分析流程"""
        # Given
        config = AgentConfig(name="test_orchestrator")
        orchestrator = PostmortemOrchestrator(config)
        
        incident_data = {
            "id": "INC-001",
            "service": "api-service",
            "start_time": "2024-12-17T10:00:00Z"
        }
        
        # Mock子Agent
        mock_data_collector = AsyncMock()
        mock_data_collector.collect_data.return_value = {"metrics": []}
        orchestrator.data_collector = mock_data_collector
        
        # When
        result = await orchestrator.analyze_incident(incident_data)
        
        # Then
        assert result is not None
        assert "analysis" in result
        mock_data_collector.collect_data.assert_called_once()
    
    async def test_analyze_incident_data_collection_failure(self):
        """測試數據收集失敗的情況"""
        # Given
        config = AgentConfig(name="test_orchestrator")
        orchestrator = PostmortemOrchestrator(config)
        
        incident_data = {"id": "INC-001"}
        
        # Mock失敗的數據收集
        mock_data_collector = AsyncMock()
        mock_data_collector.collect_data.side_effect = DataCollectionError("無法連接到監控系統")
        orchestrator.data_collector = mock_data_collector
        
        # When & Then
        with pytest.raises(AnalysisError) as exc_info:
            await orchestrator.analyze_incident(incident_data)
        
        assert "數據收集失敗" in str(exc_info.value)
```

## 🔒 安全規範

### 敏感信息管理
```yaml
# 配置文件中不應包含敏感信息
database:
  host: "${DB_HOST:-localhost}"
  port: "${DB_PORT:-5432}"
  username: "${DB_USERNAME}"
  password: "${DB_PASSWORD}"  # 必須通過環境變數
  
api:
  key: "${API_KEY}"  # 不能硬編碼
  secret: "${API_SECRET}"
```

### 輸入驗證
```go
// Go端輸入驗證
func (h *HealthAggregatorPlugin) validateRequest(req *pb.InvokeRequest) error {
    if req.Payload == nil {
        return fmt.Errorf("請求payload不能為空")
    }
    
    // 驗證payload大小
    if proto.Size(req.Payload) > MaxPayloadSize {
        return fmt.Errorf("請求payload過大: %d > %d", proto.Size(req.Payload), MaxPayloadSize)
    }
    
    // 解析並驗證業務參數
    var params HealthQueryParams
    if err := json.Unmarshal(req.Payload.Value, &params); err != nil {
        return fmt.Errorf("無效的請求格式: %w", err)
    }
    
    if params.ServiceName == "" {
        return fmt.Errorf("服務名稱不能為空")
    }
    
    return nil
}
```

```python
# Python端輸入驗證
from pydantic import BaseModel, validator

class IncidentRequest(BaseModel):
    """事故請求模型，包含完整的數據驗證"""
    id: str
    service_name: str
    start_time: datetime
    end_time: Optional[datetime] = None
    
    @validator('id')
    def validate_id(cls, v):
        if not v or len(v) < 3:
            raise ValueError('事故ID長度必須至少3個字符')
        return v
    
    @validator('service_name')
    def validate_service_name(cls, v):
        if not v or not v.replace('-', '').replace('_', '').isalnum():
            raise ValueError('服務名稱只能包含字母、數字、連字符和下劃線')
        return v

async def analyze_incident(self, request_data: Dict[str, Any]) -> Dict[str, Any]:
    # 使用Pydantic進行數據驗證
    try:
        incident = IncidentRequest(**request_data)
    except ValidationError as e:
        raise ValueError(f"請求數據驗證失敗: {e}")
    
    # 業務邏輯處理...
```

## 📝 文檔標準

### 模組文檔結構
每個模組都應包含：
1. **README.md** - 模組概述和快速開始
2. **llm.txt** - AI協作專用指南
3. **module.card.json** - 模組元數據
4. **API文檔** - 接口說明和使用範例

### 模組卡範例
```json
{
  "name": "detectviz.plugins.health_aggregator",
  "version": "0.1.0",
  "description": "健康數據聚合插件，從多個監控源收集和聚合健康指標",
  "entrypoint": "internal/pluginhost/plugins/observability/health_aggregator/plugin.go",
  "language": "go",
  "role": "plugin.gateway",
  "category": "aggregate.aggregator",
  "requires": [
    {
      "name": "influxdb-client",
      "version": ">=2.0.0"
    }
  ],
  "resources": {
    "memory_mb": 128,
    "cpu_millicores": 100
  },
  "config": {
    "schema_uri": "contracts/schemas/health_aggregator.schema.json"
  },
  "observability": {
    "tags": ["health", "monitoring", "aggregation"]
  },
  "contracts": {
    "min_proto": "0.1.0"
  }
}
```

---

**維護說明**：
- 更新頻率：代碼標準變更時更新
- 維護責任：技術架構師
- 引用方式：`{{ docs/development/MODULE_STANDARDS.md#section }}`