package contracts

import (
	"context"
)

// AuthProvider 定義了身份驗證與授權服務的統一介面。
// 職責: 提供完整的認證和授權功能，包括身份驗證、權限檢查、密碼處理和 CSRF 保護。
// AI_PLUGIN_TYPE: "auth_provider"
// AI_IMPL_PACKAGE: "detectviz-platform/internal/infrastructure/platform/auth"
// AI_IMPL_CONSTRUCTOR: "NewAuthProvider"
// @See: internal/infrastructure/platform/auth/auth_provider.go
type AuthProvider interface {
	// === 身份驗證 ===

	// Authenticate 根據提供的憑據驗證用戶身份，成功則返回用戶ID。
	Authenticate(ctx context.Context, credentials string) (userID string, err error)
	// VerifyToken 驗證一個 JWT token 的有效性並返回用戶ID。
	VerifyToken(ctx context.Context, token string) (string, error)

	// === 授權檢查 ===

	// Authorize 檢查指定用戶是否有權限對某個資源執行特定操作。
	Authorize(ctx context.Context, userID string, resource string, action string) (bool, error)
	// CheckPermissions 查詢外部服務以檢查用戶的詳細權限。
	CheckPermissions(ctx context.Context, userID, resource, action string) (bool, error)

	// === 密碼管理 ===

	// HashPassword 將明文密碼轉換為安全的散列值。
	HashPassword(ctx context.Context, plainPassword string) (string, error)
	// VerifyPassword 驗證明文密碼是否與給定的散列值匹配。
	VerifyPassword(ctx context.Context, plainPassword, hashedPassword string) (bool, error)

	// === CSRF 保護 ===

	// GenerateCSRFToken 為當前會話生成一個新的 CSRF token。
	GenerateCSRFToken(ctx context.Context) (string, error)
	// ValidateCSRFToken 驗證傳入的 CSRF token 是否有效。
	ValidateCSRFToken(ctx context.Context, token string) error

	// === 通用方法 ===

	// GetName 返回認證提供者的名稱。
	GetName() string
}

// AuthStorageProvider 定義了認證相關數據存儲的統一介面。
// 職責: 管理會話、令牌、認證狀態等認證相關數據的存儲和檢索。
// AI_PLUGIN_TYPE: "auth_storage_provider"
// AI_IMPL_PACKAGE: "detectviz-platform/internal/infrastructure/platform/auth/storage"
// AI_IMPL_CONSTRUCTOR: "NewAuthStorageProvider"
// @See: internal/infrastructure/platform/auth/storage/auth_storage_provider.go
type AuthStorageProvider interface {
	// === 會話管理 ===

	// SetSession 創建或更新一個用戶會話。
	SetSession(ctx context.Context, sessionID string, data map[string]any) error
	// GetSession 根據 sessionID 檢索會話數據。
	GetSession(ctx context.Context, sessionID string) (map[string]any, error)
	// DeleteSession 刪除一個會話。
	DeleteSession(ctx context.Context, sessionID string) error
	// RefreshSession 刷新會話的過期時間。
	RefreshSession(ctx context.Context, sessionID string) error

	// === 令牌管理 ===

	// StoreToken 存儲認證令牌（如 JWT refresh token）。
	StoreToken(ctx context.Context, userID, tokenType, token string, expiry int64) error
	// GetToken 檢索存儲的令牌。
	GetToken(ctx context.Context, userID, tokenType string) (string, error)
	// RevokeToken 撤銷指定的令牌。
	RevokeToken(ctx context.Context, userID, tokenType string) error
	// CleanExpiredTokens 清理過期的令牌。
	CleanExpiredTokens(ctx context.Context) error

	// === CSRF 令牌存儲 ===

	// StoreCSRFToken 存儲 CSRF 令牌。
	StoreCSRFToken(ctx context.Context, sessionID, token string) error
	// ValidateStoredCSRFToken 驗證存儲的 CSRF 令牌。
	ValidateStoredCSRFToken(ctx context.Context, sessionID, token string) (bool, error)
	// CleanExpiredCSRFTokens 清理過期的 CSRF 令牌。
	CleanExpiredCSRFTokens(ctx context.Context) error

	// === 通用方法 ===

	// GetName 返回存儲提供者的名稱。
	GetName() string
}
