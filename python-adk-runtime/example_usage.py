#!/usr/bin/env python3
"""
DetectViz ADK Runtime 使用範例
展示如何使用符合 ADK 標準的事後檢討系統
"""

import asyncio
from detectviz_adk import run_postmortem_analysis, PostmortemRunner


async def simple_example():
    """簡單使用範例 - 使用便利函式"""
    print("=== 簡單使用範例 ===")
    
    # 事件資料
    incident = {
        "incident_id": "INC-2024-001",
        "time_range": {
            "start": "2024-01-15T10:00:00Z",
            "end": "2024-01-15T12:00:00Z"
        },
        "affected_services": ["payment-service", "api-gateway"],
        "severity": "P2"
    }
    
    # 執行事後檢討分析
    result = await run_postmortem_analysis(incident)
    
    print(f"會話 ID: {result.get('session_id', 'N/A')}")
    print(f"事件 ID: {result.get('incident_id', 'N/A')}")
    print("分析結果:")
    print(result.get('response', 'No response generated'))


async def advanced_example():
    """進階使用範例 - 使用 PostmortemRunner 類別"""
    print("\n=== 進階使用範例 ===")
    
    # 建立執行器
    runner = PostmortemRunner(app_name="detectviz_demo")
    
    # 事件資料
    incident = {
        "incident_id": "INC-2024-002", 
        "time_range": {
            "start": "2024-01-16T14:30:00Z",
            "end": "2024-01-16T16:45:00Z"
        },
        "affected_services": ["user-service", "notification-service", "database"],
        "severity": "P1"
    }
    
    # 執行分析
    result = await runner.execute_postmortem(
        incident_request=incident,
        user_id="demo_user",
        session_id="demo_session_001"
    )
    
    print(f"會話 ID: {result.get('session_id', 'N/A')}")
    print(f"事件 ID: {result.get('incident_id', 'N/A')}")
    print("分析結果:")
    print(result.get('response', 'No response generated'))
    
    # 檢查會話狀態
    session_state = await runner.get_session_state("demo_user", "demo_session_001")
    if session_state:
        print("\n會話狀態:")
        for key, value in session_state.items():
            print(f"  {key}: {value}")


async def main():
    """主執行函式"""
    print("DetectViz ADK Runtime 示範")
    print("符合 Google ADK 標準的事後檢討系統")
    print("=" * 50)
    
    try:
        # 執行簡單範例
        await simple_example()
        
        # 執行進階範例  
        await advanced_example()
        
    except Exception as e:
        print(f"執行錯誤: {e}")
        print("注意：此範例需要完整的 ADK 環境和 API 金鑰設定")


if __name__ == "__main__":
    # 執行示範
    asyncio.run(main())
