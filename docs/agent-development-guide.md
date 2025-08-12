# Agent 開發指南

## 概覽

本指南基於對 17 個 agents 範例的深度分析，提供完整的 Agent 開發最佳實務。涵蓋 4 種核心架構模式、Sub-Agent 共享機制，以及從設計到部署的完整工作流程。

## Agent 架構模式詳解

### 模式 1: Simple Agent Pattern (Tool-Driven)

#### 適用場景
- 功能性服務 (客服、RAG、專業查詢)
- 單一領域專精應用
- 工具豐富但邏輯相對簡單

#### 架構特點
```python
# 特點：root_agent 直接配置工具，無 sub_agent
root_agent = Agent(
    model="gemini-2.5-flash",
    tools=[tool1, tool2, tool3, ...],  # 豐富的工具集
    instruction=INSTRUCTION,
    callbacks={...}
)
```

#### 實作範例 (ecommerce_assistant)
**檔案結構：**
```bash
agents/ecommerce_assistant/
├── agent.py                    # 主要 Agent 實作
├── prompts.py                  # 提示詞管理  
├── config.py                   # 配置管理 (可選)
├── tools/                      # 專屬工具
│   ├── order_lookup_tool.py
│   └── product_search_tool.py
├── shared/                     # 共享邏輯 (對應 shared_libraries)
│   ├── callbacks.py            # 生命週期回調
│   └── constants.py            # 常數定義
├── module.card.json            # 模組卡 (role: agent.tool_exec)
└── tests/                      # 測試套件
```

**實作程式碼：**
```python
# agents/ecommerce_assistant/agent.py
from detectviz_adk.agents import Agent
from detectviz_adk.tools import ToolRegistry
from .prompts import GLOBAL_INSTRUCTION, MAIN_INSTRUCTION
from .tools.order_lookup_tool import order_lookup_tool
from .shared.callbacks import before_tool_callback, after_tool_callback

# 註冊工具 (支援跨 Agent 共享)
registry = ToolRegistry.get_instance()
registry.register_tool("order_lookup", order_lookup_tool)

root_agent = Agent(
    model="gemini-2.5-flash",
    name="ecommerce_assistant",
    global_instruction=GLOBAL_INSTRUCTION,
    instruction=MAIN_INSTRUCTION,
    tools=[
        order_lookup_tool,           # 本地工具
        registry.get_tool("http_client"),  # 共享工具
        # RemoteTool 連接 Go ToolBridge
    ],
    before_tool_callback=before_tool_callback,
    after_tool_callback=after_tool_callback,
)
```

### 模式 2: Coordinator Pattern (Multi-Agent)

#### 適用場景
- 複雜業務流程 (金融分析、學術研究)
- 需要多個專家 Agent 協作
- 每個子領域有獨立工具和知識

#### 架構特點
```python
# 特點：root_agent 透過 AgentTool 協調多個專家 sub_agent
coordinator = LlmAgent(
    tools=[
        AgentTool(agent=expert_agent_1),
        AgentTool(agent=expert_agent_2),
        AgentTool(agent=expert_agent_3),
    ]
)
```

#### 實作範例 (financial_advisor)
**檔案結構：**
```bash
agents/financial_advisor/
├── agent.py                    # 協調器 Agent
├── prompts.py                  # 協調邏輯提示詞
├── sub_agents/                 # 專家 Agent 組織
│   ├── data_analyst/
│   │   ├── agent.py
│   │   ├── prompts.py
│   │   └── tools/
│   ├── risk_analyst/
│   │   ├── agent.py
│   │   └── prompts.py
│   └── trading_analyst/
│       ├── agent.py
│       └── prompts.py
├── module.card.json            # role: agent.coordinator
└── tests/
```

**實作程式碼：**
```python
# agents/financial_advisor/agent.py
from detectviz_adk.agents import LlmAgent
from detectviz_adk.tools import AgentTool
from .sub_agents.data_analyst.agent import data_analyst_agent
from .sub_agents.risk_analyst.agent import risk_analyst_agent

financial_coordinator = LlmAgent(
    name="financial_coordinator",
    model="gemini-2.5-pro",
    description="多專家金融分析協調器",
    instruction="""
    你是金融顧問協調器，負責：
    1. 分析用戶需求
    2. 分派任務給專家 Agent
    3. 整合專家意見
    4. 提供綜合建議
    """,
    tools=[
        AgentTool(agent=data_analyst_agent),    # 數據分析專家
        AgentTool(agent=risk_analyst_agent),    # 風險分析專家
        AgentTool(agent=trading_analyst_agent), # 交易分析專家
    ],
)

root_agent = financial_coordinator
```

