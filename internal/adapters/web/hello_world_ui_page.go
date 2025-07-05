package web

import (
	"context"
	"fmt"

	"github.com/labstack/echo/v4"

	"detectviz-platform/pkg/domain/interfaces/plugins"
	"detectviz-platform/pkg/platform/contracts"
)

// HelloWorldUIPagePlugin 實現了 pkg/domain/interfaces/plugins.UIPagePlugin 介面。
// 職責: 提供一個簡單的 Hello World Web UI 頁面。
// 測試說明: 單元測試將驗證路由註冊和頁面內容生成的正確性。
type HelloWorldUIPagePlugin struct {
	logger contracts.Logger
	config HelloWorldConfig
}

// HelloWorldConfig 定義 Hello World UI 頁面的配置結構
type HelloWorldConfig struct {
	Route   string `yaml:"route" json:"route"`
	Title   string `yaml:"title" json:"title"`
	Message string `yaml:"message" json:"message"`
}

// NewHelloWorldUIPagePlugin 構造函數，創建新的 Hello World UI 頁面插件實例。
func NewHelloWorldUIPagePlugin(config map[string]interface{}, logger contracts.Logger) (plugins.UIPagePlugin, error) {
	// 解析配置
	uiConfig := HelloWorldConfig{
		Route:   "/ui/hello",
		Title:   "Hello World - Detectviz Platform",
		Message: "歡迎使用 Detectviz 平台！這是一個示例 UI 頁面。",
	}

	if route, ok := config["route"].(string); ok {
		uiConfig.Route = route
	}
	if title, ok := config["title"].(string); ok {
		uiConfig.Title = title
	}
	if message, ok := config["message"].(string); ok {
		uiConfig.Message = message
	}

	logger.Info("[PLUGIN][HelloWorldUIPagePlugin] Hello World UI 頁面插件初始化完成，路由: %s", uiConfig.Route)

	return &HelloWorldUIPagePlugin{
		logger: logger,
		config: uiConfig,
	}, nil
}

func (h *HelloWorldUIPagePlugin) GetName() string {
	return "hello_world_ui_page"
}

// Init 插件初始化，接收配置
func (h *HelloWorldUIPagePlugin) Init(ctx context.Context, cfg map[string]interface{}) error {
	h.logger.Info("初始化 Hello World UI 頁面插件")
	return nil
}

// Start 插件啟動，例如啟動背景任務
func (h *HelloWorldUIPagePlugin) Start(ctx context.Context) error {
	h.logger.Info("啟動 Hello World UI 頁面插件")
	return nil
}

// Stop 插件停止，清理資源
func (h *HelloWorldUIPagePlugin) Stop(ctx context.Context) error {
	h.logger.Info("停止 Hello World UI 頁面插件")
	return nil
}

// GetRoute 返回該 UI 頁面對應的 URL 路徑
func (h *HelloWorldUIPagePlugin) GetRoute() string {
	return h.config.Route
}

// GetHTMLContent 返回該 UI 頁面所需的 HTML 內容
func (h *HelloWorldUIPagePlugin) GetHTMLContent() string {
	return fmt.Sprintf(`
<!DOCTYPE html>
<html lang="zh-TW">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>%s</title>
    <style>
        body {
            font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif;
            background: linear-gradient(135deg, #667eea 0%%, #764ba2 100%%);
            margin: 0;
            padding: 20px;
            min-height: 100vh;
            display: flex;
            align-items: center;
            justify-content: center;
        }
        .container {
            background: white;
            border-radius: 15px;
            padding: 40px;
            box-shadow: 0 20px 40px rgba(0,0,0,0.1);
            text-align: center;
            max-width: 600px;
        }
        .logo {
            font-size: 3em;
            color: #667eea;
            margin-bottom: 20px;
        }
        h1 {
            color: #333;
            margin-bottom: 20px;
        }
        .message {
            color: #666;
            font-size: 1.2em;
            line-height: 1.6;
            margin-bottom: 30px;
        }
        .features {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
            gap: 20px;
            margin-top: 30px;
        }
        .feature {
            background: #f8f9fa;
            padding: 20px;
            border-radius: 10px;
            border-left: 4px solid #667eea;
        }
        .feature h3 {
            margin: 0 0 10px 0;
            color: #333;
        }
        .feature p {
            margin: 0;
            color: #666;
            font-size: 0.9em;
        }
        .api-links {
            margin-top: 30px;
            padding-top: 20px;
            border-top: 1px solid #eee;
        }
        .api-links a {
            display: inline-block;
            margin: 0 10px;
            padding: 10px 20px;
            background: #667eea;
            color: white;
            text-decoration: none;
            border-radius: 5px;
            transition: background 0.3s;
        }
        .api-links a:hover {
            background: #764ba2;
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="logo">🔍</div>
        <h1>%s</h1>
        <div class="message">%s</div>
        
        <div class="features">
            <div class="feature">
                <h3>插件系統</h3>
                <p>一切皆插件的設計理念，支持動態載入和配置</p>
            </div>
            <div class="feature">
                <h3>Clean Architecture</h3>
                <p>嚴格分層，依賴反轉，確保代碼的可維護性</p>
            </div>
            <div class="feature">
                <h3>AI 驅動</h3>
                <p>為 AI 輔助開發和自動化奠定基礎</p>
            </div>
        </div>
        
        <div class="api-links">
            <a href="/health">健康檢查</a>
            <a href="/api/v1/info">API 資訊</a>
        </div>
    </div>
</body>
</html>`, h.config.Title, h.config.Title, h.config.Message)
}

// RegisterRoute 將插件的 HTTP 路由註冊到平台 HTTP 伺服器實例
func (h *HelloWorldUIPagePlugin) RegisterRoute(router interface{}, logger interface{}) error {
	// 類型斷言獲取 Echo 路由器和日誌器
	echoRouter, ok := router.(*echo.Echo)
	if !ok {
		return fmt.Errorf("router is not an *echo.Echo instance")
	}

	contractsLogger, ok := logger.(contracts.Logger)
	if !ok {
		return fmt.Errorf("logger is not a contracts.Logger instance")
	}

	// 註冊路由
	echoRouter.GET(h.config.Route, func(c echo.Context) error {
		return c.HTML(200, h.GetHTMLContent())
	})

	contractsLogger.Info("[PLUGIN][HelloWorldUIPagePlugin] 已註冊路由: %s", h.config.Route)

	return nil
}
