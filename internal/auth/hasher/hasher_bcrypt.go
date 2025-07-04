package hasher

import (
	"context"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

// BcryptPasswordHasher 是 PasswordHasher 的 bcrypt 實現。
// 職責: 使用 bcrypt 演算法進行密碼散列和驗證。
type BcryptPasswordHasher struct {
	cost int // bcrypt 的成本參數，控制散列的複雜度
}

// NewBcryptPasswordHasher 創建一個新的 bcrypt 密碼散列器。
// 參數: cost - bcrypt 成本參數，建議使用 bcrypt.DefaultCost (10) 或更高
// 返回: BcryptPasswordHasher 實例和可能的錯誤
func NewBcryptPasswordHasher(cost int) (*BcryptPasswordHasher, error) {
	if cost < bcrypt.MinCost || cost > bcrypt.MaxCost {
		return nil, fmt.Errorf("invalid bcrypt cost: %d, must be between %d and %d",
			cost, bcrypt.MinCost, bcrypt.MaxCost)
	}

	return &BcryptPasswordHasher{
		cost: cost,
	}, nil
}

// NewDefaultBcryptPasswordHasher 創建一個使用默認成本的 bcrypt 密碼散列器。
// 返回: BcryptPasswordHasher 實例
func NewDefaultBcryptPasswordHasher() *BcryptPasswordHasher {
	return &BcryptPasswordHasher{
		cost: bcrypt.DefaultCost,
	}
}

// HashPassword 使用 bcrypt 散列明文密碼。
// 實現 PasswordHasher 介面。
func (h *BcryptPasswordHasher) HashPassword(ctx context.Context, plainPassword string) (string, error) {
	if plainPassword == "" {
		return "", fmt.Errorf("password cannot be empty")
	}

	// 檢查上下文是否已取消
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	default:
	}

	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(plainPassword), h.cost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}

	return string(hashedBytes), nil
}

// VerifyPassword 使用 bcrypt 驗證明文密碼與散列值是否匹配。
// 實現 PasswordHasher 介面。
func (h *BcryptPasswordHasher) VerifyPassword(ctx context.Context, plainPassword, hashedPassword string) (bool, error) {
	if plainPassword == "" {
		return false, fmt.Errorf("password cannot be empty")
	}
	if hashedPassword == "" {
		return false, fmt.Errorf("hashed password cannot be empty")
	}

	// 檢查上下文是否已取消
	select {
	case <-ctx.Done():
		return false, ctx.Err()
	default:
	}

	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(plainPassword))
	if err != nil {
		if err == bcrypt.ErrMismatchedHashAndPassword {
			return false, nil // 密碼不匹配，但不是錯誤
		}
		return false, fmt.Errorf("failed to verify password: %w", err)
	}

	return true, nil
}

// GetName 返回散列器的名稱。
// 實現 PasswordHasher 介面。
func (h *BcryptPasswordHasher) GetName() string {
	return fmt.Sprintf("bcrypt_hasher_cost_%d", h.cost)
}
