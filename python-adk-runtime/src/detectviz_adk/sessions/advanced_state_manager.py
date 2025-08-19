"""
高級狀態管理器
擴展現有狀態管理，增加事件驅動、狀態同步、性能優化等功能
"""
import asyncio
import json
import time
import weakref
from typing import Dict, Any, Optional, List, Callable, Set
from dataclasses import dataclass, field
from enum import Enum
from concurrent.futures import ThreadPoolExecutor

from .redis_state_manager import RedisStateManager, FallbackStateManager, AgentState
from .enhanced_session_manager import EnhancedSessionManager


class StateEvent(Enum):
    """狀態事件類型"""
    STATE_CREATED = "state_created"
    STATE_UPDATED = "state_updated"
    STATE_DELETED = "state_deleted"
    STATE_EXPIRED = "state_expired"
    AGENT_CONNECTED = "agent_connected"
    AGENT_DISCONNECTED = "agent_disconnected"
    STATE_SYNCED = "state_synced"
    STATE_CONFLICT = "state_conflict"


@dataclass
class StateEventData:
    """狀態事件數據"""
    event_type: StateEvent
    session_id: str
    agent_id: str
    timestamp: float
    data: Dict[str, Any] = field(default_factory=dict)
    metadata: Dict[str, Any] = field(default_factory=dict)


@dataclass
class StateSnapshot:
    """狀態快照"""
    session_id: str
    agent_id: str
    state_data: Dict[str, Any]
    version: int
    timestamp: float
    checksum: str


class StateSyncStrategy(Enum):
    """狀態同步策略"""
    IMMEDIATE = "immediate"  # 立即同步
    BATCHED = "batched"     # 批次同步
    PERIODIC = "periodic"   # 定期同步
    LAZY = "lazy"          # 懶惰同步


@dataclass
class AgentConnection:
    """Agent 連接狀態"""
    agent_id: str
    session_id: str
    connected_at: float
    last_activity: float
    heartbeat_interval: float = 30.0
    is_active: bool = True
    metadata: Dict[str, Any] = field(default_factory=dict)

    def is_alive(self) -> bool:
        """檢查連接是否存活"""
        return (time.time() - self.last_activity) < (self.heartbeat_interval * 2)


