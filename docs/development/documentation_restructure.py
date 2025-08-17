#!/usr/bin/env python3
"""
文檔重構執行工具
自動化重構docs/目錄，標準化命名並整合重複內容
"""

import os
import shutil
from pathlib import Path
from typing import Dict, List, Tuple
import json
import re

class DocumentationRestructurer:
    def __init__(self, project_root: str):
        self.project_root = Path(project_root)
        self.docs_dir = self.project_root / "docs"
        
        # 重構映射表：舊檔案 -> 新檔案路徑
        self.restructure_map = {
            # 測試和手動操作指南 -> guides/
            "mvp-manual-testing-guide.md": "guides/TESTING_GUIDE.md",
            
            # 快速參考 -> reference/
            "quick-reference.md": "reference/QUICK_REFERENCE.md",
            
            # 開發指南 -> development/
            "agent-development-guide.md": "development/AGENT_DEVELOPMENT_GUIDE.md",
            "tool-development-guide.md": "development/TOOL_DEVELOPMENT_GUIDE.md",
            
            # 現有SSOT文檔保持不變，但確保命名一致性
            # 這些已經是正確的UPPER_CASE格式
        }
        
        # 需要整合到SSOT的內容映射
        self.content_integration_map = {
            "quick-reference.md": {
                "target": "reference/QUICK_REFERENCE.md",
                "extract_sections": ["Agent 架構模式選擇", "Tool vs Go Plugin 選擇", "核心 API 速查"]
            },
            "agent-development-guide.md": {
                "target": "development/AGENT_DEVELOPMENT_GUIDE.md", 
                "extract_sections": ["Agent 架構模式詳解", "Sub-Agent 共享機制", "開發最佳實務"]
            }
        }

    def analyze_current_structure(self) -> Dict:
        """分析當前docs/目錄結構"""
        structure = {
            "files": [],
            "directories": [],
            "issues": []
        }
        
        for item in self.docs_dir.iterdir():
            if item.is_file() and item.suffix == '.md':
                structure["files"].append(str(item.relative_to(self.docs_dir)))
                
                # 檢查命名約定
                if '-' in item.stem and item.stem != item.stem.upper():
                    structure["issues"].append(f"命名不一致: {item.name} (應使用UPPER_CASE)")
                    
            elif item.is_dir():
                structure["directories"].append(str(item.relative_to(self.docs_dir)))
        
        return structure

    def create_target_directories(self):
        """建立目標目錄結構"""
        target_dirs = [
            "guides",
            "reference", 
            "development",
            "status",
            "architecture"
        ]
        
        for dir_name in target_dirs:
            target_dir = self.docs_dir / dir_name
            target_dir.mkdir(exist_ok=True)
            print(f"✓ 建立目錄: {target_dir}")

    def move_and_rename_files(self):
        """移動並重新命名檔案"""
        for old_path, new_path in self.restructure_map.items():
            old_file = self.docs_dir / old_path
            new_file = self.docs_dir / new_path
            
            if old_file.exists():
                # 確保目標目錄存在
                new_file.parent.mkdir(parents=True, exist_ok=True)
                
                # 移動檔案
                shutil.move(str(old_file), str(new_file))
                print(f"✓ 移動: {old_path} -> {new_path}")
            else:
                print(f"⚠ 檔案不存在: {old_path}")

    def extract_and_integrate_content(self):
        """提取並整合重複內容到SSOT文檔"""
        for source_file, integration_config in self.content_integration_map.items():
            source_path = self.docs_dir / source_file
            target_path = self.docs_dir / integration_config["target"]
            
            if source_path.exists() and target_path.exists():
                self._merge_content(source_path, target_path, integration_config["extract_sections"])

    def _merge_content(self, source_path: Path, target_path: Path, extract_sections: List[str]):
        """合併內容到目標檔案"""
        try:
            with open(source_path, 'r', encoding='utf-8') as f:
                source_content = f.read()
            
            with open(target_path, 'r', encoding='utf-8') as f:
                target_content = f.read()
            
            # 提取指定章節
            extracted_content = self._extract_sections(source_content, extract_sections)
            
            if extracted_content:
                # 添加整合內容到目標檔案
                merged_content = self._append_content(target_content, extracted_content, source_path.name)
                
                with open(target_path, 'w', encoding='utf-8') as f:
                    f.write(merged_content)
                
                print(f"✓ 整合內容: {source_path.name} -> {target_path.name}")
            
        except Exception as e:
            print(f"✗ 內容整合失敗: {source_path.name} - {e}")

    def _extract_sections(self, content: str, section_names: List[str]) -> str:
        """從內容中提取指定章節"""
        extracted = []
        
        for section_name in section_names:
            # 尋找章節
            pattern = rf'^#+\s*{re.escape(section_name)}.*?(?=^#+|\Z)'
            match = re.search(pattern, content, re.MULTILINE | re.DOTALL)
            
            if match:
                extracted.append(match.group(0).strip())
        
        return '\n\n'.join(extracted)

    def _append_content(self, target_content: str, new_content: str, source_filename: str) -> str:
        """將新內容附加到目標檔案"""
        integration_header = f"\n\n---\n\n## 整合內容 (來源: {source_filename})\n\n"
        return target_content + integration_header + new_content

    def create_redirect_files(self):
        """建立重定向檔案指向新位置"""
        redirect_template = """# 檔案已移動

此檔案已移動到新位置：**[{new_path}]({new_path})**

請更新您的書籤和引用。

---
*此重定向檔案將在下個版本中移除*
"""
        
        for old_path, new_path in self.restructure_map.items():
            old_file = self.docs_dir / old_path
            
            # 如果舊檔案已不存在（已移動），建立重定向
            if not old_file.exists():
                redirect_content = redirect_template.format(new_path=new_path)
                
                with open(old_file, 'w', encoding='utf-8') as f:
                    f.write(redirect_content)
                
                print(f"✓ 建立重定向: {old_path}")

    def update_internal_references(self):
        """更新內部檔案引用"""
        md_files = list(self.docs_dir.rglob("*.md"))
        
        for md_file in md_files:
            try:
                with open(md_file, 'r', encoding='utf-8') as f:
                    content = f.read()
                
                updated_content = content
                
                # 更新相對路徑引用
                for old_path, new_path in self.restructure_map.items():
                    # 匹配各種引用格式
                    patterns = [
                        rf'\[([^\]]*)\]\({re.escape(old_path)}\)',  # [text](old_path)
                        rf'\[([^\]]*)\]\(\.\/{re.escape(old_path)}\)',  # [text](./old_path)
                        rf'`{re.escape(old_path)}`',  # `old_path`
                    ]
                    
                    for pattern in patterns:
                        if re.search(pattern, updated_content):
                            updated_content = re.sub(pattern, lambda m: m.group(0).replace(old_path, new_path), updated_content)
                
                if updated_content != content:
                    with open(md_file, 'w', encoding='utf-8') as f:
                        f.write(updated_content)
                    
                    print(f"✓ 更新引用: {md_file.relative_to(self.docs_dir)}")
                    
            except Exception as e:
                print(f"✗ 更新引用失敗: {md_file.name} - {e}")

    def generate_restructure_report(self) -> str:
        """生成重構報告"""
        report = []
        report.append("# 文檔重構報告\n")
        report.append(f"**執行時間**: {Path(__file__).stat().st_mtime}\n")
        
        report.append("## 檔案移動清單\n")
        for old_path, new_path in self.restructure_map.items():
            status = "✓" if (self.docs_dir / new_path).exists() else "✗"
            report.append(f"- {status} `{old_path}` → `{new_path}`")
        
        report.append("\n## 新目錄結構\n")
        report.append("```")
        report.append("docs/")
        report.append("├── status/")
        report.append("│   └── PROJECT_STATUS.md")
        report.append("├── development/")
        report.append("│   ├── AI_COLLABORATION_RULES.md")
        report.append("│   ├── AGENT_DEVELOPMENT_GUIDE.md")
        report.append("│   ├── TOOL_DEVELOPMENT_GUIDE.md")
        report.append("│   └── DOCUMENTATION_AUTOMATION.md")
        report.append("├── reference/")
        report.append("│   └── QUICK_REFERENCE.md")
        report.append("├── guides/")
        report.append("│   └── TESTING_GUIDE.md")
        report.append("└── architecture/")
        report.append("    └── DOCUMENTATION_ARCHITECTURE.md")
        report.append("```")
        
        report.append("\n## SSOT 引用架構\n")
        report.append("- 所有重複內容已整合到對應SSOT文檔")
        report.append("- 建立統一的引用機制")
        report.append("- 實現清晰的職責邊界")
        
        return '\n'.join(report)

    def execute_restructure(self):
        """執行完整重構流程"""
        print("🚀 開始文檔重構...")
        
        # 1. 分析當前結構
        print("\n📊 分析當前結構...")
        structure = self.analyze_current_structure()
        print(f"發現 {len(structure['files'])} 個檔案，{len(structure['issues'])} 個問題")
        
        # 2. 建立目標目錄
        print("\n📁 建立目標目錄...")
        self.create_target_directories()
        
        # 3. 移動和重新命名檔案
        print("\n🔄 重新組織檔案...")
        self.move_and_rename_files()
        
        # 4. 整合重複內容
        print("\n🔗 整合重複內容...")
        self.extract_and_integrate_content()
        
        # 5. 建立重定向
        print("\n↪️ 建立重定向檔案...")
        self.create_redirect_files()
        
        # 6. 更新內部引用
        print("\n🔗 更新內部引用...")
        self.update_internal_references()
        
        # 7. 生成報告
        print("\n📋 生成重構報告...")
        report = self.generate_restructure_report()
        report_path = self.docs_dir / "RESTRUCTURE_REPORT.md"
        
        with open(report_path, 'w', encoding='utf-8') as f:
            f.write(report)
        
        print(f"✅ 重構完成！報告已保存到: {report_path}")
        print("\n📋 重構摘要:")
        print(f"- 移動檔案: {len(self.restructure_map)} 個")
        print(f"- 整合內容: {len(self.content_integration_map)} 個來源檔案")
        print("- 統一命名約定: UPPER_CASE")
        print("- 建立SSOT引用架構")


if __name__ == "__main__":
    import sys
    
    if len(sys.argv) > 1:
        project_root = sys.argv[1]
    else:
        project_root = "/Users/zoe/Documents/detectviz-platform"
    
    restructurer = DocumentationRestructurer(project_root)
    restructurer.execute_restructure()