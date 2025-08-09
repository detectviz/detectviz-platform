package pluginnew

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// Scaffold 產生 Go 插件骨架到 go-platform/internal/pluginhost/plugins/<category>/<name>/
// 允許的 category: capability.gateway | collector.input | transform.processor | sink.output
func Scaffold(arg string) error {
	cat, name, err := parse(arg)
	if err != nil {
		return err
	}
	base := os.Getenv("DETECTVIZ_GO_PLUGIN_BASE")
	if base == "" {
		base = "go-platform/internal/pluginhost/plugins"
	}
	dir := filepath.Join(base, cat, name)
	if err := os.MkdirAll(filepath.Join(dir, "tests"), 0o755); err != nil {
		return fmt.Errorf("建立目錄失敗: %w", err)
	}

	// 準備檔案內容
	packageName := sanitizePkg(name)
	pluginID := fmt.Sprintf("detectviz.plugins.%s", name)
	toolID := fmt.Sprintf("detectviz.tools.%s", name)

	pluginGo := renderPluginGo(packageName, pluginID, toolID)
	moduleCard, _ := json.MarshalIndent(defaultModuleCard(cat, pluginID), "", "  ")
	readme := renderReadme(cat, name, toolID)
	test := renderTest(name)

	// 寫入檔案
	write := func(p string, b []byte) error { return os.WriteFile(p, b, 0o644) }
	if err := write(filepath.Join(dir, "plugin.go"), []byte(pluginGo)); err != nil {
		return err
	}
	if err := write(filepath.Join(dir, "module.card.json"), moduleCard); err != nil {
		return err
	}
	if err := write(filepath.Join(dir, "README.md"), []byte(readme)); err != nil {
		return err
	}
	if err := write(filepath.Join(dir, "tests", "plugin_e2e_test.go"), []byte(test)); err != nil {
		return err
	}

	fmt.Println("已建立插件骨架：")
	fmt.Println("  ", filepath.Join(dir, "plugin.go"))
	fmt.Println("  ", filepath.Join(dir, "module.card.json"))
	fmt.Println("  ", filepath.Join(dir, "README.md"))
	fmt.Println("  ", filepath.Join(dir, "tests", "plugin_e2e_test.go"))
	return nil
}

func parse(arg string) (string, string, error) {
	parts := strings.Split(arg, "/")
	if len(parts) != 2 {
		return "", "", fmt.Errorf("參數需為 <category>/<name>")
	}
	cat := parts[0]
	name := parts[1]
	switch cat {
	case "capability.gateway", "collector.input", "transform.processor", "sink.output":
	default:
		return "", "", fmt.Errorf("不支援的 category: %s", cat)
	}
	if !regexp.MustCompile(`^[a-z0-9_]+$`).MatchString(name) {
		return "", "", fmt.Errorf("name 嚴格限制 [a-z0-9_]+: %s", name)
	}
	return cat, name, nil
}

func sanitizePkg(name string) string {
	// Go package 名稱：移除非字母數字，保留底線
	out := strings.ToLower(name)
	out = strings.ReplaceAll(out, "-", "_")
	out = strings.ReplaceAll(out, ".", "_")
	return out
}

func renderPluginGo(pkg, pluginID, toolID string) string {
	ts := time.Now().UTC().Format(time.RFC3339)
	return fmt.Sprintf(`package %s

import (
    "context"
    "time"

    contractspb "github.com/detectviz/detectviz-platform/contracts/gen/go/detectviz/contracts/v1"
    statuspb "google.golang.org/genproto/googleapis/rpc/status"
    structpb "google.golang.org/protobuf/types/known/structpb"
)

// 自動產生於 %s
// 最小能力：回傳輸入 payload（echo）
type Plugin struct{}

func New() *Plugin { return &Plugin{} }

func (p *Plugin) Invoke(ctx context.Context, req *contractspb.ToolInvokeRequest) (*contractspb.ToolInvokeReply, error) {
    start := time.Now()

    var res *structpb.Struct
    if req.GetPayload() != nil {
        res = req.GetPayload()
    } else {
        res, _ = structpb.NewStruct(map[string]interface{}{"echo": true})
    }

    return &contractspb.ToolInvokeReply{
        Result: res,
        Status: &statuspb.Status{Code: 0, Message: "OK"},
        ExecMeta: &contractspb.ToolExecutionMeta{
            Attempt:    1,
            DurationMs: uint64(time.Since(start) / time.Millisecond),
            PluginId:   "%s",
        },
    }, nil
}
`, pkg, ts, pluginID)
}

func defaultModuleCard(category, pluginID string) map[string]interface{} {
	return map[string]interface{}{
		"specVersion": "1.1.0",
		"kind":        "plugin",
		"category":    category,
		"id":          pluginID,
		"version":     "0.1.0",
		"observability": map[string]interface{}{
			"spans": []string{"capability.request"},
		},
		"rate_limit": map[string]interface{}{
			"rps":        64,
			"burst":      128,
			"per_tenant": true,
		},
		"a2a": map[string]interface{}{
			"concurrency_limit": 32,
			"queue_depth":       200,
		},
		"permissions": []string{},
	}
}

func renderReadme(category, name, toolID string) string {
	return fmt.Sprintf("# %s/%s\n\n最小骨架插件，回傳輸入 payload。\n\n- Tool ID：`%s`\n- 分類：`%s`\n- 觀測：span `capability.request`（由宿主框架統一打點）\n\n## 參考\n- contracts/schemas/module.card.schema.json\n- spec.md（插件開發規範）\n", category, name, toolID, category)
}

func renderTest(name string) string {
	return fmt.Sprintf(`package tests

import "testing"

func Test_%s_Echo(t *testing.T) {
    // TODO: 啟動 ToolBridge，呼叫 Invoke，驗證回傳
}
`, name)
}
