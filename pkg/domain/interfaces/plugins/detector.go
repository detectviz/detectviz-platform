package plugins

import (
	"context"
	"detectviz-platform/pkg/domain/entities"
)

// DetectorPlugin 定義了具體偵測器實現的介面。
// 職責: 執行特定類型的數據偵測邏輯。
// AI_PLUGIN_TYPE: "detector_plugin"
// AI_IMPL_PACKAGE: "detectviz-platform/internal/plugins/detectors"
// AI_IMPL_CONSTRUCTOR: "NewDetectorPlugin"
type DetectorPlugin interface {
	Plugin
	Execute(ctx context.Context, data map[string]interface{}, detectorConfig map[string]interface{}) (*entities.AnalysisResult, error)
}
