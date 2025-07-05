package plugins

import "context"

// CLIPlugin 定義了命令行界面擴展插件的介面。
// 職責: 允許插件向平台的 CLI 工具註冊新的子命令，以擴展命令行功能。
// AI_PLUGIN_TYPE: "cli_plugin"
// AI_IMPL_PACKAGE: "detectviz-platform/internal/domain_logic/plugins/cli/detector_cli"
// AI_IMPL_CONSTRUCTOR: "NewDetectorCLIPlugin"
type CLIPlugin interface {
	Plugin
	// GetCommandName 返回要註冊的命令名稱。
	GetCommandName() string
	// GetDescription 返回命令的簡短描述。
	GetDescription() string
	// Execute 包含命令的執行邏輯。
	Execute(ctx context.Context, args []string) (string, error)
}
