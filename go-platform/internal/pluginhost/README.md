# Plugin Host 系統 - 完整實作指南

Plugin Host 是 Detectviz 平台的核心組件，負責管理和執行各種工具插件。經過優化後的系統提供健康狀態管理、負載保護、熔斷機制和 Prometheus 監控整合，確保生產環境的高可用性。

## 📁 檔案結構說明

- **`bridge_server.go`**：ToolBridge gRPC 伺服端，掛載 mTLS 與攔截器，分派請求至 Registry
- **`registry.go`**：優化的插件註冊中心，支援健康狀態管理、負載保護、熔斷機制和 Prometheus 監控
- **`metrics.go`**：Prometheus 監控整合，提供插件負載、健康狀態和調用統計指標
- **`registry_test.go`**：完整的單元測試，涵蓋並發安全、資源釋放、生命週期管理
- **`runtime.go`**：插件運行時管理，啟停控制、健康檢查
- **`observability.go`**：可觀測性整合，OpenTelemetry 監控與追蹤
- **`security.go`**：mTLS 憑證載入與安全邊界控制
- **`interceptors.go`**：Metadata 攔截、追蹤上下文透傳

## 🔄 Plugin 生命週期管理

### 註冊階段
```go
// 創建優化的註冊中心
config := &RegistryConfig{
    MaxLoad:        100,             // 最大負載閾值
    HealthInterval: 30 * time.Second, // 健康檢查間隔
}
registry := NewRegistryWithConfig(config)

// 啟動健康檢查
registry.StartHealthChecks()

// 註冊插件 - 不允許重複註冊
err := registry.Register("http_request", handler)
if err != nil {
    log.Fatal("插件註冊失敗:", err)
}

// 熱替換註冊 - 開發環境使用
err = registry.RegisterOrReplace("http_request", newHandler)
```

### 服務階段
- 接收 gRPC 請求並路由到對應 Handler
- 自動健康狀態檢查和熔斷保護
- 負載保護機制，防止插件過載
- 自動注入追蹤上下文和監控指標
- 處理超時控制和錯誤回傳
- Prometheus 指標自動收集和上報

### 關閉階段
```go
// 優雅關機 - 三階段關閉流程
registry.Shutdown()
// 1. 停止健康檢查，標記所有插件為關閉中
// 2. 等待進行中請求完成（最多2秒）
// 3. 釋放所有插件資源
```

## 🛡️ 資源治理與安全

### 插件介面
```go
// 基本處理器介面
type Handler interface {
    Invoke(ctx context.Context, req *v1.InvokeRequest) (*v1.InvokeResponse, error)
}

// 支援資源釋放的處理器
type ClosableHandler interface {
    Handler
    Close() error  // 實作資源釋放邏輯
}

// 支援健康檢查的處理器（推薦）
type HealthAwareHandler interface {
    ClosableHandler
    HealthCheck() error  // 自定義健康檢查邏輯
}
```

### 併發安全保證
- 使用 `sync.RWMutex` 保證 Registry 操作的線程安全
- 讀操作（Lookup, GetPluginStatus）允許並發
- 寫操作（Register, Unregister）互斥執行
- 原子操作管理插件負載計數器

### 健康狀態管理
- **四級健康狀態**：未知、正常、降級、異常、關閉中
- **自動健康檢查**：定期檢查插件狀態，支援自定義邏輯
- **熔斷機制**：異常插件自動熔斷，避免級聯故障
- **負載保護**：超載時自動拒絕新請求，防止系統過載

### 資源洩漏防護
- 註冊替換時自動關閉舊 Handler
- 三階段優雅關機，確保所有資源正確釋放
- 等待進行中請求完成，避免數據丟失

## 🧪 測試覆蓋

測試涵蓋以下場景：
- ✅ 基本註冊與查詢功能
- ✅ 健康狀態管理和熔斷機制
- ✅ 負載保護和過載防護
- ✅ 熱替換與資源自動釋放
- ✅ 併發操作安全性
- ✅ 三階段優雅關機流程
- ✅ Prometheus 監控指標
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
    closed atomic.Bool     // 關閉狀態標記
}

func (h *MyHandler) Invoke(ctx context.Context, req *v1.InvokeRequest) (*v1.InvokeResponse, error) {
    if h.closed.Load() {
        return nil, fmt.Errorf("handler is closed")
    }
    
    h.wg.Add(1)
    defer h.wg.Done()
    
    // 實作處理邏輯...
}

func (h *MyHandler) Close() error {
    h.closed.Store(true)
    h.wg.Wait()  // 等待所有請求完成
    if h.client != nil {
        h.client.CloseIdleConnections()
    }
    return nil
}

func (h *MyHandler) HealthCheck() error {
    if h.closed.Load() {
        return fmt.Errorf("handler is closed")
    }
    // 實作自定義健康檢查邏輯
    return nil
}
```

### 監控與狀態查詢
```go
// 獲取插件狀態
status := registry.GetPluginStatus()
for name, pluginStatus := range status {
    statusMap := pluginStatus.(map[string]interface{})
    fmt.Printf("插件 %s: 健康=%s, 負載=%d\n", 
        name, statusMap["health_str"], statusMap["load"])
}

// 更新 Prometheus 指標
registry.UpdateMetrics()

// 獲取插件列表
pluginNames := registry.GetPluginNames()
pluginCount := registry.GetPluginCount()
```

### 錯誤處理模式
```go
// 統一錯誤回傳格式
return &v1.InvokeResponse{
    Result: result,
    Status: &statuspb.Status{
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