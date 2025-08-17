#!/usr/bin/env python3
"""
模組卡生成器
自動生成符合最新 schema 規範的 module.card.json 文件
"""

import json
import sys
from pathlib import Path
from datetime import datetime
from typing import Dict, Any, List, Optional

# 最新 Schema 規範
CURRENT_SPEC_VERSION = "1.1.0"

# 有效的 categories 和 roles (與 schema 同步)
VALID_CATEGORIES = [
    'collector.input', 'transform.processor', 'aggregate.aggregator', 
    'sink.output', 'gateway', 'llm', 'retriever', 'workflow', 'a2a', 
    'capability', 'memory.backend', 'security.authn', 'security.authz',
    'observability.exporter', 'observability.processor', 'storage.blob',
    'storage.kv', 'storage.vector'
]

VALID_ROLES = [
    'agent.coordinator', 'agent.tool_exec', 'tool', 'capability',
    'plugin.gateway', 'memory.backend', 'security.module',
    'observability.module', 'storage.module'
]

def generate_module_card(
    name: str,
    role: str,
    category: str,
    description: str,
    version: str = "1.0.0",
    output_path: Optional[str] = None
) -> Dict[str, Any]:
    """生成標準化的模組卡"""
    
    # 驗證輸入
    if role not in VALID_ROLES:
        raise ValueError(f"Invalid role '{role}'. Must be one of: {VALID_ROLES}")
    
    if category not in VALID_CATEGORIES:
        raise ValueError(f"Invalid category '{category}'. Must be one of: {VALID_CATEGORIES}")
    
    # 基本模組卡結構
    module_card = {
        "specVersion": CURRENT_SPEC_VERSION,
        "kind": "ModuleCard",
        "id": f"detectviz.{name}",
        "name": name,
        "version": version,
        "role": role,
        "category": category,
        "responsibility": description,
        "description": description,
        "metadata": {
            "created_at": datetime.now().isoformat(),
            "generator": "generate_module_card.py",
            "schema_version": CURRENT_SPEC_VERSION
        },
        "observability": {
            "metrics_enabled": True,
            "tracing_enabled": True,
            "logging_enabled": True
        },
        "resources": {
            "cpu": "100m",
            "memory": "128Mi"
        },
        "author": "Detectviz Platform Team",
        "license": "MIT"
    }
    
    # 根據角色添加特定配置
    if role.startswith('plugin.'):
        module_card["type"] = "plugin"
        module_card["permissions"] = ["net"]
        
    elif role.startswith('agent.'):
        module_card["type"] = "agent"
        module_card["capabilities"] = []
        
    elif role == 'tool':
        module_card["type"] = "tool"
        module_card["execution"] = {
            "timeout": "30s",
            "retry_max": 3
        }
    
    # 根據類別添加特定配置
    if category.startswith('observability.'):
        module_card["observability"]["exporters"] = ["console", "prometheus", "otlp"]
        
    elif category.startswith('storage.'):
        module_card["persistence"] = {
            "type": "persistent",
            "backup_required": True
        }
    
    # 輸出文件
    if output_path:
        output_file = Path(output_path)
        output_file.parent.mkdir(parents=True, exist_ok=True)
        
        with open(output_file, 'w', encoding='utf-8') as f:
            json.dump(module_card, f, indent=2, ensure_ascii=False)
        
        print(f"✅ 模組卡已生成: {output_file}")
    
    return module_card

def main():
    """命令行接口"""
    if len(sys.argv) < 5:
        print("使用方法: python generate_module_card.py <name> <role> <category> <description> [output_path]")
        print(f"有效角色: {', '.join(VALID_ROLES)}")
        print(f"有效類別: {', '.join(VALID_CATEGORIES)}")
        sys.exit(1)
    
    name = sys.argv[1]
    role = sys.argv[2] 
    category = sys.argv[3]
    description = sys.argv[4]
    output_path = sys.argv[5] if len(sys.argv) > 5 else f"{name}_module.card.json"
    
    try:
        generate_module_card(name, role, category, description, output_path=output_path)
    except ValueError as e:
        print(f"❌ 錯誤: {e}")
        sys.exit(1)

if __name__ == "__main__":
    main()