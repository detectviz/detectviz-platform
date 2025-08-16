"""
符合 ADK 標準的遠端工具實作
透過 gRPC ToolBridge 與 Go Platform 通訊的 RemoteTool
基於 DetectViz 平台的 SSOT（Single Source of Truth）設計
"""
# -*- coding: utf-8 -*-
from __future__ import annotations

import os
import asyncio
from typing import Any, Dict, Optional, List, Tuple

import grpc
from google.protobuf.struct_pb2 import Struct
from google.protobuf import json_format

# ---- Protobuf 存根（容錯匯入：優先 v1 命名空間） ------------------------------
try:  # buf 產生後的相對匯入（依據 python 輸出路徑調整）
    from contracts.gen.python.detectviz.contracts.v1 import adk_bridge_pb2 as pb  # type: ignore
    from contracts.gen.python.detectviz.contracts.v1 import adk_bridge_pb2_grpc as pbg  # type: ignore
except Exception:
    try:
        # 舊路徑向後相容
        from contracts.gen.python.detectviz.contracts import adk_bridge_pb2 as pb  # type: ignore
        from contracts.gen.python.detectviz.contracts import adk_bridge_pb2_grpc as pbg  # type: ignore
    except Exception:
        pb = None  # type: ignore
        pbg = None  # type: ignore

# ---- ADK 工具介面容錯 ----------------------------------------------------
try:
    from google.adk.tools.base_tool import BaseTool  # 使用 ADK 的 BaseTool
except Exception:  # 容錯處理，用於樣板/測試
    class BaseTool:  # type: ignore
        async def invoke(self, payload: Dict[str, Any]) -> Dict[str, Any]:  # pragma: no cover
            raise NotImplementedError

# ---- 設定：對齊 DetectViz SSOT ------------------------------------------
try:
    from detectviz_adk.config.loader import get_toolbridge_addr  # 統一以 DETECTVIZ_TOOLBRIDGE_ADDR 為主
except Exception:
    def get_toolbridge_addr(default: str = "127.0.0.1:5002") -> str:  # 後備功能
        return os.getenv("DETECTVIZ_TOOLBRIDGE_ADDR", default)

# ---- 可選：從 OpenTelemetry Context 注入 traceparent/tracestate --------------------
try:
    from opentelemetry.propagate import inject
    from opentelemetry.trace.propagation.tracecontext import TraceContextTextMapPropagator
    _HAS_OTEL = True
except Exception:
    _HAS_OTEL = False


