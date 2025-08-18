#!/usr/bin/env python3
"""
Proto 健康檢查工具
檢查並修復常見的 proto 文件問題
"""

import re
import sys
from pathlib import Path
from typing import List, Dict, Any, Tuple

class ProtoHealthChecker:
    def __init__(self):
        self.issues = []
        self.fixes = []
    
    def check_proto_file(self, file_path: str) -> Tuple[List[str], List[str]]:
        """檢查單個 proto 文件"""
        self.issues = []
        self.fixes = []
        
        path = Path(file_path)
        if not path.exists():
            self.issues.append(f"文件不存在: {file_path}")
            return self.issues, self.fixes
        
        content = path.read_text(encoding='utf-8')
        lines = content.split('\n')
        
        # 檢查基本格式
        self._check_basic_format(lines)
        
        # 檢查 Java 選項
        self._check_java_options(lines)
        
        # 檢查枚舉命名
        self._check_enum_naming(lines)
        
        # 檢查 RPC 命名
        self._check_rpc_naming(lines)
        
        # 檢查未使用的 import
        self._check_unused_imports(lines, content)
        
        return self.issues, self.fixes
    
    def _check_basic_format(self, lines: List[str]):
        """檢查基本格式"""
        has_syntax = False
        has_package = False
        has_go_package = False
        
        for line in lines:
            line = line.strip()
            if line.startswith('syntax ='):
                has_syntax = True
            elif line.startswith('package '):
                has_package = True
            elif line.startswith('option go_package'):
                has_go_package = True
        
        if not has_syntax:
            self.issues.append("缺少 syntax 聲明")
            self.fixes.append("添加 syntax = \"proto3\";")
        
        if not has_package:
            self.issues.append("缺少 package 聲明")
        
        if not has_go_package:
            self.issues.append("缺少 go_package 選項")
    
    def _check_java_options(self, lines: List[str]):
        """檢查 Java 選項一致性"""
        has_java_package = False
        has_java_multiple_files = False
        
        for line in lines:
            line = line.strip()
            if line.startswith('option java_package'):
                has_java_package = True
            elif line.startswith('option java_multiple_files'):
                has_java_multiple_files = True
        
        if not has_java_package:
            self.issues.append("缺少 java_package 選項")
            self.fixes.append("添加 option java_package = \"io.detectviz.contracts.v1\";")
        
        if not has_java_multiple_files:
            self.issues.append("缺少 java_multiple_files 選項")
            self.fixes.append("添加 option java_multiple_files = true;")
    
    def _check_enum_naming(self, lines: List[str]):
        """檢查枚舉命名規範"""
        current_enum = None
        
        for i, line in enumerate(lines):
            line = line.strip()
            
            # 檢測枚舉開始
            enum_match = re.match(r'enum\s+(\w+)\s*{', line)
            if enum_match:
                current_enum = enum_match.group(1)
                continue
            
            # 檢測枚舉結束
            if line == '}' and current_enum:
                current_enum = None
                continue
            
            # 檢查枚舉值命名
            if current_enum:
                enum_value_match = re.match(r'(\w+)\s*=\s*\d+', line)
                if enum_value_match:
                    value_name = enum_value_match.group(1)
                    expected_prefix = self._get_enum_prefix(current_enum)
                    
                    if not value_name.startswith(expected_prefix):
                        self.issues.append(f"枚舉值 {value_name} 應該以 {expected_prefix} 開頭")
                        correct_name = expected_prefix + value_name.split('_', 1)[-1]
                        self.fixes.append(f"將 {value_name} 重命名為 {correct_name}")
    
    def _get_enum_prefix(self, enum_name: str) -> str:
        """獲取枚舉前綴"""
        # 轉換 CamelCase 到 UPPER_SNAKE_CASE
        prefix = re.sub('([a-z0-9])([A-Z])', r'\1_\2', enum_name).upper()
        return prefix + '_'
    
    def _check_rpc_naming(self, lines: List[str]):
        """檢查 RPC 命名規範"""
        current_service = None
        
        for line in lines:
            line = line.strip()
            
            # 檢測服務開始
            service_match = re.match(r'service\s+(\w+)\s*{', line)
            if service_match:
                current_service = service_match.group(1)
                continue
            
            # 檢測服務結束
            if line == '}' and current_service:
                current_service = None
                continue
            
            # 檢查 RPC 方法
            if current_service:
                rpc_match = re.match(r'rpc\s+(\w+)\s*\(\s*(\w+)\s*\)\s*returns\s*\(\s*(?:stream\s+)?(\w+)\s*\)', line)
                if rpc_match:
                    method_name = rpc_match.group(1)
                    request_type = rpc_match.group(2)
                    response_type = rpc_match.group(3)
                    
                    expected_request = f"{method_name}Request"
                    expected_response = f"{method_name}Response"
                    
                    if request_type != expected_request:
                        self.issues.append(f"RPC {method_name} 的請求類型應該是 {expected_request}，而不是 {request_type}")
                    
                    if response_type != expected_response:
                        self.issues.append(f"RPC {method_name} 的響應類型應該是 {expected_response}，而不是 {response_type}")
    
    def _check_unused_imports(self, lines: List[str], content: str):
        """檢查未使用的 import（改進版）"""
        imports = {}  # Store path and line number

        for i, line in enumerate(lines):
            line = line.strip()
            import_match = re.match(r'import\s+"([^"]+)";', line)
            if import_match:
                import_path = import_match.group(1)
                imports[import_path] = i

        # 創建一個不包含 import 行的內容版本，以避免自我匹配
        content_without_imports = "\n".join([
            line for i, line in enumerate(lines) if i not in imports.values()
        ])

        for import_path in imports:
            # 提取基本文件名作為檢查依據
            base_filename = import_path.split('/')[-1].replace('.proto', '')

            # 啟發式：將文件名轉換為 PascalCase 作為潛在的類型名稱
            # e.g., 'google/protobuf/struct.proto' -> 'Struct'
            # e.g., 'my_custom_message.proto' -> 'MyCustomMessage'
            potential_type = self._to_pascal_case(base_filename)
            
            # 檢查潛在類型是否在文件的其餘部分被使用
            # 這是啟發式的，可能不完美，但比單純的文件名檢查要好
            if potential_type not in content_without_imports:
                self.issues.append(f"可能未使用的 import: {import_path}")
                self.fixes.append(f"考慮移除未使用的 import: {import_path}")

    def _to_pascal_case(self, snake_str: str) -> str:
        """將 snake_case 或 camelCase 轉換為 PascalCase"""
        # 處理 camelCase
        s1 = re.sub('(.)([A-Z][a-z]+)', r'\1_\2', snake_str)
        # 處理 snake_case
        s2 = re.sub('([a-z0-9])([A-Z])', r'\1_\2', s1).lower()

        return "".join(word.capitalize() for word in s2.split('_'))

def main():
    """命令行接口"""
    if len(sys.argv) != 2:
        print("使用方法: python proto_health_check.py <proto_file>")
        sys.exit(1)
    
    file_path = sys.argv[1]
    checker = ProtoHealthChecker()
    issues, fixes = checker.check_proto_file(file_path)
    
    if issues:
        print(f"❌ 發現 {len(issues)} 個問題:")
        for issue in issues:
            print(f"  - {issue}")
        
        if fixes:
            print(f"\n🔧 建議修復:")
            for fix in fixes:
                print(f"  - {fix}")
    else:
        print("✅ Proto 文件健康狀況良好")
    
    return len(issues)

if __name__ == "__main__":
    sys.exit(main())