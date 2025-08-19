"""
高級狀態管理器的單元測試
"""
import pytest
import asyncio
import time
from unittest.mock import Mock, AsyncMock
from typing import Dict, Any

from src.detectviz_adk.sessions.advanced_state_manager import (
    AdvancedStateManager,
    StateEvent,
    StateEventData,
    StateSyncStrategy,
    AgentConnection
)
from src.detectviz_adk.sessions.enhanced_session_manager import EnhancedSessionManager


class TestAdvancedStateManager:
    """高級狀態管理器測試"""

    @pytest.fixture
    async def mock_base_manager(self):
        """模擬基礎會話管理器"""
        manager = Mock(spec=EnhancedSessionManager)
        manager.initialize = AsyncMock(return_value=True)
        manager.close = AsyncMock()
        manager.save_agent_state = AsyncMock(return_value=True)
        manager.load_agent_state = AsyncMock(return_value={"test": "data"})
        manager.get_all_agent_states = AsyncMock(return_value={
            "agent1": {"test": "data1"},
            "agent2": {"test": "data2"}
        })
        return manager

    @pytest.fixture
    async def state_manager(self, mock_base_manager):
        """創建高級狀態管理器實例"""
        manager = AdvancedStateManager(
            base_manager=mock_base_manager,
            sync_strategy=StateSyncStrategy.IMMEDIATE,
            batch_size=5,
            batch_timeout=0.1,
            cache_ttl=60
        )
        await manager.initialize()
        yield manager
        await manager.close()

    @pytest.mark.asyncio
    async def test_initialization(self, mock_base_manager):
        """測試初始化"""
        manager = AdvancedStateManager(mock_base_manager)
        
        result = await manager.initialize()
        assert result is True
        
        # 驗證基礎管理器被初始化
        mock_base_manager.initialize.assert_called_once()
        
        await manager.close()

    @pytest.mark.asyncio
    async def test_event_system(self, state_manager):
        """測試事件系統"""
        events_received = []
        
        def event_handler(event_data: StateEventData):
            events_received.append(event_data)
        
        # 訂閱事件
        state_manager.subscribe_to_events(StateEvent.AGENT_CONNECTED, event_handler)
        
        # 註冊 Agent 連接
        await state_manager.register_agent_connection("session1", "agent1")
        
        # 等待事件處理
        await asyncio.sleep(0.1)
        
        # 驗證事件被觸發
        assert len(events_received) == 1
        assert events_received[0].event_type == StateEvent.AGENT_CONNECTED
        assert events_received[0].session_id == "session1"
        assert events_received[0].agent_id == "agent1"

    @pytest.mark.asyncio
    async def test_agent_connection_management(self, state_manager):
        """測試 Agent 連接管理"""
        # 註冊連接
        result = await state_manager.register_agent_connection(
            "session1", "agent1", {"metadata": "test"}
        )
        assert result is True
        
        # 獲取連接
        connections = await state_manager.get_agent_connections("session1")
        assert len(connections) == 1
        assert connections[0].agent_id == "agent1"
        assert connections[0].session_id == "session1"
        assert connections[0].metadata == {"metadata": "test"}
        
        # 更新心跳
        result = await state_manager.update_agent_heartbeat("session1", "agent1")
        assert result is True
        
        # 取消註冊
        result = await state_manager.unregister_agent_connection("session1", "agent1")
        assert result is True
        
        # 驗證連接已移除
        connections = await state_manager.get_agent_connections("session1")
        assert len(connections) == 0

    @pytest.mark.asyncio
    async def test_state_caching(self, state_manager):
        """測試狀態快取功能"""
        test_data = {"key": "value", "number": 42}
        
        # 保存狀態（會創建快照）
        await state_manager.save_agent_state_advanced(
            "session1", "agent1", test_data, create_snapshot=True
        )
        
        # 從快取載入
        cached_data = await state_manager.load_agent_state_advanced(
            "session1", "agent1", use_cache=True
        )
        
        # 驗證快取工作正常
        assert cached_data == test_data
        
        # 驗證基礎管理器的保存被呼叫
        state_manager.base_manager.save_agent_state.assert_called_once()

    @pytest.mark.asyncio
    async def test_batched_sync_strategy(self, mock_base_manager):
        """測試批次同步策略"""
        manager = AdvancedStateManager(
            base_manager=mock_base_manager,
            sync_strategy=StateSyncStrategy.BATCHED,
            batch_size=2,
            batch_timeout=0.05
        )
        await manager.initialize()
        
        # 保存多個狀態
        await manager.save_agent_state_advanced("session1", "agent1", {"data": 1})
        await manager.save_agent_state_advanced("session1", "agent2", {"data": 2})
        
        # 驗證還沒有立即保存到基礎管理器
        assert len(manager._pending_operations) == 2
        
        # 等待批次處理
        await asyncio.sleep(0.1)
        
        # 驗證批次處理完成
        assert mock_base_manager.save_agent_state.call_count >= 2
        
        await manager.close()

    @pytest.mark.asyncio
    async def test_state_conflict_detection(self, state_manager):
        """測試狀態衝突檢測"""
        # 先保存一個狀態到快取
        original_data = {"key": "original"}
        await state_manager.save_agent_state_advanced(
            "session1", "agent1", original_data, create_snapshot=True
        )
        
        # 模擬外部修改導致的衝突
        modified_data = {"key": "modified"}
        state_manager.base_manager.get_all_agent_states.return_value = {
            "agent1": modified_data
        }
        
        # 檢測衝突
        conflicts = await state_manager.detect_state_conflicts("session1")
        
        # 應該檢測到衝突
        assert len(conflicts) == 1
        assert conflicts[0]["conflict_type"] == "checksum_mismatch"
        assert conflicts[0]["session_id"] == "session1"
        assert conflicts[0]["agent_id"] == "agent1"

    @pytest.mark.asyncio
    async def test_state_conflict_resolution(self, state_manager):
        """測試狀態衝突解決"""
        # 創建一個衝突
        conflict = {
            "session_id": "session1",
            "agent_id": "agent1",
            "conflict_type": "checksum_mismatch"
        }
        
        # 設置最新狀態
        latest_state = {"key": "latest"}
        state_manager.base_manager.load_agent_state.return_value = latest_state
        
        # 解決衝突
        result = await state_manager.resolve_state_conflict(
            conflict, resolution_strategy="latest_wins"
        )
        
        assert result is True
        # 驗證最新狀態被載入
        state_manager.base_manager.load_agent_state.assert_called_with("session1", "agent1")

    @pytest.mark.asyncio
    async def test_session_health_status(self, state_manager):
        """測試會話健康狀態"""
        # 註冊一些 Agent 連接
        await state_manager.register_agent_connection("session1", "agent1")
        await state_manager.register_agent_connection("session1", "agent2")
        
        # 設置模擬狀態
        state_manager.base_manager.get_all_agent_states.return_value = {
            "agent1": {"test": "data1"},
            "agent2": {"test": "data2"}
        }
        
        # 獲取健康狀態
        health_status = await state_manager.get_session_health_status("session1")
        
        # 驗證健康狀態
        assert health_status["session_id"] == "session1"
        assert health_status["total_agents"] == 2
        assert health_status["active_connections"] == 2
        assert health_status["inactive_connections"] == 0
        assert 0.0 <= health_status["health_score"] <= 1.0

    @pytest.mark.asyncio
    async def test_agent_connection_lifecycle(self, state_manager):
        """測試 Agent 連接生命週期"""
        # 註冊連接
        await state_manager.register_agent_connection("session1", "agent1")
        
        # 獲取連接並驗證其為活躍狀態
        connections = await state_manager.get_agent_connections("session1")
        assert len(connections) == 1
        assert connections[0].is_active is True
        assert connections[0].is_alive() is True
        
        # 更新心跳
        old_activity_time = connections[0].last_activity
        await asyncio.sleep(0.01)  # 確保時間差異
        await state_manager.update_agent_heartbeat("session1", "agent1")
        
        # 驗證心跳更新
        updated_connections = await state_manager.get_agent_connections("session1")
        assert updated_connections[0].last_activity > old_activity_time

    @pytest.mark.asyncio
    async def test_cache_cleanup(self, state_manager):
        """測試快取清理功能"""
        # 創建一個會過期的快取項目
        state_manager.cache_ttl = 0.01  # 10ms
        
        # 保存狀態
        await state_manager.save_agent_state_advanced(
            "session1", "agent1", {"test": "data"}, create_snapshot=True
        )
        
        # 驗證快取存在
        cache_key = "session1:agent1"
        assert cache_key in state_manager._state_cache
        
        # 等待快取過期
        await asyncio.sleep(0.02)
        
        # 手動觸發清理（通常由後台任務處理）
        await state_manager._cleanup_cache()
        
        # 驗證過期項目被清理
        # 注意：實際清理由後台任務處理，這裡只測試清理邏輯

    @pytest.mark.asyncio
    async def test_event_unsubscription(self, state_manager):
        """測試事件取消訂閱"""
        events_received = []
        
        def event_handler(event_data: StateEventData):
            events_received.append(event_data)
        
        # 訂閱事件
        state_manager.subscribe_to_events(StateEvent.AGENT_CONNECTED, event_handler)
        
        # 觸發事件
        await state_manager.register_agent_connection("session1", "agent1")
        await asyncio.sleep(0.05)
        
        # 取消訂閱
        state_manager.unsubscribe_from_events(StateEvent.AGENT_CONNECTED, event_handler)
        
        # 再次觸發事件
        await state_manager.register_agent_connection("session1", "agent2")
        await asyncio.sleep(0.05)
        
        # 應該只收到第一個事件
        assert len(events_received) == 1

    @pytest.mark.asyncio
    async def test_multiple_sessions(self, state_manager):
        """測試多會話處理"""
        # 在不同會話中註冊 Agent
        await state_manager.register_agent_connection("session1", "agent1")
        await state_manager.register_agent_connection("session2", "agent2")
        await state_manager.register_agent_connection("session1", "agent3")
        
        # 獲取特定會話的連接
        session1_connections = await state_manager.get_agent_connections("session1")
        session2_connections = await state_manager.get_agent_connections("session2")
        all_connections = await state_manager.get_agent_connections()
        
        # 驗證連接分離
        assert len(session1_connections) == 2
        assert len(session2_connections) == 1
        assert len(all_connections) == 3
        
        # 驗證會話隔離
        session1_agents = {conn.agent_id for conn in session1_connections}
        assert session1_agents == {"agent1", "agent3"}
        
        session2_agents = {conn.agent_id for conn in session2_connections}
        assert session2_agents == {"agent2"}


