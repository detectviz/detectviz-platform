# Configuration loader for Python ADK Runtime
# According to spec.md: 統一配置管理，支持環境變數覆蓋和架構驗證
import os
import json
import yaml
import jsonschema
import logging
from typing import Dict, Any, Optional
from pathlib import Path

logger = logging.getLogger(__name__)

def load_config(path: str = "./configs/config.yaml") -> Dict[str, Any]:
    """Load configuration with validation and environment variable override"""
    try:
        # Load base configuration
        config = {}
        
        if os.path.exists(path):
            with open(path, 'r', encoding='utf-8') as f:
                config = yaml.safe_load(f) or {}
            logger.info(f"Loaded configuration from {path}")
        else:
            logger.warning(f"Configuration file not found: {path}, using defaults")
        
        # Apply default values
        config = _apply_defaults(config)
        
        # Override with environment variables
        config = _apply_env_overrides(config)
        
        # Validate against schema if available
        _validate_config(config)
        
        logger.info(f"Configuration loaded successfully: {list(config.keys())}")
        return config
        
    except Exception as e:
        logger.error(f"Failed to load configuration: {e}")
        # Return minimal default config for emergency operation
        return _get_emergency_config()

def _apply_defaults(config: Dict[str, Any]) -> Dict[str, Any]:
    """Apply default configuration values"""
    defaults = {
        "env": "development",
        "observability": {
            "mode": "lgtm_local",
            "otlpEndpoint": "http://localhost:4317",
            "serviceName": "detectviz-adk-runtime",
            "serviceVersion": "1.0.0"
        },
        "memory": {
            "backend": "inmem",
            "redis": {
                "url": "redis://localhost:6379",
                "db": 0
            },
            "vector": {
                "provider": "chroma",
                "endpoint": "http://localhost:8000"
            }
        },
        "plugin": {
            "paths": ["./plugins", "./src/plugins"],
            "hotReload": True,
            "timeout": "30s"
        },
        "agent": {
            "goPluginApiUrl": "http://localhost:8080",
            "defaultTimeout": "60s",
            "maxConcurrentRuns": 100
        },
        "workflow": {
            "maxSteps": 50,
            "defaultTimeout": "300s",
            "persistResults": True
        }
    }
    
    # Deep merge defaults with config
    merged = _deep_merge(defaults, config)
    return merged

def _apply_env_overrides(config: Dict[str, Any]) -> Dict[str, Any]:
    """Override configuration with environment variables"""
    env_mappings = {
        # Observability
        "OTEL_ENDPOINT": ["observability", "otlpEndpoint"],
        "OTEL_SERVICE_NAME": ["observability", "serviceName"],
        "DETECTVIZ_ENV": ["env"],
        
        # Memory
        "REDIS_URL": ["memory", "redis", "url"],
        "VECTOR_ENDPOINT": ["memory", "vector", "endpoint"],
        
        # Agent
        "GO_PLUGIN_API_URL": ["agent", "goPluginApiUrl"],
        
        # Plugin
        "PLUGIN_PATHS": ["plugin", "paths"],  # comma-separated
    }
    
    for env_var, config_path in env_mappings.items():
        env_value = os.getenv(env_var)
        if env_value:
            # Handle special cases
            if env_var == "PLUGIN_PATHS":
                env_value = [p.strip() for p in env_value.split(",")]
            
            # Set nested config value
            _set_nested_value(config, config_path, env_value)
            logger.info(f"Environment override: {env_var} -> {'.'.join(config_path)}")
    
    return config

def _validate_config(config: Dict[str, Any]) -> None:
    """Validate configuration against JSON schema"""
    try:
        # Look for schema in contracts directory
        schema_paths = [
            "../contracts/schemas/config.schema.json",
            "../../contracts/schemas/config.schema.json",
            "./contracts/schemas/config.schema.json"
        ]
        
        schema_path = None
        for path in schema_paths:
            if os.path.exists(path):
                schema_path = path
                break
        
        if schema_path:
            with open(schema_path, 'r') as f:
                schema = json.load(f)
            
            jsonschema.Draft202012Validator(schema).validate(config)
            logger.info(f"Configuration validated against schema: {schema_path}")
        else:
            logger.warning("Configuration schema not found, skipping validation")
            
    except jsonschema.ValidationError as e:
        logger.error(f"Configuration validation failed: {e.message}")
        raise ValueError(f"Invalid configuration: {e.message}")
    except Exception as e:
        logger.warning(f"Configuration validation error: {e}")

def _deep_merge(base: Dict[str, Any], override: Dict[str, Any]) -> Dict[str, Any]:
    """Deep merge two dictionaries"""
    result = base.copy()
    
    for key, value in override.items():
        if key in result and isinstance(result[key], dict) and isinstance(value, dict):
            result[key] = _deep_merge(result[key], value)
        else:
            result[key] = value
    
    return result

def _set_nested_value(config: Dict[str, Any], path: list, value: Any) -> None:
    """Set nested dictionary value using path list"""
    current = config
    for key in path[:-1]:
        if key not in current:
            current[key] = {}
        current = current[key]
    current[path[-1]] = value

def _get_emergency_config() -> Dict[str, Any]:
    """Get minimal emergency configuration for basic operation"""
    return {
        "env": "emergency",
        "observability": {
            "mode": "disabled",
            "serviceName": "detectviz-adk-runtime"
        },
        "memory": {"backend": "inmem"},
        "plugin": {"paths": ["./plugins"]},
        "agent": {"goPluginApiUrl": "http://localhost:8080"}
    }
