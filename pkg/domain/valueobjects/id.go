package valueobjects

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/google/uuid"
)

// IDVO 封裝 ID 的值對象
// 職責: 確保 ID 的格式正確性和不可變性，支持 UUID 格式
type IDVO struct {
	value string
}

// UUID 格式的正則表達式
var uuidRegex = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[1-5][0-9a-fA-F]{3}-[89abAB][0-9a-fA-F]{3}-[0-9a-fA-F]{12}$`)

// NewIDVO 創建新的 ID 值對象
func NewIDVO(id string) (IDVO, error) {
	// 清理輸入
	id = strings.TrimSpace(id)

	// 驗證 ID 格式
	if err := validateID(id); err != nil {
		return IDVO{}, err
	}

	return IDVO{value: id}, nil
}

// NewIDVOFromUUID 從 UUID 創建 ID 值對象
func NewIDVOFromUUID(id uuid.UUID) IDVO {
	return IDVO{value: id.String()}
}

// GenerateNewIDVO 生成新的 UUID ID 值對象
func GenerateNewIDVO() IDVO {
	return IDVO{value: uuid.New().String()}
}

// validateID 驗證 ID 格式
func validateID(id string) error {
	if id == "" {
		return fmt.Errorf("ID 不能為空")
	}

	// 檢查是否為有效的 UUID 格式
	if !uuidRegex.MatchString(id) {
		return fmt.Errorf("ID 格式無效，必須為 UUID 格式: %s", id)
	}

	// 使用 Google UUID 庫進行額外驗證
	if _, err := uuid.Parse(id); err != nil {
		return fmt.Errorf("ID 格式無效: %s, 錯誤: %v", id, err)
	}

	return nil
}

// String 返回 ID 的字符串表示
func (i IDVO) String() string {
	return i.value
}

// Value 返回 ID 的值
func (i IDVO) Value() string {
	return i.value
}

// UUID 返回 ID 的 UUID 對象
func (i IDVO) UUID() (uuid.UUID, error) {
	return uuid.Parse(i.value)
}

// Equals 比較兩個 ID 值對象是否相等
func (i IDVO) Equals(other IDVO) bool {
	return strings.EqualFold(i.value, other.value)
}

// IsEmpty 檢查 ID 值對象是否為空
func (i IDVO) IsEmpty() bool {
	return i.value == ""
}

// IsNil 檢查 ID 是否為 nil UUID
func (i IDVO) IsNil() bool {
	if i.IsEmpty() {
		return true
	}

	nilUUID := uuid.Nil.String()
	return strings.EqualFold(i.value, nilUUID)
}

// Version 返回 UUID 的版本
func (i IDVO) Version() (int, error) {
	if i.IsEmpty() {
		return 0, fmt.Errorf("ID 為空")
	}

	parsedUUID, err := uuid.Parse(i.value)
	if err != nil {
		return 0, err
	}

	return int(parsedUUID.Version()), nil
}

// Variant 返回 UUID 的變體
func (i IDVO) Variant() (uuid.Variant, error) {
	if i.IsEmpty() {
		return uuid.Invalid, fmt.Errorf("ID 為空")
	}

	parsedUUID, err := uuid.Parse(i.value)
	if err != nil {
		return uuid.Invalid, err
	}

	return parsedUUID.Variant(), nil
}

// ShortString 返回 ID 的短字符串表示 (前8個字符)
func (i IDVO) ShortString() string {
	if len(i.value) < 8 {
		return i.value
	}
	return i.value[:8]
}

// IsValidV4 檢查是否為有效的 UUID v4
func (i IDVO) IsValidV4() bool {
	version, err := i.Version()
	if err != nil {
		return false
	}

	variant, err := i.Variant()
	if err != nil {
		return false
	}

	return version == 4 && variant == uuid.RFC4122
}

// ToBytes 返回 ID 的字節表示
func (i IDVO) ToBytes() ([]byte, error) {
	if i.IsEmpty() {
		return nil, fmt.Errorf("ID 為空")
	}

	parsedUUID, err := uuid.Parse(i.value)
	if err != nil {
		return nil, err
	}

	return parsedUUID[:], nil
}

// MarshalJSON 實現 JSON 序列化
func (i IDVO) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, i.value)), nil
}

// UnmarshalJSON 實現 JSON 反序列化
func (i *IDVO) UnmarshalJSON(data []byte) error {
	// 去除引號
	id := strings.Trim(string(data), `"`)

	idVO, err := NewIDVO(id)
	if err != nil {
		return err
	}

	*i = idVO
	return nil
}

// MarshalText 實現文本序列化
func (i IDVO) MarshalText() ([]byte, error) {
	return []byte(i.value), nil
}

// UnmarshalText 實現文本反序列化
func (i *IDVO) UnmarshalText(data []byte) error {
	idVO, err := NewIDVO(string(data))
	if err != nil {
		return err
	}

	*i = idVO
	return nil
}
