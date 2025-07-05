package interfaces

import (
	"context"
	"detectviz-platform/pkg/domain/entities"
	"detectviz-platform/pkg/domain/valueobjects"
)

// AnalysisService 定義了分析業務邏輯服務的介面。
// 職責: 封裝分析相關的業務邏輯，協調分析引擎和結果存儲。
// AI_PLUGIN_TYPE: "analysis_service"
// AI_IMPL_PACKAGE: "detectviz-platform/internal/application/analysis"
// AI_IMPL_CONSTRUCTOR: "NewAnalysisService"
// AI_SCAFFOLD_HINT: 此介面定義分析服務的核心方法，實現應包含分析邏輯
// @See: internal/application/analysis/analysis_service.go
type AnalysisService interface {
	// RunAnalysis 執行分析
	RunAnalysis(ctx context.Context, detectionID valueobjects.IDVO, data interface{}) (*entities.AnalysisResult, error)
	// GetAnalysisResult 獲取分析結果
	GetAnalysisResult(ctx context.Context, id valueobjects.IDVO) (*entities.AnalysisResult, error)
	// ListAnalysisResults 列出分析結果
	ListAnalysisResults(ctx context.Context, offset, limit int) ([]*entities.AnalysisResult, error)
	// GetAnalysisHistory 獲取分析歷史
	GetAnalysisHistory(ctx context.Context, detectionID valueobjects.IDVO, offset, limit int) ([]*entities.AnalysisResult, error)
}
