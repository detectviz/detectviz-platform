package metrics

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/prometheus/client_golang/api"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
	"go.uber.org/zap"
)

// CircuitBreakerConfig 電路斷路器配置
type CircuitBreakerConfig struct {
	Enabled          bool          `yaml:"enabled" json:"enabled"`
	FailureThreshold int64         `yaml:"failure_threshold" json:"failure_threshold"`
	RecoveryTimeout  time.Duration `yaml:"recovery_timeout" json:"recovery_timeout"`
}

// PrometheusConfig Prometheus Provider 配置
type PrometheusConfig struct {
	// Prometheus 服務器地址
	Address string `yaml:"address" json:"address"`

	// 基本認證
	Username string `yaml:"username" json:"username"`
	Password string `yaml:"password" json:"password"`

	// Bearer Token 認證
	BearerToken string `yaml:"bearer_token" json:"bearer_token"`

	// TLS 配置
	TLS struct {
		Enabled            bool   `yaml:"enabled" json:"enabled"`
		InsecureSkipVerify bool   `yaml:"insecure_skip_verify" json:"insecure_skip_verify"`
		CertFile           string `yaml:"cert_file" json:"cert_file"`
		KeyFile            string `yaml:"key_file" json:"key_file"`
		CAFile             string `yaml:"ca_file" json:"ca_file"`
	} `yaml:"tls" json:"tls"`

	// 查詢配置
	Query struct {
		Timeout    time.Duration `yaml:"timeout" json:"timeout"`
		MaxSamples int           `yaml:"max_samples" json:"max_samples"`
	} `yaml:"query" json:"query"`

	// 快取配置
	Cache struct {
		Enabled bool          `yaml:"enabled" json:"enabled"`
		TTL     time.Duration `yaml:"ttl" json:"ttl"`
		MaxSize int           `yaml:"max_size" json:"max_size"`
	} `yaml:"cache" json:"cache"`

	// 並發控制
	Concurrency struct {
		MaxConcurrent int `yaml:"max_concurrent" json:"max_concurrent"`
	} `yaml:"concurrency" json:"concurrency"`

	// 電路斷路器配置
	CircuitBreaker CircuitBreakerConfig `yaml:"circuit_breaker" json:"circuit_breaker"`
}

// CachedResult 快取的查詢結果
type CachedResult struct {
	Result    *QueryResult
	Timestamp time.Time
}

// CircuitBreakerState 表示電路斷路器的狀態
type CircuitBreakerState int

const (
	// StateClosed 關閉狀態，請求正常通過
	StateClosed CircuitBreakerState = iota
	// StateOpen 開啟狀態，請求被阻斷
	StateOpen
	// StateHalfOpen 半開狀態，嘗試性地允許部分請求通過
	StateHalfOpen
)

func (s CircuitBreakerState) String() string {
	switch s {
	case StateClosed:
		return "closed"
	case StateOpen:
		return "open"
	case StateHalfOpen:
		return "half-open"
	default:
		return "unknown"
	}
}

// PrometheusProvider Prometheus 指標提供者
type PrometheusProvider struct {
	client v1.API
	config *PrometheusConfig
	logger *zap.Logger

	// 快取管理
	cache   map[string]*CachedResult
	cacheMu sync.RWMutex

	// 並發控制
	concurrencyLimiter chan struct{}

	// 電路斷路器
	cbState       CircuitBreakerState
	cbFailures    int64
	cbLastFailure time.Time
	cbMu          sync.Mutex
}

