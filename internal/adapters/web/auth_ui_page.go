package web

import (
	"context"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"

	"detectviz-platform/pkg/domain/interfaces/plugins"
	"detectviz-platform/pkg/platform/contracts"
)

// AuthUIPagePlugin 實現了用戶登錄和註冊的 Web UI 頁面
// 職責: 提供用戶身份驗證相關的 Web 界面
type AuthUIPagePlugin struct {
	logger       contracts.Logger
	authProvider contracts.AuthProvider
	config       AuthUIConfig
}

// AuthUIConfig 定義認證 UI 頁面的配置
type AuthUIConfig struct {
	LoginRoute    string `yaml:"login_route" json:"login_route"`
	RegisterRoute string `yaml:"register_route" json:"register_route"`
	LogoutRoute   string `yaml:"logout_route" json:"logout_route"`
	Title         string `yaml:"title" json:"title"`
	BrandName     string `yaml:"brand_name" json:"brand_name"`
}

// NewAuthUIPagePlugin 創建新的認證 UI 頁面插件實例
func NewAuthUIPagePlugin(authProvider contracts.AuthProvider, logger contracts.Logger) plugins.UIPagePlugin {
	config := AuthUIConfig{
		LoginRoute:    "/auth/login",
		RegisterRoute: "/auth/register",
		LogoutRoute:   "/auth/logout",
		Title:         "Detectviz 平台 - 用戶認證",
		BrandName:     "Detectviz",
	}

	logger.Info("初始化認證 UI 頁面插件",
		"login_route", config.LoginRoute,
		"register_route", config.RegisterRoute)

	return &AuthUIPagePlugin{
		logger:       logger,
		authProvider: authProvider,
		config:       config,
	}
}

func (a *AuthUIPagePlugin) GetName() string {
	return "auth_ui_page"
}

// Init 插件初始化
func (a *AuthUIPagePlugin) Init(ctx context.Context, cfg map[string]interface{}) error {
	a.logger.Info("初始化認證 UI 頁面插件")

	// 解析配置
	if loginRoute, ok := cfg["login_route"].(string); ok {
		a.config.LoginRoute = loginRoute
	}
	if registerRoute, ok := cfg["register_route"].(string); ok {
		a.config.RegisterRoute = registerRoute
	}
	if title, ok := cfg["title"].(string); ok {
		a.config.Title = title
	}
	if brandName, ok := cfg["brand_name"].(string); ok {
		a.config.BrandName = brandName
	}

	return nil
}

// Start 插件啟動
func (a *AuthUIPagePlugin) Start(ctx context.Context) error {
	a.logger.Info("啟動認證 UI 頁面插件")
	return nil
}

// Stop 插件停止
func (a *AuthUIPagePlugin) Stop(ctx context.Context) error {
	a.logger.Info("停止認證 UI 頁面插件")
	return nil
}

// GetRoute 返回主要的登錄頁面路由
func (a *AuthUIPagePlugin) GetRoute() string {
	return a.config.LoginRoute
}

// GetHTMLContent 返回登錄頁面的 HTML 內容
func (a *AuthUIPagePlugin) GetHTMLContent() string {
	return a.generateLoginPageHTML()
}

// RegisterRoute 註冊所有認證相關的路由
func (a *AuthUIPagePlugin) RegisterRoute(router interface{}, logger interface{}) error {
	echoRouter, ok := router.(*echo.Echo)
	if !ok {
		return fmt.Errorf("expected *echo.Echo, got %T", router)
	}

	// 註冊登錄頁面
	echoRouter.GET(a.config.LoginRoute, a.handleLoginPage)
	echoRouter.POST(a.config.LoginRoute, a.handleLoginSubmit)

	// 註冊註冊頁面
	echoRouter.GET(a.config.RegisterRoute, a.handleRegisterPage)
	echoRouter.POST(a.config.RegisterRoute, a.handleRegisterSubmit)

	// 註冊登出處理
	echoRouter.POST(a.config.LogoutRoute, a.handleLogout)

	a.logger.Info("認證路由註冊完成",
		"login_route", a.config.LoginRoute,
		"register_route", a.config.RegisterRoute,
		"logout_route", a.config.LogoutRoute)

	return nil
}

// handleLoginPage 處理登錄頁面請求
func (a *AuthUIPagePlugin) handleLoginPage(c echo.Context) error {
	html := a.generateLoginPageHTML()
	return c.HTML(http.StatusOK, html)
}

// handleLoginSubmit 處理登錄表單提交
func (a *AuthUIPagePlugin) handleLoginSubmit(c echo.Context) error {
	username := c.FormValue("username")
	password := c.FormValue("password")

	if username == "" || password == "" {
		return c.HTML(http.StatusBadRequest, a.generateLoginPageHTML("用戶名和密碼不能為空"))
	}

	// 使用認證提供者驗證用戶
	credentials := fmt.Sprintf("%s:%s", username, password)
	userID, err := a.authProvider.Authenticate(c.Request().Context(), credentials)
	if err != nil {
		a.logger.Warn("用戶登錄失敗", "username", username, "error", err)
		return c.HTML(http.StatusUnauthorized, a.generateLoginPageHTML("用戶名或密碼錯誤"))
	}

	a.logger.Info("用戶登錄成功", "username", username, "user_id", userID)

	// 設置會話或 JWT token（簡化實現）
	// 在實際實現中，這裡應該設置適當的會話管理
	c.SetCookie(&http.Cookie{
		Name:     "user_id",
		Value:    userID,
		Path:     "/",
		HttpOnly: true,
		MaxAge:   3600, // 1小時
	})

	// 重定向到主頁
	return c.Redirect(http.StatusFound, "/ui/hello")
}

