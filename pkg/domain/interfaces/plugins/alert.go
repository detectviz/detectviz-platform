package plugins

import (
	"context"

	"detectviz-platform/pkg/domain/entities"
)

// AlertPlugin 定義了告警觸發插件的介面。
// 職責: 將偵測到的異常轉換為標準化告警，並集成到外部告警系統（如 PagerDuty, OpsGenie）。
// AI_PLUGIN_TYPE: "alert_plugin"
// AI_IMPL_PACKAGE: "detectviz-platform/internal/domain_logic/plugins/alert/slack_alerter"
// AI_IMPL_CONSTRUCTOR: "NewSlackAlerterPlugin"
type AlertPlugin interface {
	Plugin
	// TriggerAlert 根據分析結果和配置觸發一個告警。
	TriggerAlert(ctx context.Context, result *entities.AnalysisResult, alertConfig map[string]interface{}) error
}