### 模式 3: Hierarchy Pattern (Sub-Agent Chain)

#### 適用場景
- 階段性任務處理 (旅遊規劃、專案管理)
- 有明確的業務流程步驟
- 每個階段需要不同的專業知識

#### 架構特點
```python
# 特點：root_agent 使用 sub_agents 屬性進行階層管理
root_agent = Agent(
    sub_agents=[
        phase1_agent,  # 階段 1
        phase2_agent,  # 階段 2
        phase3_agent,  # 階段 3
    ],
    before_agent_callback=load_context_callback
)
```

#### 實作範例 (travel_concierge)
**檔案結構：**
```bash
agents/travel_concierge/
├── agent.py                    # 階層協調 Agent
├── prompts.py                  # 協調邏輯
├── sub_agents/                 # 階段性 Agent
│   ├── inspiration/
│   │   ├── agent.py            # 靈感生成
│   │   └── prompts.py
│   ├── planning/
│   │   ├── agent.py            # 規劃制定
│   │   └── prompts.py
│   ├── booking/
│   │   ├── agent.py            # 預訂執行
│   │   └── tools.py
│   ├── pre_trip/
│   │   ├── agent.py            # 行前準備
│   │   └── prompts.py
│   └── in_trip/
│       ├── agent.py            # 行程中支援
│       └── tools.py
├── shared/
│   ├── constants.py            # 旅遊相關常數
│   └── types.py                # 資料結構定義
├── tools/
│   ├── memory.py               # 行程記憶管理
│   └── places.py               # 地點資訊工具
└── profiles/                   # 預設行程模板
    └── itinerary_templates.json
```

**實作程式碼：**
```python
# agents/travel_concierge/agent.py
from detectviz_adk.agents import Agent
from .sub_agents.inspiration.agent import inspiration_agent
from .sub_agents.planning.agent import planning_agent
from .tools.memory import load_precreated_itinerary

root_agent = Agent(
    model="gemini-2.5-flash",
    name="travel_concierge",
    description="旅遊禮賓服務，提供全程旅遊支援",
    instruction="""
    你是專業的旅遊禮賓服務，會根據旅遊階段調用相應的專家：
    - 靈感階段：inspiration_agent
    - 規劃階段：planning_agent  
    - 預訂階段：booking_agent
    - 行前準備：pre_trip_agent
    - 行程中：in_trip_agent
    - 行後總結：post_trip_agent
    """,
    sub_agents=[
        inspiration_agent,   # 靈感生成
        planning_agent,      # 規劃制定
        booking_agent,       # 預訂執行
        pre_trip_agent,      # 行前準備
        in_trip_agent,       # 行程中支援
        post_trip_agent,     # 行後總結
    ],
    before_agent_callback=load_precreated_itinerary,
)
```

### 模式 4: Workflow Pattern (Sequential/Loop)

#### 適用場景
- 複雜工作流 (內容生成、品質評估)
- 需要迭代優化的任務
- 有明確的處理步驟和條件判斷

#### 架構特點
```python
# 特點：使用 SequentialAgent, LoopAgent 進行工作流編排
workflow = SequentialAgent(sub_agents=[step1, step2, step3])
root_agent = LoopAgent(
    sub_agents=[workflow, checker_agent],
    termination_condition=quality_check
)
```

#### 實作範例 (image_scoring)
**檔案結構：**
```bash
agents/image_scoring/
├── agent.py                    # 工作流編排 Agent
├── sub_agents/
│   ├── prompt/
│   │   ├── prompt_agent.py     # 提示詞生成
│   │   └── prompts.py
│   ├── image/
│   │   ├── image_agent.py      # 圖像生成
│   │   └── tools/
│   │       └── image_generation_tool.py
│   ├── scoring/
│   │   ├── scoring_agent.py    # 品質評分
│   │   └── tools/
│   │       ├── get_images_tool.py
│   │       └── set_score_tool.py
│   └── tools/
│       └── fetch_policy_tool.py
├── checker_agent.py            # 終止條件檢查
├── tools/
│   └── loop_condition_tool.py  # 循環控制工具
└── config.py                   # 品質閾值配置
```

