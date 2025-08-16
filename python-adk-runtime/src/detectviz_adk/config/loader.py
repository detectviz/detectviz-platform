# Configuration loader for Python ADK Runtime
# 對齊 Detectviz SSOT：統一搜尋順序、環境變數覆蓋、Schema 驗證
# 生效優先序（高→低）：
#  1. 函式參數 path / CLI --config
#  2. 環境變數 DETECTVIZ_CONFIG_FILE
#  3. `./config.yaml`（工作目錄）
#  4. `../contracts/config.yaml`（團隊覆蓋，可選）
#  5. `./contracts/samples/config.yaml`（SSOT 樣本兜底）

from __future__ import annotations

import os
import json
import yaml
import jsonschema
import logging
from typing import Dict, Any, Optional, List, Tuple
from pathlib import Path

logger = logging.getLogger(__name__)

# ---- Public API -----------------------------------------------------------

def load_config(path: Optional[str] = None) -> Dict[str, Any]:
    """載入 Detectviz 統一設定並通過 Schema 驗證。

    1) 解析設定檔（依優先序尋找）
    2) 套用預設值（僅補缺，不覆蓋）
    3) 環境變數覆蓋（DETECTVIZ__* 鍵位對齊 Go）
    4.以 contracts/schemas/config.schema.json 驗證
    """
    resolved, tried = _resolve_config_path(path)

    config: Dict[str, Any] = {}
    if resolved:
        with open(resolved, "r", encoding="utf-8") as f:
            config = yaml.safe_load(f) or {}
        logger.info("Loaded configuration", extra={"config.path": str(resolved)})
    else:
        logger.warning("No config file found; using defaults only", extra={"tried": tried})

    config = _apply_defaults(config)
    _apply_env_overrides(config)

    _validate_config(config)

    # 脫敏摘要輸出
    logger.info(
        "Configuration effective",
        extra={
            "env": config.get("env"),
            "observability.mode": _dget(config, "observability.mode"),
            "otlp.protocol": _dget(config, "observability.otlp.protocol"),
            "otlp.endpoint": _dget(config, "observability.otlp.endpoint"),
        },
    )

    return config


def get_toolbridge_addr(default: str = "127.0.0.1:5002") -> str:
    """取得 Python RemoteTool 連線位址（不寫入 SSOT Config，以環境變數為主）。"""
    return os.getenv("DETECTVIZ_TOOLBRIDGE_ADDR", default)


# ---- Internal helpers -----------------------------------------------------

_DEF: Dict[str, Any] = {
    "env": "development",
    "observability": {
        "mode": "lgtm_local",
        "otlp": {
            "protocol": "grpc",
            "endpoint": "127.0.0.1:4317",
            "insecure": True,
            "headers": {},
        },
        # Python 僅傳遞 span context；logs/pprof 由 Go + Alloy 負責
        "logs": {"mode": "off", "file": {"path": ""}},
        "profiling": {"enabled": False, "pprof_address": "", "application_name": "", "tags": {}},
        "resource": {
            "service_name": "detectviz-adk-runtime",
            "service_version": "0.1.0",
            "environment": "dev",
        },
        "sampling": {"ratio": 1.0},
    },
    "plugin": {"paths": [], "registry": ""},
    "memory": {"backend": "inmem", "dsn": "", "default_ttl_seconds": 3600},
}


def _apply_defaults(cfg: Dict[str, Any]) -> Dict[str, Any]:
    return _deep_merge(_DEF, cfg)


