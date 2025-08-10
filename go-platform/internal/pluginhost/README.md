# Plugin Host 系統 - 完整實作指南

Plugin Host 是 Detectviz 平台的核心組件，負責管理和執行各種工具插件，確保資源安全釋放和生命週期完整性。

## 📁 檔案結構說明

- **`bridge_server.go`**：ToolBridge gRPC 伺服端，掛載 mTLS 與攔截器，分派請求至 Registry
- **`registry.go`**：維護 `tool_id → Handler` 的映射，提供線程安全的插件註冊管理
- **`registry_test.go`**：完整的單元測試，涵蓋並發安全、資源釋放、生命週期管理
- **`runtime.go`**：插件運行時管理，啟停控制、健康檢查
- **`observability.go`**：可觀測性整合，OpenTelemetry 監控與追蹤
- **`security.go`**：mTLS 憑證載入與安全邊界控制
- **`interceptors.go`**：Metadata 攔截、追蹤上下文透傳

## 🔄 Plugin 生命週期管理

### 註冊階段
```go
// 嚴格註冊 - 生產環境推薦
err := registry.RegisterStrict("http_request", handler)
if errors.Is(err, ErrHandlerExists) {
    log.Fatal("插件已存在，防止意外覆蓋")
}

// 熱替換註冊 - 開發環境使用
registry.RegisterOrReplace("http_request", newHandler)
```

### 服務階段
- 接收 gRPC 請求並路由到對應 Handler
- 自動注入追蹤上下文和監控指標
- 處理超時控制和錯誤回傳

### 關閉階段
```go
// 優雅關機 - 自動釋放所有已註冊插件的資源
err := registry.Shutdown()
if err != nil {
    log.Printf("插件關閉過程中發生錯誤: %v", err)
}
```

## 🛡️ 資源治理與安全

### ClosableHandler 接口
```go
type ClosableHandler interface {
    Handler
    Close() error  // 實作資源釋放邏輯
}
```

### 併發安全保證
- 使用 `sync.RWMutex` 保證 Registry 操作的線程安全
- 讀操作（Lookup, List）允許並發
- 寫操作（Register, Unregister）互斥執行

### 資源洩漏防護
- 註冊替換時自動關閉舊 Handler
- Shutdown 時批次釋放所有資源
- 錯誤聚合，確保所有清理嘗試都被執行

## 🧪 測試覆蓋

測試涵蓋以下場景：
- ✅ 基本註冊與查詢功能
- ✅ 嚴格註冊防重複機制
- ✅ 熱替換與資源自動釋放
- ✅ 併發操作安全性
- ✅ 優雅關機流程
- ✅ 錯誤處理與恢復
- ✅ 性能基準測試

執行測試：
```bash
cd go-platform
go test -v ./internal/pluginhost/... -race
```

## 🔧 最佳實踐

### Handler 實作建議
```go
type MyHandler struct {
    client *http.Client
    wg     sync.WaitGroup  // 追蹤活躍請求
}

func (h *MyHandler) Invoke(ctx context.Context, req *contractspb.ToolInvokeRequest) (*contractspb.ToolInvokeReply, error) {
    h.wg.Add(1)
    defer h.wg.Done()
    
    // 實作處理邏輯...
}

func (h *MyHandler) Close() error {
    h.wg.Wait()  // 等待所有請求完成
    if h.client != nil {
        h.client.CloseIdleConnections()
    }
    return nil
}
```

### 錯誤處理模式
```go
// 統一錯誤回傳格式
return &contractspb.ToolInvokeReply{
    Status: &rpc.Status{
        Code:    int32(codes.InvalidArgument),
        Message: "參數驗證失敗",
    },
}, nil  // gRPC 層面不報錯，錯誤信息在 Status 中
```

## 🔗 相關資源

- **契約來源**：`contracts/proto/adk_bridge.proto`
- **生成指令**：`cd contracts && make gen`
- **健康檢查**：`GET /contracts` - 檢視契約版本信息
- **監控端點**：集成 OpenTelemetry，支援 traces 和 metrics

---

> 📝 **維護提醒**：任何對插件接口的修改都需要同步更新契約定義，並重新生成跨語言存根檔案。