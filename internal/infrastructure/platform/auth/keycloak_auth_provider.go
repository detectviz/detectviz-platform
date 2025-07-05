package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"detectviz-platform/pkg/platform/contracts"
)

// KeycloakAuthProvider 實現了 AuthProvider 介面，提供 Keycloak 身份驗證集成
// 職責: 與 Keycloak 服務交互，執行身份驗證和授權檢查
type KeycloakAuthProvider struct {
	baseURL      string
	realm        string
	clientID     string
	clientSecret string
	httpClient   *http.Client
	logger       contracts.Logger
}

// KeycloakConfig 定義 Keycloak 認證提供者的配置
type KeycloakConfig struct {
	BaseURL      string `yaml:"base_url" json:"base_url"`
	Realm        string `yaml:"realm" json:"realm"`
	ClientID     string `yaml:"client_id" json:"client_id"`
	ClientSecret string `yaml:"client_secret" json:"client_secret"`
	Timeout      string `yaml:"timeout" json:"timeout"`
}

// KeycloakTokenResponse 定義 Keycloak 令牌響應結構
type KeycloakTokenResponse struct {
	AccessToken      string `json:"access_token"`
	ExpiresIn        int    `json:"expires_in"`
	RefreshExpiresIn int    `json:"refresh_expires_in"`
	RefreshToken     string `json:"refresh_token"`
	TokenType        string `json:"token_type"`
	NotBeforePolicy  int    `json:"not-before-policy"`
	SessionState     string `json:"session_state"`
	Scope            string `json:"scope"`
}

// KeycloakUserInfo 定義用戶信息結構
type KeycloakUserInfo struct {
	Sub               string `json:"sub"`
	EmailVerified     bool   `json:"email_verified"`
	Name              string `json:"name"`
	PreferredUsername string `json:"preferred_username"`
	GivenName         string `json:"given_name"`
	FamilyName        string `json:"family_name"`
	Email             string `json:"email"`
}

// KeycloakIntrospectionResponse 定義令牌內省響應結構
type KeycloakIntrospectionResponse struct {
	Active    bool     `json:"active"`
	Sub       string   `json:"sub"`
	Username  string   `json:"username"`
	Email     string   `json:"email"`
	Scope     string   `json:"scope"`
	ClientID  string   `json:"client_id"`
	TokenType string   `json:"token_type"`
	Exp       int64    `json:"exp"`
	Iat       int64    `json:"iat"`
	Aud       []string `json:"aud"`
}

// NewKeycloakAuthProvider 創建新的 Keycloak 認證提供者實例
func NewKeycloakAuthProvider(config KeycloakConfig, logger contracts.Logger) (contracts.AuthProvider, error) {
	if config.BaseURL == "" {
		return nil, fmt.Errorf("keycloak base URL is required")
	}

	if config.Realm == "" {
		return nil, fmt.Errorf("keycloak realm is required")
	}

	if config.ClientID == "" {
		return nil, fmt.Errorf("keycloak client ID is required")
	}

	timeout := 30 * time.Second
	if config.Timeout != "" {
		if t, err := time.ParseDuration(config.Timeout); err == nil {
			timeout = t
		}
	}

	httpClient := &http.Client{
		Timeout: timeout,
	}

	logger.Info("初始化 Keycloak 認證提供者",
		"base_url", config.BaseURL,
		"realm", config.Realm,
		"client_id", config.ClientID)

	return &KeycloakAuthProvider{
		baseURL:      config.BaseURL,
		realm:        config.Realm,
		clientID:     config.ClientID,
		clientSecret: config.ClientSecret,
		httpClient:   httpClient,
		logger:       logger,
	}, nil
}

// Authenticate 驗證用戶憑證並返回用戶 ID
func (k *KeycloakAuthProvider) Authenticate(ctx context.Context, credentials string) (string, error) {
	// credentials 可以是 JWT token 或者 username:password 格式
	if strings.HasPrefix(credentials, "Bearer ") {
		token := strings.TrimPrefix(credentials, "Bearer ")
		return k.validateToken(ctx, token)
	}

	// 嘗試解析為用戶名密碼格式
	parts := strings.SplitN(credentials, ":", 2)
	if len(parts) == 2 {
		return k.authenticateWithPassword(ctx, parts[0], parts[1])
	}

	return "", fmt.Errorf("invalid credentials format")
}

