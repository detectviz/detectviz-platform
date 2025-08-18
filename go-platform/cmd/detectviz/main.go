package main

// main.go
// Detectviz 平台 CLI 入口
// 支援兩大子命令: plugin 與 config
// - plugin serve : 啟動 ToolBridge gRPC 服務並管理 plugin runtime
// - config validate: 驗證設定檔

import (
	"context"
	"flag"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/detectviz/detectviz-platform/go-platform/internal/configx"
	"github.com/detectviz/detectviz-platform/go-platform/internal/contracts"
	"github.com/detectviz/detectviz-platform/go-platform/internal/health"
	"github.com/detectviz/detectviz-platform/go-platform/internal/observability"
	"github.com/detectviz/detectviz-platform/go-platform/internal/pluginhost"
	register "github.com/detectviz/detectviz-platform/go-platform/internal/pluginhost/plugins/register"
	"github.com/detectviz/detectviz-platform/go-platform/tools"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(2)
	}

	switch os.Args[1] {
	case "plugin":
		handlePluginCommands()
	case "config":
		handleConfigCommands()
	default:
		fmt.Println("未知命令")
		printUsage()
		os.Exit(2)
	}
}

// printUsage 顯示 CLI 使用說明
func printUsage() {
	fmt.Println("用法: detectviz <plugin|config> [子命令]")
	fmt.Println()
	fmt.Println("commands:")
	fmt.Println("  plugin  - 插件管理與 ToolBridge 服務 (符合 spec.md 規範)")
	fmt.Println("  config  - 配置管理工具")
}

// handleConfigCommands 處理 config 子指令, 目前支援 validate
func handleConfigCommands() {
	if len(os.Args) < 3 {
		fmt.Println("用法: detectviz config validate -f <config.yaml>")
		os.Exit(2)
	}
	switch os.Args[2] {
	case "validate":
		fs := flag.NewFlagSet("validate", flag.ExitOnError)
		f := fs.String("f", "./config.yaml", "設定檔路徑")
		_ = fs.Parse(os.Args[3:])
		cfg, err := configx.LoadAndValidate(*f)
		if err != nil {
			fmt.Printf("Configuration validation failed: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Config validation passed: env=%s, mode=%s\n", cfg.Env, cfg.Observability.Mode)
	default:
		fmt.Println("未知子命令")
		os.Exit(2)
	}
}

// handlePluginCommands 處理 plugin 相關子命令 (serve, new, validate, ...)
func handlePluginCommands() {
	if len(os.Args) < 3 {
		fmt.Println("插件管理系統 - 符合 spec.md 規範")
		fmt.Println()
		fmt.Println("用法:")
		fmt.Println("  detectviz plugin <command> [args...]")
		fmt.Println()
		fmt.Println("可用命令:")
		fmt.Println("  serve                        - 啟動 ToolBridge gRPC 服務 (pluginhost)")
		fmt.Println("  new <category>/<name>        - 創建 Go 插件骨架（放至 internal/pluginhost/plugins/）")
		fmt.Println("  validate <path>              - 驗證 module.card.json（委派 contracts/tools）")
		fmt.Println("  list                         - 列出已註冊插件（預留）")
		fmt.Println("  register|unregister|publish  - 預留子命令")
		fmt.Println()
		fmt.Println("注意: Go 平台僅負責通訊/註冊/觀測；AI/業務邏輯於 Python ADK Runtime 執行")
		os.Exit(2)
	}

	command := os.Args[2]
	switch command {
	case "serve":
		cmdPluginServe()
	case "new":
		if len(os.Args) < 4 {
			fmt.Println("用法: detectviz plugin new <category>/<name>")
			os.Exit(2)
		}
		if err := tools.ScaffoldPlugin(os.Args[3]); err != nil {
			fmt.Printf("創建插件骨架失敗: %v\n", err)
			os.Exit(1)
		}
	case "validate":
		// 最小版：僅提示轉交 contracts/tools；避免引入額外依賴
		if len(os.Args) < 4 {
			fmt.Println("用法: detectviz plugin validate <path-to-module.card.json-or-plugin-dir>")
			os.Exit(2)
		}
		path := os.Args[3]
		fmt.Printf("請執行: python3 contracts/tools/validate_module_card.py %s\n", path)
	case "list", "register", "unregister", "publish":
		fmt.Println("此命令為預留，將在後續版本提供。")
	default:
		fmt.Printf("未知插件命令: %s\n", command)
		fmt.Println("執行 'detectviz plugin' 查看可用命令")
		os.Exit(2)
	}
}

// StartupContext 封裝啟動過程中的共享狀態
type StartupContext struct {
	Config    *configx.Config
	HealthSrv *health.Server
	HTTPSrv   *http.Server
	Runtime   *pluginhost.Runtime
	Shutdown  func()
	StartTime time.Time
}

// cmdPluginServe 啟動 plugin serve 模式
// 優化後的啟動流程：
//  1. 解析參數並驗證
//  2. 初始化基礎服務（health、observability）
//  3. 驗證 contracts 一致性
//  4. 載入配置並初始化觀測
//  5. 設置 TLS 和註冊插件
//  6. 啟動服務並優雅關機
func cmdPluginServe() {
	startTime := time.Now()
	ctx := &StartupContext{StartTime: startTime}

	// 解析命令行參數
	flags, err := parseServeFlags()
	if err != nil {
		fmt.Printf("參數解析失敗: %v\n", err)
		os.Exit(1)
	}

	// 設置優雅關機處理
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("服務啟動過程中發生嚴重錯誤: %v\n", r)
			cleanupServices(ctx)
			os.Exit(1)
		}
	}()

	// 執行啟動流程
	if err := executeStartupSequence(ctx, flags); err != nil {
		fmt.Printf("服務啟動失敗: %v\n", err)
		cleanupServices(ctx)
		os.Exit(1)
	}

	// 等待關機信號
	waitForShutdown(ctx)
}