// NewPrometheusProvider 創建 Prometheus Provider
func NewPrometheusProvider(config *PrometheusConfig, logger *zap.Logger) (*PrometheusProvider, error) {
	if config == nil {
		return nil, fmt.Errorf("prometheus config is required")
	}

	if logger == nil {
		logger = zap.NewNop()
	}

	// 設置默認值
	if config.Query.Timeout == 0 {
		config.Query.Timeout = 30 * time.Second
	}
	if config.Query.MaxSamples == 0 {
		config.Query.MaxSamples = 10000
	}
	if config.Cache.TTL == 0 {
		config.Cache.TTL = 5 * time.Minute
	}
	if config.Cache.MaxSize == 0 {
		config.Cache.MaxSize = 1000
	}
	if config.Concurrency.MaxConcurrent == 0 {
		config.Concurrency.MaxConcurrent = 10
	}

	// 設置電路斷路器默認值
	if config.CircuitBreaker.Enabled {
		if config.CircuitBreaker.FailureThreshold == 0 {
			config.CircuitBreaker.FailureThreshold = 5
		}
		if config.CircuitBreaker.RecoveryTimeout == 0 {
			config.CircuitBreaker.RecoveryTimeout = 30 * time.Second
		}
	}

	// 創建 HTTP 客戶端
	httpClient := &http.Client{
		Timeout: config.Query.Timeout,
		Transport: &http.Transport{
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 10,
			IdleConnTimeout:     90 * time.Second,
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: config.TLS.InsecureSkipVerify,
			},
		},
	}

	// 創建 Prometheus 客戶端
	clientConfig := api.Config{
		Address: config.Address,
		Client:  httpClient,
	}

	// 設置認證
	if config.BearerToken != "" {
		clientConfig.RoundTripper = &bearerTokenRoundTripper{
			token: config.BearerToken,
			rt:    httpClient.Transport,
		}
	} else if config.Username != "" && config.Password != "" {
		clientConfig.RoundTripper = &basicAuthRoundTripper{
			username: config.Username,
			password: config.Password,
			rt:       httpClient.Transport,
		}
	}

	client, err := api.NewClient(clientConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create prometheus client: %w", err)
	}

	provider := &PrometheusProvider{
		client:             v1.NewAPI(client),
		config:             config,
		logger:             logger,
		cache:              make(map[string]*CachedResult),
		concurrencyLimiter: make(chan struct{}, config.Concurrency.MaxConcurrent),
		cbState:            StateClosed,
	}

	// 啟動快取清理 goroutine
	if config.Cache.Enabled {
		go provider.cleanupCacheRoutine()
	}

	return provider, nil
}

// Query 執行單一指標查詢，包含電路斷路器邏輯
func (p *PrometheusProvider) Query(ctx context.Context, query MetricQuery) (*QueryResult, error) {
	if !p.config.CircuitBreaker.Enabled {
		return p.executeQuery(ctx, query)
	}

	// 檢查電路斷路器狀態
	err := p.checkCircuitBreaker()
	if err != nil {
		return nil, err
	}

	// 執行查詢並更新斷路器狀態
	result, err := p.executeQuery(ctx, query)
	if err != nil {
		p.handleFailure(err)
		return nil, err
	}

	p.handleSuccess()
	return result, nil
}

// executeQuery 執行實際的查詢邏輯
func (p *PrometheusProvider) executeQuery(ctx context.Context, query MetricQuery) (*QueryResult, error) {
	// 檢查快取
	if p.config.Cache.Enabled {
		if cached := p.getCachedResult(query); cached != nil {
			p.logger.Debug("returning cached result", zap.String("metric", query.Metric))
			return cached, nil
		}
	}

	// 獲取並發控制權限
	select {
	case p.concurrencyLimiter <- struct{}{}:
		defer func() { <-p.concurrencyLimiter }()
	case <-ctx.Done():
		return nil, ctx.Err()
	}

	// 構建 PromQL 查詢
	promQL := p.buildPromQL(query)
	if promQL == "" {
		return nil, fmt.Errorf("failed to build valid PromQL query")
	}

	p.logger.Debug("executing prometheus query",
		zap.String("promql", promQL),
		zap.String("metric", query.Metric),
	)

	// 執行查詢
	result, warnings, err := p.client.QueryRange(ctx, promQL, v1.Range{
		Start: query.TimeRange.Start,
		End:   query.TimeRange.End,
		Step:  query.Step,
	})

	if err != nil {
		p.logger.Error("prometheus query failed",
			zap.Error(err),
			zap.String("promql", promQL),
		)
		return nil, fmt.Errorf("prometheus query failed: %w", err)
	}

	// 轉換結果
	queryResult := p.convertResult(query, result, warnings)

	// 快取結果
	if p.config.Cache.Enabled {
		p.cacheResult(query, queryResult)
	}

	return queryResult, nil
}

// checkCircuitBreaker 檢查電路斷路器狀態
func (p *PrometheusProvider) checkCircuitBreaker() error {
	p.cbMu.Lock()
	defer p.cbMu.Unlock()

	switch p.cbState {
	case StateOpen:
		// 如果在開啟狀態，檢查是否可以轉換到半開狀態
		if time.Since(p.cbLastFailure) > p.config.CircuitBreaker.RecoveryTimeout {
			p.logger.Info("recovery timeout elapsed, transitioning to half-open state",
				zap.Duration("recoveryTimeout", p.config.CircuitBreaker.RecoveryTimeout))
			p.cbState = StateHalfOpen
			p.cbFailures = 0 // 重置失敗計數器以進行半開測試
			return nil
		}
		// 否則，保持開啟並返回錯誤
		return fmt.Errorf("circuit breaker is open for metric provider")
	case StateHalfOpen:
		// 在半開狀態下，只允許一個請求通過
		p.logger.Debug("circuit breaker is half-open, allowing one request")
		return nil
	case StateClosed:
		// 關閉狀態，允許請求
		return nil
	}
	return nil
}

