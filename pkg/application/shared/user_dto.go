package shared

import "time"

// CreateUserRequest 創建用戶請求 DTO
type CreateUserRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
}

// UpdateUserRequest 更新用戶請求 DTO
type UpdateUserRequest struct {
	ID    string `json:"id" validate:"required"`
	Email string `json:"email" validate:"required,email"`
}

// UserResponse 用戶響應 DTO
type UserResponse struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// AuthenticateRequest 用戶認證請求 DTO
type AuthenticateRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// ListUsersRequest 列出用戶請求 DTO
type ListUsersRequest struct {
	Limit  int `json:"limit" validate:"min=1,max=100"`
	Offset int `json:"offset" validate:"min=0"`
}