// ServeFlags 封装 serve 命令的所有参数
type ServeFlags struct {
	Listen         string
	ConfigPath     string
	HTTPDemo       bool
	HTTPDemoListen string
}

// parseServeFlags 解析 serve 命令的參數
func parseServeFlags() (*ServeFlags, error) {
	fs := flag.NewFlagSet("plugin serve", flag.ExitOnError)

	// 支援環境變數覆寫
	defaultListen := getenvDefault("DETECTVIZ__GRPC__LISTEN", ":6606")
	defaultCfg := getenvDefault("DETECTVIZ_CONFIG_FILE", "./config.yaml")
	defaultHTTPDemo := getenvDefault("DETECTVIZ_HTTP_DEMO", "0") == "1"
	defaultHTTPDemoListen := getenvDefault("DETECTVIZ_HTTP_DEMO_LISTEN", ":7777")

	flags := &ServeFlags{}
	fs.StringVar(&flags.Listen, "listen", defaultListen, "ToolBridge gRPC 監聽地址")
	fs.StringVar(&flags.ConfigPath, "config", defaultCfg, "平台設定檔 (用於 observability 等)")
	fs.BoolVar(&flags.HTTPDemo, "http-demo", defaultHTTPDemo, "啟動示範 HTTP 服務 (otelhttp instrumentation)")
	fs.StringVar(&flags.HTTPDemoListen, "http-demo-listen", defaultHTTPDemoListen, "示範 HTTP 服務監聽地址")

	if err := fs.Parse(os.Args[3:]); err != nil {
		return nil, fmt.Errorf("解析參數失敗: %w", err)
	}

	return flags, nil
}

