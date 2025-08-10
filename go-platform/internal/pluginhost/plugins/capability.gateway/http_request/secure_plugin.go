package http_request

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	pb "github.com/detectviz/detectviz-platform/contracts/gen/go/detectviz/contracts/v1"
	statuspb "google.golang.org/genproto/googleapis/rpc/status"
	structpb "google.golang.org/protobuf/types/known/structpb"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.uber.org/zap"
)

// SecurePlugin 提供具備安全邊界的 HTTP 請求功能
type SecurePlugin struct {
	client   *http.Client
	security *SecurityConfig
	metrics  *PluginMetrics

	// 併發控制
	semaphore chan struct{}  // 限制並發請求數
	wg        sync.WaitGroup // 追蹤活躍請求

	// 資源監控
	startTime        int64  // 插件啟動時間
	memoryBaseline   uint64 // 記憶體基線
	activeGoroutines int64  // 活躍 Goroutine 數
	maxMemoryBytes   int64  // 記憶體限制
	maxGoroutines    int32  // Goroutine 限制
	maxConnections   int32  // 連接限制
}

// PluginMetrics HTTP 插件的監控指標
type PluginMetrics struct {
	RequestsTotal   int64 // 總請求數
	RequestsBlocked int64 // 被安全策略阻擋的請求數
	RequestsFailed  int64 // 失敗請求數
	ResponseSizeSum int64 // 回應體總大小
}

// NewSecurePlugin 創建具備安全控制的 HTTP 請求插件
func NewSecurePlugin() *SecurePlugin {
	return NewSecurePluginWithConfig(DefaultSecurityConfig())
}

// NewSecurePluginWithConfig 使用自訂安全配置創建插件
func NewSecurePluginWithConfig(config *SecurityConfig) *SecurePlugin {
	// 創建具備安全控制的 HTTP 客戶端
	transport := &http.Transport{
		MaxIdleConns:       100,
		IdleConnTimeout:    90 * time.Second,
		DisableCompression: false,
		DisableKeepAlives:  false,
	}

	client := &http.Client{
		Transport: otelhttp.NewTransport(transport),
		Timeout:   time.Duration(config.MaxTimeoutMs) * time.Millisecond,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= config.MaxRedirects {
				return fmt.Errorf("超過最大重定向次數 %d", config.MaxRedirects)
			}
			// 對重定向的 URL 也進行安全檢查
			return config.validateURL(req.URL.String())
		},
	}

	// 記錄資源基線
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	return &SecurePlugin{
		client:         client,
		security:       config,
		metrics:        &PluginMetrics{},
		semaphore:      make(chan struct{}, 10), // 最多 10 個並發請求
		startTime:      time.Now().UnixMilli(),
		memoryBaseline: memStats.Alloc,
		maxMemoryBytes: 50 * 1024 * 1024, // 預設 50MB 限制
		maxGoroutines:  100,              // 預設 100 個 Goroutine 限制
		maxConnections: 50,               // 預設 50 個連接限制
	}
}

// Close 實作資源釋放
func (p *SecurePlugin) Close() error {
	// 等待所有活躍請求完成
	p.wg.Wait()

	// 關閉 HTTP 連接 - 直接處理可能的 HTTP Transport
	if tr, ok := p.client.Transport.(*http.Transport); ok {
		tr.CloseIdleConnections()
	} else if otelTr, ok := p.client.Transport.(*otelhttp.Transport); ok {
		// 對於 otelhttp.Transport，嘗試關閉底層連接
		// 注意：otelhttp.Transport 可能不直接暴露底層 Transport
		_ = otelTr // 避免未使用變數警告
		zap.L().Debug("無法直接關閉 otelhttp.Transport 的連接")
	}

	zap.L().Info("SecurePlugin 已關閉",
		zap.Int64("total_requests", p.metrics.RequestsTotal),
		zap.Int64("blocked_requests", p.metrics.RequestsBlocked),
		zap.Int64("failed_requests", p.metrics.RequestsFailed),
	)

	return nil
}