// handleFailure 處理失敗的請求
func (p *PrometheusProvider) handleFailure(err error) {
	p.cbMu.Lock()
	defer p.cbMu.Unlock()

	// 忽略 context canceled 錯誤
	if err == context.Canceled || err == context.DeadlineExceeded {
		return
	}

	switch p.cbState {
	case StateClosed:
		p.cbFailures++
		if p.cbFailures >= p.config.CircuitBreaker.FailureThreshold {
			p.logger.Error("failure threshold reached, opening circuit breaker",
				zap.Int64("failures", p.cbFailures),
				zap.Int64("threshold", p.config.CircuitBreaker.FailureThreshold))
			p.cbState = StateOpen
			p.cbLastFailure = time.Now()
		}
	case StateHalfOpen:
		p.logger.Warn("request failed in half-open state, re-opening circuit breaker")
		p.cbState = StateOpen
		p.cbLastFailure = time.Now()
	}
}

// handleSuccess 處理成功的請求
func (p *PrometheusProvider) handleSuccess() {
	p.cbMu.Lock()
	defer p.cbMu.Unlock()

	switch p.cbState {
	case StateHalfOpen:
		p.logger.Info("request succeeded in half-open state, closing circuit breaker")
		p.cbState = StateClosed
		p.cbFailures = 0
	case StateClosed:
		// 如果之前有失敗，重置計數器
		if p.cbFailures > 0 {
			p.cbFailures = 0
		}
	}
}

// BatchQuery 並行執行多個查詢
func (p *PrometheusProvider) BatchQuery(ctx context.Context, queries []MetricQuery) ([]*QueryResult, error) {
	if len(queries) == 0 {
		return nil, nil
	}

	// 檢查 context 取消
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	results := make([]*QueryResult, len(queries))
	errors := make([]error, len(queries))

	// 使用 WaitGroup 進行並發控制
	var wg sync.WaitGroup

	for i, query := range queries {
		wg.Add(1)
		go func(index int, q MetricQuery) {
			defer wg.Done()

			result, err := p.Query(ctx, q)
			results[index] = result
			errors[index] = err
		}(i, query)
	}

	wg.Wait()

	// 檢查是否有錯誤
	for i, err := range errors {
		if err != nil {
			return nil, fmt.Errorf("query %d failed: %w", i, err)
		}
	}

	return results, nil
}

// GetAggregation 執行聚合查詢
func (p *PrometheusProvider) GetAggregation(ctx context.Context, opts AggregationOptions) (*AggregationResult, error) {
	// 為每個指標創建查詢
	var queries []MetricQuery

	for _, metric := range opts.Metrics {
		query := MetricQuery{
			Metric:      metric,
			Labels:      opts.Filters,
			TimeRange:   opts.TimeRange,
			Step:        time.Minute, // 默認步長
			Aggregation: string(opts.Function),
		}
		queries = append(queries, query)
	}

	// 執行批量查詢
	results, err := p.BatchQuery(ctx, queries)
	if err != nil {
		return nil, err
	}

	// 轉換為聚合結果
	aggregationResult := &AggregationResult{
		Options: &opts,
		Values:  make([]AggregatedValue, 0, len(results)),
	}

	for i, result := range results {
		if result != nil && len(result.Series) > 0 {
			for _, series := range result.Series {
				if len(series.Values) > 0 {
					// 計算聚合值
					value := p.calculateAggregation(series.Values, opts.Function)

					aggregationResult.Values = append(aggregationResult.Values, AggregatedValue{
						Metric:      opts.Metrics[i],
						Labels:      series.Labels,
						Value:       value,
						SampleCount: int64(len(series.Values)),
					})
				}
			}
		}
	}

	return aggregationResult, nil
}

// HealthCheck 驗證 Prometheus 連接
func (p *PrometheusProvider) HealthCheck(ctx context.Context) error {
	// 執行簡單的查詢來測試連接
	_, _, err := p.client.Query(ctx, "up", time.Now())
	if err != nil {
		return fmt.Errorf("prometheus health check failed: %w", err)
	}
	return nil
}

