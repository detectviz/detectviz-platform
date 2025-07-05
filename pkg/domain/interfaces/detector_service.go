package interfaces

import (
	"context"
	"detectviz-platform/pkg/domain/entities"
	"detectviz-platform/pkg/domain/valueobjects"
)

// DetectorService 定義了檢測器業務邏輯服務的介面。
// 職責: 封裝檢測器相關的業務邏輯，協調檢測器和分析引擎。
// AI_PLUGIN_TYPE: "detector_service"
// AI_IMPL_PACKAGE: "detectviz-platform/internal/application/detector"
// AI_IMPL_CONSTRUCTOR: "NewDetectorService"
// AI_SCAFFOLD_HINT: 此介面定義檢測器服務的核心方法，實現應包含檢測器管理邏輯
// @See: internal/application/detector/detector_service.go
type DetectorService interface {
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
	// RunDetection 執行檢測
	RunDetection(ctx context.Context, detectorID valueobjects.IDVO, data interface{}) (*entities.DetectionResult, error)
	// GetDetectionHistory 獲取檢測歷史
	GetDetectionHistory(ctx context.Context, detectorID valueobjects.IDVO, offset, limit int) ([]*entities.DetectionResult, error)
}
