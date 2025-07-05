package interfaces

import (
	"context"
	"detectviz-platform/pkg/domain/entities"
	"detectviz-platform/pkg/domain/valueobjects"
)

// UserService 定義了用戶業務邏輯服務的介面。
// 職責: 封裝用戶相關的業務邏輯，協調多個倉儲和外部服務。
// AI_PLUGIN_TYPE: "user_service"
// AI_IMPL_PACKAGE: "detectviz-platform/internal/application/user"
// AI_IMPL_CONSTRUCTOR: "NewUserService"
// AI_SCAFFOLD_HINT: 此介面定義用戶服務的核心方法，實現應包含用戶管理邏輯
// @See: internal/application/user/user_service.go
type UserService interface {
	// CreateUser 創建新用戶
	CreateUser(ctx context.Context, user *entities.User) error
	// GetUserByID 根據 ID 獲取用戶
	GetUserByID(ctx context.Context, id valueobjects.IDVO) (*entities.User, error)
	// GetUserByEmail 根據郵箱獲取用戶
	GetUserByEmail(ctx context.Context, email valueobjects.EmailVO) (*entities.User, error)
	// UpdateUser 更新用戶信息
	UpdateUser(ctx context.Context, user *entities.User) error
	// DeleteUser 刪除用戶
	DeleteUser(ctx context.Context, id valueobjects.IDVO) error
	// ListUsers 列出所有用戶
	ListUsers(ctx context.Context, offset, limit int) ([]*entities.User, error)
	// AuthenticateUser 驗證用戶憑證
	AuthenticateUser(ctx context.Context, email valueobjects.EmailVO, password string) (*entities.User, error)
	// ChangePassword 更改用戶密碼
	ChangePassword(ctx context.Context, userID valueobjects.IDVO, oldPassword, newPassword string) error
}
