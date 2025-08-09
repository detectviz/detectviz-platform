# -*- coding: utf-8 -*-
"""Agent-as-a-Tool 樣板。"""
from typing import Any, Dict

try:
    from adk.tools import BaseTool  # type: ignore
except Exception:
    class BaseTool:  # type: ignore
        async def invoke(self, payload: Dict[str, Any]) -> Dict[str, Any]:
            raise NotImplementedError("ADK BaseTool not available - replace with real imports.")

class AgentTool(BaseTool):
    def __init__(self, *, factory=None, config: Dict[str, Any] | None = None) -> None:
        self._factory = factory
        self._config = config or {}

    async def invoke(self, payload: Dict[str, Any]) -> Dict[str, Any]:
        return {"ok": True, "echo": payload}
