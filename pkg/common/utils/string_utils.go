package utils

import (
	"regexp"
	"strings"
	"unicode"
)

// StringUtils 提供字串處理相關的工具函數
// 職責: 統一管理平台內字串處理邏輯
// AI_SCAFFOLD_HINT: 字串工具集，AI 可根據需求自動選擇合適的處理函數
type StringUtils struct{}

// NewStringUtils 創建新的字串工具
func NewStringUtils() *StringUtils {
	return &StringUtils{}
}

// ToSnakeCase 將字串轉換為 snake_case
// AI_SCAFFOLD_HINT: 用於文件名、變數名轉換
func (s *StringUtils) ToSnakeCase(str string) string {
	// 在大寫字母前插入下劃線
	re := regexp.MustCompile("([a-z0-9])([A-Z])")
	snake := re.ReplaceAllString(str, "${1}_${2}")
	return strings.ToLower(snake)
}

// ToCamelCase 將字串轉換為 camelCase
// AI_SCAFFOLD_HINT: 用於 JSON 字段名轉換
func (s *StringUtils) ToCamelCase(str string) string {
	words := strings.Split(str, "_")
	if len(words) == 0 {
		return str
	}

	result := strings.ToLower(words[0])
	for i := 1; i < len(words); i++ {
		if len(words[i]) > 0 {
			result += strings.ToUpper(string(words[i][0])) + strings.ToLower(words[i][1:])
		}
	}
	return result
}

// ToPascalCase 將字串轉換為 PascalCase
// AI_SCAFFOLD_HINT: 用於類型名、介面名轉換
func (s *StringUtils) ToPascalCase(str string) string {
	words := strings.Split(str, "_")
	result := ""
	for _, word := range words {
		if len(word) > 0 {
			result += strings.ToUpper(string(word[0])) + strings.ToLower(word[1:])
		}
	}
	return result
}

// ToKebabCase 將字串轉換為 kebab-case
// AI_SCAFFOLD_HINT: 用於文檔名、URL 路徑轉換
func (s *StringUtils) ToKebabCase(str string) string {
	re := regexp.MustCompile("([a-z0-9])([A-Z])")
	kebab := re.ReplaceAllString(str, "${1}-${2}")
	return strings.ToLower(kebab)
}

// Sanitize 清理字串，移除特殊字符
// AI_SCAFFOLD_HINT: 用於用戶輸入清理
func (s *StringUtils) Sanitize(str string) string {
	// 移除控制字符和非打印字符
	return strings.Map(func(r rune) rune {
		if unicode.IsControl(r) || !unicode.IsPrint(r) {
			return -1
		}
		return r
	}, str)
}

// Truncate 截斷字串到指定長度
// AI_SCAFFOLD_HINT: 用於日誌輸出、顯示文本截斷
func (s *StringUtils) Truncate(str string, length int) string {
	if len(str) <= length {
		return str
	}
	return str[:length] + "..."
}

// IsEmpty 檢查字串是否為空（包括只有空白字符）
// AI_SCAFFOLD_HINT: 用於輸入驗證
func (s *StringUtils) IsEmpty(str string) bool {
	return strings.TrimSpace(str) == ""
}

// ContainsAny 檢查字串是否包含任意一個子字串
// AI_SCAFFOLD_HINT: 用於關鍵字匹配
func (s *StringUtils) ContainsAny(str string, substrings []string) bool {
	for _, substr := range substrings {
		if strings.Contains(str, substr) {
			return true
		}
	}
	return false
}

// MaskSensitive 遮蔽敏感信息
// AI_SCAFFOLD_HINT: 用於日誌安全，遮蔽密碼、令牌等敏感數據
func (s *StringUtils) MaskSensitive(str string, visibleChars int) string {
	if len(str) <= visibleChars {
		return strings.Repeat("*", len(str))
	}

	visible := str[:visibleChars]
	masked := strings.Repeat("*", len(str)-visibleChars)
	return visible + masked
}