class RemoteTool(BaseTool):
    """
    符合 ADK 標準的遠端工具
    透過 ToolBridge.Invoke（gRPC）呼叫 Go Platform 端的工具

    設定來源（優先序）：
    - 端點：`DETECTVIZ_TOOLBRIDGE_ADDR`，預設 127.0.0.1:5002
    - 安全性：
      - `DETECTVIZ_TOOLBRIDGE_TLS_CERT`/`DETECTVIZ_TOOLBRIDGE_TLS_KEY`/`DETECTVIZ_TOOLBRIDGE_TLS_CA`
      - `DETECTVIZ_TOOLBRIDGE_INSECURE`（true/false）
    - 時限：`timeout_ms` 以建構參數為主

    特色：
    - 本類別使用 `grpc.aio`，避免阻塞 ADK 的事件迴圈
    - metadata 會攜帶 `tenant_id`、`owner.root_agent_id`、`traceparent`/`tracestate`（若可取得）
    - 完全相容於 ADK FunctionTool 生態系統
    """

    def __init__(self, tool_id: str, tool_version: str = "0.1.0", timeout_ms: int = 5000) -> None:
        self.tool_id = tool_id
        self.tool_version = tool_version
        self.timeout_ms = timeout_ms

        self._channel: Optional[grpc.aio.Channel] = None
        self._stub: Optional[pbg.ToolBridgeStub] = None  # type: ignore

        self._init_channel_and_stub()

    # ---- 連線初始化 -------------------------------------------------------
    def _init_channel_and_stub(self) -> None:
        if not pbg:
            return

        addr = os.getenv("DETECTVIZ_TOOLBRIDGE_ADDR") or get_toolbridge_addr()
        insecure = _env_bool("DETECTVIZ_TOOLBRIDGE_INSECURE", default=False)

        cert = os.getenv("DETECTVIZ_TOOLBRIDGE_TLS_CERT")
        key = os.getenv("DETECTVIZ_TOOLBRIDGE_TLS_KEY")
        ca = os.getenv("DETECTVIZ_TOOLBRIDGE_TLS_CA")

        if cert and key:  # 優先 mTLS/單向 TLS
            with open(cert, "rb") as f:
                cert_pem = f.read()
            with open(key, "rb") as f:
                key_pem = f.read()
            root = None
            if ca:
                with open(ca, "rb") as f:
                    root = f.read()
            creds = grpc.ssl_channel_credentials(root_certificates=root, private_key=key_pem, certificate_chain=cert_pem)
            self._channel = grpc.aio.secure_channel(addr, creds)
        elif insecure:
            self._channel = grpc.aio.insecure_channel(addr)
        else:
            # 無憑證但也未明確允許明文 → 開發環境預設允許，生產環境請設定憑證或 INSECURE=true
            self._channel = grpc.aio.insecure_channel(addr)

        self._stub = pbg.ToolBridgeStub(self._channel)  # type: ignore

    # ---- 公開 API ---------------------------------------------------------
    async def invoke(self, payload: Dict[str, Any]) -> Dict[str, Any]:
        """呼叫遠端 Go Platform 工具"""
        if not pb or not self._stub:
            return {"ok": False, "error": "protobuf 存根未產生；請先執行 buf generate"}

        # 轉換 payload → Struct（保留使用者傳入的所有鍵值）
        s = Struct()
        s.update(payload)

        req = pb.ToolInvokeRequest(
            tool_id=self.tool_id,
            tool_version=self.tool_version,
            payload=s,
            timeout_ms=self.timeout_ms,
            metadata={},  # 可選：未來擴充為規範化字典欄位
        )

        # 構造 metadata（header），維持向後相容的鍵名
        md = self._build_metadata(payload)

        try:
            res: pb.ToolInvokeReply = await self._stub.Invoke(
                req,
                metadata=md,  # type: ignore[arg-type]
                timeout=self.timeout_ms / 1000.0,
            )
        except grpc.aio.AioRpcError as e:
            code = getattr(e.code(), "value", (2, "UNKNOWN"))[0]  # type: ignore
            return {
                "ok": False,
                "status": {"code": code, "message": e.details() or str(e)},
                "result": {},
                "exec_meta": {"attempt": 1, "duration_ms": 0, "plugin_id": "", "route_id": ""},
                "error": e.details() or str(e),
            }

        result: Dict[str, Any] = {}
        if getattr(res, "result", None):
            result = json_format.MessageToDict(res.result, preserving_proto_field_name=True)

        return {
            "ok": (getattr(res.status, "code", 2) == 0),
            "status": {
                "code": getattr(res.status, "code", 2),
                "message": getattr(res.status, "message", ""),
            },
            "result": result,
            "exec_meta": {
                "attempt": getattr(res.exec_meta, "attempt", 0),
                "duration_ms": getattr(res.exec_meta, "duration_ms", 0),
                "plugin_id": getattr(res.exec_meta, "plugin_id", ""),
                "route_id": getattr(res.exec_meta, "route_id", ""),
            },
        }

    async def aclose(self) -> None:
        """關閉 gRPC 通道"""
        ch = getattr(self, "_channel", None)
        if ch is not None:
            try:
                await ch.close()
            except Exception:
                pass

    def __del__(self) -> None:  # 盡力而為，避免事件迴圈阻塞
        ch = getattr(self, "_channel", None)
        if ch is None:
            return
        try:
            loop = asyncio.get_event_loop()
            if loop.is_running():
                loop.create_task(ch.close())
            else:
                loop.run_until_complete(ch.close())
        except Exception:
            pass

    # ---- 輔助方法 ----------------------------------------------------------
    def _build_metadata(self, payload: Dict[str, Any]) -> List[Tuple[str, str]]:
        """建構 gRPC metadata"""
        md: List[Tuple[str, str]] = []

        # 與 Go 端現有橋接邏輯相容的鍵名
        for k in ("tenant_id", "owner.root_agent_id"):
            v = payload.get(k)
            if isinstance(v, str) and v:
                md.append((k, v))

        # traceparent/tracestate：payload 優先；否則嘗試自動注入
        tp = str(payload.get("traceparent", "") or "")
        ts = str(payload.get("tracestate", "") or "")

        if not tp and _HAS_OTEL:
            carrier: Dict[str, str] = {}
            try:
                # 以 W3C Trace Context 注入目前的 span
                TraceContextTextMapPropagator().inject(carrier)  # type: ignore[name-defined]
                tp = carrier.get("traceparent", tp)
                ts = carrier.get("tracestate", ts)
            except Exception:
                pass

        if tp:
            md.append(("traceparent", tp))
        if ts:
            md.append(("tracestate", ts))

        return md


# ---- 本地輔助函數 --------------------------------------------------------
def _env_bool(key: str, default: bool = False) -> bool:
    """將環境變數轉換為布林值"""
    v = os.getenv(key)
    if v is None:
        return default
    s = v.strip().lower()
    return s in ("1", "true", "t", "yes", "y", "on")