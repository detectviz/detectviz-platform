package interfaces

import (
	"context"
	"detectviz-platform/pkg/domain/entities"
	"detectviz-platform/pkg/domain/valueobjects"
)

// AnalysisResultRepository 定義了分析結果數據持久化的介面。
// 職責: 封裝分析結果實體的 CRUD 操作，隔離數據存儲細節。
// AI_PLUGIN_TYPE: "analysis_result_repository"
// AI_IMPL_PACKAGE: "detectviz-platform/internal/repositories/mysql"
// AI_IMPL_CONSTRUCTOR: "NewAnalysisResultRepository"
// @See: internal/repositories/mysql/analysis_result_repository.go
type AnalysisResultRepository interface {
	// Create 創建新分析結果
	Create(ctx context.Context, result *entities.AnalysisResult) error
	// GetByID 根據 ID 獲取分析結果
	GetByID(ctx context.Context, id valueobjects.IDVO) (*entities.AnalysisResult, error)
	// Update 更新分析結果
	Update(ctx context.Context, result *entities.AnalysisResult) error
	// Delete 刪除分析結果
	Delete(ctx context.Context, id valueobjects.IDVO) error
	// List 列出分析結果
	List(ctx context.Context, offset, limit int) ([]*entities.AnalysisResult, error)
	// GetByDetectionID 根據檢測 ID 獲取分析結果
	GetByDetectionID(ctx context.Context, detectionID valueobjects.IDVO) ([]*entities.AnalysisResult, error)
}
