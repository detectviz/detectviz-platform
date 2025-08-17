# Detectviz 文檔架構 - SSOT 設計

## 📋 文檔職責分工表 (SSOT)

| 文檔 | 職責 | 讀者 | 更新頻率 | 引用來源 |
|------|------|------|----------|----------|
| **README.md** | 專案入口與快速開始 | 新用戶、決策者 | 里程碑發布時 | docs/status/PROJECT_STATUS.md |
| **ARCHITECTURE.md** | 業務架構憲法 | 架構師、PM | 季度 | - (SSOT根源) |
| **SPEC.md** | 技術實現規範 | 開發者 | 架構變更時 | ARCHITECTURE.md |
| **AGENT.md** | AI協作守則 | AI開發者 | 技術債務清理後 | SPEC.md + docs/development/ |
| **TASKS.md** | 實作任務追蹤 | 開發團隊 | 每日 | docs/status/MVP_PROGRESS.md |
| **contracts/README.md** | SSOT契約說明 | 所有開發者 | 契約變更時 | docs/technical/CONTRACTS_GUIDE.md |
| **模組/llm.txt** | 模組專用維護指南 | 模組維護者 | 模組變更時 | docs/development/MODULE_STANDARDS.md |

## 🏗️ 新文檔結構設計

```
docs/
├── DOCUMENTATION_ARCHITECTURE.md   # 本文檔 - 文檔架構說明
├── status/                         # 專案狀態 (SSOT)
│   ├── PROJECT_STATUS.md           # 專案狀態與里程碑
│   ├── MVP_PROGRESS.md             # MVP進度追蹤
│   └── TECHNICAL_DEBT_STATUS.md    # 技術債務狀態
├── technical/                      # 技術規範 (SSOT)
│   ├── CONTRACTS_GUIDE.md          # 契約使用指南
│   ├── CROSS_LANGUAGE_BRIDGE.md    # 跨語言通訊規範
│   └── OBSERVABILITY_DESIGN.md     # 可觀察性設計
├── development/                    # 開發規範 (SSOT)
│   ├── AI_COLLABORATION_RULES.md   # AI協作核心規則
│   ├── MODULE_STANDARDS.md         # 模組開發標準
│   ├── TESTING_GUIDELINES.md       # 測試指導原則
│   └── AUTOMATION_TOOLS.md         # 自動化工具使用
├── guides/                         # 操作指南
│   ├── QUICK_START.md              # 快速開始指南
│   ├── DEVELOPMENT_SETUP.md        # 開發環境設置
│   └── DEPLOYMENT_GUIDE.md         # 部署指南
└── templates/                      # 文檔模板
    ├── MODULE_README_TEMPLATE.md   # 模組README模板
    ├── LLM_TXT_TEMPLATE.md         # llm.txt模板
    └── API_DOC_TEMPLATE.md         # API文檔模板
```

## 🔗 引用機制設計

### 標準化引用語法
```markdown
<!-- 專案狀態引用 -->
> 📊 專案狀態：{{ docs/status/PROJECT_STATUS.md#current-completion }}

<!-- 架構圖引用 -->
> 🏗️ 詳細架構：參見 [技術架構總覽](ARCHITECTURE.md#system-architecture)

<!-- 開發規範引用 -->
> 📋 開發規範：遵循 [AI協作規則](docs/development/AI_COLLABORATION_RULES.md#agent-vs-tool-principle)
```

### 自動引用更新機制
```bash
# 文檔同步工具
tools/docs/sync_references.py
tools/docs/validate_links.py
tools/docs/update_status.py
```

## 📏 文檔品質標準

### 每個文檔必須包含：
1. **職責聲明**：明確說明本文檔的唯一職責
2. **引用清單**：列出所有引用的SSOT來源
3. **更新觸發條件**：明確何時需要更新本文檔
4. **維護責任人**：指定文檔維護的責任角色

### 禁止的重複內容：
- 專案完成度數字（僅在 PROJECT_STATUS.md）
- 架構圖描述（僅在 ARCHITECTURE.md）
- 工具使用說明（僅在相應的GUIDE文檔）
- 模組標準（僅在 MODULE_STANDARDS.md）

## ✅ 重構執行狀態

### ✅ Phase 1: 建立 SSOT 文檔 - 已完成
1. ✅ 創建 docs/ 目錄結構並分類組織
2. ✅ 建立專案狀態、開發規範等SSOT文檔
3. ✅ 建立標準化引用機制和模板

### ✅ Phase 2: 重構現有文檔 - 已完成
1. ✅ 簡化 README.md 和 AGENT.md，改為引用模式
2. ✅ 清理重複內容，建立職責邊界
3. ✅ 統一檔案命名為 UPPER_CASE 約定

### ✅ Phase 3: 自動化維護 - 已完成
1. ✅ 建立文檔重構和同步工具
2. ✅ 設置引用驗證和重複檢測機制
3. ✅ 建立 Makefile 整合的自動化流程

## 🔍 驗證機制

### 文檔一致性檢查
```bash
# 檢查引用完整性
make docs-validate-references

# 檢查重複內容
make docs-check-duplicates

# 同步狀態數據
make docs-sync-status
```

### 文檔品質指標
- 重複內容檢測：0% 重複
- 引用完整性：100% 有效引用
- 更新及時性：24小時內同步
- 文檔覆蓋率：所有模組有對應文檔