**實作程式碼：**
```python
# agents/image_scoring/agent.py
from detectviz_adk.agents import SequentialAgent, LoopAgent
from detectviz_adk.agents.callback_context import CallbackContext
from .sub_agents.prompt.prompt_agent import image_generation_prompt_agent
from .sub_agents.image.image_agent import image_generation_agent
from .sub_agents.scoring.scoring_agent import scoring_images_agent
from .checker_agent import checker_agent_instance

def set_session(callback_context: CallbackContext):
    """設置會話 ID 和時間戳"""
    callback_context.state["unique_id"] = str(uuid.uuid4())
    callback_context.state["timestamp"] = datetime.now().isoformat()

# 順序工作流：提示詞 → 圖像生成 → 評分
image_generation_workflow = SequentialAgent(
    name="image_generation_scoring_agent",
    description="圖像生成與評分的順序工作流",
    sub_agents=[
        image_generation_prompt_agent,  # 步驟 1: 生成提示詞
        image_generation_agent,         # 步驟 2: 生成圖像
        scoring_images_agent,           # 步驟 3: 評分
    ]
)

# 循環執行直到品質達標
root_agent = LoopAgent(
    name="image_scoring",
    description="迭代執行圖像生成直到品質達標",
    sub_agents=[
        image_generation_workflow,  # 主要工作流
        checker_agent_instance,     # 品質檢查與終止條件
    ],
    before_agent_callback=set_session,
)
```

## Sub-Agent 共享機制與最佳實務

### 工廠模式 (推薦 - 獨立實例)

#### 適用場景
- 有狀態的 Agent (對話記憶、上下文)
- 需要避免狀態污染
- 大部分業務場景

#### 實作方式
```python
# src/detectviz_adk/agents/factory.py
class AgentFactory:
    _templates = {}
    
    @classmethod
    def register_template(cls, name: str, agent_class, default_config: Dict):
        cls._templates[name] = {
            "class": agent_class,
            "default_config": default_config
        }
    
    @classmethod  
    def create_agent(cls, template_name: str, instance_id: str, custom_config: Dict = None):
        template = cls._templates[template_name]
        config = template["default_config"].copy()
        if custom_config:
            config.update(custom_config)
        
        return template["class"](
            name=f"{template_name}_{instance_id}",
            config=config,
            instance_id=instance_id
        )

# 使用範例
AgentFactory.register_template("data_analyst", DataAnalystAgent, {...})

# 每個 coordinator 創建獨立實例
financial_coordinator = Agent(
    tools=[AgentTool(agent=AgentFactory.create_agent("data_analyst", "financial"))]
)
market_coordinator = Agent(
    tools=[AgentTool(agent=AgentFactory.create_agent("data_analyst", "market"))]
)
```

### 無狀態共享模式 (高級場景)

#### 適用場景
- 純函數處理的 Agent
- 數據查詢、分析類任務
- 高併發需求

#### 實作方式
```python
# src/detectviz_adk/agents/stateless_agent.py
class StatelessAgent(BaseAgent, ABC):
    def __init__(self, name: str, **kwargs):
        super().__init__(name=name, **kwargs)
        self._is_stateless = True
        self._shared_instance = True
    
    async def execute(self, input_text: str, context: AgentContext) -> str:
        # 所有狀態從 context 獲取，不保存內部狀態
        session_id = context.session_id
        conversation_history = context.get_conversation_history()
        
        result = await self._process(input_text, conversation_history)
        
        # 狀態更新通過 context 回寫
        context.add_to_conversation_history("assistant", result)
        return result

# 創建可共享實例
shared_market_analyst = StatelessMarketAnalyst("market_analyst")

# 多個 Agent 安全共享
financial_agent = Agent(tools=[AgentTool(agent=shared_market_analyst)])
investment_agent = Agent(tools=[AgentTool(agent=shared_market_analyst)])
```

### Agent Pool 模式 (高併發場景)

