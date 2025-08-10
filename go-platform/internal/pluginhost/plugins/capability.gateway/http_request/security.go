package http_request

import (
	"errors"
	"fmt"
	"net"
	"net/url"
	"regexp"
	"strings"
)

// 安全邊界配置 - 防範惡意或誤用的外部調用
type SecurityConfig struct {
	// URL 白名單和黑名單
	AllowedDomains    []string         `json:"allowed_domains"` // 允許的域名列表
	BlockedDomains    []string         `json:"blocked_domains"` // 禁止的域名列表
	AllowedURLPattern []*regexp.Regexp `json:"-"`               // 允許的 URL 正則模式
	BlockedURLPattern []*regexp.Regexp `json:"-"`               // 禁止的 URL 正則模式

	// 資源限制
	MaxRequestSize  int64 `json:"max_request_size"`  // 最大請求體大小（位元組）
	MaxResponseSize int64 `json:"max_response_size"` // 最大回應體大小（位元組）
	MaxRedirects    int   `json:"max_redirects"`     // 最大重定向次數

	// 超時控制
	MinTimeoutMs int `json:"min_timeout_ms"` // 最小超時時間（毫秒）
	MaxTimeoutMs int `json:"max_timeout_ms"` // 最大超時時間（毫秒）

	// 網路限制
	BlockPrivateIPs bool     `json:"block_private_ips"` // 禁止私有 IP 地址
	BlockLocalhost  bool     `json:"block_localhost"`   // 禁止 localhost
	AllowedSchemes  []string `json:"allowed_schemes"`   // 允許的協定方案

	// Headers 控制
	BlockedHeaders  []string          `json:"blocked_headers"`  // 禁止的標頭名稱
	RequiredHeaders map[string]string `json:"required_headers"` // 必要的標頭
}

// DefaultSecurityConfig 返回預設的安全配置
func DefaultSecurityConfig() *SecurityConfig {
	return &SecurityConfig{
		AllowedDomains: []string{}, // 空白表示允許所有（由其他規則限制）
		BlockedDomains: []string{
			"metadata.google.internal", // GCP metadata 服務
			"169.254.169.254",          // AWS/Azure metadata 服務
			"localhost",                // 本機服務
			"127.0.0.1",                // 本機 IP
		},

		MaxRequestSize:  1024 * 1024,     // 1MB
		MaxResponseSize: 5 * 1024 * 1024, // 5MB
		MaxRedirects:    5,

		MinTimeoutMs: 1000,  // 1 秒
		MaxTimeoutMs: 30000, // 30 秒

		BlockPrivateIPs: true,
		BlockLocalhost:  true,
		AllowedSchemes:  []string{"http", "https"},

		BlockedHeaders: []string{
			"authorization", // 防止洩漏認證資訊
			"x-api-key",     // 防止洩漏 API 金鑰
			"cookie",        // 防止 cookie 洩漏
		},
		RequiredHeaders: map[string]string{
			"user-agent": "detectviz-platform/1.0", // 識別請求來源
		},
	}
}

// ValidateRequest 驗證 HTTP 請求是否符合安全要求
func (sc *SecurityConfig) ValidateRequest(method, urlStr string, headers map[string]string, body []byte, timeoutMs int) error {
	// 1. 驗證 URL
	if err := sc.validateURL(urlStr); err != nil {
		return fmt.Errorf("URL 驗證失敗: %w", err)
	}

	// 2. 驗證 HTTP 方法
	if err := sc.validateMethod(method); err != nil {
		return fmt.Errorf("HTTP 方法驗證失敗: %w", err)
	}

	// 3. 驗證 Headers
	if err := sc.validateHeaders(headers); err != nil {
		return fmt.Errorf("Headers 驗證失敗: %w", err)
	}

	// 4. 驗證請求體大小
	if err := sc.validateRequestSize(body); err != nil {
		return fmt.Errorf("請求體大小驗證失敗: %w", err)
	}

	// 5. 驗證超時時間
	if err := sc.validateTimeout(timeoutMs); err != nil {
		return fmt.Errorf("超時時間驗證失敗: %w", err)
	}

	return nil
}

// validateURL 驗證 URL 的安全性
func (sc *SecurityConfig) validateURL(urlStr string) error {
	u, err := url.Parse(urlStr)
	if err != nil {
		return fmt.Errorf("無效的 URL: %w", err)
	}

	// 檢查協定方案
	if !sc.isSchemeAllowed(u.Scheme) {
		return fmt.Errorf("不允許的協定方案: %s", u.Scheme)
	}

	// 檢查域名黑名單
	if sc.isDomainBlocked(u.Hostname()) {
		return fmt.Errorf("域名被禁止: %s", u.Hostname())
	}

	// 檢查域名白名單
	if len(sc.AllowedDomains) > 0 && !sc.isDomainAllowed(u.Hostname()) {
		return fmt.Errorf("域名不在白名單中: %s", u.Hostname())
	}

	// 檢查私有 IP 和 localhost
	if sc.BlockPrivateIPs || sc.BlockLocalhost {
		if err := sc.validateIPAddress(u.Hostname()); err != nil {
			return err
		}
	}

	return nil
}

// validateMethod 驗證 HTTP 方法
func (sc *SecurityConfig) validateMethod(method string) error {
	allowedMethods := []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"}
	method = strings.ToUpper(method)

	for _, allowed := range allowedMethods {
		if method == allowed {
			return nil
		}
	}

	return fmt.Errorf("不允許的 HTTP 方法: %s", method)
}