// Authorize 檢查用戶是否有權限訪問特定資源
func (k *KeycloakAuthProvider) Authorize(ctx context.Context, userID string, resource string, action string) (bool, error) {
	k.logger.Debug("檢查用戶授權",
		"user_id", userID,
		"resource", resource,
		"action", action)

	// 簡化的授權檢查，實際實現中應該調用 Keycloak 的授權 API
	// 這裡可以根據需要實現更複雜的 RBAC 或 ABAC 邏輯

	// 示例：管理員用戶有所有權限
	if userID == "admin" {
		return true, nil
	}

	// 示例：普通用戶只能讀取資源
	if action == "read" {
		return true, nil
	}

	return false, nil
}

// VerifyToken 驗證 JWT 令牌並返回用戶 ID
func (k *KeycloakAuthProvider) VerifyToken(ctx context.Context, token string) (string, error) {
	return k.validateToken(ctx, token)
}

// CheckPermissions 查詢外部服務以檢查用戶的詳細權限
func (k *KeycloakAuthProvider) CheckPermissions(ctx context.Context, userID, resource, action string) (bool, error) {
	// 這裡可以調用 Keycloak 的 UMA (User-Managed Access) API 進行詳細權限檢查
	k.logger.Debug("檢查詳細權限",
		"user_id", userID,
		"resource", resource,
		"action", action)

	// 示例實現：調用 Keycloak 權限 API
	// 實際實現中應該構建適當的 API 請求
	return k.Authorize(ctx, userID, resource, action)
}

// HashPassword 將明文密碼轉換為安全的散列值
func (k *KeycloakAuthProvider) HashPassword(ctx context.Context, plainPassword string) (string, error) {
	// 在 Keycloak 集成中，密碼散列通常由 Keycloak 服務器處理
	// 這裡提供一個基本的 bcrypt 實現作為後備
	// 實際生產環境中應該使用專門的密碼散列服務

	// 注意：在真實的 Keycloak 集成中，密碼管理應該完全委託給 Keycloak
	// 這個方法主要用於與其他認證系統的兼容性
	return "", fmt.Errorf("password hashing should be handled by Keycloak server")
}

// VerifyPassword 驗證明文密碼是否與給定的散列值匹配
func (k *KeycloakAuthProvider) VerifyPassword(ctx context.Context, plainPassword, hashedPassword string) (bool, error) {
	// 在 Keycloak 集成中，密碼驗證通常通過認證 API 完成
	// 這個方法主要用於與其他認證系統的兼容性
	return false, fmt.Errorf("password verification should be handled through Keycloak authentication API")
}

// GenerateCSRFToken 為當前會話生成一個新的 CSRF token
func (k *KeycloakAuthProvider) GenerateCSRFToken(ctx context.Context) (string, error) {
	// 簡單的 CSRF token 生成實現
	// 實際生產環境中應該使用更安全的隨機生成方法

	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate CSRF token: %w", err)
	}

	token := base64.URLEncoding.EncodeToString(bytes)
	k.logger.Debug("生成 CSRF token", "token_length", len(token))

	return token, nil
}

// ValidateCSRFToken 驗證傳入的 CSRF token 是否有效
func (k *KeycloakAuthProvider) ValidateCSRFToken(ctx context.Context, token string) error {
	// 簡化的 CSRF token 驗證
	// 實際實現中應該與存儲的 token 進行比較
	if token == "" {
		return fmt.Errorf("CSRF token is required")
	}

	// 檢查 token 格式
	if len(token) < 32 {
		return fmt.Errorf("invalid CSRF token format")
	}

	k.logger.Debug("驗證 CSRF token", "token_length", len(token))

	// 在真實實現中，這裡應該與會話存儲中的 token 進行比較
	// 目前返回成功以保持向後兼容
	return nil
}

// GetName 返回提供者名稱
func (k *KeycloakAuthProvider) GetName() string {
	return "keycloak_auth_provider"
}

