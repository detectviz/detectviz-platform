package pluginhost

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

// StartSpan 創建 pluginhost 相關的追蹤 span
func StartSpan(ctx context.Context, name string) (context.Context, func()) {
	tracer := otel.Tracer("go-platform/pluginhost")

	spanCtx, span := tracer.Start(ctx, name,
		trace.WithSpanKind(trace.SpanKindInternal))

	endFunc := func() {
		span.End()
	}

	return spanCtx, endFunc
}
