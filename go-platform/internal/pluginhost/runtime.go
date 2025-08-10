package pluginhost

import (
	"context"
	"crypto/tls"
	"net"
	"sync"
	"time"

	"google.golang.org/grpc"
	grpc_health "google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"

	contractspb "github.com/detectviz/detectviz-platform/contracts/gen/go/detectviz/contracts/v1"
)

// ReadyFunc 用於在 runtime 完成就緒時通知外部（例如健康檢查）
type ReadyFunc func()

type Runtime struct {
	addr   string
	tlsCfg *tls.Config
	reg    *Registry

	mu   sync.Mutex
	gs   *grpc.Server
	lis  net.Listener
	onReady ReadyFunc
}

func NewRuntime(addr string, tlsCfg *tls.Config, reg *Registry) *Runtime {
	return &Runtime{addr: addr, tlsCfg: tlsCfg, reg: reg}
}

// SetOnReady 設置就緒回呼（可選）
func (rt *Runtime) SetOnReady(f ReadyFunc) { rt.onReady = f }

// Start 啟動 gRPC ToolBridge 服務（阻塞 Serve 放進 goroutine）
func (rt *Runtime) Start(ctx context.Context, unary ...grpc.UnaryServerInterceptor) error {
	rlis, err := net.Listen("tcp", rt.addr)
	if err != nil { return err }
	gs := NewGRPCServer(rt.tlsCfg, unary...)
	contractspb.RegisterToolBridgeServer(gs, NewServer(rt.reg))
	// gRPC Health
	hs := grpc_health.NewServer()
	healthpb.RegisterHealthServer(gs, hs)

	rt.mu.Lock()
	rt.gs = gs
	rt.lis = rlis
	rt.mu.Unlock()

	go func() {
		_ = gs.Serve(rlis)
	}()

	// 簡單就緒等待，亦可改以健康服務事件
	t := time.NewTimer(200 * time.Millisecond)
	select {
	case <-ctx.Done():
		_ = rt.Stop(context.Background())
		return ctx.Err()
	case <-t.C:
		if rt.onReady != nil { rt.onReady() }
		return nil
	}
}

func (rt *Runtime) Stop(_ context.Context) error {
	rt.mu.Lock()
	defer rt.mu.Unlock()
	if rt.gs != nil {
		stopped := make(chan struct{})
		go func() { rt.gs.GracefulStop(); close(stopped) }()
		select {
		case <-stopped:
		case <-time.After(3 * time.Second):
			// 超時則強制停止
			rt.gs.Stop()
		}
	}
	if rt.lis != nil {
		_ = rt.lis.Close()
	}
	return nil
}
