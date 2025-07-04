# Detectviz Platform API 文檔

本目錄包含 Detectviz Platform 的完整 API 文檔。

## 📋 文檔內容

- **openapi.yaml** - 完整的 OpenAPI 3.1 規範文檔
- **README.md** - 本文檔，說明如何使用 API 文檔

## 🔍 查看 API 文檔

### 方法一：使用 Swagger UI

1. 安裝 Swagger UI：
   ```bash
   npm install -g swagger-ui-serve
   ```

2. 在項目根目錄運行：
   ```bash
   swagger-ui-serve docs/api/openapi.yaml
   ```

3. 在瀏覽器中打開 `http://localhost:3000` 查看交互式 API 文檔

### 方法二：使用 VS Code 擴展

1. 安裝 VS Code 擴展：`OpenAPI (Swagger) Editor`
2. 在 VS Code 中打開 `docs/api/openapi.yaml`
3. 使用 `Ctrl+Shift+P` 打開命令面板，選擇 `OpenAPI: Preview`

### 方法三：使用在線工具

1. 訪問 [Swagger Editor](https://editor.swagger.io/)
2. 將 `openapi.yaml` 文件內容複製到編輯器中
3. 在右側查看渲染後的文檔

## 🚀 API 使用指南

### 認證

Detectviz Platform 使用 Keycloak 進行身份驗證。要訪問受保護的端點，需要在請求頭中包含有效的 Bearer token：

```bash
curl -H "Authorization: Bearer YOUR_JWT_TOKEN" \
     http://localhost:8080/api/v1/users
```

### 基本端點

#### 健康檢查
```bash
# 檢查系統整體健康狀態
curl http://localhost:8080/health

# 檢查詳細健康狀態（包含所有插件）
curl http://localhost:8080/health/detailed

# 檢查特定插件健康狀態
curl http://localhost:8080/health/plugin/csv_importer
```

#### 監控指標
```bash
# 獲取 Prometheus 指標
curl http://localhost:8080/metrics
```

### 錯誤處理

API 使用標準的 HTTP 狀態碼，並返回結構化的錯誤響應：

```json
{
  "error": "用戶未找到",
  "timestamp": "2025-01-04T10:30:00Z"
}
```

對於驗證錯誤，會返回詳細的字段錯誤信息：

```json
{
  "error": "請求參數無效",
  "timestamp": "2025-01-04T10:30:00Z",
  "details": [
    {
      "field": "email",
      "message": "郵箱格式無效"
    }
  ]
}
```

## 📊 API 端點概覽

### 健康檢查
- `GET /health` - 獲取系統健康狀態
- `GET /health/detailed` - 獲取詳細健康狀態
- `GET /health/plugin/{plugin}` - 獲取特定插件健康狀態

### 監控
- `GET /metrics` - 獲取 Prometheus 指標

## 🔧 開發者指南

### 新增 API 端點

1. 在 `openapi.yaml` 中添加新的路徑定義
2. 定義相應的請求/響應模型
3. 添加適當的標籤和描述
4. 更新相關的實現代碼

### 模型定義

所有的數據模型都在 `components/schemas` 部分定義。遵循以下命名規範：

- **實體模型**：使用名詞，如 `User`、`Detector`
- **請求模型**：使用 `Create/Update + 實體名 + Request`，如 `CreateUserRequest`
- **響應模型**：使用 `實體名 + Response`，如 `UserListResponse`

### 版本控制

API 版本通過 URL 路徑進行管理：
- `/api/v1/` - 第一版 API
- `/api/v2/` - 第二版 API（未來）

## 🧪 測試 API

### 使用 curl

```bash
# 健康檢查
curl -i http://localhost:8080/health

# 獲取指標
curl -i http://localhost:8080/metrics
```

### 使用 Postman

1. 導入 `openapi.yaml` 文件到 Postman
2. Postman 會自動生成所有端點的請求模板
3. 設置環境變量（如 base URL 和 token）
4. 執行請求進行測試

## 📝 文檔維護

### 更新文檔

1. 修改 `openapi.yaml` 文件
2. 驗證 OpenAPI 規範的正確性
3. 更新相關的示例和描述
4. 提交更改並更新版本號

### 驗證規範

使用 OpenAPI 驗證工具確保規範的正確性：

```bash
# 使用 swagger-codegen 驗證
swagger-codegen validate -i docs/api/openapi.yaml

# 使用 spectral 進行 linting
spectral lint docs/api/openapi.yaml
```

## 🤝 貢獻指南

歡迎貢獻 API 文檔的改進！請遵循以下步驟：

1. Fork 項目
2. 創建功能分支
3. 修改 API 文檔
4. 提交 Pull Request
5. 等待代碼審查

## 📞 支持

如果在使用 API 過程中遇到問題，請：

1. 查看本文檔的常見問題
2. 查看 GitHub Issues
3. 創建新的 Issue 描述問題
4. 聯繫開發團隊

---

**最後更新：2025-01-04** 