// validateHeaders 驗證請求標頭
func (sc *SecurityConfig) validateHeaders(headers map[string]string) error {
	// 檢查禁止的標頭
	for _, blocked := range sc.BlockedHeaders {
		for key := range headers {
			if strings.EqualFold(key, blocked) {
				return fmt.Errorf("禁止使用的標頭: %s", key)
			}
		}
	}

	// 檢查必要的標頭
	for requiredKey, requiredValue := range sc.RequiredHeaders {
		found := false
		for key, value := range headers {
			if strings.EqualFold(key, requiredKey) {
				if requiredValue != "" && value != requiredValue {
					return fmt.Errorf("標頭 %s 的值不正確，期望: %s，實際: %s", key, requiredValue, value)
				}
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("缺少必要的標頭: %s", requiredKey)
		}
	}

	return nil
}

// validateRequestSize 驗證請求體大小
func (sc *SecurityConfig) validateRequestSize(body []byte) error {
	if int64(len(body)) > sc.MaxRequestSize {
		return fmt.Errorf("請求體大小 %d 超過限制 %d", len(body), sc.MaxRequestSize)
	}
	return nil
}

// validateTimeout 驗證超時時間
func (sc *SecurityConfig) validateTimeout(timeoutMs int) error {
	if timeoutMs < sc.MinTimeoutMs {
		return fmt.Errorf("超時時間 %d 小於最小值 %d", timeoutMs, sc.MinTimeoutMs)
	}
	if timeoutMs > sc.MaxTimeoutMs {
		return fmt.Errorf("超時時間 %d 超過最大值 %d", timeoutMs, sc.MaxTimeoutMs)
	}
	return nil
}

// isSchemeAllowed 檢查協定方案是否被允許
func (sc *SecurityConfig) isSchemeAllowed(scheme string) bool {
	scheme = strings.ToLower(scheme)
	for _, allowed := range sc.AllowedSchemes {
		if scheme == strings.ToLower(allowed) {
			return true
		}
	}
	return false
}

// isDomainBlocked 檢查域名是否被禁止
func (sc *SecurityConfig) isDomainBlocked(hostname string) bool {
	hostname = strings.ToLower(hostname)
	for _, blocked := range sc.BlockedDomains {
		if hostname == strings.ToLower(blocked) {
			return true
		}
		// 支援通配符匹配
		if strings.HasPrefix(blocked, "*.") {
			suffix := strings.TrimPrefix(blocked, "*.")
			if strings.HasSuffix(hostname, "."+suffix) || hostname == suffix {
				return true
			}
		}
	}
	return false
}

// isDomainAllowed 檢查域名是否在白名單中
func (sc *SecurityConfig) isDomainAllowed(hostname string) bool {
	if len(sc.AllowedDomains) == 0 {
		return true // 沒有白名單限制
	}

	hostname = strings.ToLower(hostname)
	for _, allowed := range sc.AllowedDomains {
		if hostname == strings.ToLower(allowed) {
			return true
		}
		// 支援通配符匹配
		if strings.HasPrefix(allowed, "*.") {
			suffix := strings.TrimPrefix(allowed, "*.")
			if strings.HasSuffix(hostname, "."+suffix) || hostname == suffix {
				return true
			}
		}
	}
	return false
}

// validateIPAddress 驗證 IP 位址限制
func (sc *SecurityConfig) validateIPAddress(hostname string) error {
	ip := net.ParseIP(hostname)
	if ip == nil {
		// 不是 IP 位址，跳過檢查
		return nil
	}

	// 檢查 localhost
	if sc.BlockLocalhost {
		if ip.IsLoopback() {
			return errors.New("禁止訪問 localhost")
		}
	}

	// 檢查私有 IP
	if sc.BlockPrivateIPs {
		if isPrivateIP(ip) {
			return errors.New("禁止訪問私有 IP 位址")
		}
	}

	return nil
}

// isPrivateIP 判斷是否為私有 IP 位址
func isPrivateIP(ip net.IP) bool {
	// RFC 1918 私有位址範圍
	privateBlocks := []*net.IPNet{
		{IP: net.IPv4(10, 0, 0, 0), Mask: net.CIDRMask(8, 32)},     // 10.0.0.0/8
		{IP: net.IPv4(172, 16, 0, 0), Mask: net.CIDRMask(12, 32)},  // 172.16.0.0/12
		{IP: net.IPv4(192, 168, 0, 0), Mask: net.CIDRMask(16, 32)}, // 192.168.0.0/16
		{IP: net.IPv4(169, 254, 0, 0), Mask: net.CIDRMask(16, 32)}, // 169.254.0.0/16 (link-local)
	}

	for _, block := range privateBlocks {
		if block.Contains(ip) {
			return true
		}
	}

	return false
}

// SanitizeHeaders 清理和標準化請求標頭
func (sc *SecurityConfig) SanitizeHeaders(headers map[string]string) map[string]string {
	sanitized := make(map[string]string)

	// 複製允許的標頭
	for key, value := range headers {
		blocked := false
		for _, blockedHeader := range sc.BlockedHeaders {
			if strings.EqualFold(key, blockedHeader) {
				blocked = true
				break
			}
		}
		if !blocked {
			sanitized[key] = value
		}
	}

	// 添加必要的標頭
	for key, value := range sc.RequiredHeaders {
		sanitized[key] = value
	}

	return sanitized
}

// ValidateResponseSize 驗證回應體大小
func (sc *SecurityConfig) ValidateResponseSize(size int64) error {
	if size > sc.MaxResponseSize {
		return fmt.Errorf("回應體大小 %d 超過限制 %d", size, sc.MaxResponseSize)
	}
	return nil
}
