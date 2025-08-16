#!/usr/bin/env python3
"""
簡化的 ADK 測試
檢查基本的 Agent 建立是否可行
"""

import sys
from pathlib import Path

# 添加 src 到 Python 路徑
sys.path.insert(0, str(Path(__file__).parent / "src"))

try:
    print("=== 測試基本 ADK 匯入 ===")
    
    from google.adk import Agent
    from google.adk.tools import FunctionTool
    print("✅ ADK 基本匯入成功")
    
    print("\n=== 測試簡單函式工具 ===")
    
    # 簡單的測試函式
    def simple_test_tool(message: str = "Hello") -> str:
        """簡單的測試工具"""
        return f"工具回應: {message}"
    
    # 建立 FunctionTool
    test_tool = FunctionTool(simple_test_tool)
    print(f"✅ FunctionTool 建立成功: {test_tool.name}")
    
    print("\n=== 測試 Agent 建立 ===")
    
    # 嘗試建立最簡單的 Agent
    try:
        simple_agent = Agent(
            name="test_agent",
            description="測試代理",
            instruction="你是一個測試代理",
            tools=[test_tool]
        )
        print(f"✅ Agent 建立成功: {simple_agent.name}")
        print(f"   描述: {simple_agent.description}")
        print(f"   工具數量: {len(simple_agent.tools)}")
        
    except Exception as e:
        print(f"❌ Agent 建立失敗: {e}")
        # 嘗試不同的參數組合
        print("嘗試簡化參數...")
        try:
            simple_agent = Agent(
                name="test_agent",
                instruction="你是一個測試代理"
            )
            print(f"✅ 簡化 Agent 建立成功: {simple_agent.name}")
        except Exception as e2:
            print(f"❌ 簡化 Agent 也失敗: {e2}")
            
    print("\n=== 測試 Runner 匯入 ===")
    from google.adk import Runner
    print("✅ Runner 匯入成功")
    
    # 檢查是否有 sessions 模組
    try:
        from google.adk.sessions import InMemorySessionService
        print("✅ InMemorySessionService 匯入成功")
    except ImportError as e:
        print(f"⚠️ InMemorySessionService 匯入失敗: {e}")
        # 檢查替代方案
        try:
            from google.adk import sessions
            print(f"可用的 sessions: {dir(sessions)}")
        except:
            print("❌ 無法找到 sessions 模組")
    
    print("\n🎉 基本 ADK 結構檢查完成！")
    
except ImportError as e:
    print(f"❌ ADK 匯入失敗: {e}")
    print("請確認 google-adk 已正確安裝")
    
except Exception as e:
    print(f"❌ 測試失敗: {e}")
    import traceback
    traceback.print_exc()