// executeStartupSequence 執行完整的啟動序列
func executeStartupSequence(ctx *StartupContext, flags *ServeFlags) error {
	// 1. 啟動健康檢查服務
	if err := initHealthService(ctx); err != nil {
		return fmt.Errorf("健康檢查服務初始化失敗: %w", err)
	}

	// 2. 驗證 contracts 一致性
	if err := validateContracts(); err != nil {
		return fmt.Errorf("合約驗證失敗: %w", err)
	}

	// 3. 載入和驗證配置
	if err := loadConfiguration(ctx, flags.ConfigPath); err != nil {
		return fmt.Errorf("配置載入失敗: %w", err)
	}

	// 4. 初始化可觀測性
	if err := initObservability(ctx); err != nil {
		return fmt.Errorf("可觀測性初始化失敗: %w", err)
	}

	// 5. 設置 TLS 和註冊插件
	if err := setupPluginSystem(ctx, flags); err != nil {
		return fmt.Errorf("插件系統設置失敗: %w", err)
	}

	// 6. 啟動 HTTP Demo 服務（如果啟用）
	if flags.HTTPDemo {
		if err := startHTTPDemo(ctx, flags.HTTPDemoListen); err != nil {
			return fmt.Errorf("HTTP Demo 服務啟動失敗: %w", err)
		}
	}

	// 7. 啟動主要的 gRPC ToolBridge 服務
	if err := startToolBridge(ctx, flags.Listen); err != nil {
		return fmt.Errorf("ToolBridge 服務啟動失敗: %w", err)
	}

	// 記錄啟動成功
	startupDuration := time.Since(ctx.StartTime)
	zap.L().Info("服務啟動完成",
		zap.Duration("startup_duration", startupDuration),
		zap.String("grpc_listen", flags.Listen),
		zap.Bool("http_demo_enabled", flags.HTTPDemo),
		zap.Bool("mtls_enabled", ctx.Config.GRPC.TLS.Enabled),
	)

	return nil
}

// initHealthService 初始化健康檢查服務
func initHealthService(ctx *StartupContext) error {
	healthAddr := getenvDefault("DETECTVIZ_HEALTH_ADDR", ":8081")
	ctx.HealthSrv = health.NewServer(healthAddr)
	ctx.HealthSrv.Start()
	return nil
}

// validateContracts 驗證合約一致性
func validateContracts() error {
	if err := contracts.ValidateContractVersion(); err != nil {
		return fmt.Errorf("合約版本驗證失敗: %w (這會阻止啟動以確保 SSOT 合規性)", err)
	}

	if err := contracts.ValidateContractConsistency(); err != nil {
		return fmt.Errorf("合約一致性驗證失敗: %w", err)
	}

	return nil
}

// loadConfiguration 載入和驗證配置
func loadConfiguration(ctx *StartupContext, configPath string) error {
	cfg, err := configx.LoadAndValidate(configPath)
	if err != nil {
		return fmt.Errorf("配置載入失敗 (路徑: %s): %w", configPath, err)
	}
	ctx.Config = cfg
	return nil
}

// initObservability 初始化可觀測性系統
func initObservability(ctx *StartupContext) error {
	obsConfig := ctx.Config.GetObservabilityConfig()
	shutdown, err := observability.InitFromConfig(obsConfig)
	if err != nil {
		return fmt.Errorf("可觀測性初始化失敗: %w", err)
	}
	ctx.Shutdown = shutdown
	return nil
}

// setupPluginSystem 設置插件系統
func setupPluginSystem(ctx *StartupContext, flags *ServeFlags) error {
	// 載入 TLS 配置
	// loader 已根據設定中的路徑將憑證檔案讀入記憶體。
	// 此處傳遞的是憑證的位元組內容，而非路徑。
	tlsCfg, err := pluginhost.LoadTLSConfig(
		ctx.Config.GRPC.TLS.CertData,
		ctx.Config.GRPC.TLS.KeyData,
		ctx.Config.GRPC.TLS.CAData,
	)
	if err != nil {
		return fmt.Errorf("載入 mTLS 憑證失敗: %w", err)
	}

	// 創建和註冊插件
	reg := pluginhost.NewRegistry()
	if err := register.RegisterAll(reg); err != nil {
		return fmt.Errorf("插件註冊失敗: %w", err)
	}

	// 決定最終的監聽地址（配置檔案優先於命令行預設值）
	finalListen := flags.Listen
	if flags.Listen == getenvDefault("DETECTVIZ__GRPC__LISTEN", ":6606") && ctx.Config.GRPC.Listen != "" {
		finalListen = ctx.Config.GRPC.Listen
		zap.L().Info("使用配置檔案中的 gRPC 監聽地址", zap.String("address", finalListen))
	}

	// 創建 Runtime
	ctx.Runtime = pluginhost.NewRuntime(finalListen, tlsCfg, reg)
	ctx.Runtime.SetOnReady(func() {
		ctx.HealthSrv.SetReady(true)
		zap.L().Info("ToolBridge 已就緒", zap.String("listen", finalListen))
	})

	return nil
}

