# Go Platform 開發工具

本目錄包含 go-platform 的開發工具和腳手架。

## 插件腳手架工具

### 使用方式

```bash
# 在 go-platform 目錄下運行
go run tools/scaffold.go <category>/<name>
```

### 支援的插件類別

| 類別 | 說明 | 範例 |
|------|------|------|
| `gateway` | 能力閘道插件 | HTTP 請求、API 調用 |
| `observability` | 可觀測性插件 | 健康檢查、指標聚合 |
| `collector.input` | 數據收集插件 | 日誌收集、指標採集 |
| `transform.processor` | 數據處理插件 | 數據轉換、過濾 |
| `sink.output` | 數據輸出插件 | 數據庫寫入、消息發送 |

### 範例

```bash
# 創建健康檢查插件
go run tools/scaffold.go observability/health_checker

# 創建 HTTP 請求插件  
go run tools/scaffold.go gateway/http_client

# 創建指標收集插件
go run tools/scaffold.go collector.input/prometheus_scraper
```

### 生成的檔案結構

```
internal/pluginhost/plugins/<category>/<name>/
├── plugin.go           # 插件主要實作
├── plugin_test.go      # 單元測試
├── module.card.json    # 插件配置卡
└── README.md          # 插件說明文檔
```

## 開發流程

1. **生成腳手架**：使用 scaffold 工具創建基本結構
2. **實作邏輯**：編輯 `plugin.go` 實作業務邏輯
3. **配置模組**：更新 `module.card.json` 設定
4. **註冊插件**：在 `register/all.go` 中註冊
5. **編寫測試**：完善單元測試和整合測試
6. **更新文檔**：補充 README 和使用範例

## 最佳實踐

- **遵循介面**：實作 `Handler` 和 `ClosableHandler` 介面
- **錯誤處理**：使用結構化錯誤和適當的日誌級別
- **資源管理**：實作 `Close()` 方法清理資源
- **可觀測性**：添加適當的 span 和 metrics
- **測試覆蓋**：確保單元測試覆蓋率 > 90%

## 工具擴展

未來可能新增的工具：
- 插件性能分析工具
- 插件依賴檢查工具
- 插件文檔生成工具
- 插件部署驗證工具