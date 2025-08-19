// Package knowledge 實作知識庫管理插件
package knowledge

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	pb "github.com/detectviz/detectviz-platform/contracts/gen/go/detectviz/contracts/v1"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/structpb"
)

// Plugin 實作知識庫管理插件，符合 Handler 介面
type Plugin struct {
	provider Provider
	factory  *ProviderFactory
	config   *Config
	logger   *zap.Logger
}

// KnowledgeStoreRequest 知識儲存請求
type KnowledgeStoreRequest struct {
	Item *KnowledgeItem `json:"item"`
}

// KnowledgeStoreResponse 知識儲存回應
type KnowledgeStoreResponse struct {
	Success   bool   `json:"success"`
	ItemID    string `json:"item_id"`
	Message   string `json:"message,omitempty"`
	Timestamp time.Time `json:"timestamp"`
}

// KnowledgeRetrieveRequest 知識檢索請求
type KnowledgeRetrieveRequest struct {
	ItemID string `json:"item_id"`
}

// KnowledgeRetrieveResponse 知識檢索回應
type KnowledgeRetrieveResponse struct {
	Success   bool           `json:"success"`
	Item      *KnowledgeItem `json:"item,omitempty"`
	Message   string         `json:"message,omitempty"`
	Timestamp time.Time      `json:"timestamp"`
}

// KnowledgeSearchRequest 知識搜索請求
type KnowledgeSearchRequest struct {
	Query *SearchQuery `json:"query"`
}

// KnowledgeSearchResponse 知識搜索回應
type KnowledgeSearchResponse struct {
	Success   bool          `json:"success"`
	Result    *SearchResult `json:"result,omitempty"`
	Message   string        `json:"message,omitempty"`
	Timestamp time.Time     `json:"timestamp"`
}

// KnowledgeSimilaritySearchRequest 相似性搜索請求
type KnowledgeSimilaritySearchRequest struct {
	Content string `json:"content"`
	Limit   int    `json:"limit,omitempty"`
}

// KnowledgeSimilaritySearchResponse 相似性搜索回應
type KnowledgeSimilaritySearchResponse struct {
	Success   bool          `json:"success"`
	Result    *SearchResult `json:"result,omitempty"`
	Message   string        `json:"message,omitempty"`
	Timestamp time.Time     `json:"timestamp"`
}

// KnowledgeDeleteRequest 知識刪除請求
type KnowledgeDeleteRequest struct {
	ItemID string `json:"item_id"`
}

// KnowledgeDeleteResponse 知識刪除回應
type KnowledgeDeleteResponse struct {
	Success   bool   `json:"success"`
	ItemID    string `json:"item_id"`
	Message   string `json:"message,omitempty"`
	Timestamp time.Time `json:"timestamp"`
}

// New 創建新的知識庫插件
func New() *Plugin {
	logger := zap.NewNop()
	factory := NewProviderFactory(logger)
	config := factory.CreateDefaultConfig()

	// 使用 Memory Provider 作為默認
	provider, err := factory.CreateProvider(config)
	if err != nil {
		logger.Error("Failed to create default knowledge provider", zap.Error(err))
		return nil
	}

	return &Plugin{
		provider: provider,
		factory:  factory,
		config:   config,
		logger:   logger,
	}
}

// Initialize 初始化插件
func (p *Plugin) Initialize(logger *zap.Logger) {
	if logger != nil {
		p.logger = logger
		p.factory = NewProviderFactory(logger)
	}
}

// ConfigureProvider 配置知識庫提供者
func (p *Plugin) ConfigureProvider(config *Config) error {
	if config == nil {
		return fmt.Errorf("config cannot be nil")
	}

	if err := p.factory.ValidateConfig(config); err != nil {
		return fmt.Errorf("invalid config: %w", err)
	}

	// 關閉現有 provider
	if p.provider != nil {
		if err := p.provider.Close(); err != nil {
			p.logger.Warn("Failed to close existing provider", zap.Error(err))
		}
	}

	// 創建新 provider
	provider, err := p.factory.CreateProvider(config)
	if err != nil {
		return fmt.Errorf("failed to create provider: %w", err)
	}

	p.provider = provider
	p.config = config

	p.logger.Info("Knowledge provider configured successfully", 
		zap.String("provider_type", config.Provider))

	return nil
}

// KnowledgeGenericRequest 通用知識庫請求，包含方法和具體請求
type KnowledgeGenericRequest struct {
	Method string      `json:"method"`
	Data   interface{} `json:"data"`
}