def _apply_env_overrides(cfg: Dict[str, Any]) -> None:
    """以 DETECTVIZ_* 規範覆蓋設定（就地修改）。

    對齊 Go 端鍵位：
      - DETECTVIZ_ENV → env
      - DETECTVIZ__OBSERVABILITY__MODE → observability.mode
      - DETECTVIZ__OBSERVABILITY__OTLP__PROTOCOL → observability.otlp.protocol
      - DETECTVIZ__OBSERVABILITY__OTLP__ENDPOINT → observability.otlp.endpoint
      - DETECTVIZ__OBSERVABILITY__OTLP__INSECURE → observability.otlp.insecure
      - DETECTVIZ__OBSERVABILITY__OTLP__HEADERS → observability.otlp.headers (k=v,k2=v2)
      - DETECTVIZ__OBSERVABILITY__RESOURCE__SERVICE_NAME → observability.resource.service_name
      - DETECTVIZ__OBSERVABILITY__RESOURCE__SERVICE_VERSION → observability.resource.service_version
      - DETECTVIZ__OBSERVABILITY__RESOURCE__ENVIRONMENT → observability.resource.environment
      - DETECTVIZ__OBSERVABILITY__SAMPLING__RATIO → observability.sampling.ratio
      - DETECTVIZ__PLUGIN__PATHS → plugin.paths (csv)
      - DETECTVIZ__PLUGIN__REGISTRY → plugin.registry
      - DETECTVIZ__MEMORY__BACKEND → memory.backend
      - DETECTVIZ__MEMORY__DSN → memory.dsn
      - DETECTVIZ__MEMORY__DEFAULT_TTL_SECONDS → memory.default_ttl_seconds
    """

    mapping = {
        "DETECTVIZ_ENV": ("env", _as_str),
        # Observability / OTLP
        "DETECTVIZ__OBSERVABILITY__MODE": ("observability.mode", _as_str),
        "DETECTVIZ__OBSERVABILITY__OTLP__PROTOCOL": ("observability.otlp.protocol", _as_str),
        "DETECTVIZ__OBSERVABILITY__OTLP__ENDPOINT": ("observability.otlp.endpoint", _as_str),
        "DETECTVIZ__OBSERVABILITY__OTLP__INSECURE": ("observability.otlp.insecure", _as_bool),
        "DETECTVIZ__OBSERVABILITY__OTLP__HEADERS": ("observability.otlp.headers", _as_headers),
        # Resource & Sampling
        "DETECTVIZ__OBSERVABILITY__RESOURCE__SERVICE_NAME": ("observability.resource.service_name", _as_str),
        "DETECTVIZ__OBSERVABILITY__RESOURCE__SERVICE_VERSION": ("observability.resource.service_version", _as_str),
        "DETECTVIZ__OBSERVABILITY__RESOURCE__ENVIRONMENT": ("observability.resource.environment", _as_str),
        "DETECTVIZ__OBSERVABILITY__SAMPLING__RATIO": ("observability.sampling.ratio", _as_float),
        # Plugin & Memory
        "DETECTVIZ__PLUGIN__PATHS": ("plugin.paths", _as_csv),
        "DETECTVIZ__PLUGIN__REGISTRY": ("plugin.registry", _as_str),
        "DETECTVIZ__MEMORY__BACKEND": ("memory.backend", _as_str),
        "DETECTVIZ__MEMORY__DSN": ("memory.dsn", _as_str),
        "DETECTVIZ__MEMORY__DEFAULT_TTL_SECONDS": ("memory.default_ttl_seconds", _as_int),
    }

    for env, (keypath, caster) in mapping.items():
        val = os.getenv(env)
        if val is None or val == "":
            continue
        try:
            casted = caster(val)
            _dset(cfg, keypath, casted)
            logger.info("Env override", extra={"env": env, "key": keypath})
        except Exception as e:
            logger.warning("Failed to apply env override", extra={"env": env, "err": str(e)})


def _validate_config(cfg: Dict[str, Any]) -> None:
    schema_path = _resolve_schema_path()
    if not schema_path:
        logger.warning("Configuration schema not found, skip validation")
        return

    with open(schema_path, "r", encoding="utf-8") as f:
        schema = json.load(f)

    try:
        jsonschema.Draft202012Validator(schema).validate(cfg)
        logger.info("Configuration validated", extra={"schema": str(schema_path)})
    except jsonschema.ValidationError as e:
        raise ValueError(f"Invalid configuration: {e.message}") from e


# ---- Path resolvers -------------------------------------------------------

