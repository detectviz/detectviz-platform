package knowledge

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	pb "github.com/detectviz/detectviz-platform/contracts/gen/go/detectviz/contracts/v1"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/structpb"
)

// createGenericRequest 創建通用請求
func createGenericRequest(method string, data interface{}) (*pb.InvokeRequest, error) {
	genericReq := &KnowledgeGenericRequest{
		Method: method,
		Data:   data,
	}

	payload, err := json.Marshal(genericReq)
	if err != nil {
		return nil, err
	}

	payloadStruct := &structpb.Struct{}
	if err := payloadStruct.UnmarshalJSON(payload); err != nil {
		return nil, err
	}

	return &pb.InvokeRequest{
		Payload: payloadStruct,
	}, nil
}

func TestKnowledgePlugin_Store(t *testing.T) {
	plugin := New()
	if plugin == nil {
		t.Fatal("Failed to create knowledge plugin")
	}

	logger, _ := zap.NewDevelopment()
	plugin.Initialize(logger)

	// 準備測試數據
	item := &KnowledgeItem{
		Title:       "Test Incident",
		Content:     "This is a test incident for knowledge storage",
		Category:    string(CategoryPostmortem),
		Tags:        []string{"test", "incident", "memory"},
		CreatedBy:   "test-user",
		Severity:    string(SeverityMedium),
		Status:      string(StatusPublished),
		IncidentID:  "INC-001",
		RootCause:   "Test root cause",
		Resolution:  "Test resolution",
		LessonsLearned: []string{"Lesson 1", "Lesson 2"},
		ActionItems: []ActionItem{
			{
				ID:          "action-1",
				Description: "Test action item",
				Assignee:    "test-assignee",
				Status:      "open",
				Priority:    "high",
			},
		},
		Metadata: map[string]string{
			"test_key": "test_value",
		},
	}

	// 準備請求
	storeReq := &KnowledgeStoreRequest{
		Item: item,
	}

	req, err := createGenericRequest("store", storeReq)
	if err != nil {
		t.Fatalf("Failed to create generic request: %v", err)
	}

	// 執行測試
	ctx := context.Background()
	resp, err := plugin.Invoke(ctx, req)
	if err != nil {
		t.Fatalf("Store operation failed: %v", err)
	}

	// 驗證回應
	respBytes, err := resp.Result.MarshalJSON()
	if err != nil {
		t.Fatalf("Failed to marshal response: %v", err)
	}

	var storeResp KnowledgeStoreResponse
	if err := json.Unmarshal(respBytes, &storeResp); err != nil {
		t.Fatalf("Failed to unmarshal store response: %v", err)
	}

	if !storeResp.Success {
		t.Errorf("Store operation should succeed, got: %s", storeResp.Message)
	}

	if storeResp.ItemID == "" {
		t.Error("Item ID should not be empty")
	}

	t.Logf("Store test passed, item ID: %s", storeResp.ItemID)
}

