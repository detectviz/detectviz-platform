# pluginhost（宿主框架）

- `bridge_server.go`：ToolBridge gRPC 伺服端；掛載 mTLS 與攔截器；分派請求至 `Registry`。
- `registry.go`：維護 `tool_id → Handler` 的映射。
- `runtime.go`：啟停、健康檢查等待；未來擴充熱載功能。
- `observability.go`：觀測保留位；後續接 OpenTelemetry。
- `security.go`：mTLS 憑證載入。
- `interceptors.go`：Metadata 攔截與透傳。

> 契約來源：`contracts/proto/adk_bridge.proto`（內含 ToolBridge）。請在 `contracts/` 執行 `buf generate` 產生 Go stub。