// Invoke 執行 HTTP 請求（具備完整安全檢查）
func (p *SecurePlugin) Invoke(ctx context.Context, req *pb.ToolInvokeRequest) (*pb.ToolInvokeReply, error) {
	// 併發控制
	select {
	case p.semaphore <- struct{}{}:
		defer func() { <-p.semaphore }()
	case <-ctx.Done():
		return wrapErr(ctx.Err()), nil
	}

	// 追蹤活躍請求
	p.wg.Add(1)
	defer p.wg.Done()

	p.metrics.RequestsTotal++
	start := time.Now()

	// 解析請求參數
	pl := req.GetPayload().AsMap()

	method := strings.ToUpper(getString(pl, "method", "GET"))
	urlStr := getString(pl, "url", "")
	timeoutMs := getInt(pl, "timeout_ms", int(p.security.MaxTimeoutMs))

	// 準備請求體
	var bodyBytes []byte
	var contentType string
	var err error

	if obj, ok := pl["json"]; ok && obj != nil {
		bodyBytes, err = json.Marshal(obj)
		if err != nil {
			p.metrics.RequestsFailed++
			return wrapErr(fmt.Errorf("JSON 序列化失敗: %w", err)), nil
		}
		contentType = "application/json"
	} else if formMap, ok := getMapStringString(pl, "form"); ok {
		vals := url.Values{}
		for k, v := range formMap {
			vals.Set(k, v)
		}
		bodyBytes = []byte(vals.Encode())
		contentType = "application/x-www-form-urlencoded"
	} else if bodyStr, ok := pl["body"].(string); ok && bodyStr != "" {
		bodyBytes = []byte(bodyStr)
	}

	// 準備請求標頭
	headers := make(map[string]string)
	if hmap, ok := getMapStringString(pl, "headers"); ok {
		headers = hmap
	}

	// **安全驗證** - 這是關鍵的安全檢查點
	if err := p.security.ValidateRequest(method, urlStr, headers, bodyBytes, timeoutMs); err != nil {
		p.metrics.RequestsBlocked++
		zap.L().Warn("HTTP 請求被安全策略阻擋",
			zap.String("url", urlStr),
			zap.String("method", method),
			zap.Error(err),
		)
		return wrapErr(fmt.Errorf("安全檢查失敗: %w", err)), nil
	}

	// 清理和標準化標頭
	headers = p.security.SanitizeHeaders(headers)

	// 處理查詢參數
	if qmap, ok := getMapStringString(pl, "query"); ok && len(qmap) > 0 {
		u, err := url.Parse(urlStr)
		if err != nil {
			p.metrics.RequestsFailed++
			return wrapErr(fmt.Errorf("URL 解析失敗: %w", err)), nil
		}
		q := u.Query()
		for k, v := range qmap {
			q.Set(k, v)
		}
		u.RawQuery = q.Encode()
		urlStr = u.String()
	}

	// 設置超時上下文
	ctxWithTimeout, cancel := context.WithTimeout(ctx, time.Duration(timeoutMs)*time.Millisecond)
	defer cancel()

	// 創建 HTTP 請求
	var body io.Reader
	if len(bodyBytes) > 0 {
		body = bytes.NewReader(bodyBytes)
	}

	httpReq, err := http.NewRequestWithContext(ctxWithTimeout, method, urlStr, body)
	if err != nil {
		p.metrics.RequestsFailed++
		return wrapErr(fmt.Errorf("創建 HTTP 請求失敗: %w", err)), nil
	}

	// 設置標頭
	for k, v := range headers {
		httpReq.Header.Set(k, v)
	}
	if contentType != "" && httpReq.Header.Get("Content-Type") == "" {
		httpReq.Header.Set("Content-Type", contentType)
	}

	// 執行 HTTP 請求
	resp, err := p.client.Do(httpReq)
	if err != nil {
		p.metrics.RequestsFailed++
		return wrapErr(fmt.Errorf("HTTP 請求執行失敗: %w", err)), nil
	}
	defer resp.Body.Close()

	// 檢查回應體大小限制
	if resp.ContentLength > 0 {
		if err := p.security.ValidateResponseSize(resp.ContentLength); err != nil {
			p.metrics.RequestsFailed++
			return wrapErr(err), nil
		}
	}

	// 讀取回應體（帶大小限制）
	limitedReader := io.LimitReader(resp.Body, p.security.MaxResponseSize)
	responseBody, err := io.ReadAll(limitedReader)
	if err != nil {
		p.metrics.RequestsFailed++
		return wrapErr(fmt.Errorf("讀取回應體失敗: %w", err)), nil
	}

	// 更新指標
	p.metrics.ResponseSizeSum += int64(len(responseBody))

	// 處理回應標頭
	responseHeaders := make(map[string]interface{}, len(resp.Header))
	for k, vs := range resp.Header {
		responseHeaders[k] = strings.Join(vs, ",")
	}

	// 判斷請求狀態
	status := &statuspb.Status{Code: 0, Message: "OK"}
	if resp.StatusCode >= 400 {
		status = &statuspb.Status{
			Code:    2, // UNKNOWN
			Message: fmt.Sprintf("HTTP 錯誤: %s", resp.Status),
		}
	}

	// 構造回應結果
	result := map[string]interface{}{
		"status":      resp.StatusCode,
		"status_text": resp.Status,
		"headers":     responseHeaders,
		"body":        string(responseBody),
		"body_size":   len(responseBody),
		"final_url":   resp.Request.URL.String(), // 最終 URL（可能經過重定向）
	}

	resultStruct, _ := structpb.NewStruct(result)

	duration := time.Since(start)

	// 記錄請求日誌
	zap.L().Info("HTTP 請求完成",
		zap.String("method", method),
		zap.String("url", urlStr),
		zap.Int("status", resp.StatusCode),
		zap.Duration("duration", duration),
		zap.Int("response_size", len(responseBody)),
	)

	return &pb.ToolInvokeReply{
		Result: resultStruct,
		Status: status,
		ExecMeta: &pb.ToolExecutionMeta{
			Attempt:    1,
			DurationMs: uint64(duration / time.Millisecond),
			PluginId:   PluginID + ".secure",
			RouteId:    fmt.Sprintf("%s:%s", method, urlStr),
		},
	}, nil
}

