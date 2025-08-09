// Package observability provides unified observability initialization for Go platform
// According to spec.md: "統一可觀察性系統（Console、Prometheus、OTLP 導出器）"
// "應用端僅輸出 OTLP 至 Alloy/Collector；憑證與 API Key 僅配置於 Alloy"
package observability

import (
	"context"
	"fmt"
	"net"
	"net/url"
	"os"
	"strings"
	"time"

	"net/http"
	_ "net/http/pprof"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/propagation"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
	"go.opentelemetry.io/otel/trace"

	runtimemetrics "go.opentelemetry.io/contrib/instrumentation/runtime"

	"go.uber.org/zap"
)

// InitFromConfig initializes unified observability according to spec.md requirements
// - Single OTLP exporter endpoint (to Alloy/Collector)
// - JSON structured logs to stdout (collected by Alloy)
// - Resource attributes as required by spec.md
func InitFromConfig(cfg interface{}) (func(), error) {
	// 支援兩種輸入型態：
	// 1) 整體 config（含 observability 節點）
	// 2) 直接傳入 observability 節點
	var root map[string]interface{}
	if m, ok := cfg.(map[string]interface{}); ok {
		root = m
	} else if m, ok := cfg.(map[string]any); ok {
		root = map[string]interface{}(m)
	} else {
		return nil, fmt.Errorf("invalid config type: %T", cfg)
	}

	obsConfig := getMap(root, "observability")
	if len(obsConfig) == 0 {
		// 若呼叫者直接傳入 observability 節點
		obsConfig = root
	}

	mode := getStringValue(obsConfig, "mode", "lgtm_local")

	otlpNode := getMap(obsConfig, "otlp")
	protocol := strings.ToLower(getStringValue(otlpNode, "protocol", "grpc")) // grpc|http
	endpoint := getStringValue(otlpNode, "endpoint", defaultEndpoint(protocol))
	insecure := getBoolValue(otlpNode, "insecure", true)
	headers := getStringMap(otlpNode, "headers")

	// 正規化端點
	var normEndpoint string
	var err error
	if protocol == "grpc" {
		normEndpoint, err = normalizeGRPCEndpoint(endpoint)
	} else {
		normEndpoint, err = normalizeHTTPEndpoint(endpoint)
	}
	if err != nil {
		return nil, fmt.Errorf("invalid OTLP endpoint: %w", err)
	}

	// Optional file logging (for Alloy file tail -> Grafana Cloud Loki)
	// observability.logs.file: { path, max_size_mb, max_backups, max_age_days, compress }
	var logCloser func() error
	if logsNode := getMap(obsConfig, "logs"); len(logsNode) > 0 {
		fileNode := getMap(logsNode, "file")
		if strings.EqualFold(getStringValue(logsNode, "mode", ""), "file") || getStringValue(fileNode, "path", "") != "" {
			path := getStringValue(fileNode, "path", "./var/log/detectviz/detectviz.log")
			if _, closer, err := InitZapLoggerToFile(path); err != nil {
				// zap 還未完成初始化時，使用標準輸出備援
				fmt.Printf("file logger init error: %v\n", err)
			} else {
				logCloser = closer
				zap.L().Info("file logger enabled", zap.String("path", path))
			}
		}
	}

	zap.L().Info("Initializing observability", zap.String("mode", mode), zap.String("protocol", protocol), zap.String("endpoint", normEndpoint), zap.Bool("insecure", insecure))

	// 初始化 Continuous Profiling（pprof 模式）
	var profilerStop func()
	profilerConfig := getMap(obsConfig, "profiling")
	enabledProfiling := getBoolValue(profilerConfig, "enabled", false)
	zap.L().Info("Profiling configuration check", zap.Bool("enabled", enabledProfiling), zap.Any("profiler_config", profilerConfig))

	if enabledProfiling {
		stop, err := initContinuousProfiling(profilerConfig)
		if err != nil {
			// 關鍵服務（已啟用 profiling）初始化失敗 → 停止啟動並回傳錯誤
			if logCloser != nil {
				_ = logCloser()
			}
			return nil, fmt.Errorf("failed to initialize continuous profiling (pprof): %w", err)
		}
		profilerStop = stop
		zap.L().Info("Continuous profiling initialized")
	} else {
		zap.L().Info("Continuous profiling disabled or not configured")
	}

	// 建立 Resource（遵循 spec.md）
	res, err := createResource(mode)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	// 初始化 Tracing/Metrics（依協定切換 HTTP/GRPC）
	var (
		tp       *sdktrace.TracerProvider
		mp       *sdkmetric.MeterProvider
		cleanupT func()
		cleanupM func()
	)

	if protocol == "grpc" {
		tp, cleanupT, err = initTracingGRPC(normEndpoint, insecure, headers, res)
		if err != nil {
			return nil, fmt.Errorf("failed to init tracing(grpc): %w", err)
		}
		mp, cleanupM, err = initMetricsGRPC(normEndpoint, insecure, headers, res)
		if err != nil {
			cleanupT()
			return nil, fmt.Errorf("failed to init metrics(grpc): %w", err)
		}
	} else { // http
		tp, cleanupT, err = initTracingHTTP(normEndpoint, insecure, headers, res)
		if err != nil {
			return nil, fmt.Errorf("failed to init tracing(http): %w", err)
		}
		mp, cleanupM, err = initMetricsHTTP(normEndpoint, insecure, headers, res)
		if err != nil {
			cleanupT()
			return nil, fmt.Errorf("failed to init metrics(http): %w", err)
		}
	}

	otel.SetTracerProvider(tp)
	otel.SetMeterProvider(mp)
	// 啟用 Go runtime 指標（GC、mem、goroutines 等）
	_ = runtimemetrics.Start(runtimemetrics.WithMeterProvider(mp))
	// 設定全域 Propagator（W3C TraceContext + Baggage）
	otel.SetTextMapPropagator(
		propagation.NewCompositeTextMapPropagator(
			propagation.TraceContext{}, propagation.Baggage{},
		),
	)

	// 統一清理
	cleanup := func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := tp.Shutdown(ctx); err != nil {
			zap.L().Warn("trace provider shutdown error", zap.Error(err))
		}
		if err := mp.Shutdown(ctx); err != nil {
			zap.L().Warn("meter provider shutdown error", zap.Error(err))
		}
		if logCloser != nil {
			if err := logCloser(); err != nil {
				zap.L().Warn("file logger close error", zap.Error(err))
			}
		}
		if profilerStop != nil {
			profilerStop()
			zap.L().Info("Continuous profiling stopped")
		}
		cleanupT()
		cleanupM()
	}

	zap.L().Info("Observability initialized successfully")
	return cleanup, nil
}

