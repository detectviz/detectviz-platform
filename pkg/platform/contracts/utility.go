package contracts

import (
	"context"
)

// ErrorFactory 定義了錯誤創建和標準化的介面。
// 職責: 提供統一的、標準化的錯誤創建機制，確保錯誤包含錯誤碼、可讀訊息和詳細信息。
// AI_PLUGIN_TYPE: "error_factory_provider"
// AI_IMPL_PACKAGE: "detectviz-platform/internal/platform/providers/error_factory/standard_error_factory"
// AI_IMPL_CONSTRUCTOR: "NewStandardErrorFactory"
// @See: internal/platform/providers/error_factory/standard_error_factory.go
type ErrorFactory interface {
	// NewBadRequestError 創建一個表示客戶端錯誤（400）的標準錯誤。
	NewBadRequestError(message string, details ...map[string]any) error
	// NewNotFoundError 創建一個表示資源未找到（404）的標準錯誤。
	NewNotFoundError(message string, details ...map[string]any) error
	// NewUnauthorizedError 創建一個表示未授權（401）的標準錯誤。
	NewUnauthorizedError(message string, details ...map[string]any) error
	// NewInternalServerError 創建一個表示服務器內部錯誤（500）的標準錯誤。
	NewInternalServerError(message string, details ...map[string]any) error
	// NewErrorf 類似 fmt.Errorf，但返回一個標準的內部服務器錯誤類型。
	NewErrorf(format string, args ...any) error
	// GetName 返回錯誤工廠的名稱。
	GetName() string
}

// ServiceDiscoveryProvider 定義了服務發現的介面。
// 職責: 在分佈式系統中註冊、註銷服務實例，並查詢可用服務實例的地址。
// AI_PLUGIN_TYPE: "service_discovery_provider"
// AI_IMPL_PACKAGE: "detectviz-platform/internal/platform/providers/service_discovery/k8s_discovery"
// AI_IMPL_CONSTRUCTOR: "NewK8sServiceDiscoveryProvider"
// @See: internal/platform/providers/service_discovery/k8s_discovery.go
type ServiceDiscoveryProvider interface {
	// RegisterService 註冊一個新的服務實例。
	RegisterService(ctx context.Context, serviceName string, instanceID string, address string, port int, metadata map[string]string) error
	// DeregisterService 註銷一個服務實例。
	DeregisterService(ctx context.Context, serviceName string, instanceID string) error
	// GetInstances 獲取指定服務的所有健康實例。
	GetInstances(ctx context.Context, serviceName string) ([]ServiceInstance, error)
	// GetName 返回服務發現提供者的名稱。
	GetName() string
}
