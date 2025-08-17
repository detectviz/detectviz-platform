# Detectviz 平台開發貢獻指南

本文檔為 Detectviz 平台所有貢獻者（包含 AI 與人類協作者）的統一開發指南，旨在確保程式碼品質、文件同步與高效協作。

## 文檔層級關係

本文檔在整體文檔體系中的定位：

```
sre-services-map.md (架構憲法 - 業務邏輯與服務地圖)
↓ 指導實現
spec.md (技術規格 - 詳細實作規範) 
↓ 開發指導
AGENT.md (本文檔 - 開發守則與協作規範)
↓ 模組執行
各模組 llm.txt (模組專用開發檢查清單與維護指南)
↓ 契約管理
contracts/ (SSOT - Protocol Buffers, Schema, 配置範本)
```

## 專案狀態

> 當前專案狀態：參見 [專案狀態文檔](docs/status/PROJECT_STATUS.md)  
> 技術債務清理成果：參見 [技術債務狀態](docs/status/TECHNICAL_DEBT_STATUS.md)

## 核心開發原則

### Agent vs Tool 職責分離

> 核心規則詳述：參見 [AI協作核心規則](docs/development/AI_COLLABORATION_RULES.md)

**黃金準則簡述**：
- **Agent (決策層)**：負責 WHY、WHAT、WHEN - 決策制定、工作流編排
- **Tool (執行層)**：負責 HOW、WHERE、WITH - 具體執行、數據操作

### SSOT 契約優先

> 詳細契約管理：參見 [contracts 使用指南](contracts/README.md)

任何跨語言介面或配置變更**必須**優先在 `contracts/` 完成：
1. 更新 proto/schema/samples
2. 執行 `buf lint && buf generate` 
3. 同步到下游專案

### 自動化工具鏈

> 完整工具使用：參見 [自動化工具文檔](docs/development/AUTOMATION_TOOLS.md)

**關鍵命令**：
```bash
# 🚀 主要開發命令（AI 協作者常用）
make validate-implementation   # 完整實作驗證 (推薦)
make setup-development         # 設置開發環境
make fix-common-issues         # 修復常見問題

# 📋 模組卡管理
make generate-module-card NAME=<name> ROLE=<role> CATEGORY=<category> DESC="<description>"
make fix-module-cards

# 🛠️ Proto 維護
make health-check-proto
make maintain-proto

# ✅ 完整驗證
make validate-with-versions
```

## llm.txt 強制執行機制

每個主要模組都有專用的 `llm.txt` 檔案，作為該模組的開發檢查清單：

- `contracts/llm.txt` - SSOT 契約維護規範
- `go-platform/llm.txt` - Go 平台開發指南  
- `python-adk-runtime/llm.txt` - Python ADK Runtime 維護指南

### 三級檢查制度

**第一級：開發前檢查**
1. 識別本次開發涉及的所有模組
2. 閱讀所有相關模組的 `llm.txt` 
3. 制定符合規範的執行計畫

**第二級：開發過程檢查**
- 每完成一個功能點，回顧檢查清單
- 確認沒有違反模組的必守規範

**第三級：提交前檢查**
- 確保所有相關模組 `llm.txt` 的檢查清單 100% 完成
- 運行自動化驗證工具
- 同步更新相關文檔

## 模組開發標準

> 詳細標準：參見 [模組開發標準](docs/development/MODULE_STANDARDS.md)

### 代碼品質要求
- 類型標註完整 (Python) / 類型安全 (Go)
- 單元測試覆蓋率 > 90%
- 錯誤處理機制完善
- 結構化日誌記錄

### 可觀察性整合
- 使用 OpenTelemetry 進行分散式追蹤
- 導出 Prometheus 指標
- 結構化日誌輸出
- 支援 pprof 性能分析

### 安全規範
- 敏感資訊通過環境變數管理
- 禁止硬編碼密鑰或憑證
- 實施適當的輸入驗證
- 遵循最小權限原則

## 測試策略

> 完整測試指南：參見 [測試指導原則](docs/development/TESTING_GUIDELINES.md)

### 測試層級
- **單元測試**：模組內部邏輯驗證
- **整合測試**：跨模組通訊驗證
- **端到端測試**：完整工作流程驗證

### 測試要求
- 每個公開介面都要有測試
- 錯誤場景覆蓋完整
- 性能基準測試
- 回歸測試自動化

## 故障排除

### 常見問題快速診斷
1. **gRPC 通訊失敗**：檢查 ToolBridge 連接狀態
2. **Proto 消息錯誤**：驗證 contracts 同步狀態  
3. **模組卡驗證失敗**：使用自動修復工具
4. **測試失敗**：檢查模組 llm.txt 指南

### 緊急修復流程
1. 識別問題模組
2. 查閱對應的 llm.txt 緊急修復指南
3. 使用提供的自動化工具
4. 驗證修復結果

## 開發工作流程

### 功能開發流程
1. **需求分析**：參考 sre-services-map.md 了解業務邏輯
2. **技術設計**：依據 spec.md 進行技術設計
3. **開發實現**：遵循本文檔和模組 llm.txt 指南
4. **測試驗證**：執行完整測試套件
5. **文檔更新**：同步更新相關文檔

### 代碼審查檢查清單
- [ ] 符合 Agent vs Tool 職責分離 (`docs/development/AI_COLLABORATION_RULES.md`)
- [ ] 遵循 SSOT 契約優先原則 (`contracts/README.md`)
- [ ] 通過所有自動化驗證 (`make validate-implementation`)
- [ ] 包含完整的測試覆蓋 (`docs/development/TESTING_GUIDELINES.md`)
- [ ] 文檔已同步更新 (相關 SSOT 文檔)

## AI 協作者特殊注意事項

1. **理解文檔層級**：按照文檔層級關係逐步深入理解
2. **嚴格遵循職責分離**：Agent 和 Tool 的邊界不可模糊
3. **優先使用自動化工具**：避免手動操作導致的錯誤
4. **保持 SSOT 一致性**：所有變更都要回到契約源頭
5. **定期檢查健康狀態**：使用提供的健康檢查工具

記住：本指南是開發工作的基石，任何疑問都要回到對應的 SSOT 文檔尋找答案！