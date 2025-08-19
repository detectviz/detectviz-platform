"""
符合 ADK 規範的報告生成工具
基於 Tool (執行層) 設計原則，專門負責報告生成的具體執行
"""
import os
import json
import asyncio
from datetime import datetime, timezone
from typing import Dict, Any, List, Optional, Union
from pathlib import Path
from dataclasses import dataclass
from enum import Enum
from google.adk.tools import FunctionTool

try:
    import jinja2
    HAS_JINJA2 = True
except ImportError:
    HAS_JINJA2 = False

try:
    import matplotlib
    matplotlib.use('Agg')  # 使用非互動式後端
    import matplotlib.pyplot as plt
    import matplotlib.font_manager as fm
    import pandas as pd
    
    # 配置中文字體支持
    plt.rcParams['font.sans-serif'] = ['SimHei', 'Arial Unicode MS', 'DejaVu Sans']
    plt.rcParams['axes.unicode_minus'] = False  # 解決負號顯示問題
    
    HAS_PLOTTING = True
except ImportError:
    HAS_PLOTTING = False


class ReportLanguage(Enum):
    """報告語言"""
    CHINESE = "zh-TW"
    ENGLISH = "en-US"
    SIMPLIFIED_CHINESE = "zh-CN"


class ReportFormat(Enum):
    """報告格式"""
    MARKDOWN = "markdown"
    HTML = "html"
    PDF = "pdf"


@dataclass
class ChartConfig:
    """圖表配置"""
    title: str
    chart_type: str  # line, bar, pie, scatter
    data: Dict[str, Any]
    width: int = 800
    height: int = 600
    save_path: Optional[str] = None


@dataclass
class TableConfig:
    """表格配置"""
    title: str
    headers: List[str]
    rows: List[List[str]]
    alignment: List[str] = None  # left, center, right
    show_index: bool = False


