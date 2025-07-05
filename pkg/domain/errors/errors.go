package errors

import (
	"errors"
	"fmt"
)

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

// 錯誤代碼常量
const (
	// 領域錯誤代碼
	ErrorCodeValidation = "VALIDATION_ERROR"
	ErrorCodeBusiness   = "BUSINESS_ERROR"
	ErrorCodePlugin     = "PLUGIN_ERROR"
	ErrorCodeAuth       = "AUTH_ERROR"
	ErrorCodeNotFound   = "NOT_FOUND_ERROR"

	// 基礎設施錯誤代碼
	ErrorCodeDatabase   = "DATABASE_ERROR"
	ErrorCodeNetwork    = "NETWORK_ERROR"
	ErrorCodeFileSystem = "FILESYSTEM_ERROR"
	ErrorCodeExternal   = "EXTERNAL_SERVICE_ERROR"
	ErrorCodeConfig     = "CONFIG_ERROR"
)

// DomainError 表示領域層錯誤的統一結構化類型
// 涵蓋驗證錯誤、業務邏輯錯誤、插件錯誤等所有領域相關錯誤
type DomainError struct {
	Code      string            `json:"code"`                // 錯誤代碼，用於區分具體錯誤類型
	Message   string            `json:"message"`             // 錯誤訊息
	Field     string            `json:"field,omitempty"`     // 相關欄位（驗證錯誤使用）
	Component string            `json:"component,omitempty"` // 相關組件（插件錯誤使用）
	Phase     string            `json:"phase,omitempty"`     // 相關階段（插件錯誤使用）
	Details   map[string]string `json:"details,omitempty"`   // 詳細資訊
}

func (e DomainError) Error() string {
	switch e.Code {
	case ErrorCodeValidation:
		if e.Field != "" {
			return fmt.Sprintf("validation failed for field '%s': %s", e.Field, e.Message)
		}
		return fmt.Sprintf("validation error: %s", e.Message)
	case ErrorCodePlugin:
		if e.Component != "" && e.Phase != "" {
			return fmt.Sprintf("plugin error in %s during %s: %s", e.Component, e.Phase, e.Message)
		}
		return fmt.Sprintf("plugin error: %s", e.Message)
	default:
		return fmt.Sprintf("domain error [%s]: %s", e.Code, e.Message)
	}
}

// NewValidationError 創建新的驗證錯誤
func NewValidationError(field, message string, details ...map[string]string) DomainError {
	var detailsMap map[string]string
	if len(details) > 0 {
		detailsMap = details[0]
	}
	return DomainError{
		Code:    ErrorCodeValidation,
		Message: message,
		Field:   field,
		Details: detailsMap,
	}
}

// NewBusinessError 創建新的業務錯誤
func NewBusinessError(message string, details ...map[string]string) DomainError {
	var detailsMap map[string]string
	if len(details) > 0 {
		detailsMap = details[0]
	}
	return DomainError{
		Code:    ErrorCodeBusiness,
		Message: message,
		Details: detailsMap,
	}
}

// NewPluginError 創建新的插件錯誤
func NewPluginError(pluginName, phase, message string, details ...map[string]string) DomainError {
	var detailsMap map[string]string
	if len(details) > 0 {
		detailsMap = details[0]
	}
	return DomainError{
		Code:      ErrorCodePlugin,
		Message:   message,
		Component: pluginName,
		Phase:     phase,
		Details:   detailsMap,
	}
}

// NewAuthError 創建新的認證錯誤
func NewAuthError(message string, details ...map[string]string) DomainError {
	var detailsMap map[string]string
	if len(details) > 0 {
		detailsMap = details[0]
	}
	return DomainError{
		Code:    ErrorCodeAuth,
		Message: message,
		Details: detailsMap,
	}
}

// NewNotFoundError 創建新的未找到錯誤
func NewNotFoundError(resource, message string, details ...map[string]string) DomainError {
	var detailsMap map[string]string
	if len(details) > 0 {
		detailsMap = details[0]
	}
	return DomainError{
		Code:      ErrorCodeNotFound,
		Message:   message,
		Component: resource,
		Details:   detailsMap,
	}
}

