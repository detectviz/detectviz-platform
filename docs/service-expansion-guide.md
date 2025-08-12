# Service 擴充指南

## 概覽

本指南提供完整的 Service (Agent) 擴充流程，從需求分析到部署上線的詳細步驟。基於對 17 個 agents 範例的分析，確保新 Service 遵守 ADK 模組邊界和 SSOT 契約規範。

## Service 擴充完整流程

### 步驟 1: 需求分析與模式選型

#### 1.1 業務需求分析

**功能複雜度評估**：
```bash
# 單一功能 + 工具豐富 → Simple Agent Pattern
if 功能領域單一 AND 工具需求多樣:
    pattern = "Simple Agent Pattern"
    template = "simple_agent"
    
# 多專家協作 → Coordinator Pattern  
elif 需要多個領域專家 AND 協調複雜:
    pattern = "Coordinator Pattern"
    template = "coordinator_agent"
    
# 階段性流程 → Hierarchy Pattern
elif 有明確業務階段 AND 順序依賴:
    pattern = "Hierarchy Pattern"
    template = "hierarchy_agent"
    
# 迭代優化流程 → Workflow Pattern
elif 需要循環優化 OR 複雜工作流:
    pattern = "Workflow Pattern"
    template = "workflow_agent"
```

**共享需求分析**：
```bash
# Sub-Agent 共享策略選擇
if Agent 無狀態 AND 純函數處理:
    sharing_strategy = "stateless_sharing"
elif 高併發需求 AND 有狀態:
    sharing_strategy = "agent_pool" 
else:
    sharing_strategy = "factory_pattern"  # 預設推薦
```

#### 1.2 技術架構決策

**範例需求**：實作一個「智能投資顧問」Service
- **業務場景**：用戶提供投資需求，系統分析市場並提供投資建議
- **專家需求**：數據分析師、風險分析師、交易分析師
- **模式選擇**：Coordinator Pattern（多專家協作）

### 步驟 2: 生成專案骨架

#### 2.1 創建目錄結構

```bash
# 在 agents/ 目錄下創建新 Service
mkdir -p agents/investment_advisor
cd agents/investment_advisor

# 創建完整目錄結構
mkdir -p {sub_agents,tools,shared,tests,profiles}
mkdir -p sub_agents/{data_analyst,risk_analyst,trading_analyst}
mkdir -p shared/{callbacks,constants,types,middleware}
mkdir -p tests/{unit,integration}
```

#### 2.2 生成核心檔案

**主 Agent 檔案**：
```bash
# agents/investment_advisor/agent.py
touch agent.py
```

**提示詞管理**：
```bash
# agents/investment_advisor/prompts.py
touch prompts.py
```

**模組卡**：
```bash
# agents/investment_advisor/module.card.json
touch module.card.json
```

### 步驟 3: 實作核心邏輯

#### 3.1 定義模組卡

```json
// agents/investment_advisor/module.card.json
{
  "name": "investment_advisor",
  "version": "1.0.0",
  "role": "agent.coordinator",
  "category": "financial_services",
  "description": "Multi-agent investment advisory coordinator",
  "entrypoint": "agent.py",
  "language": "python",
  "sub_agents": [
    {
      "name": "data_analyst",
      "role": "agent.specialist",
      "required": true,
      "description": "Market data analysis specialist"
    },
    {
      "name": "risk_analyst", 
      "role": "agent.specialist",
      "required": true,
      "description": "Investment risk assessment specialist"
    },
    {
      "name": "trading_analyst",
      "role": "agent.specialist", 
      "required": true,
      "description": "Trading strategy specialist"
    }
  ],
  "requires": [
    {
      "name": "detectviz_adk",
      "version": ">=1.0.0"
    }
  ],
  "resources": {
    "memory_mb": 512,
    "cpu_cores": 2
  },
  "observability": {
    "tags": ["finance", "multi_agent", "advisory", "coordinator"]
  }
}
```

#### 3.2 實作提示詞策略