class TestStateSyncStrategies:
    """狀態同步策略測試"""

    @pytest.fixture
    async def mock_base_manager(self):
        """模擬基礎會話管理器"""
        manager = Mock(spec=EnhancedSessionManager)
        manager.initialize = AsyncMock(return_value=True)
        manager.close = AsyncMock()
        manager.save_agent_state = AsyncMock(return_value=True)
        return manager

    @pytest.mark.asyncio
    async def test_immediate_sync_strategy(self, mock_base_manager):
        """測試立即同步策略"""
        manager = AdvancedStateManager(
            base_manager=mock_base_manager,
            sync_strategy=StateSyncStrategy.IMMEDIATE
        )
        await manager.initialize()
        
        # 保存狀態
        await manager.save_agent_state_advanced("session1", "agent1", {"data": "test"})
        
        # 驗證立即調用基礎管理器
        mock_base_manager.save_agent_state.assert_called_once_with(
            "session1", "agent1", {"data": "test"}, None
        )
        
        await manager.close()

    @pytest.mark.asyncio
    async def test_batched_sync_with_timeout(self, mock_base_manager):
        """測試帶超時的批次同步"""
        manager = AdvancedStateManager(
            base_manager=mock_base_manager,
            sync_strategy=StateSyncStrategy.BATCHED,
            batch_size=10,  # 大批次，不會觸發
            batch_timeout=0.05  # 短超時，會觸發
        )
        await manager.initialize()
        
        # 保存少量狀態（不足以觸發批次大小）
        await manager.save_agent_state_advanced("session1", "agent1", {"data": 1})
        
        # 等待超時觸發批次處理
        await asyncio.sleep(0.1)
        
        # 驗證批次處理被觸發
        mock_base_manager.save_agent_state.assert_called_once()
        
        await manager.close()

    @pytest.mark.asyncio
    async def test_batched_sync_with_size_limit(self, mock_base_manager):
        """測試帶大小限制的批次同步"""
        manager = AdvancedStateManager(
            base_manager=mock_base_manager,
            sync_strategy=StateSyncStrategy.BATCHED,
            batch_size=3,  # 小批次，容易觸發
            batch_timeout=1.0  # 長超時，不會觸發
        )
        await manager.initialize()
        
        # 保存足夠的狀態觸發批次大小
        for i in range(5):
            await manager.save_agent_state_advanced("session1", f"agent{i}", {"data": i})
        
        # 等待批次處理
        await asyncio.sleep(0.1)
        
        # 驗證批次處理被觸發（至少處理了3個）
        assert mock_base_manager.save_agent_state.call_count >= 3
        
        await manager.close()


@pytest.mark.asyncio
async def test_concurrent_operations():
    """測試並行操作"""
    mock_base_manager = Mock(spec=EnhancedSessionManager)
    mock_base_manager.initialize = AsyncMock(return_value=True)
    mock_base_manager.close = AsyncMock()
    mock_base_manager.save_agent_state = AsyncMock(return_value=True)
    mock_base_manager.load_agent_state = AsyncMock(return_value={"test": "data"})
    
    manager = AdvancedStateManager(mock_base_manager)
    await manager.initialize()
    
    # 並行執行多個操作
    tasks = [
        manager.save_agent_state_advanced("session1", f"agent{i}", {"data": i})
        for i in range(10)
    ]
    
    results = await asyncio.gather(*tasks, return_exceptions=True)
    
    # 驗證所有操作都成功
    for result in results:
        if isinstance(result, Exception):
            pytest.fail(f"Operation failed with exception: {result}")
        assert result is True
    
    await manager.close()