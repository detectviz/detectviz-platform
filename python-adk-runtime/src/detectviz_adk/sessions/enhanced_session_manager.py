"""
增強的會話管理器
整合 Redis 狀態持久化和 Agent 間數據共享
"""
import asyncio
import os
from typing import Dict, Any, Optional, Union, List
from dataclasses import dataclass
from google.adk.sessions import InMemorySessionService
from google.adk.tools.tool_context import ToolContext

from .redis_state_manager import RedisStateManager, FallbackStateManager, AgentState


@dataclass
class AgentCommunication:
    """Agent 間通訊消息"""
    from_agent: str
    to_agent: str
    message_type: str
    data: Dict[str, Any]
    timestamp: float


class EnhancedSessionManager:
    """
    增強的會話管理器
    
    功能：
    - 基於 Redis 的持久化狀態管理
    - Agent 間數據共享
    - 狀態恢復機制
    - 異常處理和後備方案
    """

    def __init__(
        self,
        app_name: str = "detectviz_postmortem",
        redis_url: Optional[str] = None,
        use_redis: bool = True
    ):
        self.app_name = app_name
        self.session_service = InMemorySessionService()
        
        # 初始化狀態管理器
        if use_redis and redis_url is None:
            redis_url = os.getenv("REDIS_URL", "redis://localhost:6379")
            
        self.use_redis = use_redis
        self.redis_manager = RedisStateManager(redis_url) if use_redis else None
        self.fallback_manager = FallbackStateManager()
        self.state_manager = None  # 將在 initialize 中設定
        
        # Agent 間通訊緩存
        self._communication_cache: Dict[str, List[AgentCommunication]] = {}

    async def initialize(self) -> bool:
        """初始化會話管理器"""
        # 初始化後備管理器
        await self.fallback_manager.initialize()
        
        # 嘗試初始化 Redis 管理器
        if self.redis_manager:
            redis_ok = await self.redis_manager.initialize()
            if redis_ok:
                self.state_manager = self.redis_manager
                return True
        
        # 使用後備管理器
        self.state_manager = self.fallback_manager
        return True

    async def close(self) -> None:
        """關閉會話管理器"""
        if self.redis_manager:
            await self.redis_manager.close()
        await self.fallback_manager.close()

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
                    "include_metrics": True,
                    "language": "zh-TW"
                },
                "session_metadata": {
                    "created_by": user_id,
                    "session_type": "postmortem_analysis",
                    "version": "2.0"
                },
                "workflow_state": {
                    "current_stage": "initialized",
                    "completed_stages": [],
                    "agent_assignments": {}
                }
            }
        
        # 創建 ADK 會話
        session = await self.session_service.create_session(
            app_name=self.app_name,
            user_id=user_id,
            session_id=session_id,
            state=initial_state
        )
        
        # 初始化會話級別的狀態管理
        if self.state_manager:
            await self.state_manager.save_state(
                session_id=session_id,
                agent_id="__session__",
                state_data=initial_state
            )
        
        return session

    async def save_agent_state(
        self,
        session_id: str,
        agent_id: str,
        state_data: Dict[str, Any],
        ttl: Optional[int] = None
    ) -> bool:
        """保存 Agent 狀態"""
        if not self.state_manager:
            return False
            
        return await self.state_manager.save_state(
            session_id=session_id,
            agent_id=agent_id,
            state_data=state_data,
            ttl=ttl
        )

    async def load_agent_state(
        self,
        session_id: str,
        agent_id: str
    ) -> Optional[Dict[str, Any]]:
        """載入 Agent 狀態"""
        if not self.state_manager:
            return None
            
        agent_state = await self.state_manager.load_state(session_id, agent_id)
        return agent_state.state_data if agent_state else None

    async def update_agent_state(
        self,
        session_id: str,
        agent_id: str,
        state_updates: Dict[str, Any],
        ttl: Optional[int] = None
    ) -> bool:
        """更新 Agent 狀態"""
        if not self.state_manager:
            return False
            
        return await self.state_manager.update_state(
            session_id=session_id,
            agent_id=agent_id,
            state_updates=state_updates,
            ttl=ttl
        )

    async def get_session_state(
        self,
        user_id: str,
        session_id: str
    ) -> Optional[Dict[str, Any]]:
        """獲取會話狀態"""
        # 從 ADK 會話服務獲取
        session = await self.session_service.get_session(
            app_name=self.app_name,
            user_id=user_id,
            session_id=session_id
        )
        
        if session:
            return session.state
            
        # 嘗試從狀態管理器恢復
        if self.state_manager:
            session_state = await self.load_agent_state(session_id, "__session__")
            if session_state:
                # 重新創建會話
                await self.create_session_with_state(user_id, session_id, session_state)
                return session_state
                
        return None

    async def share_data_between_agents(
        self,
        session_id: str,
        from_agent: str,
        to_agent: str,
        data_key: str,
        data: Any
    ) -> bool:
        """在 Agent 間共享數據"""
        # 更新發送方的狀態
        await self.update_agent_state(
            session_id=session_id,
            agent_id=from_agent,
            state_updates={f"shared_data.{data_key}": data}
        )
        
        # 更新接收方的狀態
        await self.update_agent_state(
            session_id=session_id,
            agent_id=to_agent,
            state_updates={f"received_data.{data_key}": {
                "data": data,
                "from_agent": from_agent,
                "timestamp": asyncio.get_event_loop().time()
            }}
        )
        
        return True

    async def get_shared_data(
        self,
        session_id: str,
        agent_id: str,
        data_key: str
    ) -> Optional[Any]:
        """獲取共享數據"""
        agent_state = await self.load_agent_state(session_id, agent_id)
        if not agent_state:
            return None
            
        return agent_state.get("received_data", {}).get(data_key, {}).get("data")

    async def send_agent_message(
        self,
        session_id: str,
        from_agent: str,
        to_agent: str,
        message_type: str,
        data: Dict[str, Any]
    ) -> bool:
        """發送 Agent 間通訊消息"""
        message = AgentCommunication(
            from_agent=from_agent,
            to_agent=to_agent,
            message_type=message_type,
            data=data,
            timestamp=asyncio.get_event_loop().time()
        )
        
        # 儲存到接收方的消息隊列
        await self.update_agent_state(
            session_id=session_id,
            agent_id=to_agent,
            state_updates={
                f"messages.{message.timestamp}": {
                    "from_agent": from_agent,
                    "message_type": message_type,
                    "data": data,
                    "timestamp": message.timestamp
                }
            }
        )
        
        return True

    async def get_agent_messages(
        self,
        session_id: str,
        agent_id: str,
        message_type: Optional[str] = None
    ) -> List[AgentCommunication]:
        """獲取 Agent 的消息"""
        agent_state = await self.load_agent_state(session_id, agent_id)
        if not agent_state:
            return []
            
        messages = []
        for timestamp, msg_data in agent_state.get("messages", {}).items():
            if message_type is None or msg_data.get("message_type") == message_type:
                messages.append(AgentCommunication(
                    from_agent=msg_data["from_agent"],
                    to_agent=agent_id,
                    message_type=msg_data["message_type"],
                    data=msg_data["data"],
                    timestamp=msg_data["timestamp"]
                ))
        
        return sorted(messages, key=lambda x: x.timestamp)

    async def clear_agent_messages(
        self,
        session_id: str,
        agent_id: str
    ) -> bool:
        """清除 Agent 的消息"""
        return await self.update_agent_state(
            session_id=session_id,
            agent_id=agent_id,
            state_updates={"messages": {}}
        )

    async def get_all_agent_states(
        self,
        session_id: str
    ) -> Dict[str, Dict[str, Any]]:
        """獲取會話中所有 Agent 的狀態"""
        if not self.state_manager:
            return {}
            
        states = await self.state_manager.get_session_states(session_id)
        return {
            agent_id: state.state_data 
            for agent_id, state in states.items()
            if agent_id != "__session__"  # 排除會話級別狀態
        }

    async def restore_session(
        self,
        user_id: str,
        session_id: str
    ) -> bool:
        """恢復會話狀態"""
        if not self.state_manager:
            return False
            
        # 恢復會話級別狀態
        session_state = await self.load_agent_state(session_id, "__session__")
        if session_state:
            await self.create_session_with_state(user_id, session_id, session_state)
            return True
            
        return False

    async def cleanup_session(
        self,
        session_id: str
    ) -> bool:
        """清理會話狀態"""
        if not self.state_manager:
            return False
            
        # 獲取所有 Agent
        agents = await self.state_manager.get_session_agents(session_id)
        
        # 刪除所有 Agent 狀態
        for agent_id in agents:
            await self.state_manager.delete_state(session_id, agent_id)
            
        # 刪除會話狀態
        await self.state_manager.delete_state(session_id, "__session__")
        
        return True

    async def health_check(self) -> Dict[str, Any]:
        """健康檢查"""
        result = {
            "session_service": True,
            "redis_manager": False,
            "fallback_manager": False,
            "current_manager": "none"
        }
        
        # 檢查當前使用的狀態管理器
        if self.state_manager:
            is_healthy = await self.state_manager.health_check()
            if self.state_manager == self.redis_manager:
                result["redis_manager"] = is_healthy
                result["current_manager"] = "redis"
            else:
                result["fallback_manager"] = is_healthy
                result["current_manager"] = "fallback"
        
        return result


