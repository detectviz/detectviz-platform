// Package tools provides development utilities for go-platform
// This file contains plugin scaffolding functionality
package tools

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// ScaffoldPlugin 產生 Go 插件骨架到 go-platform/internal/pluginhost/plugins/<category>/<name>/
// 允許的 category: gateway | observability | collector.input | transform.processor | sink.output
func ScaffoldPlugin(arg string) error {
	cat, name, err := parsePluginArg(arg)
	if err != nil {
		return err
	}

	base := os.Getenv("DETECTVIZ_GO_PLUGIN_BASE")
	if base == "" {
		// 從 tools 目錄相對路徑回到 internal/pluginhost/plugins
		base = "../internal/pluginhost/plugins"
	}

	dir := filepath.Join(base, cat, name)
	if err := os.MkdirAll(filepath.Join(dir, "tests"), 0o755); err != nil {
		return fmt.Errorf("建立目錄失敗: %w", err)
	}

	// 準備檔案內容
	packageName := sanitizePackageName(name)
	pluginID := fmt.Sprintf("%s.%s", cat, name)

	pluginGo := renderPluginGoCode(packageName, pluginID)
	moduleCard, _ := json.MarshalIndent(createDefaultModuleCard(cat, pluginID), "", "  ")
	readme := renderPluginReadme(cat, name, pluginID)
	test := renderPluginTest(packageName)

	// 寫入檔案
	writeFile := func(path string, content []byte) error {
		return os.WriteFile(path, content, 0o644)
	}

	files := map[string][]byte{
		filepath.Join(dir, "plugin.go"):        []byte(pluginGo),
		filepath.Join(dir, "module.card.json"): moduleCard,
		filepath.Join(dir, "README.md"):        []byte(readme),
		filepath.Join(dir, "plugin_test.go"):   []byte(test),
	}

	for path, content := range files {
		if err := writeFile(path, content); err != nil {
			return fmt.Errorf("寫入檔案失敗 %s: %w", path, err)
		}
	}

	fmt.Println("✅ 已建立插件骨架：")
	for path := range files {
		fmt.Println("   📄", path)
	}

	fmt.Println("\n🔧 下一步：")
	fmt.Println("   1. 編輯 plugin.go 實作業務邏輯")
	fmt.Println("   2. 更新 module.card.json 配置")
	fmt.Println("   3. 在 internal/pluginhost/plugins/register/all.go 註冊插件")
	fmt.Println("   4. 運行測試: go test ./internal/pluginhost/plugins/" + cat + "/" + name + "/...")

	return nil
}

func parsePluginArg(arg string) (string, string, error) {
	parts := strings.Split(arg, "/")
	if len(parts) != 2 {
		return "", "", fmt.Errorf("參數格式錯誤，需為 <category>/<name>，例如: observability/health_checker")
	}

	category := parts[0]
	name := parts[1]

	// 擴展支援的類別，包含新增的 observability
	validCategories := map[string]bool{
		"gateway":             true,
		"observability":       true,
		"collector.input":     true,
		"transform.processor": true,
		"sink.output":         true,
	}

	if !validCategories[category] {
		return "", "", fmt.Errorf("不支援的 category: %s，支援的類別: %v",
			category, getMapKeys(validCategories))
	}

	if !regexp.MustCompile(`^[a-z0-9_]+$`).MatchString(name) {
		return "", "", fmt.Errorf("name 只能包含小寫字母、數字和底線: %s", name)
	}

	return category, name, nil
}

func sanitizePackageName(name string) string {
	// Go package 名稱規範化
	out := strings.ToLower(name)
	out = strings.ReplaceAll(out, "-", "_")
	out = strings.ReplaceAll(out, ".", "_")
	return out
}

func renderPluginGoCode(pkg, pluginID string) string {
	timestamp := time.Now().UTC().Format(time.RFC3339)
	return fmt.Sprintf(`// Package %s implements plugin %s
// Auto-generated on %s
package %s

import (
	"context"
	"fmt"
	"time"

	pb "github.com/detectviz/detectviz-platform/contracts/gen/go/detectviz/contracts/v1"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/structpb"
)

// Plugin implements the %s plugin
type Plugin struct {
	logger *zap.Logger
}

// New creates a new instance of the plugin
func New() *Plugin {
	return &Plugin{
		logger: zap.NewNop(), // Will be replaced during registration
	}
}

// Initialize initializes the plugin with logger and configuration
func (p *Plugin) Initialize(logger *zap.Logger) {
	if logger != nil {
		p.logger = logger
	}
}

// Invoke handles plugin invocation requests
func (p *Plugin) Invoke(ctx context.Context, req *pb.InvokeRequest) (*pb.InvokeResponse, error) {
	start := time.Now()
	
	p.logger.Debug("Plugin invoked",
		zap.String("plugin_id", "%s"),
		zap.Any("payload", req.Payload))

	// TODO: Implement your plugin logic here
	// This is a basic echo implementation
	result := req.Payload
	if result == nil {
		var err error
		result, err = structpb.NewStruct(map[string]interface{}{
			"message": "Hello from %s",
			"timestamp": time.Now().UTC().Format(time.RFC3339),
			"success": true,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to create result struct: %%w", err)
		}
	}

	duration := time.Since(start)
	p.logger.Info("Plugin execution completed",
		zap.String("plugin_id", "%s"),
		zap.Duration("duration", duration))

	return &pb.InvokeResponse{
		Result: result,
	}, nil
}

// Close cleans up plugin resources (implements ClosableHandler)
func (p *Plugin) Close() error {
	p.logger.Info("Plugin closing", zap.String("plugin_id", "%s"))
	return nil
}
`, pkg, pluginID, timestamp, pkg, pluginID, pluginID, pluginID, pluginID, pluginID)
}

