# Quick Reference Guide

## 快速決策矩陣

### Agent 架構模式選擇

```bash
需求特徵                           → 推薦模式              → 範本
─────────────────────────────────────────────────────────────
單一功能 + 工具豐富                → Simple Agent        → simple_agent
多專家協作 + 決策複雜              → Coordinator         → coordinator_agent  
階段性流程 + 順序依賴              → Hierarchy          → hierarchy_agent
迭代優化 + 複雜工作流              → Workflow           → workflow_agent
```

### Tool vs Go Plugin 選擇

```bash
場景                              → 建議技術棧           → 理由
─────────────────────────────────────────────────────────────
外部 API 調用                     → Go Plugin          → 高併發、連接池
資料庫操作                        → Go Plugin          → 連接管理、事務
檔案處理                          → Go Plugin          → I/O 效能
ML/AI 推理                       → Python Tool        → 生態豐富
文本處理                          → Python Tool        → NLP 庫豐富
數據分析                          → Python Tool        → Pandas/NumPy
業務邏輯                          → Python Tool        → 快速迭代
```

### Sub-Agent 共享策略

```bash
Agent 特性                        → 共享策略             → 實作方式
─────────────────────────────────────────────────────────────
有狀態 + 會話記憶                 → 獨立實例             → Factory Pattern
無狀態 + 純函數                   → 共享實例             → Stateless Sharing
高併發 + 資源密集                 → 池化管理             → Agent Pool
```

## 核心 API 速查

### Agent 創建

#### Simple Agent Pattern
```python
from detectviz_adk.agents import Agent
from detectviz_adk.tools import ToolRegistry

root_agent = Agent(
    name="my_service_agent",
    model="gemini-2.5-flash",
    instruction="服務描述和行為指令",
    tools=ToolRegistry.get_tools(["tool1", "tool2"]),
    before_tool_callback=before_tool_callback,
    after_tool_callback=after_tool_callback,
)
```

#### Coordinator Pattern
```python
from detectviz_adk.agents import LlmAgent
from detectviz_adk.tools import AgentTool

coordinator = LlmAgent(
    name="my_coordinator",
    model="gemini-2.5-pro",
    instruction="協調器邏輯描述",
    tools=[
        AgentTool(agent=expert_agent_1),
        AgentTool(agent=expert_agent_2),
        AgentTool(agent=expert_agent_3),
    ]
)
```

#### Hierarchy Pattern
```python
from detectviz_adk.agents import Agent

root_agent = Agent(
    name="my_hierarchy_agent",
    model="gemini-2.5-flash",
    instruction="階層協調描述",
    sub_agents=[
        phase1_agent,
        phase2_agent, 
        phase3_agent,
    ],
    before_agent_callback=load_context_callback
)
```

#### Workflow Pattern
```python
from detectviz_adk.agents import SequentialAgent, LoopAgent

workflow = SequentialAgent(
    name="my_workflow",
    sub_agents=[step1_agent, step2_agent, step3_agent]
)

root_agent = LoopAgent(
    name="my_iterative_agent", 
    sub_agents=[workflow, checker_agent],
    termination_condition=quality_check
)
```

### Tool 實作

#### Python Local Tool
```python
from detectviz_adk.tools import BaseTool

class MyLocalTool(BaseTool):
    name = "my_local_tool"
    version = "1.0.0"
    description = "本地工具描述"
    
    async def execute(self, input_data: str, **kwargs) -> str:
        # 實作工具邏輯
        result = await self._process(input_data)
        return result
```

#### RemoteTool (連接 Go Plugin)
```python
from detectviz_adk.tools import RemoteTool

remote_tool = RemoteTool(
    name="http_request",
    version="1.0.0",
    bridge_address="127.0.0.1:6606",
    description="HTTP 請求工具"
)
```

#### Tool Registry
```python
from detectviz_adk.tools import ToolRegistry

# 註冊工具
registry = ToolRegistry.get_instance()
registry.register_tool("my_tool", my_tool_instance)

# 獲取工具
tool = registry.get_tool("my_tool")
tools = registry.get_tools(["tool1", "tool2", "tool3"])
```

### 回調系統

#### 標準回調
```python
from detectviz_adk.agents.callback_context import CallbackContext

def before_tool_callback(callback_context: CallbackContext):
    # 工具執行前：權限檢查、追蹤、驗證
    logger.info(f"Executing tool: {callback_context.tool_name}")
    
def after_tool_callback(callback_context: CallbackContext):
    # 工具執行後：結果處理、狀態更新
    logger.info(f"Tool completed: {callback_context.tool_name}")

def before_agent_callback(callback_context: CallbackContext):
    # Agent 執行前：上下文載入、會話準備
    session_id = callback_context.state.get("session_id")
    # 載入會話上下文...

def after_agent_callback(callback_context: CallbackContext):
    # Agent 執行後：狀態保存、記憶更新
    # 保存對話記憶...
```