#### 適用場景
- 有狀態但需要併發支援
- 資源密集的 Agent
- 需要限制並發數量

#### 實作方式
```python
# src/detectviz_adk/agents/agent_pool.py
class AgentPool:
    def __init__(self, template_name: str, pool_size: int = 5):
        self.template_name = template_name
        self.available_agents = asyncio.Queue()
        self._init_pool(pool_size)
    
    async def acquire(self) -> BaseAgent:
        return await self.available_agents.get()
    
    async def release(self, agent: BaseAgent):
        await agent.reset_state()  # 清理狀態
        await self.available_agents.put(agent)
    
    async def execute_with_agent(self, input_text: str, context: dict) -> str:
        agent = await self.acquire()
        try:
            result = await agent.execute(input_text, context)
            return result
        finally:
            await self.release(agent)

# 使用 Agent Pool
data_analyst_pool = AgentPool("data_analyst", pool_size=3)

class PooledAgentTool:
    def __init__(self, agent_pool: AgentPool):
        self.agent_pool = agent_pool
    
    async def execute(self, input_text: str, **kwargs) -> str:
        return await self.agent_pool.execute_with_agent(input_text, kwargs)

# 多個 Agent 使用同一個池
pooled_tool = PooledAgentTool(data_analyst_pool)
agent_a = Agent(tools=[pooled_tool])
agent_b = Agent(tools=[pooled_tool])
```

## 完整開發工作流程

### 步驟 1: 需求分析與架構選型

#### 1.1 確定 Agent 類型
```bash
# 決策矩陣
if 單一功能 + 工具豐富:
    選擇 Simple Agent Pattern
elif 多專家協作:
    選擇 Coordinator Pattern  
elif 階段性流程:
    選擇 Hierarchy Pattern
elif 迭代優化流程:
    選擇 Workflow Pattern
```

#### 1.2 分析共享需求
```bash
# Sub-Agent 共享決策
if 無狀態 + 純函數:
    選擇無狀態共享模式
elif 高併發需求:
    選擇 Agent Pool 模式
else:
    選擇工廠模式 (獨立實例)
```

### 步驟 2: 生成專案骨架

```bash
# 使用範本生成
python scripts/generate_agent.py \
  --template coordinator_agent \
  --name financial_advisor \
  --output ./agents/financial_advisor \
  --sub-agents "data_analyst,risk_analyst,trading_analyst"

# 自動生成完整目錄結構
agents/financial_advisor/
├── agent.py                    # 主 Agent
├── prompts.py                  # 提示詞
├── sub_agents/                 # 子 Agent 骨架
├── shared/                     # 共享邏輯
├── module.card.json            # 模組卡
└── tests/                      # 測試框架
```

### 步驟 3: 實作核心邏輯

#### 3.1 定義提示詞策略
```python
# agents/financial_advisor/prompts.py
COORDINATOR_PROMPT = """
你是金融顧問協調器，負責：

1. 理解用戶的投資需求和風險偏好
2. 決定需要調用哪些專家 Agent
3. 整合專家意見形成綜合建議

可用專家：
- data_analyst: 數據分析與市場趨勢
- risk_analyst: 風險評估與管理  
- trading_analyst: 交易策略制定

工作流程：
1. 分析用戶查詢
2. 確定需要的專家類型
3. 調用相應的專家 Agent
4. 整合結果並給出建議
"""
```

#### 3.2 實作 Agent 邏輯
```python
# agents/financial_advisor/agent.py
from detectviz_adk.agents import LlmAgent, AgentFactory
from detectviz_adk.tools import AgentTool

# 註冊專家 Agent 模板
AgentFactory.register_template("data_analyst", DataAnalystAgent, {
    "model": "gemini-2.5-pro",
    "tools": ["market_data", "statistical_analysis"]
})

financial_coordinator = LlmAgent(
    name="financial_coordinator", 
    model="gemini-2.5-pro",
    instruction=COORDINATOR_PROMPT,
    tools=[
        AgentTool(agent=AgentFactory.create_agent("data_analyst", "financial")),
        AgentTool(agent=AgentFactory.create_agent("risk_analyst", "financial")),
        AgentTool(agent=AgentFactory.create_agent("trading_analyst", "financial")),
    ]
)

root_agent = financial_coordinator
```

