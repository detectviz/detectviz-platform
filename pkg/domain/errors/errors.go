package errors

import "errors"

// 領域錯誤 (Domain Errors)
// 定義領域層專屬的錯誤類型，這些錯誤是業務相關的，不應包含技術細節。

var (
	ErrUserNotFound          = errors.New("user not found")
	ErrInvalidUserFields     = errors.New("invalid user fields")
	ErrUserAlreadyExists     = errors.New("user already exists")
	ErrDatabaseOperation     = errors.New("database operation failed") // 這是領域層錯誤，但實際操作層會將其包裝
	ErrAuthenticationFailed  = errors.New("authentication failed")
	ErrUnauthorized          = errors.New("unauthorized access")
	ErrConfigLoadFailed      = errors.New("configuration loading failed")
	ErrPluginNotFound        = errors.New("plugin not found")
	ErrServiceNotInitialized = errors.New("service not initialized")
)
