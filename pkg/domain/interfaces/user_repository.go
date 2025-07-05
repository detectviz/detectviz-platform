package interfaces

import (
	"context"
	"detectviz-platform/pkg/domain/entities"
	"detectviz-platform/pkg/domain/valueobjects"
)

// UserRepository 定義了用戶數據持久化的介面。
// 職責: 封裝用戶實體的 CRUD 操作，隔離數據存儲細節。
// AI_PLUGIN_TYPE: "user_repository"
// AI_IMPL_PACKAGE: "detectviz-platform/internal/repositories/mysql"
// AI_IMPL_CONSTRUCTOR: "NewUserRepository"
// @See: internal/repositories/mysql/user_repository.go
type UserRepository interface {
	// Create 創建新用戶
	Create(ctx context.Context, user *entities.User) error
	// GetByID 根據 ID 獲取用戶
	GetByID(ctx context.Context, id valueobjects.IDVO) (*entities.User, error)
	// GetByEmail 根據郵箱獲取用戶
	GetByEmail(ctx context.Context, email valueobjects.EmailVO) (*entities.User, error)
	// Update 更新用戶信息
	Update(ctx context.Context, user *entities.User) error
	// Delete 刪除用戶
	Delete(ctx context.Context, id valueobjects.IDVO) error
	// List 列出用戶
	List(ctx context.Context, offset, limit int) ([]*entities.User, error)
}
