"""
測試 Markdown 報告生成引擎
在 Docker 環境中運行測試
"""
import asyncio
import time
import json
from datetime import datetime, timedelta
from pathlib import Path
from templates.report_engine import (
    PostmortemReportEngine,
    ReportLanguage,
    ChartConfig,
    TableConfig
)


async def test_report_generation():
    """測試報告生成功能"""
    print("🚀 開始測試 Markdown 報告生成引擎...")
    
    # 創建報告引擎
    engine = PostmortemReportEngine(
        template_dir="./templates",
        output_dir="./reports",
        default_language=ReportLanguage.CHINESE
    )
    
    # 準備測試數據
    incident_data = {
        'summary': '由於資料庫連接池耗盡導致的服務中斷事件。影響了主要 API 服務，持續約 2 小時。',
        'incident': {
            'severity': 'high',
            'affected_services': ['user-api', 'payment-api', 'notification-service'],
            'duration': 7200,  # 2 小時 (秒)
            'detection_time': datetime.now() - timedelta(hours=2),
            'resolution_time': datetime.now(),
            'responsible_team': 'Backend Infrastructure Team'
        },
        'impact': {
            'customer_affected': 15000,
            'affected_regions': ['US-East', 'EU-West', 'Asia-Pacific'],
            'revenue_loss': 50000,
            'system_degradation': 85,
            'data_loss': False,
            'customer': '約 15,000 名用戶受到影響，無法正常使用支付功能。客服收到大量投訴。',
            'business': '預估收入損失約 $50,000，主要來自支付服務中斷。品牌聲譽受到一定影響。',
            'technical': '資料庫連接池耗盡，API 響應時間增加到 30+ 秒，系統整體可用性降至 15%。'
        },
        'timeline': [
            {
                'timestamp': datetime.now() - timedelta(hours=2),
                'title': '異常檢測',
                'severity': 'info',
                'description': '監控系統檢測到 API 響應時間異常升高',
                'actions': ['觸發自動告警', '通知 on-call 工程師'],
                'metrics': {'response_time': '15s', 'error_rate': '5%'}
            },
            {
                'timestamp': datetime.now() - timedelta(minutes=115),
                'title': '問題確認',
                'severity': 'medium',
                'description': '工程師確認是資料庫連接問題',
                'actions': ['檢查資料庫狀態', '查看連接池配置'],
                'metrics': {'db_connections': '100/100', 'queue_length': '500'}
            },
            {
                'timestamp': datetime.now() - timedelta(minutes=105),
                'title': '問題升級',
                'severity': 'high',
                'description': '服務完全不可用，升級為 P0 事件',
                'actions': ['啟動 incident response', '通知管理層', '成立應急小組'],
                'metrics': {'uptime': '0%', 'error_rate': '100%'}
            },
            {
                'timestamp': datetime.now() - timedelta(hours=1),
                'title': '臨時修復',
                'severity': 'medium', 
                'description': '重啟服務並增加連接池大小',
                'actions': ['重啟應用服務', '調整連接池配置', '監控系統恢復'],
                'metrics': {'uptime': '50%', 'error_rate': '25%'}
            },
            {
                'timestamp': datetime.now() - timedelta(minutes=30),
                'title': '服務恢復',
                'severity': 'low',
                'description': '所有服務恢復正常，開始事後調查',
                'actions': ['確認服務穩定', '開始根因分析', '準備事後複盤'],
                'metrics': {'uptime': '100%', 'error_rate': '0.1%'}
            }
        ],
        'root_cause': {
            'primary': '資料庫連接池配置不當，最大連接數設置過低無法處理高峰流量',
            'contributing': [
                '缺乏連接池監控告警',
                '負載測試未覆蓋極端場景', 
                '沒有連接池動態調整機制',
                '資料庫查詢存在未優化的慢查詢'
            ],
            'root': '系統容量規劃不足，缺乏對資料庫連接數的準確評估和動態調整能力',
            'technical_details': '''
資料庫連接池配置:
- maxPoolSize: 20 (應該是 100+)
- connectionTimeout: 30000ms
- idleTimeout: 600000ms

慢查詢分析:
- SELECT * FROM users WHERE status = 'active' AND last_login > ... (平均 2.5s)
- JOIN 查詢未使用索引，導致全表掃描

流量分析:
- 正常: 50 QPS
- 高峰: 200 QPS  
- 事件期間: 300+ QPS
            '''
        },
        'action_items': [
            {
                'priority': 'High',
                'description': '立即將資料庫連接池大小調整到 100，並在所有環境部署',
                'assignee': 'Alice Chen',
                'due_date': time.time() + 86400,  # 1天後
                'status': '進行中'
            },
            {
                'priority': 'High',
                'description': '添加資料庫連接池監控告警，當使用率超過 80% 時告警',
                'assignee': 'Bob Liu',
                'due_date': time.time() + 172800,  # 2天後
                'status': '待處理'
            },
            {
                'priority': 'Medium',
                'description': '優化慢查詢，添加必要的資料庫索引',
                'assignee': 'Carol Wang',
                'due_date': time.time() + 604800,  # 1週後
                'status': '待處理'
            },
            {
                'priority': 'Medium',
                'description': '實現資料庫連接池動態調整機制',
                'assignee': 'David Zhang',
                'due_date': time.time() + 1209600,  # 2週後
                'status': '待處理'
            },
            {
                'priority': 'Low',
                'description': '更新負載測試場景，包含極端高並發情況',
                'assignee': 'Eve Li',
                'due_date': time.time() + 1814400,  # 3週後
                'status': '待處理'
            }
        ],
        'lessons_learned': {
            'positive': [
                '監控系統及時發現了異常',
                'incident response 流程運作良好',
                '團隊協作效率高，快速定位問題',
                '臨時修復方案有效'
            ],
            'improvements': [
                '需要更全面的容量規劃流程',
                '資料庫相關監控需要加強',
                '自動化故障恢復機制需要完善',
                '負載測試覆蓋度需要提升'
            ],
            'process': [
                '建立資料庫連接數的定期審查流程',
                '在部署前必須進行容量評估',
                '定期進行災難恢復演練',
                '建立預防性維護計劃'
            ],
            'technical': [
                '實現資料庫連接池的自動擴縮容',
                '建立更精細的監控指標體系',
                '實現查詢性能的自動分析和優化建議',
                '建立多層級的降級和熔斷機制'
            ]
        },
        'metrics': [
            {'name': 'API 響應時間', 'before': '200ms', 'during': '30s+', 'after': '180ms'},
            {'name': '錯誤率', 'before': '0.1%', 'during': '100%', 'after': '0.05%'},
            {'name': '系統可用性', 'before': '99.9%', 'during': '15%', 'after': '99.95%'},
            {'name': 'DB 連接使用率', 'before': '60%', 'during': '100%', 'after': '25%'}
        ],
        'charts': [
            {
                'title': '事件期間 API 響應時間趨勢',
                'chart_type': 'line',
                'data': {
                    '事件前': {'14:00': 0.2, '14:30': 0.18, '15:00': 0.22},
                    '事件中': {'15:30': 15.0, '16:00': 30.0, '16:30': 25.0},
                    '事件後': {'17:00': 0.18, '17:30': 0.16, '18:00': 0.19}
                }
            },
            {
                'title': '系統錯誤率分布',
                'chart_type': 'pie',
                'data': {
                    '資料庫超時': 60,
                    '連接池耗盡': 25,  
                    '網路錯誤': 10,
                    '其他錯誤': 5
                }
            }
        ],
        'tables': [
            {
                'title': '影響服務詳細信息',
                'headers': ['服務名', '影響程度', '恢復時間', '負責團隊'],
                'rows': [
                    ['user-api', '完全中斷', '2小時', 'Backend Team'],
                    ['payment-api', '完全中斷', '2小時', 'Payment Team'], 
                    ['notification-service', '部分影響', '1.5小時', 'Platform Team']
                ],
                'alignment': ['left', 'center', 'center', 'left']
            }
        ],
        'appendix': {
            'documents': [
                {'title': 'Incident Response Playbook', 'url': 'https://wiki.company.com/incident-response'},
                {'title': 'Database Configuration Guide', 'url': 'https://wiki.company.com/db-config'}
            ],
            'references': [
                'PostgreSQL Connection Pooling Best Practices',
                'High Availability Database Design Patterns',
                'Microservices Monitoring and Alerting Guidelines'
            ],
            'config': '''
database:
  pool:
    maxSize: 100
    minSize: 10
    acquireTimeout: 30000
    idleTimeout: 600000
    maxLifetime: 1800000
  monitoring:
    enablePoolMetrics: true
    alertThreshold: 80
'''
        }
    }
    
    print("📊 測試數據準備完成")
    
    # 測試中文報告生成
    print("📝 生成中文報告...")
    try:
        chinese_report_path = await engine.generate_full_report(
            incident_data,
            language=ReportLanguage.CHINESE,
            output_filename="incident_report_zh.md"
        )
        print(f"✅ 中文報告生成成功: {chinese_report_path}")
    except Exception as e:
        print(f"❌ 中文報告生成失敗: {e}")
        import traceback
        traceback.print_exc()
    
    # 測試英文報告生成  
    print("📝 生成英文報告...")
    try:
        english_report_path = await engine.generate_full_report(
            incident_data,
            language=ReportLanguage.ENGLISH,
            output_filename="incident_report_en.md"
        )
        print(f"✅ 英文報告生成成功: {english_report_path}")
    except Exception as e:
        print(f"❌ 英文報告生成失敗: {e}")
        import traceback
        traceback.print_exc()
    
    # 測試圖表生成
    print("📈 測試獨立圖表生成...")
    try:
        chart_config = ChartConfig(
            title="測試響應時間趨勢",
            chart_type="line",
            data={
                "API-A": {"10:00": 100, "11:00": 150, "12:00": 80},
                "API-B": {"10:00": 200, "11:00": 300, "12:00": 180}
            }
        )
        chart_markdown = await engine.generate_chart(chart_config)
        print(f"✅ 圖表生成成功: {chart_markdown}")
    except Exception as e:
        print(f"❌ 圖表生成失敗: {e}")
    
    # 測試表格生成
    print("📋 測試表格生成...")
    try:
        table_config = TableConfig(
            title="測試服務狀態表",
            headers=["服務", "狀態", "響應時間"],
            rows=[
                ["API Gateway", "正常", "50ms"],
                ["User Service", "警告", "200ms"], 
                ["Payment Service", "錯誤", "超時"]
            ],
            alignment=["left", "center", "right"]
        )
        table_markdown = engine.generate_table(table_config)
        print("✅ 表格生成成功:")
        print(table_markdown)
    except Exception as e:
        print(f"❌ 表格生成失敗: {e}")
    
    print("\n🎉 測試完成！")
    print(f"報告文件保存在: {engine.output_dir}")