func TestKnowledgePlugin_RetrieveAndSearch(t *testing.T) {
	plugin := New()
	if plugin == nil {
		t.Fatal("Failed to create knowledge plugin")
	}

	logger, _ := zap.NewDevelopment()
	plugin.Initialize(logger)

	ctx := context.Background()

	// 首先儲存一個測試項目
	item := &KnowledgeItem{
		ID:          "test-item-001",
		Title:       "Database Connection Timeout",
		Content:     "Investigation of database connection timeout issues during peak hours",
		Category:    string(CategoryPostmortem),
		Tags:        []string{"database", "timeout", "performance"},
		CreatedBy:   "sre-team",
		Severity:    string(SeverityHigh),
		Status:      string(StatusPublished),
		IncidentID:  "INC-DB-001",
		RootCause:   "Connection pool exhaustion due to long-running queries",
		Resolution:  "Optimized queries and increased connection pool size",
		LessonsLearned: []string{
			"Monitor connection pool metrics",
			"Set query timeouts",
		},
		ActionItems: []ActionItem{
			{
				ID:          "action-db-001",
				Description: "Implement connection pool monitoring",
				Assignee:    "db-team",
				Status:      "in-progress",
				Priority:    "high",
			},
		},
		Metadata: map[string]string{
			"environment": "production",
			"service":     "api-gateway",
		},
	}

	// 儲存項目
	storeReq := &KnowledgeStoreRequest{Item: item}
	storeReqPb, err := createGenericRequest("store", storeReq)
	if err != nil {
		t.Fatalf("Failed to create store request: %v", err)
	}

	_, err = plugin.Invoke(ctx, storeReqPb)
	if err != nil {
		t.Fatalf("Failed to store test item: %v", err)
	}

	// 測試檢索
	t.Run("Retrieve", func(t *testing.T) {
		retrieveReq := &KnowledgeRetrieveRequest{
			ItemID: "test-item-001",
		}

		req, err := createGenericRequest("retrieve", retrieveReq)
		if err != nil {
			t.Fatalf("Failed to create retrieve request: %v", err)
		}

		resp, err := plugin.Invoke(ctx, req)
		if err != nil {
			t.Fatalf("Retrieve operation failed: %v", err)
		}

		respBytes, _ := resp.Result.MarshalJSON()
		var retrieveResp KnowledgeRetrieveResponse
		json.Unmarshal(respBytes, &retrieveResp)

		if !retrieveResp.Success {
			t.Errorf("Retrieve operation should succeed, got: %s", retrieveResp.Message)
		}

		if retrieveResp.Item == nil {
			t.Error("Retrieved item should not be nil")
		} else {
			if retrieveResp.Item.Title != item.Title {
				t.Errorf("Expected title %s, got %s", item.Title, retrieveResp.Item.Title)
			}
			if retrieveResp.Item.RootCause != item.RootCause {
				t.Errorf("Expected root cause %s, got %s", item.RootCause, retrieveResp.Item.RootCause)
			}
		}
	})

	// 測試搜索
	t.Run("Search", func(t *testing.T) {
		searchReq := &KnowledgeSearchRequest{
			Query: &SearchQuery{
				Query:    "database timeout",
				Category: string(CategoryPostmortem),
				Limit:    10,
			},
		}

		req, err := createGenericRequest("search", searchReq)
		if err != nil {
			t.Fatalf("Failed to create search request: %v", err)
		}

		resp, err := plugin.Invoke(ctx, req)
		if err != nil {
			t.Fatalf("Search operation failed: %v", err)
		}

		respBytes, _ := resp.Result.MarshalJSON()
		var searchResp KnowledgeSearchResponse
		json.Unmarshal(respBytes, &searchResp)

		if !searchResp.Success {
			t.Errorf("Search operation should succeed, got: %s", searchResp.Message)
		}

		if searchResp.Result == nil {
			t.Error("Search result should not be nil")
		} else {
			if len(searchResp.Result.Items) == 0 {
				t.Error("Search should return at least one item")
			} else {
				found := false
				for _, foundItem := range searchResp.Result.Items {
					if foundItem.ID == "test-item-001" {
						found = true
						break
					}
				}
				if !found {
					t.Error("Search should find the stored item")
				}
			}
		}
	})

	// 測試相似性搜索
	t.Run("SimilaritySearch", func(t *testing.T) {
		simSearchReq := &KnowledgeSimilaritySearchRequest{
			Content: "connection pool exhaustion timeout",
			Limit:   5,
		}

		req, err := createGenericRequest("similarity_search", simSearchReq)
		if err != nil {
			t.Fatalf("Failed to create similarity search request: %v", err)
		}

		resp, err := plugin.Invoke(ctx, req)
		if err != nil {
			t.Fatalf("Similarity search operation failed: %v", err)
		}

		respBytes, _ := resp.Result.MarshalJSON()
		var simSearchResp KnowledgeSimilaritySearchResponse
		json.Unmarshal(respBytes, &simSearchResp)

		if !simSearchResp.Success {
			t.Errorf("Similarity search operation should succeed, got: %s", simSearchResp.Message)
		}

		if simSearchResp.Result != nil && len(simSearchResp.Result.Items) > 0 {
			t.Logf("Similarity search found %d items", len(simSearchResp.Result.Items))
		}
	})
}

func TestKnowledgePlugin_HealthCheck(t *testing.T) {
	plugin := New()
	if plugin == nil {
		t.Fatal("Failed to create knowledge plugin")
	}

	logger, _ := zap.NewDevelopment()
	plugin.Initialize(logger)

	// 測試健康檢查
	err := plugin.HealthCheck()
	if err != nil {
		t.Errorf("Health check should pass, got error: %v", err)
	}
}