// Invoke 實作 Handler 介面 - 處理插件調用
func (p *Plugin) Invoke(ctx context.Context, req *pb.InvokeRequest) (*pb.InvokeResponse, error) {
	if p.provider == nil {
		return nil, fmt.Errorf("knowledge provider not initialized")
	}

	// 解析請求參數
	payloadBytes, err := req.Payload.MarshalJSON()
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload: %w", err)
	}

	// 首先嘗試解析為通用請求格式
	var genericReq KnowledgeGenericRequest
	if err := json.Unmarshal(payloadBytes, &genericReq); err != nil {
		return nil, fmt.Errorf("failed to parse generic request: %w", err)
	}

	p.logger.Debug("Processing knowledge request",
		zap.String("method", genericReq.Method),
		zap.Int("payload_size", len(payloadBytes)),
	)

	// 根據方法路由請求
	switch genericReq.Method {
	case "store":
		return p.handleStoreGeneric(ctx, genericReq.Data)
	case "retrieve":
		return p.handleRetrieveGeneric(ctx, genericReq.Data)
	case "search":
		return p.handleSearchGeneric(ctx, genericReq.Data)
	case "similarity_search":
		return p.handleSimilaritySearchGeneric(ctx, genericReq.Data)
	case "delete":
		return p.handleDeleteGeneric(ctx, genericReq.Data)
	default:
		return nil, fmt.Errorf("unknown method: %s", genericReq.Method)
	}
}

// handleStoreGeneric 處理知識儲存
func (p *Plugin) handleStoreGeneric(ctx context.Context, data interface{}) (*pb.InvokeResponse, error) {
	// 轉換數據為 JSON 再解析為具體類型
	dataBytes, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal data: %w", err)
	}

	var storeReq KnowledgeStoreRequest
	if err := json.Unmarshal(dataBytes, &storeReq); err != nil {
		return nil, fmt.Errorf("failed to parse store request: %w", err)
	}

	if storeReq.Item == nil {
		return nil, fmt.Errorf("knowledge item is required")
	}

	// 生成 ID 如果沒有提供
	if storeReq.Item.ID == "" {
		storeReq.Item.ID = generateKnowledgeID()
	}

	// 儲存知識項目
	err = p.provider.Store(ctx, storeReq.Item)
	
	response := &KnowledgeStoreResponse{
		Success:   err == nil,
		ItemID:    storeReq.Item.ID,
		Timestamp: time.Now(),
	}

	if err != nil {
		response.Message = err.Error()
		p.logger.Error("Failed to store knowledge item", 
			zap.String("item_id", storeReq.Item.ID), 
			zap.Error(err))
	} else {
		p.logger.Info("Knowledge item stored successfully", 
			zap.String("item_id", storeReq.Item.ID),
			zap.String("title", storeReq.Item.Title))
	}

	return p.buildResponse(response)
}

// handleRetrieveGeneric 處理知識檢索
func (p *Plugin) handleRetrieveGeneric(ctx context.Context, data interface{}) (*pb.InvokeResponse, error) {
	// 轉換數據為 JSON 再解析為具體類型
	dataBytes, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal data: %w", err)
	}

	var retrieveReq KnowledgeRetrieveRequest
	if err := json.Unmarshal(dataBytes, &retrieveReq); err != nil {
		return nil, fmt.Errorf("failed to parse retrieve request: %w", err)
	}

	if retrieveReq.ItemID == "" {
		return nil, fmt.Errorf("item ID is required")
	}

	// 檢索知識項目
	item, err := p.provider.Retrieve(ctx, retrieveReq.ItemID)
	
	response := &KnowledgeRetrieveResponse{
		Success:   err == nil,
		Timestamp: time.Now(),
	}

	if err != nil {
		response.Message = err.Error()
		p.logger.Error("Failed to retrieve knowledge item", 
			zap.String("item_id", retrieveReq.ItemID), 
			zap.Error(err))
	} else {
		response.Item = item
		p.logger.Debug("Knowledge item retrieved successfully", 
			zap.String("item_id", retrieveReq.ItemID))
	}

	return p.buildResponse(response)
}

// handleSearchGeneric 處理知識搜索
func (p *Plugin) handleSearchGeneric(ctx context.Context, data interface{}) (*pb.InvokeResponse, error) {
	// 轉換數據為 JSON 再解析為具體類型
	dataBytes, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal data: %w", err)
	}

	var searchReq KnowledgeSearchRequest
	if err := json.Unmarshal(dataBytes, &searchReq); err != nil {
		return nil, fmt.Errorf("failed to parse search request: %w", err)
	}

	if searchReq.Query == nil {
		return nil, fmt.Errorf("search query is required")
	}

	// 執行搜索
	result, err := p.provider.Search(ctx, searchReq.Query)
	
	response := &KnowledgeSearchResponse{
		Success:   err == nil,
		Timestamp: time.Now(),
	}

	if err != nil {
		response.Message = err.Error()
		p.logger.Error("Failed to search knowledge", 
			zap.String("query", searchReq.Query.Query), 
			zap.Error(err))
	} else {
		response.Result = result
		p.logger.Debug("Knowledge search completed successfully", 
			zap.String("query", searchReq.Query.Query),
			zap.Int("results", len(result.Items)))
	}

	return p.buildResponse(response)
}