def _resolve_config_path(arg: Optional[str]) -> Tuple[Optional[Path], List[str]]:
    tried: List[str] = []

    # 1) explicit arg
    if arg and str(arg).strip():
        p = Path(arg)
        tried.append(str(p))
        if p.exists() and p.is_file():
            return p, tried
        return None, tried

    # 2) env
    env_path = os.getenv("DETECTVIZ_CONFIG_FILE")
    if env_path:
        p = Path(env_path)
        tried.append(str(p))
        if p.exists() and p.is_file():
            return p, tried

    # 3) ./config.yaml
    for cand in (Path("./config.yaml"), Path("config.yaml")):
        tried.append(str(cand))
        if cand.exists() and cand.is_file():
            return cand, tried

    # 4../contracts/config.yaml
    for cand in (Path("./contracts/config.yaml"), Path("contracts/config.yaml")):
        tried.append(str(cand))
        if cand.exists() and cand.is_file():
            return cand, tried

    # 5) ./contracts/samples/config.yaml
    for cand in (
        Path("./contracts/samples/config.yaml"),
        Path("contracts/samples/config.yaml"),
    ):
        tried.append(str(cand))
        if cand.exists() and cand.is_file():
            return cand, tried

    return None, tried


def _resolve_schema_path() -> Optional[Path]:
    # 掃描常見位置（以 repo 根為準）
    candidates = [
        Path("./contracts/schemas/config.schema.json"),
        Path("contracts/schemas/config.schema.json"),
        Path(__file__).resolve().parents[4] / "contracts/schemas/config.schema.json",
        Path(__file__).resolve().parents[3] / "contracts/schemas/config.schema.json",
    ]
    for p in candidates:
        if p.exists():
            return p
    return None


# ---- Generic utils --------------------------------------------------------

def _deep_merge(base: Dict[str, Any], override: Dict[str, Any]) -> Dict[str, Any]:
    out = dict(base)
    for k, v in (override or {}).items():
        if isinstance(v, dict) and isinstance(out.get(k), dict):
            out[k] = _deep_merge(out[k], v)
        else:
            out[k] = v
    return out


def _dget(d: Dict[str, Any], dotted: str, default: Any = None) -> Any:
    cur = d
    for part in dotted.split("."):
        if not isinstance(cur, dict) or part not in cur:
            return default
        cur = cur[part]
    return cur


def _dset(d: Dict[str, Any], dotted: str, value: Any) -> None:
    cur = d
    parts = dotted.split(".")
    for p in parts[:-1]:
        if p not in cur or not isinstance(cur[p], dict):
            cur[p] = {}
        cur = cur[p]
    cur[parts[-1]] = value


# casters

def _as_str(v: str) -> str:
    return str(v)


def _as_int(v: str) -> int:
    return int(v.strip())


def _as_float(v: str) -> float:
    return float(v.strip())


def _as_bool(v: str) -> bool:
    s = v.strip().lower()
    if s in ("1", "true", "t", "yes", "y", "on"):
        return True
    if s in ("0", "false", "f", "no", "n", "off"):
        return False
    raise ValueError(f"invalid bool: {v}")


def _as_csv(v: str) -> list:
    return [p.strip() for p in v.split(",") if p.strip()]


def _as_headers(v: str) -> Dict[str, str]:
    # 例如："authorization=Bearer xxx, x-foo=bar"
    out: Dict[str, str] = {}
    for part in v.split(","):
        part = part.strip()
        if not part:
            continue
        if "=" not in part:
            continue
        k, val = part.split("=", 1)
        k = k.strip()
        val = val.strip()
        if k:
            out[k] = val
    return out


# ---- Emergency minimal ----------------------------------------------------

def _get_emergency_config() -> Dict[str, Any]:
    return {
        "env": "emergency",
        "observability": {
            "mode": "lgtm_local",
            "otlp": {"protocol": "grpc", "endpoint": "127.0.0.1:4317", "insecure": True},
            "resource": {"service_name": "detectviz-adk-runtime", "service_version": "0.0.0", "environment": "emerg"},
        },
        "plugin": {"paths": []},
        "memory": {"backend": "inmem"},
    }