# 自動化工具文檔 - SSOT

> 📌 **文檔職責**：本文檔詳細說明平台提供的自動化工具，包括使用方法、配置選項和維護指南。

## 🛠️ 工具概覽

### 核心自動化工具
- **模組卡管理工具** - 自動生成和修復模組卡
- **Proto健康檢查器** - 檢查協議定義的健康狀態  
- **配置驗證器** - 驗證配置文件的正確性
- **契約版本檢查器** - 確保跨語言契約一致性
- **代碼生成工具** - 自動生成插件骨架

## 📋 模組卡管理工具

### 生成模組卡
```bash
# 基本用法
make generate-module-card NAME=<name> ROLE=<role> CATEGORY=<category> DESC="<description>"

# 範例：生成健康聚合器模組卡
make generate-module-card \
  NAME="detectviz.plugins.health_aggregator" \
  ROLE="plugin.gateway" \
  CATEGORY="aggregate.aggregator" \
  DESC="健康數據聚合插件，從多個監控源收集和聚合健康指標"

# 範例：生成Agent模組卡  
make generate-module-card \
  NAME="detectviz.agents.postmortem_orchestrator" \
  ROLE="agent.coordinator" \
  CATEGORY="workflow" \
  DESC="事後複盤協調器，管理整個複盤分析流程"
```

### 參數說明
| 參數 | 必填 | 說明 | 範例值 |
|------|------|------|--------|
| NAME | 是 | 模組全域唯一名稱 | detectviz.tools.http_request |
| ROLE | 是 | 模組角色類型 | agent.coordinator, tool, plugin.gateway |
| CATEGORY | 是 | 模組分類細項 | workflow, aggregator, input |
| DESC | 是 | 模組功能描述 | "HTTP請求工具，支持REST API調用" |
| VERSION | 否 | 版本號（預設1.0.0） | 0.1.0 |
| LANGUAGE | 否 | 程式語言（預設go） | go, python |

### 自動修復模組卡
```bash
# 修復所有模組卡的常見問題
make fix-module-cards

# 修復特定模組卡
python3 contracts/tools/fix_module_card.py path/to/module.card.json

# 批量修復目錄下的所有模組卡
find . -name "module.card.json" -exec python3 contracts/tools/fix_module_card.py {} \;
```

**修復功能包括**：
- 補充缺失的必填欄位
- 修正錯誤的枚舉值
- 統一版本格式為SemVer
- 修正依賴關係格式
- 添加預設資源配置

### 驗證模組卡
```bash
# 驗證所有模組卡
make validate-cards

# 驗證單個模組卡
python3 contracts/tools/validate_module_card.py path/to/module.card.json

# 驗證結果範例
✓ 模組卡格式正確: detectviz.plugins.health_aggregator
✗ 驗證失敗: detectviz.agents.example
  - 缺失必填欄位: entrypoint
  - 無效的role值: invalid_role
  - 版本格式錯誤: v1.0 (應為1.0.0)
```

## 🔍 Proto健康檢查器

### 執行健康檢查
```bash
# 完整健康檢查
make health-check-proto

# 手動執行
python3 contracts/tools/proto_health_check.py

# 檢查特定proto文件
python3 contracts/tools/proto_health_check.py contracts/proto/detectviz/contracts/v1/adk_bridge.proto
```

### 檢查項目
1. **語法檢查**
   - Proto語法正確性
   - Message和Service定義完整性
   - 欄位類型和標籤正確性

2. **命名規範檢查**
   - Enum命名是否使用正確前綴
   - RPC方法命名是否遵循約定
   - Message名稱是否符合規範

3. **相容性檢查**
   - 向後相容性驗證
   - 欄位標籤衝突檢查
   - 廢棄欄位標記檢查

4. **Java包配置檢查**
   - java_package選項是否正確
   - java_outer_classname是否設置
   - java_multiple_files選項檢查

### 健康檢查報告範例
```
================================
Proto 健康檢查報告
================================

檢查的文件：
- contracts/proto/detectviz/contracts/v1/adk_bridge.proto
- contracts/proto/detectviz/contracts/v1/postmortem.proto

✓ 語法檢查通過
✓ 命名規範檢查通過  
✓ Java包配置正確
⚠ 發現 2 個建議改進項目：
  - PostmortemService.CreateReport方法建議添加超時參數
  - HealthCheckRequest建議添加版本欄位

✗ 發現 1 個錯誤：
  - PostmortemStatus枚舉缺少STATUS_前綴

總體評分：B+ (建議修復錯誤項目)
```

## ⚙️ 配置驗證器

### 驗證配置文件
```bash
# 驗證主配置文件
detectviz config validate -f ./config.yaml

# 使用自定義schema驗證
python3 contracts/tools/validate_config.py \
  --config ./config.yaml \
  --schema contracts/schemas/config.schema.json

# 批量驗證配置文件
find . -name "config.yaml" -exec detectviz config validate -f {} \;
```

### 常見驗證錯誤和解決方案

#### 錯誤：Additional property profiling.* is not allowed
```yaml
# ❌ 錯誤配置
observability:
  profiling:
    mode: pyroscope          # 不支援的欄位
    endpoint: "http://..."   # 不支援的欄位
    username: "user"         # 不支援的欄位
    password: "pass"         # 不支援的欄位

# ✅ 正確配置  
observability:
  profiling:
    enabled: true
    pprof_address: "127.0.0.1:6060"
    application_name: "go-platform"
    tags:
      service.name: "go-platform"
      deployment.environment: "dev"
```

