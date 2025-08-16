"""
符合 ADK 標準的工具，使用 FunctionTool 包裝
基於實際 ADK API 模式實現
"""
from google.adk.tools import FunctionTool
from typing import Dict, List, Any, Optional
import asyncio
from .remote_tool import RemoteTool


async def get_health_metrics_func(
    service_name: str,
    time_range: Dict[str, str],
    metrics: Optional[List[str]] = None
) -> Dict[str, Any]:
    """
    [空實作] 查詢服務健康指標。直接回傳模擬的異常數據。
    
    Args:
        service_name: 服務名稱
        time_range: 時間範圍 {"start": "...", "end": "..."}
        metrics: 要查詢的指標清單，預設為 ["error_rate", "latency", "cpu_usage"]
    
    Returns:
        包含健康指標的字典
    """
    print(f"--- [TOOL EXECUTED] get_health_metrics_func: Returning mock anomaly data for service '{service_name}'. ---")
    
    if metrics is None:
        metrics = ["error_rate", "latency", "cpu_usage"]
    
    # 模擬異常數據
    mock_anomaly_data = {
        "status": "success",
        "service": service_name,
        "time_range": time_range,
        "data": {
            "cpu_usage": {
                "summary": f"服務 {service_name} 的 CPU 使用率從 20% 飆升至 95%",
                "anomaly_detected": True,
                "peak_value": "95%",
                "normal_baseline": "20%"
            },
            "latency": {
                "summary": f"服務 {service_name} 的 P99 請求延遲從 150ms 增加到 2500ms", 
                "anomaly_detected": True,
                "peak_value": "2500ms",
                "normal_baseline": "150ms"
            },
            "error_rate": {
                "summary": f"服務 {service_name} 的錯誤率從 1% 上升至 15%",
                "anomaly_detected": True,
                "peak_value": "15%",
                "normal_baseline": "1%"
            }
        }
    }
    
    # 模擬網路延遲
    await asyncio.sleep(1)
    return mock_anomaly_data
    
    # 原有的 RemoteTool 程式碼保留但註解掉
    # remote_tool = RemoteTool(
    #     tool_id="observability.health_aggregator",
    #     tool_version="0.1.0"
    # )
    # 
    # try:
    #     result = await remote_tool.invoke({
    #         "action": "query_health",
    #         "params": {
    #             "service": service_name,
    #             "time_range": time_range,
    #             "metrics": metrics
    #         }
    #     })
    #     return result
    # except Exception as e:
    #     return {
    #         "status": "error",
    #         "error_message": str(e),
    #         "service": service_name
    #     }


async def generate_report_func(
    incident_data: Dict[str, Any],
    format: str = "markdown"
) -> Dict[str, Any]:
    """
    [空實作] 生成事後複盤報告。
    
    Args:
        incident_data: 事件數據
        format: 報告格式，預設為 "markdown"
    
    Returns:
        生成的報告內容
    """
    print("--- [TOOL EXECUTED] generate_report_func: Generating mock report. ---")
    
    analysis_summary = incident_data.get('analysis_summary', '分析總結未提供。')
    incident_id = incident_data.get('incident_id', 'UNKNOWN-INCIDENT')
    service_name = incident_data.get('service_name', 'unknown-service')
    
    report_content = f"""# 事後複盤報告 (模擬)

## 事件資訊
- **事件 ID**: {incident_id}
- **受影響服務**: {service_name}
- **報告生成時間**: 模擬時間戳

## 根本原因分析

{analysis_summary}

## 改善建議 (佔位)

1. 增加服務器資源監控告警。
2. 對相關服務進行壓力測試。
3. 建立自動擴容機制，防止資源耗盡。
4. 實施更嚴格的負載均衡策略。

## 預防措施

- 設定 CPU 使用率告警閾值為 80%
- 建立延遲監控儀表板
- 定期進行災難復原演練

---
*本報告由 DetectViz 平台自動生成*
"""
    
    # 模擬生成延遲
    await asyncio.sleep(1)
    return {
        "status": "success",
        "report_content": report_content.strip(),
        "format": format,
        "incident_id": incident_id
    }
    
    # 原有的 RemoteTool 程式碼保留但註解掉
    # remote_tool = RemoteTool(
    #     tool_id="reporting.report_generator", 
    #     tool_version="0.1.0"
    # )
    # 
    # try:
    #     result = await remote_tool.invoke({
    #         "action": "generate_report",
    #         "data": incident_data,
    #         "format": format
    #     })
    #     return result
    # except Exception as e:
    #     return {
    #         "status": "error",
    #         "error_message": str(e),
    #         "format": format
    #     }


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
