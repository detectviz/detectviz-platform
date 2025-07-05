# AI 指令規範文檔 (AI Directives Specification)

## 概述

本文檔定義了 Detectviz 平台中 AI 在程式碼生成、配置管理、文檔撰寫等環節應遵循的專用指令標籤和語義。這些指令確保 AI 生成的內容符合平台的架構規範和最佳實踐。

## 核心原則

1. **一致性**: 所有 AI 指令必須遵循統一的命名規範和語義
2. **可追溯性**: 每個指令都應該能夠追溯到具體的實現文件或配置
3. **可驗證性**: 指令的執行結果應該是可驗證和可測試的
4. **向前兼容**: 新增指令不應破壞現有的 AI 工作流程

## AI 指令標籤規範

### 1. 插件類型指令 (`AI_PLUGIN_TYPE`)

**用途**: 標識插件的類型，用於 AI 生成對應的實現代碼

**語法**: `// AI_PLUGIN_TYPE: "<plugin_type>"`

**示例**:
```go
// AI_PLUGIN_TYPE: "csv_importer"
// AI_PLUGIN_TYPE: "threshold_detector"
// AI_PLUGIN_TYPE: "gemini_llm_provider"
```

**規則**:
- 插件類型名稱使用 snake_case 格式
- 必須與 JSON Schema 文件名對應
- 必須與 composition.yaml 中的類型字段一致

### 2. 實現包路徑指令 (`AI_IMPL_PACKAGE`)

**用途**: 指定 AI 生成實現代碼的包路徑

**語法**: `// AI_IMPL_PACKAGE: "<package_path>"`

**示例**:
```go
// AI_IMPL_PACKAGE: "detectviz-platform/internal/adapters/plugins/importers"
// AI_IMPL_PACKAGE: "detectviz-platform/internal/infrastructure/platform/llm"
```

**規則**:
- 使用完整的模組路徑
- 遵循 Go 包命名慣例
- 必須對應實際的目錄結構

### 3. 構造函數指令 (`AI_IMPL_CONSTRUCTOR`)

**用途**: 指定 AI 生成的構造函數名稱

**語法**: `// AI_IMPL_CONSTRUCTOR: "<constructor_name>"`

**示例**:
```go
// AI_IMPL_CONSTRUCTOR: "NewCSVImporterPlugin"
// AI_IMPL_CONSTRUCTOR: "NewGeminiLLMProvider"
```

**規則**:
- 使用 PascalCase 格式
- 以 "New" 前綴開始
- 函數名稱應該反映實現的具體類型

### 4. 文件引用指令 (`@See`)

**用途**: 提供實現文件的引用路徑，用於 AI 查找相關代碼

**語法**: `// @See: <file_path>`

**示例**:
```go
// @See: internal/adapters/plugins/importers/csv_importer.go
// @See: internal/infrastructure/platform/llm/gemini_llm_provider.go
```

**規則**:
- 使用相對於項目根目錄的路徑
- 必須指向實際存在的文件
- 路徑使用正斜杠分隔

### 5. 配置模式指令 (`AI_CONFIG_SCHEMA`)

**用途**: 指定插件配置的 JSON Schema 文件路徑

**語法**: `// AI_CONFIG_SCHEMA: "<schema_path>"`

**示例**:
```go
// AI_CONFIG_SCHEMA: "schemas/plugins/csv_importer.json"
// AI_CONFIG_SCHEMA: "schemas/plugins/threshold_detector.json"
```

**規則**:
- 路徑相對於項目根目錄
- 必須指向有效的 JSON Schema 文件
- Schema 文件名與插件類型對應

### 6. 依賴注入指令 (`AI_DEPENDENCIES`)

**用途**: 聲明插件或服務的依賴關係，用於 AI 生成依賴注入代碼

**語法**: `// AI_DEPENDENCIES: ["<dep1>", "<dep2>", ...]`

**示例**:
```go
// AI_DEPENDENCIES: ["Logger", "DBClientProvider"]
// AI_DEPENDENCIES: ["LLMProvider", "EmbeddingStoreProvider", "Logger"]
```

**規則**:
- 使用 JSON 陣列格式
- 依賴名稱使用介面類型名稱
- 按照依賴的重要性排序

### 7. 測試生成指令 (`AI_TEST_CASES`)

**用途**: 指導 AI 生成特定的測試用例

**語法**: `// AI_TEST_CASES: ["<test_case1>", "<test_case2>", ...]`

**示例**:
```go
// AI_TEST_CASES: ["success_import", "invalid_csv_format", "database_error"]
// AI_TEST_CASES: ["threshold_exceeded", "threshold_not_exceeded", "invalid_data"]
```

**規則**:
- 測試用例名稱使用 snake_case 格式
- 應該覆蓋正常流程和異常情況
- 每個測試用例都應該有明確的驗證目標

## AI 工作流程指令

### 8. 生成模式指令 (`AI_GENERATION_MODE`)

**用途**: 指定 AI 生成代碼的模式

**語法**: `// AI_GENERATION_MODE: "<mode>"`

**可選值**:
- `"scaffold"`: 生成基礎腳手架代碼
- `"implement"`: 生成完整實現
- `"test"`: 生成測試代碼
- `"config"`: 生成配置文件

