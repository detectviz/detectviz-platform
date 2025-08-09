# -*- coding: utf-8 -*-
"""Sub Agent（本地多實例）樣板。"""
from typing import Any, Dict

try:
    from adk.agents import BaseAgent  # type: ignore
except Exception:
    class BaseAgent:  # type: ignore
        async def run(self, input: Dict[str, Any], ctx: Dict[str, Any]) -> Dict[str, Any]:
            raise NotImplementedError("ADK BaseAgent not available - replace with real imports.")

class SubAgent(BaseAgent):
    def __init__(self, *, tools: Dict[str, Any] | None = None, memory=None, config: Dict[str, Any] | None = None) -> None:
        self._tools = tools or {}
        self._memory = memory
        self._config = config or {}

    async def run(self, input: Dict[str, Any], ctx: Dict[str, Any]) -> Dict[str, Any]:
        return {"ok": True, "echo": input}