class PostmortemEnhancedToolContext:
    """
    事後複盤專用的增強 ToolContext
    整合會話狀態管理和 Agent 間通訊
    """

    def __init__(self, session_manager: EnhancedSessionManager):
        self.session_manager = session_manager

    async def enhance_tool_context(
        self,
        tool_context: ToolContext,
        session_id: str,
        agent_id: str
    ) -> ToolContext:
        """增強 ToolContext 以支持狀態管理"""
        # 載入 Agent 狀態
        agent_state = await self.session_manager.load_agent_state(session_id, agent_id)
        if agent_state:
            tool_context.state.update(agent_state)
        
        # 確保基本狀態結構存在
        if "analysis_progress" not in tool_context.state:
            tool_context.state["analysis_progress"] = {
                "data_collected": False,
                "analysis_completed": False,
                "report_generated": False,
                "current_stage": "initialized"
            }
        
        if "collected_data" not in tool_context.state:
            tool_context.state["collected_data"] = {}
        
        if "analysis_results" not in tool_context.state:
            tool_context.state["analysis_results"] = {}

        if "agent_metadata" not in tool_context.state:
            tool_context.state["agent_metadata"] = {
                "agent_id": agent_id,
                "session_id": session_id,
                "last_activity": asyncio.get_event_loop().time()
            }
        
        return tool_context

    async def save_tool_context_state(
        self,
        tool_context: ToolContext,
        session_id: str,
        agent_id: str
    ) -> bool:
        """保存 ToolContext 狀態"""
        # 更新最後活動時間
        tool_context.state["agent_metadata"]["last_activity"] = asyncio.get_event_loop().time()
        
        return await self.session_manager.save_agent_state(
            session_id=session_id,
            agent_id=agent_id,
            state_data=tool_context.state
        )

    async def share_analysis_data(
        self,
        tool_context: ToolContext,
        session_id: str,
        from_agent: str,
        to_agent: str,
        data_key: str,
        data: Any
    ) -> bool:
        """在 Agent 間共享分析數據"""
        # 儲存到本地狀態
        if "shared_analysis" not in tool_context.state:
            tool_context.state["shared_analysis"] = {}
        tool_context.state["shared_analysis"][data_key] = data
        
        # 共享給目標 Agent
        return await self.session_manager.share_data_between_agents(
            session_id=session_id,
            from_agent=from_agent,
            to_agent=to_agent,
            data_key=data_key,
            data=data
        )

    async def get_shared_analysis_data(
        self,
        session_id: str,
        agent_id: str,
        data_key: str
    ) -> Optional[Any]:
        """獲取共享的分析數據"""
        return await self.session_manager.get_shared_data(
            session_id=session_id,
            agent_id=agent_id,
            data_key=data_key
        )