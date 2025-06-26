package interfaces

import (
	"context"
	"detectviz-platform/pkg/domain/entities"
)

// AnalysisEngine 定義了核心數據分析功能的介面 (領域服務介面)。
// 職責: 執行複雜的數據分析演算法，不關心數據的來源或輸出格式。
// AI 擴展點: AI 可生成新的分析演算法實現，或集成第三方分析服務的適配器。
type AnalysisEngine interface {
	AnalyzeData(ctx context.Context, data []byte) (entities.AnalysisResult, error)                         // 分析原始數據
	ProcessDetection(ctx context.Context, detection *entities.Detection) (entities.DetectionResult, error) // 處理偵測事件
}
