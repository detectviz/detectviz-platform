// Package knowledge 提供知識庫管理能力
package knowledge

import (
	"context"
	"time"
)

// Provider 定義知識庫提供者接口
type Provider interface {
	// Store 儲存知識項目
	Store(ctx context.Context, item *KnowledgeItem) error
	
	// Retrieve 根據 ID 檢索知識項目
	Retrieve(ctx context.Context, id string) (*KnowledgeItem, error)
	
	// Search 根據查詢條件搜索知識項目
	Search(ctx context.Context, query *SearchQuery) (*SearchResult, error)
	
	// SimilaritySearch 根據內容相似性搜索
	SimilaritySearch(ctx context.Context, content string, limit int) (*SearchResult, error)
	
	// Delete 刪除知識項目
	Delete(ctx context.Context, id string) error
	
	// Close 關閉提供者連接
	Close() error
}

// KnowledgeItem 表示一個知識項目
type KnowledgeItem struct {
	ID          string            `json:"id"`
	Title       string            `json:"title"`
	Content     string            `json:"content"`
	Category    string            `json:"category"`
	Tags        []string          `json:"tags"`
	Metadata    map[string]string `json:"metadata"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
	CreatedBy   string            `json:"created_by"`
	Severity    string            `json:"severity,omitempty"`
	Status      string            `json:"status,omitempty"`
	
	// 事後複盤特有欄位
	IncidentID    string    `json:"incident_id,omitempty"`
	RootCause     string    `json:"root_cause,omitempty"`
	Resolution    string    `json:"resolution,omitempty"`
	LessonsLearned []string `json:"lessons_learned,omitempty"`
	ActionItems   []ActionItem `json:"action_items,omitempty"`
}

// ActionItem 表示行動項目
type ActionItem struct {
	ID          string    `json:"id"`
	Description string    `json:"description"`
	Assignee    string    `json:"assignee"`
	Status      string    `json:"status"`
	Priority    string    `json:"priority"`
	DueDate     *time.Time `json:"due_date,omitempty"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
}

// SearchQuery 表示搜索查詢
type SearchQuery struct {
	Query     string            `json:"query"`
	Category  string            `json:"category,omitempty"`
	Tags      []string          `json:"tags,omitempty"`
	Filters   map[string]string `json:"filters,omitempty"`
	SortBy    string            `json:"sort_by,omitempty"`
	SortOrder string            `json:"sort_order,omitempty"`
	Limit     int               `json:"limit,omitempty"`
	Offset    int               `json:"offset,omitempty"`
}

// SearchResult 表示搜索結果
type SearchResult struct {
	Items      []*KnowledgeItem `json:"items"`
	Total      int64           `json:"total"`
	Page       int             `json:"page"`
	PageSize   int             `json:"page_size"`
	HasMore    bool            `json:"has_more"`
	Query      string          `json:"query"`
	SearchTime time.Duration   `json:"search_time"`
}

// SimilarityScore 表示相似度評分結果
type SimilarityScore struct {
	Item  *KnowledgeItem `json:"item"`
	Score float64        `json:"score"`
}

// Config 表示知識庫配置
type Config struct {
	// Provider 類型 (memory, postgresql, etc.)
	Provider string `yaml:"provider" json:"provider"`
	
	// 資料庫連接配置
	Database *DatabaseConfig `yaml:"database,omitempty" json:"database,omitempty"`
	
	// 相似性搜索配置
	Similarity *SimilarityConfig `yaml:"similarity,omitempty" json:"similarity,omitempty"`
	
	// 快取配置
	Cache *CacheConfig `yaml:"cache,omitempty" json:"cache,omitempty"`
}

// DatabaseConfig 資料庫配置
type DatabaseConfig struct {
	Host     string `yaml:"host" json:"host"`
	Port     int    `yaml:"port" json:"port"`
	Database string `yaml:"database" json:"database"`
	Username string `yaml:"username" json:"username"`
	Password string `yaml:"password" json:"password"`
	SSLMode  string `yaml:"ssl_mode" json:"ssl_mode"`
	
	// 連接池配置
	MaxOpenConns    int           `yaml:"max_open_conns" json:"max_open_conns"`
	MaxIdleConns    int           `yaml:"max_idle_conns" json:"max_idle_conns"`
	ConnMaxLifetime time.Duration `yaml:"conn_max_lifetime" json:"conn_max_lifetime"`
}

// SimilarityConfig 相似性配置
type SimilarityConfig struct {
	Algorithm string  `yaml:"algorithm" json:"algorithm"` // cosine, jaccard, levenshtein
	Threshold float64 `yaml:"threshold" json:"threshold"` // 相似度閾值
	MaxResults int    `yaml:"max_results" json:"max_results"`
}

// CacheConfig 快取配置
type CacheConfig struct {
	Enabled    bool          `yaml:"enabled" json:"enabled"`
	TTL        time.Duration `yaml:"ttl" json:"ttl"`
	MaxEntries int           `yaml:"max_entries" json:"max_entries"`
}

// ProviderType 知識庫提供者類型
type ProviderType string

const (
	ProviderTypeMemory     ProviderType = "memory"
	ProviderTypePostgreSQL ProviderType = "postgresql"
)

// KnowledgeCategory 知識類別
type KnowledgeCategory string

const (
	CategoryPostmortem   KnowledgeCategory = "postmortem"
	CategoryRunbook      KnowledgeCategory = "runbook"
	CategoryTroubleshooting KnowledgeCategory = "troubleshooting"
	CategoryBestPractice KnowledgeCategory = "best_practice"
)

// SeverityLevel 嚴重程度
type SeverityLevel string

const (
	SeverityLow      SeverityLevel = "low"
	SeverityMedium   SeverityLevel = "medium"
	SeverityHigh     SeverityLevel = "high"
	SeverityCritical SeverityLevel = "critical"
)

// StatusType 狀態類型
type StatusType string

const (
	StatusDraft      StatusType = "draft"
	StatusReview     StatusType = "review"
	StatusPublished  StatusType = "published"
	StatusArchived   StatusType = "archived"
)