func TestKnowledgePlugin_InvalidRequests(t *testing.T) {
	plugin := New()
	if plugin == nil {
		t.Fatal("Failed to create knowledge plugin")
	}

	logger, _ := zap.NewDevelopment()
	plugin.Initialize(logger)

	ctx := context.Background()

	// 測試無效方法
	t.Run("InvalidMethod", func(t *testing.T) {
		req, err := createGenericRequest("invalid_method", nil)
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}

		_, err = plugin.Invoke(ctx, req)
		if err == nil {
			t.Error("Should return error for invalid method")
		}
	})

	// 測試空的 retrieve 請求
	t.Run("EmptyRetrieveRequest", func(t *testing.T) {
		retrieveReq := &KnowledgeRetrieveRequest{
			ItemID: "",
		}

		req, err := createGenericRequest("retrieve", retrieveReq)
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}

		_, err = plugin.Invoke(ctx, req)
		if err == nil {
			t.Error("Should return error for empty item ID")
		}
	})

	// 測試檢索不存在的項目
	t.Run("RetrieveNonExistentItem", func(t *testing.T) {
		retrieveReq := &KnowledgeRetrieveRequest{
			ItemID: "non-existent-item",
		}

		req, err := createGenericRequest("retrieve", retrieveReq)
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}

		resp, err := plugin.Invoke(ctx, req)
		if err != nil {
			t.Fatalf("Invoke should not return error: %v", err)
		}

		respBytes, _ := resp.Result.MarshalJSON()
		var retrieveResp KnowledgeRetrieveResponse
		json.Unmarshal(respBytes, &retrieveResp)

		if retrieveResp.Success {
			t.Error("Should not succeed when retrieving non-existent item")
		}
	})
}

func TestKnowledgeProvider_Factory(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	factory := NewProviderFactory(logger)

	// 測試記憶體 provider 創建
	t.Run("CreateMemoryProvider", func(t *testing.T) {
		config := factory.CreateDefaultConfig()
		if config.Provider != string(ProviderTypeMemory) {
			t.Errorf("Default config should use memory provider, got: %s", config.Provider)
		}

		provider, err := factory.CreateProvider(config)
		if err != nil {
			t.Fatalf("Failed to create memory provider: %v", err)
		}

		if provider == nil {
			t.Error("Provider should not be nil")
		}

		// 測試基本操作
		ctx := context.Background()
		item := &KnowledgeItem{
			ID:      "test-factory-001",
			Title:   "Factory Test Item",
			Content: "Test content for factory test",
			Category: string(CategoryPostmortem),
			CreatedBy: "test",
		}

		err = provider.Store(ctx, item)
		if err != nil {
			t.Errorf("Failed to store item: %v", err)
		}

		retrievedItem, err := provider.Retrieve(ctx, "test-factory-001")
		if err != nil {
			t.Errorf("Failed to retrieve item: %v", err)
		}

		if retrievedItem.Title != item.Title {
			t.Errorf("Expected title %s, got %s", item.Title, retrievedItem.Title)
		}

		provider.Close()
	})

	// 測試配置驗證
	t.Run("ValidateConfig", func(t *testing.T) {
		// 有效配置
		validConfig := factory.CreateDefaultConfig()
		err := factory.ValidateConfig(validConfig)
		if err != nil {
			t.Errorf("Valid config should pass validation: %v", err)
		}

		// 無效配置 - nil
		err = factory.ValidateConfig(nil)
		if err == nil {
			t.Error("Should return error for nil config")
		}

		// 無效配置 - 空 provider
		invalidConfig := &Config{
			Provider: "",
		}
		err = factory.ValidateConfig(invalidConfig)
		if err == nil {
			t.Error("Should return error for empty provider")
		}

		// 無效配置 - 不支援的 provider
		invalidConfig = &Config{
			Provider: "unsupported_provider",
		}
		err = factory.ValidateConfig(invalidConfig)
		if err == nil {
			t.Error("Should return error for unsupported provider")
		}
	})
}

func TestKnowledgePlugin_CloseWithTimeout(t *testing.T) {
	plugin := New()
	if plugin == nil {
		t.Fatal("Failed to create knowledge plugin")
	}

	logger, _ := zap.NewDevelopment()
	plugin.Initialize(logger)

	// 測試正常關閉
	start := time.Now()
	err := plugin.Close()
	duration := time.Since(start)

	if err != nil {
		t.Errorf("Close should succeed: %v", err)
	}

	if duration > 6*time.Second {
		t.Errorf("Close should complete within timeout, took: %v", duration)
	}
}