// handleRegisterPage 處理註冊頁面請求
func (a *AuthUIPagePlugin) handleRegisterPage(c echo.Context) error {
	html := a.generateRegisterPageHTML()
	return c.HTML(http.StatusOK, html)
}

// handleRegisterSubmit 處理註冊表單提交
func (a *AuthUIPagePlugin) handleRegisterSubmit(c echo.Context) error {
	username := c.FormValue("username")
	password := c.FormValue("password")
	confirmPassword := c.FormValue("confirm_password")
	email := c.FormValue("email")

	// 基本驗證
	if username == "" || password == "" || email == "" {
		return c.HTML(http.StatusBadRequest, a.generateRegisterPageHTML("所有字段都是必填的"))
	}

	if password != confirmPassword {
		return c.HTML(http.StatusBadRequest, a.generateRegisterPageHTML("密碼確認不匹配"))
	}

	// 在實際實現中，這裡應該調用用戶註冊服務
	// 目前簡化為直接返回成功消息
	a.logger.Info("用戶註冊請求", "username", username, "email", email)

	successHTML := a.generateRegisterSuccessHTML(username)
	return c.HTML(http.StatusOK, successHTML)
}

// handleLogout 處理登出請求
func (a *AuthUIPagePlugin) handleLogout(c echo.Context) error {
	// 清除會話 cookie
	c.SetCookie(&http.Cookie{
		Name:     "user_id",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		MaxAge:   -1, // 刪除 cookie
	})

	a.logger.Info("用戶登出")
	return c.Redirect(http.StatusFound, a.config.LoginRoute)
}

// generateLoginPageHTML 生成登錄頁面 HTML
func (a *AuthUIPagePlugin) generateLoginPageHTML(errorMsg ...string) string {
	errorSection := ""
	if len(errorMsg) > 0 && errorMsg[0] != "" {
		errorSection = fmt.Sprintf(`
		<div class="error-message">
			<p>%s</p>
		</div>`, errorMsg[0])
	}

	return fmt.Sprintf(`
<!DOCTYPE html>
<html lang="zh-TW">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>%s - 登錄</title>
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
        .auth-container {
            background: white;
            border-radius: 15px;
            padding: 40px;
            box-shadow: 0 20px 40px rgba(0,0,0,0.1);
            width: 100%%;
            max-width: 400px;
        }
        .logo {
            text-align: center;
            font-size: 2.5em;
            color: #667eea;
            margin-bottom: 10px;
        }
        .brand-name {
            text-align: center;
            font-size: 1.8em;
            color: #333;
            margin-bottom: 30px;
            font-weight: 300;
        }
        .form-group {
            margin-bottom: 20px;
        }
        label {
            display: block;
            margin-bottom: 8px;
            color: #555;
            font-weight: 500;
        }
        input[type="text"], input[type="password"], input[type="email"] {
            width: 100%%;
            padding: 12px;
            border: 2px solid #e1e1e1;
            border-radius: 8px;
            font-size: 16px;
            transition: border-color 0.3s;
            box-sizing: border-box;
        }
        input[type="text"]:focus, input[type="password"]:focus, input[type="email"]:focus {
            outline: none;
            border-color: #667eea;
        }
        .submit-btn {
            width: 100%%;
            padding: 12px;
            background: #667eea;
            color: white;
            border: none;
            border-radius: 8px;
            font-size: 16px;
            cursor: pointer;
            transition: background 0.3s;
        }
        .submit-btn:hover {
            background: #5a67d8;
        }
        .error-message {
            background: #fed7d7;
            color: #c53030;
            padding: 12px;
            border-radius: 8px;
            margin-bottom: 20px;
            border: 1px solid #feb2b2;
        }
        .auth-links {
            text-align: center;
            margin-top: 20px;
        }
        .auth-links a {
            color: #667eea;
            text-decoration: none;
        }
        .auth-links a:hover {
            text-decoration: underline;
        }
    </style>
</head>
<body>
    <div class="auth-container">
        <div class="logo">🔍</div>
        <div class="brand-name">%s</div>
        
        %s
        
        <form method="POST" action="%s">
            <div class="form-group">
                <label for="username">用戶名</label>
                <input type="text" id="username" name="username" required>
            </div>
            
            <div class="form-group">
                <label for="password">密碼</label>
                <input type="password" id="password" name="password" required>
            </div>
            
            <button type="submit" class="submit-btn">登錄</button>
        </form>
        
        <div class="auth-links">
            <p>還沒有帳號？ <a href="%s">立即註冊</a></p>
        </div>
    </div>
</body>
</html>`, a.config.Title, a.config.BrandName, errorSection, a.config.LoginRoute, a.config.RegisterRoute)
}

