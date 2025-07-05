package embedding_store

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"detectviz-platform/pkg/platform/contracts"
)

// ChromaEmbeddingStoreProvider 實現了 EmbeddingStoreProvider 介面，提供 Chroma 向量資料庫集成
// 職責: 與 Chroma 向量資料庫交互，儲存和查詢向量嵌入
type ChromaEmbeddingStoreProvider struct {
	baseURL        string
	collectionName string
	httpClient     *http.Client
	logger         contracts.Logger
}

// ChromaConfig 定義 Chroma 嵌入存儲提供者的配置
type ChromaConfig struct {
	BaseURL        string `yaml:"base_url" json:"base_url"`
	CollectionName string `yaml:"collection_name" json:"collection_name"`
	Timeout        string `yaml:"timeout" json:"timeout"`
}

// ChromaCreateCollectionRequest 定義創建集合的請求結構
type ChromaCreateCollectionRequest struct {
	Name     string                 `json:"name"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// ChromaAddRequest 定義添加向量的請求結構
type ChromaAddRequest struct {
	Embeddings [][]float32              `json:"embeddings"`
	Documents  []string                 `json:"documents,omitempty"`
	Metadatas  []map[string]interface{} `json:"metadatas,omitempty"`
	IDs        []string                 `json:"ids"`
}

// ChromaQueryRequest 定義查詢向量的請求結構
type ChromaQueryRequest struct {
	QueryEmbeddings [][]float32            `json:"query_embeddings"`
	NResults        int                    `json:"n_results"`
	Where           map[string]interface{} `json:"where,omitempty"`
	Include         []string               `json:"include,omitempty"`
}

// ChromaQueryResponse 定義查詢響應結構
type ChromaQueryResponse struct {
	IDs       [][]string                 `json:"ids"`
	Distances [][]float32                `json:"distances"`
	Metadatas [][]map[string]interface{} `json:"metadatas"`
	Documents [][]string                 `json:"documents"`
}

// ChromaCollection 定義集合結構
type ChromaCollection struct {
	Name     string                 `json:"name"`
	ID       string                 `json:"id"`
	Metadata map[string]interface{} `json:"metadata"`
}

// NewChromaEmbeddingStoreProvider 創建新的 Chroma 嵌入存儲提供者實例
func NewChromaEmbeddingStoreProvider(config ChromaConfig, logger contracts.Logger) (contracts.EmbeddingStoreProvider, error) {
	if config.BaseURL == "" {
		config.BaseURL = "http://localhost:8000"
	}

	if config.CollectionName == "" {
		config.CollectionName = "detectviz_embeddings"
	}

	timeout := 30 * time.Second
	if config.Timeout != "" {
		if t, err := time.ParseDuration(config.Timeout); err == nil {
			timeout = t
		}
	}

	httpClient := &http.Client{
		Timeout: timeout,
	}

	provider := &ChromaEmbeddingStoreProvider{
		baseURL:        config.BaseURL,
		collectionName: config.CollectionName,
		httpClient:     httpClient,
		logger:         logger,
	}

	// 初始化時創建集合
	if err := provider.ensureCollection(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to ensure collection: %w", err)
	}

	logger.Info("初始化 Chroma 嵌入存儲提供者",
		"base_url", config.BaseURL,
		"collection", config.CollectionName)

	return provider, nil
}

// StoreEmbedding 儲存向量嵌入
func (c *ChromaEmbeddingStoreProvider) StoreEmbedding(ctx context.Context, id string, vector []float32, metadata map[string]any) error {
	if id == "" {
		return fmt.Errorf("id cannot be empty")
	}

	if len(vector) == 0 {
		return fmt.Errorf("vector cannot be empty")
	}

	// 構建請求
	request := ChromaAddRequest{
		Embeddings: [][]float32{vector},
		IDs:        []string{id},
	}

	if metadata != nil {
		request.Metadatas = []map[string]interface{}{metadata}
	}

	// 序列化請求
	requestBody, err := json.Marshal(request)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	// 構建 URL
	url := fmt.Sprintf("%s/api/v1/collections/%s/add", c.baseURL, c.collectionName)

	// 創建 HTTP 請求
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(requestBody))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	// 發送請求
	c.logger.Debug("儲存向量嵌入", "id", id, "vector_dim", len(vector))

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// 檢查響應狀態
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("HTTP error: %d", resp.StatusCode)
	}

	c.logger.Info("向量嵌入儲存成功", "id", id, "vector_dim", len(vector))

	return nil
}

// QueryNearest 查詢最相似的向量
func (c *ChromaEmbeddingStoreProvider) QueryNearest(ctx context.Context, queryVector []float32, topK int, filter map[string]any) ([]string, error) {
	if len(queryVector) == 0 {
		return nil, fmt.Errorf("query vector cannot be empty")
	}

	if topK <= 0 {
		topK = 10
	}

	// 構建請求
	request := ChromaQueryRequest{
		QueryEmbeddings: [][]float32{queryVector},
		NResults:        topK,
		Include:         []string{"metadatas", "distances"},
	}

	if filter != nil {
		request.Where = filter
	}

	// 序列化請求
	requestBody, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// 構建 URL
	url := fmt.Sprintf("%s/api/v1/collections/%s/query", c.baseURL, c.collectionName)

	// 創建 HTTP 請求
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	// 發送請求
	c.logger.Debug("查詢相似向量", "vector_dim", len(queryVector), "top_k", topK)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// 檢查響應狀態
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP error: %d", resp.StatusCode)
	}

	// 解析響應
	var queryResp ChromaQueryResponse
	if err := json.NewDecoder(resp.Body).Decode(&queryResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// 提取 ID
	var results []string
	if len(queryResp.IDs) > 0 {
		results = queryResp.IDs[0]
	}

	c.logger.Info("相似向量查詢成功",
		"query_vector_dim", len(queryVector),
		"results_count", len(results),
		"top_k", topK)

	return results, nil
}

// GetName 返回提供者名稱
func (c *ChromaEmbeddingStoreProvider) GetName() string {
	return "chroma_embedding_store_provider"
}

// ensureCollection 確保集合存在
func (c *ChromaEmbeddingStoreProvider) ensureCollection(ctx context.Context) error {
	// 檢查集合是否存在
	url := fmt.Sprintf("%s/api/v1/collections/%s", c.baseURL, c.collectionName)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create get collection request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to check collection: %w", err)
	}
	defer resp.Body.Close()

	// 如果集合存在，直接返回
	if resp.StatusCode == http.StatusOK {
		c.logger.Debug("集合已存在", "collection", c.collectionName)
		return nil
	}

	// 如果集合不存在，創建它
	if resp.StatusCode == http.StatusNotFound {
		return c.createCollection(ctx)
	}

	return fmt.Errorf("unexpected status code when checking collection: %d", resp.StatusCode)
}

// createCollection 創建集合
func (c *ChromaEmbeddingStoreProvider) createCollection(ctx context.Context) error {
	request := ChromaCreateCollectionRequest{
		Name: c.collectionName,
		Metadata: map[string]interface{}{
			"description": "Detectviz platform embeddings collection",
			"created_by":  "detectviz-platform",
		},
	}

	requestBody, err := json.Marshal(request)
	if err != nil {
		return fmt.Errorf("failed to marshal create collection request: %w", err)
	}

	url := fmt.Sprintf("%s/api/v1/collections", c.baseURL)

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(requestBody))
	if err != nil {
		return fmt.Errorf("failed to create collection request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to create collection: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("failed to create collection, status code: %d", resp.StatusCode)
	}

	c.logger.Info("集合創建成功", "collection", c.collectionName)
	return nil
}