// handleSimilaritySearchGeneric 處理相似性搜索
func (p *Plugin) handleSimilaritySearchGeneric(ctx context.Context, data interface{}) (*pb.InvokeResponse, error) {
	// 轉換數據為 JSON 再解析為具體類型
	dataBytes, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal data: %w", err)
	}

	var simSearchReq KnowledgeSimilaritySearchRequest
	if err := json.Unmarshal(dataBytes, &simSearchReq); err != nil {
		return nil, fmt.Errorf("failed to parse similarity search request: %w", err)
	}

	if simSearchReq.Content == "" {
		return nil, fmt.Errorf("content for similarity search is required")
	}

	if simSearchReq.Limit <= 0 {
		simSearchReq.Limit = 10
	}

	// 執行相似性搜索
	result, err := p.provider.SimilaritySearch(ctx, simSearchReq.Content, simSearchReq.Limit)
	
	response := &KnowledgeSimilaritySearchResponse{
		Success:   err == nil,
		Timestamp: time.Now(),
	}

	if err != nil {
		response.Message = err.Error()
		p.logger.Error("Failed to perform similarity search", 
			zap.String("content", simSearchReq.Content[:min(50, len(simSearchReq.Content))]), 
			zap.Error(err))
	} else {
		response.Result = result
		p.logger.Debug("Similarity search completed successfully", 
			zap.String("content", simSearchReq.Content[:min(50, len(simSearchReq.Content))]),
			zap.Int("results", len(result.Items)))
	}

	return p.buildResponse(response)
}

// handleDeleteGeneric 處理知識刪除
func (p *Plugin) handleDeleteGeneric(ctx context.Context, data interface{}) (*pb.InvokeResponse, error) {
	// 轉換數據為 JSON 再解析為具體類型
	dataBytes, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal data: %w", err)
	}

	var deleteReq KnowledgeDeleteRequest
	if err := json.Unmarshal(dataBytes, &deleteReq); err != nil {
		return nil, fmt.Errorf("failed to parse delete request: %w", err)
	}

	if deleteReq.ItemID == "" {
		return nil, fmt.Errorf("item ID is required")
	}

	// 刪除知識項目
	err = p.provider.Delete(ctx, deleteReq.ItemID)
	
	response := &KnowledgeDeleteResponse{
		Success:   err == nil,
		ItemID:    deleteReq.ItemID,
		Timestamp: time.Now(),
	}

	if err != nil {
		response.Message = err.Error()
		p.logger.Error("Failed to delete knowledge item", 
			zap.String("item_id", deleteReq.ItemID), 
			zap.Error(err))
	} else {
		p.logger.Info("Knowledge item deleted successfully", 
			zap.String("item_id", deleteReq.ItemID))
	}

	return p.buildResponse(response)
}


// Close 實作 ClosableHandler 介面 - 清理資源
func (p *Plugin) Close() error {
	return p.CloseWithContext(context.Background())
}

// CloseWithContext 帶超時控制的關閉方法
func (p *Plugin) CloseWithContext(ctx context.Context) error {
	// 設置 5 秒超時
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// 等待進行中的請求完成
	done := make(chan error, 1)
	go func() {
		// 關閉 provider
		if p.provider != nil {
			if err := p.provider.Close(); err != nil {
				p.logger.Warn("Failed to close knowledge provider", zap.Error(err))
				done <- err
				return
			}
		}

		p.logger.Info("Knowledge plugin closed")
		done <- nil
	}()

	select {
	case err := <-done:
		return err
	case <-ctx.Done():
		p.logger.Error("Close timeout exceeded", zap.Error(ctx.Err()))
		return fmt.Errorf("close timeout: %w", ctx.Err())
	}
}

// HealthCheck 實作 HealthAwareHandler 介面 - 健康檢查
func (p *Plugin) HealthCheck() error {
	// 檢查 provider 是否可用
	if p.provider == nil {
		return fmt.Errorf("knowledge provider not initialized")
	}

	// 嘗試執行簡單的搜索來驗證 provider 功能
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	testQuery := &SearchQuery{
		Query: "test",
		Limit: 1,
	}

	_, err := p.provider.Search(ctx, testQuery)
	if err != nil {
		return fmt.Errorf("knowledge provider health check failed: %w", err)
	}

	return nil
}

// buildResponse 建構回應
func (p *Plugin) buildResponse(response interface{}) (*pb.InvokeResponse, error) {
	// 序列化回應
	responseBytes, err := json.Marshal(response)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal response: %w", err)
	}

	// 轉換為 structpb.Struct
	result := &structpb.Struct{}
	if err := result.UnmarshalJSON(responseBytes); err != nil {
		return nil, fmt.Errorf("failed to convert response to struct: %w", err)
	}

	return &pb.InvokeResponse{
		Result: result,
	}, nil
}

// generateKnowledgeID 生成知識項目 ID
func generateKnowledgeID() string {
	return fmt.Sprintf("knowledge_%d", time.Now().UnixNano())
}