// --- Resource ---

func createResource(mode string) (*resource.Resource, error) {
	attrs := []attribute.KeyValue{
		semconv.ServiceName("go-platform"),
		semconv.ServiceVersion(getVersion()),
		semconv.DeploymentEnvironment(getEnvironment()),
		semconv.TelemetrySDKLanguageGo,
		attribute.String("platform.component", "toolbridge"),
		attribute.String("observability.mode", mode),
	}
	return resource.NewWithAttributes(semconv.SchemaURL, attrs...), nil
}

// --- Exporters (gRPC) ---

func initTracingGRPC(endpoint string, insecure bool, headers map[string]string, res *resource.Resource) (*sdktrace.TracerProvider, func(), error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	opts := []otlptracegrpc.Option{otlptracegrpc.WithEndpoint(endpoint)}
	if insecure {
		opts = append(opts, otlptracegrpc.WithInsecure())
	}
	if len(headers) > 0 {
		opts = append(opts, otlptracegrpc.WithHeaders(headers))
	}
	exp, err := otlptracegrpc.New(ctx, opts...)
	if err != nil {
		return nil, nil, err
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exp),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
	)
	cleanup := func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = exp.Shutdown(ctx)
	}
	return tp, cleanup, nil
}

func initMetricsGRPC(endpoint string, insecure bool, headers map[string]string, res *resource.Resource) (*sdkmetric.MeterProvider, func(), error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	opts := []otlpmetricgrpc.Option{otlpmetricgrpc.WithEndpoint(endpoint)}
	if insecure {
		opts = append(opts, otlpmetricgrpc.WithInsecure())
	}
	if len(headers) > 0 {
		opts = append(opts, otlpmetricgrpc.WithHeaders(headers))
	}
	exp, err := otlpmetricgrpc.New(ctx, opts...)
	if err != nil {
		return nil, nil, err
	}

	// 使用標準 PeriodicReader
	reader := sdkmetric.NewPeriodicReader(exp,
		sdkmetric.WithInterval(10*time.Second))

	mp := sdkmetric.NewMeterProvider(
		sdkmetric.WithResource(res),
		sdkmetric.WithReader(reader),
	)
	cleanup := func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = exp.Shutdown(ctx)
	}
	return mp, cleanup, nil
}

// --- Exporters (HTTP) ---

func initTracingHTTP(endpoint string, insecure bool, headers map[string]string, res *resource.Resource) (*sdktrace.TracerProvider, func(), error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	opts := []otlptracehttp.Option{otlptracehttp.WithEndpoint(endpoint)}
	if insecure {
		opts = append(opts, otlptracehttp.WithInsecure())
	}
	if len(headers) > 0 {
		opts = append(opts, otlptracehttp.WithHeaders(headers))
	}
	exp, err := otlptracehttp.New(ctx, opts...)
	if err != nil {
		return nil, nil, err
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exp),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
	)
	cleanup := func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = exp.Shutdown(ctx)
	}
	return tp, cleanup, nil
}