```python
# agents/investment_advisor/prompts.py

GLOBAL_INSTRUCTION = """
你是專業的投資顧問協調器，負責整合多個金融專家的意見，為用戶提供綜合性投資建議。

核心職責：
1. 理解用戶投資需求和風險偏好
2. 決定需要調用哪些專家 Agent
3. 整合專家意見形成可執行的投資建議
4. 確保建議符合用戶風險承受能力

工作原則：
- 基於數據做決策，避免主觀臆測
- 風險管理優先，收益預期其次
- 提供具體可執行的投資建議
- 保持建議的時效性和相關性
"""

COORDINATOR_PROMPT = """
作為投資顧問協調器，你可以調用以下專家：

1. **data_analyst**：市場數據分析與趨勢預測
   - 市場趨勢分析
   - 行業比較分析  
   - 技術指標解讀
   - 基本面數據分析

2. **risk_analyst**：投資風險評估與管理
   - 投資組合風險評估
   - 風險控制策略
   - 壓力測試分析
   - 風險預算配置

3. **trading_analyst**：交易策略制定與執行
   - 交易時機分析
   - 資產配置建議
   - 投資策略制定
   - 執行方案設計

工作流程：
1. 分析用戶投資查詢，識別關鍵需求
2. 根據查詢性質，決定需要調用的專家類型
3. 向相應專家發送具體分析請求
4. 整合專家分析結果
5. 生成綜合投資建議

輸出要求：
- 提供具體投資建議和理由
- 包含風險提示和應對措施
- 建議具體執行步驟和時間點
- 設定合理的預期收益和風險範圍
"""
```

#### 3.3 實作主 Agent

```python
# agents/investment_advisor/agent.py
from detectviz_adk.agents import LlmAgent
from detectviz_adk.tools import AgentTool
from detectviz_adk.agents.factory import AgentFactory

from .prompts import GLOBAL_INSTRUCTION, COORDINATOR_PROMPT
from .sub_agents.data_analyst.agent import create_data_analyst
from .sub_agents.risk_analyst.agent import create_risk_analyst  
from .sub_agents.trading_analyst.agent import create_trading_analyst
from .shared.callbacks import before_agent_callback, after_agent_callback

def create_investment_advisor(instance_id: str = "default"):
    """創建投資顧問協調器 Agent
    
    Args:
        instance_id: 實例 ID，用於區分不同協調器實例
        
    Returns:
        投資顧問協調器 Agent 實例
    """
    
    # 創建專家 Agent 實例（獨立實例避免狀態污染）
    data_analyst = create_data_analyst(f"investment_{instance_id}")
    risk_analyst = create_risk_analyst(f"investment_{instance_id}")
    trading_analyst = create_trading_analyst(f"investment_{instance_id}")
    
    # 創建協調器 Agent
    investment_advisor = LlmAgent(
        name=f"investment_advisor_{instance_id}",
        model="gemini-2.5-pro",
        global_instruction=GLOBAL_INSTRUCTION,
        instruction=COORDINATOR_PROMPT,
        description="Multi-agent investment advisory coordinator",
        tools=[
            AgentTool(agent=data_analyst),     # 數據分析專家
            AgentTool(agent=risk_analyst),     # 風險分析專家
            AgentTool(agent=trading_analyst),  # 交易分析專家
        ],
        before_agent_callback=before_agent_callback,
        after_agent_callback=after_agent_callback,
    )
    
    return investment_advisor

# 預設實例（向後相容）
root_agent = create_investment_advisor("default")
```

### 步驟 4: 實作 Sub-Agent

#### 4.1 數據分析師 Agent

```python
# agents/investment_advisor/sub_agents/data_analyst/agent.py
from detectviz_adk.agents import Agent
from detectviz_adk.tools import ToolRegistry

from .prompts import DATA_ANALYST_PROMPT

def create_data_analyst(instance_id: str):
    """創建數據分析師 Agent"""
    
    # 獲取共享工具
    registry = ToolRegistry.get_instance()
    
    data_analyst = Agent(
        name=f"data_analyst_{instance_id}",
        model="gemini-2.5-flash",
        instruction=DATA_ANALYST_PROMPT,
        description="Market data analysis and trend prediction specialist",
        tools=[
            registry.get_tool("market_data_api"),    # 市場數據 API
            registry.get_tool("technical_analysis"), # 技術分析工具
            registry.get_tool("fundamental_data"),   # 基本面數據
        ]
    )
    
    return data_analyst
```