class AdvancedStateManager:
    """
    高級狀態管理器
    
    功能增強：
    1. 事件驅動架構
    2. Agent 連接狀態管理
    3. 狀態同步和衝突解決
    4. 性能優化（批次處理、快取）
    5. 狀態快照和版本控制
    6. 狀態變更訂閱機制
    """

    def __init__(
        self,
        base_manager: EnhancedSessionManager,
        sync_strategy: StateSyncStrategy = StateSyncStrategy.BATCHED,
        batch_size: int = 10,
        batch_timeout: float = 1.0,
        cache_ttl: int = 300  # 5分鐘
    ):
        self.base_manager = base_manager
        self.sync_strategy = sync_strategy
        self.batch_size = batch_size
        self.batch_timeout = batch_timeout
        self.cache_ttl = cache_ttl
        
        # 事件系統
        self._event_listeners: Dict[StateEvent, List[Callable]] = {}
        self._event_queue: asyncio.Queue = asyncio.Queue()
        self._event_processor_task: Optional[asyncio.Task] = None
        
        # Agent 連接管理
        self._agent_connections: Dict[str, AgentConnection] = {}
        self._connection_monitor_task: Optional[asyncio.Task] = None
        
        # 狀態快取
        self._state_cache: Dict[str, StateSnapshot] = {}
        self._cache_cleanup_task: Optional[asyncio.Task] = None
        
        # 批次處理
        self._pending_operations: List[Dict[str, Any]] = []
        self._batch_processor_task: Optional[asyncio.Task] = None
        
        # 狀態同步
        self._sync_conflicts: List[Dict[str, Any]] = []
        
        # 線程池用於計算密集型操作
        self._executor = ThreadPoolExecutor(max_workers=4)

    async def initialize(self) -> bool:
        """初始化高級狀態管理器"""
        # 初始化基礎管理器
        if not await self.base_manager.initialize():
            return False
        
        # 啟動事件處理器
        self._event_processor_task = asyncio.create_task(self._process_events())
        
        # 啟動連接監控
        self._connection_monitor_task = asyncio.create_task(self._monitor_connections())
        
        # 啟動快取清理
        self._cache_cleanup_task = asyncio.create_task(self._cleanup_cache())
        
        # 啟動批次處理器（如果需要）
        if self.sync_strategy == StateSyncStrategy.BATCHED:
            self._batch_processor_task = asyncio.create_task(self._process_batches())
        
        return True

    async def close(self) -> None:
        """關閉高級狀態管理器"""
        # 取消所有任務
        tasks = [
            self._event_processor_task,
            self._connection_monitor_task,
            self._cache_cleanup_task,
            self._batch_processor_task
        ]
        
        for task in tasks:
            if task:
                task.cancel()
        
        # 等待任務完成
        for task in tasks:
            if task:
                try:
                    await task
                except asyncio.CancelledError:
                    pass
        
        # 關閉線程池
        self._executor.shutdown(wait=True)
        
        # 關閉基礎管理器
        await self.base_manager.close()

    def subscribe_to_events(
        self,
        event_type: StateEvent,
        callback: Callable[[StateEventData], None]
    ) -> None:
        """訂閱狀態事件"""
        if event_type not in self._event_listeners:
            self._event_listeners[event_type] = []
        self._event_listeners[event_type].append(callback)

    def unsubscribe_from_events(
        self,
        event_type: StateEvent,
        callback: Callable[[StateEventData], None]
    ) -> None:
        """取消訂閱狀態事件"""
        if event_type in self._event_listeners:
            try:
                self._event_listeners[event_type].remove(callback)
            except ValueError:
                pass

    async def _emit_event(self, event_data: StateEventData) -> None:
        """發送事件"""
        await self._event_queue.put(event_data)

    async def _process_events(self) -> None:
        """處理事件隊列"""
        while True:
            try:
                event_data = await self._event_queue.get()
                
                # 執行事件監聽器
                listeners = self._event_listeners.get(event_data.event_type, [])
                for listener in listeners:
                    try:
                        if asyncio.iscoroutinefunction(listener):
                            await listener(event_data)
                        else:
                            listener(event_data)
                    except Exception as e:
                        # 記錄錯誤但不影響其他監聽器
                        pass
                
            except asyncio.CancelledError:
                break
            except Exception:
                pass

    async def register_agent_connection(
        self,
        session_id: str,
        agent_id: str,
        metadata: Optional[Dict[str, Any]] = None
    ) -> bool:
        """註冊 Agent 連接"""
        connection_key = f"{session_id}:{agent_id}"
        
        connection = AgentConnection(
            agent_id=agent_id,
            session_id=session_id,
            connected_at=time.time(),
            last_activity=time.time(),
            metadata=metadata or {}
        )
        
        self._agent_connections[connection_key] = connection
        
        # 發送連接事件
        await self._emit_event(StateEventData(
            event_type=StateEvent.AGENT_CONNECTED,
            session_id=session_id,
            agent_id=agent_id,
            timestamp=time.time(),
            data={"connection": connection.__dict__}
        ))
        
        return True

    async def unregister_agent_connection(
        self,
        session_id: str,
        agent_id: str
    ) -> bool:
        """取消註冊 Agent 連接"""
        connection_key = f"{session_id}:{agent_id}"
        
        if connection_key in self._agent_connections:
            del self._agent_connections[connection_key]
            
            # 發送斷開連接事件
            await self._emit_event(StateEventData(
                event_type=StateEvent.AGENT_DISCONNECTED,
                session_id=session_id,
                agent_id=agent_id,
                timestamp=time.time()
            ))
            
            return True
        
        return False

    async def update_agent_heartbeat(
        self,
        session_id: str,
        agent_id: str
    ) -> bool:
        """更新 Agent 心跳"""
        connection_key = f"{session_id}:{agent_id}"
        
        if connection_key in self._agent_connections:
            self._agent_connections[connection_key].last_activity = time.time()
            return True
        
        return False

    async def _monitor_connections(self) -> None:
        """監控 Agent 連接狀態"""
        while True:
            try:
                await asyncio.sleep(30)  # 每30秒檢查一次
                
                expired_connections = []
                for connection_key, connection in self._agent_connections.items():
                    if not connection.is_alive():
                        expired_connections.append(connection_key)
                
                # 清理過期連接
                for connection_key in expired_connections:
                    connection = self._agent_connections[connection_key]
                    await self.unregister_agent_connection(
                        connection.session_id,
                        connection.agent_id
                    )
                
            except asyncio.CancelledError:
                break
            except Exception:
                pass

    async def save_agent_state_advanced(
        self,
        session_id: str,
        agent_id: str,
        state_data: Dict[str, Any],
        ttl: Optional[int] = None,
        create_snapshot: bool = True
    ) -> bool:
        """高級狀態保存（支持快照和事件）"""
        # 更新心跳
        await self.update_agent_heartbeat(session_id, agent_id)
        
        # 計算狀態校驗和
        checksum = await self._calculate_checksum(state_data)
        
        # 創建狀態快照
        if create_snapshot:
            snapshot = StateSnapshot(
                session_id=session_id,
                agent_id=agent_id,
                state_data=state_data.copy(),
                version=int(time.time()),
                timestamp=time.time(),
                checksum=checksum
            )
            
            cache_key = f"{session_id}:{agent_id}"
            self._state_cache[cache_key] = snapshot
        
        # 根據同步策略處理
        if self.sync_strategy == StateSyncStrategy.IMMEDIATE:
            result = await self.base_manager.save_agent_state(
                session_id, agent_id, state_data, ttl
            )
        elif self.sync_strategy == StateSyncStrategy.BATCHED:
            # 添加到批次隊列
            self._pending_operations.append({
                "operation": "save",
                "session_id": session_id,
                "agent_id": agent_id,
                "state_data": state_data,
                "ttl": ttl,
                "timestamp": time.time()
            })
            result = True
        else:
            result = await self.base_manager.save_agent_state(
                session_id, agent_id, state_data, ttl
            )
        
        if result:
            # 發送狀態更新事件
            await self._emit_event(StateEventData(
                event_type=StateEvent.STATE_UPDATED,
                session_id=session_id,
                agent_id=agent_id,
                timestamp=time.time(),
                data={"checksum": checksum}
            ))
        
        return result

    async def load_agent_state_advanced(
        self,
        session_id: str,
        agent_id: str,
        use_cache: bool = True
    ) -> Optional[Dict[str, Any]]:
        """高級狀態載入（支持快取）"""
        cache_key = f"{session_id}:{agent_id}"
        
        # 嘗試從快取載入
        if use_cache and cache_key in self._state_cache:
            snapshot = self._state_cache[cache_key]
            if time.time() - snapshot.timestamp < self.cache_ttl:
                return snapshot.state_data.copy()
        
        # 從基礎管理器載入
        state_data = await self.base_manager.load_agent_state(session_id, agent_id)
        
        if state_data and use_cache:
            # 更新快取
            checksum = await self._calculate_checksum(state_data)
            snapshot = StateSnapshot(
                session_id=session_id,
                agent_id=agent_id,
                state_data=state_data.copy(),
                version=int(time.time()),
                timestamp=time.time(),
                checksum=checksum
            )
            self._state_cache[cache_key] = snapshot
        
        return state_data

    async def _process_batches(self) -> None:
        """處理批次操作"""
        while True:
            try:
                await asyncio.sleep(self.batch_timeout)
                
                if not self._pending_operations:
                    continue
                
                # 獲取批次操作
                batch = self._pending_operations[:self.batch_size]
                self._pending_operations = self._pending_operations[self.batch_size:]
                
                # 執行批次操作
                for op in batch:
                    try:
                        if op["operation"] == "save":
                            await self.base_manager.save_agent_state(
                                op["session_id"],
                                op["agent_id"],
                                op["state_data"],
                                op["ttl"]
                            )
                    except Exception:
                        pass  # 記錄錯誤但繼續處理其他操作
                
            except asyncio.CancelledError:
                break
            except Exception:
                pass

    async def _cleanup_cache(self) -> None:
        """清理過期快取"""
        while True:
            try:
                await asyncio.sleep(300)  # 5分鐘清理一次
                
                now = time.time()
                expired_keys = []
                
                for cache_key, snapshot in self._state_cache.items():
                    if now - snapshot.timestamp > self.cache_ttl:
                        expired_keys.append(cache_key)
                
                for key in expired_keys:
                    del self._state_cache[key]
                
            except asyncio.CancelledError:
                break
            except Exception:
                pass

    async def _calculate_checksum(self, data: Dict[str, Any]) -> str:
        """計算狀態數據校驗和"""
        def _compute():
            import hashlib
            json_str = json.dumps(data, sort_keys=True, ensure_ascii=False)
            return hashlib.md5(json_str.encode()).hexdigest()
        
        loop = asyncio.get_event_loop()
        return await loop.run_in_executor(self._executor, _compute)

    async def detect_state_conflicts(
        self,
        session_id: str
    ) -> List[Dict[str, Any]]:
        """檢測狀態衝突"""
        conflicts = []
        
        # 獲取所有 Agent 狀態
        all_states = await self.base_manager.get_all_agent_states(session_id)
        
        # 檢查快取中的版本與實際版本是否一致
        for agent_id, state_data in all_states.items():
            cache_key = f"{session_id}:{agent_id}"
            
            if cache_key in self._state_cache:
                cached_snapshot = self._state_cache[cache_key]
                current_checksum = await self._calculate_checksum(state_data)
                
                if cached_snapshot.checksum != current_checksum:
                    conflicts.append({
                        "session_id": session_id,
                        "agent_id": agent_id,
                        "conflict_type": "checksum_mismatch",
                        "cached_version": cached_snapshot.version,
                        "cached_checksum": cached_snapshot.checksum,
                        "current_checksum": current_checksum,
                        "timestamp": time.time()
                    })
        
        return conflicts

    async def resolve_state_conflict(
        self,
        conflict: Dict[str, Any],
        resolution_strategy: str = "latest_wins"
    ) -> bool:
        """解決狀態衝突"""
        session_id = conflict["session_id"]
        agent_id = conflict["agent_id"]
        
        if resolution_strategy == "latest_wins":
            # 重新載入最新狀態
            latest_state = await self.base_manager.load_agent_state(session_id, agent_id)
            if latest_state:
                # 更新快取
                await self.save_agent_state_advanced(
                    session_id, agent_id, latest_state, create_snapshot=True
                )
                return True
        
        return False

    async def get_agent_connections(
        self,
        session_id: Optional[str] = None
    ) -> List[AgentConnection]:
        """獲取 Agent 連接列表"""
        connections = []
        
        for connection in self._agent_connections.values():
            if session_id is None or connection.session_id == session_id:
                connections.append(connection)
        
        return connections

    async def get_session_health_status(
        self,
        session_id: str
    ) -> Dict[str, Any]:
        """獲取會話健康狀態"""
        # 獲取連接狀態
        connections = await self.get_agent_connections(session_id)
        active_agents = [conn for conn in connections if conn.is_active and conn.is_alive()]
        
        # 檢測衝突
        conflicts = await self.detect_state_conflicts(session_id)
        
        # 獲取狀態統計
        all_states = await self.base_manager.get_all_agent_states(session_id)
        
        return {
            "session_id": session_id,
            "total_agents": len(all_states),
            "active_connections": len(active_agents),
            "inactive_connections": len(connections) - len(active_agents),
            "state_conflicts": len(conflicts),
            "cache_hit_ratio": self._calculate_cache_hit_ratio(),
            "pending_operations": len(self._pending_operations),
            "health_score": self._calculate_health_score(
                len(active_agents), len(conflicts), len(self._pending_operations)
            ),
            "timestamp": time.time()
        }

    def _calculate_cache_hit_ratio(self) -> float:
        """計算快取命中率（簡化實現）"""
        # 這裡應該跟踪實際的命中統計
        return 0.8  # 暫時返回固定值

    def _calculate_health_score(
        self,
        active_connections: int,
        conflicts: int,
        pending_ops: int
    ) -> float:
        """計算健康評分（0-1之間）"""
        base_score = 1.0
        
        # 根據衝突數量減分
        conflict_penalty = min(conflicts * 0.1, 0.3)
        
        # 根據待處理操作減分
        pending_penalty = min(pending_ops * 0.01, 0.2)
        
        # 根據活躍連接加分
        connection_bonus = min(active_connections * 0.1, 0.2)
        
        return max(0.0, min(1.0, base_score - conflict_penalty - pending_penalty + connection_bonus))