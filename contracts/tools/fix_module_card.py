#!/usr/bin/env python3
"""
模組卡自動修復工具
修復常見的 module.card.json 格式問題
"""

import json
import sys
from pathlib import Path
from typing import Dict, Any, List

# 最新規範
CURRENT_SPEC_VERSION = "1.1.0"
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

# 角色映射 (舊 -> 新)
ROLE_MIGRATION = {
    'plugin.observability': 'observability.module',
    'plugin.storage': 'storage.module',
    'plugin.security': 'security.module'
}

# 類別映射 (舊 -> 新)
CATEGORY_MIGRATION = {
    'observability': 'observability.processor',
    'storage': 'storage.blob',
    'security': 'security.authn'
}

def fix_module_card(file_path: str) -> bool:
    """修復模組卡常見問題"""
    path = Path(file_path)
    if not path.exists():
        print(f"❌ 文件不存在: {file_path}")
        return False
    
    try:
        with open(path, 'r', encoding='utf-8') as f:
            data = json.load(f)
    except json.JSONDecodeError as e:
        print(f"❌ JSON 格式錯誤 {file_path}: {e}")
        return False
    
    modified = False
    
    # 1. 添加缺失的必要欄位
    if 'specVersion' not in data:
        data['specVersion'] = CURRENT_SPEC_VERSION
        modified = True
        print(f"  ✓ 添加 specVersion: {CURRENT_SPEC_VERSION}")
    
    if 'kind' not in data:
        data['kind'] = 'ModuleCard'
        modified = True
        print(f"  ✓ 添加 kind: ModuleCard")
    
    if 'id' not in data and 'name' in data:
        data['id'] = f"detectviz.{data['name']}"
        modified = True
        print(f"  ✓ 添加 id: {data['id']}")
    
    if 'responsibility' not in data and 'description' in data:
        data['responsibility'] = data['description']
        modified = True
        print(f"  ✓ 添加 responsibility")
    
    # 2. 更新 specVersion
    if data.get('specVersion', '0.0.0') < CURRENT_SPEC_VERSION:
        data['specVersion'] = CURRENT_SPEC_VERSION
        modified = True
        print(f"  ✓ 更新 specVersion 到 {CURRENT_SPEC_VERSION}")
    
    # 3. 修復角色
    old_role = data.get('role')
    if old_role in ROLE_MIGRATION:
        data['role'] = ROLE_MIGRATION[old_role]
        modified = True
        print(f"  ✓ 遷移角色: {old_role} -> {data['role']}")
    elif old_role and old_role not in VALID_ROLES:
        print(f"  ⚠️  未知角色: {old_role}，請手動修復")
    
    # 4. 修復類別
    old_category = data.get('category')
    if old_category in CATEGORY_MIGRATION:
        data['category'] = CATEGORY_MIGRATION[old_category]
        modified = True
        print(f"  ✓ 遷移類別: {old_category} -> {data['category']}")
    elif old_category and old_category not in VALID_CATEGORIES:
        print(f"  ⚠️  未知類別: {old_category}，請手動修復")
    
    # 5. 添加預設觀測設定
    if 'observability' not in data:
        data['observability'] = {
            "metrics_enabled": True,
            "tracing_enabled": True,
            "logging_enabled": True
        }
        modified = True
        print(f"  ✓ 添加預設觀測設定")
    
    # 6. 添加預設資源設定
    if 'resources' not in data:
        data['resources'] = {
            "cpu": "100m",
            "memory": "128Mi"
        }
        modified = True
        print(f"  ✓ 添加預設資源設定")
    
    # 保存修改
    if modified:
        with open(path, 'w', encoding='utf-8') as f:
            json.dump(data, f, indent=2, ensure_ascii=False)
        print(f"✅ 已修復並保存: {file_path}")
        return True
    else:
        print(f"✅ 無需修復: {file_path}")
        return False

def main():
    """命令行接口"""
    if len(sys.argv) != 2:
        print("使用方法: python fix_module_card.py <module.card.json>")
        sys.exit(1)
    
    file_path = sys.argv[1]
    fix_module_card(file_path)

if __name__ == "__main__":
    main()