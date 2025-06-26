package entities

import (
	"errors"
	"time"
)

var ErrInvalidUserFields = errors.New("invalid user fields")

// User 是 Detectviz 平台的核心用戶實體。
// 職責: 封裝用戶的基本資訊及與用戶身份相關的業務行為。
// 測試說明: 領域實體應包含其內部的業務規則驗證，可獨立於數據庫和外部服務進行單元測試。
type User struct {
	ID        string
	Name      string
	Email     string
	Password  string // 在領域層，密碼通常是業務概念上的密碼，具體儲存形式(如散列)由持久化層處理。
	CreatedAt time.Time
	UpdatedAt time.Time
}

// NewUser 是一個工廠函數，用於創建新的用戶實體。
// 職責: 確保新創建的用戶實體符合領域的業務規則。
func NewUser(id, name, email, password string) (*User, error) {
	if name == "" || email == "" || password == "" {
		return nil, ErrInvalidUserFields
	}
	return &User{
		ID:        id,
		Name:      name,
		Email:     email,
		Password:  password,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}, nil
}