**示例**:
```go
// AI_GENERATION_MODE: "implement"
// AI_GENERATION_MODE: "test"
```

### 9. 驗證指令 (`AI_VALIDATION`)

**用途**: 指定 AI 生成代碼後需要執行的驗證步驟

**語法**: `// AI_VALIDATION: ["<validation1>", "<validation2>", ...]`

**示例**:
```go
// AI_VALIDATION: ["compile_check", "json_schema_validation", "interface_compliance"]
// AI_VALIDATION: ["unit_tests", "integration_tests", "lint_check"]
```

### 10. 版本控制指令 (`AI_VERSION`)

**用途**: 標識 AI 指令的版本，用於向前兼容

**語法**: `// AI_VERSION: "<version>"`

**示例**:
```go
// AI_VERSION: "1.0"
// AI_VERSION: "1.1"
```

## 配置文件中的 AI 指令

### JSON Schema 中的 AI 指令

在 JSON Schema 文件中，AI 指令通過特殊的屬性來定義：

```json
{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "title": "CSV Importer Configuration",
  "ai_directives": {
    "plugin_type": "csv_importer",
    "impl_package": "detectviz-platform/internal/adapters/plugins/importers",
    "constructor": "NewCSVImporterPlugin",
    "dependencies": ["DBClientProvider", "Logger"]
  },
  "type": "object",
  "properties": {
    // ... 配置屬性
  }
}
```

### YAML 配置中的 AI 指令

在 composition.yaml 等配置文件中：

```yaml
plugins:
  - type: csv_importer  # AI_PLUGIN_TYPE
    name: main_csv_importer
    config:
      # AI 將根據 JSON Schema 生成對應的配置驗證代碼
      table_name: "imported_data"
      batch_size: 1000
```

## AI 指令的執行順序

1. **解析階段**: AI 首先解析所有相關的指令標籤
2. **依賴分析**: 根據 `AI_DEPENDENCIES` 分析依賴關係
3. **代碼生成**: 根據 `AI_PLUGIN_TYPE` 和 `AI_IMPL_PACKAGE` 生成代碼
4. **配置生成**: 根據 `AI_CONFIG_SCHEMA` 生成配置驗證代碼
5. **測試生成**: 根據 `AI_TEST_CASES` 生成測試代碼
6. **驗證執行**: 根據 `AI_VALIDATION` 執行驗證步驟

## 最佳實踐

### 1. 指令的完整性

每個介面或插件都應該包含完整的 AI 指令集：

```go
// AI_PLUGIN_TYPE: "threshold_detector"
// AI_IMPL_PACKAGE: "detectviz-platform/internal/adapters/plugins/detectors"
// AI_IMPL_CONSTRUCTOR: "NewThresholdDetectorPlugin"
// AI_CONFIG_SCHEMA: "schemas/plugins/threshold_detector.json"
// AI_DEPENDENCIES: ["Logger", "MetricsProvider"]
// AI_TEST_CASES: ["threshold_exceeded", "threshold_not_exceeded", "invalid_config"]
// @See: internal/adapters/plugins/detectors/threshold_detector.go
type DetectorPlugin interface {
    Execute(ctx context.Context, data map[string]interface{}) (*entities.AnalysisResult, error)
    GetName() string
}
```

### 2. 指令的一致性

同類型的插件應該使用一致的指令格式：

```go
// 所有 Importer 插件都應該遵循相同的模式
// AI_PLUGIN_TYPE: "csv_importer"
// AI_IMPL_PACKAGE: "detectviz-platform/internal/adapters/plugins/importers"
// AI_IMPL_CONSTRUCTOR: "NewCSVImporterPlugin"

// AI_PLUGIN_TYPE: "json_importer"
// AI_IMPL_PACKAGE: "detectviz-platform/internal/adapters/plugins/importers"
// AI_IMPL_CONSTRUCTOR: "NewJSONImporterPlugin"
```

### 3. 指令的可維護性

定期檢查和更新 AI 指令，確保它們與實際的實現保持同步：

```go
// 當實現文件路徑變更時，及時更新 @See 指令
// 當依賴關係變化時，更新 AI_DEPENDENCIES 指令
// 當測試用例增加時，更新 AI_TEST_CASES 指令
```

## 錯誤處理和調試

### 常見錯誤

1. **指令不匹配**: AI 指令與實際實現不符
2. **路徑錯誤**: 文件路徑指向不存在的文件
3. **依賴缺失**: 聲明的依賴在實際代碼中未使用
4. **版本不兼容**: 使用了過時的指令格式

### 調試方法

1. **指令驗證**: 定期執行指令驗證腳本
2. **自動化測試**: 在 CI/CD 中包含指令一致性檢查
3. **文檔同步**: 確保指令文檔與實際使用保持同步

## 版本歷史

- **v1.0** (2025-01-04): 初始版本，定義基礎 AI 指令規範
- **v1.1** (未來): 計劃增加更多高級指令支持

## 相關文檔

- [Scaffold Workflow](scaffold_workflow.md)
- [Interface Specification](../architecture/interface_spec.md)
- [Engineering Specification](../ENGINEERING_SPEC.md)
- [Architecture Overview](../ARCHITECTURE.md) 