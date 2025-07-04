package tracing

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestNewJaegerTracingProvider(t *testing.T) {
	tests := []struct {
		name    string
		config  JaegerConfig
		wantErr bool
	}{
		{
			name: "有效配置",
			config: JaegerConfig{
				ServiceName:    "test-service",
				ServiceVersion: "1.0.0",
				Environment:    "test",
				OTLPEndpoint:   "localhost:4318",
				SamplingRate:   1.0,
				Enabled:        true,
			},
			wantErr: false,
		},
		{
			name: "禁用追蹤",
			config: JaegerConfig{
				ServiceName: "test-service",
				Enabled:     false,
			},
			wantErr: false,
		},
		{
			name: "無效的 OTLP 端點",
			config: JaegerConfig{
				ServiceName:    "test-service",
				ServiceVersion: "1.0.0",
				Environment:    "test",
				OTLPEndpoint:   "invalid-endpoint",
				SamplingRate:   1.0,
				Enabled:        true,
			},
			wantErr: false, // OTLP exporter 不會立即驗證端點
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider, err := NewJaegerTracingProvider(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewJaegerTracingProvider() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if provider == nil {
					t.Error("期望返回非空的提供者")
				}

				// 驗證提供者名稱
				expectedName := "jaeger_tracing_provider"
				if !tt.config.Enabled {
					expectedName = "noop_tracing_provider"
				}
				if provider.GetName() != expectedName {
					t.Errorf("GetName() = %v, want %v", provider.GetName(), expectedName)
				}
			}
		})
	}
}

func TestJaegerTracingProvider_StartSpan(t *testing.T) {
	// 測試啟用的追蹤提供者
	t.Run("啟用的追蹤提供者", func(t *testing.T) {
		config := JaegerConfig{
			ServiceName:    "test-service",
			ServiceVersion: "1.0.0",
			Environment:    "test",
			OTLPEndpoint:   "localhost:4318",
			SamplingRate:   1.0,
			Enabled:        true,
		}

		provider, err := NewJaegerTracingProvider(config)
		if err != nil {
			t.Fatalf("創建提供者失敗: %v", err)
		}

		ctx := context.Background()
		newCtx, span := provider.StartSpan(ctx, "test-operation")

		if newCtx == nil {
			t.Error("期望返回非空的 context")
		}

		if span == nil {
			t.Error("期望返回非空的 span")
		}

		// 測試 span 操作
		span.SetTag("test-key", "test-value")
		span.SetTag("test-number", 42)
		span.SetTag("test-bool", true)
		span.SetError(errors.New("test error"))
		span.Finish()

		// 清理
		if jaegerProvider, ok := provider.(*JaegerTracingProvider); ok {
			jaegerProvider.Shutdown(context.Background())
		}
	})

	// 測試禁用的追蹤提供者
	t.Run("禁用的追蹤提供者", func(t *testing.T) {
		config := JaegerConfig{
			ServiceName: "test-service",
			Enabled:     false,
		}

		provider, err := NewJaegerTracingProvider(config)
		if err != nil {
			t.Fatalf("創建提供者失敗: %v", err)
		}

		ctx := context.Background()
		newCtx, span := provider.StartSpan(ctx, "test-operation")

		if newCtx == nil {
			t.Error("期望返回非空的 context")
		}

		if span == nil {
			t.Error("期望返回非空的 span")
		}

		// 測試 NoOp span 操作（不應該崩潰）
		span.SetTag("test-key", "test-value")
		span.SetError(errors.New("test error"))
		span.Finish()
	})
}

func TestJaegerSpan_SetTag(t *testing.T) {
	config := JaegerConfig{
		ServiceName:    "test-service",
		ServiceVersion: "1.0.0",
		Environment:    "test",
		OTLPEndpoint:   "localhost:4318",
		SamplingRate:   1.0,
		Enabled:        true,
	}

	provider, err := NewJaegerTracingProvider(config)
	if err != nil {
		t.Fatalf("創建提供者失敗: %v", err)
	}

	ctx := context.Background()
	_, span := provider.StartSpan(ctx, "test-operation")

	tests := []struct {
		name  string
		key   string
		value interface{}
	}{
		{"字符串標籤", "string-key", "string-value"},
		{"整數標籤", "int-key", 42},
		{"長整數標籤", "int64-key", int64(42)},
		{"浮點數標籤", "float64-key", 3.14},
		{"布爾標籤", "bool-key", true},
		{"其他類型標籤", "other-key", []string{"test"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 這些操作不應該崩潰
			span.SetTag(tt.key, tt.value)
		})
	}

	span.Finish()

	// 清理
	if jaegerProvider, ok := provider.(*JaegerTracingProvider); ok {
		jaegerProvider.Shutdown(context.Background())
	}
}