// Close 關閉 Provider
func (p *PrometheusProvider) Close() error {
	// 清理快取
	p.cacheMu.Lock()
	p.cache = make(map[string]*CachedResult)
	p.cacheMu.Unlock()

	return nil
}

// buildCacheKey 構建快取鍵名（修復 FIXME.md 中的碰撞問題）
func (p *PrometheusProvider) buildCacheKey(query MetricQuery) string {
	// 按照 FIXME.md 的建議，確保標籤順序一致
	labels := make([]string, 0, len(query.Labels))
	for k, v := range query.Labels {
		labels = append(labels, fmt.Sprintf("%s=%s", k, v))
	}
	sort.Strings(labels) // 確保順序一致

	return fmt.Sprintf("%s_%s_%d_%d_%s_%d",
		query.Metric,
		strings.Join(labels, ","),
		query.TimeRange.Start.Unix(),
		query.TimeRange.End.Unix(),
		query.Aggregation,
		int64(query.Step.Seconds()),
	)
}

// buildPromQL 構建 PromQL 查詢（添加安全驗證）
func (p *PrometheusProvider) buildPromQL(query MetricQuery) string {
	// 驗證指標名稱（防止 PromQL 注入）
	if !isValidMetricName(query.Metric) {
		p.logger.Warn("invalid metric name", zap.String("metric", query.Metric))
		return ""
	}

	// 構建基礎查詢
	promql := query.Metric

	// 添加標籤過濾器
	if len(query.Labels) > 0 {
		var labelFilters []string
		for k, v := range query.Labels {
			// 驗證標籤名稱和值（防止注入攻擊）
			if !isValidLabelName(k) || !isValidLabelValue(v) {
				p.logger.Warn("invalid label", zap.String("label", k), zap.String("value", v))
				continue
			}
			labelFilters = append(labelFilters, fmt.Sprintf(`%s="%s"`, k, v))
		}

		if len(labelFilters) > 0 {
			promql = fmt.Sprintf("%s{%s}", promql, strings.Join(labelFilters, ","))
		}
	}

	// 添加聚合函數
	if query.Aggregation != "" && isValidAggregation(query.Aggregation) {
		promql = fmt.Sprintf("%s(%s)", query.Aggregation, promql)
	}

	return promql
}

// getCachedResult 獲取快取結果
func (p *PrometheusProvider) getCachedResult(query MetricQuery) *QueryResult {
	if !p.config.Cache.Enabled {
		return nil
	}

	key := p.buildCacheKey(query)

	p.cacheMu.RLock()
	cached, exists := p.cache[key]
	p.cacheMu.RUnlock()

	if !exists {
		return nil
	}

	// 檢查是否過期
	if time.Since(cached.Timestamp) > p.config.Cache.TTL {
		p.cacheMu.Lock()
		delete(p.cache, key)
		p.cacheMu.Unlock()
		return nil
	}

	return cached.Result
}

// cacheResult 快取查詢結果
func (p *PrometheusProvider) cacheResult(query MetricQuery, result *QueryResult) {
	if !p.config.Cache.Enabled {
		return
	}

	key := p.buildCacheKey(query)

	p.cacheMu.Lock()
	defer p.cacheMu.Unlock()

	// 檢查快取大小限制
	if len(p.cache) >= p.config.Cache.MaxSize {
		// 清理最舊的條目
		p.cleanupOldestCacheEntries(p.config.Cache.MaxSize / 4) // 清理 25%
	}

	p.cache[key] = &CachedResult{
		Result:    result,
		Timestamp: time.Now(),
	}
}

// cleanupOldestCacheEntries 清理最舊的快取條目
func (p *PrometheusProvider) cleanupOldestCacheEntries(count int) {
	type cacheEntry struct {
		key       string
		timestamp time.Time
	}

	entries := make([]cacheEntry, 0, len(p.cache))
	for k, v := range p.cache {
		entries = append(entries, cacheEntry{key: k, timestamp: v.Timestamp})
	}

	// 按時間排序
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].timestamp.Before(entries[j].timestamp)
	})

	// 刪除最舊的條目
	for i := 0; i < count && i < len(entries); i++ {
		delete(p.cache, entries[i].key)
	}
}

// cleanupCacheRoutine 定期清理過期快取
func (p *PrometheusProvider) cleanupCacheRoutine() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		p.cleanupExpiredCache()
	}
}

