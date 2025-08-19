"""
符合 ADK 標準的工具，使用 FunctionTool 包裝
基於實際 ADK API 模式實現
"""
from google.adk.tools import FunctionTool
from typing import Dict, List, Any, Optional
import asyncio
from .remote_tool import RemoteTool
from .report_tools import generate_postmortem_report, generate_chart


async def get_health_metrics_func(
    service_name: str,
    time_range: Dict[str, str],
    metrics: Optional[List[str]] = None
) -> Dict[str, Any]:
    """
    查詢服務健康指標。
    
    Args:
        service_name: 服務名稱
        time_range: 時間範圍 {"start": "...", "end": "..."}
        metrics: 要查詢的指標清單，預設為 ["error_rate", "latency", "cpu_usage"]
    
    Returns:
        包含健康指標的字典
    """
    if metrics is None:
        metrics = ["error_rate", "latency", "cpu_usage"]

    remote_tool = RemoteTool(
        tool_id="observability.health_aggregator",
        tool_version="0.1.0"
    )
    
    try:
        # Go 插件期望直接的 payload，而非巢狀的 "params"
        payload = {
            "service_name": service_name,
            "time_range": time_range,
            "metrics": metrics
        }
        return await remote_tool.invoke(payload)
    except Exception as e:
        return {
            "status": "error",
            "error_message": str(e),
            "service": service_name
        }


async def generate_report_func(
    incident_data: Dict[str, Any],
    format: str = "markdown"
) -> Dict[str, Any]:
    """
    生成事後複盤報告。
    
    Args:
        incident_data: 事件數據
        format: 報告格式，預設為 "markdown"
    
    Returns:
        生成的報告內容
    """
    remote_tool = RemoteTool(
        tool_id="reporting.report_generator",
        tool_version="0.1.0"
    )
    
    try:
        # Go 插件可能期望直接的資料結構
        payload = {
            "incident_data": incident_data,
            "format": format
        }
        return await remote_tool.invoke(payload)
    except Exception as e:
        return {
            "status": "error",
            "error_message": str(e),
            "format": format
        }


async def create_dashboard_func(
    incident_id: str,
    panels: List[Dict[str, Any]],
    time_range: Dict[str, str]
) -> Dict[str, Any]:
    """
    建立 Grafana 儀表板
    
    Args:
        incident_id: 事件 ID
        panels: 儀表板面板設定
        time_range: 時間範圍
    
    Returns:
        儀表板 URL 和相關資訊
    """
    remote_tool = RemoteTool(
        tool_id="reporting.dashboard_builder",
        tool_version="0.1.0"
    )
    
    try:
        result = await remote_tool.invoke({
            "action": "create_dashboard",
            "incident_id": incident_id,
            "panels": panels,
            "time_range": time_range
        })
        return result
    except Exception as e:
                return {
            "status": "error",
            "error_message": str(e),
            "incident_id": incident_id
        }


async def update_knowledge_base_func(
    incident_id: str,
    lessons_learned: List[str],
    root_causes: List[str],
    prevention_measures: List[str]
) -> Dict[str, Any]:
    """
    更新知識庫
    
    Args:
        incident_id: 事件 ID
        lessons_learned: 學習到的教訓
        root_causes: 根本原因
        prevention_measures: 預防措施
    
    Returns:
        更新結果
    """
    remote_tool = RemoteTool(
        tool_id="knowledge.knowledge_base",
        tool_version="0.1.0"
    )
    
    try:
        result = await remote_tool.invoke({
            "action": "update_knowledge",
            "incident_id": incident_id,
            "lessons_learned": lessons_learned,
            "root_causes": root_causes,
            "prevention_measures": prevention_measures
        })
        return result
    except Exception as e:
        return {
            "status": "error",
            "error_message": str(e),
            "incident_id": incident_id
        }


# 建立 FunctionTool 包裝器
get_health_metrics = FunctionTool(get_health_metrics_func)
generate_report = FunctionTool(generate_report_func)
create_dashboard = FunctionTool(create_dashboard_func)
update_knowledge_base = FunctionTool(update_knowledge_base_func)
