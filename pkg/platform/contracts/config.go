package contracts

// ConfigProvider 定義了平台統一的設定載入和存取介面。
// 職責: 從多種來源（如文件、環境變數）抽象化配置的讀取，並支持將配置反序列化到結構體。
// AI_PLUGIN_TYPE: "config_provider"
// AI_IMPL_PACKAGE: "detectviz-platform/internal/infrastructure/platform/config/viper_config_provider"
// AI_IMPL_CONSTRUCTOR: "NewViperConfigProvider"
// @See: internal/infrastructure/platform/config/viper_config_provider.go
type ConfigProvider interface {
	// GetString 根據鍵名獲取字符串類型的配置值。
	GetString(key string) string
	// GetInt 根據鍵名獲取整數類型的配置值。
	GetInt(key string) int
	// GetBool 根據鍵名獲取布爾類型的配置值。
	GetBool(key string) bool
	// Unmarshal 將整個配置或指定部分的配置反序列化到 Go 結構體中。
	Unmarshal(rawVal interface{}) error
	// GetName 返回配置提供者的名稱。
	GetName() string
}
