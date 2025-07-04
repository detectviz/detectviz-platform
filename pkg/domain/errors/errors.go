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

// ValidationError 表示驗證錯誤的結構化類型
type ValidationError struct {
	Field   string            `json:"field"`
	Message string            `json:"message"`
	Details map[string]string `json:"details,omitempty"`
}

func (e ValidationError) Error() string {
	if e.Details != nil && len(e.Details) > 0 {
		return fmt.Sprintf("validation failed for field '%s': %s (details: %v)", e.Field, e.Message, e.Details)
	}
	return fmt.Sprintf("validation failed for field '%s': %s", e.Field, e.Message)
}

// NewValidationError 創建新的驗證錯誤
func NewValidationError(field, message string, details ...map[string]string) ValidationError {
	var detailsMap map[string]string
	if len(details) > 0 {
		detailsMap = details[0]
	}
	return ValidationError{
		Field:   field,
		Message: message,
		Details: detailsMap,
	}
}

// BusinessError 表示業務邏輯錯誤的結構化類型
type BusinessError struct {
	Code    string            `json:"code"`
	Message string            `json:"message"`
	Details map[string]string `json:"details,omitempty"`
}

func (e BusinessError) Error() string {
	if e.Details != nil && len(e.Details) > 0 {
		return fmt.Sprintf("business error [%s]: %s (details: %v)", e.Code, e.Message, e.Details)
	}
	return fmt.Sprintf("business error [%s]: %s", e.Code, e.Message)
}

// NewBusinessError 創建新的業務錯誤
func NewBusinessError(code, message string, details ...map[string]string) BusinessError {
	var detailsMap map[string]string
	if len(details) > 0 {
		detailsMap = details[0]
	}
	return BusinessError{
		Code:    code,
		Message: message,
		Details: detailsMap,
	}
}

// InfrastructureError 表示基礎設施錯誤的結構化類型
type InfrastructureError struct {
	Component string            `json:"component"`
	Operation string            `json:"operation"`
	Message   string            `json:"message"`
	Details   map[string]string `json:"details,omitempty"`
}

func (e InfrastructureError) Error() string {
	if e.Details != nil && len(e.Details) > 0 {
		return fmt.Sprintf("infrastructure error in %s.%s: %s (details: %v)", e.Component, e.Operation, e.Message, e.Details)
	}
	return fmt.Sprintf("infrastructure error in %s.%s: %s", e.Component, e.Operation, e.Message)
}

// NewInfrastructureError 創建新的基礎設施錯誤
func NewInfrastructureError(component, operation, message string, details ...map[string]string) InfrastructureError {
	var detailsMap map[string]string
	if len(details) > 0 {
		detailsMap = details[0]
	}
	return InfrastructureError{
		Component: component,
		Operation: operation,
		Message:   message,
		Details:   detailsMap,
	}
}

// PluginError 表示插件相關錯誤的結構化類型
type PluginError struct {
	PluginName string            `json:"plugin_name"`
	Phase      string            `json:"phase"` // Init, Start, Stop, Execute
	Message    string            `json:"message"`
	Details    map[string]string `json:"details,omitempty"`
}

func (e PluginError) Error() string {
	if e.Details != nil && len(e.Details) > 0 {
		return fmt.Sprintf("plugin error in %s during %s: %s (details: %v)", e.PluginName, e.Phase, e.Message, e.Details)
	}
	return fmt.Sprintf("plugin error in %s during %s: %s", e.PluginName, e.Phase, e.Message)
}

// NewPluginError 創建新的插件錯誤
func NewPluginError(pluginName, phase, message string, details ...map[string]string) PluginError {
	var detailsMap map[string]string
	if len(details) > 0 {
		detailsMap = details[0]
	}
	return PluginError{
		PluginName: pluginName,
		Phase:      phase,
		Message:    message,
		Details:    detailsMap,
	}
}

// IsValidationError 檢查錯誤是否為驗證錯誤
func IsValidationError(err error) bool {
	_, ok := err.(ValidationError)
	return ok
}

// IsBusinessError 檢查錯誤是否為業務錯誤
func IsBusinessError(err error) bool {
	_, ok := err.(BusinessError)
	return ok
}

// IsInfrastructureError 檢查錯誤是否為基礎設施錯誤
func IsInfrastructureError(err error) bool {
	_, ok := err.(InfrastructureError)
	return ok
}

// IsPluginError 檢查錯誤是否為插件錯誤
func IsPluginError(err error) bool {
	_, ok := err.(PluginError)
	return ok
}
