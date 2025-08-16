"""
DetectViz ADK Runtime - 符合 Google ADK 標準的智慧代理架構
基於 adk_tutorial.ipynb 模式實作的事後檢討系統
"""

# 主要元件匯出
from .agents.postmortem.orchestrator import postmortem_orchestrator
from .agents.postmortem.data_collector import data_collector_agent
from .agents.postmortem.analyzer import root_cause_analyzer
from .agents.postmortem.report_writer import report_writer

from .runners.postmortem_runner import PostmortemRunner, run_postmortem_analysis
from .sessions.session_manager import StateAwareSessionManager, PostmortemToolContext

from .tools.adk_tools import (
    get_health_metrics,
    generate_report,
    create_dashboard,
    update_knowledge_base
)

from .tools.memory_tools import (
    remember_analysis,
    recall_analysis,
    store_postmortem_response,
    get_postmortem_history,
    clear_session_memory
)

__version__ = "1.0.0"
__author__ = "DetectViz Team"
__description__ = "符合 ADK 標準的事後檢討智慧代理系統"

# 便利的匯出清單
__all__ = [
    # 代理
    "postmortem_orchestrator",
    "data_collector_agent", 
    "root_cause_analyzer",
    "report_writer",
    
    # 執行器
    "PostmortemRunner",
    "run_postmortem_analysis",
    
    # 會話管理
    "StateAwareSessionManager",
    "PostmortemToolContext",
    
    # 工具
    "get_health_metrics",
    "generate_report", 
    "create_dashboard",
    "update_knowledge_base",
    
    # 記憶體工具
    "remember_analysis",
    "recall_analysis",
    "store_postmortem_response",
    "get_postmortem_history",
    "clear_session_memory",
]
