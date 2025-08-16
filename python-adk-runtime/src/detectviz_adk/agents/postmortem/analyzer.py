"""
根因分析子代理
基於 adk_tutorial.ipynb 模式 - 專門負責分析的代理
"""
from google import adk


# 根因分析代理（子代理）
root_cause_analyzer = adk.Agent(
    model="gemini-2.0-flash",
    name="root_cause_analyzer",
    instruction="""你是根因分析專家。基於收集的資料：
    1. 識別異常模式和相關性
    2. 關聯相關事件和指標變化
    3. 推論可能的根本原因
    4. 提供基於證據的分析結論
    
    重要原則：
    - 只基於提供的資料進行分析，不要猜測
    - 區分直接原因和根本原因
    - 考慮時間序列和因果關係
    - 提供可執行的洞察""",
    description="分析資料並識別根本原因",
    tools=[]  # 純分析，不需要工具
)
