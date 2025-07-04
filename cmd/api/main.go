package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"detectviz-platform/internal/adapters/plugins/web_ui"
	"detectviz-platform/internal/infrastructure/platform/config"
	"detectviz-platform/internal/infrastructure/platform/http_server"
	"detectviz-platform/internal/infrastructure/platform/logger"
	"detectviz-platform/internal/infrastructure/platform/registry"
)

func main() {
	// 步驟 1: 創建基礎配置提供者 (使用默認配置進行引導)
	bootstrapConfigProvider, err := config.NewViperConfigProvider("configs/app_config.yaml", nil)
	if err != nil {
		log.Fatalf("無法創建引導配置提供者: %v", err)
	}

	// 步驟 2: 從配置文件讀取日誌器配置
	loggerConfig := map[string]interface{}{
		"level":            bootstrapConfigProvider.GetString("logger.level"),
		"encoding":         bootstrapConfigProvider.GetString("logger.encoding"),
		"outputPaths":      []string{"stdout"}, // 可以從配置中讀取，但這裡簡化處理
		"errorOutputPaths": []string{"stderr"},
		"initialFields": map[string]interface{}{
			"service":   bootstrapConfigProvider.GetString("logger.initialFields.service"),
			"component": bootstrapConfigProvider.GetString("logger.initialFields.component"),
		},
	}

	otelZapLogger := logger.NewOtelZapLogger(loggerConfig)

	// 步驟 3: 創建主配置提供者
	configProvider, err := config.NewViperConfigProvider("configs/composition.yaml", otelZapLogger)
	if err != nil {
		log.Fatalf("無法創建配置提供者: %v", err)
	}
	otelZapLogger.Info("[主程序] 配置提供者創建完成")

	otelZapLogger.Info("[主程序] Detectviz 平台啟動中...")
	otelZapLogger.Info("[主程序] 日誌器初始化完成")

	// 步驟 4: 創建插件註冊表
	pluginRegistry := registry.NewPluginRegistryProvider(otelZapLogger)
	otelZapLogger.Info("[主程序] 插件註冊表創建完成")

	// 註冊核心組件到插件註冊表
	err = pluginRegistry.Register("configProvider", configProvider)
	if err != nil {
		otelZapLogger.Error("註冊配置提供者失敗: %v", err)
		os.Exit(1)
	}

	err = pluginRegistry.Register("logger", otelZapLogger)
	if err != nil {
		otelZapLogger.Error("註冊日誌器失敗: %v", err)
		os.Exit(1)
	}

	err = pluginRegistry.Register("pluginRegistry", pluginRegistry)
	if err != nil {
		otelZapLogger.Error("註冊插件註冊表失敗: %v", err)
		os.Exit(1)
	}

	// 步驟 5: 創建 HTTP 服務器 (從配置文件讀取)
	httpServerConfig := map[string]interface{}{
		"port":         bootstrapConfigProvider.GetInt("server.port"),
		"readTimeout":  bootstrapConfigProvider.GetString("server.readTimeout"),
		"writeTimeout": bootstrapConfigProvider.GetString("server.writeTimeout"),
	}

	httpServer, err := http_server.NewEchoHttpServerProvider(httpServerConfig, otelZapLogger)
	if err != nil {
		otelZapLogger.Error("創建 HTTP 服務器失敗: %v", err)
		os.Exit(1)
	}

	err = pluginRegistry.Register("httpServer", httpServer)
	if err != nil {
		otelZapLogger.Error("註冊 HTTP 服務器失敗: %v", err)
		os.Exit(1)
	}

	otelZapLogger.Info("[主程序] HTTP 服務器創建完成")

	// 步驟 6: 創建並註冊 Hello World UI 頁面插件 (從配置文件讀取)
	helloUIConfig := map[string]interface{}{
		"route":   bootstrapConfigProvider.GetString("ui.helloWorld.route"),
		"title":   bootstrapConfigProvider.GetString("ui.helloWorld.title"),
		"message": bootstrapConfigProvider.GetString("ui.helloWorld.message"),
	}

	helloWorldUI, err := web_ui.NewHelloWorldUIPagePlugin(helloUIConfig, otelZapLogger)
	if err != nil {
		otelZapLogger.Error("創建 Hello World UI 插件失敗: %v", err)
		os.Exit(1)
	}

	err = pluginRegistry.Register("helloWorldUI", helloWorldUI)
	if err != nil {
		otelZapLogger.Error("註冊 Hello World UI 插件失敗: %v", err)
		os.Exit(1)
	}

	otelZapLogger.Info("[主程序] Hello World UI 頁面插件創建完成")

	// 步驟 7: 將 UI 插件路由註冊到 HTTP 服務器
	echoHttpServer, ok := httpServer.(*http_server.EchoHttpServerProvider)
	if !ok {
		otelZapLogger.Error("HTTP 服務器類型轉換失敗")
		os.Exit(1)
	}

	err = helloWorldUI.RegisterRoute(echoHttpServer.GetRouter(), otelZapLogger)
	if err != nil {
		otelZapLogger.Error("註冊 UI 路由失敗: %v", err)
		os.Exit(1)
	}

	otelZapLogger.Info("[主程序] UI 路由註冊完成")

	// 步驟 8: 打印註冊的插件列表
	registeredPlugins := pluginRegistry.List()
	otelZapLogger.Info("[主程序] 已註冊插件列表: %v", registeredPlugins)

	// 顯示插件詳細信息
	for _, pluginName := range registeredPlugins {
		metadata, err := pluginRegistry.GetMetadata(pluginName)
		if err == nil {
			otelZapLogger.Info("[主程序] 插件 %s 元數據: %v", pluginName, metadata)
		}
	}

	// 步驟 9: 啟動 HTTP 服務器 (背景執行)
	serverPort := bootstrapConfigProvider.GetString("server.port")
	if serverPort == "" {
		serverPort = "8080" // 默認端口
	}
	go func() {
		otelZapLogger.Info("[主程序] 正在啟動 HTTP 服務器，端口: %s", serverPort)
		if err := httpServer.Start(serverPort); err != nil {
			otelZapLogger.Error("HTTP 服務器啟動失敗: %v", err)
		}
	}()

	otelZapLogger.Info("[主程序] Detectviz 平台啟動完成")
	otelZapLogger.Info("[主程序] 可存取的端點:")
	otelZapLogger.Info("[主程序]   - 健康檢查: http://localhost:%s/health", serverPort)
	otelZapLogger.Info("[主程序]   - API 資訊: http://localhost:%s/api/v1/info", serverPort)
	otelZapLogger.Info("[主程序]   - Hello World UI: http://localhost:%s%s", serverPort, bootstrapConfigProvider.GetString("ui.helloWorld.route"))

	// 步驟 10: 等待中斷信號
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	otelZapLogger.Info("[主程序] 正在關閉 Detectviz 平台...")

	// 步驟 11: 優雅關閉 (使用配置的超時時間)
	shutdownTimeout := bootstrapConfigProvider.GetString("server.shutdownTimeout")
	if shutdownTimeout == "" {
		shutdownTimeout = "30s" // 默認超時時間
	}
	timeout, err := time.ParseDuration(shutdownTimeout)
	if err != nil {
		timeout = 30 * time.Second // 解析失敗時使用默認值
		otelZapLogger.Error("解析關閉超時時間失敗，使用默認值: %v", err)
	}
	shutdownCtx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	if err := httpServer.Stop(shutdownCtx); err != nil {
		otelZapLogger.Error("HTTP 服務器關閉失敗: %v", err)
	}

	otelZapLogger.Info("[主程序] Detectviz 平台已關閉")
}
