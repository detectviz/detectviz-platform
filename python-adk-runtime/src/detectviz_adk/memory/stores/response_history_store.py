from typing import Dict, Any, List, Optional
from google.adk.tools.tool_context import ToolContext

class ResponseHistoryStore:
    """
    符合 ADK 標準的回應歷史儲存器
    
    整合 ADK Session State 機制，透過 ToolContext 進行狀態管理。
    在生產環境中，可替換為持久化後端如 Redis、Firestore 或向量資料庫。
    """

    def __init__(self):
        self._store: Dict[str, Dict[str, Any]] = {}
        print("初始化記憶體回應歷史儲存器")

    @staticmethod
    def store_response_in_session(
        tool_context: ToolContext, 
        incident_id: str, 
        response: Dict[str, Any]
    ):
        """
        將事後檢討分析結果儲存到 ADK Session State
        
        Args:
            tool_context: ADK 工具上下文
            incident_id: 事件唯一識別碼
            response: 要儲存的分析結果
        """
        if "postmortem_history" not in tool_context.state:
            tool_context.state["postmortem_history"] = {}
        
        tool_context.state["postmortem_history"][incident_id] = response
        print(f"已將事件 {incident_id} 的回應儲存到會話狀態")

    @staticmethod
    def get_response_from_session(
        tool_context: ToolContext, 
        incident_id: str
    ) -> Optional[Dict[str, Any]]:
        """
        從 ADK Session State 取得特定事件的回應
        
        Args:
            tool_context: ADK 工具上下文
            incident_id: 事件唯一識別碼
            
        Returns:
            儲存的回應，如果未找到則回傳 None
        """
        history = tool_context.state.get("postmortem_history", {})
        return history.get(incident_id)

    @staticmethod
    def get_all_history_from_session(
        tool_context: ToolContext
    ) -> List[Dict[str, Any]]:
        """
        從 ADK Session State 取得所有歷史回應
        
        Args:
            tool_context: ADK 工具上下文
            
        Returns:
            所有儲存回應的清單
        """
        history = tool_context.state.get("postmortem_history", {})
        return list(history.values())

    # 保留舊有方法以維持向後相容性
    async def add_response(self, incident_id: str, response: Dict[str, Any]):
        """
        新增事後檢討分析回應到歷史儲存器
        
        Args:
            incident_id: 事件唯一識別碼
            response: 要儲存的分析結果
        """
        print(f"儲存事件回應：{incident_id}")
        self._store[incident_id] = response

    async def get_response(self, incident_id: str) -> Optional[Dict[str, Any]]:
        """
        根據事件 ID 取得特定的事後檢討回應
        
        Args:
            incident_id: 事件唯一識別碼
            
        Returns:
            儲存的回應，如果未找到則回傳 None
        """
        return self._store.get(incident_id)

    async def get_all_history(self) -> List[Dict[str, Any]]:
        """
        取得歷史儲存器中的所有回應
        
        Returns:
            所有儲存回應的清單
        """
        return list(self._store.values())
