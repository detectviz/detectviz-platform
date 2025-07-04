package valueobjects

import (
	"fmt"
	"regexp"
	"strings"
)

// EmailVO 封裝電子郵件地址的值對象
// 職責: 確保電子郵件地址的格式正確性和不可變性
type EmailVO struct {
	value string
}

// 電子郵件驗證的正則表達式
var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

// NewEmailVO 創建新的電子郵件值對象
func NewEmailVO(email string) (EmailVO, error) {
	// 清理輸入
	email = strings.TrimSpace(email)
	email = strings.ToLower(email)

	// 驗證電子郵件格式
	if err := validateEmail(email); err != nil {
		return EmailVO{}, err
	}

	return EmailVO{value: email}, nil
}

// validateEmail 驗證電子郵件格式
func validateEmail(email string) error {
	if email == "" {
		return fmt.Errorf("電子郵件地址不能為空")
	}

	if len(email) > 254 {
		return fmt.Errorf("電子郵件地址長度不能超過 254 個字符")
	}

	if !emailRegex.MatchString(email) {
		return fmt.Errorf("電子郵件地址格式無效: %s", email)
	}

	// 檢查本地部分長度 (@ 符號前面的部分)
	parts := strings.Split(email, "@")
	if len(parts[0]) > 64 {
		return fmt.Errorf("電子郵件地址本地部分長度不能超過 64 個字符")
	}

	// 檢查域名部分
	if len(parts[1]) > 253 {
		return fmt.Errorf("電子郵件地址域名部分長度不能超過 253 個字符")
	}

	return nil
}

// String 返回電子郵件地址的字符串表示
func (e EmailVO) String() string {
	return e.value
}

// Value 返回電子郵件地址的值
func (e EmailVO) Value() string {
	return e.value
}

// Equals 比較兩個電子郵件值對象是否相等
func (e EmailVO) Equals(other EmailVO) bool {
	return e.value == other.value
}

// IsEmpty 檢查電子郵件值對象是否為空
func (e EmailVO) IsEmpty() bool {
	return e.value == ""
}

// Domain 返回電子郵件地址的域名部分
func (e EmailVO) Domain() string {
	if e.IsEmpty() {
		return ""
	}

	parts := strings.Split(e.value, "@")
	if len(parts) != 2 {
		return ""
	}

	return parts[1]
}

// LocalPart 返回電子郵件地址的本地部分 (@ 符號前面的部分)
func (e EmailVO) LocalPart() string {
	if e.IsEmpty() {
		return ""
	}

	parts := strings.Split(e.value, "@")
	if len(parts) != 2 {
		return ""
	}

	return parts[0]
}

// IsGmailAddress 檢查是否為 Gmail 地址
func (e EmailVO) IsGmailAddress() bool {
	return strings.HasSuffix(e.value, "@gmail.com")
}

// IsCompanyEmail 檢查是否為公司電子郵件 (非免費郵件服務提供商)
func (e EmailVO) IsCompanyEmail() bool {
	if e.IsEmpty() {
		return false
	}

	// 常見的免費郵件服務提供商
	freeProviders := []string{
		"@gmail.com", "@yahoo.com", "@hotmail.com", "@outlook.com",
		"@live.com", "@msn.com", "@aol.com", "@icloud.com",
		"@me.com", "@mac.com", "@protonmail.com", "@yandex.com",
	}

	domain := "@" + e.Domain()
	for _, provider := range freeProviders {
		if strings.EqualFold(domain, provider) {
			return false
		}
	}

	return true
}

// MarshalJSON 實現 JSON 序列化
func (e EmailVO) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, e.value)), nil
}

// UnmarshalJSON 實現 JSON 反序列化
func (e *EmailVO) UnmarshalJSON(data []byte) error {
	// 去除引號
	email := strings.Trim(string(data), `"`)

	emailVO, err := NewEmailVO(email)
	if err != nil {
		return err
	}

	*e = emailVO
	return nil
}
