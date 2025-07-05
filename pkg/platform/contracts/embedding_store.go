package contracts

import "context"

// EmbeddingStoreProvider 定義了向量嵌入儲存與查詢功能的介面。
// 職責: 儲存和檢索高維向量，支持相似性搜索。
// AI_PLUGIN_TYPE: "embedding_store_provider"
// AI_IMPL_PACKAGE: "detectviz-platform/internal/infrastructure/platform/embedding_store/chroma_embedding_store"
// AI_IMPL_CONSTRUCTOR: "NewChromaEmbeddingStoreProvider"
type EmbeddingStoreProvider interface {
	StoreEmbedding(ctx context.Context, id string, vector []float32, metadata map[string]any) error
	QueryNearest(ctx context.Context, queryVector []float32, topK int, filter map[string]any) ([]string, error) // 返回最相似的 ID
	GetName() string
}
