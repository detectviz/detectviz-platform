# 文檔維護自動化工具

> 📌 **文檔職責**：提供自動化工具來維護文檔同步、檢查引用完整性，確保SSOT原則執行

## 🎯 自動化工具概覽

### 核心自動化功能
- **文檔同步檢查器** - 檢查引用鏈的完整性
- **重複內容檢測器** - 識別違反SSOT原則的重複內容
- **文檔結構驗證器** - 驗證文檔層級結構正確性
- **引用連結檢查器** - 確保所有SSOT引用有效
- **文檔元數據管理器** - 自動更新維護信息

## 🔧 工具實現

### 1. 文檔同步檢查器

```python
#!/usr/bin/env python3
# tools/docs/sync_checker.py

import os
import re
import json
from pathlib import Path
from typing import Dict, List, Set, Tuple
from dataclasses import dataclass
from datetime import datetime

@dataclass
class ReferenceCheck:
    """引用檢查結果"""
    source_file: str
    reference: str
    target_file: str
    target_section: str
    status: str  # 'valid', 'missing_file', 'missing_section', 'invalid_format'
    line_number: int

@dataclass
class SyncReport:
    """同步檢查報告"""
    valid_references: List[ReferenceCheck]
    invalid_references: List[ReferenceCheck]
    orphaned_files: List[str]
    duplicate_content: List[Tuple[str, str]]
    
class DocumentSyncChecker:
    """文檔同步檢查器"""
    
    def __init__(self, docs_root: str = "docs"):
        self.docs_root = Path(docs_root)
        self.reference_pattern = re.compile(r'\{\{\s*([^}]+\.md)(?:#([^}]+))?\s*\}\}')
        self.ssot_files = self._identify_ssot_files()
        
    def _identify_ssot_files(self) -> Set[str]:
        """識別SSOT文檔"""
        ssot_patterns = [
            "**/PROJECT_STATUS.md",
            "**/AI_COLLABORATION_RULES.md", 
            "**/MODULE_STANDARDS.md",
            "**/AUTOMATION_TOOLS.md",
            "**/DEVELOPMENT_SETUP.md",
            "**/DOCUMENTATION_ARCHITECTURE.md"
        ]
        
        ssot_files = set()
        for pattern in ssot_patterns:
            for file_path in self.docs_root.glob(pattern):
                ssot_files.add(str(file_path.relative_to(self.docs_root)))
        
        return ssot_files
    
    def check_all_references(self) -> SyncReport:
        """檢查所有文檔引用"""
        valid_refs = []
        invalid_refs = []
        
        for md_file in self.docs_root.rglob("*.md"):
            refs = self._extract_references(md_file)
            for ref in refs:
                check_result = self._validate_reference(md_file, ref)
                if check_result.status == 'valid':
                    valid_refs.append(check_result)
                else:
                    invalid_refs.append(check_result)
        
        orphaned = self._find_orphaned_files()
        duplicates = self._detect_duplicate_content()
        
        return SyncReport(
            valid_references=valid_refs,
            invalid_references=invalid_refs,
            orphaned_files=orphaned,
            duplicate_content=duplicates
        )
    
    def _extract_references(self, file_path: Path) -> List[Dict]:
        """提取文檔中的引用"""
        references = []
        
        try:
            with open(file_path, 'r', encoding='utf-8') as f:
                for line_num, line in enumerate(f, 1):
                    matches = self.reference_pattern.findall(line)
                    for match in matches:
                        target_file, target_section = match
                        references.append({
                            'target_file': target_file,
                            'target_section': target_section or None,
                            'line_number': line_num,
                            'full_match': f'{{{{{target_file}#{target_section or ""}}}}}',
                            'source_line': line.strip()
                        })
        except Exception as e:
            print(f"Error reading {file_path}: {e}")
            
        return references
    
    def _validate_reference(self, source_file: Path, reference: Dict) -> ReferenceCheck:
        """驗證單個引用"""
        target_file = reference['target_file']
        target_section = reference['target_section']
        
        # 解析目標文件路徑
        if target_file.startswith('/'):
            # 絕對路徑
            target_path = Path(target_file[1:])  # 移除開頭的 /
        else:
            # 相對路徑
            target_path = source_file.parent / target_file
        
        # 檢查文件是否存在
        if not target_path.exists():
            return ReferenceCheck(
                source_file=str(source_file.relative_to(self.docs_root)),
                reference=reference['full_match'],
                target_file=target_file,
                target_section=target_section or "",
                status='missing_file',
                line_number=reference['line_number']
            )
        
        # 檢查章節是否存在（如果指定了章節）
        if target_section:
            if not self._section_exists(target_path, target_section):
                return ReferenceCheck(
                    source_file=str(source_file.relative_to(self.docs_root)),
                    reference=reference['full_match'],
                    target_file=target_file,
                    target_section=target_section,
                    status='missing_section',
                    line_number=reference['line_number']
                )
        
        return ReferenceCheck(
            source_file=str(source_file.relative_to(self.docs_root)),
            reference=reference['full_match'],
            target_file=target_file,
            target_section=target_section or "",
            status='valid',
            line_number=reference['line_number']
        )
    
    def _section_exists(self, file_path: Path, section: str) -> bool:
        """檢查文檔中是否存在指定章節"""
        try:
            with open(file_path, 'r', encoding='utf-8') as f:
                content = f.read()
                
            # 查找標題（支持 # ## ### 等）
            section_patterns = [
                rf'^#{1,6}\s+.*{re.escape(section)}.*$',  # 標題包含章節名
                rf'^#{1,6}\s+{re.escape(section)}\s*$',   # 標題完全匹配
                rf'#{re.escape(section)}',                 # 錨點
            ]
            
            for pattern in section_patterns:
                if re.search(pattern, content, re.MULTILINE | re.IGNORECASE):
                    return True
                    
            return False
        except Exception:
            return False
    
    def _find_orphaned_files(self) -> List[str]:
        """找出沒有被引用的文檔文件"""
        all_files = set()
        referenced_files = set()
        
        # 收集所有markdown文件
        for md_file in self.docs_root.rglob("*.md"):
            all_files.add(str(md_file.relative_to(self.docs_root)))
        
        # 收集被引用的文件
        for md_file in self.docs_root.rglob("*.md"):
            refs = self._extract_references(md_file)
            for ref in refs:
                referenced_files.add(ref['target_file'])
        
        # 排除SSOT文件（它們不需要被引用）
        orphaned = all_files - referenced_files - self.ssot_files
        
        # 排除特殊文件
        special_files = {'README.md', 'CHANGELOG.md', 'LICENSE.md'}
        orphaned = orphaned - special_files
        
        return sorted(list(orphaned))
    
    def _detect_duplicate_content(self) -> List[Tuple[str, str]]:
        """檢測重複內容"""
        duplicates = []
        file_contents = {}
        
        # 讀取所有文件內容
        for md_file in self.docs_root.rglob("*.md"):
            try:
                with open(md_file, 'r', encoding='utf-8') as f:
                    content = f.read()
                    file_contents[str(md_file.relative_to(self.docs_root))] = content
            except Exception:
                continue
        
        # 檢查重複段落（簡化版本）
        for file1, content1 in file_contents.items():
            paragraphs1 = self._extract_paragraphs(content1)
            for file2, content2 in file_contents.items():
                if file1 >= file2:  # 避免重複比較
                    continue
                    
                paragraphs2 = self._extract_paragraphs(content2)
                common_paragraphs = set(paragraphs1) & set(paragraphs2)
                
                # 如果有超過3個相同段落且每個段落長度>100字符
                significant_duplicates = [p for p in common_paragraphs if len(p) > 100]
                if len(significant_duplicates) >= 3:
                    duplicates.append((file1, file2))
        
        return duplicates
    
    def _extract_paragraphs(self, content: str) -> List[str]:
        """提取文檔段落"""
        # 簡化實現：按空行分割
        paragraphs = content.split('\n\n')
        # 過濾掉標題、代碼塊等
        clean_paragraphs = []
        for p in paragraphs:
            p = p.strip()
            if (len(p) > 50 and 
                not p.startswith('#') and 
                not p.startswith('```') and
                not p.startswith('|')):  # 排除表格
                clean_paragraphs.append(p)
        return clean_paragraphs
    
    def generate_report(self, report: SyncReport) -> str:
        """生成檢查報告"""
        output = []
        output.append("# 文檔同步檢查報告")
        output.append(f"生成時間：{datetime.now().strftime('%Y-%m-%d %H:%M:%S')}")
        output.append("")
        
        # 統計信息
        output.append("## 📊 統計摘要")
        output.append(f"- 有效引用：{len(report.valid_references)}")
        output.append(f"- 無效引用：{len(report.invalid_references)}")
        output.append(f"- 孤立文件：{len(report.orphaned_files)}")
        output.append(f"- 重複內容：{len(report.duplicate_content)}")
        output.append("")
        
        # 無效引用詳情
        if report.invalid_references:
            output.append("## ❌ 無效引用")
            for ref in report.invalid_references:
                output.append(f"- **{ref.source_file}:{ref.line_number}**")
                output.append(f"  - 引用：`{ref.reference}`")
                output.append(f"  - 目標：`{ref.target_file}`")
                output.append(f"  - 狀態：{ref.status}")
                output.append("")
        
        # 孤立文件
        if report.orphaned_files:
            output.append("## 🔍 孤立文件")
            output.append("以下文件沒有被其他文檔引用：")
            for file in report.orphaned_files:
                output.append(f"- {file}")
            output.append("")
        
        # 重複內容
        if report.duplicate_content:
            output.append("## ⚠️ 疑似重複內容")
            for file1, file2 in report.duplicate_content:
                output.append(f"- {file1} ↔ {file2}")
            output.append("")
        
        # 建議
        output.append("## 💡 修復建議")
        if report.invalid_references:
            output.append("### 修復無效引用")
            output.append("1. 檢查目標文件路徑是否正確")
            output.append("2. 確認章節標題是否存在")
            output.append("3. 更新引用格式為：`{{ docs/path/file.md#section }}`")
            output.append("")
        
        if report.orphaned_files:
            output.append("### 處理孤立文件")
            output.append("1. 考慮是否需要添加到其他文檔的引用")
            output.append("2. 評估是否可以合併到相關SSOT文檔")
            output.append("3. 如果不再需要，考慮刪除")
            output.append("")
        
        if report.duplicate_content:
            output.append("### 消除重複內容")
            output.append("1. 確定哪個文檔應該是SSOT")
            output.append("2. 將重複內容合併到SSOT文檔")
            output.append("3. 在其他文檔中使用引用替代重複內容")
            output.append("")
        
        return '\n'.join(output)

