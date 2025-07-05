# AI Scaffold 前置條件與目錄結構指南

## 📁 優化後的目錄結構

DetectViz Platform 已完成架構優化，採用簡化且高效的目錄結構：

```
detectviz-platform/
├── cmd/api/                    # 應用程式入口
├── internal/                   # 內部實現代碼
│   ├── adapters/              # 適配器層
│   │   ├── http_handlers/     # HTTP 請求處理器
│   │   └── web/               # Web UI 組件（整合）
│   ├── application/           # 應用層
│   ├── bootstrap/             # 啟動配置管理
│   ├── infrastructure/        # 基礎設施層
│   │   └── platform/          # 平台核心服務實現
│   │       ├── auth/          # 身份驗證服務
│   │       ├── config/        # 配置管理
│   │       ├── di/            # 依賴注入容器
│   │       ├── health/        # 健康檢查
│   │       ├── http_server/   # HTTP 服務器
│   │       ├── registry/      # 插件註冊表
│   │       └── telemetry/     # 遙測服務（整合）
│   ├── plugins/               # 插件實現（集中管理）
│   ├── repositories/          # 倉儲層
│   └── testdata/              # 測試數據
├── pkg/                       # 公共代碼庫
│   ├── application/           # 應用層公共組件
│   │   └── shared/            # 共享組件（DTO + Mapper）
│   ├── common/                # 通用工具
│   │   └── utils/             # 工具函數集合
│   ├── domain/                # 領域層
│   │   ├── entities/          # 領域實體
│   │   ├── errors/            # 自定義錯誤
│   │   ├── interfaces/        # 領域介面（扁平化）
│   │   │   └── plugins/       # 插件介面
│   │   └── valueobjects/      # 值對象
│   └── platform/              # 平台契約
│       └── contracts/         # 平台服務介面
└── docs/                      # 文檔
```

## 🔧 關鍵優化點

### 1. 目錄簡化與合併
- **DTO + Mapper 合併**: 統一在 `pkg/application/shared/`
- **Interface 扁平化**: 移除深層嵌套，集中在 `pkg/domain/interfaces/`
- **Telemetry 整合**: logger、tracing、metrics 合併為統一模組
- **Web UI 整合**: web_ui 插件合併到 web adapters

### 2. AI Scaffold 友好性增強
- **標準化路徑**: 減少路徑深度，簡化 import
- **豐富的 AI 提示**: 所有關鍵組件都包含 AI_SCAFFOLD_HINT 註解
- **工具函數支援**: 提供標準化的工具函數庫

## 🚀 AI Scaffold 使用指南

### 1. 創建新功能模組
```bash
# 創建新的領域實體
AI_SCAFFOLD_HINT: "在 pkg/domain/entities/ 創建新實體"

# 創建對應的服務介面
AI_SCAFFOLD_HINT: "在 pkg/domain/interfaces/ 創建服務介面"

# 創建應用服務實現
AI_SCAFFOLD_HINT: "在 internal/application/ 創建應用服務"
```

### 2. 創建新插件
```bash
# 插件實現統一放在 internal/plugins/
AI_SCAFFOLD_HINT: "插件命名格式：{type}_{implementation}"

# 插件介面定義在 pkg/domain/interfaces/plugins/
AI_SCAFFOLD_HINT: "插件介面遵循標準化命名"
```

### 3. 數據轉換層
```bash
# DTO 和 Mapper 統一管理
AI_SCAFFOLD_HINT: "在 pkg/application/shared/ 創建 DTO 和對應的 Mapper"

# 利用工具函數
AI_SCAFFOLD_HINT: "使用 pkg/common/utils/ 中的標準化工具"
```

## 📊 路徑映射參考

### 舊路徑 → 新路徑
| 舊路徑 | 新路徑 |
|--------|--------|
| `pkg/application/dto/` | `pkg/application/shared/` |
| `pkg/application/mapper/` | `pkg/application/shared/` |
| `pkg/domain/interfaces/services/` | `pkg/domain/interfaces/` |
| `pkg/domain/interfaces/repositories/` | `pkg/domain/interfaces/` |
| `internal/infrastructure/platform/logger/` | `internal/infrastructure/platform/telemetry/` |
| `internal/infrastructure/platform/monitoring/` | `internal/infrastructure/platform/telemetry/` |
| `internal/infrastructure/platform/tracing/` | `internal/infrastructure/platform/telemetry/` |
| `internal/plugins/web_ui/` | `internal/adapters/web/` |

## 🔍 AI 提示註解規範

### 1. 實體層 (Entities)
```go
// AI_SCAFFOLD_HINT: 領域實體，包含核心業務邏輯和不變性約束
type User struct {
    // 實體字段
}

// AI_SCAFFOLD_HINT: 業務方法，封裝領域邏輯
func (u *User) UpdateProfile(profile Profile) error {
    // 業務邏輯
}
```

### 2. 服務層 (Services)
```go
// AI_SCAFFOLD_HINT: 應用服務，協調多個領域對象完成業務用例
type UserService struct {
    // 依賴注入
}

// AI_SCAFFOLD_HINT: 用例方法，實現特定業務場景
func (s *UserService) CreateUser(ctx context.Context, req *CreateUserRequest) (*User, error) {
    // 用例邏輯
}
```

### 3. 插件層 (Plugins)
```go
// AI_SCAFFOLD_HINT: 插件實現，遵循標準化介面
// AI_PLUGIN_TYPE: "detector"
// AI_IMPL_PACKAGE: "detectviz-platform/internal/plugins/detectors"
type ThresholdDetector struct {
    // 插件字段
}
```

### 4. 共享層 (Shared)
```go
// AI_SCAFFOLD_HINT: DTO 定義，用於數據傳輸
type CreateUserRequest struct {
    Email    string `json:"email" validate:"required,email"`
    Password string `json:"password" validate:"required,min=8"`
}

// AI_SCAFFOLD_HINT: Mapper 實現，處理 DTO 與實體的轉換
func (m *UserMapper) ToEntity(req *CreateUserRequest) (*entities.User, error) {
    // 轉換邏輯
}
```

## 📝 最佳實踐

### 1. 命名規範
- **實體**: PascalCase，如 `User`, `DetectionResult`
- **服務**: `{Domain}Service`，如 `UserService`, `DetectorService`
- **倉儲**: `{Domain}Repository`，如 `UserRepository`
- **插件**: `{Type}{Implementation}Plugin`，如 `ThresholdDetectorPlugin`

### 2. 檔案組織
- 一個檔案一個主要類型
- 相關的 DTO 和 Mapper 放在同一個檔案
- 測試檔案與實現檔案同目錄

### 3. 依賴管理
- 使用介面進行依賴注入
- 避免循環依賴
- 遵循依賴反轉原則

## 🎯 總結

優化後的架構提供了：
- **更簡潔的目錄結構**
- **更好的 AI 理解性**
- **更高的開發效率**
- **更低的維護成本**

所有 AI Scaffold 操作都應基於此結構進行，確保生成的代碼符合平台的架構標準。