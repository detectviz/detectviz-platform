"""
符合 ADK 標準的記憶體管理工具
基於實際 ADK API 的 Session State 模式實現
"""
from google.adk.tools import FunctionTool
from google.adk.tools.tool_context import ToolContext
from typing import Dict, List, Any, Optional
from ..memory.stores.response_history_store import ResponseHistoryStore


async def remember_analysis_func(
    key: str,
    value: Any,
    tool_context: ToolContext
) -> str:
    """
    在會話中記住分析結果
    
    Args:
        key: 儲存鍵值
        value: 要儲存的值
        tool_context: ADK 工具上下文
    
    Returns:
        確認訊息
    """
    # 使用 ADK 的 session state
    tool_context.state[key] = value
    return f"已記住 {key}"


async def recall_analysis_func(
    key: str,
    tool_context: ToolContext
) -> Any:
    """
    從會話中回憶分析結果
    
    Args:
        key: 要取得的鍵值
        tool_context: ADK 工具上下文
    
    Returns:
        儲存的值，如果不存在則回傳 None
    """
    return tool_context.state.get(key, None)


async def store_postmortem_response_func(
    incident_id: str,
    response: Dict[str, Any],
    tool_context: ToolContext
) -> str:
    """
    儲存事後檢討回應到會話狀態
    
    Args:
        incident_id: 事件 ID
        response: 事後檢討回應
        tool_context: ADK 工具上下文
    
    Returns:
        確認訊息
    """
    ResponseHistoryStore.store_response_in_session(
        tool_context, incident_id, response
    )
    return f"已儲存事件 {incident_id} 的檢討結果"


async def get_postmortem_history_func(
    incident_id: Optional[str] = None,
    tool_context: ToolContext = None
) -> Dict[str, Any]:
    """
    取得事後檢討歷史記錄
    
    Args:
        incident_id: 特定事件 ID（可選）
        tool_context: ADK 工具上下文
    
    Returns:
        歷史記錄資料
    """
    if incident_id:
        # 取得特定事件的記錄
        response = ResponseHistoryStore.get_response_from_session(
            tool_context, incident_id
        )
        return {
            "incident_id": incident_id,
            "response": response,
            "found": response is not None
        }
    else:
        # 取得所有歷史記錄
        all_history = ResponseHistoryStore.get_all_history_from_session(tool_context)
        return {
            "total_count": len(all_history),
            "history": all_history
        }


async def clear_session_memory_func(
    confirm: bool = False,
    tool_context: ToolContext = None
) -> str:
    """
    清除會話記憶體（需要確認）
    
    Args:
        confirm: 確認清除（必須為 True）
        tool_context: ADK 工具上下文
    
    Returns:
        操作結果訊息
    """
    if not confirm:
        return "清除記憶體需要明確確認。請設定 confirm=True"
    
    # 清除分析相關的狀態
    keys_to_clear = [
        "postmortem_history",
        "analysis_progress", 
        "collected_data",
        "analysis_results"
    ]
    
    cleared_count = 0
    for key in keys_to_clear:
        if key in tool_context.state:
            del tool_context.state[key]
            cleared_count += 1
    
    return f"已清除 {cleared_count} 個記憶體項目"


# 建立 FunctionTool 包裝器
remember_analysis = FunctionTool(remember_analysis_func)
recall_analysis = FunctionTool(recall_analysis_func)
store_postmortem_response = FunctionTool(store_postmortem_response_func)
get_postmortem_history = FunctionTool(get_postmortem_history_func)
clear_session_memory = FunctionTool(clear_session_memory_func)