```python
# agents/investment_advisor/sub_agents/data_analyst/prompts.py
DATA_ANALYST_PROMPT = """
你是專業的市場數據分析師，專精於：

核心能力：
1. 市場趨勢分析：識別短期、中期、長期趨勢
2. 行業比較分析：橫向比較不同行業表現
3. 技術指標分析：使用 RSI、MACD、布林帶等指標
4. 基本面分析：財務數據、估值指標分析

分析方法：
- 結合量價分析確定趨勢強度
- 使用多時間週期確認信號
- 關注成交量變化和資金流向
- 考慮宏觀經濟因素影響

輸出要求：
- 提供具體數據支撐的分析結論
- 標明分析的時間範圍和可信度
- 包含關鍵風險點和注意事項
- 給出明確的投資時機建議
"""
```

#### 4.2 風險分析師 Agent

```python
# agents/investment_advisor/sub_agents/risk_analyst/agent.py  
from detectviz_adk.agents import Agent
from detectviz_adk.tools import ToolRegistry

from .prompts import RISK_ANALYST_PROMPT

def create_risk_analyst(instance_id: str):
    """創建風險分析師 Agent"""
    
    registry = ToolRegistry.get_instance()
    
    risk_analyst = Agent(
        name=f"risk_analyst_{instance_id}",
        model="gemini-2.5-flash", 
        instruction=RISK_ANALYST_PROMPT,
        description="Investment risk assessment and management specialist",
        tools=[
            registry.get_tool("risk_calculator"),    # 風險計算工具
            registry.get_tool("portfolio_analyzer"), # 投資組合分析
            registry.get_tool("stress_test"),        # 壓力測試
        ]
    )
    
    return risk_analyst
```

#### 4.3 交易分析師 Agent

```python
# agents/investment_advisor/sub_agents/trading_analyst/agent.py
from detectviz_adk.agents import Agent
from detectviz_adk.tools import ToolRegistry

from .prompts import TRADING_ANALYST_PROMPT

def create_trading_analyst(instance_id: str):
    """創建交易分析師 Agent"""
    
    registry = ToolRegistry.get_instance()
    
    trading_analyst = Agent(
        name=f"trading_analyst_{instance_id}",
        model="gemini-2.5-flash",
        instruction=TRADING_ANALYST_PROMPT, 
        description="Trading strategy and execution specialist",
        tools=[
            registry.get_tool("strategy_backtester"), # 策略回測
            registry.get_tool("position_calculator"),  # 倉位計算
            registry.get_tool("execution_optimizer"),  # 執行優化
        ]
    )
    
    return trading_analyst
```

### 步驟 5: 實作共享邏輯

#### 5.1 生命週期回調

```python
# agents/investment_advisor/shared/callbacks.py
from detectviz_adk.agents.callback_context import CallbackContext
import logging

logger = logging.getLogger(__name__)

def before_agent_callback(callback_context: CallbackContext):
    """Agent 執行前回調"""
    
    # 設置追蹤資訊
    session_id = callback_context.state.get("session_id")
    agent_name = callback_context.agent_name
    
    logger.info(
        f"Starting agent execution",
        extra={
            "agent_name": agent_name,
            "session_id": session_id,
            "input_length": len(callback_context.input_text)
        }
    )
    
    # 載入投資偏好上下文
    user_profile = callback_context.state.get("user_profile", {})
    risk_tolerance = user_profile.get("risk_tolerance", "moderate")
    investment_horizon = user_profile.get("investment_horizon", "medium_term")
    
    # 注入上下文到 Agent
    callback_context.context.update({
        "risk_tolerance": risk_tolerance,
        "investment_horizon": investment_horizon,
        "session_id": session_id
    })

def after_agent_callback(callback_context: CallbackContext):
    """Agent 執行後回調"""
    
    agent_name = callback_context.agent_name
    session_id = callback_context.state.get("session_id")
    
    logger.info(
        f"Agent execution completed",
        extra={
            "agent_name": agent_name,
            "session_id": session_id,
            "output_length": len(callback_context.output_text),
            "execution_time_ms": callback_context.execution_time_ms
        }
    )
    
    # 記錄專家建議到會話記憶
    if agent_name.startswith("data_analyst"):
        callback_context.state["latest_market_analysis"] = callback_context.output_text
    elif agent_name.startswith("risk_analyst"):
        callback_context.state["latest_risk_assessment"] = callback_context.output_text
    elif agent_name.startswith("trading_analyst"):
        callback_context.state["latest_trading_strategy"] = callback_context.output_text
```