def main():
    """主函數"""
    import argparse
    
    parser = argparse.ArgumentParser(description='檢查文檔同步狀態')
    parser.add_argument('--docs-root', default='docs', help='文檔根目錄')
    parser.add_argument('--output', help='輸出報告文件路徑')
    parser.add_argument('--fix', action='store_true', help='自動修復可修復的問題')
    
    args = parser.parse_args()
    
    checker = DocumentSyncChecker(args.docs_root)
    print("正在檢查文檔同步狀態...")
    
    report = checker.check_all_references()
    report_content = checker.generate_report(report)
    
    if args.output:
        with open(args.output, 'w', encoding='utf-8') as f:
            f.write(report_content)
        print(f"報告已保存到：{args.output}")
    else:
        print(report_content)
    
    # 返回狀態碼
    if report.invalid_references or report.duplicate_content:
        return 1
    return 0

if __name__ == "__main__":
    exit(main())
```

### 2. 文檔結構重構工具

```python
#!/usr/bin/env python3
# tools/docs/restructure.py

import os
import shutil
from pathlib import Path
from typing import Dict, List
import yaml

class DocumentRestructurer:
    """文檔結構重構工具"""
    
    def __init__(self):
        self.restructure_plan = {
            # 目標結構
            "target_structure": {
                "docs/": {
                    "DOCUMENTATION_ARCHITECTURE.md": "已存在",
                    "status/": {
                        "PROJECT_STATUS.md": "已存在", 
                        "TECHNICAL_DEBT_STATUS.md": "需創建"
                    },
                    "development/": {
                        "AI_COLLABORATION_RULES.md": "已存在",
                        "MODULE_STANDARDS.md": "已存在", 
                        "AUTOMATION_TOOLS.md": "已存在",
                        "TESTING_GUIDELINES.md": "需創建"
                    },
                    "guides/": {
                        "DEVELOPMENT_SETUP.md": "已存在",
                        "AGENT_DEVELOPMENT.md": "從現有整合",
                        "TOOL_DEVELOPMENT.md": "從現有整合",
                        "QUICK_REFERENCE.md": "從現有整合"
                    },
                    "technical/": {
                        "API_REFERENCE.md": "需創建"
                    },
                    "mvp/": {
                        "TESTING_GUIDE.md": "從現有遷移"
                    }
                }
            },
            
            # 重構映射
            "restructure_mapping": {
                "agent-development-guide.md": "guides/AGENT_DEVELOPMENT.md",
                "tool-development-guide.md": "guides/TOOL_DEVELOPMENT.md", 
                "quick-reference.md": "guides/QUICK_REFERENCE.md",
                "mvp-manual-testing-guide.md": "mvp/TESTING_GUIDE.md"
            },
            
            # 需要刪除的重複文件
            "files_to_remove": [
                # 這些文件的內容已整合到新的SSOT文檔中
            ]
        }
    
    def analyze_current_structure(self) -> Dict:
        """分析當前文檔結構"""
        current = {}
        docs_path = Path("docs")
        
        if docs_path.exists():
            for item in docs_path.rglob("*"):
                if item.is_file():
                    rel_path = str(item.relative_to(docs_path))
                    current[rel_path] = {
                        "size": item.stat().st_size,
                        "exists": True
                    }
        
        return current
    
    def create_migration_plan(self) -> Dict:
        """創建遷移計劃"""
        current = self.analyze_current_structure()
        plan = {
            "moves": [],
            "merges": [],
            "creates": [],
            "deletes": []
        }
        
        # 處理重構映射
        for old_path, new_path in self.restructure_plan["restructure_mapping"].items():
            if old_path in current:
                plan["moves"].append({
                    "from": f"docs/{old_path}",
                    "to": f"docs/{new_path}",
                    "action": "move_and_update_references"
                })
        
        # 處理需要創建的新文件
        target_files = self._flatten_structure(
            self.restructure_plan["target_structure"]["docs/"]
        )
        
        for file_path, status in target_files.items():
            if status == "需創建" and file_path not in current:
                plan["creates"].append({
                    "path": f"docs/{file_path}",
                    "template": self._get_template_for_file(file_path)
                })
        
        return plan
    
    def _flatten_structure(self, structure: Dict, prefix: str = "") -> Dict:
        """扁平化結構字典"""
        result = {}
        for key, value in structure.items():
            current_path = f"{prefix}{key}" if prefix else key
            
            if isinstance(value, dict):
                result.update(self._flatten_structure(value, f"{current_path}/"))
            else:
                result[current_path] = value
        
        return result
    
    def _get_template_for_file(self, file_path: str) -> str:
        """獲取文件模板"""
        templates = {
            "status/TECHNICAL_DEBT_STATUS.md": '''# 技術債務狀態 - SSOT

> 📌 **文檔職責**：記錄技術債務清理狀態和進度追蹤

## 🎯 當前技術債務狀態

### 總體評估
- **技術債務等級**：極低 ✅
- **最後清理時間**：2024年12月17日  
- **下次計劃清理**：2025年1月15日

### 模組清理狀態

| 模組 | 清理狀態 | 完成日期 | 主要改進 |
|------|----------|----------|----------|
| go-platform | ✅ 已完成 | 2024-12-15 | OpenTelemetry整合、插件工具重構 |
| python-adk-runtime | ✅ 已完成 | 2024-12-17 | RemoteTool修復、架構驗證 |
| contracts | ✅ 已完成 | 2024-12-16 | Proto規範化、自動化工具 |

## 📊 清理成果指標

> 詳細指標：參見 [專案狀態文檔](PROJECT_STATUS.md#技術債務清理成果)

---

**維護說明**：
- 更新頻率：技術債務清理完成時更新
- 維護責任：技術負責人
- 引用方式：`{{ docs/status/TECHNICAL_DEBT_STATUS.md#section }}`
''',
            
            "development/TESTING_GUIDELINES.md": '''# 測試指導原則 - SSOT

> 📌 **文檔職責**：定義統一的測試標準、策略和最佳實踐

## 🎯 測試分層策略

### 測試金字塔
```
    E2E Tests (5%)
   ─────────────────
  Integration Tests (15%)  
 ─────────────────────────
Unit Tests (80%)
```

### 測試類型定義

#### 單元測試 (Unit Tests)
- **覆蓋率要求**：> 90%
- **測試範圍**：單個函數、方法、類
- **測試工具**：Go (testing), Python (pytest)

#### 整合測試 (Integration Tests)  
- **覆蓋率要求**：> 70%
- **測試範圍**：模組間通訊、外部依賴
- **測試工具**：TestContainers, MockServer

#### 端到端測試 (E2E Tests)
- **覆蓋率要求**：核心流程100%
- **測試範圍**：完整用戶流程
- **測試工具**：ADK測試框架

## 📋 測試標準

> 詳細測試實現：參見 [模組開發標準](../development/MODULE_STANDARDS.md#測試標準)

---

**維護說明**：
- 更新頻率：測試策略變更時更新
- 維護責任：測試團隊負責人
- 引用方式：`{{ docs/development/TESTING_GUIDELINES.md#section }}`
''',
            
            "ARCHITECTURE.md": '''# 技術架構概覽

> 📌 **文檔職責**：提供系統整體技術架構的高層視圖和關鍵設計決策

## 🏗️ 系統架構

### 混合語言架構設計
```
┌─────────────────────────────────────────────────────────────┐
│                     Go Platform Core                        │
│  ┌─────────────┐ ┌─────────────┐ ┌─────────────────────────┐ │
│  │ gRPC Gateway│ │ Event System│ │  System Metrics &       │ │
│  │             │ │             │ │  Distributed Tracing    │ │
│  └─────────────┘ └─────────────┘ └─────────────────────────┘ │
└─────────────────────────────────────────────────────────────┘
                              │
                              │ gRPC Bridge
                              │
┌─────────────────────────────────────────────────────────────┐
│                  Python ADK Runtime                         │
│  ┌─────────────┐ ┌─────────────┐ ┌─────────────────────────┐ │
│  │Agent Engine │ │Memory System│ │    Tool System &        │ │
│  │             │ │             │ │   RAG Knowledge         │ │  
│  └─────────────┘ └─────────────┘ └─────────────────────────┘ │
└─────────────────────────────────────────────────────────────┘
```

## 🎯 核心設計原則

> 詳細設計原則：參見 [AI協作核心規則](../development/AI_COLLABORATION_RULES.md)

### 1. 職責分離
- **Go平台**：高性能執行、系統集成、資源管理
- **Python運行時**：AI邏輯、決策制定、業務流程

### 2. SSOT契約優先  
- **contracts/**：所有跨語言介面的唯一事實來源
- **Protocol Buffers**：確保類型安全的通訊

### 3. 可觀察性優先
- **分散式追蹤**：OpenTelemetry端到端整合
- **統一指標**：Prometheus + Grafana監控
- **結構化日誌**：統一日誌格式和聚合

## 📊 技術棧選擇

| 層級 | 技術棧 | 選擇理由 |
|------|--------|----------|
| 平台核心 | Go 1.24+ | 高並發、低延遲、優秀的gRPC支持 |
| Agent運行時 | Python 3.11+ | 豐富的AI生態、快速開發迭代 |
| 通訊協議 | gRPC + HTTP | 類型安全、高性能、跨語言兼容 |
| 可觀察性 | OpenTelemetry | 廠商中立、標準化、完整覆蓋 |

## 🔗 相關文檔

- [專案狀態](../status/PROJECT_STATUS.md) - 當前實現進度
- [技術規格文檔](../../SPEC.md) - 詳細技術實現  
- [開發設置指南](../guides/DEVELOPMENT_SETUP.md) - 環境配置

---

**維護說明**：
- 更新頻率：架構重大變更時更新
- 維護責任：系統架構師
- 引用方式：`{{ ARCHITECTURE.md#section }}`
'''
        }
        
        return templates.get(file_path, f'''# {Path(file_path).stem.replace('_', ' ').title()}

> 📌 **文檔職責**：待補充

## 待補充內容

此文檔需要補充具體內容。

---

**維護說明**：
- 更新頻率：待定
- 維護責任：待分配
- 引用方式：`{{ docs/{file_path}#section }}`
''')
    
    def execute_migration(self, plan: Dict, dry_run: bool = True) -> bool:
        """執行遷移計劃"""
        print(f"{'🔍 預覽' if dry_run else '🚀 執行'}遷移計劃...")
        
        success = True
        
        # 創建新文件
        for create_action in plan["creates"]:
            path = Path(create_action["path"])
            print(f"創建：{path}")
            
            if not dry_run:
                path.parent.mkdir(parents=True, exist_ok=True)
                with open(path, 'w', encoding='utf-8') as f:
                    f.write(create_action["template"])
        
        # 移動文件
        for move_action in plan["moves"]:
            from_path = Path(move_action["from"])
            to_path = Path(move_action["to"])
            print(f"移動：{from_path} → {to_path}")
            
            if not dry_run and from_path.exists():
                to_path.parent.mkdir(parents=True, exist_ok=True)
                shutil.move(str(from_path), str(to_path))
        
        # 刪除文件
        for delete_action in plan["deletes"]:
            path = Path(delete_action["path"])
            print(f"刪除：{path}")
            
            if not dry_run and path.exists():
                path.unlink()
        
        return success
    
    def update_references(self, plan: Dict) -> None:
        """更新文檔中的引用"""
        print("🔗 更新文檔引用...")
        
        # 建立路徑映射
        path_mapping = {}
        for move_action in plan["moves"]:
            old_path = Path(move_action["from"]).name
            new_path = move_action["to"].replace("docs/", "")
            path_mapping[old_path] = new_path
        
        # 更新所有markdown文件中的引用
        for md_file in Path("docs").rglob("*.md"):
            self._update_file_references(md_file, path_mapping)
        
        # 更新根目錄文檔中的引用
        for md_file in Path(".").glob("*.md"):
            self._update_file_references(md_file, path_mapping)
    
    def _update_file_references(self, file_path: Path, mapping: Dict) -> None:
        """更新單個文件中的引用"""
        try:
            with open(file_path, 'r', encoding='utf-8') as f:
                content = f.read()
            
            original_content = content
            
            # 更新引用
            for old_name, new_path in mapping.items():
                # 更新直接文件引用
                content = content.replace(f"]({old_name})", f"](docs/{new_path})")
                content = content.replace(f"](./{old_name})", f"](docs/{new_path})")
                
                # 更新SSOT引用格式
                content = content.replace(
                    f"[{old_name}]({old_name})",
                    f"[{Path(new_path).stem}](docs/{new_path})"
                )
            
            # 如果內容有變化，寫回文件
            if content != original_content:
                with open(file_path, 'w', encoding='utf-8') as f:
                    f.write(content)
                print(f"  已更新：{file_path}")
                
        except Exception as e:
            print(f"  ⚠️ 更新失敗：{file_path} - {e}")