### 共享機制

#### Agent Factory
```python
from detectviz_adk.agents import AgentFactory

# 註冊 Agent 模板
AgentFactory.register_template("data_analyst", DataAnalystAgent, {
    "model": "gemini-2.5-pro",
    "tools": ["data_api", "analysis_tool"]
})

# 創建獨立實例
agent_instance = AgentFactory.create_agent(
    template_name="data_analyst",
    instance_id="unique_id",
    custom_config={"custom_param": "value"}
)
```

#### Stateless Agent
```python
from detectviz_adk.agents import StatelessAgent

class MyStatelessAgent(StatelessAgent):
    async def execute(self, input_text: str, context: AgentContext) -> str:
        # 從 context 獲取所有狀態，不保存內部狀態
        session_history = context.get_conversation_history()
        result = await self._process(input_text, session_history)
        
        # 狀態更新通過 context 回寫
        context.add_to_conversation_history("assistant", result)
        return result

# 可安全共享的實例
shared_agent = MyStatelessAgent("shared_analyzer")
```

### 配置管理

#### 配置載入
```python
from detectviz_adk.config import ConfigLoader

# 載入配置（搜尋順序：參數 > 環境變數 > 檔案）
config = ConfigLoader.load()

# 存取配置
grpc_listen = config.grpc.listen
otlp_endpoint = config.observability.otlp.endpoint
```

#### 環境變數格式
```bash
# gRPC 配置
DETECTVIZ__GRPC__LISTEN=:6606
DETECTVIZ__GRPC__MAX_RECV_BYTES=4194304

# 可觀測性配置  
DETECTVIZ__OBSERVABILITY__OTLP__ENDPOINT=127.0.0.1:4317
DETECTVIZ__OBSERVABILITY__OTLP__INSECURE=true

# 工具橋接配置
DETECTVIZ_TOOLBRIDGE_ADDR=127.0.0.1:6606
DETECTVIZ_TOOLBRIDGE_INSECURE=true

# 記憶配置
DETECTVIZ__MEMORY__BACKEND=redis
DETECTVIZ__MEMORY__DSN=redis://localhost:6379/0
```

### 觀測性

#### 分散式追蹤
```python
from detectviz_adk.observability import setup_observability, get_current_trace_context

# 初始化追蹤
setup_observability(
    service_name="my-agent-service",
    otlp_endpoint=config.observability.otlp.endpoint
)

# 獲取追蹤上下文
trace_context = get_current_trace_context()
# 傳遞給 RemoteTool 進行跨語言追蹤
```

#### 結構化日誌
```python
import structlog

logger = structlog.get_logger("detectviz.agent")

logger.info(
    "Agent processing started",
    agent_name=agent.name,
    session_id=session_id,
    input_length=len(input_data)
)
```

#### 業務指標
```python
from prometheus_client import Counter, Histogram

# 定義指標
agent_requests = Counter(
    'detectviz_agent_requests_total',
    'Total agent requests',
    ['agent_name', 'status']
)

response_time = Histogram(
    'detectviz_agent_response_time_seconds', 
    'Agent response time',
    ['agent_name']
)

# 使用指標
with response_time.labels(agent_name=agent.name).time():
    result = await agent.run(input_text)
    agent_requests.labels(agent_name=agent.name, status='success').inc()
```

## 開發工作流

### 1. 新 Agent 開發

```bash
# 1. 選擇架構模式和範本
python scripts/generate_agent.py \
  --template coordinator_agent \
  --name my_new_agent \
  --output ./agents/my_new_agent

# 2. 實作核心邏輯
# - 編輯 agent.py
# - 定義 prompts.py  
# - 配置 module.card.json

# 3. 實作 Sub-Agent (如需要)
# - 創建 sub_agents/ 目錄結構
# - 實作各專家 Agent

# 4. 編寫測試
# - 單元測試：tests/unit/
# - 整合測試：tests/integration/

# 5. 驗證和部署
python scripts/validate_contracts.py
python scripts/run_tests.py --all
```

### 2. 新 Tool 開發

#### Python Tool
```bash
# 1. 創建 Tool 檔案
# tools/my_new_tool.py

# 2. 實作 Tool 類別
class MyNewTool(BaseTool):
    # 實作 execute 方法

# 3. 註冊 Tool
registry.register_tool("my_new_tool", MyNewTool())

# 4. 編寫測試
# tests/test_my_new_tool.py
```

