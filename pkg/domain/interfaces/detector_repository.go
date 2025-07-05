package interfaces

import (
	"context"
	"detectviz-platform/pkg/domain/entities"
	"detectviz-platform/pkg/domain/valueobjects"
)

// DetectorRepository 定義了檢測器數據持久化的介面。
// 職責: 封裝檢測器實體的 CRUD 操作，隔離數據存儲細節。
// AI_PLUGIN_TYPE: "detector_repository"
// AI_IMPL_PACKAGE: "detectviz-platform/internal/repositories/mysql"
// AI_IMPL_CONSTRUCTOR: "NewDetectorRepository"
// @See: internal/repositories/mysql/detector_repository.go
type DetectorRepository interface {
	// Create 創建新檢測器
	Create(ctx context.Context, detector *entities.Detector) error
	// GetByID 根據 ID 獲取檢測器
	GetByID(ctx context.Context, id valueobjects.IDVO) (*entities.Detector, error)
	// Update 更新檢測器
	Update(ctx context.Context, detector *entities.Detector) error
	// Delete 刪除檢測器
	Delete(ctx context.Context, id valueobjects.IDVO) error
	// List 列出檢測器
	List(ctx context.Context, offset, limit int) ([]*entities.Detector, error)
}
