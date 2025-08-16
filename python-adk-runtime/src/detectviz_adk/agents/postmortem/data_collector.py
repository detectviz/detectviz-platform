"""
資料收集子代理
基於 adk_tutorial.ipynb 模式 - 專門負責資料收集的代理
"""
from google import adk
from ...tools.adk_tools import get_health_metrics


# 資料收集代理（子代理）
data_collector_agent = adk.Agent(
    model="gemini-2.0-flash",
    name="data_collector",
    instruction="""你是資料收集專員。你的職責是：
    1. 從各個資料來源收集相關指標
    2. 確保資料的完整性和時間對齊
    3. 回傳結構化的資料集合
    
    使用 get_health_metrics 工具取得資料。
    專注於資料收集，不要進行分析或報告產生。""",
    description="負責收集和整理事故相關資料",
    tools=[get_health_metrics]
)
