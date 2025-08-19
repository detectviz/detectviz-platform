package knowledge

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"go.uber.org/zap"
	// PostgreSQL driver - 註解掉以避免編譯錯誤，實際使用時需要添加到 go.mod
	// _ "github.com/lib/pq"
)

// PostgreSQLProvider 實作基於 PostgreSQL 的知識庫提供者
type PostgreSQLProvider struct {
	db     *sql.DB
	logger *zap.Logger
	config *DatabaseConfig
}

// NewPostgreSQLProvider 創建新的 PostgreSQL 知識庫提供者
func NewPostgreSQLProvider(config *DatabaseConfig, logger *zap.Logger) (*PostgreSQLProvider, error) {
	if config == nil {
		return nil, fmt.Errorf("database config is required")
	}

	// 建構連接字串
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		config.Host, config.Port, config.Username, config.Password, config.Database, config.SSLMode)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	// 設置連接池參數
	if config.MaxOpenConns > 0 {
		db.SetMaxOpenConns(config.MaxOpenConns)
	}
	if config.MaxIdleConns > 0 {
		db.SetMaxIdleConns(config.MaxIdleConns)
	}
	if config.ConnMaxLifetime > 0 {
		db.SetConnMaxLifetime(config.ConnMaxLifetime)
	}

	// 測試連接
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	provider := &PostgreSQLProvider{
		db:     db,
		logger: logger,
		config: config,
	}

	// 初始化資料庫 schema
	if err := provider.initSchema(); err != nil {
		return nil, fmt.Errorf("failed to initialize database schema: %w", err)
	}

	return provider, nil
}

// initSchema 初始化資料庫 schema
func (p *PostgreSQLProvider) initSchema() error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 創建知識項目表
	createKnowledgeTable := `
	CREATE TABLE IF NOT EXISTS knowledge_items (
		id VARCHAR(255) PRIMARY KEY,
		title TEXT NOT NULL,
		content TEXT NOT NULL,
		category VARCHAR(100) NOT NULL,
		tags TEXT[], -- PostgreSQL 數組類型
		metadata JSONB,
		created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
		updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
		created_by VARCHAR(255) NOT NULL,
		severity VARCHAR(50),
		status VARCHAR(50),
		
		-- 事後複盤特有欄位
		incident_id VARCHAR(255),
		root_cause TEXT,
		resolution TEXT,
		lessons_learned TEXT[],
		action_items JSONB,
		
		-- 全文檢索索引
		content_vector tsvector GENERATED ALWAYS AS (
			setweight(to_tsvector('english', title), 'A') ||
			setweight(to_tsvector('english', content), 'B') ||
			setweight(to_tsvector('english', COALESCE(root_cause, '')), 'C')
		) STORED
	);`

	if _, err := p.db.ExecContext(ctx, createKnowledgeTable); err != nil {
		return fmt.Errorf("failed to create knowledge_items table: %w", err)
	}

	// 創建索引
	indexes := []string{
		"CREATE INDEX IF NOT EXISTS idx_knowledge_category ON knowledge_items(category);",
		"CREATE INDEX IF NOT EXISTS idx_knowledge_created_at ON knowledge_items(created_at);",
		"CREATE INDEX IF NOT EXISTS idx_knowledge_incident_id ON knowledge_items(incident_id);",
		"CREATE INDEX IF NOT EXISTS idx_knowledge_severity ON knowledge_items(severity);",
		"CREATE INDEX IF NOT EXISTS idx_knowledge_status ON knowledge_items(status);",
		"CREATE INDEX IF NOT EXISTS idx_knowledge_tags ON knowledge_items USING GIN(tags);",
		"CREATE INDEX IF NOT EXISTS idx_knowledge_metadata ON knowledge_items USING GIN(metadata);",
		"CREATE INDEX IF NOT EXISTS idx_knowledge_content_vector ON knowledge_items USING GIN(content_vector);",
	}

	for _, indexSQL := range indexes {
		if _, err := p.db.ExecContext(ctx, indexSQL); err != nil {
			p.logger.Warn("Failed to create index", zap.String("sql", indexSQL), zap.Error(err))
		}
	}

	p.logger.Info("Database schema initialized successfully")
	return nil
}

