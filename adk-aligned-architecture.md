# ADK 對齊架構調整方案

基於 [adk_tutorial.ipynb](./adk_tutorial.ipynb) 的官方範例，我發現我們的架構有幾個地方需要調整：

### ❌ 關鍵差異

1. **Agent 實現方式錯誤**
   - 我們：繼承自定義的 BaseAgent
   - ADK：應直接使用 `google.adk.Agent`

2. **缺少 Agent 團隊協作模式**
   - 我們：單一 Agent 做所有事
   - ADK：Root Agent + Sub-Agents 委派模式

3. **Tool 定義不符合規範**
   - 我們：RemoteTool 直接調用
   - ADK：使用 `@adk.tool` 裝飾器

4. **缺少核心組件**
   - 缺少 Runner
   - 缺少 SessionService
   - 缺少 ToolContext

### ✅ 調整建議

我已經在上面的文件中提供了完整的調整方案，包括：

1. **正確的 Agent 團隊結構**
   - PostmortemOrchestrator（Root Agent）
   - DataCollectorAgent（Sub-Agent）
   - RootCauseAnalyzer（Sub-Agent）
   - ReportWriter（Sub-Agent）

2. **正確的 Tool 實現**
   - 使用 @adk.tool 裝飾器
   - RemoteTool 包裝在 ADK Tool 內部

3. **完整的執行流程**
   - Runner + SessionService
   - 正確的會話管理
   - 狀態持久化

### 🎯 影響評估

**如果不調整**：
- 無法充分利用 ADK 的 Agent 協作能力
- 缺少會話管理，無法保持上下文
- 不符合 Google ADK 生態系統標準

**調整後的優勢**：
- 完全符合 ADK 最佳實踐
- 可以利用 ADK 的所有高級功能
- 更容易擴展和維護


### ✅ 對齊 ADK 結構
```bash
python-adk-runtime/
├── src/detectviz_adk/
│   ├── agents/
│   │   ├── __init__.py
│   │   ├── postmortem/
│   │   │   ├── orchestrator.py      # Root Agent
│   │   │   ├── data_collector.py    # Sub Agent
│   │   │   ├── analyzer.py          # Sub Agent
│   │   │   └── report_writer.py     # Sub Agent
│   │   └── registry.py              # Agent 註冊
│   │
│   ├── tools/
│   │   ├── __init__.py
│   │   ├── health_metrics.py        # ADK Tool
│   │   ├── report_generator.py      # ADK Tool
│   │   └── remote_tool.py           # RemoteTool 包裝器
│   │
│   ├── runners/
│   │   ├── __init__.py
│   │   └── postmortem_runner.py     # Runner 實現
│   │
│   └── sessions/
│       ├── __init__.py
│       └── session_manager.py       # Session 管理
```


## 主要調整總結

### 🔴 必須調整項目

1. **使用 ADK Agent 類**，不是自定義 BaseAgent
2. **實現 Sub-Agents 委派模式**
3. **使用 @adk.tool 裝飾器**定義工具
4. **實現 Runner 和 SessionService**
5. **使用 ToolContext** 進行狀態管理

### 🟡 建議調整項目

1. 將 RemoteTool 包裝在 ADK Tool 內部
2. 實現正確的錯誤處理模式
3. 使用 ADK 的測試模式

### 🟢 已符合項目

1. 使用 Gemini 模型
2. Agent 職責分離
3. 工具與決策分離

這樣調整後，我們的架構將完全符合 ADK 官方最佳實踐！

### 📊 關鍵差異對比

| ADK 最佳實踐 | 我們的當前設計 | 符合度 | 需要調整 |
|-------------|--------------|--------|----------|
| **Agent 繼承 ADK Agent 類** | 繼承 BaseAgent（自定義） | ⚠️ 部分 | 需要直接使用 ADK Agent |
| **Tool 定義方式** | RemoteTool（gRPC） | ⚠️ 不同 | 應使用 ADK Tool 裝飾器 |
| **Agent 委派模式** | 未明確實現 | ❌ 缺失 | 需要實現 sub_agents |
| **SessionService** | 未提及 | ❌ 缺失 | 需要添加會話管理 |
| **Runner 模式** | 未實現 | ❌ 缺失 | 需要實現 Runner |
| **Model 配置** | 有配置 | ✅ 符合 | gemini-2.0-flash |

## 1. Agent 實現調整（遵循 ADK 官方模式）

### ❌ 當前錯誤實現
```python
# 我們現在的方式（不符合 ADK）
from detectviz_adk.agents.base import BaseAgent

class PostmortemOrchestratorAgent(BaseAgent):
    def __init__(self):
        super().__init__()
        self.health_aggregator = RemoteTool("health_aggregator")
```

