package bootstrap

import (
	"context"
	"fmt"

	"detectviz-platform/internal/infrastructure/platform/config"
	"detectviz-platform/internal/infrastructure/platform/registry"
	"detectviz-platform/internal/infrastructure/platform/telemetry"
	"detectviz-platform/pkg/platform/contracts"
)

// PlatformInitializer 負責初始化平台的核心組件
// 職責: 按照正確的順序初始化所有平台服務和插件
type PlatformInitializer struct {
	configProvider   contracts.ConfigProvider
	loggerProvider   contracts.Logger
	registryProvider contracts.PluginRegistryProvider
}

// NewPlatformInitializer 創建新的平台初始化器
func NewPlatformInitializer() (*PlatformInitializer, error) {
	// 首先初始化配置提供者
	configProvider, err := config.NewViperConfigProvider("configs/app_config.yaml", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create config provider: %w", err)
	}

	// 然後初始化日誌提供者
	loggerProvider := telemetry.NewOtelZapLogger(map[string]interface{}{
		"level": "info",
	})

	// 最後初始化插件註冊表
	registryProvider := registry.NewPluginRegistryProvider(loggerProvider)

	return &PlatformInitializer{
		configProvider:   configProvider,
		loggerProvider:   loggerProvider,
		registryProvider: registryProvider,
	}, nil
}

// Initialize 初始化平台
func (p *PlatformInitializer) Initialize(ctx context.Context) error {
	p.loggerProvider.Info("開始初始化平台")

	// 這裡可以添加更多的初始化邏輯
	// 例如: 數據庫連接、插件加載等

	p.loggerProvider.Info("平台初始化完成")
	return nil
}

// GetConfigProvider 返回配置提供者
func (p *PlatformInitializer) GetConfigProvider() contracts.ConfigProvider {
	return p.configProvider
}

// GetLogger 返回日誌提供者
func (p *PlatformInitializer) GetLogger() contracts.Logger {
	return p.loggerProvider
}

// GetRegistry 返回插件註冊表
func (p *PlatformInitializer) GetRegistry() contracts.PluginRegistryProvider {
	return p.registryProvider
}