#### 5.2 常數定義

```python
# agents/investment_advisor/shared/constants.py

# 風險等級定義
RISK_LEVELS = {
    "conservative": {"max_volatility": 0.1, "max_drawdown": 0.05},
    "moderate": {"max_volatility": 0.15, "max_drawdown": 0.1}, 
    "aggressive": {"max_volatility": 0.25, "max_drawdown": 0.2}
}

# 投資期限定義
INVESTMENT_HORIZONS = {
    "short_term": {"days": 90, "rebalance_frequency": "weekly"},
    "medium_term": {"days": 365, "rebalance_frequency": "monthly"},
    "long_term": {"days": 1825, "rebalance_frequency": "quarterly"}
}

# 資產類別權重
ASSET_ALLOCATION_TEMPLATES = {
    "conservative": {"stocks": 0.3, "bonds": 0.6, "cash": 0.1},
    "moderate": {"stocks": 0.6, "bonds": 0.3, "cash": 0.1},
    "aggressive": {"stocks": 0.8, "bonds": 0.15, "cash": 0.05}
}

# 效能指標
PERFORMANCE_METRICS = [
    "annual_return",
    "sharpe_ratio", 
    "max_drawdown",
    "volatility",
    "win_rate"
]
```

### 步驟 6: 編寫測試

#### 6.1 單元測試

```python
# agents/investment_advisor/tests/unit/test_agent.py
import pytest
import asyncio
from unittest.mock import AsyncMock, patch

from agents.investment_advisor.agent import create_investment_advisor

@pytest.mark.asyncio
async def test_investment_advisor_creation():
    """測試投資顧問 Agent 創建"""
    advisor = create_investment_advisor("test_instance")
    
    assert advisor.name == "investment_advisor_test_instance"
    assert advisor.model == "gemini-2.5-pro"
    assert len(advisor.tools) == 3  # 三個專家 Agent

@pytest.mark.asyncio
async def test_investment_advisor_basic_query():
    """測試基本投資查詢"""
    advisor = create_investment_advisor("test")
    
    input_text = "我想投資科技股，但風險偏好保守，請給我建議"
    
    # Mock 專家 Agent 回應
    with patch.object(advisor, 'run') as mock_run:
        mock_run.return_value = "基於風險評估，建議配置 30% 大型科技股..."
        
        result = await advisor.run(input_text)
        
        assert "建議" in result
        assert len(result) > 50

@pytest.mark.asyncio 
async def test_sub_agent_independence():
    """測試 Sub-Agent 獨立性"""
    advisor_1 = create_investment_advisor("instance_1")
    advisor_2 = create_investment_advisor("instance_2")
    
    # 確認創建了不同的 sub_agent 實例
    agent_1_tools = advisor_1.tools
    agent_2_tools = advisor_2.tools
    
    # 每個協調器應該有獨立的專家實例
    for tool_1, tool_2 in zip(agent_1_tools, agent_2_tools):
        assert tool_1.agent != tool_2.agent
        assert tool_1.agent.name != tool_2.agent.name

@pytest.mark.asyncio
async def test_concurrent_execution():
    """測試並發執行"""
    advisor_1 = create_investment_advisor("concurrent_1")
    advisor_2 = create_investment_advisor("concurrent_2")
    
    # 模擬並發查詢
    queries = [
        "分析 AAPL 股票投資價值",
        "評估 TSLA 投資風險"
    ]
    
    with patch.object(advisor_1, 'run') as mock_1, \
         patch.object(advisor_2, 'run') as mock_2:
        
        mock_1.return_value = "AAPL 分析結果..."
        mock_2.return_value = "TSLA 風險評估..."
        
        results = await asyncio.gather(
            advisor_1.run(queries[0]),
            advisor_2.run(queries[1])
        )
        
        assert "AAPL" in results[0] 
        assert "TSLA" in results[1]
```