def main():
    """主函數"""
    import argparse
    
    parser = argparse.ArgumentParser(description='重構文檔結構')
    parser.add_argument('--dry-run', action='store_true', help='預覽模式，不實際執行')
    parser.add_argument('--update-refs', action='store_true', help='更新文檔引用')
    
    args = parser.parse_args()
    
    restructurer = DocumentRestructurer()
    
    # 分析當前結構
    print("📊 分析當前文檔結構...")
    current = restructurer.analyze_current_structure()
    print(f"發現 {len(current)} 個文檔文件")
    
    # 創建遷移計劃
    print("📋 創建遷移計劃...")
    plan = restructurer.create_migration_plan()
    
    print(f"計劃操作：")
    print(f"  - 創建：{len(plan['creates'])} 個文件")
    print(f"  - 移動：{len(plan['moves'])} 個文件") 
    print(f"  - 刪除：{len(plan['deletes'])} 個文件")
    
    # 執行遷移
    success = restructurer.execute_migration(plan, dry_run=args.dry_run)
    
    # 更新引用
    if args.update_refs and not args.dry_run:
        restructurer.update_references(plan)
    
    if args.dry_run:
        print("\n✅ 預覽完成。使用 --no-dry-run 執行實際遷移")
    elif success:
        print("\n✅ 文檔重構完成")
    else:
        print("\n❌ 文檔重構失敗")
        return 1
    
    return 0

