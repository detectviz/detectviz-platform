package entities

import (
	"context"
	"errors"
	"time"

	"detectviz-platform/internal/infrastructure/platform/auth/hasher"
)

var ErrInvalidUserFields = errors.New("invalid user fields")

// User 是 Detectviz 平台的核心用戶實體。
// 職責: 封裝用戶的基本資訊及與用戶身份相關的業務行為。
// 測試說明: 領域實體應包含其內部的業務規則驗證，可獨立於數據庫和外部服務進行單元測試。
type User struct {
	ID           string
	Name         string
	Email        string
	PasswordHash string `json:"-"` // 儲存密碼的散列值，不在JSON序列化中暴露
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// NewUser 是一個工廠函數，用於創建新的用戶實體。
// 職責: 確保新創建的用戶實體符合領域的業務規則。
// 注意: 此函數接受已經散列的密碼，不應傳入明文密碼。
func NewUser(id, name, email, passwordHash string) (*User, error) {
	if name == "" || email == "" || passwordHash == "" {
		return nil, ErrInvalidUserFields
	}
	return &User{
		ID:           id,
		Name:         name,
		Email:        email,
		PasswordHash: passwordHash,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}, nil
}

// SetPassword 設置用戶密碼（加密存儲）
func (u *User) SetPassword(password string) error {
	passwordHasher := hasher.NewDefaultBcryptPasswordHasher()
	hash, err := passwordHasher.HashPassword(context.Background(), password)
	if err != nil {
		return err
	}
	u.PasswordHash = hash
	return nil
}

// CheckPassword 驗證用戶密碼
func (u *User) CheckPassword(password string) bool {
	passwordHasher := hasher.NewDefaultBcryptPasswordHasher()
	isValid, err := passwordHasher.VerifyPassword(context.Background(), password, u.PasswordHash)
	if err != nil {
		return false
	}
	return isValid
}
