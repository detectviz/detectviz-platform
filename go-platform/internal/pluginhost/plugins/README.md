# plugins（插件實作）

分類與 `module.card.json.category` 對齊：
- `capability.gateway`：以 gRPC 暴露能力給 ADK Tool 遠端調用
- `collector.input`：資料來源介接（Events/HTTP Pull 等）
- `transform.processor`：遮罩/正規化/聚合
- `sink.output`：寫入外部系統（OnCall/Kafka/Webhook）

每個插件目錄必含：
- `plugin.go`：實作 `pluginhost.Handler`
- `module.card.json`：聲明分類、速率/容量、觀測、權限
- `README.md`：入參/出參與 Profiles 變數
- `tests/`：單元與最小 E2E