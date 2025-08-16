"""
事後檢討執行器實作
基於 [agent_team](https://github.com/google/adk-docs/tree/main/examples/python/tutorial/agent_team) 模式
"""
import asyncio
from typing import Dict, Any, Optional
from google.adk import Runner
from google.adk.sessions import InMemorySessionService
from google.genai import types
from ..agents.postmortem.orchestrator import postmortem_orchestrator


class PostmortemRunner:
    """
    符合 ADK 的事後檢討分析執行器
    管理會話和執行生命週期
    """
    
    def __init__(self, app_name: str = "detectviz_postmortem"):
        self.app_name = app_name
        self.session_service = InMemorySessionService()
        self.runner = Runner(
            agent=postmortem_orchestrator,
            app_name=app_name,
            session_service=self.session_service
        )
    
    async def execute_postmortem(
        self, 
        incident_request: Dict[str, Any],
        user_id: str = "system",
        session_id: Optional[str] = None
    ) -> Dict[str, Any]:
        """
        執行事後檢討分析
        
        Args:
            incident_request: 事件請求資料
            user_id: 使用者 ID
            session_id: 會話 ID，如果未提供將自動產生
        
        Returns:
            事後檢討分析結果
        """
        if session_id is None:
            session_id = f"incident-{incident_request.get('incident_id', 'unknown')}"
        
        # 建立會話
        session = await self.session_service.create_session(
            app_name=self.app_name,
            user_id=user_id,
            session_id=session_id
        )
        
        # 建構查詢
        query = self._build_query(incident_request)
        
        # 執行代理
        final_response = None
        async for event in self.runner.run_async(
            user_id=user_id,
            session_id=session_id,
            new_message=types.Content(
                role='user',
                parts=[types.Part(text=query)]
            )
        ):
            if event.is_final_response():
                if event.content and event.content.parts:
                    final_response = event.content.parts[0].text
                elif event.actions and event.actions.escalate:
                    final_response = f"Agent escalated: {event.error_message or 'No specific message.'}"
                break
        
        return {
            "status": "success" if final_response else "error",
            "response": final_response or "No response generated",
            "session_id": session_id,
            "incident_id": incident_request.get('incident_id')
        }
    
    def _build_query(self, incident_request: Dict[str, Any]) -> str:
        """建構事後檢討查詢"""
        incident_id = incident_request.get('incident_id', 'unknown')
        time_range = incident_request.get('time_range', {})
        affected_services = incident_request.get('affected_services', [])
        severity = incident_request.get('severity', 'unknown')
        
        query = f"""
        請為事件 {incident_id} 執行完整的事後檢討分析。

        事件詳情：
        - 事件 ID：{incident_id}
        - 時間範圍：{time_range.get('start', '未知')} 到 {time_range.get('end', '未知')}
        - 受影響服務：{', '.join(affected_services) if affected_services else '未指定'}
        - 嚴重程度：{severity}
        
        請按照標準流程：
        1. 收集相關資料和指標
        2. 分析根本原因
        3. 產生完整的事後檢討報告
        
        確保報告包含時間軸、根因分析、影響評估和改善建議。
        """
        
        return query.strip()
    
    async def get_session_state(self, user_id: str, session_id: str) -> Optional[Dict[str, Any]]:
        """獲取會話狀態"""
        session = await self.session_service.get_session(
            app_name=self.app_name,
            user_id=user_id,
            session_id=session_id
        )
        return session.state if session else None


# 便利函式用於快速使用
async def run_postmortem_analysis(incident_request: Dict[str, Any]) -> Dict[str, Any]:
    """
    快速執行事後檢討分析的便利函式
    
    Usage:
        incident = {
            "incident_id": "INC-2024-001",
            "time_range": {
                "start": "2024-01-15T10:00:00Z",
                "end": "2024-01-15T12:00:00Z"
            },
            "affected_services": ["payment-service", "api-gateway"],
            "severity": "P2"
        }
        
        result = await run_postmortem_analysis(incident)
    """
    runner = PostmortemRunner()
    return await runner.execute_postmortem(incident_request)
