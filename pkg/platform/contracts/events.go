package contracts

import "context"

// EventBusProvider 定義了事件總線服務的介面。
// 職責: 提供異步事件的發布和訂閱機制，實現服務間的解耦。
// AI_PLUGIN_TYPE: "event_bus_provider"
// AI_IMPL_PACKAGE: "detectviz-platform/internal/platform/providers/event_bus/nats_event_bus"
// AI_IMPL_CONSTRUCTOR: "NewNATSEventBusProvider"
// @See: internal/platform/providers/event_bus/nats_event_bus.go
type EventBusProvider interface {
	// Publish 向指定的主題發布一個事件。
	Publish(ctx context.Context, topic string, event interface{}) error
	// Subscribe 訂閱一個主題，並使用提供的處理函數來處理接收到的事件。
	Subscribe(ctx context.Context, topic string, handler func(event interface{})) error
	// GetName 返回事件總線提供者的名稱。
	GetName() string
}

// AuditLogProvider 定義了審計記錄的儲存與查詢功能。
// 職責: 記錄平台中的關鍵操作（誰、在何時、做了什麼），以滿足安全和合規需求。
// AI_PLUGIN_TYPE: "audit_log_provider"
// AI_IMPL_PACKAGE: "detectviz-platform/internal/platform/providers/audit_log/db_audit_log"
// AI_IMPL_CONSTRUCTOR: "NewDBAuditLogProvider"
// @See: internal/platform/providers/audit_log/db_audit_log.go
type AuditLogProvider interface {
	// LogAction 記錄一個審計日誌條目。
	LogAction(ctx context.Context, userID, action, resource string, metadata map[string]any) error
	// GetName 返回審計日誌提供者的名稱。
	GetName() string
}
