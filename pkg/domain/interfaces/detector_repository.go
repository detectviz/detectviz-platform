package interfaces

import (
	"context"
	"detectviz-platform/pkg/domain/entities"
)

// DetectorRepository 定義了偵測器領域實體的持久化操作介面。
// 職責: 為偵測服務提供偵測器數據的 CRUD 抽象。
// AI 擴展點: 類似 UserRepository，AI 可生成不同的數據庫實現。
type DetectorRepository interface {
	Save(ctx context.Context, detector *entities.Detector) error
	FindByID(ctx context.Context, id string) (*entities.Detector, error)
	FindAll(ctx context.Context) ([]*entities.Detector, error) // 獲取所有偵測器
	Delete(ctx context.Context, id string) error
}