// Store 儲存知識項目
func (p *PostgreSQLProvider) Store(ctx context.Context, item *KnowledgeItem) error {
	if item == nil {
		return fmt.Errorf("knowledge item cannot be nil")
	}

	if item.ID == "" {
		return fmt.Errorf("knowledge item ID is required")
	}

	// 序列化 metadata 和 action_items
	metadataJSON, err := json.Marshal(item.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	actionItemsJSON, err := json.Marshal(item.ActionItems)
	if err != nil {
		return fmt.Errorf("failed to marshal action items: %w", err)
	}

	// 準備標籤數組
	tags := fmt.Sprintf("{%s}", strings.Join(item.Tags, ","))
	lessonsLearned := fmt.Sprintf("{%s}", strings.Join(item.LessonsLearned, ","))

	// 設置時間戳
	now := time.Now()
	if item.CreatedAt.IsZero() {
		item.CreatedAt = now
	}
	item.UpdatedAt = now

	// 使用 UPSERT 操作
	query := `
	INSERT INTO knowledge_items (
		id, title, content, category, tags, metadata,
		created_at, updated_at, created_by, severity, status,
		incident_id, root_cause, resolution, lessons_learned, action_items
	) VALUES (
		$1, $2, $3, $4, $5, $6,
		$7, $8, $9, $10, $11,
		$12, $13, $14, $15, $16
	)
	ON CONFLICT (id) DO UPDATE SET
		title = EXCLUDED.title,
		content = EXCLUDED.content,
		category = EXCLUDED.category,
		tags = EXCLUDED.tags,
		metadata = EXCLUDED.metadata,
		updated_at = EXCLUDED.updated_at,
		severity = EXCLUDED.severity,
		status = EXCLUDED.status,
		incident_id = EXCLUDED.incident_id,
		root_cause = EXCLUDED.root_cause,
		resolution = EXCLUDED.resolution,
		lessons_learned = EXCLUDED.lessons_learned,
		action_items = EXCLUDED.action_items
	`

	_, err = p.db.ExecContext(ctx, query,
		item.ID, item.Title, item.Content, item.Category, tags, metadataJSON,
		item.CreatedAt, item.UpdatedAt, item.CreatedBy, item.Severity, item.Status,
		item.IncidentID, item.RootCause, item.Resolution, lessonsLearned, actionItemsJSON,
	)

	if err != nil {
		return fmt.Errorf("failed to store knowledge item: %w", err)
	}

	p.logger.Debug("Knowledge item stored successfully", zap.String("id", item.ID))
	return nil
}

// Retrieve 根據 ID 檢索知識項目
func (p *PostgreSQLProvider) Retrieve(ctx context.Context, id string) (*KnowledgeItem, error) {
	if id == "" {
		return nil, fmt.Errorf("knowledge item ID is required")
	}

	query := `
	SELECT id, title, content, category, tags, metadata,
		   created_at, updated_at, created_by, severity, status,
		   incident_id, root_cause, resolution, lessons_learned, action_items
	FROM knowledge_items 
	WHERE id = $1
	`

	row := p.db.QueryRowContext(ctx, query, id)

	item := &KnowledgeItem{}
	var tagsArray, lessonsLearnedArray sql.NullString
	var metadataJSON, actionItemsJSON []byte

	err := row.Scan(
		&item.ID, &item.Title, &item.Content, &item.Category, &tagsArray, &metadataJSON,
		&item.CreatedAt, &item.UpdatedAt, &item.CreatedBy, &item.Severity, &item.Status,
		&item.IncidentID, &item.RootCause, &item.Resolution, &lessonsLearnedArray, &actionItemsJSON,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("knowledge item not found: %s", id)
		}
		return nil, fmt.Errorf("failed to retrieve knowledge item: %w", err)
	}

	// 解析標籤
	if tagsArray.Valid {
		item.Tags = parsePostgreSQLArray(tagsArray.String)
	}

	// 解析經驗教訓
	if lessonsLearnedArray.Valid {
		item.LessonsLearned = parsePostgreSQLArray(lessonsLearnedArray.String)
	}

	// 解析 metadata
	if len(metadataJSON) > 0 {
		if err := json.Unmarshal(metadataJSON, &item.Metadata); err != nil {
			p.logger.Warn("Failed to unmarshal metadata", zap.Error(err))
			item.Metadata = make(map[string]string)
		}
	}

	// 解析 action_items
	if len(actionItemsJSON) > 0 {
		if err := json.Unmarshal(actionItemsJSON, &item.ActionItems); err != nil {
			p.logger.Warn("Failed to unmarshal action items", zap.Error(err))
			item.ActionItems = []ActionItem{}
		}
	}

	p.logger.Debug("Knowledge item retrieved successfully", zap.String("id", id))
	return item, nil
}