if __name__ == "__main__":
    exit(main())
```

### 3. Makefile 整合

```makefile
# tools/docs/Makefile

.PHONY: help docs-check docs-sync docs-restructure docs-validate docs-clean

help: ## 顯示幫助信息
	@echo "文檔維護自動化工具"
	@echo ""
	@echo "可用命令："
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  %-20s %s\n", $$1, $$2}'

docs-check: ## 檢查文檔同步狀態
	@echo "🔍 檢查文檔同步狀態..."
	@python3 tools/docs/sync_checker.py --docs-root docs --output docs-sync-report.md

docs-sync: ## 修復文檔同步問題
	@echo "🔧 修復文檔同步問題..."
	@python3 tools/docs/sync_checker.py --docs-root docs --fix

docs-restructure: ## 重構文檔結構（預覽）
	@echo "📋 預覽文檔重構計劃..."
	@python3 tools/docs/restructure.py --dry-run

docs-restructure-execute: ## 執行文檔重構
	@echo "🚀 執行文檔重構..."
	@python3 tools/docs/restructure.py --update-refs

docs-validate: ## 驗證文檔完整性
	@echo "✅ 驗證文檔完整性..."
	@python3 tools/docs/sync_checker.py --docs-root docs
	@echo "檢查contracts模組卡..."
	@make -C contracts validate-cards
	@echo "檢查proto健康狀態..."
	@make -C contracts health-check-proto

