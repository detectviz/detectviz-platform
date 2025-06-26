package plugins

import "context"

// Plugin 是所有 Detectviz 應用程式級插件的基本契約。
// 職責: 提供插件的通用識別和生命週期管理。所有應用程式級插件都應實現此介面。
// AI 擴展點: AI 生成新插件時，需確保其實現 Plugin 介面。
type Plugin interface {
	GetName() string                                            // 返回插件的唯一名稱
	Init(ctx context.Context, cfg map[string]interface{}) error // 插件初始化，接收配置
	Start(ctx context.Context) error                            // 插件啟動，例如啟動背景任務
	Stop(ctx context.Context) error                             // 插件停止，清理資源
}
