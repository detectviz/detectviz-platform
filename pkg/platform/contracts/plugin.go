package contracts

import "context"

// PluginRegistryProvider 定義了插件註冊與查詢的介面。
// 職責: 作為平台的核心，管理所有已載入和可用的插件實例，提供統一的查詢和元數據獲取功能。
// AI_PLUGIN_TYPE: "plugin_registry_provider"
// AI_IMPL_PACKAGE: "detectviz-platform/internal/infrastructure/platform/registry/plugin_registry_provider"
// AI_IMPL_CONSTRUCTOR: "NewPluginRegistryProvider"
// @See: internal/infrastructure/platform/registry/plugin_registry_provider.go
type PluginRegistryProvider interface {
	// Register 註冊一個具名的插件實例。
	Register(name string, provider any) error
	// Get 根據名稱獲取指定的插件實例。
	Get(name string) (any, error)
	// List 列出所有已註冊插件的名稱。
	List() []string
	// GetMetadata 返回特定插件的描述資訊（版本、作者、狀態等）。
	GetMetadata(name string) (map[string]any, error)
	// GetName 返回插件註冊表的名稱，例如 "core_registry"。
	GetName() string
}

// PluginMetadataProvider 定義了插件元資訊的查詢與註冊介面。
// 職責: 提供插件名稱、版本、依賴等元數據的存儲和查詢功能，有利於平台治理和可觀測性。
// AI_PLUGIN_TYPE: "plugin_metadata_provider"
// AI_IMPL_PACKAGE: "detectviz-platform/internal/platform/providers/plugin_metadata/in_memory_plugin_metadata"
// AI_IMPL_CONSTRUCTOR: "NewInMemoryPluginMetadataProvider"
// @See: internal/platform/providers/plugin_metadata/in_memory_plugin_metadata.go
type PluginMetadataProvider interface {
	// GetMetadata 根據插件名稱檢索其元數據。
	GetMetadata(ctx context.Context, pluginName string) (map[string]any, error)
	// RegisterMetadata 為指定的插件註冊元數據。
	RegisterMetadata(ctx context.Context, pluginName string, metadata map[string]any) error
	// GetName 返回元數據提供者的名稱。
	GetName() string
}