#### 3.3 配置模組卡
```json
// agents/financial_advisor/module.card.json
{
  "name": "financial_advisor",
  "version": "1.0.0",
  "role": "agent.coordinator",
  "category": "financial_services",
  "sub_agents": [
    {
      "name": "data_analyst",
      "role": "agent.specialist",
      "required": true
    },
    {
      "name": "risk_analyst", 
      "role": "agent.specialist",
      "required": true
    }
  ],
  "observability": {
    "tags": ["finance", "multi_agent", "advisory"]
  }
}
```

### 步驟 4: 測試與驗證

#### 4.1 單元測試
```python
# agents/financial_advisor/tests/test_agent.py
@pytest.mark.asyncio
async def test_financial_advisor_integration():
    input_text = "我想投資科技股，但風險偏好保守，請給我建議"
    
    result = await financial_coordinator.run(input_text)
    
    # 驗證協調器有調用專家 Agent
    assert "數據分析" in result or "市場趨勢" in result
    assert "風險評估" in result or "風險管理" in result
    assert "建議" in result

@pytest.mark.asyncio 
async def test_sub_agent_independence():
    # 測試多個 coordinator 使用獨立的 sub_agent 實例
    coordinator_1 = create_financial_coordinator("instance_1")
    coordinator_2 = create_financial_coordinator("instance_2")
    
    # 並發執行，不應該有狀態衝突
    results = await asyncio.gather(
        coordinator_1.run("分析 AAPL 股票"),
        coordinator_2.run("評估 TSLA 風險")
    )
    
    assert "AAPL" in results[0]
    assert "TSLA" in results[1]
```

#### 4.2 整合測試
```python
@pytest.mark.integration
async def test_agent_registry_integration():
    from detectviz_adk.agents.registry import AgentRegistry
    
    registry = AgentRegistry()
    registry.register_agent("financial_advisor", financial_coordinator)
    
    agent = registry.get_agent("financial_advisor")
    result = await agent.run("幫我分析當前市場狀況")
    
    assert result is not None
    assert len(result) > 50  # 應該有詳細回應
```

### 步驟 5: 部署與監控

#### 5.1 註冊到 Runtime
```python
# src/detectviz_adk/agents/registry.py
from agents.financial_advisor.agent import financial_coordinator

class AgentRegistry:
    def _register_builtin_agents(self):
        self.register_agent("financial_advisor", financial_coordinator)
```

#### 5.2 配置監控
```yaml
# monitoring/financial-advisor-dashboard.json
{
  "panels": [
    {
      "title": "Financial Advisor Response Time",
      "targets": [
        {
          "expr": "histogram_quantile(0.95, detectviz_agent_response_time{agent_name=\"financial_advisor\"})"
        }
      ]
    },
    {
      "title": "Sub-Agent Usage Distribution", 
      "targets": [
        {
          "expr": "rate(detectviz_sub_agent_calls_total{parent_agent=\"financial_advisor\"}[5m])"
        }
      ]
    }
  ]
}
```

## 開發最佳實務

### 1. **狀態管理**
- 使用工廠模式避免 Sub-Agent 狀態污染
- 通過 CallbackContext 管理會話狀態
- 無狀態 Agent 優先考慮共享

### 2. **錯誤處理**
- 每個 Agent 都要有完整的異常處理
- Sub-Agent 失敗不應影響整個系統
- 提供降級策略和備用方案

### 3. **性能優化**
- 並發執行獨立的 Sub-Agent 調用
- 使用 Agent Pool 管理資源密集型 Agent
- 適當的快取策略減少重複計算

### 4. **可觀測性**
- 每個 Agent 調用都要有追蹤
- 記錄 Sub-Agent 的調用路徑
- 監控各個模式的性能差異

### 5. **模組化設計**
- 清晰的模組邊界和介面
- 可插拔的 Sub-Agent 設計
- 標準化的模組卡管理

這套 Agent 開發指南確保了：
- **模式清晰**: 4 種架構模式涵蓋所有業務場景
- **共享安全**: 多種共享策略避免狀態衝突
- **流程標準**: 完整的開發到部署工作流
- **可監控**: 內建的可觀測性和性能監控
- **穩定可靠**: 完善的測試和錯誤處理機制