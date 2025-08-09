package pluginhost

import (
	"context"
	"crypto/tls"
	"time"
)

type Runtime struct {
	addr   string
	tlsCfg *tls.Config
	reg    *Registry
}

func NewRuntime(addr string, tlsCfg *tls.Config, reg *Registry) *Runtime {
	return &Runtime{addr: addr, tlsCfg: tlsCfg, reg: reg}
}

func (rt *Runtime) Start(ctx context.Context) error {
	// TODO: 載入/驗證 module.card.json、容量限制、白名單等
	go func() { _ = ListenAndServe(rt.addr, rt.tlsCfg, rt.reg) }()
	// 等待就緒（健康檢查）
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(200 * time.Millisecond):
		return nil
	}
}

func (rt *Runtime) Stop(_ context.Context) error {
	// TODO: 優雅停機
	return nil
}