### ✅ 正確的 ADK 實現
```python
# 遵循 adk_tutorial.ipynb 的正確方式
from google import adk

# 1. 定義 Tool（使用 ADK 裝飾器）
@adk.tool
async def get_health_metrics(
    service_name: str,
    time_range: dict,
    metrics: list[str] = ["error_rate", "latency", "cpu_usage"]
) -> dict:
    """
    查詢服務健康指標
    
    Args:
        service_name: 服務名稱
        time_range: 時間範圍 {"start": "...", "end": "..."}
        metrics: 要查詢的指標列表
    
    Returns:
        包含健康指標的字典
    """
    # 這裡調用 RemoteTool 到 Go 端
    remote_tool = RemoteTool("health_aggregator")
    return await remote_tool.invoke({
        "service": service_name,
        "time_range": time_range,
        "metrics": metrics
    })

@adk.tool
async def generate_report(
    incident_data: dict,
    format: str = "markdown"
) -> str:
    """生成事後複盤報告"""
    remote_tool = RemoteTool("report_generator")
    return await remote_tool.invoke({
        "data": incident_data,
        "format": format
    })

# 2. 定義 Sub-Agents（專門化的子代理）
data_collector_agent = adk.Agent(
    model="gemini-2.0-flash",
    name="data_collector",
    instruction="""你是數據收集專員。你的職責是：
    1. 從各個數據源收集相關指標
    2. 確保數據的完整性和時間對齊
    3. 返回結構化的數據集合
    使用 get_health_metrics 工具獲取數據。""",
    description="負責收集和整理事故相關數據",
    tools=[get_health_metrics]
)

root_cause_analyzer = adk.Agent(
    model="gemini-2.0-flash",
    name="root_cause_analyzer",
    instruction="""你是根因分析專家。基於收集的數據：
    1. 識別異常模式
    2. 關聯相關事件
    3. 推斷根本原因
    不要猜測，只基於數據分析。""",
    description="分析數據並識別根本原因",
    tools=[]  # 純分析，不需要工具
)

report_writer = adk.Agent(
    model="gemini-2.0-flash",
    name="report_writer",
    instruction="""你是技術文檔專家。你的任務是：
    1. 將分析結果整理成專業報告
    2. 包含時間線、根因、影響和建議
    3. 使用 generate_report 工具生成最終報告
    """,
    description="生成專業的事後複盤報告",
    tools=[generate_report]
)

# 3. 定義 Root Agent（協調器）
postmortem_orchestrator = adk.Agent(
    model="gemini-2.0-flash",
    name="postmortem_orchestrator",
    instruction="""你是事後複盤協調器，負責管理整個複盤流程。
    
    你有以下子代理可以委派任務：
    1. 'data_collector': 收集事故數據
    2. 'root_cause_analyzer': 分析根本原因
    3. 'report_writer': 生成報告
    
    工作流程：
    1. 首先委派 data_collector 收集所需數據
    2. 將數據交給 root_cause_analyzer 進行分析
    3. 最後讓 report_writer 生成完整報告
    
    協調各個代理完成任務，確保流程順暢。""",
    description="協調事後複盤流程的主代理",
    tools=[],  # Root Agent 不直接使用工具
    sub_agents=[data_collector_agent, root_cause_analyzer, report_writer]
)
```

## 2. Session 和 Runner 實現（必需組件）

### ✅ 正確實現會話管理
```python
from google.adk import Runner, InMemorySessionService
import asyncio

# 創建會話服務
session_service = InMemorySessionService()

# 創建 Runner
runner = Runner(
    agent=postmortem_orchestrator,
    app_name="detectviz_postmortem",
    session_service=session_service
)

# 執行複盤分析
async def execute_postmortem(incident_request):
    """執行事後複盤分析"""
    
    # 創建會話
    session = await session_service.create_session(
        app_name="detectviz_postmortem",
        user_id="system",
        session_id=f"incident-{incident_request['incident_id']}"
    )
    
    # 構建查詢
    query = f"""
    請為事件 {incident_request['incident_id']} 執行完整的事後複盤分析。
    
    事件詳情：
    - 時間範圍：{incident_request['time_range']['start']} 到 {incident_request['time_range']['end']}
    - 受影響服務：{', '.join(incident_request['affected_services'])}
    - 嚴重程度：{incident_request['severity']}
    
    請收集數據、分析根因並生成報告。
    """
    
    # 執行代理
    response = await runner.run(
        query=query,
        user_id="system",
        session_id=session.id
    )
    
    return response

# 使用範例
async def main():
    incident = {
        "incident_id": "INC-2024-001",
        "time_range": {
            "start": "2024-01-15T10:00:00Z",
            "end": "2024-01-15T12:00:00Z"
        },
        "affected_services": ["payment-service", "api-gateway"],
        "severity": "P2"
    }
    
    result = await execute_postmortem(incident)
    print(result)

# 運行
asyncio.run(main())
```

