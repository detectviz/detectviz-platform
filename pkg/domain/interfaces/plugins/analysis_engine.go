package plugins

import (
	"context"

	"detectviz-platform/pkg/domain/entities"
)

// AnalysisPostProcessorPlugin 定義了分析結果後處理插件的介面。
// 職責: 對偵測結果進行深度分析、歸因或提供洞察，通常涉及更複雜的模型或LLM。
// AI_PLUGIN_TYPE: "analysis_post_processor_plugin"
// AI_IMPL_PACKAGE: "detectviz-platform/internal/plugins/analysis_engine"
// AI_IMPL_CONSTRUCTOR: "NewAnalysisPostProcessorPlugin"
type AnalysisPostProcessorPlugin interface {
	Plugin
	// PostProcess 對一個已有的分析結果進行再處理，提供更深層次的見解。
	PostProcess(ctx context.Context, result *entities.AnalysisResult, processingConfig map[string]interface{}) (map[string]interface{}, error)
}
