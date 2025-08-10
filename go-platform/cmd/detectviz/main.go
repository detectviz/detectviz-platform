package main

import (
	"context"
	"flag"
	"fmt"
	"log"
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
	"github.com/detectviz/detectviz-platform/go-platform/internal/pluginnew"
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

func printUsage() {
	fmt.Println("用法: detectviz <plugin|config> [子命令]")
	fmt.Println()
	fmt.Println("commands:")
	fmt.Println("  plugin  - 插件管理與 ToolBridge 服務 (符合 spec.md 規範)")
	fmt.Println("  config  - 配置管理工具")
	fmt.Println()
	fmt.Println("注意: 已移除舊的 HTTP gateway，統一以 gRPC ToolBridge 對外服務")
}

func handleConfigCommands() {
	if len(os.Args) < 3 {
		fmt.Println("用法: detectviz config validate -f <config.yaml>")
		os.Exit(2)
	}
	switch os.Args[2] {
	case "validate":
		fs := flag.NewFlagSet("validate", flag.ExitOnError)
		f := fs.String("f", "./configs/config.yaml", "設定檔路徑")
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

// handlePluginCommands 處理插件相關命令（包含啟動 gRPC ToolBridge）
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
		if err := pluginnew.Scaffold(os.Args[3]); err != nil {
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

func cmdPluginServe() {
	fs := flag.NewFlagSet("plugin serve", flag.ExitOnError)
	// 支援環境變數覆寫
	defaultListen := getenvDefault("A2A_LISTEN", ":6606")
	defaultCfg := getenvDefault("DETECTVIZ_CONFIG", "./config.yaml") // 修正預設路徑
	defaultHTTPDemo := getenvDefault("DETECTVIZ_HTTP_DEMO", "0") == "1"
	defaultHTTPDemoListen := getenvDefault("DETECTVIZ_HTTP_DEMO_LISTEN", ":7777")

	listen := fs.String("listen", defaultListen, "ToolBridge gRPC 監聽地址")
	cfgPath := fs.String("config", defaultCfg, "平台設定檔 (用於 observability 等)")
	mtlsCert := fs.String("mtls-cert", os.Getenv("A2A_CERT_PATH"), "mTLS 證書路徑")
	mtlsKey := fs.String("mtls-key", os.Getenv("A2A_KEY_PATH"), "mTLS 私鑰路徑")
	mtlsCA := fs.String("mtls-ca", os.Getenv("A2A_CA_PATH"), "mTLS CA 路徑")
	httpDemo := fs.Bool("http-demo", defaultHTTPDemo, "啟動示範 HTTP 服務 (otelhttp instrumentation)")
	httpDemoListen := fs.String("http-demo-listen", defaultHTTPDemoListen, "示範 HTTP 服務監聽地址")
	_ = fs.Parse(os.Args[3:])

	// 啟動 Health HTTP 服務（僅供 K8s/監控用），可用環境變數 DETECTVIZ_HEALTH_ADDR 覆蓋
	healthAddr := getenvDefault("DETECTVIZ_HEALTH_ADDR", ":8081")
	healthSrv := health.NewServer(healthAddr)
	healthSrv.Start()
	defer healthSrv.Stop(context.Background())
	
	// CRITICAL: Validate contract version consistency before proceeding
	if err := contracts.ValidateContractVersion(); err != nil {
		fmt.Printf("Contract version validation failed: %v\n", err)
		fmt.Printf("This prevents startup to ensure SSOT compliance.\n")
		os.Exit(1)
	}
	
	// Additional contract consistency check
	if err := contracts.ValidateContractConsistency(); err != nil {
		fmt.Printf("Contract consistency validation failed: %v\n", err)
		os.Exit(1)
	}
	
	cfg, err := configx.LoadAndValidate(*cfgPath)
	if err != nil {
		fmt.Printf("Failed to load config: %v\n", err)
		os.Exit(1)
	}

	// 如果命令行參數使用默認值且配置檔案有設定，則優先使用配置檔案的值
	finalListen := *listen
	if *listen == defaultListen && cfg.GRPC.Listen != "" {
		finalListen = cfg.GRPC.Listen
		fmt.Printf("Using gRPC listen address from config: %s\n", finalListen)
	}

	// 初始化觀測（Grafana/Cloud/Otel 由 config 決定）
	obsConfig := cfg.GetObservabilityConfig()
	shutdown, err := observability.InitFromConfig(obsConfig)
	if err != nil {
		fmt.Printf("Failed to initialize observability: %v\n", err)
		os.Exit(1)
	}
	defer shutdown()

	tlsCfg, err := pluginhost.LoadTLSConfig(*mtlsCert, *mtlsKey, *mtlsCA)
	if err != nil {
		zap.L().Fatal("載入 mTLS 憑證失敗", zap.Error(err))
	}

	reg := pluginhost.NewRegistry()
	if err := register.RegisterAll(reg); err != nil {
		log.Fatalf("register plugins: %v", err)
	}
	// 範例：在此註冊插件
	// import http_request "github.com/detectviz/detectviz-platform/go-platform/internal/pluginhost/plugins/capability.gateway/http_request"
	// reg.Register("detectviz.tools.http_request", http_request.New())

	fmt.Printf("ToolBridge listening on %s (mTLS=%v)\n", finalListen, tlsCfg != nil)

	rt := pluginhost.NewRuntime(finalListen, tlsCfg, reg)
	rt.SetOnReady(func() { healthSrv.SetReady(true) })

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// 可選：啟動示範 HTTP 服務，使用 otelhttp 自動注入 tracing
	var httpSrv *http.Server
	if *httpDemo {
		mux := setupHTTPDemoHandlers()
		httpSrv = &http.Server{Addr: *httpDemoListen, Handler: mux}
		go func() {
			zap.L().Info("HTTP demo listening", zap.String("listen", *httpDemoListen))
			if err := httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				zap.L().Error("HTTP demo server error", zap.Error(err))
			}
		}()
	}

	// 啟動 ToolBridge gRPC 服務
	go func() {
		if err := rt.Start(ctx); err != nil {
			zap.L().Fatal("ToolBridge 啟動失敗", zap.Error(err))
		}
	}()

	<-ctx.Done()
	if httpSrv != nil {
		_ = httpSrv.Shutdown(context.Background())
	}
	_ = rt.Stop(context.Background())
}

func getenvDefault(key, def string) string {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	return v
}

// setupHTTPDemoHandlers 設置豐富的示範 HTTP endpoints
// 每個 endpoint 產生不同的 traces, metrics 和 logs 以展示 Grafana Drilldown 功能
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