// GetMetrics 返回插件的監控指標
func (p *SecurePlugin) GetMetrics() *PluginMetrics {
	return &PluginMetrics{
		RequestsTotal:   p.metrics.RequestsTotal,
		RequestsBlocked: p.metrics.RequestsBlocked,
		RequestsFailed:  p.metrics.RequestsFailed,
		ResponseSizeSum: p.metrics.ResponseSizeSum,
	}
}

// UpdateSecurityConfig 更新安全配置（熱更新）
func (p *SecurePlugin) UpdateSecurityConfig(config *SecurityConfig) {
	p.security = config

	// 更新客戶端超時時間
	p.client.Timeout = time.Duration(config.MaxTimeoutMs) * time.Millisecond

	zap.L().Info("安全配置已更新",
		zap.Int("max_timeout_ms", config.MaxTimeoutMs),
		zap.Int64("max_response_size", config.MaxResponseSize),
		zap.Bool("block_private_ips", config.BlockPrivateIPs),
		zap.Strings("blocked_domains", config.BlockedDomains),
	)
}

// GetResourceUsage 實作 ResourceAwareHandler 接口 - 返回當前資源使用情況
func (p *SecurePlugin) GetResourceUsage() (memoryBytes int64, goroutines int32, connections int32) {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	// 計算相對於基線的記憶體使用增量
	currentMemory := memStats.Alloc
	if currentMemory > p.memoryBaseline {
		memoryBytes = int64(currentMemory - p.memoryBaseline)
	} else {
		memoryBytes = 0
	}

	// 估算當前活躍的 Goroutine 數（簡化實作）
	currentGoroutines := int64(runtime.NumGoroutine())
	goroutines = int32(atomic.LoadInt64(&p.activeGoroutines) + currentGoroutines/10) // 估算

	// 估算連接數（基於 HTTP Transport 的空閒連接數）
	connections = int32(len(p.semaphore)) // 使用 semaphore 作為連接數的近似值

	return memoryBytes, goroutines, connections
}

