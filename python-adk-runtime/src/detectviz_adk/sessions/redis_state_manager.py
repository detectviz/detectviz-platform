"""
Redis 狀態管理器
提供生產級的 Agent 狀態持久化和共享機制
"""
import json
import time
import asyncio
from typing import Dict, Any, Optional, Union
from dataclasses import dataclass, asdict
from datetime import datetime, timedelta

try:
    import redis.asyncio as redis
    HAS_REDIS = True
except ImportError:
    HAS_REDIS = False


@dataclass
class AgentState:
    """Agent 狀態數據結構"""
    session_id: str
    agent_id: str
    state_data: Dict[str, Any]
    created_at: float
    updated_at: float
    expires_at: Optional[float] = None
    version: int = 1

    def to_dict(self) -> Dict[str, Any]:
        """轉換為字典格式"""
        return asdict(self)

    @classmethod
    def from_dict(cls, data: Dict[str, Any]) -> 'AgentState':
        """從字典創建實例"""
        return cls(**data)

    def is_expired(self) -> bool:
        """檢查狀態是否過期"""
        if self.expires_at is None:
            return False
        return time.time() > self.expires_at


class RedisStateManager:
    """
    Redis 狀態管理器
    
    功能：
    - Agent 狀態持久化
    - 多 Agent 狀態共享
    - 狀態過期管理
    - 異常恢復機制
    """

    def __init__(
        self,
        redis_url: str = "redis://localhost:6379",
        key_prefix: str = "detectviz:agent:state",
        default_ttl: int = 3600,  # 1小時
        max_retries: int = 3
    ):
        self.redis_url = redis_url
        self.key_prefix = key_prefix
        self.default_ttl = default_ttl
        self.max_retries = max_retries
        self._pool: Optional[redis.ConnectionPool] = None
        self._client: Optional[redis.Redis] = None

    async def initialize(self) -> bool:
        """初始化 Redis 連接池"""
        if not HAS_REDIS:
            return False

        try:
            self._pool = redis.ConnectionPool.from_url(
                self.redis_url,
                max_connections=20,
                retry_on_timeout=True,
                health_check_interval=30
            )
            self._client = redis.Redis(connection_pool=self._pool)
            
            # 測試連接
            await self._client.ping()
            return True
        except Exception:
            return False

    async def close(self) -> None:
        """關閉連接池"""
        if self._client:
            await self._client.close()
        if self._pool:
            await self._pool.disconnect()

    def _get_state_key(self, session_id: str, agent_id: str) -> str:
        """生成狀態鍵名"""
        return f"{self.key_prefix}:{session_id}:{agent_id}"

    def _get_session_key(self, session_id: str) -> str:
        """生成會話鍵名"""
        return f"{self.key_prefix}:session:{session_id}"

    async def save_state(
        self,
        session_id: str,
        agent_id: str,
        state_data: Dict[str, Any],
        ttl: Optional[int] = None
    ) -> bool:
        """
        保存 Agent 狀態
        
        Args:
            session_id: 會話 ID
            agent_id: Agent ID
            state_data: 狀態數據
            ttl: 過期時間（秒），None 使用默認值
            
        Returns:
            是否保存成功
        """
        if not self._client:
            return False

        ttl = ttl or self.default_ttl
        now = time.time()
        
        agent_state = AgentState(
            session_id=session_id,
            agent_id=agent_id,
            state_data=state_data,
            created_at=now,
            updated_at=now,
            expires_at=now + ttl,
            version=1
        )

        try:
            key = self._get_state_key(session_id, agent_id)
            value = json.dumps(agent_state.to_dict())
            
            await self._client.setex(key, ttl, value)
            
            # 更新會話級別的 Agent 列表
            session_key = self._get_session_key(session_id)
            await self._client.sadd(session_key, agent_id)
            await self._client.expire(session_key, ttl)
            
            return True
        except Exception:
            return False

    async def load_state(
        self,
        session_id: str,
        agent_id: str
    ) -> Optional[AgentState]:
        """
        載入 Agent 狀態
        
        Args:
            session_id: 會話 ID
            agent_id: Agent ID
            
        Returns:
            Agent 狀態或 None
        """
        if not self._client:
            return None

        try:
            key = self._get_state_key(session_id, agent_id)
            value = await self._client.get(key)
            
            if not value:
                return None
                
            data = json.loads(value)
            agent_state = AgentState.from_dict(data)
            
            # 檢查是否過期
            if agent_state.is_expired():
                await self.delete_state(session_id, agent_id)
                return None
                
            return agent_state
        except Exception:
            return None

    async def update_state(
        self,
        session_id: str,
        agent_id: str,
        state_updates: Dict[str, Any],
        ttl: Optional[int] = None
    ) -> bool:
        """
        更新 Agent 狀態
        
        Args:
            session_id: 會話 ID
            agent_id: Agent ID
            state_updates: 狀態更新數據
            ttl: 過期時間（秒）
            
        Returns:
            是否更新成功
        """
        # 載入現有狀態
        current_state = await self.load_state(session_id, agent_id)
        
        if current_state:
            # 合併更新
            current_state.state_data.update(state_updates)
            current_state.updated_at = time.time()
            current_state.version += 1
            
            # 保存更新後的狀態
            return await self.save_state(
                session_id,
                agent_id,
                current_state.state_data,
                ttl
            )
        else:
            # 創建新狀態
            return await self.save_state(
                session_id,
                agent_id,
                state_updates,
                ttl
            )

    async def delete_state(
        self,
        session_id: str,
        agent_id: str
    ) -> bool:
        """
        刪除 Agent 狀態
        
        Args:
            session_id: 會話 ID
            agent_id: Agent ID
            
        Returns:
            是否刪除成功
        """
        if not self._client:
            return False

        try:
            key = self._get_state_key(session_id, agent_id)
            await self._client.delete(key)
            
            # 從會話 Agent 列表中移除
            session_key = self._get_session_key(session_id)
            await self._client.srem(session_key, agent_id)
            
            return True
        except Exception:
            return False

    async def get_session_agents(self, session_id: str) -> list[str]:
        """
        獲取會話中的所有 Agent ID
        
        Args:
            session_id: 會話 ID
            
        Returns:
            Agent ID 列表
        """
        if not self._client:
            return []

        try:
            session_key = self._get_session_key(session_id)
            agents = await self._client.smembers(session_key)
            return [agent.decode() if isinstance(agent, bytes) else agent for agent in agents]
        except Exception:
            return []

    async def get_session_states(self, session_id: str) -> Dict[str, AgentState]:
        """
        獲取會話中所有 Agent 的狀態
        
        Args:
            session_id: 會話 ID
            
        Returns:
            Agent ID 到狀態的映射
        """
        agents = await self.get_session_agents(session_id)
        states = {}
        
        for agent_id in agents:
            state = await self.load_state(session_id, agent_id)
            if state:
                states[agent_id] = state
                
        return states

    async def cleanup_expired_states(self) -> int:
        """
        清理過期的狀態
        
        Returns:
            清理的狀態數量
        """
        if not self._client:
            return 0

        try:
            # 使用 SCAN 遍歷所有狀態鍵
            pattern = f"{self.key_prefix}:*:*"
            cleaned_count = 0
            
            async for key in self._client.scan_iter(match=pattern):
                try:
                    value = await self._client.get(key)
                    if value:
                        data = json.loads(value)
                        agent_state = AgentState.from_dict(data)
                        
                        if agent_state.is_expired():
                            await self._client.delete(key)
                            cleaned_count += 1
                except Exception:
                    # 如果解析失敗，也刪除這個鍵
                    await self._client.delete(key)
                    cleaned_count += 1
                    
            return cleaned_count
        except Exception:
            return 0

    async def health_check(self) -> bool:
        """
        健康檢查
        
        Returns:
            Redis 連接是否正常
        """
        if not self._client:
            return False
            
        try:
            await self._client.ping()
            return True
        except Exception:
            return False


