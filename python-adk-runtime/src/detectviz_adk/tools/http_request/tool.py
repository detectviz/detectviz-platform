# -*- coding: utf-8 -*-
from __future__ import annotations
from typing import Any, Dict

from detectviz_adk.tools.remote_tool import RemoteTool

class HttpRequestTool(RemoteTool):
    """
    detectviz.tools.http_request 的最小包裝
    """
    def __init__(self, version: str = "0.1.0", timeout_ms: int = 5000):
        super().__init__(tool_id="detectviz.tools.http_request", tool_version=version, timeout_ms=timeout_ms)

    async def invoke(self, payload: Dict[str, Any]) -> Dict[str, Any]:
        # 可於此加入輸入檢查 / 白名單校驗（method/url/header）
        return await super().invoke(payload)