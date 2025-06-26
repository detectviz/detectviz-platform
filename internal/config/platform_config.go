package config

// PlatformConfig 模擬了 `composition.yaml` 檔案的內容，用於配置驅動平台的組裝。
// 檔案位置: configs/composition.yaml
type PlatformConfig struct {
	Database struct {
		Type string `mapstructure:"type"`
		DSN  string `mapstructure:"dsn"`
	} `mapstructure:"database"`
	Logger struct {
		Type  string `mapstructure:"type"`
		Level string `mapstructure:"level"`
	} `mapstructure:"logger"`
	Auth struct {
		Provider string `mapstructure:"provider"`
		Keycloak struct {
			URL string `mapstructure:"url"`
		} `mapstructure:"keycloak"`
	} `mapstructure:"auth"`
	Plugins struct {
		Importers []string `mapstructure:"importers"`
		// Authenticator 的插件化通常由 AuthProvider 管理，這裡保留 Authenticator 列表僅供參考
		Authenticators  []string `mapstructure:"authenticators"`
		UIComponents    []string `mapstructure:"uicomponents"`
		LLMProviders    []string `mapstructure:"llm_providers"`
		EmbeddingStores []string `mapstructure:"embedding_stores"`
	} `mapstructure:"plugins"`
	Server struct {
		Port string `mapstructure:"port"`
	} `mapstructure:"server"`
	Routes map[string]string `mapstructure:"routes"` // 示例路由配置
}
