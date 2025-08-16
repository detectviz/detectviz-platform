"""
DetectViz ADK Web Server Entry Point
Usage: adk web python-adk-runtime/web_server.py
"""
import sys
from pathlib import Path

# 確保能正確匯入 src 目錄下的 agent
sys.path.insert(0, str(Path(__file__).parent / "src"))

from detectviz_adk.agents.postmortem import postmortem_orchestrator

# `adk web` 會自動尋找名為 `agent` 的變數
agent = postmortem_orchestrator