#!/usr/bin/env python3
"""
ADK 整合測試
測試新重構的 ADK 對齊架構是否正常運作
"""

import asyncio
import sys
import os
from pathlib import Path

# 添加 src 到 Python 路徑
sys.path.insert(0, str(Path(__file__).parent / "src"))

try:
    # 測試基本匯入
    print("=== 測試基本匯入 ===")
    
    # 測試 ADK 工具匯入
    from detectviz_adk.tools.adk_tools import (
        get_health_metrics,
        generate_report,
        create_dashboard,
        update_knowledge_base
    )
    print("✅ ADK 工具匯入成功")
    
    # 測試記憶體工具匯入
    from detectviz_adk.tools.memory_tools import (
        remember_analysis,
        recall_analysis,
        store_postmortem_response,
        get_postmortem_history
    )
    print("✅ 記憶體工具匯入成功")
    
    # 測試代理匯入
    from detectviz_adk.agents.postmortem import (
        data_collector_agent,
        root_cause_analyzer,
        report_writer,
        postmortem_orchestrator
    )
    print("✅ ADK 代理匯入成功")
    
    # 測試執行器匯入
    from detectviz_adk.runners.postmortem_runner import PostmortemRunner, run_postmortem_analysis
    print("✅ 執行器匯入成功")
    
    # 測試會話管理匯入
    from detectviz_adk.sessions.session_manager import StateAwareSessionManager, PostmortemToolContext
    print("✅ 會話管理匯入成功")
    
    # 測試記憶體儲存器匯入
    from detectviz_adk.memory.stores.response_history_store import ResponseHistoryStore
    print("✅ 記憶體儲存器匯入成功")
    
    print("\n=== 測試基本功能 ===")
    
    # 測試代理基本屬性
    print(f"根代理名稱: {postmortem_orchestrator.name}")
    print(f"根代理模型: {postmortem_orchestrator.model}")
    print(f"子代理數量: {len(postmortem_orchestrator.sub_agents)}")
    print(f"子代理清單: {[agent.name for agent in postmortem_orchestrator.sub_agents]}")
    
    # 測試工具屬性
    print(f"資料收集代理工具: {[tool.name for tool in data_collector_agent.tools]}")
    print(f"報告撰寫代理工具: {[tool.name for tool in report_writer.tools]}")
    
    print("\n=== 測試執行器建立 ===")
    
    # 測試執行器建立
    runner = PostmortemRunner(app_name="test_app")
    print(f"✅ 執行器建立成功，應用名稱: {runner.app_name}")
    
    # 測試會話管理器建立
    session_manager = StateAwareSessionManager(app_name="test_app")
    print(f"✅ 會話管理器建立成功，應用名稱: {session_manager.app_name}")
    
    # 測試記憶體儲存器建立
    memory_store = ResponseHistoryStore()
    print("✅ 記憶體儲存器建立成功")
    
    print("\n=== 模擬測試（不需要真實 API） ===")
    
    async def test_basic_workflow():
        """測試基本工作流程（模擬）"""
        print("開始基本工作流程測試...")
        
        # 模擬事件資料
        test_incident = {
            "incident_id": "TEST-001",
            "time_range": {
                "start": "2024-01-20T10:00:00Z",
                "end": "2024-01-20T11:00:00Z"
            },
            "affected_services": ["test-service"],
            "severity": "P3"
        }
        
        # 測試查詢建構
        query = runner._build_query(test_incident)
        print(f"✅ 查詢建構成功，長度: {len(query)} 字元")
        
        # 測試會話狀態建立
        session = await session_manager.create_session_with_state(
            user_id="test_user",
            session_id="test_session_001"
        )
        print("✅ 會話狀態建立成功")
        
        # 測試狀態取得
        state = await session_manager.get_session_state(
            user_id="test_user",
            session_id="test_session_001"
        )
        print(f"✅ 會話狀態取得成功，包含 {len(state)} 個項目")
        
        # 測試記憶體儲存（傳統方式）
        await memory_store.add_response("TEST-001", {"status": "completed", "summary": "測試分析"})
        stored_response = await memory_store.get_response("TEST-001")
        print(f"✅ 記憶體儲存測試成功：{stored_response['status']}")
        
        print("基本工作流程測試完成！")
    
    # 執行異步測試
    asyncio.run(test_basic_workflow())
    
    print("\n🎉 所有測試通過！")
    print("ADK 對齊架構重構成功，系統已準備就緒。")
    
except ImportError as e:
    print(f"❌ 匯入錯誤: {e}")
    print("請確認 requirements.txt 中的依賴項目已安裝")
    print("執行: pip install -r requirements.txt")
    sys.exit(1)
    
except Exception as e:
    print(f"❌ 測試失敗: {e}")
    import traceback
    traceback.print_exc()
    sys.exit(1)

print("\n📋 下一步:")
print("1. 安裝完整的 google-adk 依賴")
print("2. 設定 API 金鑰（GOOGLE_API_KEY 等）")
print("3. 執行 example_usage.py 進行完整測試")
print("4. 開始 MVP 開發流程")
