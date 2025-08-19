package knowledge

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"
)

// MemoryProvider 實作基於記憶體的知識庫提供者（用於測試）
type MemoryProvider struct {
	mu     sync.RWMutex
	items  map[string]*KnowledgeItem
	logger *zap.Logger
}

// NewMemoryProvider 創建新的記憶體知識庫提供者
func NewMemoryProvider(logger *zap.Logger) *MemoryProvider {
	return &MemoryProvider{
		items:  make(map[string]*KnowledgeItem),
		logger: logger,
	}
}

// Store 儲存知識項目
func (m *MemoryProvider) Store(ctx context.Context, item *KnowledgeItem) error {
	if item == nil {
		return fmt.Errorf("knowledge item cannot be nil")
	}

	if item.ID == "" {
		return fmt.Errorf("knowledge item ID is required")
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	// 複製項目以避免外部修改
	itemCopy := *item
	if itemCopy.Tags == nil {
		itemCopy.Tags = []string{}
	}
	if itemCopy.Metadata == nil {
		itemCopy.Metadata = make(map[string]string)
	}
	if itemCopy.LessonsLearned == nil {
		itemCopy.LessonsLearned = []string{}
	}
	if itemCopy.ActionItems == nil {
		itemCopy.ActionItems = []ActionItem{}
	}

	// 設置時間戳
	now := time.Now()
	if itemCopy.CreatedAt.IsZero() {
		itemCopy.CreatedAt = now
	}
	itemCopy.UpdatedAt = now

	m.items[item.ID] = &itemCopy

	m.logger.Debug("Knowledge item stored in memory", zap.String("id", item.ID))
	return nil
}

// Retrieve 根據 ID 檢索知識項目
func (m *MemoryProvider) Retrieve(ctx context.Context, id string) (*KnowledgeItem, error) {
	if id == "" {
		return nil, fmt.Errorf("knowledge item ID is required")
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	item, exists := m.items[id]
	if !exists {
		return nil, fmt.Errorf("knowledge item not found: %s", id)
	}

	// 返回複製品以避免外部修改
	itemCopy := *item
	if item.Tags != nil {
		itemCopy.Tags = make([]string, len(item.Tags))
		copy(itemCopy.Tags, item.Tags)
	}
	if item.Metadata != nil {
		itemCopy.Metadata = make(map[string]string)
		for k, v := range item.Metadata {
			itemCopy.Metadata[k] = v
		}
	}
	if item.LessonsLearned != nil {
		itemCopy.LessonsLearned = make([]string, len(item.LessonsLearned))
		copy(itemCopy.LessonsLearned, item.LessonsLearned)
	}
	if item.ActionItems != nil {
		itemCopy.ActionItems = make([]ActionItem, len(item.ActionItems))
		copy(itemCopy.ActionItems, item.ActionItems)
	}

	m.logger.Debug("Knowledge item retrieved from memory", zap.String("id", id))
	return &itemCopy, nil
}

// Search 根據查詢條件搜索知識項目
func (m *MemoryProvider) Search(ctx context.Context, query *SearchQuery) (*SearchResult, error) {
	if query == nil {
		return nil, fmt.Errorf("search query is required")
	}

	startTime := time.Now()

	m.mu.RLock()
	defer m.mu.RUnlock()

	var matchedItems []*KnowledgeItem

	// 遍歷所有項目並檢查匹配條件
	for _, item := range m.items {
		if m.matchesQuery(item, query) {
			// 創建複製品
			itemCopy := *item
			if item.Tags != nil {
				itemCopy.Tags = make([]string, len(item.Tags))
				copy(itemCopy.Tags, item.Tags)
			}
			if item.Metadata != nil {
				itemCopy.Metadata = make(map[string]string)
				for k, v := range item.Metadata {
					itemCopy.Metadata[k] = v
				}
			}
			if item.LessonsLearned != nil {
				itemCopy.LessonsLearned = make([]string, len(item.LessonsLearned))
				copy(itemCopy.LessonsLearned, item.LessonsLearned)
			}
			if item.ActionItems != nil {
				itemCopy.ActionItems = make([]ActionItem, len(item.ActionItems))
				copy(itemCopy.ActionItems, item.ActionItems)
			}
			matchedItems = append(matchedItems, &itemCopy)
		}
	}

	// 排序
	m.sortItems(matchedItems, query.SortBy, query.SortOrder)

	// 分頁
	total := int64(len(matchedItems))
	limit := 50
	if query.Limit > 0 {
		limit = query.Limit
	}
	offset := 0
	if query.Offset > 0 {
		offset = query.Offset
	}

	end := offset + limit
	if end > len(matchedItems) {
		end = len(matchedItems)
	}

	var pagedItems []*KnowledgeItem
	if offset < len(matchedItems) {
		pagedItems = matchedItems[offset:end]
	}

	searchTime := time.Since(startTime)

	result := &SearchResult{
		Items:      pagedItems,
		Total:      total,
		Page:       offset/limit + 1,
		PageSize:   limit,
		HasMore:    int64(offset+limit) < total,
		Query:      query.Query,
		SearchTime: searchTime,
	}

	m.logger.Debug("Memory search completed",
		zap.String("query", query.Query),
		zap.Int("results", len(pagedItems)),
		zap.Int64("total", total),
		zap.Duration("duration", searchTime),
	)

	return result, nil
}

// SimilaritySearch 根據內容相似性搜索
func (m *MemoryProvider) SimilaritySearch(ctx context.Context, content string, limit int) (*SearchResult, error) {
	if content == "" {
		return nil, fmt.Errorf("content for similarity search is required")
	}

	if limit <= 0 {
		limit = 10
	}

	startTime := time.Now()

	m.mu.RLock()
	defer m.mu.RUnlock()

	type itemWithScore struct {
		item  *KnowledgeItem
		score float64
	}

	var scoredItems []itemWithScore

	// 計算每個項目的相似度評分
	for _, item := range m.items {
		score := m.calculateSimilarity(content, item)
		if score > 0.1 { // 只保留相似度超過閾值的項目
			scoredItems = append(scoredItems, itemWithScore{
				item:  item,
				score: score,
			})
		}
	}

	// 按相似度評分排序
	sort.Slice(scoredItems, func(i, j int) bool {
		return scoredItems[i].score > scoredItems[j].score
	})

	// 限制結果數量
	if len(scoredItems) > limit {
		scoredItems = scoredItems[:limit]
	}

	// 轉換為結果項目
	var items []*KnowledgeItem
	for _, scored := range scoredItems {
		// 創建複製品
		itemCopy := *scored.item
		if scored.item.Tags != nil {
			itemCopy.Tags = make([]string, len(scored.item.Tags))
			copy(itemCopy.Tags, scored.item.Tags)
		}
		if scored.item.Metadata != nil {
			itemCopy.Metadata = make(map[string]string)
			for k, v := range scored.item.Metadata {
				itemCopy.Metadata[k] = v
			}
		} else {
			itemCopy.Metadata = make(map[string]string)
		}
		if scored.item.LessonsLearned != nil {
			itemCopy.LessonsLearned = make([]string, len(scored.item.LessonsLearned))
			copy(itemCopy.LessonsLearned, scored.item.LessonsLearned)
		}
		if scored.item.ActionItems != nil {
			itemCopy.ActionItems = make([]ActionItem, len(scored.item.ActionItems))
			copy(itemCopy.ActionItems, scored.item.ActionItems)
		}

		// 添加相似度評分到 metadata
		itemCopy.Metadata["similarity_score"] = fmt.Sprintf("%.4f", scored.score)
		items = append(items, &itemCopy)
	}

	searchTime := time.Since(startTime)

	result := &SearchResult{
		Items:      items,
		Total:      int64(len(items)),
		Page:       1,
		PageSize:   limit,
		HasMore:    false,
		Query:      content,
		SearchTime: searchTime,
	}

	m.logger.Debug("Memory similarity search completed",
		zap.String("content", content[:min(50, len(content))]),
		zap.Int("results", len(items)),
		zap.Duration("duration", searchTime),
	)

	return result, nil
}

// Delete 刪除知識項目
func (m *MemoryProvider) Delete(ctx context.Context, id string) error {
	if id == "" {
		return fmt.Errorf("knowledge item ID is required")
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.items[id]; !exists {
		return fmt.Errorf("knowledge item not found: %s", id)
	}

	delete(m.items, id)

	m.logger.Debug("Knowledge item deleted from memory", zap.String("id", id))
	return nil
}

// Close 關閉提供者（記憶體提供者無需特別清理）
func (m *MemoryProvider) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.items = make(map[string]*KnowledgeItem)
	m.logger.Info("Memory provider closed")
	return nil
}

// matchesQuery 檢查項目是否符合查詢條件
func (m *MemoryProvider) matchesQuery(item *KnowledgeItem, query *SearchQuery) bool {
	// 文字搜索 - 分詞匹配，任一詞匹配即可
	if query.Query != "" {
		searchWords := strings.Fields(strings.ToLower(query.Query))
		searchableText := strings.ToLower(
			item.Title + " " + item.Content + " " + item.RootCause + " " + item.Resolution,
		)
		
		// 只要任一搜索詞匹配即可
		matched := false
		for _, word := range searchWords {
			if strings.Contains(searchableText, word) {
				matched = true
				break
			}
		}
		
		if !matched {
			return false
		}
	}

	// 類別篩選
	if query.Category != "" && item.Category != query.Category {
		return false
	}

	// 標籤篩選
	if len(query.Tags) > 0 {
		if !m.hasAnyTag(item.Tags, query.Tags) {
			return false
		}
	}

	// 其他篩選條件
	for key, value := range query.Filters {
		switch key {
		case "severity":
			if item.Severity != value {
				return false
			}
		case "status":
			if item.Status != value {
				return false
			}
		case "incident_id":
			if item.IncidentID != value {
				return false
			}
		}
	}

	return true
}

// hasAnyTag 檢查是否包含任一指定標籤
func (m *MemoryProvider) hasAnyTag(itemTags, queryTags []string) bool {
	for _, queryTag := range queryTags {
		for _, itemTag := range itemTags {
			if itemTag == queryTag {
				return true
			}
		}
	}
	return false
}

// sortItems 排序項目
func (m *MemoryProvider) sortItems(items []*KnowledgeItem, sortBy, sortOrder string) {
	if sortBy == "" {
		sortBy = "updated_at"
	}
	if sortOrder == "" {
		sortOrder = "desc"
	}

	ascending := sortOrder == "asc"

	sort.Slice(items, func(i, j int) bool {
		var result bool
		switch sortBy {
		case "title":
			result = items[i].Title < items[j].Title
		case "category":
			result = items[i].Category < items[j].Category
		case "created_at":
			result = items[i].CreatedAt.Before(items[j].CreatedAt)
		case "updated_at":
			result = items[i].UpdatedAt.Before(items[j].UpdatedAt)
		case "severity":
			result = items[i].Severity < items[j].Severity
		default:
			result = items[i].UpdatedAt.Before(items[j].UpdatedAt)
		}

		if ascending {
			return result
		}
		return !result
	})
}

// calculateSimilarity 計算內容相似度（簡單的詞彙匹配）
func (m *MemoryProvider) calculateSimilarity(query string, item *KnowledgeItem) float64 {
	queryWords := strings.Fields(strings.ToLower(query))
	if len(queryWords) == 0 {
		return 0
	}

	// 合併所有可搜索的文字
	searchableText := strings.ToLower(
		item.Title + " " + item.Content + " " + item.RootCause + " " + item.Resolution,
	)

	matchCount := 0
	for _, word := range queryWords {
		if strings.Contains(searchableText, word) {
			matchCount++
		}
	}

	// 簡單的相似度計算：匹配詞彙數 / 總詞彙數
	return float64(matchCount) / float64(len(queryWords))
}