#### 6.2 整合測試

```python
# agents/investment_advisor/tests/integration/test_agent_integration.py
import pytest
from detectviz_adk.agents.registry import AgentRegistry

from agents.investment_advisor.agent import create_investment_advisor

@pytest.mark.integration
async def test_agent_registry_integration():
    """測試與 Agent Registry 整合"""
    registry = AgentRegistry()
    
    advisor = create_investment_advisor("registry_test")
    registry.register_agent("investment_advisor", advisor)
    
    # 測試從 registry 獲取 Agent
    retrieved_agent = registry.get_agent("investment_advisor")
    assert retrieved_agent.name == advisor.name
    
    # 測試基本功能
    result = await retrieved_agent.run("幫我分析當前市場狀況")
    assert result is not None
    assert len(result) > 20

@pytest.mark.integration  
async def test_tool_registry_integration():
    """測試與 Tool Registry 整合"""
    from detectviz_adk.tools import ToolRegistry
    
    registry = ToolRegistry.get_instance()
    
    # 驗證必要工具已註冊
    required_tools = [
        "market_data_api",
        "technical_analysis", 
        "risk_calculator",
        "strategy_backtester"
    ]
    
    for tool_name in required_tools:
        tool = registry.get_tool(tool_name)
        assert tool is not None, f"Tool {tool_name} not found in registry"

@pytest.mark.integration
async def test_memory_persistence():
    """測試記憶持久化"""
    advisor = create_investment_advisor("memory_test")
    
    # 模擬多輪對話
    session_context = {"session_id": "test_session_123"}
    
    # 第一輪：設定投資偏好
    result_1 = await advisor.run(
        "我的風險偏好是保守型，投資期限是長期",
        context=session_context
    )
    
    # 第二輪：基於之前偏好進行查詢
    result_2 = await advisor.run(
        "請推薦適合的投資組合",
        context=session_context
    )
    
    # 驗證第二輪回應考慮了第一輪的偏好設定
    assert "保守" in result_2 or "穩健" in result_2
```

### 步驟 7: 配置與部署

#### 7.1 註冊到 Runtime

```python
# src/detectviz_adk/agents/registry.py
from agents.investment_advisor.agent import create_investment_advisor

class AgentRegistry:
    def _register_builtin_agents(self):
        """註冊內建 Agent"""
        
        # 註冊投資顧問
        investment_advisor = create_investment_advisor("default")
        self.register_agent("investment_advisor", investment_advisor)
        
        # 其他 Agent 註冊...
```

#### 7.2 配置監控

```yaml
# monitoring/investment-advisor-dashboard.json
{
  "title": "Investment Advisor Dashboard",
  "panels": [
    {
      "title": "Response Time Distribution",
      "targets": [
        {
          "expr": "histogram_quantile(0.95, detectviz_agent_response_time{agent_name=\"investment_advisor\"})"
        }
      ]
    },
    {
      "title": "Sub-Agent Usage",
      "targets": [
        {
          "expr": "rate(detectviz_sub_agent_calls_total{parent_agent=\"investment_advisor\"}[5m])"
        }
      ]
    },
    {
      "title": "Error Rate",
      "targets": [
        {
          "expr": "rate(detectviz_agent_errors_total{agent_name=\"investment_advisor\"}[5m])"
        }
      ]
    }
  ]
}
```

### 步驟 8: 驗證與上線

#### 8.1 功能驗證

```bash
# 運行單元測試
pytest agents/investment_advisor/tests/unit/ -v

# 運行整合測試  
pytest agents/investment_advisor/tests/integration/ -v

# 驗證模組卡
python contracts/tools/validate_module_card.py agents/investment_advisor/module.card.json

# 契約版本檢查
python scripts/validate_contracts.py
```

#### 8.2 性能測試

