package hasher

import "context"

// PasswordHasher 定義了密碼散列和驗證的抽象介面。
// 職責: 提供安全的密碼散列和驗證功能，隔離具體的散列演算法實現。
// AI_PLUGIN_TYPE: "password_hasher"
// AI_IMPL_PACKAGE: "detectviz-platform/internal/auth/hasher/bcrypt_hasher"
// AI_IMPL_CONSTRUCTOR: "NewBcryptPasswordHasher"
type PasswordHasher interface {
	// HashPassword 將明文密碼轉換為安全的散列值
	// 參數: ctx - 上下文，plainPassword - 明文密碼
	// 返回: 散列後的密碼字符串和可能的錯誤
	HashPassword(ctx context.Context, plainPassword string) (string, error)

	// VerifyPassword 驗證明文密碼是否與散列值匹配
	// 參數: ctx - 上下文，plainPassword - 明文密碼，hashedPassword - 散列值
	// 返回: 是否匹配的布爾值和可能的錯誤
	VerifyPassword(ctx context.Context, plainPassword, hashedPassword string) (bool, error)

	// GetName 返回散列器的名稱，用於識別和日誌記錄
	GetName() string
}