class PostmortemReportEngine:
    """
    事後複盤報告生成引擎 (Tool 層實現)
    
    功能：
    1. 支援多語言模板 (中文/英文)
    2. Jinja2 模板渲染
    3. 圖表生成和嵌入
    4. 表格格式化
    5. 報告結構化組織
    """

    def __init__(
        self,
        template_dir: Optional[str] = None,
        output_dir: Optional[str] = None,
        default_language: ReportLanguage = ReportLanguage.CHINESE
    ):
        if not HAS_JINJA2:
            raise ImportError("jinja2 is required for report generation")
        
        self.template_dir = Path(template_dir) if template_dir else Path(__file__).parent.parent / "templates"
        self.output_dir = Path(output_dir) if output_dir else Path("./reports")
        self.default_language = default_language
        
        # 確保輸出目錄存在
        self.output_dir.mkdir(parents=True, exist_ok=True)
        
        # 初始化 Jinja2 環境
        self.jinja_env = jinja2.Environment(
            loader=jinja2.FileSystemLoader(str(self.template_dir)),
            autoescape=jinja2.select_autoescape(['html', 'xml']),
            enable_async=True
        )
        
        # 註冊自定義過濾器
        self._register_filters()
        
        # 語言字典
        self._translations = self._load_translations()

    def _register_filters(self) -> None:
        """註冊 Jinja2 自定義過濾器"""
        
        def format_timestamp(timestamp: Union[float, int, str, datetime], lang: str = None) -> str:
            """格式化時間戳"""
            lang = lang or self.default_language.value
            
            # 處理 datetime 對象
            if isinstance(timestamp, datetime):
                dt = timestamp
            elif isinstance(timestamp, str):
                try:
                    timestamp = float(timestamp)
                    dt = datetime.fromtimestamp(timestamp, tz=timezone.utc)
                except ValueError:
                    return str(timestamp)
            else:
                try:
                    dt = datetime.fromtimestamp(float(timestamp), tz=timezone.utc)
                except (ValueError, TypeError):
                    return str(timestamp)
            
            if lang.startswith('zh'):
                return dt.strftime('%Y年%m月%d日 %H:%M:%S UTC')
            else:
                return dt.strftime('%Y-%m-%d %H:%M:%S UTC')
        
        def format_duration(seconds: Union[float, int], lang: str = None) -> str:
            """格式化持續時間"""
            lang = lang or self.default_language.value
            
            if isinstance(seconds, str):
                try:
                    seconds = float(seconds)
                except ValueError:
                    return str(seconds)
            
            hours, remainder = divmod(int(seconds), 3600)
            minutes, seconds = divmod(remainder, 60)
            
            if lang.startswith('zh'):
                if hours > 0:
                    return f"{hours}小時{minutes}分鐘{seconds}秒"
                elif minutes > 0:
                    return f"{minutes}分鐘{seconds}秒"
                else:
                    return f"{seconds}秒"
            else:
                if hours > 0:
                    return f"{hours}h {minutes}m {seconds}s"
                elif minutes > 0:
                    return f"{minutes}m {seconds}s"
                else:
                    return f"{seconds}s"
        
        def format_severity(severity: str, lang: str = None) -> str:
            """格式化嚴重程度"""
            lang = lang or self.default_language.value
            
            severity_map = {
                'critical': {'zh': '嚴重', 'en': 'Critical'},
                'high': {'zh': '高', 'en': 'High'},
                'medium': {'zh': '中等', 'en': 'Medium'},
                'low': {'zh': '低', 'en': 'Low'},
                'info': {'zh': '信息', 'en': 'Info'}
            }
            
            lang_key = 'zh' if lang.startswith('zh') else 'en'
            return severity_map.get(severity.lower(), {}).get(lang_key, severity)
        
        # 註冊過濾器
        self.jinja_env.filters['format_timestamp'] = format_timestamp
        self.jinja_env.filters['format_duration'] = format_duration
        self.jinja_env.filters['format_severity'] = format_severity

    def _load_translations(self) -> Dict[str, Dict[str, str]]:
        """載入翻譯字典"""
        return {
            'zh-TW': {
                'incident_report': '事故複盤報告',
                'executive_summary': '執行摘要',
                'incident_timeline': '事件時間線',
                'impact_analysis': '影響分析',
                'root_cause_analysis': '根因分析',
                'action_items': '行動項目',
                'lessons_learned': '經驗教訓',
                'appendix': '附錄',
                'incident_overview': '事件概覽',
                'affected_services': '影響服務',
                'duration': '持續時間',
                'severity': '嚴重程度',
                'customer_impact': '客戶影響',
                'business_impact': '業務影響',
                'technical_impact': '技術影響',
                'detection_time': '發現時間',
                'resolution_time': '解決時間',
                'responsible_team': '負責團隊',
                'status': '狀態',
                'priority': '優先級',
                'assignee': '負責人',
                'due_date': '截止日期',
                'description': '描述',
                'metrics_chart': '指標圖表',
                'system_metrics': '系統指標',
                'error_rates': '錯誤率',
                'response_times': '響應時間',
                'resource_usage': '資源使用',
                'generated_at': '生成時間',
                'report_version': '報告版本',
                'incident_summary_report': '事件摘要報告',
                'overview': '概覽',
                'total_incidents': '總事件數',
                'total_downtime': '總停機時間',
                'avg_resolution_time': '平均解決時間',
                'severity_breakdown': '嚴重程度分布',
                'severity_distribution': '嚴重程度分佈',
                'service_impact': '服務影響',
                'most_affected_services': '受影響最大的服務',
                'incidents': '事件',
                'language': '語言'
            },
            'en-US': {
                'incident_report': 'Incident Postmortem Report',
                'executive_summary': 'Executive Summary',
                'incident_timeline': 'Incident Timeline',
                'impact_analysis': 'Impact Analysis',
                'root_cause_analysis': 'Root Cause Analysis',
                'action_items': 'Action Items',
                'lessons_learned': 'Lessons Learned',
                'appendix': 'Appendix',
                'incident_overview': 'Incident Overview',
                'affected_services': 'Affected Services',
                'duration': 'Duration',
                'severity': 'Severity',
                'customer_impact': 'Customer Impact',
                'business_impact': 'Business Impact',
                'technical_impact': 'Technical Impact',
                'detection_time': 'Detection Time',
                'resolution_time': 'Resolution Time',
                'responsible_team': 'Responsible Team',
                'status': 'Status',
                'priority': 'Priority',
                'assignee': 'Assignee',
                'due_date': 'Due Date',
                'description': 'Description',
                'metrics_chart': 'Metrics Chart',
                'system_metrics': 'System Metrics',
                'error_rates': 'Error Rates',
                'response_times': 'Response Times',
                'resource_usage': 'Resource Usage',
                'generated_at': 'Generated At',
                'report_version': 'Report Version',
                'incident_summary_report': 'Incident Summary Report',
                'overview': 'Overview',
                'total_incidents': 'Total Incidents',
                'total_downtime': 'Total Downtime',
                'avg_resolution_time': 'Average Resolution Time',
                'severity_breakdown': 'Severity Breakdown',
                'severity_distribution': 'Severity Distribution',
                'service_impact': 'Service Impact',
                'most_affected_services': 'Most Affected Services',
                'incidents': 'Incidents',
                'language': 'Language'
            }
        }

    def get_translation(self, key: str, language: ReportLanguage = None) -> str:
        """獲取翻譯"""
        lang = language or self.default_language
        return self._translations.get(lang.value, {}).get(key, key)

    async def generate_chart(self, config: ChartConfig) -> str:
        """生成圖表並返回文件路徑"""
        if not HAS_PLOTTING:
            return f"![{config.title}](chart_not_available.png)"
        
        try:
            # 設置中文字體
            plt.rcParams['font.sans-serif'] = ['SimHei', 'Arial Unicode MS', 'DejaVu Sans']
            plt.rcParams['axes.unicode_minus'] = False
            
            fig, ax = plt.subplots(figsize=(config.width/100, config.height/100))
            
            if config.chart_type == "line":
                for series_name, series_data in config.data.items():
                    if isinstance(series_data, dict):
                        x_data = list(series_data.keys())
                        y_data = list(series_data.values())
                        ax.plot(x_data, y_data, label=series_name, marker='o')
                    
            elif config.chart_type == "bar":
                if isinstance(config.data, dict):
                    keys = list(config.data.keys())
                    values = list(config.data.values())
                    ax.bar(keys, values)
                    
            elif config.chart_type == "pie":
                if isinstance(config.data, dict):
                    labels = list(config.data.keys())
                    sizes = list(config.data.values())
                    ax.pie(sizes, labels=labels, autopct='%1.1f%%')
            
            ax.set_title(config.title)
            ax.legend()
            
            # 保存圖表
            if config.save_path:
                chart_path = self.output_dir / config.save_path
            else:
                chart_path = self.output_dir / f"chart_{hash(config.title)}.png"
                
            chart_path.parent.mkdir(parents=True, exist_ok=True)
            plt.savefig(chart_path, dpi=100, bbox_inches='tight')
            plt.close()
            
            # 返回相對路徑
            relative_path = chart_path.name
            return f"![{config.title}]({relative_path})"
            
        except Exception as e:
            plt.close('all')  # 確保清理
            print(f"Chart generation error: {str(e)}")
            return f"*圖表生成失敗: {str(e)}*"

    def generate_table(self, config: TableConfig, language: ReportLanguage = None) -> str:
        """生成 Markdown 表格"""
        if not config.headers or not config.rows:
            return ""
        
        lines = []
        
        # 標題
        if config.title:
            lines.append(f"### {config.title}")
            lines.append("")
        
        # 表頭
        header_line = "| " + " | ".join(config.headers) + " |"
        lines.append(header_line)
        
        # 分隔線
        if config.alignment:
            separator_parts = []
            for align in config.alignment:
                if align == "center":
                    separator_parts.append(":---:")
                elif align == "right":
                    separator_parts.append("---:")
                else:
                    separator_parts.append("---")
        else:
            separator_parts = ["---"] * len(config.headers)
        
        separator_line = "| " + " | ".join(separator_parts) + " |"
        lines.append(separator_line)
        
        # 數據行
        for i, row in enumerate(config.rows):
            if config.show_index:
                row_data = [str(i+1)] + [str(cell) for cell in row]
            else:
                row_data = [str(cell) for cell in row]
            
            # 確保行長度與表頭一致
            while len(row_data) < len(config.headers):
                row_data.append("")
            
            row_line = "| " + " | ".join(row_data[:len(config.headers)]) + " |"
            lines.append(row_line)
        
        lines.append("")  # 表格後的空行
        return "\\n".join(lines)

    async def render_template(
        self,
        template_name: str,
        context: Dict[str, Any],
        language: ReportLanguage = None
    ) -> str:
        """渲染模板"""
        lang = language or self.default_language
        
        # 添加語言和翻譯到上下文
        def safe_translation(key: str) -> str:
            """安全的翻譯函數，提供回退機制"""
            try:
                return self.get_translation(key, lang)
            except Exception as e:
                print(f"Translation error for key '{key}': {e}")
                return key
        
        context.update({
            'language': lang.value,
            'translations': self._translations.get(lang.value, {}),
            't': safe_translation,
            'now': datetime.now(),
            'report_meta': {
                'generated_at': datetime.now().isoformat(),
                'language': lang.value,
                'version': '1.0.0'
            }
        })
        
        try:
            template = self.jinja_env.get_template(template_name)
            return await template.render_async(**context)
        except jinja2.TemplateNotFound:
            # 嘗試使用默認模板
            default_template = self._create_default_template()
            template = self.jinja_env.from_string(default_template)
            return await template.render_async(**context)

    def _create_default_template(self) -> str:
        """創建默認模板"""
        return """# {{ t('incident_report') }}

## {{ t('executive_summary') }}

{{ summary | default('暫無摘要') }}

## {{ t('incident_overview') }}

- **{{ t('severity') }}**: {{ incident.severity | format_severity(language) | default('未知') }}
- **{{ t('duration') }}**: {{ incident.duration | format_duration(language) | default('未知') }}
- **{{ t('detection_time') }}**: {{ incident.detection_time | format_timestamp(language) | default('未知') }}
- **{{ t('resolution_time') }}**: {{ incident.resolution_time | format_timestamp(language) | default('未知') }}

## {{ t('impact_analysis') }}

### {{ t('customer_impact') }}
{{ impact.customer | default('暫無數據') }}

### {{ t('business_impact') }}  
{{ impact.business | default('暫無數據') }}

### {{ t('technical_impact') }}
{{ impact.technical | default('暫無數據') }}

## {{ t('incident_timeline') }}

{% for event in timeline | default([]) %}
- **{{ event.timestamp | format_timestamp(language) }}**: {{ event.description }}
{% endfor %}

## {{ t('root_cause_analysis') }}

{{ root_cause | default('根因分析待補充') }}

## {{ t('action_items') }}

{% for item in action_items | default([]) %}
- [ ] **{{ item.title }}** ({{ t('priority') }}: {{ item.priority }}, {{ t('assignee') }}: {{ item.assignee }})
  - {{ item.description }}
{% endfor %}

## {{ t('lessons_learned') }}

{{ lessons_learned | default('經驗教訓待補充') }}

---

*{{ t('generated_at') }}: {{ report_meta.generated_at }}*  
*{{ t('report_version') }}: {{ report_meta.version }}*
"""


