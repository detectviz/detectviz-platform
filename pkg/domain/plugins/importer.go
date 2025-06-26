package plugins

import "context"

// Importer 定義了數據導入功能的通用介面。
// 職責: 從不同來源（文件、API、數據庫）導入數據到平台。
// AI 擴展點: AI 可生成 `CSVImporter`、`S3Importer` 等具體實現骨架。
type Importer interface {
	Plugin                                               // 繼承通用 Plugin 介面
	ImportData(ctx context.Context, source string) error // 根據來源導入數據
}
