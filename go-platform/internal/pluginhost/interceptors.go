package pluginhost

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// UnaryMetaInterceptor 取得與透傳必要 metadata（tenant/traceparent/owner 等）
func UnaryMetaInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		if md, ok := metadata.FromIncomingContext(ctx); ok {
			// 未來可擴展：驗證或記錄 metadata 欄位
			_ = md
		}
		return handler(ctx, req)
	}
}
