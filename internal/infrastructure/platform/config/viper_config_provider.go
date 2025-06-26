package config

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"detectviz-platform/pkg/platform/contracts"

	"github.com/spf13/viper"
	"github.com/xeipuuv/gojsonschema"
)

// ViperConfigProvider 實現了 pkg/platform/contracts.ConfigProvider 介面。
// 職責: 負責從 YAML 文件等載入配置，並提供 Get 方法存取。
// 測試說明: 單元測試將驗證配置文件的正確解析和數據提取。
type ViperConfigProvider struct {
	viper *viper.Viper
}

// NewViperConfigProvider 構造函數，根據配置路徑讀取配置。
func NewViperConfigProvider(configFilePath string, logger contracts.Logger) (contracts.ConfigProvider, error) {
	v := viper.New()
	v.SetConfigType("yaml")
	v.SetConfigFile(configFilePath)
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			return nil, fmt.Errorf("ConfigProvider: 配置檔未找到 %s: %w", configFilePath, err)
		}
		return nil, fmt.Errorf("ConfigProvider: 讀取配置檔失敗 %s: %w", configFilePath, err)
	}
	logger.Info("[INFRA][ViperConfigProvider] 從 %s 初始化完成。\n", configFilePath)
	return &ViperConfigProvider{viper: v}, nil
}

func (c *ViperConfigProvider) GetName() string                    { return "viper" }
func (c *ViperConfigProvider) GetString(key string) string        { return c.viper.GetString(key) }
func (c *ViperConfigProvider) GetInt(key string) int              { return c.viper.GetInt(key) }
func (c *ViperConfigProvider) GetBool(key string) bool            { return c.viper.GetBool(key) }
func (c *ViperConfigProvider) Unmarshal(rawVal interface{}) error { return c.viper.Unmarshal(rawVal) }

// LoadCompositionConfig 載入 composition.yaml 配置
func (v *ViperConfigProvider) LoadCompositionConfig(ctx context.Context) (map[string]interface{}, error) {
	v.viper.SetConfigName("composition")
	if err := v.viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("無法讀取 composition.yaml: %w", err)
	}

	config := v.viper.AllSettings()

	// 執行 JSON Schema 驗證
	if err := v.validateConfig(config, "composition"); err != nil {
		return nil, fmt.Errorf("composition.yaml 配置驗證失敗: %w", err)
	}

	return config, nil
}

// LoadAppConfig 載入 app_config.yaml 配置
func (v *ViperConfigProvider) LoadAppConfig(ctx context.Context) (map[string]interface{}, error) {
	// 重新初始化以載入不同的配置文件
	v.viper = viper.New()
	v.viper.SetConfigType("yaml")
	v.viper.AddConfigPath("./configs")
	v.viper.AddConfigPath("./")
	v.viper.SetConfigName("app_config")

	if err := v.viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("無法讀取 app_config.yaml: %w", err)
	}

	config := v.viper.AllSettings()

	// 執行 JSON Schema 驗證
	if err := v.validateConfig(config, "app_config"); err != nil {
		return nil, fmt.Errorf("app_config.yaml 配置驗證失敗: %w", err)
	}

	return config, nil
}

// validateConfig 根據 JSON Schema 驗證配置
func (v *ViperConfigProvider) validateConfig(config map[string]interface{}, configType string) error {
	// 構建 Schema 文件路徑
	schemaPath := filepath.Join("schemas", fmt.Sprintf("%s.json", configType))

	// 檢查 Schema 文件是否存在
	if _, err := os.Stat(schemaPath); os.IsNotExist(err) {
		log.Printf("警告: 未找到 %s 的 JSON Schema 文件，跳過驗證", schemaPath)
		return nil
	}

	// 讀取 Schema 文件
	schemaBytes, err := os.ReadFile(schemaPath)
	if err != nil {
		return fmt.Errorf("無法讀取 Schema 文件 %s: %w", schemaPath, err)
	}

	// 載入 Schema
	schemaLoader := gojsonschema.NewBytesLoader(schemaBytes)

	// 將配置轉換為 JSON 字節
	configBytes, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("無法序列化配置為 JSON: %w", err)
	}

	// 載入配置數據
	documentLoader := gojsonschema.NewBytesLoader(configBytes)

	// 執行驗證
	result, err := gojsonschema.Validate(schemaLoader, documentLoader)
	if err != nil {
		return fmt.Errorf("Schema 驗證過程出錯: %w", err)
	}

	// 檢查驗證結果
	if !result.Valid() {
		var errors []string
		for _, desc := range result.Errors() {
			errors = append(errors, desc.String())
		}
		return fmt.Errorf("配置驗證失敗，錯誤:\n- %s", fmt.Sprintf("\n- %s", errors))
	}

	log.Printf("✅ %s 配置通過 JSON Schema 驗證", configType)
	return nil
}

// ValidatePluginConfig 驗證單個插件配置
func (v *ViperConfigProvider) ValidatePluginConfig(pluginType string, config map[string]interface{}) error {
	// 構建插件 Schema 文件路徑
	schemaPath := filepath.Join("schemas", "plugins", fmt.Sprintf("%s.json", pluginType))

	// 檢查 Schema 文件是否存在
	if _, err := os.Stat(schemaPath); os.IsNotExist(err) {
		log.Printf("警告: 未找到插件 %s 的 JSON Schema 文件，跳過驗證", pluginType)
		return nil
	}

	// 讀取 Schema 文件
	schemaBytes, err := os.ReadFile(schemaPath)
	if err != nil {
		return fmt.Errorf("無法讀取插件 Schema 文件 %s: %w", schemaPath, err)
	}

	// 載入 Schema
	schemaLoader := gojsonschema.NewBytesLoader(schemaBytes)

	// 將配置轉換為 JSON 字節
	configBytes, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("無法序列化插件配置為 JSON: %w", err)
	}

	// 載入配置數據
	documentLoader := gojsonschema.NewBytesLoader(configBytes)

	// 執行驗證
	result, err := gojsonschema.Validate(schemaLoader, documentLoader)
	if err != nil {
		return fmt.Errorf("插件 Schema 驗證過程出錯: %w", err)
	}

	// 檢查驗證結果
	if !result.Valid() {
		var errors []string
		for _, desc := range result.Errors() {
			errors = append(errors, desc.String())
		}
		return fmt.Errorf("插件 %s 配置驗證失敗，錯誤:\n- %s", pluginType, fmt.Sprintf("\n- %s", errors))
	}

	log.Printf("✅ 插件 %s 配置通過 JSON Schema 驗證", pluginType)
	return nil
}

// 確保實現了 ConfigProvider 介面
var _ contracts.ConfigProvider = (*ViperConfigProvider)(nil)