## 3. 工具設計調整（ADK Tool Pattern）

### ✅ 正確的 Tool 包裝模式
```python
# tools/health_aggregator_tool.py
from google import adk
from detectviz_adk.tools.remote_tool import RemoteTool

@adk.tool
async def query_service_health(
    service: str,
    start_time: str,
    end_time: str,
    metrics: list[str] = None
) -> dict:
    """
    查詢服務健康數據（ADK Tool）
    
    這是一個 ADK Tool，內部調用 Go 端的 RemoteTool
    """
    # Tool 只負責執行，不做決策
    remote = RemoteTool(
        tool_id="observability.health_aggregator",
        tool_version="0.1.0"
    )
    
    try:
        result = await remote.invoke({
            "action": "query_health",
            "params": {
                "service": service,
                "time_range": {
                    "start": start_time,
                    "end": end_time
                },
                "metrics": metrics or ["error_rate", "latency", "cpu"]
            }
        })
        return result
    except Exception as e:
        return {"error": str(e), "status": "failed"}

@adk.tool  
async def create_dashboard(
    incident_id: str,
    panels: list[dict],
    time_range: dict
) -> str:
    """創建 Grafana 儀表板"""
    remote = RemoteTool(
        tool_id="reporting.dashboard_builder",
        tool_version="0.1.0"
    )
    
    result = await remote.invoke({
        "action": "create_dashboard",
        "incident_id": incident_id,
        "panels": panels,
        "time_range": time_range
    })
    
    return result.get("dashboard_url", "")
```

## 4. 記憶體管理（Session State）

### ✅ 使用 ADK Session State
```python
from google.adk import ToolContext

@adk.tool
async def remember_analysis(
    ctx: ToolContext,
    key: str,
    value: any
) -> str:
    """在會話中記住分析結果"""
    # 使用 ADK 的 session state
    ctx.session.state[key] = value
    return f"已記住 {key}"

@adk.tool
async def recall_analysis(
    ctx: ToolContext,
    key: str
) -> any:
    """從會話中回憶分析結果"""
    return ctx.session.state.get(key, None)

# Agent 使用記憶
memory_agent = adk.Agent(
    model="gemini-2.0-flash",
    name="memory_agent",
    instruction="""你可以記住和回憶分析過程中的信息。
    使用 remember_analysis 保存重要發現。
    使用 recall_analysis 獲取之前的分析結果。""",
    tools=[remember_analysis, recall_analysis]
)
```

## 5. 測試模式（遵循 ADK 範例）

### ✅ 正確的測試方式
```python
# tests/test_postmortem_agent.py
import asyncio
from google.adk import Runner, InMemorySessionService

async def test_agent_delegation():
    """測試 Agent 委派機制"""
    
    # 創建測試會話
    session_service = InMemorySessionService()
    
    # 創建 Runner
    runner = Runner(
        agent=postmortem_orchestrator,
        app_name="test_postmortem",
        session_service=session_service
    )
    
    # 測試簡單查詢
    response = await runner.run(
        query="分析昨天下午 3 點的服務中斷事件",
        user_id="test_user",
        session_id="test_session_001"
    )
    
    # 驗證委派發生
    assert "data_collector" in response
    assert "root_cause_analyzer" in response
    assert "report_writer" in response

# 運行測試
asyncio.run(test_agent_delegation())
```

## 6. 配置調整

### ✅ ADK 標準配置
```python
# config.py
ADK_CONFIG = {
    "agents": {
        "postmortem_orchestrator": {
            "model": "gemini-2.0-flash",
            "temperature": 0.3,  # 降低隨機性
            "max_tokens": 2048,
            "tools_config": {
                "function_calling_config": {
                    "mode": "AUTO"  # 自動決定何時調用工具
                }
            }
        }
    },
    "session": {
        "backend": "memory",  # 或 "redis" for production
        "ttl_seconds": 3600
    },
    "runner": {
        "timeout_seconds": 300,
        "max_retries": 3
    }
}
```

## 7. 目錄結構調整

## 8. 需要更新的文件

基於這個發現，以下文件需要更新：
1. `TODO.md` - 調整實作任務
2. `CLAUDE.md` - 更新開發規範
3. `python-adk-runtime/README.md` - 更新架構說明
4. `docs/sre-services-map.md` - 調整 Agent 協作模式
