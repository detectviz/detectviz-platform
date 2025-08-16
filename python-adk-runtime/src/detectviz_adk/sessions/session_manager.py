"""
ADK 會話管理
基於 [agent_team](https://github.com/google/adk-docs/tree/main/examples/python/tutorial/agent_team) 模式 - 會話狀態和 ToolContext 管理
"""
from typing import Dict, Any, Optional
from google.adk.sessions import InMemorySessionService
from google.adk.tools.tool_context import ToolContext


class StateAwareSessionManager:
    """
    擴展的會話管理器，支援狀態管理和 ToolContext
    """
    
    def __init__(self, app_name: str = "detectviz_postmortem"):
        self.app_name = app_name
        self.session_service = InMemorySessionService()
    
    async def create_session_with_state(
        self,
        user_id: str,
        session_id: str,
        initial_state: Optional[Dict[str, Any]] = None
    ):
        """建立帶有初始狀態的會話"""
        if initial_state is None:
            initial_state = {
                "postmortem_preferences": {
                    "report_format": "markdown",
                    "include_timeline": True,
                    "include_metrics": True
                },
                "session_metadata": {
                    "created_by": user_id,
                    "session_type": "postmortem_analysis"
                }
            }
        
        return await self.session_service.create_session(
            app_name=self.app_name,
            user_id=user_id,
            session_id=session_id,
            state=initial_state
        )
    
    async def get_session_state(
        self,
        user_id: str,
        session_id: str
    ) -> Optional[Dict[str, Any]]:
        """獲取會話狀態"""
        session = await self.session_service.get_session(
            app_name=self.app_name,
            user_id=user_id,
            session_id=session_id
        )
        return session.state if session else None
    
    async def update_session_state(
        self,
        user_id: str,
        session_id: str,
        state_updates: Dict[str, Any]
    ) -> bool:
        """更新會話狀態"""
        try:
            # 注意：這是針對 InMemorySessionService 的直接操作
            # 在生產環境中，應該使用更正式的 API
            session = await self.session_service.get_session(
                app_name=self.app_name,
                user_id=user_id,
                session_id=session_id
            )
            
            if session:
                session.state.update(state_updates)
                return True
            return False
        except Exception:
            return False


class PostmortemToolContext:
    """
    事後複盤專用的 ToolContext 擴展
    """
    
    @staticmethod
    def enhance_tool_context(tool_context: ToolContext) -> ToolContext:
        """增強 ToolContext 以支持事後複盤特定的狀態管理"""
        # 確保事後複盤相關的狀態鍵存在
        if "analysis_progress" not in tool_context.state:
            tool_context.state["analysis_progress"] = {
                "data_collected": False,
                "analysis_completed": False,
                "report_generated": False
            }
        
        if "collected_data" not in tool_context.state:
            tool_context.state["collected_data"] = {}
        
        if "analysis_results" not in tool_context.state:
            tool_context.state["analysis_results"] = {}
        
        return tool_context
    
    @staticmethod
    def mark_progress(tool_context: ToolContext, stage: str, completed: bool = True):
        """標記分析進度"""
        if "analysis_progress" in tool_context.state:
            tool_context.state["analysis_progress"][stage] = completed
    
    @staticmethod
    def store_analysis_data(
        tool_context: ToolContext,
        data_type: str,
        data: Any
    ):
        """存儲分析數據"""
        if "collected_data" not in tool_context.state:
            tool_context.state["collected_data"] = {}
        tool_context.state["collected_data"][data_type] = data
    
    @staticmethod
    def get_analysis_data(
        tool_context: ToolContext,
        data_type: str
    ) -> Optional[Any]:
        """獲取分析數據"""
        return tool_context.state.get("collected_data", {}).get(data_type)