// generateRegisterPageHTML 生成註冊頁面 HTML
func (a *AuthUIPagePlugin) generateRegisterPageHTML(errorMsg ...string) string {
	errorSection := ""
	if len(errorMsg) > 0 && errorMsg[0] != "" {
		errorSection = fmt.Sprintf(`
		<div class="error-message">
			<p>%s</p>
		</div>`, errorMsg[0])
	}

	return fmt.Sprintf(`
<!DOCTYPE html>
<html lang="zh-TW">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>%s - 註冊</title>
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
        .auth-container {
            background: white;
            border-radius: 15px;
            padding: 40px;
            box-shadow: 0 20px 40px rgba(0,0,0,0.1);
            width: 100%%;
            max-width: 400px;
        }
        .logo {
            text-align: center;
            font-size: 2.5em;
            color: #667eea;
            margin-bottom: 10px;
        }
        .brand-name {
            text-align: center;
            font-size: 1.8em;
            color: #333;
            margin-bottom: 30px;
            font-weight: 300;
        }
        .form-group {
            margin-bottom: 20px;
        }
        label {
            display: block;
            margin-bottom: 8px;
            color: #555;
            font-weight: 500;
        }
        input[type="text"], input[type="password"], input[type="email"] {
            width: 100%%;
            padding: 12px;
            border: 2px solid #e1e1e1;
            border-radius: 8px;
            font-size: 16px;
            transition: border-color 0.3s;
            box-sizing: border-box;
        }
        input[type="text"]:focus, input[type="password"]:focus, input[type="email"]:focus {
            outline: none;
            border-color: #667eea;
        }
        .submit-btn {
            width: 100%%;
            padding: 12px;
            background: #667eea;
            color: white;
            border: none;
            border-radius: 8px;
            font-size: 16px;
            cursor: pointer;
            transition: background 0.3s;
        }
        .submit-btn:hover {
            background: #5a67d8;
        }
        .error-message {
            background: #fed7d7;
            color: #c53030;
            padding: 12px;
            border-radius: 8px;
            margin-bottom: 20px;
            border: 1px solid #feb2b2;
        }
        .auth-links {
            text-align: center;
            margin-top: 20px;
        }
        .auth-links a {
            color: #667eea;
            text-decoration: none;
        }
        .auth-links a:hover {
            text-decoration: underline;
        }
    </style>
</head>
<body>
    <div class="auth-container">
        <div class="logo">🔍</div>
        <div class="brand-name">%s</div>
        
        %s
        
        <form method="POST" action="%s">
            <div class="form-group">
                <label for="username">用戶名</label>
                <input type="text" id="username" name="username" required>
            </div>
            
            <div class="form-group">
                <label for="email">電子郵件</label>
                <input type="email" id="email" name="email" required>
            </div>
            
            <div class="form-group">
                <label for="password">密碼</label>
                <input type="password" id="password" name="password" required>
            </div>
            
            <div class="form-group">
                <label for="confirm_password">確認密碼</label>
                <input type="password" id="confirm_password" name="confirm_password" required>
            </div>
            
            <button type="submit" class="submit-btn">註冊</button>
        </form>
        
        <div class="auth-links">
            <p>已有帳號？ <a href="%s">立即登錄</a></p>
        </div>
    </div>
</body>
</html>`, a.config.Title, a.config.BrandName, errorSection, a.config.RegisterRoute, a.config.LoginRoute)
}

// generateRegisterSuccessHTML 生成註冊成功頁面 HTML
func (a *AuthUIPagePlugin) generateRegisterSuccessHTML(username string) string {
	return fmt.Sprintf(`
<!DOCTYPE html>
<html lang="zh-TW">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>%s - 註冊成功</title>
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
        .success-container {
            background: white;
            border-radius: 15px;
            padding: 40px;
            box-shadow: 0 20px 40px rgba(0,0,0,0.1);
            width: 100%%;
            max-width: 400px;
            text-align: center;
        }
        .success-icon {
            font-size: 4em;
            color: #48bb78;
            margin-bottom: 20px;
        }
        .success-title {
            font-size: 1.8em;
            color: #333;
            margin-bottom: 20px;
        }
        .success-message {
            color: #666;
            margin-bottom: 30px;
            line-height: 1.6;
        }
        .login-btn {
            display: inline-block;
            padding: 12px 24px;
            background: #667eea;
            color: white;
            text-decoration: none;
            border-radius: 8px;
            transition: background 0.3s;
        }
        .login-btn:hover {
            background: #5a67d8;
        }
    </style>
</head>
<body>
    <div class="success-container">
        <div class="success-icon">✅</div>
        <div class="success-title">註冊成功！</div>
        <div class="success-message">
            歡迎 %s！<br>
            您的帳號已成功創建。請使用您的用戶名和密碼登錄。
        </div>
        <a href="%s" class="login-btn">前往登錄</a>
    </div>
</body>
</html>`, a.config.Title, username, a.config.LoginRoute)
}
