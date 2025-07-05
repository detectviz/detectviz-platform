package interfaces

import (
	"context"
	"detectviz-platform/pkg/domain/entities"
	"detectviz-platform/pkg/domain/valueobjects"
)

// DetectionService 定義了檢測與分析業務邏輯服務的統一介面。
// 職責: 封裝檢測器管理、檢測執行、分析處理等完整的檢測流程業務邏輯。
// AI_PLUGIN_TYPE: "detection_service"
// AI_IMPL_PACKAGE: "detectviz-platform/internal/application/detection"
// AI_IMPL_CONSTRUCTOR: "NewDetectionService"
// AI_SCAFFOLD_HINT: 此介面定義檢測服務的核心方法，實現應包含檢測和分析的整合邏輯
// @See: internal/application/detection/detection_service.go
type DetectionService interface {
	// === 檢測器管理 ===

	// CreateDetector 創建新檢測器
	CreateDetector(ctx context.Context, detector *entities.Detector) error
	// GetDetectorByID 根據 ID 獲取檢測器
	GetDetectorByID(ctx context.Context, id valueobjects.IDVO) (*entities.Detector, error)
	// UpdateDetector 更新檢測器
	UpdateDetector(ctx context.Context, detector *entities.Detector) error
	// DeleteDetector 刪除檢測器
	DeleteDetector(ctx context.Context, id valueobjects.IDVO) error
	// ListDetectors 列出所有檢測器
	ListDetectors(ctx context.Context, offset, limit int) ([]*entities.Detector, error)

	// === 檢測執行 ===

	// RunDetection 執行檢測
	RunDetection(ctx context.Context, detectorID valueobjects.IDVO, data interface{}) (*entities.DetectionResult, error)
	// GetDetectionHistory 獲取檢測歷史
	GetDetectionHistory(ctx context.Context, detectorID valueobjects.IDVO, offset, limit int) ([]*entities.DetectionResult, error)

	// === 分析處理 ===

	// RunAnalysis 執行分析（基於檢測結果）
	RunAnalysis(ctx context.Context, detectionID valueobjects.IDVO, data interface{}) (*entities.AnalysisResult, error)
	// GetAnalysisResult 獲取分析結果
	GetAnalysisResult(ctx context.Context, id valueobjects.IDVO) (*entities.AnalysisResult, error)
	// ListAnalysisResults 列出分析結果
	ListAnalysisResults(ctx context.Context, offset, limit int) ([]*entities.AnalysisResult, error)
	// GetAnalysisHistory 獲取分析歷史
	GetAnalysisHistory(ctx context.Context, detectionID valueobjects.IDVO, offset, limit int) ([]*entities.AnalysisResult, error)

	// === 整合流程 ===

	// RunDetectionAndAnalysis 執行完整的檢測和分析流程
	RunDetectionAndAnalysis(ctx context.Context, detectorID valueobjects.IDVO, data interface{}) (*entities.DetectionResult, *entities.AnalysisResult, error)
}