#### Go Plugin  
```bash
# 1. 生成插件骨架
detectviz plugin new capability.gateway/my_plugin

# 2. 實作插件邏輯
# internal/pluginhost/plugins/capability.gateway/my_plugin/plugin.go

# 3. 註冊插件
# 在 registry.go 中註冊

# 4. 啟動 ToolBridge
detectviz plugin serve --config ./config.yaml
```

### 3. 測試策略

#### 單元測試
```python
@pytest.mark.asyncio
async def test_agent_basic_functionality():
    agent = create_my_agent("test_instance")
    result = await agent.run("test input")
    assert result is not None
    assert len(result) > 0

@pytest.mark.asyncio
async def test_tool_execution():
    tool = MyTool()
    result = await tool.execute("test data")
    assert result.status == "success"
```

#### 整合測試
```python
@pytest.mark.integration
async def test_agent_tool_integration():
    # 測試 Agent 與 Tool 的整合
    agent = create_agent_with_tools()
    result = await agent.run("需要調用工具的請求")
    # 驗證工具被正確調用
```

#### 性能測試
```python
@pytest.mark.performance 
async def test_concurrent_requests():
    # 測試並發性能
    tasks = [agent.run(f"request_{i}") for i in range(50)]
    results = await asyncio.gather(*tasks)
    # 驗證所有請求都成功處理
```

## 故障排除

### 常見問題

#### Agent 創建失敗
```bash
問題：Agent 初始化時拋出異常
排查：
1. 檢查模組卡格式：python scripts/validate_module_card.py
2. 檢查依賴版本：pip list | grep detectviz
3. 檢查配置檔案：python scripts/validate_config.py
```

#### Tool 調用失敗
```bash
問題：工具調用時返回錯誤
排查：
1. 檢查工具註冊：registry.list_tools()
2. 檢查 ToolBridge 連接：telnet 127.0.0.1 6606
3. 檢查工具權限：查看 audit 日誌
```

#### Sub-Agent 狀態污染
```bash
問題：多個 Agent 共享狀態導致錯誤
解決：
1. 使用工廠模式創建獨立實例
2. 檢查回調函數中的狀態管理
3. 使用 Stateless Agent 如果可能
```

#### 記憶管理問題
```bash
問題：Agent 無法記住之前的對話
排查：
1. 檢查記憶後端連接：redis-cli ping
2. 檢查會話 ID 設定：callback_context.state["session_id"]
3. 檢查記憶策略配置：memory.default_ttl_seconds
```

### 效能調優

#### Agent 響應時間優化
- 使用並發調用獨立的 Sub-Agent
- 優化提示詞長度和複雜度
- 使用快取減少重複計算
- 選擇適當的模型大小

#### 工具執行效能
- Go Plugin 用於 I/O 密集任務
- Python Tool 用於計算密集任務
- 實作工具結果快取
- 監控工具執行時間

#### 記憶體使用優化
- 定期清理過期記憶
- 使用記憶壓縮策略
- 監控記憶後端性能
- 適當設定 TTL 策略

## 部署檢查清單

### 開發環境
- [ ] 配置檔案正確設定
- [ ] 所有依賴套件已安裝
- [ ] ToolBridge 服務正常運行
- [ ] 記憶後端服務可用

### 測試環境
- [ ] 單元測試全部通過
- [ ] 整合測試全部通過
- [ ] 性能測試達標
- [ ] 契約驗證通過

### 生產環境
- [ ] 監控和告警配置完成
- [ ] 日誌記錄充分
- [ ] 安全配置檢查
- [ ] 備份和恢復計劃

## 最佳實務摘要

### 設計原則
1. **模組邊界清晰**：嚴格遵守 ADK 模組邊界
2. **職責單一**：每個 Agent/Tool 有明確職責
3. **狀態隔離**：避免 Agent 間狀態污染
4. **可觀測性**：充分的日誌、監控、追蹤

### 程式碼品質
1. **類型安全**：使用類型提示和驗證
2. **錯誤處理**：完整的異常處理機制
3. **測試覆蓋**：充分的單元和整合測試
4. **文檔完整**：清晰的註釋和 README

### 效能考量
1. **適當選擇**：Go vs Python 根據場景選擇
2. **資源管理**：適當的連接池和快取
3. **並發處理**：合理的並發策略
4. **監控調優**：持續的效能監控和優化

這份 Quick Reference 提供了開發過程中最常用的 API、模式和工作流程，讓開發者能夠快速上手並遵循最佳實務。