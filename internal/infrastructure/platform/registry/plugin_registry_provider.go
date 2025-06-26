package registry

import (
	"detectviz-platform/pkg/platform/contracts"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/xeipuuv/gojsonschema"
)

// PluginRegistryProvider 實現了 pkg/platform/contracts.PluginRegistryProvider 介面。
// 職責: 管理已載入和可用的插件，提供插件查詢和元數據獲取功能。
// 測試說明: 單元測試將驗證插件註冊、查詢和元數據管理的正確性。
type PluginRegistryProvider struct {
	plugins  map[string]interface{}    // 儲存插件實例
	metadata map[string]map[string]any // 儲存插件元數據
	mutex    sync.RWMutex              // 保護併發存取
	logger   contracts.Logger          // 添加 logger 字段
}

// NewPluginRegistryProvider 構造函數，創建新的插件註冊表實例。
func NewPluginRegistryProvider(logger contracts.Logger) contracts.PluginRegistryProvider {
	if logger == nil {
		// 創建簡單的控制台 logger 作為備用
		logger = &SimpleConsoleLogger{}
	}
	return &PluginRegistryProvider{
		plugins:  make(map[string]interface{}),
		metadata: make(map[string]map[string]any),
		logger:   logger,
	}
}

// SimpleConsoleLogger 簡單的控制台日誌實現，作為備用
type SimpleConsoleLogger struct{}

func (l *SimpleConsoleLogger) Debug(format string, args ...interface{}) {
	fmt.Printf("[DEBUG] "+format+"\n", args...)
}
func (l *SimpleConsoleLogger) Info(format string, args ...interface{}) {
	fmt.Printf("[INFO] "+format+"\n", args...)
}
func (l *SimpleConsoleLogger) Warn(format string, args ...interface{}) {
	fmt.Printf("[WARN] "+format+"\n", args...)
}
func (l *SimpleConsoleLogger) Error(format string, args ...interface{}) {
	fmt.Printf("[ERROR] "+format+"\n", args...)
}
func (l *SimpleConsoleLogger) Fatal(format string, args ...interface{}) {
	fmt.Printf("[FATAL] "+format+"\n", args...)
}
func (l *SimpleConsoleLogger) WithFields(fields ...interface{}) contracts.Logger { return l }
func (l *SimpleConsoleLogger) WithContext(ctx interface{}) contracts.Logger      { return l }
func (l *SimpleConsoleLogger) GetName() string                                   { return "simple_console" }

func (p *PluginRegistryProvider) GetName() string {
	return "core_registry"
}

// Register 註冊一個具名的插件實例
func (p *PluginRegistryProvider) Register(name string, provider any) error {
	if name == "" {
		return fmt.Errorf("plugin name cannot be empty")
	}

	p.mutex.Lock()
	defer p.mutex.Unlock()

	if _, exists := p.plugins[name]; exists {
		return fmt.Errorf("plugin '%s' is already registered", name)
	}

	p.plugins[name] = provider

	// 設置預設元數據
	p.metadata[name] = map[string]any{
		"name":   name,
		"type":   fmt.Sprintf("%T", provider),
		"status": "registered",
	}

	return nil
}

// Get 獲取指定名稱的插件實例
func (p *PluginRegistryProvider) Get(name string) (any, error) {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	plugin, exists := p.plugins[name]
	if !exists {
		return nil, fmt.Errorf("plugin '%s' not found", name)
	}

	return plugin, nil
}

// List 列出所有已註冊插件的名稱
func (p *PluginRegistryProvider) List() []string {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	names := make([]string, 0, len(p.plugins))
	for name := range p.plugins {
		names = append(names, name)
	}

	return names
}

// GetMetadata 回傳特定插件的描述資訊（版本、作者、狀態等）
func (p *PluginRegistryProvider) GetMetadata(name string) (map[string]any, error) {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	metadata, exists := p.metadata[name]
	if !exists {
		return nil, fmt.Errorf("metadata for plugin '%s' not found", name)
	}

	// 複製元數據以避免外部修改
	result := make(map[string]any)
	for k, v := range metadata {
		result[k] = v
	}

	return result, nil
}

// UpdateMetadata 更新插件的元數據
func (p *PluginRegistryProvider) UpdateMetadata(name string, metadata map[string]any) error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if _, exists := p.plugins[name]; !exists {
		return fmt.Errorf("plugin '%s' not found", name)
	}

	if p.metadata[name] == nil {
		p.metadata[name] = make(map[string]any)
	}

	// 更新元數據
	for k, v := range metadata {
		p.metadata[name][k] = v
	}

	return nil
}

// ValidatePluginsConfig 驗證所有插件配置的 JSON Schema
func (pr *PluginRegistryProvider) ValidatePluginsConfig(plugins []map[string]interface{}) error {
	for _, pluginConfig := range plugins {
		// 提取插件類型、名稱和配置
		pluginType, ok := pluginConfig["type"].(string)
		if !ok {
			return fmt.Errorf("插件配置缺少必需的 'type' 欄位")
		}

		pluginName, ok := pluginConfig["name"].(string)
		if !ok {
			return fmt.Errorf("插件配置缺少必需的 'name' 欄位")
		}

		config, ok := pluginConfig["config"].(map[string]interface{})
		if !ok {
			return fmt.Errorf("插件 %s 配置缺少必需的 'config' 欄位", pluginName)
		}

		// 驗證插件配置 (使用 JSON Schema)
		if err := pr.validatePluginConfig(pluginType, config); err != nil {
			return fmt.Errorf("插件 %s (類型: %s) 配置驗證失敗: %w", pluginName, pluginType, err)
		}
	}
	return nil
}

// validatePluginConfig 驗證插件配置符合其 JSON Schema
func (pr *PluginRegistryProvider) validatePluginConfig(pluginType string, config map[string]interface{}) error {
	// 構建 Schema 文件路徑
	schemaPath := filepath.Join("schemas", "plugins", fmt.Sprintf("%s.json", pluginType))

	// 檢查 Schema 文件是否存在
	if _, err := os.Stat(schemaPath); os.IsNotExist(err) {
		pr.logger.Info("警告: 未找到插件 %s 的 JSON Schema 文件，跳過驗證", pluginType)
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
		return fmt.Errorf("插件 %s 配置驗證失敗，錯誤:\n- %s", pluginType, strings.Join(errors, "\n- "))
	}

	pr.logger.Info("✅ 插件 %s 配置通過 JSON Schema 驗證", pluginType)
	return nil
}