// InfrastructureError 表示基礎設施錯誤的結構化類型
// 涵蓋數據庫、網絡、文件系統、外部服務等基礎設施相關錯誤
type InfrastructureError struct {
	Code      string            `json:"code"`              // 錯誤代碼
	Component string            `json:"component"`         // 相關組件
	Operation string            `json:"operation"`         // 相關操作
	Message   string            `json:"message"`           // 錯誤訊息
	Details   map[string]string `json:"details,omitempty"` // 詳細資訊
}

func (e InfrastructureError) Error() string {
	if e.Component != "" && e.Operation != "" {
		return fmt.Sprintf("infrastructure error [%s] in %s.%s: %s", e.Code, e.Component, e.Operation, e.Message)
	}
	return fmt.Sprintf("infrastructure error [%s]: %s", e.Code, e.Message)
}

// NewDatabaseError 創建新的數據庫錯誤
func NewDatabaseError(component, operation, message string, details ...map[string]string) InfrastructureError {
	var detailsMap map[string]string
	if len(details) > 0 {
		detailsMap = details[0]
	}
	return InfrastructureError{
		Code:      ErrorCodeDatabase,
		Component: component,
		Operation: operation,
		Message:   message,
		Details:   detailsMap,
	}
}

// NewNetworkError 創建新的網絡錯誤
func NewNetworkError(component, operation, message string, details ...map[string]string) InfrastructureError {
	var detailsMap map[string]string
	if len(details) > 0 {
		detailsMap = details[0]
	}
	return InfrastructureError{
		Code:      ErrorCodeNetwork,
		Component: component,
		Operation: operation,
		Message:   message,
		Details:   detailsMap,
	}
}

// NewConfigError 創建新的配置錯誤
func NewConfigError(component, operation, message string, details ...map[string]string) InfrastructureError {
	var detailsMap map[string]string
	if len(details) > 0 {
		detailsMap = details[0]
	}
	return InfrastructureError{
		Code:      ErrorCodeConfig,
		Component: component,
		Operation: operation,
		Message:   message,
		Details:   detailsMap,
	}
}

// NewExternalServiceError 創建新的外部服務錯誤
func NewExternalServiceError(component, operation, message string, details ...map[string]string) InfrastructureError {
	var detailsMap map[string]string
	if len(details) > 0 {
		detailsMap = details[0]
	}
	return InfrastructureError{
		Code:      ErrorCodeExternal,
		Component: component,
		Operation: operation,
		Message:   message,
		Details:   detailsMap,
	}
}

// 錯誤類型檢查函數
func IsDomainError(err error) bool {
	_, ok := err.(DomainError)
	return ok
}

func IsInfrastructureError(err error) bool {
	_, ok := err.(InfrastructureError)
	return ok
}

func IsValidationError(err error) bool {
	if domainErr, ok := err.(DomainError); ok {
		return domainErr.Code == ErrorCodeValidation
	}
	return false
}

func IsBusinessError(err error) bool {
	if domainErr, ok := err.(DomainError); ok {
		return domainErr.Code == ErrorCodeBusiness
	}
	return false
}

func IsPluginError(err error) bool {
	if domainErr, ok := err.(DomainError); ok {
		return domainErr.Code == ErrorCodePlugin
	}
	return false
}

func IsAuthError(err error) bool {
	if domainErr, ok := err.(DomainError); ok {
		return domainErr.Code == ErrorCodeAuth
	}
	return false
}

func IsNotFoundError(err error) bool {
	if domainErr, ok := err.(DomainError); ok {
		return domainErr.Code == ErrorCodeNotFound
	}
	return false
}

func IsDatabaseError(err error) bool {
	if infraErr, ok := err.(InfrastructureError); ok {
		return infraErr.Code == ErrorCodeDatabase
	}
	return false
}

func IsNetworkError(err error) bool {
	if infraErr, ok := err.(InfrastructureError); ok {
		return infraErr.Code == ErrorCodeNetwork
	}
	return false
}

func IsConfigError(err error) bool {
	if infraErr, ok := err.(InfrastructureError); ok {
		return infraErr.Code == ErrorCodeConfig
	}
	return false
}

func IsExternalServiceError(err error) bool {
	if infraErr, ok := err.(InfrastructureError); ok {
		return infraErr.Code == ErrorCodeExternal
	}
	return false
}