func TestJaegerSpan_SetError(t *testing.T) {
	config := JaegerConfig{
		ServiceName:    "test-service",
		ServiceVersion: "1.0.0",
		Environment:    "test",
		OTLPEndpoint:   "localhost:4318",
		SamplingRate:   1.0,
		Enabled:        true,
	}

	provider, err := NewJaegerTracingProvider(config)
	if err != nil {
		t.Fatalf("創建提供者失敗: %v", err)
	}

	ctx := context.Background()
	_, span := provider.StartSpan(ctx, "test-operation")

	// 測試設置錯誤
	testErr := errors.New("test error message")
	span.SetError(testErr)

	span.Finish()

	// 清理
	if jaegerProvider, ok := provider.(*JaegerTracingProvider); ok {
		jaegerProvider.Shutdown(context.Background())
	}
}

func TestJaegerTracingProvider_Shutdown(t *testing.T) {
	config := JaegerConfig{
		ServiceName:    "test-service",
		ServiceVersion: "1.0.0",
		Environment:    "test",
		OTLPEndpoint:   "localhost:4318",
		SamplingRate:   1.0,
		Enabled:        true,
	}

	provider, err := NewJaegerTracingProvider(config)
	if err != nil {
		t.Fatalf("創建提供者失敗: %v", err)
	}

	jaegerProvider, ok := provider.(*JaegerTracingProvider)
	if !ok {
		t.Fatal("期望 JaegerTracingProvider 類型")
	}

	// 測試正常關閉
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = jaegerProvider.Shutdown(ctx)
	if err != nil {
		t.Errorf("Shutdown() error = %v", err)
	}

	// 測試重複關閉（不應該崩潰）
	err = jaegerProvider.Shutdown(ctx)
	if err != nil {
		t.Errorf("Second Shutdown() error = %v", err)
	}
}

func TestNoOpTracingProvider(t *testing.T) {
	provider := &NoOpTracingProvider{}

	// 測試基本操作
	if provider.GetName() != "noop_tracing_provider" {
		t.Errorf("GetName() = %v, want noop_tracing_provider", provider.GetName())
	}

	ctx := context.Background()
	newCtx, span := provider.StartSpan(ctx, "test-operation")

	if newCtx == nil {
		t.Error("期望返回非空的 context")
	}

	if span == nil {
		t.Error("期望返回非空的 span")
	}

	// 測試 NoOp span 操作（不應該崩潰）
	span.SetTag("test-key", "test-value")
	span.SetError(errors.New("test error"))
	span.Finish()
}

func TestNoOpSpan(t *testing.T) {
	span := &NoOpSpan{}

	// 所有操作都應該是安全的 no-op
	span.SetTag("test-key", "test-value")
	span.SetTag("test-number", 42)
	span.SetTag("test-bool", true)
	span.SetError(errors.New("test error"))
	span.Finish()

	// 如果沒有崩潰，測試就通過了
}

// 基準測試
func BenchmarkJaegerTracingProvider_StartSpan(b *testing.B) {
	config := JaegerConfig{
		ServiceName:    "benchmark-service",
		ServiceVersion: "1.0.0",
		Environment:    "test",
		OTLPEndpoint:   "localhost:4318",
		SamplingRate:   0.1, // 降低採樣率以減少開銷
		Enabled:        true,
	}

	provider, err := NewJaegerTracingProvider(config)
	if err != nil {
		b.Fatalf("創建提供者失敗: %v", err)
	}

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, span := provider.StartSpan(ctx, "benchmark-operation")
		span.SetTag("iteration", i)
		span.Finish()
	}

	// 清理
	if jaegerProvider, ok := provider.(*JaegerTracingProvider); ok {
		jaegerProvider.Shutdown(context.Background())
	}
}

func BenchmarkNoOpTracingProvider_StartSpan(b *testing.B) {
	provider := &NoOpTracingProvider{}
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, span := provider.StartSpan(ctx, "benchmark-operation")
		span.SetTag("iteration", i)
		span.Finish()
	}
}