// startHTTPDemo 啟動 HTTP Demo 服務
func startHTTPDemo(ctx *StartupContext, listen string) error {
	mux := setupHTTPDemoHandlers()
	ctx.HTTPSrv = &http.Server{Addr: listen, Handler: mux}

	go func() {
		zap.L().Info("HTTP demo 服務啟動", zap.String("listen", listen))
		if err := ctx.HTTPSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			zap.L().Error("HTTP demo 服務錯誤", zap.Error(err))
		}
	}()

	return nil
}

// startToolBridge 啟動 ToolBridge gRPC 服務
func startToolBridge(ctx *StartupContext, listen string) error {
	go func() {
		if err := ctx.Runtime.Start(context.Background()); err != nil {
			zap.L().Fatal("ToolBridge 啟動失敗", zap.Error(err))
		}
	}()

	return nil
}

// waitForShutdown 等待關機信號並執行優雅關機
func waitForShutdown(ctx *StartupContext) {
	sigCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	<-sigCtx.Done()
	zap.L().Info("收到關機信號，開始優雅關機...")

	cleanupServices(ctx)
}

// cleanupServices 清理所有服務
func cleanupServices(ctx *StartupContext) {
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 1. 標記服務不健康
	if ctx.HealthSrv != nil {
		ctx.HealthSrv.SetReady(false)
	}

	// 2. 關閉 HTTP Demo 服務
	if ctx.HTTPSrv != nil {
		zap.L().Info("關閉 HTTP Demo 服務")
		if err := ctx.HTTPSrv.Shutdown(shutdownCtx); err != nil {
			zap.L().Warn("HTTP Demo 服務關閉失敗", zap.Error(err))
		}
	}

	// 3. 關閉 ToolBridge
	if ctx.Runtime != nil {
		zap.L().Info("關閉 ToolBridge 服務")
		if err := ctx.Runtime.Stop(shutdownCtx); err != nil {
			zap.L().Warn("ToolBridge 關閉失敗", zap.Error(err))
		}
	}

	// 4. 關閉可觀測性系統
	if ctx.Shutdown != nil {
		zap.L().Info("關閉可觀測性系統")
		ctx.Shutdown()
	}

	// 5. 關閉健康檢查服務
	if ctx.HealthSrv != nil {
		zap.L().Info("關閉健康檢查服務")
		ctx.HealthSrv.Stop(shutdownCtx)
	}

	shutdownDuration := time.Since(ctx.StartTime)
	zap.L().Info("服務已完全關閉",
		zap.Duration("total_uptime", shutdownDuration))
}

// getenvDefault 從環境變數讀取值，若不存在則回傳預設值
func getenvDefault(key, def string) string {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	return v
}

// setupHTTPDemoHandlers 設置豐富的示範 HTTP endpoints
// 每個 endpoint 產生不同的 traces, metrics 和 logs 以展示 Grafana Drilldown 功能

