package telemetry

import (
	"context"
	"fmt"

	"detectviz-platform/pkg/platform/contracts"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	"go.opentelemetry.io/otel/trace"
)

// JaegerTracingProvider 實現 TracingProvider 介面，提供 OTLP 分散式追蹤功能
// 職責: 初始化 OpenTelemetry tracer，創建和管理 spans，發送追蹤數據到 OTLP endpoint (compatible with Jaeger)
type JaegerTracingProvider struct {
	tracer   trace.Tracer
	provider *sdktrace.TracerProvider
}

// JaegerConfig 定義追蹤提供者的配置
type JaegerConfig struct {
	ServiceName    string  `yaml:"service_name" json:"service_name"`
	ServiceVersion string  `yaml:"service_version" json:"service_version"`
	Environment    string  `yaml:"environment" json:"environment"`
	OTLPEndpoint   string  `yaml:"otlp_endpoint" json:"otlp_endpoint"` // OTLP HTTP endpoint, e.g., "http://localhost:4318"
	SamplingRate   float64 `yaml:"sampling_rate" json:"sampling_rate"`
	Enabled        bool    `yaml:"enabled" json:"enabled"`
}

// NewJaegerTracingProvider 創建新的追蹤提供者實例
func NewJaegerTracingProvider(config JaegerConfig) (contracts.TracingProvider, error) {
	if !config.Enabled {
		return &NoOpTracingProvider{}, nil
	}

	// 創建 OTLP HTTP exporter
	exporter, err := otlptracehttp.New(context.Background(),
		otlptracehttp.WithEndpoint(config.OTLPEndpoint),
		otlptracehttp.WithInsecure(),            // Use HTTP instead of HTTPS for local development
		otlptracehttp.WithURLPath("/v1/traces"), // Specify the correct path
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create OTLP exporter: %w", err)
	}

	// 創建 resource，包含服務資訊
	res := resource.NewWithAttributes(
		semconv.SchemaURL,
		semconv.ServiceNameKey.String(config.ServiceName),
		semconv.ServiceVersionKey.String(config.ServiceVersion),
		semconv.DeploymentEnvironmentKey.String(config.Environment),
	)

	// 創建 TracerProvider
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sdktrace.TraceIDRatioBased(config.SamplingRate)),
	)

	// 設置全局 TracerProvider
	otel.SetTracerProvider(tp)

	// 創建 tracer
	tracer := tp.Tracer(config.ServiceName)

	return &JaegerTracingProvider{
		tracer:   tracer,
		provider: tp,
	}, nil
}

// StartSpan 開始一個新的 span
func (j *JaegerTracingProvider) StartSpan(ctx context.Context, operationName string) (context.Context, contracts.Span) {
	ctx, span := j.tracer.Start(ctx, operationName)
	return ctx, &JaegerSpan{span: span}
}

// GetName 返回提供者名稱
func (j *JaegerTracingProvider) GetName() string {
	return "jaeger_tracing_provider"
}

// Shutdown 優雅關閉追蹤提供者
func (j *JaegerTracingProvider) Shutdown(ctx context.Context) error {
	if j.provider != nil {
		if err := j.provider.Shutdown(ctx); err != nil {
			return fmt.Errorf("failed to shutdown tracer provider: %w", err)
		}
	}
	return nil
}

// JaegerSpan 實現 contracts.Span 介面
type JaegerSpan struct {
	span trace.Span
}

// SetTag 設置 span 標籤
func (s *JaegerSpan) SetTag(key string, value interface{}) {
	switch v := value.(type) {
	case string:
		s.span.SetAttributes(attribute.String(key, v))
	case int:
		s.span.SetAttributes(attribute.Int(key, v))
	case int64:
		s.span.SetAttributes(attribute.Int64(key, v))
	case float64:
		s.span.SetAttributes(attribute.Float64(key, v))
	case bool:
		s.span.SetAttributes(attribute.Bool(key, v))
	default:
		s.span.SetAttributes(attribute.String(key, fmt.Sprintf("%v", v)))
	}
}

// SetError 記錄錯誤到 span
func (s *JaegerSpan) SetError(err error) {
	s.span.RecordError(err)
	s.span.SetStatus(codes.Error, err.Error())
}

// Finish 完成 span
func (s *JaegerSpan) Finish() {
	s.span.End()
}

// NoOpTracingProvider 提供空實現，當追蹤被禁用時使用
type NoOpTracingProvider struct{}

func (n *NoOpTracingProvider) StartSpan(ctx context.Context, operationName string) (context.Context, contracts.Span) {
	return ctx, &NoOpSpan{}
}

func (n *NoOpTracingProvider) GetName() string {
	return "noop_tracing_provider"
}

// NoOpSpan 提供空的 span 實現
type NoOpSpan struct{}

func (s *NoOpSpan) SetTag(key string, value interface{}) {
	// No-op
}

func (s *NoOpSpan) SetError(err error) {
	// No-op
}

func (s *NoOpSpan) Finish() {
	// No-op
}
