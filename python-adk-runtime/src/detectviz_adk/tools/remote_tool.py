# -*- coding: utf-8 -*-
from __future__ import annotations
import os
from typing import Any, Dict, Optional

import grpc

try:
    # buf 產碼後的相對匯入（依你的 python out 路徑調整）
    from contracts.gen.python.detectviz.contracts import adk_bridge_pb2 as pb
    from contracts.gen.python.detectviz.contracts import adk_bridge_pb2_grpc as pbg
except Exception as e:  # 開發期允許無產碼
    pb = None
    pbg = None

try:
    # ADK Tool 介面
    from adk.tools import BaseTool  # type: ignore
except Exception:
    class BaseTool:  # fallback，用於樣板測試
        async def invoke(self, payload: Dict[str, Any]) -> Dict[str, Any]:
            raise NotImplementedError

class RemoteTool(BaseTool):
    """
    將 ADK Tool 的 invoke 轉交 ToolBridge.Invoke（gRPC）
    - 端點／mTLS 憑證由 Profiles 注入環境變數，不得硬編碼：
      A2A_ENDPOINT, A2A_CERT_PATH, A2A_KEY_PATH, A2A_CA_PATH
    """

    def __init__(self, tool_id: str, tool_version: str = "0.1.0", timeout_ms: int = 5000):
        self.tool_id = tool_id
        self.tool_version = tool_version
        self.timeout_ms = timeout_ms

        endpoint = os.getenv("A2A_ENDPOINT", "127.0.0.1:6606")
        cert = os.getenv("A2A_CERT_PATH")
        key = os.getenv("A2A_KEY_PATH")
        ca = os.getenv("A2A_CA_PATH")

        if cert and key:
            with open(cert, "rb") as f:
                cert_pem = f.read()
            with open(key, "rb") as f:
                key_pem = f.read()
            root = None
            if ca:
                with open(ca, "rb") as f:
                    root = f.read()
            creds = grpc.ssl_channel_credentials(root_certificates=root, private_key=key_pem, certificate_chain=cert_pem)
            self._channel = grpc.secure_channel(endpoint, creds)
        else:
            # 開發模式允許明文通道；正式環境請務必提供 mTLS
            self._channel = grpc.insecure_channel(endpoint)

        self._stub: Optional[pbg.ToolBridgeStub] = pbg.ToolBridgeStub(self._channel) if pbg else None

    async def invoke(self, payload: Dict[str, Any]) -> Dict[str, Any]:
        if not pb or not self._stub:
            return {"ok": False, "error": "proto stubs not generated; run buf generate first"}

        md = [
            ("tenant_id", payload.get("tenant_id", "")),
            ("owner.root_agent_id", payload.get("owner.root_agent_id", "")),
            ("traceparent", payload.get("traceparent", "")),
        ]
        req = pb.ToolInvokeRequest(
            tool_id=self.tool_id,
            tool_version=self.tool_version,
            payload=pb.google_dot_protobuf_dot_struct__pb2.Struct(fields={}),  # 由 helper 轉換
            timeout_ms=self.timeout_ms,
            metadata={},
        )
        # 將 dict 轉 Struct（簡化版）
        from google.protobuf.struct_pb2 import Struct
        s = Struct()
        s.update(payload)
        req.payload.CopyFrom(s)

        # deadline
        import time
        deadline = time.time() + (self.timeout_ms / 1000.0)
        res: pb.ToolInvokeReply = self._stub.Invoke(req, metadata=md, deadline=deadline)  # type: ignore

        # 映射最小回傳
        result = {}
        if res.result:
            result = dict(res.result)

        return {
            "ok": (getattr(res.status, "code", 2) == 0),
            "status": {"code": getattr(res.status, "code", 2), "message": getattr(res.status, "message", "")},
            "result": result,
            "exec_meta": {
                "attempt": getattr(res.exec_meta, "attempt", 0),
                "duration_ms": getattr(res.exec_meta, "duration_ms", 0),
                "plugin_id": getattr(res.exec_meta, "plugin_id", ""),
                "route_id": getattr(res.exec_meta, "route_id", ""),
            },
        }