// setupHTTPDemoHandlers 建立示範 HTTP handler，展示 tracing/metrics/logs
func setupHTTPDemoHandlers() *http.ServeMux {
	mux := http.NewServeMux()

	// 初始化 metrics
	meter := observability.GetMeter("http-demo")
	requestCounter, _ := meter.Int64Counter(
		"http_requests_total",
		metric.WithDescription("Total number of HTTP requests"),
	)
	requestDuration, _ := meter.Float64Histogram(
		"http_request_duration_seconds",
		metric.WithDescription("HTTP request duration in seconds"),
	)
	activeConnections, _ := meter.Int64UpDownCounter(
		"http_active_connections",
		metric.WithDescription("Active HTTP connections"),
	)

	// Handler wrapper for metrics and enhanced tracing
	instrumentHandler := func(name, operation string, handler http.HandlerFunc) http.Handler {
		return otelhttp.NewHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// 增加活躍連接數
			activeConnections.Add(r.Context(), 1, metric.WithAttributes(
				attribute.String("method", r.Method),
				attribute.String("endpoint", name),
			))
			defer activeConnections.Add(r.Context(), -1, metric.WithAttributes(
				attribute.String("method", r.Method),
				attribute.String("endpoint", name),
			))

			// 獲取 tracer 並增加自訂屬性
			tracer := observability.GetTracer("http-demo")
			ctx, span := tracer.Start(r.Context(), operation,
				trace.WithAttributes(
					attribute.String("http.endpoint", name),
					attribute.String("http.method", r.Method),
					attribute.String("http.user_agent", r.UserAgent()),
					attribute.String("component", "go-platform"),
					attribute.String("service.component", "http-demo"),
				))
			defer span.End()

			// 更新 request context
			r = r.WithContext(ctx)

			// 執行實際 handler
			handler(w, r)

			// 記錄 metrics
			duration := time.Since(start).Seconds()
			labels := metric.WithAttributes(
				attribute.String("method", r.Method),
				attribute.String("endpoint", name),
				attribute.String("status", "200"), // 簡化，實際可以從 response 取得
			)

			requestCounter.Add(ctx, 1, labels)
			requestDuration.Record(ctx, duration, labels)

			// 在 span 中加入執行時間
			span.SetAttributes(
				attribute.Float64("http.duration_seconds", duration),
				attribute.Int("http.status_code", 200),
			)

			// 結構化日誌 (會被 Alloy 收集)
			zap.L().Info("HTTP request processed",
				zap.String("method", r.Method),
				zap.String("endpoint", name),
				zap.String("operation", operation),
				zap.Float64("duration_seconds", duration),
				zap.String("user_agent", r.UserAgent()),
				zap.String("trace_id", span.SpanContext().TraceID().String()),
				zap.String("span_id", span.SpanContext().SpanID().String()),
			)
		}), name)
	}

	// 1. 基本 Hello endpoint
	mux.Handle("/hello", instrumentHandler("hello", "greet_user", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "Hello, world! This is Detectviz Platform HTTP Demo\n")
	}))

	// 2. 健康檢查 endpoint
	mux.Handle("/health", instrumentHandler("health", "health_check", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"status":"healthy","service":"go-platform","component":"http-demo"}`)
	}))

	// 3. 模擬業務邏輯 endpoint (包含子 spans)
	mux.Handle("/business", instrumentHandler("business", "process_business_logic", func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		tracer := observability.GetTracer("http-demo")

		// 創建子 span - 資料庫查詢模擬
		_, dbSpan := tracer.Start(ctx, "database.query",
			trace.WithAttributes(
				attribute.String("db.system", "postgresql"),
				attribute.String("db.operation", "SELECT"),
				attribute.String("db.table", "users"),
			))

		// 模擬資料庫延遲
		time.Sleep(time.Duration(rand.Intn(100)) * time.Millisecond)
		dbSpan.End()

		// 創建另一個子 span - 外部 API 調用模擬
		_, apiSpan := tracer.Start(ctx, "external.api_call",
			trace.WithAttributes(
				attribute.String("http.url", "https://api.example.com/data"),
				attribute.String("http.method", "GET"),
			))

		// 模擬 API 延遲
		time.Sleep(time.Duration(rand.Intn(200)) * time.Millisecond)
		apiSpan.End()

		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"result":"success","processed_items":42,"total_time_ms":150}`)

		zap.L().Info("Business logic processed",
			zap.Int("processed_items", 42),
			zap.String("operation", "complex_business_flow"),
		)
	}))

	// 4. 錯誤演示 endpoint
	mux.Handle("/error", instrumentHandler("error", "simulate_error", func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		span := trace.SpanFromContext(ctx)

		// 記錄錯誤到 span
		span.SetStatus(codes.Error, "Simulated error for demo")
		span.SetAttributes(
			attribute.String("error.type", "simulation"),
			attribute.String("error.message", "This is a demo error"),
		)

		// 錯誤日誌
		zap.L().Error("Simulated error occurred",
			zap.String("error_type", "demo_error"),
			zap.String("endpoint", "/error"),
			zap.String("trace_id", span.SpanContext().TraceID().String()),
		)

		http.Error(w, "This is a simulated error for observability demo", http.StatusInternalServerError)
	}))

	// 5. 參數化 endpoint 展示不同 attributes
	mux.Handle("/user/", instrumentHandler("user_profile", "get_user_profile", func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		span := trace.SpanFromContext(ctx)

		// 從 URL 提取用戶 ID
		userID := r.URL.Path[len("/user/"):]
		if userID == "" {
			userID = "anonymous"
		}

		span.SetAttributes(
			attribute.String("user.id", userID),
			attribute.String("user.type", "standard"),
		)

		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"user_id":"%s","name":"Demo User","profile":"active"}`, userID)

		zap.L().Info("User profile accessed",
			zap.String("user_id", userID),
			zap.String("action", "profile_view"),
		)
	}))

	// 6. 負載測試 endpoint (產生大量 metrics)
	mux.Handle("/load", instrumentHandler("load_test", "handle_load", func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		// 從查詢參數讀取負載數量
		countParam := r.URL.Query().Get("count")
		count, _ := strconv.Atoi(countParam)
		if count <= 0 || count > 1000 {
			count = 100
		}

		span := trace.SpanFromContext(ctx)
		span.SetAttributes(
			attribute.Int("load.count", count),
			attribute.String("load.type", "simulation"),
		)

		// 模擬處理多個項目
		for i := 0; i < count; i++ {
			// 輕量級模擬工作
			_ = fmt.Sprintf("processing item %d", i)
			if i%50 == 0 {
				// 每 50 個項目記錄一次進度
				zap.L().Debug("Load processing progress",
					zap.Int("current", i),
					zap.Int("total", count),
				)
			}
		}

		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"processed":%d,"status":"completed"}`, count)

		zap.L().Info("Load test completed",
			zap.Int("items_processed", count),
			zap.String("test_type", "synthetic_load"),
		)
	}))

	// 7. 主頁面，提供所有可用 endpoints 的說明
	mux.Handle("/", instrumentHandler("index", "show_endpoints", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}

		w.Header().Set("Content-Type", "text/html")
		fmt.Fprint(w, `<!DOCTYPE html>
<html>
<head>
    <title>Detectviz Platform HTTP Demo</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 40px; }
        .endpoint { margin: 20px 0; padding: 10px; border-left: 3px solid #007cba; }
        code { background: #f4f4f4; padding: 2px 4px; }
    </style>
</head>
<body>
    <h1>Detectviz Platform HTTP Demo</h1>
    <p>此 demo 服務展示完整的 OpenTelemetry 整合，包含 metrics, traces, 和 logs，支援 Grafana Cloud Drilldown 功能。</p>
    
    <h2>可用的 Endpoints:</h2>
    
    <div class="endpoint">
        <h3><code>GET /hello</code></h3>
        <p>基本問候 endpoint，展示簡單的 tracing 和 metrics。</p>
    </div>
    
    <div class="endpoint">
        <h3><code>GET /health</code></h3>
        <p>健康檢查 endpoint，返回服務狀態。</p>
    </div>
    
    <div class="endpoint">
        <h3><code>GET /business</code></h3>
        <p>模擬複雜業務邏輯，包含資料庫和 API 調用的子 spans。</p>
    </div>
    
    <div class="endpoint">
        <h3><code>GET /error</code></h3>
        <p>模擬錯誤情況，展示錯誤追蹤和日誌。</p>
    </div>
    
    <div class="endpoint">
        <h3><code>GET /user/{id}</code></h3>
        <p>參數化 endpoint，展示不同的 trace attributes。<br>
        例如: <code>/user/123</code>, <code>/user/alice</code></p>
    </div>
    
    <div class="endpoint">
        <h3><code>GET /load?count=N</code></h3>
        <p>負載測試 endpoint，產生大量 metrics 和 logs。<br>
        例如: <code>/load?count=500</code></p>
    </div>
    
    <h2>Grafana Cloud Drilldown 測試:</h2>
    <ol>
        <li>訪問各個 endpoints 產生遙測數據</li>
        <li>在 Grafana Cloud 的 Metrics 頁面查看 <code>http_requests_total</code> 和 <code>http_request_duration_seconds</code></li>
        <li>使用 Drilldown 功能從 metrics 跳轉到對應的 traces</li>
        <li>在 traces 中查看關聯的 logs</li>
        <li>通過 trace_id 和 span_id 實現跨 observability signal 的關聯</li>
    </ol>
</body>
</html>`)
	}))

	return mux
}