// Search 根據查詢條件搜索知識項目
func (p *PostgreSQLProvider) Search(ctx context.Context, query *SearchQuery) (*SearchResult, error) {
	if query == nil {
		return nil, fmt.Errorf("search query is required")
	}

	startTime := time.Now()
	
	// 建構 WHERE 條件
	whereClauses := []string{}
	args := []interface{}{}
	argIndex := 1

	// 全文檢索
	if query.Query != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("content_vector @@ plainto_tsquery('english', $%d)", argIndex))
		args = append(args, query.Query)
		argIndex++
	}

	// 類別篩選
	if query.Category != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("category = $%d", argIndex))
		args = append(args, query.Category)
		argIndex++
	}

	// 標籤篩選
	if len(query.Tags) > 0 {
		whereClauses = append(whereClauses, fmt.Sprintf("tags && $%d", argIndex))
		args = append(args, fmt.Sprintf("{%s}", strings.Join(query.Tags, ",")))
		argIndex++
	}

	// 其他篩選條件
	for key, value := range query.Filters {
		switch key {
		case "severity":
			whereClauses = append(whereClauses, fmt.Sprintf("severity = $%d", argIndex))
			args = append(args, value)
			argIndex++
		case "status":
			whereClauses = append(whereClauses, fmt.Sprintf("status = $%d", argIndex))
			args = append(args, value)
			argIndex++
		case "incident_id":
			whereClauses = append(whereClauses, fmt.Sprintf("incident_id = $%d", argIndex))
			args = append(args, value)
			argIndex++
		}
	}

	// 建構完整查詢
	whereClause := ""
	if len(whereClauses) > 0 {
		whereClause = "WHERE " + strings.Join(whereClauses, " AND ")
	}

	// 排序
	orderBy := "ORDER BY updated_at DESC"
	if query.SortBy != "" {
		direction := "DESC"
		if query.SortOrder == "asc" {
			direction = "ASC"
		}
		orderBy = fmt.Sprintf("ORDER BY %s %s", query.SortBy, direction)
	}

	// 分頁
	limit := 50
	if query.Limit > 0 {
		limit = query.Limit
	}
	offset := 0
	if query.Offset > 0 {
		offset = query.Offset
	}

	// 計算總數
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM knowledge_items %s", whereClause)
	var total int64
	if err := p.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, fmt.Errorf("failed to count knowledge items: %w", err)
	}

	// 執行查詢
	mainQuery := fmt.Sprintf(`
		SELECT id, title, content, category, tags, metadata,
			   created_at, updated_at, created_by, severity, status,
			   incident_id, root_cause, resolution, lessons_learned, action_items
		FROM knowledge_items 
		%s %s 
		LIMIT $%d OFFSET $%d
	`, whereClause, orderBy, argIndex, argIndex+1)

	args = append(args, limit, offset)

	rows, err := p.db.QueryContext(ctx, mainQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to execute search query: %w", err)
	}
	defer rows.Close()

	items := []*KnowledgeItem{}
	for rows.Next() {
		item := &KnowledgeItem{}
		var tagsArray, lessonsLearnedArray sql.NullString
		var metadataJSON, actionItemsJSON []byte

		err := rows.Scan(
			&item.ID, &item.Title, &item.Content, &item.Category, &tagsArray, &metadataJSON,
			&item.CreatedAt, &item.UpdatedAt, &item.CreatedBy, &item.Severity, &item.Status,
			&item.IncidentID, &item.RootCause, &item.Resolution, &lessonsLearnedArray, &actionItemsJSON,
		)

		if err != nil {
			p.logger.Error("Failed to scan knowledge item", zap.Error(err))
			continue
		}

		// 解析數組和 JSON 欄位
		if tagsArray.Valid {
			item.Tags = parsePostgreSQLArray(tagsArray.String)
		}
		if lessonsLearnedArray.Valid {
			item.LessonsLearned = parsePostgreSQLArray(lessonsLearnedArray.String)
		}
		if len(metadataJSON) > 0 {
			json.Unmarshal(metadataJSON, &item.Metadata)
		}
		if len(actionItemsJSON) > 0 {
			json.Unmarshal(actionItemsJSON, &item.ActionItems)
		}

		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating search results: %w", err)
	}

	searchTime := time.Since(startTime)

	result := &SearchResult{
		Items:      items,
		Total:      total,
		Page:       offset/limit + 1,
		PageSize:   limit,
		HasMore:    int64(offset+limit) < total,
		Query:      query.Query,
		SearchTime: searchTime,
	}

	p.logger.Debug("Search completed",
		zap.String("query", query.Query),
		zap.Int("results", len(items)),
		zap.Int64("total", total),
		zap.Duration("duration", searchTime),
	)

	return result, nil
}