```python
# agents/investment_advisor/tests/performance/test_load.py
import pytest
import asyncio
import time
from concurrent.futures import ThreadPoolExecutor

from agents.investment_advisor.agent import create_investment_advisor

@pytest.mark.performance
async def test_concurrent_load():
    """測試並發負載"""
    advisor = create_investment_advisor("load_test")
    
    async def single_query():
        start_time = time.time()
        result = await advisor.run("分析當前市場趨勢")
        end_time = time.time()
        return end_time - start_time
    
    # 並發執行 50 個查詢
    tasks = [single_query() for _ in range(50)]
    response_times = await asyncio.gather(*tasks)
    
    # 驗證性能指標
    avg_response_time = sum(response_times) / len(response_times)
    max_response_time = max(response_times)
    
    assert avg_response_time < 5.0  # 平均回應時間 < 5 秒
    assert max_response_time < 10.0  # 最大回應時間 < 10 秒
    
    print(f"Average response time: {avg_response_time:.2f}s")
    print(f"Max response time: {max_response_time:.2f}s")
```

#### 8.3 部署檢查清單

**技術檢查**：
- [ ] 模組卡通過驗證
- [ ] 單元測試全部通過
- [ ] 整合測試全部通過
- [ ] 性能測試達標
- [ ] 契約版本一致性檢查通過

**業務檢查**：
- [ ] 核心業務流程測試通過
- [ ] 錯誤處理機制完善
- [ ] 日誌記錄充分
- [ ] 監控指標配置完成

**安全檢查**：
- [ ] 輸入驗證機制
- [ ] 敏感資訊保護
- [ ] 權限控制適當
- [ ] 審計日誌記錄

## 檔案對應清單

基於投資顧問範例，新 Service 需要創建的檔案：

```bash
agents/investment_advisor/                # 新 Service 根目錄
├── agent.py                            # 主 Agent 實作
├── prompts.py                          # 提示詞定義
├── module.card.json                    # 模組卡 (符合 SSOT 規範)
├── README.md                           # Service 說明文檔
├── sub_agents/                         # Sub-Agent 組織
│   ├── data_analyst/
│   │   ├── agent.py                    # 數據分析師實作
│   │   └── prompts.py                  # 專業提示詞
│   ├── risk_analyst/
│   │   ├── agent.py                    # 風險分析師實作
│   │   └── prompts.py                  # 專業提示詞
│   └── trading_analyst/
│       ├── agent.py                    # 交易分析師實作
│       └── prompts.py                  # 專業提示詞
├── shared/                             # 共享邏輯 (對應 shared_libraries)
│   ├── callbacks.py                   # 生命週期回調
│   ├── constants.py                   # 常數定義
│   ├── types.py                       # 型別定義
│   └── middleware.py                  # 中間件邏輯
├── tools/                             # 專屬工具 (可選)
│   └── investment_calculator.py       # 投資計算工具
├── tests/                             # 測試套件
│   ├── unit/
│   │   ├── test_agent.py              # 主 Agent 單元測試
│   │   ├── test_sub_agents.py         # Sub-Agent 測試
│   │   └── test_callbacks.py          # 回調函數測試
│   ├── integration/
│   │   ├── test_agent_integration.py   # 整合測試
│   │   └── test_tool_integration.py    # 工具整合測試
│   └── performance/
│       └── test_load.py               # 性能測試
└── profiles/                          # 配置檔案 (可選)
    └── risk_profiles.json             # 風險偏好配置
```

## 開發最佳實務

### 1. 狀態管理
- 使用工廠模式創建 Sub-Agent 獨立實例
- 通過 CallbackContext 管理會話狀態
- 避免 Agent 間的狀態污染

### 2. 錯誤處理
- 每個 Agent 都要有完整的異常處理
- Sub-Agent 失敗不應影響整個系統
- 提供降級策略和備用方案

### 3. 性能優化
- 並發執行獨立的 Sub-Agent 調用
- 使用適當的快取策略
- 監控和優化回應時間

### 4. 可觀測性
- 每個 Agent 調用都要有追蹤
- 記錄 Sub-Agent 的調用路徑
- 監控業務指標和技術指標

### 5. 測試策略
- 完整的單元測試覆蓋
- 整合測試驗證 Agent 協作
- 性能測試確保可擴展性

這套 Service 擴充指南確保新 Service 符合平台架構規範，提供穩定可靠的多 Agent 協作能力。