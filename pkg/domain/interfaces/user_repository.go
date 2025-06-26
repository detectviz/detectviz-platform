package interfaces

import (
	"context"

	"detectviz-platform/pkg/domain/entities"
)

// UserRepository 定義了用戶領域實體的持久化操作介面。
// 職責: 為應用程式層的服務提供用戶數據的 CRUD 抽象，不關心具體儲存技術。
// AI 擴展點: AI 可以基於此介面生成新的數據庫適配器 (e.g., MongoDBUserRepository)。
type UserRepository interface {
	Save(ctx context.Context, user *entities.User) error // 創建或更新用戶
	FindByID(ctx context.Context, id string) (*entities.User, error)
	FindByEmail(ctx context.Context, email string) (*entities.User, error)
	Delete(ctx context.Context, id string) error
}