// cleanupExpiredCache 清理過期的快取條目
func (p *PrometheusProvider) cleanupExpiredCache() {
	p.cacheMu.Lock()
	defer p.cacheMu.Unlock()

	now := time.Now()
	for key, cached := range p.cache {
		if now.Sub(cached.Timestamp) > p.config.Cache.TTL {
			delete(p.cache, key)
		}
	}
}

// convertResult 轉換 Prometheus 結果為統一格式
func (p *PrometheusProvider) convertResult(query MetricQuery, result model.Value, warnings v1.Warnings) *QueryResult {
	queryResult := &QueryResult{
		Query:    &query,
		Series:   make([]TimeSeries, 0),
		Warnings: make([]string, len(warnings)),
		Stats: &QueryStats{
			Duration: 0, // 這裡可以添加實際的查詢時間統計
		},
	}

	// 轉換警告
	for i, warning := range warnings {
		queryResult.Warnings[i] = string(warning)
	}

	// 轉換時間序列數據
	switch v := result.(type) {
	case model.Matrix:
		for _, sample := range v {
			series := TimeSeries{
				Labels: make(map[string]string),
				Values: make([]DataPoint, len(sample.Values)),
			}

			// 轉換標籤
			for name, value := range sample.Metric {
				series.Labels[string(name)] = string(value)
			}

			// 轉換數據點
			for i, value := range sample.Values {
				series.Values[i] = DataPoint{
					Timestamp: value.Timestamp.Time(),
					Value:     float64(value.Value),
				}
			}

			queryResult.Series = append(queryResult.Series, series)
		}
	}

	return queryResult
}

// calculateAggregation 計算聚合值
func (p *PrometheusProvider) calculateAggregation(values []DataPoint, function AggregationFunc) float64 {
	if len(values) == 0 {
		return 0
	}

	switch function {
	case AggregationAvg:
		sum := 0.0
		for _, v := range values {
			sum += v.Value
		}
		return sum / float64(len(values))

	case AggregationMax:
		max := values[0].Value
		for _, v := range values {
			if v.Value > max {
				max = v.Value
			}
		}
		return max

	case AggregationMin:
		min := values[0].Value
		for _, v := range values {
			if v.Value < min {
				min = v.Value
			}
		}
		return min

	case AggregationSum:
		sum := 0.0
		for _, v := range values {
			sum += v.Value
		}
		return sum

	case AggregationCount:
		return float64(len(values))

	default:
		return values[len(values)-1].Value // 返回最後一個值
	}
}

// 驗證函數（防止 PromQL 注入攻擊）
func isValidMetricName(name string) bool {
	if name == "" {
		return false
	}
	// 檢查是否符合 Prometheus 指標名稱規範
	for _, char := range name {
		if !((char >= 'a' && char <= 'z') ||
			(char >= 'A' && char <= 'Z') ||
			(char >= '0' && char <= '9') ||
			char == '_' || char == ':') {
			return false
		}
	}
	return true
}

func isValidLabelName(name string) bool {
	if name == "" || name[0] == '_' && name[1] == '_' {
		return false // 禁止以 __ 開頭的標籤
	}
	for _, char := range name {
		if !((char >= 'a' && char <= 'z') ||
			(char >= 'A' && char <= 'Z') ||
			(char >= '0' && char <= '9') ||
			char == '_') {
			return false
		}
	}
	return true
}

func isValidLabelValue(value string) bool {
	// 基本檢查，防止特殊字符注入
	forbiddenChars := []string{`"`, `\`, "\n", "\r", "\t"}
	for _, forbidden := range forbiddenChars {
		if strings.Contains(value, forbidden) {
			return false
		}
	}
	return true
}

func isValidAggregation(agg string) bool {
	validAggregations := []string{
		"sum", "avg", "max", "min", "count",
		"rate", "increase", "delta",
		"topk", "bottomk",
		"quantile",
	}

	for _, valid := range validAggregations {
		if agg == valid {
			return true
		}
	}
	return false
}

// HTTP Transport 實作

type bearerTokenRoundTripper struct {
	token string
	rt    http.RoundTripper
}

func (b *bearerTokenRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Set("Authorization", "Bearer "+b.token)
	return b.rt.RoundTrip(req)
}

type basicAuthRoundTripper struct {
	username string
	password string
	rt       http.RoundTripper
}

func (b *basicAuthRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	req.SetBasicAuth(b.username, b.password)
	return b.rt.RoundTrip(req)
}