// SimilaritySearch 根據內容相似性搜索
func (p *PostgreSQLProvider) SimilaritySearch(ctx context.Context, content string, limit int) (*SearchResult, error) {
	if content == "" {
		return nil, fmt.Errorf("content for similarity search is required")
	}

	if limit <= 0 {
		limit = 10
	}

	startTime := time.Now()

	// 使用 PostgreSQL 的相似性搜索功能
	query := `
	SELECT id, title, content, category, tags, metadata,
		   created_at, updated_at, created_by, severity, status,
		   incident_id, root_cause, resolution, lessons_learned, action_items,
		   ts_rank(content_vector, plainto_tsquery('english', $1)) as rank
	FROM knowledge_items 
	WHERE content_vector @@ plainto_tsquery('english', $1)
	ORDER BY rank DESC, updated_at DESC
	LIMIT $2
	`

	rows, err := p.db.QueryContext(ctx, query, content, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to execute similarity search: %w", err)
	}
	defer rows.Close()

	items := []*KnowledgeItem{}
	for rows.Next() {
		item := &KnowledgeItem{}
		var tagsArray, lessonsLearnedArray sql.NullString
		var metadataJSON, actionItemsJSON []byte
		var rank float64

		err := rows.Scan(
			&item.ID, &item.Title, &item.Content, &item.Category, &tagsArray, &metadataJSON,
			&item.CreatedAt, &item.UpdatedAt, &item.CreatedBy, &item.Severity, &item.Status,
			&item.IncidentID, &item.RootCause, &item.Resolution, &lessonsLearnedArray, &actionItemsJSON,
			&rank,
		)

		if err != nil {
			p.logger.Error("Failed to scan similarity search result", zap.Error(err))
			continue
		}

		// 解析數組和 JSON 欄位
		if tagsArray.Valid {
			item.Tags = parsePostgreSQLArray(tagsArray.String)
		}
		if lessonsLearnedArray.Valid {
			item.LessonsLearned = parsePostgreSQLArray(lessonsLearnedArray.String)
		}
		if len(metadataJSON) > 0 {
			json.Unmarshal(metadataJSON, &item.Metadata)
		}
		if len(actionItemsJSON) > 0 {
			json.Unmarshal(actionItemsJSON, &item.ActionItems)
		}

		// 將相似度評分存儲在 metadata 中
		if item.Metadata == nil {
			item.Metadata = make(map[string]string)
		}
		item.Metadata["similarity_score"] = fmt.Sprintf("%.4f", rank)

		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating similarity search results: %w", err)
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

	p.logger.Debug("Similarity search completed",
		zap.String("content", content[:min(50, len(content))]),
		zap.Int("results", len(items)),
		zap.Duration("duration", searchTime),
	)

	return result, nil
}

// Delete 刪除知識項目
func (p *PostgreSQLProvider) Delete(ctx context.Context, id string) error {
	if id == "" {
		return fmt.Errorf("knowledge item ID is required")
	}

	query := "DELETE FROM knowledge_items WHERE id = $1"
	result, err := p.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete knowledge item: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("knowledge item not found: %s", id)
	}

	p.logger.Debug("Knowledge item deleted successfully", zap.String("id", id))
	return nil
}

// Close 關閉資料庫連接
func (p *PostgreSQLProvider) Close() error {
	if p.db != nil {
		return p.db.Close()
	}
	return nil
}

// parsePostgreSQLArray 解析 PostgreSQL 數組格式
func parsePostgreSQLArray(arrayStr string) []string {
	if arrayStr == "" || arrayStr == "{}" {
		return []string{}
	}

	// 移除大括號
	arrayStr = strings.Trim(arrayStr, "{}")
	if arrayStr == "" {
		return []string{}
	}

	// 分割元素
	elements := strings.Split(arrayStr, ",")
	result := make([]string, len(elements))
	for i, element := range elements {
		result[i] = strings.Trim(element, " ")
	}

	return result
}

// min 輔助函數
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}