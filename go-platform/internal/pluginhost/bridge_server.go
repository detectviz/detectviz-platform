package pluginhost

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	grpc_health "google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/metadata"

	v1 "github.com/detectviz/detectviz-platform/contracts/gen/go/detectviz/contracts/v1"
	statuspb "google.golang.org/genproto/googleapis/rpc/status"
)

// Server 實作 ToolBridge，將請求分派至已註冊的 Handler。
type Server struct {
	v1.UnimplementedToolBridgeServiceServer
	reg *Registry
}

func NewServer(reg *Registry) *Server { return &Server{reg: reg} }

// 建立 gRPC 伺服器（可掛載 mTLS 與攔截器）
func NewGRPCServer(tlsCfg *tls.Config, unary ...grpc.UnaryServerInterceptor) *grpc.Server {
	opts := []grpc.ServerOption{}
	if tlsCfg != nil {
		opts = append(opts, grpc.Creds(credentials.NewTLS(tlsCfg)))
	}
	if len(unary) > 0 {
		opts = append(opts, grpc.ChainUnaryInterceptor(unary...))
	}
	return grpc.NewServer(opts...)
}

func (s *Server) Healthz(ctx context.Context, _ *v1.HealthzRequest) (*v1.HealthzResponse, error) {
	return &v1.HealthzResponse{Status: v1.HealthzResponse_SERVING_STATUS_SERVING, Version: "0.1.0"}, nil
}

func (s *Server) Invoke(ctx context.Context, req *v1.InvokeRequest) (*v1.InvokeResponse, error) {
	// 可在此處讀取 metadata（tenant/traceparent/owner 等）
	_, _ = metadata.FromIncomingContext(ctx)

	h, ok := s.reg.Lookup(req.GetToolId())
	if !ok {
		return &v1.InvokeResponse{
			Result: nil,
			Status: &statuspb.Status{Code: 5, Message: "tool_id not found"}, // NOT_FOUND
		}, nil
	}
	return h.Invoke(ctx, req)
}

// Streaming 版（保留擴充）
func (s *Server) InvokeStream(_ *v1.InvokeStreamRequest, _ v1.ToolBridgeService_InvokeStreamServer) error {
	return fmt.Errorf("InvokeStream not implemented")
}

// 監聽與啟動服務
func ListenAndServe(addr string, tlsCfg *tls.Config, reg *Registry, unary ...grpc.UnaryServerInterceptor) error {
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	gs := NewGRPCServer(tlsCfg, unary...)
	hs := grpc_health.NewServer()
	healthpb.RegisterHealthServer(gs, hs)
	v1.RegisterToolBridgeServiceServer(gs, NewServer(reg))
	return gs.Serve(lis)
}