class FallbackStateManager:
    """
    後備狀態管理器（記憶體實作）
    當 Redis 不可用時使用
    """

    def __init__(self, default_ttl: int = 3600):
        self.default_ttl = default_ttl
        self._states: Dict[str, AgentState] = {}
        self._session_agents: Dict[str, set[str]] = {}
        self._cleanup_task: Optional[asyncio.Task] = None

    async def initialize(self) -> bool:
        """初始化（啟動清理任務）"""
        self._cleanup_task = asyncio.create_task(self._periodic_cleanup())
        return True

    async def close(self) -> None:
        """關閉（停止清理任務）"""
        if self._cleanup_task:
            self._cleanup_task.cancel()
            try:
                await self._cleanup_task
            except asyncio.CancelledError:
                pass

    def _get_state_key(self, session_id: str, agent_id: str) -> str:
        """生成狀態鍵名"""
        return f"{session_id}:{agent_id}"

    async def save_state(
        self,
        session_id: str,
        agent_id: str,
        state_data: Dict[str, Any],
        ttl: Optional[int] = None
    ) -> bool:
        """保存 Agent 狀態"""
        ttl = ttl or self.default_ttl
        now = time.time()
        
        agent_state = AgentState(
            session_id=session_id,
            agent_id=agent_id,
            state_data=state_data,
            created_at=now,
            updated_at=now,
            expires_at=now + ttl,
            version=1
        )

        key = self._get_state_key(session_id, agent_id)
        self._states[key] = agent_state
        
        # 更新會話 Agent 列表
        if session_id not in self._session_agents:
            self._session_agents[session_id] = set()
        self._session_agents[session_id].add(agent_id)
        
        return True

    async def load_state(
        self,
        session_id: str,
        agent_id: str
    ) -> Optional[AgentState]:
        """載入 Agent 狀態"""
        key = self._get_state_key(session_id, agent_id)
        agent_state = self._states.get(key)
        
        if agent_state and agent_state.is_expired():
            await self.delete_state(session_id, agent_id)
            return None
            
        return agent_state

    async def update_state(
        self,
        session_id: str,
        agent_id: str,
        state_updates: Dict[str, Any],
        ttl: Optional[int] = None
    ) -> bool:
        """更新 Agent 狀態"""
        current_state = await self.load_state(session_id, agent_id)
        
        if current_state:
            current_state.state_data.update(state_updates)
            current_state.updated_at = time.time()
            current_state.version += 1
            return True
        else:
            return await self.save_state(session_id, agent_id, state_updates, ttl)

    async def delete_state(
        self,
        session_id: str,
        agent_id: str
    ) -> bool:
        """刪除 Agent 狀態"""
        key = self._get_state_key(session_id, agent_id)
        self._states.pop(key, None)
        
        if session_id in self._session_agents:
            self._session_agents[session_id].discard(agent_id)
            if not self._session_agents[session_id]:
                del self._session_agents[session_id]
                
        return True

    async def get_session_agents(self, session_id: str) -> list[str]:
        """獲取會話中的所有 Agent ID"""
        return list(self._session_agents.get(session_id, set()))

    async def get_session_states(self, session_id: str) -> Dict[str, AgentState]:
        """獲取會話中所有 Agent 的狀態"""
        agents = await self.get_session_agents(session_id)
        states = {}
        
        for agent_id in agents:
            state = await self.load_state(session_id, agent_id)
            if state:
                states[agent_id] = state
                
        return states

    async def cleanup_expired_states(self) -> int:
        """清理過期的狀態"""
        now = time.time()
        expired_keys = []
        
        for key, state in self._states.items():
            if state.is_expired():
                expired_keys.append(key)
                
        for key in expired_keys:
            session_id, agent_id = key.split(":", 1)
            await self.delete_state(session_id, agent_id)
            
        return len(expired_keys)

    async def health_check(self) -> bool:
        """健康檢查"""
        return True

    async def _periodic_cleanup(self) -> None:
        """定期清理過期狀態"""
        while True:
            try:
                await asyncio.sleep(300)  # 5分鐘清理一次
                await self.cleanup_expired_states()
            except asyncio.CancelledError:
                break
            except Exception:
                pass  # 忽略清理錯誤