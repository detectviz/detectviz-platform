"""
測試符合 ADK 規範的報告工具
"""
import pytest
import asyncio
from datetime import datetime, timedelta
from pathlib import Path
import sys
import os

# 添加 src 目錄到 Python 路徑
sys.path.insert(0, os.path.join(os.path.dirname(__file__), '../../src'))

from detectviz_adk.tools.report_tools import (
    generate_postmortem_report_func,
    generate_chart_func,
    PostmortemReportEngine,
    ReportLanguage,
    ChartConfig,
    TableConfig
)


class TestReportTools:
    """測試報告工具功能"""

    @pytest.mark.asyncio
    async def test_generate_postmortem_report_func(self):
        """測試報告生成函數"""
        incident_data = {
            'summary': '測試事件摘要',
            'incident': {
                'severity': 'high',
                'affected_services': ['test-api'],
                'duration': 3600,
                'detection_time': datetime.now() - timedelta(hours=1),
                'resolution_time': datetime.now(),
                'responsible_team': 'Test Team'
            },
            'impact': {
                'customer': '測試客戶影響',
                'business': '測試業務影響',
                'technical': '測試技術影響'
            },
            'timeline': [
                {
                    'timestamp': datetime.now() - timedelta(minutes=30),
                    'title': '測試事件',
                    'severity': 'medium',
                    'description': '測試描述'
                }
            ],
            'root_cause': {
                'primary': '測試根因',
                'contributing': ['因素1', '因素2'],
                'root': '根本原因'
            },
            'action_items': [
                {
                    'priority': 'High',
                    'description': '測試行動項目',
                    'assignee': 'Test Person',
                    'due_date': datetime.now() + timedelta(days=1),
                    'status': '待處理'
                }
            ]
        }
        
        result = await generate_postmortem_report_func(
            incident_data=incident_data,
            language="zh-TW",
            output_filename="test_report.md"
        )
        
        assert result["status"] == "success"
        assert "report_path" in result
        assert "zh-TW" in result["language"]
        
        # 檢查檔案是否存在
        report_path = Path(result["report_path"])
        assert report_path.exists()
        
        # 檢查檔案內容
        content = report_path.read_text(encoding='utf-8')
        assert '測試事件摘要' in content
        assert '事故複盤報告' in content

    @pytest.mark.asyncio
    async def test_generate_chart_func(self):
        """測試圖表生成函數"""
        chart_data = {
            'Series 1': {'A': 10, 'B': 20, 'C': 15},
            'Series 2': {'A': 15, 'B': 25, 'C': 10}
        }
        
        result = await generate_chart_func(
            title="測試圖表",
            chart_type="line",
            data=chart_data,
            width=800,
            height=600
        )
        
        assert result["status"] == "success"
        assert "chart_markdown" in result
        assert "測試圖表" in result["chart_markdown"]
        assert result["chart_type"] == "line"

    def test_postmortem_report_engine_init(self):
        """測試報告引擎初始化"""
        engine = PostmortemReportEngine(
            default_language=ReportLanguage.CHINESE
        )
        
        assert engine.default_language == ReportLanguage.CHINESE
        assert engine.jinja_env is not None
        assert engine._translations is not None

    def test_chart_config(self):
        """測試圖表配置"""
        config = ChartConfig(
            title="測試圖表",
            chart_type="bar",
            data={'A': 10, 'B': 20}
        )
        
        assert config.title == "測試圖表"
        assert config.chart_type == "bar"
        assert config.width == 800  # 默認值
        assert config.height == 600  # 默認值

    def test_table_config(self):
        """測試表格配置"""
        config = TableConfig(
            title="測試表格",
            headers=["列1", "列2"],
            rows=[["數據1", "數據2"]]
        )
        
        assert config.title == "測試表格"
        assert len(config.headers) == 2
        assert len(config.rows) == 1

    def test_report_language_enum(self):
        """測試語言枚舉"""
        assert ReportLanguage.CHINESE.value == "zh-TW"
        assert ReportLanguage.ENGLISH.value == "en-US"
        assert ReportLanguage.SIMPLIFIED_CHINESE.value == "zh-CN"

    def test_translation_system(self):
        """測試翻譯系統"""
        engine = PostmortemReportEngine()
        
        # 測試中文翻譯
        zh_translation = engine.get_translation('incident_report', ReportLanguage.CHINESE)
        assert zh_translation == '事故複盤報告'
        
        # 測試英文翻譯
        en_translation = engine.get_translation('incident_report', ReportLanguage.ENGLISH)
        assert en_translation == 'Incident Postmortem Report'
        
        # 測試未知鍵值
        unknown_translation = engine.get_translation('unknown_key', ReportLanguage.CHINESE)
        assert unknown_translation == 'unknown_key'

    @pytest.mark.asyncio
    async def test_template_rendering(self):
        """測試模板渲染"""
        engine = PostmortemReportEngine()
        
        context = {
            'summary': '測試摘要',
            'incident': {
                'severity': 'high'
            }
        }
        
        # 使用默認模板
        result = await engine.render_template(
            "non_existent_template.md",
            context,
            ReportLanguage.CHINESE
        )
        
        assert '測試摘要' in result
        assert '事故複盤報告' in result

    def test_error_handling(self):
        """測試錯誤處理"""
        # 測試沒有 jinja2 的情況 (這個測試在實際環境中會跳過)
        try:
            import jinja2
        except ImportError:
            with pytest.raises(ImportError):
                PostmortemReportEngine()


@pytest.mark.asyncio
async def test_integration_with_docker_logs():
    """整合測試：使用 Docker 日誌數據"""
    incident_data = {
        'summary': 'Docker 環境測試報告生成',
        'incident': {
            'severity': 'medium',
            'affected_services': ['detectviz-api', 'detectviz-db'],
            'duration': 1800,
            'detection_time': datetime.now() - timedelta(minutes=30),
            'resolution_time': datetime.now(),
            'responsible_team': 'DevOps Team'
        },
        'impact': {
            'customer': 'Docker 測試環境中的模擬影響',
            'business': '無實際業務影響',
            'technical': '測試環境正常運行驗證'
        },
        'charts': [
            {
                'title': 'Docker 測試響應時間',
                'chart_type': 'line',
                'data': {'響應時間': {str(i): i*100 for i in range(5)}},
                'save_path': 'docker_test_chart.png'
            }
        ],
        'tables': [
            {
                'title': 'Docker 服務狀態',
                'headers': ['服務', '狀態'],
                'rows': [['detectviz-api', '運行中'], ['detectviz-db', '運行中']]
            }
        ]
    }
    
    result = await generate_postmortem_report_func(
        incident_data=incident_data,
        language="zh-TW",
        output_filename="docker_integration_test.md"
    )
    
    assert result["status"] == "success"
    assert result["charts_generated"] >= 1
    assert result["tables_generated"] >= 1
    
    # 檢查檔案內容
    report_path = Path(result["report_path"])
    content = report_path.read_text(encoding='utf-8')
    assert 'Docker 環境測試報告生成' in content
    assert 'detectviz-api' in content
    

if __name__ == "__main__":
    # 運行簡單測試
    asyncio.run(test_integration_with_docker_logs())
    print("✅ 報告工具集成測試通過！")