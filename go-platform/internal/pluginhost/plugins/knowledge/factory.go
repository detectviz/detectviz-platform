package knowledge

import (
	"fmt"

	"go.uber.org/zap"
)

// ProviderFactory 知識庫提供者工廠
type ProviderFactory struct {
	logger *zap.Logger
}

// NewProviderFactory 創建新的提供者工廠
func NewProviderFactory(logger *zap.Logger) *ProviderFactory {
	return &ProviderFactory{
		logger: logger,
	}
}

// CreateProvider 根據配置創建知識庫提供者
func (f *ProviderFactory) CreateProvider(config *Config) (Provider, error) {
	if config == nil {
		return nil, fmt.Errorf("knowledge provider config is required")
	}

	switch ProviderType(config.Provider) {
	case ProviderTypeMemory:
		f.logger.Info("Creating memory knowledge provider")
		return NewMemoryProvider(f.logger), nil

	case ProviderTypePostgreSQL:
		if config.Database == nil {
			return nil, fmt.Errorf("database config is required for PostgreSQL provider")
		}
		f.logger.Info("Creating PostgreSQL knowledge provider", 
			zap.String("host", config.Database.Host),
			zap.Int("port", config.Database.Port),
			zap.String("database", config.Database.Database),
		)
		return NewPostgreSQLProvider(config.Database, f.logger)

	default:
		return nil, fmt.Errorf("unsupported knowledge provider type: %s", config.Provider)
	}
}

// CreateDefaultConfig 創建默認配置
func (f *ProviderFactory) CreateDefaultConfig() *Config {
	return &Config{
		Provider: string(ProviderTypeMemory),
		Similarity: &SimilarityConfig{
			Algorithm:  "cosine",
			Threshold:  0.3,
			MaxResults: 10,
		},
		Cache: &CacheConfig{
			Enabled:    true,
			TTL:        300000000000, // 5 分鐘 (nanoseconds)
			MaxEntries: 1000,
		},
	}
}

// CreatePostgreSQLConfig 創建 PostgreSQL 配置
func (f *ProviderFactory) CreatePostgreSQLConfig(host string, port int, database, username, password string) *Config {
	return &Config{
		Provider: string(ProviderTypePostgreSQL),
		Database: &DatabaseConfig{
			Host:            host,
			Port:            port,
			Database:        database,
			Username:        username,
			Password:        password,
			SSLMode:         "prefer",
			MaxOpenConns:    25,
			MaxIdleConns:    5,
			ConnMaxLifetime: 300000000000, // 5 分鐘 (nanoseconds)
		},
		Similarity: &SimilarityConfig{
			Algorithm:  "postgresql_ts_rank",
			Threshold:  0.1,
			MaxResults: 20,
		},
		Cache: &CacheConfig{
			Enabled:    true,
			TTL:        600000000000, // 10 分鐘 (nanoseconds)
			MaxEntries: 5000,
		},
	}
}

// ValidateConfig 驗證配置
func (f *ProviderFactory) ValidateConfig(config *Config) error {
	if config == nil {
		return fmt.Errorf("config cannot be nil")
	}

	if config.Provider == "" {
		return fmt.Errorf("provider type is required")
	}

	providerType := ProviderType(config.Provider)
	switch providerType {
	case ProviderTypeMemory:
		// Memory provider 不需要額外配置
		
	case ProviderTypePostgreSQL:
		if config.Database == nil {
			return fmt.Errorf("database config is required for PostgreSQL provider")
		}
		if err := f.validateDatabaseConfig(config.Database); err != nil {
			return fmt.Errorf("invalid database config: %w", err)
		}
		
	default:
		return fmt.Errorf("unsupported provider type: %s", config.Provider)
	}

	// 驗證相似性配置
	if config.Similarity != nil {
		if err := f.validateSimilarityConfig(config.Similarity); err != nil {
			return fmt.Errorf("invalid similarity config: %w", err)
		}
	}

	// 驗證快取配置
	if config.Cache != nil {
		if err := f.validateCacheConfig(config.Cache); err != nil {
			return fmt.Errorf("invalid cache config: %w", err)
		}
	}

	return nil
}

// validateDatabaseConfig 驗證資料庫配置
func (f *ProviderFactory) validateDatabaseConfig(config *DatabaseConfig) error {
	if config.Host == "" {
		return fmt.Errorf("database host is required")
	}
	if config.Port <= 0 || config.Port > 65535 {
		return fmt.Errorf("invalid database port: %d", config.Port)
	}
	if config.Database == "" {
		return fmt.Errorf("database name is required")
	}
	if config.Username == "" {
		return fmt.Errorf("database username is required")
	}
	if config.Password == "" {
		return fmt.Errorf("database password is required")
	}

	// 驗證連接池設置
	if config.MaxOpenConns < 0 {
		return fmt.Errorf("max_open_conns cannot be negative")
	}
	if config.MaxIdleConns < 0 {
		return fmt.Errorf("max_idle_conns cannot be negative")
	}
	if config.MaxIdleConns > config.MaxOpenConns && config.MaxOpenConns > 0 {
		return fmt.Errorf("max_idle_conns cannot be greater than max_open_conns")
	}

	return nil
}

// validateSimilarityConfig 驗證相似性配置
func (f *ProviderFactory) validateSimilarityConfig(config *SimilarityConfig) error {
	validAlgorithms := []string{"cosine", "jaccard", "levenshtein", "postgresql_ts_rank"}
	isValid := false
	for _, alg := range validAlgorithms {
		if config.Algorithm == alg {
			isValid = true
			break
		}
	}
	if !isValid {
		return fmt.Errorf("unsupported similarity algorithm: %s, supported: %v", config.Algorithm, validAlgorithms)
	}

	if config.Threshold < 0 || config.Threshold > 1 {
		return fmt.Errorf("similarity threshold must be between 0 and 1, got: %f", config.Threshold)
	}

	if config.MaxResults <= 0 {
		return fmt.Errorf("max_results must be positive, got: %d", config.MaxResults)
	}

	return nil
}

// validateCacheConfig 驗證快取配置
func (f *ProviderFactory) validateCacheConfig(config *CacheConfig) error {
	if config.TTL <= 0 {
		return fmt.Errorf("cache TTL must be positive")
	}

	if config.MaxEntries <= 0 {
		return fmt.Errorf("cache max_entries must be positive, got: %d", config.MaxEntries)
	}

	return nil
}