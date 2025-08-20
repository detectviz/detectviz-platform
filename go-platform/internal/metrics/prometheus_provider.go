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

// PrometheusConfig Prometheus Provider 配置 (對齊 config.yaml.example)
type PrometheusConfig struct {
	URL         string        `yaml:"url" json:"url"`
	Timeout     time.Duration `yaml:"timeout" json:"timeout"`
	BearerToken string        `yaml:"bearer_token" json:"bearer_token"`
	BasicAuth   struct {
		Username string `yaml:"username" json:"username"`
		Password string `yaml:"password" json:"password"`
	} `yaml:"basic_auth" json:"basic_auth"`
	TLS struct {
		Enabled            bool   `yaml:"enabled" json:"enabled"`
		InsecureSkipVerify bool   `yaml:"insecure_skip_verify" json:"insecure_skip_verify"`
		CertFile           string `yaml:"cert_file" json:"cert_file"`
		KeyFile            string `yaml:"key_file" json:"key_file"`
		CAFile             string `yaml:"ca_file" json:"ca_file"`
	} `yaml:"tls" json:"tls"`
	Query struct {
		MaxSamples    int `yaml:"max_samples" json:"max_samples"`
		MaxConcurrent int `yaml:"max_concurrent" json:"max_concurrent"`
	} `yaml:"query" json:"query"`
	Cache struct {
		Enabled bool          `yaml:"enabled" json:"enabled"`
		TTL     time.Duration `yaml:"ttl" json:"ttl"`
		MaxSize int           `yaml:"max_size" json:"max_size"`
	} `yaml:"cache" json:"cache"`
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
	cache      map[string]*CachedResult
	cacheMu    sync.RWMutex
	inflight   map[string]chan struct{} // 用於防止快取失效風暴
	inflightMu sync.Mutex

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
	if config.Timeout == 0 {
		config.Timeout = 30 * time.Second
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
	if config.Query.MaxConcurrent == 0 {
		config.Query.MaxConcurrent = 10
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
		Timeout: config.Timeout,
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
		Address: config.URL,
		Client:  httpClient,
	}

	// 設置認證
	if config.BearerToken != "" {
		clientConfig.RoundTripper = &bearerTokenRoundTripper{
			token: config.BearerToken,
			rt:    httpClient.Transport,
		}
	} else if config.BasicAuth.Username != "" && config.BasicAuth.Password != "" {
		clientConfig.RoundTripper = &basicAuthRoundTripper{
			username: config.BasicAuth.Username,
			password: config.BasicAuth.Password,
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
		inflight:           make(map[string]chan struct{}),
		concurrencyLimiter: make(chan struct{}, config.Query.MaxConcurrent),
		cbState:            StateClosed,
	}

	// 啟動快取清理 goroutine
	if config.Cache.Enabled {
		go provider.cleanupCacheRoutine()
	}

	return provider, nil
}

// Query 執行單一指標查詢，包含電路斷路器和指標記錄邏輯
func (p *PrometheusProvider) Query(ctx context.Context, query MetricQuery) (*QueryResult, error) {
	if !p.config.CircuitBreaker.Enabled {
		result, err := p.executeQuery(ctx, query)
		if err != nil {
			providerQueriesTotal.WithLabelValues("error").Inc()
			return nil, err
		}
		providerQueriesTotal.WithLabelValues("success").Inc()
		return result, nil
	}

	// 檢查電路斷路器狀態
	if err := p.checkCircuitBreaker(); err != nil {
		providerQueriesTotal.WithLabelValues("circuit_breaker_rejected").Inc()
		return nil, err
	}

	// 執行查詢並更新斷路器狀態
	result, err := p.executeQuery(ctx, query)
	if err != nil {
		p.handleFailure(err)
		providerQueriesTotal.WithLabelValues("error").Inc()
		return nil, err
	}

	p.handleSuccess()
	providerQueriesTotal.WithLabelValues("success").Inc()
	return result, nil
}

// executeQuery 執行實際的查詢邏輯，包含 single-flight 機制防止快取失效風暴
func (p *PrometheusProvider) executeQuery(ctx context.Context, query MetricQuery) (*QueryResult, error) {
	if !p.config.Cache.Enabled {
		// 如果快取未啟用，直接執行查詢
		promQL := p.buildPromQL(query)
		result, warnings, err := p.executePrometheusQuery(ctx, promQL, query.TimeRange, query.Step)
		if err != nil {
			return nil, err
		}
		return p.convertResult(query, result, warnings), nil
	}

	// 1. 檢查快取
	key := p.buildCacheKey(query)
	if cached := p.getCachedResult(query); cached != nil {
		providerQueriesTotal.WithLabelValues("cache_hit").Inc()
		providerCacheHitsTotal.Inc()
		p.logger.Debug("returning cached result", zap.String("key", key))
		return cached, nil
	}

	// 2. Single-flight 邏輯
	p.inflightMu.Lock()
	ch, inFlight := p.inflight[key]
	if inFlight {
		// 如果有其他 goroutine 正在查詢此 key，等待結果
		p.inflightMu.Unlock()
		p.logger.Debug("waiting for in-flight query", zap.String("key", key))
		select {
		case <-ch:
			// 查詢已完成，再次從快取中獲取
			p.logger.Debug("in-flight query finished, retrying from cache", zap.String("key", key))
			return p.getCachedResult(query), nil
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}

	// 標記此 key 為正在查詢中
	ch = make(chan struct{})
	p.inflight[key] = ch
	p.inflightMu.Unlock()

	// 確保無論如何都清理 in-flight 標記
	defer func() {
		p.inflightMu.Lock()
		if c, ok := p.inflight[key]; ok {
			close(c)
			delete(p.inflight, key)
		}
		p.inflightMu.Unlock()
	}()

	// 3. 快取未命中，執行查詢
	providerCacheMissesTotal.Inc()
	promQL := p.buildPromQL(query)
	if promQL == "" {
		return nil, fmt.Errorf("failed to build valid PromQL query")
	}

	result, warnings, err := p.executePrometheusQuery(ctx, promQL, query.TimeRange, query.Step)
	if err != nil {
		return nil, err
	}

	// 4. 轉換並快取結果
	queryResult := p.convertResult(query, result, warnings)
	p.cacheResult(query, queryResult)

	return queryResult, nil
}

// checkCircuitBreaker 檢查電路斷路器狀態
func (p *PrometheusProvider) checkCircuitBreaker() error {
	p.cbMu.Lock()
	defer p.cbMu.Unlock()

	switch p.cbState {
	case StateOpen:
		if time.Since(p.cbLastFailure) > p.config.CircuitBreaker.RecoveryTimeout {
			p.logger.Info("recovery timeout elapsed, transitioning to half-open state",
				zap.Duration("recoveryTimeout", p.config.CircuitBreaker.RecoveryTimeout))
			p.cbState = StateHalfOpen
			providerCircuitBreakerState.Set(2) // 2 for half-open
			p.cbFailures = 0
			return nil
		}
		return fmt.Errorf("circuit breaker is open for metric provider")
	case StateHalfOpen:
		p.logger.Debug("circuit breaker is half-open, allowing one request")
		return nil
	case StateClosed:
		return nil
	}
	return nil
}

// handleFailure 處理失敗的請求
func (p *PrometheusProvider) handleFailure(err error) {
	p.cbMu.Lock()
	defer p.cbMu.Unlock()

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
			providerCircuitBreakerState.Set(1) // 1 for open
			p.cbLastFailure = time.Now()
		}
	case StateHalfOpen:
		p.logger.Warn("request failed in half-open state, re-opening circuit breaker")
		p.cbState = StateOpen
		providerCircuitBreakerState.Set(1) // 1 for open
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

// BatchQuery 並行執行多個查詢，並對相似的查詢進行合併優化
func (p *PrometheusProvider) BatchQuery(ctx context.Context, queries []MetricQuery) ([]*QueryResult, error) {
	if len(queries) == 0 {
		return nil, nil
	}

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	// 1. Group queries that can be merged
	type queryGroupKey string
	type originalQuery struct {
		query MetricQuery
		index int
	}

	groups := make(map[queryGroupKey][]*originalQuery)
	for i, q := range queries {
		labels := make([]string, 0, len(q.Labels))
		for k, v := range q.Labels {
			labels = append(labels, fmt.Sprintf("%s=%s", k, v))
		}
		sort.Strings(labels)

		key := fmt.Sprintf("%d-%d-%d-%s-%s",
			q.TimeRange.Start.UnixNano(),
			q.TimeRange.End.UnixNano(),
			q.Step.Nanoseconds(),
			q.Aggregation,
			strings.Join(labels, ","))

		groupKey := queryGroupKey(key)
		groups[groupKey] = append(groups[groupKey], &originalQuery{query: q, index: i})
	}

	results := make([]*QueryResult, len(queries))
	errs := make(chan error, len(groups))
	var wg sync.WaitGroup

	// 2. Execute each group
	for _, group := range groups {
		wg.Add(1)
		go func(g []*originalQuery) {
			defer wg.Done()

			if len(g) == 1 {
				// Execute as a single query
				result, err := p.Query(ctx, g[0].query)
				if err != nil {
					errs <- fmt.Errorf("query for metric %s failed: %w", g[0].query.Metric, err)
					return
				}
				results[g[0].index] = result
				return
			}

			// Execute as a merged query
			var groupQueries []MetricQuery
			for _, oq := range g {
				groupQueries = append(groupQueries, oq.query)
			}

			mergedPromQL := p.buildMergedPromQL(groupQueries)
			if mergedPromQL == "" {
				errs <- fmt.Errorf("failed to build merged promql for group")
				return
			}

			// Execute query directly, bypassing individual query cache/circuit breaker
			mergedResult, warnings, err := p.executePrometheusQuery(ctx, mergedPromQL, groupQueries[0].TimeRange, groupQueries[0].Step)
			if err != nil {
				errs <- fmt.Errorf("merged query failed: %w", err)
				return
			}

			// De-multiplex the result
			demuxResults := p.demultiplexResults(mergedResult, warnings, groupQueries)

			// Assign results back to their original positions
			for _, oq := range g {
				if res, ok := demuxResults[oq.query.Metric]; ok {
					results[oq.index] = res
				} else {
					results[oq.index] = &QueryResult{Query: &oq.query, Series: []TimeSeries{}}
				}
			}
		}(group)
	}

	wg.Wait()
	close(errs)

	// Check for errors
	if len(errs) > 0 {
		var errStrings []string
		for err := range errs {
			errStrings = append(errStrings, err.Error())
		}
		return nil, fmt.Errorf("one or more batch queries failed: %s", strings.Join(errStrings, "; "))
	}

	return results, nil
}

// executePrometheusQuery is a helper to run the actual PromQL query, wrapped with concurrency limiting.
func (p *PrometheusProvider) executePrometheusQuery(ctx context.Context, promQL string, timeRange TimeRange, step time.Duration) (model.Value, v1.Warnings, error) {
	// Get concurrency limiter token
	select {
	case p.concurrencyLimiter <- struct{}{}:
		defer func() { <-p.concurrencyLimiter }()
	case <-ctx.Done():
		return nil, nil, ctx.Err()
	}

	p.logger.Debug("executing prometheus query", zap.String("promql", promQL))

	result, warnings, err := p.client.QueryRange(ctx, promQL, v1.Range{
		Start: timeRange.Start,
		End:   timeRange.End,
		Step:  step,
	})
	if err != nil {
		p.logger.Error("prometheus query failed", zap.Error(err), zap.String("promql", promQL))
		return nil, nil, fmt.Errorf("prometheus query failed: %w", err)
	}
	return result, warnings, nil
}

// buildMergedPromQL builds a PromQL query for a set of mergeable queries.
func (p *PrometheusProvider) buildMergedPromQL(queries []MetricQuery) string {
	if len(queries) == 0 {
		return ""
	}
	sampleQuery := queries[0]

	var metricNames []string
	for _, q := range queries {
		if isValidMetricName(q.Metric) {
			metricNames = append(metricNames, q.Metric)
		}
	}
	if len(metricNames) == 0 {
		return ""
	}
	metricRegex := strings.Join(metricNames, "|")

	// Combine labels from the sample query with the __name__ regex
	var labelFilters []string
	for k, v := range sampleQuery.Labels {
		if isValidLabelName(k) && isValidLabelValue(v) {
			labelFilters = append(labelFilters, fmt.Sprintf(`%s="%s"`, k, v))
		}
	}
	labelFilters = append(labelFilters, fmt.Sprintf(`__name__=~"%s"`, metricRegex))
	sort.Strings(labelFilters)

	promql := fmt.Sprintf("{%s}", strings.Join(labelFilters, ","))

	// Apply aggregation, preserving the metric name for de-multiplexing
	if sampleQuery.Aggregation != "" && isValidAggregation(sampleQuery.Aggregation) {
		promql = fmt.Sprintf("%s by (__name__) (%s)", sampleQuery.Aggregation, promql)
	}

	return promql
}

// demultiplexResults splits a merged query result back into individual results.
func (p *PrometheusProvider) demultiplexResults(mergedResult model.Value, warnings v1.Warnings, originalQueries []MetricQuery) map[string]*QueryResult {
	demuxResults := make(map[string]*QueryResult)

	matrix, ok := mergedResult.(model.Matrix)
	if !ok {
		return demuxResults
	}

	// Pre-create result holders for each original query
	queryMap := make(map[string]MetricQuery)
	for _, q := range originalQueries {
		queryMap[q.Metric] = q
		demuxResults[q.Metric] = &QueryResult{
			Query:    &q,
			Series:   []TimeSeries{},
			Warnings: warnings,
		}
	}

	// Distribute the series from the merged result
	for _, series := range matrix {
		if metricName, ok := series.Metric["__name__"]; ok {
			if _, exists := demuxResults[string(metricName)]; exists {
				convertedSeries := p.convertSeries(series)
				demuxResults[string(metricName)].Series = append(demuxResults[string(metricName)].Series, convertedSeries)
			}
		}
	}

	return demuxResults
}

// convertSeries converts a single Prometheus model.SampleStream to a TimeSeries.
func (p *PrometheusProvider) convertSeries(series *model.SampleStream) TimeSeries {
	ts := TimeSeries{
		Labels: make(map[string]string),
		Values: make([]DataPoint, len(series.Values)),
	}
	for name, value := range series.Metric {
		ts.Labels[string(name)] = string(value)
	}
	for i, value := range series.Values {
		ts.Values[i] = DataPoint{Timestamp: value.Timestamp.Time(), Value: float64(value.Value)}
	}
	return ts
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
