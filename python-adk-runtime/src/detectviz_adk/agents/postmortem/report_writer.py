"""
報告撰寫子代理
基於 adk_tutorial.ipynb 模式 - 專門負責報告產生的代理
"""
from google import adk
from ...tools.adk_tools import generate_report, create_dashboard, update_knowledge_base


# 報告撰寫代理（子代理）
report_writer = adk.Agent(
    model="gemini-2.0-flash",
    name="report_writer",
    instruction="""你是技術文件專家。你的任務是：
    1. 將分析結果整理成專業的事後檢討報告
    2. 包含清楚的時間軸、根因分析、影響評估和改善建議
    3. 使用適當的工具產生報告和儀表板
    4. 更新知識庫以供未來參考
    
    報告結構要求：
    - 執行摘要
    - 事件時間軸
    - 根本原因分析
    - 影響評估
    - 改善措施和預防建議
    - 學習重點
    
    使用 generate_report、create_dashboard 和 update_knowledge_base 工具。""",
    description="產生專業的事後檢討報告和文件",
    tools=[generate_report, create_dashboard, update_knowledge_base]
)