// SetResourceLimits 實作 ResourceAwareHandler 接口 - 設置資源限制
func (p *SecurePlugin) SetResourceLimits(maxMemoryBytes int64, maxGoroutines int32, maxConnections int32) error {
	// 驗證資源限制的合理性
	if maxMemoryBytes <= 0 || maxGoroutines <= 0 || maxConnections <= 0 {
		return fmt.Errorf("資源限制必須為正數值")
	}

	// 檢查當前資源使用是否超過新限制
	currentMemory, currentGoroutines, currentConnections := p.GetResourceUsage()

	if currentMemory > maxMemoryBytes {
		zap.L().Warn("當前記憶體使用量超過新限制",
			zap.Int64("current_bytes", currentMemory),
			zap.Int64("new_limit_bytes", maxMemoryBytes))
	}

	if currentGoroutines > maxGoroutines {
		zap.L().Warn("當前 Goroutine 數量超過新限制",
			zap.Int32("current_count", currentGoroutines),
			zap.Int32("new_limit", maxGoroutines))
	}

	if currentConnections > maxConnections {
		zap.L().Warn("當前連接數量超過新限制",
			zap.Int32("current_count", currentConnections),
			zap.Int32("new_limit", maxConnections))

		// 調整 semaphore 大小以限制併發連接
		newSemaphore := make(chan struct{}, maxConnections)
		// 將現有的 tokens 轉移到新的 semaphore
	transfer:
		for i := int32(0); i < min(int32(len(p.semaphore)), maxConnections); i++ {
			select {
			case <-p.semaphore:
				newSemaphore <- struct{}{}
			default:
				break transfer
			}
		}
		p.semaphore = newSemaphore
	}

	// 更新限制值
	atomic.StoreInt64(&p.maxMemoryBytes, maxMemoryBytes)
	atomic.StoreInt32(&p.maxGoroutines, maxGoroutines)
	atomic.StoreInt32(&p.maxConnections, maxConnections)

	zap.L().Info("資源限制已更新",
		zap.Int64("max_memory_bytes", maxMemoryBytes),
		zap.Int32("max_goroutines", maxGoroutines),
		zap.Int32("max_connections", maxConnections))

	return nil
}

// CheckResourceLimits 檢查是否超過資源限制
func (p *SecurePlugin) CheckResourceLimits() error {
	memory, goroutines, connections := p.GetResourceUsage()

	if memory > atomic.LoadInt64(&p.maxMemoryBytes) {
		return fmt.Errorf("記憶體使用量 %d bytes 超過限制 %d bytes",
			memory, atomic.LoadInt64(&p.maxMemoryBytes))
	}

	if goroutines > atomic.LoadInt32(&p.maxGoroutines) {
		return fmt.Errorf("Goroutine 數量 %d 超過限制 %d",
			goroutines, atomic.LoadInt32(&p.maxGoroutines))
	}

	if connections > atomic.LoadInt32(&p.maxConnections) {
		return fmt.Errorf("連接數量 %d 超過限制 %d",
			connections, atomic.LoadInt32(&p.maxConnections))
	}

	return nil
}

// GetDetailedResourceMetrics 獲取詳細的資源指標
func (p *SecurePlugin) GetDetailedResourceMetrics() map[string]interface{} {
	memory, goroutines, connections := p.GetResourceUsage()

	return map[string]interface{}{
		"memory_usage": map[string]interface{}{
			"current_bytes":  memory,
			"limit_bytes":    atomic.LoadInt64(&p.maxMemoryBytes),
			"baseline_bytes": p.memoryBaseline,
			"utilization":    float64(memory) / float64(atomic.LoadInt64(&p.maxMemoryBytes)) * 100,
		},
		"goroutines": map[string]interface{}{
			"current_count": goroutines,
			"limit_count":   atomic.LoadInt32(&p.maxGoroutines),
			"utilization":   float64(goroutines) / float64(atomic.LoadInt32(&p.maxGoroutines)) * 100,
		},
		"connections": map[string]interface{}{
			"current_count": connections,
			"limit_count":   atomic.LoadInt32(&p.maxConnections),
			"utilization":   float64(connections) / float64(atomic.LoadInt32(&p.maxConnections)) * 100,
		},
		"requests": map[string]interface{}{
			"total_count":   p.metrics.RequestsTotal,
			"blocked_count": p.metrics.RequestsBlocked,
			"failed_count":  p.metrics.RequestsFailed,
			"success_rate":  float64(p.metrics.RequestsTotal-p.metrics.RequestsFailed) / float64(p.metrics.RequestsTotal) * 100,
		},
		"response_data": map[string]interface{}{
			"total_size_bytes": p.metrics.ResponseSizeSum,
			"avg_size_bytes":   float64(p.metrics.ResponseSizeSum) / float64(p.metrics.RequestsTotal),
		},
		"uptime_ms": time.Now().UnixMilli() - p.startTime,
	}
}

// min 輔助函數
func min(a, b int32) int32 {
	if a < b {
		return a
	}
	return b
}