#### 錯誤：plugin.paths: Invalid type. Expected: array
```yaml
# ❌ 錯誤配置
plugin:
  paths: "./plugins"  # 應該是陣列

# ✅ 正確配置
plugin:
  paths: 
    - "./plugins"
    - "./internal/pluginhost/plugins"
```

## 🔄 契約版本檢查器

### 版本一致性檢查
```bash
# 檢查所有契約版本
make validate-with-versions

# 手動檢查
go run go-platform/internal/contracts/version_check.go

# 檢查特定生成文件
python3 contracts/tools/check_generated_versions.py
```

### 版本檢查內容
1. **Proto生成碼版本對齊**
   - Go生成碼版本
   - Python生成碼版本
   - 生成時間戳一致性

2. **Buf版本檢查**
   - buf.lock檔案完整性
   - 依賴版本一致性
   - 生成配置正確性

3. **Cross-language一致性**
   - Go和Python的proto定義一致
   - 包名和導入路徑正確
   - 生成選項配置對齊

### 版本不一致處理
```bash
# 當版本檢查失敗時的修復流程

# 1. 重新生成契約
cd contracts
make gen

# 2. 更新Go模組
cd ../go-platform  
go mod tidy

# 3. 更新Python依賴
cd ../python-adk-runtime
pip install -e .

# 4. 重新驗證
make validate-with-versions
```

## 🏗️ 代碼生成工具

### 生成Go插件骨架
```bash
# 生成新插件
detectviz plugin new <category>/<name>

# 範例：生成HTTP請求插件
detectviz plugin new gateway/http_request

# 範例：生成健康檢查插件  
detectviz plugin new observability/health_checker

# 範例：生成數據處理插件
detectviz plugin new transform/data_processor
```

### 生成的文件結構
```
internal/pluginhost/plugins/<category>/<name>/
├── plugin.go              # 主要插件實現
├── config.go              # 配置結構定義
├── plugin_test.go         # 單元測試
├── module.card.json       # 模組卡
└── README.md              # 插件文檔
```

### 自定義模板
```bash
# 創建自定義插件模板
mkdir -p tools/templates/plugin/custom

# 使用自定義模板生成
detectviz plugin new --template custom <category>/<name>
```

## 🔧 Makefile整合

### 完整的自動化流程
```bash
# 查看所有可用命令
make help

# 完整的驗證流程
make validate-all

# 生成和驗證一體化
make gen-and-validate

# 清理和重建
make clean && make build
```

### 自定義Make目標
```makefile
# 在contracts/Makefile中添加自定義目標

.PHONY: validate-all
validate-all: validate validate-cards health-check-proto validate-with-versions
	@echo "所有驗證完成"

.PHONY: fix-all
fix-all: fix-module-cards fix-proto-format
	@echo "所有修復完成"

.PHONY: gen-and-validate
gen-and-validate: gen validate-all
	@echo "生成和驗證完成"
```

## 📊 工具性能指標

### 執行時間基準
| 工具 | 小型專案 | 中型專案 | 大型專案 |
|------|----------|----------|----------|
| 模組卡生成 | <1s | <2s | <5s |
| Proto健康檢查 | <2s | <5s | <10s |
| 配置驗證 | <1s | <2s | <3s |
| 版本檢查 | <3s | <8s | <15s |

### 最佳化建議
1. **並行處理**：批量操作時使用並行處理
2. **快取機制**：重複檢查時利用快取
3. **增量更新**：只處理變更的文件
4. **早期失敗**：遇到致命錯誤時立即停止

## 🚨 故障排除

### 常見問題和解決方案

#### 問題：模組卡生成失敗
```bash
# 檢查Python環境
python3 --version
pip3 list | grep jsonschema

# 重新安裝依賴
pip3 install jsonschema jinja2

# 檢查模板文件
ls -la contracts/tools/templates/
```

#### 問題：Proto健康檢查通不過
```bash
# 檢查buf安裝
buf --version

# 重新安裝buf
curl -sSL "https://github.com/bufbuild/buf/releases/latest/download/buf-$(uname -s)-$(uname -m)" -o /usr/local/bin/buf
chmod +x /usr/local/bin/buf

# 檢查proto語法
buf lint contracts/proto
```

#### 問題：版本檢查失敗
```bash
# 清理生成文件
rm -rf contracts/gen/

# 重新生成
cd contracts && make gen

# 檢查權限問題
ls -la contracts/gen/
chmod -R 644 contracts/gen/
```

## 📝 工具開發指南

### 添加新的自動化工具
1. **創建工具腳本**
   ```python
   # contracts/tools/new_tool.py
   #!/usr/bin/env python3
   """新工具的說明文檔"""
   
   import argparse
   import sys
   from pathlib import Path
   
   def main():
       parser = argparse.ArgumentParser(description='新工具功能描述')
       parser.add_argument('--input', required=True, help='輸入參數')
       args = parser.parse_args()
       
       # 工具邏輯實現
       print(f"處理輸入：{args.input}")
   
   if __name__ == '__main__':
       main()
   ```

2. **更新Makefile**
   ```makefile
   .PHONY: new-tool
   new-tool:
   	python3 contracts/tools/new_tool.py --input $(INPUT)
   ```

3. **添加測試**
   ```python
   # contracts/tools/test_new_tool.py
   import unittest
   from new_tool import main
   
   class TestNewTool(unittest.TestCase):
       def test_basic_functionality(self):
           # 測試基本功能
           pass
   ```

4. **更新文檔**
   - 在本文件中添加工具說明
   - 更新README.md
   - 添加使用範例

---

**維護說明**：
- 更新頻率：新增工具或工具功能變更時更新
- 維護責任：工具開發團隊
- 引用方式：`{{ docs/development/AUTOMATION_TOOLS.md#section }}`