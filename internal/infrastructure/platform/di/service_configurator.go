package di

import (
	"detectviz-platform/internal/adapters/plugins/web_ui"
	"detectviz-platform/internal/infrastructure/platform/config"
	"detectviz-platform/internal/infrastructure/platform/health"
	"detectviz-platform/internal/infrastructure/platform/http_server"
	"detectviz-platform/internal/infrastructure/platform/logger"
	"detectviz-platform/internal/infrastructure/platform/registry"
	"detectviz-platform/pkg/platform/contracts"
	"time"
)

// ServiceConfigurator 負責配置和註冊所有平台服務
type ServiceConfigurator struct {
	container *Container
}

// NewServiceConfigurator 創建新的服務配置器
func NewServiceConfigurator(container *Container) *ServiceConfigurator {
	return &ServiceConfigurator{
		container: container,
	}
}

// ConfigureServices 配置所有平台服務
func (sc *ServiceConfigurator) ConfigureServices() error {
	// 註冊配置提供者工廠
	err := sc.container.RegisterSingleton((*contracts.ConfigProvider)(nil), func() (contracts.ConfigProvider, error) {
		// 首先創建引導配置提供者
		return config.NewViperConfigProvider("configs/app_config.yaml", nil)
	})
	if err != nil {
		return err
	}

	// 註冊日誌器工廠
	err = sc.container.RegisterSingleton((*contracts.Logger)(nil), func(configProvider contracts.ConfigProvider) (contracts.Logger, error) {
		loggerConfig := map[string]interface{}{
			"level":            configProvider.GetString("logger.level"),
			"encoding":         configProvider.GetString("logger.encoding"),
			"outputPaths":      []string{"stdout"},
			"errorOutputPaths": []string{"stderr"},
			"initialFields": map[string]interface{}{
				"service":   configProvider.GetString("logger.initialFields.service"),
				"component": configProvider.GetString("logger.initialFields.component"),
			},
		}
		return logger.NewOtelZapLogger(loggerConfig), nil
	})
	if err != nil {
		return err
	}

	// 註冊主配置提供者工廠
	err = sc.container.RegisterSingleton((*contracts.ConfigProvider)(nil), func(logger contracts.Logger) (contracts.ConfigProvider, error) {
		return config.NewViperConfigProvider("configs/composition.yaml", logger)
	})
	if err != nil {
		return err
	}

	// 註冊插件註冊表工廠
	err = sc.container.RegisterSingleton((*contracts.PluginRegistryProvider)(nil), func(logger contracts.Logger) (contracts.PluginRegistryProvider, error) {
		return registry.NewPluginRegistryProvider(logger), nil
	})
	if err != nil {
		return err
	}

	// 註冊 HTTP 服務器工廠
	err = sc.container.RegisterSingleton((*contracts.HttpServerProvider)(nil), func(configProvider contracts.ConfigProvider, logger contracts.Logger) (contracts.HttpServerProvider, error) {
		httpServerConfig := map[string]interface{}{
			"port":         configProvider.GetInt("server.port"),
			"readTimeout":  configProvider.GetString("server.readTimeout"),
			"writeTimeout": configProvider.GetString("server.writeTimeout"),
		}
		return http_server.NewEchoHttpServerProvider(httpServerConfig, logger)
	})
	if err != nil {
		return err
	}

	// 註冊健康檢查管理器工廠
	err = sc.container.RegisterSingleton((*health.HealthCheckManager)(nil), func(logger contracts.Logger) (*health.HealthCheckManager, error) {
		return health.NewHealthCheckManager(logger, 30*time.Second), nil
	})
	if err != nil {
		return err
	}

	// 註冊 Hello World UI 插件工廠
	err = sc.container.RegisterSingleton((*web_ui.HelloWorldUIPagePlugin)(nil), func(configProvider contracts.ConfigProvider, logger contracts.Logger) (*web_ui.HelloWorldUIPagePlugin, error) {
		helloUIConfig := map[string]interface{}{
			"route":   configProvider.GetString("ui.helloWorld.route"),
			"title":   configProvider.GetString("ui.helloWorld.title"),
			"message": configProvider.GetString("ui.helloWorld.message"),
		}
		plugin, err := web_ui.NewHelloWorldUIPagePlugin(helloUIConfig, logger)
		if err != nil {
			return nil, err
		}
		return plugin.(*web_ui.HelloWorldUIPagePlugin), nil
	})
	if err != nil {
		return err
	}

	return nil
}

// GetService 從容器中獲取服務
func (sc *ServiceConfigurator) GetService(serviceType interface{}) (interface{}, error) {
	return sc.container.Resolve(serviceType)
}

// GetContainer 獲取底層的 DI 容器
func (sc *ServiceConfigurator) GetContainer() *Container {
	return sc.container
}