docs-clean: ## 清理臨時文件
	@echo "🧹 清理臨時文件..."
	@rm -f docs-sync-report.md
	@rm -f docs-restructure-plan.json

docs-full-check: docs-validate docs-check ## 完整文檔檢查
	@echo "✅ 完整文檔檢查完成"

# CI/CD 整合
docs-ci: docs-validate ## CI環境文檔檢查
	@echo "🤖 CI環境文檔檢查..."
	@python3 tools/docs/sync_checker.py --docs-root docs || exit 1
```

## 🎯 使用方法

### 基本命令

```bash
# 檢查文檔同步狀態
make docs-check

# 預覽文檔重構計劃
make docs-restructure

# 執行文檔重構
make docs-restructure-execute

# 完整文檔驗證
make docs-validate

# CI/CD使用
make docs-ci
```

### 高級用法

```bash
# 自定義檢查
python3 tools/docs/sync_checker.py \
  --docs-root docs \
  --output custom-report.md

# 自動修復同步問題
python3 tools/docs/sync_checker.py \
  --docs-root docs \
  --fix

# 預覽重構（安全模式）
python3 tools/docs/restructure.py --dry-run
```

## 📊 自動化集成

### CI/CD Pipeline

```yaml
# .github/workflows/docs-check.yml
name: Documentation Check

on:
  pull_request:
    paths:
      - 'docs/**'
      - '*.md'

jobs:
  docs-check:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      
      - name: Setup Python
        uses: actions/setup-python@v4
        with:
          python-version: '3.11'
      
      - name: Check documentation sync
        run: |
          make docs-ci
      
      - name: Upload report
        if: failure()
        uses: actions/upload-artifact@v3
        with:
          name: docs-sync-report
          path: docs-sync-report.md
```

### Git Pre-commit Hook

```bash
#!/bin/bash
# .git/hooks/pre-commit

echo "檢查文檔同步狀態..."
if ! make docs-check > /dev/null 2>&1; then
    echo "❌ 文檔同步檢查失敗"
    echo "請運行 'make docs-check' 查看詳細報告"
    exit 1
fi

echo "✅ 文檔同步檢查通過"
```

---

**維護說明**：
- 更新頻率：工具功能變更時更新
- 維護責任：文檔維護團隊
- 引用方式：`{{ docs/development/DOCUMENTATION_AUTOMATION.md#section }}`