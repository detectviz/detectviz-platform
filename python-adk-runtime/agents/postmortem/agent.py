"""
ADK Web 相容的 postmortem agent 入口點
"""
import sys
from pathlib import Path

# 確保能正確匯入 src 目錄下的 agent
sys.path.insert(0, str(Path(__file__).parent.parent.parent / "src"))

from detectviz_adk.agents.postmortem import postmortem_orchestrator

# ADK Web 會自動尋找名為 root_agent 的變數
root_agent = postmortem_orchestrator
agent = postmortem_orchestrator  # 也保留這個以防萬一