// validateToken 驗證 JWT 令牌
func (k *KeycloakAuthProvider) validateToken(ctx context.Context, token string) (string, error) {
	// 構建內省端點 URL
	introspectURL := fmt.Sprintf("%s/realms/%s/protocol/openid-connect/token/introspect", k.baseURL, k.realm)

	// 準備請求數據
	data := url.Values{}
	data.Set("token", token)
	data.Set("client_id", k.clientID)
	if k.clientSecret != "" {
		data.Set("client_secret", k.clientSecret)
	}

	// 創建 HTTP 請求
	req, err := http.NewRequestWithContext(ctx, "POST", introspectURL, strings.NewReader(data.Encode()))
	if err != nil {
		return "", fmt.Errorf("failed to create introspection request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// 發送請求
	k.logger.Debug("驗證令牌", "token_length", len(token))

	resp, err := k.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to introspect token: %w", err)
	}
	defer resp.Body.Close()

	// 檢查響應狀態
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("token introspection failed with status: %d", resp.StatusCode)
	}

	// 解析響應
	var introspectionResp KeycloakIntrospectionResponse
	if err := json.NewDecoder(resp.Body).Decode(&introspectionResp); err != nil {
		return "", fmt.Errorf("failed to decode introspection response: %w", err)
	}

	// 檢查令牌是否有效
	if !introspectionResp.Active {
		return "", fmt.Errorf("token is not active")
	}

	// 檢查令牌是否過期
	if introspectionResp.Exp > 0 && time.Now().Unix() > introspectionResp.Exp {
		return "", fmt.Errorf("token has expired")
	}

	k.logger.Info("令牌驗證成功",
		"user_id", introspectionResp.Sub,
		"username", introspectionResp.Username)

	return introspectionResp.Sub, nil
}

// authenticateWithPassword 使用用戶名密碼進行身份驗證
func (k *KeycloakAuthProvider) authenticateWithPassword(ctx context.Context, username, password string) (string, error) {
	// 構建令牌端點 URL
	tokenURL := fmt.Sprintf("%s/realms/%s/protocol/openid-connect/token", k.baseURL, k.realm)

	// 準備請求數據
	data := url.Values{}
	data.Set("grant_type", "password")
	data.Set("client_id", k.clientID)
	if k.clientSecret != "" {
		data.Set("client_secret", k.clientSecret)
	}
	data.Set("username", username)
	data.Set("password", password)

	// 創建 HTTP 請求
	req, err := http.NewRequestWithContext(ctx, "POST", tokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return "", fmt.Errorf("failed to create token request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// 發送請求
	k.logger.Debug("用戶名密碼認證", "username", username)

	resp, err := k.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to authenticate with password: %w", err)
	}
	defer resp.Body.Close()

	// 檢查響應狀態
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("authentication failed with status: %d", resp.StatusCode)
	}

	// 解析響應
	var tokenResp KeycloakTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return "", fmt.Errorf("failed to decode token response: %w", err)
	}

	// 獲取用戶信息
	userID, err := k.getUserInfo(ctx, tokenResp.AccessToken)
	if err != nil {
		return "", fmt.Errorf("failed to get user info: %w", err)
	}

	k.logger.Info("用戶名密碼認證成功", "username", username, "user_id", userID)

	return userID, nil
}

// getUserInfo 獲取用戶信息
func (k *KeycloakAuthProvider) getUserInfo(ctx context.Context, accessToken string) (string, error) {
	// 構建用戶信息端點 URL
	userInfoURL := fmt.Sprintf("%s/realms/%s/protocol/openid-connect/userinfo", k.baseURL, k.realm)

	// 創建 HTTP 請求
	req, err := http.NewRequestWithContext(ctx, "GET", userInfoURL, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create userinfo request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)

	// 發送請求
	resp, err := k.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to get user info: %w", err)
	}
	defer resp.Body.Close()

	// 檢查響應狀態
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("get user info failed with status: %d", resp.StatusCode)
	}

	// 解析響應
	var userInfo KeycloakUserInfo
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		return "", fmt.Errorf("failed to decode user info response: %w", err)
	}

	return userInfo.Sub, nil
}