func initMetricsHTTP(endpoint string, insecure bool, headers map[string]string, res *resource.Resource) (*sdkmetric.MeterProvider, func(), error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	opts := []otlpmetrichttp.Option{otlpmetrichttp.WithEndpoint(endpoint)}
	if insecure {
		opts = append(opts, otlpmetrichttp.WithInsecure())
	}
	if len(headers) > 0 {
		opts = append(opts, otlpmetrichttp.WithHeaders(headers))
	}
	exp, err := otlpmetrichttp.New(ctx, opts...)
	if err != nil {
		return nil, nil, err
	}

	// 使用標準 PeriodicReader
	reader := sdkmetric.NewPeriodicReader(exp,
		sdkmetric.WithInterval(10*time.Second))

	mp := sdkmetric.NewMeterProvider(
		sdkmetric.WithResource(res),
		sdkmetric.WithReader(reader),
	)
	cleanup := func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = exp.Shutdown(ctx)
	}
	return mp, cleanup, nil
}

// --- Helpers ---

func defaultEndpoint(protocol string) string {
	if protocol == "http" {
		return "http://localhost:4318"
	}
	return "localhost:4317"
}

func normalizeGRPCEndpoint(ep string) (string, error) {
	ep = strings.TrimSpace(ep)
	if strings.HasPrefix(ep, "http://") || strings.HasPrefix(ep, "https://") {
		u, err := url.Parse(ep)
		if err != nil {
			return "", err
		}
		host := u.Host
		if _, _, err := net.SplitHostPort(host); err != nil {
			// 無埠號時，預設 4317
			host = net.JoinHostPort(host, "4317")
		}
		return host, nil
	}
	// 無 scheme，應為 host:port 或 host
	if _, _, err := net.SplitHostPort(ep); err != nil {
		// 無埠號 => 4317
		return net.JoinHostPort(ep, "4317"), nil
	}
	return ep, nil
}

func normalizeHTTPEndpoint(ep string) (string, error) {
	ep = strings.TrimSpace(ep)
	if !strings.HasPrefix(ep, "http://") && !strings.HasPrefix(ep, "https://") {
		// 預設 http
		ep = "http://" + ep
	}
	u, err := url.Parse(ep)
	if err != nil {
		return "", err
	}
	host := u.Host
	if _, _, err := net.SplitHostPort(host); err != nil {
		// 無埠號 => 4318
		u.Host = net.JoinHostPort(host, "4318")
	}
	return u.String(), nil
}

func getMap(m map[string]interface{}, key string) map[string]interface{} {
	if v, ok := m[key]; ok {
		if mm, ok := v.(map[string]interface{}); ok {
			return mm
		}
		if mm, ok := v.(map[string]any); ok {
			return map[string]interface{}(mm)
		}
	}
	return map[string]interface{}{}
}

func getStringValue(m map[string]interface{}, key, def string) string {
	if v, ok := m[key].(string); ok && v != "" {
		return v
	}
	return def
}

func getBoolValue(m map[string]interface{}, key string, def bool) bool {
	if v, ok := m[key].(bool); ok {
		return v
	}
	return def
}

func getIntValue(m map[string]interface{}, key string, def int) int {
	if v, ok := m[key].(float64); ok {
		return int(v)
	}
	if v, ok := m[key].(int); ok {
		return v
	}
	return def
}

func getStringMap(m map[string]interface{}, key string) map[string]string {
	out := map[string]string{}
	if v, ok := m[key].(map[string]interface{}); ok {
		for k, vv := range v {
			if s, ok := vv.(string); ok {
				out[k] = s
			}
		}
	} else if v, ok := m[key].(map[string]any); ok {
		for k, vv := range v {
			if s, ok := vv.(string); ok {
				out[k] = s
			}
		}
	}
	return out
}

func getVersion() string {
	if v := os.Getenv("SERVICE_VERSION"); v != "" {
		return v
	}
	return "dev"
}

func getEnvironment() string {
	if env := os.Getenv("DEPLOYMENT_ENVIRONMENT"); env != "" {
		return env
	}
	if env := os.Getenv("ENV"); env != "" {
		return env
	}
	return "development"
}

// GetTracer returns a tracer for the Go platform with standard naming
func GetTracer(name string) trace.Tracer {
	return otel.Tracer(fmt.Sprintf("go-platform/%s", name))
}

// GetMeter returns a meter for the Go platform with standard naming
func GetMeter(name string) metric.Meter {
	return otel.Meter(fmt.Sprintf("go-platform/%s", name))
}

// initContinuousProfiling initializes continuous profiling.
// 應用端僅支援 pprof（scrape 模式）。Grafana Cloud 的寫入由 Alloy/Pyroscope 完成。
func initContinuousProfiling(config map[string]interface{}) (func(), error) {
	enabled := getBoolValue(config, "enabled", false)
	if !enabled {
		return func() {}, nil
	}
	// 僅支援 pprof；若設定為其他值，一律視為 pprof
	addr := getStringValue(config, "pprof_address", "127.0.0.1:6060")

	srv := &http.Server{Addr: addr}
	go func() {
		zap.L().Info("pprof server starting", zap.String("addr", addr))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			zap.L().Warn("pprof server stopped with error", zap.Error(err))
		} else {
			zap.L().Info("pprof server stopped")
		}
	}()

	stop := func() {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		if err := srv.Shutdown(ctx); err != nil {
			zap.L().Warn("pprof server shutdown error", zap.Error(err))
		}
	}
	zap.L().Info("Continuous profiling enabled (pprof)", zap.String("addr", addr))
	return stop, nil
}