async def test_summary_report():
    """測試摘要報告"""
    print("\n📊 測試摘要報告生成...")
    
    engine = PostmortemReportEngine()
    
    # 多個事件數據
    incidents = [
        {
            'severity': 'high',
            'affected_services': ['api-gateway', 'user-service'],
            'duration': 3600,
            'detection_time': time.time() - 86400,
            'resolution_time': time.time() - 82800
        },
        {
            'severity': 'medium',
            'affected_services': ['payment-service'],
            'duration': 1800,
            'detection_time': time.time() - 172800,
            'resolution_time': time.time() - 171000
        },
        {
            'severity': 'low',
            'affected_services': ['notification-service'],
            'duration': 900,
            'detection_time': time.time() - 259200,
            'resolution_time': time.time() - 258300
        }
    ]
    
    try:
        summary = await engine.generate_summary_report(incidents, ReportLanguage.CHINESE)
        print("✅ 摘要報告生成成功")
        
        # 保存摘要報告
        summary_path = engine.output_dir / "summary_report.md"
        with open(summary_path, 'w', encoding='utf-8') as f:
            f.write(summary)
        print(f"📄 摘要報告保存到: {summary_path}")
        
    except Exception as e:
        print(f"❌ 摘要報告生成失敗: {e}")
        import traceback
        traceback.print_exc()


if __name__ == "__main__":
    # 確保必要的目錄存在
    Path("./reports").mkdir(exist_ok=True)
    Path("./templates").mkdir(exist_ok=True)
    
    print("🔧 開始在 Docker 環境中測試報告引擎...")
    print(f"📍 當前工作目錄: {Path.cwd()}")
    print(f"📂 報告輸出目錄: {Path('./reports').resolve()}")
    
    # 運行測試
    asyncio.run(test_report_generation())
    asyncio.run(test_summary_report())
    
    # 顯示生成的文件
    reports_dir = Path("./reports")
    if reports_dir.exists():
        print(f"\n📁 生成的報告文件:")
        for report_file in reports_dir.iterdir():
            if report_file.is_file():
                print(f"  📄 {report_file.name} ({report_file.stat().st_size} bytes)")
    
    print("\n✨ 所有測試完成！")