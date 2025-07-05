package plugins

import "context"

// NotificationPlugin 定義了通知發送插件的介面。
// 職責: 負責通過不同渠道（如郵件、簡訊、Slack）發送格式化的通知。
// AI_PLUGIN_TYPE: "notification_plugin"
// AI_IMPL_PACKAGE: "detectviz-platform/internal/domain_logic/plugins/notification/email_notifier"
// AI_IMPL_CONSTRUCTOR: "NewEmailNotifierPlugin"
type NotificationPlugin interface {
	Plugin
	// SendNotification 向指定的接收者發送通知。
	SendNotification(ctx context.Context, recipient, subject, body string, metadata map[string]interface{}) error
}