# ============= ADK Function Tools =============

async def generate_postmortem_report_func(
    incident_data: Dict[str, Any],
    language: str = "zh-TW",
    output_filename: Optional[str] = None
) -> Dict[str, Any]:
    """
    生成事後複盤報告
    
    Args:
        incident_data: 事件數據
        language: 語言設置 ("zh-TW", "en-US")
        output_filename: 輸出檔案名
        
    Returns:
        報告生成結果
    """
    try:
        engine = PostmortemReportEngine(
            default_language=ReportLanguage(language)
        )
        
        # 處理圖表
        if 'charts' in incident_data:
            for chart_config in incident_data['charts']:
                if isinstance(chart_config, dict):
                    config = ChartConfig(**chart_config)
                    chart_markdown = await engine.generate_chart(config)
                    incident_data.setdefault('chart_embeddings', []).append(chart_markdown)
        
        # 處理表格
        if 'tables' in incident_data:
            for table_config in incident_data['tables']:
                if isinstance(table_config, dict):
                    config = TableConfig(**table_config)
                    table_markdown = engine.generate_table(config, ReportLanguage(language))
                    incident_data.setdefault('table_embeddings', []).append(table_markdown)
        
        # 渲染報告
        report_content = await engine.render_template("postmortem_template.md", incident_data, ReportLanguage(language))
        
        # 保存到文件
        if output_filename:
            output_path = engine.output_dir / output_filename
        else:
            timestamp = datetime.now().strftime("%Y%m%d_%H%M%S")
            output_path = engine.output_dir / f"postmortem_report_{timestamp}_{language}.md"
        
        output_path.parent.mkdir(parents=True, exist_ok=True)
        
        with open(output_path, 'w', encoding='utf-8') as f:
            f.write(report_content)
        
        return {
            "status": "success",
            "report_path": str(output_path),
            "language": language,
            "charts_generated": len(incident_data.get('chart_embeddings', [])),
            "tables_generated": len(incident_data.get('table_embeddings', []))
        }
        
    except Exception as e:
        return {
            "status": "error",
            "error_message": str(e)
        }


async def generate_chart_func(
    title: str,
    chart_type: str,
    data: Dict[str, Any],
    width: int = 800,
    height: int = 600,
    save_path: Optional[str] = None
) -> Dict[str, Any]:
    """
    生成圖表
    
    Args:
        title: 圖表標題
        chart_type: 圖表類型 (line, bar, pie, scatter)
        data: 圖表數據
        width: 寬度
        height: 高度
        save_path: 保存路徑
        
    Returns:
        圖表生成結果
    """
    try:
        engine = PostmortemReportEngine()
        config = ChartConfig(
            title=title,
            chart_type=chart_type,
            data=data,
            width=width,
            height=height,
            save_path=save_path
        )
        
        chart_markdown = await engine.generate_chart(config)
        
        return {
            "status": "success",
            "chart_markdown": chart_markdown,
            "chart_type": chart_type
        }
        
    except Exception as e:
        return {
            "status": "error",
            "error_message": str(e)
        }


# ============= ADK Tools 註冊 =============

# 註冊為 ADK Function Tools
generate_postmortem_report = FunctionTool(generate_postmortem_report_func)
generate_chart = FunctionTool(generate_chart_func)