func createDefaultModuleCard(category, pluginID string) map[string]interface{} {
	return map[string]interface{}{
		"specVersion": "1.1.0",
		"kind":        "plugin",
		"category":    category,
		"id":          pluginID,
		"version":     "0.1.0",
		"description": fmt.Sprintf("Auto-generated plugin for %s", category),
		"observability": map[string]interface{}{
			"spans":   []string{"plugin.invoke"},
			"metrics": []string{"plugin.duration", "plugin.requests_total"},
		},
		"rate_limit": map[string]interface{}{
			"rps":        100,
			"burst":      200,
			"per_tenant": true,
		},
		"resource_limits": map[string]interface{}{
			"concurrency_limit": 50,
			"queue_depth":       500,
			"memory_limit_mb":   128,
		},
		"permissions":  []string{},
		"dependencies": []string{},
	}
}

func renderPluginReadme(category, name, pluginID string) string {
	return fmt.Sprintf(`# %s Plugin

## 概述
此插件屬於 **%s** 類別，提供 %s 相關功能。

- **Plugin ID**: `+"`%s`"+`
- **類別**: `+"`%s`"+`
- **版本**: 0.1.0

## 功能特色
- ✅ 基於 gRPC 的高效能通訊
- ✅ 完整的 OpenTelemetry 可觀測性
- ✅ 資源使用監控和限制
- ✅ 優雅的錯誤處理和重試機制

## 開發指南

### 1. 實作業務邏輯
編輯 `+"`plugin.go`"+` 中的 `+"`Invoke`"+` 方法：

`+"```go"+`
func (p *Plugin) Invoke(ctx context.Context, req *pb.InvokeRequest) (*pb.InvokeResponse, error) {
    // 實作你的業務邏輯
    return &pb.InvokeResponse{Result: result}, nil
}
`+"```"+`

### 2. 配置模組卡
更新 `+"`module.card.json`"+` 中的配置項目：
- `+"`permissions`"+`: 所需權限
- `+"`dependencies`"+`: 依賴的其他插件
- `+"`rate_limit`"+`: 請求限制設定

### 3. 註冊插件
在 `+"`internal/pluginhost/plugins/register/all.go`"+` 中註冊：

`+"```go"+`
import "%s_plugin "github.com/detectviz/detectviz-platform/go-platform/internal/pluginhost/plugins/%s/%s"

func RegisterAll(reg *pluginhost.Registry) error {
    // 其他註冊...
    reg.Register("%s", %s_plugin.New())
    return nil
}
`+"```"+`

### 4. 測試
運行單元測試：
`+"```bash"+`
go test ./internal/pluginhost/plugins/%s/%s/...
`+"```"+`

運行整合測試：
`+"```bash"+`
go test ./internal/pluginhost/... -tags=integration
`+"```"+`

## 部署
插件會隨 go-platform 自動載入，無需額外部署步驟。

## 監控
插件提供以下監控指標：
- `+"`plugin.requests_total`"+`: 總請求數
- `+"`plugin.duration`"+`: 執行時間
- `+"`plugin.errors_total`"+`: 錯誤總數

## 故障排除
1. 檢查日誌: `+"`kubectl logs -f deployment/go-platform | grep %s`"+`
2. 檢查指標: 在 Grafana 中查看插件相關指標
3. 檢查追蹤: 在 Jaeger 中查看請求鏈路

## 相關文檔
- [插件開發指南](../../README.md)
- [可觀測性規範](../../../../docs/observability.md)
- [安全性指南](../../../../docs/security.md)
`,
		strings.Title(strings.ReplaceAll(name, "_", " ")),
		category, name, pluginID, category,
		name, category, name, name, pluginID, name, category, name, pluginID)
}

func renderPluginTest(packageName string) string {
	return fmt.Sprintf(`package %s

import (
	"context"
	"testing"

	pb "github.com/detectviz/detectviz-platform/contracts/gen/go/detectviz/contracts/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/structpb"
)

func TestPlugin_Invoke(t *testing.T) {
	plugin := New()
	ctx := context.Background()

	t.Run("basic invocation", func(t *testing.T) {
		req := &pb.InvokeRequest{
			Payload: nil, // Test with empty payload
		}

		resp, err := plugin.Invoke(ctx, req)
		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.NotNil(t, resp.Result)
	})

	t.Run("with payload", func(t *testing.T) {
		payload, err := structpb.NewStruct(map[string]interface{}{
			"test_key": "test_value",
		})
		require.NoError(t, err)

		req := &pb.InvokeRequest{
			Payload: payload,
		}

		resp, err := plugin.Invoke(ctx, req)
		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.Equal(t, payload, resp.Result)
	})
}

func TestPlugin_Close(t *testing.T) {
	plugin := New()
	err := plugin.Close()
	assert.NoError(t, err)
}

// BenchmarkPlugin_Invoke benchmarks the plugin invoke performance
func BenchmarkPlugin_Invoke(b *testing.B) {
	plugin := New()
	ctx := context.Background()
	
	payload, _ := structpb.NewStruct(map[string]interface{}{
		"benchmark": true,
	})
	
	req := &pb.InvokeRequest{Payload: payload}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := plugin.Invoke(ctx, req)
		if err != nil {
			b.Fatal(err)
		}
	}
}
`, packageName)
}

// Helper function to get map keys
func getMapKeys(m map[string]bool) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
