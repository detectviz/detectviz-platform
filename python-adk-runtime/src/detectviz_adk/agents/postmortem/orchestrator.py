"""
事後檢討協調器 - 根代理
基於 [agent_team](https://github.com/google/adk-docs/tree/main/examples/python/tutorial/agent_team) 模式 - 主要協調代理
"""
from google import adk
from .data_collector import data_collector_agent
from .analyzer import root_cause_analyzer
from .report_writer import report_writer


# 事後檢討協調器（根代理）
postmortem_orchestrator = adk.Agent(
    model="gemini-2.0-flash",
    name="postmortem_orchestrator",
    instruction="""你是事後檢討協調器，負責管理整個檢討流程。

    你有以下子代理可以委派任務：
    1. 'data_collector': 收集事故相關資料和指標
    2. 'root_cause_analyzer': 分析根本原因和相關性
    3. 'report_writer': 產生完整報告和文件
    
    標準工作流程：
    1. 首先委派 data_collector 收集事件期間的所有相關資料
    2. 將收集的資料交給 root_cause_analyzer 進行深度分析
    3. 最後讓 report_writer 基於分析結果產生完整的事後檢討報告
    
    職責：
    - 理解事後檢討請求的範圍和需求
    - 協調各個子代理完成任務
    - 確保資料流在代理間正確傳遞
    - 監控整體流程的完整性
    - 提供最終的總結和建議
    
    重要：你不直接使用工具，而是透過委派給專門的子代理來完成任務。""",
    description="協調事後檢討流程的主代理",
    tools=[],  # Root Agent 不直接使用工具
    sub_agents=[data_collector_agent, root_cause_analyzer, report_writer]
)
