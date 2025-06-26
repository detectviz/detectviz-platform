package http_server

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"detectviz-platform/pkg/platform/contracts"
)

// EchoHttpServerProvider 實現了 pkg/platform/contracts.HttpServerProvider 介面。
// 職責: 提供基於 Echo 框架的 HTTP 伺服器功能。
// 測試說明: 單元測試將驗證服務啟動、停止和路由註冊的正確性。
type EchoHttpServerProvider struct {
	echo   *echo.Echo
	config HttpServerConfig
	logger contracts.Logger
}

// HttpServerConfig 定義 HTTP 服務器的配置結構
type HttpServerConfig struct {
	Port         string `yaml:"port" json:"port"`
	ReadTimeout  string `yaml:"readTimeout" json:"readTimeout"`
	WriteTimeout string `yaml:"writeTimeout" json:"writeTimeout"`
}

// NewEchoHttpServerProvider 構造函數，根據配置創建 Echo HTTP 服務器實例。
func NewEchoHttpServerProvider(config map[string]interface{}, logger contracts.Logger) (contracts.HttpServerProvider, error) {
	// 解析配置
	serverConfig := HttpServerConfig{
		Port:         "8080",
		ReadTimeout:  "5s",
		WriteTimeout: "10s",
	}

	if port, ok := config["port"].(string); ok {
		serverConfig.Port = port
	}
	if portInt, ok := config["port"].(int); ok {
		serverConfig.Port = fmt.Sprintf("%d", portInt)
	}
	if readTimeout, ok := config["readTimeout"].(string); ok {
		serverConfig.ReadTimeout = readTimeout
	}
	if writeTimeout, ok := config["writeTimeout"].(string); ok {
		serverConfig.WriteTimeout = writeTimeout
	}

	// 創建 Echo 實例
	e := echo.New()

	// 設置中介層
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())

	// 註冊健康檢查端點
	e.GET("/health", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{
			"status":    "healthy",
			"service":   "detectviz-platform",
			"timestamp": time.Now().Format(time.RFC3339),
		})
	})

	// 註冊基本的 API 端點
	e.GET("/api/v1/info", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{
			"name":    "Detectviz Platform",
			"version": "0.1.0",
			"status":  "running",
		})
	})

	logger.Info("[INFRA][EchoHttpServerProvider] HTTP Server 初始化完成，端口: %s", serverConfig.Port)

	return &EchoHttpServerProvider{
		echo:   e,
		config: serverConfig,
		logger: logger,
	}, nil
}

func (h *EchoHttpServerProvider) GetName() string {
	return "echo_http_server"
}

// Start 啟動 HTTP 服務
func (h *EchoHttpServerProvider) Start(port string) error {
	if port != "" {
		h.config.Port = port
	}

	h.logger.Info("啟動 HTTP 服務器，端口: %s", h.config.Port)

	// 設置服務器超時
	h.echo.Server.ReadTimeout = h.parseTimeout(h.config.ReadTimeout, 5*time.Second)
	h.echo.Server.WriteTimeout = h.parseTimeout(h.config.WriteTimeout, 10*time.Second)

	// 啟動服務器
	address := ":" + h.config.Port
	if err := h.echo.Start(address); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("HTTP 服務器啟動失敗: %w", err)
	}

	return nil
}

// Stop 停止 HTTP 服務
func (h *EchoHttpServerProvider) Stop(ctx context.Context) error {
	h.logger.Info("正在停止 HTTP 服務器...")

	if err := h.echo.Shutdown(ctx); err != nil {
		return fmt.Errorf("HTTP 服務器停止失敗: %w", err)
	}

	h.logger.Info("HTTP 服務器已停止")
	return nil
}

// GetRouter 獲取底層路由實例，用於註冊路由和中介層
func (h *EchoHttpServerProvider) GetRouter() *echo.Echo {
	return h.echo
}

// parseTimeout 解析超時字符串為 time.Duration
func (h *EchoHttpServerProvider) parseTimeout(timeoutStr string, defaultTimeout time.Duration) time.Duration {
	if timeoutStr == "" {
		return defaultTimeout
	}

	timeout, err := time.ParseDuration(timeoutStr)
	if err != nil {
		h.logger.Warn("無法解析超時值 '%s'，使用預設值 %v", timeoutStr, defaultTimeout)
		return defaultTimeout
	}

	return timeout
}
